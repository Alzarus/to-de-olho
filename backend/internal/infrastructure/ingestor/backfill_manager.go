package ingestor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB abstracts the subset of pgxpool.Pool used, enabling mocking in unit tests.
type DB interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

// BackfillCheckpoint representa um checkpoint do processo de backfill
type BackfillCheckpoint struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`   // "deputados", "proposicoes", "despesas"
	Status       string                 `json:"status"` // "pending", "in_progress", "completed", "failed"
	Progress     BackfillProgress       `json:"progress"`
	StartedAt    *time.Time             `json:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// BackfillProgress rastreia o progresso do backfill
type BackfillProgress struct {
	TotalItems      int    `json:"total_items"`
	ProcessedItems  int    `json:"processed_items"`
	FailedItems     int    `json:"failed_items"`
	LastProcessedID string `json:"last_processed_id,omitempty"`
}

// BackfillManager gerencia o processo de backfill histórico
type BackfillManager struct {
	db DB
}

// NewBackfillManager cria um novo gerenciador de backfill
func NewBackfillManager(db *pgxpool.Pool) *BackfillManager {
	return &BackfillManager{db: db}
}

// NewBackfillManagerWithDB cria um novo gerenciador de backfill com interface DB (para testes)
func NewBackfillManagerWithDB(db DB) *BackfillManager {
	return &BackfillManager{db: db}
}

// CreateCheckpoint cria um novo checkpoint
func (bm *BackfillManager) CreateCheckpoint(ctx context.Context, checkpointType string, metadata map[string]interface{}) (*BackfillCheckpoint, error) {
	checkpoint := &BackfillCheckpoint{
		ID:        fmt.Sprintf("%s_%d", checkpointType, time.Now().Unix()),
		Type:      checkpointType,
		Status:    "pending",
		Progress:  BackfillProgress{},
		Metadata:  metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Serializar dados para salvar no banco
	metadataJSON, err := json.Marshal(checkpoint.Metadata)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar metadata: %w", err)
	}

	progressJSON, err := json.Marshal(checkpoint.Progress)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar progress: %w", err)
	}

	query := `
		INSERT INTO backfill_checkpoints 
		(id, type, status, progress, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = bm.db.Exec(ctx, query,
		checkpoint.ID,
		checkpoint.Type,
		checkpoint.Status,
		progressJSON,
		metadataJSON,
		checkpoint.CreatedAt,
		checkpoint.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("erro ao criar checkpoint: %w", err)
	}

	return checkpoint, nil
}

// UpdateCheckpoint atualiza um checkpoint existente
func (bm *BackfillManager) UpdateCheckpoint(ctx context.Context, checkpoint *BackfillCheckpoint) error {
	checkpoint.UpdatedAt = time.Now()

	progressJSON, err := json.Marshal(checkpoint.Progress)
	if err != nil {
		return fmt.Errorf("erro ao serializar progress: %w", err)
	}

	metadataJSON, err := json.Marshal(checkpoint.Metadata)
	if err != nil {
		return fmt.Errorf("erro ao serializar metadata: %w", err)
	}

	query := `
		UPDATE backfill_checkpoints 
		SET status = $2, progress = $3, metadata = $4, 
		    started_at = $5, completed_at = $6, error_message = $7, updated_at = $8
		WHERE id = $1
	`

	_, err = bm.db.Exec(ctx, query,
		checkpoint.ID,
		checkpoint.Status,
		progressJSON,
		metadataJSON,
		checkpoint.StartedAt,
		checkpoint.CompletedAt,
		checkpoint.ErrorMessage,
		checkpoint.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("erro ao atualizar checkpoint: %w", err)
	}

	return nil
}

// GetCheckpoint recupera um checkpoint por ID
func (bm *BackfillManager) GetCheckpoint(ctx context.Context, id string) (*BackfillCheckpoint, error) {
	query := `
		SELECT id, type, status, progress, metadata, started_at, completed_at, 
		       error_message, created_at, updated_at
		FROM backfill_checkpoints 
		WHERE id = $1
	`

	var checkpoint BackfillCheckpoint
	var progressJSON, metadataJSON []byte
	var errorMessage *string

	err := bm.db.QueryRow(ctx, query, id).Scan(
		&checkpoint.ID,
		&checkpoint.Type,
		&checkpoint.Status,
		&progressJSON,
		&metadataJSON,
		&checkpoint.StartedAt,
		&checkpoint.CompletedAt,
		&errorMessage,
		&checkpoint.CreatedAt,
		&checkpoint.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("erro ao buscar checkpoint: %w", err)
	}

	// Tratar campos nullable
	if errorMessage != nil {
		checkpoint.ErrorMessage = *errorMessage
	}

	if err := json.Unmarshal(progressJSON, &checkpoint.Progress); err != nil {
		return nil, fmt.Errorf("erro ao deserializar progress: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &checkpoint.Metadata); err != nil {
		return nil, fmt.Errorf("erro ao deserializar metadata: %w", err)
	}

	return &checkpoint, nil
}

// GetPendingCheckpoints retorna checkpoints pendentes ou em progresso
func (bm *BackfillManager) GetPendingCheckpoints(ctx context.Context) ([]*BackfillCheckpoint, error) {
	query := `
		SELECT id, type, status, progress, metadata, started_at, completed_at,
		       error_message, created_at, updated_at
		FROM backfill_checkpoints 
		WHERE status IN ('pending', 'in_progress')
		ORDER BY created_at ASC
	`

	rows, err := bm.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar checkpoints pendentes: %w", err)
	}
	defer rows.Close()

	var checkpoints []*BackfillCheckpoint

	for rows.Next() {
		var checkpoint BackfillCheckpoint
		var progressJSON, metadataJSON []byte
		var errorMessage *string

		err := rows.Scan(
			&checkpoint.ID,
			&checkpoint.Type,
			&checkpoint.Status,
			&progressJSON,
			&metadataJSON,
			&checkpoint.StartedAt,
			&checkpoint.CompletedAt,
			&errorMessage,
			&checkpoint.CreatedAt,
			&checkpoint.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("erro ao escanear checkpoint: %w", err)
		}

		// Tratar campos nullable
		if errorMessage != nil {
			checkpoint.ErrorMessage = *errorMessage
		}

		if err := json.Unmarshal(progressJSON, &checkpoint.Progress); err != nil {
			log.Printf("Erro ao deserializar progress para checkpoint %s: %v", checkpoint.ID, err)
			continue
		}

		if err := json.Unmarshal(metadataJSON, &checkpoint.Metadata); err != nil {
			log.Printf("Erro ao deserializar metadata para checkpoint %s: %v", checkpoint.ID, err)
			continue
		}

		checkpoints = append(checkpoints, &checkpoint)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar checkpoints: %w", err)
	}

	return checkpoints, nil
}

// MarkAsStarted marca um checkpoint como iniciado
func (bm *BackfillManager) MarkAsStarted(ctx context.Context, checkpoint *BackfillCheckpoint) error {
	now := time.Now()
	checkpoint.Status = "in_progress"
	checkpoint.StartedAt = &now
	return bm.UpdateCheckpoint(ctx, checkpoint)
}

// MarkAsCompleted marca um checkpoint como completado
func (bm *BackfillManager) MarkAsCompleted(ctx context.Context, checkpoint *BackfillCheckpoint) error {
	now := time.Now()
	checkpoint.Status = "completed"
	checkpoint.CompletedAt = &now
	return bm.UpdateCheckpoint(ctx, checkpoint)
}

// MarkAsFailed marca um checkpoint como falhado
func (bm *BackfillManager) MarkAsFailed(ctx context.Context, checkpoint *BackfillCheckpoint, errorMsg string) error {
	checkpoint.Status = "failed"
	checkpoint.ErrorMessage = errorMsg
	return bm.UpdateCheckpoint(ctx, checkpoint)
}

// UpdateProgress atualiza o progresso de um checkpoint
func (bm *BackfillManager) UpdateProgress(ctx context.Context, checkpoint *BackfillCheckpoint, processed, failed int, lastID string) error {
	checkpoint.Progress.ProcessedItems = processed
	checkpoint.Progress.FailedItems = failed
	if lastID != "" {
		checkpoint.Progress.LastProcessedID = lastID
	}
	return bm.UpdateCheckpoint(ctx, checkpoint)
}

// GetBackfillStats retorna estatísticas do backfill
func (bm *BackfillManager) GetBackfillStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT 
			type,
			status,
			COUNT(*) as count,
			SUM(CASE WHEN progress->>'processed_items' != '' 
			    THEN (progress->>'processed_items')::int ELSE 0 END) as total_processed,
			SUM(CASE WHEN progress->>'failed_items' != '' 
			    THEN (progress->>'failed_items')::int ELSE 0 END) as total_failed
		FROM backfill_checkpoints 
		GROUP BY type, status
		ORDER BY type, status
	`

	rows, err := bm.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar estatísticas: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]interface{})

	for rows.Next() {
		var checkpointType, status string
		var count, totalProcessed, totalFailed int

		err := rows.Scan(&checkpointType, &status, &count, &totalProcessed, &totalFailed)
		if err != nil {
			return nil, fmt.Errorf("erro ao escanear estatísticas: %w", err)
		}

		if stats[checkpointType] == nil {
			stats[checkpointType] = make(map[string]interface{})
		}

		typeStats := stats[checkpointType].(map[string]interface{})
		typeStats[status] = map[string]interface{}{
			"count":           count,
			"total_processed": totalProcessed,
			"total_failed":    totalFailed,
		}
	}

	return stats, nil
}

// BackfillStrategy define estratégias de backfill
type BackfillStrategy struct {
	YearStart  int           // Ano inicial (ex: 2019)
	YearEnd    int           // Ano final (ex: 2025)
	BatchSize  int           // Tamanho do lote
	MaxRetries int           // Máximo de tentativas por lote
	RetryDelay time.Duration // Delay entre tentativas
}

// DefaultBackfillStrategy retorna estratégia padrão seguindo as melhores práticas
func DefaultBackfillStrategy() BackfillStrategy {
	currentYear := time.Now().Year()
	return BackfillStrategy{
		YearStart:  2019, // Dados históricos desde 2019
		YearEnd:    currentYear,
		BatchSize:  100,             // Lotes de 100 itens
		MaxRetries: 3,               // 3 tentativas por lote
		RetryDelay: 5 * time.Second, // 5s entre tentativas
	}
}
