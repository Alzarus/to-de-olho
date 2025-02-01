package models

import "gorm.io/gorm"

type Frequency struct {
	gorm.Model
	SessionNumber string `gorm:"column:numero_sessao" json:"session_number"`
	SessionYear   string `gorm:"column:ano_sessao" json:"session_year"`
	ShortName     string `gorm:"column:nome_abreviado" json:"short_name"`
	Status        string `gorm:"column:status_presenca" json:"status"`
}
