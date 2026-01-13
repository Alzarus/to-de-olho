package votacao

import "time"

// Votacao representa um voto de um senador em uma sessao
type Votacao struct {
	ID              int       `gorm:"primaryKey" json:"id"`
	SenadorID       int       `gorm:"index:idx_votacao_senador;not null" json:"senador_id"`
	SessaoID        string    `gorm:"index:idx_votacao_sessao;not null" json:"sessao_id"`
	CodigoSessao    string    `json:"codigo_sessao"`
	Data            time.Time `gorm:"index" json:"data"`
	Voto            string    `json:"voto"` // Sim, Nao, Abstencao, Obstrucao, NCom
	DescricaoVotacao string   `json:"descricao_votacao,omitempty"`
	Materia         string    `json:"materia,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName define o nome da tabela
func (Votacao) TableName() string {
	return "votacoes"
}

// VotacaoStats representa estatisticas de votacao de um senador
type VotacaoStats struct {
	SenadorID           int     `json:"senador_id"`
	TotalVotacoes       int     `json:"total_votacoes"`
	VotosRegistrados    int     `json:"votos_registrados"`    // Sim + Nao + Abstencao
	Ausencias           int     `json:"ausencias"`            // NCom (Nao Compareceu)
	Obstrucoes          int     `json:"obstrucoes"`
	TaxaPresenca        float64 `json:"taxa_presenca"`        // 0-100
	TaxaParticipacao    float64 `json:"taxa_participacao"`    // Votos efetivos / Total
}

// VotosPorTipo representa contagem de votos por tipo
type VotosPorTipo struct {
	Voto  string `json:"voto"`
	Total int    `json:"total"`
}
