package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"to-de-olho-backend/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BackfillRepository gerencia operações de controle de backfill
type BackfillRepository struct {
	db *pgxpool.Pool
}

// NewBackfillRepository cria nova instância do repositório
func NewBackfillRepository(db *pgxpool.Pool) *BackfillRepository {
	return &BackfillRepository{db: db}
}

// HasSuccessfulHistoricalBackfill verifica se já foi feito backfill histórico com sucesso
func (r *BackfillRepository) HasSuccessfulHistoricalBackfill(ctx context.Context, startYear, endYear int) (bool, error) {
	query := `SELECT has_successful_historical_backfill($1, $2)`

	var hasBackfill bool
	err := r.db.QueryRow(ctx, query, startYear, endYear).Scan(&hasBackfill)
	if err != nil {
		return false, fmt.Errorf("erro ao verificar backfill histórico: %w", err)
	}

	return hasBackfill, nil
}

// GetLastExecution retorna última execução por tipo
func (r *BackfillRepository) GetLastExecution(ctx context.Context, executionType string) (*domain.BackfillExecution, error) {
	query := `
		SELECT 
			id, execution_id, tipo, ano_inicio, ano_fim, status,
			deputados_processados, proposicoes_processadas, 
			despesas_processadas, votacoes_processadas,
			started_at, completed_at, duration_seconds,
			triggered_by, error_message, 
			COALESCE(config::text, '{}') as config
		FROM backfill_executions 
		WHERE tipo = $1 
		ORDER BY started_at DESC 
		LIMIT 1
	`

	var exec domain.BackfillExecution
	var configStr string

	err := r.db.QueryRow(ctx, query, executionType).Scan(
		&exec.ID, &exec.ExecutionID, &exec.Tipo, &exec.AnoInicio, &exec.AnoFim, &exec.Status,
		&exec.DeputadosProcessados, &exec.ProposicoesProcessadas,
		&exec.DespesasProcessadas, &exec.VotacoesProcessadas,
		&exec.StartedAt, &exec.CompletedAt, &exec.DurationSeconds,
		&exec.TriggeredBy, &exec.ErrorMessage, &configStr,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, domain.ErrBackfillNaoEncontrado
		}
		return nil, fmt.Errorf("erro ao buscar última execução: %w", err)
	}

	// Parse do config JSON
	if configStr != "" && configStr != "{}" {
		if err := json.Unmarshal([]byte(configStr), &exec.Config); err != nil {
			// Log do erro mas não falha a consulta
			exec.Config = make(map[string]interface{})
		}
	} else {
		exec.Config = make(map[string]interface{})
	}

	return &exec, nil
}

// CreateExecution cria nova execução de backfill
func (r *BackfillRepository) CreateExecution(ctx context.Context, config *domain.BackfillConfig) (*domain.BackfillExecution, error) {
	// Serializar config para JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar config: %w", err)
	}

	query := `
		INSERT INTO backfill_executions (
			tipo, ano_inicio, ano_fim, status, triggered_by, config
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, execution_id, started_at
	`

	var exec domain.BackfillExecution
	err = r.db.QueryRow(ctx, query,
		config.Tipo, config.AnoInicio, config.AnoFim,
		domain.BackfillStatusRunning, config.TriggeredBy, configJSON,
	).Scan(&exec.ID, &exec.ExecutionID, &exec.StartedAt)

	if err != nil {
		return nil, fmt.Errorf("erro ao criar execução: %w", err)
	}

	// Preencher dados da execução
	exec.Tipo = config.Tipo
	exec.AnoInicio = config.AnoInicio
	exec.AnoFim = config.AnoFim
	exec.Status = domain.BackfillStatusRunning
	exec.TriggeredBy = config.TriggeredBy
	exec.Config = make(map[string]interface{})
	json.Unmarshal(configJSON, &exec.Config)

	return &exec, nil
}

// UpdateExecutionProgress atualiza progresso da execução
func (r *BackfillRepository) UpdateExecutionProgress(ctx context.Context, executionID string, update domain.BackfillStatus) error {
	query := `
		UPDATE backfill_executions 
		SET 
			deputados_processados = $2,
			proposicoes_processadas = $3,
			despesas_processadas = $4,
			votacoes_processadas = $5,
			current_operation = $6
		WHERE execution_id = $1
	`

	_, err := r.db.Exec(ctx, query, executionID,
		update.DeputadosProcessados, update.ProposicoesProcessadas,
		update.DespesasProcessadas, update.VotacoesProcessadas,
		update.CurrentOperation,
	)

	return err
}

// CompleteExecution marca execução como concluída
func (r *BackfillRepository) CompleteExecution(ctx context.Context, executionID string, status string, errorMessage *string) error {
	now := time.Now()

	// Buscar started_at para calcular duração
	var startedAt time.Time
	err := r.db.QueryRow(ctx, "SELECT started_at FROM backfill_executions WHERE execution_id = $1", executionID).Scan(&startedAt)
	if err != nil {
		return fmt.Errorf("erro ao buscar started_at: %w", err)
	}

	duration := int(now.Sub(startedAt).Seconds())

	query := `
		UPDATE backfill_executions 
		SET 
			status = $2,
			completed_at = $3,
			duration_seconds = $4,
			error_message = $5
		WHERE execution_id = $1
	`

	_, err = r.db.Exec(ctx, query, executionID, status, now, duration, errorMessage)
	return err
}

// GetRunningExecution verifica se há execução em andamento
func (r *BackfillRepository) GetRunningExecution(ctx context.Context) (*domain.BackfillExecution, error) {
	query := `
		SELECT 
			id, execution_id, tipo, ano_inicio, ano_fim, status,
			deputados_processados, proposicoes_processadas,
			despesas_processadas, votacoes_processadas,
			started_at, triggered_by
		FROM backfill_executions 
		WHERE status = $1
		ORDER BY started_at DESC 
		LIMIT 1
	`

	var exec domain.BackfillExecution

	err := r.db.QueryRow(ctx, query, domain.BackfillStatusRunning).Scan(
		&exec.ID, &exec.ExecutionID, &exec.Tipo, &exec.AnoInicio, &exec.AnoFim, &exec.Status,
		&exec.DeputadosProcessados, &exec.ProposicoesProcessadas,
		&exec.DespesasProcessadas, &exec.VotacoesProcessadas,
		&exec.StartedAt, &exec.TriggeredBy,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, domain.ErrBackfillNaoEncontrado
		}
		return nil, fmt.Errorf("erro ao buscar execução em andamento: %w", err)
	}

	return &exec, nil
}

// ListExecutions lista execuções por período
func (r *BackfillRepository) ListExecutions(ctx context.Context, limit int, offset int) ([]domain.BackfillExecution, int, error) {
	// Query para buscar execuções
	query := `
		SELECT 
			id, execution_id, tipo, ano_inicio, ano_fim, status,
			deputados_processados, proposicoes_processadas,
			despesas_processadas, votacoes_processadas,
			started_at, completed_at, duration_seconds,
			triggered_by, error_message
		FROM backfill_executions 
		ORDER BY started_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao listar execuções: %w", err)
	}
	defer rows.Close()

	var executions []domain.BackfillExecution

	for rows.Next() {
		var exec domain.BackfillExecution
		err := rows.Scan(
			&exec.ID, &exec.ExecutionID, &exec.Tipo, &exec.AnoInicio, &exec.AnoFim, &exec.Status,
			&exec.DeputadosProcessados, &exec.ProposicoesProcessadas,
			&exec.DespesasProcessadas, &exec.VotacoesProcessadas,
			&exec.StartedAt, &exec.CompletedAt, &exec.DurationSeconds,
			&exec.TriggeredBy, &exec.ErrorMessage,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("erro ao escanear execução: %w", err)
		}

		executions = append(executions, exec)
	}

	// Query para contar total
	var total int
	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM backfill_executions").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao contar execuções: %w", err)
	}

	return executions, total, nil
}

// DeleteOldExecutions remove execuções antigas (manter apenas as últimas N por tipo)
func (r *BackfillRepository) DeleteOldExecutions(ctx context.Context, keepLast int) error {
	query := `
		WITH ranked_executions AS (
			SELECT id, 
				   ROW_NUMBER() OVER (PARTITION BY tipo ORDER BY started_at DESC) as rn
			FROM backfill_executions
		)
		DELETE FROM backfill_executions 
		WHERE id IN (
			SELECT id FROM ranked_executions WHERE rn > $1
		)
	`

	_, err := r.db.Exec(ctx, query, keepLast)
	return err
}
