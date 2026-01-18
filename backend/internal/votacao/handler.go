package votacao

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler gerencia endpoints REST de votacoes
type Handler struct {
	repo *Repository
}

// NewHandler cria um novo handler
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// ListBySenador godoc
// @Summary Lista votacoes de um senador
// @Tags votacoes
// @Produce json
// @Param id path int true "ID do senador"
// @Param limit query int false "Limite de resultados (default 50)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/senadores/{id}/votacoes [get]
func (h *Handler) ListBySenador(c *gin.Context) {
	senadorIDStr := c.Param("id")
	senadorID, err := strconv.Atoi(senadorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalido"})
		return
	}

	// Parametros de paginacao
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	votoType := c.Query("voto")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	votacoes, total, err := h.repo.FindBySenadorID(senadorID, limit, offset, votoType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao buscar votacoes"})
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, gin.H{
		"senador_id":  senadorID,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
		"votacoes":    votacoes,
	})
}

// GetStats godoc
// @Summary Retorna estatisticas de votacao de um senador
// @Tags votacoes
// @Produce json
// @Param id path int true "ID do senador"
// @Success 200 {object} VotacaoStats
// @Router /api/v1/senadores/{id}/votacoes/stats [get]
func (h *Handler) GetStats(c *gin.Context) {
	senadorIDStr := c.Param("id")
	senadorID, err := strconv.Atoi(senadorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalido"})
		return
	}

	stats, err := h.repo.GetStats(senadorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao calcular estatisticas"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetVotosPorTipo godoc
// @Summary Retorna contagem de votos por tipo
// @Tags votacoes
// @Produce json
// @Param id path int true "ID do senador"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/senadores/{id}/votacoes/tipos [get]
func (h *Handler) GetVotosPorTipo(c *gin.Context) {
	senadorIDStr := c.Param("id")
	senadorID, err := strconv.Atoi(senadorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalido"})
		return
	}

	tipos, err := h.repo.GetVotosPorTipo(senadorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao buscar tipos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"senador_id": senadorID,
		"por_tipo":   tipos,
	})
}

// GetAll godoc
// @Summary Lista todas as votacoes (agrupadas por sessao)
// @Tags votacoes
// @Produce json
// @Param page query int false "Pagina (default 1)"
// @Param limit query int false "Limite (default 20)"
// @Param ano query int false "Ano (default atual)"
// @Param materia query string false "Filtro por materia/descricao"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/votacoes [get]
func (h *Handler) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	ano, _ := strconv.Atoi(c.Query("ano"))
	materia := c.Query("materia")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	votacoes, total, err := h.repo.FindAll(limit, offset, ano, materia)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao buscar votacoes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  votacoes,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetByID godoc
// @Summary Retorna detalhes de uma votacao e lista de votos
// @Tags votacoes
// @Produce json
// @Param id path string true "ID da Sessao de Votacao"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/votacoes/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")

	// Metadata da votacao
	votacao, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "votacao nao encontrada"})
		return
	}

	// Lista de votos
	votos, err := h.repo.FindVotosBySessaoID(id)
	if err != nil {
		// DEBUG: Exposing error details to frontend/curl
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao buscar votos", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"votacao": votacao,
		"votos":   votos,
	})
}
