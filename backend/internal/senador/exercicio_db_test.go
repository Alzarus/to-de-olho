package senador

import (
	"testing"
	"time"

	"github.com/Alzarus/to-de-olho/internal/testdb"
)

// Item 14: o ranking da legislatura encerrada usa quem ocupava a cadeira no
// ultimo dia dela, mesmo que ja tenha saido (em_exercicio = false).
func TestFindEmExercicioEm(t *testing.T) {
	db := testdb.Abrir(t, &Senador{}, &Mandato{})
	repo := NewRepository(db)
	dia := func(s string) time.Time { v, _ := time.Parse("2006-01-02", s); return v }
	fim := func(s string) *time.Time { v := dia(s); return &v }

	senadores := []Senador{
		{CodigoParlamentar: 1, Nome: "Saiu em fevereiro", EmExercicio: false},
		{CodigoParlamentar: 2, Nome: "Reeleito", EmExercicio: true},
		{CodigoParlamentar: 3, Nome: "Novo na 58a", EmExercicio: true},
		{CodigoParlamentar: 4, Nome: "Saiu em 2025", EmExercicio: false},
	}
	if err := db.Create(&senadores).Error; err != nil {
		t.Fatal(err)
	}
	mandatos := []Mandato{
		{SenadorID: senadores[0].ID, Inicio: dia("2019-02-01")}, // fim nao atualizado depois de sair
		{SenadorID: senadores[1].ID, Inicio: dia("2019-02-01"), Fim: fim("2027-01-31")},
		{SenadorID: senadores[1].ID, Inicio: dia("2027-02-01")},
		{SenadorID: senadores[2].ID, Inicio: dia("2027-02-01")},
		{SenadorID: senadores[3].ID, Inicio: dia("2023-02-01"), Fim: fim("2025-06-30")},
	}
	if err := db.Create(&mandatos).Error; err != nil {
		t.Fatal(err)
	}

	got, err := repo.FindEmExercicioEm(dia("2027-01-31"))
	if err != nil {
		t.Fatal(err)
	}
	var nomes []string
	for _, s := range got {
		nomes = append(nomes, s.Nome)
	}
	if len(nomes) != 2 || nomes[0] != "Reeleito" || nomes[1] != "Saiu em fevereiro" {
		t.Errorf("em exercicio em 31/01/2027: %v", nomes)
	}
}
