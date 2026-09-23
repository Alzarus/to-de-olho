package comissao

import (
	"time"

	"gorm.io/gorm"

	"github.com/Alzarus/to-de-olho/internal/utils"
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

// GetStats retorna estatisticas de comissoes no periodo do mandato (desde o
// inicio do recorte)
func (r *Repository) GetStats(senadorID int) (*ComissaoStats, error) {
	inicio, fim := utils.PeriodoDoMandato()
	return r.GetStatsPeriodo(senadorID, inicio, fim)
}

// GetStatsPeriodo pontua as participacoes que tocam [inicio, fim). Regras em
// pontuacao.go (itens 6 e 7 da auditoria).
func (r *Repository) GetStatsPeriodo(senadorID int, inicio, fim time.Time) (*ComissaoStats, error) {
	var participacoes []ComissaoMembro
	err := r.db.Where("senador_id = ? AND (data_inicio IS NULL OR data_inicio < ?) AND (data_fim IS NULL OR data_fim >= ?)",
		senadorID, fim, inicio).Find(&participacoes).Error
	if err != nil {
		return nil, err
	}
	return calcularStats(senadorID, participacoes, fim), nil
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

// SubstituirDoSenador troca todas as participacoes do senador pela lista da
// API, numa transacao. A API devolve o historico completo, uma entrada por
// periodo; o upsert antigo por (senador, comissao) juntava periodos
// diferentes numa linha so e deixava data_fim anterior a data_inicio.
func (r *Repository) SubstituirDoSenador(senadorID int, participacoes []ComissaoMembro) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("senador_id = ?", senadorID).Delete(&ComissaoMembro{}).Error; err != nil {
			return err
		}
		if len(participacoes) == 0 {
			return nil
		}
		return tx.CreateInBatches(participacoes, 500).Error
	})
}

// DeleteBySenadorID remove todas as comissoes de um senador
// Usado antes de re-sincronizar para evitar duplicatas
func (r *Repository) DeleteBySenadorID(senadorID int) error {
	return r.db.Where("senador_id = ?", senadorID).Delete(&ComissaoMembro{}).Error
}

// GetStatsByAno retorna estatisticas de comissoes no ano, dentro do recorte.
// Mesma formula do mandato (item 7).
func (r *Repository) GetStatsByAno(senadorID int, ano int) (*ComissaoStats, error) {
	inicio, fim := utils.PeriodoDoAno(ano)
	return r.GetStatsPeriodo(senadorID, inicio, fim)
}
