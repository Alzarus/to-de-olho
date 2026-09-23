package proposicao

import "testing"

func TestPosicaoNoTexto(t *testing.T) {
	pec := "Senador Magno Malta (PL/ES), Senador Styvenson Valentim (PODEMOS/RN), Senador Eduardo Girão (NOVO/CE), Senador Eduardo Gomes (PL/TO), Senadora Damares Alves (REPUBLICANOS/DF)"

	casos := []struct {
		nome    string
		autoria string
		senador string
		posicao int
		total   int
		ok      bool
	}{
		{"primeiro autor", pec, "Magno Malta", 1, 5, true},
		{"coautor", pec, "Eduardo Gomes", 4, 5, true},
		{"senadora, sem acento no cadastro", pec, "Damares Alves", 5, 5, true},
		{"acento so no texto", pec, "Eduardo Girao", 3, 5, true},
		{"nome que e prefixo de outro nao casa", pec, "Eduardo", 0, 0, false},
		{"autor unico", "Senador Alan Rick (UNIÃO/AC)", "Alan Rick", 1, 1, true},
		{"caixa e espacos", "Senador  ALAN   rick (UNIÃO/AC)", "Alan Rick", 1, 1, true},
		{"e outros, primeiro autor", "Senador Magno Malta (PL/ES) e outros.", "Magno Malta", 1, 0, true},
		{"e outros, senador oculto", "Senador Magno Malta (PL/ES) e outros.", "Alan Rick", 0, 0, false},
		{"nao citado", pec, "Alan Rick", 0, 0, false},
		{"autor nao parlamentar", "Câmara dos Deputados", "Alan Rick", 0, 0, false},
		{"comissao com virgula no nome", "Comissão de Constituição, Justiça e Cidadania", "Alan Rick", 0, 0, false},
		{"nome repetido e ambiguo", "Senador Alan Rick (UNIÃO/AC), Senador Alan Rick (PL/AC)", "Alan Rick", 0, 0, false},
		{"vazio", "", "Alan Rick", 0, 0, false},
		{"nome com ponto", "Senador Dr. Hiran (PP/RR), Senador Alan Rick (UNIÃO/AC)", "Dr. Hiran", 1, 2, true},
		{"como deputado vai para o detalhe", "Deputado Alan Rick (UNIÃO/AC)", "Alan Rick", 0, 0, false},
		{"como lider vai para o detalhe", "Líder do PL Alan Rick (PL/AC), Senador Magno Malta (PL/ES)", "Alan Rick", 0, 0, false},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			pos, total, ok := PosicaoNoTexto(c.autoria, c.senador)
			if pos != c.posicao || total != c.total || ok != c.ok {
				t.Errorf("PosicaoNoTexto(%q) = (%d, %d, %v); esperado (%d, %d, %v)",
					c.senador, pos, total, ok, c.posicao, c.total, c.ok)
			}
		})
	}
}
