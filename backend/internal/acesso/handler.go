package acesso

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler recebe o aviso de visita enviado pelo navegador.
type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// Registrar godoc
// @Summary Conta uma visita (visitante unico por dia, sem cookie nem IP guardado)
// @Tags acessos
// @Success 204
// @Router /api/v1/acessos [post]
func (h *Handler) Registrar(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	ua := c.GetHeader("User-Agent")
	if EhRobo(ua) {
		c.Status(http.StatusNoContent)
		return
	}
	if err := h.repo.Registrar(c.ClientIP(), ua); err != nil {
		// A contagem nunca atrapalha a navegacao: loga e segue.
		slog.Warn("falha ao registrar acesso", "error", err)
	}
	c.Status(http.StatusNoContent)
}
