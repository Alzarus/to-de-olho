package senado

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// servidorFixture responde recursos-utilizados com a resposta real gravada em
// testdata (Randolfe Rodrigues, 5012, ano 2026, gravada em 23/09/2026).
func servidorFixture(t *testing.T, status int, corpo string, chamadas *atomic.Int32) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if chamadas != nil {
			chamadas.Add(1)
		}
		if !strings.HasSuffix(r.URL.Path, "/api/v1/senadores/5012/recursos-utilizados") || r.URL.Query().Get("ano") != "2026" {
			http.Error(w, "rota inesperada "+r.URL.String(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(corpo))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestRecursosUtilizados_ContratoFixture(t *testing.T) {
	fixture, err := os.ReadFile("testdata/recursos_utilizados_5012_2026.json")
	if err != nil {
		t.Fatal(err)
	}
	srv := servidorFixture(t, http.StatusOK, string(fixture), nil)
	r, err := NewAdmClientURL(srv.URL, 1).RecursosUtilizados(context.Background(), 5012, 2026)
	if err != nil {
		t.Fatal(err)
	}
	if r == nil {
		t.Fatal("esperava dados, veio nil")
	}
	if r.Ano != 2026 || r.Parlamentar.Nome != "Randolfe Rodrigues" {
		t.Errorf("cabecalho errado: %+v", r.Parlamentar)
	}

	// valores conferidos com /servidores/servidores (validacao cruzada)
	esperado := map[string]map[string]int{
		"Gabinete":               {"Comissionado": 21, "Efetivo": 3},
		"Escritório(s) de Apoio": {"Comissionado": 48},
	}
	if len(r.Pessoal) != len(esperado) {
		t.Fatalf("esperava %d locais, veio %+v", len(esperado), r.Pessoal)
	}
	for _, p := range r.Pessoal {
		vinc, ok := esperado[p.Local]
		if !ok {
			t.Errorf("local inesperado %q", p.Local)
			continue
		}
		soma := 0
		for _, v := range p.Vinculos {
			if vinc[v.Vinculo] != v.Quantidade {
				t.Errorf("%s/%s: esperado %d, veio %d", p.Local, v.Vinculo, vinc[v.Vinculo], v.Quantidade)
			}
			soma += v.Quantidade
		}
		if soma != p.Quantidade {
			t.Errorf("%s: soma dos vinculos %d difere do total %d", p.Local, soma, p.Quantidade)
		}
	}

	if len(r.Beneficios) != 2 || r.Beneficios[0].Beneficio == "" || r.Beneficios[0].Utilizacao == "" {
		t.Errorf("beneficios fora do contrato: %+v", r.Beneficios)
	}
	if r.Cotas.TotalValor <= 0 || len(r.GastosNaoInclusos.Despesas) == 0 {
		t.Errorf("cota ou gastos nao inclusos vazios: %+v %+v", r.Cotas, r.GastosNaoInclusos)
	}
}

func TestRecursosUtilizados_Respostas(t *testing.T) {
	casos := []struct {
		nome         string
		status       int
		corpo        string
		tentativas   int
		chamadas     int32
		erroEsperado error
		qualquerErro bool
		esperaNilNil bool
	}{
		{
			nome:         "404 nao repete e devolve ErrSenadorNaoEncontrado",
			status:       http.StatusNotFound,
			corpo:        `{"statusCode":404,"msg":"Senador não encontrado para o código: 5012","data":[]}`,
			tentativas:   3,
			chamadas:     1,
			erroEsperado: ErrSenadorNaoEncontrado,
		},
		{
			nome:         "data vazio devolve nil sem erro",
			status:       http.StatusOK,
			corpo:        `{"statusCode":200,"msg":"ok","data":[]}`,
			tentativas:   3,
			chamadas:     1,
			esperaNilNil: true,
		},
		{
			nome:         "500 repete ate esgotar as tentativas",
			status:       http.StatusInternalServerError,
			corpo:        `{}`,
			tentativas:   2,
			chamadas:     2,
			qualquerErro: true,
		},
		{
			nome:         "JSON cortado vira erro",
			status:       http.StatusOK,
			corpo:        `{"statusCode":200,"data":[{"ano":20`,
			tentativas:   1,
			chamadas:     1,
			qualquerErro: true,
		},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			var chamadas atomic.Int32
			srv := servidorFixture(t, c.status, c.corpo, &chamadas)
			r, err := NewAdmClientURL(srv.URL, c.tentativas).RecursosUtilizados(context.Background(), 5012, 2026)
			switch {
			case c.erroEsperado != nil:
				if !errors.Is(err, c.erroEsperado) {
					t.Errorf("esperava %v, veio %v", c.erroEsperado, err)
				}
			case c.qualquerErro:
				if err == nil {
					t.Error("esperava erro")
				}
			case c.esperaNilNil:
				if err != nil || r != nil {
					t.Errorf("esperava (nil, nil), veio (%+v, %v)", r, err)
				}
			}
			if got := chamadas.Load(); got != c.chamadas {
				t.Errorf("esperava %d chamadas, houve %d", c.chamadas, got)
			}
		})
	}
}

// TestContract_RecursosUtilizados consulta a API real (fora do CI).
func TestContract_RecursosUtilizados(t *testing.T) {
	if os.Getenv("CI") != "" || testing.Short() {
		t.Skip("teste de contrato com a API real")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	r, err := NewAdmClient().RecursosUtilizados(ctx, 5012, 2024)
	if err != nil {
		t.Fatalf("contrato quebrado ou API indisponivel: %v", err)
	}
	if r == nil || len(r.Pessoal) == 0 {
		t.Fatalf("API sem pessoal para 5012/2024: %+v", r)
	}
	for _, p := range r.Pessoal {
		if p.Local == "" || len(p.Vinculos) == 0 {
			t.Errorf("local sem nome ou sem vinculos: %+v", p)
		}
	}
}
