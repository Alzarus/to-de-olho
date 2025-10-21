package http

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"to-de-olho-backend/internal/application"
	"to-de-olho-backend/internal/domain"

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

// GetProposicoesHandler retorna handler para listar proposições
func GetProposicoesHandler(svc application.ProposicoesServiceInterface) gin.HandlerFunc {
	logger := slog.Default()

	return func(c *gin.Context) {
		start := time.Now()

		// Extrair filtros dos parâmetros de query
		filtros := extrairFiltrosProposicoes(c)

		// Validar filtros
		if err := filtros.Validate(); err != nil {
			logger.Error("filtros inválidos para busca de proposições",
				slog.String("error", err.Error()),
				slog.Any("filtros", filtros))

			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Parâmetros inválidos",
				"details": err.Error(),
			})
			return
		}

		// Buscar proposições
		proposicoes, total, source, err := svc.ListarProposicoes(c.Request.Context(), filtros)
		if err != nil {
			logger.Error("erro ao buscar proposições",
				slog.String("error", err.Error()),
				slog.Any("filtros", filtros),
				slog.Duration("duration", time.Since(start)))

			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Erro interno do servidor",
				"details": err.Error(),
			})
			return
		}

		// Resposta de sucesso
		response := gin.H{
			"data":    proposicoes,
			"total":   total,
			"pagina":  filtros.Pagina,
			"limite":  filtros.Limite,
			"source":  source,
			"filtros": buildFiltrosResponse(filtros),
		}

		c.JSON(http.StatusOK, response)

		logger.Info("proposições listadas com sucesso",
			slog.Int("total", total),
			slog.String("source", source),
			slog.Duration("duration", time.Since(start)))
	}
}

// GetProposicaoPorIDHandler retorna handler para buscar proposição por ID
func GetProposicaoPorIDHandler(svc application.ProposicoesServiceInterface) gin.HandlerFunc {
	logger := slog.Default()

	return func(c *gin.Context) {
		start := time.Now()

		// Extrair e validar ID
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			logger.Warn("ID de proposição inválido",
				slog.String("id_string", idStr),
				slog.String("error", err.Error()))

			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "ID de proposição inválido",
				"details": "ID deve ser um número inteiro positivo",
			})
			return
		}

		// Buscar proposição
		proposicao, source, err := svc.BuscarProposicaoPorID(c.Request.Context(), id)
		if err != nil {
			if err == domain.ErrProposicaoNaoEncontrada {
				logger.Info("proposição não encontrada",
					slog.Int("id", id),
					slog.Duration("duration", time.Since(start)))

				c.JSON(http.StatusNotFound, gin.H{
					"error":   "Proposição não encontrada",
					"details": "Não foi possível encontrar proposição com o ID informado",
				})
				return
			}

			logger.Error("erro ao buscar proposição por ID",
				slog.Int("id", id),
				slog.String("error", err.Error()),
				slog.Duration("duration", time.Since(start)))

			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Erro interno do servidor",
				"details": err.Error(),
			})
			return
		}

		// Resposta de sucesso
		response := gin.H{
			"data":   proposicao,
			"source": source,
		}

		c.JSON(http.StatusOK, response)

		logger.Info("proposição encontrada com sucesso",
			slog.Int("id", id),
			slog.String("identificacao", proposicao.GetIdentificacao()),
			slog.String("source", source),
			slog.Duration("duration", time.Since(start)))
	}
}

// extrairFiltrosProposicoes extrai filtros dos parâmetros de query
func extrairFiltrosProposicoes(c *gin.Context) *domain.ProposicaoFilter {
	filtros := &domain.ProposicaoFilter{}

	// Parâmetros básicos
	if limite := c.Query("limite"); limite != "" {
		if l, err := strconv.Atoi(limite); err == nil && l > 0 {
			filtros.Limite = l
		}
	}

	if pagina := c.Query("pagina"); pagina != "" {
		if p, err := strconv.Atoi(pagina); err == nil && p > 0 {
			filtros.Pagina = p
		}
	}

	if ordem := c.Query("ordem"); ordem != "" {
		filtros.Ordem = ordem
	}

	if ordenarPor := c.Query("ordenarPor"); ordenarPor != "" {
		filtros.OrdenarPor = ordenarPor
	}

	// Filtros específicos
	if siglaTipo := c.Query("siglaTipo"); siglaTipo != "" {
		filtros.SiglaTipo = siglaTipo
	}

	if numero := c.Query("numero"); numero != "" {
		if n, err := strconv.Atoi(numero); err == nil && n > 0 {
			filtros.Numero = &n
		}
	}

	if ano := c.Query("ano"); ano != "" {
		if a, err := strconv.Atoi(ano); err == nil {
			filtros.Ano = &a
		}
	}

	if codSituacao := c.Query("codSituacao"); codSituacao != "" {
		if cs, err := strconv.Atoi(codSituacao); err == nil {
			filtros.CodSituacao = &cs
		}
	}

	if siglaUfAutor := c.Query("siglaUfAutor"); siglaUfAutor != "" {
		filtros.SiglaUfAutor = siglaUfAutor
	}

	if siglaPartidoAutor := c.Query("siglaPartidoAutor"); siglaPartidoAutor != "" {
		filtros.SiglaPartidoAutor = siglaPartidoAutor
	}

	if nomeAutor := c.Query("nomeAutor"); nomeAutor != "" {
		filtros.NomeAutor = nomeAutor
	}

	if tema := c.Query("tema"); tema != "" {
		filtros.Tema = tema
	}

	if keywords := c.Query("keywords"); keywords != "" {
		filtros.Keywords = keywords
	}

	// Aplicar padrões
	filtros.SetDefaults()

	return filtros
}

// buildFiltrosResponse constrói resposta com filtros aplicados
func buildFiltrosResponse(filtros *domain.ProposicaoFilter) gin.H {
	response := gin.H{
		"limite":     filtros.Limite,
		"pagina":     filtros.Pagina,
		"ordem":      filtros.Ordem,
		"ordenarPor": filtros.OrdenarPor,
	}

	if filtros.SiglaTipo != "" {
		response["siglaTipo"] = filtros.SiglaTipo
	}

	if filtros.Numero != nil {
		response["numero"] = *filtros.Numero
	}

	if filtros.Ano != nil {
		response["ano"] = *filtros.Ano
	}

	if filtros.CodSituacao != nil {
		response["codSituacao"] = *filtros.CodSituacao
	}

	if filtros.SiglaUfAutor != "" {
		response["siglaUfAutor"] = filtros.SiglaUfAutor
	}

	if filtros.SiglaPartidoAutor != "" {
		response["siglaPartidoAutor"] = filtros.SiglaPartidoAutor
	}

	if filtros.NomeAutor != "" {
		response["nomeAutor"] = filtros.NomeAutor
	}

	if filtros.Tema != "" {
		response["tema"] = filtros.Tema
	}

	if filtros.Keywords != "" {
		response["keywords"] = filtros.Keywords
	}

	return response
}

// Analytics Handlers

// GetRankingGastosHandler retorna ranking de gastos dos deputados
func GetRankingGastosHandler(svc application.AnalyticsServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parâmetros opcionais
		anoStr := c.DefaultQuery("ano", strconv.Itoa(time.Now().Year()))
		limiteStr := c.DefaultQuery("limite", "50")

		ano, err := strconv.Atoi(anoStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetro 'ano' inválido"})
			return
		}

		limite, err := strconv.Atoi(limiteStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetro 'limite' inválido"})
			return
		}

		ranking, source, err := svc.GetRankingGastos(c.Request.Context(), ano, limite)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar ranking de gastos", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":   ranking,
			"source": source,
		})
	}
}

// GetRankingProposicoesHandler retorna ranking de proposições dos deputados
func GetRankingProposicoesHandler(svc application.AnalyticsServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parâmetros opcionais
		anoStr := c.DefaultQuery("ano", strconv.Itoa(time.Now().Year()))
		limiteStr := c.DefaultQuery("limite", "50")

		ano, err := strconv.Atoi(anoStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetro 'ano' inválido"})
			return
		}

		limite, err := strconv.Atoi(limiteStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetro 'limite' inválido"})
			return
		}

		ranking, source, err := svc.GetRankingProposicoes(c.Request.Context(), ano, limite)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar ranking de proposições", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":   ranking,
			"source": source,
		})
	}
}

// GetRankingPresencaHandler retorna ranking de presença dos deputados
func GetRankingPresencaHandler(svc application.AnalyticsServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parâmetros opcionais
		anoStr := c.DefaultQuery("ano", strconv.Itoa(time.Now().Year()))
		limiteStr := c.DefaultQuery("limite", "50")

		ano, err := strconv.Atoi(anoStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetro 'ano' inválido"})
			return
		}

		limite, err := strconv.Atoi(limiteStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetro 'limite' inválido"})
			return
		}

		ranking, source, err := svc.GetRankingPresenca(c.Request.Context(), ano, limite)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar ranking de presença", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":   ranking,
			"source": source,
		})
	}
}

// GetInsightsGeraisHandler retorna insights gerais sobre os dados
func GetInsightsGeraisHandler(svc application.AnalyticsServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		insights, source, err := svc.GetInsightsGerais(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar insights gerais", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":   insights,
			"source": source,
		})
	}
}

// PostAtualizarRankingsHandler força atualização de todos os rankings
func PostAtualizarRankingsHandler(svc application.AnalyticsServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := svc.AtualizarRankings(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar rankings", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Rankings atualizados com sucesso",
		})
	}
}

// GetRankingDeputadosVotacaoHandler retorna ranking de deputados por participação/votos
func GetRankingDeputadosVotacaoHandler(svc application.AnalyticsServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		anoStr := c.DefaultQuery("ano", strconv.Itoa(time.Now().Year()))
		limiteStr := c.DefaultQuery("limite", "50")

		ano, err := strconv.Atoi(anoStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetro 'ano' inválido"})
			return
		}

		limite, err := strconv.Atoi(limiteStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetro 'limite' inválido"})
			return
		}

		ranking, source, err := svc.GetRankingDeputadosVotacao(c.Request.Context(), ano, limite)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar ranking de votações", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": ranking, "source": source})
	}
}

// GetRankingPartidosDisciplinaHandler retorna ranking de disciplina partidária
func GetRankingPartidosDisciplinaHandler(svc application.AnalyticsServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		anoStr := c.DefaultQuery("ano", strconv.Itoa(time.Now().Year()))

		ano, err := strconv.Atoi(anoStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetro 'ano' inválido"})
			return
		}

		ranking, source, err := svc.GetRankingPartidosDisciplina(c.Request.Context(), ano)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar ranking de disciplina", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": ranking, "source": source})
	}
}

// GetStatsVotacoesHandler retorna estatísticas agregadas de votações
func GetStatsVotacoesHandler(svc application.AnalyticsServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		periodo := c.DefaultQuery("periodo", "ano")

		stats, source, err := svc.GetStatsVotacoes(c.Request.Context(), periodo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar estatísticas de votações", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": stats, "source": source})
	}
}
