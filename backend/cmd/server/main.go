package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	app "to-de-olho-backend/internal/application"
	"to-de-olho-backend/internal/infrastructure/cache"
	"to-de-olho-backend/internal/infrastructure/db"
	"to-de-olho-backend/internal/infrastructure/httpclient"
	"to-de-olho-backend/internal/infrastructure/repository"
	httpif "to-de-olho-backend/internal/interfaces/http"
	"to-de-olho-backend/internal/interfaces/http/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: arquivo .env não encontrado")
	}

	ctx := context.Background()

	pgPool, err := db.NewPostgresPool(ctx)
	if err != nil {
		log.Printf("Aviso: não foi possível conectar ao Postgres: %v", err)
	}
	cacheClient := cache.New()
	repo := repository.NewDeputadoRepository(pgPool)
	client := httpclient.NewCamaraClient("", 30*time.Second, 2, 4)

	svc := app.NewDeputadosService(client, cacheClient, repo)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.RateLimitPerIP(100, time.Minute))

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
		api.GET("/deputados", httpif.GetDeputadosHandler(svc))
		api.GET("/deputados/:id", httpif.GetDeputadoByIDHandler(svc))
		api.GET("/deputados/:id/despesas", httpif.GetDespesasDeputadoHandler(svc))
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Servidor rodando na porta %s", port)
	log.Printf("📊 API disponível em: http://localhost:%s/api/v1", port)
	_ = r.Run(":" + port)
}
