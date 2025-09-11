package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"to-de-olho-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB abstracts the subset of pgxpool.Pool used, enabling mocking in unit tests.
type DB interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type DeputadoRepository struct {
	db DB
}

func NewDeputadoRepository(db *pgxpool.Pool) *DeputadoRepository { // keep signature for existing callers
	return &DeputadoRepository{db: db}
}

func (r *DeputadoRepository) UpsertDeputados(ctx context.Context, deps []domain.Deputado) error {
	if r == nil || r.db == nil || len(deps) == 0 {
		return nil
	}
	for _, d := range deps {
		b, _ := json.Marshal(d)
		_, err := r.db.Exec(ctx, `INSERT INTO deputados_cache (id, payload, updated_at)
            VALUES ($1, $2, NOW())
            ON CONFLICT (id) DO UPDATE SET payload = EXCLUDED.payload, updated_at = NOW()`, d.ID, string(b))
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *DeputadoRepository) ListFromCache(ctx context.Context, limit int) ([]domain.Deputado, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	rows, err := r.db.Query(ctx, `SELECT payload FROM deputados_cache ORDER BY updated_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Deputado
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var d domain.Deputado
		if err := json.Unmarshal([]byte(payload), &d); err != nil {
			return nil, fmt.Errorf("erro ao decodificar cache: %w", err)
		}
		out = append(out, d)
	}
	return out, nil
}
