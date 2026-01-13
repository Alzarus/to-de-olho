package ceaps

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler gerencia endpoints REST de despesas CEAPS
type Handler struct {
	repo *Repository
}

// NewHandler cria um novo handler
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// ListBySenador godoc
// @Summary Lista despesas de um senador
// @Tags despesas
// @Produce json
// @Param senador_id path int true "ID do senador"
// @Param ano query int false "Ano de referencia"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/senadores/{senador_id}/despesas [get]
func (h *Handler) ListBySenador(c *gin.Context) {
	senadorIDStr := c.Param("id")
	senadorID, err := strconv.Atoi(senadorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalido"})
		return
	}

	// Parametro opcional: ano
	var ano *int
	if anoStr := c.Query("ano"); anoStr != "" {
		if anoVal, err := strconv.Atoi(anoStr); err == nil {
			ano = &anoVal
		}
	}

	despesas, err := h.repo.FindBySenadorID(senadorID, ano)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao buscar despesas"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"senador_id": senadorID,
		"total":      len(despesas),
		"despesas":   despesas,
	})
}

// AggregateBySenador godoc
// @Summary Retorna gastos agregados por tipo de despesa
// @Tags despesas
// @Produce json
// @Param senador_id path int true "ID do senador"
// @Param ano query int false "Ano de referencia"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/senadores/{senador_id}/despesas/agregado [get]
func (h *Handler) AggregateBySenador(c *gin.Context) {
	senadorIDStr := c.Param("id")
	senadorID, err := strconv.Atoi(senadorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalido"})
		return
	}

	var ano *int
	if anoStr := c.Query("ano"); anoStr != "" {
		if anoVal, err := strconv.Atoi(anoStr); err == nil {
			ano = &anoVal
		}
	}

	agregados, err := h.repo.AggregateByTipo(senadorID, ano)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao agregar despesas"})
		return
	}

	// Calcular total geral
	var totalGeral float64
	for _, a := range agregados {
		totalGeral += a.Total
	}

	c.JSON(http.StatusOK, gin.H{
		"senador_id":  senadorID,
		"total_geral": totalGeral,
		"por_tipo":    agregados,
	})
}
