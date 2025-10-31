package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"to-de-olho-backend/internal/domain"
	"to-de-olho-backend/internal/infrastructure/resilience"
)

// CamaraAPIPort define interface para API da Câmara para votações
type CamaraAPIPort interface {
	GetVotacoes(ctx context.Context, dataInicio, dataFim time.Time) ([]*domain.Votacao, error)
	GetVotacao(ctx context.Context, id string) (*domain.Votacao, error)
	GetVotosPorVotacao(ctx context.Context, idVotacao string) ([]*domain.VotoDeputado, error)
	GetOrientacoesPorVotacao(ctx context.Context, idVotacao string) ([]*domain.OrientacaoPartido, error)
}

// VotacoesService gerencia operações relacionadas a votações
type VotacoesService struct {
	votacaoRepo  domain.VotacaoRepository
	camaraClient CamaraAPIPort
	cache        CachePort
	logger       *slog.Logger
}

type votacoesProgressContextKey struct{}

var progressReporterKey = votacoesProgressContextKey{}

// WithVotacoesProgressReporter injeta um callback opcional no contexto para acompanhar progresso.
// O service chamará esse callback com o total processado e o total previsto sempre que registrar um progresso relevante.
func WithVotacoesProgressReporter(ctx context.Context, reporter func(processed, total int)) context.Context {
	if ctx == nil || reporter == nil {
		return ctx
	}
	return context.WithValue(ctx, progressReporterKey, reporter)
}

func getVotacoesProgressReporter(ctx context.Context) func(int, int) {
	if ctx == nil {
		return nil
	}
	if v := ctx.Value(progressReporterKey); v != nil {
		if reporter, ok := v.(func(int, int)); ok {
			return reporter
		}
	}
	return nil
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
func (vs *VotacoesService) SincronizarVotacoes(ctx context.Context, dataInicio, dataFim time.Time) (int, error) {
	vs.logger.Info("iniciando sincronização de votações",
		slog.Time("data_inicio", dataInicio),
		slog.Time("data_fim", dataFim))

	// Buscar votações da API da Câmara com tratamento melhorado de circuit breaker
	votacoes, err := vs.camaraClient.GetVotacoes(ctx, dataInicio, dataFim)
	if err != nil {
		if resilience.IsCircuitBreakerOpen(err) {
			vs.logger.Warn("circuit breaker aberto ao buscar votações - pulando período",
				slog.Time("data_inicio", dataInicio),
				slog.Time("data_fim", dataFim),
				slog.String("error", err.Error()))
			return 0, fmt.Errorf("circuit breaker aberto para período %s-%s: %w",
				dataInicio.Format("2006-01-02"), dataFim.Format("2006-01-02"), err)
		}
		vs.logger.Error("erro ao buscar votações da API da Câmara",
			slog.String("error", err.Error()),
			slog.Time("data_inicio", dataInicio),
			slog.Time("data_fim", dataFim))
		return 0, fmt.Errorf("erro ao buscar votações da API da Câmara: %w", err)
	}

	vs.logger.Info("votações obtidas da API da Câmara", slog.Int("count", len(votacoes)))

	processedCount := 0
	skippedByCircuit := 0
	voteErrors := 0
	var breakerErr error
	reporter := getVotacoesProgressReporter(ctx)

	// Sincronizar cada votação
	for _, votacao := range votacoes {
		if err := vs.votacaoRepo.UpsertVotacao(ctx, votacao); err != nil {
			vs.logger.Error("erro ao salvar votação",
				slog.Int64("id", votacao.ID),
				slog.String("error", err.Error()))
			voteErrors++
			continue
		}

		processedCount++
		if reporter != nil {
			reporter(processedCount, len(votacoes))
		}

		// Sincronizar votos dos deputados
		if err := vs.sincronizarVotos(ctx, votacao); err != nil {
			if resilience.IsCircuitBreakerOpen(err) || errors.Is(err, context.DeadlineExceeded) {
				vs.logger.Warn("circuit breaker ativo ao sincronizar votos; interrompendo processamento restante",
					slog.Int64("votacao_id", votacao.ID),
					slog.String("id_camara", votacao.IDCamara),
					slog.String("error", err.Error()))
				skippedByCircuit++
				breakerErr = err
				break
			}

			vs.logger.Error("erro ao sincronizar votos da votação",
				slog.Int64("votacao_id", votacao.ID),
				slog.String("id_camara", votacao.IDCamara),
				slog.String("error", err.Error()))
		}

		// Sincronizar orientações partidárias
		if breakerErr != nil {
			break
		}
		if err := vs.sincronizarOrientacoes(ctx, votacao); err != nil {
			if resilience.IsCircuitBreakerOpen(err) || errors.Is(err, context.DeadlineExceeded) {
				vs.logger.Warn("circuit breaker ativo ao sincronizar orientações; interrompendo processamento restante",
					slog.Int64("votacao_id", votacao.ID),
					slog.String("id_camara", votacao.IDCamara),
					slog.String("error", err.Error()))
				skippedByCircuit++
				breakerErr = err
				break
			}

			vs.logger.Error("erro ao sincronizar orientações da votação",
				slog.Int64("votacao_id", votacao.ID),
				slog.String("id_camara", votacao.IDCamara),
				slog.String("error", err.Error()))
		}
	}

	// Invalidar caches relacionados
	vs.invalidarCachesVotacao(ctx)

	vs.logger.Info("sincronização de votações concluída",
		slog.Int("total_votacoes", len(votacoes)),
		slog.Int("processadas", processedCount),
		slog.Int("erros_votos", voteErrors),
		slog.Int("circuit_breaker_skips", skippedByCircuit))

	if breakerErr != nil {
		return processedCount, fmt.Errorf("circuit breaker interrompeu sincronização de votações: %w", breakerErr)
	}

	return processedCount, nil
}

// sincronizarVotos sincroniza votos dos deputados para uma votação
func (vs *VotacoesService) sincronizarVotos(ctx context.Context, votacao *domain.Votacao) error {
	if votacao == nil {
		return fmt.Errorf("votação inválida para sincronização de votos")
	}
	if votacao.IDCamara == "" {
		return fmt.Errorf("votação %d sem idCamara associado", votacao.ID)
	}

	votos, err := vs.camaraClient.GetVotosPorVotacao(ctx, votacao.IDCamara)
	if err != nil {
		return fmt.Errorf("erro ao buscar votos da votação %s: %w", votacao.IDCamara, err)
	}

	for _, voto := range votos {
		voto.IDVotacao = votacao.ID
		if err := vs.votacaoRepo.CreateVotoDeputado(ctx, voto); err != nil {
			vs.logger.Error("erro ao salvar voto do deputado",
				slog.Int64("votacao_id", votacao.ID),
				slog.String("id_camara", votacao.IDCamara),
				slog.Int("deputado_id", voto.IDDeputado),
				slog.String("error", err.Error()))
		}
	}

	return nil
}

// sincronizarOrientacoes sincroniza orientações partidárias para uma votação
func (vs *VotacoesService) sincronizarOrientacoes(ctx context.Context, votacao *domain.Votacao) error {
	if votacao == nil {
		return fmt.Errorf("votação inválida para sincronização de orientações")
	}
	if votacao.IDCamara == "" {
		return fmt.Errorf("votação %d sem idCamara associado", votacao.ID)
	}

	orientacoes, err := vs.camaraClient.GetOrientacoesPorVotacao(ctx, votacao.IDCamara)
	if err != nil {
		return fmt.Errorf("erro ao buscar orientações da votação %s: %w", votacao.IDCamara, err)
	}

	for _, orientacao := range orientacoes {
		orientacao.IDVotacao = votacao.ID
		if err := vs.votacaoRepo.CreateOrientacaoPartido(ctx, orientacao); err != nil {
			vs.logger.Error("erro ao salvar orientação do partido",
				slog.Int64("votacao_id", votacao.ID),
				slog.String("id_camara", votacao.IDCamara),
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
				slog.String("idCamara", votacao.IDCamara),
				slog.String("titulo", votacao.Titulo),
				slog.String("error", err.Error()))
			continue
		}

		// Sincronizar votos individuais dos deputados para esta votação
		if err := vs.sincronizarVotosPorVotacao(ctx, votacao); err != nil {
			vs.logger.Warn("erro ao sincronizar votos da votação",
				slog.Int64("votacao_id", votacao.ID),
				slog.String("idCamara", votacao.IDCamara),
				slog.String("error", err.Error()))
		}

		// Sincronizar orientações partidárias para esta votação
		if err := vs.sincronizarOrientacoesPorVotacao(ctx, votacao); err != nil {
			vs.logger.Warn("erro ao sincronizar orientações da votação",
				slog.Int64("votacao_id", votacao.ID),
				slog.String("idCamara", votacao.IDCamara),
				slog.String("error", err.Error()))
		}

		totalProcessadas++
		vs.logger.Debug("votação processada",
			slog.Int64("votacao_id", votacao.ID),
			slog.String("idCamara", votacao.IDCamara),
			slog.String("titulo", votacao.Titulo))
	} // Invalidar caches
	vs.invalidarCachesVotacao(ctx)

	vs.logger.Info("sincronização de votações concluída",
		slog.Int("totalEncontradas", len(votacoes)),
		slog.Int("totalProcessadas", totalProcessadas))

	return totalProcessadas, nil
}

// sincronizarVotosPorVotacao sincroniza votos individuais dos deputados
func (vs *VotacoesService) sincronizarVotosPorVotacao(ctx context.Context, votacao *domain.Votacao) error {
	if votacao == nil {
		return fmt.Errorf("votação inválida para sincronização de votos")
	}
	if votacao.IDCamara == "" {
		return fmt.Errorf("votação %d sem idCamara associado", votacao.ID)
	}

	votos, err := vs.camaraClient.GetVotosPorVotacao(ctx, votacao.IDCamara)
	if err != nil {
		return fmt.Errorf("erro ao buscar votos da votação %s: %w", votacao.IDCamara, err)
	}

	for _, voto := range votos {
		voto.IDVotacao = votacao.ID
		if err := vs.votacaoRepo.CreateVotoDeputado(ctx, voto); err != nil {
			vs.logger.Debug("erro ao salvar voto (pode já existir)",
				slog.Int64("votacao_id", votacao.ID),
				slog.String("idCamara", votacao.IDCamara),
				slog.Int("idDeputado", voto.IDDeputado),
				slog.String("error", err.Error()))
		}
	}

	vs.logger.Debug("votos sincronizados",
		slog.Int64("votacao_id", votacao.ID),
		slog.String("idCamara", votacao.IDCamara),
		slog.Int("totalVotos", len(votos)))

	return nil
}

// sincronizarOrientacoesPorVotacao sincroniza orientações partidárias
func (vs *VotacoesService) sincronizarOrientacoesPorVotacao(ctx context.Context, votacao *domain.Votacao) error {
	if votacao == nil {
		return fmt.Errorf("votação inválida para sincronização de orientações")
	}
	if votacao.IDCamara == "" {
		return fmt.Errorf("votação %d sem idCamara associado", votacao.ID)
	}

	orientacoes, err := vs.camaraClient.GetOrientacoesPorVotacao(ctx, votacao.IDCamara)
	if err != nil {
		return fmt.Errorf("erro ao buscar orientações da votação %s: %w", votacao.IDCamara, err)
	}

	for _, orientacao := range orientacoes {
		orientacao.IDVotacao = votacao.ID
		if err := vs.votacaoRepo.CreateOrientacaoPartido(ctx, orientacao); err != nil {
			vs.logger.Debug("erro ao salvar orientação (pode já existir)",
				slog.Int64("votacao_id", votacao.ID),
				slog.String("idCamara", votacao.IDCamara),
				slog.String("partido", orientacao.Partido),
				slog.String("error", err.Error()))
		}
	}

	vs.logger.Debug("orientações sincronizadas",
		slog.Int64("votacao_id", votacao.ID),
		slog.String("idCamara", votacao.IDCamara),
		slog.Int("totalOrientacoes", len(orientacoes)))

	return nil
}
