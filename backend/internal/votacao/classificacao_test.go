package votacao

import (
	"math"
	"testing"
	"time"

	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/internal/testdb"
)

func perto(a, b float64) bool { return math.Abs(a-b) < 0.05 }

// Numeros do PLANO-MIGRACAO.md §7, recalculados sobre as 423 votacoes do mandato
func TestCalcularStatsCasosDoPlano(t *testing.T) {
	casos := []struct {
		nome     string
		porSigla map[string]int
		a, b     float64
	}{
		// 423 registros, 402 presentes, 0 AP, 14 justificadas, 7 NA
		{"Marcelo Castro", map[string]int{"Sim": 300, "Não": 60, "Votou": 30, "P-NRV": 12, "LS": 10, "MIS": 4, "NA": 7}, 96.6, 100.0},
		// 233 presentes, 166 AP, 17 justificadas, 7 NCom
		{"Giordano", map[string]int{"Sim": 200, "Votou": 33, "AP": 166, "LS": 17, "NCom": 7}, 55.1, 57.4},
	}
	for _, c := range casos {
		st := calcularStats(1, c.porSigla)
		if !perto(st.PresencaBruta, c.a) || !perto(st.PresencaAjustada, c.b) || st.TaxaPresenca != st.PresencaAjustada {
			t.Errorf("%s: A=%.2f B=%.2f taxa=%.2f; esperado A=%.1f B=%.1f", c.nome, st.PresencaBruta, st.PresencaAjustada, st.TaxaPresenca, c.a, c.b)
		}
		if st.TotalVotacoes != 423 || !st.DadosSuficientes {
			t.Errorf("%s: total=%d dados=%v", c.nome, st.TotalVotacoes, st.DadosSuficientes)
		}
	}

	g := calcularStats(1, casos[1].porSigla)
	if g.AusenciasAP != 166 || g.NaoCompareceu != 7 || g.Ausencias != 173 || g.AusenciasJustificadas != 17 || g.Presentes != 233 || g.VotosRegistrados != 200 {
		t.Errorf("contagens do Giordano erradas: %+v", g)
	}
}

// Metodologia v2.1: LP conta como falta; obstrucao fica fora da conta
func TestLicencaParticularEObstrucao(t *testing.T) {
	st := calcularStats(1, map[string]int{"Sim": 8, "LP": 2, "LS": 5, "Obstrução": 3, "P-OD": 1})
	if !perto(st.PresencaAjustada, 80) || !perto(st.PresencaBruta, 53.33) {
		t.Errorf("LP deveria ser falta e obstrucao fora da conta: A=%.2f B=%.2f", st.PresencaBruta, st.PresencaAjustada)
	}
	if st.Ausencias != 2 || st.AusenciasJustificadas != 5 || st.Obstrucoes != 4 || st.Presentes != 8 {
		t.Errorf("contagens erradas: %+v", st)
	}
}

func TestCalcularStatsSemDados(t *testing.T) {
	// so licencas: sem nenhum registro que conte em B. Antes virava presenca 0
	st := calcularStats(1, map[string]int{"LS": 5, "NA": 1})
	if st.DadosSuficientes || st.TaxaPresenca != 0 {
		t.Errorf("sem denominador deveria marcar dados insuficientes: %+v", st)
	}
	vazio := calcularStats(1, map[string]int{})
	if vazio.DadosSuficientes {
		t.Error("sem registros deveria marcar dados insuficientes")
	}
}

func TestClassificarCodigoDesconhecidoNaoConta(t *testing.T) {
	if Classificar("XYZ") != NaoConta {
		t.Error("codigo desconhecido deveria ficar fora do calculo")
	}
	st := calcularStats(1, map[string]int{"Sim": 9, "XYZ": 1, "AP": 1})
	if !perto(st.PresencaBruta, 90) {
		t.Errorf("desconhecido nao pode entrar no denominador: A=%.2f", st.PresencaBruta)
	}
}

func TestGetStatsUsaRecorteEAno(t *testing.T) {
	t.Setenv("RECORTE_INICIO", "2023-02-01")
	db := testdb.Abrir(t, &senador.Senador{}, &Votacao{})
	repo := NewRepository(db)
	em := func(codigo int, data string, sigla string) Votacao {
		d, _ := time.Parse("2006-01-02", data)
		return Votacao{SenadorID: 1, CodigoVotacao: codigo, SessaoID: "1", Data: d.Add(12 * time.Hour), SiglaVoto: sigla, Voto: rotuloVoto(sigla)}
	}
	if err := repo.UpsertBatch([]Votacao{
		em(1, "2022-12-10", "AP"), // legislatura anterior: fora
		em(2, "2023-03-01", "Sim"),
		em(3, "2024-05-01", "AP"),
		em(4, "2024-05-02", "LS"),
		em(5, "2024-05-03", "Votou"),
	}); err != nil {
		t.Fatal(err)
	}

	st, err := repo.GetStats(1)
	if err != nil {
		t.Fatal(err)
	}
	if st.TotalVotacoes != 4 || !perto(st.PresencaAjustada, 66.67) || !perto(st.PresencaBruta, 50) {
		t.Errorf("mandato: %+v", st)
	}
	ano, _ := repo.GetStatsByAno(1, 2024)
	if ano.TotalVotacoes != 3 || !perto(ano.PresencaAjustada, 50) {
		t.Errorf("2024: %+v", ano)
	}
}
