package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"to-de-olho-backend/internal/domain"
	"to-de-olho-backend/internal/infrastructure/resilience"

	"golang.org/x/time/rate"
)

// Helper para criar client de teste com circuit breaker
func newTestCamaraClient(baseURL string) *CamaraClient {
	config := resilience.DefaultCircuitBreakerConfig()
	config.Timeout = time.Second

	return &CamaraClient{
		baseURL:        baseURL,
		httpClient:     &http.Client{Timeout: time.Second},
		limiter:        rate.NewLimiter(100, 100),
		circuitBreaker: resilience.NewCircuitBreaker(config),
	}
}

func TestNewCamaraClient(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		timeout     time.Duration
		rps         int
		burst       int
		expectedURL string
	}{
		{
			name:        "valores padrão",
			baseURL:     "",
			timeout:     30 * time.Second,
			rps:         100,
			burst:       100,
			expectedURL: DefaultBaseURLCamara,
		},
		{
			name:        "URL customizada",
			baseURL:     "https://api.custom.com/v1",
			timeout:     10 * time.Second,
			rps:         50,
			burst:       75,
			expectedURL: "https://api.custom.com/v1",
		},
		{
			name:        "valores inválidos corrigidos",
			baseURL:     "",
			timeout:     5 * time.Second,
			rps:         0,
			burst:       0,
			expectedURL: DefaultBaseURLCamara,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewCamaraClient(tt.baseURL, tt.timeout, tt.rps, tt.burst)

			if client == nil {
				t.Error("NewCamaraClient() deveria retornar uma instância válida")
				return
			}

			if client.baseURL != tt.expectedURL {
				t.Errorf("baseURL esperada: %s, recebida: %s", tt.expectedURL, client.baseURL)
			}

			if client.httpClient == nil {
				t.Error("httpClient não foi inicializado")
			}

			if client.limiter == nil {
				t.Error("limiter não foi inicializado")
			}
		})
	}
}

func TestCamaraClient_FetchDeputados(t *testing.T) {
	tests := []struct {
		name          string
		setupServer   func() *httptest.Server
		expectedError bool
		expectedCount int
	}{
		{
			name: "sucesso com deputados",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := map[string]interface{}{
						"dados": []map[string]interface{}{
							{"id": 1, "nome": "João Silva", "siglaPartido": "PT"},
							{"id": 2, "nome": "Maria Santos", "siglaPartido": "PSDB"},
						},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(response)
				}))
			},
			expectedError: false,
			expectedCount: 2,
		},
		{
			name: "resposta vazia",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := map[string]interface{}{"dados": []map[string]interface{}{}}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(response)
				}))
			},
			expectedError: false,
			expectedCount: 0,
		},
		{
			name: "erro servidor",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			expectedError: true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupServer()
			defer server.Close()

			client := newTestCamaraClient(server.URL)

			ctx := context.Background()
			deputados, err := client.FetchDeputados(ctx, "", "", "")

			if tt.expectedError && err == nil {
				t.Error("esperava erro mas não ocorreu")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("não esperava erro: %v", err)
			}
			if len(deputados) != tt.expectedCount {
				t.Errorf("esperava %d deputados, recebeu %d", tt.expectedCount, len(deputados))
			}
		})
	}
}

func TestCamaraClient_FetchDeputadoByID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"dados": map[string]interface{}{
				"id":   123,
				"nome": "João Silva",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestCamaraClient(server.URL)

	ctx := context.Background()
	deputado, err := client.FetchDeputadoByID(ctx, "123")

	if err != nil {
		t.Errorf("não esperava erro: %v", err)
	}
	if deputado == nil {
		t.Error("deputado não deveria ser nil")
		return
	}
	if deputado.Nome != "João Silva" {
		t.Errorf("nome esperado: João Silva, recebido: %s", deputado.Nome)
	}
}

func TestCamaraClient_FetchDespesas(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"dados": []map[string]interface{}{
				{
					"ano":          2024,
					"mes":          1,
					"tipoDespesa":  "PASSAGEM AÉREA",
					"valorLiquido": 1500.50,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestCamaraClient(server.URL)

	ctx := context.Background()
	despesas, err := client.FetchDespesas(ctx, "123", "2024")

	if err != nil {
		t.Errorf("não esperava erro: %v", err)
	}
	if len(despesas) != 1 {
		t.Errorf("esperava 1 despesa, recebeu %d", len(despesas))
	}
	if despesas[0].Ano != 2024 {
		t.Errorf("ano esperado: 2024, recebido: %d", despesas[0].Ano)
	}
}

func TestCamaraClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestCamaraClient(server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.FetchDeputados(ctx, "", "", "")
	if err == nil {
		t.Error("esperava erro de contexto cancelado")
	}
}

func TestCamaraClient_GetVotosPorVotacao_NormalizaDeputadoEVoto(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := map[string]any{
			"dados": []map[string]any{
				{
					"deputado":         map[string]any{"id": 123, "nome": "Dep Primario"},
					"tipoVoto":         "Não",
					"voto":             "Nao",
					"dataRegistroVoto": "2024-01-01T12:00:00",
				},
				{
					"deputado_":        map[string]any{"id": 456, "nome": "Dep Legado"},
					"tipoVoto":         "Sim",
					"voto":             "",
					"dataRegistroVoto": "2024-02-02T15:30:00",
				},
				{
					"tipoVoto":         "Abstenção",
					"voto":             "",
					"dataRegistroVoto": "2024-03-03T10:10:00",
				},
			},
		}
		json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	client := newTestCamaraClient(server.URL)
	votos, err := client.GetVotosPorVotacao(context.Background(), "123456")
	if err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}
	if len(votos) != 2 {
		t.Fatalf("esperava 2 votos válidos, obteve %d", len(votos))
	}

	if votos[0].IDDeputado != 123 {
		t.Fatalf("esperava ID 123 para primeiro voto, obteve %d", votos[0].IDDeputado)
	}
	if votos[0].Voto != "Nao" {
		t.Fatalf("esperava voto 'Nao', obteve %s", votos[0].Voto)
	}
	if nome, _ := votos[0].Payload["nomeDeputado"].(string); nome != "Dep Primario" {
		t.Fatalf("esperava nome 'Dep Primario', obteve %s", nome)
	}

	if votos[1].IDDeputado != 456 {
		t.Fatalf("esperava fallback de deputado legado com id 456, obteve %d", votos[1].IDDeputado)
	}
	if votos[1].Voto != "Sim" {
		t.Fatalf("esperava voto 'Sim' a partir de tipoVoto, obteve %s", votos[1].Voto)
	}
	if fonte, _ := votos[1].Payload["fonteDeputado"].(string); fonte != "deputado_" {
		t.Fatalf("esperava fonte 'deputado_', obteve %s", fonte)
	}
	if nome, _ := votos[1].Payload["nomeDeputado"].(string); nome != "Dep Legado" {
		t.Fatalf("esperava nome 'Dep Legado', obteve %s", nome)
	}
}

// ---- Novos testes adicionais para cobertura ----

func TestCamaraClient_FetchDeputadosPaged_NormalizaParametros(t *testing.T) {
	var captured []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = append(captured, r.URL.RawQuery)
		json.NewEncoder(w).Encode(map[string]any{"dados": []map[string]any{}})
	}))
	defer server.Close()
	client := newTestCamaraClient(server.URL)

	// itens >100 deve virar 100; pagina <=0 vira 1 (observamos apenas que request acontece sem erro)
	_, err := client.FetchDeputadosPaged(context.Background(), "", "", "", -5, 1000)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(captured) == 0 || !containsAll(captured[0], []string{"pagina=1", "itens=100"}) {
		t.Fatalf("query não normalizada corretamente: %v", captured)
	}
}

func TestCamaraClient_FetchAllDeputados_PaginacaoCompleta(t *testing.T) {
	var page1Served, page2Served bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Query().Get("pagina")
		switch p {
		case "1":
			page1Served = true
			json.NewEncoder(w).Encode(map[string]any{"dados": []map[string]any{{"id": 1, "nome": "Dep 1"}}})
		case "2":
			page2Served = true
			json.NewEncoder(w).Encode(map[string]any{"dados": []map[string]any{{"id": 2, "nome": "Dep 2"}}})
		default:
			json.NewEncoder(w).Encode(map[string]any{"dados": []map[string]any{}})
		}
	}))
	defer server.Close()
	client := newTestCamaraClient(server.URL)
	deps, err := client.FetchAllDeputados(context.Background())
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(deps) != 2 {
		t.Fatalf("esperava 2 deps, obteve %d", len(deps))
	}
	if !page1Served || !page2Served {
		t.Fatalf("páginas não servidas corretamente p1=%v p2=%v", page1Served, page2Served)
	}
}

func TestCamaraClient_JSONInvalido(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{invalid"))
	}))
	defer server.Close()
	client := newTestCamaraClient(server.URL)
	if _, err := client.FetchDeputados(context.Background(), "", "", ""); err == nil {
		t.Fatalf("esperava erro de JSON inválido")
	}
}

func TestCamaraClient_RetryComSucessoNaSegundaTentativa(t *testing.T) {
	var count int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := atomic.AddInt32(&count, 1)
		if c < 2 { // Sucesso na 2ª tentativa (vs 3ª anteriormente)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"dados": []map[string]any{{"id": 1, "nome": "Ok"}}})
	}))
	defer server.Close()
	client := newTestCamaraClient(server.URL)
	deps, err := client.FetchDeputados(context.Background(), "", "", "")
	if err != nil {
		t.Fatalf("não esperava erro após retries: %v", err)
	}
	if len(deps) != 1 {
		t.Fatalf("esperava 1 dep, obteve %d", len(deps))
	}
	if count != 2 { // Esperamos 2 tentativas agora (vs 3 anteriormente)
		t.Fatalf("esperava 2 tentativas, obteve %d", count)
	}
}

func TestCamaraClient_RetryFalhaApos2Tentativas(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()
	client := newTestCamaraClient(server.URL)
	if _, err := client.FetchDeputados(context.Background(), "", "", ""); err == nil {
		t.Fatalf("esperava erro após 2 tentativas sem sucesso") // Atualizado para 2 tentativas
	}
}

func TestCamaraClient_LimiterContextCancel(t *testing.T) {
	// Usamos um limiter com alta latência simulada cancelando contexto antes de Wait
	client := &CamaraClient{baseURL: "http://127.0.0.1", httpClient: &http.Client{Timeout: time.Second}, limiter: rate.NewLimiter(rate.Limit(1), 0)}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.FetchDeputados(ctx, "", "", "")
	if err == nil {
		t.Fatalf("esperava erro por contexto cancelado antes do limiter")
	}
}

func TestCamaraClient_FetchDespesas_JSONInvalido(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{bad"))
	}))
	defer server.Close()
	client := newTestCamaraClient(server.URL)
	if _, err := client.FetchDespesas(context.Background(), "123", "2024"); err == nil {
		t.Fatalf("esperava erro de JSON inválido despesas")
	}
}

func TestCamaraClient_FetchDeputadoByID_ERROStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	client := newTestCamaraClient(server.URL)
	if _, err := client.FetchDeputadoByID(context.Background(), "999"); err == nil {
		t.Fatalf("esperava erro status 404")
	}
}

// helper
func containsAll(s string, parts []string) bool {
	for _, p := range parts {
		if !strings.Contains(s, p) {
			return false
		}
	}
	return true
}

// Testes para proposições

func TestCamaraClient_FetchProposicoes(t *testing.T) {
	tests := []struct {
		name           string
		filtros        *domain.ProposicaoFilter
		serverResponse string
		serverStatus   int
		expectError    bool
		expectedCount  int
	}{
		{
			name: "busca com sucesso",
			filtros: &domain.ProposicaoFilter{
				SiglaTipo:  "PL",
				Ano:        &[]int{2024}[0],
				Limite:     10,
				Pagina:     1,
				Ordem:      "DESC",
				OrdenarPor: "dataApresentacao",
			},
			serverResponse: `{
				"dados": [
					{
						"id": 123456,
						"uri": "https://dadosabertos.camara.leg.br/api/v2/proposicoes/123456",
						"siglaTipo": "PL",
						"codTipo": 139,
						"numero": 1234,
						"ano": 2024,
						"ementa": "Dispõe sobre teste de proposição.",
						"dataApresentacao": "2024-01-15T10:00:00",
						"descricaoTipo": "Projeto de Lei",
						"statusProposicao": {
							"id": 1,
							"uri": "https://dadosabertos.camara.leg.br/api/v2/referencias/situacoesProposicao/1",
							"descricaoSituacao": "Aguardando Designação de Relator",
							"codSituacao": 1,
							"dataHora": "2024-01-15T10:00:00",
							"sequencia": 1
						}
					},
					{
						"id": 789012,
						"uri": "https://dadosabertos.camara.leg.br/api/v2/proposicoes/789012",
						"siglaTipo": "PL",
						"codTipo": 139,
						"numero": 5678,
						"ano": 2024,
						"ementa": "Altera a Lei sobre testes.",
						"dataApresentacao": "2024-02-10T14:30:00",
						"descricaoTipo": "Projeto de Lei",
						"statusProposicao": {
							"id": 2,
							"uri": "https://dadosabertos.camara.leg.br/api/v2/referencias/situacoesProposicao/2",
							"descricaoSituacao": "Em tramitação",
							"codSituacao": 2,
							"dataHora": "2024-02-10T14:30:00",
							"sequencia": 1
						}
					}
				],
				"links": [
					{
						"rel": "self",
						"href": "https://dadosabertos.camara.leg.br/api/v2/proposicoes"
					}
				]
			}`,
			serverStatus:  http.StatusOK,
			expectError:   false,
			expectedCount: 2,
		},
		{
			name: "resposta vazia",
			filtros: &domain.ProposicaoFilter{
				SiglaTipo: "PEC",
				Limite:    5,
				Pagina:    1,
			},
			serverResponse: `{
				"dados": [],
				"links": []
			}`,
			serverStatus:  http.StatusOK,
			expectError:   false,
			expectedCount: 0,
		},
		{
			name: "erro do servidor",
			filtros: &domain.ProposicaoFilter{
				Limite: 10,
				Pagina: 1,
			},
			serverResponse: `{"erro": "Erro interno do servidor"}`,
			serverStatus:   http.StatusInternalServerError,
			expectError:    true,
			expectedCount:  0,
		},
		{
			name: "JSON inválido",
			filtros: &domain.ProposicaoFilter{
				Limite: 10,
				Pagina: 1,
			},
			serverResponse: `{invalid json`,
			serverStatus:   http.StatusOK,
			expectError:    true,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Criar servidor mock
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verificar parâmetros de query
				if tt.filtros.SiglaTipo != "" {
					if r.URL.Query().Get("siglaTipo") != tt.filtros.SiglaTipo {
						t.Errorf("Expected siglaTipo %s, got %s", tt.filtros.SiglaTipo, r.URL.Query().Get("siglaTipo"))
					}
				}

				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			client := NewCamaraClient(server.URL, 30*time.Second, 100, 100)

			// Aplicar padrões nos filtros
			tt.filtros.SetDefaults()

			proposicoes, err := client.FetchProposicoes(context.Background(), tt.filtros)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if !tt.expectError {
				if len(proposicoes) != tt.expectedCount {
					t.Errorf("Expected %d proposições but got %d", tt.expectedCount, len(proposicoes))
				}

				// Verificar dados da primeira proposição se houver
				if len(proposicoes) > 0 {
					p := proposicoes[0]
					if p.ID <= 0 {
						t.Errorf("Expected valid ID but got %d", p.ID)
					}
					if p.SiglaTipo == "" {
						t.Errorf("Expected SiglaTipo but got empty string")
					}
					if p.Ementa == "" {
						t.Errorf("Expected Ementa but got empty string")
					}
				}
			}
		})
	}
}

func TestCamaraClient_FetchProposicaoPorID(t *testing.T) {
	tests := []struct {
		name           string
		id             int
		serverResponse string
		serverStatus   int
		expectError    bool
		expectedID     int
	}{
		{
			name: "busca com sucesso",
			id:   123456,
			serverResponse: `{
				"dados": {
					"id": 123456,
					"uri": "https://dadosabertos.camara.leg.br/api/v2/proposicoes/123456",
					"siglaTipo": "PL",
					"codTipo": 139,
					"numero": 1234,
					"ano": 2024,
					"ementa": "Dispõe sobre teste de proposição individual.",
					"dataApresentacao": "2024-01-15T10:00:00",
					"descricaoTipo": "Projeto de Lei",
					"statusProposicao": {
						"id": 1,
						"uri": "https://dadosabertos.camara.leg.br/api/v2/referencias/situacoesProposicao/1",
						"descricaoSituacao": "Aguardando Designação de Relator",
						"codSituacao": 1,
						"dataHora": "2024-01-15T10:00:00",
						"sequencia": 1
					},
					"ultimoRelator": {
						"id": 204379,
						"nome": "Dep. Teste",
						"codTipo": 10000,
						"siglaUf": "SP",
						"siglaPartido": "PT",
						"uriPartido": "https://dadosabertos.camara.leg.br/api/v2/partidos/36835",
						"uriCamara": "https://dadosabertos.camara.leg.br/api/v2/deputados/204379",
						"urlFoto": "https://www.camara.leg.br/internet/deputado/bandep/204379.jpg"
					}
				}
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
			expectedID:   123456,
		},
		{
			name:           "ID inválido",
			id:             0,
			serverResponse: "",
			serverStatus:   http.StatusOK,
			expectError:    true,
			expectedID:     0,
		},
		{
			name:           "proposição não encontrada",
			id:             999999,
			serverResponse: `{"erro": "Proposição não encontrada"}`,
			serverStatus:   http.StatusNotFound,
			expectError:    true,
			expectedID:     0,
		},
		{
			name:           "JSON inválido",
			id:             123456,
			serverResponse: `{invalid json`,
			serverStatus:   http.StatusOK,
			expectError:    true,
			expectedID:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Criar servidor mock
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verificar se a URL contém o ID correto
				expectedPath := fmt.Sprintf("/proposicoes/%d", tt.id)
				if tt.id > 0 && !strings.Contains(r.URL.Path, expectedPath) {
					t.Errorf("Expected path to contain %s, got %s", expectedPath, r.URL.Path)
				}

				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			client := NewCamaraClient(server.URL, 30*time.Second, 100, 100)

			proposicao, err := client.FetchProposicaoPorID(context.Background(), tt.id)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if !tt.expectError {
				if proposicao == nil {
					t.Errorf("Expected proposição but got nil")
					return
				}

				if proposicao.ID != tt.expectedID {
					t.Errorf("Expected ID %d but got %d", tt.expectedID, proposicao.ID)
				}

				if proposicao.SiglaTipo == "" {
					t.Errorf("Expected SiglaTipo but got empty string")
				}

				if proposicao.Ementa == "" {
					t.Errorf("Expected Ementa but got empty string")
				}
			}
		})
	}
}
