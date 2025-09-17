package ingestor

import (
	"testing"
	"time"
)

func TestBackfillStrategy_Structure(t *testing.T) {
	strategy := BackfillStrategy{
		YearStart:  2020,
		YearEnd:    2024,
		BatchSize:  100,
		MaxRetries: 3,
		RetryDelay: 5 * time.Second,
	}

	if strategy.YearStart != 2020 {
		t.Errorf("YearStart = %v, want 2020", strategy.YearStart)
	}

	if strategy.YearEnd != 2024 {
		t.Errorf("YearEnd = %v, want 2024", strategy.YearEnd)
	}

	if strategy.BatchSize != 100 {
		t.Errorf("BatchSize = %v, want 100", strategy.BatchSize)
	}

	if strategy.MaxRetries != 3 {
		t.Errorf("MaxRetries = %v, want 3", strategy.MaxRetries)
	}

	if strategy.RetryDelay != 5*time.Second {
		t.Errorf("RetryDelay = %v, want 5s", strategy.RetryDelay)
	}
}

func TestBackfillStrategy_YearRange(t *testing.T) {
	strategy := BackfillStrategy{
		YearStart: 2019,
		YearEnd:   2025,
	}

	yearRange := strategy.YearEnd - strategy.YearStart + 1
	expectedRange := 7 // 2019, 2020, 2021, 2022, 2023, 2024, 2025

	if yearRange != expectedRange {
		t.Errorf("Year range = %v, want %v", yearRange, expectedRange)
	}
}

func TestBackfillStrategy_BatchSizeValidation(t *testing.T) {
	tests := []struct {
		name      string
		batchSize int
		isValid   bool
	}{
		{
			name:      "valid batch size",
			batchSize: 100,
			isValid:   true,
		},
		{
			name:      "small batch size",
			batchSize: 10,
			isValid:   true,
		},
		{
			name:      "large batch size",
			batchSize: 1000,
			isValid:   true,
		},
		{
			name:      "zero batch size",
			batchSize: 0,
			isValid:   false,
		},
		{
			name:      "negative batch size",
			batchSize: -1,
			isValid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := BackfillStrategy{
				BatchSize: tt.batchSize,
			}

			isValid := strategy.BatchSize > 0
			if isValid != tt.isValid {
				t.Errorf("BatchSize %v validity = %v, want %v", tt.batchSize, isValid, tt.isValid)
			}
		})
	}
}

func TestBackfillStrategy_RetryDelayValidation(t *testing.T) {
	tests := []struct {
		name       string
		retryDelay time.Duration
		isValid    bool
	}{
		{
			name:       "valid retry delay",
			retryDelay: 5 * time.Second,
			isValid:    true,
		},
		{
			name:       "short retry delay",
			retryDelay: 1 * time.Second,
			isValid:    true,
		},
		{
			name:       "long retry delay",
			retryDelay: 30 * time.Second,
			isValid:    true,
		},
		{
			name:       "zero retry delay",
			retryDelay: 0,
			isValid:    false,
		},
		{
			name:       "negative retry delay",
			retryDelay: -1 * time.Second,
			isValid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := BackfillStrategy{
				RetryDelay: tt.retryDelay,
			}

			isValid := strategy.RetryDelay > 0
			if isValid != tt.isValid {
				t.Errorf("RetryDelay %v validity = %v, want %v", tt.retryDelay, isValid, tt.isValid)
			}
		})
	}
}

func TestBackfillStrategy_MaxRetriesValidation(t *testing.T) {
	tests := []struct {
		name       string
		maxRetries int
		isValid    bool
	}{
		{
			name:       "valid max retries",
			maxRetries: 3,
			isValid:    true,
		},
		{
			name:       "single retry",
			maxRetries: 1,
			isValid:    true,
		},
		{
			name:       "many retries",
			maxRetries: 10,
			isValid:    true,
		},
		{
			name:       "zero retries",
			maxRetries: 0,
			isValid:    false,
		},
		{
			name:       "negative retries",
			maxRetries: -1,
			isValid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := BackfillStrategy{
				MaxRetries: tt.maxRetries,
			}

			isValid := strategy.MaxRetries > 0
			if isValid != tt.isValid {
				t.Errorf("MaxRetries %v validity = %v, want %v", tt.maxRetries, isValid, tt.isValid)
			}
		})
	}
}

func TestBackfillStrategy_YearValidation(t *testing.T) {
	currentYear := time.Now().Year()

	tests := []struct {
		name      string
		yearStart int
		yearEnd   int
		isValid   bool
	}{
		{
			name:      "valid year range",
			yearStart: 2020,
			yearEnd:   2024,
			isValid:   true,
		},
		{
			name:      "single year",
			yearStart: 2023,
			yearEnd:   2023,
			isValid:   true,
		},
		{
			name:      "current year range",
			yearStart: currentYear - 1,
			yearEnd:   currentYear,
			isValid:   true,
		},
		{
			name:      "invalid range - start after end",
			yearStart: 2024,
			yearEnd:   2020,
			isValid:   false,
		},
		{
			name:      "future year range",
			yearStart: currentYear + 1,
			yearEnd:   currentYear + 5,
			isValid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := BackfillStrategy{
				YearStart: tt.yearStart,
				YearEnd:   tt.yearEnd,
			}

			isValid := strategy.YearStart <= strategy.YearEnd &&
				strategy.YearEnd <= currentYear &&
				strategy.YearStart >= 1988 // Ano da constituição

			if isValid != tt.isValid {
				t.Errorf("Year range %v-%v validity = %v, want %v",
					tt.yearStart, tt.yearEnd, isValid, tt.isValid)
			}
		})
	}
}

func TestBackfillStrategy_DefaultValues(t *testing.T) {
	// Test that we can create a strategy with all default-like values
	strategy := DefaultBackfillStrategy()

	if strategy.YearStart < 2019 {
		t.Errorf("Default YearStart %v should be >= 2019", strategy.YearStart)
	}

	currentYear := time.Now().Year()
	if strategy.YearEnd > currentYear {
		t.Errorf("Default YearEnd %v should be <= current year %v", strategy.YearEnd, currentYear)
	}

	if strategy.BatchSize <= 0 {
		t.Errorf("Default BatchSize %v should be > 0", strategy.BatchSize)
	}

	if strategy.MaxRetries <= 0 {
		t.Errorf("Default MaxRetries %v should be > 0", strategy.MaxRetries)
	}

	if strategy.RetryDelay <= 0 {
		t.Errorf("Default RetryDelay %v should be > 0", strategy.RetryDelay)
	}
}
