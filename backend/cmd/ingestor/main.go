package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	app "to-de-olho-backend/internal/application"
	"to-de-olho-backend/internal/config"
	"to-de-olho-backend/internal/infrastructure/cache"
	"to-de-olho-backend/internal/infrastructure/db"
	"to-de-olho-backend/internal/infrastructure/httpclient"
	"to-de-olho-backend/internal/infrastructure/ingestor"
	"to-de-olho-backend/internal/infrastructure/migrations"
	"to-de-olho-backend/internal/infrastructure/repository"
)

func main() {
	mode := flag.String("mode", "daily", "Mode: backfill|daily|strategic")
	years := flag.Int("years", 0, "Backfill years from now backwards (0 = use config)")
	startYear := flag.Int("start-year", 0, "Specific start year for backfill (0 = use config)")
	flag.Parse()

	// Setup structured logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Falha ao carregar configuração: %v", err)
	}

	ctx := context.Background()
	pgPool, err := db.NewPostgresPoolFromConfig(ctx, &cfg.Database)
	if err != nil {
		log.Fatalf("Postgres connection error: %v", err)
	}

	// Run database migrations
	migrator := migrations.NewMigrator(pgPool)
	if err := migrator.Run(ctx); err != nil {
		log.Fatalf("Migration error: %v", err)
	}

	// Setup repositories and services
	deputadoRepo := repository.NewDeputadoRepository(pgPool)
	proposicaoRepo := repository.NewProposicaoRepository(pgPool)
	client := httpclient.NewCamaraClientFromConfig(&cfg.CamaraClient)
	cacheClient := cache.NewFromConfig(&cfg.Redis)

	deputadosService := app.NewDeputadosService(client, cacheClient, deputadoRepo)
	proposicoesService := app.NewProposicoesService(client, cacheClient, proposicaoRepo, logger)

	switch *mode {
	case "strategic":
		if err := runStrategicBackfill(ctx, pgPool, deputadosService, proposicoesService, deputadoRepo, proposicaoRepo, cfg, *years, *startYear); err != nil {
			log.Fatalf("strategic backfill failed: %v", err)
		}
	case "backfill":
		if err := runBackfill(ctx, deputadosService, deputadoRepo, *years); err != nil {
			log.Fatalf("backfill failed: %v", err)
		}
	case "daily":
		if err := runDaily(ctx, deputadosService, deputadoRepo); err != nil {
			log.Fatalf("daily ingest failed: %v", err)
		}
	default:
		log.Fatalf("unknown mode: %s", *mode)
	}
}

// runStrategicBackfill executa backfill histórico estratégico com checkpoints
func runStrategicBackfill(
	ctx context.Context,
	pgPool *pgxpool.Pool,
	deputadosService *app.DeputadosService,
	proposicoesService *app.ProposicoesService,
	deputadoRepo *repository.DeputadoRepository,
	proposicaoRepo *repository.ProposicaoRepository,
	cfg *config.Config,
	years int,
	startYear int,
) error {
	log.Println("🚀 Iniciando Backfill Histórico Estratégico")

	// Configurar estratégia baseada nos parâmetros ou configuração
	strategy := ingestor.DefaultBackfillStrategy()

	// Determinar ano inicial e final
	currentYear := time.Now().Year()

	if startYear > 0 {
		// Usar ano específico fornecido via flag
		strategy.YearStart = startYear
		strategy.YearEnd = currentYear
		log.Printf("📅 Usando ano inicial específico: %d", startYear)
	} else if years > 0 {
		// Usar número de anos atrás
		strategy.YearStart = currentYear - years + 1
		strategy.YearEnd = currentYear
		log.Printf("📅 Usando %d anos atrás: %d-%d", years, strategy.YearStart, strategy.YearEnd)
	} else {
		// Usar configuração padrão
		strategy.YearStart = cfg.Ingestor.BackfillStartYear
		strategy.YearEnd = currentYear
		log.Printf("📅 Usando configuração padrão: %d-%d", strategy.YearStart, strategy.YearEnd)
	}

	// Aplicar configurações do ingestor
	strategy.BatchSize = cfg.Ingestor.BatchSize
	strategy.MaxRetries = cfg.Ingestor.MaxRetries

	log.Printf("📊 Estratégia: %d-%d (%d anos), lotes de %d, %d tentativas",
		strategy.YearStart, strategy.YearEnd, strategy.YearEnd-strategy.YearStart+1,
		strategy.BatchSize, strategy.MaxRetries)

	// Criar gerenciador de backfill e executor estratégico
	backfillManager := ingestor.NewBackfillManager(pgPool)
	executor := ingestor.NewStrategicBackfillExecutor(
		backfillManager,
		deputadosService,
		proposicoesService,
		deputadoRepo,
		proposicaoRepo,
		strategy,
	)

	// Executar backfill estratégico
	return executor.ExecuteBackfill(ctx)
}

func runBackfill(ctx context.Context, svc *app.DeputadosService, repo *repository.DeputadoRepository, years int) error {
	fmt.Println("Starting backfill for deputies and recent expenses...")
	deps, _, err := svc.ListarDeputados(ctx, "", "", "")
	if err != nil {
		return err
	}
	fmt.Printf("Fetched %d deputies. Upserting...\n", len(deps))
	if err := repo.UpsertDeputados(ctx, deps); err != nil {
		return err
	}
	// Expenses for last N years (bounded to current year)
	year := time.Now().Year()
	minYear := year - (years - 1)
	if minYear < year-10 {
		minYear = year - 10
	}
	fmt.Printf("Fetching expenses from %d to %d (skipping for brevity placeholder)\n", minYear, year)
	// NOTE: For now, we keep only deputies cached in DB; expenses ingestion can be implemented in repository later.
	return nil
}

func runDaily(ctx context.Context, svc *app.DeputadosService, repo *repository.DeputadoRepository) error {
	fmt.Println("Running daily sync for deputies (and recent expenses placeholder)...")
	deps, _, err := svc.ListarDeputados(ctx, "", "", "")
	if err != nil {
		return err
	}
	if err := repo.UpsertDeputados(ctx, deps); err != nil {
		return err
	}
	fmt.Printf("Daily sync done. Deputies upserted: %d\n", len(deps))
	return nil
}
