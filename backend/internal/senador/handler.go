package senador

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler gerencia endpoints REST de senadores
type Handler struct {
	repo *Repository
}

// NewHandler cria um novo handler
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// ListAll godoc
// @Summary Lista todos os senadores (permite inativos)
// @Tags senadores
// @Produce json
// @Param inativos query string false "Incluir senadores fora de exercicio"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/senadores [get]
func (h *Handler) ListAll(c *gin.Context) {
	includeInactive := c.Query("inativos") == "true"
	senadores, err := h.repo.FindAll(includeInactive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao buscar senadores"})
		return
	}

	count, _ := h.repo.Count()

	c.JSON(http.StatusOK, gin.H{
		"total":     count,
		"senadores": senadores,
	})
}

// GetByID godoc
// @Summary Busca senador por ID
// @Tags senadores
// @Produce json
// @Param id path int true "ID do senador"
// @Success 200 {object} Senador
// @Failure 404 {object} map[string]string
// @Router /api/v1/senadores/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalido"})
		return
	}

	senador, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "senador nao encontrado"})
		return
	}

	c.JSON(http.StatusOK, senador)
}

// GetByCodigo godoc
// @Summary Busca senador por codigo parlamentar
// @Tags senadores
// @Produce json
// @Param codigo path int true "Codigo parlamentar"
// @Success 200 {object} Senador
// @Failure 404 {object} map[string]string
// @Router /api/v1/senadores/codigo/{codigo} [get]
func (h *Handler) GetByCodigo(c *gin.Context) {
	codigoStr := c.Param("codigo")
	codigo, err := strconv.Atoi(codigoStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "codigo invalido"})
		return
	}

	senador, err := h.repo.FindByCodigo(codigo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "senador nao encontrado"})
		return
	}

	c.JSON(http.StatusOK, senador)
}
