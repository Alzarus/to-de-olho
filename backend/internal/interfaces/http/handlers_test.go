package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"to-de-olho-backend/internal/domain"

	"github.com/gin-gonic/gin"
)

// Mock service para testes de handlers
type MockDeputadosService struct {
	deputados []domain.Deputado
	deputado  *domain.Deputado
	despesas  []domain.Despesa
	source    string
	err       error
}

func (m *MockDeputadosService) ListarDeputados(ctx context.Context, partido, uf, nome string) ([]domain.Deputado, string, error) {
	if m.err != nil {
		return nil, "", m.err
	}
	return m.deputados, m.source, nil
}

func (m *MockDeputadosService) BuscarDeputadoPorID(ctx context.Context, id string) (*domain.Deputado, string, error) {
	if m.err != nil {
		return nil, "", m.err
	}
	return m.deputado, m.source, nil
}

func (m *MockDeputadosService) ListarDespesas(ctx context.Context, deputadoID, ano string) ([]domain.Despesa, string, error) {
	if m.err != nil {
		return nil, "", m.err
	}
	return m.despesas, m.source, nil
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestGetDeputadosHandler(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockDeputados  []domain.Deputado
		mockSource     string
		mockError      error
		expectedStatus int
		expectedData   bool
	}{
		{
			name:        "sucesso - lista deputados sem filtro",
			queryParams: "",
			mockDeputados: []domain.Deputado{
				{ID: 1, Nome: "João Silva", Partido: "PT", UF: "SP"},
				{ID: 2, Nome: "Maria Santos", Partido: "PSDB", UF: "RJ"},
			},
			mockSource:     "api",
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name:        "sucesso - filtro por partido",
			queryParams: "?partido=PT",
			mockDeputados: []domain.Deputado{
				{ID: 1, Nome: "João Silva", Partido: "PT", UF: "SP"},
			},
			mockSource:     "cache",
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name:        "sucesso - filtros múltiplos",
			queryParams: "?partido=PT&uf=SP&nome=João",
			mockDeputados: []domain.Deputado{
				{ID: 1, Nome: "João Silva", Partido: "PT", UF: "SP"},
			},
			mockSource:     "api",
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name:           "erro interno do serviço",
			queryParams:    "",
			mockError:      errors.New("erro de rede"),
			expectedStatus: http.StatusInternalServerError,
			expectedData:   false,
		},
		{
			name:           "sucesso - lista vazia",
			queryParams:    "?partido=INEXISTENTE",
			mockDeputados:  []domain.Deputado{},
			mockSource:     "api",
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockDeputadosService{
				deputados: tt.mockDeputados,
				source:    tt.mockSource,
				err:       tt.mockError,
			}

			router := setupRouter()
			router.GET("/deputados", GetDeputadosHandler(mockService))

			// Execute
			req := httptest.NewRequest("GET", "/deputados"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("status esperado: %d, recebido: %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedData && tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("erro ao decodificar JSON: %v", err)
				}

				if response["data"] == nil {
					t.Error("resposta deveria conter campo 'data'")
				}

				if response["total"] == nil {
					t.Error("resposta deveria conter campo 'total'")
				}

				if response["source"] != tt.mockSource {
					t.Errorf("source esperado: %s, recebido: %v", tt.mockSource, response["source"])
				}

				data := response["data"].([]interface{})
				if len(data) != len(tt.mockDeputados) {
					t.Errorf("quantidade esperada: %d, recebida: %d", len(tt.mockDeputados), len(data))
				}
			}
		})
	}
}

func TestGetDeputadoByIDHandler(t *testing.T) {
	tests := []struct {
		name           string
		deputadoID     string
		mockDeputado   *domain.Deputado
		mockSource     string
		mockError      error
		expectedStatus int
	}{
		{
			name:       "sucesso - deputado encontrado",
			deputadoID: "123",
			mockDeputado: &domain.Deputado{
				ID:      123,
				Nome:    "João Silva",
				Partido: "PT",
				UF:      "SP",
				Email:   "joao@camara.leg.br",
			},
			mockSource:     "api",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "deputado não encontrado",
			deputadoID:     "999",
			mockError:      errors.New("deputado não encontrado"),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "sucesso - dados do cache",
			deputadoID: "456",
			mockDeputado: &domain.Deputado{
				ID:      456,
				Nome:    "Maria Santos",
				Partido: "PSDB",
				UF:      "RJ",
			},
			mockSource:     "cache",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockDeputadosService{
				deputado: tt.mockDeputado,
				source:   tt.mockSource,
				err:      tt.mockError,
			}

			router := setupRouter()
			router.GET("/deputados/:id", GetDeputadoByIDHandler(mockService))

			// Execute
			req := httptest.NewRequest("GET", "/deputados/"+tt.deputadoID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("status esperado: %d, recebido: %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("erro ao decodificar JSON: %v", err)
				}

				if response["data"] == nil {
					t.Error("resposta deveria conter campo 'data'")
				}

				if response["source"] != tt.mockSource {
					t.Errorf("source esperado: %s, recebido: %v", tt.mockSource, response["source"])
				}
			}
		})
	}
}

func TestGetDespesasDeputadoHandler(t *testing.T) {
	tests := []struct {
		name           string
		deputadoID     string
		ano            string
		mockDespesas   []domain.Despesa
		mockSource     string
		mockError      error
		expectedStatus int
	}{
		{
			name:       "sucesso - despesas encontradas",
			deputadoID: "123",
			ano:        "2024",
			mockDespesas: []domain.Despesa{
				{Ano: 2024, Mes: 1, ValorLiquido: 150.75, TipoDespesa: "COMBUSTÍVEL"},
				{Ano: 2024, Mes: 2, ValorLiquido: 300.50, TipoDespesa: "ALIMENTAÇÃO"},
			},
			mockSource:     "api",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "sucesso - ano padrão (atual)",
			deputadoID:     "123",
			ano:            "", // Teste do DefaultQuery
			mockDespesas:   []domain.Despesa{},
			mockSource:     "api",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "erro interno do serviço",
			deputadoID:     "123",
			ano:            "2024",
			mockError:      errors.New("erro ao buscar despesas"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "sucesso - lista vazia",
			deputadoID:     "123",
			ano:            "2020",
			mockDespesas:   []domain.Despesa{},
			mockSource:     "api",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockDeputadosService{
				despesas: tt.mockDespesas,
				source:   tt.mockSource,
				err:      tt.mockError,
			}

			router := setupRouter()
			router.GET("/deputados/:id/despesas", GetDespesasDeputadoHandler(mockService))

			// Execute
			url := "/deputados/" + tt.deputadoID + "/despesas"
			if tt.ano != "" {
				url += "?ano=" + tt.ano
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("status esperado: %d, recebido: %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("erro ao decodificar JSON: %v", err)
				}

				if response["data"] == nil {
					t.Error("resposta deveria conter campo 'data'")
				}

				if response["total"] == nil {
					t.Error("resposta deveria conter campo 'total'")
				}

				if response["valor_total"] == nil {
					t.Error("resposta deveria conter campo 'valor_total'")
				}

				// Verificar cálculo do total de valores
				data := response["data"].([]interface{})
				if len(data) != len(tt.mockDespesas) {
					t.Errorf("quantidade esperada: %d, recebida: %d", len(tt.mockDespesas), len(data))
				}

				// Calcular total esperado
				var expectedTotal float64
				for _, despesa := range tt.mockDespesas {
					expectedTotal += despesa.ValorLiquido
				}

				totalValor := response["valor_total"].(float64)
				if totalValor != expectedTotal {
					t.Errorf("total esperado: %.2f, recebido: %.2f", expectedTotal, totalValor)
				}
			}
		})
	}
}

// Benchmark para handlers críticos
func BenchmarkGetDeputadosHandler(b *testing.B) {
	mockService := &MockDeputadosService{
		deputados: []domain.Deputado{
			{ID: 1, Nome: "João Silva", Partido: "PT", UF: "SP"},
			{ID: 2, Nome: "Maria Santos", Partido: "PSDB", UF: "RJ"},
		},
		source: "api",
	}

	router := setupRouter()
	router.GET("/deputados", GetDeputadosHandler(mockService))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/deputados", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
