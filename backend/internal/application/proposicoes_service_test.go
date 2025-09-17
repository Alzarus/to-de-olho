package application

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"to-de-olho-backend/internal/domain"
)

// Mock implementations para testes

type mockProposicaoPort struct {
	proposicoes map[int]*domain.Proposicao
	shouldError bool
	errorMsg    string
}

func (m *mockProposicaoPort) FetchProposicoes(ctx context.Context, filtros *domain.ProposicaoFilter) ([]domain.Proposicao, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}

	var result []domain.Proposicao
	for _, prop := range m.proposicoes {
		result = append(result, *prop)
	}

	return result, nil
}

func (m *mockProposicaoPort) FetchProposicaoPorID(ctx context.Context, id int) (*domain.Proposicao, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}

	if prop, exists := m.proposicoes[id]; exists {
		return prop, nil
	}

	return nil, domain.ErrProposicaoNaoEncontrada
}

type mockCachePort struct {
	data map[string]string
}

func (m *mockCachePort) Get(ctx context.Context, key string) (string, bool) {
	if m.data == nil {
		m.data = make(map[string]string)
	}

	value, exists := m.data[key]
	return value, exists
}

func (m *mockCachePort) Set(ctx context.Context, key, value string, ttl time.Duration) {
	if m.data == nil {
		m.data = make(map[string]string)
	}

	m.data[key] = value
}

type mockProposicaoRepositoryPort struct {
	proposicoes map[int]*domain.Proposicao
	shouldError bool
	errorMsg    string
}

func (m *mockProposicaoRepositoryPort) ListProposicoes(ctx context.Context, filtros *domain.ProposicaoFilter) ([]domain.Proposicao, int, error) {
	if m.shouldError {
		return nil, 0, errors.New(m.errorMsg)
	}

	var result []domain.Proposicao
	for _, prop := range m.proposicoes {
		result = append(result, *prop)
	}

	return result, len(result), nil
}

func (m *mockProposicaoRepositoryPort) GetProposicaoPorID(ctx context.Context, id int) (*domain.Proposicao, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}

	if prop, exists := m.proposicoes[id]; exists {
		return prop, nil
	}

	return nil, domain.ErrProposicaoNaoEncontrada
}

func (m *mockProposicaoRepositoryPort) UpsertProposicoes(ctx context.Context, proposicoes []domain.Proposicao) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}

	if m.proposicoes == nil {
		m.proposicoes = make(map[int]*domain.Proposicao)
	}

	for _, prop := range proposicoes {
		propCopy := prop
		m.proposicoes[prop.ID] = &propCopy
	}

	return nil
}

func createTestProposicao(id int, siglaTipo string, numero int, ano int, ementa string) *domain.Proposicao {
	return &domain.Proposicao{
		ID:               id,
		SiglaTipo:        siglaTipo,
		Numero:           numero,
		Ano:              ano,
		Ementa:           ementa,
		DataApresentacao: "2024-01-15T10:00:00",
		DescricaoTipo:    "Projeto de Lei",
	}
}

func TestProposicoesService_ListarProposicoes(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	tests := []struct {
		name           string
		setupMocks     func() (*mockProposicaoPort, *mockCachePort, *mockProposicaoRepositoryPort)
		filtros        *domain.ProposicaoFilter
		expectedTotal  int
		expectedSource string
		expectError    bool
	}{
		{
			name: "busca com sucesso da API",
			setupMocks: func() (*mockProposicaoPort, *mockCachePort, *mockProposicaoRepositoryPort) {
				mockClient := &mockProposicaoPort{
					proposicoes: map[int]*domain.Proposicao{
						1: createTestProposicao(1, "PL", 1234, 2024, "Teste PL 1"),
						2: createTestProposicao(2, "PEC", 15, 2024, "Teste PEC 1"),
					},
				}
				mockCache := &mockCachePort{}
				mockRepo := &mockProposicaoRepositoryPort{}
				return mockClient, mockCache, mockRepo
			},
			filtros: &domain.ProposicaoFilter{
				Limite: 20,
				Pagina: 1,
			},
			expectedTotal:  2,
			expectedSource: "api",
			expectError:    false,
		},
		{
			name: "busca do cache quando disponível",
			setupMocks: func() (*mockProposicaoPort, *mockCachePort, *mockProposicaoRepositoryPort) {
				mockClient := &mockProposicaoPort{}
				mockCache := &mockCachePort{}
				mockRepo := &mockProposicaoRepositoryPort{}

				// Pré-popular cache
				cacheData := struct {
					Proposicoes []domain.Proposicao `json:"proposicoes"`
					Total       int                 `json:"total"`
				}{
					Proposicoes: []domain.Proposicao{
						*createTestProposicao(1, "PL", 1234, 2024, "Cache PL 1"),
					},
					Total: 1,
				}

				if cacheBytes, err := json.Marshal(cacheData); err == nil {
					// Criar os filtros para gerar a chave corretamente
					filtros := &domain.ProposicaoFilter{
						Limite: 20,
						Pagina: 1,
					}
					filtros.SetDefaults() // Aplicar padrões para gerar chave correta
					cacheKey := BuildProposicoesCacheKey(filtros)
					mockCache.Set(context.Background(), cacheKey, string(cacheBytes), time.Minute)
				}

				return mockClient, mockCache, mockRepo
			},
			filtros: &domain.ProposicaoFilter{
				Limite: 20,
				Pagina: 1,
			},
			expectedTotal:  1,
			expectedSource: "cache",
			expectError:    false,
		},
		{
			name: "fallback para repositório quando API falha",
			setupMocks: func() (*mockProposicaoPort, *mockCachePort, *mockProposicaoRepositoryPort) {
				mockClient := &mockProposicaoPort{
					shouldError: true,
					errorMsg:    "API indisponível",
				}
				mockCache := &mockCachePort{}
				mockRepo := &mockProposicaoRepositoryPort{
					proposicoes: map[int]*domain.Proposicao{
						1: createTestProposicao(1, "PL", 5678, 2024, "Repo PL 1"),
					},
				}
				return mockClient, mockCache, mockRepo
			},
			filtros: &domain.ProposicaoFilter{
				Limite: 20,
				Pagina: 1,
			},
			expectedTotal:  1,
			expectedSource: "repository",
			expectError:    false,
		},
		{
			name: "erro quando filtros são inválidos",
			setupMocks: func() (*mockProposicaoPort, *mockCachePort, *mockProposicaoRepositoryPort) {
				return &mockProposicaoPort{}, &mockCachePort{}, &mockProposicaoRepositoryPort{}
			},
			filtros: &domain.ProposicaoFilter{
				Limite: 200, // Limite muito alto
				Pagina: 1,
			},
			expectedTotal:  0,
			expectedSource: "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient, mockCache, mockRepo := tt.setupMocks()

			service := NewProposicoesService(mockClient, mockCache, mockRepo, logger)

			proposicoes, total, source, err := service.ListarProposicoes(context.Background(), tt.filtros)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if !tt.expectError {
				if total != tt.expectedTotal {
					t.Errorf("Total = %d, want %d", total, tt.expectedTotal)
				}

				if source != tt.expectedSource {
					t.Errorf("Source = %s, want %s", source, tt.expectedSource)
				}

				if len(proposicoes) != tt.expectedTotal {
					t.Errorf("Proposicoes length = %d, want %d", len(proposicoes), tt.expectedTotal)
				}
			}
		})
	}
}

func TestProposicoesService_BuscarProposicaoPorID(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	tests := []struct {
		name           string
		setupMocks     func() (*mockProposicaoPort, *mockCachePort, *mockProposicaoRepositoryPort)
		id             int
		expectedSource string
		expectError    bool
		expectedError  error
	}{
		{
			name: "busca com sucesso da API",
			setupMocks: func() (*mockProposicaoPort, *mockCachePort, *mockProposicaoRepositoryPort) {
				mockClient := &mockProposicaoPort{
					proposicoes: map[int]*domain.Proposicao{
						123: createTestProposicao(123, "PL", 1234, 2024, "API PL 1"),
					},
				}
				mockCache := &mockCachePort{}
				mockRepo := &mockProposicaoRepositoryPort{}
				return mockClient, mockCache, mockRepo
			},
			id:             123,
			expectedSource: "api",
			expectError:    false,
		},
		{
			name: "busca do cache quando disponível",
			setupMocks: func() (*mockProposicaoPort, *mockCachePort, *mockProposicaoRepositoryPort) {
				mockClient := &mockProposicaoPort{}
				mockCache := &mockCachePort{}
				mockRepo := &mockProposicaoRepositoryPort{}

				// Pré-popular cache
				proposicao := createTestProposicao(456, "PEC", 15, 2024, "Cache PEC 1")
				if cacheBytes, err := json.Marshal(proposicao); err == nil {
					mockCache.Set(context.Background(), "proposicao:456", string(cacheBytes), time.Minute)
				}

				return mockClient, mockCache, mockRepo
			},
			id:             456,
			expectedSource: "cache",
			expectError:    false,
		},
		{
			name: "fallback para repositório quando API falha",
			setupMocks: func() (*mockProposicaoPort, *mockCachePort, *mockProposicaoRepositoryPort) {
				mockClient := &mockProposicaoPort{
					shouldError: true,
					errorMsg:    "API indisponível",
				}
				mockCache := &mockCachePort{}
				mockRepo := &mockProposicaoRepositoryPort{
					proposicoes: map[int]*domain.Proposicao{
						789: createTestProposicao(789, "MPV", 1185, 2024, "Repo MPV 1"),
					},
				}
				return mockClient, mockCache, mockRepo
			},
			id:             789,
			expectedSource: "repository",
			expectError:    false,
		},
		{
			name: "erro quando ID é inválido",
			setupMocks: func() (*mockProposicaoPort, *mockCachePort, *mockProposicaoRepositoryPort) {
				return &mockProposicaoPort{}, &mockCachePort{}, &mockProposicaoRepositoryPort{}
			},
			id:            0,
			expectError:   true,
			expectedError: domain.ErrProposicaoIDInvalido,
		},
		{
			name: "erro quando proposição não existe",
			setupMocks: func() (*mockProposicaoPort, *mockCachePort, *mockProposicaoRepositoryPort) {
				mockClient := &mockProposicaoPort{proposicoes: make(map[int]*domain.Proposicao)}
				mockCache := &mockCachePort{}
				mockRepo := &mockProposicaoRepositoryPort{proposicoes: make(map[int]*domain.Proposicao)}
				return mockClient, mockCache, mockRepo
			},
			id:          999,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient, mockCache, mockRepo := tt.setupMocks()

			service := NewProposicoesService(mockClient, mockCache, mockRepo, logger)

			proposicao, source, err := service.BuscarProposicaoPorID(context.Background(), tt.id)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if tt.expectedError != nil && err != tt.expectedError {
				t.Errorf("Expected error %v but got %v", tt.expectedError, err)
				return
			}

			if !tt.expectError {
				if proposicao == nil {
					t.Errorf("Expected proposicao but got nil")
					return
				}

				if proposicao.ID != tt.id {
					t.Errorf("Proposicao ID = %d, want %d", proposicao.ID, tt.id)
				}

				if source != tt.expectedSource {
					t.Errorf("Source = %s, want %s", source, tt.expectedSource)
				}
			}
		})
	}
}

func TestBuildProposicoesCacheKey(t *testing.T) {
	tests := []struct {
		name     string
		filtros  *domain.ProposicaoFilter
		expected string
	}{
		{
			name: "filtro básico",
			filtros: &domain.ProposicaoFilter{
				Pagina:     1,
				Limite:     20,
				Ordem:      "DESC",
				OrdenarPor: "dataApresentacao",
			},
			expected: "proposicoes:p1:l20:oDESC:opdataApresentacao",
		},
		{
			name: "filtro com tipo",
			filtros: &domain.ProposicaoFilter{
				SiglaTipo:  "PL",
				Pagina:     2,
				Limite:     50,
				Ordem:      "ASC",
				OrdenarPor: "numero",
			},
			expected: "proposicoes:p2:l50:oASC:opnumero:stPL",
		},
		{
			name: "filtro completo",
			filtros: &domain.ProposicaoFilter{
				SiglaTipo:         "PEC",
				Ano:               &[]int{2024}[0],
				Numero:            &[]int{15}[0],
				CodSituacao:       &[]int{100}[0],
				SiglaUfAutor:      "SP",
				SiglaPartidoAutor: "PT",
				Tema:              "saúde",
				Pagina:            3,
				Limite:            25,
				Ordem:             "DESC",
				OrdenarPor:        "ano",
			},
			expected: "proposicoes:p3:l25:oDESC:opano:stPEC:a2024:n15:cs100:ufSP:ptPT:tsaúde",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildProposicoesCacheKey(tt.filtros)
			if result != tt.expected {
				t.Errorf("buildProposicoesCacheKey() = %s, want %s", result, tt.expected)
			}
		})
	}
}
