package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"to-de-olho-backend/internal/domain"
)

// Service interface - define o contrato do serviço
type DeputadosServiceInterface interface {
	ListarDeputados(ctx context.Context, partido, uf, nome string) ([]domain.Deputado, string, error)
	BuscarDeputadoPorID(ctx context.Context, id string) (*domain.Deputado, string, error)
	ListarDespesas(ctx context.Context, deputadoID, ano string) ([]domain.Despesa, string, error)
}

// Ports (interfaces)
type CamaraPort interface {
	FetchDeputados(ctx context.Context, partido, uf, nome string) ([]domain.Deputado, error)
	FetchDeputadoByID(ctx context.Context, id string) (*domain.Deputado, error)
	FetchDespesas(ctx context.Context, deputadoID, ano string) ([]domain.Despesa, error)
}

type CachePort interface {
	Get(ctx context.Context, key string) (string, bool)
	Set(ctx context.Context, key, value string, ttl time.Duration)
}

type DeputadoRepositoryPort interface {
	UpsertDeputados(ctx context.Context, deps []domain.Deputado) error
	ListFromCache(ctx context.Context, limit int) ([]domain.Deputado, error)
}

type DespesaRepositoryPort interface {
	UpsertDespesas(ctx context.Context, deputadoID int, ano int, despesas []domain.Despesa) error
	ListDespesasByDeputadoAno(ctx context.Context, deputadoID int, ano int) ([]domain.Despesa, error)
}

type DeputadosService struct {
	client      CamaraPort
	cache       CachePort
	repo        DeputadoRepositoryPort
	despesaRepo DespesaRepositoryPort
}

func NewDeputadosService(client CamaraPort, cache CachePort, repo DeputadoRepositoryPort, despesaRepo DespesaRepositoryPort) *DeputadosService {
	return &DeputadosService{
		client:      client,
		cache:       cache,
		repo:        repo,
		despesaRepo: despesaRepo,
	}
}

func (s *DeputadosService) ListarDeputados(ctx context.Context, partido, uf, nome string) ([]domain.Deputado, string, error) {
	cacheKey := BuildDeputadosCacheKey(partido, uf, nome)
	if v, ok := s.cache.Get(ctx, cacheKey); ok && v != "" {
		var cached []domain.Deputado
		if err := json.Unmarshal([]byte(v), &cached); err == nil {
			return cached, "cache", nil
		}
	}

	deputados, err := s.client.FetchDeputados(ctx, partido, uf, nome)
	if err != nil {
		// Fallback: tentar ler do Postgres via repositório local
		if s.repo != nil {
			if cached, err2 := s.repo.ListFromCache(ctx, 100); err2 == nil && len(cached) > 0 {
				return cached, "fallback-db", nil
			}
		}
		return nil, "", err
	}
	_ = s.repo.UpsertDeputados(ctx, deputados)
	if b, err := json.Marshal(deputados); err == nil {
		s.cache.Set(ctx, cacheKey, string(b), 2*time.Minute)
	}
	return deputados, "api", nil
}

// BuildDeputadosCacheKey centraliza o formato da chave de cache de listagem
// para evitar divergência entre implementação e testes.
func BuildDeputadosCacheKey(partido, uf, nome string) string {
	return fmt.Sprintf("deputados:p=%s:u=%s:n=%s", partido, uf, nome)
}

func (s *DeputadosService) BuscarDeputadoPorID(ctx context.Context, id string) (*domain.Deputado, string, error) {
	// Validar ID não vazio
	if id == "" {
		return nil, "", errors.New("ID do deputado é obrigatório")
	}

	if v, ok := s.cache.Get(ctx, "deputado:"+id); ok && v != "" {
		var d domain.Deputado
		if err := json.Unmarshal([]byte(v), &d); err == nil {
			return &d, "cache", nil
		}
	}
	dep, err := s.client.FetchDeputadoByID(ctx, id)
	if err != nil {
		return nil, "", err
	}
	if b, err := json.Marshal(dep); err == nil {
		s.cache.Set(ctx, "deputado:"+id, string(b), 5*time.Minute)
	}
	return dep, "api", nil
}

func (s *DeputadosService) ListarDespesas(ctx context.Context, id, ano string) ([]domain.Despesa, string, error) {
	deputadoID := 0
	if _, err := fmt.Sscanf(id, "%d", &deputadoID); err != nil {
		return nil, "", fmt.Errorf("ID de deputado inválido: %s", id)
	}

	anoInt := 0
	if _, err := fmt.Sscanf(ano, "%d", &anoInt); err != nil {
		return nil, "", fmt.Errorf("ano inválido: %s", ano)
	}

	// Prioridade 1: Verificar no banco de dados (source of truth)
	if s.despesaRepo != nil {
		despesas, err := s.despesaRepo.ListDespesasByDeputadoAno(ctx, deputadoID, anoInt)
		if err == nil && len(despesas) > 0 {
			return despesas, "database", nil
		}
		// Se erro ou vazio, continua para próxima source
	}

	// Prioridade 2: Verificar cache Redis
	cacheKey := "despesas:" + id + ":" + ano
	if v, ok := s.cache.Get(ctx, cacheKey); ok && v != "" {
		var cached []domain.Despesa
		if err := json.Unmarshal([]byte(v), &cached); err == nil {
			return cached, "cache", nil
		}
	}

	// Prioridade 3: Buscar na API da Câmara (fallback)
	despesas, err := s.client.FetchDespesas(ctx, id, ano)
	if err != nil {
		return nil, "", fmt.Errorf("erro ao buscar despesas na API da Câmara: %w", err)
	}

	// Armazenar no banco para próximas consultas
	if s.despesaRepo != nil && len(despesas) > 0 {
		if err := s.despesaRepo.UpsertDespesas(ctx, deputadoID, anoInt, despesas); err != nil {
			// Log error mas não falha a requisição
			fmt.Printf("Aviso: erro ao salvar despesas no banco: %v\n", err)
		}
	}

	// Armazenar no cache
	if b, err := json.Marshal(despesas); err == nil {
		s.cache.Set(ctx, cacheKey, string(b), time.Minute)
	}

	return despesas, "api", nil
}
