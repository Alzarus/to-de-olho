package application

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"to-de-olho-backend/internal/domain"
)

var testLogger = slog.New(slog.NewTextHandler(os.Stdout, nil))

// Mock do Cache para testes
type MockCacheAnalytics struct {
	data map[string]string
}

func NewMockCacheAnalytics() *MockCacheAnalytics {
	return &MockCacheAnalytics{
		data: make(map[string]string),
	}
}

func (m *MockCacheAnalytics) Get(ctx context.Context, key string) (string, bool) {
	value, exists := m.data[key]
	return value, exists
}

func (m *MockCacheAnalytics) Set(ctx context.Context, key, value string, ttl time.Duration) {
	m.data[key] = value
}

func (m *MockCacheAnalytics) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

// Mock do DeputadoRepository para testes de analytics
type MockDeputadoRepository struct {
	deputados []domain.Deputado
}

func NewMockDeputadoRepository() *MockDeputadoRepository {
	return &MockDeputadoRepository{
		deputados: []domain.Deputado{
			{ID: 1, Nome: "Deputado A", Partido: "PT", UF: "SP"},
			{ID: 2, Nome: "Deputado B", Partido: "PSDB", UF: "RJ"},
			{ID: 3, Nome: "Deputado C", Partido: "PT", UF: "MG"},
		},
	}
}

func NewMockDeputadoRepositoryWithCount(count int) *MockDeputadoRepository {
	deputados := make([]domain.Deputado, count)
	partidos := []string{"PT", "PSDB", "PL", "MDB", "UNIÃO", "PDT", "PSB", "REPUBLICANOS", "PSOL", "PCdoB"}
	ufs := []string{"SP", "RJ", "MG", "BA", "PR", "RS", "PE", "CE", "PA", "SC"}

	for i := 0; i < count; i++ {
		deputados[i] = domain.Deputado{
			ID:      i + 1,
			Nome:    fmt.Sprintf("Deputado %d", i+1),
			Partido: partidos[i%len(partidos)],
			UF:      ufs[i%len(ufs)],
		}
	}

	return &MockDeputadoRepository{deputados: deputados}
}

func (m *MockDeputadoRepository) ListFromCache(ctx context.Context, limit int) ([]domain.Deputado, error) {
	if limit > 0 && limit < len(m.deputados) {
		return m.deputados[:limit], nil
	}
	return m.deputados, nil
}

func (m *MockDeputadoRepository) UpsertDeputados(ctx context.Context, deps []domain.Deputado) error {
	m.deputados = append(m.deputados, deps...)
	return nil
}

// Mock do ProposicaoRepository para testes de analytics
type MockProposicaoRepository struct {
	proposicoes []domain.Proposicao
}

func NewMockProposicaoRepository() *MockProposicaoRepository {
	return &MockProposicaoRepository{
		proposicoes: []domain.Proposicao{
			{ID: 1, Numero: 123, SiglaTipo: "PL", Ano: 2025},
			{ID: 2, Numero: 456, SiglaTipo: "PEC", Ano: 2025},
		},
	}
}

func (m *MockProposicaoRepository) ListProposicoes(ctx context.Context, filtros *domain.ProposicaoFilter) ([]domain.Proposicao, int, error) {
	return m.proposicoes, len(m.proposicoes), nil
}

func (m *MockProposicaoRepository) UpsertProposicoes(ctx context.Context, proposicoes []domain.Proposicao) error {
	m.proposicoes = append(m.proposicoes, proposicoes...)
	return nil
}

func TestAnalyticsService_GetRankingGastos(t *testing.T) {
	tests := []struct {
		name         string
		ano          int
		limite       int
		cacheData    string
		wantSource   string
		wantError    bool
		wantPosicoes []int // Verificar ordem do ranking
	}{
		{
			name:         "Sucesso - cálculo novo",
			ano:          2024,
			limite:       3,
			wantSource:   "computed",
			wantError:    false,
			wantPosicoes: []int{1, 2, 3},
		},
		{
			name:       "Sucesso - dados do cache",
			ano:        2024,
			limite:     2,
			cacheData:  `{"ano":2024,"total_geral":10000,"media_gastos":3333.33,"deputados":[{"id":2,"nome":"Deputado B","posicao":1}],"ultima_atualizacao":"2024-01-01T00:00:00Z"}`,
			wantSource: "cache",
			wantError:  false,
		},
		{
			name:       "Sucesso - limite aplicado",
			ano:        2024,
			limite:     1,
			wantSource: "computed",
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cache := NewMockCacheAnalytics()
			if tt.cacheData != "" {
				cacheKey := "ranking:gastos:2024:2"
				cache.Set(context.Background(), cacheKey, tt.cacheData, time.Hour)
			}

			deputadoRepo := NewMockDeputadoRepository()
			proposicaoRepo := NewMockProposicaoRepository()

			service := NewAnalyticsService(
				deputadoRepo,
				proposicaoRepo,
				cache,
				testLogger,
			)

			// Act
			ranking, source, err := service.GetRankingGastos(context.Background(), tt.ano, tt.limite)

			// Assert
			if tt.wantError {
				if err == nil {
					t.Errorf("GetRankingGastos() error = %v, wantErr %v", err, tt.wantError)
				}
				return
			}

			if err != nil {
				t.Errorf("GetRankingGastos() error = %v, wantErr %v", err, tt.wantError)
				return
			}

			if source != tt.wantSource {
				t.Errorf("GetRankingGastos() source = %v, want %v", source, tt.wantSource)
			}

			if ranking == nil {
				t.Error("GetRankingGastos() ranking is nil")
				return
			}

			if ranking.Ano != tt.ano {
				t.Errorf("GetRankingGastos() ano = %v, want %v", ranking.Ano, tt.ano)
			}

			if tt.limite > 0 && len(ranking.Deputados) > tt.limite {
				t.Errorf("GetRankingGastos() deputados count = %v, want <= %v", len(ranking.Deputados), tt.limite)
			}

			// Verificar se posições estão corretas (para dados computados)
			if tt.wantSource == "computed" && len(tt.wantPosicoes) > 0 {
				if len(ranking.Deputados) < len(tt.wantPosicoes) {
					t.Errorf("GetRankingGastos() deputados count = %v, want >= %v", len(ranking.Deputados), len(tt.wantPosicoes))
					return
				}

				for i, expectedPos := range tt.wantPosicoes {
					if ranking.Deputados[i].Posicao != expectedPos {
						t.Errorf("GetRankingGastos() deputado[%d].Posicao = %v, want %v", i, ranking.Deputados[i].Posicao, expectedPos)
					}
				}
			}
		})
	}
}

func TestAnalyticsService_GetRankingProposicoes(t *testing.T) {
	tests := []struct {
		name       string
		ano        int
		limite     int
		cacheData  string
		wantSource string
		wantError  bool
	}{
		{
			name:       "Sucesso - cálculo novo",
			ano:        2024,
			limite:     3,
			wantSource: "computed",
			wantError:  false,
		},
		{
			name:       "Sucesso - dados do cache",
			ano:        2024,
			limite:     2,
			cacheData:  `{"ano":2024,"total_geral":100,"media_proposicoes":33.33,"deputados":[{"id":1,"nome":"Deputado A","posicao":1}],"ultima_atualizacao":"2024-01-01T00:00:00Z"}`,
			wantSource: "cache",
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cache := NewMockCacheAnalytics()
			if tt.cacheData != "" {
				cacheKey := "ranking:proposicoes:2024:2"
				cache.Set(context.Background(), cacheKey, tt.cacheData, time.Hour)
			}

			deputadoRepo := NewMockDeputadoRepository()
			proposicaoRepo := NewMockProposicaoRepository()

			service := NewAnalyticsService(
				deputadoRepo,
				proposicaoRepo,
				cache,
				testLogger,
			)

			// Act
			ranking, source, err := service.GetRankingProposicoes(context.Background(), tt.ano, tt.limite)

			// Assert
			if tt.wantError {
				if err == nil {
					t.Errorf("GetRankingProposicoes() error = %v, wantErr %v", err, tt.wantError)
				}
				return
			}

			if err != nil {
				t.Errorf("GetRankingProposicoes() error = %v, wantErr %v", err, tt.wantError)
				return
			}

			if source != tt.wantSource {
				t.Errorf("GetRankingProposicoes() source = %v, want %v", source, tt.wantSource)
			}

			if ranking == nil {
				t.Error("GetRankingProposicoes() ranking is nil")
				return
			}

			if ranking.Ano != tt.ano {
				t.Errorf("GetRankingProposicoes() ano = %v, want %v", ranking.Ano, tt.ano)
			}

			if tt.limite > 0 && len(ranking.Deputados) > tt.limite {
				t.Errorf("GetRankingProposicoes() deputados count = %v, want <= %v", len(ranking.Deputados), tt.limite)
			}
		})
	}
}

func TestAnalyticsService_GetRankingPresenca(t *testing.T) {
	tests := []struct {
		name       string
		ano        int
		limite     int
		cacheData  string
		wantSource string
		wantError  bool
	}{
		{
			name:       "Sucesso - cálculo novo",
			ano:        2024,
			limite:     3,
			wantSource: "computed",
			wantError:  false,
		},
		{
			name:       "Sucesso - dados do cache",
			ano:        2024,
			limite:     1,
			cacheData:  `{"ano":2024,"total_sessoes":100,"media_presenca":85.5,"deputados":[{"id":1,"nome":"Deputado A","posicao":1}],"ultima_atualizacao":"2024-01-01T00:00:00Z"}`,
			wantSource: "cache",
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cache := NewMockCacheAnalytics()
			if tt.cacheData != "" {
				cacheKey := "ranking:presenca:2024:1"
				cache.Set(context.Background(), cacheKey, tt.cacheData, time.Hour)
			}

			deputadoRepo := NewMockDeputadoRepository()
			proposicaoRepo := NewMockProposicaoRepository()

			service := NewAnalyticsService(
				deputadoRepo,
				proposicaoRepo,
				cache,
				testLogger,
			)

			// Act
			ranking, source, err := service.GetRankingPresenca(context.Background(), tt.ano, tt.limite)

			// Assert
			if tt.wantError {
				if err == nil {
					t.Errorf("GetRankingPresenca() error = %v, wantErr %v", err, tt.wantError)
				}
				return
			}

			if err != nil {
				t.Errorf("GetRankingPresenca() error = %v, wantErr %v", err, tt.wantError)
				return
			}

			if source != tt.wantSource {
				t.Errorf("GetRankingPresenca() source = %v, want %v", source, tt.wantSource)
			}

			if ranking == nil {
				t.Error("GetRankingPresenca() ranking is nil")
				return
			}

			if ranking.Ano != tt.ano {
				t.Errorf("GetRankingPresenca() ano = %v, want %v", ranking.Ano, tt.ano)
			}

			if tt.limite > 0 && len(ranking.Deputados) > tt.limite {
				t.Errorf("GetRankingPresenca() deputados count = %v, want <= %v", len(ranking.Deputados), tt.limite)
			}

			// Verificar se percentual de presença está no range válido
			for i, deputado := range ranking.Deputados {
				if deputado.PercentualPresenca < 0 || deputado.PercentualPresenca > 100 {
					t.Errorf("GetRankingPresenca() deputado[%d].PercentualPresenca = %v, want between 0-100", i, deputado.PercentualPresenca)
				}
			}
		})
	}
}

func TestAnalyticsService_GetInsightsGerais(t *testing.T) {
	tests := []struct {
		name       string
		cacheData  string
		wantSource string
		wantError  bool
	}{
		{
			name:       "Sucesso - cálculo novo",
			wantSource: "computed",
			wantError:  false,
		},
		{
			name:       "Sucesso - dados do cache",
			cacheData:  `{"total_deputados":3,"total_gasto_ano":10000,"total_proposicoes_ano":1000,"media_gastos_deputado":3333.33,"partido_maior_gasto":"PT","uf_maior_gasto":"SP","ultima_atualizacao":"2024-01-01T00:00:00Z"}`,
			wantSource: "cache",
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cache := NewMockCacheAnalytics()
			if tt.cacheData != "" {
				cache.Set(context.Background(), "insights:gerais", tt.cacheData, time.Hour)
			}

			deputadoRepo := NewMockDeputadoRepository()
			proposicaoRepo := NewMockProposicaoRepository()

			service := NewAnalyticsService(
				deputadoRepo,
				proposicaoRepo,
				cache,
				testLogger,
			)

			// Act
			insights, source, err := service.GetInsightsGerais(context.Background())

			// Assert
			if tt.wantError {
				if err == nil {
					t.Errorf("GetInsightsGerais() error = %v, wantErr %v", err, tt.wantError)
				}
				return
			}

			if err != nil {
				t.Errorf("GetInsightsGerais() error = %v, wantErr %v", err, tt.wantError)
				return
			}

			if source != tt.wantSource {
				t.Errorf("GetInsightsGerais() source = %v, want %v", source, tt.wantSource)
			}

			if insights == nil {
				t.Error("GetInsightsGerais() insights is nil")
				return
			}

			if insights.TotalDeputados <= 0 {
				t.Errorf("GetInsightsGerais() TotalDeputados = %v, want > 0", insights.TotalDeputados)
			}

			if insights.MediaGastosDeputado < 0 {
				t.Errorf("GetInsightsGerais() MediaGastosDeputado = %v, want >= 0", insights.MediaGastosDeputado)
			}
		})
	}
}

func TestAnalyticsService_AtualizarRankings(t *testing.T) {
	tests := []struct {
		name      string
		wantError bool
	}{
		{
			name:      "Sucesso - atualização completa",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cache := NewMockCacheAnalytics()
			deputadoRepo := NewMockDeputadoRepository()
			proposicaoRepo := NewMockProposicaoRepository()

			service := NewAnalyticsService(
				deputadoRepo,
				proposicaoRepo,
				cache,
				testLogger,
			)

			// Act
			err := service.AtualizarRankings(context.Background())

			// Assert
			if tt.wantError {
				if err == nil {
					t.Errorf("AtualizarRankings() error = %v, wantErr %v", err, tt.wantError)
				}
				return
			}

			if err != nil {
				t.Errorf("AtualizarRankings() error = %v, wantErr %v", err, tt.wantError)
				return
			}
		})
	}
}

// Testes para funções auxiliares
func TestFindMaxKey(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]float64
		expected string
	}{
		{
			name:     "Map com valores únicos",
			input:    map[string]float64{"A": 10.0, "B": 20.0, "C": 5.0},
			expected: "B",
		},
		{
			name:     "Map vazio",
			input:    map[string]float64{},
			expected: "",
		},
		{
			name:     "Map com um elemento",
			input:    map[string]float64{"ÚNICO": 42.0},
			expected: "ÚNICO",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findMaxKey(tt.input)
			if result != tt.expected {
				t.Errorf("findMaxKey() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Benchmark para operações de ranking
func BenchmarkAnalyticsService_GetRankingGastos(b *testing.B) {
	cache := NewMockCacheAnalytics()
	deputadoRepo := NewMockDeputadoRepository()
	proposicaoRepo := NewMockProposicaoRepository()

	service := NewAnalyticsService(
		deputadoRepo,
		proposicaoRepo,
		cache,
		testLogger,
	)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := service.GetRankingGastos(ctx, 2024, 10)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark para stress test com muitos deputados
func BenchmarkAnalyticsService_GetRankingGastosStress(b *testing.B) {
	cache := NewMockCacheAnalytics()
	// Simular 513 deputados (número real da Câmara)
	deputadoRepo := NewMockDeputadoRepositoryWithCount(513)
	proposicaoRepo := NewMockProposicaoRepository()

	service := NewAnalyticsService(
		deputadoRepo,
		proposicaoRepo,
		cache,
		testLogger,
	)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := service.GetRankingGastos(ctx, 2024, 50)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Teste para verificar performance com timeout
func TestAnalyticsService_GetRankingGastosTimeout(t *testing.T) {
	cache := NewMockCacheAnalytics()
	// Muitos deputados para forçar timeout
	deputadoRepo := NewMockDeputadoRepositoryWithCount(1000)
	proposicaoRepo := NewMockProposicaoRepository()

	service := NewAnalyticsService(
		deputadoRepo,
		proposicaoRepo,
		cache,
		testLogger,
	)

	// Contexto com timeout muito curto
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	start := time.Now()
	ranking, source, err := service.GetRankingGastos(ctx, 2024, 10)
	duration := time.Since(start)

	// Deve funcionar mesmo com timeout porque usa dados simulados
	if err != nil {
		t.Logf("Erro com timeout: %v", err)
	}

	t.Logf("Performance com 1000 deputados: %v", duration)
	t.Logf("Source: %s", source)
	if ranking != nil {
		t.Logf("Deputados processados: %d", len(ranking.Deputados))
	}

	// Verificar se não demora mais que 1 segundo mesmo com muitos deputados
	if duration > time.Second {
		t.Errorf("GetRankingGastos demorou muito: %v (esperado < 1s)", duration)
	}
}
