package proposicao

import (
	"testing"
	"time"

	"github.com/Alzarus/to-de-olho/internal/materia"
	"github.com/Alzarus/to-de-olho/internal/testdb"
)

func TestFindBySenadorIDFiltraAutoria(t *testing.T) {
	db := testdb.Abrir(t, &Proposicao{}, &materia.Materia{}, &materia.ApelidoCurado{})
	repo := NewRepository(db)
	data := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	nova := func(codigo, sigla string, posicao *int, tipoAutor string) Proposicao {
		return Proposicao{SenadorID: 1, CodigoMateria: codigo, SiglaSubtipoMateria: sigla,
			PosicaoAutoria: posicao, TipoAutor: tipoAutor, DataApresentacao: &data}
	}
	lote := []Proposicao{
		nova("1", "PL", ptr(1), "SENADOR"),  // principal
		nova("2", "RQS", ptr(1), "LIDER"),   // principal, como lider
		nova("3", "PEC", ptr(4), "SENADOR"), // coautoria
		nova("4", "VET", ptr(1), "SENADOR"), // veto: nem principal nem coautoria
		nova("5", "PL", ptr(1), "DEPUTADO"), // autoria como deputado
		nova("6", "REQ", nil, ""),           // institucional
	}
	if err := repo.UpsertBatch(lote); err != nil {
		t.Fatal(err)
	}

	casos := []struct {
		autoria string
		total   int64
		tipos   map[string]int
	}{
		{"", 6, map[string]int{"PL": 2, "RQS": 1, "PEC": 1, "VET": 1, "REQ": 1}},
		{"principal", 2, map[string]int{"PL": 1, "RQS": 1}},
		{"coautoria", 1, map[string]int{"PEC": 1}},
		{"qualquer", 6, nil}, // valor desconhecido: sem filtro
	}
	for _, c := range casos {
		t.Run(c.autoria, func(t *testing.T) {
			_, total, err := repo.FindBySenadorID(1, 20, 0, "", 0, "", "", "", c.autoria)
			if err != nil {
				t.Fatal(err)
			}
			if total != c.total {
				t.Errorf("total = %d, queria %d", total, c.total)
			}
			if c.tipos == nil {
				return
			}
			tipos, err := repo.ContarPorSigla(1, c.autoria)
			if err != nil {
				t.Fatal(err)
			}
			got := map[string]int{}
			for _, tp := range tipos {
				got[tp.Tipo] = int(tp.Total)
			}
			if len(got) != len(c.tipos) {
				t.Errorf("tipos = %v, queria %v", got, c.tipos)
			}
			for k, v := range c.tipos {
				if got[k] != v {
					t.Errorf("tipos[%s] = %d, queria %d", k, got[k], v)
				}
			}
		})
	}

	// Combina com o filtro de sigla
	_, total, err := repo.FindBySenadorID(1, 20, 0, "", 0, "PL", "", "", "principal")
	if err != nil || total != 1 {
		t.Errorf("PL principal: total = %d, err = %v; queria 1", total, err)
	}
}
