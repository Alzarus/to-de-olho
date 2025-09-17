package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"to-de-olho-backend/internal/config"
	"to-de-olho-backend/internal/domain"
	"to-de-olho-backend/internal/infrastructure/resilience"

	"golang.org/x/time/rate"
)

const (
	DefaultBaseURLCamara = "https://dadosabertos.camara.leg.br/api/v2"
	UserAgent            = "ToDeOlho/1.0 (+https://github.com/alzarus/to-de-olho)"
)

type CamaraClient struct {
	httpClient     *http.Client
	baseURL        string
	limiter        *rate.Limiter
	circuitBreaker *resilience.CircuitBreaker
}

func NewCamaraClient(baseURL string, timeout time.Duration, rps int, burst int) *CamaraClient {
	if baseURL == "" {
		baseURL = DefaultBaseURLCamara
	}
	if rps <= 0 {
		rps = 100
	}
	if burst <= 0 {
		burst = rps
	}

	// Configuração otimizada para API da Câmara
	circuitConfig := resilience.CircuitBreakerConfig{
		MaxFailures:      3,                // 3 falhas para abrir
		ResetTimeout:     60 * time.Second, // 1 minuto para retry (API pode estar sobrecarregada)
		SuccessThreshold: 2,                // 2 sucessos para fechar
		Timeout:          timeout,          // Mesmo timeout do HTTP client
	}

	return &CamaraClient{
		httpClient:     &http.Client{Timeout: timeout},
		baseURL:        baseURL,
		limiter:        rate.NewLimiter(rate.Limit(float64(rps)), burst),
		circuitBreaker: resilience.NewCircuitBreaker(circuitConfig),
	}
}

// NewCamaraClientFromConfig creates a client from config
func NewCamaraClientFromConfig(cfg *config.CamaraClientConfig) *CamaraClient {
	// Configuração otimizada para API da Câmara
	circuitConfig := resilience.CircuitBreakerConfig{
		MaxFailures:      3,                // 3 falhas para abrir
		ResetTimeout:     60 * time.Second, // 1 minuto para retry
		SuccessThreshold: 2,                // 2 sucessos para fechar
		Timeout:          cfg.Timeout,      // Timeout da configuração
	}

	return &CamaraClient{
		httpClient:     &http.Client{Timeout: cfg.Timeout},
		baseURL:        cfg.BaseURL,
		limiter:        rate.NewLimiter(rate.Limit(float64(cfg.RPS)), cfg.Burst),
		circuitBreaker: resilience.NewCircuitBreaker(circuitConfig),
	}
}

func (c *CamaraClient) doRequest(ctx context.Context, method, url string) ([]byte, error) {
	// Aplicar rate limiting primeiro
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	// Executar requisição com circuit breaker
	var result []byte
	err := c.circuitBreaker.Execute(ctx, func(ctx context.Context) error {
		req, err := http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return fmt.Errorf("erro ao criar requisição: %w", err)
		}
		req.Header.Set("User-Agent", UserAgent)
		req.Header.Set("Accept", "application/json")

		// Retry com backoff exponencial
		var lastErr error
		backoff := 200 * time.Millisecond

		for i := 0; i < 3; i++ {
			resp, err := c.httpClient.Do(req)
			if err != nil {
				lastErr = fmt.Errorf("erro na requisição HTTP (tentativa %d/3): %w", i+1, err)
				if i < 2 { // Não fazer sleep na última tentativa
					time.Sleep(backoff)
					backoff *= 2
				}
				continue
			}

			defer resp.Body.Close()

			// Verificar status codes
			switch resp.StatusCode {
			case http.StatusOK:
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return fmt.Errorf("erro ao ler resposta: %w", err)
				}
				result = body
				return nil

			case http.StatusTooManyRequests:
				lastErr = fmt.Errorf("rate limit excedido (429), tentativa %d/3", i+1)

			case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
				lastErr = fmt.Errorf("erro do servidor (%d), tentativa %d/3", resp.StatusCode, i+1)

			default:
				// Para outros códigos, não tenta novamente
				return fmt.Errorf("erro HTTP %d: %s", resp.StatusCode, resp.Status)
			}

			if i < 2 { // Não fazer sleep na última tentativa
				time.Sleep(backoff)
				backoff *= 2
			}
		}

		return fmt.Errorf("falha após 3 tentativas: %w", lastErr)
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetCircuitBreakerMetrics retorna métricas do circuit breaker para monitoramento
func (c *CamaraClient) GetCircuitBreakerMetrics() map[string]interface{} {
	return c.circuitBreaker.GetMetrics()
}

func (c *CamaraClient) FetchDeputados(ctx context.Context, partido, uf, nome string) ([]domain.Deputado, error) {
	urlStr := fmt.Sprintf("%s/deputados?ordem=ASC&ordenarPor=nome", c.baseURL)
	if partido != "" {
		urlStr += "&siglaPartido=" + url.QueryEscape(partido)
	}
	if uf != "" {
		urlStr += "&siglaUf=" + url.QueryEscape(uf)
	}
	if nome != "" {
		urlStr += "&nome=" + url.QueryEscape(nome)
	}
	body, err := c.doRequest(ctx, http.MethodGet, urlStr)
	if err != nil {
		return nil, err
	}
	var apiResp struct {
		Dados []domain.Deputado `json:"dados"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON: %w", err)
	}
	return apiResp.Dados, nil
}

// FetchDeputadosPaged allows explicit pagination to retrieve all items (itens up to 100)
func (c *CamaraClient) FetchDeputadosPaged(ctx context.Context, partido, uf, nome string, pagina, itens int) ([]domain.Deputado, error) {
	if itens <= 0 || itens > 100 {
		itens = 100
	}
	if pagina <= 0 {
		pagina = 1
	}
	urlStr := fmt.Sprintf("%s/deputados?ordem=ASC&ordenarPor=nome&pagina=%d&itens=%d", c.baseURL, pagina, itens)
	if partido != "" {
		urlStr += "&siglaPartido=" + url.QueryEscape(partido)
	}
	if uf != "" {
		urlStr += "&siglaUf=" + url.QueryEscape(uf)
	}
	if nome != "" {
		urlStr += "&nome=" + url.QueryEscape(nome)
	}
	body, err := c.doRequest(ctx, http.MethodGet, urlStr)
	if err != nil {
		return nil, err
	}
	var apiResp struct {
		Dados []domain.Deputado `json:"dados"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON: %w", err)
	}
	return apiResp.Dados, nil
}

// FetchAllDeputados retrieves all deputies by paginating until empty page
func (c *CamaraClient) FetchAllDeputados(ctx context.Context) ([]domain.Deputado, error) {
	var all []domain.Deputado
	pagina := 1
	for {
		batch, err := c.FetchDeputadosPaged(ctx, "", "", "", pagina, 100)
		if err != nil {
			return nil, err
		}
		if len(batch) == 0 {
			break
		}
		all = append(all, batch...)
		pagina++
		// small pause to be gentle
		time.Sleep(50 * time.Millisecond)
	}
	return all, nil
}

func (c *CamaraClient) FetchDeputadoByID(ctx context.Context, id string) (*domain.Deputado, error) {
	url := fmt.Sprintf("%s/deputados/%s", c.baseURL, id)
	body, err := c.doRequest(ctx, http.MethodGet, url)
	if err != nil {
		return nil, err
	}
	var apiResp struct {
		Dados domain.Deputado `json:"dados"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON: %w", err)
	}
	return &apiResp.Dados, nil
}

func (c *CamaraClient) FetchDespesas(ctx context.Context, deputadoID, ano string) ([]domain.Despesa, error) {
	url := fmt.Sprintf("%s/deputados/%s/despesas?ano=%s&ordem=DESC&ordenarPor=dataDocumento", c.baseURL, deputadoID, ano)
	body, err := c.doRequest(ctx, http.MethodGet, url)
	if err != nil {
		return nil, err
	}
	var apiResp struct {
		Dados []domain.Despesa `json:"dados"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON: %w", err)
	}
	return apiResp.Dados, nil
}

// FetchProposicoes busca proposições da API da Câmara com filtros
func (c *CamaraClient) FetchProposicoes(ctx context.Context, filtros *domain.ProposicaoFilter) ([]domain.Proposicao, error) {
	// Construir URL com parâmetros de query
	params := filtros.BuildAPIQueryParams()

	baseURL := fmt.Sprintf("%s/proposicoes", c.baseURL)
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao construir URL: %w", err)
	}

	// Adicionar parâmetros de query
	query := u.Query()
	for key, value := range params {
		query.Set(key, value)
	}
	u.RawQuery = query.Encode()

	body, err := c.doRequest(ctx, http.MethodGet, u.String())
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar proposições: %w", err)
	}

	var apiResp struct {
		Dados []domain.Proposicao `json:"dados"`
		Links []struct {
			Rel  string `json:"rel"`
			Href string `json:"href"`
		} `json:"links"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON de proposições: %w", err)
	}

	return apiResp.Dados, nil
}

// FetchProposicaoPorID busca uma proposição específica por ID
func (c *CamaraClient) FetchProposicaoPorID(ctx context.Context, id int) (*domain.Proposicao, error) {
	if id <= 0 {
		return nil, domain.ErrProposicaoIDInvalido
	}

	url := fmt.Sprintf("%s/proposicoes/%d", c.baseURL, id)
	body, err := c.doRequest(ctx, http.MethodGet, url)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar proposição %d: %w", id, err)
	}

	var apiResp struct {
		Dados domain.Proposicao `json:"dados"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON da proposição %d: %w", id, err)
	}

	return &apiResp.Dados, nil
}
