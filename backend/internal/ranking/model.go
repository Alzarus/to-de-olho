package ranking

import "time"

// SenadorScore representa o score completo de um senador
type SenadorScore struct {
	SenadorID int     `json:"senador_id"`
	Nome      string  `json:"nome"`
	Partido   string  `json:"partido"`
	UF        string  `json:"uf"`
	FotoURL   string  `json:"foto_url,omitempty"`

	// Scores individuais normalizados (0-100)
	Produtividade float64 `json:"produtividade"`
	Presenca      float64 `json:"presenca"`
	EconomiaCota  float64 `json:"economia_cota"`
	Comissoes     float64 `json:"comissoes"`

	// Score final ponderado (0-100)
	ScoreFinal float64 `json:"score_final"`
	Posicao    int     `json:"posicao"`

	// Detalhes para transparencia/auditoria
	Detalhes    ScoreDetalhes `json:"detalhes"`
	CalculadoEm time.Time     `json:"calculado_em"`
}

// ScoreDetalhes contem os dados brutos utilizados no calculo
type ScoreDetalhes struct {
	// Produtividade
	TotalProposicoes     int `json:"total_proposicoes"`
	ProposicoesAprovadas int `json:"proposicoes_aprovadas"`
	TransformadasEmLei   int `json:"transformadas_em_lei"`
	PontuacaoProposicoes int `json:"pontuacao_proposicoes"`

	// Presenca
	TotalVotacoes        int     `json:"total_votacoes"`
	VotacoesParticipadas int     `json:"votacoes_participadas"`
	TaxaPresencaBruta    float64 `json:"taxa_presenca_bruta"`

	// Economia CEAPS
	GastoCEAPS float64 `json:"gasto_ceaps"`
	TetoCEAPS  float64 `json:"teto_ceaps"`

	// Comissoes
	ComissoesAtivas   int     `json:"comissoes_ativas"`
	ComissoesTitular  int     `json:"comissoes_titular"`
	ComissoesSuplente int     `json:"comissoes_suplente"`
	PontosComissoes   float64 `json:"pontos_comissoes"`
}

// RankingResponse representa a resposta do endpoint de ranking
type RankingResponse struct {
	Ranking     []SenadorScore `json:"ranking"`
	Total       int            `json:"total"`
	CalculadoEm time.Time      `json:"calculado_em"`
	Metodologia string         `json:"metodologia"`
}

// Pesos dos criterios conforme metodologia-ranking.md
const (
	PesoProdutividade = 0.35
	PesoPresenca      = 0.25
	PesoEconomia      = 0.20
	PesoComissoes     = 0.20

	// Teto CEAPS anual estimado (media nacional * 12 meses)
	TetoCEAPSMedia = 40000.0 * 12
)

// TetoCEAPSPorUF define o valor mensal do teto por estado (referencia 2024)
var TetoCEAPSPorUF = map[string]float64{
	"DF": 33178.69,
	"GO": 38166.97,
	"SC": 42091.22,
	"SP": 40381.18,
	"RJ": 39599.93,
	"MG": 39078.68,
	"PR": 41851.68,
	"RS": 43533.91,
	"AC": 44670.36, "AL": 40995.53, "AM": 44092.38, "AP": 43574.61,
	"BA": 40228.75, "CE": 41348.70, "ES": 37637.52, "MA": 41785.49,
	"MT": 39434.72, "MS": 40049.91, "PA": 41656.76, "PB": 40702.40,
	"PE": 41042.81, "PI": 40954.10, "RN": 40875.05, "RO": 43632.74,
	"RR": 44321.49, "SE": 40166.45, "TO": 39458.74,
}
