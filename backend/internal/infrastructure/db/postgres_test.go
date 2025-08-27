package db

import (
	"context"
	"os"
	"testing"
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
	// Força erro alterando variável para uma URL inválida (porta não numérica) - alterando apenas porta deve ainda parsear, então usamos user com caracteres
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "invalid user\n") // quebra parse
	os.Setenv("DB_PASSWORD", "x")
	os.Setenv("DB_NAME", "foo")
	if _, err := NewPostgresPool(context.Background()); err == nil {
		t.Fatalf("esperava erro de parse config")
	}
}
