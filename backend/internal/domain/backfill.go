package domain

import (
	"errors"
	"fmt"
	"time"
)

// BackfillExecution representa uma execução de backfill
type BackfillExecution struct {
	ID          int    `json:"id" db:"id"`
	ExecutionID string `json:"execution_id" db:"execution_id"`
	Tipo        string `json:"tipo" db:"tipo"` // 'historico', 'incremental', 'manual'
	AnoInicio   int    `json:"ano_inicio" db:"ano_inicio"`
	AnoFim      int    `json:"ano_fim" db:"ano_fim"`
	Status      string `json:"status" db:"status"` // 'running', 'success', 'failed', 'partial'

	// Métricas de execução
	DeputadosProcessados   int `json:"deputados_processados" db:"deputados_processados"`
	ProposicoesProcessadas int `json:"proposicoes_processadas" db:"proposicoes_processadas"`
	DespesasProcessadas    int `json:"despesas_processadas" db:"despesas_processadas"`
	VotacoesProcessadas    int `json:"votacoes_processadas" db:"votacoes_processadas"`

	// Controle temporal
	StartedAt       time.Time  `json:"started_at" db:"started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	DurationSeconds *int       `json:"duration_seconds,omitempty" db:"duration_seconds"`

	// Metadados
	TriggeredBy  string                 `json:"triggered_by" db:"triggered_by"`
	ErrorMessage *string                `json:"error_message,omitempty" db:"error_message"`
	Config       map[string]interface{} `json:"config,omitempty" db:"config"`
}

// BackfillConfig representa configurações para execução de backfill
type BackfillConfig struct {
	AnoInicio           int      `json:"ano_inicio"`
	AnoFim              int      `json:"ano_fim"`
	Tipo                string   `json:"tipo"`
	TriggeredBy         string   `json:"triggered_by"`
	IncluirDeputados    bool     `json:"incluir_deputados"`
	IncluirProposicoes  bool     `json:"incluir_proposicoes"`
	IncluirDespesas     bool     `json:"incluir_despesas"`
	IncluirVotacoes     bool     `json:"incluir_votacoes"`
	ForcarReexecucao    bool     `json:"forcar_reexecucao"`
	BatchSize           int      `json:"batch_size"`
	ParallelWorkers     int      `json:"parallel_workers"`
	UfsFiltro           []string `json:"ufs_filtro,omitempty"`
	PartidosFiltro      []string `json:"partidos_filtro,omitempty"`
	DelayBetweenBatches int      `json:"delay_between_batches_ms"` // ms
}

// BackfillStatus representa status de uma execução
type BackfillStatus struct {
	ExecutionID            string    `json:"execution_id"`
	Status                 string    `json:"status"`
	ProgressPercentage     float64   `json:"progress_percentage"`
	CurrentOperation       string    `json:"current_operation"`
	DeputadosProcessados   int       `json:"deputados_processados"`
	ProposicoesProcessadas int       `json:"proposicoes_processadas"`
	DespesasProcessadas    int       `json:"despesas_processadas"`
	VotacoesProcessadas    int       `json:"votacoes_processadas"`
	EstimatedTimeRemaining *int      `json:"estimated_time_remaining_seconds,omitempty"`
	StartedAt              time.Time `json:"started_at"`
	LastUpdate             time.Time `json:"last_update"`
	ErrorsCount            int       `json:"errors_count"`
	LatestError            *string   `json:"latest_error,omitempty"`
}

// Constantes para tipos de backfill
const (
	BackfillTipoHistorico   = "historico"
	BackfillTipoIncremental = "incremental"
	BackfillTipoManual      = "manual"
)

// Constantes para tipos de scheduler
const (
	SchedulerTipoDiario  = "diario"
	SchedulerTipoRapido  = "rapido"
	SchedulerTipoManual  = "manual"
	SchedulerTipoInicial = "inicial"
)

// Constantes para status de execução
const (
	BackfillStatusRunning = "running"
	BackfillStatusSuccess = "success"
	BackfillStatusFailed  = "failed"
	BackfillStatusPartial = "partial"
)

// Constantes para quem disparou
const (
	BackfillTriggeredByScheduler = "scheduler"
	BackfillTriggeredByManual    = "manual"
	BackfillTriggeredByDeploy    = "deploy"
	BackfillTriggeredByAPI       = "api"
)

// SetDefaults define valores padrão para configuração
func (c *BackfillConfig) SetDefaults() {
	if c.AnoFim == 0 {
		c.AnoFim = time.Now().Year()
	}
	if c.AnoInicio == 0 {
		c.AnoInicio = 2022 // Padrão: dados dos últimos 3 anos
	}
	if c.Tipo == "" {
		c.Tipo = BackfillTipoHistorico
	}
	if c.TriggeredBy == "" {
		c.TriggeredBy = BackfillTriggeredByScheduler
	}
	if c.BatchSize == 0 {
		c.BatchSize = 100
	}
	if c.ParallelWorkers == 0 {
		c.ParallelWorkers = 3
	}
	if c.DelayBetweenBatches == 0 {
		c.DelayBetweenBatches = 500 // 500ms
	}

	// Por padrão, incluir tudo
	if !c.ForcarReexecucao {
		c.IncluirDeputados = true
		c.IncluirProposicoes = true
		c.IncluirDespesas = true
		c.IncluirVotacoes = true
	}
}

// Validate valida configuração de backfill
func (c *BackfillConfig) Validate() error {
	if c.AnoInicio < 1988 || c.AnoInicio > time.Now().Year() {
		return ErrBackfillAnoInvalido
	}
	if c.AnoFim < c.AnoInicio || c.AnoFim > time.Now().Year()+1 {
		return ErrBackfillAnoInvalido
	}
	if c.Tipo != BackfillTipoHistorico && c.Tipo != BackfillTipoIncremental && c.Tipo != BackfillTipoManual {
		return ErrBackfillTipoInvalido
	}
	if c.BatchSize < 10 || c.BatchSize > 1000 {
		return ErrBackfillConfigInvalida
	}
	if c.ParallelWorkers < 1 || c.ParallelWorkers > 10 {
		return ErrBackfillConfigInvalida
	}

	return nil
}

// IsCompleted verifica se execução foi completada
func (e *BackfillExecution) IsCompleted() bool {
	return e.Status == BackfillStatusSuccess || e.Status == BackfillStatusFailed || e.Status == BackfillStatusPartial
}

// IsSuccessful verifica se execução foi bem-sucedida
func (e *BackfillExecution) IsSuccessful() bool {
	return e.Status == BackfillStatusSuccess
}

// CalculateProgress calcula progresso baseado em métricas
func (e *BackfillExecution) CalculateProgress() float64 {
	if e.IsCompleted() {
		return 100.0
	}

	return CalculateProgressFromMetrics(
		e.DeputadosProcessados,
		e.ProposicoesProcessadas,
		e.DespesasProcessadas,
		e.VotacoesProcessadas,
	)
}

// CalculateProgressFromMetrics calcula o progresso percentual com base nas métricas acumuladas
func CalculateProgressFromMetrics(deputados, proposicoes, despesas, votacoes int) float64 {
	totalProcessed := deputados + proposicoes + despesas + votacoes
	if totalProcessed <= 0 {
		return 0
	}

	// Estimativa: ~600 deputados + 50k proposições + 500k despesas + 10k votações
	estimatedTotal := 560600

	progress := (float64(totalProcessed) / float64(estimatedTotal)) * 100.0
	if progress > 95 {
		progress = 95 // Nunca mostrar 100% até estar completo
	}

	return progress
}

// GetPeriodDescription retorna descrição amigável do período
func (e *BackfillExecution) GetPeriodDescription() string {
	if e.AnoInicio == e.AnoFim {
		return fmt.Sprintf("Dados de %d", e.AnoInicio)
	}
	return fmt.Sprintf("Dados de %d a %d", e.AnoInicio, e.AnoFim)
}

// Errors relacionados a backfill
var (
	ErrBackfillJaExecutado    = errors.New("backfill já foi executado com sucesso para este período")
	ErrBackfillEmExecucao     = errors.New("já existe um backfill em execução")
	ErrBackfillNaoEncontrado  = errors.New("execução de backfill não encontrada")
	ErrBackfillAnoInvalido    = errors.New("ano deve estar entre 1988 e o ano atual")
	ErrBackfillTipoInvalido   = errors.New("tipo de backfill inválido")
	ErrBackfillConfigInvalida = errors.New("configuração de backfill inválida")
	ErrBackfillStatusInvalido = errors.New("status de backfill inválido")
	ErrDespesasFonteDegradada = errors.New("dados de despesas retornados em modo degradado")
)
