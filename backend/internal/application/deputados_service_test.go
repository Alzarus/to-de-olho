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
	deputados    []domain.Deputado
	err          error
	upsertedData []domain.Deputado
}

func (m *MockRepository) UpsertDeputados(ctx context.Context, deps []domain.Deputado) error {
	if m.err != nil {
		return m.err
	}
	m.upsertedData = deps
	return nil
}

func (m *MockRepository) ListFromCache(ctx context.Context, limit int) ([]domain.Deputado, error) {
	if m.err != nil {
		return nil, m.err
	}
	if len(m.deputados) > limit {
		return m.deputados[:limit], nil
	}
	return m.deputados, nil
}

type MockDespesaRepository struct {
	despesas    []domain.Despesa
	err         error
	upsertCalls []UpsertCall
}

type UpsertCall struct {
	DeputadoID int
	Ano        int
	Despesas   []domain.Despesa
}

func (m *MockDespesaRepository) UpsertDespesas(ctx context.Context, deputadoID int, ano int, despesas []domain.Despesa) error {
	if m.err != nil {
		return m.err
	}
	m.upsertCalls = append(m.upsertCalls, UpsertCall{
		DeputadoID: deputadoID,
		Ano:        ano,
		Despesas:   despesas,
	})
	return nil
}

func (m *MockDespesaRepository) ListDespesasByDeputadoAno(ctx context.Context, deputadoID int, ano int) ([]domain.Despesa, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.despesas, nil
}

func boolPtr(v bool) *bool {
	return &v
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
			cacheData: func() map[string]string {
				m := make(map[string]string)
				// Usar a função centralizada para gerar a chave
				m[BuildDeputadosCacheKey("PSDB", "RJ", "")] = `[{"id":3,"nome":"Pedro Oliveira","siglaPartido":"PSDB","siglaUf":"RJ"}]`
				return m
			}(),
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
			name:      "fallback para repositório quando API falha",
			partido:   "PSL",
			uf:        "RS",
			nome:      "",
			mockError: errors.New("API indisponível"),
			// Mock repositório terá dados de fallback
			mockDeputados: []domain.Deputado{
				{ID: 99, Nome: "Deputado Fallback", Partido: "PSL", UF: "RS"},
			},
			expectedSource: "fallback-db",
			expectError:    false,
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
			// Se esperamos fallback, configurar dados no repositório
			if tt.expectedSource == "fallback-db" {
				mockRepo.deputados = tt.mockDeputados
			}

			service := NewDeputadosService(mockClient, mockCache, mockRepo, &MockDespesaRepository{})

			// Execute
			ctx := context.Background()
			result, source, err := service.ListarDeputados(ctx, tt.partido, tt.uf, tt.nome) // Assert
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

			service := NewDeputadosService(mockClient, mockCache, mockRepo, &MockDespesaRepository{})

			// Execute
			ctx := context.Background()
			result, source, err := service.BuscarDeputadoPorID(ctx, tt.deputadoID) // Assert
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
		name           string
		deputadoID     string
		ano            string
		mockDespesas   []domain.Despesa
		mockError      error
		cacheData      map[string]string
		repoDespesas   []domain.Despesa
		repoError      error
		expectedSource string
		expectError    bool
		expectedCount  int
		ctx            context.Context
		expectUpsert   *bool
	}{
		{
			name:       "sucesso busca despesas na API",
			deputadoID: "123",
			ano:        "2024",
			mockDespesas: []domain.Despesa{
				{Ano: 2024, Mes: 1, ValorLiquido: 150.0, TipoDespesa: "COMBUSTÍVEL"},
				{Ano: 2024, Mes: 2, ValorLiquido: 300.0, TipoDespesa: "ALIMENTAÇÃO"},
			},
			expectedSource: "api",
			expectError:    false,
			expectedCount:  2,
		},
		{
			name:       "sucesso dados do cache",
			deputadoID: "456",
			ano:        "2024",
			cacheData: map[string]string{
				"despesas:456:2024": `[{"ano":2024,"mes":3,"valorLiquido":200.0,"tipoDespesa":"COMBUSTÍVEL"}]`,
			},
			expectedSource: "cache",
			expectError:    false,
			expectedCount:  1,
		},
		{
			name:           "API indisponível retorna dados vazios do banco",
			deputadoID:     "123",
			ano:            "2024",
			mockError:      errors.New("erro de rede"),
			repoDespesas:   []domain.Despesa{},
			expectedSource: "database_fallback",
			expectError:    false,
			expectedCount:  0,
		},
		{
			name:           "nenhuma despesa encontrada",
			deputadoID:     "123",
			ano:            "2024",
			mockDespesas:   []domain.Despesa{},
			expectedSource: "api",
			expectError:    false,
			expectedCount:  0,
		},
		{
			name:       "usa dados do banco quando disponíveis",
			deputadoID: "789",
			ano:        "2023",
			repoDespesas: []domain.Despesa{
				{Ano: 2023, Mes: 5, ValorLiquido: 500.0, TipoDespesa: "DIVULGAÇÃO DA ATIVIDADE PARLAMENTAR"},
			},
			mockError:      errors.New("API não deveria ser chamada"),
			expectedSource: "database",
			expectError:    false,
			expectedCount:  1,
		},
		{
			name:       "força remoto ignora cache e banco",
			deputadoID: "321",
			ano:        "2022",
			mockDespesas: []domain.Despesa{
				{Ano: 2022, Mes: 7, ValorLiquido: 100.0, TipoDespesa: "COMBUSTÍVEL"},
			},
			cacheData: map[string]string{
				"despesas:321:2022": `[{"ano":2022,"mes":1,"valorLiquido":999.0,"tipoDespesa":"CACHE"}]`,
			},
			repoDespesas: []domain.Despesa{
				{Ano: 2022, Mes: 2, ValorLiquido: 200.0, TipoDespesa: "BD"},
			},
			expectedSource: "api",
			expectError:    false,
			expectedCount:  1,
			ctx:            domain.WithForceDespesaRemote(context.Background()),
			expectUpsert:   boolPtr(true),
		},
		{
			name:       "skip persist evita upsert de despesas",
			deputadoID: "654",
			ano:        "2021",
			mockDespesas: []domain.Despesa{
				{Ano: 2021, Mes: 3, ValorLiquido: 300.0, TipoDespesa: "PASSAGENS"},
			},
			expectedSource: "api",
			expectError:    false,
			expectedCount:  1,
			ctx:            domain.WithSkipDespesaPersist(context.Background()),
			expectUpsert:   boolPtr(false),
		},
		{
			name:        "erro quando banco e API falham",
			deputadoID:  "555",
			ano:         "2022",
			mockError:   errors.New("API indisponível"),
			repoError:   errors.New("DB indisponível"),
			expectError: true,
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
			if tt.cacheData != nil {
				for k, v := range tt.cacheData {
					mockCache.data[k] = v
				}
			}

			mockRepo := &MockRepository{}

			mockDespesaRepo := &MockDespesaRepository{
				despesas: tt.repoDespesas,
				err:      tt.repoError,
			}

			service := NewDeputadosService(mockClient, mockCache, mockRepo, mockDespesaRepo)

			// Execute
			ctx := context.Background()
			if tt.ctx != nil {
				ctx = tt.ctx
			}
			result, source, err := service.ListarDespesas(ctx, tt.deputadoID, tt.ano) // Assert
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

			if source != tt.expectedSource {
				t.Errorf("source esperado: %s, recebido: %s", tt.expectedSource, source)
			}

			if tt.expectUpsert != nil {
				realUpsert := len(mockDespesaRepo.upsertCalls) > 0
				if *tt.expectUpsert && !realUpsert {
					t.Errorf("esperava persistência de despesas, mas não ocorreu")
				}
				if !*tt.expectUpsert && realUpsert {
					t.Errorf("não esperava persistência de despesas, mas ocorreu")
				}
			}
		})
	}
}
