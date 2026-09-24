package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Alzarus/to-de-olho/internal/acesso"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// StatsResponse representa as estatisticas gerais da plataforma
type StatsResponse struct {
	TotalSenadores    int64      `json:"total_senadores"`
	TotalVotos        int64      `json:"total_votos"`    // votos individuais (um por senador por votacao)
	TotalVotacoes     int64      `json:"total_votacoes"` // votacoes nominais distintas
	VotacoesDesde     *time.Time `json:"votacoes_desde"` // data da primeira votacao no banco
	TotalDespesasCEAP float64    `json:"total_despesas_ceaps"`
	CeapsAnoInicio    int        `json:"ceaps_ano_inicio"` // periodo somado em total_despesas_ceaps
	CeapsAnoFim       int        `json:"ceaps_ano_fim"`
	TotalEmendas      int64      `json:"total_emendas"`
	TotalAcessos      int64      `json:"total_acessos"` // visitantes unicos por dia, somados
	AcessosDesde      *time.Time `json:"acessos_desde"` // inicio da contagem; null antes da primeira visita
	UltimaAtualizacao time.Time  `json:"ultima_atualizacao"`
}

func statsHandler(db *gorm.DB, acessos *acesso.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var stats StatsResponse

		// Total de senadores em exercicio
		db.Raw("SELECT COUNT(*) FROM senadores WHERE em_exercicio = true").Scan(&stats.TotalSenadores)

		// Votos individuais e votacoes distintas
		var votos struct {
			Votos    int64
			Votacoes int64
			Desde    *time.Time
		}
		db.Raw("SELECT COUNT(*) AS votos, COUNT(DISTINCT codigo_votacao) AS votacoes, MIN(data) AS desde FROM votacoes").Scan(&votos)
		stats.TotalVotos, stats.TotalVotacoes, stats.VotacoesDesde = votos.Votos, votos.Votacoes, votos.Desde

		// Total de despesas CEAPS e o periodo coberto
		var ceaps struct {
			Total  float64
			Inicio int
			Fim    int
		}
		db.Raw("SELECT COALESCE(SUM(valor), 0) AS total, COALESCE(MIN(ano), 0) AS inicio, COALESCE(MAX(ano), 0) AS fim FROM despesas_ceaps").Scan(&ceaps)
		stats.TotalDespesasCEAP, stats.CeapsAnoInicio, stats.CeapsAnoFim = ceaps.Total, ceaps.Inicio, ceaps.Fim

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

		// Acessos reais (internal/acesso); antes era um numero fixo
		if resumo, err := acessos.Resumo(); err != nil {
			slog.Warn("falha ao ler acessos", "error", err)
		} else {
			stats.TotalAcessos, stats.AcessosDesde = resumo.Total, resumo.Desde
		}

		c.JSON(http.StatusOK, stats)
	}
}
