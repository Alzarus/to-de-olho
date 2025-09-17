package ingestor

import (
	"context"
	"fmt"
	"log"
	"time"

	app "to-de-olho-backend/internal/application"
	"to-de-olho-backend/internal/domain"
	"to-de-olho-backend/internal/infrastructure/repository"
)

// StrategicBackfillExecutor executa backfill histórico com estratégia inteligente
type StrategicBackfillExecutor struct {
	manager            *BackfillManager
	deputadosService   *app.DeputadosService
	proposicoesService *app.ProposicoesService
	deputadoRepo       *repository.DeputadoRepository
	proposicaoRepo     *repository.ProposicaoRepository
	strategy           BackfillStrategy
}

// NewStrategicBackfillExecutor cria um novo executor estratégico
func NewStrategicBackfillExecutor(
	manager *BackfillManager,
	deputadosService *app.DeputadosService,
	proposicoesService *app.ProposicoesService,
	deputadoRepo *repository.DeputadoRepository,
	proposicaoRepo *repository.ProposicaoRepository,
	strategy BackfillStrategy,
) *StrategicBackfillExecutor {
	return &StrategicBackfillExecutor{
		manager:            manager,
		deputadosService:   deputadosService,
		proposicoesService: proposicoesService,
		deputadoRepo:       deputadoRepo,
		proposicaoRepo:     proposicaoRepo,
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

	log.Printf("✅ Plano criado: 1 checkpoint deputados + %d proposições + %d despesas",
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
	default:
		return fmt.Errorf("tipo de checkpoint desconhecido: %s", checkpoint.Type)
	}
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
	itensPorPagina := 100

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

// executeDespesasBackfill executa backfill de despesas por ano (placeholder)
func (sbe *StrategicBackfillExecutor) executeDespesasBackfill(ctx context.Context, checkpoint *BackfillCheckpoint) error {
	year, ok := checkpoint.Metadata["year"].(float64)
	if !ok {
		return fmt.Errorf("metadado 'year' não encontrado ou inválido no checkpoint")
	}

	yearInt := int(year)
	log.Printf("💰 Executando backfill de despesas para o ano %d (PLACEHOLDER)", yearInt)

	// TODO: Implementar quando tivermos repositório de despesas
	// Por enquanto, simular processamento para não bloquear o desenvolvimento
	time.Sleep(2 * time.Second)

	checkpoint.Progress.TotalItems = 1
	checkpoint.Progress.ProcessedItems = 1
	checkpoint.Progress.LastProcessedID = fmt.Sprintf("despesas_%d_placeholder", yearInt)

	log.Printf("✅ Backfill despesas %d simulado com sucesso", yearInt)
	return nil
}

// resumeCheckpoint resume um checkpoint que foi interrompido
func (sbe *StrategicBackfillExecutor) resumeCheckpoint(ctx context.Context, checkpoint *BackfillCheckpoint) error {
	log.Printf("🔄 Resumindo checkpoint %s a partir do progresso: %d/%d",
		checkpoint.ID, checkpoint.Progress.ProcessedItems, checkpoint.Progress.TotalItems)

	return sbe.executeCheckpoint(ctx, checkpoint)
}
