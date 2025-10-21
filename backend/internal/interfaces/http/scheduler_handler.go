package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"to-de-olho-backend/internal/application"
	"to-de-olho-backend/internal/domain"
)

// SchedulerHandler handles HTTP requests related to scheduler operations
type SchedulerHandler struct {
	schedulerService *application.SmartSchedulerService
	logger           *slog.Logger
}

// NewSchedulerHandler creates a new SchedulerHandler
func NewSchedulerHandler(schedulerService *application.SmartSchedulerService, logger *slog.Logger) *SchedulerHandler {
	return &SchedulerHandler{
		schedulerService: schedulerService,
		logger:           logger,
	}
}

// GetCurrentStatus retorna o status atual da execução de scheduler
func (h *SchedulerHandler) GetCurrentStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parâmetro opcional para filtrar por tipo
	schedulerTipo := r.URL.Query().Get("tipo")
	var tipoPtr *string
	if schedulerTipo != "" {
		tipoPtr = &schedulerTipo
	}

	status, err := h.schedulerService.GetCurrentStatus(ctx, tipoPtr)
	if err != nil {
		h.logger.Error("failed to get scheduler status",
			slog.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if status == nil {
		// Nenhuma execução em andamento
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "idle",
			"message": "Nenhuma execução de scheduler em andamento",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

// ListExecutions lista execuções de scheduler com paginação
func (h *SchedulerHandler) ListExecutions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	schedulerTipo := r.URL.Query().Get("tipo")

	limit := 10
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var tipoPtr *string
	if schedulerTipo != "" {
		tipoPtr = &schedulerTipo
	}

	executions, total, err := h.schedulerService.ListExecutions(ctx, limit, offset, tipoPtr)
	if err != nil {
		h.logger.Error("failed to list scheduler executions",
			slog.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"executions": executions,
		"pagination": map[string]interface{}{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetExecutionStatus retorna detalhes de uma execução específica
func (h *SchedulerHandler) GetExecutionStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extrair ID da URL (assumindo formato /scheduler/executions/{id})
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/scheduler/executions/")
	executionIDStr := strings.Split(path, "/")[0]

	executionID, err := strconv.Atoi(executionIDStr)
	if err != nil {
		http.Error(w, "Invalid execution ID", http.StatusBadRequest)
		return
	}

	executions, _, err := h.schedulerService.ListExecutions(ctx, 1000, 0, nil) // Buscar todas para encontrar por ID
	if err != nil {
		h.logger.Error("failed to get scheduler execution",
			slog.Int("execution_id", executionID),
			slog.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Procurar execução por ID
	var execution *domain.SchedulerExecution
	for _, exec := range executions {
		if exec.ID == executionID {
			execution = &exec
			break
		}
	}

	if execution == nil {
		http.Error(w, "Execution not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(execution)
}

// TriggerManualScheduler dispara uma execução manual de scheduler
func (h *SchedulerHandler) TriggerManualScheduler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// Parse request body
	var requestBody struct {
		Tipo string `json:"tipo"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validar tipo
	validTypes := map[string]bool{
		domain.SchedulerTipoDiario: true,
		domain.SchedulerTipoRapido: true,
		domain.SchedulerTipoManual: true,
	}

	if !validTypes[requestBody.Tipo] {
		http.Error(w, "Invalid scheduler type", http.StatusBadRequest)
		return
	}

	// Criar configuração para execução manual
	config := domain.GetDefaultSchedulerConfig(requestBody.Tipo)
	config.TriggeredBy = "api-manual"
	config.MinIntervalHours = 0 // Sempre pode executar quando manual

	execution, err := h.schedulerService.ExecuteIntelligentScheduler(ctx, config)
	if err != nil {
		h.logger.Error("failed to trigger manual scheduler",
			slog.String("tipo", requestBody.Tipo),
			slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	h.logger.Info("manual scheduler triggered",
		slog.String("execution_id", execution.ExecutionID),
		slog.String("tipo", execution.Tipo))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(execution)
}

// RegisterRoutes registra as rotas do handler no ServeMux
func (h *SchedulerHandler) RegisterRoutes(mux *http.ServeMux) {
	// Rotas de scheduler
	mux.HandleFunc("/api/v1/scheduler/status", h.GetCurrentStatus)
	mux.HandleFunc("/api/v1/scheduler/executions", h.ListExecutions)
	mux.HandleFunc("/api/v1/scheduler/executions/", h.GetExecutionStatus) // Tratará sub-paths
	mux.HandleFunc("/api/v1/scheduler/trigger", h.TriggerManualScheduler) // POST para execução manual
}
