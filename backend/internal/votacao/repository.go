package votacao

import (
	"gorm.io/gorm"
)

// Repository encapsula operacoes de banco de dados para Votacao
type Repository struct {
	db *gorm.DB
}

// NewRepository cria um novo repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// FindBySenadorID retorna votacoes de um senador
func (r *Repository) FindBySenadorID(senadorID int, limit int) ([]Votacao, error) {
	var votacoes []Votacao
	query := r.db.Where("senador_id = ?", senadorID).Order("data DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	result := query.Find(&votacoes)
	return votacoes, result.Error
}

// CountBySenadorID retorna total de votacoes de um senador
func (r *Repository) CountBySenadorID(senadorID int) (int64, error) {
	var count int64
	result := r.db.Model(&Votacao{}).Where("senador_id = ?", senadorID).Count(&count)
	return count, result.Error
}

// GetStats retorna estatisticas de votacao de um senador
func (r *Repository) GetStats(senadorID int) (*VotacaoStats, error) {
	var stats VotacaoStats
	stats.SenadorID = senadorID

	var total, registrados, ausencias, obstrucoes int64

	// Total de votacoes
	r.db.Model(&Votacao{}).Where("senador_id = ?", senadorID).Count(&total)
	stats.TotalVotacoes = int(total)

	// Votos registrados (Sim, Nao, Abstencao)
	r.db.Model(&Votacao{}).Where(
		"senador_id = ? AND voto IN (?, ?, ?)", senadorID, "Sim", "Nao", "Abstencao",
	).Count(&registrados)
	stats.VotosRegistrados = int(registrados)

	// Ausencias (NCom)
	r.db.Model(&Votacao{}).Where(
		"senador_id = ? AND voto = ?", senadorID, "NCom",
	).Count(&ausencias)
	stats.Ausencias = int(ausencias)

	// Obstrucoes
	r.db.Model(&Votacao{}).Where(
		"senador_id = ? AND voto = ?", senadorID, "Obstrucao",
	).Count(&obstrucoes)
	stats.Obstrucoes = int(obstrucoes)

	// Calcular taxas
	if stats.TotalVotacoes > 0 {
		// Presenca = (Total - Ausencias) / Total
		stats.TaxaPresenca = float64(stats.TotalVotacoes-stats.Ausencias) / float64(stats.TotalVotacoes) * 100

		// Participacao = Votos efetivos / Total
		stats.TaxaParticipacao = float64(stats.VotosRegistrados) / float64(stats.TotalVotacoes) * 100
	}

	return &stats, nil
}

// GetVotosPorTipo retorna contagem de votos por tipo
func (r *Repository) GetVotosPorTipo(senadorID int) ([]VotosPorTipo, error) {
	var result []VotosPorTipo
	err := r.db.Model(&Votacao{}).
		Select("voto, COUNT(*) as total").
		Where("senador_id = ?", senadorID).
		Group("voto").
		Order("total DESC").
		Scan(&result).Error
	return result, err
}

// Upsert insere ou atualiza uma votacao
func (r *Repository) Upsert(votacao *Votacao) error {
	return r.db.Save(votacao).Error
}

// UpsertBatch insere ou atualiza multiplas votacoes
func (r *Repository) UpsertBatch(votacoes []Votacao) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, v := range votacoes {
			if err := tx.Save(&v).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
