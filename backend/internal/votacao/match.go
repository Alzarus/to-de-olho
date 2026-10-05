package votacao

import (
	"regexp"
	"sort"
	"strings"
	"time"
)

// Regra publica de selecao das votacoes do "match" (Quem Votar): votacoes
// nominais abertas de merito, em que ao menos MinMinoriaMatch dos votos Sim/Nao
// ficaram do lado vencido. Votacoes quase unanimes nao diferenciam ninguem, e
// emendas, destaques e requerimentos sao procedimentais e dificeis de explicar.
// Nenhuma escolha por tema ou por lado: o criterio e so o placar.
const (
	MinVotosMatch   = 40   // votos Sim + Nao (quorum de uma votacao real de merito)
	MinMinoriaMatch = 0.15 // parcela minima do lado vencido
)

// Placar e o resultado agregado de uma votacao nominal aberta.
type Placar struct {
	CodigoVotacao    int       `json:"codigo_votacao"`
	Data             time.Time `json:"data"`
	Materia          string    `json:"materia"`
	SiglaMateria     string    `json:"sigla_materia"`
	DescricaoVotacao string    `json:"descricao_votacao"`
	Sim              int       `json:"sim"`
	Nao              int       `json:"nao"`
}

// Minoria e a parcela do lado vencido entre os votos Sim e Nao.
func (p Placar) Minoria() float64 {
	total := p.Sim + p.Nao
	if total == 0 {
		return 0
	}
	return float64(min(p.Sim, p.Nao)) / float64(total)
}

var (
	// "Proposta de Emenda a Constituicao" contem "Emenda": trocada antes do filtro
	rePEC          = regexp.MustCompile(`(?i)proposta de emenda [àa] constitui[çc][ãa]o`)
	reProcedimento = regexp.MustCompile(`(?i)emenda|requerimento|destac|art\.|§|par[áa]grafo|inciso|dispositivo|adiamento|urg[êe]ncia|calend[áa]rio|interst`)
	reMerito       = regexp.MustCompile(`(?i)^vota[çc][ãa]o nominal d[ao]s? (projeto|pec|medida|substitutivo)`)
)

// EhMerito indica a votacao do texto principal da materia (e nao de emenda,
// destaque ou requerimento).
func EhMerito(descricao string) bool {
	d := rePEC.ReplaceAllString(strings.TrimSpace(descricao), "PEC")
	return reMerito.MatchString(d) && !reProcedimento.MatchString(d)
}

// SelecionarMatch aplica a regra publica: uma votacao de merito por materia (a
// mais recente; na PEC, o segundo turno), com ao menos MinVotosMatch votos
// Sim/Nao e MinMinoriaMatch do lado vencido. Ordem cronologica.
func SelecionarMatch(placares []Placar) []Placar {
	porMateria := map[string]Placar{}
	for _, p := range placares {
		if !elegivel(p) {
			continue
		}
		if atual, ok := porMateria[p.Materia]; !ok || p.Data.After(atual.Data) {
			porMateria[p.Materia] = p
		}
	}
	out := make([]Placar, 0, len(porMateria))
	for _, p := range porMateria {
		if p.Minoria() >= MinMinoriaMatch {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Data.Before(out[j].Data) })
	return out
}

func elegivel(p Placar) bool {
	if p.Sim+p.Nao < MinVotosMatch || !EhMerito(p.DescricaoVotacao) {
		return false
	}
	primeiroTurno := strings.Contains(strings.ToLower(p.DescricaoVotacao), "primeiro turno")
	return !(p.SiglaMateria == "PEC" && primeiroTurno)
}

// VotacaoMatch e uma pergunta do match, com o texto oficial da materia.
type VotacaoMatch struct {
	Placar
	Titulo           string  `json:"titulo"` // nome popular ou identificacao
	Ementa           string  `json:"ementa"`
	ExplicacaoEmenta *string `json:"explicacao_ementa,omitempty"`
	Resultado        string  `json:"resultado"`
}

// SenadorMatch traz os votos de um senador (ou ex-senador) nas votacoes do match.
// NomeCompleto permite ao Quem Votar achar o senador entre os candidatos.
type SenadorMatch struct {
	ID           int               `json:"id"`
	Nome         string            `json:"nome"`
	NomeCompleto string            `json:"nome_completo"`
	UF           string            `json:"uf"`
	Partido      string            `json:"partido"`
	Votos        map[string]string `json:"votos"` // codigo_votacao -> Sim, Nao, Abstencao...
}

// RespostaMatch e o corpo de GET /api/v1/votacoes/match.
type RespostaMatch struct {
	Regra     string         `json:"regra"`
	Votacoes  []VotacaoMatch `json:"votacoes"`
	Senadores []SenadorMatch `json:"senadores"`
}

// TextoRegraMatch descreve a regra na resposta, para a metodologia publica.
const TextoRegraMatch = "Votações nominais abertas do Plenário do Senado na legislatura, só as de mérito (o texto principal; na PEC, o segundo turno), " +
	"uma por matéria, com ao menos 40 votos Sim/Não e em que ao menos 15% dos votos ficaram do lado vencido. Nenhuma escolha por tema: o critério é só o placar."
