package utils

import (
	"testing"
)

func TestGetInicioLegislaturaAtual(t *testing.T) {
	anoInicio := GetInicioLegislaturaAtual()
	if anoInicio <= 2000 || anoInicio > 2100 {
		t.Errorf("GetInicioLegislaturaAtual() = %d; valor fora do intervalo esperado", anoInicio)
	}

	// Verificar se a regra de calculo modulo 4 e respeitada (ex: 2023 - 2003 % 4 == 2023 - 0 = 2023)
	if (anoInicio-3)%4 != 0 {
		t.Errorf("GetInicioLegislaturaAtual() = %d; deve alinhar com a sequencia de legislaturas (2019, 2023, 2027)", anoInicio)
	}
}
