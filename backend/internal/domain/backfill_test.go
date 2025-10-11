package domain

import (
	"testing"
	"time"
)

func TestBackfillConfig_SetDefaults(t *testing.T) {
	c := &BackfillConfig{}
	c.SetDefaults()

	if c.AnoFim == 0 {
		t.Fatalf("AnoFim deve ser preenchido pelo SetDefaults")
	}
	if c.AnoInicio == 0 {
		t.Fatalf("AnoInicio deve ser preenchido pelo SetDefaults")
	}
	if c.BatchSize == 0 {
		t.Fatalf("BatchSize deve ter valor padrão após SetDefaults")
	}
	if c.DelayBetweenBatches == 0 {
		t.Fatalf("DelayBetweenBatches deve ter valor padrão após SetDefaults")
	}

	// Validar que AnoFim >= AnoInicio
	if c.AnoFim < c.AnoInicio {
		t.Fatalf("AnoFim (%d) não pode ser menor que AnoInicio (%d)", c.AnoFim, c.AnoInicio)
	}

	// Testar que valores razoáveis permanecem
	c2 := &BackfillConfig{AnoInicio: 2019, AnoFim: 2020, BatchSize: 50}
	c2.SetDefaults()
	if c2.BatchSize != 50 {
		t.Fatalf("BatchSize não deve ser sobrescrito se já definido")
	}

	// Validar validade
	if err := c2.Validate(); err != nil {
		t.Fatalf("Config válida falhou na validação: %v", err)
	}

	// Teste limite de BatchSize inválido
	c3 := &BackfillConfig{BatchSize: 5}
	if err := c3.Validate(); err == nil {
		t.Fatalf("BatchSize muito pequeno deve retornar erro na validação")
	}

	_ = time.Now()
}
