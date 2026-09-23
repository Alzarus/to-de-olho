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
