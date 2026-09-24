package proposicao

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Alzarus/to-de-olho/internal/utils"
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
// @Param limit query int false "Limite de resultados (default 20)"
// @Param page query int false "Pagina (default 1)"
// @Param q query string false "Termo de busca"
// @Param ano query int false "Ano da materia"
// @Param sigla query string false "Sigla do subtipo (PEC, PL, etc)"
// @Param sort query string false "Ordenacao (data_desc, data_asc, ano_desc)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/senadores/{id}/proposicoes [get]
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
	sigla := c.Query("sigla")
	status := c.Query("status")
	sort := c.Query("sort")
	
	var ano int
	if anoStr := c.Query("ano"); anoStr != "" {
		if a, err := strconv.Atoi(anoStr); err == nil {
			ano = a
		}
	}

	proposicoes, total, err := h.repo.FindBySenadorID(senadorID, limit, offset, queryStr, ano, sigla, status, sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao buscar proposicoes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"senador_id":   senadorID,
		"total":        total,
		"limit":        limit,
		"page":         page,
		"total_pages":  (int(total) + limit - 1) / limit,
		"proposicoes":  utils.NaoNulo(proposicoes),
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
		"por_tipo":   utils.NaoNulo(tipos),
	})
}
