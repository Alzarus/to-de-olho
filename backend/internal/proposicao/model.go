package proposicao

import (
	"time"

	"github.com/Alzarus/to-de-olho/internal/materia"
)

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
	TipoAutor      string `json:"tipo_autor,omitempty"` // SENADOR, LIDER, PRESIDENTE_SF, DEPUTADO (siglaTipo da API)
	Autoria        string `json:"autoria,omitempty"` // texto bruto da API, para auditoria

	// Para calculo de score
	EstagioTramitacao string `json:"estagio_tramitacao"` // Apresentado, EmComissao, AprovadoComissao, AprovadoPlenario, TransformadoLei
	Pontuacao         float64 `json:"pontuacao"`          // Pontos calculados

	// IdProcesso e o id de /processo/{id} (detalhe da materia, tabela materias)
	IdProcesso *int `json:"id_processo,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Nome popular e descricao, via LEFT JOIN de materias (read-only, sem
	// coluna em proposicoes). ApelidoFonte: "oficial" ou "curadoria".
	Apelido          *string       `gorm:"->;-:migration" json:"apelido,omitempty"`
	ApelidoFonte     *string       `gorm:"->;-:migration" json:"apelido_fonte,omitempty"`
	ApelidoFonteURL  *string       `gorm:"->;-:migration" json:"apelido_fonte_url,omitempty"`
	ExplicacaoEmenta *string       `gorm:"->;-:migration" json:"explicacao_ementa,omitempty"`
	Temas            materia.Temas `gorm:"->;-:migration" json:"temas,omitempty"`
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
	TotalSemPontos      int     `json:"total_sem_pontos"`     // vetos, autoria como deputado ou institucional
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

// AutoriaPrincipal indica se a materia pontua para o senador: primeiro autor,
// na condicao de senador (nao de deputado), e nao e veto (veto e ato do
// Presidente da Republica sobre materia ja aprovada).
func (p *Proposicao) AutoriaPrincipal() bool {
	return p.PosicaoAutoria != nil && *p.PosicaoAutoria == 1 && TiposSenador[p.TipoAutor] && p.SiglaSubtipoMateria != "VET"
}

// pesosPorTipo e o multiplicador de cada sigla (metodologia v2.2): segue a
// forca juridica do instrumento e a dificuldade de aprova-lo. Vale zero o que
// nao e iniciativa legislativa do proprio senador (emenda ou substitutivo da
// Camara a mesma materia, oficio, mensagem, peticao, denuncia).
var pesosPorTipo = map[string]float64{
	// Constituicao: 3/5 em dois turnos
	"PEC": 3.0,
	// Lei complementar: maioria absoluta
	"PLP": 2.0,
	// Normas com efeito proprio. PLS e PDS sao as siglas antigas de PL e PDL
	"PL": 1.0, "PLS": 1.0, "PDL": 1.0, "PDS": 1.0, "PRS": 1.0, "PRN": 1.0,
	// Requerimento ao Plenario, mocao e proposta de fiscalizacao e controle
	"RQS": 0.5, "MOC": 0.5, "PFS": 0.5,
	// Requerimentos de comissao e indicacao (sugestao sem efeito vinculante)
	"REQ": 0.1, "INS": 0.1, "RDH": 0.1, "RQN": 0.1, "RAS": 0.1, "RCE": 0.1, "RQJ": 0.1,
	"RMA": 0.1, "RCT": 0.1, "RQE": 0.1, "RQI": 0.1, "RDR": 0.1, "RRA": 0.1, "RRE": 0.1,
	"RFF": 0.1, "RTG": 0.1, "RQR": 0.1, "R.S": 0.1, "R.C": 0.1,
	// Nao e autoria legislativa do senador
	"ECD": 0, "SCD": 0, "OFS": 0, "MSG": 0, "OFN": 0, "PET": 0, "DEN": 0,
	"CON": 0, "SIN": 0, "DIV": 0, "ATS": 0, "PCE": 0, "PRM": 0,
}

// PesoTipo devolve o multiplicador da sigla. Sigla fora da tabela vale zero
// (ok = false): antes da v2.2 ela valia 1, como um projeto de lei, sem
// ninguem decidir isso. O sync registra a sigla nova no log.
func PesoTipo(sigla string) (peso float64, ok bool) {
	peso, ok = pesosPorTipo[sigla]
	return peso, ok
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

	// Multiplicador por tipo de proposicao (metodologia v2.2)
	multiplicador, _ := PesoTipo(p.SiglaSubtipoMateria)

	pontos := pontosBase[p.EstagioTramitacao]
	if pontos == 0 {
		pontos = 1 // Default para apresentado
	}

	return float64(pontos) * multiplicador
}
