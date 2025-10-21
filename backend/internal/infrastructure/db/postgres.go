package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"to-de-olho-backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestDB é um wrapper simples usado em testes que esperam NewPostgreSQL
type TestDB struct {
	pool *pgxpool.Pool
}

func (t *TestDB) GetDB() *pgxpool.Pool { return t.pool }
func (t *TestDB) Close() {
	if t.pool != nil {
		t.pool.Close()
	}
}

// NewPostgreSQL cria um TestDB a partir de uma connection string (compatibilidade com testes)
func NewPostgreSQL(connStr string) (*TestDB, error) {
	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}
	return &TestDB{pool: pool}, nil
}

func NewPostgresPool(ctx context.Context) (*pgxpool.Pool, error) {
	host := getenv("POSTGRES_HOST", "postgres")
	port := getenv("POSTGRES_PORT", "5432")
	user := getenv("POSTGRES_USER", "postgres")
	pass := getenv("POSTGRES_PASSWORD", "postgres123")
	db := getenv("POSTGRES_DB", "to_de_olho")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?pool_max_conns=5", user, pass, host, port, db)
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 5
	cfg.MinConns = 0
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 10 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return pool, nil
}

// NewPostgresPoolFromConfig creates a connection pool using config struct
func NewPostgresPoolFromConfig(ctx context.Context, cfg *config.DatabaseConfig) (*pgxpool.Pool, error) {
	dsn := cfg.ConnectionString()

	pgxCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Apply configuration coming from config
	pgxCfg.MaxConns = int32(cfg.MaxConns)
	pgxCfg.MinConns = cfg.MinConns
	pgxCfg.MaxConnLifetime = cfg.MaxConnLifetime
	pgxCfg.MaxConnIdleTime = cfg.MaxConnIdleTime

	// Create pool but don't immediately return on first ping failure; implement retry with backoff
	var pool *pgxpool.Pool
	var lastErr error

	// Retry parameters - configurable via env if needed in future
	maxAttempts := 6
	backoff := 500 * time.Millisecond

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			// small backoff before retrying
			time.Sleep(backoff)
			backoff = backoff * 2
		}

		pool, err = pgxpool.NewWithConfig(ctx, pgxCfg)
		if err != nil {
			lastErr = fmt.Errorf("unable to create connection pool: %w", err)
			continue
		}

		// Test connection with a short timeout derived from context
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = pool.Ping(pingCtx)
		cancel()
		if err == nil {
			// success
			return pool, nil
		}

		// Close pool and save error to retry
		pool.Close()
		lastErr = fmt.Errorf("unable to ping database: %w", err)
	}

	// All attempts failed
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("failed to create postgres pool")
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
