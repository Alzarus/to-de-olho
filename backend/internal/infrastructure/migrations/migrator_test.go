package migrations

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestNewMigrator(t *testing.T) {
	pool := &pgxpool.Pool{}
	migrator := NewMigrator(pool)

	if migrator == nil {
		t.Fatal("NewMigrator não deveria retornar nil")
	}

	if migrator.db != pool {
		t.Error("migrator deveria conter a referência do pool")
	}
}

func TestMigrator_GetMigrations(t *testing.T) {
	migrator := &Migrator{}
	migrations := migrator.getMigrations()

	if len(migrations) == 0 {
		t.Error("getMigrations() deveria retornar pelo menos uma migração")
	}

	// Verificar primeira migração
	firstMigration := migrations[0]
	if firstMigration.Version != 1 {
		t.Errorf("primeira migração deveria ter versão 1, obteve %d", firstMigration.Version)
	}

	if firstMigration.Name != "create_deputados_cache" {
		t.Errorf("primeira migração deveria se chamar 'create_deputados_cache', obteve '%s'", firstMigration.Name)
	}

	if firstMigration.SQL == "" {
		t.Error("primeira migração deveria ter SQL não vazio")
	}

	// Verificar se o SQL contém comandos esperados
	expectedCommands := []string{"CREATE TABLE", "deputados_cache", "CREATE INDEX"}
	for _, cmd := range expectedCommands {
		if !containsString(firstMigration.SQL, cmd) {
			t.Errorf("SQL da migração deveria conter '%s'", cmd)
		}
	}
}

// Função auxiliar para verificar se uma string contém outra
func containsString(text, substr string) bool {
	return len(text) >= len(substr) && findInString(text, substr)
}

func findInString(text, substr string) bool {
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestMigrator_IsMigrationApplied(t *testing.T) {
	migrator := &Migrator{}

	applied := map[int]bool{
		1: true,
		2: true,
		5: true,
	}

	tests := []struct {
		version  int
		expected bool
	}{
		{1, true},
		{2, true},
		{3, false},
		{4, false},
		{5, true},
		{6, false},
	}

	for _, tt := range tests {
		result := migrator.isMigrationApplied(applied, tt.version)
		if result != tt.expected {
			t.Errorf("isMigrationApplied(%d) = %v, esperado %v", tt.version, result, tt.expected)
		}
	}
}

func TestMigrator_IsMigrationApplied_EmptyMap(t *testing.T) {
	migrator := &Migrator{}
	applied := make(map[int]bool)

	if migrator.isMigrationApplied(applied, 1) {
		t.Error("migração não deveria estar aplicada quando map está vazio")
	}
}

func TestMigrator_GetMigrations_ContainsExpectedSQL(t *testing.T) {
	migrator := &Migrator{}
	migrations := migrator.getMigrations()

	if len(migrations) == 0 {
		t.Fatal("deveria ter pelo menos uma migração")
	}

	migration := migrations[0]

	// Verificar elementos essenciais do SQL
	requiredElements := []string{
		"CREATE TABLE IF NOT EXISTS deputados_cache",
		"id INT PRIMARY KEY",
		"payload JSONB NOT NULL",
		"updated_at TIMESTAMP",
		"CREATE INDEX IF NOT EXISTS idx_deputados_cache_updated_at",
	}

	for _, element := range requiredElements {
		if !containsString(migration.SQL, element) {
			t.Errorf("SQL deveria conter '%s'", element)
		}
	}
}

func TestMigration_Struct(t *testing.T) {
	migration := Migration{
		Version: 42,
		Name:    "test_migration_name",
		SQL:     "CREATE TABLE example (id INT PRIMARY KEY);",
	}

	if migration.Version != 42 {
		t.Errorf("migration.Version = %d, esperado 42", migration.Version)
	}

	if migration.Name != "test_migration_name" {
		t.Errorf("migration.Name = %q, esperado %q", migration.Name, "test_migration_name")
	}

	expectedSQL := "CREATE TABLE example (id INT PRIMARY KEY);"
	if migration.SQL != expectedSQL {
		t.Errorf("migration.SQL = %q, esperado %q", migration.SQL, expectedSQL)
	}
}

func TestMigrator_GetMigrations_VersionSequence(t *testing.T) {
	migrator := &Migrator{}
	migrations := migrator.getMigrations()

	// Verificar se as versões começam em 1 e são sequenciais
	for i, migration := range migrations {
		expectedVersion := i + 1
		if migration.Version != expectedVersion {
			t.Errorf("migração %d deveria ter versão %d, obteve %d", i, expectedVersion, migration.Version)
		}
	}
}

func TestMigrator_GetMigrations_NamesNotEmpty(t *testing.T) {
	migrator := &Migrator{}
	migrations := migrator.getMigrations()

	for i, migration := range migrations {
		if migration.Name == "" {
			t.Errorf("migração %d não deveria ter nome vazio", i)
		}

		if migration.SQL == "" {
			t.Errorf("migração %d não deveria ter SQL vazio", i)
		}
	}
}
