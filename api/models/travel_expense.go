package models

import "gorm.io/gorm"

type TravelExpense struct {
	gorm.Model
	Data          string  `gorm:"column:data" json:"date"`
	Tipo          string  `gorm:"column:tipo" json:"type"`
	Usuario       string  `gorm:"column:usuario" json:"user"`
	Valor         float64 `gorm:"column:valor" json:"value"`
	Localidade    string  `gorm:"column:localidade" json:"location"`
	Justificativa string  `gorm:"column:justificativa" json:"justification"`
}
