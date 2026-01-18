package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alzarus/to-de-olho/internal/api"
	"github.com/Alzarus/to-de-olho/internal/ceaps"
	"github.com/Alzarus/to-de-olho/internal/comissao"
	"github.com/Alzarus/to-de-olho/internal/emenda"
	"github.com/Alzarus/to-de-olho/internal/proposicao"
	"github.com/Alzarus/to-de-olho/internal/scheduler"
	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/internal/votacao"
	"github.com/redis/go-redis/v9"
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
		&comissao.ComissaoMembro{},
		&proposicao.Proposicao{},
		&emenda.Emenda{},
	); err != nil {
		slog.Error("falha no auto-migrate", "error", err)
		os.Exit(1)
	}

	// Conectar ao Redis
	redisClient := connectRedis()
	if redisClient != nil {
		slog.Info("conectado ao redis")
	} else {
		slog.Warn("rodando sem cache redis")
	}

	// Configurar router
	router := api.SetupRouter(db, redisClient)

	// Criar servidor HTTP
	srv := &http.Server{
		Addr:    getPort(),
		Handler: router,
	}

	// --- Inicializar Services para Scheduler (Duplicado do Router por enquanto) ---
	// Repositorios
	senadorRepo := senador.NewRepository(db)
	votacaoRepo := votacao.NewRepository(db)
	ceapsRepo := ceaps.NewRepository(db)
	emendaRepo := emenda.NewRepository(db)

	// Clients
	legisClient := senado.NewLegisClient()
	admClient := senado.NewAdmClient()
	transparenciaKey := os.Getenv("TRANSPARENCIA_API_KEY")

	// Sync Services
	senadorSync := senador.NewSyncService(senadorRepo, legisClient)
	votacaoSync := votacao.NewSyncService(votacaoRepo, senadorRepo, legisClient)
	ceapsSync := ceaps.NewSyncService(ceapsRepo, senadorRepo, admClient)
	emendaSync := emenda.NewSyncService(emendaRepo, senadorRepo, transparenciaKey)

	// Iniciar Scheduler
	sched := scheduler.NewScheduler(senadorSync, votacaoSync, ceapsSync, emendaSync)
	
	// Contexto para o scheduler (cancelado no shutdown)
	ctxSched, cancelSched := context.WithCancel(context.Background())
	defer cancelSched()

	sched.Start(ctxSched)
	// -----------------------------------------------------------------------------

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

func connectRedis() *redis.Client {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		slog.Warn("redis indisponivel", "error", err)
		return nil
	}

	return client
}
