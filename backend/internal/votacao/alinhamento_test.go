package votacao

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func votoAlinhamento(senadorID, codigo int, sigla, siglaMateria string, secreta bool, ano int) Votacao {
	s := secreta
	return Votacao{
		SenadorID: senadorID, CodigoVotacao: codigo, SessaoID: "1", CodigoSessao: "1",
		Data:      time.Date(ano, 5, 1, 12, 0, 0, 0, time.UTC).Add(time.Duration(codigo) * time.Hour),
		SiglaVoto: sigla, Voto: rotuloVoto(sigla), Materia: siglaMateria + " 1/2025",
		SiglaMateria: siglaMateria, Secreta: &s,
	}
}

// cenarioAlinhamento: senadores 1, 2 e 3 (tabela de novoRepo).
//
//	votacao  tipo  secreta  sen1       sen2       sen3
//	10       PL    nao      Sim        Sim        Não
//	11       PL    nao      Não        Sim        Não
//	12       PEC   nao      Abstenção  Abstenção  -
//	13       PEC   nao      Sim        P-NRV      AP       (so o 1 votou: nao conta)
//	14       PL    sim      Votou      Votou      Votou    (secreta: fora)
//	15       PL    nao      Sim        Não        Sim      (2024)
//	16       MPV   nulo     Sim        Sim        Sim      (secreta nula de linha antiga: aberta)
func cenarioAlinhamento(t *testing.T) *Repository {
	t.Helper()
	repo := novoRepo(t)
	lote := []Votacao{
		votoAlinhamento(1, 10, "Sim", "PL", false, 2025), votoAlinhamento(2, 10, "Sim", "PL", false, 2025), votoAlinhamento(3, 10, "Não", "PL", false, 2025),
		votoAlinhamento(1, 11, "Não", "PL", false, 2025), votoAlinhamento(2, 11, "Sim", "PL", false, 2025), votoAlinhamento(3, 11, "Não", "PL", false, 2025),
		votoAlinhamento(1, 12, "Abstenção", "PEC", false, 2025), votoAlinhamento(2, 12, "Abstenção", "PEC", false, 2025),
		votoAlinhamento(1, 13, "Sim", "PEC", false, 2025), votoAlinhamento(2, 13, "P-NRV", "PEC", false, 2025), votoAlinhamento(3, 13, "AP", "PEC", false, 2025),
		votoAlinhamento(1, 14, "Votou", "PL", true, 2025), votoAlinhamento(2, 14, "Votou", "PL", true, 2025), votoAlinhamento(3, 14, "Votou", "PL", true, 2025),
		votoAlinhamento(1, 15, "Sim", "PL", false, 2024), votoAlinhamento(2, 15, "Não", "PL", false, 2024), votoAlinhamento(3, 15, "Sim", "PL", false, 2024),
	}
	for _, s := range []int{1, 2, 3} {
		v := votoAlinhamento(s, 16, "Sim", "MPV", false, 2025)
		v.Secreta = nil
		lote = append(lote, v)
	}
	if err := repo.UpsertBatch(lote); err != nil {
		t.Fatal(err)
	}
	return repo
}

type parEsperado struct {
	a, b, comuns, iguais int
	pct                  *float64
}

func porcento(v float64) *float64 { return &v }

func TestAlinhamento(t *testing.T) {
	repo := cenarioAlinhamento(t)

	casos := []struct {
		nome         string
		filtro       FiltroAlinhamento
		pares        []parEsperado
		divergencias []int // codigos, da mais recente para a mais antiga
	}{
		{
			nome:   "2025, todos os tipos",
			filtro: FiltroAlinhamento{IDs: []int{3, 1, 2}, Ano: 2025},
			pares: []parEsperado{
				// 1x2: 10 igual, 11 diferente, 12 igual, 16 igual -> 3/4
				{1, 2, 4, 3, porcento(75)},
				// 1x3: 10 diferente, 11 igual, 16 igual -> 2/3
				{1, 3, 3, 2, porcento(66.7)},
				// 2x3: 10 diferente, 11 diferente, 16 igual -> 1/3
				{2, 3, 3, 1, porcento(33.3)},
			},
			divergencias: []int{11, 10},
		},
		{
			nome:   "todos os anos inclui 2024",
			filtro: FiltroAlinhamento{IDs: []int{1, 2}},
			pares:  []parEsperado{{1, 2, 5, 3, porcento(60)}},
			// 15 e de 2024, mas a hora soma o codigo: a ordem e pela data
			divergencias: []int{11, 15},
		},
		{
			nome:         "so PEC",
			filtro:       FiltroAlinhamento{IDs: []int{1, 2}, Ano: 2025, Tipos: []string{"PEC"}},
			pares:        []parEsperado{{1, 2, 1, 1, porcento(100)}},
			divergencias: nil,
		},
		{
			nome:         "par sem votacao em comum tem percentual nulo",
			filtro:       FiltroAlinhamento{IDs: []int{2, 3}, Ano: 2025, Tipos: []string{"PEC"}},
			pares:        []parEsperado{{2, 3, 0, 0, nil}},
			divergencias: nil,
		},
	}

	for _, tc := range casos {
		t.Run(tc.nome, func(t *testing.T) {
			res, err := repo.Alinhamento(tc.filtro)
			if err != nil {
				t.Fatal(err)
			}
			if len(res.Pares) != len(tc.pares) {
				t.Fatalf("esperados %d pares, vieram %+v", len(tc.pares), res.Pares)
			}
			for i, esp := range tc.pares {
				p := res.Pares[i]
				if p.SenadorA != esp.a || p.SenadorB != esp.b || p.VotacoesComuns != esp.comuns || p.VotosIguais != esp.iguais {
					t.Errorf("par %d: esperado %+v, obtido %+v", i, esp, p)
				}
				switch {
				case esp.pct == nil && p.Percentual != nil:
					t.Errorf("par %d: percentual deveria ser nulo, veio %v", i, *p.Percentual)
				case esp.pct != nil && (p.Percentual == nil || *p.Percentual != *esp.pct):
					t.Errorf("par %d: percentual esperado %v, obtido %v", i, *esp.pct, p.Percentual)
				}
			}
			if res.TotalDivergencias != len(tc.divergencias) || len(res.Divergencias) != len(tc.divergencias) {
				t.Fatalf("esperadas %d divergencias, vieram total=%d %+v", len(tc.divergencias), res.TotalDivergencias, res.Divergencias)
			}
			for i, cod := range tc.divergencias {
				if res.Divergencias[i].CodigoVotacao != cod {
					t.Errorf("divergencia %d: esperado codigo %d, obtido %d", i, cod, res.Divergencias[i].CodigoVotacao)
				}
			}
		})
	}
}

func TestAlinhamentoDivergenciaTrazVotosEfetivos(t *testing.T) {
	repo := cenarioAlinhamento(t)
	res, err := repo.Alinhamento(FiltroAlinhamento{IDs: []int{1, 2, 3}, Ano: 2025})
	if err != nil {
		t.Fatal(err)
	}
	d := res.Divergencias[len(res.Divergencias)-1] // votacao 10
	if d.CodigoVotacao != 10 || d.SiglaMateria != "PL" || d.Materia != "PL 1/2025" {
		t.Fatalf("metadados errados: %+v", d)
	}
	esperado := []VotoSenador{{1, "Sim"}, {2, "Sim"}, {3, "Não"}}
	if len(d.Votos) != len(esperado) {
		t.Fatalf("esperados %d votos, vieram %+v", len(esperado), d.Votos)
	}
	for i, v := range esperado {
		if d.Votos[i] != v {
			t.Errorf("voto %d: esperado %+v, obtido %+v", i, v, d.Votos[i])
		}
	}
}

func TestParseIDs(t *testing.T) {
	casos := []struct {
		entrada string
		ids     []int
		erro    bool
	}{
		{"1,2", []int{1, 2}, false},
		{" 3 , 1,2,3 ", []int{3, 1, 2}, false},
		{"1,2,3,4,5", []int{1, 2, 3, 4, 5}, false},
		{"1", nil, true},
		{"1,1", nil, true},
		{"", nil, true},
		{"1,2,3,4,5,6", nil, true},
		{"1,abc", nil, true},
		{"1,-2", nil, true},
	}
	for _, tc := range casos {
		t.Run(tc.entrada, func(t *testing.T) {
			ids, err := parseIDs(tc.entrada)
			if (err != nil) != tc.erro {
				t.Fatalf("erro esperado=%v, obtido %v", tc.erro, err)
			}
			if len(ids) != len(tc.ids) {
				t.Fatalf("esperado %v, obtido %v", tc.ids, ids)
			}
			for i := range ids {
				if ids[i] != tc.ids[i] {
					t.Errorf("esperado %v, obtido %v", tc.ids, ids)
				}
			}
		})
	}
}

func TestGetAlinhamentoValidaParametros(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(nil) // parametros invalidos nao chegam ao banco
	casos := []struct {
		nome  string
		query string
	}{
		{"sem ids", ""},
		{"um id so", "?ids=1"},
		{"ano invalido", "?ids=1,2&ano=abc"},
		{"ano negativo", "?ids=1,2&ano=-1"},
	}
	for _, tc := range casos {
		t.Run(tc.nome, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/votacoes/alinhamento"+tc.query, nil)
			h.GetAlinhamento(c)
			if w.Code != http.StatusBadRequest {
				t.Errorf("esperado 400, obtido %d", w.Code)
			}
		})
	}
}
