package background

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"
)

// JobType define os tipos de jobs em background
type JobType string

const (
	JobTypeCacheWarm JobType = "cache_warm"
	JobTypeDataSync  JobType = "data_sync"
	JobTypeAnalytics JobType = "analytics"
	JobTypeCleanup   JobType = "cleanup"
)

// Job representa um trabalho em background
type Job struct {
	ID         string                 `json:"id"`
	Type       JobType                `json:"type"`
	Payload    map[string]interface{} `json:"payload"`
	Priority   int                    `json:"priority"` // 1 = alta, 5 = baixa
	CreatedAt  time.Time              `json:"created_at"`
	StartedAt  *time.Time             `json:"started_at,omitempty"`
	DoneAt     *time.Time             `json:"done_at,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Retries    int                    `json:"retries"`
	MaxRetries int                    `json:"max_retries"`
}

// JobHandler define como processar cada tipo de job
type JobHandler interface {
	Handle(ctx context.Context, job *Job) error
}

// BackgroundProcessor gerencia jobs em background
type BackgroundProcessor struct {
	jobs     chan *Job
	handlers map[JobType]JobHandler
	workers  int
	logger   *slog.Logger
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc

	// Métricas
	processed int64
	failed    int64
	mutex     sync.RWMutex
}

// NewBackgroundProcessor cria novo processador de jobs
func NewBackgroundProcessor(workers int, queueSize int, logger *slog.Logger) *BackgroundProcessor {
	if workers <= 0 {
		workers = 5 // default
	}
	if queueSize <= 0 {
		queueSize = 1000 // default
	}
	if logger == nil {
		logger = slog.Default()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &BackgroundProcessor{
		jobs:     make(chan *Job, queueSize),
		handlers: make(map[JobType]JobHandler),
		workers:  workers,
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// RegisterHandler registra handler para um tipo de job
func (bp *BackgroundProcessor) RegisterHandler(jobType JobType, handler JobHandler) {
	bp.handlers[jobType] = handler
	bp.logger.Info("job handler registered", slog.String("type", string(jobType)))
}

// Start inicia os workers
func (bp *BackgroundProcessor) Start() {
	bp.logger.Info("starting background processor", slog.Int("workers", bp.workers))

	for i := 0; i < bp.workers; i++ {
		bp.wg.Add(1)
		go bp.worker(i)
	}

	// Worker de métricas
	bp.wg.Add(1)
	go bp.metricsWorker()
}

// Stop para o processador graciosamente
func (bp *BackgroundProcessor) Stop() {
	bp.logger.Info("stopping background processor")
	bp.cancel()
	close(bp.jobs)
	bp.wg.Wait()
	bp.logger.Info("background processor stopped")
}

// SubmitJob envia job para processamento
func (bp *BackgroundProcessor) SubmitJob(job *Job) error {
	if job.ID == "" {
		job.ID = generateJobID()
	}
	job.CreatedAt = time.Now()

	if job.MaxRetries == 0 {
		job.MaxRetries = 3 // default
	}

	select {
	case bp.jobs <- job:
		bp.logger.Debug("job submitted",
			slog.String("id", job.ID),
			slog.String("type", string(job.Type)))
		return nil
	case <-bp.ctx.Done():
		return fmt.Errorf("processor is shutting down")
	default:
		return fmt.Errorf("job queue is full")
	}
}

// SubmitJobAsync envia job de forma assíncrona (non-blocking)
func (bp *BackgroundProcessor) SubmitJobAsync(job *Job) {
	go func() {
		if err := bp.SubmitJob(job); err != nil {
			bp.logger.Error("failed to submit async job",
				slog.String("id", job.ID),
				slog.String("error", err.Error()))
		}
	}()
}

// worker processa jobs continuamente
func (bp *BackgroundProcessor) worker(id int) {
	defer bp.wg.Done()

	bp.logger.Debug("worker started", slog.Int("worker_id", id))

	for {
		select {
		case <-bp.ctx.Done():
			bp.logger.Debug("worker stopping", slog.Int("worker_id", id))
			return
		case job, ok := <-bp.jobs:
			if !ok {
				bp.logger.Debug("job channel closed", slog.Int("worker_id", id))
				return
			}

			bp.processJob(id, job)
		}
	}
}

// processJob processa um job individual
func (bp *BackgroundProcessor) processJob(workerID int, job *Job) {
	start := time.Now()
	now := time.Now()
	job.StartedAt = &now

	bp.logger.Debug("processing job",
		slog.Int("worker_id", workerID),
		slog.String("job_id", job.ID),
		slog.String("job_type", string(job.Type)))

	handler, exists := bp.handlers[job.Type]
	if !exists {
		bp.logger.Error("no handler for job type",
			slog.String("job_id", job.ID),
			slog.String("job_type", string(job.Type)))
		bp.incrementFailed()
		return
	}

	// Timeout por job (5 minutos por default)
	jobCtx, cancel := context.WithTimeout(bp.ctx, 5*time.Minute)
	defer cancel()

	err := handler.Handle(jobCtx, job)
	done := time.Now()
	job.DoneAt = &done

	if err != nil {
		job.Error = err.Error()
		job.Retries++

		bp.logger.Error("job failed",
			slog.String("job_id", job.ID),
			slog.String("job_type", string(job.Type)),
			slog.String("error", err.Error()),
			slog.Int("retries", job.Retries),
			slog.Duration("duration", time.Since(start)))

		// Retry se ainda temos tentativas
		if job.Retries < job.MaxRetries {
			// Backoff exponencial: base 2^retry com jitter
			baseDelay := time.Duration(1<<job.Retries) * time.Second

			// Adicionar jitter (±25%) para evitar thundering herd
			jitter := time.Duration(rand.Float64()*0.5-0.25) * baseDelay
			delay := baseDelay + jitter

			// Cap máximo de 5 minutos
			maxDelay := 5 * time.Minute
			if delay > maxDelay {
				delay = maxDelay
			}

			bp.logger.Info("scheduling job retry",
				slog.String("job_id", job.ID),
				slog.Duration("delay", delay),
				slog.Duration("base_delay", baseDelay),
				slog.Duration("jitter", jitter))

			go func() {
				time.Sleep(delay)
				job.StartedAt = nil
				job.DoneAt = nil
				bp.SubmitJobAsync(job)
			}()
		} else {
			bp.logger.Error("job max retries exceeded",
				slog.String("job_id", job.ID),
				slog.String("job_type", string(job.Type)))
		}

		bp.incrementFailed()
	} else {
		bp.logger.Info("job completed successfully",
			slog.String("job_id", job.ID),
			slog.String("job_type", string(job.Type)),
			slog.Duration("duration", time.Since(start)))

		bp.incrementProcessed()
	}
}

// metricsWorker reporta métricas periodicamente
func (bp *BackgroundProcessor) metricsWorker() {
	defer bp.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-bp.ctx.Done():
			return
		case <-ticker.C:
			bp.reportMetrics()
		}
	}
}

// reportMetrics reporta métricas do processador
func (bp *BackgroundProcessor) reportMetrics() {
	bp.mutex.RLock()
	processed := bp.processed
	failed := bp.failed
	bp.mutex.RUnlock()

	queueLen := len(bp.jobs)

	bp.logger.Info("background processor metrics",
		slog.Int64("jobs_processed", processed),
		slog.Int64("jobs_failed", failed),
		slog.Int("queue_length", queueLen),
		slog.Int("active_workers", bp.workers))
}

// GetMetrics retorna métricas atuais
func (bp *BackgroundProcessor) GetMetrics() map[string]interface{} {
	bp.mutex.RLock()
	defer bp.mutex.RUnlock()

	return map[string]interface{}{
		"processed":      bp.processed,
		"failed":         bp.failed,
		"queue_length":   len(bp.jobs),
		"active_workers": bp.workers,
	}
}

// incrementProcessed incrementa contador de jobs processados
func (bp *BackgroundProcessor) incrementProcessed() {
	bp.mutex.Lock()
	bp.processed++
	bp.mutex.Unlock()
}

// incrementFailed incrementa contador de jobs falhados
func (bp *BackgroundProcessor) incrementFailed() {
	bp.mutex.Lock()
	bp.failed++
	bp.mutex.Unlock()
}

// generateJobID gera ID único para job
func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}

// Helper functions para criar jobs comuns

// CreateCacheWarmJob cria job para aquecimento de cache
func CreateCacheWarmJob(cacheKeys []string, priority int) *Job {
	return &Job{
		Type:     JobTypeCacheWarm,
		Priority: priority,
		Payload: map[string]interface{}{
			"cache_keys": cacheKeys,
		},
		MaxRetries: 2,
	}
}

// CreateDataSyncJob cria job para sincronização de dados
func CreateDataSyncJob(dataType string, entityIDs []int, priority int) *Job {
	return &Job{
		Type:     JobTypeDataSync,
		Priority: priority,
		Payload: map[string]interface{}{
			"data_type":  dataType,
			"entity_ids": entityIDs,
		},
		MaxRetries: 3,
	}
}

// CreateAnalyticsJob cria job para computação de analytics
func CreateAnalyticsJob(analyticsType string, year int, priority int) *Job {
	return &Job{
		Type:     JobTypeAnalytics,
		Priority: priority,
		Payload: map[string]interface{}{
			"analytics_type": analyticsType,
			"year":           year,
		},
		MaxRetries: 2,
	}
}

// CreateCleanupJob cria job para limpeza de dados
func CreateCleanupJob(cleanupType string, olderThan time.Time, priority int) *Job {
	return &Job{
		Type:     JobTypeCleanup,
		Priority: priority,
		Payload: map[string]interface{}{
			"cleanup_type": cleanupType,
			"older_than":   olderThan.Unix(),
		},
		MaxRetries: 1,
	}
}
