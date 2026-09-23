package votacao

import "time"

// Votacao representa o voto de um senador em uma votacao nominal.
//
// A chave e (senador_id, codigo_votacao), onde codigo_votacao e o
// codigoSessaoVotacao da API: unico por votacao e nunca nulo. A chave antiga
// (senador_id, codigoSessao_ano) colapsava as votacoes de uma mesma sessao
// numa linha so (item 3 da auditoria).
type Votacao struct {
	ID                int       `gorm:"primaryKey" json:"id"`
	SenadorID         int       `gorm:"uniqueIndex:idx_votacao_senador_votacao,priority:1;index:idx_votacao_senador;not null" json:"senador_id"`
	CodigoVotacao     int       `gorm:"uniqueIndex:idx_votacao_senador_votacao,priority:2;index:idx_votacao_codigo;not null" json:"codigo_votacao"`
	SessaoID          string    `gorm:"index:idx_votacao_sessao;not null" json:"sessao_id"` // codigoSessao: agrupa as votacoes de uma sessao
	CodigoSessao      string    `json:"codigo_sessao"`
	SequencialVotacao *int      `json:"sequencial_votacao"` // a API manda null em parte das votacoes
	Data              time.Time `gorm:"index" json:"data"`
	SiglaVoto         string    `gorm:"not null" json:"sigla_voto"` // codigo bruto da API: Votou, P-NRV, AP, LS, MIS...
	Voto              string    `json:"voto"`                       // rotulo para exibicao: Sim, Nao, Abstencao, Obstrucao ou a sigla
	DescricaoVotacao  string    `json:"descricao_votacao,omitempty"`
	Materia           string    `json:"materia,omitempty"` // identificacao: "PLP 124/2022 (Substitutivo-CD)"
	Ementa            string    `json:"ementa,omitempty"`
	Resultado         string    `json:"resultado,omitempty"` // A (aprovada), R (rejeitada)...

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Campos populados via Join (read-only)
	SenadorNome    string `gorm:"->" json:"senador_nome,omitempty"`
	SenadorPartido string `gorm:"->" json:"senador_partido,omitempty"`
	SenadorUF      string `gorm:"->" json:"senador_uf,omitempty"`
	SenadorFoto    string `gorm:"->" json:"senador_foto,omitempty"`
}

// TableName define o nome da tabela
func (Votacao) TableName() string {
	return "votacoes"
}

// VotacaoStats representa estatisticas de votacao de um senador.
// Classificacao dos codigos em classificacao.go.
type VotacaoStats struct {
	SenadorID             int     `json:"senador_id"`
	TotalVotacoes         int     `json:"total_votacoes"`         // registros no periodo, de qualquer tipo
	VotosRegistrados      int     `json:"votos_registrados"`      // Sim + Nao + Abstencao
	Presentes             int     `json:"presentes"`              // inclui Votou (secreta), P-NRV, presidencia
	Ausencias             int     `json:"ausencias"`              // AP + NCom: contam como falta
	AusenciasAP           int     `json:"ausencias_ap"`           // "Atividade parlamentar", declarada pelo senador
	NaoCompareceu         int     `json:"nao_compareceu"`         // NCom
	AusenciasJustificadas int     `json:"ausencias_justificadas"` // licencas e missoes: fora do denominador em B
	Obstrucoes            int     `json:"obstrucoes"`
	PresencaBruta         float64 `json:"presenca_bruta"`    // A: presentes / (total - NA)
	PresencaAjustada      float64 `json:"presenca_ajustada"` // B: presentes / (total - NA - justificadas)
	TaxaPresenca          float64 `json:"taxa_presenca"`     // = B (metrica oficial, 0-100)
	TaxaParticipacao      float64 `json:"taxa_participacao"` // votos efetivos / denominador de B
	DadosSuficientes      bool    `json:"dados_suficientes"` // false: nenhum registro que conte no periodo
}

// VotosPorTipo representa contagem de votos por tipo
type VotosPorTipo struct {
	Voto  string `json:"voto"`
	Total int    `json:"total"`
}
