package comissao

import "time"

// ComissaoMembro representa a participacao de um senador em uma comissao
type ComissaoMembro struct {
	ID                 int        `gorm:"primaryKey" json:"id"`
	SenadorID          int        `gorm:"uniqueIndex:idx_comissao_membro_unico,priority:1;index:idx_comissao_senador;not null" json:"senador_id"`
	CodigoComissao     string     `gorm:"uniqueIndex:idx_comissao_membro_unico,priority:2;index:idx_comissao_codigo" json:"codigo_comissao"`
	SiglaComissao      string     `json:"sigla_comissao"`
	NomeComissao       string     `json:"nome_comissao"`
	SiglaCasaComissao  string     `json:"sigla_casa_comissao"` // SF, CN
	DescricaoParticipacao string  `json:"descricao_participacao"` // Titular, Suplente
	DataInicio         *time.Time `json:"data_inicio,omitempty"`
	DataFim            *time.Time `json:"data_fim,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName define o nome da tabela
func (ComissaoMembro) TableName() string {
	return "comissao_membros"
}

// ComissaoStats representa estatisticas de participacao em comissoes
type ComissaoStats struct {
	SenadorID            int     `json:"senador_id"`
	TotalComissoes       int     `json:"total_comissoes"`
	ComissoesTitular     int     `json:"comissoes_titular"`
	ComissoesSuplente    int     `json:"comissoes_suplente"`
	ComissoesAtivas      int     `json:"comissoes_ativas"`      // Sem data_fim
	TaxaTitularidade     float64 `json:"taxa_titularidade"`     // Titular / Total * 100
}

// ComissoesPorCasa representa contagem de comissoes por casa
type ComissoesPorCasa struct {
	Casa  string `json:"casa"`
	Total int    `json:"total"`
}
