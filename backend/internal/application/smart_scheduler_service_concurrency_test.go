package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"log/slog"

	"to-de-olho-backend/internal/domain"
	"to-de-olho-backend/internal/pkg/metrics"
)

// mockSchedulerRepo implementa SchedulerRepositoryPort para testes
type mockSchedulerRepo struct {
	createCalls int
}

func (m *mockSchedulerRepo) CreateExecution(ctx context.Context, config *domain.SchedulerConfig) (*domain.SchedulerExecution, error) {
	m.createCalls++
	if m.createCalls == 1 {
		// first call succeeds
		return &domain.SchedulerExecution{ExecutionID: "exec-1", Tipo: config.Tipo, StartedAt: time.Now(), Status: domain.BackfillStatusRunning}, nil
	}
	// second call simulates lock contention
	return nil, domain.ErrSchedulerAlreadyRunning
}
func (m *mockSchedulerRepo) UpdateExecutionProgress(ctx context.Context, executionID string, update map[string]interface{}) error {
	return nil
}
func (m *mockSchedulerRepo) CompleteExecution(ctx context.Context, executionID string, status string, errorMessage *string, nextExecution *time.Time) error {
	return nil
}
func (m *mockSchedulerRepo) ShouldSchedulerExecute(ctx context.Context, schedulerTipo string, minIntervalHours int) (*domain.ShouldExecuteResult, error) {
	return &domain.ShouldExecuteResult{ShouldRun: true, Reason: "ok"}, nil
}
func (m *mockSchedulerRepo) GetCurrentStatus(ctx context.Context, schedulerTipo *string) (*domain.SchedulerStatus, error) {
	return nil, nil
}
func (m *mockSchedulerRepo) ListExecutions(ctx context.Context, limit, offset int, schedulerTipo *string) ([]domain.SchedulerExecution, int, error) {
	return nil, 0, nil
}
func (m *mockSchedulerRepo) GetLastSuccessfulExecution(ctx context.Context, schedulerTipo string) (*domain.SchedulerExecution, error) {
	return nil, nil
}
func (m *mockSchedulerRepo) CleanupOldExecutions(ctx context.Context) (int, error) { return 0, nil }

func TestExecuteIntelligentScheduler_ConcurrentSkip(t *testing.T) {
	// reset metric
	// ...existing code...
	metrics.IncSchedulerSkip("clear-me") // harmless call to ensure map is initialized

	repo := &mockSchedulerRepo{}
	s := NewSmartSchedulerService(repo, nil, nil, nil, nil, slog.Default())

	cfg := &domain.SchedulerConfig{Tipo: domain.SchedulerTipoRapido, MinIntervalHours: 0, TriggeredBy: "test"}

	// First attempt should create execution
	exec, err := s.ExecuteIntelligentScheduler(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error on first execution: %v", err)
	}
	if exec == nil {
		t.Fatalf("expected execution on first attempt")
	}

	// Second attempt (simulated) should be skipped due to ErrSchedulerAlreadyRunning
	_, err = s.ExecuteIntelligentScheduler(context.Background(), cfg)
	if err == nil {
		t.Fatalf("expected error on second attempt due to already running, got nil")
	}
	if !errors.Is(err, domain.ErrSchedulerAlreadyRunning) {
		t.Fatalf("expected ErrSchedulerAlreadyRunning, got: %v", err)
	}

	// Verify metric incremented for skip
	skips := metrics.GetSchedulerSkips(domain.SchedulerTipoRapido)
	if skips < 1 {
		t.Fatalf("expected at least 1 skip metric, got %d", skips)
	}
}
