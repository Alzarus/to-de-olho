package utils

import "testing"

func TestLerPaginacao(t *testing.T) {
	casos := []struct {
		nome           string
		limit, page    string
		padrao, maximo int
		quer           Paginacao
	}{
		{"ausentes usam o padrão", "", "", 20, 100, Paginacao{Limit: 20, Page: 1, Offset: 0}},
		{"valores normais", "50", "3", 20, 100, Paginacao{Limit: 50, Page: 3, Offset: 100}},
		{"limit acima do máximo é truncado", "100000", "1", 20, 100, Paginacao{Limit: 100, Page: 1, Offset: 0}},
		{"limit no máximo passa", "100", "2", 20, 100, Paginacao{Limit: 100, Page: 2, Offset: 100}},
		{"limit zero volta ao padrão", "0", "1", 20, 100, Paginacao{Limit: 20, Page: 1, Offset: 0}},
		{"limit negativo volta ao padrão", "-5", "1", 20, 100, Paginacao{Limit: 20, Page: 1, Offset: 0}},
		{"limit não numérico volta ao padrão", "abc", "1", 20, 100, Paginacao{Limit: 20, Page: 1, Offset: 0}},
		{"page zero vira 1", "20", "0", 20, 100, Paginacao{Limit: 20, Page: 1, Offset: 0}},
		{"page não numérica vira 1", "20", "x", 20, 100, Paginacao{Limit: 20, Page: 1, Offset: 0}},
		{"page enorme para no offset máximo", "20", "99999999", 20, 100,
			Paginacao{Limit: 20, Page: OffsetMaximo/20 + 1, Offset: OffsetMaximo}},
		{"page perto do int máximo não estoura", "100", "9223372036854775807", 20, 100,
			Paginacao{Limit: 100, Page: OffsetMaximo/100 + 1, Offset: OffsetMaximo}},
		{"page acima do int vira 1", "100", "99999999999999999999", 20, 100, Paginacao{Limit: 100, Page: 1, Offset: 0}},
		{"exportação do front (500 páginas de 100) cabe", "100", "500", 20, 100, Paginacao{Limit: 100, Page: 500, Offset: 49_900}},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			got := LerPaginacao(c.limit, c.page, c.padrao, c.maximo)
			if got != c.quer {
				t.Errorf("LerPaginacao(%q, %q) = %+v; quer %+v", c.limit, c.page, got, c.quer)
			}
			if got.Offset < 0 || got.Offset > OffsetMaximo {
				t.Errorf("offset fora do intervalo: %d", got.Offset)
			}
		})
	}
}
