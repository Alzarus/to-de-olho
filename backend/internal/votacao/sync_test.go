package votacao

import (
	"testing"
	"time"

	senadoapi "github.com/Alzarus/to-de-olho/pkg/senado"
)

func ptr(v int) *int { return &v }

func TestConverterVotacao(t *testing.T) {
	v := senadoapi.VotacaoSessaoAPI{
		Ano:                 2022, // ano da MATERIA, nao da sessao: nao pode entrar na chave
		CodigoSessao:        581816,
		CodigoSessaoVotacao: 7101,
		SequencialVotacao:   nil,
		DataSessao:          "2026-08-12",
		DescricaoVotacao:    "Votação nominal do Substitutivo",
		Identificacao:       "PLP 124/2022 (Substitutivo-CD)",
		Ementa:              "Altera o CTN",
		ResultadoVotacao:    "A",
		Sigla:               "PLP",
		VotacaoSecreta:      "N",
		Votos: []senadoapi.VotoParlamentar{
			{CodigoParlamentar: 5672, SiglaVoto: "Sim"},
			{CodigoParlamentar: 742, SiglaVoto: "Não"},
			{CodigoParlamentar: 5894, SiglaVoto: "P-OD"},
			{CodigoParlamentar: 9999, SiglaVoto: "AP"}, // fora da tabela senadores
		},
	}
	votos, ignorados, err := converterVotacao(v, map[int]int{5672: 1, 742: 2, 5894: 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(votos) != 3 || ignorados != 1 {
		t.Fatalf("esperado 3 votos e 1 ignorado, obtido %d e %d", len(votos), ignorados)
	}
	nao := votos[1]
	if nao.CodigoVotacao != 7101 || nao.SessaoID != "581816" || nao.SiglaVoto != "Não" || nao.Voto != "Nao" {
		t.Errorf("voto convertido errado: %+v", nao)
	}
	if nao.Materia != "PLP 124/2022 (Substitutivo-CD)" || nao.Ementa != "Altera o CTN" || nao.Resultado != "A" {
		t.Errorf("metadados errados: %+v", nao)
	}
	if nao.SiglaMateria != "PLP" || nao.Secreta == nil || *nao.Secreta {
		t.Errorf("sigla/secreta errados: %q %v", nao.SiglaMateria, nao.Secreta)
	}
	if !nao.Data.Equal(time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC)) {
		t.Errorf("data errada: %v", nao.Data)
	}
	if votos[2].Voto != "P-OD" {
		t.Errorf("P-OD nao pode virar Obstrucao (item 15): %q", votos[2].Voto)
	}
}

func TestConverterVotacaoSemCodigoOuDataFalha(t *testing.T) {
	if _, _, err := converterVotacao(senadoapi.VotacaoSessaoAPI{DataSessao: "2026-08-12"}, nil); err == nil {
		t.Error("votacao sem codigoSessaoVotacao deveria falhar")
	}
	if _, _, err := converterVotacao(senadoapi.VotacaoSessaoAPI{CodigoSessaoVotacao: 1, DataSessao: "12/08/2026"}, nil); err == nil {
		t.Error("data fora do formato deveria falhar")
	}
}

func TestJanelasMensais(t *testing.T) {
	d := func(s string) time.Time { t, _ := time.Parse("2006-01-02", s); return t }
	janelas := janelasMensais(d("2023-02-01"), d("2023-04-10"))
	esperado := [][2]string{{"2023-02-01", "2023-02-28"}, {"2023-03-01", "2023-03-31"}, {"2023-04-01", "2023-04-10"}}
	if len(janelas) != len(esperado) {
		t.Fatalf("esperado %d janelas, obtido %d", len(esperado), len(janelas))
	}
	for i, j := range janelas {
		if j[0].Format("2006-01-02") != esperado[i][0] || j[1].Format("2006-01-02") != esperado[i][1] {
			t.Errorf("janela %d = %s..%s; esperado %v", i, j[0].Format("2006-01-02"), j[1].Format("2006-01-02"), esperado[i])
		}
	}
	if n := len(janelasMensais(d("2026-09-23"), d("2026-09-23"))); n != 1 {
		t.Errorf("um dia so deveria dar 1 janela, deu %d", n)
	}
}

func TestConverterVotacaoSemSiglaDerivaDaIdentificacao(t *testing.T) {
	v := senadoapi.VotacaoSessaoAPI{
		CodigoSessaoVotacao: 6966, DataSessao: "2025-08-13", Identificacao: "MSF 81/2024", VotacaoSecreta: "S",
		Votos: []senadoapi.VotoParlamentar{{CodigoParlamentar: 1, SiglaVoto: "Votou"}},
	}
	votos, _, err := converterVotacao(v, map[int]int{1: 1})
	if err != nil || len(votos) != 1 {
		t.Fatalf("%v %v", votos, err)
	}
	if votos[0].SiglaMateria != "MSF" || votos[0].Secreta == nil || !*votos[0].Secreta {
		t.Errorf("sigla/secreta errados: %q %v", votos[0].SiglaMateria, votos[0].Secreta)
	}
}

func TestSiglaDaIdentificacao(t *testing.T) {
	casos := []struct{ entrada, esperado string }{
		{"PLP 124/2022 (Substitutivo-CD)", "PLP"},
		{"  msf 81/2024", "MSF"},
		{"PEC", "PEC"},
		{"", ""},
		{"   ", ""},
	}
	for _, c := range casos {
		if obtido := siglaDaIdentificacao(c.entrada); obtido != c.esperado {
			t.Errorf("siglaDaIdentificacao(%q) = %q; esperado %q", c.entrada, obtido, c.esperado)
		}
	}
}

func TestVotacaoSecreta(t *testing.T) {
	casos := []struct {
		entrada  string
		esperado *bool
	}{
		{"S", boolPtr(true)},
		{"s", boolPtr(true)},
		{"N", boolPtr(false)},
		{"", nil},
		{"?", nil},
	}
	for _, c := range casos {
		obtido := votacaoSecreta(c.entrada)
		if (obtido == nil) != (c.esperado == nil) || (obtido != nil && *obtido != *c.esperado) {
			t.Errorf("votacaoSecreta(%q) = %v; esperado %v", c.entrada, obtido, c.esperado)
		}
	}
}
