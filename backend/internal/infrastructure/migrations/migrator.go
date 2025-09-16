package migrations

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Migration represents a database migration
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// Migrator handles database migrations
type Migrator struct {
	db *pgxpool.Pool
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *pgxpool.Pool) *Migrator {
	return &Migrator{db: db}
}

// Run executes all pending migrations
func (m *Migrator) Run(ctx context.Context) error {
	if err := m.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	migrations := m.getMigrations()
	appliedMigrations, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	for _, migration := range migrations {
		if m.isMigrationApplied(appliedMigrations, migration.Version) {
			log.Printf("Migration %03d_%s already applied, skipping", migration.Version, migration.Name)
			continue
		}

		log.Printf("Applying migration %03d_%s", migration.Version, migration.Name)
		if err := m.applyMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to apply migration %03d_%s: %w", migration.Version, migration.Name, err)
		}
		log.Printf("Successfully applied migration %03d_%s", migration.Version, migration.Name)
	}

	log.Println("All migrations completed successfully")
	return nil
}

// getMigrations returns all available migrations (hardcoded for CI/CD compatibility)
func (m *Migrator) getMigrations() []Migration {
	return []Migration{
		{
			Version: 1,
			Name:    "create_deputados_cache",
			SQL: `-- Migration: Create deputados_cache table
-- Version: 001
-- Description: Initial table for caching deputy data from API

CREATE TABLE IF NOT EXISTS deputados_cache (
    id INT PRIMARY KEY,
    payload JSONB NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_deputados_cache_updated_at ON deputados_cache(updated_at);`,
		},
		{
			Version: 2,
			Name:    "create_proposicoes_cache",
			SQL: `-- Migration: 002_create_proposicoes_cache.sql
-- Descrição: Cria tabela para cache de proposições da Câmara dos Deputados

-- Criar extensão pg_trgm se não existir (para busca textual)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS proposicoes_cache (
    id INTEGER PRIMARY KEY,
    payload JSONB NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Índices para melhorar performance das queries
CREATE INDEX IF NOT EXISTS idx_proposicoes_cache_updated_at ON proposicoes_cache(updated_at DESC);

-- Índices BTREE para campos de texto/número (mais eficientes para igualdade e range)
CREATE INDEX IF NOT EXISTS idx_proposicoes_cache_sigla_tipo ON proposicoes_cache((payload->>'siglaTipo'));
CREATE INDEX IF NOT EXISTS idx_proposicoes_cache_numero ON proposicoes_cache(((payload->>'numero')::int));
CREATE INDEX IF NOT EXISTS idx_proposicoes_cache_ano ON proposicoes_cache(((payload->>'ano')::int));
CREATE INDEX IF NOT EXISTS idx_proposicoes_cache_uf_autor ON proposicoes_cache((payload->>'siglaUfAutor'));
CREATE INDEX IF NOT EXISTS idx_proposicoes_cache_partido_autor ON proposicoes_cache((payload->>'siglaPartidoAutor'));

-- Índice GIN apenas para busca textual na ementa
CREATE INDEX IF NOT EXISTS idx_proposicoes_cache_ementa ON proposicoes_cache USING GIN ((payload->>'ementa') gin_trgm_ops);

-- Comentários para documentação
COMMENT ON TABLE proposicoes_cache IS 'Cache de proposições da Câmara dos Deputados para melhorar performance';
COMMENT ON COLUMN proposicoes_cache.id IS 'ID único da proposição na API da Câmara';
COMMENT ON COLUMN proposicoes_cache.payload IS 'Dados completos da proposição em formato JSON';
COMMENT ON COLUMN proposicoes_cache.updated_at IS 'Data e hora da última atualização do cache';`,
		},
		{
			Version: 3,
			Name:    "create_backfill_checkpoints",
			SQL: `-- Migration: 003_create_backfill_checkpoints.sql
-- Descrição: Cria tabela para checkpoints do processo de backfill histórico

CREATE TABLE IF NOT EXISTS backfill_checkpoints (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,                              -- 'deputados', 'proposicoes', 'despesas'
    status VARCHAR(20) NOT NULL DEFAULT 'pending',          -- 'pending', 'in_progress', 'completed', 'failed'
    progress JSONB NOT NULL DEFAULT '{}',                   -- Progresso serializado
    metadata JSONB NOT NULL DEFAULT '{}',                   -- Metadados adicionais
    started_at TIMESTAMP WITH TIME ZONE,                    -- Quando foi iniciado
    completed_at TIMESTAMP WITH TIME ZONE,                  -- Quando foi completado
    error_message TEXT,                                      -- Mensagem de erro se falhou
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),      -- Quando foi criado
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()       -- Última atualização
);

-- Índices para performance
CREATE INDEX IF NOT EXISTS idx_backfill_checkpoints_type ON backfill_checkpoints(type);
CREATE INDEX IF NOT EXISTS idx_backfill_checkpoints_status ON backfill_checkpoints(status);
CREATE INDEX IF NOT EXISTS idx_backfill_checkpoints_created_at ON backfill_checkpoints(created_at);

-- Comentários para documentação
COMMENT ON TABLE backfill_checkpoints IS 'Checkpoints do processo de backfill histórico para resumabilidade';
COMMENT ON COLUMN backfill_checkpoints.id IS 'ID único do checkpoint (tipo_timestamp)';
COMMENT ON COLUMN backfill_checkpoints.type IS 'Tipo de dados sendo processados';
COMMENT ON COLUMN backfill_checkpoints.status IS 'Status atual do checkpoint';
COMMENT ON COLUMN backfill_checkpoints.progress IS 'Progresso detalhado em JSON';
COMMENT ON COLUMN backfill_checkpoints.metadata IS 'Metadados específicos do tipo de backfill';`,
		},
		{
			Version: 4,
			Name:    "create_sync_metrics",
			SQL: `-- Migration: 004_create_sync_metrics.sql
-- Descrição: Cria tabela para métricas de sincronização incremental

CREATE TABLE IF NOT EXISTS sync_metrics (
    id BIGSERIAL PRIMARY KEY,
    sync_type VARCHAR(20) NOT NULL,                     -- 'daily', 'quick'
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    duration_ms INTEGER NOT NULL,
    deputados_updated INTEGER DEFAULT 0,
    proposicoes_updated INTEGER DEFAULT 0,
    errors_count INTEGER DEFAULT 0,
    errors TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Índices para performance
CREATE INDEX IF NOT EXISTS idx_sync_metrics_type_time ON sync_metrics(sync_type, start_time DESC);
CREATE INDEX IF NOT EXISTS idx_sync_metrics_start_time ON sync_metrics(start_time DESC);

-- Comentários para documentação
COMMENT ON TABLE sync_metrics IS 'Métricas de sincronização incremental e diária';
COMMENT ON COLUMN sync_metrics.sync_type IS 'Tipo de sincronização: daily ou quick';
COMMENT ON COLUMN sync_metrics.duration_ms IS 'Duração da sincronização em milissegundos';
COMMENT ON COLUMN sync_metrics.errors IS 'Lista de erros ocorridos durante a sincronização';`,
		},
	}
}

// createMigrationsTable creates the migrations tracking table
func (m *Migrator) createMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err := m.db.Exec(ctx, query)
	return err
}

// getAppliedMigrations returns a list of applied migration versions
func (m *Migrator) getAppliedMigrations(ctx context.Context) (map[int]bool, error) {
	rows, err := m.db.Query(ctx, "SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// isMigrationApplied checks if a migration has been applied
func (m *Migrator) isMigrationApplied(applied map[int]bool, version int) bool {
	return applied[version]
}

// applyMigration applies a single migration
func (m *Migrator) applyMigration(ctx context.Context, migration Migration) error {
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Execute the migration SQL
	_, err = tx.Exec(ctx, migration.SQL)
	if err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record the migration as applied
	_, err = tx.Exec(ctx,
		"INSERT INTO schema_migrations (version) VALUES ($1)",
		migration.Version)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit(ctx)
}
