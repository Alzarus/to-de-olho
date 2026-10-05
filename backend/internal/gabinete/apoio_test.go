package gabinete

import (
	"context"
	"errors"
	"testing"

	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/internal/testdb"
	"github.com/Alzarus/to-de-olho/pkg/senado"
)

func TestNormalizarNome(t *testing.T) {
	casos := map[string]string{
		"Hamilton Mourão":       "HAMILTON MOURAO",
		"  sérgio   petecão ":   "SERGIO PETECAO",
		"PLÍNIO VALÉRIO":        "PLINIO VALERIO",
		"Escritório nº 1 - AC.": "ESCRITORIO NO 1 AC",
	}
	for entrada, esperado := range casos {
		if got := normalizarNome(entrada); got != esperado {
			t.Errorf("normalizarNome(%q) = %q, queria %q", entrada, got, esperado)
		}
	}
}

// Nomes reais da fonte (05/10/2026) contra os nomes do banco
var senadoresCasador = []senador.Senador{
	{ID: 1, Nome: "Alan Rick", NomeCompleto: "Alan Rick Miranda"},
	{ID: 2, Nome: "Weverton", NomeCompleto: "Weverton Rocha Marques de Sousa"},
	{ID: 3, Nome: "Nelsinho Trad", NomeCompleto: "Nelson Trad Filho"},
	{ID: 4, Nome: "Hamilton Mourão", NomeCompleto: "Antônio Hamilton Martins Mourão"},
	{ID: 5, Nome: "Sérgio Petecão", NomeCompleto: "Sérgio de Oliveira Cunha"},
	{ID: 6, Nome: "Jaques Wagner", NomeCompleto: "Jaques Wagner"},
	{ID: 7, Nome: "Otto Alencar", NomeCompleto: "Otto Roberto Mendonça de Alencar"},
}

func TestCasador(t *testing.T) {
	c := NovoCasador(senadoresCasador)
	casos := []struct {
		nome string
		id   int
		ok   bool
	}{
		{"ALAN RICK", 1, true},
		{"WEVERTON ROCHA", 2, true},      // palavras contidas no nome completo
		{"NELSINHO TRAD FILHO", 3, true}, // nome parlamentar + nome completo
		{"HAMILTON MOURAO", 4, true},     // sem acento
		{"RODRIGO PACHECO", 0, false},    // fora de exercicio: nao esta no banco
		{"RODRIGO CUNHA", 0, false},      // "CUNHA" casa com Petecao, "RODRIGO" nao
		{"", 0, false},
	}
	for _, cs := range casos {
		id, ok := c.Casar(cs.nome)
		if ok != cs.ok || id != cs.id {
			t.Errorf("Casar(%q) = (%d, %v), queria (%d, %v)", cs.nome, id, ok, cs.id, cs.ok)
		}
	}
}

func TestCasador_Ambiguo(t *testing.T) {
	c := NovoCasador([]senador.Senador{
		{ID: 1, Nome: "Fulano Silva", NomeCompleto: "Fulano Souza Silva"},
		{ID: 2, Nome: "Beltrano Silva", NomeCompleto: "Beltrano Souza Silva"},
	})
	if id, ok := c.Casar("SOUZA SILVA"); ok {
		t.Errorf("dois senadores casam: nao deveria escolher, veio %d", id)
	}
}

func TestDonoDoGabinete(t *testing.T) {
	casos := map[string]string{
		"GABINETE DO SENADOR CLEITINHO":      "CLEITINHO",
		"GABINETE DA SENADORA ZENAIDE MAIA":  "ZENAIDE MAIA",
		"Gabinete do Senador Plínio Valério": "PLINIO VALERIO",
		"GAB SEN ASTRONAUTA MARCOS PONTES":   "ASTRONAUTA MARCOS PONTES",
		"GAB. SEN. VENEZIANO VITAL RÊGO":     "VENEZIANO VITAL REGO",
		"GABINETE DA DIRETORIA GERAL":        "",
		"GABINETE ADMINISTRATIVO DA SINFRA":  "",
		"SERVIÇO DE EDIÇÃO":                  "",
	}
	for entrada, esperado := range casos {
		if got := donoDoGabinete(entrada); got != esperado {
			t.Errorf("donoDoGabinete(%q) = %q, queria %q", entrada, got, esperado)
		}
	}
}

// fonteApoioFake devolve listas fixas; erro simula a fonte fora do ar
type fonteApoioFake struct {
	escritorios   []senado.EscritorioApoioAPI
	terceirizados []senado.TerceirizadoAPI
	estagiarios   []senado.EstagiarioAPI
	falhaEsc      bool
}

func (f fonteApoioFake) EscritoriosApoio(context.Context) ([]senado.EscritorioApoioAPI, error) {
	if f.falhaEsc {
		return nil, errors.New("falha simulada")
	}
	return f.escritorios, nil
}
func (f fonteApoioFake) Terceirizados(context.Context) ([]senado.TerceirizadoAPI, error) {
	return f.terceirizados, nil
}
func (f fonteApoioFake) Estagiarios(context.Context) ([]senado.EstagiarioAPI, error) {
	return f.estagiarios, nil
}

func escritorio(nome, setor, endereco string) senado.EscritorioApoioAPI {
	var e senado.EscritorioApoioAPI
	e.Parlamentar.Nome = nome
	e.Setor.Nome, e.Setor.Endereco = setor, endereco
	return e
}

func terceirizado(lotacao, situacao string) senado.TerceirizadoAPI {
	var t senado.TerceirizadoAPI
	t.Situacao, t.Lotacao.Nome = situacao, lotacao
	return t
}

func TestSyncApoio(t *testing.T) {
	repo := NewRepository(testdb.Abrir(t, Modelos()...))
	sens := senadoresFake(senadoresCasador)
	fonte := fonteApoioFake{
		escritorios: []senado.EscritorioApoioAPI{
			escritorio("ALAN RICK", "Escritório de Apoio nº 1 do Senador Alan Rick", "RIO BRANCO."),
			escritorio("WEVERTON ROCHA", "Escritório de Apoio do Senador Weverton", "AV. DOS HOLANDESES, SAO LUIS, MA"),
			escritorio("RODRIGO PACHECO", "Escritório de Apoio do Senador Rodrigo Pacheco", "BELO HORIZONTE, MG"),
		},
		terceirizados: []senado.TerceirizadoAPI{
			terceirizado("GABINETE DO SENADOR ALAN RICK", "Ativo"),
			terceirizado("GABINETE DO SENADOR ALAN RICK", "Ativo"),
			terceirizado("GABINETE DO SENADOR ALAN RICK", "Inativo"),
			terceirizado("GABINETE DO SENADOR HAMILTON MOURÃO", "Ativo"),
			terceirizado("SERVIÇO DE CONTROLE DE EQUIPAMENTOS", "Ativo"),
		},
		estagiarios: []senado.EstagiarioAPI{{NomeOrgao: "GABINETE DO SENADOR ALAN RICK"}, {NomeOrgao: "SERVIÇO DE EDIÇÃO"}},
	}
	svc := NewSyncService(repo, sens, nil, nil)
	svc.apoio = fonte

	res, err := svc.SyncApoio(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.Escritorios != 2 || len(res.NaoCasados) != 1 || res.NaoCasados[0] != "RODRIGO PACHECO" {
		t.Errorf("resultado = %+v", res)
	}

	escs, _ := repo.Escritorios(1)
	if len(escs) != 1 || escs[0].Endereco != "RIO BRANCO" {
		t.Errorf("escritorios de Alan Rick = %+v (o ponto final da fonte sai)", escs)
	}
	cols, _ := repo.Colaboradores(1)
	got := map[string]int{}
	for _, c := range cols {
		got[c.Tipo] = c.Quantidade
	}
	if got[ColaboradorTerceirizado] != 2 || got[ColaboradorEstagiario] != 1 {
		t.Errorf("colaboradores de Alan Rick = %v, queria 2 terceirizados ativos e 1 estagiario", got)
	}
	if cols, _ := repo.Colaboradores(4); len(cols) != 1 || cols[0].Quantidade != 1 {
		t.Errorf("colaboradores de Mourao = %+v", cols)
	}

	// Escritorios fora do ar: os enderecos gravados ficam; a contagem atualiza
	fonte.falhaEsc = true
	fonte.terceirizados = fonte.terceirizados[3:4] // so o de Mourao
	svc.apoio = fonte
	if _, err := svc.SyncApoio(context.Background()); err == nil {
		t.Error("falha de uma fonte deveria voltar como erro")
	}
	if escs, _ := repo.Escritorios(1); len(escs) != 1 {
		t.Errorf("escritorios apagados com a fonte fora do ar: %+v", escs)
	}
	cols, _ = repo.Colaboradores(1)
	for _, c := range cols {
		if c.Tipo == ColaboradorTerceirizado {
			t.Errorf("Alan Rick saiu da lista de terceirizados e deveria ficar sem linha: %+v", c)
		}
	}
}
