package votacao

import (
	"fmt"
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

// GetStats retorna estatisticas de votacao de um senador restritas ao mandato (2023+)
func (r *Repository) GetStats(senadorID int) (*VotacaoStats, error) {
	var stats VotacaoStats
	stats.SenadorID = senadorID

	var total, registrados, ausencias, obstrucoes int64

	mandatoFilter := fmt.Sprintf("senador_id = ? AND data >= '%d-01-01'", utils.GetInicioLegislaturaAtual())

	// Total de votacoes
	r.db.Model(&Votacao{}).Where(mandatoFilter, senadorID).Count(&total)
	stats.TotalVotacoes = int(total)

	// Votos registrados (Sim, Nao, Abstencao)
	r.db.Model(&Votacao{}).Where(
		mandatoFilter+" AND voto IN (?, ?, ?)", senadorID, "Sim", "Nao", "Abstencao",
	).Count(&registrados)
	stats.VotosRegistrados = int(registrados)

	// Ausencias (NCom)
	r.db.Model(&Votacao{}).Where(
		mandatoFilter+" AND voto = ?", senadorID, "NCom",
	).Count(&ausencias)
	stats.Ausencias = int(ausencias)

	// Obstrucoes
	r.db.Model(&Votacao{}).Where(
		mandatoFilter+" AND voto = ?", senadorID, "Obstrucao",
	).Count(&obstrucoes)
	stats.Obstrucoes = int(obstrucoes)

	// Calcular taxas
	if stats.TotalVotacoes > 0 {
		// Presenca (calculada em cima dos que de fato ele devia estar: registrados + ausencias + obstrucoes)
		// Ignorando fatores como Licenca, Missao, Presidencia do Senado (P-OD)
		baseCalculoPresenca := stats.VotosRegistrados + stats.Ausencias + stats.Obstrucoes
		
		if baseCalculoPresenca > 0 {
			stats.TaxaPresenca = float64(stats.VotosRegistrados+stats.Obstrucoes) / float64(baseCalculoPresenca) * 100
		} else {
			stats.TaxaPresenca = 0
		}

		// Participacao = Votos efetivos (Sim, Nao, Abstencao) / Total real baseCalculada
		if baseCalculoPresenca > 0 {
			stats.TaxaParticipacao = float64(stats.VotosRegistrados) / float64(baseCalculoPresenca) * 100
		} else {
			stats.TaxaParticipacao = 0
		}
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

// GetStatsByAno retorna estatisticas de votacao filtradas por ano
func (r *Repository) GetStatsByAno(senadorID int, ano int) (*VotacaoStats, error) {
	var stats VotacaoStats
	stats.SenadorID = senadorID

	var total, registrados, ausencias, obstrucoes int64

	// Filtro de data: inicio do ano e inicio do proximo ano
	dataInicio := fmt.Sprintf("%d-01-01", ano)
	dataProximoAno := fmt.Sprintf("%d-01-01", ano+1)
	dateFilter := "data >= ? AND data < ?"

	// Total de votacoes
	r.db.Debug().Model(&Votacao{}).Where("senador_id = ? AND "+dateFilter, senadorID, dataInicio, dataProximoAno).Count(&total)
	stats.TotalVotacoes = int(total)

	// Votos registrados
	r.db.Model(&Votacao{}).Where(
		"senador_id = ? AND voto IN (?, ?, ?) AND "+dateFilter,
		senadorID, "Sim", "Nao", "Abstencao", dataInicio, dataProximoAno,
	).Count(&registrados)
	stats.VotosRegistrados = int(registrados)

	// Ausencias
	r.db.Model(&Votacao{}).Where(
		"senador_id = ? AND voto = ? AND "+dateFilter,
		senadorID, "NCom", dataInicio, dataProximoAno,
	).Count(&ausencias)
	stats.Ausencias = int(ausencias)

	// Obstrucoes
	r.db.Model(&Votacao{}).Where(
		"senador_id = ? AND voto = ? AND "+dateFilter,
		senadorID, "Obstrucao", dataInicio, dataProximoAno,
	).Count(&obstrucoes)
	stats.Obstrucoes = int(obstrucoes)

	// Calcular taxas
	if stats.TotalVotacoes > 0 {
		baseCalculoPresenca := stats.VotosRegistrados + stats.Ausencias + stats.Obstrucoes

		if baseCalculoPresenca > 0 {
			stats.TaxaPresenca = float64(stats.VotosRegistrados+stats.Obstrucoes) / float64(baseCalculoPresenca) * 100
			stats.TaxaParticipacao = float64(stats.VotosRegistrados) / float64(baseCalculoPresenca) * 100
		} else {
			stats.TaxaPresenca = 0
			stats.TaxaParticipacao = 0
		}
	}

	return &stats, nil
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
