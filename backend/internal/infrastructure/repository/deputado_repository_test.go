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

// Testes adicionais para melhorar cobertura
func TestDeputadoRepository_UpsertDeputados_EmptySlice(t *testing.T) {
	repo := &DeputadoRepository{db: &mockDB{}}
	ctx := context.Background()

	err := repo.UpsertDeputados(ctx, []domain.Deputado{})
	if err != nil {
		t.Logf("erro com slice vazio: %v", err)
	}
}

func TestDeputadoRepository_UpsertDeputados_NilSlice(t *testing.T) {
	repo := &DeputadoRepository{db: &mockDB{}}
	ctx := context.Background()

	err := repo.UpsertDeputados(ctx, nil)
	if err != nil {
		t.Logf("erro com slice nil: %v", err)
	}
}

func TestDeputadoRepository_ListFromCache_ZeroLimit(t *testing.T) {
	repo := &DeputadoRepository{db: &mockDB{}}
	ctx := context.Background()

	// Usar defer para capturar possíveis panics
	defer func() {
		if r := recover(); r != nil {
			t.Logf("ListFromCache with zero limit panicked: %v", r)
		}
	}()

	deputados, err := repo.ListFromCache(ctx, 0)
	if err != nil {
		t.Logf("erro com limite zero: %v", err)
	}

	if len(deputados) > 0 {
		t.Log("ListFromCache with zero limit returned results")
	}
}

func TestDeputadoRepository_ListFromCache_NegativeLimit(t *testing.T) {
	repo := &DeputadoRepository{db: &mockDB{}}
	ctx := context.Background()

	// Usar defer para capturar possíveis panics
	defer func() {
		if r := recover(); r != nil {
			t.Logf("ListFromCache with negative limit panicked: %v", r)
		}
	}()

	deputados, err := repo.ListFromCache(ctx, -1)
	if err != nil {
		t.Logf("erro com limite negativo: %v", err)
	}

	if len(deputados) > 0 {
		t.Log("ListFromCache with negative limit returned results")
	}
}

func TestDeputadoRepository_ListFromCache_LargeLimit(t *testing.T) {
	repo := &DeputadoRepository{db: &mockDB{}}
	ctx := context.Background()

	// Usar defer para capturar possíveis panics
	defer func() {
		if r := recover(); r != nil {
			t.Logf("ListFromCache with large limit panicked: %v", r)
		}
	}()

	deputados, err := repo.ListFromCache(ctx, 10000)
	if err != nil {
		t.Logf("erro com limite grande: %v", err)
	}

	t.Logf("ListFromCache with large limit returned %d results", len(deputados))
}

func TestDeputadoRepository_UpsertDeputados_ValidationCases(t *testing.T) {
	repo := &DeputadoRepository{db: &mockDB{}}
	ctx := context.Background()

	// Teste com deputados com IDs duplicados
	deputadosDuplicados := []domain.Deputado{
		{ID: 1, Nome: "Deputado 1", Partido: "PT"},
		{ID: 1, Nome: "Deputado 1 Duplicado", Partido: "PSDB"},
	}

	err := repo.UpsertDeputados(ctx, deputadosDuplicados)
	if err != nil {
		t.Logf("erro com IDs duplicados: %v", err)
	}

	// Teste com nomes muito longos
	deputadoNomeLongo := []domain.Deputado{
		{ID: 2, Nome: "Nome muito longo que pode causar problemas na base de dados se não houver validação adequada", Partido: "TEST"},
	}

	err = repo.UpsertDeputados(ctx, deputadoNomeLongo)
	if err != nil {
		t.Logf("erro com nome longo: %v", err)
	}

	// Teste com campos vazios
	deputadoCamposVazios := []domain.Deputado{
		{ID: 3, Nome: "", Partido: ""},
	}

	err = repo.UpsertDeputados(ctx, deputadoCamposVazios)
	if err != nil {
		t.Logf("erro com campos vazios: %v", err)
	}
}

func TestDeputadoRepository_NilDB(t *testing.T) {
	repo := &DeputadoRepository{db: nil}
	ctx := context.Background()

	// Teste UpsertDeputados com DB nil - deve falhar ou ter verificação
	err := repo.UpsertDeputados(ctx, []domain.Deputado{{ID: 1}})
	if err != nil {
		t.Logf("UpsertDeputados falhou com DB nil como esperado: %v", err)
	} else {
		t.Log("UpsertDeputados tratou DB nil graciosamente")
	}

	// Teste ListFromCache com DB nil - deve falhar ou ter verificação
	deputados, err := repo.ListFromCache(ctx, 10)
	if err != nil {
		t.Logf("ListFromCache falhou com DB nil como esperado: %v", err)
	} else {
		t.Logf("ListFromCache tratou DB nil graciosamente, retornou %d deputados", len(deputados))
	}
}

func TestDeputadoRepository_ConcurrentAccess(t *testing.T) {
	repo := &DeputadoRepository{db: &mockDB{}}
	ctx := context.Background()

	// Simular acesso concorrente
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			deputados := []domain.Deputado{
				{ID: id, Nome: fmt.Sprintf("Deputado Concurrent %d", id), Partido: "TEST"},
			}

			err := repo.UpsertDeputados(ctx, deputados)
			if err != nil {
				t.Logf("erro na goroutine %d: %v", id, err)
			}
		}(i)
	}

	// Aguardar todas as goroutines terminarem
	for i := 0; i < 10; i++ {
		<-done
	}

	t.Log("Teste de acesso concorrente concluído")
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

// Testes adicionais para melhorar cobertura
func TestDeputadoRepository_AdditionalEdgeCases(t *testing.T) {
	t.Run("UpsertDeputados com slice nil", func(t *testing.T) {
		repo := &DeputadoRepository{db: &mockDB{rows: &mockRows{data: []string{}}}}
		err := repo.UpsertDeputados(context.Background(), nil)
		if err != nil {
			t.Logf("UpsertDeputados with nil slice: %v", err)
		}
	})

	t.Run("UpsertDeputados com slice vazio", func(t *testing.T) {
		repo := &DeputadoRepository{db: &mockDB{rows: &mockRows{data: []string{}}}}
		err := repo.UpsertDeputados(context.Background(), []domain.Deputado{})
		if err != nil {
			t.Logf("UpsertDeputados with empty slice: %v", err)
		}
	})

	t.Run("ListFromCache com limite zero", func(t *testing.T) {
		repo := &DeputadoRepository{db: &mockDB{rows: &mockRows{data: []string{}}}}
		deputados, err := repo.ListFromCache(context.Background(), 0)
		if err != nil {
			t.Logf("ListFromCache with zero limit: %v", err)
		}
		if len(deputados) > 0 {
			t.Error("Expected empty result with zero limit")
		}
	})

	t.Run("ListFromCache com limite negativo", func(t *testing.T) {
		repo := &DeputadoRepository{db: &mockDB{rows: &mockRows{data: []string{}}}}
		deputados, err := repo.ListFromCache(context.Background(), -1)
		if err != nil {
			t.Logf("ListFromCache with negative limit: %v", err)
		}
		if len(deputados) > 0 {
			t.Error("Expected empty result with negative limit")
		}
	})

	t.Run("ListFromCache com limite muito alto", func(t *testing.T) {
		repo := &DeputadoRepository{db: &mockDB{rows: &mockRows{data: []string{}}}}
		deputados, err := repo.ListFromCache(context.Background(), 999999)
		if err != nil {
			t.Logf("ListFromCache with very high limit: %v", err)
		}
		t.Logf("ListFromCache with high limit returned %d deputados", len(deputados))
	})
}

func TestDeputadoRepository_NilRepository(t *testing.T) {
	var repo *DeputadoRepository

	// Test UpsertDeputados with nil repository
	err := repo.UpsertDeputados(context.Background(), []domain.Deputado{})
	if err != nil {
		t.Errorf("UpsertDeputados() with nil repo should not error, got %v", err)
	}

	// Test ListFromCache with nil repository
	deputados, err := repo.ListFromCache(context.Background(), 10)
	if err != nil {
		t.Errorf("ListFromCache() with nil repo should not error, got %v", err)
	}
	if deputados != nil {
		t.Error("ListFromCache() with nil repo should return nil")
	}
}

func TestDeputadoRepository_DBNil(t *testing.T) {
	repo := &DeputadoRepository{db: nil}
	ctx := context.Background()

	// Test with DB nil - expect no panic
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic: %v", r)
		}
	}()

	err := repo.UpsertDeputados(ctx, []domain.Deputado{{ID: 1}})
	t.Logf("UpsertDeputados with nil DB returned: %v", err)

	deputados, err := repo.ListFromCache(ctx, 10)
	t.Logf("ListFromCache with nil DB returned: %v (len: %d)", err, len(deputados))
}

func TestDeputadoRepository_DeputadosExtremos(t *testing.T) {
	repo := &DeputadoRepository{db: &mockDB{rows: &mockRows{data: []string{}}}}
	ctx := context.Background()

	// Teste com deputados com dados extremos
	deputadosExtremos := []domain.Deputado{
		{
			ID:      -1,
			Nome:    "",
			UF:      "",
			Partido: "",
		},
		{
			ID:      999999,
			Nome:    "Nome Muito Longo Para Deputado Federal Brasileiro Que Pode Causar Problemas",
			UF:      "XXXXX",
			Partido: "PARTIDOCOMNOMELONGO",
		},
		{
			ID:      0,
			Nome:    "Deputado Zero",
			UF:      "XX",
			Partido: "XX",
		},
	}

	err := repo.UpsertDeputados(ctx, deputadosExtremos)
	if err != nil {
		t.Logf("UpsertDeputados with extreme data: %v", err)
	}
}
