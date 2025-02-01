package models

import "gorm.io/gorm"

type Proposition struct {
	gorm.Model
	Proposicao       string `gorm:"column:proposicao" json:"proposicao"`
	AutorProposicao  string `gorm:"column:autor_proposicao" json:"autorProposicao"`
	Ementa           string `gorm:"column:ementa" json:"proEmenta"`
	DataMovimentacao string `gorm:"column:data_movimentacao" json:"traDtMovimentacao"`
	Destino          string `gorm:"column:destino" json:"destino"`
	SituacaoFutura   string `gorm:"column:situacao_futura" json:"sitNomeFuturo"`
	AutorDocumento   string `gorm:"column:autor_documento" json:"autordoc"`
}
