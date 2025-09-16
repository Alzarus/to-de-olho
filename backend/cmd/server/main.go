package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	app "to-de-olho-backend/internal/application"
	"to-de-olho-backend/internal/config"
	"to-de-olho-backend/internal/infrastructure/cache"
	"to-de-olho-backend/internal/infrastructure/db"
	"to-de-olho-backend/internal/infrastructure/httpclient"
	"to-de-olho-backend/internal/infrastructure/migrations"
	"to-de-olho-backend/internal/infrastructure/repository"
	httpif "to-de-olho-backend/internal/interfaces/http"
	"to-de-olho-backend/internal/interfaces/http/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: arquivo .env não encontrado")
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Falha ao carregar configuração: %v", err)
	}

	ctx := context.Background()

	pgPool, err := db.NewPostgresPoolFromConfig(ctx, &cfg.Database)
	if err != nil {
		log.Printf("Aviso: não foi possível conectar ao Postgres: %v", err)
	}

	// Run database migrations if database is available
	if pgPool != nil {
		migrator := migrations.NewMigrator(pgPool)
		if err := migrator.Run(ctx); err != nil {
			log.Printf("Aviso: falha ao executar migrações: %v", err)
		}
	}

	cacheClient := cache.NewFromConfig(&cfg.Redis)
	repo := repository.NewDeputadoRepository(pgPool)
	proposicoesRepo := repository.NewProposicaoRepository(pgPool)
	client := httpclient.NewCamaraClientFromConfig(&cfg.CamaraClient)

	logger := slog.Default()
	svc := app.NewDeputadosService(client, cacheClient, repo)
	proposicoesSvc := app.NewProposicoesService(client, cacheClient, proposicoesRepo, logger)
	analyticsSvc := app.NewAnalyticsService(repo, proposicoesRepo, cacheClient, logger)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.RateLimitPerIP(cfg.Server.RateLimit, time.Minute))

	api := r.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "OK",
				"message":   "API Tô De Olho funcionando!",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"version":   "1.0.0",
			})
		})

		// Rotas para deputados
		api.GET("/deputados", httpif.GetDeputadosHandler(svc))
		api.GET("/deputados/:id", httpif.GetDeputadoByIDHandler(svc))
		api.GET("/deputados/:id/despesas", httpif.GetDespesasDeputadoHandler(svc))

		// Rotas para proposições
		api.GET("/proposicoes", httpif.GetProposicoesHandler(proposicoesSvc))
		api.GET("/proposicoes/:id", httpif.GetProposicaoPorIDHandler(proposicoesSvc))

		// Rotas para analytics
		analytics := api.Group("/analytics")
		{
			analytics.GET("/rankings/gastos", httpif.GetRankingGastosHandler(analyticsSvc))
			analytics.GET("/rankings/proposicoes", httpif.GetRankingProposicoesHandler(analyticsSvc))
			analytics.GET("/rankings/presenca", httpif.GetRankingPresencaHandler(analyticsSvc))
			analytics.GET("/insights", httpif.GetInsightsGeraisHandler(analyticsSvc))
			analytics.POST("/rankings/atualizar", httpif.PostAtualizarRankingsHandler(analyticsSvc))
		}
	}

	log.Printf("🚀 Servidor rodando na porta %s", cfg.Server.Port)
	log.Printf("📊 API disponível em: http://localhost:%s/api/v1", cfg.Server.Port)
	_ = r.Run(":" + cfg.Server.Port)
}
