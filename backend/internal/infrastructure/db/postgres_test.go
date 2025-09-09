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
