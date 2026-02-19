package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
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
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter configura todas as rotas da API
func SetupRouter(db *gorm.DB, transparenciaAPIKey string) *gin.Engine {
	router := gin.Default()

	// Middleware CORS
	router.Use(corsMiddleware())

	// Health check
	router.GET("/health", healthHandler(db))

	// Clients das APIs externas
	legisClient := senado.NewLegisClient()
	admClient := senado.NewAdmClient()

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Senadores
		senadorRepo := senador.NewRepository(db)
		senadorHandler := senador.NewHandler(senadorRepo)
		senadorSync := senador.NewSyncService(senadorRepo, legisClient)

		// Despesas CEAPS
		ceapsRepo := ceaps.NewRepository(db)
		ceapsHandler := ceaps.NewHandler(ceapsRepo)
		ceapsSync := ceaps.NewSyncService(ceapsRepo, senadorRepo, admClient)

		// Votacoes
		votacaoRepo := votacao.NewRepository(db)
		votacaoHandler := votacao.NewHandler(votacaoRepo)
		votacaoSync := votacao.NewSyncService(votacaoRepo, senadorRepo, legisClient)

		// Comissoes
		comissaoRepo := comissao.NewRepository(db)
		comissaoHandler := comissao.NewHandler(comissaoRepo)
		comissaoSync := comissao.NewSyncService(comissaoRepo, senadorRepo, legisClient)

		// Proposicoes
		proposicaoRepo := proposicao.NewRepository(db)
		proposicaoHandler := proposicao.NewHandler(proposicaoRepo)

		proposicaoSync := proposicao.NewSyncService(proposicaoRepo, senadorRepo, legisClient)

		// Emendas (RF08-RF10)
		emendaRepo := emenda.NewRepository(db)
		emendaService := emenda.NewService(emendaRepo, senadorRepo)
		emendaHandler := emenda.NewHandler(emendaService)
		emendaSync := emenda.NewSyncService(emendaRepo, senadorRepo, transparenciaAPIKey)

		// Ranking
		rankingService := ranking.NewService(senadorRepo, proposicaoRepo, votacaoRepo, ceapsRepo, comissaoRepo)
		rankingHandler := ranking.NewHandler(rankingService)

		senadores := v1.Group("/senadores")
		{
			senadores.GET("", senadorHandler.ListAll)
			senadores.GET("/:id", senadorHandler.GetByID)
			senadores.GET("/codigo/:codigo", senadorHandler.GetByCodigo)
			senadores.GET("/:id/despesas", ceapsHandler.ListBySenador)
			senadores.GET("/:id/despesas/agregado", ceapsHandler.AggregateBySenador)
			senadores.GET("/:id/votacoes", votacaoHandler.ListBySenador)
			senadores.GET("/:id/votacoes/stats", votacaoHandler.GetStats)
			senadores.GET("/:id/votacoes/tipos", votacaoHandler.GetVotosPorTipo)
			// Comissoes
			senadores.GET("/:id/comissoes", comissaoHandler.ListBySenador)
			senadores.GET("/:id/comissoes/ativas", comissaoHandler.GetAtivas)
			senadores.GET("/:id/comissoes/stats", comissaoHandler.GetStats)
			senadores.GET("/:id/comissoes/casas", comissaoHandler.GetPorCasa)
			// Proposicoes
			senadores.GET("/:id/proposicoes", proposicaoHandler.ListBySenador)
			senadores.GET("/:id/proposicoes/stats", proposicaoHandler.GetStats)
			senadores.GET("/:id/proposicoes/tipos", proposicaoHandler.GetPorTipo)
			// Score individual

			senadores.GET("/:id/score", rankingHandler.GetScoreSenador)
			// Emendas
			senadores.GET("/:id/emendas", emendaHandler.GetBySenador)
		}

		// Votacoes (Geral)
		votacoes := v1.Group("/votacoes")
		{
			votacoes.GET("", votacaoHandler.GetAll)
			votacoes.GET("/:id", votacaoHandler.GetByID)
		}

		// Sync (trigger manual para desenvolvimento)
		v1.POST("/sync/senadores", func(c *gin.Context) {
			if err := senadorSync.SyncFromAPI(c.Request.Context()); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			count, _ := senadorRepo.Count()
			c.JSON(http.StatusOK, gin.H{
				"message": "sync concluido",
				"total":   count,
			})
		})

		v1.POST("/sync/despesas/:ano", func(c *gin.Context) {
			anoStr := c.Param("ano")
			ano := 2024 // default
			if _, err := fmt.Sscanf(anoStr, "%d", &ano); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "ano invalido"})
				return
			}
			if err := ceapsSync.SyncFromAPI(c.Request.Context(), ano); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "sync de despesas concluido",
				"ano":     ano,
			})
		})

		v1.POST("/sync/votacoes", func(c *gin.Context) {
			if err := votacaoSync.SyncFromAPI(c.Request.Context()); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "sync de votacoes concluido",
			})
		})

		v1.POST("/sync/comissoes", func(c *gin.Context) {
			if err := comissaoSync.SyncFromAPI(c.Request.Context()); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "sync de comissoes concluido",
			})
		})

		v1.POST("/sync/proposicoes", func(c *gin.Context) {
			if err := proposicaoSync.SyncFromAPI(c.Request.Context()); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "sync de proposicoes concluido",
			})
		})


        
		v1.POST("/sync/emendas/:ano", func(c *gin.Context) {
			anoStr := c.Param("ano")
			ano := 2024
			if _, err := fmt.Sscanf(anoStr, "%d", &ano); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "ano invalido"})
				return
			}
            
			if err := emendaSync.SyncAll(c.Request.Context(), ano); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "sync de emendas concluido",
				"ano":     ano,
			})
		})

		// Metadata
		v1.GET("/metadata/last-sync", func(c *gin.Context) {
			
			var lastUpdate time.Time
			// Estrategia: Maior timestamp entre updated_at de senadores e data de votacoes
			// Usando UNION ALL para pegar o maior de todos
			query := `
				SELECT MAX(ts) FROM (
					SELECT MAX(updated_at) as ts FROM senadores
					UNION ALL
					SELECT MAX(created_at) as ts FROM votacoes
				) as updates
			`
			if err := db.Raw(query).Scan(&lastUpdate).Error; err != nil {
				lastUpdate = time.Now()
			}
			
			response := gin.H{"last_sync": lastUpdate}

			c.Header("X-Cache", "MISS")
			c.JSON(http.StatusOK, response)
		})

		// Stats (dados reais para a home page)
		v1.GET("/stats", statsHandler(db))

		// Ranking
		v1.GET("/ranking", rankingHandler.GetRanking)
		v1.GET("/ranking/metodologia", rankingHandler.GetMetodologia)
	}

	return router
}

// DailySyncRunner define a interface para executar o sync diario
// Implementada por scheduler.Scheduler
// SyncRunner define a interface para o scheduler executar syncs
type SyncRunner interface {
	RunDailySync(ctx context.Context)
	RunBackfill(ctx context.Context)
}

// RegisterSchedulerRoutes registra os endpoints de sync
// para serem chamados pelo Google Cloud Scheduler ou manualmente
func RegisterSchedulerRoutes(router *gin.Engine, runner SyncRunner) {
	syncSecret := os.Getenv("SYNC_SECRET")

	// Middleware de autenticacao por header secreto
	authSync := func(c *gin.Context) bool {
		if syncSecret != "" && c.GetHeader("X-Sync-Secret") != syncSecret {
			c.JSON(http.StatusForbidden, gin.H{"error": "acesso negado"})
			return false
		}
		return true
	}

	// POST /api/v1/sync/daily - Sync diario (Cloud Scheduler)
	// Executa sincronamente para manter o container vivo no Cloud Run
	router.POST("/api/v1/sync/daily", func(c *gin.Context) {
		if !authSync(c) {
			return
		}

		slog.Info("sync diario disparado via HTTP")
		runner.RunDailySync(c.Request.Context())

		c.JSON(http.StatusOK, gin.H{
			"message": "sync diario concluido",
		})
	})

	// POST /api/v1/sync/backfill - Backfill completo (manual)
	// Executa sincronamente; pode levar 30-60 min
	router.POST("/api/v1/sync/backfill", func(c *gin.Context) {
		if !authSync(c) {
			return
		}

		slog.Info("backfill completo disparado via HTTP")
		runner.RunBackfill(c.Request.Context())

		c.JSON(http.StatusOK, gin.H{
			"message": "backfill completo concluido",
		})
	})
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, X-Sync-Secret")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func healthHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verificar conexao com banco
		sqlDB, err := db.DB()
		dbStatus := "ok"
		if err != nil || sqlDB.Ping() != nil {
			dbStatus = "error"
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
			"database":  dbStatus,
		})
	}
}
