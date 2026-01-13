package proposicao

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler gerencia endpoints REST de proposicoes
type Handler struct {
	repo *Repository
}

// NewHandler cria um novo handler
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// ListBySenador godoc
// @Summary Lista proposicoes de um senador
// @Tags proposicoes
// @Produce json
// @Param id path int true "ID do senador"
// @Param limit query int false "Limite de resultados (default 50)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/senadores/{id}/proposicoes [get]
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

	proposicoes, err := h.repo.FindBySenadorID(senadorID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao buscar proposicoes"})
		return
	}

	total, _ := h.repo.CountBySenadorID(senadorID)

	c.JSON(http.StatusOK, gin.H{
		"senador_id":   senadorID,
		"total":        total,
		"limit":        limit,
		"proposicoes":  proposicoes,
	})
}

// GetStats godoc
// @Summary Retorna estatisticas de proposicoes de um senador
// @Tags proposicoes
// @Produce json
// @Param id path int true "ID do senador"
// @Success 200 {object} ProposicaoStats
// @Router /api/v1/senadores/{id}/proposicoes/stats [get]
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

// GetPorTipo godoc
// @Summary Retorna contagem de proposicoes por tipo (PEC/PLP/PL)
// @Tags proposicoes
// @Produce json
// @Param id path int true "ID do senador"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/senadores/{id}/proposicoes/tipos [get]
func (h *Handler) GetPorTipo(c *gin.Context) {
	senadorIDStr := c.Param("id")
	senadorID, err := strconv.Atoi(senadorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalido"})
		return
	}

	tipos, err := h.repo.GetProposicoesPorTipo(senadorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao buscar tipos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"senador_id": senadorID,
		"por_tipo":   tipos,
	})
}
