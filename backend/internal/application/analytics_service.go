package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"to-de-olho-backend/internal/domain"
)

// Interfaces para repositórios
type DeputadoRepositoryInterface interface {
	ListFromCache(ctx context.Context, limit int) ([]domain.Deputado, error)
	UpsertDeputados(ctx context.Context, deps []domain.Deputado) error
}

type ProposicaoRepositoryInterface interface {
	ListProposicoes(ctx context.Context, filtros *domain.ProposicaoFilter) ([]domain.Proposicao, int, error)
	UpsertProposicoes(ctx context.Context, proposicoes []domain.Proposicao) error
	// GetProposicoesCountByDeputadoAno retorna contagens por deputado para um ano
	GetProposicoesCountByDeputadoAno(ctx context.Context, ano int) ([]domain.ProposicaoCount, error)
}

// DespesaRepositoryInterface define o contrato necessário para despesas usado pelo Analytics
type DespesaRepositoryInterface interface {
	ListDespesasByDeputadoAno(ctx context.Context, deputadoID int, ano int) ([]domain.Despesa, error)
	GetDespesasStats(ctx context.Context, deputadoID int, ano int) (*domain.DespesaStats, error)
}

// AnalyticsServiceInterface define o contrato para o serviço de analytics
type AnalyticsServiceInterface interface {
	// Rankings
	GetRankingGastos(ctx context.Context, ano int, limite int) (*RankingGastos, string, error)
	GetRankingProposicoes(ctx context.Context, ano int, limite int) (*RankingProposicoes, string, error)
	GetRankingPresenca(ctx context.Context, ano int, limite int) (*RankingPresenca, string, error)

	// Insights gerais
	GetInsightsGerais(ctx context.Context) (*InsightsGerais, string, error)

	// Atualização de rankings
	AtualizarRankings(ctx context.Context) error

	// Votações
	GetRankingDeputadosVotacao(ctx context.Context, ano int, limite int) ([]domain.RankingDeputadoVotacao, string, error)
	GetRankingPartidosDisciplina(ctx context.Context, ano int) ([]domain.VotacaoPartido, string, error)
	GetStatsVotacoes(ctx context.Context, periodo string) (*domain.VotacaoStats, string, error)
}

// Estruturas dos Rankings
type RankingGastos struct {
	Ano               int                     `json:"ano"`
	TotalGeral        float64                 `json:"total_geral"`
	MediaGastos       float64                 `json:"media_gastos"`
	Deputados         []DeputadoRankingGastos `json:"deputados"`
	UltimaAtualizacao time.Time               `json:"ultima_atualizacao"`
}

type DeputadoRankingGastos struct {
	ID              int     `json:"id"`
	Nome            string  `json:"nome"`
	Partido         string  `json:"partido"`
	UF              string  `json:"uf"`
	TotalGasto      float64 `json:"total_gasto"`
	Posicao         int     `json:"posicao"`
	PercentualMedia float64 `json:"percentual_media"` // % acima/abaixo da média
}

type RankingProposicoes struct {
	Ano               int                          `json:"ano"`
	TotalGeral        int                          `json:"total_geral"`
	MediaProposicoes  float64                      `json:"media_proposicoes"`
	Deputados         []DeputadoRankingProposicoes `json:"deputados"`
	UltimaAtualizacao time.Time                    `json:"ultima_atualizacao"`
}

type DeputadoRankingProposicoes struct {
	ID               int     `json:"id"`
	Nome             string  `json:"nome"`
	Partido          string  `json:"partido"`
	UF               string  `json:"uf"`
	TotalProposicoes int     `json:"total_proposicoes"`
	Posicao          int     `json:"posicao"`
	PercentualMedia  float64 `json:"percentual_media"`
}

type RankingPresenca struct {
	Ano               int                       `json:"ano"`
	TotalSessoes      int                       `json:"total_sessoes"`
	MediaPresenca     float64                   `json:"media_presenca"`
	Deputados         []DeputadoRankingPresenca `json:"deputados"`
	UltimaAtualizacao time.Time                 `json:"ultima_atualizacao"`
}

type DeputadoRankingPresenca struct {
	ID                 int     `json:"id"`
	Nome               string  `json:"nome"`
	Partido            string  `json:"partido"`
	UF                 string  `json:"uf"`
	SessoesPresente    int     `json:"sessoes_presente"`
	SessoesFaltou      int     `json:"sessoes_faltou"`
	PercentualPresenca float64 `json:"percentual_presenca"`
	Posicao            int     `json:"posicao"`
}

type InsightsGerais struct {
	TotalDeputados      int       `json:"total_deputados"`
	TotalGastoAno       float64   `json:"total_gasto_ano"`
	TotalProposicoesAno int       `json:"total_proposicoes_ano"`
	MediaGastosDeputado float64   `json:"media_gastos_deputado"`
	PartidoMaiorGasto   string    `json:"partido_maior_gasto"`
	UFMaiorGasto        string    `json:"uf_maior_gasto"`
	UltimaAtualizacao   time.Time `json:"ultima_atualizacao"`
}

// AnalyticsService implementa o serviço de analytics usando dados internos
type AnalyticsService struct {
	deputadoRepo   DeputadoRepositoryInterface
	proposicaoRepo ProposicaoRepositoryInterface
	votacaoRepo    domain.VotacaoRepository
	despesaRepo    DespesaRepositoryInterface
	cache          CachePort
	logger         *slog.Logger
}

// NewAnalyticsService cria uma nova instância do serviço de analytics
func NewAnalyticsService(
	deputadoRepo DeputadoRepositoryInterface,
	proposicaoRepo ProposicaoRepositoryInterface,
	votacaoRepo domain.VotacaoRepository,
	despesaRepo DespesaRepositoryInterface,
	cache CachePort,
	logger *slog.Logger,
) *AnalyticsService {
	return &AnalyticsService{
		deputadoRepo:   deputadoRepo,
		proposicaoRepo: proposicaoRepo,
		votacaoRepo:    votacaoRepo,
		despesaRepo:    despesaRepo,
		cache:          cache,
		logger:         logger,
	}
}

// GetRankingGastos retorna ranking de gastos dos deputados
func (s *AnalyticsService) GetRankingGastos(ctx context.Context, ano int, limite int) (*RankingGastos, string, error) {
	if ano < 2000 || ano > time.Now().Year() {
		return nil, "", domain.ErrProposicaoAnoInvalido
	}

	cacheKey := fmt.Sprintf("ranking:gastos:%d:%d", ano, limite)

	// Tentar buscar do cache primeiro
	if cached, ok := s.cache.Get(ctx, cacheKey); ok && cached != "" {
		var ranking RankingGastos
		if err := json.Unmarshal([]byte(cached), &ranking); err == nil {
			return &ranking, "cache", nil
		}
	}

	// Criar contexto com timeout para evitar travamento
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Buscar dados dos deputados do nosso banco
	// Usar limite de 600 para cobrir todos os deputados (513 + margem)
	deputadosCache, err := s.deputadoRepo.ListFromCache(timeoutCtx, 600)
	if err != nil {
		return nil, "", fmt.Errorf("erro ao buscar deputados do banco: %w", err)
	}

	// Converter para slice de ponteiros para manter compatibilidade
	deputados := make([]*domain.Deputado, len(deputadosCache))
	for i := range deputadosCache {
		deputados[i] = &deputadosCache[i]
	}

	s.logger.Info("calculando ranking de gastos com dados internos",
		slog.Int("deputados_count", len(deputados)),
		slog.Int("ano", ano))

	// Calcular gastos para cada deputado com processamento otimizado
	deputadosRanking := make([]DeputadoRankingGastos, 0, len(deputados))
	var totalGeral float64

	// Processar usando DespesaRepository para obter despesas por deputado
	for _, deputado := range deputados {
		// Consultar estatísticas de despesas por deputado
		var totalGasto float64
		if s.despesaRepo != nil {
			if stats, err := s.despesaRepo.GetDespesasStats(timeoutCtx, deputado.ID, ano); err == nil && stats != nil {
				totalGasto = stats.TotalValor
			} else {
				s.logger.Debug("erro ao obter despesas stats, fallback para 0",
					slog.Int("deputado_id", deputado.ID),
					slog.String("error", func() string {
						if err != nil {
							return err.Error()
						}
						return ""
					}()))
			}
		}

		deputadosRanking = append(deputadosRanking, DeputadoRankingGastos{
			ID:         deputado.ID,
			Nome:       deputado.Nome,
			Partido:    deputado.Partido,
			UF:         deputado.UF,
			TotalGasto: totalGasto,
		})
		totalGeral += totalGasto
	}

	// Calcular média (verificar se há deputados para evitar divisão por zero)
	var mediaGastos float64
	if len(deputadosRanking) > 0 {
		mediaGastos = totalGeral / float64(len(deputadosRanking))
	}

	// Ordenar por gasto (maior para menor)
	sort.Slice(deputadosRanking, func(i, j int) bool {
		return deputadosRanking[i].TotalGasto > deputadosRanking[j].TotalGasto
	})

	// Aplicar posições e calcular percentual
	for i := range deputadosRanking {
		deputadosRanking[i].Posicao = i + 1
		if mediaGastos > 0 {
			deputadosRanking[i].PercentualMedia = ((deputadosRanking[i].TotalGasto - mediaGastos) / mediaGastos) * 100
		}
	}

	// Aplicar limite
	if limite > 0 && limite < len(deputadosRanking) {
		deputadosRanking = deputadosRanking[:limite]
	}

	ranking := &RankingGastos{
		Ano:               ano,
		TotalGeral:        totalGeral,
		MediaGastos:       mediaGastos,
		Deputados:         deputadosRanking,
		UltimaAtualizacao: time.Now(),
	}

	// Salvar no cache por 1 hora
	if data, err := json.Marshal(ranking); err == nil {
		s.cache.Set(ctx, cacheKey, string(data), time.Hour)
	}

	return ranking, "computed", nil
}

// GetRankingProposicoes retorna ranking de proposições dos deputados
func (s *AnalyticsService) GetRankingProposicoes(ctx context.Context, ano int, limite int) (*RankingProposicoes, string, error) {
	cacheKey := fmt.Sprintf("ranking:proposicoes:%d:%d", ano, limite)

	// Tentar buscar do cache primeiro
	if cached, ok := s.cache.Get(ctx, cacheKey); ok && cached != "" {
		var ranking RankingProposicoes
		if err := json.Unmarshal([]byte(cached), &ranking); err == nil {
			return &ranking, "cache", nil
		}
	}

	// Buscar dados dos deputados do banco
	deputadosCache, err := s.deputadoRepo.ListFromCache(ctx, 600)
	if err != nil {
		return nil, "", fmt.Errorf("erro ao buscar deputados: %w", err)
	}

	// Converter para slice de ponteiros para manter compatibilidade
	deputados := make([]*domain.Deputado, len(deputadosCache))
	for i := range deputadosCache {
		deputados[i] = &deputadosCache[i]
	}

	// Usar repository aggregation para obter contagem de proposições por deputado
	counts := make(map[int]int)
	if s.proposicaoRepo != nil {
		if rows, err := s.proposicaoRepo.GetProposicoesCountByDeputadoAno(ctx, ano); err == nil {
			for _, r := range rows {
				counts[r.IDDeputado] = r.Count
			}
		} else {
			s.logger.Debug("erro ao obter contagens de proposições, fallback para 0",
				slog.String("error", err.Error()))
		}
	}

	deputadosRanking := make([]DeputadoRankingProposicoes, 0, len(deputados))
	var totalGeral int

	for _, deputado := range deputados {
		totalProposicoes := counts[deputado.ID]
		deputadosRanking = append(deputadosRanking, DeputadoRankingProposicoes{
			ID:               deputado.ID,
			Nome:             deputado.Nome,
			Partido:          deputado.Partido,
			UF:               deputado.UF,
			TotalProposicoes: totalProposicoes,
		})
		totalGeral += totalProposicoes
	}

	// Calcular média (verificar se há deputados para evitar divisão por zero)
	var mediaProposicoes float64
	if len(deputadosRanking) > 0 {
		mediaProposicoes = float64(totalGeral) / float64(len(deputadosRanking))
	}

	// Ordenar por proposições (maior para menor)
	sort.Slice(deputadosRanking, func(i, j int) bool {
		return deputadosRanking[i].TotalProposicoes > deputadosRanking[j].TotalProposicoes
	})

	// Aplicar posições e calcular percentual
	for i := range deputadosRanking {
		deputadosRanking[i].Posicao = i + 1
		if mediaProposicoes > 0 {
			deputadosRanking[i].PercentualMedia = ((float64(deputadosRanking[i].TotalProposicoes) - mediaProposicoes) / mediaProposicoes) * 100
		}
	}

	// Aplicar limite
	if limite > 0 && limite < len(deputadosRanking) {
		deputadosRanking = deputadosRanking[:limite]
	}

	ranking := &RankingProposicoes{
		Ano:               ano,
		TotalGeral:        totalGeral,
		MediaProposicoes:  mediaProposicoes,
		Deputados:         deputadosRanking,
		UltimaAtualizacao: time.Now(),
	}

	// Salvar no cache por 2 horas
	if data, err := json.Marshal(ranking); err == nil {
		s.cache.Set(ctx, cacheKey, string(data), 2*time.Hour)
	}

	return ranking, "computed", nil
}

// GetRankingPresenca retorna ranking de presença dos deputados
func (s *AnalyticsService) GetRankingPresenca(ctx context.Context, ano int, limite int) (*RankingPresenca, string, error) {
	cacheKey := fmt.Sprintf("ranking:presenca:%d:%d", ano, limite)

	// Tentar buscar do cache primeiro
	if cached, ok := s.cache.Get(ctx, cacheKey); ok && cached != "" {
		var ranking RankingPresenca
		if err := json.Unmarshal([]byte(cached), &ranking); err == nil {
			return &ranking, "cache", nil
		}
	}

	// Usar repositório para agregar presença (votos registrados por deputado) no ano
	deputadosCache, err := s.deputadoRepo.ListFromCache(ctx, 600)
	if err != nil {
		return nil, "", fmt.Errorf("erro ao buscar deputados: %w", err)
	}

	// Converter para slice de ponteiros para manter compatibilidade
	deputados := make([]*domain.Deputado, len(deputadosCache))
	for i := range deputadosCache {
		deputados[i] = &deputadosCache[i]
	}

	// Obter contagens de presença do repositório de votações (votos registrados por deputado)
	presencas := map[int]int{}
	totalSessoes := 0
	if s.votacaoRepo != nil {
		if rows, err := s.votacaoRepo.GetPresencaPorDeputadoAno(ctx, ano); err == nil {
			for _, p := range rows {
				presencas[p.IDDeputado] = p.Participacoes
				if p.Participacoes > totalSessoes {
					totalSessoes = p.Participacoes
				}
			}
		} else {
			s.logger.Debug("erro ao obter presencas por deputado", slog.String("error", err.Error()))
		}
	}

	deputadosRanking := make([]DeputadoRankingPresenca, 0, len(deputados))
	var somaPresenca float64

	for _, deputado := range deputados {
		sessoesPresente := presencas[deputado.ID]
		sessoesFaltou := 0
		percentualPresenca := 0.0
		if totalSessoes > 0 {
			percentualPresenca = (float64(sessoesPresente) / float64(totalSessoes)) * 100
			sessoesFaltou = totalSessoes - sessoesPresente
		}

		deputadosRanking = append(deputadosRanking, DeputadoRankingPresenca{
			ID:                 deputado.ID,
			Nome:               deputado.Nome,
			Partido:            deputado.Partido,
			UF:                 deputado.UF,
			SessoesPresente:    sessoesPresente,
			SessoesFaltou:      sessoesFaltou,
			PercentualPresenca: percentualPresenca,
		})

		somaPresenca += percentualPresenca
	}

	// Calcular média de presença
	mediaPresenca := somaPresenca / float64(len(deputadosRanking))

	// Ordenar por presença (maior para menor)
	sort.Slice(deputadosRanking, func(i, j int) bool {
		return deputadosRanking[i].PercentualPresenca > deputadosRanking[j].PercentualPresenca
	})

	// Aplicar posições
	for i := range deputadosRanking {
		deputadosRanking[i].Posicao = i + 1
	}

	// Aplicar limite
	if limite > 0 && limite < len(deputadosRanking) {
		deputadosRanking = deputadosRanking[:limite]
	}

	ranking := &RankingPresenca{
		Ano:               ano,
		TotalSessoes:      totalSessoes,
		MediaPresenca:     mediaPresenca,
		Deputados:         deputadosRanking,
		UltimaAtualizacao: time.Now(),
	}

	// Salvar no cache por 4 horas
	if data, err := json.Marshal(ranking); err == nil {
		s.cache.Set(ctx, cacheKey, string(data), 4*time.Hour)
	}

	return ranking, "computed", nil
}

// GetInsightsGerais retorna insights gerais sobre os dados
func (s *AnalyticsService) GetInsightsGerais(ctx context.Context) (*InsightsGerais, string, error) {
	cacheKey := "insights:gerais"

	// Tentar buscar do cache primeiro
	if cached, ok := s.cache.Get(ctx, cacheKey); ok && cached != "" {
		var insights InsightsGerais
		if err := json.Unmarshal([]byte(cached), &insights); err == nil {
			return &insights, "cache", nil
		}
	}

	// Criar contexto com timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Buscar dados base do banco
	deputadosCache, err := s.deputadoRepo.ListFromCache(timeoutCtx, 600)
	if err != nil {
		return nil, "", fmt.Errorf("erro ao buscar deputados: %w", err)
	}

	// Converter para slice de ponteiros para manter compatibilidade
	deputados := make([]*domain.Deputado, len(deputadosCache))
	for i := range deputadosCache {
		deputados[i] = &deputadosCache[i]
	}

	anoAtual := time.Now().Year()
	var totalGastoAno float64
	gastoPorPartido := make(map[string]float64)
	gastoPorUF := make(map[string]float64)

	// Calcular gastos totais e por categoria usando DespesaRepository
	for _, deputado := range deputados {
		var gastoDeputado float64
		if s.despesaRepo != nil {
			if stats, err := s.despesaRepo.GetDespesasStats(timeoutCtx, deputado.ID, anoAtual); err == nil && stats != nil {
				gastoDeputado = stats.TotalValor
			} else {
				s.logger.Debug("erro ao obter despesas stats para insights",
					slog.Int("deputado_id", deputado.ID),
					slog.String("error", func() string {
						if err != nil {
							return err.Error()
						}
						return ""
					}()))
			}
		}

		totalGastoAno += gastoDeputado
		gastoPorPartido[deputado.Partido] += gastoDeputado
		gastoPorUF[deputado.UF] += gastoDeputado
	}

	// Encontrar partido e UF com maior gasto
	partidoMaiorGasto := findMaxKey(gastoPorPartido)
	ufMaiorGasto := findMaxKey(gastoPorUF)

	// Buscar total de proposições (usar repositorio de proposicoes para contagem)
	totalProposicoesAno := 0
	if s.proposicaoRepo != nil {
		filtros := &domain.ProposicaoFilter{Pagina: 1, Limite: 1}
		ano := time.Now().Year()
		filtros.Ano = &ano
		// Lista apenas para obter total aproximado (repo simplificado retorna len as total)
		if _, total, err := s.proposicaoRepo.ListProposicoes(timeoutCtx, filtros); err == nil {
			totalProposicoesAno = total
		}
	}

	var mediaGastoDeputado float64
	if len(deputados) > 0 {
		mediaGastoDeputado = totalGastoAno / float64(len(deputados))
	}

	insights := &InsightsGerais{
		TotalDeputados:      len(deputados),
		TotalGastoAno:       totalGastoAno,
		TotalProposicoesAno: totalProposicoesAno,
		MediaGastosDeputado: mediaGastoDeputado,
		PartidoMaiorGasto:   partidoMaiorGasto,
		UFMaiorGasto:        ufMaiorGasto,
		UltimaAtualizacao:   time.Now(),
	}

	// Salvar no cache por 6 horas
	if data, err := json.Marshal(insights); err == nil {
		s.cache.Set(ctx, cacheKey, string(data), 6*time.Hour)
	}

	return insights, "computed", nil
}

// AtualizarRankings força atualização de todos os rankings
func (s *AnalyticsService) AtualizarRankings(ctx context.Context) error {
	anoAtual := time.Now().Year()

	// O Redis tem TTL automático, então apenas forçar recálculo
	s.logger.Info("iniciando atualização de rankings",
		slog.Int("ano", anoAtual))

	// 🔧 MELHORIA: Invalidar cache sobrescrevendo com valores vazios e TTL zero
	cacheKeys := []string{
		fmt.Sprintf("ranking:gastos:%d:50", anoAtual),
		fmt.Sprintf("ranking:proposicoes:%d:50", anoAtual),
		fmt.Sprintf("ranking:presenca:%d:50", anoAtual),
		"insights:gerais",
	}

	for _, key := range cacheKeys {
		s.cache.Set(ctx, key, "", 1*time.Nanosecond) // TTL mínimo para invalidar
	}

	s.logger.Info("cache de rankings invalidado, recalculando...")

	// Pré-computar rankings principais
	_, _, err := s.GetRankingGastos(ctx, anoAtual, 50)
	if err != nil {
		s.logger.Error("erro ao atualizar ranking de gastos", slog.String("error", err.Error()))
	}

	_, _, err = s.GetRankingProposicoes(ctx, anoAtual, 50)
	if err != nil {
		s.logger.Error("erro ao atualizar ranking de proposições", slog.String("error", err.Error()))
	}

	_, _, err = s.GetRankingPresenca(ctx, anoAtual, 50)
	if err != nil {
		s.logger.Error("erro ao atualizar ranking de presença", slog.String("error", err.Error()))
	}

	_, _, err = s.GetInsightsGerais(ctx)
	if err != nil {
		s.logger.Error("erro ao atualizar insights gerais", slog.String("error", err.Error()))
	}

	s.logger.Info("atualização de rankings concluída")
	return nil
}

// GetRankingDeputadosVotacao retorna ranking de deputados baseado em participação e votos no ano
func (s *AnalyticsService) GetRankingDeputadosVotacao(ctx context.Context, ano int, limite int) ([]domain.RankingDeputadoVotacao, string, error) {
	cacheKey := fmt.Sprintf("ranking:votacao:deputados:%d:%d", ano, limite)

	if cached, ok := s.cache.Get(ctx, cacheKey); ok && cached != "" {
		var ranking []domain.RankingDeputadoVotacao
		if err := json.Unmarshal([]byte(cached), &ranking); err == nil {
			return ranking, "cache", nil
		}
	}

	// Buscar deputados
	deputadosCache, err := s.deputadoRepo.ListFromCache(ctx, 600)
	if err != nil {
		return nil, "", fmt.Errorf("erro ao buscar deputados: %w", err)
	}

	deputados := make([]*domain.Deputado, len(deputadosCache))
	for i := range deputadosCache {
		deputados[i] = &deputadosCache[i]
	}

	// Buscar todas as votações do ano iterando páginas (limite grande)
	var allVotacoes []*domain.Votacao
	page := 1
	pageSize := 500
	for {
		vots, total, err := s.votacaoRepo.ListVotacoes(ctx, domain.FiltrosVotacao{Ano: ano}, domain.Pagination{Page: page, Limit: pageSize})
		if err != nil {
			return nil, "", fmt.Errorf("erro ao listar votações: %w", err)
		}
		if len(vots) == 0 {
			break
		}
		allVotacoes = append(allVotacoes, vots...)
		// parar se coletamos todos
		if len(allVotacoes) >= total {
			break
		}
		page++
	}

	// Map para acumular estatísticas por deputado
	stats := make(map[int]*domain.RankingDeputadoVotacao)
	for _, d := range deputados {
		stats[d.ID] = &domain.RankingDeputadoVotacao{IDDeputado: d.ID}
	}

	// Para cada votação, buscar votos e acumular
	for _, vot := range allVotacoes {
		votos, err := s.votacaoRepo.GetVotosPorVotacao(ctx, vot.ID)
		if err != nil {
			s.logger.Error("erro ao obter votos por votação", slog.Int64("votacao_id", vot.ID), slog.String("error", err.Error()))
			continue
		}
		for _, voto := range votos {
			entry, ok := stats[voto.IDDeputado]
			if !ok {
				// deputado não conhecido na lista em cache -> criar entrada
				stats[voto.IDDeputado] = &domain.RankingDeputadoVotacao{IDDeputado: voto.IDDeputado}
				entry = stats[voto.IDDeputado]
			}
			entry.TotalVotacoes++
			switch voto.Voto {
			case "Sim":
				entry.VotosFavoraveis++
			case "Não":
				entry.VotosContrarios++
			default:
				entry.Abstencoes++
			}
		}
	}

	// Converter mapa em slice e calcular taxa
	rankings := make([]domain.RankingDeputadoVotacao, 0, len(stats))
	for _, st := range stats {
		taxa := 0.0
		if st.TotalVotacoes > 0 {
			taxa = float64(st.VotosFavoraveis) / float64(st.TotalVotacoes) * 100
		}
		st.TaxaAprovacao = taxa
		rankings = append(rankings, *st)
	}

	// Ordenar por TotalVotacoes desc, depois por TaxaAprovacao desc
	sort.Slice(rankings, func(i, j int) bool {
		if rankings[i].TotalVotacoes == rankings[j].TotalVotacoes {
			return rankings[i].TaxaAprovacao > rankings[j].TaxaAprovacao
		}
		return rankings[i].TotalVotacoes > rankings[j].TotalVotacoes
	})

	if limite > 0 && limite < len(rankings) {
		rankings = rankings[:limite]
	}

	if data, err := json.Marshal(rankings); err == nil {
		s.cache.Set(ctx, cacheKey, string(data), 2*time.Hour)
	}

	return rankings, "computed", nil
}

// GetRankingPartidosDisciplina calcula disciplina partidária agregada por partido no ano
func (s *AnalyticsService) GetRankingPartidosDisciplina(ctx context.Context, ano int) ([]domain.VotacaoPartido, string, error) {
	cacheKey := fmt.Sprintf("ranking:votacao:partidos:disciplina:%d", ano)

	if cached, ok := s.cache.Get(ctx, cacheKey); ok && cached != "" {
		var resp []domain.VotacaoPartido
		if err := json.Unmarshal([]byte(cached), &resp); err == nil {
			return resp, "cache", nil
		}
	}

	// Construir mapa de partido -> stats
	deputadosCache, err := s.deputadoRepo.ListFromCache(ctx, 1000)
	if err != nil {
		return nil, "", fmt.Errorf("erro ao buscar deputados: %w", err)
	}

	// contagem de membros por partido e mapa id->partido
	membrosPorPartido := make(map[string]int)
	partidoPorDeputado := make(map[int]string)
	for _, d := range deputadosCache {
		membrosPorPartido[d.Partido]++
		partidoPorDeputado[d.ID] = d.Partido
	}

	resultadosMap := make(map[string]*domain.VotacaoPartido)
	for partido, qtd := range membrosPorPartido {
		resultadosMap[partido] = &domain.VotacaoPartido{Partido: partido, TotalMembros: qtd}
	}

	// Buscar todas as votações do ano
	var allVotacoes []*domain.Votacao
	page := 1
	pageSize := 500
	for {
		vots, total, err := s.votacaoRepo.ListVotacoes(ctx, domain.FiltrosVotacao{Ano: ano}, domain.Pagination{Page: page, Limit: pageSize})
		if err != nil {
			return nil, "", fmt.Errorf("erro ao listar votações: %w", err)
		}
		if len(vots) == 0 {
			break
		}
		allVotacoes = append(allVotacoes, vots...)
		if len(allVotacoes) >= total {
			break
		}
		page++
	}

	// Para cada votação, agregar orientações e votos por partido
	for _, vot := range allVotacoes {
		orientacoes, err := s.votacaoRepo.GetOrientacoesPorVotacao(ctx, vot.ID)
		if err != nil {
			s.logger.Error("erro ao obter orientações", slog.Int64("votacao_id", vot.ID), slog.String("error", err.Error()))
			continue
		}
		if len(orientacoes) == 0 {
			continue
		}

		votos, err := s.votacaoRepo.GetVotosPorVotacao(ctx, vot.ID)
		if err != nil {
			s.logger.Error("erro ao obter votos", slog.Int64("votacao_id", vot.ID), slog.String("error", err.Error()))
			continue
		}

		// construir mapa partido->contagens nesta votação
		contagens := make(map[string]struct{ favor, contra, abst int })
		for _, v := range votos {
			partido := partidoPorDeputado[v.IDDeputado]
			c := contagens[partido]
			switch v.Voto {
			case "Sim":
				c.favor++
			case "Não":
				c.contra++
			default:
				c.abst++
			}
			contagens[partido] = c
		}

		// aplicar orientações
		for _, o := range orientacoes {
			partido := o.Partido
			c := contagens[partido]
			rp, exists := resultadosMap[partido]
			if !exists {
				// criar entrada se partido não estava no cache
				rp = &domain.VotacaoPartido{Partido: partido, TotalMembros: membrosPorPartido[partido]}
				resultadosMap[partido] = rp
			}
			rp.VotaramFavor += c.favor
			rp.VotaramContra += c.contra
			rp.VotaramAbstencao += c.abst
			// registrar a orientacao mais recente (não perfeito, mas útil para cálculo)
			rp.Orientacao = o.Orientacao
		}
	}

	// Converter mapa em slice e calcular disciplina
	resultados := make([]domain.VotacaoPartido, 0, len(resultadosMap))
	for _, rp := range resultadosMap {
		rp.CalcularDisciplina()
		resultados = append(resultados, *rp)
	}

	// Ordenar por disciplina desc
	sort.Slice(resultados, func(i, j int) bool {
		return resultados[i].Disciplina > resultados[j].Disciplina
	})

	if data, err := json.Marshal(resultados); err == nil {
		s.cache.Set(ctx, cacheKey, string(data), 6*time.Hour)
	}

	return resultados, "computed", nil
}

// GetStatsVotacoes retorna estatísticas agregadas de votações para um período (ex: "ano", "mes")
func (s *AnalyticsService) GetStatsVotacoes(ctx context.Context, periodo string) (*domain.VotacaoStats, string, error) {
	cacheKey := fmt.Sprintf("stats:votacoes:%s", periodo)

	if cached, ok := s.cache.Get(ctx, cacheKey); ok && cached != "" {
		var stats domain.VotacaoStats
		if err := json.Unmarshal([]byte(cached), &stats); err == nil {
			return &stats, "cache", nil
		}
	}

	// Agregar estatísticas a partir do repositório de votações
	// Para 'periodo' suportamos 'ano' com valor numérico (ex: "2024") ou "ano"(current year)
	// Se periodo for um ano numérico, usamos esse ano; caso contrário usamos o ano atual
	var ano int
	if len(periodo) == 4 {
		fmt.Sscanf(periodo, "%d", &ano)
	}
	if ano == 0 {
		ano = time.Now().Year()
	}

	// Listar todas as votações do ano via paginação
	var allVotacoes []*domain.Votacao
	page := 1
	pageSize := 500
	for {
		vots, total, err := s.votacaoRepo.ListVotacoes(ctx, domain.FiltrosVotacao{Ano: ano}, domain.Pagination{Page: page, Limit: pageSize})
		if err != nil {
			return nil, "", fmt.Errorf("erro ao listar votações para stats: %w", err)
		}
		if len(vots) == 0 {
			break
		}
		allVotacoes = append(allVotacoes, vots...)
		if len(allVotacoes) >= total {
			break
		}
		page++
	}

	total := len(allVotacoes)
	aprovadas := 0
	votosTotais := 0
	porMes := make([]int, 12)
	porRelevancia := map[string]int{}

	for _, v := range allVotacoes {
		if v.Aprovacao == "Aprovada" {
			aprovadas++
		}
		// contar por mês
		mes := int(v.DataVotacao.Month()) - 1
		if mes >= 0 && mes < 12 {
			porMes[mes]++
		}
		porRelevancia[v.Relevancia]++

		// estimar participação real usando votos registrados
		votos, err := s.votacaoRepo.GetVotosPorVotacao(ctx, v.ID)
		if err != nil {
			s.logger.Debug("erro ao obter votos para stats, ignorando", slog.Int64("votacao_id", v.ID), slog.String("error", err.Error()))
			continue
		}
		votosTotais += len(votos)
	}

	rejeitadas := total - aprovadas
	mediaParticipacao := 0.0
	if total > 0 {
		mediaParticipacao = float64(votosTotais) / float64(total)
	}

	stats := &domain.VotacaoStats{
		TotalVotacoes:         total,
		VotacoesAprovadas:     aprovadas,
		VotacoesRejeitadas:    rejeitadas,
		MediaParticipacao:     mediaParticipacao,
		VotacoesPorMes:        porMes,
		VotacoesPorRelevancia: porRelevancia,
	}

	if data, err := json.Marshal(stats); err == nil {
		s.cache.Set(ctx, cacheKey, string(data), 6*time.Hour)
	}

	return stats, "computed", nil
}

// Funções auxiliares para simulação (temporárias)

// simularGastoDeputado removed - replaced by repository-backed implementation

// Funções auxiliares para simulação (temporárias)

// simularContagemProposicoes removed - replaced by repository-backed implementation

func (s *AnalyticsService) simularPresenca(deputadoID, totalSessoes int) int {
	// Simulação de presença entre 60% e 95%
	seed := deputadoID % 100
	percentual := 60 + (seed % 36) // 60% a 95%
	return (totalSessoes * percentual) / 100
}

func findMaxKey(m map[string]float64) string {
	var maxKey string
	var maxValue float64

	for key, value := range m {
		if value > maxValue {
			maxValue = value
			maxKey = key
		}
	}

	return maxKey
}

// Funções de simulação para votações (temporárias até integrar repositório real)
// Note: previously there were simulation helpers here; analytics now uses repository-backed queries.
