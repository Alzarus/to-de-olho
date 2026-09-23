package ceaps

import (
	"math"
	"time"

	"gorm.io/gorm"
)

// DespesaCEAPS representa um lancamento da Cota para o Exercicio da Atividade Parlamentar
type DespesaCEAPS struct {
	ID int `gorm:"primaryKey" json:"id"`
	// IDOrigem e o id do lancamento na API do Senado: a chave. A chave antiga
	// (senador, cnpj, data, valor) juntava lancamentos distintos com o mesmo
	// fornecedor, dia e valor (ex.: 28 de 71 do Davi Alcolumbre em 2025).
	IDOrigem  int `gorm:"uniqueIndex:idx_despesa_origem;not null" json:"id_origem"`
	SenadorID int `gorm:"index:idx_despesa_senador_ano;not null" json:"senador_id"`
	Ano       int `gorm:"index:idx_despesa_senador_ano;not null" json:"ano"`
	Mes       int `json:"mes"`

	// Dados do lancamento
	TipoDespesa string     `json:"tipo_despesa"`
	Fornecedor  string     `json:"fornecedor"`
	CNPJCPF     string     `gorm:"column:cnpj_cpf" json:"cnpj_cpf"`
	Documento   string     `json:"documento,omitempty"`
	DataEmissao *time.Time `json:"data_emissao,omitempty"`
	Valor       float64    `json:"valor"`

	ValorCentavos int64 `json:"-"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName define o nome da tabela
func (DespesaCEAPS) TableName() string {
	return "despesas_ceaps"
}

// BeforeCreate converte valor para centavos antes de inserir
func (d *DespesaCEAPS) BeforeCreate(_ *gorm.DB) error {
	d.ValorCentavos = int64(math.Round(d.Valor * 100))
	return nil
}

// AggregatedDespesa representa gastos agregados por categoria
type AggregatedDespesa struct {
	TipoDespesa string  `json:"tipo_despesa"`
	Total       float64 `json:"total"`
	Quantidade  int     `json:"quantidade"`
}

// SenadorGastoMensal representa gasto mensal de um senador
type SenadorGastoMensal struct {
	Ano   int     `json:"ano"`
	Mes   int     `json:"mes"`
	Total float64 `json:"total"`
}
