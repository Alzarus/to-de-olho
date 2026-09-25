package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// relogioFalso controla o tempo do limitador nos testes.
type relogioFalso struct{ t time.Time }

func (r *relogioFalso) agora() time.Time        { return r.t }
func (r *relogioFalso) avancar(d time.Duration) { r.t = r.t.Add(d) }

// configTeste usa números pequenos para os testes ficarem legíveis.
func configTeste() ConfigLimite {
	return ConfigLimite{
		Visitante:     Taxa{PorSegundo: 1, Rajada: 3},
		Interno:       Taxa{PorSegundo: 1, Rajada: 5},
		Acessos:       Taxa{PorSegundo: 1.0 / 60, Rajada: 2},
		MaxVisitantes: 100,
	}
}

func routerLimitado(cfg ConfigLimite, r *relogioFalso) *gin.Engine {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	usarIPDoVisitante(e)
	e.Use(limitarRequisicoes(cfg, r.agora))
	ok := func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }
	e.GET("/health", ok)
	e.GET("/api/v1/ranking", ok)
	e.POST("/api/v1/acessos", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	return e
}

// origem descreve de onde vem a requisição de teste.
type origem struct {
	remoto string // RemoteAddr (IP da conexão)
	cf     string // CF-Connecting-IP; vazio = sem header
	xff    string // X-Forwarded-For; vazio = sem header
}

var (
	visitanteA = origem{remoto: "172.21.0.4:40000", cf: "203.0.113.7"}
	visitanteB = origem{remoto: "172.21.0.4:40000", cf: "198.51.100.9"}
	ssrDoNext  = origem{remoto: "172.21.0.4:40000"}
)

func pedir(e *gin.Engine, metodo, path string, o origem) *httptest.ResponseRecorder {
	req := httptest.NewRequest(metodo, path, nil)
	req.RemoteAddr = o.remoto
	if o.cf != "" {
		req.Header.Set("CF-Connecting-IP", o.cf)
	}
	if o.xff != "" {
		req.Header.Set("X-Forwarded-For", o.xff)
	}
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)
	return w
}

// liberadas conta quantas de n requisições seguidas passam.
func liberadas(e *gin.Engine, metodo, path string, o origem, n int) int {
	ok := 0
	for i := 0; i < n; i++ {
		if pedir(e, metodo, path, o).Code != http.StatusTooManyRequests {
			ok++
		}
	}
	return ok
}

func TestLimitePorOrigem(t *testing.T) {
	casos := []struct {
		nome     string
		metodo   string
		path     string
		origem   origem
		n        int
		esperado int
	}{
		{"visitante passa a rajada e para", http.MethodGet, "/api/v1/ranking", visitanteA, 10, 3},
		{"SSR sem CF-Connecting-IP usa o balde interno", http.MethodGet, "/api/v1/ranking", ssrDoNext, 10, 5},
		{"loopback sem header também é interno", http.MethodGet, "/api/v1/ranking", origem{remoto: "127.0.0.1:5000"}, 10, 5},
		{"IP público direto sem header vira visitante", http.MethodGet, "/api/v1/ranking", origem{remoto: "192.0.2.10:5000"}, 10, 3},
		{"/health nunca é limitado", http.MethodGet, "/health", visitanteA, 50, 50},
		{"POST /acessos tem limite próprio", http.MethodPost, "/api/v1/acessos", visitanteA, 10, 2},
		{"POST /acessos sem header também é limitado", http.MethodPost, "/api/v1/acessos", ssrDoNext, 10, 2},
		{"CF-Connecting-IP inválido cai no IP da conexão", http.MethodGet, "/api/v1/ranking", origem{remoto: "172.21.0.4:1", cf: "nao-e-ip"}, 10, 5},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			e := routerLimitado(configTeste(), &relogioFalso{t: time.Unix(1_800_000_000, 0)})
			if got := liberadas(e, c.metodo, c.path, c.origem, c.n); got != c.esperado {
				t.Errorf("liberadas = %d; esperado %d", got, c.esperado)
			}
		})
	}
}

func TestLimite429TemRetryAfterENaoFicaEmCache(t *testing.T) {
	relogio := &relogioFalso{t: time.Unix(1_800_000_000, 0)}
	e := routerLimitado(configTeste(), relogio)
	liberadas(e, http.MethodGet, "/api/v1/ranking", visitanteA, 3)

	w := pedir(e, http.MethodGet, "/api/v1/ranking", visitanteA)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("esperado 429; veio %d", w.Code)
	}
	if ra, err := strconv.Atoi(w.Header().Get("Retry-After")); err != nil || ra < 1 {
		t.Errorf("Retry-After deve ser inteiro >= 1; veio %q", w.Header().Get("Retry-After"))
	}
	if cc := w.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("429 deve sair com no-store; veio %q", cc)
	}

	// O Retry-After do contador de acessos reflete a taxa (1/min).
	liberadas(e, http.MethodPost, "/api/v1/acessos", visitanteA, 2)
	w = pedir(e, http.MethodPost, "/api/v1/acessos", visitanteA)
	if w.Code != http.StatusTooManyRequests || w.Header().Get("Retry-After") != "60" {
		t.Errorf("acessos: esperado 429 com Retry-After 60; veio %d %q", w.Code, w.Header().Get("Retry-After"))
	}
}

func TestLimiteRecuperaComOTempo(t *testing.T) {
	relogio := &relogioFalso{t: time.Unix(1_800_000_000, 0)}
	e := routerLimitado(configTeste(), relogio)
	liberadas(e, http.MethodGet, "/api/v1/ranking", visitanteA, 3)
	if pedir(e, http.MethodGet, "/api/v1/ranking", visitanteA).Code != http.StatusTooManyRequests {
		t.Fatal("balde deveria estar vazio")
	}
	relogio.avancar(time.Second)
	if w := pedir(e, http.MethodGet, "/api/v1/ranking", visitanteA); w.Code != http.StatusOK {
		t.Errorf("depois de 1 s deve haver uma ficha nova; veio %d", w.Code)
	}
}

func TestLimiteSeparaVisitantes(t *testing.T) {
	e := routerLimitado(configTeste(), &relogioFalso{t: time.Unix(1_800_000_000, 0)})
	liberadas(e, http.MethodGet, "/api/v1/ranking", visitanteA, 10)
	if w := pedir(e, http.MethodGet, "/api/v1/ranking", visitanteB); w.Code != http.StatusOK {
		t.Errorf("outro visitante não pode ser afetado; veio %d", w.Code)
	}
	if w := pedir(e, http.MethodGet, "/api/v1/ranking", ssrDoNext); w.Code != http.StatusOK {
		t.Errorf("o SSR não pode ser afetado por um visitante; veio %d", w.Code)
	}
}

func TestLimiteIgnoraXForwardedFor(t *testing.T) {
	// Sem Cloudflare na frente, trocar X-Forwarded-For não pode render um
	// balde novo: o header é escrito pelo cliente.
	e := routerLimitado(configTeste(), &relogioFalso{t: time.Unix(1_800_000_000, 0)})
	ok := 0
	for i := 0; i < 10; i++ {
		o := origem{remoto: "192.0.2.10:5000", xff: fmt.Sprintf("10.0.0.%d", i)}
		if pedir(e, http.MethodGet, "/api/v1/ranking", o).Code == http.StatusOK {
			ok++
		}
	}
	if ok != 3 {
		t.Errorf("X-Forwarded-For variado liberou %d; esperado 3 (um balde só)", ok)
	}
}

func TestLimiteAgrupaIPv6Por64(t *testing.T) {
	e := routerLimitado(configTeste(), &relogioFalso{t: time.Unix(1_800_000_000, 0)})
	ok := 0
	for i := 0; i < 10; i++ {
		o := origem{remoto: "172.21.0.4:1", cf: fmt.Sprintf("2001:db8:1:2::%x", i+1)}
		if pedir(e, http.MethodGet, "/api/v1/ranking", o).Code == http.StatusOK {
			ok++
		}
	}
	if ok != 3 {
		t.Errorf("endereços do mesmo /64 liberaram %d; esperado 3 (um balde só)", ok)
	}
}

func TestLimiteDesativadoComTaxaZero(t *testing.T) {
	cfg := configTeste()
	cfg.Visitante.PorSegundo = 0
	e := routerLimitado(cfg, &relogioFalso{t: time.Unix(1_800_000_000, 0)})
	if got := liberadas(e, http.MethodGet, "/api/v1/ranking", visitanteA, 50); got != 50 {
		t.Errorf("com taxa 0 o limite deve ficar desligado; liberou %d de 50", got)
	}
}

func TestBaldesLimpamOciosos(t *testing.T) {
	inicio := time.Unix(1_800_000_000, 0)
	b := novosBaldes(Taxa{PorSegundo: 1, Rajada: 3}, 100, inicio)
	for i := 0; i < 50; i++ {
		b.reservar(fmt.Sprintf("198.51.100.%d", i), inicio)
	}
	if b.tamanho() != 50 {
		t.Fatalf("esperado 50 baldes; veio %d", b.tamanho())
	}

	// Um visitante continua ativo; os outros somem depois da ociosidade.
	depois := inicio.Add(b.ociosidade + intervaloLimpeza)
	b.reservar("198.51.100.0", depois.Add(-time.Second))
	b.reservar("203.0.113.1", depois)
	if got := b.tamanho(); got != 2 {
		t.Errorf("depois da limpeza devem sobrar 2 baldes (o ativo e o novo); veio %d", got)
	}
}

func TestBaldesTetoDeMemoria(t *testing.T) {
	// Forjar um IP diferente a cada requisição não pode crescer a tabela
	// sem limite nem dar vazão ilimitada: acima do teto, todos dividem o
	// balde de excedente.
	agora := time.Unix(1_800_000_000, 0)
	b := novosBaldes(Taxa{PorSegundo: 1, Rajada: 3}, 10, agora)
	liberados := 0
	for i := 0; i < 1000; i++ {
		if b.reservar(fmt.Sprintf("ip-forjado-%d", i), agora) == 0 {
			liberados++
		}
	}
	if b.tamanho() != 10 {
		t.Errorf("a tabela deve parar no teto de 10; veio %d", b.tamanho())
	}
	// Cada IP forjado usa uma vez: 10 passam pelos baldes próprios e os
	// outros 990 disputam as 3 fichas do excedente.
	if liberados != 10+3 {
		t.Errorf("liberados = %d; esperado %d (um por balde próprio + a rajada do excedente)", liberados, 10+3)
	}
}

func TestOciosidadeNaoApagaBaldeAntesDeEncher(t *testing.T) {
	casos := []struct {
		nome     string
		taxa     Taxa
		esperado time.Duration
	}{
		{"visitante enche rápido: vale o mínimo", Taxa{PorSegundo: 10, Rajada: 40}, intervaloLimpeza},
		{"acessos enche em 10 min", Taxa{PorSegundo: 1.0 / 60, Rajada: 10}, 10 * time.Minute},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			b := novosBaldes(c.taxa, 10, time.Now())
			if diff := b.ociosidade - c.esperado; diff < -time.Millisecond || diff > time.Millisecond {
				t.Errorf("ociosidade = %v; esperado %v", b.ociosidade, c.esperado)
			}
		})
	}
}

func TestConfigLimiteDoAmbiente(t *testing.T) {
	casos := []struct {
		nome     string
		env      map[string]string
		conferir func(ConfigLimite) bool
	}{
		{"sem env usa o padrão", nil, func(c ConfigLimite) bool { return c == ConfigLimitePadrao() }},
		{"RPS e BURST do visitante", map[string]string{"RATE_LIMIT_RPS": "5", "RATE_LIMIT_BURST": "20"},
			func(c ConfigLimite) bool { return c.Visitante == Taxa{PorSegundo: 5, Rajada: 20} }},
		{"acessos por minuto", map[string]string{"RATE_LIMIT_ACESSOS_POR_MINUTO": "6"},
			func(c ConfigLimite) bool { return c.Acessos.PorSegundo == 0.1 }},
		{"valor inválido fica no padrão", map[string]string{"RATE_LIMIT_RPS": "muito", "RATE_LIMIT_BURST": "-3"},
			func(c ConfigLimite) bool { return c.Visitante == ConfigLimitePadrao().Visitante }},
		{"zero desliga", map[string]string{"RATE_LIMIT_INTERNO_RPS": "0"},
			func(c ConfigLimite) bool { return !c.Interno.ativa() }},
	}
	nomes := []string{"RATE_LIMIT_RPS", "RATE_LIMIT_BURST", "RATE_LIMIT_INTERNO_RPS", "RATE_LIMIT_INTERNO_BURST",
		"RATE_LIMIT_ACESSOS_POR_MINUTO", "RATE_LIMIT_ACESSOS_BURST", "RATE_LIMIT_MAX_VISITANTES"}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			for _, n := range nomes {
				t.Setenv(n, c.env[n])
			}
			if cfg := ConfigLimiteDoAmbiente(); !c.conferir(cfg) {
				t.Errorf("config inesperada: %+v", cfg)
			}
		})
	}
}
