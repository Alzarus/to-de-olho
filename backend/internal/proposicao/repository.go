package proposicao

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Alzarus/to-de-olho/internal/utils"
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

// GetStats retorna estatisticas de proposicoes de um senador apresentadas a
// partir do inicio do recorte (posse da legislatura atual)
func (r *Repository) GetStats(senadorID int) (*ProposicaoStats, error) {
	return r.stats(senadorID, "data_apresentacao >= ?", utils.InicioRecorte())
}

// condicaoPrincipal marca a linha em que o senador e o primeiro autor
const condicaoPrincipal = "posicao_autoria = 1"

// stats agrega as estatisticas de um senador em uma query. Contagens e
// pontuacao consideram so a autoria principal; coautorias sao contadas a parte.
func (r *Repository) stats(senadorID int, filtro string, args ...any) (*ProposicaoStats, error) {
	var linha struct {
		Total, Coautorias, Pecs, Plps, Pls, Leis, Plenario, Tramitacao int
		Pontuacao                                                       float64
	}
	p := condicaoPrincipal
	err := r.db.Model(&Proposicao{}).
		Select(`
			COUNT(*) FILTER (WHERE `+p+`) AS total,
			COUNT(*) FILTER (WHERE COALESCE(posicao_autoria, 0) <> 1) AS coautorias,
			COUNT(*) FILTER (WHERE `+p+` AND sigla_subtipo_materia = 'PEC') AS pecs,
			COUNT(*) FILTER (WHERE `+p+` AND sigla_subtipo_materia = 'PLP') AS plps,
			COUNT(*) FILTER (WHERE `+p+` AND sigla_subtipo_materia = 'PL') AS pls,
			COUNT(*) FILTER (WHERE `+p+` AND estagio_tramitacao = 'TransformadoLei') AS leis,
			COUNT(*) FILTER (WHERE `+p+` AND estagio_tramitacao IN ('AprovadoPlenario', 'TransformadoLei')) AS plenario,
			COUNT(*) FILTER (WHERE `+p+` AND estagio_tramitacao IN ('Apresentado', 'EmComissao', 'AprovadoComissao')) AS tramitacao,
			COALESCE(SUM(pontuacao) FILTER (WHERE `+p+`), 0) AS pontuacao`).
		Where("senador_id = ?", senadorID).
		Where(filtro, args...).
		Scan(&linha).Error
	if err != nil {
		return nil, err
	}

	return &ProposicaoStats{
		SenadorID:          senadorID,
		TotalProposicoes:   linha.Total,
		TotalCoautorias:    linha.Coautorias,
		TotalPECs:          linha.Pecs,
		TotalPLPs:          linha.Plps,
		TotalPLs:           linha.Pls,
		TotalOutros:        linha.Total - linha.Pecs - linha.Plps - linha.Pls,
		TransformadasEmLei: linha.Leis,
		AprovadosPlenario:  linha.Plenario,
		EmTramitacao:       linha.Tramitacao,
		PontuacaoTotal:     linha.Pontuacao,
	}, nil
}

// GetProposicoesPorTipo retorna contagem de proposicoes por tipo
func (r *Repository) GetProposicoesPorTipo(senadorID int) ([]ProposicaoPorTipo, error) {
	var result []ProposicaoPorTipo
	err := r.db.Model(&Proposicao{}).
		Select("sigla_subtipo_materia as tipo, COUNT(*) as total").
		Where("senador_id = ? AND "+condicaoPrincipal, senadorID).
		Group("sigla_subtipo_materia").
		Order("total DESC").
		Scan(&result).Error
	return result, err
}

// colunasAtualizaveis sao reescritas no upsert, para uma correcao na API
// (autoria, ementa, identificacao) chegar ao banco
var colunasAtualizaveis = []string{
	"sigla_subtipo_materia", "numero_materia", "ano_materia", "descricao_identificacao",
	"ementa", "situacao_atual", "data_apresentacao", "estagio_tramitacao", "pontuacao",
	"posicao_autoria", "total_autores", "autoria", "updated_at",
}

var conflitoSenadorMateria = clause.OnConflict{
	Columns:   []clause.Column{{Name: "senador_id"}, {Name: "codigo_materia"}},
	DoUpdates: clause.AssignmentColumns(colunasAtualizaveis),
}

// Upsert insere ou atualiza uma proposicao pela chave (senador_id, codigo_materia)
func (r *Repository) Upsert(proposicao *Proposicao) error {
	return r.db.Clauses(conflitoSenadorMateria).Create(proposicao).Error
}

// UpsertBatch insere ou atualiza multiplas proposicoes, em lotes de 500
func (r *Repository) UpsertBatch(proposicoes []Proposicao) error {
	if len(proposicoes) == 0 {
		return nil
	}
	return r.db.Clauses(conflitoSenadorMateria).CreateInBatches(proposicoes, 500).Error
}

// DeleteBySenadorID remove todas as proposicoes de um senador
func (r *Repository) DeleteBySenadorID(senadorID int) error {
	return r.db.Where("senador_id = ?", senadorID).Delete(&Proposicao{}).Error
}

// GetStatsByAno retorna estatisticas de proposicoes filtradas pelo ano da
// materia (producao por safra)
func (r *Repository) GetStatsByAno(senadorID int, ano int) (*ProposicaoStats, error) {
	return r.stats(senadorID, "ano_materia = ?", ano)
}
