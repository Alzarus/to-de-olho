package models

import "time"

type Contract struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	NumeroContrato      string    `json:"numero_contrato"`
	NomeContratado      string    `json:"nome_contratado"`
	DataAssinatura      time.Time `json:"data_assinatura"`
	DataInicio          time.Time `json:"data_inicio"`
	DataFim             time.Time `json:"data_fim"`
	ValorContrato       float64   `json:"valor_contrato"`
	TempoMaximoExecucao string    `json:"tempo_maximo_execucao"`
	DataPublicacao      time.Time `json:"data_publicacao"`
	DiarioOficial       *string   `json:"diario_oficial"`
	LinkPDF             *string   `json:"link_pdf"`
}
