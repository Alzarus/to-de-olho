package ingestor

import (
	"context"
	"fmt"
	"log"
	"time"

	"to-de-olho-backend/internal/application"
	"to-de-olho-backend/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

// IncrementalSyncManager gerencia sincronizações incrementais
type IncrementalSyncManager struct {
	deputadosService   *application.DeputadosService
	proposicoesService *application.ProposicoesService
	analyticsService   application.AnalyticsServiceInterface
	db                 *pgxpool.Pool
	cache              application.CachePort
}

// SyncMetrics métricas da sincronização
type SyncMetrics struct {
	StartTime          time.Time     `json:"start_time"`
	EndTime            time.Time     `json:"end_time"`
	Duration           time.Duration `json:"duration"`
	DeputadosUpdated   int           `json:"deputados_updated"`
	ProposicoesUpdated int           `json:"proposicoes_updated"`
	ErrorsCount        int           `json:"errors_count"`
	Errors             []string      `json:"errors,omitempty"`
	SyncType           string        `json:"sync_type"` // "daily", "quick"
}

// NewIncrementalSyncManager cria novo gerenciador de sync incremental
func NewIncrementalSyncManager(
	deputadosService *application.DeputadosService,
	proposicoesService *application.ProposicoesService,
	analyticsService application.AnalyticsServiceInterface,
	db *pgxpool.Pool,
	cache application.CachePort,
) *IncrementalSyncManager {
	return &IncrementalSyncManager{
		deputadosService:   deputadosService,
		proposicoesService: proposicoesService,
		analyticsService:   analyticsService,
		db:                 db,
		cache:              cache,
	}
}

// ExecuteDailySync executa sincronização completa diária
func (ism *IncrementalSyncManager) ExecuteDailySync(ctx context.Context) error {
	metrics := &SyncMetrics{
		StartTime: time.Now(),
		SyncType:  "daily",
		Errors:    []string{},
	}

	log.Println("🌅 Iniciando sincronização diária completa")

	// 1. Sincronizar deputados (base fundamental)
	if err := ism.syncDeputados(ctx, metrics); err != nil {
		metrics.Errors = append(metrics.Errors, fmt.Sprintf("Deputados: %v", err))
		metrics.ErrorsCount++
		log.Printf("❌ Erro na sincronização de deputados: %v", err)
	}

	// 2. Sincronizar proposições das últimas 24h
	if err := ism.syncRecentProposicoes(ctx, metrics); err != nil {
		metrics.Errors = append(metrics.Errors, fmt.Sprintf("Proposições: %v", err))
		metrics.ErrorsCount++
		log.Printf("❌ Erro na sincronização de proposições: %v", err)
	}

	// 3. Limpar cache antigo
	if err := ism.cleanupOldCache(ctx); err != nil {
		log.Printf("⚠️  Aviso: erro na limpeza de cache: %v", err)
	}

	// Finalizar métricas
	metrics.EndTime = time.Now()
	metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)

	// Atualizar rankings analytics após sincronização completa
	if ism.analyticsService != nil {
		log.Println("📊 Atualizando rankings e analytics...")
		if err := ism.analyticsService.AtualizarRankings(ctx); err != nil {
			metrics.Errors = append(metrics.Errors, fmt.Sprintf("Analytics: %v", err))
			metrics.ErrorsCount++
			log.Printf("⚠️  Erro ao atualizar analytics: %v", err)
		} else {
			log.Println("✅ Rankings atualizados com sucesso")
		}
	}

	// Persistir métricas
	if err := ism.saveSyncMetrics(ctx, metrics); err != nil {
		log.Printf("⚠️  Erro ao salvar métricas: %v", err)
	}

	log.Printf("📊 Sync diário: %d deputados, %d proposições, %d erros em %v",
		metrics.DeputadosUpdated, metrics.ProposicoesUpdated,
		metrics.ErrorsCount, metrics.Duration)

	if metrics.ErrorsCount > 0 {
		return fmt.Errorf("sincronização com %d erros", metrics.ErrorsCount)
	}

	return nil
}

// ExecuteQuickSync executa sincronização rápida (apenas dados críticos)
func (ism *IncrementalSyncManager) ExecuteQuickSync(ctx context.Context) error {
	metrics := &SyncMetrics{
		StartTime: time.Now(),
		SyncType:  "quick",
		Errors:    []string{},
	}

	log.Println("⚡ Iniciando sincronização rápida")

	// Apenas proposições das últimas 4h (mais voláteis)
	if err := ism.syncRecentProposicoes(ctx, metrics); err != nil {
		metrics.Errors = append(metrics.Errors, fmt.Sprintf("Proposições: %v", err))
		metrics.ErrorsCount++
	}

	metrics.EndTime = time.Now()
	metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)

	log.Printf("⚡ Sync rápido: %d proposições, %d erros em %v",
		metrics.ProposicoesUpdated, metrics.ErrorsCount, metrics.Duration)

	return nil
}

// syncDeputados sincroniza lista completa de deputados
func (ism *IncrementalSyncManager) syncDeputados(ctx context.Context, metrics *SyncMetrics) error {
	log.Println("👥 Sincronizando deputados...")

	// Buscar deputados atuais
	deputados, source, err := ism.deputadosService.ListarDeputados(ctx, "", "", "")
	if err != nil {
		return fmt.Errorf("erro ao buscar deputados: %w", err)
	}

	// Só contar como atualização se veio da API
	if source == "api" {
		metrics.DeputadosUpdated = len(deputados)
		log.Printf("✅ %d deputados sincronizados da API", len(deputados))
	} else {
		log.Printf("📄 Deputados obtidos do %s", source)
	}

	return nil
}

// syncRecentProposicoes sincroniza proposições recentes
func (ism *IncrementalSyncManager) syncRecentProposicoes(ctx context.Context, metrics *SyncMetrics) error {
	log.Println("📜 Sincronizando proposições recentes...")

	// Filtro para proposições das últimas 24h
	filtros := &domain.ProposicaoFilter{
		DataApresentacaoInicio: func() *time.Time {
			t := time.Now().AddDate(0, 0, -1) // Últimas 24h
			return &t
		}(),
		Ordem:      "DESC",
		OrdenarPor: "dataApresentacao",
		Limite:     50, // Limite conservador para sync incremental
	}

	proposicoes, _, source, err := ism.proposicoesService.ListarProposicoes(ctx, filtros)
	if err != nil {
		return fmt.Errorf("erro ao buscar proposições: %w", err)
	}

	if source == "api" {
		metrics.ProposicoesUpdated = len(proposicoes)
		log.Printf("✅ %d proposições sincronizadas da API", len(proposicoes))
	} else {
		log.Printf("📄 Proposições obtidas do %s", source)
	}

	return nil
}

// cleanupOldCache remove entradas de cache antigas
func (ism *IncrementalSyncManager) cleanupOldCache(ctx context.Context) error {
	log.Println("🧹 Limpando cache antigo...")

	// Por enquanto é placeholder - Redis tem TTL automático
	// Futuramente podemos implementar limpeza manual de chaves específicas

	return nil
}

// saveSyncMetrics persiste métricas da sincronização
func (ism *IncrementalSyncManager) saveSyncMetrics(ctx context.Context, metrics *SyncMetrics) error {
	query := `
		INSERT INTO sync_metrics (
			sync_type, start_time, end_time, duration_ms,
			deputados_updated, proposicoes_updated, errors_count, errors
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	durationMS := int(metrics.Duration.Milliseconds())
	var errorsJSON *string
	if len(metrics.Errors) > 0 {
		errorsStr := fmt.Sprintf("%v", metrics.Errors)
		errorsJSON = &errorsStr
	}

	_, err := ism.db.Exec(ctx, query,
		metrics.SyncType,
		metrics.StartTime,
		metrics.EndTime,
		durationMS,
		metrics.DeputadosUpdated,
		metrics.ProposicoesUpdated,
		metrics.ErrorsCount,
		errorsJSON,
	)

	if err != nil {
		return fmt.Errorf("erro ao salvar métricas: %w", err)
	}

	return nil
}

// GetSyncStats retorna estatísticas de sincronização
func (ism *IncrementalSyncManager) GetSyncStats(ctx context.Context, days int) ([]SyncMetrics, error) {
	query := `
		SELECT sync_type, start_time, end_time, duration_ms,
		       deputados_updated, proposicoes_updated, errors_count, errors
		FROM sync_metrics 
		WHERE start_time >= NOW() - INTERVAL '%d days'
		ORDER BY start_time DESC
		LIMIT 50
	`

	rows, err := ism.db.Query(ctx, fmt.Sprintf(query, days))
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar estatísticas: %w", err)
	}
	defer rows.Close()

	var stats []SyncMetrics
	for rows.Next() {
		var metric SyncMetrics
		var durationMS int
		var errorsJSON *string

		err := rows.Scan(
			&metric.SyncType,
			&metric.StartTime,
			&metric.EndTime,
			&durationMS,
			&metric.DeputadosUpdated,
			&metric.ProposicoesUpdated,
			&metric.ErrorsCount,
			&errorsJSON,
		)
		if err != nil {
			continue
		}

		metric.Duration = time.Duration(durationMS) * time.Millisecond
		if errorsJSON != nil {
			metric.Errors = []string{*errorsJSON}
		}

		stats = append(stats, metric)
	}

	return stats, nil
}
