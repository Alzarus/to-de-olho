package senador

import (
	"math"
	"testing"
	"time"

	senadoapi "github.com/Alzarus/to-de-olho/pkg/senado"
)

func TestMesesEmExercicio(t *testing.T) {
	dia := func(s string) time.Time { v, _ := time.Parse("2006-01-02", s); return v }
	fim := func(s string) *time.Time { v := dia(s); return &v }
	inicio, agora := dia("2023-02-01"), dia("2026-09-23")

	// Renan Filho: dois periodos curtos em 2023 e volta em abril de 2026
	renan, err := converterExercicios(1, []senadoapi.Exercicio{
		{Inicio: "2023-02-01", Fim: "2023-02-02"},
		{Inicio: "2023-12-12", Fim: "2023-12-20"},
		{Inicio: "2026-04-01"},
	})
	if err != nil {
		t.Fatal(err)
	}
	// 2 + 9 + 175 dias
	if got := mesesEmExercicio(renan, inicio, agora); math.Abs(got-186/diasPorMes) > 0.01 {
		t.Errorf("Renan Filho: %.2f meses", got)
	}

	// titular desde 2019: so conta a partir do recorte
	titular := []Mandato{{Inicio: dia("2019-02-01")}}
	if got := mesesEmExercicio(titular, inicio, agora); math.Abs(got-43.7) > 0.1 {
		t.Errorf("titular desde 2019: %.2f meses", got)
	}

	// periodo anterior ao recorte nao conta
	antigo := []Mandato{{Inicio: dia("2015-02-01"), Fim: fim("2019-01-31")}}
	if got := mesesEmExercicio(antigo, inicio, agora); got != 0 {
		t.Errorf("periodo anterior contou %.2f meses", got)
	}
}

func TestConverterExerciciosDataInvalida(t *testing.T) {
	if _, err := converterExercicios(1, []senadoapi.Exercicio{{Inicio: "01/02/2023"}}); err == nil {
		t.Error("data fora do formato deveria falhar")
	}
}
