package senado

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestComposicaoMesaSF(t *testing.T) {
	casos := []struct {
		nome     string
		corpo    string
		esperado []CargoMesaAPI
	}{
		{
			nome: "lista de cargos (formato atual, trecho real de 23/09/2026)",
			corpo: `{"MesaSenado":{"Colegiados":{"Colegiado":[{"SiglaColegiado":"CDIR","Cargos":{"Cargo":[
				{"Cargo":["PRESIDENTE"],"NomeParlamentar":"Senador Davi Alcolumbre","Http":"3830"},
				{"Cargo":["1ª SECRETÁRIA"],"NomeParlamentar":"Senadora Daniella Ribeiro","Http":"5998"}]}}]}}}`,
			esperado: []CargoMesaAPI{{3830, "PRESIDENTE"}, {5998, "1ª SECRETÁRIA"}},
		},
		{
			nome:     "objeto unico no lugar de lista",
			corpo:    `{"MesaSenado":{"Colegiados":{"Colegiado":{"Cargos":{"Cargo":{"Cargo":"PRESIDENTE","Http":"3830"}}}}}}`,
			esperado: []CargoMesaAPI{{3830, "PRESIDENTE"}},
		},
		{
			nome:  "codigo invalido e ignorado",
			corpo: `{"MesaSenado":{"Colegiados":{"Colegiado":[{"Cargos":{"Cargo":[{"Cargo":["PRESIDENTE"],"Http":""}]}}]}}}`,
		},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(c.corpo))
			}))
			defer srv.Close()
			cli := &LegisClient{baseURL: srv.URL, httpClient: srv.Client()}
			got, err := cli.ComposicaoMesaSF(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(c.esperado) {
				t.Fatalf("esperado %+v, veio %+v", c.esperado, got)
			}
			for i := range got {
				if got[i] != c.esperado[i] {
					t.Errorf("item %d: esperado %+v, veio %+v", i, c.esperado[i], got[i])
				}
			}
		})
	}
}
