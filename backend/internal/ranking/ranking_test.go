package ranking

import (
	"testing"
)

// TestArredondar verifica se a funcao auxiliar de arredondamento funciona como esperado
func TestArredondar(t *testing.T) {
	testes := []struct {
		entrada  float64
		esperado float64
	}{
		{10.555, 10.56},
		{10.554, 10.55},
		{99.999, 100.00},
		{0.0, 0.0},
		{33.3333, 33.33},
	}

	for _, teste := range testes {
		resultado := arredondar(teste.entrada)
		if resultado != teste.esperado {
			t.Errorf("arredondar(%f) = %f; esperado %f", teste.entrada, resultado, teste.esperado)
		}
	}
}

// TestCalculoPesosSimples verifica se a logica basica dos pesos esta correta
// Este é um "white-box test" simplificado para validar a metodologia
func TestCalculoPesosSimples(t *testing.T) {
	// Ponderacao oficial
	// (Produtividade * 0.35) + (Presenca * 0.25) + (Economia * 0.20) + (Comissoes * 0.20)
	
	produtividade := 100.0
	presenca := 100.0
	economia := 100.0
	comissoes := 100.0

	scoreFinal := (produtividade * 0.35) + 
				  (presenca * 0.25) + 
				  (economia * 0.20) + 
				  (comissoes * 0.20)

	if scoreFinal != 100.0 {
		t.Errorf("Score maximo esperado 100.0, obtido %f", scoreFinal)
	}

	// Cenaria onde so tem presenca
	produtividade = 0
	presenca = 100
	economia = 0
	comissoes = 0

	scoreFinal = (produtividade * 0.35) + 
				  (presenca * 0.25) + 
				  (economia * 0.20) + 
				  (comissoes * 0.20)
	
	if scoreFinal != 25.0 {
		t.Errorf("Score apenas com presenca esperado 25.0, obtido %f", scoreFinal)
	}
}
