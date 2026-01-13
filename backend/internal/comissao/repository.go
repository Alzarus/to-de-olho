package comissao

import (
	"gorm.io/gorm"
)

// Repository encapsula operacoes de banco de dados para ComissaoMembro
type Repository struct {
	db *gorm.DB
}

// NewRepository cria um novo repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// FindBySenadorID retorna comissoes de um senador
func (r *Repository) FindBySenadorID(senadorID int, limit int) ([]ComissaoMembro, error) {
	var comissoes []ComissaoMembro
	query := r.db.Where("senador_id = ?", senadorID).Order("data_inicio DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	result := query.Find(&comissoes)
	return comissoes, result.Error
}

// FindAtivasBySenadorID retorna comissoes ativas de um senador (sem data_fim)
func (r *Repository) FindAtivasBySenadorID(senadorID int) ([]ComissaoMembro, error) {
	var comissoes []ComissaoMembro
	result := r.db.Where("senador_id = ? AND data_fim IS NULL", senadorID).
		Order("data_inicio DESC").
		Find(&comissoes)
	return comissoes, result.Error
}

// CountBySenadorID retorna total de comissoes de um senador
func (r *Repository) CountBySenadorID(senadorID int) (int64, error) {
	var count int64
	result := r.db.Model(&ComissaoMembro{}).Where("senador_id = ?", senadorID).Count(&count)
	return count, result.Error
}

// GetStats retorna estatisticas de comissoes de um senador
func (r *Repository) GetStats(senadorID int) (*ComissaoStats, error) {
	var stats ComissaoStats
	stats.SenadorID = senadorID

	var total, titular, suplente, ativas int64

	// Total de participacoes
	r.db.Model(&ComissaoMembro{}).Where("senador_id = ?", senadorID).Count(&total)
	stats.TotalComissoes = int(total)

	// Titular
	r.db.Model(&ComissaoMembro{}).Where(
		"senador_id = ? AND descricao_participacao = ?", senadorID, "Titular",
	).Count(&titular)
	stats.ComissoesTitular = int(titular)

	// Suplente
	r.db.Model(&ComissaoMembro{}).Where(
		"senador_id = ? AND descricao_participacao = ?", senadorID, "Suplente",
	).Count(&suplente)
	stats.ComissoesSuplente = int(suplente)

	// Ativas (sem data_fim)
	r.db.Model(&ComissaoMembro{}).Where(
		"senador_id = ? AND data_fim IS NULL", senadorID,
	).Count(&ativas)
	stats.ComissoesAtivas = int(ativas)

	// Calcular taxa de titularidade
	if stats.TotalComissoes > 0 {
		stats.TaxaTitularidade = float64(stats.ComissoesTitular) / float64(stats.TotalComissoes) * 100
	}

	return &stats, nil
}

// GetComissoesPorCasa retorna contagem de comissoes por casa
func (r *Repository) GetComissoesPorCasa(senadorID int) ([]ComissoesPorCasa, error) {
	var result []ComissoesPorCasa
	err := r.db.Model(&ComissaoMembro{}).
		Select("sigla_casa_comissao as casa, COUNT(*) as total").
		Where("senador_id = ?", senadorID).
		Group("sigla_casa_comissao").
		Order("total DESC").
		Scan(&result).Error
	return result, err
}

// Upsert insere ou atualiza uma comissao
func (r *Repository) Upsert(comissao *ComissaoMembro) error {
	return r.db.Save(comissao).Error
}

// UpsertBatch insere ou atualiza multiplas comissoes
func (r *Repository) UpsertBatch(comissoes []ComissaoMembro) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, c := range comissoes {
			if err := tx.Save(&c).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteBySenadorID remove todas as comissoes de um senador
// Usado antes de re-sincronizar para evitar duplicatas
func (r *Repository) DeleteBySenadorID(senadorID int) error {
	return r.db.Where("senador_id = ?", senadorID).Delete(&ComissaoMembro{}).Error
}
