package ceaps

import "time"

// DespesaCEAPS representa um lancamento da Cota para o Exercicio da Atividade Parlamentar
type DespesaCEAPS struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	SenadorID int       `gorm:"index:idx_despesa_senador_ano;not null" json:"senador_id"`
	Ano       int       `gorm:"index:idx_despesa_senador_ano;not null" json:"ano"`
	Mes       int       `json:"mes"`

	// Dados do lancamento
	TipoDespesa  string  `json:"tipo_despesa"`
	Fornecedor   string  `json:"fornecedor"`
	CNPJCPF      string  `gorm:"column:cnpj_cpf" json:"cnpj_cpf"`
	Documento    string  `json:"documento,omitempty"`
	DataEmissao  *time.Time `json:"data_emissao,omitempty"`
	Valor        float64 `json:"valor"`

	// Chave natural para idempotencia (upsert)
	// Composta: (senador_id, cnpj_cpf, data_emissao, valor_centavos)
	ValorCentavos int64 `gorm:"uniqueIndex:idx_despesa_unica,priority:4" json:"-"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName define o nome da tabela
func (DespesaCEAPS) TableName() string {
	return "despesas_ceaps"
}

// BeforeCreate converte valor para centavos antes de inserir
func (d *DespesaCEAPS) BeforeCreate() {
	d.ValorCentavos = int64(d.Valor * 100)
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
