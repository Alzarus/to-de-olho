package domain

import (
	"context"
	"errors"
	"time"
)

// Votacao representa uma votação na Câmara dos Deputados
type Votacao struct {
	ID                    int64                  `json:"id"`
	IDVotacaoCamara       int64                  `json:"idVotacaoCamara"`
	Titulo                string                 `json:"titulo"`
	Ementa                string                 `json:"ementa"`
	DataVotacao           time.Time              `json:"dataVotacao"`
	Aprovacao             string                 `json:"aprovacao"` // "Aprovada", "Rejeitada"
	PlacarSim             int                    `json:"placarSim"`
	PlacarNao             int                    `json:"placarNao"`
	PlacarAbstencao       int                    `json:"placarAbstencao"`
	PlacarOutros          int                    `json:"placarOutros"`
	IDProposicaoPrincipal *int64                 `json:"idProposicaoPrincipal,omitempty"`
	TipoProposicao        string                 `json:"tipoProposicao"`
	NumeroProposicao      string                 `json:"numeroProposicao"`
	AnoProposicao         *int                   `json:"anoProposicao,omitempty"`
	Relevancia            string                 `json:"relevancia"` // "alta", "média", "baixa"
	Payload               map[string]interface{} `json:"payload"`
	CreatedAt             time.Time              `json:"createdAt"`
	UpdatedAt             time.Time              `json:"updatedAt"`
}

// VotoDeputado representa o voto individual de um deputado
type VotoDeputado struct {
	ID            int64                  `json:"id"`
	IDVotacao     int64                  `json:"idVotacao"`
	IDDeputado    int                    `json:"idDeputado"`
	Voto          string                 `json:"voto"` // "Sim", "Não", "Abstenção", "Obstrução", "Art17"
	Justificativa *string                `json:"justificativa,omitempty"`
	Payload       map[string]interface{} `json:"payload,omitempty"`
	CreatedAt     time.Time              `json:"createdAt"`
}

// OrientacaoPartido representa a orientação oficial de um partido
type OrientacaoPartido struct {
	ID         int64     `json:"id"`
	IDVotacao  int64     `json:"idVotacao"`
	Partido    string    `json:"partido"`
	Orientacao string    `json:"orientacao"` // "Sim", "Não", "Liberado", "Obstrução"
	CreatedAt  time.Time `json:"createdAt"`
}

// VotacaoDetalhada contém votação completa com votos e orientações
type VotacaoDetalhada struct {
	Votacao     Votacao              `json:"votacao"`
	Votos       []*VotoDeputado      `json:"votos"`
	Orientacoes []*OrientacaoPartido `json:"orientacoes"`
}

// RankingDeputadoVotacao representa ranking de deputados por votações
type RankingDeputadoVotacao struct {
	IDDeputado      int     `json:"idDeputado"`
	TotalVotacoes   int     `json:"totalVotacoes"`
	VotosFavoraveis int     `json:"votosFavoraveis"`
	VotosContrarios int     `json:"votosContrarios"`
	Abstencoes      int     `json:"abstencoes"`
	TaxaAprovacao   float64 `json:"taxaAprovacao"`
}

// Estruturas para análise partidária
type VotacaoPartido struct {
	Partido          string  `json:"partido"`
	Orientacao       string  `json:"orientacao"`
	VotaramFavor     int     `json:"votaramFavor"`
	VotaramContra    int     `json:"votaramContra"`
	VotaramAbstencao int     `json:"votaramAbstencao"`
	TotalMembros     int     `json:"totalMembros"`
	Disciplina       float64 `json:"disciplina"` // % que seguiu orientação
}

// Filtros para consultas
type FiltrosVotacao struct {
	Busca          string `json:"busca,omitempty"`
	Ano            int    `json:"ano,omitempty"`
	Aprovacao      string `json:"aprovacao,omitempty"`
	Relevancia     string `json:"relevancia,omitempty"`
	TipoProposicao string `json:"tipoProposicao,omitempty"`
}

type FiltrosRanking struct {
	Partido    string     `json:"partido,omitempty"`
	DataInicio *time.Time `json:"dataInicio,omitempty"`
	DataFim    *time.Time `json:"dataFim,omitempty"`
}

// Validações e regras de negócio
func (v *Votacao) Validate() error {
	if v.IDVotacaoCamara <= 0 {
		return errors.New("ID da votação na Câmara é obrigatório")
	}

	if v.Titulo == "" {
		return errors.New("título da votação é obrigatório")
	}

	if v.DataVotacao.IsZero() {
		return errors.New("data da votação é obrigatória")
	}

	if v.Aprovacao != "Aprovada" && v.Aprovacao != "Rejeitada" {
		return errors.New("aprovação deve ser 'Aprovada' ou 'Rejeitada'")
	}

	if v.Relevancia != "alta" && v.Relevancia != "média" && v.Relevancia != "baixa" {
		return errors.New("relevância deve ser 'alta', 'média' ou 'baixa'")
	}

	if v.PlacarSim < 0 || v.PlacarNao < 0 || v.PlacarAbstencao < 0 {
		return errors.New("placares não podem ser negativos")
	}

	return nil
}

func (vd *VotoDeputado) Validate() error {
	if vd.IDVotacao <= 0 {
		return errors.New("ID da votação é obrigatório")
	}

	if vd.IDDeputado <= 0 {
		return errors.New("ID do deputado é obrigatório")
	}

	votosValidos := []string{"Sim", "Não", "Abstenção", "Obstrução", "Art17"}
	for _, voto := range votosValidos {
		if vd.Voto == voto {
			return nil
		}
	}

	return errors.New("voto deve ser: Sim, Não, Abstenção, Obstrução ou Art17")
}

func (op *OrientacaoPartido) Validate() error {
	if op.IDVotacao <= 0 {
		return errors.New("ID da votação é obrigatório")
	}

	if op.Partido == "" {
		return errors.New("partido é obrigatório")
	}

	orientacoesValidas := []string{"Sim", "Não", "Liberado", "Obstrução"}
	for _, orientacao := range orientacoesValidas {
		if op.Orientacao == orientacao {
			return nil
		}
	}

	return errors.New("orientação deve ser: Sim, Não, Liberado ou Obstrução")
}

// Métodos auxiliares para análise
func (v *Votacao) TotalVotos() int {
	return v.PlacarSim + v.PlacarNao + v.PlacarAbstencao + v.PlacarOutros
}

func (v *Votacao) PorcentagemAprovacao() float64 {
	total := v.TotalVotos()
	if total == 0 {
		return 0
	}
	return float64(v.PlacarSim) / float64(total) * 100
}

func (v *Votacao) IsRelevante() bool {
	return v.Relevancia == "alta" || v.Relevancia == "média"
}

func (vp *VotacaoPartido) CalcularDisciplina() {
	if vp.TotalMembros == 0 {
		vp.Disciplina = 0
		return
	}

	var seguiramOrientacao int
	switch vp.Orientacao {
	case "Sim":
		seguiramOrientacao = vp.VotaramFavor
	case "Não":
		seguiramOrientacao = vp.VotaramContra
	case "Abstenção":
		seguiramOrientacao = vp.VotaramAbstencao
	case "Liberado":
		// Se liberado, considera 100% de disciplina
		vp.Disciplina = 100.0
		return
	default:
		vp.Disciplina = 0
		return
	}

	vp.Disciplina = float64(seguiramOrientacao) / float64(vp.TotalMembros) * 100
}

// VotacaoRepository define interface para persistência de votações
type VotacaoRepository interface {
	CreateVotacao(ctx context.Context, votacao *Votacao) error
	GetVotacaoByID(ctx context.Context, id int64) (*Votacao, error)
	ListVotacoes(ctx context.Context, filtros FiltrosVotacao, pag Pagination) ([]*Votacao, int, error)
	UpdateVotacao(ctx context.Context, votacao *Votacao) error
	DeleteVotacao(ctx context.Context, id int64) error

	// Métodos para votos de deputados
	CreateVotoDeputado(ctx context.Context, voto *VotoDeputado) error
	GetVotosPorVotacao(ctx context.Context, idVotacao int64) ([]*VotoDeputado, error)
	GetVotoPorDeputado(ctx context.Context, idVotacao int64, idDeputado int) (*VotoDeputado, error)

	// Métodos para orientações partidárias
	CreateOrientacaoPartido(ctx context.Context, orientacao *OrientacaoPartido) error
	GetOrientacoesPorVotacao(ctx context.Context, idVotacao int64) ([]*OrientacaoPartido, error)

	// Métodos para análise completa
	GetVotacaoDetalhada(ctx context.Context, id int64) (*VotacaoDetalhada, error)
	UpsertVotacao(ctx context.Context, votacao *Votacao) error
	// Agregação de presença (votos registrados por deputado em um ano)
	GetPresencaPorDeputadoAno(ctx context.Context, ano int) ([]PresencaCount, error)

	// Consultas agregadas para analytics (evitam N+1 queries)
	GetRankingDeputadosAggregated(ctx context.Context, ano int) ([]RankingDeputadoVotacao, error)
	GetDisciplinaPartidosAggregated(ctx context.Context, ano int) ([]VotacaoPartido, error)
	GetVotacaoStatsAggregated(ctx context.Context, ano int) (*VotacaoStats, error)
}

// VotacaoStats representa estatísticas de votações
type VotacaoStats struct {
	TotalVotacoes         int            `json:"totalVotacoes"`
	VotacoesAprovadas     int            `json:"votacoesAprovadas"`
	VotacoesRejeitadas    int            `json:"votacoesRejeitadas"`
	MediaParticipacao     float64        `json:"mediaParticipacao"`
	VotacoesPorMes        []int          `json:"votacoesPorMes"`
	VotacoesPorRelevancia map[string]int `json:"votacoesPorRelevancia"`
}

// Errors específicos do domínio
var (
	ErrVotacaoNaoEncontrada      = errors.New("votação não encontrada")
	ErrVotoDeputadoNaoEncontrado = errors.New("voto do deputado não encontrado")
	ErrOrientacaoNaoEncontrada   = errors.New("orientação do partido não encontrada")
	ErrVotacaoJaExiste           = errors.New("votação já existe")
	ErrVotoJaRegistrado          = errors.New("voto já registrado para este deputado")
	ErrOrientacaoJaRegistrada    = errors.New("orientação já registrada para este partido")
)
