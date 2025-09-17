package repository

import (
	"log/slog"
	"os"
	"testing"
)

// setupTestLogger configura um logger silencioso para testes
func setupTestLogger() {
	// Durante os testes, usar um logger silencioso para evitar poluir o output
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError, // Só mostrar erros críticos durante testes
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// TestMain permite configuração global para todos os testes do package
func TestMain(m *testing.M) {
	setupTestLogger()
	code := m.Run()
	os.Exit(code)
}
