package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"to-de-olho-backend/internal/domain"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DespesaRepository struct {
	db DB
}

func NewDespesaRepository(db *pgxpool.Pool) *DespesaRepository {
	return &DespesaRepository{db: db}
}

// UpsertDespesas insere ou atualiza despesas no cache
func (r *DespesaRepository) UpsertDespesas(ctx context.Context, deputadoID int, ano int, despesas []domain.Despesa) error {
	if r == nil || r.db == nil || len(despesas) == 0 {
		return nil
	}

	for _, d := range despesas {
		payload, err := json.Marshal(d)
		if err != nil {
			return fmt.Errorf("erro ao serializar despesa: %w", err)
		}

		// Primeiro, tentamos inserir; se houver conflito, fazemos update
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
		` // Parse data_documento se não estiver vazia
		var dataDocumento interface{}
		if d.DataDocumento != "" {
			// Assumindo formato YYYY-MM-DD da API da Câmara
			if parsedDate, err := time.Parse("2006-01-02", d.DataDocumento); err == nil {
				dataDocumento = parsedDate
			} else {
				dataDocumento = nil // ou podemos logar o erro
			}
		} else {
			dataDocumento = nil
		}

		_, err = r.db.Exec(ctx, query,
			deputadoID, d.Ano, d.Mes, d.TipoDespesa, d.CodDocumento,
			d.TipoDocumento, d.CodTipoDocumento, dataDocumento, d.NumDocumento,
			d.ValorDocumento, d.URLDocumento, d.NomeFornecedor, d.CNPJCPFFornecedor,
			d.ValorLiquido, d.ValorBruto, d.ValorGlosa, d.NumRessarcimento,
			d.CodLote, d.Parcela, string(payload),
		)

		if err != nil {
			// Se foi erro de constraint única, ignoramos (despesa já existe)
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" { // unique_violation
				continue // Ignora duplicatas
			}
			return fmt.Errorf("erro ao inserir despesa do deputado %d: %w", deputadoID, err)
		}
	}

	return nil
}

// ListDespesasByDeputadoAno busca despesas de um deputado em um ano específico
func (r *DespesaRepository) ListDespesasByDeputadoAno(ctx context.Context, deputadoID int, ano int) ([]domain.Despesa, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}

	query := `
		SELECT payload 
		FROM despesas_cache 
		WHERE deputado_id = $1 AND ano = $2 
		ORDER BY data_documento DESC, valor_documento DESC
	`

	rows, err := r.db.Query(ctx, query, deputadoID, ano)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar despesas do deputado %d no ano %d: %w", deputadoID, ano, err)
	}
	defer rows.Close()

	var despesas []domain.Despesa
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, fmt.Errorf("erro ao ler payload: %w", err)
		}

		var despesa domain.Despesa
		if err := json.Unmarshal([]byte(payload), &despesa); err != nil {
			return nil, fmt.Errorf("erro ao decodificar despesa: %w", err)
		}

		despesas = append(despesas, despesa)
	}

	return despesas, nil
}

// GetDespesasStats retorna estatísticas de despesas de um deputado
func (r *DespesaRepository) GetDespesasStats(ctx context.Context, deputadoID int, ano int) (*DespesaStats, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}

	query := `
		SELECT 
			COUNT(*) as total_despesas,
			COALESCE(SUM(valor_liquido), 0) as total_valor,
			COALESCE(AVG(valor_liquido), 0) as valor_medio,
			COALESCE(MAX(valor_liquido), 0) as maior_valor,
			COUNT(DISTINCT tipo_despesa) as tipos_diferentes
		FROM despesas_cache 
		WHERE deputado_id = $1 AND ano = $2
	`

	row := r.db.QueryRow(ctx, query, deputadoID, ano)

	var stats DespesaStats
	err := row.Scan(
		&stats.TotalDespesas,
		&stats.TotalValor,
		&stats.ValorMedio,
		&stats.MaiorValor,
		&stats.TiposDiferentes,
	)

	if err != nil {
		return nil, fmt.Errorf("erro ao buscar estatísticas de despesas: %w", err)
	}

	return &stats, nil
}

// DespesaStats representa estatísticas de despesas
type DespesaStats struct {
	TotalDespesas   int     `json:"total_despesas"`
	TotalValor      float64 `json:"total_valor"`
	ValorMedio      float64 `json:"valor_medio"`
	MaiorValor      float64 `json:"maior_valor"`
	TiposDiferentes int     `json:"tipos_diferentes"`
}
