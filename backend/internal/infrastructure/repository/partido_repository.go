package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"to-de-olho-backend/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PartidoRepository struct {
	db DB
}

func NewPartidoRepository(db *pgxpool.Pool) *PartidoRepository {
	return &PartidoRepository{db: db}
}

// UpsertPartidos insere ou atualiza a lista de partidos
func (r *PartidoRepository) UpsertPartidos(ctx context.Context, partidos []domain.Partido) error {
	if r == nil || r.db == nil || len(partidos) == 0 {
		return nil
	}

	for _, p := range partidos {
		b, _ := json.Marshal(p.Payload)
		_, err := r.db.Exec(ctx, `INSERT INTO partidos (id, sigla, nome, uri, payload, updated_at)
            VALUES ($1, $2, $3, $4, $5, NOW())
            ON CONFLICT (id) DO UPDATE SET sigla = EXCLUDED.sigla, nome = EXCLUDED.nome, uri = EXCLUDED.uri, payload = EXCLUDED.payload, updated_at = NOW()`,
			p.ID, p.Sigla, p.Nome, p.URI, string(b))
		if err != nil {
			return fmt.Errorf("erro ao upsert partido %d: %w", p.ID, err)
		}
	}

	return nil
}
