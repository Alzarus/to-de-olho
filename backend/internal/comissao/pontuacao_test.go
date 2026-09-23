package comissao

import (
	"testing"
	"time"

	"github.com/Alzarus/to-de-olho/internal/testdb"
)

func d(s string) *time.Time { t, _ := time.Parse("2006-01-02", s); return &t }

func TestCategorizar(t *testing.T) {
	casos := map[string]Categoria{
		"Comissão de Assuntos Econômicos":                    Colegiado,
		"CPMI - INSS":                                        Colegiado,
		"Frente Parlamentar da Eletromobilidade":             Frente,
		"Frente Parlamentar Mista das Ferrovias Autorizadas": Frente,
		"Grupo Parlamentar Brasil-China":                     Grupo,
		"Grupo Brasileiro do Parlatino":                      Grupo,
		"Conselho do Diploma José Ermírio de Moraes":         Honraria,
		"Comenda Santa Dulce dos Pobres":                     Honraria,
		"Conselho do Prêmio Adoção Tardia":                   Honraria,
		"Subcomissão Temporária sobre Doenças Raras":         Colegiado,
	}
	for nome, esperado := range casos {
		if got := Categorizar(nome); got != esperado {
			t.Errorf("Categorizar(%q) = %s; esperado %s", nome, got, esperado)
		}
	}
}

func TestCalcularStatsUmaVezPorColegiadoMesmaFormulaSempre(t *testing.T) {
	fim := time.Date(2026, 9, 23, 0, 0, 0, 0, time.UTC)
	participacoes := []ComissaoMembro{
		// CAE: suplente e depois titular (reconducao) -> conta uma vez, como titular
		{CodigoComissao: "38", NomeComissao: "Comissão de Assuntos Econômicos", DescricaoParticipacao: "Suplente", DataInicio: d("2023-03-01"), DataFim: d("2024-12-31")},
		{CodigoComissao: "38", NomeComissao: "Comissão de Assuntos Econômicos", DescricaoParticipacao: "Titular", DataInicio: d("2025-02-18")},
		// CAS: so suplente, ja encerrada
		{CodigoComissao: "40", NomeComissao: "Comissão de Assuntos Sociais", DescricaoParticipacao: "Suplente", DataInicio: d("2023-03-01"), DataFim: d("2025-02-17")},
		// fora da conta
		{CodigoComissao: "900", NomeComissao: "Frente Parlamentar Católica", DescricaoParticipacao: "Titular", DataInicio: d("2023-03-01")},
		{CodigoComissao: "901", NomeComissao: "Grupo Parlamentar Brasil-Japão", DescricaoParticipacao: "Titular", DataInicio: d("2023-03-01")},
	}
	st := calcularStats(1, participacoes, fim)
	if st.TotalComissoes != 2 || st.ComissoesTitular != 1 || st.ComissoesSuplente != 1 || st.Pontos != 3 || st.ForaDaConta != 2 || st.ComissoesAtivas != 1 {
		t.Errorf("stats erradas: %+v", st)
	}
}

func TestSubstituirDoSenadorEPeriodo(t *testing.T) {
	repo := NewRepository(testdb.Abrir(t, &ComissaoMembro{}))
	velha := []ComissaoMembro{{SenadorID: 1, CodigoComissao: "38", NomeComissao: "CAE", DescricaoParticipacao: "Titular", DataInicio: d("2019-02-01"), DataFim: d("2023-01-20")}}
	if err := repo.SubstituirDoSenador(1, velha); err != nil {
		t.Fatal(err)
	}
	nova := []ComissaoMembro{
		velha[0],
		{SenadorID: 1, CodigoComissao: "38", NomeComissao: "CAE", DescricaoParticipacao: "Suplente", DataInicio: d("2023-03-01")},
	}
	if err := repo.SubstituirDoSenador(1, nova); err != nil {
		t.Fatal(err)
	}
	var n int64
	repo.db.Model(&ComissaoMembro{}).Where("senador_id = 1").Count(&n)
	if n != 2 {
		t.Fatalf("substituicao deveria deixar 2 periodos, deixou %d", n)
	}
	// periodo que terminou em janeiro de 2023 e da legislatura anterior
	st, err := repo.GetStatsPeriodo(1, time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 9, 23, 0, 0, 0, 0, time.UTC))
	if err != nil || st.Pontos != 1 || st.ComissoesSuplente != 1 {
		t.Errorf("no recorte so a suplencia conta: %+v %v", st, err)
	}
}
