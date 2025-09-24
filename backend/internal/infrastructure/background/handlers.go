package background

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"to-de-olho-backend/internal/application"
	"to-de-olho-backend/internal/infrastructure/cache"
)

// CacheWarmHandler aquece cache com dados frequentemente acessados
type CacheWarmHandler struct {
	deputadosService *application.DeputadosService
	analyticsService *application.AnalyticsService
	cache            cache.CacheInterface
	logger           *slog.Logger
}

// NewCacheWarmHandler cria handler para aquecimento de cache
func NewCacheWarmHandler(
	deputadosService *application.DeputadosService,
	analyticsService *application.AnalyticsService,
	cache cache.CacheInterface,
	logger *slog.Logger,
) *CacheWarmHandler {
	return &CacheWarmHandler{
		deputadosService: deputadosService,
		analyticsService: analyticsService,
		cache:            cache,
		logger:           logger,
	}
}

// Handle processa job de aquecimento de cache
func (h *CacheWarmHandler) Handle(ctx context.Context, job *Job) error {
	h.logger.Info("starting cache warm job", slog.String("job_id", job.ID))

	cacheType, ok := job.Payload["cache_type"].(string)
	if !ok {
		cacheType = "all" // default
	}

	switch cacheType {
	case "deputados":
		return h.warmDeputadosCache(ctx)
	case "analytics":
		return h.warmAnalyticsCache(ctx)
	case "all":
		if err := h.warmDeputadosCache(ctx); err != nil {
			return err
		}
		return h.warmAnalyticsCache(ctx)
	default:
		return fmt.Errorf("unknown cache type: %s", cacheType)
	}
}

// warmDeputadosCache aquece cache de deputados
func (h *CacheWarmHandler) warmDeputadosCache(ctx context.Context) error {
	h.logger.Info("warming deputados cache")

	// Buscar dados mais acessados
	_, _, err := h.deputadosService.ListarDeputados(ctx, "", "", "")
	if err != nil {
		return fmt.Errorf("erro ao aquecer cache de deputados: %w", err)
	}

	h.logger.Info("deputados cache warmed successfully")
	return nil
}

// warmAnalyticsCache aquece cache de analytics
func (h *CacheWarmHandler) warmAnalyticsCache(ctx context.Context) error {
	h.logger.Info("warming analytics cache")

	currentYear := time.Now().Year()

	// Aquecer rankings mais consultados
	rankings := []string{"gastos", "proposicoes", "presenca"}

	for _, ranking := range rankings {
		switch ranking {
		case "gastos":
			_, _, err := h.analyticsService.GetRankingGastos(ctx, currentYear, 50)
			if err != nil {
				h.logger.Warn("failed to warm gastos ranking", slog.String("error", err.Error()))
			}
		case "proposicoes":
			_, _, err := h.analyticsService.GetRankingProposicoes(ctx, currentYear, 50)
			if err != nil {
				h.logger.Warn("failed to warm proposicoes ranking", slog.String("error", err.Error()))
			}
		case "presenca":
			_, _, err := h.analyticsService.GetRankingPresenca(ctx, currentYear, 50)
			if err != nil {
				h.logger.Warn("failed to warm presenca ranking", slog.String("error", err.Error()))
			}
		}
	}

	// Aquecer insights gerais
	_, _, err := h.analyticsService.GetInsightsGerais(ctx)
	if err != nil {
		h.logger.Warn("failed to warm insights", slog.String("error", err.Error()))
	}

	h.logger.Info("analytics cache warmed successfully")
	return nil
}

// DataSyncHandler sincroniza dados com API da Câmara
type DataSyncHandler struct {
	deputadosService *application.DeputadosService
	logger           *slog.Logger
}

// NewDataSyncHandler cria handler para sincronização de dados
func NewDataSyncHandler(
	deputadosService *application.DeputadosService,
	logger *slog.Logger,
) *DataSyncHandler {
	return &DataSyncHandler{
		deputadosService: deputadosService,
		logger:           logger,
	}
}

// Handle processa job de sincronização de dados
func (h *DataSyncHandler) Handle(ctx context.Context, job *Job) error {
	h.logger.Info("starting data sync job", slog.String("job_id", job.ID))

	dataType, ok := job.Payload["data_type"].(string)
	if !ok {
		return fmt.Errorf("data_type not specified")
	}

	switch dataType {
	case "deputados":
		return h.syncDeputados(ctx, job)
	case "despesas":
		return h.syncDespesas(ctx, job)
	default:
		return fmt.Errorf("unknown data type: %s", dataType)
	}
}

// syncDeputados sincroniza dados de deputados
func (h *DataSyncHandler) syncDeputados(ctx context.Context, job *Job) error {
	h.logger.Info("syncing deputados data")

	// Buscar deputados para forçar sincronização
	_, _, err := h.deputadosService.ListarDeputados(ctx, "", "", "")
	if err != nil {
		return fmt.Errorf("erro ao sincronizar deputados: %w", err)
	}

	h.logger.Info("deputados data synced successfully")
	return nil
}

// syncDespesas sincroniza despesas de deputados específicos
func (h *DataSyncHandler) syncDespesas(ctx context.Context, job *Job) error {
	h.logger.Info("syncing despesas data")

	entityIDs, ok := job.Payload["entity_ids"].([]interface{})
	if !ok {
		return fmt.Errorf("entity_ids not specified or invalid format")
	}

	year := time.Now().Year()
	if yearPayload, exists := job.Payload["year"]; exists {
		if yearFloat, ok := yearPayload.(float64); ok {
			year = int(yearFloat)
		}
	}

	// Sincronizar despesas para cada deputado
	for _, idInterface := range entityIDs {
		var deputadoID string

		switch id := idInterface.(type) {
		case float64:
			deputadoID = strconv.Itoa(int(id))
		case int:
			deputadoID = strconv.Itoa(id)
		case string:
			deputadoID = id
		default:
			h.logger.Warn("invalid entity_id format", slog.Any("id", id))
			continue
		}

		_, _, err := h.deputadosService.ListarDespesas(ctx, deputadoID, strconv.Itoa(year))
		if err != nil {
			h.logger.Warn("failed to sync despesas",
				slog.String("deputado_id", deputadoID),
				slog.Int("year", year),
				slog.String("error", err.Error()))
		}
	}

	h.logger.Info("despesas data synced successfully")
	return nil
}

// AnalyticsHandler computa rankings e insights
type AnalyticsHandler struct {
	analyticsService *application.AnalyticsService
	logger           *slog.Logger
}

// NewAnalyticsHandler cria handler para analytics
func NewAnalyticsHandler(
	analyticsService *application.AnalyticsService,
	logger *slog.Logger,
) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		logger:           logger,
	}
}

// Handle processa job de analytics
func (h *AnalyticsHandler) Handle(ctx context.Context, job *Job) error {
	h.logger.Info("starting analytics job", slog.String("job_id", job.ID))

	analyticsType, ok := job.Payload["analytics_type"].(string)
	if !ok {
		return fmt.Errorf("analytics_type not specified")
	}

	year := time.Now().Year()
	if yearPayload, exists := job.Payload["year"]; exists {
		if yearFloat, ok := yearPayload.(float64); ok {
			year = int(yearFloat)
		}
	}

	switch analyticsType {
	case "rankings":
		return h.computeRankings(ctx, year)
	case "insights":
		return h.computeInsights(ctx, year)
	case "all":
		if err := h.computeRankings(ctx, year); err != nil {
			return err
		}
		return h.computeInsights(ctx, year)
	default:
		return fmt.Errorf("unknown analytics type: %s", analyticsType)
	}
}

// computeRankings computa todos os rankings
func (h *AnalyticsHandler) computeRankings(ctx context.Context, year int) error {
	h.logger.Info("computing rankings", slog.Int("year", year))

	// Computar ranking de gastos
	_, _, err := h.analyticsService.GetRankingGastos(ctx, year, 100)
	if err != nil {
		return fmt.Errorf("erro ao computar ranking de gastos: %w", err)
	}

	// Computar ranking de proposições
	_, _, err = h.analyticsService.GetRankingProposicoes(ctx, year, 100)
	if err != nil {
		return fmt.Errorf("erro ao computar ranking de proposições: %w", err)
	}

	// Computar ranking de presença
	_, _, err = h.analyticsService.GetRankingPresenca(ctx, year, 100)
	if err != nil {
		return fmt.Errorf("erro ao computar ranking de presença: %w", err)
	}

	h.logger.Info("rankings computed successfully", slog.Int("year", year))
	return nil
}

// computeInsights computa insights gerais
func (h *AnalyticsHandler) computeInsights(ctx context.Context, year int) error {
	h.logger.Info("computing insights", slog.Int("year", year))

	_, _, err := h.analyticsService.GetInsightsGerais(ctx)
	if err != nil {
		return fmt.Errorf("erro ao computar insights: %w", err)
	}

	h.logger.Info("insights computed successfully", slog.Int("year", year))
	return nil
}

// CleanupHandler limpa dados antigos
type CleanupHandler struct {
	cache  cache.CacheInterface
	logger *slog.Logger
}

// NewCleanupHandler cria handler para limpeza
func NewCleanupHandler(cache cache.CacheInterface, logger *slog.Logger) *CleanupHandler {
	return &CleanupHandler{
		cache:  cache,
		logger: logger,
	}
}

// Handle processa job de limpeza
func (h *CleanupHandler) Handle(ctx context.Context, job *Job) error {
	h.logger.Info("starting cleanup job", slog.String("job_id", job.ID))

	cleanupType, ok := job.Payload["cleanup_type"].(string)
	if !ok {
		return fmt.Errorf("cleanup_type not specified")
	}

	switch cleanupType {
	case "cache":
		return h.cleanupCache(ctx)
	case "logs":
		return h.cleanupLogs(ctx, job)
	default:
		return fmt.Errorf("unknown cleanup type: %s", cleanupType)
	}
}

// cleanupCache limpa cache expirado
func (h *CleanupHandler) cleanupCache(ctx context.Context) error {
	h.logger.Info("cleaning up expired cache")

	// Se cache implementa CleanupProvider, usar método específico
	if cleanupProvider, ok := h.cache.(cache.CleanupProvider); ok {
		cleanupProvider.CleanupExpired()
		h.logger.Info("cache cleanup completed using CleanupProvider")
	} else {
		h.logger.Info("cache does not support cleanup interface")
	}

	return nil
}

// cleanupLogs limpa logs antigos (placeholder)
func (h *CleanupHandler) cleanupLogs(ctx context.Context, job *Job) error {
	h.logger.Info("cleaning up old logs")

	// Implementar limpeza de logs se necessário
	// Por enquanto é apenas um placeholder

	h.logger.Info("log cleanup completed")
	return nil
}
