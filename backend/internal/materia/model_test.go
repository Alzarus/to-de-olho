package materia

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	senadoapi "github.com/Alzarus/to-de-olho/pkg/senado"
)

func TestNormalizarApelido(t *testing.T) {
	casos := []struct {
		nome, entrada string
		esperado      *string
	}{
		{"vazio", "", nil},
		{"so espacos", "  \n ", nil},
		{"simples", "Lei das Bets", ptr("Lei das Bets")},
		{"espaco no fim", "PL Antifacção ", ptr("PL Antifacção")},
		{"duas linhas", "Lei da Reciprocidade\nLei da Reciprocidade Econômica", ptr("Lei da Reciprocidade")},
		{"crlf", "Propag\r\nOutro", ptr("Propag")},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			got := NormalizarApelido(c.entrada)
			if !reflect.DeepEqual(got, c.esperado) {
				t.Errorf("NormalizarApelido(%q) = %v; esperado %v", c.entrada, val(got), val(c.esperado))
			}
		})
	}
}

// Fixtures: respostas reais de /processo/{id} (23/09/2026), sem os blocos
// grandes que o sync nao le (despachos, autuacoes, ordensDoDia...).
func TestDoProcessoFixtureReal(t *testing.T) {
	casos := []struct {
		arquivo       string
		codigo, id    int
		identificacao string
		sigla         string
		apelido       *string
		explicacao    bool
		temas         []string
	}{
		{"processo_8441243.json", 157233, 8441243, "PL 2338/2023", "PL", ptr("Marco Legal da Inteligência Artificial"), false,
			[]string{"Ciência, Tecnologia e Informática", "Responsabilidade Civil", "Direitos Individuais e Coletivos"}},
		{"processo_8701679.json", 164914, 8701679, "PLP 68/2024", "PLP", nil, false,
			[]string{"Tributos", "Administração Tributária"}},
		{"processo_7804756.json", 0, 7804756, "PEC 137/2019", "PEC", nil, true, []string{"Educação"}},
		// apelido com duas linhas na API: fica a primeira
		{"processo_8421084.json", 0, 8421084, "PL 2088/2023", "PL", ptr("Lei da Reciprocidade"), false, nil},
	}
	for _, c := range casos {
		t.Run(c.arquivo, func(t *testing.T) {
			b, err := os.ReadFile(filepath.Join("testdata", c.arquivo))
			if err != nil {
				t.Fatal(err)
			}
			var det senadoapi.ProcessoDetalhe
			if err := json.Unmarshal(b, &det); err != nil {
				t.Fatal(err)
			}
			m := DoProcesso(&det)
			if c.codigo != 0 && m.CodigoMateria != c.codigo {
				t.Errorf("codigo_materia %d; esperado %d", m.CodigoMateria, c.codigo)
			}
			if m.CodigoMateria == 0 {
				t.Error("codigo_materia vazio")
			}
			if m.IdProcesso != c.id || m.Identificacao != c.identificacao || m.Sigla != c.sigla {
				t.Errorf("id/identificacao/sigla = %d %q %q", m.IdProcesso, m.Identificacao, m.Sigla)
			}
			if !reflect.DeepEqual(m.Apelido, c.apelido) {
				t.Errorf("apelido %v; esperado %v", val(m.Apelido), val(c.apelido))
			}
			if (m.ExplicacaoEmenta != nil) != c.explicacao {
				t.Errorf("explicacao_ementa %v; esperado presente=%v", val(m.ExplicacaoEmenta), c.explicacao)
			}
			if m.Ementa == "" {
				t.Error("ementa vazia")
			}
			if !strings.HasPrefix(m.UrlDocumento, "https://legis.senado.gov.br/sdleg-getter/documento?dm=") {
				t.Errorf("url_documento inesperada: %q", m.UrlDocumento)
			}
			if c.temas != nil && !reflect.DeepEqual([]string(m.Temas), c.temas) {
				t.Errorf("temas %v; esperado %v", m.Temas, c.temas)
			}
			if len(m.Temas) == 0 {
				t.Error("sem temas")
			}
		})
	}
}

func TestTemasValueScan(t *testing.T) {
	casos := []struct {
		nome  string
		temas Temas
	}{
		{"vazio vira nulo", nil},
		{"com acentos", Temas{"Tributos", "Educação"}},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			v, err := c.temas.Value()
			if err != nil {
				t.Fatal(err)
			}
			var volta Temas
			if err := volta.Scan(v); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(volta, c.temas) {
				t.Errorf("ida e volta: %v -> %v", c.temas, volta)
			}
		})
	}
	var tb Temas
	if err := tb.Scan([]byte(`["A"]`)); err != nil || len(tb) != 1 {
		t.Errorf("scan de []byte: %v %v", tb, err)
	}
	if err := tb.Scan(42); err == nil {
		t.Error("esperado erro para tipo inesperado")
	}
}

func TestCuradoriaTemFonteOficial(t *testing.T) {
	dominios := []string{"https://www12.senado.leg.br/", "https://www25.senado.leg.br/", "https://www.camara.leg.br/", "https://www.gov.br/"}
	vistos := map[int]bool{}
	for _, a := range ApelidosCurados() {
		if vistos[a.CodigoMateria] {
			t.Errorf("codigo %d repetido na curadoria", a.CodigoMateria)
		}
		vistos[a.CodigoMateria] = true
		if a.Apelido == "" || a.Identificacao == "" || a.CodigoMateria == 0 {
			t.Errorf("entrada incompleta: %+v", a)
		}
		oficial := false
		for _, d := range dominios {
			oficial = oficial || strings.HasPrefix(a.FonteURL, d)
		}
		if !oficial {
			t.Errorf("%s: fonte fora dos dominios oficiais: %s", a.Identificacao, a.FonteURL)
		}
	}
}

func ptr(s string) *string { return &s }

func val(p *string) any {
	if p == nil {
		return nil
	}
	return *p
}
