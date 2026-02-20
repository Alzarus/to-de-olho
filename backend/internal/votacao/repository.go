package votacao

import (
	"fmt"
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

// FindBySenadorID retorna votacoes de um senador com paginacao e filtros
func (r *Repository) FindBySenadorID(senadorID int, limit, offset int, votoType string) ([]Votacao, int64, error) {
	var votacoes []Votacao
	var total int64

	query := r.db.Model(&Votacao{}).Where("senador_id = ?", senadorID)

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

// Upsert insere ou atualiza uma votacao usando chave composta (senador_id, sessao_id)
func (r *Repository) Upsert(votacao *Votacao) error {
	return r.db.Where("senador_id = ? AND sessao_id = ?", votacao.SenadorID, votacao.SessaoID).
		Assign(*votacao).FirstOrCreate(votacao).Error
}

// UpdateMetadata atualiza metadados de uma sessao de votacao
func (r *Repository) UpdateMetadata(sessaoID string, updates map[string]interface{}) error {
	return r.db.Model(&Votacao{}).Where("sessao_id = ?", sessaoID).Updates(updates).Error
}

// UpdateVoteBatch atualiza o tipo de voto em massa (para normalizacao)
func (r *Repository) UpdateVoteBatch(oldVoto, newVoto string) error {
	return r.db.Model(&Votacao{}).Where("voto = ?", oldVoto).Update("voto", newVoto).Error
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
		stats.TaxaPresenca = float64(stats.TotalVotacoes-stats.Ausencias) / float64(stats.TotalVotacoes) * 100
		stats.TaxaParticipacao = float64(stats.VotosRegistrados) / float64(stats.TotalVotacoes) * 100
	}

	return &stats, nil
}

// FindAll retorna votacoes com paginacao e filtros (ordem: "asc" ou "desc")
func (r *Repository) FindAll(limit, offset, ano int, materia, ordem string) ([]Votacao, int64, error) {
	var votacoes []Votacao
	var total int64

	// Base query com filtros para count e subquery
	baseQuery := r.db.Model(&Votacao{})

	if ano > 0 {
		baseQuery = baseQuery.Where("EXTRACT(YEAR FROM data) = ?", ano)
	}

	if materia != "" {
		like := "%" + materia + "%"
		baseQuery = baseQuery.Where("materia ILIKE ? OR descricao_votacao ILIKE ? OR codigo_sessao ILIKE ?", like, like, like)
	}

	// Contar total de sessoes unicas
	if err := baseQuery.Session(&gorm.Session{}).Select("COUNT(DISTINCT sessao_id)").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Subquery para obter sessoes unicas
	// Nota: DISTINCT ON requer que o ORDER BY comece com a coluna distinct
	subQuery := baseQuery.Session(&gorm.Session{}).
		Select("DISTINCT ON (sessao_id) *").
		Order("sessao_id, data DESC")

	// Query principal ordenando o resultado da subquery pela data
	sortOrder := "data DESC"
	if ordem == "asc" {
		sortOrder = "data ASC"
	}

	err := r.db.Table("(?) as v", subQuery).
		Order(sortOrder).
		Limit(limit).
		Offset(offset).
		Find(&votacoes).Error

	return votacoes, total, err
}

// FindByID retorna uma votacao pelo ID (sessao_id)
func (r *Repository) FindByID(id string) (*Votacao, error) {
	var votacao Votacao
	err := r.db.Where("sessao_id = ?", id).First(&votacao).Error
	if err != nil {
		return nil, err
	}
	return &votacao, nil
}

// FindVotosBySessaoID retorna todos os votos de uma sessao especifica
func (r *Repository) FindVotosBySessaoID(sessaoID string) ([]Votacao, error) {
	var votacoes []Votacao
	err := r.db.Debug().Table("votacoes").
		Select("votacoes.*, senadores.nome as senador_nome, senadores.partido as senador_partido, senadores.uf as senador_uf, senadores.foto_url as senador_foto").
		Joins("JOIN senadores ON senadores.id = votacoes.senador_id").
		Where("votacoes.sessao_id = ?", sessaoID).
		Order("senadores.nome ASC").
		Find(&votacoes).Error

	return votacoes, err
}

// GetAllSessoesIDs retorna IDs de sessoes distintas para um ano
func (r *Repository) GetAllSessoesIDs(ano int) ([]string, error) {
	var ids []string
	// sessao_id format: "CODE_YEAR"
	like := fmt.Sprintf("%%_%d", ano)
	err := r.db.Model(&Votacao{}).
		Distinct("sessao_id").
		Where("sessao_id LIKE ?", like).
		Pluck("sessao_id", &ids).Error
	return ids, err
}
