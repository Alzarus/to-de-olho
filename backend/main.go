package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Deputado struct {
	ID       int    `json:"id"`
	Nome     string `json:"nome"`
	Partido  string `json:"siglaPartido"`
	UF       string `json:"siglaUf"`
	URLFoto  string `json:"urlFoto"`
	Situacao string `json:"condicaoEleitoral"`
	Email    string `json:"email"`
}

type APIResponseDeputados struct {
	Dados []Deputado `json:"dados"`
	Links []struct {
		Rel  string `json:"rel"`
		Href string `json:"href"`
	} `json:"links"`
}

type Despesa struct {
	Ano            int     `json:"ano"`
	Mes            int     `json:"mes"`
	TipoDespesa    string  `json:"tipoDespesa"`
	CodDocumento   int     `json:"codDocumento"`
	TipoDocumento  string  `json:"tipoDocumento"`
	CodTipoDoc     int     `json:"codTipoDocumento"`
	DataDocumento  string  `json:"dataDocumento"`
	NumDocumento   string  `json:"numDocumento"`
	ValorLiquido   float64 `json:"valorLiquido"`
	Fornecedor     string  `json:"nomeFornecedor"`
	CNPJFornecedor string  `json:"cnpjCpfFornecedor"`
}

type APIResponseDespesas struct {
	Dados []Despesa `json:"dados"`
	Links []struct {
		Rel  string `json:"rel"`
		Href string `json:"href"`
	} `json:"links"`
}

func main() {
	// Carregar variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: arquivo .env não encontrado")
	}

	// Contexto base
	ctx := context.Background()

	// Conexões e clients
	pgPool, err := newPostgresPool(ctx)
	if err != nil {
		log.Printf("Aviso: não foi possível conectar ao Postgres: %v", err)
	}
	cache := newCache()
	repo := NewDeputadoRepository(pgPool)
	client := NewCamaraClient("", 30*time.Second, 2, 4)

	// Configurar Gin
	r := gin.Default()

	// Configurar CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// Middleware de logging
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Rotas da API
	api := r.Group("/api/v1")
	{
		api.GET("/health", healthCheck)
		api.GET("/deputados", func(c *gin.Context) { getDeputados(c, ctx, cache, client, repo) })
		api.GET("/deputados/:id", func(c *gin.Context) { getDeputadoByID(c, ctx, cache, client) })
		api.GET("/deputados/:id/despesas", func(c *gin.Context) { getDespesasDeputado(c, ctx, cache, client) })
	}

	// Porta do servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Servidor rodando na porta %s", port)
	log.Printf("📊 API disponível em: http://localhost:%s/api/v1", port)
	log.Printf("🏥 Health check: http://localhost:%s/api/v1/health", port)

	r.Run(":" + port)
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "OK",
		"message":   "API Tô De Olho funcionando!",
		"timestamp": "2025-08-12T00:00:00Z",
		"version":   "1.0.0",
	})
}

func getDeputados(c *gin.Context, ctx context.Context, cache *Cache, client *CamaraClient, repo *DeputadoRepository) {
	// Parâmetros de query
	partido := c.Query("partido")
	uf := c.Query("uf")
	nome := c.Query("nome")

	// Cache key
	keyBytes, _ := json.Marshal(map[string]string{"p": partido, "u": uf, "n": nome})
	cacheKey := "deputados:" + string(keyBytes)

	// Tenta cache
	if v, ok := cache.Get(ctx, cacheKey); ok && v != "" {
		var cached []Deputado
		if err := json.Unmarshal([]byte(v), &cached); err == nil {
			c.JSON(http.StatusOK, gin.H{"data": cached, "total": len(cached), "source": "cache"})
			return
		}
	}

	// Chamar API da Câmara
	deputados, err := client.FetchDeputados(ctx, partido, uf, nome)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Erro ao buscar deputados",
			"details": err.Error(),
		})
		return
	}

	// Persistência leve e cache
	_ = repo.UpsertDeputados(ctx, deputados)
	if b, err := json.Marshal(deputados); err == nil {
		cache.Set(ctx, cacheKey, string(b), 2*time.Minute)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   deputados,
		"total":  len(deputados),
		"source": "API Câmara dos Deputados",
	})
}

func getDeputadoByID(c *gin.Context, ctx context.Context, cache *Cache, client *CamaraClient) {
	id := c.Param("id")
	if v, ok := cache.Get(ctx, "deputado:"+id); ok && v != "" {
		var d Deputado
		if err := json.Unmarshal([]byte(v), &d); err == nil {
			c.JSON(http.StatusOK, gin.H{"data": d, "source": "cache"})
			return
		}
	}

	deputado, err := client.FetchDeputadoByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Deputado não encontrado",
			"details": err.Error(),
		})
		return
	}

	if b, err := json.Marshal(deputado); err == nil {
		cache.Set(ctx, "deputado:"+id, string(b), 5*time.Minute)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   deputado,
		"source": "API Câmara dos Deputados",
	})
}

func getDespesasDeputado(c *gin.Context, ctx context.Context, cache *Cache, client *CamaraClient) {
	id := c.Param("id")
	ano := c.DefaultQuery("ano", "2025")
	cacheKey := "despesas:" + id + ":" + ano
	if v, ok := cache.Get(ctx, cacheKey); ok && v != "" {
		var cached []Despesa
		if err := json.Unmarshal([]byte(v), &cached); err == nil {
			var t float64
			for _, d := range cached {
				t += d.ValorLiquido
			}
			c.JSON(http.StatusOK, gin.H{"data": cached, "total": len(cached), "valor_total": t, "ano": ano, "source": "cache"})
			return
		}
	}

	despesas, err := client.FetchDespesas(ctx, id, ano)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Erro ao buscar despesas",
			"details": err.Error(),
		})
		return
	}

	// Calcular estatísticas básicas
	var total float64
	for _, despesa := range despesas {
		total += despesa.ValorLiquido
	}

	if b, err := json.Marshal(despesas); err == nil {
		cache.Set(ctx, cacheKey, string(b), 1*time.Minute)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        despesas,
		"total":       len(despesas),
		"valor_total": total,
		"ano":         ano,
		"source":      "API Câmara dos Deputados",
	})
}
