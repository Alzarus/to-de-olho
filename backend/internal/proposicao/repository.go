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

// FindBySenadorID retorna proposicoes de um senador com paginacao, busca e filtros
func (r *Repository) FindBySenadorID(senadorID int, limit int, offset int, queryStr string, ano int, sigla string, tramitacao string, sort string) ([]Proposicao, int64, error) {
	var proposicoes []Proposicao
	var total int64
	
	dbQuery := r.db.Model(&Proposicao{}).Where("senador_id = ?", senadorID)

	if queryStr != "" {
		search := "%" + queryStr + "%"
		dbQuery = dbQuery.Where("(ementa ILIKE ? OR descricao_identificacao ILIKE ? OR codigo_materia ILIKE ?)", search, search, search)
	}

	if ano > 0 {
		dbQuery = dbQuery.Where("ano_materia = ?", ano)
	}

	if sigla != "" {
		// Use TRIM to handle potential whitespace in DB or input
		dbQuery = dbQuery.Where("TRIM(sigla_subtipo_materia) ILIKE TRIM(?)", sigla)
	}

	if tramitacao != "" {
		dbQuery = dbQuery.Where("(estagio_tramitacao = ? OR situacao_atual ILIKE ?)", tramitacao, tramitacao)
	}

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sorting
	// Default: Data DESC (NULLS LAST to keep invalid dates at bottom), fallback to Ano/Codigo
	order := "data_apresentacao DESC NULLS LAST, ano_materia DESC, codigo_materia DESC"
	
	if sort == "data_asc" {
		order = "data_apresentacao ASC NULLS LAST, ano_materia ASC, codigo_materia ASC"
	} else if sort == "ano_desc" {
		order = "ano_materia DESC, data_apresentacao DESC NULLS LAST"
	}
	
	dbQuery = dbQuery.Order(order)

	if limit > 0 {
		dbQuery = dbQuery.Limit(limit)
	}
	if offset > 0 {
		dbQuery = dbQuery.Offset(offset)
	}

	result := dbQuery.Find(&proposicoes)
	return proposicoes, total, result.Error
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

// GetStatsByAno retorna estatisticas de proposicoes filtradas por ano de apresentacao
func (r *Repository) GetStatsByAno(senadorID int, ano int) (*ProposicaoStats, error) {
	var stats ProposicaoStats
	stats.SenadorID = senadorID

	var total, pecs, plps, pls, leis, plenario, tramitacao int64
	var pontuacaoTotal int

	// Filtro de data: baseado no ano_materia ou data_apresentacao
	// O mais confiavel para "ano" costuma ser o ano_materia para proposicoes legislativas
	yearFilter := "ano_materia = ?"

	// Total de proposicoes
	r.db.Model(&Proposicao{}).Where("senador_id = ? AND "+yearFilter, senadorID, ano).Count(&total)
	stats.TotalProposicoes = int(total)

	// PECs
	r.db.Model(&Proposicao{}).Where("senador_id = ? AND sigla_subtipo_materia = ? AND "+yearFilter, senadorID, "PEC", ano).Count(&pecs)
	stats.TotalPECs = int(pecs)

	// PLPs
	r.db.Model(&Proposicao{}).Where("senador_id = ? AND sigla_subtipo_materia = ? AND "+yearFilter, senadorID, "PLP", ano).Count(&plps)
	stats.TotalPLPs = int(plps)

	// PLs
	r.db.Model(&Proposicao{}).Where("senador_id = ? AND sigla_subtipo_materia = ? AND "+yearFilter, senadorID, "PL", ano).Count(&pls)
	stats.TotalPLs = int(pls)

	// Outros
	stats.TotalOutros = stats.TotalProposicoes - stats.TotalPECs - stats.TotalPLPs - stats.TotalPLs

	// Transformadas em lei (filtro yearFilter se aplica a quando foi apresentada, nao quando virou lei, conforme metodologia de produtividade por safra)
	r.db.Model(&Proposicao{}).Where(
		"senador_id = ? AND estagio_tramitacao = ? AND "+yearFilter, senadorID, "TransformadoLei", ano,
	).Count(&leis)
	stats.TransformadasEmLei = int(leis)

	// Aprovados plenario
	r.db.Model(&Proposicao{}).Where(
		"senador_id = ? AND estagio_tramitacao IN (?, ?) AND "+yearFilter, senadorID, "AprovadoPlenario", "TransformadoLei", ano,
	).Count(&plenario)
	stats.AprovadosPlenario = int(plenario)

	// Em tramitacao
	r.db.Model(&Proposicao{}).Where(
		"senador_id = ? AND estagio_tramitacao IN (?, ?, ?) AND "+yearFilter, senadorID, "Apresentado", "EmComissao", "AprovadoComissao", ano,
	).Count(&tramitacao)
	stats.EmTramitacao = int(tramitacao)

	// Soma de pontuacao
	var soma struct {
		Total int
	}
	r.db.Model(&Proposicao{}).
		Select("COALESCE(SUM(pontuacao), 0) as total").
		Where("senador_id = ? AND "+yearFilter, senadorID, ano).
		Scan(&soma)
	pontuacaoTotal = soma.Total
	stats.PontuacaoTotal = pontuacaoTotal

	return &stats, nil
}
