package ceaps

import (
	"testing"
	"time"

	"github.com/Alzarus/to-de-olho/internal/testdb"
)

func TestGetTotalPeriodoPorCompetencia(t *testing.T) {
	db := testdb.Abrir(t, &DespesaCEAPS{})
	repo := NewRepository(db)
	for i, c := range []struct {
		ano, mes int
		valor    float64
	}{
		{2023, 1, 1000}, // legislatura anterior
		{2023, 2, 200},
		{2024, 12, 30},
		{2026, 9, 4},
	} {
		dt := time.Date(c.ano, time.Month(c.mes), 10, 0, 0, 0, 0, time.UTC)
		if err := db.Create(&DespesaCEAPS{IDOrigem: i + 1, SenadorID: 1, Ano: c.ano, Mes: c.mes, Valor: c.valor, CNPJCPF: string(rune('a' + i)), DataEmissao: &dt}).Error; err != nil {
			t.Fatal(err)
		}
	}
	mandato, _ := repo.GetTotalPeriodo(1, time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 9, 23, 0, 0, 0, 0, time.UTC))
	if mandato != 234 {
		t.Errorf("mandato deveria somar fev/2023 em diante (234), somou %v", mandato)
	}
	ano, _ := repo.GetTotalPeriodo(1, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	if ano != 30 {
		t.Errorf("2024 deveria somar 30, somou %v", ano)
	}
}

func TestSubstituirAnoGuardaLancamentosIguais(t *testing.T) {
	repo := NewRepository(testdb.Abrir(t, &DespesaCEAPS{}))
	dt := time.Date(2025, 1, 29, 0, 0, 0, 0, time.UTC)
	// mesmo fornecedor, dia e valor: a chave antiga guardava um so
	iguais := []DespesaCEAPS{
		{IDOrigem: 1, SenadorID: 1, Ano: 2025, Mes: 1, CNPJCPF: "51.407.456/0001-01", DataEmissao: &dt, Valor: 6000},
		{IDOrigem: 2, SenadorID: 1, Ano: 2025, Mes: 1, CNPJCPF: "51.407.456/0001-01", DataEmissao: &dt, Valor: 6000},
	}
	if err := repo.SubstituirAno(2025, iguais); err != nil {
		t.Fatal(err)
	}
	// nova rodada com uma exclusao na origem: o ano vira o retrato da API
	if err := repo.SubstituirAno(2025, iguais[:1]); err != nil {
		t.Fatal(err)
	}
	total, _ := repo.GetTotalPeriodo(1, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if total != 6000 {
		t.Errorf("esperado 6000 apos a substituicao, obtido %v", total)
	}
	if err := repo.SubstituirAno(2025, iguais); err != nil {
		t.Fatal(err)
	}
	total, _ = repo.GetTotalPeriodo(1, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if total != 12000 {
		t.Errorf("lancamentos iguais com ids distintos: esperado 12000, obtido %v", total)
	}
}
