package ranking

import "time"

// SenadorScore representa o score completo de um senador
type SenadorScore struct {
	SenadorID int    `json:"senador_id"`
	Nome      string `json:"nome"`
	Partido   string `json:"partido"`
	UF        string `json:"uf"`
	FotoURL   string `json:"foto_url,omitempty"`
	Cargo     string `json:"cargo,omitempty"`
	Titular   string `json:"titular,omitempty"`

	// Scores individuais normalizados (0-100)
	Produtividade float64  `json:"produtividade"`
	Presenca      *float64 `json:"presenca"` // null: sem registro de votacao no periodo (item 4)
	EconomiaCota  float64  `json:"economia_cota"`
	Comissoes     float64  `json:"comissoes"`

	// Score final ponderado (0-100)
	ScoreFinal float64 `json:"score_final"`
	Posicao    int     `json:"posicao"` // 0 quando fora da ordenacao

	// DadosInsuficientes: sem dado para algum criterio. O senador fica fora da
	// ordenacao (D4): posicao 81 seria uma afirmacao que o dado nao sustenta.
	DadosInsuficientes bool   `json:"dados_insuficientes"`
	Motivo             string `json:"motivo,omitempty"` // por que ficou fora da ordenacao

	// Detalhes para transparencia/auditoria
	Detalhes    ScoreDetalhes `json:"detalhes"`
	CalculadoEm time.Time     `json:"calculado_em"`
}

// ScoreDetalhes contem os dados brutos utilizados no calculo
type ScoreDetalhes struct {
	// Produtividade
	TotalProposicoes     int     `json:"total_proposicoes"` // autoria principal
	TotalCoautorias      int     `json:"total_coautorias"`  // nao pontuam (item 9)
	TotalSemPontos       int     `json:"total_sem_pontos"`  // vetos, autoria como deputado ou institucional
	ProposicoesAprovadas int     `json:"proposicoes_aprovadas"`
	TransformadasEmLei   int     `json:"transformadas_em_lei"`
	PontuacaoProposicoes float64 `json:"pontuacao_proposicoes"`

	// Presenca
	TotalVotacoes         int     `json:"total_votacoes"`
	VotacoesParticipadas  int     `json:"votacoes_participadas"` // Sim + Nao + Abstencao
	Presentes             int     `json:"presentes"`
	AusenciasAP           int     `json:"ausencias_ap"`
	NaoCompareceu         int     `json:"nao_compareceu"`
	AusenciasJustificadas int     `json:"ausencias_justificadas"`
	TaxaPresencaBruta     float64 `json:"taxa_presenca_bruta"`    // A
	TaxaPresencaAjustada  float64 `json:"taxa_presenca_ajustada"` // B, a que entra no score

	// Economia CEAPS
	GastoCEAPS     float64 `json:"gasto_ceaps"`
	TetoCEAPS      float64 `json:"teto_ceaps"`      // teto mensal da UF x meses em exercicio
	MesesExercicio float64 `json:"meses_exercicio"` // no periodo (item 8)

	// Comissoes
	ComissoesAtivas      int     `json:"comissoes_ativas"`
	ComissoesTitular     int     `json:"comissoes_titular"`
	ComissoesSuplente    int     `json:"comissoes_suplente"`
	PontosComissoes      float64 `json:"pontos_comissoes"`
	ComissoesForaDaConta int     `json:"comissoes_fora_da_conta"` // frentes, grupos de amizade, honrarias
}

// RankingResponse representa a resposta do endpoint de ranking
type RankingResponse struct {
	Ranking     []SenadorScore `json:"ranking"`
	SemDados    []SenadorScore `json:"sem_dados"` // fora da ordenacao, "dados insuficientes"
	Total       int            `json:"total"`     // len(Ranking)
	CalculadoEm time.Time      `json:"calculado_em"`
	Metodologia string         `json:"metodologia"`
}

// Pesos dos criterios conforme metodologia-ranking.md
const (
	PesoProdutividade = 0.35
	PesoPresenca      = 0.25
	PesoEconomia      = 0.20
	PesoComissoes     = 0.20

	// PisoMesesExercicio: com menos que isso em exercicio no periodo, o senador
	// fica fora da ordenacao (decisao D6, 23/09/2026)
	PisoMesesExercicio = 6.0

	// Teto CEAPS anual estimado (media nacional * 12 meses)
	TetoCEAPSMedia = 40000.0 * 12
)

// TetoCEAPSPorUF define o valor mensal do teto por estado (referencia marco 2025, reajuste 12%)
// Fonte: Senado Federal - Ato da Comissao Diretora (atualizado 10/01/2026)
// Media nacional: R$ 46.402,62/mes
var TetoCEAPSPorUF = map[string]float64{
	"AC": 50426.26, "AL": 44500.00, "AM": 52798.82, "AP": 51103.82,
	"BA": 45000.00, "CE": 48245.57, "DF": 36582.46, "ES": 42000.00,
	"GO": 36582.46, "MA": 47500.00, "MG": 40000.00, "MS": 42000.00,
	"MT": 44500.00, "PA": 48207.30, "PB": 45000.00, "PE": 46000.00,
	"PI": 49000.00, "PR": 43000.00, "RJ": 42000.00, "RN": 46000.00,
	"RO": 44000.00, "RR": 51500.00, "RS": 45500.00, "SC": 42000.00,
	"SE": 53000.00, "SP": 40000.00, "TO": 36582.46,
}
