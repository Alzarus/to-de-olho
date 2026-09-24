package acesso

import (
	"bytes"
	"testing"
	"time"

	"github.com/Alzarus/to-de-olho/internal/testdb"
)

func TestEhRobo(t *testing.T) {
	casos := []struct {
		ua   string
		robo bool
	}{
		{"", true},
		{"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)", true},
		{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 HeadlessChrome/120.0", true},
		{"curl/8.4.0", true},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/128.0 Safari/537.36", false},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/605.1.15 Mobile/15E148", false},
	}
	for _, c := range casos {
		if got := EhRobo(c.ua); got != c.robo {
			t.Errorf("EhRobo(%q) = %v, esperado %v", c.ua, got, c.robo)
		}
	}
}

func TestDiaDeUsaHorarioDeBrasilia(t *testing.T) {
	// 01:30 UTC de 24/09 ainda e 23/09 em Brasilia
	got := diaDe(time.Date(2026, 9, 24, 1, 30, 0, 0, time.UTC))
	if want := time.Date(2026, 9, 23, 0, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Errorf("diaDe = %v, esperado %v", got, want)
	}
}

func TestHashMudaComOSal(t *testing.T) {
	a := hashVisitante([]byte("sal-1"), "200.1.2.3", "ua")
	b := hashVisitante([]byte("sal-2"), "200.1.2.3", "ua")
	if bytes.Equal(a, b) {
		t.Error("o mesmo visitante em dias diferentes nao pode gerar o mesmo hash")
	}
}

func TestRegistrarContaUmaVezPorDiaEApagaSalVelho(t *testing.T) {
	db := testdb.Abrir(t, &Visita{}, &Sal{})
	repo := NewRepository(db)
	agora := time.Date(2026, 9, 23, 15, 0, 0, 0, time.UTC)
	repo.agora = func() time.Time { return agora }

	for _, v := range []struct{ ip, ua string }{
		{"200.1.2.3", "chrome"},
		{"200.1.2.3", "chrome"}, // repetido no mesmo dia: nao conta
		{"200.1.2.3", "firefox"},
		{"200.9.9.9", "chrome"},
	} {
		if err := repo.Registrar(v.ip, v.ua); err != nil {
			t.Fatal(err)
		}
	}
	agora = agora.AddDate(0, 0, 1)
	if err := repo.Registrar("200.1.2.3", "chrome"); err != nil { // outro dia: conta
		t.Fatal(err)
	}

	r, err := repo.Resumo()
	if err != nil {
		t.Fatal(err)
	}
	if r.Total != 4 {
		t.Errorf("total = %d, esperado 4", r.Total)
	}
	if r.Desde == nil || r.Desde.Format("2006-01-02") != "2026-09-23" {
		t.Errorf("desde = %v, esperado 2026-09-23", r.Desde)
	}

	var sais int64
	db.Model(&Sal{}).Count(&sais)
	if sais != 1 {
		t.Errorf("deveria sobrar so o sal do dia; sobraram %d", sais)
	}
}

func TestResumoSemVisitas(t *testing.T) {
	r, err := NewRepository(testdb.Abrir(t, &Visita{}, &Sal{})).Resumo()
	if err != nil {
		t.Fatal(err)
	}
	if r.Total != 0 || r.Desde != nil {
		t.Errorf("sem visitas: esperado 0 e nil, veio %d e %v", r.Total, r.Desde)
	}
}
