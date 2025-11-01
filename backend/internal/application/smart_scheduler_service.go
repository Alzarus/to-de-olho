package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"to-de-olho-backend/internal/domain"
	"to-de-olho-backend/internal/pkg/envutils"
	"to-de-olho-backend/internal/pkg/metrics"
)

// SchedulerRepositoryPort interface para o repository de scheduler
type SchedulerRepositoryPort interface {
	CreateExecution(ctx context.Context, config *domain.SchedulerConfig) (*domain.SchedulerExecution, error)
	UpdateExecutionProgress(ctx context.Context, executionID string, update map[string]interface{}) error
	CompleteExecution(ctx context.Context, executionID string, status string, errorMessage *string, nextExecution *time.Time) error
	ShouldSchedulerExecute(ctx context.Context, schedulerTipo string, minIntervalHours int) (*domain.ShouldExecuteResult, error)
	GetCurrentStatus(ctx context.Context, schedulerTipo *string) (*domain.SchedulerStatus, error)
	ListExecutions(ctx context.Context, limit, offset int, schedulerTipo *string) ([]domain.SchedulerExecution, int, error)
	GetLastSuccessfulExecution(ctx context.Context, schedulerTipo string) (*domain.SchedulerExecution, error)
	CleanupOldExecutions(ctx context.Context) (int, error)
}

// IngestorPort representa o contrato mínimo que o scheduler precisa do ingestor
type IngestorPort interface {
	// ExecuteDailySync executa a sincronização diária (dados recentes)
	ExecuteDailySync(ctx context.Context) error
}

// SmartSchedulerService gerencia execuções inteligentes de scheduler
type SmartSchedulerService struct {
	schedulerRepo      SchedulerRepositoryPort
	ingestor           IngestorPort
	deputadosService   *DeputadosService
	proposicoesService *ProposicoesService
	votacoesService    *VotacoesService
	logger             *slog.Logger
	mu                 sync.Mutex // Previne execuções simultâneas
}

// NewSmartSchedulerService cria uma nova instância do serviço
func NewSmartSchedulerService(
	schedulerRepo SchedulerRepositoryPort,
	ingestor IngestorPort,
	deputadosService *DeputadosService,
	proposicoesService *ProposicoesService,
	votacoesService *VotacoesService,
	logger *slog.Logger,
) *SmartSchedulerService {
	return &SmartSchedulerService{
		schedulerRepo:      schedulerRepo,
		ingestor:           ingestor,
		deputadosService:   deputadosService,
		proposicoesService: proposicoesService,
		votacoesService:    votacoesService,
		logger:             logger,
	}
}

// ShouldRunScheduler verifica se um scheduler deve executar baseado no tipo e intervalo
func (s *SmartSchedulerService) ShouldRunScheduler(ctx context.Context, schedulerTipo string, config *domain.SchedulerConfig) (bool, string, error) {
	result, err := s.schedulerRepo.ShouldSchedulerExecute(ctx, schedulerTipo, config.MinIntervalHours)
	if err != nil {
		return false, "", fmt.Errorf("erro ao verificar se scheduler deve executar: %w", err)
	}

	s.logger.Info("🤖 Decisão inteligente de scheduler",
		slog.String("tipo", schedulerTipo),
		slog.Bool("deve_executar", result.ShouldRun),
		slog.String("razao", result.Reason),
		slog.Int("intervalo_minimo_horas", config.MinIntervalHours))

	return result.ShouldRun, result.Reason, nil
}

// ExecuteIntelligentScheduler executa scheduler com controle inteligente
func (s *SmartSchedulerService) ExecuteIntelligentScheduler(ctx context.Context, config *domain.SchedulerConfig) (*domain.SchedulerExecution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validar configuração
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuração inválida: %w", err)
	}

	// Verificar se deve executar
	shouldRun, reason, err := s.ShouldRunScheduler(ctx, config.Tipo, config)
	if err != nil {
		return nil, err
	}

	if !shouldRun {
		s.logger.Info("⏭️ Scheduler pulado",
			slog.String("tipo", config.Tipo),
			slog.String("razao", reason))
		return nil, fmt.Errorf("scheduler não deve executar: %s", reason)
	}

	s.logger.Info("🚀 Iniciando execução inteligente de scheduler",
		slog.String("tipo", config.Tipo),
		slog.String("razao", reason),
		slog.String("triggered_by", config.TriggeredBy))

	// Criar execução no banco
	execution, err := s.schedulerRepo.CreateExecution(ctx, config)
	if err != nil {
		if errors.Is(err, domain.ErrSchedulerAlreadyRunning) {
			s.logger.Info("⏳ Scheduler já em execução, pulando criação de nova execução",
				slog.String("tipo", config.Tipo),
				slog.String("triggered_by", config.TriggeredBy))
			// Increment metric for observability
			metrics.IncSchedulerSkip(config.Tipo)
			// Graceful skip: return with wrapped error so callers can decide
			return nil, fmt.Errorf("scheduler já em execução: %w", err)
		}
		return nil, fmt.Errorf("erro ao criar execução: %w", err)
	}

	// Executar em goroutine para não bloquear
	go s.runSchedulerExecution(context.Background(), execution, config)

	return execution, nil
}

// runSchedulerExecution executa o scheduler e atualiza o progresso
func (s *SmartSchedulerService) runSchedulerExecution(ctx context.Context, execution *domain.SchedulerExecution, config *domain.SchedulerConfig) {
	defer func() {
		if r := recover(); r != nil {
			errorMsg := fmt.Sprintf("Panic durante execução: %v", r)
			s.schedulerRepo.CompleteExecution(ctx, execution.ExecutionID, domain.BackfillStatusFailed, &errorMsg, nil)
			s.logger.Error("💥 Panic em execução de scheduler",
				slog.String("execution_id", execution.ExecutionID),
				slog.Any("error", r))
		}
	}()

	s.logger.Info("⚡ Executando scheduler",
		slog.String("execution_id", execution.ExecutionID),
		slog.String("tipo", execution.Tipo))

	var totalSincronizados int
	var executionError error

	// Executar sincronizações baseado na configuração
	if config.IncluirDeputados {
		if count, err := s.sincronizarDeputados(ctx, execution.ExecutionID); err != nil {
			s.logger.Error("❌ Erro ao sincronizar deputados", slog.String("error", err.Error()))
			executionError = err
		} else {
			totalSincronizados += count
			s.schedulerRepo.UpdateExecutionProgress(ctx, execution.ExecutionID, map[string]interface{}{
				"deputados_sincronizados": count,
			})
		}
	}

	// Proposições: podem ser desabilitadas via SCHEDULER_INCLUDE_PROPOSICOES=false quando necessário
	if config.IncluirProposicoes && executionError == nil {
		if envutils.IsEnabled(os.Getenv("SCHEDULER_INCLUDE_PROPOSICOES"), true) {
			if count, err := s.sincronizarProposicoes(ctx, execution.ExecutionID); err != nil {
				s.logger.Error("❌ Erro ao sincronizar proposições", slog.String("error", err.Error()))
				executionError = err
			} else {
				totalSincronizados += count
				s.schedulerRepo.UpdateExecutionProgress(ctx, execution.ExecutionID, map[string]interface{}{
					"proposicoes_sincronizadas": count,
				})
			}
		} else {
			s.logger.Info("📋 Sincronização de proposições desativada via flag", slog.String("execution_id", execution.ExecutionID))
		}
	}

	if config.IncluirDespesas && executionError == nil {
		if envutils.IsEnabled(os.Getenv("SCHEDULER_INCLUDE_DESPESAS"), true) {
			if count, err := s.sincronizarDespesas(ctx, execution.ExecutionID); err != nil {
				s.logger.Error("❌ Erro ao sincronizar despesas", slog.String("error", err.Error()))
				executionError = err
			} else {
				totalSincronizados += count
				s.schedulerRepo.UpdateExecutionProgress(ctx, execution.ExecutionID, map[string]interface{}{
					"despesas_sincronizadas": count,
				})
			}
		} else {
			s.logger.Info("💤 Sincronização de despesas desativada via flag", slog.String("execution_id", execution.ExecutionID))
		}
	}

	// Votações: podem ser desabilitadas via SCHEDULER_INCLUDE_VOTACOES=false em situações de mitigação de carga
	if config.IncluirVotacoes && executionError == nil {
		if envutils.IsEnabled(os.Getenv("SCHEDULER_INCLUDE_VOTACOES"), true) {
			if count, err := s.sincronizarVotacoes(ctx, execution.ExecutionID); err != nil {
				s.logger.Error("❌ Erro ao sincronizar votações", slog.String("error", err.Error()))
				executionError = err
			} else {
				totalSincronizados += count
				s.schedulerRepo.UpdateExecutionProgress(ctx, execution.ExecutionID, map[string]interface{}{
					"votacoes_sincronizadas": count,
				})
			}
		} else {
			s.logger.Info("🗳️ Sincronização de votações desativada via flag", slog.String("execution_id", execution.ExecutionID))
		}
	}

	// Determinar status final e próxima execução
	var status string
	var errorMessage *string
	var nextExecution *time.Time

	if executionError != nil {
		status = domain.BackfillStatusFailed
		errMsg := executionError.Error()
		errorMessage = &errMsg
		s.logger.Error("❌ Execução de scheduler falhou",
			slog.String("execution_id", execution.ExecutionID),
			slog.String("error", errMsg))
	} else {
		status = domain.BackfillStatusSuccess
		nextExec := s.calculateNextExecution(config.Tipo)
		nextExecution = &nextExec
		s.logger.Info("✅ Execução de scheduler bem-sucedida",
			slog.String("execution_id", execution.ExecutionID),
			slog.String("tipo", execution.Tipo),
			slog.Int("total_sincronizados", totalSincronizados),
			slog.Time("proxima_execucao", nextExec))
	}

	// Marcar como concluída
	if err := s.schedulerRepo.CompleteExecution(ctx, execution.ExecutionID, status, errorMessage, nextExecution); err != nil {
		s.logger.Error("❌ Erro ao completar execução",
			slog.String("execution_id", execution.ExecutionID),
			slog.String("error", err.Error()))
	}
}

// sincronizarDeputados executa sincronização incremental de deputados
func (s *SmartSchedulerService) sincronizarDeputados(ctx context.Context, executionID string) (int, error) {
	s.logger.Info("👥 Sincronizando deputados", slog.String("execution_id", executionID))

	// Buscar deputados atuais da API
	deputados, _, err := s.deputadosService.ListarDeputados(ctx, "", "", "")
	if err != nil {
		return 0, fmt.Errorf("erro ao listar deputados: %w", err)
	}

	// Aqui seria implementada a lógica incremental real
	// Por ora, retorna contagem simulada
	count := len(deputados)
	s.logger.Info("✅ Deputados sincronizados",
		slog.String("execution_id", executionID),
		slog.Int("count", count))

	return count, nil
}

// sincronizarProposicoes executa sincronização incremental de proposições
func (s *SmartSchedulerService) sincronizarProposicoes(ctx context.Context, executionID string) (int, error) {
	s.logger.Info("📋 Sincronizando proposições", slog.String("execution_id", executionID))

	// Definir intervalo: dia anterior (00:00 -> 23:59:59)
	now := time.Now()
	y := now.AddDate(0, 0, -1)
	start := time.Date(y.Year(), y.Month(), y.Day(), 0, 0, 0, 0, now.Location())
	end := time.Date(y.Year(), y.Month(), y.Day(), 23, 59, 59, 0, now.Location())

	filtros := &domain.ProposicaoFilter{
		DataApresentacaoInicio: &start,
		DataApresentacaoFim:    &end,
		// A API da Câmara tem limite máximo de 100 itens por página
		// (validação em domain.ProposicaoFilter.Validate). Usar 100
		// evita erro de validação e permite paginação quando necessário.
		Limite: 100,
		Pagina: 1,
	}

	proposicoes, total, source, err := s.proposicoesService.ListarProposicoes(ctx, filtros)
	if err != nil {
		s.logger.Error("❌ Erro ao sincronizar proposições", slog.String("error", err.Error()))
		return 0, fmt.Errorf("erro ao sincronizar proposições: %w", err)
	}

	count := len(proposicoes)
	s.logger.Info("✅ Proposições sincronizadas",
		slog.String("execution_id", executionID),
		slog.Int("count", count),
		slog.Int("total", total),
		slog.String("source", source))

	return count, nil
}

// sincronizarDespesas executa sincronização incremental de despesas
func (s *SmartSchedulerService) sincronizarDespesas(ctx context.Context, executionID string) (int, error) {
	s.logger.Info("💰 Sincronizando despesas", slog.String("execution_id", executionID))

	// Buscar despesas do dia anterior para os deputados
	now := time.Now()
	y := now.AddDate(0, 0, -1)
	targetYear := y.Year()
	targetMonth := int(y.Month())
	targetDay := y.Day()

	deputies, _, err := s.deputadosService.ListarDeputados(ctx, "", "", "")
	if err != nil {
		s.logger.Error("❌ Erro ao listar deputados para sincronizar despesas", slog.String("error", err.Error()))
		return 0, fmt.Errorf("erro ao listar deputados: %w", err)
	}

	var totalFound int64

	// TODO: Otimizar para evitar problema N+1 queries
	// Atualmente fazemos uma query para cada deputado (N+1 pattern)
	// Considerar criar método ListarDespesasPorPeriodoBatch no repositório
	// para buscar despesas de múltiplos deputados de uma vez

	// worker pool to process deputies in parallel
	deputiesCh := make(chan domain.Deputado)
	var wg sync.WaitGroup

	workers := runtime.NumCPU()
	if workers < 2 {
		workers = 2
	}
	// Limitar workers para reduzir carga no banco durante sync diário
	if workers > 4 {
		workers = 4
	}

	worker := func() {
		defer wg.Done()
		for d := range deputiesCh {
			anoStr := fmt.Sprintf("%d", targetYear)
			despesas, _, err := s.deputadosService.ListarDespesas(ctx, fmt.Sprintf("%d", d.ID), anoStr)
			if err != nil {
				s.logger.Debug("Aviso: erro ao listar despesas para deputado", slog.Int("deputado_id", d.ID), slog.String("error", err.Error()))
				continue
			}

			localCount := 0
			for _, desp := range despesas {
				if desp.Ano != targetYear || desp.Mes != targetMonth {
					continue
				}
				dayMatch := false
				if desp.DataDocumento != "" {
					layouts := []string{"2006-01-02T15:04:05", "2006-01-02T15:04", "2006-01-02", "02/01/2006"}
					for _, l := range layouts {
						if t, err := time.Parse(l, desp.DataDocumento); err == nil {
							if t.Day() == targetDay && t.Month() == time.Month(targetMonth) && t.Year() == targetYear {
								dayMatch = true
								break
							}
						}
					}
				}
				if dayMatch || desp.DataDocumento == "" {
					localCount++
				}
			}

			if localCount > 0 {
				atomic.AddInt64(&totalFound, int64(localCount))
				// update progress per deputy batch
				s.schedulerRepo.UpdateExecutionProgress(ctx, executionID, map[string]interface{}{
					"despesas_sincronizadas": atomic.LoadInt64(&totalFound),
				})
			}
		}
	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker()
	}

	go func() {
		for _, d := range deputies {
			deputiesCh <- d
		}
		close(deputiesCh)
	}()

	wg.Wait()

	s.logger.Info("✅ Despesas sincronizadas",
		slog.String("execution_id", executionID),
		slog.Int64("count", atomic.LoadInt64(&totalFound)))

	return int(atomic.LoadInt64(&totalFound)), nil
}

// sincronizarVotacoes executa sincronização incremental de votações
func (s *SmartSchedulerService) sincronizarVotacoes(ctx context.Context, executionID string) (int, error) {
	s.logger.Info("🗳️ Sincronizando votações", slog.String("execution_id", executionID))

	now := time.Now()
	y := now.AddDate(0, 0, -1)
	start := time.Date(y.Year(), y.Month(), y.Day(), 0, 0, 0, 0, now.Location())
	end := time.Date(y.Year(), y.Month(), y.Day(), 23, 59, 59, 0, now.Location())

	filtros := map[string]interface{}{
		"dataInicio": start,
		"dataFim":    end,
		"limite":     500,
	}

	total, err := s.votacoesService.SincronizarVotacoesRecentes(ctx, filtros)
	if err != nil {
		s.logger.Error("❌ Erro ao sincronizar votações", slog.String("error", err.Error()))
		return 0, fmt.Errorf("erro ao sincronizar votações: %w", err)
	}

	s.logger.Info("✅ Votações sincronizadas",
		slog.String("execution_id", executionID),
		slog.Int("count", total))

	return total, nil
}

// calculateNextExecution calcula próxima execução baseada no tipo
func (s *SmartSchedulerService) calculateNextExecution(schedulerTipo string) time.Time {
	now := time.Now()

	switch schedulerTipo {
	case domain.SchedulerTipoDiario:
		// Próximo dia às 6h da manhã
		tomorrow := now.AddDate(0, 0, 1)
		return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 6, 0, 0, 0, tomorrow.Location())
	case domain.SchedulerTipoRapido:
		// Próxima execução em 4 horas
		return now.Add(4 * time.Hour)
	default:
		// Padrão: próxima execução em 1 hora
		return now.Add(time.Hour)
	}
}

// GetCurrentStatus retorna status atual de execuções de scheduler
func (s *SmartSchedulerService) GetCurrentStatus(ctx context.Context, schedulerTipo *string) (*domain.SchedulerStatus, error) {
	return s.schedulerRepo.GetCurrentStatus(ctx, schedulerTipo)
}

// ListExecutions lista execuções de scheduler com paginação
func (s *SmartSchedulerService) ListExecutions(ctx context.Context, limit, offset int, schedulerTipo *string) ([]domain.SchedulerExecution, int, error) {
	return s.schedulerRepo.ListExecutions(ctx, limit, offset, schedulerTipo)
}

// GetLastSuccessfulExecution retorna última execução bem-sucedida
func (s *SmartSchedulerService) GetLastSuccessfulExecution(ctx context.Context, schedulerTipo string) (*domain.SchedulerExecution, error) {
	return s.schedulerRepo.GetLastSuccessfulExecution(ctx, schedulerTipo)
}

// CleanupOldExecutions remove execuções antigas
func (s *SmartSchedulerService) CleanupOldExecutions(ctx context.Context) (int, error) {
	return s.schedulerRepo.CleanupOldExecutions(ctx)
}
