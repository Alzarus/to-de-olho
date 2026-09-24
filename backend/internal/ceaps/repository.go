package ceaps

import (
	"time"

	"gorm.io/gorm"
)

// Repository encapsula operacoes de banco de dados para DespesaCEAPS
type Repository struct {
	db *gorm.DB
}

// NewRepository cria um novo repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// FindBySenadorID retorna despesas de um senador com paginacao, busca e filtros
func (r *Repository) FindBySenadorID(senadorID int, ano *int, limit int, offset int, queryStr string, tipo string, sort string) ([]DespesaCEAPS, int64, error) {
	var despesas []DespesaCEAPS
	var total int64

	dbQuery := r.db.Model(&DespesaCEAPS{}).Where("senador_id = ?", senadorID)

	if ano != nil {
		dbQuery = dbQuery.Where("ano = ?", *ano)
	}

	if tipo != "" && tipo != "todos" {
		dbQuery = dbQuery.Where("tipo_despesa = ?", tipo)
	}

	if queryStr != "" {
		search := "%" + queryStr + "%"
		dbQuery = dbQuery.Where("(fornecedor ILIKE ? OR tipo_despesa ILIKE ?)", search, search)
	}

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sorting
	// Default: Date DESC
	order := "ano DESC, mes DESC, data_emissao DESC"
	switch sort {
	case "data_asc":
		order = "ano ASC, mes ASC, data_emissao ASC"
	case "valor_desc":
		order = "valor DESC"
	case "valor_asc":
		order = "valor ASC"
	case "fornecedor_asc":
		order = "fornecedor ASC"
	case "fornecedor_desc":
		order = "fornecedor DESC"
	}

	result := dbQuery.Order(order).
		Limit(limit).
		Offset(offset).
		Find(&despesas)
		
	return despesas, total, result.Error
}

// AggregateByTipo retorna gastos agregados por tipo de despesa
func (r *Repository) AggregateByTipo(senadorID int, ano *int) ([]AggregatedDespesa, error) {
	var result []AggregatedDespesa

	query := r.db.Model(&DespesaCEAPS{}).
		Select("tipo_despesa, SUM(valor) as total, COUNT(*) as quantidade").
		Where("senador_id = ?", senadorID).
		Group("tipo_despesa").
		Order("total DESC")

	if ano != nil {
		query = query.Where("ano = ?", *ano)
	}

	err := query.Scan(&result).Error
	return result, err
}

// GastoMensal soma as despesas por mes de competencia. Sem ano, traz todos os
// meses do banco (mesmo recorte do agregado por tipo).
func (r *Repository) GastoMensal(senadorID int, ano *int) ([]SenadorGastoMensal, error) {
	var result []SenadorGastoMensal

	query := r.db.Model(&DespesaCEAPS{}).
		Select("ano, mes, SUM(valor) as total").
		Where("senador_id = ?", senadorID).
		Group("ano, mes").
		Order("ano ASC, mes ASC")
	if ano != nil {
		query = query.Where("ano = ?", *ano)
	}

	err := query.Scan(&result).Error
	return result, err
}

// Fornecedores soma as despesas por fornecedor, do maior para o menor. O
// fornecedor e identificado pelo CNPJ/CPF; sem documento, pelo nome.
func (r *Repository) Fornecedores(senadorID int, ano *int) ([]FornecedorAgregado, error) {
	var result []FornecedorAgregado

	query := r.db.Model(&DespesaCEAPS{}).
		Select(`MAX(fornecedor) AS fornecedor, COALESCE(NULLIF(cnpj_cpf, ''), '') AS cnpj_cpf,
			SUM(valor) AS total, COUNT(*) AS quantidade`).
		Where("senador_id = ?", senadorID).
		Group("COALESCE(NULLIF(cnpj_cpf, ''), ''), CASE WHEN COALESCE(cnpj_cpf, '') = '' THEN fornecedor END").
		Order("total DESC")
	if ano != nil {
		query = query.Where("ano = ?", *ano)
	}

	err := query.Scan(&result).Error
	return result, err
}

// GetTotalByAno retorna total gasto por um senador em um ano
func (r *Repository) GetTotalByAno(senadorID int, ano int) (float64, error) {
	var total float64
	err := r.db.Model(&DespesaCEAPS{}).
		Select("COALESCE(SUM(valor), 0)").
		Where("senador_id = ? AND ano = ?", senadorID, ano).
		Scan(&total).Error
	return total, err
}

// DeleteByAno remove todas as despesas de um determinado ano
func (r *Repository) DeleteByAno(ano int) error {
	return r.db.Where("ano = ?", ano).Delete(&DespesaCEAPS{}).Error
}

// GetTotalPeriodo soma as despesas com competencia (ano, mes) nos meses que
// tocam [inicio, fim). E a fonte do criterio de economia: mesmo periodo das
// votacoes e proposicoes (janeiro de 2023 e da legislatura anterior e fica de
// fora do mandato).
func (r *Repository) GetTotalPeriodo(senadorID int, inicio, fim time.Time) (float64, error) {
	var total float64
	ultimo := fim.Add(-time.Nanosecond)
	err := r.db.Model(&DespesaCEAPS{}).
		Select("COALESCE(SUM(valor), 0)").
		Where("senador_id = ? AND ano * 100 + mes BETWEEN ? AND ?", senadorID,
			inicio.Year()*100+int(inicio.Month()), ultimo.Year()*100+int(ultimo.Month())).
		Scan(&total).Error
	return total, err
}

// SubstituirAno troca todas as despesas do ano numa transacao
func (r *Repository) SubstituirAno(ano int, despesas []DespesaCEAPS) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("ano = ?", ano).Delete(&DespesaCEAPS{}).Error; err != nil {
			return err
		}
		if len(despesas) == 0 {
			return nil
		}
		return tx.CreateInBatches(despesas, 1000).Error
	})
}
