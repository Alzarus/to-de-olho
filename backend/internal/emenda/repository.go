package emenda

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Upsert(emenda *Emenda) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "numero"}, {Name: "senador_id"}, {Name: "ano"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"tipo",
			"funcional_programatica",
			"localidade",
			"valor_empenhado",
			"valor_pago",
			"data_ultima_atualizacao",
		}),
	}).Create(emenda).Error
}

func (r *Repository) ListBySenador(senadorID uint, ano int) ([]Emenda, error) {
	var emendas []Emenda
	query := r.db.Where("senador_id = ?", senadorID)
	if ano > 0 {
		query = query.Where("ano = ?", ano)
	}
	err := query.Order("valor_pago DESC").Find(&emendas).Error
	return emendas, err
}

func (r *Repository) GetResumo(senadorID uint, ano int) (*ResumoEmendas, error) {
	var resumo ResumoEmendas

	type Result struct {
		TotalEmpenhado float64
		TotalPago      float64
		Quantidade     int64
	}
	var res Result

	query := r.db.Model(&Emenda{}).
		Where("senador_id = ?", senadorID)
	if ano > 0 {
		query = query.Where("ano = ?", ano)
	}

	err := query.
		Select("COALESCE(sum(valor_empenhado), 0) as total_empenhado, COALESCE(sum(valor_pago), 0) as total_pago, count(*) as quantidade").
		Scan(&res).Error

	if err != nil {
		return nil, err
	}

	resumo.TotalEmpenhado = res.TotalEmpenhado
	resumo.TotalPago = res.TotalPago
	resumo.Quantidade = res.Quantidade

	// Top Localidades (apenas com valor pago > 0)
	localidadesQuery := r.db.Model(&Emenda{}).
		Where("senador_id = ? AND valor_pago > 0", senadorID)
	if ano > 0 {
		localidadesQuery = localidadesQuery.Where("ano = ?", ano)
	}

	rows, err := localidadesQuery.
		Select("localidade, sum(valor_pago) as valor").
		Group("localidade").
		Order("valor DESC").
		Limit(10).
		Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var lv LocalidadeValor
		// Scan direto em struct não funciona com Rows em GORM v2 as vezes, melhor scan manual
		var loc string
		var val float64
		if err := rows.Scan(&loc, &val); err == nil {
			lv.Localidade = loc
			lv.Valor = val
			resumo.TopLocalidades = append(resumo.TopLocalidades, lv)
		}
	}

	return &resumo, nil
}
