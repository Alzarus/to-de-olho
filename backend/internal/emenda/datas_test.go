package emenda_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Alzarus/to-de-olho/internal/emenda"
	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/internal/testdb"
	"github.com/Alzarus/to-de-olho/pkg/transparencia"
)

func doc(data, fase string) transparencia.DocumentoEmendaDTO {
	return transparencia.DocumentoEmendaDTO{Data: data, Fase: fase}
}

func dia(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func mesmoDia(t *testing.T, nome string, got *time.Time, want string) {
	t.Helper()
	if want == "" {
		if got != nil {
			t.Errorf("%s = %v, queria nulo", nome, got)
		}
		return
	}
	if got == nil || !got.Equal(dia(want)) {
		t.Errorf("%s = %v, queria %s", nome, got, want)
	}
}

func TestResumirDatas(t *testing.T) {
	d := emenda.ResumirDatas([]transparencia.DocumentoEmendaDTO{
		doc("19/06/2024", "Empenho"),
		doc("16/04/2024", "Empenho"),
		doc("17/05/2024", "Liquidação"),
		doc("13/12/2024", "Pagamento"),
		doc("17/05/2024", "Pagamento"),
		doc("data ruim", "Pagamento"),
		doc("01/01/2024", "Outra fase"),
	})
	mesmoDia(t, "PrimeiroEmpenho", d.PrimeiroEmpenho, "2024-04-16")
	mesmoDia(t, "PrimeiroPagamento", d.PrimeiroPagamento, "2024-05-17")
	mesmoDia(t, "UltimoPagamento", d.UltimoPagamento, "2024-12-13")
}

func TestResumirDatas_SemPagamento(t *testing.T) {
	d := emenda.ResumirDatas([]transparencia.DocumentoEmendaDTO{doc("02/09/2025", "Empenho")})
	mesmoDia(t, "PrimeiroEmpenho", d.PrimeiroEmpenho, "2025-09-02")
	mesmoDia(t, "PrimeiroPagamento", d.PrimeiroPagamento, "")
	mesmoDia(t, "UltimoPagamento", d.UltimoPagamento, "")
}

// clienteFalso devolve as paginas de documentos por codigo (15 por pagina na
// fonte; aqui o tamanho nao importa)
type clienteFalso struct {
	paginas  map[string][][]transparencia.DocumentoEmendaDTO
	falha    map[string]bool
	chamadas map[string]int
}

func (c *clienteFalso) GetDocumentosEmenda(_ context.Context, codigo string, pagina int) ([]transparencia.DocumentoEmendaDTO, error) {
	c.chamadas[codigo]++
	if c.falha[codigo] {
		return nil, errors.New("API returned server error: 500")
	}
	ps := c.paginas[codigo]
	if pagina > len(ps) {
		return nil, nil
	}
	return ps[pagina-1], nil
}

func TestSyncDatas(t *testing.T) {
	db := testdb.Abrir(t, &senador.Senador{}, &emenda.Emenda{})
	repo := emenda.NewRepository(db)
	if err := db.Create(&[]senador.Senador{{ID: 1, CodigoParlamentar: 1, Nome: "S1"}, {ID: 2, CodigoParlamentar: 2, Nome: "S2"}}).Error; err != nil {
		t.Fatal(err)
	}
	linhas := []emenda.Emenda{
		{SenadorID: 1, Ano: 2024, Numero: "A", ValorEmpenhado: 100, ValorPago: 100},
		{SenadorID: 2, Ano: 2024, Numero: "A", ValorEmpenhado: 100, ValorPago: 100}, // mesmo codigo, outro senador
		{SenadorID: 1, Ano: 2025, Numero: "B", ValorEmpenhado: 50, ValorPago: 0},
		{SenadorID: 1, Ano: 2025, Numero: "C", ValorEmpenhado: 10, ValorPago: 0},  // fonte vazia
		{SenadorID: 1, Ano: 2023, Numero: "D", ValorEmpenhado: 10, ValorPago: 10}, // fonte falha
	}
	if err := db.Create(&linhas).Error; err != nil {
		t.Fatal(err)
	}

	cli := &clienteFalso{
		paginas: map[string][][]transparencia.DocumentoEmendaDTO{
			"A": {{doc("16/04/2024", "Empenho"), doc("17/05/2024", "Pagamento")}, {doc("13/12/2024", "Pagamento")}},
			"B": {{doc("02/09/2025", "Empenho")}},
		},
		falha:    map[string]bool{"D": true},
		chamadas: map[string]int{},
	}
	svc := emenda.NewSyncDatas(repo, cli, 0)

	resumo, err := svc.SyncDatas(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if resumo != (emenda.ResumoDatas{Pendentes: 4, Gravadas: 2, SemDados: 1, Falhas: 1}) {
		t.Errorf("resumo = %+v", resumo)
	}
	if cli.chamadas["A"] != 3 { // 2 paginas + a vazia, uma vez para as duas linhas
		t.Errorf("chamadas A = %d, queria 3", cli.chamadas["A"])
	}

	var a []emenda.Emenda
	db.Where("numero = ?", "A").Find(&a)
	for _, e := range a {
		mesmoDia(t, "A.PrimeiroEmpenho", e.DataPrimeiroEmpenho, "2024-04-16")
		mesmoDia(t, "A.UltimoPagamento", e.DataUltimoPagamento, "2024-12-13")
	}

	// Segunda rodada: so C e D seguem pendentes
	pend, err := repo.DatasPendentes(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(pend) != 2 {
		t.Errorf("pendentes = %+v, queria C e D", pend)
	}

	// Pagamento novo em B: valor pago muda, B volta a ser consultada
	db.Model(&emenda.Emenda{}).Where("numero = ?", "B").Update("valor_pago", 50)
	pend, _ = repo.DatasPendentes(0)
	var numeros []string
	for _, p := range pend {
		numeros = append(numeros, p.Numero)
	}
	// 2025 antes de 2023; dentro do ano, codigo decrescente
	if got := strings.Join(numeros, ","); got != "C,B,D" {
		t.Errorf("pendentes depois do pagamento = %s, queria C,B,D", got)
	}
}
