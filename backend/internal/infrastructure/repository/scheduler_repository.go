package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"to-de-olho-backend/internal/domain"
	"to-de-olho-backend/internal/pkg/metrics"
)

// generateExecutionID gera um UUID único para execução
func generateExecutionID() string {
	return uuid.New().String()
}

// SchedulerRepository gerencia execuções de scheduler no banco
type SchedulerRepository struct {
	db *pgxpool.Pool
}

// Note: use domain.ErrSchedulerAlreadyRunning for consistency across packages

// computeLockKey calcula uma chave int64 consistente para advisory lock baseada no tipo
func computeLockKey(schedulerTipo string) int64 {
	h := fnv.New64a()
	h.Write([]byte("scheduler:"))
	h.Write([]byte(schedulerTipo))
	return int64(h.Sum64())
}

// NewSchedulerRepository cria uma nova instância do repository
func NewSchedulerRepository(db *pgxpool.Pool) *SchedulerRepository {
	return &SchedulerRepository{db: db}
}

// CreateExecution inicia uma nova execução de scheduler
func (r *SchedulerRepository) CreateExecution(ctx context.Context, config *domain.SchedulerConfig) (*domain.SchedulerExecution, error) {
	// Tentar adquirir um advisory lock baseado no tipo do scheduler para prevenir concorrência entre processos
	lockKey := computeLockKey(config.Tipo)
	var gotLock bool
	err := r.db.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", lockKey).Scan(&gotLock)
	if err != nil {
		return nil, fmt.Errorf("erro ao tentar adquirir advisory lock: %w", err)
	}
	if !gotLock {
		// Increment in-memory metric for observability
		metrics.IncSchedulerSkip(config.Tipo)
		return nil, domain.ErrSchedulerAlreadyRunning
	}

	execution := &domain.SchedulerExecution{
		ExecutionID: generateExecutionID(),
		Tipo:        config.Tipo,
		Status:      domain.BackfillStatusRunning,
		StartedAt:   time.Now(),
		TriggeredBy: config.TriggeredBy,
		Config:      config.Config,
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar config: %w", err)
	}

	query := `
		INSERT INTO scheduler_executions (
			execution_id, tipo, status, started_at, triggered_by, config
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRow(ctx, query,
		execution.ExecutionID,
		execution.Tipo,
		execution.Status,
		execution.StartedAt,
		execution.TriggeredBy,
		configJSON,
	).Scan(&execution.ID, &execution.CreatedAt, &execution.UpdatedAt)

	if err != nil {
		// Se falhou ao inserir, liberar o advisory lock para evitar deadlock
		_, _ = r.db.Exec(ctx, "SELECT pg_advisory_unlock($1)", lockKey)
		return nil, fmt.Errorf("erro ao criar execução de scheduler: %w", err)
	}

	return execution, nil
}

// UpdateExecutionProgress atualiza o progresso de uma execução
func (r *SchedulerRepository) UpdateExecutionProgress(ctx context.Context, executionID string, update map[string]interface{}) error {
	query := `
		UPDATE scheduler_executions
		SET deputados_sincronizados = COALESCE($2, deputados_sincronizados),
		    proposicoes_sincronizadas = COALESCE($3, proposicoes_sincronizadas),
		    despesas_sincronizadas = COALESCE($4, despesas_sincronizadas),
		    votacoes_sincronizadas = COALESCE($5, votacoes_sincronizadas),
		    updated_at = NOW()
		WHERE execution_id = $1
	`

	var deputados, proposicoes, despesas, votacoes *int
	if val, ok := update["deputados_sincronizados"]; ok {
		if v, ok := val.(int); ok {
			deputados = &v
		}
	}
	if val, ok := update["proposicoes_sincronizadas"]; ok {
		if v, ok := val.(int); ok {
			proposicoes = &v
		}
	}
	if val, ok := update["despesas_sincronizadas"]; ok {
		if v, ok := val.(int); ok {
			despesas = &v
		}
	}
	if val, ok := update["votacoes_sincronizadas"]; ok {
		if v, ok := val.(int); ok {
			votacoes = &v
		}
	}

	_, err := r.db.Exec(ctx, query, executionID, deputados, proposicoes, despesas, votacoes)
	return err
}

// CompleteExecution marca uma execução como concluída
func (r *SchedulerRepository) CompleteExecution(ctx context.Context, executionID string, status string, errorMessage *string, nextExecution *time.Time) error {
	completedAt := time.Now()

	query := `
		UPDATE scheduler_executions
		SET status = $2,
		    completed_at = $3,
		    error_message = $4,
		    next_execution = $5,
		    updated_at = NOW()
		WHERE execution_id = $1
	`

	_, err := r.db.Exec(ctx, query, executionID, status, completedAt, errorMessage, nextExecution)
	if err != nil {
		return err
	}

	// Após marcar concluída, tentar liberar o advisory lock baseado no tipo desta execução
	var tipo string
	row := r.db.QueryRow(ctx, "SELECT tipo FROM scheduler_executions WHERE execution_id = $1", executionID)
	if err := row.Scan(&tipo); err != nil {
		// Se não conseguir obter o tipo, apenas retornar sem falha (unlock é tentativa de conveniência)
		return nil
	}

	lockKey := computeLockKey(tipo)
	_, _ = r.db.Exec(ctx, "SELECT pg_advisory_unlock($1)", lockKey)

	return nil
}

// ShouldSchedulerExecute verifica se um scheduler deve executar baseado no intervalo mínimo
func (r *SchedulerRepository) ShouldSchedulerExecute(ctx context.Context, schedulerTipo string, minIntervalHours int) (*domain.ShouldExecuteResult, error) {
	query := `SELECT should_run, reason, last_execution, hours_since_last FROM should_scheduler_execute($1, $2)`

	var result domain.ShouldExecuteResult
	var lastExecution sql.NullTime
	var hoursSinceLast sql.NullFloat64

	err := r.db.QueryRow(ctx, query, schedulerTipo, minIntervalHours).Scan(
		&result.ShouldRun,
		&result.Reason,
		&lastExecution,
		&hoursSinceLast,
	)

	if err != nil {
		return nil, fmt.Errorf("erro ao verificar se scheduler deve executar: %w", err)
	}

	if lastExecution.Valid {
		result.LastExecution = &lastExecution.Time
	}
	if hoursSinceLast.Valid {
		result.HoursSinceLast = &hoursSinceLast.Float64
	}

	return &result, nil
}

// GetCurrentStatus retorna o status da execução atual (running) ou da última concluída
func (r *SchedulerRepository) GetCurrentStatus(ctx context.Context, schedulerTipo *string) (*domain.SchedulerStatus, error) {
	// Primeiro, tentar encontrar execução em andamento
	query := `
		SELECT execution_id, tipo, status, started_at, next_execution,
		       deputados_sincronizados, proposicoes_sincronizadas, 
		       despesas_sincronizadas, votacoes_sincronizadas
		FROM scheduler_executions
		WHERE status = 'running' AND ($1 IS NULL OR tipo = $1)
		ORDER BY started_at DESC
		LIMIT 1
	`

	var status domain.SchedulerStatus
	var nextExec sql.NullTime

	err := r.db.QueryRow(ctx, query, schedulerTipo).Scan(
		&status.ExecutionID,
		&status.Tipo,
		&status.Status,
		&status.StartedAt,
		&nextExec,
		&status.DeputadosSincronizados,
		&status.ProposicoesSincronizadas,
		&status.DespesasSincronizadas,
		&status.VotacoesSincronizadas,
	)

	if err == nil {
		// Execução em andamento encontrada
		status.CurrentOperation = "Sincronização em andamento"
		status.LastUpdate = time.Now()
		if nextExec.Valid {
			status.NextExecution = &nextExec.Time
		}
		return &status, nil
	}

	if err != pgx.ErrNoRows {
		return nil, fmt.Errorf("erro ao buscar status atual: %w", err)
	}

	// Não há execução em andamento, buscar a última concluída
	query = `
		SELECT execution_id, tipo, status, started_at, next_execution,
		       deputados_sincronizados, proposicoes_sincronizadas,
		       despesas_sincronizadas, votacoes_sincronizadas, completed_at
		FROM scheduler_executions
		WHERE status IN ('success', 'failed', 'partial') AND ($1 IS NULL OR tipo = $1)
		ORDER BY completed_at DESC
		LIMIT 1
	`

	var completedAt sql.NullTime
	err = r.db.QueryRow(ctx, query, schedulerTipo).Scan(
		&status.ExecutionID,
		&status.Tipo,
		&status.Status,
		&status.StartedAt,
		&nextExec,
		&status.DeputadosSincronizados,
		&status.ProposicoesSincronizadas,
		&status.DespesasSincronizadas,
		&status.VotacoesSincronizadas,
		&completedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil // Nenhuma execução encontrada
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar última execução: %w", err)
	}

	status.CurrentOperation = "Aguardando próxima execução"
	if completedAt.Valid {
		status.LastUpdate = completedAt.Time
	}
	if nextExec.Valid {
		status.NextExecution = &nextExec.Time
	}

	return &status, nil
}

// ListExecutions lista execuções com paginação
func (r *SchedulerRepository) ListExecutions(ctx context.Context, limit, offset int, schedulerTipo *string) ([]domain.SchedulerExecution, int, error) {
	// Count total
	countQuery := `SELECT COUNT(*) FROM scheduler_executions WHERE ($1 IS NULL OR tipo = $1)`
	var total int
	err := r.db.QueryRow(ctx, countQuery, schedulerTipo).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao contar execuções: %w", err)
	}

	// Lista com paginação
	query := `
		SELECT id, execution_id, tipo, status, 
		       deputados_sincronizados, proposicoes_sincronizadas,
		       despesas_sincronizadas, votacoes_sincronizadas,
		       started_at, completed_at, duration_seconds, next_execution,
		       triggered_by, error_message, created_at, updated_at
		FROM scheduler_executions
		WHERE ($1 IS NULL OR tipo = $1)
		ORDER BY started_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, schedulerTipo, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao listar execuções: %w", err)
	}
	defer rows.Close()

	var executions []domain.SchedulerExecution
	for rows.Next() {
		var exec domain.SchedulerExecution
		var completedAt, nextExecution sql.NullTime
		var durationSeconds sql.NullInt32
		var errorMessage sql.NullString

		err := rows.Scan(
			&exec.ID, &exec.ExecutionID, &exec.Tipo, &exec.Status,
			&exec.DeputadosSincronizados, &exec.ProposicoesSincronizadas,
			&exec.DespesasSincronizadas, &exec.VotacoesSincronizadas,
			&exec.StartedAt, &completedAt, &durationSeconds, &nextExecution,
			&exec.TriggeredBy, &errorMessage, &exec.CreatedAt, &exec.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("erro ao scanear execução: %w", err)
		}

		if completedAt.Valid {
			exec.CompletedAt = &completedAt.Time
		}
		if nextExecution.Valid {
			exec.NextExecution = &nextExecution.Time
		}
		if durationSeconds.Valid {
			duration := int(durationSeconds.Int32)
			exec.DurationSeconds = &duration
		}
		if errorMessage.Valid {
			exec.ErrorMessage = &errorMessage.String
		}

		executions = append(executions, exec)
	}

	return executions, total, nil
}

// GetLastSuccessfulExecution retorna a última execução bem-sucedida de um tipo
func (r *SchedulerRepository) GetLastSuccessfulExecution(ctx context.Context, schedulerTipo string) (*domain.SchedulerExecution, error) {
	query := `
		SELECT id, execution_id, tipo, status, 
		       deputados_sincronizados, proposicoes_sincronizadas,
		       despesas_sincronizadas, votacoes_sincronizadas,
		       started_at, completed_at, duration_seconds, next_execution,
		       triggered_by, error_message, created_at, updated_at
		FROM scheduler_executions
		WHERE tipo = $1 AND status = 'success'
		ORDER BY completed_at DESC
		LIMIT 1
	`

	var exec domain.SchedulerExecution
	var completedAt, nextExecution sql.NullTime
	var durationSeconds sql.NullInt32
	var errorMessage sql.NullString

	err := r.db.QueryRow(ctx, query, schedulerTipo).Scan(
		&exec.ID, &exec.ExecutionID, &exec.Tipo, &exec.Status,
		&exec.DeputadosSincronizados, &exec.ProposicoesSincronizadas,
		&exec.DespesasSincronizadas, &exec.VotacoesSincronizadas,
		&exec.StartedAt, &completedAt, &durationSeconds, &nextExecution,
		&exec.TriggeredBy, &errorMessage, &exec.CreatedAt, &exec.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar última execução bem-sucedida: %w", err)
	}

	if completedAt.Valid {
		exec.CompletedAt = &completedAt.Time
	}
	if nextExecution.Valid {
		exec.NextExecution = &nextExecution.Time
	}
	if durationSeconds.Valid {
		duration := int(durationSeconds.Int32)
		exec.DurationSeconds = &duration
	}
	if errorMessage.Valid {
		exec.ErrorMessage = &errorMessage.String
	}

	return &exec, nil
}

// CleanupOldExecutions remove execuções antigas (mais de 30 dias)
func (r *SchedulerRepository) CleanupOldExecutions(ctx context.Context) (int, error) {
	query := `SELECT cleanup_old_scheduler_executions()`

	var deletedCount int
	err := r.db.QueryRow(ctx, query).Scan(&deletedCount)
	if err != nil {
		return 0, fmt.Errorf("erro ao limpar execuções antigas: %w", err)
	}

	return deletedCount, nil
}
