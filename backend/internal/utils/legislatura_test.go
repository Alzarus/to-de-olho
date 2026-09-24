package utils

import (
	"testing"
	"time"
)

func TestInicioLegislatura(t *testing.T) {
	casos := []struct {
		data     time.Time
		esperado string
	}{
		{time.Date(2026, 9, 23, 0, 0, 0, 0, time.UTC), "2023-02-01"},
		{time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC), "2023-02-01"},
		{time.Date(2023, 1, 31, 23, 0, 0, 0, time.UTC), "2019-02-01"},
		{time.Date(2027, 1, 15, 0, 0, 0, 0, time.UTC), "2023-02-01"},
		{time.Date(2027, 2, 1, 0, 0, 0, 0, time.UTC), "2027-02-01"},
		{time.Date(2030, 12, 31, 0, 0, 0, 0, time.UTC), "2027-02-01"},
	}
	for _, c := range casos {
		if got := InicioLegislatura(c.data).Format("2006-01-02"); got != c.esperado {
			t.Errorf("InicioLegislatura(%s) = %s; esperado %s", c.data.Format("2006-01-02"), got, c.esperado)
		}
	}
}

func TestInicioRecorteSobrescritoPorVariavel(t *testing.T) {
	t.Setenv("RECORTE_INICIO", "2019-02-01")
	if got := InicioRecorte().Format("2006-01-02"); got != "2019-02-01" {
		t.Errorf("InicioRecorte() = %s; esperado 2019-02-01", got)
	}
	t.Setenv("RECORTE_INICIO", "lixo")
	if got := InicioRecorte(); !got.Equal(InicioLegislatura(time.Now())) {
		t.Errorf("valor invalido deveria cair no padrao, obtido %s", got)
	}
}

func TestPeriodoDoAno(t *testing.T) {
	t.Setenv("RECORTE_INICIO", "2023-02-01")
	ini, fim := PeriodoDoAno(2023)
	if ini.Format("2006-01-02") != "2023-02-01" || fim.Format("2006-01-02") != "2024-01-01" {
		t.Errorf("2023: %s..%s (janeiro e da legislatura anterior)", ini, fim)
	}
	ini, fim = PeriodoDoAno(2024)
	if ini.Format("2006-01-02") != "2024-01-01" || fim.Format("2006-01-02") != "2025-01-01" {
		t.Errorf("2024: %s..%s", ini, fim)
	}
	ano := time.Now().Year()
	if _, fim = PeriodoDoAno(ano); fim.After(time.Now()) {
		t.Errorf("ano corrente termina hoje, nao em dezembro: %s", fim)
	}
}

// Item 14: nos 6 primeiros meses da 58a legislatura o ranking do mandato
// continua na 57a, ja encerrada; depois passa para a 58a.
func TestPeriodoDoMandatoNaViradaDeLegislatura(t *testing.T) {
	t.Setenv("RECORTE_INICIO", "")
	d := func(s string) time.Time { v, _ := time.Parse("2006-01-02", s); return v }
	casos := []struct {
		agora, inicio, fim string
		encerrada          bool
	}{
		{"2026-09-23", "2023-02-01", "2026-09-23", false},
		{"2027-01-31", "2023-02-01", "2027-01-31", false},
		{"2027-02-01", "2023-02-01", "2027-02-01", true},
		{"2027-07-31", "2023-02-01", "2027-02-01", true},
		{"2027-08-01", "2027-02-01", "2027-08-01", false},
	}
	for _, c := range casos {
		ini, fim, enc := periodoDoMandatoEm(d(c.agora))
		if ini.Format("2006-01-02") != c.inicio || fim.Format("2006-01-02") != c.fim || enc != c.encerrada {
			t.Errorf("%s: %s..%s encerrada=%v; esperado %s..%s encerrada=%v",
				c.agora, ini.Format("2006-01-02"), fim.Format("2006-01-02"), enc, c.inicio, c.fim, c.encerrada)
		}
	}
}

func TestPeriodoDoMandatoRespeitaRecorteForcado(t *testing.T) {
	t.Setenv("RECORTE_INICIO", "2019-02-01")
	agora := time.Date(2027, 3, 1, 0, 0, 0, 0, time.UTC)
	ini, fim, enc := periodoDoMandatoEm(agora)
	if ini.Format("2006-01-02") != "2019-02-01" || !fim.Equal(agora) || enc {
		t.Errorf("recorte forcado ignorado: %s..%s encerrada=%v", ini, fim, enc)
	}
}

// Sem RECORTE_INICIO, cada ano e cortado pela posse da propria legislatura:
// 2026 inteiro e da 57a mesmo depois da posse da 58a.
func TestPeriodoDoAnoPorLegislatura(t *testing.T) {
	t.Setenv("RECORTE_INICIO", "")
	ini, fim := PeriodoDoAno(2023)
	if ini.Format("2006-01-02") != "2023-02-01" || fim.Format("2006-01-02") != "2024-01-01" {
		t.Errorf("2023: %s..%s", ini, fim)
	}
	ini, _ = PeriodoDoAno(2026)
	if ini.Format("2006-01-02") != "2026-01-01" {
		t.Errorf("2026 deveria comecar em 01/01: %s", ini)
	}
}
