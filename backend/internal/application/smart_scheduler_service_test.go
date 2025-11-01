package application

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"to-de-olho-backend/internal/domain"
)

// schedulerRepoStub oferece hooks configuráveis para os testes do scheduler
// e captura chamadas relevantes para asserções.
type schedulerRepoStub struct {
	mu                  sync.Mutex
	shouldExecuteResult *domain.ShouldExecuteResult
	shouldExecuteErr    error
	createFn            func(ctx context.Context, config *domain.SchedulerConfig) (*domain.SchedulerExecution, error)
	updateCalls         []map[string]interface{}
	completeStatus      string
	completeError       *string
	completeNext        *time.Time
	createCalled        bool
}

var _ SchedulerRepositoryPort = (*schedulerRepoStub)(nil)

func (s *schedulerRepoStub) CreateExecution(ctx context.Context, config *domain.SchedulerConfig) (*domain.SchedulerExecution, error) {
	s.mu.Lock()
	s.createCalled = true
	fn := s.createFn
	s.mu.Unlock()

	if fn != nil {
		return fn(ctx, config)
	}

	return &domain.SchedulerExecution{
		ExecutionID: "exec-stub",
		Tipo:        config.Tipo,
		StartedAt:   time.Now(),
		Status:      domain.BackfillStatusRunning,
	}, nil
}

func (s *schedulerRepoStub) UpdateExecutionProgress(ctx context.Context, executionID string, update map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updateCalls = append(s.updateCalls, update)
	return nil
}

func (s *schedulerRepoStub) CompleteExecution(ctx context.Context, executionID string, status string, errorMessage *string, nextExecution *time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.completeStatus = status
	s.completeError = errorMessage
	s.completeNext = nextExecution
	return nil
}

func (s *schedulerRepoStub) ShouldSchedulerExecute(ctx context.Context, schedulerTipo string, minIntervalHours int) (*domain.ShouldExecuteResult, error) {
	if s.shouldExecuteErr != nil {
		return nil, s.shouldExecuteErr
	}
	if s.shouldExecuteResult != nil {
		return s.shouldExecuteResult, nil
	}
	return &domain.ShouldExecuteResult{ShouldRun: true, Reason: "default"}, nil
}

func (s *schedulerRepoStub) GetCurrentStatus(ctx context.Context, schedulerTipo *string) (*domain.SchedulerStatus, error) {
	return nil, nil
}

func (s *schedulerRepoStub) ListExecutions(ctx context.Context, limit, offset int, schedulerTipo *string) ([]domain.SchedulerExecution, int, error) {
	return nil, 0, nil
}

func (s *schedulerRepoStub) GetLastSuccessfulExecution(ctx context.Context, schedulerTipo string) (*domain.SchedulerExecution, error) {
	return nil, nil
}

func (s *schedulerRepoStub) CleanupOldExecutions(ctx context.Context) (int, error) {
	return 0, nil
}

func TestShouldRunScheduler_UsesRepositoryDecision(t *testing.T) {
	repo := &schedulerRepoStub{
		shouldExecuteResult: &domain.ShouldExecuteResult{ShouldRun: false, Reason: "executado recentemente"},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := NewSmartSchedulerService(repo, nil, nil, nil, nil, logger)

	cfg := &domain.SchedulerConfig{Tipo: domain.SchedulerTipoRapido, MinIntervalHours: 2}

	shouldRun, reason, err := svc.ShouldRunScheduler(context.Background(), cfg.Tipo, cfg)
	if err != nil {
		t.Fatalf("ShouldRunScheduler retornou erro inesperado: %v", err)
	}
	if shouldRun {
		t.Fatalf("esperava shouldRun false, mas recebeu true")
	}
	if reason != "executado recentemente" {
		t.Fatalf("motivo incorreto, obtido: %s", reason)
	}
}

func TestExecuteIntelligentScheduler_SkipsWhenNotDue(t *testing.T) {
	repo := &schedulerRepoStub{
		shouldExecuteResult: &domain.ShouldExecuteResult{ShouldRun: false, Reason: "ainda dentro do intervalo"},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := NewSmartSchedulerService(repo, nil, nil, nil, nil, logger)

	cfg := &domain.SchedulerConfig{Tipo: domain.SchedulerTipoRapido, MinIntervalHours: 3, TriggeredBy: "test"}

	exec, err := svc.ExecuteIntelligentScheduler(context.Background(), cfg)
	if err == nil {
		t.Fatal("esperava erro indicando que o scheduler não deve executar")
	}
	if exec != nil {
		t.Fatal("execução não deveria ser criada quando scheduler é pulado")
	}
	repo.mu.Lock()
	created := repo.createCalled
	repo.mu.Unlock()
	if created {
		t.Fatal("CreateExecution não deve ser chamado quando ShouldRun retorna false")
	}
}

func TestRunSchedulerExecution_CompletesSuccessfully(t *testing.T) {
	repo := &schedulerRepoStub{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	deputadosClient := &stubCamaraClient{
		deputados: []domain.Deputado{{ID: 101, Nome: "Deputada Teste"}},
	}
	cache := newStubCache()
	depRepo := &stubDeputadoRepo{}
	despesaRepo := newTrackingDespesaRepository()
	deputadosService := NewDeputadosService(deputadosClient, cache, depRepo, despesaRepo)

	svc := NewSmartSchedulerService(repo, nil, deputadosService, nil, nil, logger)

	execution := &domain.SchedulerExecution{ExecutionID: "exec-123", Tipo: domain.SchedulerTipoRapido, StartedAt: time.Now(), Status: domain.BackfillStatusRunning}
	cfg := &domain.SchedulerConfig{Tipo: domain.SchedulerTipoRapido, IncluirDeputados: true}

	svc.runSchedulerExecution(context.Background(), execution, cfg)

	repo.mu.Lock()
	defer repo.mu.Unlock()
	if repo.completeStatus != domain.BackfillStatusSuccess {
		t.Fatalf("status final inesperado: %s", repo.completeStatus)
	}
	if repo.completeNext == nil {
		t.Fatal("esperava próxima execução calculada")
	}
	if len(repo.updateCalls) == 0 {
		t.Fatal("esperava ao menos um registro de progresso")
	}
	found := false
	for _, call := range repo.updateCalls {
		if v, ok := call["deputados_sincronizados"]; ok {
			if count, ok := v.(int); ok && count == 1 {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatalf("não encontrou atualização de deputados_sincronizados nas chamadas: %#v", repo.updateCalls)
	}
}

func TestRunSchedulerExecution_RespectsFeatureFlags(t *testing.T) {
	t.Setenv("SCHEDULER_INCLUDE_DESPESAS", "false")
	t.Setenv("SCHEDULER_INCLUDE_VOTACOES", "false")
	t.Setenv("SCHEDULER_INCLUDE_PROPOSICOES", "false")

	repo := &schedulerRepoStub{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// preparar serviço de deputados mínimo
	deputadosClient := &stubCamaraClient{
		deputados: []domain.Deputado{{ID: 101, Nome: "Deputada Teste"}},
	}
	cache := newStubCache()
	depRepo := &stubDeputadoRepo{}
	deputadosService := NewDeputadosService(deputadosClient, cache, depRepo, newTrackingDespesaRepository())

	svc := NewSmartSchedulerService(repo, nil, deputadosService, nil, nil, logger)

	execution := &domain.SchedulerExecution{ExecutionID: "exec-flag", Tipo: domain.SchedulerTipoRapido, StartedAt: time.Now(), Status: domain.BackfillStatusRunning}
	cfg := &domain.SchedulerConfig{
		Tipo:               domain.SchedulerTipoRapido,
		IncluirDeputados:   true,
		IncluirDespesas:    true,
		IncluirVotacoes:    false,
		IncluirProposicoes: false,
	}

	svc.runSchedulerExecution(context.Background(), execution, cfg)

	repo.mu.Lock()
	defer repo.mu.Unlock()
	for _, update := range repo.updateCalls {
		if _, ok := update["despesas_sincronizadas"]; ok {
			t.Fatalf("despesas_sincronizadas não deveria ser registrado quando flag está desativada: %#v", update)
		}
	}
}
