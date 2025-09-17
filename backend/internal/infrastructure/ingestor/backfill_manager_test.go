package ingestor

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// MockDB implementa interface para testes do backfill manager
type MockDBBackfill struct {
	execFunc  func(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	queryFunc func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

func (m *MockDBBackfill) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	if m.execFunc != nil {
		return m.execFunc(ctx, sql, arguments...)
	}
	return pgconn.CommandTag{}, nil
}

func (m *MockDBBackfill) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, sql, args...)
	}
	return nil, nil
}

func (m *MockDBBackfill) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	// Para testes simples, retornamos um mock row
	return &MockRowBackfill{}
}

// MockRowBackfill implementa pgx.Row para testes
type MockRowBackfill struct {
	scanFunc func(dest ...interface{}) error
}

func (m *MockRowBackfill) Scan(dest ...interface{}) error {
	if m.scanFunc != nil {
		return m.scanFunc(dest...)
	}
	return nil
}

func TestNewBackfillManager(t *testing.T) {
	mockDB := &MockDBBackfill{}
	manager := NewBackfillManagerWithDB(mockDB)

	if manager == nil {
		t.Error("NewBackfillManager should not return nil")
	}
}

func TestBackfillManager_CreateCheckpoint(t *testing.T) {
	tests := []struct {
		name           string
		checkpointType string
		metadata       map[string]interface{}
		mockExec       func(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
		wantErr        bool
	}{
		{
			name:           "success - create deputados checkpoint",
			checkpointType: "deputados",
			metadata: map[string]interface{}{
				"legislatura": "56",
				"priority":    1,
			},
			mockExec: func(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, nil
			},
			wantErr: false,
		},
		{
			name:           "success - create proposicoes checkpoint",
			checkpointType: "proposicoes",
			metadata: map[string]interface{}{
				"year":     2024,
				"priority": 2,
			},
			mockExec: func(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockDBBackfill{
				execFunc: tt.mockExec,
			}

			manager := NewBackfillManagerWithDB(mockDB)
			checkpoint, err := manager.CreateCheckpoint(context.Background(), tt.checkpointType, tt.metadata)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCheckpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if checkpoint == nil {
					t.Error("CreateCheckpoint() should return checkpoint on success")
					return
				}

				if checkpoint.Type != tt.checkpointType {
					t.Errorf("CreateCheckpoint() type = %v, want %v", checkpoint.Type, tt.checkpointType)
				}

				if checkpoint.Status != "pending" {
					t.Errorf("CreateCheckpoint() status = %v, want pending", checkpoint.Status)
				}
			}
		})
	}
}

func TestBackfillManager_MarkAsStarted(t *testing.T) {
	mockDB := &MockDBBackfill{
		execFunc: func(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, nil
		},
	}

	manager := NewBackfillManagerWithDB(mockDB)
	checkpoint := &BackfillCheckpoint{
		ID:     "test_123",
		Type:   "deputados",
		Status: "pending",
	}

	err := manager.MarkAsStarted(context.Background(), checkpoint)
	if err != nil {
		t.Errorf("MarkAsStarted() error = %v", err)
	}

	if checkpoint.Status != "in_progress" {
		t.Errorf("MarkAsStarted() status = %v, want in_progress", checkpoint.Status)
	}

	if checkpoint.StartedAt == nil {
		t.Error("MarkAsStarted() should set StartedAt")
	}
}

func TestBackfillManager_MarkAsCompleted(t *testing.T) {
	mockDB := &MockDBBackfill{
		execFunc: func(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, nil
		},
	}

	manager := NewBackfillManagerWithDB(mockDB)
	checkpoint := &BackfillCheckpoint{
		ID:     "test_123",
		Type:   "deputados",
		Status: "in_progress",
	}

	err := manager.MarkAsCompleted(context.Background(), checkpoint)
	if err != nil {
		t.Errorf("MarkAsCompleted() error = %v", err)
	}

	if checkpoint.Status != "completed" {
		t.Errorf("MarkAsCompleted() status = %v, want completed", checkpoint.Status)
	}

	if checkpoint.CompletedAt == nil {
		t.Error("MarkAsCompleted() should set CompletedAt")
	}
}

func TestBackfillManager_MarkAsFailed(t *testing.T) {
	mockDB := &MockDBBackfill{
		execFunc: func(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, nil
		},
	}

	manager := NewBackfillManagerWithDB(mockDB)
	checkpoint := &BackfillCheckpoint{
		ID:     "test_123",
		Type:   "deputados",
		Status: "in_progress",
	}

	errorMsg := "simulated error"
	err := manager.MarkAsFailed(context.Background(), checkpoint, errorMsg)
	if err != nil {
		t.Errorf("MarkAsFailed() error = %v", err)
	}

	if checkpoint.Status != "failed" {
		t.Errorf("MarkAsFailed() status = %v, want failed", checkpoint.Status)
	}

	if checkpoint.ErrorMessage != errorMsg {
		t.Errorf("MarkAsFailed() error message = %v, want %v", checkpoint.ErrorMessage, errorMsg)
	}
}

func TestBackfillManager_UpdateProgress(t *testing.T) {
	mockDB := &MockDBBackfill{
		execFunc: func(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, nil
		},
	}

	manager := NewBackfillManagerWithDB(mockDB)
	checkpoint := &BackfillCheckpoint{
		ID:     "test_123",
		Type:   "deputados",
		Status: "in_progress",
		Progress: BackfillProgress{
			TotalItems: 100,
		},
	}

	err := manager.UpdateProgress(context.Background(), checkpoint, 50, 2, "last_id_123")
	if err != nil {
		t.Errorf("UpdateProgress() error = %v", err)
	}

	if checkpoint.Progress.ProcessedItems != 50 {
		t.Errorf("UpdateProgress() processed = %v, want 50", checkpoint.Progress.ProcessedItems)
	}

	if checkpoint.Progress.FailedItems != 2 {
		t.Errorf("UpdateProgress() failed = %v, want 2", checkpoint.Progress.FailedItems)
	}

	if checkpoint.Progress.LastProcessedID != "last_id_123" {
		t.Errorf("UpdateProgress() lastID = %v, want last_id_123", checkpoint.Progress.LastProcessedID)
	}
}

func TestDefaultBackfillStrategy(t *testing.T) {
	strategy := DefaultBackfillStrategy()

	currentYear := time.Now().Year()

	if strategy.YearStart != 2019 {
		t.Errorf("DefaultBackfillStrategy() YearStart = %v, want 2019", strategy.YearStart)
	}

	if strategy.YearEnd != currentYear {
		t.Errorf("DefaultBackfillStrategy() YearEnd = %v, want %v", strategy.YearEnd, currentYear)
	}

	if strategy.BatchSize != 100 {
		t.Errorf("DefaultBackfillStrategy() BatchSize = %v, want 100", strategy.BatchSize)
	}

	if strategy.MaxRetries != 3 {
		t.Errorf("DefaultBackfillStrategy() MaxRetries = %v, want 3", strategy.MaxRetries)
	}

	if strategy.RetryDelay != 5*time.Second {
		t.Errorf("DefaultBackfillStrategy() RetryDelay = %v, want 5s", strategy.RetryDelay)
	}
}

func TestBackfillProgress_Validation(t *testing.T) {
	progress := BackfillProgress{
		TotalItems:      100,
		ProcessedItems:  50,
		FailedItems:     5,
		LastProcessedID: "id_123",
	}

	// Test that progress structure is valid
	if progress.TotalItems != 100 {
		t.Errorf("TotalItems = %v, want 100", progress.TotalItems)
	}

	if progress.ProcessedItems != 50 {
		t.Errorf("ProcessedItems = %v, want 50", progress.ProcessedItems)
	}

	if progress.FailedItems != 5 {
		t.Errorf("FailedItems = %v, want 5", progress.FailedItems)
	}

	if progress.LastProcessedID != "id_123" {
		t.Errorf("LastProcessedID = %v, want id_123", progress.LastProcessedID)
	}
}

func TestBackfillCheckpoint_Validation(t *testing.T) {
	now := time.Now()
	checkpoint := BackfillCheckpoint{
		ID:          "test_123",
		Type:        "deputados",
		Status:      "pending",
		Progress:    BackfillProgress{TotalItems: 100},
		StartedAt:   &now,
		CompletedAt: nil,
		Metadata: map[string]interface{}{
			"priority": 1,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Test that checkpoint structure is valid
	if checkpoint.ID != "test_123" {
		t.Errorf("ID = %v, want test_123", checkpoint.ID)
	}

	if checkpoint.Type != "deputados" {
		t.Errorf("Type = %v, want deputados", checkpoint.Type)
	}

	if checkpoint.Status != "pending" {
		t.Errorf("Status = %v, want pending", checkpoint.Status)
	}

	if checkpoint.StartedAt == nil {
		t.Error("StartedAt should not be nil")
	}

	if checkpoint.CompletedAt != nil {
		t.Error("CompletedAt should be nil for pending checkpoint")
	}

	if checkpoint.Metadata["priority"] != 1 {
		t.Errorf("Metadata priority = %v, want 1", checkpoint.Metadata["priority"])
	}
}
