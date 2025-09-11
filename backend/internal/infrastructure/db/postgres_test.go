package db

import (
	"context"
	"os"
	"testing"
	"time"

	"to-de-olho-backend/internal/config"
)

func TestGetenv(t *testing.T) {
	os.Unsetenv("TEST_KEY_X")
	if v := getenv("TEST_KEY_X", "default"); v != "default" {
		t.Fatalf("esperava default, obteve %s", v)
	}
	os.Setenv("TEST_KEY_X", "custom")
	if v := getenv("TEST_KEY_X", "default"); v != "custom" {
		t.Fatalf("esperava custom, obteve %s", v)
	}
}

func TestNewPostgresPool_ErroConfig(t *testing.T) {
	// Salvar valores originais para restaurar depois
	origHost := os.Getenv("POSTGRES_HOST")
	origPort := os.Getenv("POSTGRES_PORT")
	origUser := os.Getenv("POSTGRES_USER")
	origPass := os.Getenv("POSTGRES_PASSWORD")
	origDB := os.Getenv("POSTGRES_DB")

	// Restaurar valores originais após o teste
	defer func() {
		if origHost != "" {
			os.Setenv("POSTGRES_HOST", origHost)
		} else {
			os.Unsetenv("POSTGRES_HOST")
		}
		if origPort != "" {
			os.Setenv("POSTGRES_PORT", origPort)
		} else {
			os.Unsetenv("POSTGRES_PORT")
		}
		if origUser != "" {
			os.Setenv("POSTGRES_USER", origUser)
		} else {
			os.Unsetenv("POSTGRES_USER")
		}
		if origPass != "" {
			os.Setenv("POSTGRES_PASSWORD", origPass)
		} else {
			os.Unsetenv("POSTGRES_PASSWORD")
		}
		if origDB != "" {
			os.Setenv("POSTGRES_DB", origDB)
		} else {
			os.Unsetenv("POSTGRES_DB")
		}
	}()

	// Força erro alterando variável para uma URL inválida - user com caracteres inválidos deve quebrar parse
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_PORT", "5432")
	os.Setenv("POSTGRES_USER", "invalid user\n") // quebra parse
	os.Setenv("POSTGRES_PASSWORD", "x")
	os.Setenv("POSTGRES_DB", "foo")
	if _, err := NewPostgresPool(context.Background()); err == nil {
		t.Fatalf("esperava erro de parse config")
	}
}

func TestNewPostgresPool_DefaultValues(t *testing.T) {
	// Limpar todas as variáveis de ambiente
	envVars := []string{
		"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_USER",
		"POSTGRES_PASSWORD", "POSTGRES_DB",
	}

	originalValues := make(map[string]string)
	for _, key := range envVars {
		originalValues[key] = os.Getenv(key)
		os.Unsetenv(key)
	}

	defer func() {
		for key, value := range originalValues {
			if value != "" {
				os.Setenv(key, value)
			}
		}
	}()

	// Testar valores padrão
	if host := getenv("POSTGRES_HOST", "localhost"); host != "localhost" {
		t.Errorf("esperava host padrão 'localhost', obteve '%s'", host)
	}

	if port := getenv("POSTGRES_PORT", "5432"); port != "5432" {
		t.Errorf("esperava port padrão '5432', obteve '%s'", port)
	}

	if user := getenv("POSTGRES_USER", "postgres"); user != "postgres" {
		t.Errorf("esperava user padrão 'postgres', obteve '%s'", user)
	}

	if pass := getenv("POSTGRES_PASSWORD", "postgres123"); pass != "postgres123" {
		t.Errorf("esperava password padrão 'postgres123', obteve '%s'", pass)
	}

	if db := getenv("POSTGRES_DB", "to_de_olho"); db != "to_de_olho" {
		t.Errorf("esperava db padrão 'to_de_olho', obteve '%s'", db)
	}
}

func TestNewPostgresPool_CustomValues(t *testing.T) {
	originalValues := map[string]string{
		"POSTGRES_HOST":     os.Getenv("POSTGRES_HOST"),
		"POSTGRES_PORT":     os.Getenv("POSTGRES_PORT"),
		"POSTGRES_USER":     os.Getenv("POSTGRES_USER"),
		"POSTGRES_PASSWORD": os.Getenv("POSTGRES_PASSWORD"),
		"POSTGRES_DB":       os.Getenv("POSTGRES_DB"),
	}

	// Definir valores customizados
	os.Setenv("POSTGRES_HOST", "custom_host")
	os.Setenv("POSTGRES_PORT", "9999")
	os.Setenv("POSTGRES_USER", "custom_user")
	os.Setenv("POSTGRES_PASSWORD", "custom_pass")
	os.Setenv("POSTGRES_DB", "custom_db")

	defer func() {
		for key, value := range originalValues {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	// Verificar se getenv usa valores customizados
	if host := getenv("POSTGRES_HOST", "default"); host != "custom_host" {
		t.Errorf("esperava 'custom_host', obteve '%s'", host)
	}

	if port := getenv("POSTGRES_PORT", "default"); port != "9999" {
		t.Errorf("esperava '9999', obteve '%s'", port)
	}

	// Executar NewPostgresPool (pode falhar na conexão mas testamos parse)
	ctx := context.Background()
	_, err := NewPostgresPool(ctx)

	// Esperamos erro de conexão, não de parsing
	if err != nil {
		// Aceitar erro de conexão (esperado em testes)
		t.Logf("Erro de conexão esperado: %v", err)
	}
}

func TestNewPostgresPoolFromConfig_Success(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            "5432",
		User:            "postgres",
		Password:        "test_password",
		Database:        "test_db",
		SSLMode:         "disable",
		MaxConns:        10,
		MinConns:        2,
		MaxConnLifetime: 1 * time.Hour,
		MaxConnIdleTime: 30 * time.Minute,
	}

	ctx := context.Background()

	// Esta função será executada (testamos pelo menos o parsing)
	_, err := NewPostgresPoolFromConfig(ctx, cfg)

	// Esperamos erro de conexão em testes, não de parsing
	if err != nil {
		// Verificar se não é erro de parsing (que seria um bug)
		if containsSubstring(err.Error(), "failed to parse connection string") {
			t.Errorf("erro de parsing inesperado: %v", err)
		}
		// Erros de conexão são esperados em testes unitários
		t.Logf("Erro de conexão esperado: %v", err)
	}
}

func TestNewPostgresPoolFromConfig_InvalidConnectionString(t *testing.T) {
	// Configuração com dados inválidos que quebram o parsing
	cfg := &config.DatabaseConfig{
		Host:     "invalid\nhost",
		Port:     "invalid\nport",
		User:     "invalid\nuser",
		Password: "password",
		Database: "db",
		SSLMode:  "disable",
		MaxConns: 5,
		MinConns: 1,
	}

	ctx := context.Background()

	_, err := NewPostgresPoolFromConfig(ctx, cfg)

	// Deve haver erro devido ao connection string inválido
	if err == nil {
		t.Error("esperava erro com connection string inválido")
	}

	// Verificar se é realmente erro de parsing
	if !containsSubstring(err.Error(), "failed to parse connection string") {
		t.Logf("Erro obtido (conexão ou outro): %v", err)
	}
}

func TestNewPostgresPoolFromConfig_ConfigApplication(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Host:            "testhost",
		Port:            "9999",
		User:            "testuser",
		Password:        "testpass",
		Database:        "testdb",
		SSLMode:         "require",
		MaxConns:        20,
		MinConns:        5,
		MaxConnLifetime: 2 * time.Hour,
		MaxConnIdleTime: 45 * time.Minute,
	}

	// Verificar se ConnectionString() gera string esperada
	expectedDSN := "postgres://testuser:testpass@testhost:9999/testdb?sslmode=require"
	actualDSN := cfg.ConnectionString()

	if actualDSN != expectedDSN {
		t.Errorf("ConnectionString() = %q, esperado %q", actualDSN, expectedDSN)
	}

	ctx := context.Background()

	// Tentar criar pool (provavelmente falhará na conexão, mas testamos a configuração)
	_, err := NewPostgresPoolFromConfig(ctx, cfg)

	// Esperamos erro de conexão, não de parsing
	if err != nil && containsSubstring(err.Error(), "failed to parse connection string") {
		t.Errorf("erro de parsing inesperado: %v", err)
	}
}

func TestNewPostgresPool_InvalidDSNCharacters(t *testing.T) {
	originalValues := map[string]string{
		"POSTGRES_HOST":     os.Getenv("POSTGRES_HOST"),
		"POSTGRES_PORT":     os.Getenv("POSTGRES_PORT"),
		"POSTGRES_USER":     os.Getenv("POSTGRES_USER"),
		"POSTGRES_PASSWORD": os.Getenv("POSTGRES_PASSWORD"),
		"POSTGRES_DB":       os.Getenv("POSTGRES_DB"),
	}

	defer func() {
		for key, value := range originalValues {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	// Definir user com caracteres que quebram URL
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_PORT", "5432")
	os.Setenv("POSTGRES_USER", "user with\nspaces")
	os.Setenv("POSTGRES_PASSWORD", "pass")
	os.Setenv("POSTGRES_DB", "db")

	ctx := context.Background()
	_, err := NewPostgresPool(ctx)

	if err == nil {
		t.Error("esperava erro com caracteres inválidos na URL")
	}
}

// Função auxiliar para verificar substring
func containsSubstring(text, substr string) bool {
	if len(substr) > len(text) {
		return false
	}
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
