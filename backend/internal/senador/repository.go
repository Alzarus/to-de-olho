package senador

import (
	"gorm.io/gorm"
)

// Repository encapsula operacoes de banco de dados para Senador
type Repository struct {
	db *gorm.DB
}

// NewRepository cria um novo repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// FindAll retorna todos os senadores em exercicio
func (r *Repository) FindAll() ([]Senador, error) {
	var senadores []Senador
	result := r.db.Where("em_exercicio = ?", true).
		Order("nome ASC").
		Find(&senadores)
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
	return r.db.Save(senador).Error
}

// UpsertBatch insere ou atualiza multiplos senadores
func (r *Repository) UpsertBatch(senadores []Senador) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, s := range senadores {
			if err := tx.Save(&s).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Count retorna o total de senadores em exercicio
func (r *Repository) Count() (int64, error) {
	var count int64
	result := r.db.Model(&Senador{}).Where("em_exercicio = ?", true).Count(&count)
	return count, result.Error
}
