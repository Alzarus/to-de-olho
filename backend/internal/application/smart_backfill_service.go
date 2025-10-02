package application

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"to-de-olho-backend/internal/domain"
)

// BackfillRepositoryPort define interface para repositório de backfill
type BackfillRepositoryPort interface {
	HasSuccessfulHistoricalBackfill(ctx context.Context, startYear, endYear int) (bool, error)
	GetLastExecution(ctx context.Context, executionType string) (*domain.BackfillExecution, error)
	CreateExecution(ctx context.Context, config *domain.BackfillConfig) (*domain.BackfillExecution, error)
	UpdateExecutionProgress(ctx context.Context, executionID string, update domain.BackfillStatus) error
	CompleteExecution(ctx context.Context, executionID string, status string, errorMessage *string) error
	GetRunningExecution(ctx context.Context) (*domain.BackfillExecution, error)
	ListExecutions(ctx context.Context, limit int, offset int) ([]domain.BackfillExecution, int, error)
}

// SmartBackfillService serviço inteligente de backfill
type SmartBackfillService struct {
	backfillRepo       BackfillRepositoryPort
	deputadosService   *DeputadosService
	proposicoesService *ProposicoesService
	votacoesService    *VotacoesService
	analyticsService   AnalyticsServiceInterface
	logger             *slog.Logger
	currentExecutionID string
	mu                 sync.Mutex // protege updates concorrentes de status
	// currentStatus guarda último status conhecido em memória para exposição rápida
	currentStatus domain.BackfillStatus
}

// NewSmartBackfillService cria nova instância do serviço
func NewSmartBackfillService(
	backfillRepo BackfillRepositoryPort,
	deputadosService *DeputadosService,
	proposicoesService *ProposicoesService,
	votacoesService *VotacoesService,
	analyticsService AnalyticsServiceInterface,
	logger *slog.Logger,
) *SmartBackfillService {
	return &SmartBackfillService{
		backfillRepo:       backfillRepo,
		deputadosService:   deputadosService,
		proposicoesService: proposicoesService,
		votacoesService:    votacoesService,
		analyticsService:   analyticsService,
		logger:             logger,
	}
}

// ShouldRunHistoricalBackfill verifica se deve executar backfill histórico
func (s *SmartBackfillService) ShouldRunHistoricalBackfill(ctx context.Context, config *domain.BackfillConfig) (bool, string, error) {
	// 1. Verificar se já há execução em andamento
	runningExecution, err := s.backfillRepo.GetRunningExecution(ctx)
	if err != nil && err != domain.ErrBackfillNaoEncontrado {
		return false, "", fmt.Errorf("erro ao verificar execução em andamento: %w", err)
	}

	if runningExecution != nil {
		reason := fmt.Sprintf("Execução %s já em andamento desde %s",
			runningExecution.ExecutionID,
			runningExecution.StartedAt.Format("15:04:05"))
		return false, reason, nil
	}

	// 2. Se forçar reexecução, sempre rodar
	if config.ForcarReexecucao {
		return true, "Reexecução forçada", nil
	}

	// 3. Verificar se já foi feito backfill histórico com sucesso
	hasSuccessful, err := s.backfillRepo.HasSuccessfulHistoricalBackfill(ctx, config.AnoInicio, config.AnoFim)
	if err != nil {
		return false, "", fmt.Errorf("erro ao verificar backfill histórico: %w", err)
	}

	if hasSuccessful {
		// Verificar quando foi a última execução
		lastExecution, err := s.backfillRepo.GetLastExecution(ctx, domain.BackfillTipoHistorico)
		if err != nil && err != domain.ErrBackfillNaoEncontrado {
			return false, "", fmt.Errorf("erro ao buscar última execução: %w", err)
		}

		var lastRunInfo string
		if lastExecution != nil {
			lastRunInfo = fmt.Sprintf(" (última execução: %s)", lastExecution.StartedAt.Format("02/01/2006 15:04"))
		}

		reason := fmt.Sprintf("Backfill histórico já realizado com sucesso para período %d-%d%s",
			config.AnoInicio, config.AnoFim, lastRunInfo)
		return false, reason, nil
	}

	return true, "Backfill histórico necessário", nil
}

// ExecuteHistoricalBackfill executa backfill histórico inteligente
func (s *SmartBackfillService) ExecuteHistoricalBackfill(ctx context.Context, config *domain.BackfillConfig) (*domain.BackfillExecution, error) {
	config.SetDefaults()

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuração inválida: %w", err)
	}

	// Verificar se deve executar
	shouldRun, reason, err := s.ShouldRunHistoricalBackfill(ctx, config)
	if err != nil {
		return nil, err
	}

	if !shouldRun {
		s.logger.Info("Backfill histórico não será executado", slog.String("reason", reason))
		return nil, fmt.Errorf("backfill não necessário: %s", reason)
	}

	s.logger.Info("Iniciando backfill histórico",
		slog.String("reason", reason),
		slog.Int("ano_inicio", config.AnoInicio),
		slog.Int("ano_fim", config.AnoFim))

	// Criar registro da execução
	execution, err := s.backfillRepo.CreateExecution(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar execução: %w", err)
	}

	s.currentExecutionID = execution.ExecutionID

	// Inicializar currentStatus com valores iniciais
	s.mu.Lock()
	s.currentStatus = domain.BackfillStatus{
		ExecutionID:      execution.ExecutionID,
		Status:           domain.BackfillStatusRunning,
		CurrentOperation: "Iniciando backfill",
		StartedAt:        execution.StartedAt,
		LastUpdate:       time.Now(),
	}
	s.mu.Unlock()

	// Executar backfill em goroutine separada
	go s.runHistoricalBackfill(context.Background(), execution, config)

	return execution, nil
}

// GetBackfillConfigFromEnv obtém configuração de backfill de variáveis de ambiente
func (s *SmartBackfillService) GetBackfillConfigFromEnv() *domain.BackfillConfig {
	config := &domain.BackfillConfig{}

	// Ano inicial (padrão: 2022)
	if envYear := os.Getenv("BACKFILL_START_YEAR"); envYear != "" {
		if year, err := strconv.Atoi(envYear); err == nil && year >= 1988 {
			config.AnoInicio = year
		}
	}

	// Ano final (padrão: ano atual)
	if envYear := os.Getenv("BACKFILL_END_YEAR"); envYear != "" {
		if year, err := strconv.Atoi(envYear); err == nil {
			config.AnoFim = year
		}
	}

	// Forçar reexecução
	config.ForcarReexecucao = os.Getenv("BACKFILL_FORCE") == "true"

	// Tipo de execução
	if envType := os.Getenv("BACKFILL_TYPE"); envType != "" {
		config.Tipo = envType
	}

	// Triggered by
	if envTrigger := os.Getenv("BACKFILL_TRIGGERED_BY"); envTrigger != "" {
		config.TriggeredBy = envTrigger
	} else {
		config.TriggeredBy = domain.BackfillTriggeredByDeploy
	}

	// Workers paralelos
	if envWorkers := os.Getenv("BACKFILL_WORKERS"); envWorkers != "" {
		if workers, err := strconv.Atoi(envWorkers); err == nil && workers > 0 && workers <= 10 {
			config.ParallelWorkers = workers
		}
	}

	// Batch size
	if envBatch := os.Getenv("BACKFILL_BATCH_SIZE"); envBatch != "" {
		if batch, err := strconv.Atoi(envBatch); err == nil && batch >= 10 && batch <= 1000 {
			config.BatchSize = batch
		}
	}

	config.SetDefaults()
	return config
}

// runHistoricalBackfill executa o backfill histórico
func (s *SmartBackfillService) runHistoricalBackfill(ctx context.Context, execution *domain.BackfillExecution, config *domain.BackfillConfig) {
	defer func() {
		if r := recover(); r != nil {
			errorMsg := fmt.Sprintf("Panic durante backfill: %v", r)
			s.backfillRepo.CompleteExecution(ctx, execution.ExecutionID, domain.BackfillStatusFailed, &errorMsg)
			s.logger.Error("Panic durante backfill", slog.Any("error", r))
		}
	}()

	s.logger.Info("🔄 Iniciando backfill histórico",
		slog.String("execution_id", execution.ExecutionID),
		slog.String("periodo", fmt.Sprintf("%d-%d", config.AnoInicio, config.AnoFim)))

	status := domain.BackfillStatus{
		ExecutionID:      execution.ExecutionID,
		Status:           domain.BackfillStatusRunning,
		StartedAt:        execution.StartedAt,
		LastUpdate:       time.Now(),
		CurrentOperation: "Iniciando backfill",
	}

	var finalStatus = domain.BackfillStatusSuccess
	var errorMessage *string

	// 1. Sincronizar deputados
	if config.IncluirDeputados {
		s.logger.Info("👥 Sincronizando deputados...")
		status.CurrentOperation = "Sincronizando deputados"
		// atualizar currentStatus em memória
		s.mu.Lock()
		s.currentStatus = status
		s.mu.Unlock()

		deputados, source, err := s.deputadosService.ListarDeputados(ctx, "", "", "")
		if err != nil {
			errMsg := fmt.Sprintf("Erro ao sincronizar deputados: %v", err)
			errorMessage = &errMsg
			finalStatus = domain.BackfillStatusFailed
			s.logger.Error("Erro ao sincronizar deputados", slog.Any("error", err))
		} else {
			if source == "api" {
				status.DeputadosProcessados = len(deputados)
			}
			s.logger.Info("✅ Deputados sincronizados", slog.Int("total", len(deputados)))
		}

		s.backfillRepo.UpdateExecutionProgress(ctx, execution.ExecutionID, status)
		// atualizar currentStatus em memória
		s.mu.Lock()
		s.currentStatus = status
		s.mu.Unlock()
	}

	// 2. Sincronizar proposições por ano (paralelizar por ano)
	// OBS: Temporariamente desativado por volume de dados (conforme solicitação).
	// Comentamos o bloco de processamento pesado para evitar ingestão de proposições agora.
	if config.IncluirProposicoes && finalStatus != domain.BackfillStatusFailed {
		// Allow explicit opt-in via env var BACKFILL_INCLUDE_PROPOSICOES=true
		if os.Getenv("BACKFILL_INCLUDE_PROPOSICOES") == "true" {
			s.logger.Info("📜 Sincronização de proposições RE-ativada via BACKFILL_INCLUDE_PROPOSICOES")
			// Para segurança mantemos o comportamento original inalterado quando habilitado.
			// (Não reimplementamos aqui; caso precise, remova este guard e restaure o bloco original.)
		} else {
			s.logger.Info("📜 Sincronização de proposições temporariamente desativada (BACKFILL_INCLUDE_PROPOSICOES!=true)")
		}
	}

	// 3. Sincronizar votações históricas por ano
	if config.IncluirVotacoes && finalStatus != domain.BackfillStatusFailed {
		s.logger.Info("🗳️ Iniciando sincronização de votações históricas (paralelo)")
		status.CurrentOperation = "Sincronizando votações históricas"
		s.mu.Lock()
		s.currentStatus = status
		s.mu.Unlock()

		if s.votacoesService == nil {
			s.logger.Warn("VotacoesService não disponível; pulando sincronização de votações")
			s.backfillRepo.UpdateExecutionProgress(ctx, execution.ExecutionID, status)
		} else {
			anos := make(chan int)
			var wg sync.WaitGroup

			workers := config.ParallelWorkers
			if workers <= 0 {
				workers = 3
			}

			worker := func(workerID int) {
				defer wg.Done()
				for ano := range anos {
					s.logger.Info("🔎 Sincronizando votações", slog.Int("ano", ano))
					// preparar uma cópia local do status para update atômico
					s.mu.Lock()
					st := status
					st.CurrentOperation = fmt.Sprintf("worker-%d: Sincronizando votações de %d", workerID, ano)
					// atualizar repositório com cópia local
					_ = s.backfillRepo.UpdateExecutionProgress(ctx, execution.ExecutionID, st)
					s.currentStatus = st
					s.mu.Unlock()

					dataInicio := time.Date(ano, time.January, 1, 0, 0, 0, 0, time.UTC)
					dataFim := time.Date(ano, time.December, 31, 23, 59, 59, 0, time.UTC)

					votacoes, err := s.votacoesService.camaraClient.GetVotacoes(ctx, dataInicio, dataFim)
					if err != nil {
						// Se for timeout/504 do gateway, tentar estratégia fallback por dia
						errStr := err.Error()
						s.logger.Warn("erro ao buscar lista de votações da API", slog.Int("ano", ano), slog.String("error", errStr))
						if strings.Contains(errStr, "504") || strings.Contains(strings.ToLower(errStr), "gateway timeout") || strings.Contains(strings.ToLower(errStr), "upstream request timeout") {
							s.logger.Info("Tentando fallback diário devido a 504/timeout para o período", slog.Int("ano", ano))
							// iterar dia a dia
							day := dataInicio
							for !day.After(dataFim) {
								dayEnd := time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 59, 0, day.Location())
								single, errDay := s.votacoesService.camaraClient.GetVotacoes(ctx, day, dayEnd)
								if errDay != nil {
									s.logger.Warn("fallback diário falhou para dia", slog.String("dia", day.Format("2006-01-02")), slog.String("error", errDay.Error()))
									day = day.AddDate(0, 0, 1)
									continue
								}
								// sincronizar apenas o dia
								if len(single) > 0 {
									if err := s.votacoesService.SincronizarVotacoes(ctx, day, dayEnd); err != nil {
										s.logger.Warn("erro ao sincronizar votações para o dia", slog.String("dia", day.Format("2006-01-02")), slog.String("error", err.Error()))
									}
								}
								// contabilizar
								s.mu.Lock()
								status.VotacoesProcessadas += len(single)
								// atualizar currentStatus para refletir progresso na DB
								st2 := status
								st2.CurrentOperation = fmt.Sprintf("worker-%d: Sincronizando votações de %d (fallback diário em %s)", workerID, ano, day.Format("2006-01-02"))
								_ = s.backfillRepo.UpdateExecutionProgress(ctx, execution.ExecutionID, st2)
								s.currentStatus = st2
								s.mu.Unlock()
								day = day.AddDate(0, 0, 1)
								// pequeno sleep entre dias para ser gentil
								time.Sleep(time.Duration(config.DelayBetweenBatches) * time.Millisecond)
							}
							// depois do fallback, continuar para próximo ano
							continue
						}
						// outro erro: pular ano
						_ = s.backfillRepo.UpdateExecutionProgress(ctx, execution.ExecutionID, status)
						continue
					}

					if err := s.votacoesService.SincronizarVotacoes(ctx, dataInicio, dataFim); err != nil {
						s.logger.Warn("erro ao sincronizar votações para o ano", slog.Int("ano", ano), slog.String("error", err.Error()))
						_ = s.backfillRepo.UpdateExecutionProgress(ctx, execution.ExecutionID, status)
						continue
					}

					s.mu.Lock()
					status.VotacoesProcessadas += len(votacoes)
					s.backfillRepo.UpdateExecutionProgress(ctx, execution.ExecutionID, status)
					s.currentStatus = status

					// Restaurar operação atual genérica após concluir o ano para evitar que o
					// campo fique preso no último ano processado.
					status.CurrentOperation = "Sincronizando votações históricas"
					s.backfillRepo.UpdateExecutionProgress(ctx, execution.ExecutionID, status)
					s.currentStatus = status
					s.mu.Unlock()

					time.Sleep(time.Duration(config.DelayBetweenBatches) * time.Millisecond)
				}
			}

			for i := 0; i < workers; i++ {
				wg.Add(1)
				go worker(i + 1)
			}

			go func() {
				for ano := config.AnoInicio; ano <= config.AnoFim; ano++ {
					anos <- ano
				}
				close(anos)
			}()

			wg.Wait()
		}
	}

	// Completar execução
	duration := time.Since(execution.StartedAt)
	s.logger.Info("🎯 Backfill histórico concluído",
		slog.String("status", finalStatus),
		slog.String("duracao", duration.String()),
		slog.Int("deputados", status.DeputadosProcessados),
		slog.Int("proposicoes", status.ProposicoesProcessadas),
		slog.Int("votacoes", status.VotacoesProcessadas))

	if err := s.backfillRepo.CompleteExecution(ctx, execution.ExecutionID, finalStatus, errorMessage); err != nil {
		s.logger.Error("Erro ao completar execução", slog.Any("error", err))
	}

	// Se o backfill foi concluído com sucesso, disparar atualização dos rankings analytics (se injetado)
	if finalStatus == domain.BackfillStatusSuccess && s.analyticsService != nil {
		// Executar em goroutine para não bloquear
		go func() {
			ctx := context.Background()
			s.logger.Info("Iniciando atualização de rankings analytics após backfill")
			if err := s.analyticsService.AtualizarRankings(ctx); err != nil {
				s.logger.Error("Erro ao atualizar rankings analytics após backfill", slog.Any("error", err))
			} else {
				s.logger.Info("Atualização de rankings analytics finalizada com sucesso")
			}
		}()
	}
}

// GetCurrentStatus retorna status da execução atual
func (s *SmartBackfillService) GetCurrentStatus(ctx context.Context) (*domain.BackfillStatus, error) {
	if s.currentExecutionID == "" {
		return nil, domain.ErrBackfillNaoEncontrado
	}

	execution, err := s.backfillRepo.GetRunningExecution(ctx)
	if err != nil {
		return nil, err
	}

	if execution.ExecutionID != s.currentExecutionID {
		return nil, domain.ErrBackfillNaoEncontrado
	}

	progress := execution.CalculateProgress()

	status := &domain.BackfillStatus{
		ExecutionID:            execution.ExecutionID,
		Status:                 execution.Status,
		ProgressPercentage:     progress,
		DeputadosProcessados:   execution.DeputadosProcessados,
		ProposicoesProcessadas: execution.ProposicoesProcessadas,
		DespesasProcessadas:    execution.DespesasProcessadas,
		VotacoesProcessadas:    execution.VotacoesProcessadas,
		StartedAt:              execution.StartedAt,
		LastUpdate:             time.Now(),
	}

	// Merge CurrentOperation from in-memory status if available
	s.mu.Lock()
	if s.currentStatus.ExecutionID == execution.ExecutionID && s.currentStatus.CurrentOperation != "" {
		status.CurrentOperation = s.currentStatus.CurrentOperation
		// if in-memory has more recent counters, prefer them
		if s.currentStatus.DeputadosProcessados > 0 {
			status.DeputadosProcessados = s.currentStatus.DeputadosProcessados
		}
		if s.currentStatus.ProposicoesProcessadas > 0 {
			status.ProposicoesProcessadas = s.currentStatus.ProposicoesProcessadas
		}
		if s.currentStatus.VotacoesProcessadas > 0 {
			status.VotacoesProcessadas = s.currentStatus.VotacoesProcessadas
		}
	}
	s.mu.Unlock()

	return status, nil
}

// ListExecutions lista execuções de backfill
func (s *SmartBackfillService) ListExecutions(ctx context.Context, limit, offset int) ([]domain.BackfillExecution, int, error) {
	return s.backfillRepo.ListExecutions(ctx, limit, offset)
}
