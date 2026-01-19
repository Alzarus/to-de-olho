package ceaps

import (
	"fmt"
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

// GetGastoMensal retorna evolucao mensal de gastos
func (r *Repository) GetGastoMensal(senadorID int, ano int) ([]SenadorGastoMensal, error) {
	var result []SenadorGastoMensal

	err := r.db.Model(&DespesaCEAPS{}).
		Select("ano, mes, SUM(valor) as total").
		Where("senador_id = ? AND ano = ?", senadorID, ano).
		Group("ano, mes").
		Order("mes ASC").
		Scan(&result).Error

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

// GetTotal retorna total gasto por um senador em todo o mandato (todas os anos)
func (r *Repository) GetTotal(senadorID int) (float64, error) {
	var total float64
	err := r.db.Model(&DespesaCEAPS{}).
		Select("COALESCE(SUM(valor), 0)").
		Where("senador_id = ?", senadorID).
		Scan(&total).Error
	fmt.Printf("[DEBUG] GetTotal SenadorID=%d Total=%f Err=%v\n", senadorID, total, err)
	return total, err
}

// Upsert insere ou atualiza uma despesa usando chave composta
func (r *Repository) Upsert(despesa *DespesaCEAPS) error {
	_ = despesa.BeforeCreate(nil)
	return r.db.Where("senador_id = ? AND cnpj_cpf = ? AND data_emissao = ? AND valor_centavos = ?",
		despesa.SenadorID, despesa.CNPJCPF, despesa.DataEmissao, despesa.ValorCentavos).
		Assign(*despesa).FirstOrCreate(despesa).Error
}

// DeleteByAno remove todas as despesas de um determinado ano
func (r *Repository) DeleteByAno(ano int) error {
	return r.db.Where("ano = ?", ano).Delete(&DespesaCEAPS{}).Error
}
