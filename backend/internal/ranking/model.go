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
	TetoCEAPSAnual = 528000.0 // R$ 44.000/mes * 12
)
