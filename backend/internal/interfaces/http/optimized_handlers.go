package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"to-de-olho-backend/internal/application"
	"to-de-olho-backend/internal/domain"
	"to-de-olho-backend/internal/interfaces/http/middleware"

	"github.com/gin-gonic/gin"
)

// OptimizedHandlers handlers otimizados para performance
type OptimizedHandlers struct {
	deputadosService   *application.DeputadosService
	proposicoesService *application.ProposicoesService
	analyticsService   *application.AnalyticsService
}

// NewOptimizedHandlers cria handlers otimizados
func NewOptimizedHandlers(
	deputadosService *application.DeputadosService,
	proposicoesService *application.ProposicoesService,
	analyticsService *application.AnalyticsService,
) *OptimizedHandlers {
	return &OptimizedHandlers{
		deputadosService:   deputadosService,
		proposicoesService: proposicoesService,
		analyticsService:   analyticsService,
	}
}

// ListDeputadosOptimized lista deputados com otimizações avançadas
func (h *OptimizedHandlers) ListDeputadosOptimized(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()

	// Parse pagination
	var paginationReq domain.PaginationRequest
	if err := c.ShouldBindQuery(&paginationReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Parâmetros de paginação inválidos",
		})
		return
	}
	paginationReq.ValidateAndNormalize()

	// Filtros
	uf := c.Query("uf")
	partido := c.Query("partido")
	nome := c.Query("nome")

	// Cache key
	cacheKey := buildCacheKey("deputados", uf, partido, nome, &paginationReq)

	// Verificar cache primeiro
	if cachedData, found := h.checkCache(ctx, cacheKey); found {
		h.sendOptimizedResponse(c, cachedData, true, startTime)
		return
	}

	// Buscar todos os dados primeiro para aplicar paginação
	allDeputados, _, err := h.deputadosService.ListarDeputados(
		ctx, partido, uf, nome,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao buscar deputados",
		})
		return
	}

	// Aplicar paginação manual nos dados
	total := int64(len(allDeputados))

	// Calcular offset e limite
	offset := (paginationReq.Page - 1) * paginationReq.Limit
	end := offset + paginationReq.Limit

	var deputados []domain.Deputado
	if offset < len(allDeputados) {
		if end > len(allDeputados) {
			end = len(allDeputados)
		}
		deputados = allDeputados[offset:end]
	} else {
		deputados = []domain.Deputado{}
	}

	// Build response
	response := domain.BuildPagination(&paginationReq, total, deputados)

	// Cache result
	h.cacheResult(ctx, cacheKey, response, 5*time.Minute)

	// Send optimized response
	h.sendOptimizedResponse(c, response, false, startTime)
}

// StreamDeputados stream de deputados para grandes volumes
func (h *OptimizedHandlers) StreamDeputados(c *gin.Context) {
	ctx := c.Request.Context()

	// Setup streaming headers
	c.Header("Content-Type", "application/json")
	c.Header("Transfer-Encoding", "chunked")
	c.Header("Cache-Control", "no-cache")

	// Parse filters
	uf := c.Query("uf")
	partido := c.Query("partido")

	// Stream deputados em chunks
	deputados, _, err := h.deputadosService.ListarDeputados(ctx, partido, uf, "")
	if err != nil {
		c.Writer.WriteString(`],"error":"Erro no streaming"}`)
		return
	}

	isFirst := true
	for _, deputado := range deputados {
		if !isFirst {
			c.Writer.WriteString(",")
		}
		isFirst = false

		data, err := json.Marshal(deputado)
		if err != nil {
			continue
		}

		c.Writer.Write(data)

		// Flush periodicamente
		if flusher, ok := c.Writer.(http.Flusher); ok {
			flusher.Flush()
		}
	}

	c.Writer.WriteString(`]}`)

	// Final flush
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

// GetDeputadoOptimized busca deputado otimizada
func (h *OptimizedHandlers) GetDeputadoOptimized(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()

	deputadoID := c.Param("id")
	if deputadoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID do deputado é obrigatório",
		})
		return
	}

	// Cache check
	cacheKey := "deputado:" + deputadoID
	if cachedData, found := h.checkCache(ctx, cacheKey); found {
		h.sendOptimizedResponse(c, cachedData, true, startTime)
		return
	}

	// Buscar deputado
	deputado, _, err := h.deputadosService.BuscarDeputadoPorID(ctx, deputadoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Deputado não encontrado",
		})
		return
	}

	// Cache por 1 hora
	h.cacheResult(ctx, cacheKey, deputado, time.Hour)

	h.sendOptimizedResponse(c, deputado, false, startTime)
}

// GetAnalyticsOptimized analytics otimizadas
func (h *OptimizedHandlers) GetAnalyticsOptimized(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()

	tipoAnalise := c.Query("tipo")
	periodo := c.Query("periodo")

	// Cache key
	cacheKey := buildCacheKey("analytics", tipoAnalise, periodo)

	// Check cache
	if cachedData, found := h.checkCache(ctx, cacheKey); found {
		h.sendOptimizedResponse(c, cachedData, true, startTime)
		return
	}

	// Executar análise com base no tipo
	var resultado interface{}
	var err error

	switch tipoAnalise {
	case "gastos":
		resultado, _, err = h.analyticsService.GetRankingGastos(ctx, 2024, 100)
	case "proposicoes":
		resultado, _, err = h.analyticsService.GetRankingProposicoes(ctx, 2024, 100)
	case "presenca":
		resultado, _, err = h.analyticsService.GetRankingPresenca(ctx, 2024, 100)
	default:
		resultado, _, err = h.analyticsService.GetRankingGastos(ctx, 2024, 100)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao executar análise",
		})
		return
	}

	// Cache por 30 minutos
	h.cacheResult(ctx, cacheKey, resultado, 30*time.Minute)

	h.sendOptimizedResponse(c, resultado, false, startTime)
}

// Helper methods

func (h *OptimizedHandlers) checkCache(ctx context.Context, key string) (interface{}, bool) {
	// Implementar check de cache multi-level
	// Por simplicidade, retornando false
	return nil, false
}

func (h *OptimizedHandlers) cacheResult(ctx context.Context, key string, data interface{}, ttl time.Duration) {
	// Implementar cache multi-level
	// Por simplicidade, não implementando
}

func (h *OptimizedHandlers) sendOptimizedResponse(c *gin.Context, data interface{}, cacheHit bool, startTime time.Time) {
	processTime := time.Since(startTime)

	// Headers de performance
	c.Header("X-Process-Time", processTime.String())
	c.Header("X-Cache-Hit", strconv.FormatBool(cacheHit))

	// Verificar se deve usar compressão
	if shouldCompress(c.Request, data) {
		c.Header("Content-Encoding", "gzip")
	}

	// Response otimizada
	response := map[string]interface{}{
		"data": data,
		"meta": map[string]interface{}{
			"process_time": processTime.Milliseconds(),
			"cache_hit":    cacheHit,
			"timestamp":    time.Now().Unix(),
		},
	}

	// Usar streaming se necessário
	if shouldUseStreaming(c.Request) {
		middleware.WriteStreamingJSON(c, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

func buildCacheKey(prefix string, params ...interface{}) string {
	key := prefix
	for _, param := range params {
		if param != nil && param != "" {
			key += ":" + toString(param)
		}
	}
	return key
}

func toString(v interface{}) string {
	if v == nil {
		return "nil"
	}

	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	case *domain.PaginationRequest:
		if val == nil {
			return "pagination:nil"
		}
		return fmt.Sprintf("pagination:page=%d,limit=%d,sort=%s,order=%s,cursor=%s",
			val.Page, val.Limit, val.SortBy, val.Order, val.Cursor)
	default:
		// Para outros tipos, usar reflexão para criar uma chave única
		return fmt.Sprintf("type:%T,value:%+v", v, v)
	}
}

func shouldCompress(req *http.Request, data interface{}) bool {
	// Comprimir apenas responses grandes
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return false
	}

	return len(dataBytes) > 1024 // > 1KB
}

func shouldUseStreaming(req *http.Request) bool {
	// Streaming para grandes datasets
	limitStr := req.URL.Query().Get("limit")
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 100 {
			return true
		}
	}

	return req.URL.Query().Get("stream") == "true"
}
