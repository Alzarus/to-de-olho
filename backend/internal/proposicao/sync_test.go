package proposicao

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Alzarus/to-de-olho/internal/senador"
	senadoapi "github.com/Alzarus/to-de-olho/pkg/senado"
)

type clienteFalso struct {
	mu        sync.Mutex
	processos map[int]*senadoapi.ProcessoDetalhe
	chamadas  map[int]int
	falhar    bool
}

func (c *clienteFalso) ListarProposicoesParlamentar(context.Context, int) ([]senadoapi.MateriaAPI, error) {
	return nil, nil
}

func (c *clienteFalso) ObterProcesso(_ context.Context, id int) (*senadoapi.ProcessoDetalhe, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.chamadas[id]++
	if c.falhar {
		return nil, errors.New("corpo cortado")
	}
	return c.processos[id], nil
}

func ptr(v int) *int { return &v }

func TestMontarProposicoes(t *testing.T) {
	cliente := &clienteFalso{
		chamadas: map[int]int{},
		processos: map[int]*senadoapi.ProcessoDetalhe{
			// "e outros": o texto nao mostra o Alan Rick, o detalhe mostra na 3a posicao
			200: {ID: 200, AutoriaIniciativa: []senadoapi.AutorIniciativa{
				{Autor: "Magno Malta", Ordem: 1, CodigoParlamentar: ptr(631)},
				{Autor: "Styvenson Valentim", Ordem: 2, CodigoParlamentar: ptr(5959)},
				{Autor: "Alan Rick", Ordem: 3, CodigoParlamentar: ptr(5672)},
			}},
			// autor nao parlamentar: senador fora da autoria
			300: {ID: 300, AutoriaIniciativa: []senadoapi.AutorIniciativa{
				{Autor: "Câmara dos Deputados", Ordem: 1},
			}},
		},
	}
	s := &SyncService{client: cliente, detalhes: map[int]*senadoapi.ProcessoDetalhe{}}
	sen := senador.Senador{ID: 7, CodigoParlamentar: 5672, Nome: "Alan Rick"}

	lista := []senadoapi.MateriaAPI{
		{ID: 100, CodigoMateria: 1, Identificacao: "PEC 5/2024", Autoria: "Senador Alan Rick (UNIÃO/AC), Senador Magno Malta (PL/ES)", DataApresentacao: "2024-03-01"},
		{ID: 101, CodigoMateria: 2, Identificacao: "PEC 6/2024", Autoria: "Senador Magno Malta (PL/ES), Senador Alan Rick (UNIÃO/AC)", DataApresentacao: "2024-03-01"},
		{ID: 200, CodigoMateria: 3, Identificacao: "PEC 7/2024", Autoria: "Senador Magno Malta (PL/ES) e outros.", DataApresentacao: "2024-03-01"},
		{ID: 200, CodigoMateria: 3, Identificacao: "PEC 7/2024", Autoria: "Senador Magno Malta (PL/ES) e outros."}, // duplicada
		{ID: 300, CodigoMateria: 4, Identificacao: "PL 1/2024", Autoria: "Câmara dos Deputados"},
	}

	got, resumo, err := s.montarProposicoes(context.Background(), sen, lista)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Fatalf("esperado 4 linhas (duplicata removida), obtido %d", len(got))
	}

	esperado := []struct {
		posicao, total *int
		pontua         bool
	}{
		{ptr(1), ptr(2), true},  // primeiro autor: PEC apresentada = 1 x 3
		{ptr(2), ptr(2), false}, // coautor: nao pontua
		{ptr(3), ptr(3), false}, // posicao vinda do detalhe
		{nil, ptr(1), false},    // fora da autoria
	}
	for i, e := range esperado {
		p := got[i]
		if !igual(p.PosicaoAutoria, e.posicao) || !igual(p.TotalAutores, e.total) {
			t.Errorf("linha %d: posicao=%v total=%v; esperado %v %v", i, val(p.PosicaoAutoria), val(p.TotalAutores), val(e.posicao), val(e.total))
		}
		if (p.Pontuacao > 0) != e.pontua {
			t.Errorf("linha %d: pontuacao %v; pontua esperado %v", i, p.Pontuacao, e.pontua)
		}
	}
	if got[0].Pontuacao != 3 {
		t.Errorf("PEC apresentada de autoria principal vale 3, obtido %v", got[0].Pontuacao)
	}
	if resumo != (ResumoAutoria{Texto: 2, Detalhe: 1, SemPosicao: 1}) {
		t.Errorf("resumo inesperado: %+v", resumo)
	}

	// o detalhe do mesmo processo e buscado uma vez so, mesmo para outro senador
	outro := senador.Senador{ID: 8, CodigoParlamentar: 5959, Nome: "Styvenson Valentim"}
	if _, _, err := s.montarProposicoes(context.Background(), outro, lista[2:3]); err != nil {
		t.Fatal(err)
	}
	if cliente.chamadas[200] != 1 {
		t.Errorf("processo 200 buscado %d vezes; esperado 1 (cache)", cliente.chamadas[200])
	}
}

func TestMontarProposicoesFalhaNoDetalheDevolveErro(t *testing.T) {
	s := &SyncService{client: &clienteFalso{chamadas: map[int]int{}, falhar: true}, detalhes: map[int]*senadoapi.ProcessoDetalhe{}}
	sen := senador.Senador{ID: 7, CodigoParlamentar: 5672, Nome: "Alan Rick"}
	lista := []senadoapi.MateriaAPI{{ID: 200, CodigoMateria: 3, Autoria: "Senador Magno Malta (PL/ES) e outros."}}

	if _, _, err := s.montarProposicoes(context.Background(), sen, lista); err == nil {
		t.Fatal("esperado erro quando o detalhe falha: a linha nao pode entrar sem posicao")
	}
}

func TestCalcularPontuacaoSoPrimeiroAutor(t *testing.T) {
	casos := []struct {
		posicao  *int
		esperado float64
	}{
		{ptr(1), 3 * 16},
		{ptr(2), 0},
		{nil, 0},
	}
	for _, c := range casos {
		p := Proposicao{SiglaSubtipoMateria: "PEC", EstagioTramitacao: "TransformadoLei", PosicaoAutoria: c.posicao}
		if got := p.CalcularPontuacao(); got != c.esperado {
			t.Errorf("posicao %v: pontuacao %v; esperado %v", val(c.posicao), got, c.esperado)
		}
	}
}

func igual(a, b *int) bool { return (a == nil && b == nil) || (a != nil && b != nil && *a == *b) }

func val(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}
