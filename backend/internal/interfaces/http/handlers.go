package http

import (
	"net/http"
	"strconv"
	"time"

	"to-de-olho-backend/internal/application"

	"github.com/gin-gonic/gin"
)

func GetDeputadosHandler(svc application.DeputadosServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		partido := c.Query("partido")
		uf := c.Query("uf")
		nome := c.Query("nome")
		deps, source, err := svc.ListarDeputados(c.Request.Context(), partido, uf, nome)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar deputados", "details": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": deps, "total": len(deps), "source": source})
	}
}

func GetDeputadoByIDHandler(svc application.DeputadosServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		dep, source, err := svc.BuscarDeputadoPorID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Deputado não encontrado", "details": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": dep, "source": source})
	}
}

func GetDespesasDeputadoHandler(svc application.DeputadosServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		ano := c.DefaultQuery("ano", strconv.Itoa(time.Now().Year()))
		desp, source, err := svc.ListarDespesas(c.Request.Context(), id, ano)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar despesas", "details": err.Error()})
			return
		}
		var total float64
		for _, d := range desp {
			total += d.ValorLiquido
		}
		c.JSON(http.StatusOK, gin.H{"data": desp, "total": len(desp), "valor_total": total, "ano": ano, "source": source})
	}
}
