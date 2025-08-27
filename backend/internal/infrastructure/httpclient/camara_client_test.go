package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

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

			client := &CamaraClient{
				baseURL:    server.URL,
				httpClient: &http.Client{Timeout: 5 * time.Second},
				limiter:    rate.NewLimiter(rate.Limit(100), 100),
			}

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

	client := &CamaraClient{
		baseURL:    server.URL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		limiter:    rate.NewLimiter(rate.Limit(100), 100),
	}

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

	client := &CamaraClient{
		baseURL:    server.URL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		limiter:    rate.NewLimiter(rate.Limit(100), 100),
	}

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

	client := &CamaraClient{
		baseURL:    server.URL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		limiter:    rate.NewLimiter(rate.Limit(100), 100),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.FetchDeputados(ctx, "", "", "")
	if err == nil {
		t.Error("esperava erro de contexto cancelado")
	}
}

// end of file: camara_client_test.go (normalized to LF)
