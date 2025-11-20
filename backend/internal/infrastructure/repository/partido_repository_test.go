package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"to-de-olho-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type fakeDB struct {
	execErr   error
	execCalls []execCall
}

type execCall struct {
	sql  string
	args []interface{}
}

func (f *fakeDB) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	f.execCalls = append(f.execCalls, execCall{sql: sql, args: arguments})
	if f.execErr != nil {
		return pgconn.CommandTag{}, f.execErr
	}
	return pgconn.CommandTag{}, nil
}

func (f *fakeDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return nil, fmt.Errorf("not implemented")
}

func (f *fakeDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return nil
}

func TestPartidoRepository_UpsertPartidos(t *testing.T) {
	t.Parallel()

	sample := []domain.Partido{
		{ID: 1, Sigla: "ABC", Nome: "Partido ABC", URI: "https://example.com", Payload: map[string]interface{}{"sigla": "ABC"}},
		{ID: 2, Sigla: "XYZ", Nome: "Partido XYZ", URI: "https://example.com/xyz", Payload: map[string]interface{}{"sigla": "XYZ"}},
	}

	tests := []struct {
		name          string
		repo          *PartidoRepository
		partidos      []domain.Partido
		wantErr       bool
		wantErrString string
		wantExecs     int
	}{
		{
			name:      "repo sem db",
			repo:      &PartidoRepository{},
			partidos:  sample,
			wantExecs: 0,
		},
		{
			name:      "lista vazia",
			repo:      &PartidoRepository{db: &fakeDB{}},
			wantExecs: 0,
		},
		{
			name:      "sucesso",
			repo:      &PartidoRepository{db: &fakeDB{}},
			partidos:  sample,
			wantExecs: len(sample),
		},
		{
			name:          "erro no exec",
			repo:          &PartidoRepository{db: &fakeDB{execErr: errors.New("falha")}},
			partidos:      sample[:1],
			wantErr:       true,
			wantErrString: "upsert partido",
			wantExecs:     1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.repo.UpsertPartidos(context.Background(), tt.partidos)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("esperava erro, mas não ocorreu")
				}
				if tt.wantErrString != "" && !strings.Contains(err.Error(), tt.wantErrString) {
					t.Fatalf("erro %q não contém %q", err.Error(), tt.wantErrString)
				}
			} else if err != nil {
				t.Fatalf("não esperava erro, ocorreu: %v", err)
			}

			fake, _ := tt.repo.db.(*fakeDB)
			if fake != nil && len(fake.execCalls) != tt.wantExecs {
				t.Fatalf("execCalls = %d, esperado %d", len(fake.execCalls), tt.wantExecs)
			}

			if fake != nil && tt.wantExecs > 0 {
				for _, call := range fake.execCalls {
					if !strings.HasPrefix(call.sql, "INSERT INTO partidos") {
						t.Fatalf("sql inesperado: %s", call.sql)
					}
					if len(call.args) < 4 {
						t.Fatalf("args insuficientes: %d", len(call.args))
					}
				}
			}
		})
	}
}
