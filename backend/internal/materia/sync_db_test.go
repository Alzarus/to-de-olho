package materia_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/Alzarus/to-de-olho/internal/materia"
	"github.com/Alzarus/to-de-olho/internal/proposicao"
	"github.com/Alzarus/to-de-olho/internal/testdb"
	"github.com/Alzarus/to-de-olho/internal/votacao"
	senadoapi "github.com/Alzarus/to-de-olho/pkg/senado"
)

func abrir(t *testing.T) *gorm.DB {
	t.Helper()
	return testdb.Abrir(t, &votacao.Votacao{}, &proposicao.Proposicao{}, &materia.Materia{}, &materia.ApelidoCurado{})
}

type clienteFalso struct {
	mu        sync.Mutex
	processos map[int]*senadoapi.ProcessoDetalhe // por id do processo
	ids       map[int]int                        // codigoMateria -> id
	chamadas  map[int]int
}

func (c *clienteFalso) ObterProcesso(_ context.Context, id int) (*senadoapi.ProcessoDetalhe, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.chamadas[id]++
	p, ok := c.processos[id]
	if !ok {
		return nil, fmt.Errorf("%w: processo %d", senadoapi.ErrNaoEncontrado, id)
	}
	return p, nil
}

func (c *clienteFalso) ObterIDProcesso(_ context.Context, codigo int) (int, error) {
	return c.ids[codigo], nil
}

func processo(id, codigo int, ident, apelido, explicacao string, temas ...string) *senadoapi.ProcessoDetalhe {
	d := &senadoapi.ProcessoDetalhe{ID: id, CodigoMateria: codigo, Identificacao: ident, Apelido: apelido}
	d.Conteudo.Ementa = "Ementa de " + ident
	d.Conteudo.ExplicacaoEmenta = explicacao
	for _, t := range temas {
		d.Classificacoes = append(d.Classificacoes, senadoapi.ClassificacaoProcesso{Descricao: t})
	}
	d.Documento.URL = "https://legis.senado.gov.br/sdleg-getter/documento?dm=1"
	return d
}

func ip(v int) *int { return &v }

func votoEm(codigoVotacao int, sigla, ident string, codigo, id *int) votacao.Votacao {
	return votacao.Votacao{
		SenadorID: 1, CodigoVotacao: codigoVotacao, SessaoID: "1", CodigoSessao: "1",
		Data: time.Date(2025, 5, 1, 12, 0, 0, 0, time.UTC), SiglaVoto: "Sim", Voto: "Sim",
		Materia: ident, SiglaMateria: sigla, CodigoMateria: codigo, IdProcesso: id,
	}
}

func TestSyncMaterias(t *testing.T) {
	db := abrir(t)
	votos := []votacao.Votacao{
		votoEm(1, "PL", "PL 2338/2023", ip(157233), ip(8441243)),
		votoEm(2, "PLP", "PLP 68/2024", ip(164914), nil), // linha antiga sem id: resolve pelo codigo
		votoEm(3, "MSF", "MSF 32/2024", ip(165001), ip(8705702)),
		votoEm(4, "PEC", "PEC 1/2020", ip(999), ip(123)), // 404
	}
	if err := db.Create(&votos).Error; err != nil {
		t.Fatal(err)
	}
	apresentada := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	props := []proposicao.Proposicao{
		{SenadorID: 1, CodigoMateria: "170000", SiglaSubtipoMateria: "PEC", DescricaoIdentificacao: "PEC 9/2024", DataApresentacao: &apresentada, IdProcesso: ip(9000000)},
		{SenadorID: 1, CodigoMateria: "170001", SiglaSubtipoMateria: "REQ", DescricaoIdentificacao: "REQ 1/2024", DataApresentacao: &apresentada, IdProcesso: ip(9000001)},
	}
	if err := db.Create(&props).Error; err != nil {
		t.Fatal(err)
	}

	cliente := &clienteFalso{
		chamadas: map[int]int{},
		ids:      map[int]int{164914: 8701679},
		processos: map[int]*senadoapi.ProcessoDetalhe{
			8441243: processo(8441243, 157233, "PL 2338/2023", "Marco Legal da Inteligência Artificial ", "", "Ciência, Tecnologia e Informática"),
			8701679: processo(8701679, 164914, "PLP 68/2024", "", "", "Tributos", "Tributos"),
			9000000: processo(9000000, 170000, "PEC 9/2024", "", "Explica a PEC.", "Educação"),
		},
	}
	s := materia.NewSyncService(db, cliente).ComPausa(0)

	pend, err := s.Pendentes(0)
	if err != nil {
		t.Fatal(err)
	}
	// votacoes primeiro (sem MSF), depois proposicoes (sem REQ)
	var codigos []int
	for _, p := range pend {
		codigos = append(codigos, p.CodigoMateria)
	}
	// mesma data: codigo decrescente
	if fmt.Sprint(codigos) != "[164914 157233 999 170000]" {
		t.Fatalf("pendentes inesperados: %v", codigos)
	}

	resumo, err := s.Sync(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	esperado := materia.Resumo{Pendentes: 4, Gravadas: 3, ComApelido: 1, ComExplicacao: 1, NaoEncontradas: 1}
	if resumo != esperado {
		t.Errorf("resumo %+v; esperado %+v", resumo, esperado)
	}
	if cliente.chamadas[123] != 1 {
		t.Errorf("404 repetido %d vezes; esperado 1", cliente.chamadas[123])
	}

	var ia materia.Materia
	if err := db.First(&ia, "codigo_materia = ?", 157233).Error; err != nil {
		t.Fatal(err)
	}
	if ia.Apelido == nil || *ia.Apelido != "Marco Legal da Inteligência Artificial" || len(ia.Temas) != 1 {
		t.Errorf("materia gravada errada: %+v", ia)
	}
	var plp materia.Materia
	if err := db.First(&plp, "codigo_materia = ?", 164914).Error; err != nil {
		t.Fatal(err)
	}
	if plp.IdProcesso != 8701679 || plp.Apelido != nil || len(plp.Temas) != 1 {
		t.Errorf("PLP 68/2024: id pelo codigo, sem apelido, temas sem repeticao: %+v", plp)
	}

	// segunda rodada: so o 404 continua pendente
	pend, err = s.Pendentes(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(pend) != 1 || pend[0].CodigoMateria != 999 {
		t.Errorf("segunda rodada: esperado so o 999 pendente, obtido %+v", pend)
	}
}

func TestSemearApelidosCuradosIdempotente(t *testing.T) {
	db := abrir(t)
	// entrada antiga que saiu da curadoria deve sumir
	if err := db.Create(&materia.ApelidoCurado{CodigoMateria: 1, Identificacao: "PL 1/2000", Apelido: "Velho", FonteURL: "https://www12.senado.leg.br/x"}).Error; err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := materia.SemearApelidosCurados(db); err != nil {
			t.Fatal(err)
		}
	}
	var n int64
	db.Model(&materia.ApelidoCurado{}).Count(&n)
	if int(n) != len(materia.ApelidosCurados()) {
		t.Errorf("curadoria com %d linhas; esperado %d", n, len(materia.ApelidosCurados()))
	}
	var velho int64
	db.Model(&materia.ApelidoCurado{}).Where("codigo_materia = 1").Count(&velho)
	if velho != 0 {
		t.Error("entrada fora da curadoria nao foi removida")
	}
	var reforma materia.ApelidoCurado
	if err := db.First(&reforma, "codigo_materia = ?", 158930).Error; err != nil || reforma.Apelido != "Reforma tributária" {
		t.Errorf("PEC 45/2019: %+v %v", reforma, err)
	}
}

// TestSyncMateriasAPIReal busca ~20 materias votadas de verdade na API do
// Senado. So roda com TEST_SENADO_API=1 (e TEST_DATABASE_URL).
func TestSyncMateriasAPIReal(t *testing.T) {
	if os.Getenv("TEST_SENADO_API") != "1" {
		t.Skip("TEST_SENADO_API != 1")
	}
	db := abrir(t)
	client := senadoapi.NewLegisClient()
	ctx := context.Background()

	fim := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	var lista []senadoapi.VotacaoSessaoAPI
	for _, ini := range []time.Time{time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC), time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)} {
		f := ini.AddDate(0, 1, -1)
		if ini.Year() == 2024 && ini.Month() == 12 {
			f = fim
		}
		v, err := client.ListarVotacoesPeriodo(ctx, ini, f)
		if err != nil {
			t.Fatal(err)
		}
		lista = append(lista, v...)
	}
	vistos := map[int]bool{}
	var votos []votacao.Votacao
	for _, v := range lista {
		if v.CodigoMateria == nil || vistos[*v.CodigoMateria] || v.Sigla == "MSF" || v.Sigla == "OFS" || len(vistos) >= 20 {
			continue
		}
		vistos[*v.CodigoMateria] = true
		votos = append(votos, votoEm(v.CodigoSessaoVotacao, v.Sigla, v.Identificacao, v.CodigoMateria, v.IdProcesso))
	}
	if err := db.Create(&votos).Error; err != nil {
		t.Fatal(err)
	}
	if err := materia.SemearApelidosCurados(db); err != nil {
		t.Fatal(err)
	}
	inicio := time.Now()
	resumo, err := materia.NewSyncService(db, client).Sync(ctx, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("materias: %d votadas; resumo %+v; %.1fs", len(votos), resumo, time.Since(inicio).Seconds())
	if resumo.Gravadas == 0 || resumo.Falhas > 0 {
		t.Errorf("resumo inesperado: %+v", resumo)
	}
	var linhas []struct {
		Identificacao, Fonte, Apelido string
		Explicacao                    bool
		Temas                         string
	}
	db.Raw(`SELECT m.identificacao, COALESCE(CASE WHEN m.apelido IS NOT NULL THEN 'oficial' WHEN c.apelido IS NOT NULL THEN 'curadoria' END, '-') AS fonte,
		COALESCE(m.apelido, c.apelido, '') AS apelido, m.explicacao_ementa IS NOT NULL AS explicacao, COALESCE(m.temas::text, '') AS temas
		FROM materias m LEFT JOIN materias_apelidos_curados c USING (codigo_materia) ORDER BY m.identificacao`).Scan(&linhas)
	for _, l := range linhas {
		t.Logf("%-32s %-9s %-45s explicacao=%-5v %s", l.Identificacao, l.Fonte, l.Apelido, l.Explicacao, l.Temas)
	}
}
