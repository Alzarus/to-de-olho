package proposicao

import (
	"gorm.io/gorm"
)

// Repository encapsula operacoes de banco de dados para Proposicao
type Repository struct {
	db *gorm.DB
}

// NewRepository cria um novo repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// FindBySenadorID retorna proposicoes de um senador
func (r *Repository) FindBySenadorID(senadorID int, limit int) ([]Proposicao, error) {
	var proposicoes []Proposicao
	query := r.db.Where("senador_id = ?", senadorID).Order("data_apresentacao DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	result := query.Find(&proposicoes)
	return proposicoes, result.Error
}

// CountBySenadorID retorna total de proposicoes de um senador
func (r *Repository) CountBySenadorID(senadorID int) (int64, error) {
	var count int64
	result := r.db.Model(&Proposicao{}).Where("senador_id = ?", senadorID).Count(&count)
	return count, result.Error
}

// GetStats retorna estatisticas de proposicoes de um senador
func (r *Repository) GetStats(senadorID int) (*ProposicaoStats, error) {
	var stats ProposicaoStats
	stats.SenadorID = senadorID

	var total, pecs, plps, pls, leis, plenario, tramitacao int64
	var pontuacaoTotal int

	// Total de proposicoes
	r.db.Model(&Proposicao{}).Where("senador_id = ?", senadorID).Count(&total)
	stats.TotalProposicoes = int(total)

	// PECs
	r.db.Model(&Proposicao{}).Where("senador_id = ? AND sigla_subtipo_materia = ?", senadorID, "PEC").Count(&pecs)
	stats.TotalPECs = int(pecs)

	// PLPs
	r.db.Model(&Proposicao{}).Where("senador_id = ? AND sigla_subtipo_materia = ?", senadorID, "PLP").Count(&plps)
	stats.TotalPLPs = int(plps)

	// PLs
	r.db.Model(&Proposicao{}).Where("senador_id = ? AND sigla_subtipo_materia = ?", senadorID, "PL").Count(&pls)
	stats.TotalPLs = int(pls)

	// Outros
	stats.TotalOutros = stats.TotalProposicoes - stats.TotalPECs - stats.TotalPLPs - stats.TotalPLs

	// Transformadas em lei
	r.db.Model(&Proposicao{}).Where(
		"senador_id = ? AND estagio_tramitacao = ?", senadorID, "TransformadoLei",
	).Count(&leis)
	stats.TransformadasEmLei = int(leis)

	// Aprovados plenario
	r.db.Model(&Proposicao{}).Where(
		"senador_id = ? AND estagio_tramitacao IN (?, ?)", senadorID, "AprovadoPlenario", "TransformadoLei",
	).Count(&plenario)
	stats.AprovadosPlenario = int(plenario)

	// Em tramitacao
	r.db.Model(&Proposicao{}).Where(
		"senador_id = ? AND estagio_tramitacao IN (?, ?, ?)", senadorID, "Apresentado", "EmComissao", "AprovadoComissao",
	).Count(&tramitacao)
	stats.EmTramitacao = int(tramitacao)

	// Soma de pontuacao
	var soma struct {
		Total int
	}
	r.db.Model(&Proposicao{}).
		Select("COALESCE(SUM(pontuacao), 0) as total").
		Where("senador_id = ?", senadorID).
		Scan(&soma)
	pontuacaoTotal = soma.Total
	stats.PontuacaoTotal = pontuacaoTotal

	return &stats, nil
}

// GetProposicoesPorTipo retorna contagem de proposicoes por tipo
func (r *Repository) GetProposicoesPorTipo(senadorID int) ([]ProposicaoPorTipo, error) {
	var result []ProposicaoPorTipo
	err := r.db.Model(&Proposicao{}).
		Select("sigla_subtipo_materia as tipo, COUNT(*) as total").
		Where("senador_id = ?", senadorID).
		Group("sigla_subtipo_materia").
		Order("total DESC").
		Scan(&result).Error
	return result, err
}

// Upsert insere ou atualiza uma proposicao
func (r *Repository) Upsert(proposicao *Proposicao) error {
	return r.db.Save(proposicao).Error
}

// UpsertBatch insere ou atualiza multiplas proposicoes
func (r *Repository) UpsertBatch(proposicoes []Proposicao) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, p := range proposicoes {
			if err := tx.Save(&p).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteBySenadorID remove todas as proposicoes de um senador
func (r *Repository) DeleteBySenadorID(senadorID int) error {
	return r.db.Where("senador_id = ?", senadorID).Delete(&Proposicao{}).Error
}
