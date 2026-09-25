package senador

import (
	"testing"
	"time"

	"github.com/Alzarus/to-de-olho/internal/testdb"
)

func TestNormalizarFotoURL(t *testing.T) {
	casos := []struct {
		nome, entrada, esperado string
	}{
		{"foto do Senado em http vira https",
			"http://www.senado.leg.br/senadores/img/fotos-oficiais/senador5895.jpg",
			"https://www.senado.leg.br/senadores/img/fotos-oficiais/senador5895.jpg"},
		{"esquema em maiúsculas", "HTTP://www.senado.leg.br/a.jpg", "https://www.senado.leg.br/a.jpg"},
		{"outro subdomínio do Senado", "http://legis.senado.leg.br/senadores/fotos-oficiais/5895", "https://legis.senado.leg.br/senadores/fotos-oficiais/5895"},
		{"já em https fica igual", "https://www.senado.leg.br/a.jpg", "https://www.senado.leg.br/a.jpg"},
		{"host de fora do Senado fica igual", "http://exemplo.com/a.jpg", "http://exemplo.com/a.jpg"},
		{"host que só termina parecido fica igual", "http://naosenado.leg.br/a.jpg", "http://naosenado.leg.br/a.jpg"},
		{"porta explícita fica igual", "http://www.senado.leg.br:80/a.jpg", "http://www.senado.leg.br:80/a.jpg"},
		{"vazio fica vazio", "", ""},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if got := NormalizarFotoURL(c.entrada); got != c.esperado {
				t.Errorf("NormalizarFotoURL(%q) = %q; esperado %q", c.entrada, got, c.esperado)
			}
		})
	}
}

func TestNormalizarFotosHTTPSNoBanco(t *testing.T) {
	db := testdb.Abrir(t, &Senador{}, &Mandato{})

	antigo := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	senadores := []Senador{
		{CodigoParlamentar: 1, Nome: "Http do Senado", FotoURL: "http://www.senado.leg.br/senadores/img/fotos-oficiais/senador1.jpg"},
		{CodigoParlamentar: 2, Nome: "Já em https", FotoURL: "https://www.senado.leg.br/senadores/img/fotos-oficiais/senador2.jpg"},
		{CodigoParlamentar: 3, Nome: "Outro host", FotoURL: "http://exemplo.com/3.jpg"},
		{CodigoParlamentar: 4, Nome: "Sem foto"},
		{CodigoParlamentar: 5, Nome: "Parecido", FotoURL: "http://naosenado.leg.br/5.jpg"},
	}
	if err := db.Create(&senadores).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("UPDATE senadores SET updated_at = ?", antigo).Error; err != nil {
		t.Fatal(err)
	}

	// Duas vezes: a segunda não pode mudar nada (idempotente).
	for i := 0; i < 2; i++ {
		if err := NormalizarFotosHTTPS(db); err != nil {
			t.Fatalf("execução %d: %v", i+1, err)
		}
	}

	esperado := map[int]string{
		1: "https://www.senado.leg.br/senadores/img/fotos-oficiais/senador1.jpg",
		2: "https://www.senado.leg.br/senadores/img/fotos-oficiais/senador2.jpg",
		3: "http://exemplo.com/3.jpg",
		4: "",
		5: "http://naosenado.leg.br/5.jpg",
	}
	var gravados []Senador
	if err := db.Order("codigo_parlamentar").Find(&gravados).Error; err != nil {
		t.Fatal(err)
	}
	for _, s := range gravados {
		if s.FotoURL != esperado[s.CodigoParlamentar] {
			t.Errorf("senador %d: foto_url = %q; esperado %q", s.CodigoParlamentar, s.FotoURL, esperado[s.CodigoParlamentar])
		}
		if !s.UpdatedAt.Equal(antigo) {
			t.Errorf("senador %d: updated_at mudou para %v (não é dado novo do Senado)", s.CodigoParlamentar, s.UpdatedAt)
		}
	}
}
