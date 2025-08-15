package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"to-de-olho-backend/internal/domain"

	"golang.org/x/time/rate"
)

const (
	DefaultBaseURLCamara = "https://dadosabertos.camara.leg.br/api/v2"
	UserAgent            = "ToDeOlho/1.0 (+https://github.com/alzarus/to-de-olho)"
)

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
		rps = 100
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
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/json")

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
		time.Sleep(backoff)
		backoff *= 2
	}
	if lastErr == nil {
		lastErr = errors.New("falha desconhecida na requisição")
	}
	return nil, lastErr
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
