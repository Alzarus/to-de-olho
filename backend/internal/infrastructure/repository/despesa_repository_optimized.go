package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"to-de-olho-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OptimizedDespesaRepository implementa operações otimizadas para despesas
type OptimizedDespesaRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

// NewOptimizedDespesaRepository cria repositório com prepared statements
func NewOptimizedDespesaRepository(db *pgxpool.Pool, logger *slog.Logger) *OptimizedDespesaRepository {
	if logger == nil {
		logger = slog.Default()
	}

	repo := &OptimizedDespesaRepository{
		db:     db,
		logger: logger,
	}

	// Preparar statements na inicialização
	if err := repo.prepareStatements(context.Background()); err != nil {
		logger.Warn("falha ao preparar statements", slog.String("error", err.Error()))
	}

	return repo
}

// prepareStatements documenta as queries otimizadas utilizadas
func (r *OptimizedDespesaRepository) prepareStatements(ctx context.Context) error {
	// Note: pgx v5 não suporta prepared statements globais no pool
	// Em vez disso, usaremos QueryRow/Query diretamente que são otimizadas pelo pgx
	r.logger.Info("query optimization ready for batch operations and connection pooling")
	return nil
}

// UpsertDespesasBatch insere/atualiza despesas em lote para máxima performance
func (r *OptimizedDespesaRepository) UpsertDespesasBatch(ctx context.Context, deputadoID int, ano int, despesas []domain.Despesa) error {
	if len(despesas) == 0 {
		return nil
	}

	start := time.Now()

	// Usar transação para batch
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback(ctx)

	// Preparar batch com CopyFrom para máxima performance
	rows := make([][]interface{}, len(despesas))
	for i, d := range despesas {
		payload, _ := json.Marshal(d)

		var dataDocumento interface{}
		if d.DataDocumento != "" {
			if parsedDate, err := time.Parse("2006-01-02", d.DataDocumento); err == nil {
				dataDocumento = parsedDate
			}
		}

		rows[i] = []interface{}{
			deputadoID, d.Ano, d.Mes, d.TipoDespesa, d.CodDocumento,
			d.TipoDocumento, d.CodTipoDocumento, dataDocumento, d.NumDocumento,
			d.ValorDocumento, d.URLDocumento, d.NomeFornecedor, d.CNPJCPFFornecedor,
			d.ValorLiquido, d.ValorBruto, d.ValorGlosa, d.NumRessarcimento,
			d.CodLote, d.Parcela, string(payload), time.Now(),
		}
	}

	// Usar COPY para inserção em massa ultra-rápida
	copyCount, err := tx.CopyFrom(ctx,
		pgx.Identifier{"despesas_cache"},
		[]string{
			"deputado_id", "ano", "mes", "tipo_despesa", "cod_documento",
			"tipo_documento", "cod_tipo_documento", "data_documento", "num_documento",
			"valor_documento", "url_documento", "nome_fornecedor", "cnpj_cpf_fornecedor",
			"valor_liquido", "valor_bruto", "valor_glosa", "num_ressarcimento",
			"cod_lote", "parcela", "payload", "updated_at",
		},
		pgx.CopyFromRows(rows),
	)

	if err != nil {
		// Se COPY falhar (por conflitos), usar upsert individual
		r.logger.Warn("CopyFrom failed, falling back to individual upserts",
			slog.String("error", err.Error()),
			slog.Int("deputado_id", deputadoID),
			slog.Int("total_despesas", len(despesas)))

		if err := r.upsertIndividual(ctx, tx, deputadoID, ano, despesas); err != nil {
			return fmt.Errorf("erro no upsert individual: %w", err)
		}

		// Commit da transação após upsert individual bem-sucedido
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("erro ao fazer commit da transação (fallback): %w", err)
		}

		r.logger.Info("despesas inseridas via upsert individual",
			slog.Int("deputado_id", deputadoID),
			slog.Int("ano", ano),
			slog.Int("inserted_count", len(despesas)),
			slog.Duration("duration", time.Since(start)))

		return nil
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("erro ao fazer commit da transação: %w", err)
	}

	r.logger.Info("despesas inseridas em lote",
		slog.Int("deputado_id", deputadoID),
		slog.Int("ano", ano),
		slog.Int64("inserted_count", copyCount),
		slog.Duration("duration", time.Since(start)))

	return nil
}

// upsertIndividual faz upsert individual quando batch falha
func (r *OptimizedDespesaRepository) upsertIndividual(ctx context.Context, tx pgx.Tx, deputadoID int, ano int, despesas []domain.Despesa) error {
	query := `
		INSERT INTO despesas_cache (
			deputado_id, ano, mes, tipo_despesa, cod_documento,
			tipo_documento, cod_tipo_documento, data_documento, num_documento,
			valor_documento, url_documento, nome_fornecedor, cnpj_cpf_fornecedor,
			valor_liquido, valor_bruto, valor_glosa, num_ressarcimento,
			cod_lote, parcela, payload, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, NOW()
		)
		ON CONFLICT (deputado_id, ano, cod_documento)
		DO UPDATE SET
			mes = EXCLUDED.mes,
			valor_documento = EXCLUDED.valor_documento,
			payload = EXCLUDED.payload,
			updated_at = NOW()
	`

	batch := &pgx.Batch{}
	for _, d := range despesas {
		payload, _ := json.Marshal(d)

		var dataDocumento interface{}
		if d.DataDocumento != "" {
			if parsedDate, err := time.Parse("2006-01-02", d.DataDocumento); err == nil {
				dataDocumento = parsedDate
			}
		}

		batch.Queue(query,
			deputadoID, d.Ano, d.Mes, d.TipoDespesa, d.CodDocumento,
			d.TipoDocumento, d.CodTipoDocumento, dataDocumento, d.NumDocumento,
			d.ValorDocumento, d.URLDocumento, d.NomeFornecedor, d.CNPJCPFFornecedor,
			d.ValorLiquido, d.ValorBruto, d.ValorGlosa, d.NumRessarcimento,
			d.CodLote, d.Parcela, string(payload),
		)
	}

	results := tx.SendBatch(ctx, batch)
	defer results.Close()

	// Coletar erros e retornar se algum falhou
	var errorList []error
	for i := 0; i < len(despesas); i++ {
		_, err := results.Exec()
		if err != nil {
			r.logger.Error("erro no upsert individual",
				slog.String("error", err.Error()),
				slog.Int("index", i))
			errorList = append(errorList, fmt.Errorf("erro no item %d do batch de upsert: %w", i, err))
		}
	}

	// Se houve erros, retornar todos os erros agregados (Go 1.20+)
	if len(errorList) > 0 {
		return errors.Join(errorList...)
	}

	return nil
}

// ListDespesasOptimized lista despesas com query otimizada
func (r *OptimizedDespesaRepository) ListDespesasOptimized(ctx context.Context, deputadoID int, ano int) ([]domain.Despesa, error) {
	start := time.Now()

	query := `
		SELECT 
			tipo_despesa, cod_documento, tipo_documento, cod_tipo_documento,
			data_documento, num_documento, valor_documento, url_documento,
			nome_fornecedor, cnpj_cpf_fornecedor, valor_liquido, valor_bruto,
			valor_glosa, num_ressarcimento, cod_lote, parcela
		FROM despesas_cache 
		WHERE deputado_id = $1 AND ano = $2
		ORDER BY mes DESC, valor_documento DESC
	`

	rows, err := r.db.Query(ctx, query, deputadoID, ano)
	if err != nil {
		return nil, fmt.Errorf("erro ao executar query: %w", err)
	}
	defer rows.Close()

	var despesas []domain.Despesa
	for rows.Next() {
		var d domain.Despesa
		var dataDocumento *time.Time

		err := rows.Scan(
			&d.TipoDespesa, &d.CodDocumento, &d.TipoDocumento, &d.CodTipoDocumento,
			&dataDocumento, &d.NumDocumento, &d.ValorDocumento, &d.URLDocumento,
			&d.NomeFornecedor, &d.CNPJCPFFornecedor, &d.ValorLiquido, &d.ValorBruto,
			&d.ValorGlosa, &d.NumRessarcimento, &d.CodLote, &d.Parcela,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao fazer scan: %w", err)
		}

		// Converter data para string se não for nil
		if dataDocumento != nil {
			d.DataDocumento = dataDocumento.Format("2006-01-02")
		}

		d.Ano = ano // Sempre será o ano solicitado
		despesas = append(despesas, d)
	}

	r.logger.Debug("despesas listadas",
		slog.Int("deputado_id", deputadoID),
		slog.Int("ano", ano),
		slog.Int("count", len(despesas)),
		slog.Duration("duration", time.Since(start)))

	return despesas, nil
}

// GetStatsOptimized obtém estatísticas agregadas de forma otimizada
func (r *OptimizedDespesaRepository) GetStatsOptimized(ctx context.Context, deputadoID int, ano int) (*DespesaStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_despesas,
			COALESCE(SUM(valor_liquido), 0) as total_valor,
			COALESCE(AVG(valor_liquido), 0) as media_valor,
			COALESCE(MAX(valor_liquido), 0) as maior_valor,
			COUNT(DISTINCT tipo_despesa) as tipos_distintos
		FROM despesas_cache 
		WHERE deputado_id = $1 AND ano = $2
	`

	var stats DespesaStats
	err := r.db.QueryRow(ctx, query, deputadoID, ano).Scan(
		&stats.TotalDespesas,
		&stats.TotalValor,
		&stats.ValorMedio,
		&stats.MaiorValor,
		&stats.TiposDiferentes,
	)

	if err != nil {
		return nil, fmt.Errorf("erro ao obter estatísticas: %w", err)
	}

	return &stats, nil
}

// Interface compatibility methods
func (r *OptimizedDespesaRepository) UpsertDespesas(ctx context.Context, deputadoID int, ano int, despesas []domain.Despesa) error {
	return r.UpsertDespesasBatch(ctx, deputadoID, ano, despesas)
}

func (r *OptimizedDespesaRepository) ListDespesasByDeputadoAno(ctx context.Context, deputadoID int, ano int) ([]domain.Despesa, error) {
	return r.ListDespesasOptimized(ctx, deputadoID, ano)
}
