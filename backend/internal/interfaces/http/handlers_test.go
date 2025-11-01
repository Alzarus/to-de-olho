package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"to-de-olho-backend/internal/application"
	"to-de-olho-backend/internal/domain"

	"github.com/gin-gonic/gin"
)

// Mock service para testes de handlers
type MockDeputadosService struct {
	deputados []domain.Deputado
	deputado  *domain.Deputado
	total     int
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
	return []domain.Despesa{}, m.source, nil
}

// Mock service para testes de proposições
type MockProposicoesService struct {
	proposicoes []domain.Proposicao
	proposicao  *domain.Proposicao
	total       int
	source      string
	err         error
}

func (m *MockProposicoesService) ListarProposicoes(ctx context.Context, filtros *domain.ProposicaoFilter) ([]domain.Proposicao, int, string, error) {
	if m.err != nil {
		return nil, 0, "", m.err
	}
	return m.proposicoes, m.total, m.source, nil
}

func (m *MockProposicoesService) BuscarProposicaoPorID(ctx context.Context, id int) (*domain.Proposicao, string, error) {
	if m.err != nil {
		return nil, "", m.err
	}
	return m.proposicao, m.source, nil
}

func createTestProposicao(id int, siglaTipo string, numero int, ano int, ementa string) *domain.Proposicao {
	return &domain.Proposicao{
		ID:               id,
		SiglaTipo:        siglaTipo,
		Numero:           numero,
		Ano:              ano,
		Ementa:           ementa,
		DataApresentacao: "2024-01-01",
		StatusProposicao: domain.StatusProposicao{
			DescricaoSituacao: "Em tramitação",
			DataHora:          "2024-01-01T10:00:00",
		},
	}
}

func TestGetProposicoesHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		mockService    *MockProposicoesService
		expectedCode   int
		checkSource    bool
		expectedSource string
	}{
		{
			name:        "success - proposições encontradas via API",
			queryParams: "?siglaTipo=PL&ano=2024",
			mockService: &MockProposicoesService{
				proposicoes: []domain.Proposicao{
					*createTestProposicao(1, "PL", 123, 2024, "Ementa do projeto de lei"),
				},
				total:  1,
				source: "api",
			},
			expectedCode:   http.StatusOK,
			checkSource:    true,
			expectedSource: "api",
		},
		{
			name:        "success - proposições encontradas via cache",
			queryParams: "?numero=123",
			mockService: &MockProposicoesService{
				proposicoes: []domain.Proposicao{
					*createTestProposicao(1, "PL", 123, 2024, "Ementa do projeto de lei"),
				},
				total:  1,
				source: "cache",
			},
			expectedCode:   http.StatusOK,
			checkSource:    true,
			expectedSource: "cache",
		},
		{
			name:        "success - nenhuma proposição encontrada",
			queryParams: "?siglaTipo=INEXISTENTE",
			mockService: &MockProposicoesService{
				proposicoes: []domain.Proposicao{},
				total:       0,
				source:      "api",
			},
			expectedCode: http.StatusOK,
		},
		{
			name:        "error - erro interno do serviço",
			queryParams: "?siglaTipo=PL",
			mockService: &MockProposicoesService{
				err: errors.New("erro interno"),
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/proposicoes", GetProposicoesHandler(tt.mockService))

			req := httptest.NewRequest("GET", "/proposicoes"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("código esperado: %d, recebido: %d", tt.expectedCode, w.Code)
			}

			if tt.expectedCode == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("erro ao decodificar resposta: %v", err)
				}

				if tt.checkSource {
					if response["source"] != tt.expectedSource {
						t.Errorf("source esperado: %s, recebido: %s", tt.expectedSource, response["source"])
					}
				}

				data := response["data"].([]interface{})
				if len(data) != len(tt.mockService.proposicoes) {
					t.Errorf("quantidade esperada: %d, recebida: %d", len(tt.mockService.proposicoes), len(data))
				}
			}
		})
	}
}

func TestGetProposicaoPorIDHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		proposicaoID   string
		mockService    *MockProposicoesService
		expectedCode   int
		checkSource    bool
		expectedSource string
	}{
		{
			name:         "success - proposição encontrada via API",
			proposicaoID: "123",
			mockService: &MockProposicoesService{
				proposicao: createTestProposicao(123, "PL", 456, 2024, "Ementa do projeto"),
				source:     "api",
			},
			expectedCode:   http.StatusOK,
			checkSource:    true,
			expectedSource: "api",
		},
		{
			name:         "success - proposição encontrada via cache",
			proposicaoID: "123",
			mockService: &MockProposicoesService{
				proposicao: createTestProposicao(123, "PL", 456, 2024, "Ementa do projeto"),
				source:     "cache",
			},
			expectedCode:   http.StatusOK,
			checkSource:    true,
			expectedSource: "cache",
		},
		{
			name:         "error - ID inválido",
			proposicaoID: "invalid",
			mockService:  &MockProposicoesService{},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "error - proposição não encontrada",
			proposicaoID: "999",
			mockService: &MockProposicoesService{
				err: domain.ErrProposicaoNaoEncontrada,
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "error - erro interno do serviço",
			proposicaoID: "123",
			mockService: &MockProposicoesService{
				err: errors.New("erro interno"),
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/proposicoes/:id", GetProposicaoPorIDHandler(tt.mockService))

			req := httptest.NewRequest("GET", "/proposicoes/"+tt.proposicaoID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("código esperado: %d, recebido: %d", tt.expectedCode, w.Code)
			}

			if tt.expectedCode == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("erro ao decodificar resposta: %v", err)
				}

				if tt.checkSource {
					if response["source"] != tt.expectedSource {
						t.Errorf("source esperado: %s, recebido: %s", tt.expectedSource, response["source"])
					}
				}

				data := response["data"].(map[string]interface{})
				if data["id"] != float64(tt.mockService.proposicao.ID) {
					t.Errorf("ID esperado: %d, recebido: %v", tt.mockService.proposicao.ID, data["id"])
				}
			}
		})
	}
}

// Testes dos deputados existentes
func TestGetDeputadosHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		mockService    *MockDeputadosService
		expectedCode   int
		checkSource    bool
		expectedSource string
	}{
		{
			name:        "success - deputados encontrados via API",
			queryParams: "?uf=SP&partido=PT",
			mockService: &MockDeputadosService{
				deputados: []domain.Deputado{
					{ID: 1, Nome: "Deputado 1", UF: "SP", Partido: "PT"},
				},
				total:  1,
				source: "api",
			},
			expectedCode:   http.StatusOK,
			checkSource:    true,
			expectedSource: "api",
		},
		{
			name:        "success - deputados encontrados via cache",
			queryParams: "?nome=João",
			mockService: &MockDeputadosService{
				deputados: []domain.Deputado{
					{ID: 1, Nome: "João Silva", UF: "RJ", Partido: "PSDB"},
				},
				total:  1,
				source: "cache",
			},
			expectedCode:   http.StatusOK,
			checkSource:    true,
			expectedSource: "cache",
		},
		{
			name:        "success - nenhum deputado encontrado",
			queryParams: "?uf=INEXISTENTE",
			mockService: &MockDeputadosService{
				deputados: []domain.Deputado{},
				total:     0,
				source:    "api",
			},
			expectedCode: http.StatusOK,
		},
		{
			name:        "error - erro interno do serviço",
			queryParams: "?uf=SP",
			mockService: &MockDeputadosService{
				err: errors.New("erro interno"),
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/deputados", GetDeputadosHandler(tt.mockService))

			req := httptest.NewRequest("GET", "/deputados"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("código esperado: %d, recebido: %d", tt.expectedCode, w.Code)
			}

			if tt.expectedCode == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("erro ao decodificar resposta: %v", err)
				}

				if tt.checkSource {
					if response["source"] != tt.expectedSource {
						t.Errorf("source esperado: %s, recebido: %s", tt.expectedSource, response["source"])
					}
				}

				data := response["data"].([]interface{})
				if len(data) != len(tt.mockService.deputados) {
					t.Errorf("quantidade esperada: %d, recebida: %d", len(tt.mockService.deputados), len(data))
				}
			}
		})
	}
}

func TestGetDeputadoByIDHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		deputadoID     string
		mockService    *MockDeputadosService
		expectedCode   int
		checkSource    bool
		expectedSource string
	}{
		{
			name:       "success - deputado encontrado via API",
			deputadoID: "123",
			mockService: &MockDeputadosService{
				deputado: &domain.Deputado{
					ID: 123, Nome: "João Silva", UF: "SP", Partido: "PT",
				},
				source: "api",
			},
			expectedCode:   http.StatusOK,
			checkSource:    true,
			expectedSource: "api",
		},
		{
			name:       "success - deputado encontrado via cache",
			deputadoID: "123",
			mockService: &MockDeputadosService{
				deputado: &domain.Deputado{
					ID: 123, Nome: "João Silva", UF: "SP", Partido: "PT",
				},
				source: "cache",
			},
			expectedCode:   http.StatusOK,
			checkSource:    true,
			expectedSource: "cache",
		},
		{
			name:       "error - ID inválido",
			deputadoID: "invalid",
			mockService: &MockDeputadosService{
				err: errors.New("ID do deputado é obrigatório"),
			},
			expectedCode: http.StatusNotFound, // Handler não valida ID, sempre retorna 404 para erro
		},
		{
			name:       "error - deputado não encontrado",
			deputadoID: "999",
			mockService: &MockDeputadosService{
				err: domain.ErrDeputadoNaoEncontrado,
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name:       "error - erro interno do serviço",
			deputadoID: "123",
			mockService: &MockDeputadosService{
				err: errors.New("erro interno"),
			},
			expectedCode: http.StatusNotFound, // Handler sempre retorna 404 para qualquer erro
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/deputados/:id", GetDeputadoByIDHandler(tt.mockService))

			req := httptest.NewRequest("GET", "/deputados/"+tt.deputadoID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("código esperado: %d, recebido: %d", tt.expectedCode, w.Code)
			}

			if tt.expectedCode == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("erro ao decodificar resposta: %v", err)
				}

				if tt.checkSource {
					if response["source"] != tt.expectedSource {
						t.Errorf("source esperado: %s, recebido: %s", tt.expectedSource, response["source"])
					}
				}

				data := response["data"].(map[string]interface{})
				if data["id"] != float64(tt.mockService.deputado.ID) {
					t.Errorf("ID esperado: %d, recebido: %v", tt.mockService.deputado.ID, data["id"])
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkGetDeputadosHandler(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := &MockDeputadosService{
		deputados: []domain.Deputado{
			{ID: 1, Nome: "Deputado 1", UF: "SP", Partido: "PT"},
			{ID: 2, Nome: "Deputado 2", UF: "RJ", Partido: "PSDB"},
		},
		source: "cache",
	}

	router.GET("/deputados", GetDeputadosHandler(mockService))

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/deputados", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// Mock Analytics Service para testes
type MockAnalyticsService struct {
	rankings         interface{}
	insights         *application.InsightsGerais
	rankingDeputados []domain.RankingDeputadoVotacao
	rankingPartidos  []domain.VotacaoPartido
	statsVotacoes    *domain.VotacaoStats
	err              error
	source           string
	atualizarChamado bool
}

func (m *MockAnalyticsService) GetRankingGastos(ctx context.Context, ano int, limite int) (*application.RankingGastos, string, error) {
	if m.err != nil {
		return nil, "", m.err
	}
	ranking := &application.RankingGastos{
		Ano:         ano,
		TotalGeral:  1000000.0,
		MediaGastos: 50000.0,
		Deputados:   []application.DeputadoRankingGastos{},
	}
	return ranking, m.source, nil
}

func (m *MockAnalyticsService) GetRankingProposicoes(ctx context.Context, ano int, limite int) (*application.RankingProposicoes, string, error) {
	if m.err != nil {
		return nil, "", m.err
	}
	ranking := &application.RankingProposicoes{
		Ano:       ano,
		Deputados: []application.DeputadoRankingProposicoes{},
	}
	return ranking, m.source, nil
}

func (m *MockAnalyticsService) GetRankingPresenca(ctx context.Context, ano int, limite int) (*application.RankingPresenca, string, error) {
	if m.err != nil {
		return nil, "", m.err
	}
	ranking := &application.RankingPresenca{
		Ano:       ano,
		Deputados: []application.DeputadoRankingPresenca{},
	}
	return ranking, m.source, nil
}

func (m *MockAnalyticsService) GetInsightsGerais(ctx context.Context) (*application.InsightsGerais, string, error) {
	if m.err != nil {
		return nil, "", m.err
	}
	if m.insights == nil {
		m.insights = &application.InsightsGerais{
			TotalDeputados:      513,
			TotalGastoAno:       25000000.0,
			TotalProposicoesAno: 1500,
			MediaGastosDeputado: 50000.0,
			PartidoMaiorGasto:   "PT",
			UFMaiorGasto:        "SP",
		}
	}
	return m.insights, m.source, nil
}

func (m *MockAnalyticsService) AtualizarRankings(ctx context.Context) error {
	m.atualizarChamado = true
	return m.err
}

// Novos métodos de Analytics (stubs para testes)
func (m *MockAnalyticsService) GetRankingDeputadosVotacao(ctx context.Context, ano int, limite int) ([]domain.RankingDeputadoVotacao, string, error) {
	if m.err != nil {
		return nil, "", m.err
	}
	if len(m.rankingDeputados) == 0 {
		m.rankingDeputados = []domain.RankingDeputadoVotacao{
			{IDDeputado: 1, Nome: "Deputado de Teste", SiglaPartido: "PT", SiglaUF: "SP", URLFoto: "https://example.com/foto.jpg", TotalVotacoes: 80, VotosFavoraveis: 60, VotosContrarios: 15, Abstencoes: 5},
		}
	}
	return m.rankingDeputados, m.source, nil
}

func (m *MockAnalyticsService) GetRankingPartidosDisciplina(ctx context.Context, ano int) ([]domain.VotacaoPartido, string, error) {
	if m.err != nil {
		return nil, "", m.err
	}
	if len(m.rankingPartidos) == 0 {
		m.rankingPartidos = []domain.VotacaoPartido{{Partido: "PT", Orientacao: "Sim", VotaramFavor: 10, VotaramContra: 2, VotaramAbstencao: 1, TotalMembros: 15}}
	}
	return m.rankingPartidos, m.source, nil
}

func (m *MockAnalyticsService) GetStatsVotacoes(ctx context.Context, periodo string) (*domain.VotacaoStats, string, error) {
	if m.err != nil {
		return nil, "", m.err
	}
	if m.statsVotacoes == nil {
		m.statsVotacoes = &domain.VotacaoStats{
			TotalVotacoes:      100,
			VotacoesAprovadas:  70,
			VotacoesRejeitadas: 30,
			MediaParticipacao:  75.0,
		}
	}
	return m.statsVotacoes, m.source, nil
}

func TestGetDespesasDeputadoHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		deputadoID   string
		ano          string
		mockService  *MockDeputadosService
		expectedCode int
	}{
		{
			name:       "success - despesas encontradas",
			deputadoID: "123",
			ano:        "2024",
			mockService: &MockDeputadosService{
				source: "api",
			},
			expectedCode: http.StatusOK,
		},
		{
			name:       "error - erro no serviço",
			deputadoID: "123",
			mockService: &MockDeputadosService{
				err: errors.New("erro do serviço"),
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/deputados/:id/despesas", GetDespesasDeputadoHandler(tt.mockService))

			url := "/deputados/" + tt.deputadoID + "/despesas"
			if tt.ano != "" {
				url += "?ano=" + tt.ano
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Status code = %v, want %v", w.Code, tt.expectedCode)
			}
		})
	}
}

func TestGetRankingGastosHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		limite       string
		mockService  *MockAnalyticsService
		expectedCode int
	}{
		{
			name:   "success - ranking encontrado",
			limite: "10",
			mockService: &MockAnalyticsService{
				rankings: []interface{}{
					map[string]interface{}{"nome": "Deputado 1", "gasto": 1000.0},
				},
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "success - limite padrão",
			mockService:  &MockAnalyticsService{rankings: []interface{}{}},
			expectedCode: http.StatusOK,
		},
		{
			name: "error - erro no serviço",
			mockService: &MockAnalyticsService{
				err: errors.New("erro analytics"),
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/rankings/gastos", GetRankingGastosHandler(tt.mockService))

			url := "/rankings/gastos"
			if tt.limite != "" {
				url += "?limite=" + tt.limite
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Status code = %v, want %v", w.Code, tt.expectedCode)
			}
		})
	}
}

func TestGetRankingProposicoesHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		mockService  *MockAnalyticsService
		expectedCode int
	}{
		{
			name: "success - ranking encontrado",
			mockService: &MockAnalyticsService{
				rankings: []interface{}{
					map[string]interface{}{"nome": "Deputado 1", "proposicoes": 10},
				},
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "error - erro no serviço",
			mockService: &MockAnalyticsService{
				err: errors.New("erro analytics"),
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/rankings/proposicoes", GetRankingProposicoesHandler(tt.mockService))

			req := httptest.NewRequest("GET", "/rankings/proposicoes", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Status code = %v, want %v", w.Code, tt.expectedCode)
			}
		})
	}
}

func TestGetRankingPresencaHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		mockService  *MockAnalyticsService
		expectedCode int
	}{
		{
			name: "success - ranking encontrado",
			mockService: &MockAnalyticsService{
				rankings: []interface{}{
					map[string]interface{}{"nome": "Deputado 1", "presenca": 95.5},
				},
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "error - erro no serviço",
			mockService: &MockAnalyticsService{
				err: errors.New("erro analytics"),
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/rankings/presenca", GetRankingPresencaHandler(tt.mockService))

			req := httptest.NewRequest("GET", "/rankings/presenca", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Status code = %v, want %v", w.Code, tt.expectedCode)
			}
		})
	}
}

func TestGetInsightsGeraisHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		mockService  *MockAnalyticsService
		expectedCode int
	}{
		{
			name: "success - insights encontrados",
			mockService: &MockAnalyticsService{
				insights: &application.InsightsGerais{
					TotalDeputados:      513,
					TotalGastoAno:       25000000.0,
					TotalProposicoesAno: 1500,
					MediaGastosDeputado: 50000.0,
					PartidoMaiorGasto:   "PT",
					UFMaiorGasto:        "SP",
				},
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "error - erro no serviço",
			mockService: &MockAnalyticsService{
				err: errors.New("erro analytics"),
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/insights", GetInsightsGeraisHandler(tt.mockService))

			req := httptest.NewRequest("GET", "/insights", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Status code = %v, want %v", w.Code, tt.expectedCode)
			}
		})
	}
}

func TestPostAtualizarRankingsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		mockService  *MockAnalyticsService
		expectedCode int
	}{
		{
			name:         "success - rankings atualizados",
			mockService:  &MockAnalyticsService{},
			expectedCode: http.StatusOK,
		},
		{
			name: "error - erro na atualização",
			mockService: &MockAnalyticsService{
				err: errors.New("erro ao atualizar"),
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.POST("/rankings/atualizar", PostAtualizarRankingsHandler(tt.mockService))

			req := httptest.NewRequest("POST", "/rankings/atualizar", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Status code = %v, want %v", w.Code, tt.expectedCode)
			}

			if tt.expectedCode == http.StatusOK && !tt.mockService.atualizarChamado {
				t.Error("AtualizarRankings should have been called")
			}
		})
	}
}

func TestGetRankingDeputadosVotacaoHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		query        string
		mockService  *MockAnalyticsService
		expectedCode int
	}{
		{
			name:         "success - ranking retornado",
			query:        "?ano=2024&limite=5",
			mockService:  &MockAnalyticsService{source: "computed"},
			expectedCode: http.StatusOK,
		},
		{
			name:  "error - serviço falha",
			query: "?ano=2024",
			mockService: &MockAnalyticsService{
				err: errors.New("falha"),
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "error - ano inválido",
			query:        "?ano=abc",
			mockService:  &MockAnalyticsService{},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/analytics/votacoes/ranking", GetRankingDeputadosVotacaoHandler(tt.mockService))

			req := httptest.NewRequest("GET", "/analytics/votacoes/ranking"+tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Status code = %v, want %v", w.Code, tt.expectedCode)
			}
		})
	}
}

func TestGetRankingPartidosDisciplinaHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		query        string
		mockService  *MockAnalyticsService
		expectedCode int
	}{
		{
			name:         "success - ranking retornado",
			query:        "?ano=2024",
			mockService:  &MockAnalyticsService{source: "computed"},
			expectedCode: http.StatusOK,
		},
		{
			name:  "error - serviço falha",
			query: "?ano=2024",
			mockService: &MockAnalyticsService{
				err: errors.New("falha"),
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "error - ano inválido",
			query:        "?ano=abc",
			mockService:  &MockAnalyticsService{},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/analytics/votacoes/partidos", GetRankingPartidosDisciplinaHandler(tt.mockService))

			req := httptest.NewRequest("GET", "/analytics/votacoes/partidos"+tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Status code = %v, want %v", w.Code, tt.expectedCode)
			}
		})
	}
}

func TestGetStatsVotacoesHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		query        string
		mockService  *MockAnalyticsService
		expectedCode int
	}{
		{
			name:         "success - estatísticas retornadas",
			query:        "?periodo=2024",
			mockService:  &MockAnalyticsService{source: "computed"},
			expectedCode: http.StatusOK,
		},
		{
			name:  "error - serviço falha",
			query: "?periodo=2024",
			mockService: &MockAnalyticsService{
				err: errors.New("falha"),
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/analytics/votacoes/stats", GetStatsVotacoesHandler(tt.mockService))

			req := httptest.NewRequest("GET", "/analytics/votacoes/stats"+tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Status code = %v, want %v", w.Code, tt.expectedCode)
			}
		})
	}
}
