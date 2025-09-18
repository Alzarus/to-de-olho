package ingestor

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"to-de-olho-backend/internal/application"
)

// TestSyncMetrics testa a struct SyncMetrics usando table-driven tests
func TestSyncMetrics(t *testing.T) {
	tests := []struct {
		name     string
		metrics  SyncMetrics
		validate func(t *testing.T, metrics SyncMetrics)
	}{
		{
			name: "complete structure with all fields",
			metrics: SyncMetrics{
				StartTime:          time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
				EndTime:            time.Date(2024, 1, 1, 11, 30, 0, 0, time.UTC),
				Duration:           90 * time.Minute,
				DeputadosUpdated:   100,
				ProposicoesUpdated: 250,
				ErrorsCount:        2,
				Errors:             []string{"error1", "error2"},
				SyncType:           "daily",
			},
			validate: func(t *testing.T, metrics SyncMetrics) {
				if metrics.DeputadosUpdated != 100 {
					t.Errorf("DeputadosUpdated = %v, want 100", metrics.DeputadosUpdated)
				}
				if metrics.ProposicoesUpdated != 250 {
					t.Errorf("ProposicoesUpdated = %v, want 250", metrics.ProposicoesUpdated)
				}
				if metrics.ErrorsCount != 2 {
					t.Errorf("ErrorsCount = %v, want 2", metrics.ErrorsCount)
				}
				if len(metrics.Errors) != 2 {
					t.Errorf("len(Errors) = %v, want 2", len(metrics.Errors))
				}
				if metrics.SyncType != "daily" {
					t.Errorf("SyncType = %v, want daily", metrics.SyncType)
				}
				expectedDuration := 90 * time.Minute
				if metrics.Duration != expectedDuration {
					t.Errorf("Duration = %v, want %v", metrics.Duration, expectedDuration)
				}
			},
		},
		{
			name: "empty errors case",
			metrics: SyncMetrics{
				ErrorsCount: 0,
				Errors:      []string{},
				SyncType:    "quick",
			},
			validate: func(t *testing.T, metrics SyncMetrics) {
				if metrics.ErrorsCount != 0 {
					t.Errorf("ErrorsCount = %v, want 0", metrics.ErrorsCount)
				}
				if len(metrics.Errors) != 0 {
					t.Errorf("len(Errors) = %v, want 0", len(metrics.Errors))
				}
				if metrics.SyncType != "quick" {
					t.Errorf("SyncType = %v, want quick", metrics.SyncType)
				}
			},
		},
		{
			name: "nil errors slice",
			metrics: SyncMetrics{
				ErrorsCount: 0,
				Errors:      nil,
				SyncType:    "incremental",
			},
			validate: func(t *testing.T, metrics SyncMetrics) {
				if metrics.ErrorsCount != 0 {
					t.Errorf("ErrorsCount = %v, want 0", metrics.ErrorsCount)
				}
				if metrics.Errors != nil {
					t.Errorf("Errors should be nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.metrics)
		})
	}
}

// TestSyncMetrics_JSONSerialization testa serialização/deserialização JSON
func TestSyncMetrics_JSONSerialization(t *testing.T) {
	originalMetrics := SyncMetrics{
		StartTime:          time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		EndTime:            time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
		Duration:           time.Hour,
		DeputadosUpdated:   50,
		ProposicoesUpdated: 150,
		ErrorsCount:        1,
		Errors:             []string{"connection timeout"},
		SyncType:           "test",
	}

	// Test marshaling
	data, err := json.Marshal(originalMetrics)
	if err != nil {
		t.Fatalf("Failed to marshal SyncMetrics: %v", err)
	}

	if len(data) == 0 {
		t.Error("Marshaled data is empty")
	}

	// Test unmarshaling
	var unmarshaled SyncMetrics
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal SyncMetrics: %v", err)
	}

	// Verify key fields
	if unmarshaled.SyncType != originalMetrics.SyncType {
		t.Errorf("Unmarshaled SyncType = %v, want %v", unmarshaled.SyncType, originalMetrics.SyncType)
	}

	if unmarshaled.DeputadosUpdated != originalMetrics.DeputadosUpdated {
		t.Errorf("Unmarshaled DeputadosUpdated = %v, want %v",
			unmarshaled.DeputadosUpdated, originalMetrics.DeputadosUpdated)
	}

	if unmarshaled.ProposicoesUpdated != originalMetrics.ProposicoesUpdated {
		t.Errorf("Unmarshaled ProposicoesUpdated = %v, want %v",
			unmarshaled.ProposicoesUpdated, originalMetrics.ProposicoesUpdated)
	}

	if unmarshaled.ErrorsCount != originalMetrics.ErrorsCount {
		t.Errorf("Unmarshaled ErrorsCount = %v, want %v",
			unmarshaled.ErrorsCount, originalMetrics.ErrorsCount)
	}
}

// TestSyncMetrics_SyncTypes testa diferentes tipos de sincronização
func TestSyncMetrics_SyncTypes(t *testing.T) {
	syncTypes := []string{"daily", "quick", "incremental", "full", "batch"}

	for _, syncType := range syncTypes {
		t.Run("sync_type_"+syncType, func(t *testing.T) {
			metrics := SyncMetrics{
				SyncType: syncType,
			}

			if metrics.SyncType != syncType {
				t.Errorf("SyncType = %v, want %v", metrics.SyncType, syncType)
			}
		})
	}
}

// TestNewIncrementalSyncManager testa a criação do manager
func TestNewIncrementalSyncManager(t *testing.T) {
	tests := []struct {
		name    string
		wantNil bool
	}{
		{
			name:    "should create manager with nil dependencies",
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewIncrementalSyncManager(
				nil, // deputadosService
				nil, // proposicoesService
				nil, // analyticsService
				nil, // db
				nil, // cache
			)

			if (got == nil) != tt.wantNil {
				t.Errorf("NewIncrementalSyncManager() = %v, wantNil %v", got, tt.wantNil)
			}
		})
	}
}

// TestIncrementalSyncManager_CleanupOldCache testa a limpeza de cache
func TestIncrementalSyncManager_CleanupOldCache(t *testing.T) {
	tests := []struct {
		name    string
		cache   application.CachePort
		wantErr bool
	}{
		{
			name:    "nil cache should not error",
			cache:   nil,
			wantErr: false,
		},
		{
			name:    "valid cache should work",
			cache:   &mockCache{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &IncrementalSyncManager{
				cache: tt.cache,
			}

			ctx := context.Background()
			err := manager.cleanupOldCache(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("cleanupOldCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestIncrementalSyncManager_GetSyncStats testa o método GetSyncStats
func TestIncrementalSyncManager_GetSyncStats(t *testing.T) {
	tests := []struct {
		name        string
		manager     *IncrementalSyncManager
		days        int
		expectError bool
		expectNil   bool
	}{
		{
			name: "nil database should return error",
			manager: &IncrementalSyncManager{
				db: nil,
			},
			days:        7,
			expectError: true,
			expectNil:   true,
		},
		{
			name: "negative days should be handled",
			manager: &IncrementalSyncManager{
				db: nil,
			},
			days:        -1,
			expectError: true,
			expectNil:   true,
		},
		{
			name: "zero days should be handled",
			manager: &IncrementalSyncManager{
				db: nil,
			},
			days:        0,
			expectError: true,
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			stats, err := tt.manager.GetSyncStats(ctx, tt.days)

			if (err != nil) != tt.expectError {
				t.Errorf("GetSyncStats() error = %v, expectError %v", err, tt.expectError)
			}

			if (stats == nil) != tt.expectNil {
				t.Errorf("GetSyncStats() stats = %v, expectNil %v", stats, tt.expectNil)
			}
		})
	}
}

// TestSyncMetrics_DurationCalculation testa cálculos de duração
func TestSyncMetrics_DurationCalculation(t *testing.T) {
	testCases := []struct {
		name      string
		startTime time.Time
		endTime   time.Time
		expected  time.Duration
	}{
		{
			name:      "30 seconds duration",
			startTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			endTime:   time.Date(2024, 1, 1, 10, 0, 30, 0, time.UTC),
			expected:  30 * time.Second,
		},
		{
			name:      "2 hours 30 minutes duration",
			startTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			endTime:   time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC),
			expected:  2*time.Hour + 30*time.Minute,
		},
		{
			name:      "zero duration",
			startTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			endTime:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			expected:  0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metrics := SyncMetrics{
				StartTime: tc.startTime,
				EndTime:   tc.endTime,
				Duration:  tc.endTime.Sub(tc.startTime),
			}

			if metrics.Duration != tc.expected {
				t.Errorf("Duration = %v, want %v", metrics.Duration, tc.expected)
			}
		})
	}
}

// TestSyncMetrics_ErrorAccumulation testa acumulação de erros
func TestSyncMetrics_ErrorAccumulation(t *testing.T) {
	metrics := SyncMetrics{
		Errors:      []string{},
		ErrorsCount: 0,
	}

	// Simular adição de erros
	errors := []string{"error 1", "error 2", "error 3"}

	for _, errMsg := range errors {
		metrics.Errors = append(metrics.Errors, errMsg)
		metrics.ErrorsCount++
	}

	if metrics.ErrorsCount != 3 {
		t.Errorf("ErrorsCount = %v, want 3", metrics.ErrorsCount)
	}

	if len(metrics.Errors) != 3 {
		t.Errorf("len(Errors) = %v, want 3", len(metrics.Errors))
	}

	// Verificar conteúdo dos erros
	for i, errMsg := range errors {
		if metrics.Errors[i] != errMsg {
			t.Errorf("Errors[%d] = %v, want %v", i, metrics.Errors[i], errMsg)
		}
	}
}

// mockCache é um mock simples do CachePort para testes
type mockCache struct{}

func (m *mockCache) Get(ctx context.Context, key string) (string, bool) {
	return "", false
}

func (m *mockCache) Set(ctx context.Context, key, value string, ttl time.Duration) {
	// Mock implementation - does nothing
}
