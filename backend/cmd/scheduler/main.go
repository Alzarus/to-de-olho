package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"to-de-olho-backend/internal/application"
	"to-de-olho-backend/internal/config"
	"to-de-olho-backend/internal/domain"
	"to-de-olho-backend/internal/infrastructure/cache"
	"to-de-olho-backend/internal/infrastructure/db"
	"to-de-olho-backend/internal/infrastructure/httpclient"
	ingestor "to-de-olho-backend/internal/infrastructure/ingestor"
	"to-de-olho-backend/internal/infrastructure/migrations"
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
	votacaoRepo := repository.NewVotacaoRepository(database)
	schedulerRepo := repository.NewSchedulerRepository(database)

	// Run database migrations if database is available (ensure functions/tables exist)
	if database != nil {
		migrator := migrations.NewMigrator(database)
		if err := migrator.Run(ctx); err != nil {
			log.Printf("Aviso: falha ao executar migrações: %v", err)
		}
	}

	// Services
	deputadosService := application.NewDeputadosService(camaraClient, redisCache, deputadoRepo, despesaRepo)
	proposicoesService := application.NewProposicoesService(camaraClient, redisCache, proposicaoRepo, logger)
	votacoesService := application.NewVotacoesService(votacaoRepo, camaraClient, redisCache)

	// Ingestor / Incremental Sync Manager (usado pelo scheduler para sync diário)
	ingestorManager := ingestor.NewIncrementalSyncManager(deputadosService, proposicoesService, votacoesService, database, redisCache)

	// 🧠 Sistema Inteligente de Controle de Scheduler
	smartSchedulerService := application.NewSmartSchedulerService(
		schedulerRepo,
		ingestorManager,
		deputadosService,
		proposicoesService,
		votacoesService,
		logger,
	)

	// Cron Scheduler com timezone do Brasil
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return fmt.Errorf("erro ao carregar fuso horário: %w", err)
	}
	c := cron.New(cron.WithLocation(loc))

	// Verificar se o scheduler diário está habilitado
	dailyEnabled := os.Getenv("SCHEDULER_ENABLE_DAILY") != "false" // default true
	dailyCron := os.Getenv("SCHEDULER_DAILY_CRON")
	if dailyCron == "" {
		dailyCron = "0 6 * * *" // default às 6:00
	}

	if dailyEnabled {
		log.Printf("🌅 Configurando scheduler diário: %s", dailyCron)
		_, err = c.AddFunc(dailyCron, func() {
			log.Println("🌅 Verificando necessidade de sincronização diária")

			// Timeout configurável via env var
			timeoutEnv := os.Getenv("SCHEDULER_TIMEOUT")
			timeout, err := time.ParseDuration(timeoutEnv)
			if err != nil || timeout == 0 {
				timeout = 60 * time.Minute // default 60 minutos
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			// Configuração inteligente para sync diária
			config := domain.GetDefaultSchedulerConfig(domain.SchedulerTipoDiario)
			// Allow override of parallel workers via env var
			if pw := os.Getenv("SCHEDULER_PARALLEL_WORKERS"); pw != "" {
				if v, err := strconv.Atoi(pw); err == nil && v > 0 {
					config.ParallelWorkers = v
				}
			}
			config.TriggeredBy = "cron-diario"

			execution, err := smartSchedulerService.ExecuteIntelligentScheduler(ctx, config)
			if err != nil {
				log.Printf("ℹ️ Sync diária não executada: %v", err)
			} else {
				log.Printf("✅ Sync diária iniciada - Execution ID: %s", execution.ExecutionID)
			}
		})

		if err != nil {
			return fmt.Errorf("erro ao configurar cron job diário: %w", err)
		}
	} else {
		log.Println("⚠️ Scheduler diário desabilitado via SCHEDULER_ENABLE_DAILY")
	}

	// Sync rápido opcional (desabilitado por padrão para execução única diária)
	quickEnabled := os.Getenv("SCHEDULER_ENABLE_QUICK") == "true" // default false
	if quickEnabled {
		quickCron := os.Getenv("SCHEDULER_QUICK_CRON")
		if quickCron == "" {
			quickCron = "0 8,12,16,20 * * *" // default a cada 4 horas
		}

		log.Printf("🔄 Configurando scheduler rápido: %s", quickCron)
		_, err = c.AddFunc(quickCron, func() {
			log.Println("🔄 Verificando necessidade de sincronização rápida")

			// Timeout menor para sync rápida
			timeout := 20 * time.Minute
			if timeoutEnv := os.Getenv("SCHEDULER_TIMEOUT"); timeoutEnv != "" {
				if parsedTimeout, err := time.ParseDuration(timeoutEnv); err == nil {
					timeout = parsedTimeout / 3 // 1/3 do timeout para sync rápida
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			// Configuração inteligente para sync rápida
			config := domain.GetDefaultSchedulerConfig(domain.SchedulerTipoRapido)
			// Allow override of parallel workers via env var
			if pw := os.Getenv("SCHEDULER_PARALLEL_WORKERS"); pw != "" {
				if v, err := strconv.Atoi(pw); err == nil && v > 0 {
					config.ParallelWorkers = v
				}
			}
			config.TriggeredBy = "cron-rapido"

			execution, err := smartSchedulerService.ExecuteIntelligentScheduler(ctx, config)
			if err != nil {
				log.Printf("ℹ️ Sync rápida não executada: %v", err)
			} else {
				log.Printf("✅ Sync rápida iniciada - Execution ID: %s", execution.ExecutionID)
			}
		})

		if err != nil {
			return fmt.Errorf("erro ao configurar cron job rápido: %w", err)
		}
	}

	// Iniciar scheduler
	c.Start()
	log.Println("⏰ Scheduler iniciado com sucesso")
	log.Println("📅 Próximas execuções:")
	for _, entry := range c.Entries() {
		log.Printf("  → %s", entry.Next.Format("2006-01-02 15:04:05"))
	}

	// Execução inicial controlada: consultar repositório de backfill para saber se há backfill em andamento.
	// Se houver uma execução de backfill com status 'running', pulamos a verificação inicial.
	backfillRepo := repository.NewBackfillRepository(database)
	backfillRunning := false

	// Checar rapidamente por até 15s se uma execução backfill aparece no DB. Isso cobre o caso onde
	// o ingestor cria o registro de execução poucos instantes antes do scheduler subir.
	waitCtx, waitCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer waitCancel()

	tick := time.NewTicker(500 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-waitCtx.Done():
			// Timeout: sair do loop e usar o resultado (possivelmente nenhum backfill em andamento)
		case <-tick.C:
			if exec, err := backfillRepo.GetRunningExecution(context.Background()); err == nil && exec != nil {
				log.Printf("ℹ️ Backfill em andamento detectado (execution_id=%s, tipo=%s). Pulando verificação inicial do scheduler.", exec.ExecutionID, exec.Tipo)
				backfillRunning = true
				// Encontrado — não precisamos continuar polling
				goto afterPolling
			} else {
				// Se a consulta tiver erro diferente de not found, logar e continuar tentando
				if err != nil && err != domain.ErrBackfillNaoEncontrado {
					log.Printf("⚠️ Erro ao verificar execução de backfill em andamento: %v. Continuando polling.", err)
				}
			}
		}
	}
afterPolling:

	// Além do check DB, ainda permitimos override via variável de ambiente para testes rápidos
	if os.Getenv("SCHEDULER_SKIP_STARTUP") == "true" || os.Getenv("BACKFILL_RUNNING") == "true" {
		backfillRunning = true
	}

	if backfillRunning {
		log.Println("ℹ️ Pulando verificação inicial do scheduler devido a backfill em andamento ou configuração de ambiente")
	} else {
		log.Println("🚀 Executando verificação inicial...")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)

		// Configuração para sync inicial
		config := domain.GetDefaultSchedulerConfig(domain.SchedulerTipoInicial)
		// Allow override of parallel workers via env var
		if pw := os.Getenv("SCHEDULER_PARALLEL_WORKERS"); pw != "" {
			if v, err := strconv.Atoi(pw); err == nil && v > 0 {
				config.ParallelWorkers = v
			}
		}
		config.TriggeredBy = "startup"

		execution, err := smartSchedulerService.ExecuteIntelligentScheduler(ctx, config)
		if err != nil {
			log.Printf("ℹ️ Verificação inicial: %v", err)
		} else {
			log.Printf("✅ Verificação inicial iniciada - Execution ID: %s", execution.ExecutionID)
		}
		cancel()
	}

	// Aguardar sinal de interrupção
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Parando scheduler...")
	c.Stop()
	log.Println("👋 Scheduler finalizado")

	return nil
}
