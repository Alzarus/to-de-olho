package domain

import (
	"errors"
	"time"
)

// SchedulerExecution representa uma execução de scheduler (sincronização incremental)
type SchedulerExecution struct {
	ID          int    `json:"id" db:"id"`
	ExecutionID string `json:"execution_id" db:"execution_id"`
	Tipo        string `json:"tipo" db:"tipo"`     // 'diario', 'rapido', 'manual', 'inicial'
	Status      string `json:"status" db:"status"` // 'running', 'success', 'failed', 'partial'

	// Métricas de execução
	DeputadosSincronizados   int `json:"deputados_sincronizados" db:"deputados_sincronizados"`
	ProposicoesSincronizadas int `json:"proposicoes_sincronizadas" db:"proposicoes_sincronizadas"`
	DespesasSincronizadas    int `json:"despesas_sincronizadas" db:"despesas_sincronizadas"`
	VotacoesSincronizadas    int `json:"votacoes_sincronizadas" db:"votacoes_sincronizadas"`

	// Controle temporal
	StartedAt       time.Time  `json:"started_at" db:"started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	DurationSeconds *int       `json:"duration_seconds,omitempty" db:"duration_seconds"`
	NextExecution   *time.Time `json:"next_execution,omitempty" db:"next_execution"`

	// Metadados
	TriggeredBy  string                 `json:"triggered_by" db:"triggered_by"`
	ErrorMessage *string                `json:"error_message,omitempty" db:"error_message"`
	Config       map[string]interface{} `json:"config,omitempty" db:"config"`

	// Auditoria
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// SchedulerConfig representa configurações para execução de scheduler
type SchedulerConfig struct {
	Tipo                string                 `json:"tipo"`
	TriggeredBy         string                 `json:"triggered_by"`
	MinIntervalHours    int                    `json:"min_interval_hours"`
	TimeoutMinutes      int                    `json:"timeout_minutes"`
	IncluirDeputados    bool                   `json:"incluir_deputados"`
	IncluirProposicoes  bool                   `json:"incluir_proposicoes"`
	IncluirDespesas     bool                   `json:"incluir_despesas"`
	IncluirVotacoes     bool                   `json:"incluir_votacoes"`
	BatchSize           int                    `json:"batch_size"`
	ParallelWorkers     int                    `json:"parallel_workers"`
	DelayBetweenBatches int                    `json:"delay_between_batches_ms"`
	Config              map[string]interface{} `json:"config,omitempty"`
}

// SchedulerStatus representa status de uma execução de scheduler
type SchedulerStatus struct {
	ExecutionID              string     `json:"execution_id"`
	Tipo                     string     `json:"tipo"`
	Status                   string     `json:"status"`
	CurrentOperation         string     `json:"current_operation"`
	DeputadosSincronizados   int        `json:"deputados_sincronizados"`
	ProposicoesSincronizadas int        `json:"proposicoes_sincronizadas"`
	DespesasSincronizadas    int        `json:"despesas_sincronizadas"`
	VotacoesSincronizadas    int        `json:"votacoes_sincronizadas"`
	StartedAt                time.Time  `json:"started_at"`
	LastUpdate               time.Time  `json:"last_update"`
	NextExecution            *time.Time `json:"next_execution,omitempty"`
	ErrorsCount              int        `json:"errors_count"`
	LatestError              *string    `json:"latest_error,omitempty"`
}

// ShouldExecuteResult representa resultado da verificação se scheduler deve executar
type ShouldExecuteResult struct {
	ShouldRun      bool       `json:"should_run"`
	Reason         string     `json:"reason"`
	LastExecution  *time.Time `json:"last_execution,omitempty"`
	HoursSinceLast *float64   `json:"hours_since_last,omitempty"`
}

// Validate valida uma configuração de scheduler
func (sc *SchedulerConfig) Validate() error {
	if sc.Tipo == "" {
		return errors.New("tipo é obrigatório")
	}

	validTypes := map[string]bool{
		SchedulerTipoDiario:  true,
		SchedulerTipoRapido:  true,
		SchedulerTipoManual:  true,
		SchedulerTipoInicial: true,
	}

	if !validTypes[sc.Tipo] {
		return errors.New("tipo inválido: deve ser diario, rapido, manual ou inicial")
	}

	if sc.MinIntervalHours < 0 {
		return errors.New("min_interval_hours deve ser >= 0")
	}

	if sc.TimeoutMinutes <= 0 {
		sc.TimeoutMinutes = 30 // padrão
	}

	if sc.BatchSize <= 0 {
		sc.BatchSize = 50 // padrão
	}

	if sc.ParallelWorkers <= 0 {
		sc.ParallelWorkers = 2 // padrão
	}

	if sc.DelayBetweenBatches < 0 {
		sc.DelayBetweenBatches = 100 // padrão
	}

	return nil
}

// GetDefaultSchedulerConfig retorna configuração padrão para um tipo de scheduler
func GetDefaultSchedulerConfig(tipo string) *SchedulerConfig {
	config := &SchedulerConfig{
		Tipo:                tipo,
		TriggeredBy:         "cron",
		IncluirDeputados:    true,
		IncluirProposicoes:  true,
		IncluirDespesas:     true,
		IncluirVotacoes:     true,
		BatchSize:           50,
		ParallelWorkers:     2,
		DelayBetweenBatches: 100,
		TimeoutMinutes:      30,
	}

	// Configurações específicas por tipo
	switch tipo {
	case SchedulerTipoDiario:
		config.MinIntervalHours = 20 // Mínimo 20h entre execuções diárias
		config.TimeoutMinutes = 60   // Timeout maior para sync completa
		config.ParallelWorkers = 3   // Mais workers para sync diária
	case SchedulerTipoRapido:
		config.MinIntervalHours = 3 // Mínimo 3h entre execuções rápidas
		config.TimeoutMinutes = 15  // Timeout menor para sync rápida
		config.ParallelWorkers = 2  // Menos workers para não sobrecarregar
	case SchedulerTipoInicial:
		config.MinIntervalHours = 0 // Sempre pode executar
		config.TimeoutMinutes = 10  // Timeout baixo para teste inicial
	case SchedulerTipoManual:
		config.MinIntervalHours = 0 // Sempre pode executar quando manual
		config.TimeoutMinutes = 45  // Timeout maior para execução manual
	}

	return config
}

// IsRunning verifica se a execução está em andamento
func (se *SchedulerExecution) IsRunning() bool {
	return se.Status == BackfillStatusRunning
}

// IsCompleted verifica se a execução foi concluída (sucesso ou falha)
func (se *SchedulerExecution) IsCompleted() bool {
	return se.CompletedAt != nil
}

// GetTotalSyncronized retorna total de itens sincronizados
func (se *SchedulerExecution) GetTotalSyncronized() int {
	return se.DeputadosSincronizados + se.ProposicoesSincronizadas + se.DespesasSincronizadas + se.VotacoesSincronizadas
}

// GetDurationMinutes retorna duração em minutos
func (se *SchedulerExecution) GetDurationMinutes() float64 {
	if se.DurationSeconds == nil {
		return 0
	}
	return float64(*se.DurationSeconds) / 60.0
}

// ErrSchedulerAlreadyRunning é retornado quando já existe uma execução em andamento para o tipo
var ErrSchedulerAlreadyRunning = errors.New("scheduler já em execução para este tipo")
