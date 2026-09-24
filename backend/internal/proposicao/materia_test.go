package proposicao

import (
	"testing"
	"time"

	"github.com/Alzarus/to-de-olho/internal/materia"
	"github.com/Alzarus/to-de-olho/internal/testdb"
)

func TestFindBySenadorIDTrazNomePopular(t *testing.T) {
	db := testdb.Abrir(t, &Proposicao{}, &materia.Materia{}, &materia.ApelidoCurado{})
	repo := NewRepository(db)
	d := func(dia int) *time.Time { v := time.Date(2024, 3, dia, 0, 0, 0, 0, time.UTC); return &v }
	lote := []Proposicao{
		{SenadorID: 1, CodigoMateria: "157233", SiglaSubtipoMateria: "PL", DescricaoIdentificacao: "PL 2338/2023", Ementa: "Dispõe sobre IA.", DataApresentacao: d(3)},
		{SenadorID: 1, CodigoMateria: "158930", SiglaSubtipoMateria: "PEC", DescricaoIdentificacao: "PEC 45/2019", Ementa: "Altera o Sistema Tributário.", DataApresentacao: d(2)},
		{SenadorID: 1, CodigoMateria: "abc", SiglaSubtipoMateria: "REQ", DescricaoIdentificacao: "REQ 1/2024", Ementa: "Requer.", DataApresentacao: d(1)},
	}
	if err := repo.UpsertBatch(lote); err != nil {
		t.Fatal(err)
	}
	apelido, explicacao := "Marco Legal da Inteligência Artificial", "Explicação."
	if err := db.Create(&materia.Materia{CodigoMateria: 157233, IdProcesso: 8441243, Identificacao: "PL 2338/2023", Ementa: "outra ementa",
		Apelido: &apelido, ExplicacaoEmenta: &explicacao, Temas: materia.Temas{"Responsabilidade Civil"}}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&materia.ApelidoCurado{CodigoMateria: 158930, Identificacao: "PEC 45/2019", Apelido: "Reforma tributária", FonteURL: "https://www12.senado.leg.br/b"}).Error; err != nil {
		t.Fatal(err)
	}

	casos := []struct {
		nome, busca string
		esperado    []string // apelido|fonte por linha, na ordem
	}{
		{"sem busca", "", []string{"Marco Legal da Inteligência Artificial|oficial", "Reforma tributária|curadoria", "|"}},
		{"busca pelo apelido oficial", "inteligência", []string{"Marco Legal da Inteligência Artificial|oficial"}},
		{"busca pelo apelido curado", "reforma", []string{"Reforma tributária|curadoria"}},
		{"busca pela ementa da proposicao, nao da materia", "outra ementa", nil},
		{"busca pela identificacao", "REQ 1", []string{"|"}},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			lista, total, err := repo.FindBySenadorID(1, 10, 0, c.busca, 0, "", "", "")
			if err != nil {
				t.Fatal(err)
			}
			if int(total) != len(c.esperado) || len(lista) != len(c.esperado) {
				t.Fatalf("total=%d len=%d; esperado %d", total, len(lista), len(c.esperado))
			}
			for i, p := range lista {
				got := deref(p.Apelido) + "|" + deref(p.ApelidoFonte)
				if got != c.esperado[i] {
					t.Errorf("linha %d: %q; esperado %q", i, got, c.esperado[i])
				}
				if p.Ementa == "outra ementa" {
					t.Error("ementa da proposicao trocada pela da materia")
				}
			}
		})
	}
	lista, _, _ := repo.FindBySenadorID(1, 1, 0, "", 0, "", "", "")
	if deref(lista[0].ExplicacaoEmenta) != explicacao || len(lista[0].Temas) != 1 {
		t.Errorf("explicacao/temas: %q %v", deref(lista[0].ExplicacaoEmenta), lista[0].Temas)
	}
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
