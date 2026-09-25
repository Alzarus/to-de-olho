package comissao

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Alzarus/to-de-olho/internal/utils"
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
// @Param limit query int false "Limite de resultados (default 20)"
// @Param page query int false "Pagina (default 1)"
// @Param q query string false "Termo de busca"
// @Param status query string false "Status (ativa/inativa)"
// @Param participacao query string false "Tipo (Titular/Suplente)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/senadores/{id}/comissoes [get]
func (h *Handler) ListBySenador(c *gin.Context) {
	senadorIDStr := c.Param("id")
	senadorID, err := strconv.Atoi(senadorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalido"})
		return
	}

	// Paginação com teto (ver utils.LerPaginacao): antes ?limit= não tinha
	// limite e ?limit=100000 devolvia a tabela inteira de uma vez.
	pag := utils.LerPaginacao(c.Query("limit"), c.Query("page"), 20, utils.LimiteMaximoPadrao)
	limit, page, offset := pag.Limit, pag.Page, pag.Offset

	queryStr := c.Query("q")
	status := c.Query("status")
	participacao := c.Query("participacao")

	comissoes, total, err := h.repo.FindBySenadorID(senadorID, limit, offset, queryStr, status, participacao)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao buscar comissoes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"senador_id":   senadorID,
		"total":        total,
		"limit":        limit,
		"page":         page,
		"total_pages":  (int(total) + limit - 1) / limit,
		"comissoes":    utils.NaoNulo(comissoes),
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
		"comissoes":  utils.NaoNulo(comissoes),
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
		"por_casa":   utils.NaoNulo(casas),
	})
}
