package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"to-de-olho-backend/internal/domain"
)

type fakeCamaraPartidosPort struct {
	partidos []domain.Partido
	err      error
}

func (f *fakeCamaraPartidosPort) FetchPartidos(ctx context.Context) ([]domain.Partido, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.partidos, nil
}

type fakePartidoRepository struct {
	called bool
	err    error
	last   []domain.Partido
}

func (f *fakePartidoRepository) UpsertPartidos(ctx context.Context, partidos []domain.Partido) error {
	f.called = true
	f.last = partidos
	return f.err
}

func TestPartidosService_ListarPartidos(t *testing.T) {
	t.Parallel()

	sample := []domain.Partido{{ID: 10, Sigla: "ABC", Nome: "Partido ABC"}}

	tests := []struct {
		name            string
		client          *fakeCamaraPartidosPort
		repo            *fakePartidoRepository
		wantErr         bool
		wantErrContains string
		wantPartidos    []domain.Partido
		repoCalled      bool
	}{
		{
			name:            "erro ao buscar na API",
			client:          &fakeCamaraPartidosPort{err: errors.New("api fora")},
			repo:            &fakePartidoRepository{},
			wantErr:         true,
			wantErrContains: "buscar partidos",
			wantPartidos:    nil,
			repoCalled:      false,
		},
		{
			name:            "erro ao persistir",
			client:          &fakeCamaraPartidosPort{partidos: sample},
			repo:            &fakePartidoRepository{err: errors.New("falha banco")},
			wantErr:         true,
			wantErrContains: "persistir partidos",
			wantPartidos:    sample,
			repoCalled:      true,
		},
		{
			name:         "sucesso",
			client:       &fakeCamaraPartidosPort{partidos: sample},
			repo:         &fakePartidoRepository{},
			wantPartidos: sample,
			repoCalled:   true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := NewPartidosService(tt.client, tt.repo)

			partidos, err := service.ListarPartidos(context.Background())

			if tt.wantErr {
				if err == nil {
					t.Fatalf("esperava erro, mas não ocorreu")
				}
				if tt.wantErrContains != "" && !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Fatalf("erro %q não contém %q", err.Error(), tt.wantErrContains)
				}
			} else if err != nil {
				t.Fatalf("não esperava erro, ocorreu: %v", err)
			}

			if diff := comparePartidos(partidos, tt.wantPartidos); diff != "" {
				t.Fatalf("partidos retornados divergentes: %s", diff)
			}

			if tt.repo != nil {
				if tt.repo.called != tt.repoCalled {
					t.Fatalf("repoCalled = %v, esperado %v", tt.repo.called, tt.repoCalled)
				}
			}
		})
	}
}

func comparePartidos(got, want []domain.Partido) string {
	if len(got) != len(want) {
		return fmt.Sprintf("len=%d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i].ID != want[i].ID || got[i].Sigla != want[i].Sigla || got[i].Nome != want[i].Nome {
			return fmt.Sprintf("índice %d divergente: %+v != %+v", i, got[i], want[i])
		}
	}
	return ""
}
