package proposicao

import "time"

// Proposicao representa uma proposicao legislativa de autoria de um senador
type Proposicao struct {
	ID                int       `gorm:"primaryKey" json:"id"`
	// Uma linha por (senador, materia): coautorias sao preservadas (item 1).
	// O indice mudou de nome de proposito: o AutoMigrate decide pelo nome, e
	// com o nome antigo nao criaria o indice composto.
	SenadorID         int       `gorm:"uniqueIndex:idx_proposicao_senador_materia,priority:1;index:idx_proposicao_senador;not null" json:"senador_id"`
	CodigoMateria     string    `gorm:"uniqueIndex:idx_proposicao_senador_materia,priority:2;index:idx_proposicao_materia;not null" json:"codigo_materia"`
	SiglaSubtipoMateria string  `json:"sigla_subtipo_materia"` // PEC, PLP, PL, etc.
	NumeroMateria     string    `json:"numero_materia"`
	AnoMateria        int       `json:"ano_materia"`
	DescricaoIdentificacao string `json:"descricao_identificacao"`
	Ementa            string    `json:"ementa,omitempty"`
	SituacaoAtual     string    `json:"situacao_atual,omitempty"` // Em tramitacao, Arquivada, Transformada em Lei
	DataApresentacao  *time.Time `json:"data_apresentacao,omitempty"`

	// Autoria (item 9): so o primeiro autor pontua
	PosicaoAutoria *int   `json:"posicao_autoria"` // 1 = primeiro autor
	TotalAutores   *int   `json:"total_autores"`   // null quando a API so informa "e outros"
	Autoria        string `json:"autoria,omitempty"` // texto bruto da API, para auditoria

	// Para calculo de score
	EstagioTramitacao string `json:"estagio_tramitacao"` // Apresentado, EmComissao, AprovadoComissao, AprovadoPlenario, TransformadoLei
	Pontuacao         float64 `json:"pontuacao"`          // Pontos calculados

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName define o nome da tabela
func (Proposicao) TableName() string {
	return "proposicoes"
}

// ProposicaoStats representa estatisticas de proposicoes de um senador
type ProposicaoStats struct {
	SenadorID           int     `json:"senador_id"`
	TotalProposicoes    int     `json:"total_proposicoes"`    // de autoria principal (primeiro autor)
	TotalCoautorias     int     `json:"total_coautorias"`     // assinadas como coautor: nao pontuam
	TotalPECs           int     `json:"total_pecs"`           // Propostas de Emenda Constitucional
	TotalPLPs           int     `json:"total_plps"`           // Projetos de Lei Complementar
	TotalPLs            int     `json:"total_pls"`            // Projetos de Lei
	TotalOutros         int     `json:"total_outros"`
	TransformadasEmLei  int     `json:"transformadas_em_lei"`
	AprovadosPlenario   int     `json:"aprovados_plenario"`
	EmTramitacao        int     `json:"em_tramitacao"`
	PontuacaoTotal      float64 `json:"pontuacao_total"`      // Score de produtividade
	ScoreNormalizado    float64 `json:"score_normalizado"`    // 0-100
}

// ProposicaoPorTipo representa contagem de proposicoes por tipo
type ProposicaoPorTipo struct {
	Tipo  string `json:"tipo"`
	Total int    `json:"total"`
}

// AutoriaPrincipal indica se o senador e o primeiro autor da materia.
func (p *Proposicao) AutoriaPrincipal() bool {
	return p.PosicaoAutoria != nil && *p.PosicaoAutoria == 1
}

// CalcularPontuacao calcula a pontuacao de uma proposicao baseado no estagio e tipo.
// Coautoria nao pontua (item 9, regra do LES de Volden & Wiseman: conta o
// sponsor, nao o cosponsor).
func (p *Proposicao) CalcularPontuacao() float64 {
	if !p.AutoriaPrincipal() {
		return 0
	}

	// Pontos base por estagio
	pontosBase := map[string]int{
		"Apresentado":       1,
		"EmComissao":        2,
		"AprovadoComissao":  4,
		"AprovadoPlenario":  8,
		"TransformadoLei":   16,
	}

	// Multiplicador por tipo de proposicao
	multiplicador := 1.0
	switch p.SiglaSubtipoMateria {
	case "PEC":
		multiplicador = 3.0
	case "PLP":
		multiplicador = 2.0
	case "RQS", "MOC":
		multiplicador = 0.5
	case "REQ":
		multiplicador = 0.1
	}

	pontos := pontosBase[p.EstagioTramitacao]
	if pontos == 0 {
		pontos = 1 // Default para apresentado
	}

	return float64(pontos) * multiplicador
}
