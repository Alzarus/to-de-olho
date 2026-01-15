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

// FindBySenadorID retorna despesas de um senador
func (r *Repository) FindBySenadorID(senadorID int, ano *int) ([]DespesaCEAPS, error) {
	var despesas []DespesaCEAPS
	query := r.db.Where("senador_id = ?", senadorID)

	if ano != nil {
		query = query.Where("ano = ?", *ano)
	}

	result := query.Order("ano DESC, mes DESC").Find(&despesas)
	return despesas, result.Error
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

// Upsert insere ou atualiza uma despesa
func (r *Repository) Upsert(despesa *DespesaCEAPS) error {
	despesa.BeforeCreate()
	return r.db.Save(despesa).Error
}
