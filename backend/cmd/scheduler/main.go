package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"to-de-olho-backend/internal/application"
	"to-de-olho-backend/internal/config"
	"to-de-olho-backend/internal/infrastructure/cache"
	"to-de-olho-backend/internal/infrastructure/db"
	"to-de-olho-backend/internal/infrastructure/httpclient"
	"to-de-olho-backend/internal/infrastructure/ingestor"
	"to-de-olho-backend/internal/infrastructure/repository"

	"github.com/robfig/cron/v3"
)

func main() {
	log.Println("🕐 Iniciando Scheduler de Sincronização Diária")

	if err := run(); err != nil {
		log.Printf("❌ Erro fatal no scheduler: %v", err)
		os.Exit(1)
	}
}

func run() error {
	// Configuração
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("erro ao carregar configuração: %w", err)
	}

	// Database
	ctx := context.Background()
	database, err := db.NewPostgresPoolFromConfig(ctx, &cfg.Database)
	if err != nil {
		return fmt.Errorf("erro ao conectar ao banco: %w", err)
	}
	defer database.Close()

	// Redis Cache
	redisCache := cache.NewFromConfig(&cfg.Redis)

	// Logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// HTTP Client
	camaraClient := httpclient.NewCamaraClient(
		cfg.CamaraClient.BaseURL,
		cfg.CamaraClient.Timeout,
		cfg.CamaraClient.RPS,
		cfg.CamaraClient.Burst,
	)

	// Repositories
	deputadoRepo := repository.NewDeputadoRepository(database)
	despesaRepo := repository.NewDespesaRepository(database)
	proposicaoRepo := repository.NewProposicaoRepository(database)

	// Services
	deputadosService := application.NewDeputadosService(camaraClient, redisCache, deputadoRepo, despesaRepo)
	proposicoesService := application.NewProposicoesService(camaraClient, redisCache, proposicaoRepo, logger)
	analyticsService := application.NewAnalyticsService(deputadoRepo, proposicaoRepo, redisCache, logger)

	// Sync Manager
	syncManager := ingestor.NewIncrementalSyncManager(
		deputadosService,
		proposicoesService,
		analyticsService,
		database,
		redisCache,
	)

	// Cron Scheduler com timezone do Brasil
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return fmt.Errorf("erro ao carregar fuso horário: %w", err)
	}
	c := cron.New(cron.WithLocation(loc))

	// Sync diário às 6h da manhã
	_, err = c.AddFunc("0 6 * * *", func() {
		log.Println("🌅 Iniciando sincronização incremental diária às 6h")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		if err := syncManager.ExecuteDailySync(ctx); err != nil {
			log.Printf("❌ Erro na sincronização diária: %v", err)
		} else {
			log.Println("✅ Sincronização diária concluída com sucesso")
		}
	})

	if err != nil {
		return fmt.Errorf("erro ao configurar cron job: %w", err)
	}

	// Sync a cada 4 horas durante horário comercial (útil para testes)
	_, err = c.AddFunc("0 8,12,16,20 * * *", func() {
		log.Println("🔄 Sincronização incremental de 4h")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		if err := syncManager.ExecuteQuickSync(ctx); err != nil {
			log.Printf("⚠️  Erro na sincronização rápida: %v", err)
		} else {
			log.Println("✅ Sincronização rápida concluída")
		}
	})

	if err != nil {
		return fmt.Errorf("erro ao configurar cron job 4h: %w", err)
	}

	// Iniciar scheduler
	c.Start()
	log.Println("⏰ Scheduler iniciado com sucesso")
	log.Println("📅 Próximas execuções:")
	for _, entry := range c.Entries() {
		log.Printf("  → %s", entry.Next.Format("2006-01-02 15:04:05"))
	}

	// Execução inicial imediata para teste
	log.Println("🚀 Executando sincronização inicial...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	if err := syncManager.ExecuteQuickSync(ctx); err != nil {
		log.Printf("⚠️  Erro na sincronização inicial: %v", err)
	} else {
		log.Println("✅ Sincronização inicial concluída")
	}
	cancel()

	// Aguardar sinal de interrupção
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Parando scheduler...")
	c.Stop()
	log.Println("👋 Scheduler finalizado")

	return nil
}
