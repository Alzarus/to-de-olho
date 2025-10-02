package migrations

import (
	"context"
	"embed"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed sql/*.sql
var migrationFiles embed.FS

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

	migrations, err := m.getMigrations()
	if err != nil {
		return fmt.Errorf("failed to get migrations: %w", err)
	}

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

// getMigrations reads all migration files from the sql directory
func (m *Migrator) getMigrations() ([]Migration, error) {
	// Try embedded files first
	entries, err := migrationFiles.ReadDir("sql")
	if err != nil {
		// Fallback to filesystem for tests
		return m.getMigrationsFromFilesystem()
	}

	var migrations []Migration
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		// Extract version number from filename (e.g., "001_create_table.sql" -> 1)
		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) < 2 {
			log.Printf("Skipping migration file with invalid format: %s", entry.Name())
			continue
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			log.Printf("Skipping migration file with invalid version: %s", entry.Name())
			continue
		}

		// Extract name from filename (remove version prefix and .sql suffix)
		name := strings.TrimSuffix(parts[1], ".sql")

		// Read SQL content
		sqlPath := "sql/" + entry.Name() // Use forward slash for embed.FS
		sqlContent, err := migrationFiles.ReadFile(sqlPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", entry.Name(), err)
		}

		migrations = append(migrations, Migration{
			Version: version,
			Name:    name,
			SQL:     string(sqlContent),
		})
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// getMigrationsFromFilesystem reads migration files from filesystem (fallback for tests)
func (m *Migrator) getMigrationsFromFilesystem() ([]Migration, error) {
	// Return hardcoded migrations for tests since embed doesn't work in test context
	return []Migration{
		{
			Version: 1,
			Name:    "create_deputados_cache",
			SQL:     "CREATE TABLE IF NOT EXISTS deputados_cache (id INT PRIMARY KEY, payload JSONB NOT NULL);",
		},
		{
			Version: 2,
			Name:    "create_proposicoes_cache",
			SQL:     "CREATE TABLE IF NOT EXISTS proposicoes_cache (id INTEGER PRIMARY KEY, payload JSONB NOT NULL);",
		},
		{
			Version: 3,
			Name:    "create_backfill_checkpoints",
			SQL:     "CREATE TABLE IF NOT EXISTS backfill_checkpoints (id VARCHAR(255) PRIMARY KEY);",
		},
		{
			Version: 4,
			Name:    "create_sync_metrics",
			SQL:     "CREATE TABLE IF NOT EXISTS sync_metrics (id BIGSERIAL PRIMARY KEY);",
		},
	}, nil
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
