package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pedroalmeida/to-de-olho/internal/ceaps"
	"github.com/pedroalmeida/to-de-olho/internal/senador"
	"github.com/pedroalmeida/to-de-olho/internal/votacao"
	"github.com/pedroalmeida/to-de-olho/pkg/senado"
	"gorm.io/gorm"
)

// SetupRouter configura todas as rotas da API
func SetupRouter(db *gorm.DB) *gin.Engine {
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

		// Ranking (placeholder)
		v1.GET("/ranking", rankingPlaceholder)
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

func rankingPlaceholder(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "ranking endpoint em desenvolvimento",
		"info":    "Este endpoint retornara o ranking de senadores por score",
	})
}
