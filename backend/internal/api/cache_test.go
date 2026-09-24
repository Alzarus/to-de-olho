package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func routerComCache() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(corsMiddleware())
	r.Use(cacheDeLeitura())
	ok := func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }
	r.GET("/health", ok)
	r.GET("/api/v1/ranking", ok)
	r.GET("/api/v1/stats", ok)
	r.GET("/api/v1/senadores/:id", func(c *gin.Context) {
		if c.Param("id") == "0" {
			c.JSON(http.StatusNotFound, gin.H{"error": "senador nao encontrado"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": c.Param("id")})
	})
	r.GET("/api/v1/falha", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha"})
	})
	r.GET("/api/v1/texto", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.GET("/api/v1/proprio", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.GET("/api/v1/sync/status", ok)
	r.POST("/api/v1/acessos", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	return r
}

func TestCacheControlNasLeituras(t *testing.T) {
	casos := []struct {
		nome, metodo, path, esperado string
	}{
		{"GET 200 público", http.MethodGet, "/api/v1/ranking", cachePublico},
		{"GET 200 com parâmetro", http.MethodGet, "/api/v1/senadores/49", cachePublico},
		{"GET 200 com c.String", http.MethodGet, "/api/v1/texto", cachePublico},
		{"stats tem cópia mais curta", http.MethodGet, "/api/v1/stats", cacheStats},
		{"404 não entra em cache", http.MethodGet, "/api/v1/senadores/0", ""},
		{"500 não entra em cache", http.MethodGet, "/api/v1/falha", ""},
		{"rota inexistente não entra em cache", http.MethodGet, "/api/v1/nao-existe", ""},
		{"health fica de fora", http.MethodGet, "/health", ""},
		{"sync fica de fora", http.MethodGet, "/api/v1/sync/status", ""},
		{"POST fica de fora", http.MethodPost, "/api/v1/acessos", ""},
		{"Cache-Control da própria rota é mantido", http.MethodGet, "/api/v1/proprio", "no-store"},
	}
	r := routerComCache()
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(c.metodo, c.path, nil))
			if got := w.Header().Get("Cache-Control"); got != c.esperado {
				t.Errorf("%s %s (status %d): Cache-Control = %q; esperado %q", c.metodo, c.path, w.Code, got, c.esperado)
			}
		})
	}
}

func TestCorsSoAnunciaLeitura(t *testing.T) {
	casos := []struct {
		nome, metodo string
		status       int
	}{
		{"GET", http.MethodGet, http.StatusOK},
		{"preflight", http.MethodOptions, http.StatusNoContent},
	}
	r := routerComCache()
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(c.metodo, "/api/v1/ranking", nil))
			if w.Code != c.status {
				t.Errorf("status = %d; esperado %d", w.Code, c.status)
			}
			if got := w.Header().Get("Access-Control-Allow-Methods"); got != "GET, OPTIONS" {
				t.Errorf("Access-Control-Allow-Methods = %q; esperado só GET, OPTIONS", got)
			}
			if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
				t.Errorf("Access-Control-Allow-Origin = %q; esperado *", got)
			}
		})
	}
}

// Um sync autenticado mais lento que o WriteTimeout do servidor ainda
// precisa entregar a resposta; uma rota pública no mesmo servidor não.
func TestSyncAutenticadoSobreviveAoWriteTimeout(t *testing.T) {
	comSegredo(t, "segredo-de-producao")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	lento := func(c *gin.Context) {
		time.Sleep(300 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"message": "concluido"})
	}
	r.POST("/api/v1/sync/votacoes", requireSyncSecret(), semPrazoDeEscrita(), lento)
	r.GET("/api/v1/ranking", lento)

	srv := httptest.NewUnstartedServer(r)
	srv.Config.WriteTimeout = 100 * time.Millisecond
	srv.Start()
	t.Cleanup(srv.Close)

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/sync/votacoes", nil)
	req.Header.Set(SyncSecretHeader, "segredo-de-producao")
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("sync autenticado perdeu a resposta: %v", err)
	}
	corpo, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("sync: status %d (%s)", resp.StatusCode, corpo)
	}

	// Controle: sem semPrazoDeEscrita a mesma lentidão perde a resposta.
	if resp, err := srv.Client().Get(srv.URL + "/api/v1/ranking"); err == nil {
		_, errCorpo := io.ReadAll(resp.Body)
		resp.Body.Close()
		if errCorpo == nil {
			t.Errorf("rota pública lenta deveria estourar o WriteTimeout; veio %d", resp.StatusCode)
		}
	}
}
