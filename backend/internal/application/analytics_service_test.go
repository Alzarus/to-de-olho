package application

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"to-de-olho-backend/internal/domain"
)

var testLogger = slog.New(slog.NewTextHandler(os.Stdout, nil))

// Minimal mocks implementing interfaces used by AnalyticsService
type MockCacheAnalytics struct{ data map[string]string }

func NewMockCacheAnalytics() *MockCacheAnalytics {
	return &MockCacheAnalytics{data: make(map[string]string)}
}
func (m *MockCacheAnalytics) Get(ctx context.Context, key string) (string, bool) {
	v, ok := m.data[key]
	return v, ok
}
func (m *MockCacheAnalytics) Set(ctx context.Context, key, value string, ttl time.Duration) {
	m.data[key] = value
}

type MockDeputadoRepository struct{ deputados []domain.Deputado }

func NewMockDeputadoRepository() *MockDeputadoRepository {
	return &MockDeputadoRepository{deputados: []domain.Deputado{{ID: 1, Nome: "Deputado A", Partido: "PT", UF: "SP"}, {ID: 2, Nome: "Deputado B", Partido: "PSDB", UF: "RJ"}}}
}
func (m *MockDeputadoRepository) ListFromCache(ctx context.Context, limit int) ([]domain.Deputado, error) {
	return m.deputados, nil
}
func (m *MockDeputadoRepository) UpsertDeputados(ctx context.Context, deps []domain.Deputado) error {
	m.deputados = append(m.deputados, deps...)
	return nil
}

type MockProposicaoRepository struct{ proposicoes []domain.Proposicao }

func NewMockProposicaoRepository() *MockProposicaoRepository {
	return &MockProposicaoRepository{proposicoes: []domain.Proposicao{{ID: 1, Numero: 123, SiglaTipo: "PL", Ano: 2025}}}
}
func (m *MockProposicaoRepository) ListProposicoes(ctx context.Context, filtros *domain.ProposicaoFilter) ([]domain.Proposicao, int, error) {
	return m.proposicoes, len(m.proposicoes), nil
}
func (m *MockProposicaoRepository) UpsertProposicoes(ctx context.Context, proposicoes []domain.Proposicao) error {
	m.proposicoes = append(m.proposicoes, proposicoes...)
	return nil
}
func (m *MockProposicaoRepository) GetProposicoesCountByDeputadoAno(ctx context.Context, ano int) ([]domain.ProposicaoCount, error) {
	// Return simple counts mapping based on existing mock deputados (IDs 1 and 2)
	return []domain.ProposicaoCount{{IDDeputado: 1, Count: 5}, {IDDeputado: 2, Count: 3}}, nil
}

type MockDespesaRepoAnalytics struct{}

func (m *MockDespesaRepoAnalytics) ListDespesasByDeputadoAno(ctx context.Context, deputadoID int, ano int) ([]domain.Despesa, error) {
	return []domain.Despesa{}, nil
}
func (m *MockDespesaRepoAnalytics) GetDespesasStats(ctx context.Context, deputadoID int, ano int) (*domain.DespesaStats, error) {
	return &domain.DespesaStats{TotalDespesas: 10, TotalValor: float64(10000 + deputadoID), ValorMedio: 1000, MaiorValor: 2000, TiposDiferentes: 5}, nil
}
func (m *MockDespesaRepoAnalytics) GetDespesasStatsByAno(ctx context.Context, ano int) (map[int]domain.DespesaStats, error) {
	return map[int]domain.DespesaStats{
		1: {TotalDespesas: 10, TotalValor: 11000, ValorMedio: 1100, MaiorValor: 2000, TiposDiferentes: 4},
		2: {TotalDespesas: 8, TotalValor: 9000, ValorMedio: 1125, MaiorValor: 1800, TiposDiferentes: 3},
	}, nil
}

// MockVotacaoRepository implements required domain.VotacaoRepository methods used by analytics
type MockVotacaoRepository struct{}

func (m *MockVotacaoRepository) CreateVotacao(ctx context.Context, votacao *domain.Votacao) error {
	return nil
}
func (m *MockVotacaoRepository) GetVotacaoByID(ctx context.Context, id int64) (*domain.Votacao, error) {
	return nil, domain.ErrVotacaoNaoEncontrada
}
func (m *MockVotacaoRepository) ListVotacoes(ctx context.Context, filtros domain.FiltrosVotacao, pag domain.Pagination) ([]*domain.Votacao, int, error) {
	return []*domain.Votacao{}, 0, nil
}
func (m *MockVotacaoRepository) UpdateVotacao(ctx context.Context, votacao *domain.Votacao) error {
	return nil
}
func (m *MockVotacaoRepository) DeleteVotacao(ctx context.Context, id int64) error { return nil }
func (m *MockVotacaoRepository) CreateVotoDeputado(ctx context.Context, voto *domain.VotoDeputado) error {
	return nil
}
func (m *MockVotacaoRepository) GetVotosPorVotacao(ctx context.Context, idVotacao int64) ([]*domain.VotoDeputado, error) {
	return []*domain.VotoDeputado{}, nil
}
func (m *MockVotacaoRepository) GetVotoPorDeputado(ctx context.Context, idVotacao int64, idDeputado int) (*domain.VotoDeputado, error) {
	return nil, domain.ErrVotoDeputadoNaoEncontrado
}
func (m *MockVotacaoRepository) CreateOrientacaoPartido(ctx context.Context, orientacao *domain.OrientacaoPartido) error {
	return nil
}
func (m *MockVotacaoRepository) GetOrientacoesPorVotacao(ctx context.Context, idVotacao int64) ([]*domain.OrientacaoPartido, error) {
	return []*domain.OrientacaoPartido{}, nil
}
func (m *MockVotacaoRepository) GetVotacaoDetalhada(ctx context.Context, id int64) (*domain.VotacaoDetalhada, error) {
	return nil, domain.ErrVotacaoNaoEncontrada
}
func (m *MockVotacaoRepository) UpsertVotacao(ctx context.Context, votacao *domain.Votacao) error {
	return nil
}
func (m *MockVotacaoRepository) GetPresencaPorDeputadoAno(ctx context.Context, ano int) ([]domain.PresencaCount, error) {
	// Return example participations for deputies 1 and 2
	return []domain.PresencaCount{{IDDeputado: 1, Participacoes: 50}, {IDDeputado: 2, Participacoes: 40}}, nil
}
func (m *MockVotacaoRepository) GetRankingDeputadosAggregated(ctx context.Context, ano int) ([]domain.RankingDeputadoVotacao, error) {
	return []domain.RankingDeputadoVotacao{
		{IDDeputado: 1, TotalVotacoes: 80, VotosFavoraveis: 60, VotosContrarios: 15, Abstencoes: 5},
		{IDDeputado: 2, TotalVotacoes: 70, VotosFavoraveis: 40, VotosContrarios: 20, Abstencoes: 10},
	}, nil
}
func (m *MockVotacaoRepository) GetDisciplinaPartidosAggregated(ctx context.Context, ano int) ([]domain.VotacaoPartido, error) {
	return []domain.VotacaoPartido{
		{Partido: "PT", Orientacao: "Sim", VotaramFavor: 120, VotaramContra: 10, VotaramAbstencao: 5, TotalMembros: 50},
		{Partido: "PSDB", Orientacao: "Não", VotaramFavor: 40, VotaramContra: 80, VotaramAbstencao: 5, TotalMembros: 45},
	}, nil
}
func (m *MockVotacaoRepository) GetVotacaoStatsAggregated(ctx context.Context, ano int) (*domain.VotacaoStats, error) {
	stats := &domain.VotacaoStats{
		TotalVotacoes:         12,
		VotacoesAprovadas:     7,
		VotacoesRejeitadas:    5,
		MediaParticipacao:     480.0,
		VotacoesPorMes:        make([]int, 12),
		VotacoesPorRelevancia: map[string]int{"alta": 5, "média": 4, "baixa": 3},
	}
	stats.VotacoesPorMes[0] = 2
	stats.VotacoesPorMes[1] = 3
	stats.VotacoesPorMes[2] = 7
	return stats, nil
}

func TestAnalyticsService_GetRankingGastos(t *testing.T) {
	cache := NewMockCacheAnalytics()
	deputadoRepo := NewMockDeputadoRepository()
	proposicaoRepo := NewMockProposicaoRepository()

	service := NewAnalyticsService(deputadoRepo, proposicaoRepo, &MockVotacaoRepository{}, &MockDespesaRepoAnalytics{}, cache, testLogger)

	ranking, source, err := service.GetRankingGastos(context.Background(), time.Now().Year(), 10)
	if err != nil {
		t.Fatalf("GetRankingGastos error: %v", err)
	}
	if ranking == nil {
		t.Fatal("expected ranking")
	}
	if source != "computed" && source != "cache" {
		t.Fatalf("unexpected source: %s", source)
	}
}

func TestAnalyticsService_GetInsightsGerais(t *testing.T) {
	cache := NewMockCacheAnalytics()
	deputadoRepo := NewMockDeputadoRepository()
	proposicaoRepo := NewMockProposicaoRepository()

	service := NewAnalyticsService(deputadoRepo, proposicaoRepo, &MockVotacaoRepository{}, &MockDespesaRepoAnalytics{}, cache, testLogger)

	insights, source, err := service.GetInsightsGerais(context.Background())
	if err != nil {
		t.Fatalf("GetInsightsGerais error: %v", err)
	}
	if insights == nil {
		t.Fatal("expected insights")
	}
	if source != "computed" && source != "cache" {
		t.Fatalf("unexpected source: %s", source)
	}
}

func TestAnalyticsService_GetRankingPresenca(t *testing.T) {
	cache := NewMockCacheAnalytics()
	deputadoRepo := NewMockDeputadoRepository()
	proposicaoRepo := NewMockProposicaoRepository()

	service := NewAnalyticsService(deputadoRepo, proposicaoRepo, &MockVotacaoRepository{}, &MockDespesaRepoAnalytics{}, cache, testLogger)

	ranking, source, err := service.GetRankingPresenca(context.Background(), time.Now().Year(), 5)
	if err != nil {
		t.Fatalf("GetRankingPresenca error: %v", err)
	}
	if ranking == nil {
		t.Fatal("expected ranking")
	}
	if source != "computed" && source != "cache" {
		t.Fatalf("unexpected source: %s", source)
	}
}

func TestAnalyticsService_GetRankingDeputadosVotacao(t *testing.T) {
	cache := NewMockCacheAnalytics()
	deputadoRepo := NewMockDeputadoRepository()
	proposicaoRepo := NewMockProposicaoRepository()

	service := NewAnalyticsService(deputadoRepo, proposicaoRepo, &MockVotacaoRepository{}, &MockDespesaRepoAnalytics{}, cache, testLogger)

	ctx := context.Background()
	ano := time.Now().Year()

	resp, source, err := service.GetRankingDeputadosVotacao(ctx, ano, 10)
	if err != nil {
		t.Fatalf("GetRankingDeputadosVotacao error: %v", err)
	}
	if len(resp) == 0 {
		t.Fatal("expected ranking entries")
	}
	if source != "computed" && source != "cache" {
		t.Fatalf("unexpected source: %s", source)
	}

	// segunda chamada deve vir do cache
	respCached, cacheSource, err := service.GetRankingDeputadosVotacao(ctx, ano, 10)
	if err != nil {
		t.Fatalf("cached GetRankingDeputadosVotacao error: %v", err)
	}
	if cacheSource != "cache" {
		t.Fatalf("expected cache source, got %s", cacheSource)
	}
	if len(respCached) != len(resp) {
		t.Fatalf("cached entries mismatch: got %d want %d", len(respCached), len(resp))
	}
}

func TestAnalyticsService_GetRankingPartidosDisciplina(t *testing.T) {
	cache := NewMockCacheAnalytics()
	deputadoRepo := NewMockDeputadoRepository()
	proposicaoRepo := NewMockProposicaoRepository()

	service := NewAnalyticsService(deputadoRepo, proposicaoRepo, &MockVotacaoRepository{}, &MockDespesaRepoAnalytics{}, cache, testLogger)

	ctx := context.Background()
	ano := time.Now().Year()

	resp, source, err := service.GetRankingPartidosDisciplina(ctx, ano)
	if err != nil {
		t.Fatalf("GetRankingPartidosDisciplina error: %v", err)
	}
	if len(resp) == 0 {
		t.Fatal("expected discipline ranking entries")
	}
	if source != "computed" && source != "cache" {
		t.Fatalf("unexpected source: %s", source)
	}

	// Validate discipline field calculated
	for _, partido := range resp {
		if partido.TotalMembros == 0 {
			t.Fatalf("expected TotalMembros to be populated for partido %s", partido.Partido)
		}
	}

	// confirm cache path
	_, cacheSource, err := service.GetRankingPartidosDisciplina(ctx, ano)
	if err != nil {
		t.Fatalf("cached GetRankingPartidosDisciplina error: %v", err)
	}
	if cacheSource != "cache" {
		t.Fatalf("expected cache source, got %s", cacheSource)
	}
}

func TestAnalyticsService_GetStatsVotacoes(t *testing.T) {
	cache := NewMockCacheAnalytics()
	deputadoRepo := NewMockDeputadoRepository()
	proposicaoRepo := NewMockProposicaoRepository()

	service := NewAnalyticsService(deputadoRepo, proposicaoRepo, &MockVotacaoRepository{}, &MockDespesaRepoAnalytics{}, cache, testLogger)

	ctx := context.Background()
	stats, source, err := service.GetStatsVotacoes(ctx, "2024")
	if err != nil {
		t.Fatalf("GetStatsVotacoes error: %v", err)
	}
	if stats == nil {
		t.Fatal("expected stats")
	}
	if len(stats.VotacoesPorMes) != 12 {
		t.Fatalf("expected 12 months, got %d", len(stats.VotacoesPorMes))
	}
	if source != "computed" && source != "cache" {
		t.Fatalf("unexpected source: %s", source)
	}

	// Cached path
	_, cacheSource, err := service.GetStatsVotacoes(ctx, "2024")
	if err != nil {
		t.Fatalf("cached GetStatsVotacoes error: %v", err)
	}
	if cacheSource != "cache" {
		t.Fatalf("expected cache source, got %s", cacheSource)
	}
}
