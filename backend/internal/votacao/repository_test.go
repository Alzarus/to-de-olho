package votacao

import (
	"testing"
	"time"

	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/internal/testdb"
)

func novoRepo(t *testing.T) *Repository {
	db := testdb.Abrir(t, &senador.Senador{}, &Votacao{})
	senadores := []senador.Senador{
		{ID: 1, CodigoParlamentar: 742, Nome: "Marcelo Castro", EmExercicio: true},
		{ID: 2, CodigoParlamentar: 5672, Nome: "Alan Rick", EmExercicio: true},
		{ID: 3, CodigoParlamentar: 1, Nome: "Sem Votos", EmExercicio: true},
	}
	if err := db.Create(&senadores).Error; err != nil {
		t.Fatal(err)
	}
	return NewRepository(db)
}

func voto(senadorID, codigo int, sessao string, dia int, sigla string) Votacao {
	return Votacao{
		SenadorID: senadorID, CodigoVotacao: codigo, SessaoID: sessao, CodigoSessao: sessao,
		Data:      time.Date(2025, 8, dia, 12, 0, 0, 0, time.UTC),
		SiglaVoto: sigla, Voto: rotuloVoto(sigla), Materia: "PL 1/2025",
	}
}

func TestVotacoesDaMesmaSessaoNaoColapsam(t *testing.T) {
	repo := novoRepo(t)
	// sessao 473484 com 3 votacoes: a chave antiga guardava so a ultima
	lote := []Votacao{
		voto(1, 101, "473484", 19, "Sim"), voto(1, 102, "473484", 19, "Não"), voto(1, 103, "473484", 19, "AP"),
		voto(2, 101, "473484", 19, "Sim"), voto(2, 102, "473484", 19, "Sim"), voto(2, 103, "473484", 19, "Sim"),
		voto(1, 201, "480000", 20, "LS"),
	}
	if err := repo.UpsertBatch(lote); err != nil {
		t.Fatal(err)
	}
	// recarga: idempotente e atualiza o voto
	lote[0].SiglaVoto, lote[0].Voto = "Abstenção", "Abstencao"
	if err := repo.UpsertBatch(lote); err != nil {
		t.Fatal(err)
	}

	n, _ := repo.CountBySenadorID(1)
	if n != 4 {
		t.Errorf("senador 1: esperado 4 votos, obtido %d", n)
	}

	lista, total, err := repo.FindAll(10, 0, 2025, "", "desc", "")
	if err != nil {
		t.Fatal(err)
	}
	if total != 4 || len(lista) != 4 {
		t.Fatalf("esperado 4 votacoes distintas, obtido total=%d len=%d", total, len(lista))
	}
	if lista[0].CodigoVotacao != 201 {
		t.Errorf("ordem desc: primeira deveria ser a 201 (dia 20), foi %d", lista[0].CodigoVotacao)
	}

	daSessao, total, _ := repo.FindAll(10, 0, 0, "", "asc", "473484")
	if total != 3 || daSessao[0].CodigoVotacao != 101 {
		t.Errorf("filtro por sessao: total=%d primeira=%v", total, daSessao)
	}

	votos, err := repo.FindVotosByCodigoVotacao(101)
	if err != nil || len(votos) != 2 {
		t.Fatalf("votos da votacao 101: %v %v", len(votos), err)
	}
	if votos[1].SenadorNome != "Marcelo Castro" || votos[1].SiglaVoto != "Abstenção" {
		t.Errorf("join com senadores ou upsert errado: %+v", votos[1])
	}

	if v, err := repo.FindByID(103); err != nil || v.SessaoID != "473484" {
		t.Errorf("FindByID(103): %+v %v", v, err)
	}
}

func TestSenadoresEmExercicioSemVotos(t *testing.T) {
	repo := novoRepo(t)
	if err := repo.UpsertBatch([]Votacao{voto(1, 101, "1", 19, "Sim"), voto(2, 101, "1", 19, "Sim")}); err != nil {
		t.Fatal(err)
	}
	nomes, err := repo.SenadoresEmExercicioSemVotos(time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(nomes) != 1 || nomes[0] != "Sem Votos" {
		t.Errorf("esperado [Sem Votos], obtido %v", nomes)
	}
}
