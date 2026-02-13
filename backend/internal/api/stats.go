package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// StatsResponse representa as estatisticas gerais da plataforma
type StatsResponse struct {
	TotalSenadores    int64     `json:"total_senadores"`
	TotalVotos        int64     `json:"total_votos"`
	TotalDespesasCEAP float64   `json:"total_despesas_ceaps"`
	TotalEmendas      int64     `json:"total_emendas"`
	UltimaAtualizacao time.Time `json:"ultima_atualizacao"`
}

func statsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var stats StatsResponse

		// Total de senadores em exercicio
		db.Raw("SELECT COUNT(*) FROM senadores WHERE em_exercicio = true").Scan(&stats.TotalSenadores)

		// Total de votos registrados
		db.Raw("SELECT COUNT(*) FROM votacoes").Scan(&stats.TotalVotos)

		// Total de despesas CEAPS (soma de valores)
		db.Raw("SELECT COALESCE(SUM(valor), 0) FROM despesa_ceaps").Scan(&stats.TotalDespesasCEAP)

		// Total de emendas
		db.Raw("SELECT COUNT(*) FROM emendas").Scan(&stats.TotalEmendas)

		// Ultima atualizacao (maior timestamp entre tabelas)
		db.Raw(`
			SELECT COALESCE(MAX(ts), NOW()) FROM (
				SELECT MAX(updated_at) as ts FROM senadores
				UNION ALL
				SELECT MAX(created_at) as ts FROM votacoes
				UNION ALL
				SELECT MAX(data_ultima_atualizacao) as ts FROM emendas
			) as updates
		`).Scan(&stats.UltimaAtualizacao)

		c.JSON(http.StatusOK, stats)
	}
}
