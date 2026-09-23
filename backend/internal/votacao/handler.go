package votacao

import (
	"net/http"
	"regexp"
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
	ano, _ := strconv.Atoi(c.Query("ano"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	votacoes, total, err := h.repo.FindBySenadorID(senadorID, limit, offset, votoType, ano)
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
// @Param sessao query string false "Codigo da sessao (todas as votacoes dela)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/votacoes [get]
func (h *Handler) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	ano, _ := strconv.Atoi(c.Query("ano"))
	materia := c.Query("materia")
	ordem := c.DefaultQuery("ordem", "desc")
	sessao := c.Query("sessao")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	votacoes, total, err := h.repo.FindAll(limit, offset, ano, materia, ordem, sessao)
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

// idLegado casa o formato antigo de id de votacao, "codigoSessao_ano"
var idLegado = regexp.MustCompile(`^(\d+)_\d+$`)

// GetByID godoc
// @Summary Retorna detalhes de uma votacao e lista de votos
// @Tags votacoes
// @Produce json
// @Param id path int true "Codigo da votacao (codigoSessaoVotacao)"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{} "nao encontrada; para id legado NNNNNN_AAAA, traz sessao_id"
// @Router /api/v1/votacoes/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")

	// Links antigos apontavam para a sessao inteira (D2): o cliente redireciona
	// para a lista de votacoes da sessao
	if m := idLegado.FindStringSubmatch(id); m != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "id de votacao legado", "sessao_id": m[1]})
		return
	}

	codigo, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id invalido"})
		return
	}

	votacao, err := h.repo.FindByID(codigo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "votacao nao encontrada"})
		return
	}

	votos, err := h.repo.FindVotosByCodigoVotacao(codigo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao buscar votos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"votacao": votacao,
		"votos":   votos,
	})
}
