package repository

import (
	"context"
	"encoding/json"
	"testing"

	"to-de-olho-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// MockDB implementa a interface DB para testes
type MockDBProposicoes struct {
	execFunc  func(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	queryFunc func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

func (m *MockDBProposicoes) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	if m.execFunc != nil {
		return m.execFunc(ctx, sql, arguments...)
	}
	return pgconn.CommandTag{}, nil
}

func (m *MockDBProposicoes) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, sql, args...)
	}
	return nil, nil
}

// MockRows implementa pgx.Rows para testes
type MockRowsProposicoes struct {
	rows     [][]interface{}
	current  int
	err      error
	scanFunc func(dest ...interface{}) error
}

func NewMockRowsProposicoes(data [][]interface{}) *MockRowsProposicoes {
	return &MockRowsProposicoes{
		rows:    data,
		current: -1,
	}
}

func (m *MockRowsProposicoes) Next() bool {
	m.current++
	return m.current < len(m.rows)
}

func (m *MockRowsProposicoes) Scan(dest ...interface{}) error {
	if m.scanFunc != nil {
		return m.scanFunc(dest...)
	}
	if m.current < 0 || m.current >= len(m.rows) {
		return pgx.ErrNoRows
	}

	row := m.rows[m.current]
	for i, dest := range dest {
		if i < len(row) {
			switch d := dest.(type) {
			case *string:
				if str, ok := row[i].(string); ok {
					*d = str
				}
			case *int:
				if num, ok := row[i].(int); ok {
					*d = num
				}
			}
		}
	}
	return nil
}

func (m *MockRowsProposicoes) Close() {}

func (m *MockRowsProposicoes) Err() error {
	return m.err
}

func (m *MockRowsProposicoes) CommandTag() pgconn.CommandTag {
	return pgconn.CommandTag{}
}

func (m *MockRowsProposicoes) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

func (m *MockRowsProposicoes) Values() ([]interface{}, error) {
	return nil, nil
}

func (m *MockRowsProposicoes) RawValues() [][]byte {
	return nil
}

func (m *MockRowsProposicoes) Conn() *pgx.Conn {
	return nil
}

func createTestProposicaoForRepo(id int, siglaTipo string, numero int, ano int, ementa string) *domain.Proposicao {
	return &domain.Proposicao{
		ID:               id,
		SiglaTipo:        siglaTipo,
		Numero:           numero,
		Ano:              ano,
		Ementa:           ementa,
		DataApresentacao: "2024-01-01",
		StatusProposicao: domain.StatusProposicao{
			DescricaoSituacao: "Em tramitação",
			DataHora:          "2024-01-01T10:00:00",
		},
	}
}

func TestProposicaoRepository_UpsertProposicoes(t *testing.T) {
	tests := []struct {
		name        string
		proposicoes []domain.Proposicao
		mockExec    func(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
		wantErr     bool
	}{
		{
			name: "success - upsert proposições",
			proposicoes: []domain.Proposicao{
				*createTestProposicaoForRepo(1, "PL", 123, 2024, "Projeto de lei teste"),
			},
			mockExec: func(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, nil
			},
			wantErr: false,
		},
		{
			name:        "success - lista vazia",
			proposicoes: []domain.Proposicao{},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockDBProposicoes{
				execFunc: tt.mockExec,
			}

			repo := &ProposicaoRepository{db: mockDB}
			err := repo.UpsertProposicoes(context.Background(), tt.proposicoes)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpsertProposicoes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProposicaoRepository_GetProposicaoPorID(t *testing.T) {
	testProposicao := createTestProposicaoForRepo(123, "PL", 456, 2024, "Ementa teste")
	payload, _ := json.Marshal(testProposicao)

	tests := []struct {
		name      string
		id        int
		mockQuery func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
		want      *domain.Proposicao
		wantErr   bool
	}{
		{
			name: "success - proposição encontrada",
			id:   123,
			mockQuery: func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
				mockRows := NewMockRowsProposicoes([][]interface{}{
					{string(payload)},
				})
				return mockRows, nil
			},
			want:    testProposicao,
			wantErr: false,
		},
		{
			name: "error - proposição não encontrada",
			id:   999,
			mockQuery: func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
				mockRows := &MockRowsProposicoes{
					scanFunc: func(dest ...interface{}) error {
						return pgx.ErrNoRows
					},
				}
				return mockRows, nil
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockDBProposicoes{
				queryFunc: tt.mockQuery,
			}

			repo := &ProposicaoRepository{db: mockDB}
			got, err := repo.GetProposicaoPorID(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetProposicaoPorID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != nil && got != nil {
				if got.ID != tt.want.ID {
					t.Errorf("GetProposicaoPorID() got.ID = %v, want %v", got.ID, tt.want.ID)
				}
			}
		})
	}
}

func TestProposicaoRepository_ListProposicoes(t *testing.T) {
	testProposicao := createTestProposicaoForRepo(1, "PL", 123, 2024, "Ementa teste")
	payload, _ := json.Marshal(testProposicao)

	tests := []struct {
		name      string
		filtros   *domain.ProposicaoFilter
		mockQuery func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
		wantCount int
		wantErr   bool
	}{
		{
			name: "success - proposições encontradas",
			filtros: &domain.ProposicaoFilter{
				SiglaTipo: "PL",
				Pagina:    1,
				Limite:    20,
			},
			mockQuery: func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
				mockRows := NewMockRowsProposicoes([][]interface{}{
					{string(payload)},
				})
				return mockRows, nil
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:    "success - nenhuma proposição encontrada",
			filtros: nil,
			mockQuery: func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
				mockRows := NewMockRowsProposicoes([][]interface{}{})
				return mockRows, nil
			},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockDBProposicoes{
				queryFunc: tt.mockQuery,
			}

			repo := &ProposicaoRepository{db: mockDB}
			proposicoes, total, err := repo.ListProposicoes(context.Background(), tt.filtros)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListProposicoes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(proposicoes) != tt.wantCount {
				t.Errorf("ListProposicoes() len = %v, want %v", len(proposicoes), tt.wantCount)
			}

			if total != tt.wantCount {
				t.Errorf("ListProposicoes() total = %v, want %v", total, tt.wantCount)
			}
		})
	}
}

func TestProposicaoRepository_buildWhereClause(t *testing.T) {
	repo := &ProposicaoRepository{}

	tests := []struct {
		name           string
		filtros        *domain.ProposicaoFilter
		wantConditions bool
		wantArgsCount  int
	}{
		{
			name:           "nil filtros",
			filtros:        nil,
			wantConditions: false,
			wantArgsCount:  0,
		},
		{
			name: "filtro por SiglaTipo",
			filtros: &domain.ProposicaoFilter{
				SiglaTipo: "PL",
			},
			wantConditions: true,
			wantArgsCount:  1,
		},
		{
			name: "múltiplos filtros",
			filtros: &domain.ProposicaoFilter{
				SiglaTipo:         "PL",
				SiglaUfAutor:      "SP",
				SiglaPartidoAutor: "PT",
			},
			wantConditions: true,
			wantArgsCount:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			whereClause, args := repo.buildWhereClause(tt.filtros)

			if tt.wantConditions && whereClause == "" {
				t.Error("buildWhereClause() expected conditions but got empty string")
			}

			if !tt.wantConditions && whereClause != "" {
				t.Error("buildWhereClause() expected no conditions but got some")
			}

			if len(args) != tt.wantArgsCount {
				t.Errorf("buildWhereClause() args count = %v, want %v", len(args), tt.wantArgsCount)
			}
		})
	}
}

func TestProposicaoRepository_NilRepository(t *testing.T) {
	var repo *ProposicaoRepository

	// Test UpsertProposicoes with nil repository
	err := repo.UpsertProposicoes(context.Background(), []domain.Proposicao{})
	if err != nil {
		t.Errorf("UpsertProposicoes() with nil repo should not error, got %v", err)
	}

	// Test ListProposicoes with nil repository
	proposicoes, total, err := repo.ListProposicoes(context.Background(), nil)
	if err != nil {
		t.Errorf("ListProposicoes() with nil repo should not error, got %v", err)
	}
	if proposicoes != nil || total != 0 {
		t.Error("ListProposicoes() with nil repo should return nil, 0")
	}

	// Test GetProposicaoPorID with nil repository
	proposicao, err := repo.GetProposicaoPorID(context.Background(), 123)
	if err != nil {
		t.Errorf("GetProposicaoPorID() with nil repo should not error, got %v", err)
	}
	if proposicao != nil {
		t.Error("GetProposicaoPorID() with nil repo should return nil")
	}
}
