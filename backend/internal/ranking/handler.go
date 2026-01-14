package ranking

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler gerencia endpoints de ranking
type Handler struct {
	service *Service
}

// NewHandler cria um novo handler de ranking
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetRanking retorna o ranking geral de senadores
// GET /api/v1/ranking
func (h *Handler) GetRanking(c *gin.Context) {
	ranking, err := h.service.CalcularRanking(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Aplicar limite se especificado
	if limitStr := c.Query("limite"); limitStr != "" {
		if limite, err := strconv.Atoi(limitStr); err == nil && limite > 0 && limite < len(ranking.Ranking) {
			ranking.Ranking = ranking.Ranking[:limite]
		}
	}

	c.JSON(http.StatusOK, ranking)
}

// GetScoreSenador retorna o score detalhado de um senador
// GET /api/v1/senadores/:id/score
func (h *Handler) GetScoreSenador(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id invalido"})
		return
	}

	score, err := h.service.CalcularScoreSenador(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "senador nao encontrado"})
		return
	}

	c.JSON(http.StatusOK, score)
}

// GetMetodologia retorna a metodologia de calculo do ranking
// GET /api/v1/ranking/metodologia
func (h *Handler) GetMetodologia(c *gin.Context) {
	metodologia := gin.H{
		"titulo":      "Metodologia do Ranking de Senadores",
		"versao":      "2.0",
		"referencia":  "Volden, C. & Wiseman, A. E. (2018). Legislative Effectiveness in the American States",
		"formula":     "Score = (Produtividade * 0.35) + (Presenca * 0.25) + (Economia * 0.20) + (Comissoes * 0.20)",
		"criterios": []gin.H{
			{
				"nome":        "Produtividade Legislativa",
				"peso":        "35%",
				"descricao":   "Capacidade de avancar proposicoes pelo processo legislativo",
				"normalizacao": "Pontuacao do senador / Maior pontuacao da casa * 100",
			},
			{
				"nome":        "Presenca em Votacoes",
				"peso":        "25%",
				"descricao":   "Participacao em votacoes nominais",
				"normalizacao": "(Total - Ausencias) / Total * 100",
			},
			{
				"nome":        "Economia na Cota (CEAPS)",
				"peso":        "20%",
				"descricao":   "Responsabilidade fiscal no uso da cota parlamentar",
				"normalizacao": "(1 - Gasto / Teto) * 100",
			},
			{
				"nome":        "Participacao em Comissoes",
				"peso":        "20%",
				"descricao":   "Trabalho tecnico em comissoes permanentes e temporarias",
				"normalizacao": "Pontos do senador / Maior pontuacao da casa * 100",
			},
		},
		"escala": "Todos os scores sao normalizados para escala 0-100 antes da ponderacao",
	}

	c.JSON(http.StatusOK, metodologia)
}
