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

	lista, total, err := repo.FindAll(10, 0, FiltroLista{Ano: 2025, Ordem: "desc"})
	if err != nil {
		t.Fatal(err)
	}
	if total != 4 || len(lista) != 4 {
		t.Fatalf("esperado 4 votacoes distintas, obtido total=%d len=%d", total, len(lista))
	}
	if lista[0].CodigoVotacao != 201 {
		t.Errorf("ordem desc: primeira deveria ser a 201 (dia 20), foi %d", lista[0].CodigoVotacao)
	}

	daSessao, total, _ := repo.FindAll(10, 0, FiltroLista{Ordem: "asc", Sessao: "473484"})
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

func boolPtr(b bool) *bool { return &b }

// votacaoCom monta os votos de uma votacao com os campos de filtro
func votacaoCom(codigo, dia int, materia, sigla, resultado string, secreta *bool, siglasVoto ...string) []Votacao {
	var out []Votacao
	for i, sv := range siglasVoto {
		v := voto(i+1, codigo, "1", dia, sv)
		v.Materia, v.SiglaMateria, v.Resultado, v.Secreta = materia, sigla, resultado, secreta
		out = append(out, v)
	}
	return out
}

func TestFindAllFiltrosEFacetas(t *testing.T) {
	repo := novoRepo(t)
	var lote []Votacao
	lote = append(lote, votacaoCom(1, 10, "MSF 81/2024", "MSF", "A", boolPtr(true), "Votou", "Votou")...)
	lote = append(lote, votacaoCom(2, 11, "MSF 90/2024", "MSF", "R", boolPtr(true), "Votou", "AP")...)
	lote = append(lote, votacaoCom(3, 12, "PEC 10/2024", "PEC", "A", boolPtr(false), "Sim", "Não")...)
	lote = append(lote, votacaoCom(4, 13, "PLP 124/2022 (Substitutivo-CD)", "PLP", "A", boolPtr(false), "Sim")...)
	if err := repo.UpsertBatch(lote); err != nil {
		t.Fatal(err)
	}

	casos := []struct {
		nome     string
		filtro   FiltroLista
		esperado []int // codigos em ordem desc
	}{
		{"sem filtro", FiltroLista{}, []int{4, 3, 2, 1}},
		{"um tipo", FiltroLista{Tipos: []string{"MSF"}}, []int{2, 1}},
		{"dois tipos", FiltroLista{Tipos: []string{"PEC", "PLP"}}, []int{4, 3}},
		{"secretas", FiltroLista{Secreta: boolPtr(true)}, []int{2, 1}},
		{"abertas", FiltroLista{Secreta: boolPtr(false)}, []int{4, 3}},
		{"rejeitadas", FiltroLista{Resultados: []string{"R"}}, []int{2}},
		{"combinados", FiltroLista{Tipos: []string{"MSF", "PEC"}, Resultados: []string{"A"}}, []int{3, 1}},
		{"ordem asc", FiltroLista{Tipos: []string{"MSF"}, Ordem: "asc"}, []int{1, 2}},
		{"ano sem dados", FiltroLista{Ano: 2020}, nil},
		{"tipo inexistente", FiltroLista{Tipos: []string{"XYZ"}}, nil},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			lista, total, err := repo.FindAll(10, 0, c.filtro)
			if err != nil {
				t.Fatal(err)
			}
			if int(total) != len(c.esperado) || len(lista) != len(c.esperado) {
				t.Fatalf("total=%d len=%d; esperado %d", total, len(lista), len(c.esperado))
			}
			for i, v := range lista {
				if v.CodigoVotacao != c.esperado[i] {
					t.Errorf("posicao %d: codigo %d; esperado %d", i, v.CodigoVotacao, c.esperado[i])
				}
			}
		})
	}

	fac, err := repo.Facetas(2025)
	if err != nil {
		t.Fatal(err)
	}
	conta := func(fs []Faceta) map[string]int {
		m := map[string]int{}
		for _, f := range fs {
			m[f.Valor] = f.Total
		}
		return m
	}
	// contagem por votacao, nao por voto: a MSF 81 tem 2 votos e conta 1
	if m := conta(fac.Tipos); m["MSF"] != 2 || m["PEC"] != 1 || m["PLP"] != 1 || len(m) != 3 {
		t.Errorf("facetas por tipo: %v", fac.Tipos)
	}
	if fac.Tipos[0].Valor != "MSF" {
		t.Errorf("facetas por tipo devem vir da maior para a menor: %v", fac.Tipos)
	}
	if m := conta(fac.Secreta); m["true"] != 2 || m["false"] != 2 {
		t.Errorf("facetas por secreta: %v", fac.Secreta)
	}
	if m := conta(fac.Resultados); m["A"] != 3 || m["R"] != 1 {
		t.Errorf("facetas por resultado: %v", fac.Resultados)
	}
	if fac.Total != 4 {
		t.Errorf("total das facetas: %d", fac.Total)
	}

	if fac, err := repo.Facetas(2020); err != nil || len(fac.Tipos) != 0 || fac.Total != 0 {
		t.Errorf("facetas de ano sem dados: %+v %v", fac, err)
	}
}

func TestPreencherHistoricoIdempotente(t *testing.T) {
	repo := novoRepo(t)
	var lote []Votacao
	// linhas antigas: sem sigla_materia e com secreta nulo
	lote = append(lote, votacaoCom(1, 10, "MSF 81/2024", "", "A", nil, "Votou", "AP")...)
	lote = append(lote, votacaoCom(2, 11, "PLP 124/2022 (Substitutivo-CD)", "", "A", nil, "Sim", "Não")...)
	lote = append(lote, votacaoCom(3, 12, "", "", "A", nil, "Sim")...)
	// linha ja gravada pelo sync novo: nao pode ser sobrescrita
	lote = append(lote, votacaoCom(4, 13, "pec 5/2024", "PEC", "A", boolPtr(true), "Sim")...)
	if err := repo.UpsertBatch(lote); err != nil {
		t.Fatal(err)
	}

	esperado := map[int]struct {
		sigla   string
		secreta bool
	}{1: {"MSF", true}, 2: {"PLP", false}, 3: {"", false}, 4: {"PEC", true}}
	for rodada := 1; rodada <= 2; rodada++ {
		if err := PreencherHistorico(repo.db); err != nil {
			t.Fatalf("rodada %d: %v", rodada, err)
		}
		var linhas []Votacao
		if err := repo.db.Order("codigo_votacao, senador_id").Find(&linhas).Error; err != nil {
			t.Fatal(err)
		}
		for _, l := range linhas {
			e := esperado[l.CodigoVotacao]
			if l.SiglaMateria != e.sigla || l.Secreta == nil || *l.Secreta != e.secreta {
				t.Errorf("rodada %d, votacao %d senador %d: sigla=%q secreta=%v; esperado %+v",
					rodada, l.CodigoVotacao, l.SenadorID, l.SiglaMateria, l.Secreta, e)
			}
		}
	}
}
