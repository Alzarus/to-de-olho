package comissao

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler gerencia endpoints REST de comissoes
type Handler struct {
	repo *Repository
}

// NewHandler cria um novo handler
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// ListBySenador godoc
// @Summary Lista comissoes de um senador
// @Tags comissoes
// @Produce json
// @Param id path int true "ID do senador"
// @Param limit query int false "Limite de resultados (default 50)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/senadores/{id}/comissoes [get]
func (h *Handler) ListBySenador(c *gin.Context) {
	senadorIDStr := c.Param("id")
	senadorID, err := strconv.Atoi(senadorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalido"})
		return
	}

	// Limite opcional (default 50)
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	comissoes, err := h.repo.FindBySenadorID(senadorID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao buscar comissoes"})
		return
	}

	total, _ := h.repo.CountBySenadorID(senadorID)

	c.JSON(http.StatusOK, gin.H{
		"senador_id": senadorID,
		"total":      total,
		"limit":      limit,
		"comissoes":  comissoes,
	})
}

// GetAtivas godoc
// @Summary Lista comissoes ativas de um senador
// @Tags comissoes
// @Produce json
// @Param id path int true "ID do senador"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/senadores/{id}/comissoes/ativas [get]
func (h *Handler) GetAtivas(c *gin.Context) {
	senadorIDStr := c.Param("id")
	senadorID, err := strconv.Atoi(senadorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalido"})
		return
	}

	comissoes, err := h.repo.FindAtivasBySenadorID(senadorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao buscar comissoes ativas"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"senador_id": senadorID,
		"total":      len(comissoes),
		"comissoes":  comissoes,
	})
}

// GetStats godoc
// @Summary Retorna estatisticas de comissoes de um senador
// @Tags comissoes
// @Produce json
// @Param id path int true "ID do senador"
// @Success 200 {object} ComissaoStats
// @Router /api/v1/senadores/{id}/comissoes/stats [get]
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

// GetPorCasa godoc
// @Summary Retorna contagem de comissoes por casa (SF/CN)
// @Tags comissoes
// @Produce json
// @Param id path int true "ID do senador"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/senadores/{id}/comissoes/casas [get]
func (h *Handler) GetPorCasa(c *gin.Context) {
	senadorIDStr := c.Param("id")
	senadorID, err := strconv.Atoi(senadorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalido"})
		return
	}

	casas, err := h.repo.GetComissoesPorCasa(senadorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao buscar casas"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"senador_id": senadorID,
		"por_casa":   casas,
	})
}
