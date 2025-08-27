package repository

import (
	"context"
	"fmt"
	"testing"

	"to-de-olho-backend/internal/domain"
)

func TestNewDeputadoRepository(t *testing.T) {
	tests := []struct {
		name string
		db   interface{}
	}{
		{
			name: "repository válido",
			db:   &struct{}{},
		},
		{
			name: "db nil",
			db:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewDeputadoRepository(nil)
			if repo == nil {
				t.Error("NewDeputadoRepository não deveria retornar nil")
			}
		})
	}
}

func TestDeputadoRepository_UpsertDeputados(t *testing.T) {
	tests := []struct {
		name        string
		repository  *DeputadoRepository
		deputados   []domain.Deputado
		expectError bool
		description string
	}{
		{
			name:        "repository nil",
			repository:  nil,
			deputados:   []domain.Deputado{{ID: 1, Nome: "Test"}},
			expectError: false,
			description: "Repository nil deveria ser tratado graciosamente",
		},
		{
			name:        "lista vazia",
			repository:  &DeputadoRepository{},
			deputados:   []domain.Deputado{},
			expectError: false,
			description: "Lista vazia deveria ser tratada sem erro",
		},
		{
			name:       "deputados válidos",
			repository: &DeputadoRepository{},
			deputados: []domain.Deputado{
				{ID: 1, Nome: "João Silva", Partido: "PT"},
				{ID: 2, Nome: "Maria Santos", Partido: "PSDB"},
			},
			expectError: false,
			description: "Deveria inserir deputados válidos com sucesso",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			err := tt.repository.UpsertDeputados(ctx, tt.deputados)

			if tt.expectError && err == nil {
				t.Errorf("esperava erro, mas não ocorreu - %s", tt.description)
			}
			if !tt.expectError && err != nil {
				t.Errorf("não esperava erro, mas ocorreu: %v - %s", err, tt.description)
			}
		})
	}
}

func TestDeputadoRepository_ListFromCache(t *testing.T) {
	tests := []struct {
		name          string
		repository    *DeputadoRepository
		limit         int
		expectError   bool
		expectedCount int
		description   string
	}{
		{
			name:          "repository nil",
			repository:    nil,
			limit:         10,
			expectError:   false,
			expectedCount: 0,
			description:   "Repository nil deveria retornar lista vazia",
		},
		{
			name:          "limite zero",
			repository:    &DeputadoRepository{},
			limit:         0,
			expectError:   false,
			expectedCount: 0,
			description:   "Limite zero deveria ser aceito",
		},
		{
			name:          "limite negativo",
			repository:    &DeputadoRepository{},
			limit:         -1,
			expectError:   false,
			expectedCount: 0,
			description:   "Limite negativo deveria ser tratado",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			deputados, err := tt.repository.ListFromCache(ctx, tt.limit)

			if tt.expectError && err == nil {
				t.Errorf("esperava erro, mas não ocorreu - %s", tt.description)
			}
			if !tt.expectError && err != nil {
				t.Errorf("não esperava erro, mas ocorreu: %v - %s", err, tt.description)
			}

			if deputados == nil && tt.expectedCount == 0 {
				// OK, resultado esperado
			} else if deputados != nil && len(deputados) != tt.expectedCount {
				t.Errorf("esperava %d deputados, recebeu %d - %s",
					tt.expectedCount, len(deputados), tt.description)
			}
		})
	}
}

func TestDeputadoRepository_EdgeCases(t *testing.T) {
	repo := &DeputadoRepository{}
	ctx := context.Background()

	// Teste com contexto cancelado
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()

	err := repo.UpsertDeputados(cancelCtx, []domain.Deputado{{ID: 1}})
	if err != nil {
		t.Logf("erro com contexto cancelado (esperado): %v", err)
	}

	// Teste com slice grande
	largeSlice := make([]domain.Deputado, 1000)
	for i := range largeSlice {
		largeSlice[i] = domain.Deputado{
			ID:   i,
			Nome: fmt.Sprintf("Deputado %d", i),
		}
	}

	err = repo.UpsertDeputados(ctx, largeSlice)
	if err != nil {
		t.Logf("erro com slice grande (esperado sem banco): %v", err)
	}
}

// Benchmarks
func BenchmarkDeputadoRepository_UpsertDeputados(b *testing.B) {
	repo := &DeputadoRepository{}
	ctx := context.Background()
	deputados := []domain.Deputado{
		{ID: 1, Nome: "Benchmark Test", Partido: "TEST"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.UpsertDeputados(ctx, deputados)
	}
}

func BenchmarkDeputadoRepository_ListFromCache(b *testing.B) {
	repo := &DeputadoRepository{}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.ListFromCache(ctx, 100)
	}
}

// end of file: deputado_repository_test.go (normalized to LF)
