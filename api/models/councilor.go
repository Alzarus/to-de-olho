package models

import "gorm.io/gorm"

type Councilor struct {
	gorm.Model
	Nome               string `gorm:"column:nome" json:"nome"`
	Partido            string `gorm:"column:partido" json:"partido"`
	Descricao          string `gorm:"column:descricao" json:"descricao"`
	LinkFoto           string `gorm:"column:link_foto" json:"link_foto"`
	EmAtividade        bool   `gorm:"column:em_atividade" json:"em_atividade"`
	Nascimento         string `gorm:"column:nascimento" json:"nascimento"`
	Telefone           string `gorm:"column:telefone" json:"telefone"`
	Email              string `gorm:"column:email" json:"email"`
	EnderecoDeGabinete string `gorm:"column:endereco_de_gabinete" json:"endereco_de_gabinete"`
}
