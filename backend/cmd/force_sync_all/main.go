package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/Alzarus/to-de-olho/internal/ceaps"
	"github.com/Alzarus/to-de-olho/internal/comissao"
	"github.com/Alzarus/to-de-olho/internal/emenda"
	"github.com/Alzarus/to-de-olho/internal/proposicao"
	"github.com/Alzarus/to-de-olho/internal/ranking"
	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/internal/votacao"
	"github.com/Alzarus/to-de-olho/pkg/senado"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// FORCE BACKFILL SCRIPT
// This script iterates from 2023 to current year and forces a sync of:
// - Votacao Metadata (Ementas, Dates)
// - CEAPS (Expenses)
// - Emendas (Amendments - requires TRANSPARENCIA_API_KEY)
// It also syncs Senators, Votes (List), Commissions, and Propositions.
//Finally, it recalculates the ranking.

func main() {
	// 1. Load Env
	_ = godotenv.Load() // Try loading .env (ignore error if missing)
	
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	
	slog.Info("=== INICIANDO FORCE BACKFILL (2023-CURRENT) ===")

	// 2. Connect DB
	db, err := connectDB()
	if err != nil {
		slog.Error("CRITICAL: falha ao conectar no banco", "error", err)
		os.Exit(1)
	}

	// 3. Connect Redis (Optional)
	redisClient := connectRedis()

	// 4. Init Services
	transparenciaKey := os.Getenv("TRANSPARENCIA_API_KEY")
	if transparenciaKey == "" {
		slog.Warn("!!! ATENCAO: TRANSPARENCIA_API_KEY NAO ENCONTRADA. SYNC DE EMENDAS FALHARA !!!")
	} else {
		slog.Info("TRANSPARENCIA_API_KEY encontrada", "len", len(transparenciaKey))
	}

	// Repos
	senadorRepo := senador.NewRepository(db)
	votacaoRepo := votacao.NewRepository(db)
	ceapsRepo := ceaps.NewRepository(db)
	emendaRepo := emenda.NewRepository(db)
	comissaoRepo := comissao.NewRepository(db)
	proposicaoRepo := proposicao.NewRepository(db)

	// Clients
	legisClient := senado.NewLegisClient()
	admClient := senado.NewAdmClient()

	// Sync Services
	senadorSync := senador.NewSyncService(senadorRepo, legisClient)
	votacaoSync := votacao.NewSyncService(votacaoRepo, senadorRepo, legisClient)
	ceapsSync := ceaps.NewSyncService(ceapsRepo, senadorRepo, admClient)
	emendaSync := emenda.NewSyncService(emendaRepo, senadorRepo, transparenciaKey)
	comissaoSync := comissao.NewSyncService(comissaoRepo, senadorRepo, legisClient)
	proposicaoSync := proposicao.NewSyncService(proposicaoRepo, senadorRepo, legisClient)
	
	rankingService := ranking.NewService(
		senadorRepo, proposicaoRepo, votacaoRepo, ceapsRepo, comissaoRepo, redisClient,
	)

	ctx := context.Background()
	startAno := 2023
	endAno := time.Now().Year()

	// --- EXECUTION ---

	// A. Senadores
	slog.Info(">>> 1. SENADORES")
	if err := senadorSync.SyncFromAPI(ctx); err != nil {
		slog.Error("falha sync senadores", "error", err)
	}

	// B. Votacoes List (All Years)
	slog.Info(">>> 2. VOTACOES (LISTA GERAL)")
	if err := votacaoSync.SyncFromAPI(ctx); err != nil {
		slog.Error("falha sync votacoes list", "error", err)
	}

	// C. Loop 2023..2026
	for ano := startAno; ano <= endAno; ano++ {
		slog.Info(">>> PROCESSANDO ANO", "ano", ano)
		
		// Metadata
		slog.Info("   > Metadata Votacoes")
		if err := votacaoSync.SyncMetadata(ctx, ano); err != nil {
			slog.Error("falha metadata", "ano", ano, "error", err)
		}

		// CEAPS
		slog.Info("   > CEAPS")
		if err := ceapsSync.SyncFromAPI(ctx, ano); err != nil {
			slog.Error("falha ceaps", "ano", ano, "error", err)
		}

		// Emendas
		slog.Info("   > Emendas (Portal Transparencia)")
		if err := emendaSync.SyncAll(ctx, ano); err != nil {
			slog.Error("falha emendas", "ano", ano, "error", err)
		}
	}

	// D. Comissoes & Proposicoes
	slog.Info(">>> 4. COMISSOES")
	if err := comissaoSync.SyncFromAPI(ctx); err != nil {
		slog.Error("falha comissoes", "error", err)
	}

	slog.Info(">>> 5. PROPOSICOES")
	if err := proposicaoSync.SyncFromAPI(ctx); err != nil {
		slog.Error("falha proposicoes", "error", err)
	}

	// E. Ranking
	slog.Info(">>> 6. RECALCULANDO RANKING")
	if _, err := rankingService.CalcularRanking(ctx, nil); err != nil {
		slog.Error("falha calculo ranking", "error", err)
	}

	slog.Info("=== FORCE BACKFILL CONCLUIDO ===")
}

// Helpers (Copied from main.go)
func connectDB() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=todeolho port=5432 sslmode=disable"
	}
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func connectRedis() *redis.Client {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		return nil // Sem redis localmente geralmente
	}
	return redis.NewClient(&redis.Options{Addr: redisAddr})
}

