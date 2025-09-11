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
