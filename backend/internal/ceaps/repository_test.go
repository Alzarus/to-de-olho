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

func TestAgregadosUsamTodosOsLancamentos(t *testing.T) {
	db := testdb.Abrir(t, &DespesaCEAPS{})
	repo := NewRepository(db)
	// 30 lancamentos: mais que a pagina padrao da lista (20), que os graficos usavam
	for i := 0; i < 30; i++ {
		mes := i%3 + 1
		cnpj, nome := "11.111.111/0001-11", "Posto A"
		if i%2 == 1 {
			cnpj, nome = "22.222.222/0001-22", "Grafica B"
		}
		if err := db.Create(&DespesaCEAPS{IDOrigem: i + 1, SenadorID: 1, Ano: 2024, Mes: mes, Valor: 10, CNPJCPF: cnpj, Fornecedor: nome}).Error; err != nil {
			t.Fatal(err)
		}
	}
	// sem documento: agrupa pelo nome; outro ano e outro senador ficam de fora
	db.Create(&DespesaCEAPS{IDOrigem: 100, SenadorID: 1, Ano: 2024, Mes: 1, Valor: 5, Fornecedor: "Taxi"})
	db.Create(&DespesaCEAPS{IDOrigem: 101, SenadorID: 1, Ano: 2024, Mes: 1, Valor: 5, Fornecedor: "Onibus"})
	db.Create(&DespesaCEAPS{IDOrigem: 102, SenadorID: 1, Ano: 2025, Mes: 1, Valor: 1000, CNPJCPF: "11.111.111/0001-11", Fornecedor: "Posto A"})
	db.Create(&DespesaCEAPS{IDOrigem: 103, SenadorID: 2, Ano: 2024, Mes: 1, Valor: 1000, CNPJCPF: "11.111.111/0001-11", Fornecedor: "Posto A"})

	ano := 2024
	meses, err := repo.GastoMensal(1, &ano, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(meses) != 3 || meses[0].Mes != 1 || meses[0].Total != 110 || meses[2].Total != 100 {
		t.Errorf("meses de 2024 errados: %+v", meses)
	}
	todos, _ := repo.GastoMensal(1, nil, nil)
	if len(todos) != 4 || todos[3].Ano != 2025 {
		t.Errorf("sem ano deveria trazer os 4 meses em ordem: %+v", todos)
	}

	forn, err := repo.Fornecedores(1, &ano, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(forn) != 4 {
		t.Fatalf("esperados 4 fornecedores (2 por CNPJ, 2 sem documento), vieram %+v", forn)
	}
	if forn[0].Total != 150 || forn[0].Quantidade != 15 || forn[0].CNPJCPF == "" {
		t.Errorf("maior fornecedor errado: %+v", forn[0])
	}
}

func TestAgregadosFiltramPorCategoria(t *testing.T) {
	db := testdb.Abrir(t, &DespesaCEAPS{})
	repo := NewRepository(db)
	combustivel := "Locomoção, hospedagem, alimentação, combustíveis e lubrificantes"
	aluguel := "Aluguel de imóveis para escritório político, compreendendo despesas concernentes a eles."
	lancamentos := []DespesaCEAPS{
		{IDOrigem: 1, SenadorID: 1, Ano: 2025, Mes: 1, Valor: 100, TipoDespesa: combustivel, CNPJCPF: "11", Fornecedor: "Posto"},
		{IDOrigem: 2, SenadorID: 1, Ano: 2025, Mes: 1, Valor: 1000, TipoDespesa: aluguel, CNPJCPF: "22", Fornecedor: "Imobiliaria"},
		{IDOrigem: 3, SenadorID: 1, Ano: 2025, Mes: 2, Valor: 50, TipoDespesa: combustivel, CNPJCPF: "11", Fornecedor: "Posto"},
		{IDOrigem: 4, SenadorID: 1, Ano: 2025, Mes: 3, Valor: 7, TipoDespesa: "Outra", CNPJCPF: "33", Fornecedor: "X"},
	}
	if err := db.Create(&lancamentos).Error; err != nil {
		t.Fatal(err)
	}
	ano := 2025

	casos := []struct {
		nome         string
		tipos        []string
		meses        []float64 // total por mes, em ordem
		fornecedores int
	}{
		{"sem filtro soma tudo", nil, []float64{1100, 50, 7}, 3},
		{"uma categoria", []string{combustivel}, []float64{100, 50}, 1},
		{"categoria com virgula no nome", []string{aluguel}, []float64{1000}, 1},
		{"duas categorias", []string{combustivel, aluguel}, []float64{1100, 50}, 2},
		{"categoria inexistente", []string{"Nada"}, nil, 0},
	}
	for _, tc := range casos {
		t.Run(tc.nome, func(t *testing.T) {
			meses, err := repo.GastoMensal(1, &ano, tc.tipos)
			if err != nil {
				t.Fatal(err)
			}
			if len(meses) != len(tc.meses) {
				t.Fatalf("esperados %d meses, vieram %+v", len(tc.meses), meses)
			}
			for i, m := range meses {
				if m.Total != tc.meses[i] {
					t.Errorf("mes %d: esperado %v, obtido %v", m.Mes, tc.meses[i], m.Total)
				}
			}
			forn, err := repo.Fornecedores(1, &ano, tc.tipos)
			if err != nil {
				t.Fatal(err)
			}
			if len(forn) != tc.fornecedores {
				t.Errorf("esperados %d fornecedores, vieram %+v", tc.fornecedores, forn)
			}
		})
	}
}

func TestAgregadoPorTipoRecortaMeses(t *testing.T) {
	db := testdb.Abrir(t, &DespesaCEAPS{})
	repo := NewRepository(db)
	for i, mes := range []int{1, 2, 3, 6, 12} {
		if err := db.Create(&DespesaCEAPS{IDOrigem: i + 1, SenadorID: 1, Ano: 2025, Mes: mes, Valor: float64(mes), TipoDespesa: "A"}).Error; err != nil {
			t.Fatal(err)
		}
	}
	ano := 2025
	casos := []struct {
		nome   string
		meses  IntervaloMeses
		total  float64
		linhas int
	}{
		{"sem recorte", IntervaloMeses{}, 24, 1},
		{"fevereiro a junho", IntervaloMeses{De: 2, Ate: 6}, 11, 1},
		{"so o inicio", IntervaloMeses{De: 6}, 18, 1},
		{"so o fim", IntervaloMeses{Ate: 2}, 3, 1},
		{"intervalo vazio", IntervaloMeses{De: 7, Ate: 11}, 0, 0},
	}
	for _, tc := range casos {
		t.Run(tc.nome, func(t *testing.T) {
			ag, err := repo.AggregateByTipo(1, &ano, tc.meses)
			if err != nil {
				t.Fatal(err)
			}
			if len(ag) != tc.linhas {
				t.Fatalf("esperadas %d linhas, vieram %+v", tc.linhas, ag)
			}
			if tc.linhas > 0 && ag[0].Total != tc.total {
				t.Errorf("esperado %v, obtido %v", tc.total, ag[0].Total)
			}
		})
	}
}
