package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"to-de-olho-backend/internal/config"
	"to-de-olho-backend/internal/domain"
	"to-de-olho-backend/internal/infrastructure/resilience"

	"golang.org/x/time/rate"
)

// isRetryableServerError detects errors that are likely transient server/timeouts
// and worth retrying with smaller chunks (e.g. 502/503/504/gateway/timeout).
func isRetryableServerError(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	if strings.Contains(s, "504") || strings.Contains(s, "gateway") || strings.Contains(s, "timeout") || strings.Contains(s, "502") || strings.Contains(s, "503") {
		return true
	}
	return false
}

const (
	DefaultBaseURLCamara = "https://dadosabertos.camara.leg.br/api/v2"
	UserAgent            = "ToDeOlho/1.0 (+https://github.com/alzarus/to-de-olho)"
)

// Int64String é um wrapper que aceita valores JSON que podem ser número ou string
// e fornece um Int64 seguro via Int64(). Isso evita erros quando a API varia o
// tipo de representação para ids.
type Int64String int64

func (i *Int64String) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "null" || s == "\"\"" || s == "" {
		*i = 0
		return nil
	}

	// Se vier com aspas, desempacotar
	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		if unq, err := strconv.Unquote(s); err == nil {
			s = unq
		}
	}

	// Alguns ids vêm no formato "12345-67" — aceitar a parte antes do '-'
	if idx := strings.Index(s, "-"); idx != -1 {
		s = s[:idx]
	}

	s = strings.TrimSpace(s)
	if s == "" {
		*i = 0
		return nil
	}

	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		*i = Int64String(n)
		return nil
	}

	return fmt.Errorf("Int64String: formato não suportado: %s", s)
}

func (i Int64String) Int64() int64 { return int64(i) }

// parseTimeFlexible tenta vários formatos comuns retornados pela API da Câmara
func parseTimeFlexible(value string, loc *time.Location) (time.Time, error) {
	// Empty value is common in the API; treat as zero time without returning an error
	// to avoid noisy WARN logs upstream. Callers should handle zero time as necessary.
	if value == "" {
		return time.Time{}, nil
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02",
	}

	for _, l := range layouts {
		if t, err := time.Parse(l, value); err == nil {
			// se layout não inclui timezone, aplicar location se fornecida
			if l == "2006-01-02T15:04:05" || l == "2006-01-02T15:04" || l == "2006-01-02" {
				if loc != nil {
					y, m, d := t.Date()
					hh, mm, ss := t.Clock()
					return time.Date(y, m, d, hh, mm, ss, t.Nanosecond(), loc), nil
				}
			}
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unrecognized time format: %s", value)
}

type CamaraClient struct {
	httpClient     *http.Client
	baseURL        string
	limiter        *rate.Limiter
	circuitBreaker *resilience.CircuitBreaker
	maxRetries     int
	backoffBase    time.Duration
	pageDelay      time.Duration
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
		maxRetries:     cfg.MaxRetries,
		backoffBase:    cfg.RetryDelay,
		pageDelay:      cfg.RetryDelay,
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

		// Retry com backoff exponencial (configurável)
		var lastErr error
		backoff := c.backoffBase
		if backoff <= 0 {
			backoff = 200 * time.Millisecond
		}

		maxRetries := c.maxRetries
		if maxRetries <= 0 {
			maxRetries = 3
		}

		for i := 0; i < maxRetries; i++ {
			resp, err := c.httpClient.Do(req)
			if err != nil {
				lastErr = fmt.Errorf("erro na requisição HTTP (tentativa %d/%d): %w", i+1, maxRetries, err)
				if i < maxRetries-1 {
					time.Sleep(backoff)
					backoff *= 2
				}
				continue
			}

			// Ler corpo sempre para ajudar no diagnóstico de erros 4xx/5xx
			bodyBytes, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr != nil {
				// Se falhar ao ler o corpo, registrar o erro de leitura mas continuar com o código de status
				lastErr = fmt.Errorf("erro ao ler resposta (tentativa %d/%d): %w", i+1, maxRetries, readErr)
				if i < maxRetries-1 {
					time.Sleep(backoff)
					backoff *= 2
				}
				continue
			}

			// Verificar status codes
			switch resp.StatusCode {
			case http.StatusOK:
				result = bodyBytes
				return nil

			case http.StatusTooManyRequests:
				lastErr = fmt.Errorf("rate limit excedido (429), tentativa %d/%d: %s", i+1, maxRetries, string(bodyBytes))

			case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
				lastErr = fmt.Errorf("erro do servidor (%d), tentativa %d/%d: %s", resp.StatusCode, i+1, maxRetries, string(bodyBytes))

			case http.StatusNotFound:
				// Para endpoints filhos (/votacoes/{id}/votos e /orientacoes) a API pode
				// retornar 404 quando não há dados. Nestes casos, retornar o corpo para o
				// chamador tratar como lista vazia, sem provocar abertura do circuito.
				if req.URL != nil {
					path := req.URL.Path
					if strings.Contains(path, "/votacoes/") && (strings.HasSuffix(path, "/votos") || strings.HasSuffix(path, "/orientacoes")) {
						result = bodyBytes
						return nil
					}
				}
				// Caso contrário, tratar 404 como erro comum
				lastErr = fmt.Errorf("erro HTTP %d: %s - %s", resp.StatusCode, resp.Status, string(bodyBytes))

			default:
				// Para outros códigos, não tenta novamente — incluir corpo para diagnóstico
				return fmt.Errorf("erro HTTP %d: %s - %s", resp.StatusCode, resp.Status, string(bodyBytes))
			}

			if i < maxRetries-1 { // Não fazer sleep na última tentativa
				time.Sleep(backoff)
				backoff *= 2
			}
		}

		return fmt.Errorf("falha após %d tentativas: %w", maxRetries, lastErr)
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetCircuitBreakerMetrics retorna métricas do circuit breaker para monitoramento
func (c *CamaraClient) GetCircuitBreakerMetrics() map[string]interface{} {
	if c == nil || c.circuitBreaker == nil {
		return nil
	}
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
	// Construir URL inicial
	urlStr := fmt.Sprintf("%s/deputados/%s/despesas?ano=%s&ordem=DESC&ordenarPor=dataDocumento", c.baseURL, deputadoID, ano)

	var all []domain.Despesa
	for {
		body, err := c.doRequest(ctx, http.MethodGet, urlStr)
		if err != nil {
			return nil, err
		}

		var apiResp struct {
			Dados []domain.Despesa `json:"dados"`
			Links []struct {
				Rel  string `json:"rel"`
				Href string `json:"href"`
			} `json:"links"`
		}

		if err := json.Unmarshal(body, &apiResp); err != nil {
			return nil, fmt.Errorf("erro ao decodificar JSON: %w", err)
		}

		all = append(all, apiResp.Dados...)

		// Procurar link 'next'
		next := ""
		for _, l := range apiResp.Links {
			if l.Rel == "next" {
				next = l.Href
				break
			}
		}

		if next == "" {
			break
		}

		// Usar next como próxima URL
		urlStr = next
		// Pequena pausa para ser gentil com a API (configurável)
		if c.pageDelay > 0 {
			time.Sleep(c.pageDelay)
		} else {
			time.Sleep(50 * time.Millisecond)
		}
	}

	return all, nil
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

	// Paginar automaticamente seguindo links 'next'
	urlStr := u.String()
	var all []domain.Proposicao

	for {
		body, err := c.doRequest(ctx, http.MethodGet, urlStr)
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

		all = append(all, apiResp.Dados...)

		// Procurar link 'next'
		next := ""
		for _, l := range apiResp.Links {
			if l.Rel == "next" {
				next = l.Href
				break
			}
		}

		if next == "" {
			break
		}

		urlStr = next
		// Pequena pausa para não bombardear a API (configurável)
		if c.pageDelay > 0 {
			time.Sleep(c.pageDelay)
		} else {
			time.Sleep(50 * time.Millisecond)
		}
	}

	return all, nil
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

// Métodos para votações (implementação da interface CamaraAPIPort)

// GetVotacoes busca votações em um período específico
func (c *CamaraClient) GetVotacoes(ctx context.Context, dataInicio, dataFim time.Time) ([]*domain.Votacao, error) {
	// Para ser conservador com a API pública da Câmara, dividimos o período
	// em fatias mensais (1 mês) e agregamos os resultados. Isso reduz a
	// chance de sobrecarga/erros 4xx/5xx por ranges muito grandes.

	var all []*domain.Votacao

	// Normalizar para data (sem hora) preservando a Location passada
	loc := dataInicio.Location()
	if loc == nil {
		loc = time.UTC
	}
	start := time.Date(dataInicio.Year(), dataInicio.Month(), dataInicio.Day(), 0, 0, 0, 0, loc)
	end := time.Date(dataFim.Year(), dataFim.Month(), dataFim.Day(), 0, 0, 0, 0, loc)

	// Estratégia de chunking adaptativo: tenta ranges maiores e degrada para menores
	// quando encontra erros retryable (ex: 504). Sequência configurada aqui:
	// 30 dias -> 7 dias -> 1 dia. Isso reduz o número de requests quando a API
	// está saudável, e permite recuperação granular quando há problemas.
	chunkStrategy := []int{30, 7, 1}

	fetchRange := func(rangeStart, rangeEnd time.Time) error {
		params := url.Values{}
		params.Add("dataInicio", rangeStart.In(loc).Format("2006-01-02"))
		params.Add("dataFim", rangeEnd.In(loc).Format("2006-01-02"))
		params.Add("ordenarPor", "dataHoraRegistro")
		params.Add("ordem", "DESC")
		params.Add("itens", "100")

		pageURL := fmt.Sprintf("%s/votacoes?%s", c.baseURL, params.Encode())

		for {
			body, err := c.doRequest(ctx, http.MethodGet, pageURL)
			if err != nil {
				return err
			}

			var response struct {
				Dados []struct {
					ID                 Int64String `json:"id"`
					URI                string      `json:"uri"`
					Titulo             string      `json:"titulo"`
					TipoVotacao        string      `json:"tipoVotacao"`
					Aprovacao          int         `json:"aprovacao"`
					DataHoraInicio     string      `json:"dataHoraInicio"`
					DataHoraRegistro   string      `json:"dataHoraRegistro"`
					DataHoraFim        string      `json:"dataHoraFim"`
					Descricao          string      `json:"descricao"`
					ProposicaoObjeto   string      `json:"proposicaoObjeto"`
					UltimaApresentacao struct {
						ID         Int64String `json:"id"`
						Proposicao struct {
							ID     Int64String `json:"id"`
							Numero string      `json:"numero"`
							Ano    int         `json:"ano"`
							Tipo   string      `json:"tipo"`
						} `json:"proposicao"`
					} `json:"ultimaApresentacaoProposicao"`
				} `json:"dados"`
				Links []struct {
					Rel  string `json:"rel"`
					Href string `json:"href"`
				} `json:"links"`
			}

			if err := json.Unmarshal(body, &response); err != nil {
				return fmt.Errorf("erro ao fazer parse da resposta de votações: %w", err)
			}

			for _, item := range response.Dados {
				// prefer dataHoraRegistro (mais consistente no API) e fallback para dataHoraInicio
				dtStr := item.DataHoraInicio
				if item.DataHoraRegistro != "" {
					dtStr = item.DataHoraRegistro
				}
				dataVotacao, err := parseTimeFlexible(dtStr, loc)
				if err != nil {
					// fallback: zero time but continue
					// Apenas log se não for string vazia (API retorna vazio frequentemente)
					if dtStr != "" {
						cErr := fmt.Errorf("erro ao parsear dataHoraInicio: %w", err)
						fmt.Printf("WARN: %v\n", cErr)
					}
					dataVotacao = time.Time{}
				}

				// Converter aprovação de int para string
				aprovacao := "Pendente"
				if item.Aprovacao == 1 {
					aprovacao = "Aprovada"
				} else if item.Aprovacao == 0 {
					aprovacao = "Rejeitada"
				}

				var proposicaoID *int64
				if item.UltimaApresentacao.Proposicao.ID.Int64() != 0 {
					id := item.UltimaApresentacao.Proposicao.ID.Int64()
					proposicaoID = &id
				}

				var anoProposicao *int
				if item.UltimaApresentacao.Proposicao.Ano != 0 {
					anoProposicao = &item.UltimaApresentacao.Proposicao.Ano
				}

				votacao := &domain.Votacao{
					IDVotacaoCamara:       item.ID.Int64(),
					Titulo:                item.Titulo,
					Ementa:                item.Descricao,
					DataVotacao:           dataVotacao,
					Aprovacao:             aprovacao,
					IDProposicaoPrincipal: proposicaoID,
					TipoProposicao:        item.UltimaApresentacao.Proposicao.Tipo,
					NumeroProposicao:      item.UltimaApresentacao.Proposicao.Numero,
					AnoProposicao:         anoProposicao,
					Relevancia:            "média", // Padrão, pode ser calculada depois
					Payload: map[string]interface{}{
						"uri":              item.URI,
						"tipoVotacao":      item.TipoVotacao,
						"proposicaoObjeto": item.ProposicaoObjeto,
						"dataHoraFim":      item.DataHoraFim,
					},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				all = append(all, votacao)
			}

			// Procurar link 'next'
			next := ""
			for _, l := range response.Links {
				if l.Rel == "next" {
					next = l.Href
					break
				}
			}

			if next == "" {
				break
			}

			pageURL = next
			if c.pageDelay > 0 {
				time.Sleep(c.pageDelay)
			} else {
				time.Sleep(50 * time.Millisecond)
			}
		}

		return nil
	}

	// Cursor para o dia atual a processar
	for cur := start; !cur.After(end); {
		progressed := false

		for _, days := range chunkStrategy {
			candidateEnd := cur.AddDate(0, 0, days-1)
			if candidateEnd.After(end) {
				candidateEnd = end
			}

			if err := fetchRange(cur, candidateEnd); err != nil {
				if isRetryableServerError(err) {
					// tentar próximo tamanho menor
					fmt.Printf("WARN: chunk %s - %s falhou com erro transitório: %v. Tentando próximo tamanho menor...\n", cur.Format("2006-01-02"), candidateEnd.Format("2006-01-02"), err)
					// pequena pausa antes de tentar subdividir
					time.Sleep(50 * time.Millisecond)
					continue
				}
				return nil, fmt.Errorf("erro ao buscar votações: %w", err)
			}

			// sucesso no chunk candidate
			cur = candidateEnd.AddDate(0, 0, 1)
			// ser gentil entre chunks
			time.Sleep(100 * time.Millisecond)
			progressed = true
			break
		}

		if !progressed {
			// mesmo um chunk de 1 dia falhou com erro retryable — pular 1 dia para evitar loop
			fmt.Printf("ERROR: todos os tamanhos de chunk falharam para %s; pulando 1 dia\n", cur.Format("2006-01-02"))
			cur = cur.AddDate(0, 0, 1)
			time.Sleep(100 * time.Millisecond)
		}
	}

	return all, nil
}

// GetVotacao busca uma votação específica por ID
func (c *CamaraClient) GetVotacao(ctx context.Context, id int64) (*domain.Votacao, error) {
	urlStr := fmt.Sprintf("%s/votacoes/%d", c.baseURL, id)

	body, err := c.doRequest(ctx, http.MethodGet, urlStr)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar votação ID %d: %w", id, err)
	}

	var response struct {
		Dados struct {
			ID                 Int64String `json:"id"`
			URI                string      `json:"uri"`
			Titulo             string      `json:"titulo"`
			TipoVotacao        string      `json:"tipoVotacao"`
			Aprovacao          int         `json:"aprovacao"`
			DataHoraInicio     string      `json:"dataHoraInicio"`
			DataHoraRegistro   string      `json:"dataHoraRegistro"`
			DataHoraFim        string      `json:"dataHoraFim"`
			Descricao          string      `json:"descricao"`
			ProposicaoObjeto   string      `json:"proposicaoObjeto"`
			UltimaApresentacao struct {
				ID         Int64String `json:"id"`
				Proposicao struct {
					ID     Int64String `json:"id"`
					Numero string      `json:"numero"`
					Ano    int         `json:"ano"`
					Tipo   string      `json:"tipo"`
				} `json:"proposicao"`
			} `json:"ultimaApresentacaoProposicao"`
		} `json:"dados"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("erro ao fazer parse da resposta da votação: %w", err)
	}

	item := response.Dados
	dtStr := item.DataHoraInicio
	if item.DataHoraRegistro != "" {
		dtStr = item.DataHoraRegistro
	}
	dataVotacao, _ := parseTimeFlexible(dtStr, time.UTC)

	// Converter aprovação
	aprovacao := "Pendente"
	if item.Aprovacao == 1 {
		aprovacao = "Aprovada"
	} else if item.Aprovacao == 0 {
		aprovacao = "Rejeitada"
	}

	var proposicaoID *int64
	if item.UltimaApresentacao.Proposicao.ID.Int64() != 0 {
		id := item.UltimaApresentacao.Proposicao.ID.Int64()
		proposicaoID = &id
	}

	var anoProposicao *int
	if item.UltimaApresentacao.Proposicao.Ano != 0 {
		anoProposicao = &item.UltimaApresentacao.Proposicao.Ano
	}

	votacao := &domain.Votacao{
		IDVotacaoCamara:       item.ID.Int64(),
		Titulo:                item.Titulo,
		Ementa:                item.Descricao,
		DataVotacao:           dataVotacao,
		Aprovacao:             aprovacao,
		IDProposicaoPrincipal: proposicaoID,
		TipoProposicao:        item.UltimaApresentacao.Proposicao.Tipo,
		NumeroProposicao:      item.UltimaApresentacao.Proposicao.Numero,
		AnoProposicao:         anoProposicao,
		Relevancia:            "média",
		Payload: map[string]interface{}{
			"uri":              item.URI,
			"tipoVotacao":      item.TipoVotacao,
			"proposicaoObjeto": item.ProposicaoObjeto,
			"dataHoraFim":      item.DataHoraFim,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return votacao, nil
}

// GetVotosPorVotacao busca votos dos deputados para uma votação
func (c *CamaraClient) GetVotosPorVotacao(ctx context.Context, idVotacao int64) ([]*domain.VotoDeputado, error) {
	urlStr := fmt.Sprintf("%s/votacoes/%d/votos", c.baseURL, idVotacao)

	body, err := c.doRequest(ctx, http.MethodGet, urlStr)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar votos da votação %d: %w", idVotacao, err)
	}

	var response struct {
		Dados []struct {
			Deputado struct {
				ID   int    `json:"id"`
				Nome string `json:"nome"`
			} `json:"deputado"`
			TipoVoto     string `json:"tipoVoto"`
			Voto         string `json:"voto"`
			DataRegistro string `json:"dataRegistroVoto"`
		} `json:"dados"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("erro ao fazer parse dos votos: %w", err)
	}

	votos := make([]*domain.VotoDeputado, 0, len(response.Dados))

	for _, item := range response.Dados {
		voto := &domain.VotoDeputado{
			IDVotacao:     idVotacao,
			IDDeputado:    item.Deputado.ID,
			Voto:          item.Voto,
			Justificativa: nil, // API da Câmara não retorna justificativa nos votos
			Payload: map[string]interface{}{
				"tipoVoto":     item.TipoVoto,
				"nomeDeputado": item.Deputado.Nome,
				"dataRegistro": item.DataRegistro,
			},
			CreatedAt: time.Now(),
		}

		votos = append(votos, voto)
	}

	return votos, nil
}

// GetOrientacoesPorVotacao busca orientações partidárias para uma votação
func (c *CamaraClient) GetOrientacoesPorVotacao(ctx context.Context, idVotacao int64) ([]*domain.OrientacaoPartido, error) {
	urlStr := fmt.Sprintf("%s/votacoes/%d/orientacoes", c.baseURL, idVotacao)

	body, err := c.doRequest(ctx, http.MethodGet, urlStr)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar orientações da votação %d: %w", idVotacao, err)
	}

	var response struct {
		Dados []struct {
			Partido    string `json:"siglaPartido"`
			Orientacao string `json:"orientacaoVoto"`
		} `json:"dados"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("erro ao fazer parse das orientações: %w", err)
	}

	orientacoes := make([]*domain.OrientacaoPartido, 0, len(response.Dados))

	for _, item := range response.Dados {
		orientacao := &domain.OrientacaoPartido{
			IDVotacao:  idVotacao,
			Partido:    item.Partido,
			Orientacao: item.Orientacao,
			CreatedAt:  time.Now(),
		}

		orientacoes = append(orientacoes, orientacao)
	}

	return orientacoes, nil
}

// FetchPartidos obtém lista de partidos da API da Câmara
func (c *CamaraClient) FetchPartidos(ctx context.Context) ([]domain.Partido, error) {
	urlStr := fmt.Sprintf("%s/partidos?itens=100", c.baseURL)

	body, err := c.doRequest(ctx, http.MethodGet, urlStr)
	if err != nil {
		return nil, err
	}

	var apiResp struct {
		Dados []struct {
			ID    int64  `json:"id"`
			Sigla string `json:"sigla"`
			Nome  string `json:"nome"`
			URI   string `json:"uri"`
		} `json:"dados"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON de partidos: %w", err)
	}

	out := make([]domain.Partido, 0, len(apiResp.Dados))
	for _, it := range apiResp.Dados {
		p := domain.Partido{
			ID:    it.ID,
			Sigla: it.Sigla,
			Nome:  it.Nome,
			URI:   it.URI,
			Payload: map[string]interface{}{
				"uri": it.URI,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		out = append(out, p)
	}

	return out, nil
}
