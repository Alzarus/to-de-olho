package votacao

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/Alzarus/to-de-olho/internal/utils"
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

	// Paginação com teto de limit e de offset (ver utils.LerPaginacao)
	pag := utils.LerPaginacao(c.Query("limit"), c.Query("page"), 20, utils.LimiteMaximoPadrao)
	limit, page, offset := pag.Limit, pag.Page, pag.Offset
	votoType := c.Query("voto")
	ano, _ := strconv.Atoi(c.Query("ano"))

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
		"votacoes":    utils.NaoNulo(votacoes),
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
		"por_tipo":   utils.NaoNulo(tipos),
	})
}

// GetAll godoc
// @Summary Lista todas as votacoes (uma linha por votacao)
// @Tags votacoes
// @Produce json
// @Param page query int false "Pagina (default 1)"
// @Param limit query int false "Limite (default 20)"
// @Param ano query int false "Ano (default: todos)"
// @Param materia query string false "Filtro por materia/descricao"
// @Param sessao query string false "Codigo da sessao (todas as votacoes dela)"
// @Param tipo query string false "Siglas da materia separadas por virgula (PEC,MSF...)"
// @Param secreta query bool false "true: so secretas; false: so abertas"
// @Param resultado query string false "Resultados separados por virgula (A,R)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/votacoes [get]
func (h *Handler) GetAll(c *gin.Context) {
	// Paginação com teto de limit e de offset (ver utils.LerPaginacao)
	pag := utils.LerPaginacao(c.Query("limit"), c.Query("page"), 20, utils.LimiteMaximoPadrao)
	limit, page, offset := pag.Limit, pag.Page, pag.Offset
	ano, _ := strconv.Atoi(c.Query("ano"))

	secreta, err := parseSecreta(c.Query("secreta"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	filtro := FiltroLista{
		Ano:        ano,
		Materia:    c.Query("materia"),
		Sessao:     c.Query("sessao"),
		Ordem:      c.DefaultQuery("ordem", "desc"),
		Tipos:      parseLista(c.Query("tipo")),
		Secreta:    secreta,
		Resultados: parseLista(c.Query("resultado")),
	}

	votacoes, total, err := h.repo.FindAll(limit, offset, filtro)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao buscar votacoes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  utils.NaoNulo(votacoes),
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetFacetas godoc
// @Summary Contagens de votacoes por tipo de materia, secreta e resultado
// @Tags votacoes
// @Produce json
// @Param ano query int false "Ano (default: todos)"
// @Success 200 {object} Facetas
// @Router /api/v1/votacoes/facetas [get]
func (h *Handler) GetFacetas(c *gin.Context) {
	ano, _ := strconv.Atoi(c.Query("ano"))
	facetas, err := h.repo.Facetas(ano)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao buscar facetas"})
		return
	}
	c.JSON(http.StatusOK, facetas)
}

// siglaValida limita os valores de tipo/resultado a siglas simples
var siglaValida = regexp.MustCompile(`^[A-Z0-9-]{1,20}$`)

// maxItensLista limita quantas siglas ?tipo= ou ?resultado= viram filtro. As
// facetas reais são poucas dezenas; sem teto, uma query string longa virava
// um IN (...) com milhares de itens em cada consulta.
const maxItensLista = 30

// parseLista le uma lista separada por virgula, em maiusculas, sem vazios,
// repetidos ou valores fora do formato de sigla.
func parseLista(valor string) []string {
	var out []string
	vistos := map[string]bool{}
	for _, item := range strings.Split(valor, ",") {
		item = strings.ToUpper(strings.TrimSpace(item))
		if item == "" || vistos[item] || !siglaValida.MatchString(item) || len(out) >= maxItensLista {
			continue
		}
		vistos[item] = true
		out = append(out, item)
	}
	return out
}

// parseSecreta le o filtro secreta: vazio e "todas" (nil).
func parseSecreta(valor string) (*bool, error) {
	if strings.TrimSpace(valor) == "" {
		return nil, nil
	}
	b, err := strconv.ParseBool(strings.TrimSpace(valor))
	if err != nil {
		return nil, fmt.Errorf("parametro secreta invalido %q: use true ou false", valor)
	}
	return &b, nil
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
		"votos":   utils.NaoNulo(votos),
	})
}
