package ceaps

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/Alzarus/to-de-olho/internal/utils"
)

// limiteMaximo e o maior limit aceito na lista de despesas
const limiteMaximo = 100

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
// @Param limit query int false "Limite (default 20)"
// @Param page query int false "Pagina (default 1)"
// @Param q query string false "Termo de busca"
// @Param tipo query string false "Tipo de despesa"
// @Param sort query string false "Ordenacao (data_desc, data_asc, valor_desc, valor_asc)"
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
		if anoVal, err := strconv.Atoi(anoStr); err == nil && anoVal > 0 {
			ano = &anoVal
		}
	}
	
	// Paginacao. Agregados tem rota propria (/mensal, /fornecedores): a
	// lista nao serve para somar e nao precisa devolver milhares de linhas.
	pag := utils.LerPaginacao(c.Query("limit"), c.Query("page"), 20, limiteMaximo)
	limit, page, offset := pag.Limit, pag.Page, pag.Offset

	queryStr := c.Query("q")
	tipo := c.Query("tipo")
	sort := c.Query("sort")

	despesas, total, err := h.repo.FindBySenadorID(senadorID, ano, limit, offset, queryStr, tipo, sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao buscar despesas"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"senador_id": senadorID,
		"total":      total,
		"limit":      limit,
		"page":       page,
		"total_pages": (int(total) + limit - 1) / limit,
		"despesas":   utils.NaoNulo(despesas),
	})
}

// AggregateBySenador godoc
// @Summary Retorna gastos agregados por tipo de despesa
// @Tags despesas
// @Produce json
// @Param senador_id path int true "ID do senador"
// @Param ano query int false "Ano de referencia"
// @Param mes_de query int false "Primeiro mes de competencia (1-12)"
// @Param mes_ate query int false "Ultimo mes de competencia (1-12)"
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

	agregados, err := h.repo.AggregateByTipo(senadorID, ano, IntervaloMeses{De: mesOpcional(c, "mes_de"), Ate: mesOpcional(c, "mes_ate")})
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
		"por_tipo":    utils.NaoNulo(agregados),
	})
}

// anoOpcional le ?ano=; ausente ou invalido vale "todos os anos"
func anoOpcional(c *gin.Context) *int {
	if anoVal, err := strconv.Atoi(c.Query("ano")); err == nil && anoVal > 0 {
		return &anoVal
	}
	return nil
}

// mesOpcional le um mes 1-12; ausente ou invalido vale 0 (sem limite)
func mesOpcional(c *gin.Context, chave string) int {
	if m, err := strconv.Atoi(c.Query(chave)); err == nil && m >= 1 && m <= 12 {
		return m
	}
	return 0
}

// maxTipos limita quantas categorias ?tipo= um agregado aceita
const maxTipos = 30

// tiposOpcionais le ?tipo= repetido (uma categoria por parametro: os nomes
// das categorias tem virgula). Ausente: todas as categorias.
func tiposOpcionais(c *gin.Context) []string {
	var tipos []string
	vistos := map[string]bool{}
	for _, t := range c.QueryArray("tipo") {
		t = strings.TrimSpace(t)
		if t == "" || vistos[t] || len(tipos) >= maxTipos {
			continue
		}
		vistos[t] = true
		tipos = append(tipos, t)
	}
	return tipos
}

// MensalBySenador godoc
// @Summary Retorna o gasto por mes de competencia
// @Tags despesas
// @Produce json
// @Param senador_id path int true "ID do senador"
// @Param ano query int false "Ano de referencia"
// @Param tipo query []string false "Categorias (tipo_despesa); repetir o parametro para varias" collectionFormat(multi)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/senadores/{senador_id}/despesas/mensal [get]
func (h *Handler) MensalBySenador(c *gin.Context) {
	senadorID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalido"})
		return
	}

	meses, err := h.repo.GastoMensal(senadorID, anoOpcional(c), tiposOpcionais(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao agregar despesas por mes"})
		return
	}
	if meses == nil {
		meses = []SenadorGastoMensal{}
	}

	c.JSON(http.StatusOK, gin.H{
		"senador_id": senadorID,
		"meses":      utils.NaoNulo(meses),
	})
}

// FornecedoresBySenador godoc
// @Summary Retorna o total pago a cada fornecedor, do maior para o menor
// @Tags despesas
// @Produce json
// @Param senador_id path int true "ID do senador"
// @Param ano query int false "Ano de referencia"
// @Param tipo query []string false "Categorias (tipo_despesa); repetir o parametro para varias" collectionFormat(multi)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/senadores/{senador_id}/despesas/fornecedores [get]
func (h *Handler) FornecedoresBySenador(c *gin.Context) {
	senadorID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalido"})
		return
	}

	fornecedores, err := h.repo.Fornecedores(senadorID, anoOpcional(c), tiposOpcionais(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao agregar despesas por fornecedor"})
		return
	}
	if fornecedores == nil {
		fornecedores = []FornecedorAgregado{}
	}

	c.JSON(http.StatusOK, gin.H{
		"senador_id":   senadorID,
		"fornecedores": utils.NaoNulo(fornecedores),
	})
}
