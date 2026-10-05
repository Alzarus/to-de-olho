package votacao

import (
	"testing"
	"time"

	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/internal/testdb"
)

func TestEhMerito(t *testing.T) {
	casos := map[string]bool{
		"Votação nominal do Projeto de Lei Complementar nº 177, de 2023, nos termos do parecer.": true,
		"Votação nominal da Proposta de Emenda à Constituição nº 45, de 2019, em segundo turno.": true,
		"Votação nominal da Medida Provisória nº 1.185, de 2023.":                                true,
		"Votação nominal da Emenda nº 2.203 ao Substitutivo do Relator ao Projeto.":              false,
		"Votação nominal do Requerimento 965, de 2023, de adiamento da votação.":                 false,
		"Votação nominal do Art. 28 do Projeto de Lei nº 2.903, de 2023, destacado.":             false,
		"Votação nominal da Emenda nº 806, destacada.":                                           false,
		"Votação nominal do Requerimento nº 857, de 2024, que solicita urgência":                 false,
	}
	for descricao, esperado := range casos {
		if got := EhMerito(descricao); got != esperado {
			t.Errorf("EhMerito(%q) = %v, queria %v", descricao, got, esperado)
		}
	}
}

func placar(codigo int, dia string, materia, sigla, descricao string, sim, nao int) Placar {
	data, _ := time.Parse("2006-01-02", dia)
	return Placar{CodigoVotacao: codigo, Data: data, Materia: materia, SiglaMateria: sigla,
		DescricaoVotacao: descricao, Sim: sim, Nao: nao}
}

func TestSelecionarMatch(t *testing.T) {
	const merito = "Votação nominal do Projeto de Lei nº 1, nos termos do parecer."
	const pec1 = "Votação nominal da Proposta de Emenda à Constituição nº 8, em primeiro turno."
	const pec2 = "Votação nominal da Proposta de Emenda à Constituição nº 8, em segundo turno."
	entrada := []Placar{
		placar(1, "2024-01-10", "PL 1/2024", "PL", merito, 37, 32),                            // disputada: entra
		placar(2, "2024-02-10", "PL 2/2024", "PL", merito, 60, 2),                             // quase unanime: sai
		placar(3, "2024-03-10", "PL 3/2024", "PL", merito, 20, 15),                            // menos de 40 votos: sai
		placar(4, "2024-04-10", "PEC 8/2021", "PEC", pec1, 50, 16),                            // primeiro turno: sai
		placar(5, "2024-04-11", "PEC 8/2021", "PEC", pec2, 52, 16),                            // segundo turno: entra
		placar(6, "2024-05-10", "PL 1/2024", "PL", merito, 30, 30),                            // mesma materia, mais recente: troca a 1
		placar(7, "2024-06-10", "PL 7/2024", "PL", "Votação nominal da Emenda nº 3.", 30, 30), // emenda: sai
	}
	got := SelecionarMatch(entrada)
	codigos := []int{}
	for _, p := range got {
		codigos = append(codigos, p.CodigoVotacao)
	}
	if len(codigos) != 2 || codigos[0] != 5 || codigos[1] != 6 {
		t.Fatalf("selecao = %v, queria [5 6] em ordem cronologica", codigos)
	}
}

func TestMinoria(t *testing.T) {
	if m := (Placar{Sim: 45, Nao: 15}).Minoria(); m != 0.25 {
		t.Errorf("Minoria = %v, queria 0.25", m)
	}
	if m := (Placar{}).Minoria(); m != 0 {
		t.Errorf("Minoria sem votos = %v, queria 0", m)
	}
}

func TestPlacaresESenadoresMatch(t *testing.T) {
	db := testdb.Abrir(t, &senador.Senador{}, &Votacao{})
	repo := NewRepository(db)
	db.Create(&[]senador.Senador{
		{ID: 1, CodigoParlamentar: 1, Nome: "Ana", NomeCompleto: "Ana Maria Souza", UF: "BA", Partido: "X"},
		{ID: 2, CodigoParlamentar: 2, Nome: "Bia", NomeCompleto: "Beatriz Lima", UF: "SP", Partido: "Y"},
	})
	aberta, secreta := false, true
	data := time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC)
	votos := []Votacao{
		{SenadorID: 1, CodigoVotacao: 10, SessaoID: "s", Data: data, SiglaVoto: "Sim", Voto: "Sim", Secreta: &aberta, Materia: "PL 1/2024"},
		{SenadorID: 2, CodigoVotacao: 10, SessaoID: "s", Data: data, SiglaVoto: "Não", Voto: "Nao", Secreta: &aberta, Materia: "PL 1/2024"},
		{SenadorID: 1, CodigoVotacao: 11, SessaoID: "s", Data: data, SiglaVoto: "Votou", Voto: "Votou", Secreta: &secreta},
	}
	if err := db.Create(&votos).Error; err != nil {
		t.Fatal(err)
	}

	placares, err := repo.PlacaresAbertos()
	if err != nil {
		t.Fatal(err)
	}
	if len(placares) != 1 || placares[0].Sim != 1 || placares[0].Nao != 1 {
		t.Fatalf("placares = %+v, queria so a votacao aberta 10 com 1x1", placares)
	}

	senadores, err := repo.SenadoresMatch([]int{10})
	if err != nil {
		t.Fatal(err)
	}
	if len(senadores) != 2 || senadores[0].NomeCompleto != "Ana Maria Souza" || senadores[1].Votos["10"] != "Nao" {
		t.Errorf("senadores = %+v", senadores)
	}
}
