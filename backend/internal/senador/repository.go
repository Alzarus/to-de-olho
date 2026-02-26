package senador

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository encapsula operacoes de banco de dados para Senador
type Repository struct {
	db *gorm.DB
}

// NewRepository cria um novo repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// FindAll retorna senadores (pode incluir inativos)
func (r *Repository) FindAll(includeInactive bool) ([]Senador, error) {
	var senadores []Senador
	query := r.db.Order("nome ASC")
	if !includeInactive {
		query = query.Where("em_exercicio = ?", true)
	}
	result := query.Find(&senadores)
	return senadores, result.Error
}

// FindByID busca senador por ID interno
func (r *Repository) FindByID(id int) (*Senador, error) {
	var senador Senador
	result := r.db.Preload("Mandatos").First(&senador, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &senador, nil
}

// FindByCodigo busca senador por codigo parlamentar
func (r *Repository) FindByCodigo(codigo int) (*Senador, error) {
	var senador Senador
	result := r.db.Preload("Mandatos").
		Where("codigo_parlamentar = ?", codigo).
		First(&senador)
	if result.Error != nil {
		return nil, result.Error
	}
	return &senador, nil
}

// Upsert insere ou atualiza um senador
func (r *Repository) Upsert(senador *Senador) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "codigo_parlamentar"}},
		DoUpdates: clause.AssignmentColumns([]string{"nome", "nome_completo", "partido", "uf", "foto_url", "email", "cargo", "titular", "em_exercicio", "updated_at"}),
	}).Create(senador).Error
}

// UpsertBatch insere ou atualiza multiplos senadores
func (r *Repository) UpsertBatch(senadores []Senador) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "codigo_parlamentar"}},
		DoUpdates: clause.AssignmentColumns([]string{"nome", "nome_completo", "partido", "uf", "foto_url", "email", "cargo", "titular", "em_exercicio", "updated_at"}),
	}).CreateInBatches(senadores, 100).Error
}

// SetInactive marca senadores não listados como fora de exercicio
func (r *Repository) SetInactive(activeCodes []int) error {
	if len(activeCodes) == 0 {
		return nil
	}
	return r.db.Model(&Senador{}).
		Where("codigo_parlamentar NOT IN ?", activeCodes).
		Update("em_exercicio", false).Error
}

// Count retorna o total de senadores em exercicio
func (r *Repository) Count() (int64, error) {
	var count int64
	result := r.db.Model(&Senador{}).Where("em_exercicio = ?", true).Count(&count)
	return count, result.Error
}
