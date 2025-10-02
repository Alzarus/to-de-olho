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

// BackfillHandler handles HTTP requests related to backfill operations
type BackfillHandler struct {
	backfillService *application.SmartBackfillService
	logger          *slog.Logger
}

// NewBackfillHandler creates a new BackfillHandler
func NewBackfillHandler(backfillService *application.SmartBackfillService, logger *slog.Logger) *BackfillHandler {
	return &BackfillHandler{
		backfillService: backfillService,
		logger:          logger,
	}
}

// GetCurrentStatus retorna o status atual da execução de backfill
func (h *BackfillHandler) GetCurrentStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	status, err := h.backfillService.GetCurrentStatus(ctx)
	if err != nil {
		h.logger.Error("failed to get backfill status",
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
			"message": "Nenhuma execução de backfill em andamento",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

// ListExecutions lista execuções de backfill com paginação
func (h *BackfillHandler) ListExecutions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

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

	executions, total, err := h.backfillService.ListExecutions(ctx, limit, offset)
	if err != nil {
		h.logger.Error("failed to list backfill executions",
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
func (h *BackfillHandler) GetExecutionStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extrair ID da URL (assumindo formato /backfill/executions/{id})
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/backfill/executions/")
	executionIDStr := strings.Split(path, "/")[0]

	executionID, err := strconv.Atoi(executionIDStr)
	if err != nil {
		http.Error(w, "Invalid execution ID", http.StatusBadRequest)
		return
	}

	executions, _, err := h.backfillService.ListExecutions(ctx, 1000, 0) // Buscar todas para encontrar por ID
	if err != nil {
		h.logger.Error("failed to get backfill execution",
			slog.Int("execution_id", executionID),
			slog.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Procurar execução por ID
	var execution *domain.BackfillExecution
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

// RegisterRoutes registra as rotas do handler no ServeMux
func (h *BackfillHandler) RegisterRoutes(mux *http.ServeMux) {
	// Rotas de backfill
	mux.HandleFunc("/api/v1/backfill/status", h.GetCurrentStatus)
	mux.HandleFunc("/api/v1/backfill/executions", h.ListExecutions)
	mux.HandleFunc("/api/v1/backfill/executions/", h.GetExecutionStatus) // Tratará sub-paths
}
