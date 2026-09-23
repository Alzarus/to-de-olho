package votacao

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Alzarus/to-de-olho/internal/utils"
)

// Repository encapsula operacoes de banco de dados para Votacao
type Repository struct {
	db *gorm.DB
}

// NewRepository cria um novo repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// FindBySenadorID retorna votacoes de um senador com paginacao e filtros
func (r *Repository) FindBySenadorID(senadorID int, limit, offset int, votoType string, ano int) ([]Votacao, int64, error) {
	var votacoes []Votacao
	var total int64

	query := r.db.Model(&Votacao{}).Where("senador_id = ?", senadorID)

	if ano > 0 {
		query = query.Where("EXTRACT(YEAR FROM data) = ?", ano)
	}

	// Filtro por tipo de voto
	if votoType != "" {
		if votoType == "Outros" {
			// Outros = tudo que NAO for Sim, Nao, Abstencao, Obstrucao
			query = query.Where("voto NOT IN (?, ?, ?, ?)", "Sim", "Nao", "Abstencao", "Obstrucao")
		} else {
			query = query.Where("voto = ?", votoType)
		}
	}

	// Contar total (considerando filtros)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Aplicar ordenacao e paginacao
	result := query.Order("data DESC").
		Limit(limit).
		Offset(offset).
		Find(&votacoes)
	
	return votacoes, total, result.Error
}

// CountBySenadorID retorna total de votacoes de um senador
func (r *Repository) CountBySenadorID(senadorID int) (int64, error) {
	var count int64
	result := r.db.Model(&Votacao{}).Where("senador_id = ?", senadorID).Count(&count)
	return count, result.Error
}

// Count retorna total de votacoes no banco
func (r *Repository) Count() (int64, error) {
	var count int64
	result := r.db.Model(&Votacao{}).Count(&count)
	return count, result.Error
}

// GetStats retorna estatisticas de votacao de um senador desde o inicio do
// recorte (posse da legislatura atual)
func (r *Repository) GetStats(senadorID int) (*VotacaoStats, error) {
	return r.stats(senadorID, "data >= ?", utils.InicioRecorte())
}

// stats conta os votos por sigla bruta e aplica a classificacao
func (r *Repository) stats(senadorID int, filtro string, args ...any) (*VotacaoStats, error) {
	var linhas []struct {
		SiglaVoto string
		Total     int
	}
	err := r.db.Model(&Votacao{}).
		Select("sigla_voto, COUNT(*) AS total").
		Where("senador_id = ?", senadorID).
		Where(filtro, args...).
		Group("sigla_voto").
		Scan(&linhas).Error
	if err != nil {
		return nil, err
	}
	porSigla := make(map[string]int, len(linhas))
	for _, l := range linhas {
		porSigla[l.SiglaVoto] = l.Total
	}
	return calcularStats(senadorID, porSigla), nil
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

// UpsertBatch grava votos em lotes de 1.000, pela chave (senador_id, codigo_votacao)
func (r *Repository) UpsertBatch(votacoes []Votacao) error {
	if len(votacoes) == 0 {
		return nil
	}
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "senador_id"}, {Name: "codigo_votacao"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"sessao_id", "codigo_sessao", "sequencial_votacao", "data", "sigla_voto", "voto",
			"descricao_votacao", "materia", "ementa", "resultado", "updated_at",
		}),
	}).CreateInBatches(votacoes, 1000).Error
}

// SenadoresEmExercicioSemVotos lista senadores em exercicio sem nenhum voto
// a partir de `desde`
func (r *Repository) SenadoresEmExercicioSemVotos(desde time.Time) ([]string, error) {
	var nomes []string
	err := r.db.Raw(`
		SELECT s.nome FROM senadores s
		WHERE s.em_exercicio
		  AND NOT EXISTS (SELECT 1 FROM votacoes v WHERE v.senador_id = s.id AND v.data >= ?)
		ORDER BY s.nome`, desde).Scan(&nomes).Error
	return nomes, err
}

// GetStatsByAno retorna estatisticas de votacao das sessoes de um ano
func (r *Repository) GetStatsByAno(senadorID int, ano int) (*VotacaoStats, error) {
	return r.stats(senadorID, "data >= ? AND data < ?",
		time.Date(ano, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(ano+1, 1, 1, 0, 0, 0, 0, time.UTC))
}

// FindAll retorna votacoes (uma linha por votacao, nao por voto) com
// paginacao e filtros. ordem: "asc" ou "desc". sessao filtra pelo codigo da
// sessao (destino dos links antigos /votacoes/NNNNNN_AAAA, decisao D2).
func (r *Repository) FindAll(limit, offset, ano int, materia, ordem, sessao string) ([]Votacao, int64, error) {
	var votacoes []Votacao
	var total int64

	baseQuery := r.db.Model(&Votacao{})

	if ano > 0 {
		baseQuery = baseQuery.Where("EXTRACT(YEAR FROM data) = ?", ano)
	}

	if materia != "" {
		like := "%" + materia + "%"
		baseQuery = baseQuery.Where("(materia ILIKE ? OR descricao_votacao ILIKE ? OR ementa ILIKE ? OR codigo_sessao ILIKE ?)", like, like, like, like)
	}

	if sessao != "" {
		baseQuery = baseQuery.Where("sessao_id = ?", sessao)
	}

	if err := baseQuery.Session(&gorm.Session{}).Select("COUNT(DISTINCT codigo_votacao)").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// DISTINCT ON exige que o ORDER BY comece pela coluna distinta
	subQuery := baseQuery.Session(&gorm.Session{}).
		Select("DISTINCT ON (codigo_votacao) *").
		Order("codigo_votacao, id")

	// Dentro do dia, a ordem das votacoes na sessao (sequencial pode ser nulo)
	sortOrder := "data DESC, sessao_id DESC, sequencial_votacao DESC NULLS LAST, codigo_votacao DESC"
	if ordem == "asc" {
		sortOrder = "data ASC, sessao_id ASC, sequencial_votacao ASC NULLS FIRST, codigo_votacao ASC"
	}

	err := r.db.Table("(?) as v", subQuery).
		Order(sortOrder).
		Limit(limit).
		Offset(offset).
		Find(&votacoes).Error

	return votacoes, total, err
}

// FindByID retorna os dados de uma votacao pelo codigo_votacao
func (r *Repository) FindByID(codigoVotacao int) (*Votacao, error) {
	var votacao Votacao
	err := r.db.Where("codigo_votacao = ?", codigoVotacao).Order("id").First(&votacao).Error
	if err != nil {
		return nil, err
	}
	return &votacao, nil
}

// FindVotosByCodigoVotacao retorna os votos de todos os senadores em uma votacao
func (r *Repository) FindVotosByCodigoVotacao(codigoVotacao int) ([]Votacao, error) {
	var votacoes []Votacao
	err := r.db.Table("votacoes").
		Select("votacoes.*, senadores.nome as senador_nome, senadores.partido as senador_partido, senadores.uf as senador_uf, senadores.foto_url as senador_foto").
		Joins("JOIN senadores ON senadores.id = votacoes.senador_id").
		Where("votacoes.codigo_votacao = ?", codigoVotacao).
		Order("senadores.nome ASC").
		Find(&votacoes).Error

	return votacoes, err
}
