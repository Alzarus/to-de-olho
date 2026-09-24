package scheduler

import (
	"testing"
	"time"
)

func TestProximaExecucao(t *testing.T) {
	brt := time.FixedZone("BRT", -3*60*60)
	casos := []struct {
		nome  string
		agora time.Time
		quer  time.Time
	}{
		{"antes do horario", time.Date(2026, 9, 24, 2, 46, 0, 0, time.UTC), time.Date(2026, 9, 24, 6, 0, 0, 0, time.UTC)},
		{"exatamente no horario", time.Date(2026, 9, 24, 6, 0, 0, 0, time.UTC), time.Date(2026, 9, 25, 6, 0, 0, 0, time.UTC)},
		{"depois do horario", time.Date(2026, 9, 24, 18, 0, 0, 0, time.UTC), time.Date(2026, 9, 25, 6, 0, 0, 0, time.UTC)},
		{"virada de mes", time.Date(2026, 9, 30, 23, 0, 0, 0, time.UTC), time.Date(2026, 10, 1, 6, 0, 0, 0, time.UTC)},
		{"fuso de Brasilia", time.Date(2026, 9, 24, 1, 0, 0, 0, brt), time.Date(2026, 9, 24, 6, 0, 0, 0, time.UTC)},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if got := proximaExecucao(c.agora); !got.Equal(c.quer) {
				t.Errorf("proximaExecucao(%v) = %v, quer %v", c.agora, got, c.quer)
			}
		})
	}
}
