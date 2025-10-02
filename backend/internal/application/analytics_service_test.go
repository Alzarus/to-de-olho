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
