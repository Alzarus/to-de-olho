package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"to-de-olho-backend/internal/domain"
)

// Mocks para as interfaces
type MockCamaraClient struct {
	deputados []domain.Deputado
	deputado  *domain.Deputado
	despesas  []domain.Despesa
	err       error
}

func (m *MockCamaraClient) FetchDeputados(ctx context.Context, partido, uf, nome string) ([]domain.Deputado, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.deputados, nil
}

func (m *MockCamaraClient) FetchDeputadoByID(ctx context.Context, id string) (*domain.Deputado, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.deputado, nil
}

func (m *MockCamaraClient) FetchDespesas(ctx context.Context, deputadoID, ano string) ([]domain.Despesa, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.despesas, nil
}

type MockCache struct {
	data map[string]string
}

func NewMockCache() *MockCache {
	return &MockCache{data: make(map[string]string)}
}

func (m *MockCache) Get(ctx context.Context, key string) (string, bool) {
	val, exists := m.data[key]
	return val, exists
}

func (m *MockCache) Set(ctx context.Context, key, value string, ttl time.Duration) {
	m.data[key] = value
}

type MockRepository struct {
	deputados []domain.Deputado
	err       error
}

func (m *MockRepository) UpsertDeputados(ctx context.Context, deps []domain.Deputado) error {
	if m.err != nil {
		return m.err
	}
	m.deputados = deps
	return nil
}

func (m *MockRepository) ListFromCache(ctx context.Context, limit int) ([]domain.Deputado, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.deputados, nil
}

func TestDeputadosService_ListarDeputados(t *testing.T) {
	tests := []struct {
		name           string
		partido        string
		uf             string
		nome           string
		mockDeputados  []domain.Deputado
		mockError      error
		cacheData      map[string]string
		expectedSource string
		expectError    bool
	}{
		{
			name:    "sucesso busca na API",
			partido: "PT",
			uf:      "SP",
			nome:    "",
			mockDeputados: []domain.Deputado{
				{ID: 1, Nome: "João Silva", Partido: "PT", UF: "SP"},
				{ID: 2, Nome: "Maria Santos", Partido: "PT", UF: "SP"},
			},
			expectedSource: "api",
			expectError:    false,
		},
		{
			name:           "sucesso dados do cache",
			partido:        "PSDB",
			uf:             "RJ",
			nome:           "",
			expectedSource: "cache",
			cacheData: map[string]string{
				`deputados:{"n":"","p":"PSDB","u":"RJ"}`: `[{"id":3,"nome":"Pedro Oliveira","siglaPartido":"PSDB","siglaUf":"RJ"}]`,
			},
			expectError: false,
		},
		{
			name:        "erro na API",
			partido:     "PL",
			uf:          "MG",
			nome:        "",
			mockError:   errors.New("erro de rede"),
			expectError: true,
		},
		{
			name:    "filtro por nome",
			partido: "",
			uf:      "",
			nome:    "João",
			mockDeputados: []domain.Deputado{
				{ID: 1, Nome: "João Silva", Partido: "PT", UF: "SP"},
			},
			expectedSource: "api",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockClient := &MockCamaraClient{
				deputados: tt.mockDeputados,
				err:       tt.mockError,
			}

			mockCache := NewMockCache()
			if tt.cacheData != nil {
				for k, v := range tt.cacheData {
					mockCache.data[k] = v
				}
			}

			mockRepo := &MockRepository{}

			service := NewDeputadosService(mockClient, mockCache, mockRepo)

			// Execute
			ctx := context.Background()
			result, source, err := service.ListarDeputados(ctx, tt.partido, tt.uf, tt.nome)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("esperava erro mas não recebeu")
				}
				return
			}

			if err != nil {
				t.Fatalf("não esperava erro: %v", err)
			}

			if source != tt.expectedSource {
				t.Errorf("source esperado: %s, recebido: %s", tt.expectedSource, source)
			}

			if tt.expectedSource == "api" && len(result) != len(tt.mockDeputados) {
				t.Errorf("quantidade esperada: %d, recebida: %d", len(tt.mockDeputados), len(result))
			}
		})
	}
}

func TestDeputadosService_BuscarDeputadoPorID(t *testing.T) {
	tests := []struct {
		name           string
		deputadoID     string
		mockDeputado   *domain.Deputado
		mockError      error
		cacheData      map[string]string
		expectedSource string
		expectError    bool
	}{
		{
			name:       "sucesso busca na API",
			deputadoID: "123",
			mockDeputado: &domain.Deputado{
				ID:      123,
				Nome:    "João Silva",
				Partido: "PT",
				UF:      "SP",
			},
			expectedSource: "api",
			expectError:    false,
		},
		{
			name:           "sucesso dados do cache",
			deputadoID:     "456",
			expectedSource: "cache",
			cacheData: map[string]string{
				"deputado:456": `{"id":456,"nome":"Maria Santos","siglaPartido":"PSDB","siglaUf":"RJ"}`,
			},
			expectError: false,
		},
		{
			name:        "erro na API - deputado não encontrado",
			deputadoID:  "999",
			mockError:   errors.New("deputado não encontrado"),
			expectError: true,
		},
		{
			name:        "ID inválido - string vazia",
			deputadoID:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockClient := &MockCamaraClient{
				deputado: tt.mockDeputado,
				err:      tt.mockError,
			}

			mockCache := NewMockCache()
			if tt.cacheData != nil {
				for k, v := range tt.cacheData {
					mockCache.data[k] = v
				}
			}

			mockRepo := &MockRepository{}

			service := NewDeputadosService(mockClient, mockCache, mockRepo)

			// Execute
			ctx := context.Background()
			result, source, err := service.BuscarDeputadoPorID(ctx, tt.deputadoID)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("esperava erro mas não recebeu")
				}
				return
			}

			if err != nil {
				t.Fatalf("não esperava erro: %v", err)
			}

			if source != tt.expectedSource {
				t.Errorf("source esperado: %s, recebido: %s", tt.expectedSource, source)
			}

			if result == nil {
				t.Errorf("resultado não deveria ser nil")
			}
		})
	}
}

func TestDeputadosService_ListarDespesas(t *testing.T) {
	tests := []struct {
		name          string
		deputadoID    string
		ano           string
		mockDespesas  []domain.Despesa
		mockError     error
		expectError   bool
		expectedCount int
	}{
		{
			name:       "sucesso busca despesas",
			deputadoID: "123",
			ano:        "2024",
			mockDespesas: []domain.Despesa{
				{Ano: 2024, Mes: 1, ValorLiquido: 150.0, TipoDespesa: "COMBUSTÍVEL"},
				{Ano: 2024, Mes: 2, ValorLiquido: 300.0, TipoDespesa: "ALIMENTAÇÃO"},
			},
			expectError:   false,
			expectedCount: 2,
		},
		{
			name:        "erro na API",
			deputadoID:  "123",
			ano:         "2024",
			mockError:   errors.New("erro de rede"),
			expectError: true,
		},
		{
			name:          "nenhuma despesa encontrada",
			deputadoID:    "123",
			ano:           "2024",
			mockDespesas:  []domain.Despesa{},
			expectError:   false,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockClient := &MockCamaraClient{
				despesas: tt.mockDespesas,
				err:      tt.mockError,
			}

			mockCache := NewMockCache()
			mockRepo := &MockRepository{}

			service := NewDeputadosService(mockClient, mockCache, mockRepo)

			// Execute
			ctx := context.Background()
			result, source, err := service.ListarDespesas(ctx, tt.deputadoID, tt.ano)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("esperava erro mas não recebeu")
				}
				return
			}

			if err != nil {
				t.Fatalf("não esperava erro: %v", err)
			}

			if len(result) != tt.expectedCount {
				t.Errorf("quantidade esperada: %d, recebida: %d", tt.expectedCount, len(result))
			}

			// Para despesas, deve sempre vir da API (por enquanto)
			if source != "api" {
				t.Errorf("source deveria ser 'api', recebido: %s", source)
			}
		})
	}
}
