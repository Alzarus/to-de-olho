package migrations

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed *.sql
var migrationFiles embed.FS

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

	migrations, err := m.loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	appliedMigrations, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	for _, migration := range migrations {
		if _, applied := appliedMigrations[migration.Version]; !applied {
			log.Printf("Applying migration %d: %s", migration.Version, migration.Name)

			if err := m.applyMigration(ctx, migration); err != nil {
				return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
			}

			log.Printf("Successfully applied migration %d", migration.Version)
		}
	}

	log.Printf("All migrations applied successfully")
	return nil
}

func (m *Migrator) createMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := m.db.Exec(ctx, query)
	return err
}

func (m *Migrator) loadMigrations() ([]Migration, error) {
	var migrations []Migration

	err := fs.WalkDir(migrationFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		content, err := migrationFiles.ReadFile(path)
		if err != nil {
			return err
		}

		// Extract version from filename (e.g., "001_create_table.sql" -> 1)
		parts := strings.Split(path, "_")
		if len(parts) < 2 {
			return fmt.Errorf("invalid migration filename: %s", path)
		}

		var version int
		if _, err := fmt.Sscanf(parts[0], "%d", &version); err != nil {
			return fmt.Errorf("invalid version in filename %s: %w", path, err)
		}

		name := strings.TrimSuffix(strings.Join(parts[1:], "_"), ".sql")

		migrations = append(migrations, Migration{
			Version: version,
			Name:    name,
			SQL:     string(content),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func (m *Migrator) getAppliedMigrations(ctx context.Context) (map[int]bool, error) {
	query := "SELECT version FROM schema_migrations"
	rows, err := m.db.Query(ctx, query)
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

func (m *Migrator) applyMigration(ctx context.Context, migration Migration) error {
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Execute migration SQL
	if _, err := tx.Exec(ctx, migration.SQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration as applied
	insertQuery := `
		INSERT INTO schema_migrations (version, name) 
		VALUES ($1, $2)
	`
	if _, err := tx.Exec(ctx, insertQuery, migration.Version, migration.Name); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit(ctx)
}
