package emenda

import (
	"time"

	"github.com/Alzarus/to-de-olho/internal/senador"
)

type Emenda struct {
	ID                    uint             `gorm:"primaryKey" json:"id"`
	SenadorID             uint             `gorm:"index;uniqueIndex:idx_emenda_unique,priority:2" json:"senador_id"`
	Senador               *senador.Senador `gorm:"foreignKey:SenadorID" json:"senador,omitempty"`
	Ano                   int              `gorm:"index;uniqueIndex:idx_emenda_unique,priority:3" json:"ano"`
	Numero                string           `gorm:"index;uniqueIndex:idx_emenda_unique,priority:1" json:"numero"`
	Tipo                  string           `json:"tipo"` // Individual, Bancada, Especial, Relator, Comissao
	FuncionalProgramatica string           `json:"funcional_programatica"`
	Localidade            string           `json:"localidade"` // UF ou Município
	ValorEmpenhado        float64          `json:"valor_empenhado"`
	ValorPago             float64          `json:"valor_pago"`
	DataUltimaAtualizacao time.Time        `json:"data_ultima_atualizacao"`

	// Datas da execucao, de /emendas/documentos/{codigo}. Nulas ate a primeira
	// consulta ou quando a fase ainda nao ocorreu.
	DataPrimeiroEmpenho   *time.Time `json:"data_primeiro_empenho"`
	DataPrimeiroPagamento *time.Time `json:"data_primeiro_pagamento"`
	DataUltimoPagamento   *time.Time `json:"data_ultimo_pagamento"`
	// DatasConsultadasEm e DatasValorPago guardam quando e com que valor pago
	// as datas foram buscadas: valor pago diferente pede nova consulta.
	DatasConsultadasEm *time.Time `json:"-"`
	DatasValorPago     float64    `json:"-"`
}

type ResumoEmendas struct {
	TotalEmpenhado float64           `json:"total_empenhado"`
	TotalPago      float64           `json:"total_pago"`
	Quantidade     int64             `json:"quantidade"`
	TopLocalidades []LocalidadeValor `json:"top_localidades"`
}

type LocalidadeValor struct {
	Localidade string  `json:"localidade"`
	Valor      float64 `json:"valor"`
}
