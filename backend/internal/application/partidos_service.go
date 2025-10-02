package application

import (
	"context"
	"fmt"
	"to-de-olho-backend/internal/domain"
	"to-de-olho-backend/internal/infrastructure/repository"
)

// CamaraPartidosPort é a abstração do cliente HTTP para obter partidos
type CamaraPartidosPort interface {
	FetchPartidos(ctx context.Context) ([]domain.Partido, error)
}

type PartidosService struct {
	client CamaraPartidosPort
	repo   *repository.PartidoRepository
}

func NewPartidosService(client CamaraPartidosPort, repo *repository.PartidoRepository) *PartidosService {
	return &PartidosService{client: client, repo: repo}
}

// ListarPartidos busca partidos da API e persiste via repository
func (ps *PartidosService) ListarPartidos(ctx context.Context) ([]domain.Partido, error) {
	partidos, err := ps.client.FetchPartidos(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar partidos da API: %w", err)
	}

	if err := ps.repo.UpsertPartidos(ctx, partidos); err != nil {
		return partidos, fmt.Errorf("erro ao persistir partidos: %w", err)
	}

	return partidos, nil
}
