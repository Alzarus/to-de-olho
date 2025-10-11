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

// DespesaRepository implementa operações otimizadas para despesas
type DespesaRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewDespesaRepository(db *pgxpool.Pool) *DespesaRepository {
	logger := slog.Default()

	repo := &DespesaRepository{
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
func (r *DespesaRepository) prepareStatements(ctx context.Context) error {
	// Note: pgx v5 não suporta prepared statements globais no pool
	// Em vez disso, usaremos QueryRow/Query diretamente que são otimizadas pelo pgx
	r.logger.Info("query optimization ready for batch operations and connection pooling")
	return nil
}

// UpsertDespesas insere/atualiza despesas em lote para máxima performance
func (r *DespesaRepository) UpsertDespesas(ctx context.Context, deputadoID int, ano int, despesas []domain.Despesa) error {
	if len(despesas) == 0 {
		return nil
	}

	start := time.Now()

	// Usar transação para batch
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	// Preparar batch com CopyFrom para máxima performance
	now := time.Now()
	rows := make([][]interface{}, len(despesas))
	for i, d := range despesas {
		payload, _ := json.Marshal(d)

		var dataDocumento interface{}
		if d.DataDocumento != "" {
			if parsedDate, err := time.Parse("2006-01-02", d.DataDocumento); err == nil {
				dataDocumento = parsedDate
			}
		}

		valorDocumento := d.ValorDocumento
		if valorDocumento <= 0 && d.ValorLiquido > 0 {
			valorDocumento = d.ValorLiquido
		}

		valorLiquido := d.ValorLiquido
		if valorLiquido <= 0 && valorDocumento > 0 {
			valorLiquido = valorDocumento
		}

		rows[i] = []interface{}{
			deputadoID, d.Ano, d.Mes, d.TipoDespesa, d.CodDocumento,
			d.TipoDocumento, d.CodTipoDocumento, dataDocumento, d.NumDocumento,
			valorDocumento, d.URLDocumento, d.NomeFornecedor, d.CNPJCPFFornecedor,
			valorLiquido, d.ValorBruto, d.ValorGlosa, d.NumRessarcimento,
			d.CodLote, d.Parcela, string(payload), now,
		}
	}

	// Usar COPY para inserção em massa ultra-rápida
	copyCount, err := tx.CopyFrom(ctx,
		pgx.Identifier{"despesas"},
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
		// Se COPY falhar (por conflitos), fazer rollback e criar nova transação
		r.logger.Warn("CopyFrom failed, rolling back and retrying with individual upserts",
			slog.String("error", err.Error()),
			slog.Int("deputado_id", deputadoID),
			slog.Int("total_despesas", len(despesas)))

		// CRITICAL: Rollback da transação abortada
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			r.logger.Error("erro ao fazer rollback da transação abortada",
				slog.String("error", rollbackErr.Error()))
		}
		committed = true // transação original já finalizada

		// CRITICAL: Criar nova transação para upsert individual
		newTx, err := r.db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("erro ao iniciar nova transação para fallback: %w", err)
		}
		fallbackCommitted := false
		defer func() {
			if !fallbackCommitted {
				_ = newTx.Rollback(ctx)
			}
		}()

		if err := r.upsertIndividual(ctx, newTx, deputadoID, ano, despesas); err != nil {
			return fmt.Errorf("erro no upsert individual: %w", err)
		}

		// Commit da nova transação após upsert individual bem-sucedido
		if err := newTx.Commit(ctx); err != nil {
			return fmt.Errorf("erro ao fazer commit da transação (fallback): %w", err)
		}
		fallbackCommitted = true

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
	committed = true

	r.logger.Info("despesas inseridas em lote",
		slog.Int("deputado_id", deputadoID),
		slog.Int("ano", ano),
		slog.Int64("inserted_count", copyCount),
		slog.Duration("duration", time.Since(start)))

	return nil
}

// upsertIndividual faz upsert individual quando batch falha
func (r *DespesaRepository) upsertIndividual(ctx context.Context, tx pgx.Tx, deputadoID int, ano int, despesas []domain.Despesa) error {
	query := `
		INSERT INTO despesas (
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
			tipo_documento = EXCLUDED.tipo_documento,
			cod_tipo_documento = EXCLUDED.cod_tipo_documento,
			data_documento = EXCLUDED.data_documento,
			num_documento = EXCLUDED.num_documento,
			valor_documento = EXCLUDED.valor_documento,
			valor_liquido = EXCLUDED.valor_liquido,
			valor_bruto = EXCLUDED.valor_bruto,
			valor_glosa = EXCLUDED.valor_glosa,
			nome_fornecedor = EXCLUDED.nome_fornecedor,
			cnpj_cpf_fornecedor = EXCLUDED.cnpj_cpf_fornecedor,
			url_documento = EXCLUDED.url_documento,
			num_ressarcimento = EXCLUDED.num_ressarcimento,
			cod_lote = EXCLUDED.cod_lote,
			parcela = EXCLUDED.parcela,
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

		valorDocumento := d.ValorDocumento
		if valorDocumento <= 0 && d.ValorLiquido > 0 {
			valorDocumento = d.ValorLiquido
		}

		valorLiquido := d.ValorLiquido
		if valorLiquido <= 0 && valorDocumento > 0 {
			valorLiquido = valorDocumento
		}

		batch.Queue(query,
			deputadoID, d.Ano, d.Mes, d.TipoDespesa, d.CodDocumento,
			d.TipoDocumento, d.CodTipoDocumento, dataDocumento, d.NumDocumento,
			valorDocumento, d.URLDocumento, d.NomeFornecedor, d.CNPJCPFFornecedor,
			valorLiquido, d.ValorBruto, d.ValorGlosa, d.NumRessarcimento,
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

// ListDespesasByDeputadoAno lista despesas com query otimizada
func (r *DespesaRepository) ListDespesasByDeputadoAno(ctx context.Context, deputadoID int, ano int) ([]domain.Despesa, error) {
	start := time.Now()

	query := `
		SELECT 
			tipo_despesa, cod_documento, tipo_documento, COALESCE(cod_tipo_documento, 0),
			data_documento, num_documento, valor_documento, url_documento,
			nome_fornecedor, cnpj_cpf_fornecedor, valor_liquido, valor_bruto,
			valor_glosa, num_ressarcimento, cod_lote, parcela
		FROM despesas 
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

// GetDespesasStats obtém estatísticas agregadas de forma otimizada
func (r *DespesaRepository) GetDespesasStats(ctx context.Context, deputadoID int, ano int) (*domain.DespesaStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_despesas,
			COALESCE(SUM(valor_liquido), 0) as total_valor,
			COALESCE(AVG(valor_liquido), 0) as media_valor,
			COALESCE(MAX(valor_liquido), 0) as maior_valor,
			COUNT(DISTINCT tipo_despesa) as tipos_distintos
		FROM despesas 
		WHERE deputado_id = $1 AND ano = $2
	`

	var stats domain.DespesaStats
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

// GetDespesasStatsByAno agrega estatísticas de despesas por deputado para o ano informado
func (r *DespesaRepository) GetDespesasStatsByAno(ctx context.Context, ano int) (map[int]domain.DespesaStats, error) {
	query := `
		SELECT 
			deputado_id,
			COUNT(*) as total_despesas,
			COALESCE(SUM(valor_liquido), 0) as total_valor,
			COALESCE(AVG(valor_liquido), 0) as media_valor,
			COALESCE(MAX(valor_liquido), 0) as maior_valor,
			COUNT(DISTINCT tipo_despesa) as tipos_distintos
		FROM despesas
		WHERE ano = $1
		GROUP BY deputado_id`

	rows, err := r.db.Query(ctx, query, ano)
	if err != nil {
		return nil, fmt.Errorf("erro ao agregar estatísticas de despesas: %w", err)
	}
	defer rows.Close()

	result := make(map[int]domain.DespesaStats)
	for rows.Next() {
		var deputadoID int
		var stats domain.DespesaStats
		if err := rows.Scan(&deputadoID, &stats.TotalDespesas, &stats.TotalValor, &stats.ValorMedio, &stats.MaiorValor, &stats.TiposDiferentes); err != nil {
			return nil, fmt.Errorf("erro ao escanear estatísticas de despesas: %w", err)
		}
		result[deputadoID] = stats
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar estatísticas de despesas: %w", err)
	}

	return result, nil
}
