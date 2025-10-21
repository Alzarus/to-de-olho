package ingestor

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	app "to-de-olho-backend/internal/application"
	"to-de-olho-backend/internal/domain"
	"to-de-olho-backend/internal/infrastructure/repository"
)

// DespesaUpserter abstrai operações de persistência utilizadas pelo backfill histórico de despesas.
type DespesaUpserter interface {
	UpsertDespesas(ctx context.Context, deputadoID int, ano int, despesas []domain.Despesa) error
}

// StrategicBackfillExecutor executa backfill histórico com estratégia inteligente
type StrategicBackfillExecutor struct {
	manager            *BackfillManager
	deputadosService   *app.DeputadosService
	proposicoesService *app.ProposicoesService
	deputadoRepo       *repository.DeputadoRepository
	proposicaoRepo     *repository.ProposicaoRepository
	votacoesService    *app.VotacoesService
	despesaRepo        DespesaUpserter
	partidosService    *app.PartidosService
	analyticsService   *app.AnalyticsService
	strategy           BackfillStrategy
}

// NewStrategicBackfillExecutor cria um novo executor estratégico
func NewStrategicBackfillExecutor(
	manager *BackfillManager,
	deputadosService *app.DeputadosService,
	proposicoesService *app.ProposicoesService,
	deputadoRepo *repository.DeputadoRepository,
	proposicaoRepo *repository.ProposicaoRepository,
	votacoesService *app.VotacoesService,
	despesaRepo DespesaUpserter,
	partidosService *app.PartidosService,
	analyticsService *app.AnalyticsService,
	strategy BackfillStrategy,
) *StrategicBackfillExecutor {
	return &StrategicBackfillExecutor{
		manager:            manager,
		deputadosService:   deputadosService,
		proposicoesService: proposicoesService,
		deputadoRepo:       deputadoRepo,
		proposicaoRepo:     proposicaoRepo,
		votacoesService:    votacoesService,
		despesaRepo:        despesaRepo,
		partidosService:    partidosService,
		analyticsService:   analyticsService,
		strategy:           strategy,
	}
}

// ExecuteBackfill executa o backfill completo seguindo a estratégia
func (sbe *StrategicBackfillExecutor) ExecuteBackfill(ctx context.Context) error {
	log.Println("🚀 Iniciando Backfill Histórico Estratégico")

	// 1. Verificar se há checkpoints pendentes para resumir
	pendingCheckpoints, err := sbe.manager.GetPendingCheckpoints(ctx)
	if err != nil {
		return fmt.Errorf("erro ao verificar checkpoints pendentes: %w", err)
	}

	if len(pendingCheckpoints) > 0 {
		log.Printf("🔄 Resumindo %d checkpoints pendentes", len(pendingCheckpoints))
		for _, checkpoint := range pendingCheckpoints {
			if err := sbe.resumeCheckpoint(ctx, checkpoint); err != nil {
				log.Printf("❌ Erro ao resumir checkpoint %s: %v", checkpoint.ID, err)
				if markErr := sbe.manager.MarkAsFailed(ctx, checkpoint, err.Error()); markErr != nil {
					log.Printf("❌ Erro ao marcar checkpoint como falhado: %v", markErr)
				}
			}
		}
	}

	// 2. Criar novos checkpoints para backfill completo
	if err := sbe.createBackfillPlan(ctx); err != nil {
		return fmt.Errorf("erro ao criar plano de backfill: %w", err)
	}

	// 3. Executar checkpoints em ordem de prioridade
	return sbe.executeBackfillPlan(ctx)
}

// createBackfillPlan cria checkpoints para todo o backfill necessário
func (sbe *StrategicBackfillExecutor) createBackfillPlan(ctx context.Context) error {
	log.Println("📋 Criando plano de backfill histórico")

	// Prioridade 1: Deputados (base fundamental)
	deputadosMetadata := map[string]interface{}{
		"legislatura": "56", // Legislatura atual
		"description": "Backfill completo dos deputados da legislatura atual",
		"priority":    1,
	}

	if _, err := sbe.manager.CreateCheckpoint(ctx, "deputados", deputadosMetadata); err != nil {
		return fmt.Errorf("erro ao criar checkpoint deputados: %w", err)
	}

	// Prioridade 2: Proposições por ano (dados mais voláteis)
	for year := sbe.strategy.YearStart; year <= sbe.strategy.YearEnd; year++ {
		proposicoesMetadata := map[string]interface{}{
			"year":        year,
			"description": fmt.Sprintf("Backfill proposições do ano %d", year),
			"priority":    2,
			"batch_size":  sbe.strategy.BatchSize,
		}

		if _, err := sbe.manager.CreateCheckpoint(ctx, "proposicoes", proposicoesMetadata); err != nil {
			log.Printf("⚠️  Erro ao criar checkpoint para proposições %d: %v", year, err)
		}
	}

	// Prioridade 3: Despesas (dados grandes, menos críticos para MVP)
	for year := sbe.strategy.YearStart; year <= sbe.strategy.YearEnd; year++ {
		despesasMetadata := map[string]interface{}{
			"year":        year,
			"description": fmt.Sprintf("Backfill despesas do ano %d", year),
			"priority":    3,
			"batch_size":  sbe.strategy.BatchSize / 2, // Lotes menores para despesas
		}

		if _, err := sbe.manager.CreateCheckpoint(ctx, "despesas", despesasMetadata); err != nil {
			log.Printf("⚠️  Erro ao criar checkpoint para despesas %d: %v", year, err)
		}
	}

	// Prioridade 4: Votações por ano (usar upsert existente) - criar checkpoints uma vez por ano
	for year := sbe.strategy.YearStart; year <= sbe.strategy.YearEnd; year++ {
		votacoesMetadata := map[string]interface{}{
			"year":        year,
			"description": fmt.Sprintf("Backfill votações do ano %d", year),
			"priority":    4,
			"batch_size":  sbe.strategy.BatchSize,
		}

		if _, err := sbe.manager.CreateCheckpoint(ctx, "votacoes", votacoesMetadata); err != nil {
			log.Printf("⚠️  Erro ao criar checkpoint para votações %d: %v", year, err)
		}
	}

	// Partidos - baixa volumetria, executar uma vez por backfill
	partidosMetadata := map[string]interface{}{
		"description": "Backfill partidos - lista completa",
		"priority":    2,
		"batch_size":  sbe.strategy.BatchSize,
	}

	if _, err := sbe.manager.CreateCheckpoint(ctx, "partidos", partidosMetadata); err != nil {
		log.Printf("⚠️  Erro ao criar checkpoint para partidos: %v", err)
	}

	log.Printf("✅ Plano criado: 1 checkpoint deputados + %d proposições + %d despesas + %d votações",
		sbe.strategy.YearEnd-sbe.strategy.YearStart+1,
		sbe.strategy.YearEnd-sbe.strategy.YearStart+1,
		sbe.strategy.YearEnd-sbe.strategy.YearStart+1)

	return nil
}

// executeBackfillPlan executa os checkpoints em ordem de prioridade
func (sbe *StrategicBackfillExecutor) executeBackfillPlan(ctx context.Context) error {
	log.Println("⚡ Executando plano de backfill")

	// Buscar todos os checkpoints pendentes ordenados por prioridade
	pendingCheckpoints, err := sbe.manager.GetPendingCheckpoints(ctx)
	if err != nil {
		return fmt.Errorf("erro ao buscar checkpoints: %w", err)
	}

	totalCheckpoints := len(pendingCheckpoints)
	log.Printf("📊 Total de checkpoints para processar: %d", totalCheckpoints)

	for i, checkpoint := range pendingCheckpoints {
		log.Printf("🔄 Processando checkpoint %d/%d: %s", i+1, totalCheckpoints, checkpoint.ID)

		startTime := time.Now()
		if err := sbe.executeCheckpoint(ctx, checkpoint); err != nil {
			log.Printf("❌ Falha no checkpoint %s após %v: %v",
				checkpoint.ID, time.Since(startTime), err)
			if markErr := sbe.manager.MarkAsFailed(ctx, checkpoint, err.Error()); markErr != nil {
				log.Printf("❌ Erro ao marcar checkpoint como falhado: %v", markErr)
			}
			continue
		}

		duration := time.Since(startTime)
		log.Printf("✅ Checkpoint %s concluído em %v", checkpoint.ID, duration)
		if markErr := sbe.manager.MarkAsCompleted(ctx, checkpoint); markErr != nil {
			log.Printf("❌ Erro ao marcar checkpoint como concluído: %v", markErr)
		}
	}

	// Estatísticas finais
	stats, err := sbe.manager.GetBackfillStats(ctx)
	if err == nil {
		log.Printf("📈 Estatísticas finais do backfill: %+v", stats)
	}

	// 🔧 CORREÇÃO: Atualizar analytics após backfill histórico concluído
	log.Println("📊 Atualizando rankings e analytics após backfill histórico...")
	if sbe.analyticsService != nil {
		if err := sbe.analyticsService.AtualizarRankings(ctx); err != nil {
			log.Printf("⚠️ Erro ao atualizar analytics após backfill: %v", err)
		} else {
			log.Println("✅ Analytics atualizados com sucesso após backfill histórico")
		}
	}

	return nil
}

// executeCheckpoint executa um checkpoint específico
func (sbe *StrategicBackfillExecutor) executeCheckpoint(ctx context.Context, checkpoint *BackfillCheckpoint) error {
	// Marcar como iniciado
	if err := sbe.manager.MarkAsStarted(ctx, checkpoint); err != nil {
		return fmt.Errorf("erro ao marcar checkpoint como iniciado: %w", err)
	}

	switch checkpoint.Type {
	case "deputados":
		return sbe.executeDeputadosBackfill(ctx, checkpoint)
	case "proposicoes":
		return sbe.executeProposicoesBackfill(ctx, checkpoint)
	case "despesas":
		return sbe.executeDespesasBackfill(ctx, checkpoint)
	case "votacoes":
		return sbe.executeVotacoesBackfill(ctx, checkpoint)
	case "partidos":
		return sbe.executePartidosBackfill(ctx, checkpoint)
	default:
		return fmt.Errorf("tipo de checkpoint desconhecido: %s", checkpoint.Type)
	}
}

// executePartidosBackfill sincroniza a lista de partidos (low volume)
func (sbe *StrategicBackfillExecutor) executePartidosBackfill(ctx context.Context, checkpoint *BackfillCheckpoint) error {
	log.Printf("🏳️ Executando backfill de partidos")

	if sbe.partidosService == nil {
		log.Printf("⚠️  PartidosService não injetada no executor; pulando partidos")
		return nil
	}

	partidos, err := sbe.partidosService.ListarPartidos(ctx)
	if err != nil {
		checkpoint.Progress.FailedItems = len(partidos)
		if markErr := sbe.manager.UpdateProgress(ctx, checkpoint, checkpoint.Progress.ProcessedItems, checkpoint.Progress.FailedItems, checkpoint.Progress.LastProcessedID); markErr != nil {
			log.Printf("⚠️  Erro ao atualizar progresso de partidos: %v", markErr)
		}
		return fmt.Errorf("erro ao listar/sincronizar partidos: %w", err)
	}

	checkpoint.Progress.TotalItems = len(partidos)
	checkpoint.Progress.ProcessedItems = len(partidos)
	if err := sbe.manager.UpdateProgress(ctx, checkpoint, checkpoint.Progress.ProcessedItems, checkpoint.Progress.FailedItems, checkpoint.Progress.LastProcessedID); err != nil {
		log.Printf("⚠️  Erro ao atualizar progresso pós-sync partidos: %v", err)
	}

	log.Printf("🎉 Backfill partidos concluído: %d processados", len(partidos))
	return nil
}

// executeVotacoesBackfill executa backfill de votações por ano
func (sbe *StrategicBackfillExecutor) executeVotacoesBackfill(ctx context.Context, checkpoint *BackfillCheckpoint) error {
	year, ok := checkpoint.Metadata["year"].(float64)
	if !ok {
		return fmt.Errorf("metadado 'year' não encontrado ou inválido no checkpoint")
	}

	yearInt := int(year)
	log.Printf("🗳️ Executando backfill de votações para o ano %d", yearInt)

	// Paginação simples - página baseada em ID ou página numérica dependendo do client
	itensPorPagina := sbe.strategy.BatchSize
	if itensPorPagina <= 0 {
		itensPorPagina = 100
	}
	if itensPorPagina > 100 {
		itensPorPagina = 100
	}

	// Use VotacoesService.SincronizarVotacoes which accepts date ranges and performs upserts internally.
	// Construir período do ano
	dataInicio := time.Date(yearInt, time.January, 1, 0, 0, 0, 0, time.UTC)
	dataFim := time.Date(yearInt, time.December, 31, 23, 59, 59, 0, time.UTC)

	// A SincronizarVotacoes já faz Upsert e sincroniza votos/orientações
	if sbe.votacoesService == nil {
		log.Printf("⚠️  VotacoesService não injetada no executor; pulando votações %d", yearInt)
		return nil
	}

	processed, err := sbe.votacoesService.SincronizarVotacoes(ctx, dataInicio, dataFim)
	if err != nil {
		return fmt.Errorf("erro ao sincronizar votações ano %d: %w", yearInt, err)
	}

	// Atualizar progresso com os totais consolidados retornados pelo serviço
	checkpoint.Progress.TotalItems += processed
	checkpoint.Progress.ProcessedItems += processed
	if err := sbe.manager.UpdateProgress(ctx, checkpoint,
		checkpoint.Progress.ProcessedItems,
		checkpoint.Progress.FailedItems,
		checkpoint.Progress.LastProcessedID); err != nil {
		log.Printf("⚠️  Erro ao atualizar progresso pós-sync votações: %v", err)
	}

	log.Printf("🎉 Backfill votações %d executado via VotacoesService.SincronizarVotacoes (processadas: %d)", yearInt, processed)
	return nil
}

// executeDeputadosBackfill executa backfill de deputados
func (sbe *StrategicBackfillExecutor) executeDeputadosBackfill(ctx context.Context, checkpoint *BackfillCheckpoint) error {
	log.Printf("👥 Executando backfill de deputados")

	// Buscar todos os deputados da API
	deputados, _, err := sbe.deputadosService.ListarDeputados(ctx, "", "", "")
	if err != nil {
		return fmt.Errorf("erro ao buscar deputados da API: %w", err)
	}

	checkpoint.Progress.TotalItems = len(deputados)

	// Processar em lotes
	batchSize := sbe.strategy.BatchSize
	for i := 0; i < len(deputados); i += batchSize {
		end := i + batchSize
		if end > len(deputados) {
			end = len(deputados)
		}

		batch := deputados[i:end]

		// Retry com backoff exponencial
		var lastErr error
		for retry := 0; retry < sbe.strategy.MaxRetries; retry++ {
			if err := sbe.deputadoRepo.UpsertDeputados(ctx, batch); err != nil {
				lastErr = err
				log.Printf("⚠️  Tentativa %d/%d falhou para lote deputados %d-%d: %v",
					retry+1, sbe.strategy.MaxRetries, i, end-1, err)

				if retry < sbe.strategy.MaxRetries-1 {
					time.Sleep(sbe.strategy.RetryDelay * time.Duration(retry+1))
				}
				continue
			}

			// Sucesso - atualizar progresso
			checkpoint.Progress.ProcessedItems = end
			if len(batch) > 0 {
				checkpoint.Progress.LastProcessedID = fmt.Sprintf("%d", batch[len(batch)-1].ID)
			}

			if err := sbe.manager.UpdateProgress(ctx, checkpoint,
				checkpoint.Progress.ProcessedItems,
				checkpoint.Progress.FailedItems,
				checkpoint.Progress.LastProcessedID); err != nil {
				log.Printf("⚠️  Erro ao atualizar progresso: %v", err)
			}

			log.Printf("✅ Lote deputados %d-%d processado com sucesso", i, end-1)
			lastErr = nil
			break
		}

		if lastErr != nil {
			checkpoint.Progress.FailedItems += len(batch)
			return fmt.Errorf("falha após %d tentativas no lote deputados %d-%d: %w",
				sbe.strategy.MaxRetries, i, end-1, lastErr)
		}
	}

	log.Printf("🎉 Backfill deputados concluído: %d processados, %d falhas",
		checkpoint.Progress.ProcessedItems, checkpoint.Progress.FailedItems)

	return nil
}

// executeProposicoesBackfill executa backfill de proposições por ano
func (sbe *StrategicBackfillExecutor) executeProposicoesBackfill(ctx context.Context, checkpoint *BackfillCheckpoint) error {
	year, ok := checkpoint.Metadata["year"].(float64) // JSON deserializa números como float64
	if !ok {
		return fmt.Errorf("metadado 'year' não encontrado ou inválido no checkpoint")
	}

	yearInt := int(year)
	log.Printf("📜 Executando backfill de proposições para o ano %d", yearInt)

	// Buscar proposições por ano com paginação
	var allProposicoes []domain.Proposicao
	pagina := 1

	// Determinar itens por página (preferir metadado do checkpoint -> strategy.BatchSize -> fallback 100)
	itensPorPagina := 100
	if bs, ok := checkpoint.Metadata["batch_size"].(float64); ok {
		// JSON numbers são float64
		if int(bs) > 0 {
			itensPorPagina = int(bs)
		}
	} else if sbe.strategy.BatchSize > 0 {
		itensPorPagina = sbe.strategy.BatchSize
	}
	// Respeitar limite máximo da API da Câmara (100 itens por página)
	if itensPorPagina > 100 {
		itensPorPagina = 100
	}

	for {
		// Criar filtros para o ano
		filtros := &domain.ProposicaoFilter{
			Ano:        &yearInt,
			Pagina:     pagina,
			Limite:     itensPorPagina,
			Ordem:      "ASC",
			OrdenarPor: "id",
		}

		proposicoes, _, _, err := sbe.proposicoesService.ListarProposicoes(ctx, filtros)
		if err != nil {
			return fmt.Errorf("erro ao buscar proposições ano %d página %d: %w", yearInt, pagina, err)
		}

		// Se não há proposições na página atual, chegamos ao final
		if len(proposicoes) == 0 {
			break
		}

		allProposicoes = append(allProposicoes, proposicoes...)

		// Log de progresso
		if pagina%10 == 0 {
			log.Printf("📄 Página %d processada, coletadas: %d proposições até agora", pagina, len(allProposicoes))
		}

		// Se a página atual retornou menos itens que o limite, provavelmente é a última página
		if len(proposicoes) < itensPorPagina {
			break
		}

		pagina++
	}

	checkpoint.Progress.TotalItems = len(allProposicoes)
	log.Printf("📊 Total de proposições coletadas para %d: %d", yearInt, len(allProposicoes))

	// Processar em lotes
	batchSize := sbe.strategy.BatchSize
	for i := 0; i < len(allProposicoes); i += batchSize {
		end := i + batchSize
		if end > len(allProposicoes) {
			end = len(allProposicoes)
		}

		batch := allProposicoes[i:end]

		// Retry com backoff exponencial
		var lastErr error
		for retry := 0; retry < sbe.strategy.MaxRetries; retry++ {
			if err := sbe.proposicaoRepo.UpsertProposicoes(ctx, batch); err != nil {
				lastErr = err
				log.Printf("⚠️  Tentativa %d/%d falhou para lote proposições %d-%d: %v",
					retry+1, sbe.strategy.MaxRetries, i, end-1, err)

				if retry < sbe.strategy.MaxRetries-1 {
					time.Sleep(sbe.strategy.RetryDelay * time.Duration(retry+1))
				}
				continue
			}

			// Sucesso - atualizar progresso
			checkpoint.Progress.ProcessedItems = end
			if len(batch) > 0 {
				checkpoint.Progress.LastProcessedID = fmt.Sprintf("%d", batch[len(batch)-1].ID)
			}

			if err := sbe.manager.UpdateProgress(ctx, checkpoint,
				checkpoint.Progress.ProcessedItems,
				checkpoint.Progress.FailedItems,
				checkpoint.Progress.LastProcessedID); err != nil {
				log.Printf("⚠️  Erro ao atualizar progresso: %v", err)
			}

			log.Printf("✅ Lote proposições %d-%d (%d) processado", i, end-1, yearInt)
			lastErr = nil
			break
		}

		if lastErr != nil {
			checkpoint.Progress.FailedItems += len(batch)
			return fmt.Errorf("falha após %d tentativas no lote proposições %d-%d: %w",
				sbe.strategy.MaxRetries, i, end-1, lastErr)
		}
	}

	log.Printf("🎉 Backfill proposições %d concluído: %d processadas, %d falhas",
		yearInt, checkpoint.Progress.ProcessedItems, checkpoint.Progress.FailedItems)

	return nil
}

// executeDespesasBackfill executa backfill de despesas por ano
func (sbe *StrategicBackfillExecutor) executeDespesasBackfill(ctx context.Context, checkpoint *BackfillCheckpoint) error {
	year, ok := checkpoint.Metadata["year"].(float64)
	if !ok {
		return fmt.Errorf("metadado 'year' não encontrado ou inválido no checkpoint")
	}

	yearInt := int(year)
	log.Printf("💰 Executando backfill de despesas para o ano %d", yearInt)

	if sbe.deputadosService == nil || sbe.despesaRepo == nil {
		log.Printf("⚠️ Dependências de despesas indisponíveis; checkpoint %s será pulado", checkpoint.ID)
		return nil
	}

	deputados, source, err := sbe.deputadosService.ListarDeputados(ctx, "", "", "")
	if err != nil {
		return fmt.Errorf("erro ao buscar deputados para despesas: %w", err)
	}

	if len(deputados) == 0 {
		log.Printf("⚠️ Nenhum deputado retornado para despesas do ano %d (source=%s)", yearInt, source)
		checkpoint.Progress.TotalItems = 0
		return nil
	}

	checkpoint.Progress.TotalItems = len(deputados)
	log.Printf("📋 %d deputados carregados para despesas %d (source=%s)", len(deputados), yearInt, source)

	startIndex := checkpoint.Progress.ProcessedItems
	if lastID := checkpoint.Progress.LastProcessedID; lastID != "" {
		if parsedID, parseErr := strconv.Atoi(lastID); parseErr == nil {
			for idx, dep := range deputados {
				if dep.ID == parsedID {
					startIndex = idx + 1
					break
				}
			}
		}
	}
	if startIndex < 0 {
		startIndex = 0
	}
	if startIndex > len(deputados) {
		startIndex = len(deputados)
	}

	maxRetries := sbe.strategy.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}
	retryDelay := sbe.strategy.RetryDelay
	if retryDelay <= 0 {
		retryDelay = 3 * time.Second
	}

	var totalDespesas int

	for idx := startIndex; idx < len(deputados); idx++ {
		dep := deputados[idx]
		log.Printf("🔎 Ingerindo despesas do deputado %d (%s) para %d [%d/%d]",
			dep.ID, dep.Nome, yearInt, idx+1, len(deputados))

		var lastErr error
		for attempt := 0; attempt < maxRetries; attempt++ {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			despesas, _, err := sbe.deputadosService.ListarDespesas(ctx,
				fmt.Sprintf("%d", dep.ID),
				fmt.Sprintf("%d", yearInt))
			if err != nil {
				lastErr = err
			} else {
				if len(despesas) == 0 {
					lastErr = nil
					break
				}

				if err := sbe.despesaRepo.UpsertDespesas(ctx, dep.ID, yearInt, despesas); err != nil {
					lastErr = err
				} else {
					totalDespesas += len(despesas)
					lastErr = nil
					break
				}
			}

			if attempt < maxRetries-1 {
				time.Sleep(retryDelay * time.Duration(attempt+1))
			}
		}

		if lastErr != nil {
			checkpoint.Progress.FailedItems++
			if err := sbe.manager.UpdateProgress(ctx, checkpoint,
				checkpoint.Progress.ProcessedItems,
				checkpoint.Progress.FailedItems,
				checkpoint.Progress.LastProcessedID); err != nil {
				log.Printf("⚠️ Erro ao atualizar progresso após falha em despesas: %v", err)
			}
			log.Printf("⚠️ Falha ao sincronizar despesas do deputado %d no ano %d: %v", dep.ID, yearInt, lastErr)
			continue
		}

		checkpoint.Progress.ProcessedItems = idx + 1
		checkpoint.Progress.LastProcessedID = strconv.Itoa(dep.ID)

		if err := sbe.manager.UpdateProgress(ctx, checkpoint,
			checkpoint.Progress.ProcessedItems,
			checkpoint.Progress.FailedItems,
			checkpoint.Progress.LastProcessedID); err != nil {
			log.Printf("⚠️ Erro ao atualizar progresso de despesas: %v", err)
		}

		time.Sleep(150 * time.Millisecond)
	}

	log.Printf("✅ Backfill despesas %d concluído: %d deputados processados, %d despesas ingeridas",
		yearInt, checkpoint.Progress.ProcessedItems, totalDespesas)

	return nil
}

// resumeCheckpoint resume um checkpoint que foi interrompido
func (sbe *StrategicBackfillExecutor) resumeCheckpoint(ctx context.Context, checkpoint *BackfillCheckpoint) error {
	log.Printf("🔄 Resumindo checkpoint %s a partir do progresso: %d/%d",
		checkpoint.ID, checkpoint.Progress.ProcessedItems, checkpoint.Progress.TotalItems)

	return sbe.executeCheckpoint(ctx, checkpoint)
}
