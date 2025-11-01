package application

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"log/slog"
	"to-de-olho-backend/internal/domain"
)

type mockBackfillRepo struct {
	hasSuccessfulFn       func(ctx context.Context, startYear, endYear int) (bool, error)
	getLastExecutionFn    func(ctx context.Context, executionType string) (*domain.BackfillExecution, error)
	createExecutionFn     func(ctx context.Context, config *domain.BackfillConfig) (*domain.BackfillExecution, error)
	updateProgressFn      func(ctx context.Context, executionID string, update domain.BackfillStatus) error
	completeExecutionFn   func(ctx context.Context, executionID string, status string, errorMessage *string) error
	getRunningExecutionFn func(ctx context.Context) (*domain.BackfillExecution, error)
	listExecutionsFn      func(ctx context.Context, limit int, offset int) ([]domain.BackfillExecution, int, error)

	mu              sync.Mutex
	progress        []domain.BackfillStatus
	completedStatus string
	completedError  *string
}

func (m *mockBackfillRepo) HasSuccessfulHistoricalBackfill(ctx context.Context, startYear, endYear int) (bool, error) {
	if m.hasSuccessfulFn != nil {
		return m.hasSuccessfulFn(ctx, startYear, endYear)
	}
	return false, nil
}

func (m *mockBackfillRepo) GetLastExecution(ctx context.Context, executionType string) (*domain.BackfillExecution, error) {
	if m.getLastExecutionFn != nil {
		return m.getLastExecutionFn(ctx, executionType)
	}
	return nil, domain.ErrBackfillNaoEncontrado
}

func (m *mockBackfillRepo) CreateExecution(ctx context.Context, config *domain.BackfillConfig) (*domain.BackfillExecution, error) {
	if m.createExecutionFn != nil {
		return m.createExecutionFn(ctx, config)
	}
	return &domain.BackfillExecution{ExecutionID: "test-exec", StartedAt: time.Now()}, nil
}

func (m *mockBackfillRepo) UpdateExecutionProgress(ctx context.Context, executionID string, update domain.BackfillStatus) error {
	if m.updateProgressFn != nil {
		return m.updateProgressFn(ctx, executionID, update)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.progress = append(m.progress, update)
	return nil
}

func (m *mockBackfillRepo) CompleteExecution(ctx context.Context, executionID string, status string, errorMessage *string) error {
	if m.completeExecutionFn != nil {
		return m.completeExecutionFn(ctx, executionID, status, errorMessage)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.completedStatus = status
	m.completedError = errorMessage
	return nil
}

func (m *mockBackfillRepo) GetRunningExecution(ctx context.Context) (*domain.BackfillExecution, error) {
	if m.getRunningExecutionFn != nil {
		return m.getRunningExecutionFn(ctx)
	}
	return nil, domain.ErrBackfillNaoEncontrado
}

func (m *mockBackfillRepo) ListExecutions(ctx context.Context, limit int, offset int) ([]domain.BackfillExecution, int, error) {
	if m.listExecutionsFn != nil {
		return m.listExecutionsFn(ctx, limit, offset)
	}
	return nil, 0, nil
}

func (m *mockBackfillRepo) lastProgress() *domain.BackfillStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.progress) == 0 {
		return nil
	}
	status := m.progress[len(m.progress)-1]
	return &status
}

type mockAnalyticsSvc struct {
	mu     sync.Mutex
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
func (m *mockAnalyticsSvc) AtualizarRankings(ctx context.Context) error {
	m.mu.Lock()
	m.called = true
	m.mu.Unlock()
	return nil
}
func (m *mockAnalyticsSvc) GetRankingDeputadosVotacao(ctx context.Context, ano int, limite int) ([]domain.RankingDeputadoVotacao, string, error) {
	return nil, "", nil
}
func (m *mockAnalyticsSvc) GetRankingPartidosDisciplina(ctx context.Context, ano int) ([]domain.VotacaoPartido, string, error) {
	return nil, "", nil
}
func (m *mockAnalyticsSvc) GetStatsVotacoes(ctx context.Context, periodo string) (*domain.VotacaoStats, string, error) {
	return nil, "", nil
}

func (m *mockAnalyticsSvc) WasCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.called
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

	svc := NewSmartBackfillService(backfillRepo, nil, nil, nil, nil, analytics, logger)

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

	if !analytics.WasCalled() {
		t.Fatalf("Esperava que analytics.AtualizarRankings fosse chamado após backfill bem-sucedido")
	}
}

func TestGetBackfillConfigFromEnv_Flags(t *testing.T) {
	svc := &SmartBackfillService{}

	t.Setenv("BACKFILL_INCLUDE_DESPESAS", "false")
	t.Setenv("BACKFILL_INCLUDE_VOTACOES", "false")
	t.Setenv("BACKFILL_INCLUDE_PROPOSICOES", "false")

	cfg := svc.GetBackfillConfigFromEnv()

	if cfg.IncluirDespesas {
		t.Fatalf("esperava IncluirDespesas desativado via env")
	}

	if cfg.IncluirVotacoes {
		t.Fatalf("esperava IncluirVotacoes desativado via env")
	}

	if cfg.IncluirProposicoes {
		t.Fatalf("esperava IncluirProposicoes desativado via env")
	}
}

func TestShouldRunHistoricalBackfill_DetectsRunningExecution(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := &mockBackfillRepo{
		getRunningExecutionFn: func(ctx context.Context) (*domain.BackfillExecution, error) {
			return &domain.BackfillExecution{ExecutionID: "exec-running", StartedAt: time.Now(), Status: domain.BackfillStatusRunning}, nil
		},
	}

	svc := NewSmartBackfillService(repo, nil, nil, nil, nil, nil, logger)

	cfg := &domain.BackfillConfig{AnoInicio: 2022, AnoFim: 2023}
	cfg.SetDefaults()

	shouldRun, reason, err := svc.ShouldRunHistoricalBackfill(context.Background(), cfg)
	if err != nil {
		t.Fatalf("ShouldRunHistoricalBackfill retornou erro inesperado: %v", err)
	}
	if shouldRun {
		t.Fatal("esperava shouldRun false quando já existe execução em andamento")
	}
	if !strings.Contains(strings.ToLower(reason), "em andamento") {
		t.Fatalf("motivo deveria mencionar execução em andamento, obtido: %s", reason)
	}
}

func TestShouldRunHistoricalBackfill_PropagatesRepositoryError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := &mockBackfillRepo{
		hasSuccessfulFn: func(ctx context.Context, startYear, endYear int) (bool, error) {
			return false, errors.New("db indisponível")
		},
	}

	svc := NewSmartBackfillService(repo, nil, nil, nil, nil, nil, logger)

	cfg := &domain.BackfillConfig{AnoInicio: 2020, AnoFim: 2021}
	cfg.SetDefaults()

	_, _, err := svc.ShouldRunHistoricalBackfill(context.Background(), cfg)
	if err == nil {
		t.Fatal("esperava erro quando repositório falha")
	}
	if !strings.Contains(err.Error(), "erro ao verificar backfill histórico") {
		t.Fatalf("erro deveria conter contexto de verificação, obtido: %v", err)
	}
}

func TestRunHistoricalBackfill_SyncsDespesasAndUpdatesProgress(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := &mockBackfillRepo{}
	trackingRepo := newTrackingDespesaRepository()

	ano := time.Now().Year()
	depID := 123
	deputado := domain.Deputado{ID: depID, Nome: "Deputado Teste"}
	deputadosClient := &stubCamaraClient{
		deputados: []domain.Deputado{deputado},
		despesas: map[string][]domain.Despesa{
			fmt.Sprintf("%d:%d", depID, ano): {
				{Ano: ano, Mes: int(time.Now().Month()), TipoDespesa: "Divulgação", CodDocumento: 1, ValorDocumento: 100, ValorLiquido: 95},
			},
		},
	}

	cache := newStubCache()
	depRepo := &stubDeputadoRepo{}
	deputadosService := NewDeputadosService(deputadosClient, cache, depRepo, trackingRepo)
	analytics := &mockAnalyticsSvc{}

	svc := NewSmartBackfillService(repo, deputadosService, nil, nil, trackingRepo, analytics, logger)

	exec := &domain.BackfillExecution{ExecutionID: "exec-1", StartedAt: time.Now()}
	cfg := &domain.BackfillConfig{AnoInicio: ano, AnoFim: ano}
	cfg.SetDefaults()
	cfg.IncluirDeputados = false
	cfg.IncluirProposicoes = false
	cfg.IncluirVotacoes = false
	cfg.DelayBetweenBatches = 0

	svc.runHistoricalBackfill(context.Background(), exec, cfg)

	// permitir execução das goroutines internas
	time.Sleep(200 * time.Millisecond)

	if repo.completedStatus != domain.BackfillStatusSuccess {
		t.Fatalf("esperava status de sucesso, obtido: %s", repo.completedStatus)
	}
	progress := repo.lastProgress()
	if progress == nil {
		t.Fatal("esperava progresso registrado para despesas")
	}
	if progress.DespesasProcessadas == 0 {
		t.Fatalf("esperava contadores de despesas atualizados, progresso: %+v", progress)
	}
	if trackingRepo.upsertCount == 0 {
		t.Fatal("esperava UpsertDespesas chamado ao menos uma vez")
	}
	if !analytics.WasCalled() {
		t.Fatal("esperava analytics.AtualizarRankings chamado após sincronização")
	}
}
