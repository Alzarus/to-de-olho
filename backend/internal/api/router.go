package api

import (
	"context"
	"fmt"
	"log/slog"
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
	"gorm.io/gorm"
)

// SetupRouter configura todas as rotas da API
func SetupRouter(db *gorm.DB, transparenciaAPIKey string) *gin.Engine {
	router := gin.Default()
	usarIPDoVisitante(router)

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

		// Sync (trigger manual). Protegido por X-Sync-Secret: sao jobs de
		// ingestao pesados, nao endpoints publicos.
		syncGroup := v1.Group("/sync", requireSyncSecret())

		syncGroup.POST("/senadores", func(c *gin.Context) {
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

		syncGroup.POST("/despesas/:ano", func(c *gin.Context) {
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

		syncGroup.POST("/votacoes", func(c *gin.Context) {
			if err := votacaoSync.SyncFromAPI(c.Request.Context()); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "sync de votacoes concluido",
			})
		})

		syncGroup.POST("/comissoes", func(c *gin.Context) {
			if err := comissaoSync.SyncFromAPI(c.Request.Context()); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "sync de comissoes concluido",
			})
		})

		syncGroup.POST("/proposicoes", func(c *gin.Context) {
			if err := proposicaoSync.SyncFromAPI(c.Request.Context()); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "sync de proposicoes concluido",
			})
		})

		syncGroup.POST("/emendas/:ano", func(c *gin.Context) {
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
	authSync := requireSyncSecret()

	// POST /api/v1/sync/daily - Sync diario (Cloud Scheduler)
	// Executa sincronamente para manter o container vivo no Cloud Run
	router.POST("/api/v1/sync/daily", authSync, func(c *gin.Context) {
		slog.Info("sync diario disparado via HTTP")
		runner.RunDailySync(c.Request.Context())

		c.JSON(http.StatusOK, gin.H{
			"message": "sync diario concluido",
		})
	})

	// POST /api/v1/sync/backfill - Backfill completo (manual)
	// Retorna 202 imediatamente; backfill roda em background (sem limite de tempo)
	router.POST("/api/v1/sync/backfill", authSync, func(c *gin.Context) {
		slog.Info("backfill completo disparado via HTTP")

		// Rodar em goroutine com contexto independente do request HTTP
		// para nao ficar preso ao timeout de 3600s do Cloud Run
		go runner.RunBackfill(context.Background())

		c.JSON(http.StatusAccepted, gin.H{
			"message": "backfill iniciado em background",
		})
	})
}

// usarIPDoVisitante faz c.ClientIP() devolver o IP real de quem fez o request.
//
// Em producao o trafego chega por Cloudflare -> Nginx Proxy Manager -> Next ->
// API, entao o IP da conexao e sempre o do conteiner do Next. A Cloudflare
// sobrescreve CF-Connecting-IP com o IP do visitante; sem o header (ambiente
// local), o gin volta ao comportamento padrao.
//
// Serve para log. Quem alcancar a origem sem passar pela Cloudflare consegue
// forjar o header, entao o valor nao deve ser usado para autorizar nada.
func usarIPDoVisitante(r *gin.Engine) {
	r.TrustedPlatform = gin.PlatformCloudflare
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Os endpoints de sync nao sao chamados por navegador: nao recebem
		// header de CORS e nao respondem preflight. Sem isso, qualquer pagina
		// aberta por um visitante podia tentar disparar um backfill.
		if isSyncPath(c.Request.URL.Path) {
			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			c.Next()
			return
		}

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
