package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

const (
	DefaultBaseURLCamara = "https://dadosabertos.camara.leg.br/api/v2"
	UserAgent            = "ToDeOlho/1.0 (+https://github.com/alzarus/to-de-olho)"
)

// CamaraClient encapsula chamadas à API da Câmara com rate limit e retry/backoff
type CamaraClient struct {
	httpClient *http.Client
	baseURL    string
	limiter    *rate.Limiter
}

func NewCamaraClient(baseURL string, timeout time.Duration, rps int, burst int) *CamaraClient {
	if baseURL == "" {
		baseURL = DefaultBaseURLCamara
	}
	if rps <= 0 {
		rps = 100 // limite da API ~100 req/min → ~1.6 rps; mantemos margem
	}
	if burst <= 0 {
		burst = rps
	}
	return &CamaraClient{
		httpClient: &http.Client{Timeout: timeout},
		baseURL:    baseURL,
		limiter:    rate.NewLimiter(rate.Limit(float64(rps)), burst),
	}
}

func (c *CamaraClient) doRequest(ctx context.Context, method, url string) ([]byte, error) {
	// Respeitar rate limit local
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/json")

	// Retry com backoff exponencial (até 3 tentativas)
	var lastErr error
	backoff := 200 * time.Millisecond
	for i := 0; i < 3; i++ {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
		} else {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				b, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("erro ao ler resposta: %w", err)
				}
				return b, nil
			}
			lastErr = fmt.Errorf("API retornou status %d", resp.StatusCode)
		}
		// backoff antes da próxima tentativa
		time.Sleep(backoff)
		backoff *= 2
	}
	if lastErr == nil {
		lastErr = errors.New("falha desconhecida na requisição")
	}
	return nil, lastErr
}

// FetchDeputados busca deputados da API oficial
func (c *CamaraClient) FetchDeputados(ctx context.Context, partido, uf, nome string) ([]Deputado, error) {
	url := fmt.Sprintf("%s/deputados?ordem=ASC&ordenarPor=nome", c.baseURL)
	if partido != "" {
		url += "&siglaPartido=" + partido
	}
	if uf != "" {
		url += "&siglaUf=" + uf
	}
	if nome != "" {
		url += "&nome=" + nome
	}

	body, err := c.doRequest(ctx, http.MethodGet, url)
	if err != nil {
		return nil, err
	}
	var apiResp APIResponseDeputados
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON: %w", err)
	}
	return apiResp.Dados, nil
}

// FetchDeputadoByID busca um deputado específico por ID
func (c *CamaraClient) FetchDeputadoByID(ctx context.Context, id string) (*Deputado, error) {
	url := fmt.Sprintf("%s/deputados/%s", c.baseURL, id)
	body, err := c.doRequest(ctx, http.MethodGet, url)
	if err != nil {
		return nil, err
	}
	var apiResp struct {
		Dados Deputado `json:"dados"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON: %w", err)
	}
	return &apiResp.Dados, nil
}

// FetchDespesas busca despesas de um deputado num ano
func (c *CamaraClient) FetchDespesas(ctx context.Context, deputadoID, ano string) ([]Despesa, error) {
	url := fmt.Sprintf("%s/deputados/%s/despesas?ano=%s&ordem=DESC&ordenarPor=dataDocumento", c.baseURL, deputadoID, ano)
	body, err := c.doRequest(ctx, http.MethodGet, url)
	if err != nil {
		return nil, err
	}
	var apiResp APIResponseDespesas
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON: %w", err)
	}
	return apiResp.Dados, nil
}
