package gabinete

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/internal/testdb"
	"github.com/Alzarus/to-de-olho/pkg/senado"
)

func pessoal(local string, vinculos ...senado.VinculoPessoalAPI) senado.PessoalLocalAPI {
	p := senado.PessoalLocalAPI{Local: local, Vinculos: vinculos}
	for _, v := range vinculos {
		p.Quantidade += v.Quantidade
	}
	return p
}

func vinc(nome string, q int) senado.VinculoPessoalAPI {
	return senado.VinculoPessoalAPI{Vinculo: nome, Quantidade: q}
}

func TestConverter(t *testing.T) {
	casos := []struct {
		nome       string
		entrada    *senado.RecursosUtilizadosAPI
		recursos   map[[2]string]int
		beneficios int
	}{
		{nome: "nil", entrada: nil, recursos: map[[2]string]int{}},
		{
			nome: "Randolfe 2026 (fixture real)",
			entrada: &senado.RecursosUtilizadosAPI{
				Pessoal: []senado.PessoalLocalAPI{
					pessoal("Gabinete", vinc("Comissionado", 21), vinc("Efetivo", 3)),
					pessoal("Escritório(s) de Apoio", vinc("Comissionado", 48)),
				},
				Beneficios: []senado.BeneficioAPI{{Beneficio: "Auxílio-Moradia", Utilizacao: "Não utilizou"}, {Beneficio: "Imóvel Funcional", Utilizacao: "Utilizou"}},
			},
			recursos: map[[2]string]int{
				{LocalGabinete, "Comissionado"}: 21, {LocalGabinete, "Efetivo"}: 3, {LocalEscritorio, "Comissionado"}: 48,
			},
			beneficios: 2,
		},
		{
			nome: "exclui o vinculo PARLAMENTAR, zeros e local desconhecido; soma repetidos",
			entrada: &senado.RecursosUtilizadosAPI{
				Pessoal: []senado.PessoalLocalAPI{
					pessoal("Gabinete", vinc("PARLAMENTAR", 1), vinc("Comissionado", 5), vinc("Comissionado", 2), vinc("Requisitado", 0)),
					pessoal("Liderança do Governo", vinc("Comissionado", 9)),
				},
				Beneficios: []senado.BeneficioAPI{{Beneficio: "Imóvel Funcional", Utilizacao: "Utilizou"}, {Beneficio: "Imóvel Funcional", Utilizacao: "Utilizou"}},
			},
			recursos:   map[[2]string]int{{LocalGabinete, "Comissionado"}: 7},
			beneficios: 1,
		},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			rec, ben := converter(c.entrada)
			if len(rec) != len(c.recursos) {
				t.Fatalf("esperava %d linhas, veio %+v", len(c.recursos), rec)
			}
			for _, r := range rec {
				if q, ok := c.recursos[[2]string{r.Local, r.Vinculo}]; !ok || q != r.Quantidade {
					t.Errorf("linha inesperada %+v", r)
				}
			}
			if len(ben) != c.beneficios {
				t.Errorf("esperava %d beneficios, veio %+v", c.beneficios, ben)
			}
		})
	}
}

func TestMontarResumo(t *testing.T) {
	ts := time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)
	casos := []struct {
		nome          string
		recursos      []Recurso
		total         int
		locais        []string
		totalGabinete int
		primeiroVinc  string
	}{
		{nome: "sem dados", total: 0, locais: []string{}},
		{
			nome: "gabinete antes de escritorio, vinculos por quantidade",
			recursos: []Recurso{
				{Local: LocalEscritorio, Vinculo: "Comissionado", Quantidade: 48, AtualizadoEm: ts},
				{Local: LocalGabinete, Vinculo: "Efetivo", Quantidade: 3, AtualizadoEm: ts},
				{Local: LocalGabinete, Vinculo: "Comissionado", Quantidade: 21, AtualizadoEm: ts},
			},
			total: 72, locais: []string{LocalGabinete, LocalEscritorio}, totalGabinete: 24, primeiroVinc: "Comissionado",
		},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			r := montarResumo(1, 2026, nil, c.recursos, nil)
			if r.Total != c.total || len(r.Locais) != len(c.locais) {
				t.Fatalf("resumo errado: %+v", r)
			}
			for i, l := range c.locais {
				if r.Locais[i].Local != l {
					t.Errorf("ordem dos locais: %+v", r.Locais)
				}
			}
			if len(c.locais) > 0 {
				if r.Locais[0].Total != c.totalGabinete || r.Locais[0].Vinculos[0].Vinculo != c.primeiroVinc {
					t.Errorf("gabinete errado: %+v", r.Locais[0])
				}
				if r.PorVinculo[0].Vinculo != "Comissionado" || r.PorVinculo[0].Quantidade != 69 {
					t.Errorf("por vinculo errado: %+v", r.PorVinculo)
				}
				if r.AtualizadoEm == nil || !r.AtualizadoEm.Equal(ts) {
					t.Errorf("atualizado_em errado: %v", r.AtualizadoEm)
				}
			} else if r.AtualizadoEm != nil || r.AnosDisponiveis == nil || r.Beneficios == nil {
				t.Errorf("resposta vazia deve ter listas vazias e sem data: %+v", r)
			}
		})
	}
}

func contar(t *testing.T, repo *Repository, senadorID, ano int) (int, int, int) {
	t.Helper()
	rec, err := repo.Recursos(senadorID, ano)
	if err != nil {
		t.Fatal(err)
	}
	ben, err := repo.Beneficios(senadorID, ano)
	if err != nil {
		t.Fatal(err)
	}
	soma := 0
	for _, r := range rec {
		soma += r.Quantidade
	}
	return len(rec), soma, len(ben)
}

func TestSubstituirSenadorAno_IdempotenteEEncolhe(t *testing.T) {
	repo := NewRepository(testdb.Abrir(t, Modelos()...))
	t0 := time.Date(2026, 9, 1, 10, 0, 0, 123456789, time.UTC)
	novo := func() ([]Recurso, []Beneficio) {
		return []Recurso{
				{Local: LocalGabinete, Vinculo: "Comissionado", Quantidade: 21},
				{Local: LocalGabinete, Vinculo: "Efetivo", Quantidade: 3},
				{Local: LocalEscritorio, Vinculo: "Comissionado", Quantidade: 48},
			}, []Beneficio{
				{Tipo: "Auxílio-Moradia", Utilizacao: "Não utilizou"},
				{Tipo: "Imóvel Funcional", Utilizacao: "Utilizou"},
			}
	}

	// duas rodadas iguais nao duplicam
	for i := 0; i < 2; i++ {
		rec, ben := novo()
		if err := repo.SubstituirSenadorAno(1, 2026, rec, ben, t0.Add(time.Duration(i)*time.Hour)); err != nil {
			t.Fatal(err)
		}
		if n, soma, nb := contar(t, repo, 1, 2026); n != 3 || soma != 72 || nb != 2 {
			t.Fatalf("rodada %d: esperado 3 linhas/72/2 beneficios, veio %d/%d/%d", i+1, n, soma, nb)
		}
	}
	// outro ano e outro senador nao sao tocados
	rec, ben := novo()
	if err := repo.SubstituirSenadorAno(2, 2026, rec, ben, t0); err != nil {
		t.Fatal(err)
	}
	rec, ben = novo()
	if err := repo.SubstituirSenadorAno(1, 2025, rec, ben, t0); err != nil {
		t.Fatal(err)
	}

	// a fonte encolheu: o efetivo saiu e um beneficio sumiu
	if err := repo.SubstituirSenadorAno(1, 2026,
		[]Recurso{{Local: LocalGabinete, Vinculo: "Comissionado", Quantidade: 20}, {Local: LocalEscritorio, Vinculo: "Comissionado", Quantidade: 48}},
		[]Beneficio{{Tipo: "Imóvel Funcional", Utilizacao: "Não utilizou"}},
		t0.Add(3*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if n, soma, nb := contar(t, repo, 1, 2026); n != 2 || soma != 68 || nb != 1 {
		t.Fatalf("apos encolher: esperado 2/68/1, veio %d/%d/%d", n, soma, nb)
	}
	bens, _ := repo.Beneficios(1, 2026)
	if bens[0].Utilizacao != "Não utilizou" {
		t.Errorf("upsert nao atualizou a utilizacao: %+v", bens[0])
	}
	for _, alvo := range [][2]int{{2, 2026}, {1, 2025}} {
		if n, soma, _ := contar(t, repo, alvo[0], alvo[1]); n != 3 || soma != 72 {
			t.Errorf("senador %d/%d foi alterado: %d linhas, %d servidores", alvo[0], alvo[1], n, soma)
		}
	}
	anos, _ := repo.AnosDisponiveis(1)
	if len(anos) != 2 || anos[0] != 2026 {
		t.Errorf("anos disponiveis errados: %v", anos)
	}
}

func TestSubstituirMesa(t *testing.T) {
	repo := NewRepository(testdb.Abrir(t, Modelos()...))
	t0 := time.Now()
	if err := repo.SubstituirMesa([]CargoMesa{{SenadorID: 1, Cargo: "PRESIDENTE"}, {SenadorID: 2, Cargo: "1º SECRETÁRIO"}}, t0); err != nil {
		t.Fatal(err)
	}
	if err := repo.SubstituirMesa(nil, t0.Add(time.Hour)); err == nil {
		t.Error("lista vazia deveria dar erro sem apagar")
	}
	if c, _ := repo.CargoMesaAtual(2); c == nil {
		t.Error("lista vazia apagou a mesa")
	}
	if err := repo.SubstituirMesa([]CargoMesa{{SenadorID: 1, Cargo: "PRESIDENTE"}}, t0.Add(2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if c, _ := repo.CargoMesaAtual(2); c != nil {
		t.Errorf("quem saiu da mesa deveria sumir: %+v", c)
	}
	if c, _ := repo.CargoMesaAtual(1); c == nil || c.Cargo != "PRESIDENTE" {
		t.Errorf("presidente sumiu: %+v", c)
	}
}

// fakes do sync
type senadoresFake []senador.Senador

func (f senadoresFake) FindAll(bool) ([]senador.Senador, error) { return f, nil }

type fonteFake map[int]*senado.RecursosUtilizadosAPI

func (f fonteFake) RecursosUtilizados(_ context.Context, codigo, _ int) (*senado.RecursosUtilizadosAPI, error) {
	r, ok := f[codigo]
	if !ok {
		return nil, errors.New("falha simulada")
	}
	return r, nil
}

func TestSyncAno(t *testing.T) {
	sens := senadoresFake{{ID: 1, CodigoParlamentar: 5012}, {ID: 2, CodigoParlamentar: 5982}, {ID: 3, CodigoParlamentar: 3830}}
	cheio := fonteFake{
		5012: {Pessoal: []senado.PessoalLocalAPI{pessoal("Gabinete", vinc("Comissionado", 21), vinc("Efetivo", 3))}},
		5982: {Pessoal: []senado.PessoalLocalAPI{pessoal("Gabinete", vinc("Comissionado", 15), vinc("Efetivo", 1))}},
		3830: {Pessoal: []senado.PessoalLocalAPI{pessoal("Gabinete", vinc("Comissionado", 9))}},
	}
	vazio := fonteFake{5012: {}, 5982: {}, 3830: {}}

	casos := []struct {
		nome       string
		fonte      fonteFake
		erro       bool
		servidores int // total no banco depois da rodada
	}{
		{nome: "carga inicial", fonte: cheio, servidores: 49},
		{nome: "repetir nao duplica", fonte: cheio, servidores: 49},
		{nome: "API vazia para todos aborta sem apagar", fonte: vazio, erro: true, servidores: 49},
		{nome: "falha de um senador preserva os dados dele", fonte: fonteFake{5012: cheio[5012], 5982: cheio[5982]}, servidores: 49},
	}
	repo := NewRepository(testdb.Abrir(t, Modelos()...))
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			svc := NewSyncService(repo, sens, c.fonte, nil)
			_, err := svc.SyncAno(context.Background(), 2026)
			if (err != nil) != c.erro {
				t.Fatalf("erro = %v, esperava erro = %v", err, c.erro)
			}
			total := 0
			for _, s := range sens {
				_, soma, _ := contar(t, repo, s.ID, 2026)
				total += soma
			}
			if total != c.servidores {
				t.Errorf("esperado %d servidores no banco, veio %d", c.servidores, total)
			}
		})
	}
}

func TestHandlerGetBySenador(t *testing.T) {
	db := testdb.Abrir(t, append([]any{&senador.Senador{}, &senador.Mandato{}}, Modelos()...)...)
	if err := db.Create(&senador.Senador{ID: 3, CodigoParlamentar: 3830, Nome: "Davi Alcolumbre", EmExercicio: true}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&senador.Senador{ID: 4, CodigoParlamentar: 1, Nome: "Sem dados", EmExercicio: true}).Error; err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(db)
	agora := time.Now()
	_ = repo.SubstituirSenadorAno(3, 2025, []Recurso{{Local: LocalGabinete, Vinculo: "Comissionado", Quantidade: 12}}, nil, agora)
	_ = repo.SubstituirSenadorAno(3, 2026, []Recurso{{Local: LocalGabinete, Vinculo: "Comissionado", Quantidade: 9}}, nil, agora)
	_ = repo.SubstituirMesa([]CargoMesa{{SenadorID: 3, Cargo: "PRESIDENTE"}}, agora)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/senadores/:id/gabinete", NewHandler(repo, senador.NewRepository(db)).GetBySenador)

	casos := []struct {
		nome   string
		url    string
		status int
		ano    int
		total  int
		mesa   bool
	}{
		{"sem ano usa o mais recente", "/senadores/3/gabinete", 200, 2026, 9, true},
		{"ano 0 (mandato) usa o mais recente", "/senadores/3/gabinete?ano=0", 200, 2026, 9, true},
		{"ano explicito", "/senadores/3/gabinete?ano=2025", 200, 2025, 12, true},
		{"ano sem dados", "/senadores/3/gabinete?ano=2023", 200, 2023, 0, true},
		{"senador sem dados", "/senadores/4/gabinete", 200, 0, 0, false},
		{"senador inexistente", "/senadores/99/gabinete", 404, 0, 0, false},
		{"id invalido", "/senadores/abc/gabinete", 400, 0, 0, false},
		{"ano invalido", "/senadores/3/gabinete?ano=x", 400, 0, 0, false},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, c.url, nil))
			if w.Code != c.status {
				t.Fatalf("status %d, esperado %d: %s", w.Code, c.status, w.Body.String())
			}
			if c.status != 200 {
				return
			}
			var res Resumo
			if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
				t.Fatal(err)
			}
			if res.Ano != c.ano || res.Total != c.total || (res.CargoMesa != nil) != c.mesa {
				t.Errorf("resposta errada: %s", w.Body.String())
			}
			if res.FonteURL == "" || res.Locais == nil || res.Beneficios == nil {
				t.Errorf("campos obrigatorios ausentes: %s", w.Body.String())
			}
		})
	}
}
