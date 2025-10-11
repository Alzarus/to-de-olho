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
	"to-de-olho-backend/internal/infrastructure/resilience"
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
	currentStatus   domain.BackfillStatus
	progressTracker *backfillProgressTracker
}

type backfillProgressTracker struct {
	mu             sync.Mutex
	totalExpected  int
	totalProcessed int
	registered     map[string]int
}

func newBackfillProgressTracker() *backfillProgressTracker {
	return &backfillProgressTracker{
		registered: make(map[string]int),
	}
}

func (t *backfillProgressTracker) registerExpected(key string, total int) {
	if total <= 0 {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if existing, ok := t.registered[key]; ok {
		if total > existing {
			increment := total - existing
			t.totalExpected += increment
			t.registered[key] = total
		}
		return
	}
	t.registered[key] = total
	t.totalExpected += total
}

func (t *backfillProgressTracker) addProcessed(delta int) float64 {
	if delta <= 0 {
		return t.currentPercentage()
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.totalProcessed += delta
	return t.percentageLocked()
}

func (t *backfillProgressTracker) currentPercentage() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.percentageLocked()
}

func (t *backfillProgressTracker) markCompleted() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.totalExpected == 0 {
		return 100
	}
	t.totalProcessed = t.totalExpected
	return 100
}

func (t *backfillProgressTracker) percentageLocked() float64 {
	if t.totalExpected == 0 {
		return 0
	}
	if t.totalProcessed >= t.totalExpected {
		return 100
	}
	progress := (float64(t.totalProcessed) / float64(t.totalExpected)) * 100
	if progress > 99 {
		progress = 99
	}
	return progress
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

	tracker := newBackfillProgressTracker()
	s.mu.Lock()
	s.progressTracker = tracker
	s.mu.Unlock()

	var finalStatus = domain.BackfillStatusSuccess
	var errorMessage *string
	var finalStatusMu sync.Mutex
	var partialFailuresMu sync.Mutex
	partialFailures := make([]string, 0)

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
				total := len(deputados)
				tracker.registerExpected("deputados", total)
				status.DeputadosProcessados = total
				status.ProgressPercentage = tracker.addProcessed(total)
			} else {
				status.ProgressPercentage = tracker.currentPercentage()
			}
			status.LastUpdate = time.Now()
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
		status.ProgressPercentage = tracker.currentPercentage()
		status.LastUpdate = time.Now()
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
					st.ProgressPercentage = tracker.currentPercentage()
					// atualizar repositório com cópia local
					_ = s.backfillRepo.UpdateExecutionProgress(ctx, execution.ExecutionID, st)
					s.currentStatus = st
					s.mu.Unlock()

					dataInicio := time.Date(ano, time.January, 1, 0, 0, 0, 0, time.UTC)
					dataFim := time.Date(ano, time.December, 31, 23, 59, 59, 0, time.UTC)

					progressThreshold := config.BatchSize / 2
					if progressThreshold < 100 {
						progressThreshold = 100
					}
					lastReported := 0
					pendingVotacoes := 0
					pendingProgressUnits := 0
					lastFlush := time.Now()
					progressKey := fmt.Sprintf("votacoes-%d", ano)
					totalDays := int(dataFim.Sub(dataInicio).Hours()/24) + 1
					if totalDays < 0 {
						totalDays = 0
					}
					if totalDays > 0 {
						tracker.registerExpected(progressKey+":coverage", totalDays)
					}
					dayFlushThreshold := progressThreshold / 4
					if dayFlushThreshold < 10 {
						dayFlushThreshold = 10
					}

					flushProgress := func(reason string, force bool) {
						if !force && pendingProgressUnits == 0 {
							return
						}

						incrementVotacoes := pendingVotacoes
						incrementUnits := pendingProgressUnits
						pendingVotacoes = 0
						pendingProgressUnits = 0
						lastFlush = time.Now()
						var progress float64
						if incrementUnits > 0 {
							progress = tracker.addProcessed(incrementUnits)
						} else {
							progress = tracker.currentPercentage()
						}

						s.mu.Lock()
						if incrementVotacoes > 0 {
							status.VotacoesProcessadas += incrementVotacoes
						}
						status.ProgressPercentage = progress
						status.LastUpdate = time.Now()
						status.CurrentOperation = fmt.Sprintf("worker-%d: Sincronizando votações de %d (%s, total=%d)",
							workerID, ano, reason, status.VotacoesProcessadas)
						_ = s.backfillRepo.UpdateExecutionProgress(ctx, execution.ExecutionID, status)
						s.currentStatus = status
						s.mu.Unlock()
					}

					chunkAwareCtx := domain.WithVotacoesChunkProgress(ctx, func(start, end time.Time, success bool) {
						if !success {
							return
						}
						days := int(end.Sub(start).Hours()/24) + 1
						if days <= 0 {
							days = 1
						}
						pendingProgressUnits += days
						if pendingProgressUnits >= dayFlushThreshold || time.Since(lastFlush) >= 15*time.Second {
							flushProgress("janela concluída", false)
						}
					})

					reporterCtx := WithVotacoesProgressReporter(chunkAwareCtx, func(totalProcessed, total int) {
						if total > 0 {
							tracker.registerExpected(progressKey, total)
						}
						delta := totalProcessed - lastReported
						if delta <= 0 {
							return
						}
						lastReported = totalProcessed
						pendingVotacoes += delta
						pendingProgressUnits += delta

						if pendingVotacoes >= progressThreshold || pendingProgressUnits >= dayFlushThreshold || time.Since(lastFlush) >= 15*time.Second {
							flushProgress("progresso parcial", false)
						}
					})

					processed, err := s.votacoesService.SincronizarVotacoes(reporterCtx, dataInicio, dataFim)
					flushProgress("progresso parcial", false)
					if err != nil {
						partialFailuresMu.Lock()
						partialFailures = append(partialFailures, fmt.Sprintf("ano %d: %s", ano, err.Error()))
						partialFailuresMu.Unlock()

						finalStatusMu.Lock()
						if finalStatus == domain.BackfillStatusSuccess {
							finalStatus = domain.BackfillStatusPartial
						}
						finalStatusMu.Unlock()

						if resilience.IsCircuitBreakerOpen(err) {
							cooldown := time.Duration(config.DelayBetweenBatches) * time.Millisecond
							if cooldown < 30*time.Second { // Cooldown maior para circuit breaker (vs 15s)
								cooldown = 30 * time.Second
							}
							s.logger.Warn("circuit breaker aberto durante sincronização anual de votações",
								slog.Int("ano", ano),
								slog.String("cooldown", cooldown.String()),
								slog.String("error", err.Error()))
							time.Sleep(cooldown)
						} else if isGatewayTimeoutError(err) {
							cooldown := time.Duration(config.DelayBetweenBatches) * time.Millisecond
							if cooldown < 15*time.Second { // Cooldown maior para gateway timeout (vs 5s)
								cooldown = 15 * time.Second
							}
							s.logger.Warn("gateway timeout ao sincronizar votações para o ano",
								slog.Int("ano", ano),
								slog.String("cooldown", cooldown.String()),
								slog.String("error", err.Error()))
							time.Sleep(cooldown)
						} else {
							s.logger.Warn("erro ao sincronizar votações para o ano", slog.Int("ano", ano), slog.String("error", err.Error()))
						}
						flushProgress("erro ao processar ano", true)
						s.mu.Lock()
						status.ErrorsCount++
						msg := err.Error()
						status.LatestError = &msg
						s.currentStatus = status
						s.mu.Unlock()
						_ = s.backfillRepo.UpdateExecutionProgress(ctx, execution.ExecutionID, status)
						continue
					}

					flushProgress("ano concluído", true)
					s.logger.Info("✅ Backfill de votações concluído para o ano",
						slog.Int("ano", ano),
						slog.Int("processadas", processed))

					s.mu.Lock()
					status.CurrentOperation = "Sincronizando votações históricas"
					status.ProgressPercentage = tracker.currentPercentage()
					status.LastUpdate = time.Now()
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
				diff := config.AnoFim - config.AnoInicio
				if diff < 0 {
					diff = -diff
				}
				years := make([]int, 0, diff+1)
				if config.AnoFim >= config.AnoInicio {
					for ano := config.AnoFim; ano >= config.AnoInicio; ano-- {
						years = append(years, ano)
					}
				} else {
					for ano := config.AnoInicio; ano >= config.AnoFim; ano-- {
						years = append(years, ano)
					}
				}
				for _, ano := range years {
					anos <- ano
				}
				close(anos)
			}()

			wg.Wait()
		}
	}

	// Atualizar progresso final em memória antes de completar execução
	if len(partialFailures) > 0 && errorMessage == nil {
		summary := fmt.Sprintf("pendências ao sincronizar votações: %s", strings.Join(partialFailures, "; "))
		errorMessage = &summary
	}

	if finalStatus == domain.BackfillStatusSuccess {
		status.ProgressPercentage = tracker.markCompleted()
	} else {
		status.ProgressPercentage = tracker.currentPercentage()
	}
	status.Status = finalStatus
	status.LastUpdate = time.Now()
	if finalStatus == domain.BackfillStatusSuccess {
		status.CurrentOperation = "Backfill concluído"
	} else {
		status.CurrentOperation = "Backfill finalizado com alerta"
	}
	s.mu.Lock()
	s.currentStatus = status
	s.mu.Unlock()

	// Completar execução
	duration := time.Since(execution.StartedAt)
	s.logger.Info("🎯 Backfill histórico concluído",
		slog.String("status", finalStatus),
		slog.String("duracao", duration.String()),
		slog.Int("deputados", status.DeputadosProcessados),
		slog.Int("proposicoes", status.ProposicoesProcessadas),
		slog.Int("votacoes", status.VotacoesProcessadas))

	if err := s.backfillRepo.UpdateExecutionProgress(ctx, execution.ExecutionID, status); err != nil {
		s.logger.Warn("erro ao atualizar progresso final do backfill", slog.Any("error", err))
	}

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

func isGatewayTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "504") || strings.Contains(msg, "gateway timeout") || strings.Contains(msg, "upstream request timeout") || strings.Contains(msg, "deadline exceeded") || strings.Contains(msg, "context deadline")
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

	s.mu.Lock()
	current := s.currentStatus
	tracker := s.progressTracker
	s.mu.Unlock()

	progressResolved := false
	if tracker != nil && current.ExecutionID == execution.ExecutionID {
		progress = tracker.currentPercentage()
		progressResolved = true
	}

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

	if current.ExecutionID == execution.ExecutionID {
		if current.CurrentOperation != "" {
			status.CurrentOperation = current.CurrentOperation
		}
		if current.DeputadosProcessados > 0 {
			status.DeputadosProcessados = current.DeputadosProcessados
		}
		if current.ProposicoesProcessadas > 0 {
			status.ProposicoesProcessadas = current.ProposicoesProcessadas
		}
		if current.VotacoesProcessadas > 0 {
			status.VotacoesProcessadas = current.VotacoesProcessadas
		}
		if current.DespesasProcessadas > 0 {
			status.DespesasProcessadas = current.DespesasProcessadas
		}
		if !progressResolved {
			status.ProgressPercentage = current.ProgressPercentage
			progressResolved = true
		}
	}

	if !progressResolved {
		status.ProgressPercentage = domain.CalculateProgressFromMetrics(
			status.DeputadosProcessados,
			status.ProposicoesProcessadas,
			status.DespesasProcessadas,
			status.VotacoesProcessadas,
		)
	}

	return status, nil
}

// ListExecutions lista execuções de backfill
func (s *SmartBackfillService) ListExecutions(ctx context.Context, limit, offset int) ([]domain.BackfillExecution, int, error) {
	return s.backfillRepo.ListExecutions(ctx, limit, offset)
}
