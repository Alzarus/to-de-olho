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
	cache          CachePort
	logger         *slog.Logger
}

// NewAnalyticsService cria uma nova instância do serviço de analytics
func NewAnalyticsService(
	deputadoRepo DeputadoRepositoryInterface,
	proposicaoRepo ProposicaoRepositoryInterface,
	cache CachePort,
	logger *slog.Logger,
) *AnalyticsService {
	return &AnalyticsService{
		deputadoRepo:   deputadoRepo,
		proposicaoRepo: proposicaoRepo,
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

	// Processar em batches para melhor performance
	batchSize := 50
	var processedCount int

deputadosBatchLoop:
	for i := 0; i < len(deputados); i += batchSize {
		// Verificar timeout a cada batch
		select {
		case <-timeoutCtx.Done():
			s.logger.Warn("timeout calculando gastos",
				slog.Int("deputados_processados", processedCount),
				slog.Int("total_deputados", len(deputados)))
			break deputadosBatchLoop
		default:
		}

		end := i + batchSize
		if end > len(deputados) {
			end = len(deputados)
		}

		// Processar batch
		for j := i; j < end; j++ {
			deputado := deputados[j]
			// TODO: Implementar busca de despesas por deputado no repositório
			// Por enquanto, simular gastos para demonstração
			totalGasto := s.simularGastoDeputado(deputado.ID, ano)

			deputadosRanking = append(deputadosRanking, DeputadoRankingGastos{
				ID:         deputado.ID,
				Nome:       deputado.Nome,
				Partido:    deputado.Partido,
				UF:         deputado.UF,
				TotalGasto: totalGasto,
			})
			totalGeral += totalGasto
			processedCount++
		}
	}

	// Calcular média
	mediaGastos := totalGeral / float64(len(deputadosRanking))

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

	// Buscar proposições por autor (simulação - seria necessário implementar busca por autor na API)
	deputadosRanking := make([]DeputadoRankingProposicoes, 0, len(deputados))
	var totalGeral int

	for _, deputado := range deputados {
		// Por enquanto simular contagem de proposições
		// Na implementação real seria necessário buscar proposições por autor
		totalProposicoes := s.simularContagemProposicoes(deputado.ID, ano)

		deputadosRanking = append(deputadosRanking, DeputadoRankingProposicoes{
			ID:               deputado.ID,
			Nome:             deputado.Nome,
			Partido:          deputado.Partido,
			UF:               deputado.UF,
			TotalProposicoes: totalProposicoes,
		})

		totalGeral += totalProposicoes
	}

	// Calcular média
	mediaProposicoes := float64(totalGeral) / float64(len(deputadosRanking))

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

	// Por enquanto simular dados de presença
	// Na implementação real seria necessário integrar com API de presença da Câmara
	deputadosCache, err := s.deputadoRepo.ListFromCache(ctx, 600)
	if err != nil {
		return nil, "", fmt.Errorf("erro ao buscar deputados: %w", err)
	}

	// Converter para slice de ponteiros para manter compatibilidade
	deputados := make([]*domain.Deputado, len(deputadosCache))
	for i := range deputadosCache {
		deputados[i] = &deputadosCache[i]
	}

	totalSessoes := 100 // Simular total de sessões no ano
	deputadosRanking := make([]DeputadoRankingPresenca, 0, len(deputados))
	var somaPresenca float64

	for _, deputado := range deputados {
		// Simular dados de presença
		sessoesPresente := s.simularPresenca(deputado.ID, totalSessoes)
		sessoesFaltou := totalSessoes - sessoesPresente
		percentualPresenca := (float64(sessoesPresente) / float64(totalSessoes)) * 100

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

	// Calcular gastos totais e por categoria (amostra limitada)
	for _, deputado := range deputados {
		// Usar dados simulados para evitar timeout nas chamadas API
		gastoDeputado := s.simularGastoDeputado(deputado.ID, anoAtual)

		totalGastoAno += gastoDeputado
		gastoPorPartido[deputado.Partido] += gastoDeputado
		gastoPorUF[deputado.UF] += gastoDeputado
	}

	// Encontrar partido e UF com maior gasto
	partidoMaiorGasto := findMaxKey(gastoPorPartido)
	ufMaiorGasto := findMaxKey(gastoPorUF)

	// Buscar total de proposições (simulado)
	totalProposicoesAno := 1000 // Simular

	insights := &InsightsGerais{
		TotalDeputados:      len(deputados),
		TotalGastoAno:       totalGastoAno,
		TotalProposicoesAno: totalProposicoesAno,
		MediaGastosDeputado: totalGastoAno / float64(len(deputados)),
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

	// Limpar cache dos rankings
	keysToDelete := []string{
		fmt.Sprintf("ranking:gastos:%d:10", anoAtual),
		fmt.Sprintf("ranking:gastos:%d:50", anoAtual),
		fmt.Sprintf("ranking:proposicoes:%d:10", anoAtual),
		fmt.Sprintf("ranking:proposicoes:%d:50", anoAtual),
		fmt.Sprintf("ranking:presenca:%d:10", anoAtual),
		fmt.Sprintf("ranking:presenca:%d:50", anoAtual),
		"insights:gerais",
	}

	// O Redis tem TTL automático, então apenas forçar recálculo
	s.logger.Info("iniciando atualização de rankings",
		slog.Int("ano", anoAtual),
		slog.Int("rankings_para_atualizar", len(keysToDelete)))

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

// Funções auxiliares para simulação (temporárias)

func (s *AnalyticsService) simularGastoDeputado(deputadoID, ano int) float64 {
	// Simulação baseada em hash do ID e ano para dados consistentes
	seed := deputadoID + ano
	base := float64(seed % 200000) // Entre 0 e R$ 200.000
	return base + 50000.0          // Entre R$ 50.000 e R$ 250.000
}

// Funções auxiliares para simulação (temporárias)

func (s *AnalyticsService) simularContagemProposicoes(deputadoID, ano int) int {
	// Simulação baseada em hash do ID
	seed := deputadoID + ano
	return (seed % 50) + 1 // Entre 1 e 50 proposições
}

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
