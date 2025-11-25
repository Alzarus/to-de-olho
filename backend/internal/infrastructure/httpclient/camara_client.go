package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"to-de-olho-backend/internal/config"
	"to-de-olho-backend/internal/domain"
	"to-de-olho-backend/internal/infrastructure/resilience"
	"to-de-olho-backend/internal/pkg/metrics"

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
	if strings.Contains(s, "deadline") || strings.Contains(s, "context deadline") || strings.Contains(s, "client.timeout") {
		return true
	}
	return false
}

// isAPIOverloaded detects when the API is overloaded and needs back-pressure
func isAPIOverloaded(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	// Padrões que indicam sobrecarga da API
	overloadIndicators := []string{
		"504", "gateway timeout", "upstream timeout",
		"503", "service unavailable", "temporarily unavailable",
		"502", "bad gateway", "connection failed",
		"context deadline exceeded", "timeout",
		"too many requests", "rate limit", "throttled",
	}

	for _, indicator := range overloadIndicators {
		if strings.Contains(s, indicator) {
			return true
		}
	}
	return false
}

const (
	DefaultBaseURLCamara = "https://dadosabertos.camara.leg.br/api/v2"
	UserAgent            = "ToDeOlho/1.0 (+https://github.com/alzarus/to-de-olho)"
)

// CamaraID representa um identificador retornado pela API da Câmara.
// Alguns endpoints retornam números, outros strings alfanuméricas. Quando
// possível, preservamos o valor numérico como referência auxiliar; caso
// contrário, apenas o valor textual é mantido.
type CamaraID struct {
	raw     string
	numeric *int64
}

func (id *CamaraID) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "" || s == "null" {
		*id = CamaraID{}
		return nil
	}

	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		unq, err := strconv.Unquote(s)
		if err != nil {
			return fmt.Errorf("CamaraID: erro ao remover aspas: %w", err)
		}
		s = strings.TrimSpace(unq)
	}

	raw := s

	candidate := raw
	if idx := strings.Index(candidate, "-"); idx > 0 {
		candidate = candidate[:idx]
	}
	if idx := strings.Index(candidate, "_"); idx > 0 {
		candidate = candidate[:idx]
	}

	var numericPtr *int64
	if isDigits(candidate) {
		if val, err := strconv.ParseInt(candidate, 10, 64); err == nil {
			numericPtr = &val
		}
	}

	*id = CamaraID{
		raw:     raw,
		numeric: numericPtr,
	}

	return nil
}

func (id CamaraID) String() string {
	return id.raw
}

func (id CamaraID) HasInt64() bool {
	return id.numeric != nil
}

func (id CamaraID) Int64() int64 {
	if id.numeric == nil {
		return 0
	}
	return *id.numeric
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

var (
	jitterRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	jitterMu   sync.Mutex
)

func randomIntn(n int) int {
	if n <= 0 {
		return 0
	}
	jitterMu.Lock()
	v := jitterRand.Intn(n)
	jitterMu.Unlock()
	return v
}

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

type deputadoResumo struct {
	ID   int    `json:"id"`
	Nome string `json:"nome"`
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
	cbConfig := resilience.CircuitBreakerConfig{
		MaxFailures:      8,                 // Mais tolerante para API da Câmara
		ResetTimeout:     120 * time.Second, // 2 minutos para retry (API pode estar sobrecarregada)
		SuccessThreshold: 3,                 // 3 sucessos para fechar
		Timeout:          timeout,           // Mesmo timeout do HTTP client
	}

	return &CamaraClient{
		httpClient:     &http.Client{Timeout: timeout},
		baseURL:        baseURL,
		limiter:        rate.NewLimiter(rate.Limit(float64(rps)), burst),
		circuitBreaker: resilience.NewCircuitBreaker(cbConfig),
	}
}

// NewCamaraClientFromConfig creates a client from config
func NewCamaraClientFromConfig(cfg *config.CamaraClientConfig) *CamaraClient {
	// Configuração otimizada para API da Câmara baseada nos logs de timeout
	circuitConfig := resilience.CircuitBreakerConfig{
		MaxFailures:      8,                 // Mais tolerante: 8 falhas para abrir
		ResetTimeout:     120 * time.Second, // 2 minutos para retry
		SuccessThreshold: 3,                 // 3 sucessos para fechar
		Timeout:          cfg.Timeout,       // Timeout da configuração
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
			backoff = 500 * time.Millisecond // Backoff inicial maior (vs 200ms)
		}

		maxRetries := c.maxRetries
		if maxRetries <= 0 {
			maxRetries = 2 // Menos tentativas para evitar circuit breaker (vs 3)
		}

		for i := 0; i < maxRetries; i++ {
			resp, err := c.httpClient.Do(req)
			if err != nil {
				lastErr = fmt.Errorf("erro na requisição HTTP (tentativa %d/%d): %w", i+1, maxRetries, err)
				if i < maxRetries-1 {
					// Back-pressure: se detectar sobrecarga da API, aguardar mais tempo
					delay := backoff
					if isAPIOverloaded(err) {
						delay = backoff * 2 // Dobrar delay se API estiver sobrecarregada
					}
					time.Sleep(delay)
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

	// Estratégia de chunking adaptativo mais conservadora baseada nos logs de timeout.
	// Começamos com períodos menores e ajustamos com cuidado para evitar circuit breaker.
	// Após mais sucessos consecutivos aumentamos gradualmente o tamanho para recuperar throughput.
	chunkLevels := []int{14, 7, 3, 1} // Começar com 14 dias (mais conservador)
	levelIndex := 0
	successStreak := 0
	errorStreak := 0
	const minSuccessBeforeLevelExpand = 10    // Mais sucessos antes de expandir (vs 6)
	const minSecondsBetweenLevelChanges = 180 // 3 minutos entre mudanças (vs 90s)
	lastLevelChange := time.Now()
	dayAttempts := make(map[string]int)
	var skippedDays []string
	const maxRetryPerDay = 2                  // Menos tentativas por dia (vs 3)
	const minDailyCooldown = 15 * time.Second // Cooldown maior (vs 8s)
	chunkProgress := domain.GetVotacoesChunkProgress(ctx)

	fetchRange := func(rangeStart, rangeEnd time.Time, itensPagina int, delayBetweenPages time.Duration) error {
		params := url.Values{}
		params.Add("dataInicio", rangeStart.In(loc).Format("2006-01-02"))
		params.Add("dataFim", rangeEnd.In(loc).Format("2006-01-02"))
		params.Add("ordenarPor", "dataHoraRegistro")
		params.Add("ordem", "DESC")
		params.Add("itens", strconv.Itoa(itensPagina))

		pageURL := fmt.Sprintf("%s/votacoes?%s", c.baseURL, params.Encode())

		for {
			body, err := c.doRequest(ctx, http.MethodGet, pageURL)
			if err != nil {
				return err
			}

			var response struct {
				Dados []struct {
					ID                 CamaraID `json:"id"`
					URI                string   `json:"uri"`
					Titulo             string   `json:"titulo"`
					TipoVotacao        string   `json:"tipoVotacao"`
					Aprovacao          int      `json:"aprovacao"`
					DataHoraInicio     string   `json:"dataHoraInicio"`
					DataHoraRegistro   string   `json:"dataHoraRegistro"`
					DataHoraFim        string   `json:"dataHoraFim"`
					Descricao          string   `json:"descricao"`
					ProposicaoObjeto   string   `json:"proposicaoObjeto"`
					UltimaApresentacao struct {
						ID         CamaraID `json:"id"`
						Proposicao struct {
							ID     CamaraID `json:"id"`
							Numero string   `json:"numero"`
							Ano    int      `json:"ano"`
							Tipo   string   `json:"tipo"`
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
				if item.UltimaApresentacao.Proposicao.ID.HasInt64() {
					id := item.UltimaApresentacao.Proposicao.ID.Int64()
					proposicaoID = &id
				}

				var anoProposicao *int
				if item.UltimaApresentacao.Proposicao.Ano != 0 {
					anoProposicao = &item.UltimaApresentacao.Proposicao.Ano
				}

				var numericID *int64
				if item.ID.HasInt64() {
					v := item.ID.Int64()
					numericID = &v
				}

				votacao := &domain.Votacao{
					IDCamara:              item.ID.String(),
					IDVotacaoCamara:       numericID,
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
			if delayBetweenPages > 0 {
				time.Sleep(delayBetweenPages)
			} else {
				time.Sleep(50 * time.Millisecond)
			}
		}

		return nil
	}

	// Cursor para o dia atual a processar
	for cur := start; !cur.After(end); {
		dayKey := cur.Format("2006-01-02")
		chunkDays := chunkLevels[levelIndex]
		candidateEnd := cur.AddDate(0, 0, chunkDays-1)
		if candidateEnd.After(end) {
			candidateEnd = end
		}

		itensPagina := 100
		switch {
		case chunkDays <= 1:
			itensPagina = 10
		case chunkDays <= 3:
			itensPagina = 30
		case chunkDays <= 7:
			itensPagina = 50
		case chunkDays <= 14:
			itensPagina = 80
		}

		delayBetweenPages := c.pageDelay
		if delayBetweenPages <= 0 {
			delayBetweenPages = 50 * time.Millisecond
		}
		if levelIndex >= 2 && delayBetweenPages < 120*time.Millisecond {
			delayBetweenPages = 120 * time.Millisecond
		}
		if levelIndex >= 3 && delayBetweenPages < 180*time.Millisecond {
			delayBetweenPages = 180 * time.Millisecond
		}
		if levelIndex >= len(chunkLevels)-1 && delayBetweenPages < 600*time.Millisecond {
			delayBetweenPages = 600 * time.Millisecond
		}
		if chunkDays <= 1 && delayBetweenPages < 600*time.Millisecond {
			delayBetweenPages = 600 * time.Millisecond
		}

		err := fetchRange(cur, candidateEnd, itensPagina, delayBetweenPages)
		if err != nil {
			metrics.IncCamaraChunkFailure(chunkDays)

			if isRetryableServerError(err) {
				errorStreak++
				successStreak = 0
				if levelIndex < len(chunkLevels)-1 {
					nextLevel := levelIndex + 1
					fmt.Printf("WARN: chunk %s - %s falhou com erro transitório: %v. Reduzindo janela para %d dias...\n",
						cur.Format("2006-01-02"), candidateEnd.Format("2006-01-02"), err, chunkLevels[nextLevel])
					levelIndex = nextLevel
					lastLevelChange = time.Now()
				} else {
					dayAttempts[dayKey]++
					attempt := dayAttempts[dayKey]
					if attempt < maxRetryPerDay {
						cooldown := time.Duration(1<<min(attempt-1, 4)) * 5 * time.Second
						if cooldown < minDailyCooldown {
							cooldown = minDailyCooldown
						}
						jitterRange := 1500
						if levelIndex == len(chunkLevels)-1 {
							jitterRange = 4000
						}
						jitter := time.Duration(randomIntn(jitterRange)) * time.Millisecond
						fmt.Printf("WARN: todos os tamanhos de chunk falharam para %s (tentativa %d/%d). Aguardando %s antes de tentar novamente com janela mínima...\n",
							dayKey, attempt, maxRetryPerDay, (cooldown + jitter).Round(time.Millisecond))
						levelIndex = len(chunkLevels) - 1
						effectiveCooldown := cooldown
						if levelIndex == len(chunkLevels)-1 && effectiveCooldown < 20*time.Second {
							effectiveCooldown = 20 * time.Second
						}
						time.Sleep(effectiveCooldown + jitter)
						continue
					}
					fmt.Printf("ERROR: esgotadas %d tentativas para %s; registrando para retry futuro e seguindo em frente\n", attempt, dayKey)
					skippedDays = append(skippedDays, dayKey)
					if chunkProgress != nil {
						chunkProgress(cur, cur, false)
					}
					cur = cur.AddDate(0, 0, 1)
					errorStreak = 0
					successStreak = 0
					time.Sleep(250 * time.Millisecond)
					continue
				}

				baseBackoff := time.Duration(1<<min(errorStreak-1, 5)) * time.Second
				if baseBackoff < 5*time.Second {
					baseBackoff = 5 * time.Second
				}
				if levelIndex == len(chunkLevels)-1 && baseBackoff < 15*time.Second {
					baseBackoff = 15 * time.Second
				}
				jitterRange := 750
				if levelIndex == len(chunkLevels)-1 {
					jitterRange = 2000
				}
				jitter := time.Duration(randomIntn(jitterRange)) * time.Millisecond
				time.Sleep(baseBackoff + jitter)
				continue
			}
			return nil, fmt.Errorf("erro ao buscar votações: %w", err)
		}
		metrics.IncCamaraChunkSuccess(chunkDays)
		if chunkProgress != nil {
			chunkProgress(cur, candidateEnd, true)
		}

		delete(dayAttempts, dayKey)
		cur = candidateEnd.AddDate(0, 0, 1)
		successStreak++
		errorStreak = 0

		var postDelay time.Duration
		switch {
		case chunkDays <= 1:
			postDelay = 1200 * time.Millisecond
		case chunkDays <= 3:
			postDelay = 800 * time.Millisecond
		case chunkDays <= 7:
			postDelay = 480 * time.Millisecond
		default:
			postDelay = time.Duration(levelIndex+1) * 80 * time.Millisecond
			if postDelay < 80*time.Millisecond {
				postDelay = 80 * time.Millisecond
			}
		}
		time.Sleep(postDelay)

		if successStreak >= minSuccessBeforeLevelExpand && levelIndex > 0 {
			if time.Since(lastLevelChange) >= minSecondsBetweenLevelChanges*time.Second {
				levelIndex--
				successStreak = 0
				lastLevelChange = time.Now()
				fmt.Printf("INFO: aumentando janela para %d dias após sequência de sucesso\n", chunkLevels[levelIndex])
			}
		}
	}

	if len(skippedDays) > 0 {
		return all, fmt.Errorf("falha ao coletar votações em %d intervalos (%s)", len(skippedDays), strings.Join(skippedDays, ", "))
	}

	return all, nil
}

// GetVotacao busca uma votação específica por ID externo
func (c *CamaraClient) GetVotacao(ctx context.Context, id string) (*domain.Votacao, error) {
	urlStr := fmt.Sprintf("%s/votacoes/%s", c.baseURL, id)

	body, err := c.doRequest(ctx, http.MethodGet, urlStr)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar votação ID %s: %w", id, err)
	}

	var response struct {
		Dados struct {
			ID                 CamaraID `json:"id"`
			URI                string   `json:"uri"`
			Titulo             string   `json:"titulo"`
			TipoVotacao        string   `json:"tipoVotacao"`
			Aprovacao          int      `json:"aprovacao"`
			DataHoraInicio     string   `json:"dataHoraInicio"`
			DataHoraRegistro   string   `json:"dataHoraRegistro"`
			DataHoraFim        string   `json:"dataHoraFim"`
			Descricao          string   `json:"descricao"`
			ProposicaoObjeto   string   `json:"proposicaoObjeto"`
			UltimaApresentacao struct {
				ID         CamaraID `json:"id"`
				Proposicao struct {
					ID     CamaraID `json:"id"`
					Numero string   `json:"numero"`
					Ano    int      `json:"ano"`
					Tipo   string   `json:"tipo"`
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
	if item.UltimaApresentacao.Proposicao.ID.HasInt64() {
		id := item.UltimaApresentacao.Proposicao.ID.Int64()
		proposicaoID = &id
	}

	var anoProposicao *int
	if item.UltimaApresentacao.Proposicao.Ano != 0 {
		anoProposicao = &item.UltimaApresentacao.Proposicao.Ano
	}

	var numericID *int64
	if item.ID.HasInt64() {
		v := item.ID.Int64()
		numericID = &v
	}

	votacao := &domain.Votacao{
		IDCamara:              item.ID.String(),
		IDVotacaoCamara:       numericID,
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
func (c *CamaraClient) GetVotosPorVotacao(ctx context.Context, idVotacao string) ([]*domain.VotoDeputado, error) {
	urlStr := fmt.Sprintf("%s/votacoes/%s/votos", c.baseURL, idVotacao)

	body, err := c.doRequest(ctx, http.MethodGet, urlStr)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar votos da votação %s: %w", idVotacao, err)
	}

	var response struct {
		Dados []struct {
			Deputado       deputadoResumo `json:"deputado"`
			DeputadoLegacy deputadoResumo `json:"deputado_"`
			TipoVoto       string         `json:"tipoVoto"`
			Voto           string         `json:"voto"`
			DataRegistro   string         `json:"dataRegistroVoto"`
		} `json:"dados"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("erro ao fazer parse dos votos: %w", err)
	}

	votos := make([]*domain.VotoDeputado, 0, len(response.Dados))

	for _, item := range response.Dados {
		deputadoInfo, fonteDeputado := pickDeputadoInfo(item.Deputado, item.DeputadoLegacy)

		votoStr := strings.TrimSpace(item.Voto)
		if votoStr == "" {
			votoStr = strings.TrimSpace(item.TipoVoto)
		}

		if deputadoInfo.ID == 0 || votoStr == "" {
			continue
		}

		voto := &domain.VotoDeputado{
			IDVotacao:     0,
			IDDeputado:    deputadoInfo.ID,
			Voto:          votoStr,
			Justificativa: nil, // API da Câmara não retorna justificativa nos votos
			Payload: map[string]interface{}{
				"tipoVoto":      item.TipoVoto,
				"nomeDeputado":  deputadoInfo.Nome,
				"dataRegistro":  item.DataRegistro,
				"fonteDeputado": fonteDeputado,
			},
			CreatedAt: time.Now(),
		}

		votos = append(votos, voto)
	}

	return votos, nil
}

func pickDeputadoInfo(primary, legacy deputadoResumo) (deputadoResumo, string) {
	if primary.ID != 0 || primary.Nome != "" {
		return primary, "deputado"
	}
	if legacy.ID != 0 || legacy.Nome != "" {
		return legacy, "deputado_"
	}
	return deputadoResumo{}, ""
}

// GetOrientacoesPorVotacao busca orientações partidárias para uma votação
func (c *CamaraClient) GetOrientacoesPorVotacao(ctx context.Context, idVotacao string) ([]*domain.OrientacaoPartido, error) {
	urlStr := fmt.Sprintf("%s/votacoes/%s/orientacoes", c.baseURL, idVotacao)

	body, err := c.doRequest(ctx, http.MethodGet, urlStr)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar orientações da votação %s: %w", idVotacao, err)
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
			IDVotacao:  0,
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
