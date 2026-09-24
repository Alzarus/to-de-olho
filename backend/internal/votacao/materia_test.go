package votacao

import (
	"strconv"
	"testing"

	"github.com/Alzarus/to-de-olho/internal/materia"
	"github.com/Alzarus/to-de-olho/internal/proposicao"
	"github.com/Alzarus/to-de-olho/internal/testdb"
)

func strPtr(s string) *string { return &s }
func intPtr(v int) *int       { return &v }

// Tres votacoes: uma com apelido oficial (e curadoria, que perde), uma so com
// curadoria e uma sem nome popular.
func semearMaterias(t *testing.T, repo *Repository) {
	t.Helper()
	var lote []Votacao
	for _, c := range []struct {
		codigoVotacao, dia, codigoMateria int
		ident, sigla                      string
	}{
		{1, 10, 157233, "PL 2338/2023", "PL"},
		{2, 11, 158930, "PEC 45/2019", "PEC"},
		{3, 12, 500, "PLP 1/2025", "PLP"},
	} {
		for _, v := range votacaoCom(c.codigoVotacao, c.dia, c.ident, c.sigla, "A", boolPtr(false), "Sim", "Não") {
			v.CodigoMateria = intPtr(c.codigoMateria)
			lote = append(lote, v)
		}
	}
	if err := repo.UpsertBatch(lote); err != nil {
		t.Fatal(err)
	}
	materias := []materia.Materia{
		{CodigoMateria: 157233, IdProcesso: 8441243, Identificacao: "PL 2338/2023", Apelido: strPtr("Marco Legal da Inteligência Artificial"),
			ExplicacaoEmenta: strPtr("Explicação oficial."), Temas: materia.Temas{"Ciência, Tecnologia e Informática"}},
		{CodigoMateria: 158930, IdProcesso: 8503515, Identificacao: "PEC 45/2019", Temas: materia.Temas{"Tributos"}},
		{CodigoMateria: 500, IdProcesso: 1, Identificacao: "PLP 1/2025"},
	}
	if err := repo.db.Create(&materias).Error; err != nil {
		t.Fatal(err)
	}
	curados := []materia.ApelidoCurado{
		{CodigoMateria: 157233, Identificacao: "PL 2338/2023", Apelido: "Nome curado que perde", FonteURL: "https://www12.senado.leg.br/a"},
		{CodigoMateria: 158930, Identificacao: "PEC 45/2019", Apelido: "Reforma tributária", FonteURL: "https://www12.senado.leg.br/b"},
	}
	if err := repo.db.Create(&curados).Error; err != nil {
		t.Fatal(err)
	}
}

func deref(p *string) string {
	if p == nil {
		return "<nil>"
	}
	return *p
}

func TestJoinMateriasNasListas(t *testing.T) {
	repo := novoRepo(t)
	semearMaterias(t, repo)

	type esperado struct{ apelido, fonte, url, explicacao string }
	porCodigo := map[int]esperado{
		1: {"Marco Legal da Inteligência Artificial", "oficial", "<nil>", "Explicação oficial."},
		2: {"Reforma tributária", "curadoria", "https://www12.senado.leg.br/b", "<nil>"},
		3: {"<nil>", "<nil>", "<nil>", "<nil>"},
	}
	conferir := func(t *testing.T, onde string, v Votacao) {
		t.Helper()
		e := porCodigo[v.CodigoVotacao]
		got := esperado{deref(v.Apelido), deref(v.ApelidoFonte), deref(v.ApelidoFonteURL), deref(v.ExplicacaoEmenta)}
		if got != e {
			t.Errorf("%s, votacao %d: %+v; esperado %+v", onde, v.CodigoVotacao, got, e)
		}
	}

	lista, total, err := repo.FindAll(10, 0, FiltroLista{})
	if err != nil || total != 3 {
		t.Fatalf("FindAll: total=%d err=%v", total, err)
	}
	for _, v := range lista {
		conferir(t, "FindAll", v)
	}
	if len(lista[2].Temas) != 1 || lista[2].Temas[0] != "Ciência, Tecnologia e Informática" {
		t.Errorf("temas da votacao 1: %v", lista[2].Temas)
	}

	for codigo := 1; codigo <= 3; codigo++ {
		v, err := repo.FindByID(codigo)
		if err != nil {
			t.Fatal(err)
		}
		conferir(t, "FindByID", *v)
	}

	votos, total, err := repo.FindBySenadorID(1, 10, 0, "", 0)
	if err != nil || total != 3 {
		t.Fatalf("FindBySenadorID: total=%d err=%v", total, err)
	}
	for _, v := range votos {
		conferir(t, "FindBySenadorID", v)
	}

	// busca pelo apelido (oficial e curado)
	casos := []struct {
		busca    string
		esperado []int
	}{
		{"inteligência artificial", []int{1}},
		{"REFORMA TRIBUT", []int{2}},
		{"nome curado que perde", []int{1}}, // a curadoria tambem e buscavel
		{"PLP 1/2025", []int{3}},
		{"inexistente", nil},
	}
	for _, c := range casos {
		lista, total, err := repo.FindAll(10, 0, FiltroLista{Materia: c.busca})
		if err != nil {
			t.Fatal(err)
		}
		var got []int
		for _, v := range lista {
			got = append(got, v.CodigoVotacao)
		}
		if int(total) != len(c.esperado) || len(got) != len(c.esperado) || (len(got) > 0 && got[0] != c.esperado[0]) {
			t.Errorf("busca %q: %v (total %d); esperado %v", c.busca, got, total, c.esperado)
		}
	}
}

func TestPreencherCodigoMateriaIdempotente(t *testing.T) {
	db := testdb.Abrir(t, &Votacao{}, &proposicao.Proposicao{}, &materia.Materia{})
	repo := NewRepository(db)
	var lote []Votacao
	lote = append(lote, votacaoCom(1, 10, "PL 2338/2023", "PL", "A", boolPtr(false), "Sim")...)
	lote = append(lote, votacaoCom(2, 11, "PLP 121/2024 (Substitutivo-CD)", "PLP", "A", boolPtr(false), "Sim")...)
	lote = append(lote, votacaoCom(3, 12, "PEC 45/2019", "PEC", "A", boolPtr(false), "Sim")...)
	lote = append(lote, votacaoCom(4, 13, "PL 9/2025", "PL", "A", boolPtr(false), "Sim")...)
	ja := votacaoCom(5, 14, "PL 2338/2023", "PL", "A", boolPtr(false), "Sim")
	ja[0].CodigoMateria = intPtr(777) // gravado pelo sync: nao muda
	lote = append(lote, ja...)
	if err := repo.UpsertBatch(lote); err != nil {
		t.Fatal(err)
	}
	props := []proposicao.Proposicao{
		{SenadorID: 1, CodigoMateria: "157233", DescricaoIdentificacao: "PL 2338/2023"},
		{SenadorID: 2, CodigoMateria: "157233", DescricaoIdentificacao: "PL 2338/2023"}, // coautoria: mesmo codigo
		{SenadorID: 1, CodigoMateria: "164599", DescricaoIdentificacao: "PLP 121/2024"}, // sem sufixo: nao casa
		{SenadorID: 1, CodigoMateria: "1", DescricaoIdentificacao: "PL 9/2025"},
		{SenadorID: 2, CodigoMateria: "2", DescricaoIdentificacao: "PL 9/2025"}, // ambiguo: nao preenche
	}
	if err := db.Create(&props).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&materia.Materia{CodigoMateria: 158930, IdProcesso: 8503515, Identificacao: "PEC 45/2019"}).Error; err != nil {
		t.Fatal(err)
	}

	esperado := map[int]string{1: "157233", 2: "<nil>", 3: "158930", 4: "<nil>", 5: "777"}
	for rodada := 1; rodada <= 2; rodada++ {
		if err := PreencherCodigoMateria(db); err != nil {
			t.Fatal(err)
		}
		var linhas []Votacao
		db.Order("codigo_votacao").Find(&linhas)
		for _, l := range linhas {
			got := "<nil>"
			if l.CodigoMateria != nil {
				got = strconv.Itoa(*l.CodigoMateria)
			}
			if got != esperado[l.CodigoVotacao] {
				t.Errorf("rodada %d, votacao %d: codigo_materia %s; esperado %s", rodada, l.CodigoVotacao, got, esperado[l.CodigoVotacao])
			}
		}
	}
	n, err := repo.ContarSemCodigoMateria(lote[0].Data.AddDate(-1, 0, 0))
	if err != nil || n != 2 {
		t.Errorf("sem codigo_materia: %d %v; esperado 2", n, err)
	}
}
