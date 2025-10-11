package http

import (
	"net/http"
	"strconv"

	"to-de-olho-backend/internal/application"
	"to-de-olho-backend/internal/domain"

	"github.com/gin-gonic/gin"
)

// VotacaoHandler gerencia endpoints relacionados a votações
type VotacaoHandler struct {
	votacoesService *application.VotacoesService
}

// NewVotacaoHandler cria um novo handler de votações
func NewVotacaoHandler(votacoesService *application.VotacoesService) *VotacaoHandler {
	return &VotacaoHandler{
		votacoesService: votacoesService,
	}
}

// ListVotacoes lista votações com filtros
// @Summary Lista votações
// @Description Lista votações da Câmara dos Deputados com filtros opcionais
// @Tags Votações
// @Accept json
// @Produce json
// @Param busca query string false "Busca textual no título/ementa"
// @Param ano query int false "Ano da votação"
// @Param aprovacao query string false "Status de aprovação (Aprovada/Rejeitada)"
// @Param relevancia query string false "Relevância (alta/média/baixa)"
// @Param tipo_proposicao query string false "Tipo da proposição"
// @Param page query int false "Página (padrão: 1)"
// @Param limit query int false "Limite por página (padrão: 20, máx: 100)"
// @Success 200 {object} domain.PaginationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/votacoes [get]
func (h *VotacaoHandler) ListVotacoes(c *gin.Context) {
	// Parse filtros
	filtros := domain.FiltrosVotacao{
		Busca:          c.Query("busca"),
		Aprovacao:      c.Query("aprovacao"),
		Relevancia:     c.Query("relevancia"),
		TipoProposicao: c.Query("tipo_proposicao"),
	}

	// Parse ano se fornecido
	if anoStr := c.Query("ano"); anoStr != "" {
		if ano, err := strconv.Atoi(anoStr); err == nil {
			filtros.Ano = ano
		}
	}

	// Parse paginação
	pag := domain.Pagination{
		Page:  1,
		Limit: 20,
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			pag.Page = page
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			pag.Limit = limit
		}
	}

	// Buscar votações
	votacoes, total, err := h.votacoesService.ListarVotacoes(c.Request.Context(), filtros, pag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Erro ao listar votações",
			Details: err.Error(),
		})
		return
	}

	// Construir resposta paginada
	response := domain.BuildPagination(&domain.PaginationRequest{
		Page:  pag.Page,
		Limit: pag.Limit,
	}, int64(total), votacoes)

	c.JSON(http.StatusOK, response)
}

// GetVotacao obtém uma votação específica
// @Summary Obtém votação por ID
// @Description Obtém dados básicos de uma votação específica
// @Tags Votações
// @Accept json
// @Produce json
// @Param id path int true "ID da votação"
// @Success 200 {object} domain.Votacao
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/votacoes/{id} [get]
func (h *VotacaoHandler) GetVotacao(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Details: "O ID deve ser um número inteiro",
		})
		return
	}

	votacao, err := h.votacoesService.ObterVotacaoDetalhada(c.Request.Context(), id)
	if err != nil {
		if err == domain.ErrVotacaoNaoEncontrada {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Votação não encontrada",
				Details: "Não foi possível encontrar uma votação com este ID",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Erro ao buscar votação",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, votacao.Votacao)
}

// GetVotacaoCompleta obtém votação com votos e orientações
// @Summary Obtém votação completa
// @Description Obtém votação com todos os votos dos deputados e orientações partidárias
// @Tags Votações
// @Accept json
// @Produce json
// @Param id path int true "ID da votação"
// @Success 200 {object} domain.VotacaoDetalhada
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/votacoes/{id}/completa [get]
func (h *VotacaoHandler) GetVotacaoCompleta(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Details: "O ID deve ser um número inteiro",
		})
		return
	}

	votacao, err := h.votacoesService.ObterVotacaoDetalhada(c.Request.Context(), id)
	if err != nil {
		if err == domain.ErrVotacaoNaoEncontrada {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Votação não encontrada",
				Details: "Não foi possível encontrar uma votação com este ID",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Erro ao buscar votação completa",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, votacao)
}

// ErrorResponse representa uma resposta de erro da API
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}
