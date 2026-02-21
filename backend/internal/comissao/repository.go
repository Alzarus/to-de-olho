package comissao

import (
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository encapsula operacoes de banco de dados para ComissaoMembro
type Repository struct {
	db *gorm.DB
}

// NewRepository cria um novo repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// FindBySenadorID retorna comissoes de um senador com paginacao, busca e filtros
func (r *Repository) FindBySenadorID(senadorID int, limit int, offset int, queryStr string, status string, participacao string) ([]ComissaoMembro, int64, error) {
	var comissoes []ComissaoMembro
	var total int64
	
	dbQuery := r.db.Model(&ComissaoMembro{}).Where("senador_id = ?", senadorID)

	if queryStr != "" {
		search := "%" + queryStr + "%"
		dbQuery = dbQuery.Where("(nome_comissao ILIKE ? OR descricao_participacao ILIKE ? OR sigla_comissao ILIKE ?)", search, search, search)
	}

	// Status: ativa | inativa
	if status == "ativa" {
		dbQuery = dbQuery.Where("data_fim IS NULL")
	} else if status == "inativa" {
		dbQuery = dbQuery.Where("data_fim IS NOT NULL")
	}

	// Participacao: Titular | Suplente
	if participacao != "" && participacao != "todos" {
		dbQuery = dbQuery.Where("descricao_participacao = ?", participacao)
	}

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	dbQuery = dbQuery.Order("data_inicio DESC")

	if limit > 0 {
		dbQuery = dbQuery.Limit(limit)
	}
	if offset > 0 {
		dbQuery = dbQuery.Offset(offset)
	}

	result := dbQuery.Find(&comissoes)
	return comissoes, total, result.Error
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

// Upsert insere ou atualiza uma comissao usando chave composta (senador_id, codigo_comissao)
func (r *Repository) Upsert(comissao *ComissaoMembro) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "senador_id"}, {Name: "codigo_comissao"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"sigla_comissao", "nome_comissao", "sigla_casa_comissao",
			"descricao_participacao", "data_inicio", "data_fim", "updated_at",
		}),
	}).Create(comissao).Error
}

// UpsertBatch insere ou atualiza multiplas comissoes
func (r *Repository) UpsertBatch(comissoes []ComissaoMembro) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "senador_id"}, {Name: "codigo_comissao"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"sigla_comissao", "nome_comissao", "sigla_casa_comissao",
			"descricao_participacao", "data_inicio", "data_fim", "updated_at",
		}),
	}).CreateInBatches(comissoes, 100).Error
}

// DeleteBySenadorID remove todas as comissoes de um senador
// Usado antes de re-sincronizar para evitar duplicatas
func (r *Repository) DeleteBySenadorID(senadorID int) error {
	return r.db.Where("senador_id = ?", senadorID).Delete(&ComissaoMembro{}).Error
}

// GetStatsByAno retorna estatisticas de comissoes filtradas por ano
func (r *Repository) GetStatsByAno(senadorID int, ano int) (*ComissaoStats, error) {
	var stats ComissaoStats
	stats.SenadorID = senadorID

	var total, titular, suplente int64

	// Filtro de participacao no ano:
	// data_inicio < inicio_proximo_ano AND (data_fim >= inicio_ano OR data_fim IS NULL)
	dataInicioAno := fmt.Sprintf("%d-01-01", ano)
	dataProximoAno := fmt.Sprintf("%d-01-01", ano+1)

	// Criterio de filtro para "ativo durante o ano"
	dateFilter := "data_inicio < ? AND (data_fim >= ? OR data_fim IS NULL)"
	args := []interface{}{dataProximoAno, dataInicioAno}

	// Total de participacoes
	r.db.Model(&ComissaoMembro{}).Where("senador_id = ? AND "+dateFilter, append([]interface{}{senadorID}, args...)...).Count(&total)
	stats.TotalComissoes = int(total)

	// Titular
	r.db.Model(&ComissaoMembro{}).Where(
		"senador_id = ? AND descricao_participacao = ? AND "+dateFilter,
		append([]interface{}{senadorID, "Titular"}, args...)...,
	).Count(&titular)
	stats.ComissoesTitular = int(titular)

	// Suplente
	r.db.Model(&ComissaoMembro{}).Where(
		"senador_id = ? AND descricao_participacao = ? AND "+dateFilter,
		append([]interface{}{senadorID, "Suplente"}, args...)...,
	).Count(&suplente)
	stats.ComissoesSuplente = int(suplente)

	// Ativas (sem data_fim ou data_fim no futuro do ano... mas simplificando, seria se estava ativa em algum momento)
	// Comissoes "Ativas" no contexto anual significa se participou.
	stats.ComissoesAtivas = int(total) // No contexto anual, todas contadas foram ativas em algum momento

	// Calcular taxa de titularidade
	if stats.TotalComissoes > 0 {
		stats.TaxaTitularidade = float64(stats.ComissoesTitular) / float64(stats.TotalComissoes) * 100
	}

	return &stats, nil
}
