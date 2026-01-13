package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pedroalmeida/to-de-olho/internal/api"
	"github.com/pedroalmeida/to-de-olho/internal/ceaps"
	"github.com/pedroalmeida/to-de-olho/internal/senador"
	"github.com/pedroalmeida/to-de-olho/internal/votacao"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Configurar logger estruturado (JSON para Cloud Run)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Conectar ao banco de dados
	db, err := connectDB()
	if err != nil {
		slog.Error("falha ao conectar ao banco", "error", err)
		os.Exit(1)
	}

	// Auto-migrate das entidades
	if err := db.AutoMigrate(
		&senador.Senador{},
		&senador.Mandato{},
		&ceaps.DespesaCEAPS{},
		&votacao.Votacao{},
	); err != nil {
		slog.Error("falha no auto-migrate", "error", err)
		os.Exit(1)
	}

	// Configurar router
	router := api.SetupRouter(db)

	// Criar servidor HTTP
	srv := &http.Server{
		Addr:    getPort(),
		Handler: router,
	}

	// Iniciar servidor em goroutine
	go func() {
		slog.Info("servidor iniciando", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("falha no servidor", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	slog.Info("encerrando servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown forcado", "error", err)
	}

	slog.Info("servidor encerrado")
}

func connectDB() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Default para desenvolvimento local
		dsn = "host=localhost user=postgres password=postgres dbname=todeolho port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true, // Cache de prepared statements
	})
	if err != nil {
		return nil, err
	}

	// Configurar connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return ":" + port
}
