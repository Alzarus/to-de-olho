package proposicao

import (
	"testing"
	"time"

	"github.com/Alzarus/to-de-olho/internal/testdb"
)

func TestUpsertPreservaCoautoriaEAtualizaCampos(t *testing.T) {
	repo := NewRepository(testdb.Abrir(t, &Proposicao{}))
	data := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)

	// a mesma materia para dois senadores: antes do item 1 o segundo sumia
	lote := []Proposicao{
		{SenadorID: 1, CodigoMateria: "155808", SiglaSubtipoMateria: "PEC", EstagioTramitacao: "Apresentado", PosicaoAutoria: ptr(1), DataApresentacao: &data},
		{SenadorID: 2, CodigoMateria: "155808", SiglaSubtipoMateria: "PEC", EstagioTramitacao: "Apresentado", PosicaoAutoria: ptr(2), DataApresentacao: &data},
	}
	for i := range lote {
		lote[i].Pontuacao = lote[i].CalcularPontuacao()
	}
	if err := repo.UpsertBatch(lote); err != nil {
		t.Fatal(err)
	}

	// correcao na API: nova ementa e novo estagio chegam ao banco
	corrigida := lote[0]
	corrigida.ID = 0
	corrigida.Ementa = "ementa corrigida"
	corrigida.EstagioTramitacao = "TransformadoLei"
	corrigida.Pontuacao = corrigida.CalcularPontuacao()
	if err := repo.Upsert(&corrigida); err != nil {
		t.Fatal(err)
	}

	var linhas []Proposicao
	repo.db.Order("senador_id").Find(&linhas)
	if len(linhas) != 2 {
		t.Fatalf("esperado 2 linhas (autor + coautor), obtido %d", len(linhas))
	}
	if linhas[0].Ementa != "ementa corrigida" || linhas[0].Pontuacao != 48 {
		t.Errorf("upsert nao atualizou: ementa=%q pontuacao=%v", linhas[0].Ementa, linhas[0].Pontuacao)
	}
}

func TestStatsSoAutoriaPrincipalPontuaEConta(t *testing.T) {
	t.Setenv("RECORTE_INICIO", "2023-02-01")
	repo := NewRepository(testdb.Abrir(t, &Proposicao{}))
	dentro := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	fora := time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC) // legislatura anterior

	lote := []Proposicao{
		{SenadorID: 1, CodigoMateria: "1", SiglaSubtipoMateria: "PEC", EstagioTramitacao: "Apresentado", PosicaoAutoria: ptr(1), DataApresentacao: &dentro, AnoMateria: 2024},
		{SenadorID: 1, CodigoMateria: "2", SiglaSubtipoMateria: "PL", EstagioTramitacao: "TransformadoLei", PosicaoAutoria: ptr(1), DataApresentacao: &dentro, AnoMateria: 2024},
		{SenadorID: 1, CodigoMateria: "3", SiglaSubtipoMateria: "PEC", EstagioTramitacao: "AprovadoPlenario", PosicaoAutoria: ptr(5), DataApresentacao: &dentro, AnoMateria: 2024},
		{SenadorID: 1, CodigoMateria: "4", SiglaSubtipoMateria: "PL", EstagioTramitacao: "Apresentado", PosicaoAutoria: ptr(1), DataApresentacao: &fora, AnoMateria: 2023},
		{SenadorID: 2, CodigoMateria: "1", SiglaSubtipoMateria: "PEC", EstagioTramitacao: "Apresentado", PosicaoAutoria: ptr(2), DataApresentacao: &dentro, AnoMateria: 2024},
	}
	for i := range lote {
		lote[i].Pontuacao = lote[i].CalcularPontuacao()
	}
	if err := repo.UpsertBatch(lote); err != nil {
		t.Fatal(err)
	}

	st, err := repo.GetStats(1)
	if err != nil {
		t.Fatal(err)
	}
	// principal no recorte: PEC apresentada (3) + PL virou lei (16) = 19
	if st.TotalProposicoes != 2 || st.TotalCoautorias != 1 || st.TotalPECs != 1 || st.TotalPLs != 1 ||
		st.TransformadasEmLei != 1 || st.AprovadosPlenario != 1 || st.EmTramitacao != 1 || st.PontuacaoTotal != 19 {
		t.Errorf("stats do mandato erradas: %+v", st)
	}

	st2, _ := repo.GetStats(2)
	if st2.TotalProposicoes != 0 || st2.TotalCoautorias != 1 || st2.PontuacaoTotal != 0 {
		t.Errorf("coautor de PEC nao pode pontuar: %+v", st2)
	}

	ano, _ := repo.GetStatsByAno(1, 2023)
	if ano.TotalProposicoes != 1 || ano.PontuacaoTotal != 1 {
		t.Errorf("stats por ano erradas: %+v", ano)
	}
}
