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

	// Verificar se as migrações estão ordenadas por versão
	for i := 1; i < len(migrations); i++ {
		if migrations[i].Version <= migrations[i-1].Version {
			t.Errorf("migrações deveriam estar ordenadas por versão, %d não é maior que %d", migrations[i].Version, migrations[i-1].Version)
		}
	}

	// Verificar se todas as migrações têm nome
	for i, migration := range migrations {
		if migration.Name == "" {
			t.Errorf("migração %d deveria ter nome não vazio", i)
		}
		if migration.SQL == "" {
			t.Errorf("migração %d (%s) deveria ter SQL não vazio", i, migration.Name)
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

// Testes adicionais para melhorar cobertura
func TestMigration_Structure(t *testing.T) {
	migration := Migration{
		Version: 1,
		Name:    "test_migration",
		SQL:     "CREATE TABLE test (id INTEGER);",
	}

	if migration.Version != 1 {
		t.Errorf("Version = %v, want 1", migration.Version)
	}

	if migration.Name != "test_migration" {
		t.Errorf("Name = %v, want test_migration", migration.Name)
	}

	if migration.SQL != "CREATE TABLE test (id INTEGER);" {
		t.Errorf("SQL = %v, want CREATE TABLE test (id INTEGER);", migration.SQL)
	}
}

func TestMigrator_IsMigrationApplied_Extended(t *testing.T) {
	migrator := &Migrator{}

	appliedMigrations := map[int]bool{
		1: true,
		3: true,
		5: true,
	}

	tests := []struct {
		version  int
		expected bool
	}{
		{1, true},
		{2, false},
		{3, true},
		{4, false},
		{5, true},
		{6, false},
	}

	for _, tt := range tests {
		result := migrator.isMigrationApplied(appliedMigrations, tt.version)
		if result != tt.expected {
			t.Errorf("isMigrationApplied(%d) = %v, want %v", tt.version, result, tt.expected)
		}
	}
}

func TestMigrator_IsMigrationApplied_EmptyMapV2(t *testing.T) {
	migrator := &Migrator{}

	appliedMigrations := map[int]bool{}
	result := migrator.isMigrationApplied(appliedMigrations, 1)

	if result != false {
		t.Errorf("isMigrationApplied with empty map should return false, got %v", result)
	}
}

func TestMigrator_GetMigrations_AllVersionsUnique(t *testing.T) {
	migrator := &Migrator{}
	migrations := migrator.getMigrations()

	versionMap := make(map[int]bool)
	for _, migration := range migrations {
		if versionMap[migration.Version] {
			t.Errorf("versão %d aparece mais de uma vez", migration.Version)
		}
		versionMap[migration.Version] = true
	}
}

func TestMigrator_GetMigrations_AllNamesUnique(t *testing.T) {
	migrator := &Migrator{}
	migrations := migrator.getMigrations()

	nameMap := make(map[string]bool)
	for _, migration := range migrations {
		if nameMap[migration.Name] {
			t.Errorf("nome '%s' aparece mais de uma vez", migration.Name)
		}
		nameMap[migration.Name] = true
	}
}

func TestMigrator_GetMigrations_SQLValidation(t *testing.T) {
	migrator := &Migrator{}
	migrations := migrator.getMigrations()

	for _, migration := range migrations {
		// Verificar se o SQL contém comandos SQL válidos
		if !containsString(migration.SQL, "CREATE") &&
			!containsString(migration.SQL, "ALTER") &&
			!containsString(migration.SQL, "INSERT") {
			t.Errorf("migração %s deveria conter comandos SQL válidos", migration.Name)
		}

		// Verificar se não contém comandos perigosos
		dangerousCommands := []string{"DROP DATABASE", "TRUNCATE", "DELETE FROM"}
		for _, cmd := range dangerousCommands {
			if containsString(migration.SQL, cmd) {
				t.Errorf("migração %s não deveria conter comando perigoso: %s", migration.Name, cmd)
			}
		}
	}
}

func TestMigrator_Fields(t *testing.T) {
	pool := &pgxpool.Pool{}
	migrator := NewMigrator(pool)

	// Verificar se todos os campos estão definidos corretamente
	if migrator.db == nil {
		t.Error("db field should not be nil")
	}

	if migrator.db != pool {
		t.Error("db field should point to the provided pool")
	}
}

func TestMigrator_NilPool(t *testing.T) {
	migrator := NewMigrator(nil)

	if migrator == nil {
		t.Error("NewMigrator should not return nil even with nil pool")
	}

	if migrator.db != nil {
		t.Error("db field should be nil when nil pool is provided")
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
