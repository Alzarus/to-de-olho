package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"to-de-olho-backend/internal/domain"
	"to-de-olho-backend/internal/pkg/metrics"
)

const (
	despesasSourceAPI              = "api"
	despesasSourceCache            = "cache"
	despesasSourceDatabase         = "database"
	despesasSourceDatabaseFallback = "database_fallback"
)

// DeputadosServiceInterface define o contrato público do serviço de deputados.
type DeputadosServiceInterface interface {
	ListarDeputados(ctx context.Context, partido, uf, nome string) ([]domain.Deputado, string, error)
	BuscarDeputadoPorID(ctx context.Context, id string) (*domain.Deputado, string, error)
	ListarDespesas(ctx context.Context, deputadoID, ano string) ([]domain.Despesa, string, error)
}

// CamaraPort representa a integração com a API da Câmara.
type CamaraPort interface {
	FetchDeputados(ctx context.Context, partido, uf, nome string) ([]domain.Deputado, error)
	FetchDeputadoByID(ctx context.Context, id string) (*domain.Deputado, error)
	FetchDespesas(ctx context.Context, deputadoID, ano string) ([]domain.Despesa, error)
}

// CachePort representa o cache utilizado pelo serviço.
type CachePort interface {
	Get(ctx context.Context, key string) (string, bool)
	Set(ctx context.Context, key, value string, ttl time.Duration)
}

// DeputadoRepositoryPort representa o armazenamento local de deputados.
type DeputadoRepositoryPort interface {
	UpsertDeputados(ctx context.Context, deps []domain.Deputado) error
	ListFromCache(ctx context.Context, limit int) ([]domain.Deputado, error)
}

// DespesaRepositoryPort representa o armazenamento local de despesas.
type DespesaRepositoryPort interface {
	UpsertDespesas(ctx context.Context, deputadoID int, ano int, despesas []domain.Despesa) error
	ListDespesasByDeputadoAno(ctx context.Context, deputadoID int, ano int) ([]domain.Despesa, error)
}

// DeputadosService concentra as operações relacionadas a deputados e suas despesas.
type DeputadosService struct {
	client      CamaraPort
	cache       CachePort
	repo        DeputadoRepositoryPort
	despesaRepo DespesaRepositoryPort
	logger      *slog.Logger
}

// NewDeputadosService cria uma nova instância do serviço.
func NewDeputadosService(
	client CamaraPort,
	cache CachePort,
	repo DeputadoRepositoryPort,
	despesaRepo DespesaRepositoryPort,
) *DeputadosService {
	return &DeputadosService{
		client:      client,
		cache:       cache,
		repo:        repo,
		despesaRepo: despesaRepo,
		logger:      slog.Default(),
	}
}

// ListarDeputados retorna a lista de deputados levando em conta cache e fallback local.
func (s *DeputadosService) ListarDeputados(ctx context.Context, partido, uf, nome string) ([]domain.Deputado, string, error) {
	cacheKey := BuildDeputadosCacheKey(partido, uf, nome)
	if s.cache != nil {
		if v, ok := s.cache.Get(ctx, cacheKey); ok && v != "" {
			var cached []domain.Deputado
			if err := json.Unmarshal([]byte(v), &cached); err == nil {
				return cached, "cache", nil
			} else {
				s.logger.Warn("erro ao desserializar cache de deputados",
					slog.String("cache_key", cacheKey),
					slog.String("error", err.Error()))
			}
		}
	}

	deputados, err := s.client.FetchDeputados(ctx, partido, uf, nome)
	if err != nil {
		if s.repo != nil {
			if cached, errRepo := s.repo.ListFromCache(ctx, 100); errRepo == nil && len(cached) > 0 {
				return cached, "fallback-db", nil
			}
		}
		return nil, "", err
	}

	if s.repo != nil {
		if err := s.repo.UpsertDeputados(ctx, deputados); err != nil {
			s.logger.Warn("erro ao persistir deputados", slog.String("error", err.Error()))
		}
	}

	if s.cache != nil {
		if b, err := json.Marshal(deputados); err == nil {
			s.cache.Set(ctx, cacheKey, string(b), 2*time.Minute)
		}
	}

	return deputados, "api", nil
}

// BuildDeputadosCacheKey centraliza o formato da chave de cache de listagem para evitar divergências.
func BuildDeputadosCacheKey(partido, uf, nome string) string {
	return fmt.Sprintf("deputados:p=%s:u=%s:n=%s", partido, uf, nome)
}

// BuscarDeputadoPorID busca um deputado específico pelo ID informado.
func (s *DeputadosService) BuscarDeputadoPorID(ctx context.Context, id string) (*domain.Deputado, string, error) {
	if id == "" {
		return nil, "", errors.New("ID do deputado é obrigatório")
	}

	cacheKey := "deputado:" + id
	if s.cache != nil {
		if v, ok := s.cache.Get(ctx, cacheKey); ok && v != "" {
			var dep domain.Deputado
			if err := json.Unmarshal([]byte(v), &dep); err == nil {
				return &dep, "cache", nil
			} else {
				s.logger.Warn("erro ao desserializar cache de deputado",
					slog.String("cache_key", cacheKey),
					slog.String("error", err.Error()))
			}
		}
	}

	dep, err := s.client.FetchDeputadoByID(ctx, id)
	if err != nil {
		return nil, "", err
	}

	if s.cache != nil {
		if b, err := json.Marshal(dep); err == nil {
			s.cache.Set(ctx, cacheKey, string(b), 5*time.Minute)
		}
	}

	return dep, "api", nil
}

// ListarDespesas retorna as despesas do deputado considerando cache, repositório e API.
func (s *DeputadosService) ListarDespesas(ctx context.Context, id, ano string) ([]domain.Despesa, string, error) {
	deputadoID, err := strconv.Atoi(id)
	if err != nil {
		return nil, "", fmt.Errorf("ID de deputado inválido: %s", id)
	}

	anoInt, err := strconv.Atoi(ano)
	if err != nil {
		return nil, "", fmt.Errorf("ano inválido: %s", ano)
	}

	forceRemote := domain.ShouldForceDespesaRemote(ctx)
	skipPersist := domain.ShouldSkipDespesaPersist(ctx)
	cacheKey := fmt.Sprintf("despesas:%s:%s", id, ano)

	var (
		dbDespesas []domain.Despesa
		dbErr      error
	)

	if s.despesaRepo != nil {
		dbDespesas, dbErr = s.despesaRepo.ListDespesasByDeputadoAno(ctx, deputadoID, anoInt)
		if dbErr != nil {
			s.logger.Warn("falha ao consultar despesas no repositório",
				slog.Int("deputado_id", deputadoID),
				slog.Int("ano", anoInt),
				slog.String("error", dbErr.Error()))
		} else if len(dbDespesas) > 0 && !forceRemote {
			return dbDespesas, despesasSourceDatabase, nil
		}
	}

	if !forceRemote && s.cache != nil {
		if v, ok := s.cache.Get(ctx, cacheKey); ok && v != "" {
			var cached []domain.Despesa
			if err := json.Unmarshal([]byte(v), &cached); err == nil {
				return cached, despesasSourceCache, nil
			} else {
				s.logger.Warn("erro ao desserializar despesas do cache",
					slog.String("cache_key", cacheKey),
					slog.String("error", err.Error()))
			}
		}
	}

	despesas, err := s.client.FetchDespesas(ctx, id, ano)
	if err != nil {
		if dbErr == nil && s.despesaRepo != nil {
			metrics.IncDespesasFallback(fmt.Sprintf("%d-%d", deputadoID, anoInt))
			if dbDespesas == nil {
				dbDespesas = []domain.Despesa{}
			}
			s.logger.Warn("API da Câmara indisponível, retornando despesas do banco",
				slog.Int("deputado_id", deputadoID),
				slog.Int("ano", anoInt),
				slog.String("error", err.Error()))
			return dbDespesas, despesasSourceDatabaseFallback, nil
		}
		return nil, "", fmt.Errorf("erro ao buscar despesas na API da Câmara: %w", err)
	}

	if s.despesaRepo != nil && len(despesas) > 0 && !skipPersist {
		if err := s.despesaRepo.UpsertDespesas(ctx, deputadoID, anoInt, despesas); err != nil {
			s.logger.Warn("erro ao persistir despesas",
				slog.Int("deputado_id", deputadoID),
				slog.Int("ano", anoInt),
				slog.String("error", err.Error()))
		}
	}

	if s.cache != nil {
		if b, err := json.Marshal(despesas); err == nil {
			s.cache.Set(ctx, cacheKey, string(b), time.Minute)
		}
	}

	return despesas, despesasSourceAPI, nil
}
