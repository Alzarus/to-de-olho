package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	app "to-de-olho-backend/internal/application"
	"to-de-olho-backend/internal/config"
	"to-de-olho-backend/internal/infrastructure/cache"
	"to-de-olho-backend/internal/infrastructure/db"
	"to-de-olho-backend/internal/infrastructure/httpclient"
	"to-de-olho-backend/internal/infrastructure/migrations"
	"to-de-olho-backend/internal/infrastructure/repository"
)

func main() {
	mode := flag.String("mode", "backfill", "Mode: backfill|daily")
	years := flag.Int("years", 5, "Backfill years from now backwards")
	flag.Parse()

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

	repo := repository.NewDeputadoRepository(pgPool)
	client := httpclient.NewCamaraClientFromConfig(&cfg.CamaraClient)
	cacheClient := cache.NewFromConfig(&cfg.Redis)
	svc := app.NewDeputadosService(client, cacheClient, repo)

	switch *mode {
	case "backfill":
		if err := runBackfill(ctx, svc, repo, *years); err != nil {
			log.Fatalf("backfill failed: %v", err)
		}
	case "daily":
		if err := runDaily(ctx, svc, repo); err != nil {
			log.Fatalf("daily ingest failed: %v", err)
		}
	default:
		log.Fatalf("unknown mode: %s", *mode)
	}
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
