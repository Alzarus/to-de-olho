package emenda

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetBySenador(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id invalido"})
		return
	}

	anoStr := c.Query("ano")
	ano := 0
	if anoStr != "" {
		ano, _ = strconv.Atoi(anoStr)
	}

	emendas, err := h.service.ListBySenador(uint(id), ano)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resumo, err := h.service.GetResumo(uint(id), ano)
	if err != nil {
		_ = c.Error(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"emendas": emendas,
		"resumo":  resumo,
	})
}
