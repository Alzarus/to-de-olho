package application

import (
	"context"
	"io"
	"testing"
	"time"

	"log/slog"
	"to-de-olho-backend/internal/domain"
)

type mockBackfillRepo struct{}

func (m *mockBackfillRepo) HasSuccessfulHistoricalBackfill(ctx context.Context, startYear, endYear int) (bool, error) {
	return false, nil
}
func (m *mockBackfillRepo) GetLastExecution(ctx context.Context, executionType string) (*domain.BackfillExecution, error) {
	return nil, domain.ErrBackfillNaoEncontrado
}
func (m *mockBackfillRepo) CreateExecution(ctx context.Context, config *domain.BackfillConfig) (*domain.BackfillExecution, error) {
	return &domain.BackfillExecution{ExecutionID: "test-exec", StartedAt: time.Now()}, nil
}
func (m *mockBackfillRepo) UpdateExecutionProgress(ctx context.Context, executionID string, update domain.BackfillStatus) error {
	return nil
}
func (m *mockBackfillRepo) CompleteExecution(ctx context.Context, executionID string, status string, errorMessage *string) error {
	return nil
}
func (m *mockBackfillRepo) GetRunningExecution(ctx context.Context) (*domain.BackfillExecution, error) {
	return nil, domain.ErrBackfillNaoEncontrado
}
func (m *mockBackfillRepo) ListExecutions(ctx context.Context, limit int, offset int) ([]domain.BackfillExecution, int, error) {
	return nil, 0, nil
}

type mockAnalyticsSvc struct {
	called bool
}

func (m *mockAnalyticsSvc) GetRankingGastos(ctx context.Context, ano int, limite int) (*RankingGastos, string, error) {
	return nil, "", nil
}
func (m *mockAnalyticsSvc) GetRankingProposicoes(ctx context.Context, ano int, limite int) (*RankingProposicoes, string, error) {
	return nil, "", nil
}
func (m *mockAnalyticsSvc) GetRankingPresenca(ctx context.Context, ano int, limite int) (*RankingPresenca, string, error) {
	return nil, "", nil
}
func (m *mockAnalyticsSvc) GetInsightsGerais(ctx context.Context) (*InsightsGerais, string, error) {
	return nil, "", nil
}
func (m *mockAnalyticsSvc) AtualizarRankings(ctx context.Context) error { m.called = true; return nil }
func (m *mockAnalyticsSvc) GetRankingDeputadosVotacao(ctx context.Context, ano int, limite int) ([]domain.RankingDeputadoVotacao, string, error) {
	return nil, "", nil
}
func (m *mockAnalyticsSvc) GetRankingPartidosDisciplina(ctx context.Context, ano int) ([]domain.VotacaoPartido, string, error) {
	return nil, "", nil
}
func (m *mockAnalyticsSvc) GetStatsVotacoes(ctx context.Context, periodo string) (*domain.VotacaoStats, string, error) {
	return nil, "", nil
}

// Mocks mínimos para serviços usados pelo SmartBackfillService
var (
	_ BackfillRepositoryPort    = (*mockBackfillRepo)(nil)
	_ AnalyticsServiceInterface = (*mockAnalyticsSvc)(nil)
)

func TestSmartBackfillService_TriggersAnalyticsOnSuccess(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	backfillRepo := &mockBackfillRepo{}
	analytics := &mockAnalyticsSvc{}

	svc := NewSmartBackfillService(backfillRepo, nil, nil, nil, analytics, logger)

	// Criar execução e config mínima
	exec := &domain.BackfillExecution{ExecutionID: "exec-1", StartedAt: time.Now()}
	cfg := &domain.BackfillConfig{}
	cfg.SetDefaults()
	// Nos testes, evitar chamadas externas/repositórios: desabilitar sincronizações
	cfg.IncluirDeputados = false
	cfg.IncluirProposicoes = false
	cfg.IncluirVotacoes = false

	// Executar diretamente o runHistoricalBackfill (sincronamente)
	svc.runHistoricalBackfill(context.Background(), exec, cfg)

	// Dar um tempo para a goroutine de AtualizarRankings executar caso tenha sido disparada
	time.Sleep(200 * time.Millisecond)

	if !analytics.called {
		t.Fatalf("Esperava que analytics.AtualizarRankings fosse chamado após backfill bem-sucedido")
	}
}
