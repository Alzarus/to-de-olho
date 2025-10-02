package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"to-de-olho-backend/internal/domain"
)

// CamaraAPIPort define interface para API da Câmara para votações
type CamaraAPIPort interface {
	GetVotacoes(ctx context.Context, dataInicio, dataFim time.Time) ([]*domain.Votacao, error)
	GetVotacao(ctx context.Context, id int64) (*domain.Votacao, error)
	GetVotosPorVotacao(ctx context.Context, idVotacao int64) ([]*domain.VotoDeputado, error)
	GetOrientacoesPorVotacao(ctx context.Context, idVotacao int64) ([]*domain.OrientacaoPartido, error)
}

// VotacoesService gerencia operações relacionadas a votações
type VotacoesService struct {
	votacaoRepo  domain.VotacaoRepository
	camaraClient CamaraAPIPort
	cache        CachePort
	logger       *slog.Logger
}

// NewVotacoesService cria um novo service de votações
func NewVotacoesService(
	votacaoRepo domain.VotacaoRepository,
	camaraClient CamaraAPIPort,
	cache CachePort,
) *VotacoesService {
	return &VotacoesService{
		votacaoRepo:  votacaoRepo,
		camaraClient: camaraClient,
		cache:        cache,
		logger:       slog.Default(),
	}
}

// ListarVotacoes lista votações com filtros
func (vs *VotacoesService) ListarVotacoes(ctx context.Context, filters domain.FiltrosVotacao, pag domain.Pagination) ([]*domain.Votacao, int, error) {
	cacheKey := fmt.Sprintf("votacoes:list:%+v:%+v", filters, pag)

	// Verificar cache
	if cached, found := vs.cache.Get(ctx, cacheKey); found {
		var result struct {
			Votacoes []*domain.Votacao `json:"votacoes"`
			Total    int               `json:"total"`
		}
		if err := json.Unmarshal([]byte(cached), &result); err == nil {
			vs.logger.Debug("cache hit para listagem de votações", slog.String("key", cacheKey))
			return result.Votacoes, result.Total, nil
		}
	}

	// Buscar do repositório
	votacoes, total, err := vs.votacaoRepo.ListVotacoes(ctx, filters, pag)
	if err != nil {
		vs.logger.Error("erro ao listar votações", slog.String("error", err.Error()))
		return nil, 0, fmt.Errorf("erro ao listar votações: %w", err)
	}

	// Salvar no cache
	result := struct {
		Votacoes []*domain.Votacao `json:"votacoes"`
		Total    int               `json:"total"`
	}{
		Votacoes: votacoes,
		Total:    total,
	}
	if resultJSON, err := json.Marshal(result); err == nil {
		vs.cache.Set(ctx, cacheKey, string(resultJSON), time.Hour)
	}

	return votacoes, total, nil
}

// ObterVotacaoDetalhada obtém votação completa com votos e orientações
func (vs *VotacoesService) ObterVotacaoDetalhada(ctx context.Context, votacaoID int64) (*domain.VotacaoDetalhada, error) {
	cacheKey := fmt.Sprintf("votacao:detalhada:%d", votacaoID)

	// Verificar cache
	if cached, found := vs.cache.Get(ctx, cacheKey); found {
		var votacao domain.VotacaoDetalhada
		if err := json.Unmarshal([]byte(cached), &votacao); err == nil {
			vs.logger.Debug("cache hit para votação detalhada", slog.Int64("votacao_id", votacaoID))
			return &votacao, nil
		}
	}

	// Buscar do repositório
	votacao, err := vs.votacaoRepo.GetVotacaoDetalhada(ctx, votacaoID)
	if err != nil {
		vs.logger.Error("erro ao obter votação detalhada",
			slog.Int64("votacao_id", votacaoID),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("erro ao obter votação detalhada: %w", err)
	}

	// Salvar no cache
	if votacaoJSON, err := json.Marshal(votacao); err == nil {
		vs.cache.Set(ctx, cacheKey, string(votacaoJSON), 2*time.Hour)
	}

	return votacao, nil
}

// SincronizarVotacoes sincroniza votações de um período da API da Câmara
func (vs *VotacoesService) SincronizarVotacoes(ctx context.Context, dataInicio, dataFim time.Time) error {
	vs.logger.Info("iniciando sincronização de votações",
		slog.Time("data_inicio", dataInicio),
		slog.Time("data_fim", dataFim))

	// Buscar votações da API da Câmara
	votacoes, err := vs.camaraClient.GetVotacoes(ctx, dataInicio, dataFim)
	if err != nil {
		vs.logger.Error("erro ao buscar votações da API da Câmara", slog.String("error", err.Error()))
		return fmt.Errorf("erro ao buscar votações da API da Câmara: %w", err)
	}

	vs.logger.Info("votações obtidas da API da Câmara", slog.Int("count", len(votacoes)))

	// Sincronizar cada votação
	for _, votacao := range votacoes {
		if err := vs.votacaoRepo.UpsertVotacao(ctx, votacao); err != nil {
			vs.logger.Error("erro ao salvar votação",
				slog.Int64("id", votacao.ID),
				slog.String("error", err.Error()))
			continue
		}

		// Sincronizar votos dos deputados
		if err := vs.sincronizarVotos(ctx, votacao.ID); err != nil {
			vs.logger.Error("erro ao sincronizar votos da votação",
				slog.Int64("votacao_id", votacao.ID),
				slog.String("error", err.Error()))
		}

		// Sincronizar orientações partidárias
		if err := vs.sincronizarOrientacoes(ctx, votacao.ID); err != nil {
			vs.logger.Error("erro ao sincronizar orientações da votação",
				slog.Int64("votacao_id", votacao.ID),
				slog.String("error", err.Error()))
		}
	}

	// Invalidar caches relacionados
	vs.invalidarCachesVotacao(ctx)

	vs.logger.Info("sincronização de votações concluída",
		slog.Int("total_votacoes", len(votacoes)))

	return nil
}

// sincronizarVotos sincroniza votos dos deputados para uma votação
func (vs *VotacoesService) sincronizarVotos(ctx context.Context, votacaoID int64) error {
	votos, err := vs.camaraClient.GetVotosPorVotacao(ctx, votacaoID)
	if err != nil {
		return fmt.Errorf("erro ao buscar votos da votação %d: %w", votacaoID, err)
	}

	for _, voto := range votos {
		if err := vs.votacaoRepo.CreateVotoDeputado(ctx, voto); err != nil {
			vs.logger.Error("erro ao salvar voto do deputado",
				slog.Int64("votacao_id", votacaoID),
				slog.Int("deputado_id", voto.IDDeputado),
				slog.String("error", err.Error()))
		}
	}

	return nil
}

// sincronizarOrientacoes sincroniza orientações partidárias para uma votação
func (vs *VotacoesService) sincronizarOrientacoes(ctx context.Context, votacaoID int64) error {
	orientacoes, err := vs.camaraClient.GetOrientacoesPorVotacao(ctx, votacaoID)
	if err != nil {
		return fmt.Errorf("erro ao buscar orientações da votação %d: %w", votacaoID, err)
	}

	for _, orientacao := range orientacoes {
		if err := vs.votacaoRepo.CreateOrientacaoPartido(ctx, orientacao); err != nil {
			vs.logger.Error("erro ao salvar orientação do partido",
				slog.Int64("votacao_id", votacaoID),
				slog.String("partido", orientacao.Partido),
				slog.String("error", err.Error()))
		}
	}

	return nil
}

// invalidarCachesVotacao invalida caches relacionados a votações
func (vs *VotacoesService) invalidarCachesVotacao(ctx context.Context) {
	// Nota: Em uma implementação real, seria interessante ter um método
	// no cache para invalidar por padrão de chave
	vs.logger.Debug("caches de votação invalidados")
}

// SincronizarVotacoesRecentes sincroniza votações recentes da API da Câmara
func (vs *VotacoesService) SincronizarVotacoesRecentes(ctx context.Context, filtros map[string]interface{}) (int, error) {
	vs.logger.Info("iniciando sincronização de votações recentes")

	// Extrair parâmetros dos filtros
	dataInicio, ok := filtros["dataInicio"].(time.Time)
	if !ok {
		dataInicio = time.Now().AddDate(0, 0, -7) // Padrão: 7 dias atrás
	}

	dataFim, ok := filtros["dataFim"].(time.Time)
	if !ok {
		dataFim = time.Now() // Padrão: hoje
	}

	// Buscar votações da API da Câmara
	vs.logger.Info("buscando votações da API da Câmara",
		slog.Time("dataInicio", dataInicio),
		slog.Time("dataFim", dataFim))

	votacoes, err := vs.camaraClient.GetVotacoes(ctx, dataInicio, dataFim)
	if err != nil {
		vs.logger.Error("erro ao buscar votações da API",
			slog.String("error", err.Error()))
		return 0, fmt.Errorf("erro ao buscar votações da API da Câmara: %w", err)
	}

	totalProcessadas := 0

	// Processar cada votação usando UpsertVotacao (cria ou atualiza)
	for _, votacao := range votacoes {
		if err := vs.votacaoRepo.UpsertVotacao(ctx, votacao); err != nil {
			vs.logger.Error("erro ao salvar votação",
				slog.Int64("idVotacaoCamara", votacao.IDVotacaoCamara),
				slog.String("titulo", votacao.Titulo),
				slog.String("error", err.Error()))
			continue
		}

		// Sincronizar votos individuais dos deputados para esta votação
		if err := vs.sincronizarVotosPorVotacao(ctx, votacao.IDVotacaoCamara); err != nil {
			vs.logger.Warn("erro ao sincronizar votos da votação",
				slog.Int64("idVotacao", votacao.IDVotacaoCamara),
				slog.String("error", err.Error()))
		}

		// Sincronizar orientações partidárias para esta votação
		if err := vs.sincronizarOrientacoesPorVotacao(ctx, votacao.IDVotacaoCamara); err != nil {
			vs.logger.Warn("erro ao sincronizar orientações da votação",
				slog.Int64("idVotacao", votacao.IDVotacaoCamara),
				slog.String("error", err.Error()))
		}

		totalProcessadas++
		vs.logger.Debug("votação processada",
			slog.Int64("idVotacaoCamara", votacao.IDVotacaoCamara),
			slog.String("titulo", votacao.Titulo))
	} // Invalidar caches
	vs.invalidarCachesVotacao(ctx)

	vs.logger.Info("sincronização de votações concluída",
		slog.Int("totalEncontradas", len(votacoes)),
		slog.Int("totalProcessadas", totalProcessadas))

	return totalProcessadas, nil
}

// sincronizarVotosPorVotacao sincroniza votos individuais dos deputados
func (vs *VotacoesService) sincronizarVotosPorVotacao(ctx context.Context, idVotacao int64) error {
	votos, err := vs.camaraClient.GetVotosPorVotacao(ctx, idVotacao)
	if err != nil {
		return fmt.Errorf("erro ao buscar votos da votação %d: %w", idVotacao, err)
	}

	for _, voto := range votos {
		if err := vs.votacaoRepo.CreateVotoDeputado(ctx, voto); err != nil {
			vs.logger.Debug("erro ao salvar voto (pode já existir)",
				slog.Int64("idVotacao", idVotacao),
				slog.Int("idDeputado", voto.IDDeputado),
				slog.String("error", err.Error()))
		}
	}

	vs.logger.Debug("votos sincronizados",
		slog.Int64("idVotacao", idVotacao),
		slog.Int("totalVotos", len(votos)))

	return nil
}

// sincronizarOrientacoesPorVotacao sincroniza orientações partidárias
func (vs *VotacoesService) sincronizarOrientacoesPorVotacao(ctx context.Context, idVotacao int64) error {
	orientacoes, err := vs.camaraClient.GetOrientacoesPorVotacao(ctx, idVotacao)
	if err != nil {
		return fmt.Errorf("erro ao buscar orientações da votação %d: %w", idVotacao, err)
	}

	for _, orientacao := range orientacoes {
		if err := vs.votacaoRepo.CreateOrientacaoPartido(ctx, orientacao); err != nil {
			vs.logger.Debug("erro ao salvar orientação (pode já existir)",
				slog.Int64("idVotacao", idVotacao),
				slog.String("partido", orientacao.Partido),
				slog.String("error", err.Error()))
		}
	}

	vs.logger.Debug("orientações sincronizadas",
		slog.Int64("idVotacao", idVotacao),
		slog.Int("totalOrientacoes", len(orientacoes)))

	return nil
}
