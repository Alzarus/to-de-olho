package comissao

import (
	"regexp"
	"time"
)

// Categoria separa colegiados legislativos do que a API de "comissoes" tambem
// devolve e nao e trabalho legislativo (item 6 da auditoria).
type Categoria string

const (
	Colegiado Categoria = "colegiado" // comissoes, subcomissoes, CPI/CPMI, mistas: pontuam
	Frente    Categoria = "frente"    // frentes parlamentares: nao pontuam
	Grupo     Categoria = "grupo"     // grupos parlamentares de amizade/relacionamento: nao pontuam
	Honraria  Categoria = "honraria"  // conselhos de comendas, diplomas e premios: nao pontuam
)

var (
	reFrente   = regexp.MustCompile(`(?i)^\s*frente parlamentar`)
	reGrupo    = regexp.MustCompile(`(?i)^\s*grupo (parlamentar|brasileiro)`)
	reHonraria = regexp.MustCompile(`(?i)(comenda|diploma|pr[eê]mio)`)
)

// Categorizar classifica um colegiado pelo nome.
func Categorizar(nome string) Categoria {
	switch {
	case reFrente.MatchString(nome):
		return Frente
	case reGrupo.MatchString(nome):
		return Grupo
	case reHonraria.MatchString(nome):
		return Honraria
	default:
		return Colegiado
	}
}

// pesoParticipacao: titular (ou membro nato) vale 2; suplente, 1.
func pesoParticipacao(descricao string) int {
	switch descricao {
	case "Titular", "Nato":
		return 2
	default:
		return 1
	}
}

// calcularStats pontua as participacoes que tocam o periodo. Cada colegiado
// conta uma vez, pelo papel mais alto que o senador teve nele no periodo
// (reconducao nao pontua de novo). Mesma formula no ranking anual e no do
// mandato (item 7). "Ativas" e so informativo: colegiados em que o senador
// continuava no fim do periodo.
func calcularStats(senadorID int, participacoes []ComissaoMembro, fim time.Time) *ComissaoStats {
	st := &ComissaoStats{SenadorID: senadorID}
	melhor := map[string]int{}
	ativa := map[string]bool{}
	for _, p := range participacoes {
		if Categorizar(p.NomeComissao) != Colegiado {
			st.ForaDaConta++
			continue
		}
		if peso := pesoParticipacao(p.DescricaoParticipacao); peso > melhor[p.CodigoComissao] {
			melhor[p.CodigoComissao] = peso
		}
		if p.DataFim == nil || !p.DataFim.Before(fim) {
			ativa[p.CodigoComissao] = true
		}
	}
	for _, peso := range melhor {
		st.TotalComissoes++
		st.Pontos += peso
		if peso == 2 {
			st.ComissoesTitular++
		} else {
			st.ComissoesSuplente++
		}
	}
	st.ComissoesAtivas = len(ativa)
	if st.TotalComissoes > 0 {
		st.TaxaTitularidade = float64(st.ComissoesTitular) / float64(st.TotalComissoes) * 100
	}
	return st
}
