package api

import (
	"fmt"
	"net/http"
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
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// SetupRouter configura todas as rotas da API
func SetupRouter(db *gorm.DB, redisClient *redis.Client, transparenciaAPIKey string) *gin.Engine {
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
		rankingService := ranking.NewService(senadorRepo, proposicaoRepo, votacaoRepo, ceapsRepo, comissaoRepo, redisClient)
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

		v1.POST("/sync/emendas", func(c *gin.Context) {
			// Start backfill in background or sync (using 2024/2025 default or param)
			// For simplicity: sync all years 2023-2026 or just current
			// Since this is heavy, maybe we just trigger one year? Or lets make it accept "ano"?
			// Implementation plan implied general sync. Lets do loop for 2023..2025 like scheduler?
			// Better: POST /sync/emendas/:ano
			
			// Let's stick to simplest: Sync All years for all senators (heavy)
			// But ctx timeout is issue.
			// Let's do like scheduler: 2023..2026 inside a go routine? 
			// No, manual sync is usually synchronous for feedback.
			
			// Let's implement /sync/emendas/:ano
			ano := time.Now().Year() // default
			
			// Using SyncAll from service
			// It iterates all senators.
			if err := emendaSync.SyncAll(c.Request.Context(), ano); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return 
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "sync de emendas concluido",
				"ano": ano,
			})
		})
        
        // Let's refine: actually we wanted /sync/emendas to match others.
		// Added the route above. But let's support :ano param optionally or make another route.
		// Let's follow pattern: /sync/emendas/:ano
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

		// Ranking
		v1.GET("/ranking", rankingHandler.GetRanking)
		v1.GET("/ranking/metodologia", rankingHandler.GetMetodologia)
	}

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

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
