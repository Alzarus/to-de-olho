package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DeputadoRepository struct {
	db *pgxpool.Pool
}

func NewDeputadoRepository(db *pgxpool.Pool) *DeputadoRepository {
	return &DeputadoRepository{db: db}
}

func (r *DeputadoRepository) UpsertDeputados(ctx context.Context, deps []Deputado) error {
	if r == nil || r.db == nil || len(deps) == 0 {
		return nil
	}
	// Tabela simples para cache/persistência leve (JSON)
	// CREATE TABLE IF NOT EXISTS deputados_cache (id INT PRIMARY KEY, payload JSONB, updated_at TIMESTAMP);
	_, err := r.db.Exec(ctx, `CREATE TABLE IF NOT EXISTS deputados_cache (
        id INT PRIMARY KEY,
        payload JSONB NOT NULL,
        updated_at TIMESTAMP NOT NULL
    )`)
	if err != nil {
		return err
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

func (r *DeputadoRepository) ListFromCache(ctx context.Context, limit int) ([]Deputado, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	rows, err := r.db.Query(ctx, `SELECT payload FROM deputados_cache ORDER BY updated_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Deputado
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var d Deputado
		if err := json.Unmarshal([]byte(payload), &d); err != nil {
			return nil, fmt.Errorf("erro ao decodificar cache: %w", err)
		}
		out = append(out, d)
	}
	return out, nil
}
