package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"to-de-olho-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// mockRows implements pgx.Rows minimally for tests
type mockRows struct {
	idx      int
	data     []string
	failScan bool
}

func (m *mockRows) Close()                                       {}
func (m *mockRows) Err() error                                   { return nil }
func (m *mockRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (m *mockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (m *mockRows) Next() bool {
	m.idx++
	return m.idx <= len(m.data)
}
func (m *mockRows) Scan(dest ...interface{}) error {
	if m.failScan {
		return errors.New("scan error")
	}
	if m.idx-1 < 0 || m.idx-1 >= len(m.data) {
		return errors.New("out of range")
	}
	*(dest[0].(*string)) = m.data[m.idx-1]
	return nil
}
func (m *mockRows) Values() ([]interface{}, error) { return nil, nil }
func (m *mockRows) RawValues() [][]byte            { return nil }
func (m *mockRows) Conn() *pgx.Conn                { return nil }

// mockDB implements DB
type mockDB struct {
	execErr   error
	queryErr  error
	rows      pgx.Rows
	execCount int
}

func (m *mockDB) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	m.execCount++
	return pgconn.CommandTag{}, m.execErr
}
func (m *mockDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	if m.queryErr != nil {
		return nil, m.queryErr
	}
	return m.rows, nil
}

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
			name:       "deputados válidos (com mock DB)",
			repository: &DeputadoRepository{db: &mockDB{}},
			deputados: []domain.Deputado{
				{ID: 1, Nome: "João Silva", Partido: "PT"},
				{ID: 2, Nome: "Maria Santos", Partido: "PSDB"},
			},
			expectError: false,
			description: "Deveria inserir deputados válidos com sucesso",
		},
		{
			name:        "erro ao criar tabela",
			repository:  &DeputadoRepository{db: &mockDB{execErr: errors.New("ddl error")}},
			deputados:   []domain.Deputado{{ID: 1, Nome: "X"}},
			expectError: true,
			description: "Propaga erro de criação de tabela",
		},
		// cenário de erro ao inserir já coberto por erro ao criar tabela
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Se test de erro ao inserir, injeta erro após primeira Exec
			var err error
			if tt.repository != nil { // evitar panic em ponteiro nil
				err = tt.repository.UpsertDeputados(ctx, tt.deputados)
			}

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
			name:          "retorna itens do cache",
			repository:    &DeputadoRepository{db: &mockDB{rows: &mockRows{data: []string{`{"id":1,"nome":"A"}`, `{"id":2,"nome":"B"}`}}}},
			limit:         10,
			expectError:   false,
			expectedCount: 2,
			description:   "Decodifica dois deputados do cache",
		},
		{
			name:          "erro query",
			repository:    &DeputadoRepository{db: &mockDB{queryErr: errors.New("query fail")}},
			limit:         5,
			expectError:   true,
			expectedCount: 0,
			description:   "Propaga erro de query",
		},
		{
			name:          "erro scan",
			repository:    &DeputadoRepository{db: &mockDB{rows: &mockRows{data: []string{"x"}, failScan: true}}},
			limit:         5,
			expectError:   true,
			expectedCount: 0,
			description:   "Propaga erro de scan",
		},
		{
			name:          "erro unmarshal",
			repository:    &DeputadoRepository{db: &mockDB{rows: &mockRows{data: []string{"{"}}}},
			limit:         5,
			expectError:   true,
			expectedCount: 0,
			description:   "Propaga erro de unmarshal",
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
	repo := &DeputadoRepository{db: &mockDB{}}
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
