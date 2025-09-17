package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"to-de-olho-backend/internal/domain"
)

// ProposicoesServiceInterface define o contrato do serviço de proposições
type ProposicoesServiceInterface interface {
	ListarProposicoes(ctx context.Context, filtros *domain.ProposicaoFilter) ([]domain.Proposicao, int, string, error)
	BuscarProposicaoPorID(ctx context.Context, id int) (*domain.Proposicao, string, error)
}

// ProposicaoPort define interface para acesso à API da Câmara
type ProposicaoPort interface {
	FetchProposicoes(ctx context.Context, filtros *domain.ProposicaoFilter) ([]domain.Proposicao, error)
	FetchProposicaoPorID(ctx context.Context, id int) (*domain.Proposicao, error)
}

// ProposicaoRepositoryPort define interface para acesso a dados
type ProposicaoRepositoryPort interface {
	ListProposicoes(ctx context.Context, filtros *domain.ProposicaoFilter) ([]domain.Proposicao, int, error)
	GetProposicaoPorID(ctx context.Context, id int) (*domain.Proposicao, error)
	UpsertProposicoes(ctx context.Context, proposicoes []domain.Proposicao) error
}

// ProposicoesService implementa os casos de uso para proposições
type ProposicoesService struct {
	client ProposicaoPort
	cache  CachePort
	repo   ProposicaoRepositoryPort
	logger *slog.Logger
}

// NewProposicoesService cria uma nova instância do serviço de proposições
func NewProposicoesService(
	client ProposicaoPort,
	cache CachePort,
	repo ProposicaoRepositoryPort,
	logger *slog.Logger,
) *ProposicoesService {
	return &ProposicoesService{
		client: client,
		cache:  cache,
		repo:   repo,
		logger: logger,
	}
}

// ListarProposicoes busca proposições com filtros aplicados
func (s *ProposicoesService) ListarProposicoes(
	ctx context.Context,
	filtros *domain.ProposicaoFilter,
) ([]domain.Proposicao, int, string, error) {
	start := time.Now()

	// Validar e aplicar padrões nos filtros
	if err := filtros.Validate(); err != nil {
		s.logger.Error("filtros inválidos para busca de proposições",
			slog.String("error", err.Error()),
			slog.Any("filtros", filtros))
		return nil, 0, "", fmt.Errorf("filtros inválidos: %w", err)
	}

	filtros.SetDefaults()

	// Tentar buscar no cache primeiro
	cacheKey := BuildProposicoesCacheKey(filtros)
	if cached, ok := s.cache.Get(ctx, cacheKey); ok && cached != "" {
		var result struct {
			Proposicoes []domain.Proposicao `json:"proposicoes"`
			Total       int                 `json:"total"`
		}

		if err := json.Unmarshal([]byte(cached), &result); err == nil {
			s.logger.Info("proposições encontradas no cache",
				slog.String("cache_key", cacheKey),
				slog.Int("total", result.Total),
				slog.Duration("duration", time.Since(start)))

			return result.Proposicoes, result.Total, "cache", nil
		} else {
			s.logger.Warn("erro ao deserializar cache de proposições",
				slog.String("cache_key", cacheKey),
				slog.String("error", err.Error()))
		}
	}

	// Buscar na API da Câmara
	proposicoes, err := s.client.FetchProposicoes(ctx, filtros)
	if err != nil {
		s.logger.Error("erro ao buscar proposições na API da Câmara",
			slog.String("error", err.Error()),
			slog.Any("filtros", filtros),
			slog.Duration("duration", time.Since(start)))

		// Tentar buscar no banco como fallback
		return s.buscarProposicoesRepository(ctx, filtros, start)
	}

	// Validar as proposições retornadas
	proposicoesValidas := make([]domain.Proposicao, 0, len(proposicoes))
	for _, p := range proposicoes {
		if err := p.Validate(); err != nil {
			s.logger.Warn("proposição inválida ignorada",
				slog.Int("id", p.ID),
				slog.String("identificacao", p.GetIdentificacao()),
				slog.String("error", err.Error()))
			continue
		}
		proposicoesValidas = append(proposicoesValidas, p)
	}

	total := len(proposicoesValidas)

	// Salvar no cache
	cacheData := struct {
		Proposicoes []domain.Proposicao `json:"proposicoes"`
		Total       int                 `json:"total"`
	}{
		Proposicoes: proposicoesValidas,
		Total:       total,
	}

	if cacheBytes, err := json.Marshal(cacheData); err == nil {
		// Cache por 5 minutos para listas de proposições
		s.cache.Set(ctx, cacheKey, string(cacheBytes), 5*time.Minute)
	}

	// Salvar no repositório em background (não bloquear a resposta)
	go func() {
		if err := s.repo.UpsertProposicoes(context.Background(), proposicoesValidas); err != nil {
			s.logger.Error("erro ao salvar proposições no repositório",
				slog.String("error", err.Error()),
				slog.Int("total", len(proposicoesValidas)))
		}
	}()

	s.logger.Info("proposições listadas com sucesso",
		slog.Int("total", total),
		slog.String("source", "api"),
		slog.Duration("duration", time.Since(start)))

	return proposicoesValidas, total, "api", nil
}

// BuscarProposicaoPorID busca uma proposição específica por ID
func (s *ProposicoesService) BuscarProposicaoPorID(
	ctx context.Context,
	id int,
) (*domain.Proposicao, string, error) {
	start := time.Now()

	if id <= 0 {
		return nil, "", domain.ErrProposicaoIDInvalido
	}

	// Tentar buscar no cache primeiro
	cacheKey := fmt.Sprintf("proposicao:%d", id)
	if cached, ok := s.cache.Get(ctx, cacheKey); ok && cached != "" {
		var proposicao domain.Proposicao
		if err := json.Unmarshal([]byte(cached), &proposicao); err == nil {
			s.logger.Info("proposição encontrada no cache",
				slog.Int("id", id),
				slog.String("identificacao", proposicao.GetIdentificacao()),
				slog.Duration("duration", time.Since(start)))

			return &proposicao, "cache", nil
		} else {
			s.logger.Warn("erro ao deserializar cache de proposição",
				slog.Int("id", id),
				slog.String("error", err.Error()))
		}
	}

	// Buscar na API da Câmara
	proposicao, err := s.client.FetchProposicaoPorID(ctx, id)
	if err != nil {
		s.logger.Error("erro ao buscar proposição na API da Câmara",
			slog.Int("id", id),
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(start)))

		// Tentar buscar no banco como fallback
		if proposicaoRepo, errRepo := s.repo.GetProposicaoPorID(ctx, id); errRepo == nil {
			s.logger.Info("proposição encontrada no repositório (fallback)",
				slog.Int("id", id),
				slog.String("identificacao", proposicaoRepo.GetIdentificacao()),
				slog.Duration("duration", time.Since(start)))

			return proposicaoRepo, "repository", nil
		}

		return nil, "", fmt.Errorf("erro ao buscar proposição %d: %w", id, err)
	}

	// Validar a proposição
	if err := proposicao.Validate(); err != nil {
		s.logger.Error("proposição retornada da API é inválida",
			slog.Int("id", id),
			slog.String("error", err.Error()))
		return nil, "", fmt.Errorf("proposição %d inválida: %w", id, err)
	}

	// Salvar no cache
	if cacheBytes, err := json.Marshal(proposicao); err == nil {
		// Cache por 10 minutos para proposições individuais
		s.cache.Set(ctx, cacheKey, string(cacheBytes), 10*time.Minute)
	}

	// Salvar no repositório em background
	go func() {
		if err := s.repo.UpsertProposicoes(context.Background(), []domain.Proposicao{*proposicao}); err != nil {
			s.logger.Error("erro ao salvar proposição no repositório",
				slog.Int("id", id),
				slog.String("error", err.Error()))
		}
	}()

	s.logger.Info("proposição encontrada com sucesso",
		slog.Int("id", id),
		slog.String("identificacao", proposicao.GetIdentificacao()),
		slog.String("source", "api"),
		slog.Duration("duration", time.Since(start)))

	return proposicao, "api", nil
}

// buscarProposicoesRepository busca proposições no repositório como fallback
func (s *ProposicoesService) buscarProposicoesRepository(
	ctx context.Context,
	filtros *domain.ProposicaoFilter,
	start time.Time,
) ([]domain.Proposicao, int, string, error) {
	proposicoes, total, err := s.repo.ListProposicoes(ctx, filtros)
	if err != nil {
		s.logger.Error("erro ao buscar proposições no repositório",
			slog.String("error", err.Error()),
			slog.Any("filtros", filtros),
			slog.Duration("duration", time.Since(start)))
		return nil, 0, "", fmt.Errorf("erro ao buscar proposições: %w", err)
	}

	s.logger.Info("proposições encontradas no repositório (fallback)",
		slog.Int("total", total),
		slog.String("source", "repository"),
		slog.Duration("duration", time.Since(start)))

	return proposicoes, total, "repository", nil
}

// BuildProposicoesCacheKey constrói a chave do cache para listas de proposições
func BuildProposicoesCacheKey(filtros *domain.ProposicaoFilter) string {
	key := fmt.Sprintf("proposicoes:p%d:l%d:o%s:op%s",
		filtros.Pagina,
		filtros.Limite,
		filtros.Ordem,
		filtros.OrdenarPor,
	)

	if filtros.SiglaTipo != "" {
		key += fmt.Sprintf(":st%s", filtros.SiglaTipo)
	}

	if filtros.Ano != nil {
		key += fmt.Sprintf(":a%d", *filtros.Ano)
	}

	if filtros.Numero != nil {
		key += fmt.Sprintf(":n%d", *filtros.Numero)
	}

	if filtros.CodSituacao != nil {
		key += fmt.Sprintf(":cs%d", *filtros.CodSituacao)
	}

	if filtros.SiglaUfAutor != "" {
		key += fmt.Sprintf(":uf%s", filtros.SiglaUfAutor)
	}

	if filtros.SiglaPartidoAutor != "" {
		key += fmt.Sprintf(":pt%s", filtros.SiglaPartidoAutor)
	}

	if filtros.Tema != "" {
		key += fmt.Sprintf(":t%s", filtros.Tema)
	}

	return key
}
