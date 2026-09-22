package api

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// comSegredo troca o leitor de SYNC_SECRET durante o teste.
func comSegredo(t *testing.T, valor string) {
	t.Helper()
	anterior := syncSecret
	syncSecret = func() string { return valor }
	t.Cleanup(func() { syncSecret = anterior })
}

// rotaProtegida monta um router minimo com o middleware aplicado.
func rotaProtegida() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	usarIPDoVisitante(r)
	r.Use(corsMiddleware())
	r.POST("/api/v1/sync/votacoes", requireSyncSecret(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "executado"})
	})
	r.GET("/api/v1/ranking", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "publico"})
	})
	return r
}

func executar(r *gin.Engine, metodo, path, segredo string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(metodo, path, nil)
	if segredo != "" {
		req.Header.Set(SyncSecretHeader, segredo)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestSyncSemSegredoConfiguradoFalhaFechado(t *testing.T) {
	comSegredo(t, "")
	w := executar(rotaProtegida(), http.MethodPost, "/api/v1/sync/votacoes", "qualquer-coisa")
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("sem SYNC_SECRET deve recusar com 503; veio %d (%s)", w.Code, w.Body.String())
	}
}

func TestSyncSemHeaderERecusado(t *testing.T) {
	comSegredo(t, "segredo-de-producao")
	w := executar(rotaProtegida(), http.MethodPost, "/api/v1/sync/votacoes", "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("sync anonimo deve receber 401; veio %d (%s)", w.Code, w.Body.String())
	}
}

func TestSyncComSegredoErradoERecusado(t *testing.T) {
	comSegredo(t, "segredo-de-producao")
	w := executar(rotaProtegida(), http.MethodPost, "/api/v1/sync/votacoes", "segredo-errado")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("segredo errado deve receber 401; veio %d (%s)", w.Code, w.Body.String())
	}
}

func TestSyncComSegredoCorretoExecuta(t *testing.T) {
	comSegredo(t, "segredo-de-producao")
	w := executar(rotaProtegida(), http.MethodPost, "/api/v1/sync/votacoes", "segredo-de-producao")
	if w.Code != http.StatusOK {
		t.Errorf("segredo correto deve executar; veio %d (%s)", w.Code, w.Body.String())
	}
}

func TestPreflightDeSyncNaoERespondido(t *testing.T) {
	comSegredo(t, "segredo-de-producao")
	w := executar(rotaProtegida(), http.MethodOptions, "/api/v1/sync/votacoes", "")
	if w.Code != http.StatusNotFound {
		t.Errorf("preflight em /sync/* deve dar 404; veio %d", w.Code)
	}
	if origem := w.Header().Get("Access-Control-Allow-Origin"); origem != "" {
		t.Errorf("/sync/* nao deve liberar CORS; veio Access-Control-Allow-Origin=%q", origem)
	}
}

func TestCorsSegueLiberadoNasRotasPublicas(t *testing.T) {
	comSegredo(t, "segredo-de-producao")
	w := executar(rotaProtegida(), http.MethodGet, "/api/v1/ranking", "")
	if w.Code != http.StatusOK {
		t.Fatalf("rota publica deve responder 200; veio %d", w.Code)
	}
	if origem := w.Header().Get("Access-Control-Allow-Origin"); origem != "*" {
		t.Errorf("rota publica deve manter CORS aberto; veio %q", origem)
	}
}

func TestIsSyncPath(t *testing.T) {
	casos := map[string]bool{
		"/api/v1/sync":               true,
		"/api/v1/sync/backfill":      true,
		"/api/v1/sync/despesas/2024": true,
		"/api/v1/ranking":            false,
		"/api/v1/senadores/49":       false,
		"/health":                    false,
	}
	for path, esperado := range casos {
		if got := isSyncPath(path); got != esperado {
			t.Errorf("isSyncPath(%q) = %v; esperado %v", path, got, esperado)
		}
	}
}

// ipLogado dispara um sync sem credencial e devolve o IP gravado no log.
func ipLogado(t *testing.T, cfConnectingIP string) string {
	t.Helper()
	comSegredo(t, "segredo-de-producao")

	var buf bytes.Buffer
	anterior := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(anterior) })

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sync/votacoes", nil)
	req.RemoteAddr = "172.21.0.4:40000" // conteiner do Next
	if cfConnectingIP != "" {
		req.Header.Set("CF-Connecting-IP", cfConnectingIP)
	}
	w := httptest.NewRecorder()
	rotaProtegida().ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401; veio %d", w.Code)
	}

	var entrada struct {
		IP string `json:"ip"`
	}
	if err := json.Unmarshal(buf.Bytes(), &entrada); err != nil {
		t.Fatalf("log invalido %q: %v", buf.String(), err)
	}
	return entrada.IP
}

func TestLogDeSyncNegadoUsaIPDaCloudflare(t *testing.T) {
	if ip := ipLogado(t, "203.0.113.7"); ip != "203.0.113.7" {
		t.Errorf("log deve trazer o IP do visitante; veio %q", ip)
	}
}

func TestLogDeSyncNegadoSemCloudflareUsaIPDaConexao(t *testing.T) {
	if ip := ipLogado(t, ""); ip != "172.21.0.4" {
		t.Errorf("sem CF-Connecting-IP o log deve trazer o IP da conexao; veio %q", ip)
	}
}
