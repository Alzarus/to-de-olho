package gabinete

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"

	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/pkg/senado"
)

// FonteApoio e a parte do AdmClient usada pelo sync de escritorios e
// colaboradores (permite teste sem rede)
type FonteApoio interface {
	EscritoriosApoio(ctx context.Context) ([]senado.EscritorioApoioAPI, error)
	Terceirizados(ctx context.Context) ([]senado.TerceirizadoAPI, error)
	Estagiarios(ctx context.Context) ([]senado.EstagiarioAPI, error)
}

// normalizarNome tira acentos e pontuacao, junta espacos e poe em maiusculas
// (NFKD tambem decompoe "º" em "o")
func normalizarNome(s string) string {
	var b strings.Builder
	for _, r := range norm.NFKD.String(s) {
		switch {
		case unicode.Is(unicode.Mn, r):
			// acento: some
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(unicode.ToUpper(r))
		default:
			b.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// Casador acha o senador pelo nome como a fonte escreve. Primeiro pelo nome
// parlamentar igual; senao, quando todas as palavras do nome da fonte estao no
// nome parlamentar ou no nome completo de um unico senador ("WEVERTON ROCHA":
// Weverton, nome completo Weverton Rocha Marques de Sousa).
type Casador struct {
	porNome  map[string]int
	palavras map[int]map[string]bool
}

// NovoCasador indexa os senadores
func NovoCasador(lista []senador.Senador) *Casador {
	c := &Casador{porNome: map[string]int{}, palavras: map[int]map[string]bool{}}
	for _, s := range lista {
		if n := normalizarNome(s.Nome); n != "" {
			c.porNome[n] = s.ID
		}
		p := map[string]bool{}
		for _, w := range strings.Fields(normalizarNome(s.Nome + " " + s.NomeCompleto)) {
			p[w] = true
		}
		c.palavras[s.ID] = p
	}
	return c
}

// Casar devolve o id do senador, ou false se nenhum (ou mais de um) casar
func (c *Casador) Casar(nome string) (int, bool) {
	n := normalizarNome(nome)
	if n == "" {
		return 0, false
	}
	if id, ok := c.porNome[n]; ok {
		return id, true
	}
	achado, quantos := 0, 0
	for id, p := range c.palavras {
		todas := true
		for _, w := range strings.Fields(n) {
			if !p[w] {
				todas = false
				break
			}
		}
		if todas {
			achado, quantos = id, quantos+1
		}
	}
	return achado, quantos == 1
}

// reGabinete extrai o nome de "GABINETE DO SENADOR X", "GABINETE DA SENADORA X"
// ou da forma abreviada dos nomes longos, "GAB SEN X" (o ponto ja saiu na
// normalizacao): "GAB SEN ASTRONAUTA MARCOS PONTES"
var reGabinete = regexp.MustCompile(`^(?:GABINETE D[OA] SENADORA?|GAB SEN) (.+)$`)

// donoDoGabinete devolve o nome do senador numa lotacao de gabinete ("" se a
// lotacao nao e gabinete de senador, como "GABINETE DA DIRETORIA GERAL")
func donoDoGabinete(lotacao string) string {
	m := reGabinete.FindStringSubmatch(normalizarNome(lotacao))
	if m == nil {
		return ""
	}
	return m[1]
}

// ResultadoApoio resume a carga de escritorios e colaboradores
type ResultadoApoio struct {
	Escritorios   int      `json:"escritorios"`
	Terceirizados int      `json:"terceirizados"`
	Estagiarios   int      `json:"estagiarios"`
	NaoCasados    []string `json:"nao_casados"` // nomes da fonte sem senador em exercicio
}

// SyncApoio grava o retrato atual dos escritorios de apoio e a contagem de
// terceirizados e estagiarios por gabinete. Cada fonte e independente: a que
// falhar ou vier vazia nao apaga o que ja estava gravado.
func (s *SyncService) SyncApoio(ctx context.Context) (ResultadoApoio, error) {
	var res ResultadoApoio
	if s.apoio == nil {
		return res, errors.New("fonte de escritorios e colaboradores nao configurada")
	}
	lista, err := s.senadores.FindAll(false)
	if err != nil {
		return res, fmt.Errorf("listar senadores em exercicio: %w", err)
	}
	casador := NovoCasador(lista)
	naoCasados := map[string]bool{}
	agora := s.agora()
	var erros []error

	if escs, err := s.apoio.EscritoriosApoio(ctx); err != nil {
		erros = append(erros, fmt.Errorf("escritorios de apoio: %w", err))
	} else if len(escs) > 0 {
		var linhas []Escritorio
		for _, e := range escs {
			id, ok := casador.Casar(e.Parlamentar.Nome)
			if !ok {
				naoCasados[e.Parlamentar.Nome] = true
				continue
			}
			linhas = append(linhas, Escritorio{
				SenadorID: id,
				Nome:      strings.TrimSpace(e.Setor.Nome),
				Endereco:  strings.TrimSuffix(strings.TrimSpace(e.Setor.Endereco), "."),
				Telefone:  strings.TrimSpace(e.Setor.Telefone),
			})
		}
		if err := s.repo.SubstituirEscritorios(linhas, agora); err != nil {
			return res, fmt.Errorf("gravar escritorios: %w", err)
		}
		res.Escritorios = len(linhas)
	}

	contagem := map[int]map[string]int{}
	contar := func(lotacao, tipo string) {
		dono := donoDoGabinete(lotacao)
		if dono == "" {
			return
		}
		id, ok := casador.Casar(dono)
		if !ok {
			naoCasados[dono] = true
			return
		}
		if contagem[id] == nil {
			contagem[id] = map[string]int{}
		}
		contagem[id][tipo]++
	}
	var tipos []string
	if terc, err := s.apoio.Terceirizados(ctx); err != nil {
		erros = append(erros, fmt.Errorf("terceirizados: %w", err))
	} else if len(terc) > 0 {
		tipos = append(tipos, ColaboradorTerceirizado)
		for _, t := range terc {
			if strings.EqualFold(strings.TrimSpace(t.Situacao), "ativo") {
				contar(t.Lotacao.Nome, ColaboradorTerceirizado)
				res.Terceirizados++
			}
		}
	}
	if est, err := s.apoio.Estagiarios(ctx); err != nil {
		erros = append(erros, fmt.Errorf("estagiarios: %w", err))
	} else if len(est) > 0 {
		tipos = append(tipos, ColaboradorEstagiario)
		for _, e := range est {
			contar(e.NomeOrgao, ColaboradorEstagiario)
			res.Estagiarios++
		}
	}
	if len(tipos) > 0 {
		var linhas []Colaborador
		for id, porTipo := range contagem {
			for tipo, q := range porTipo {
				linhas = append(linhas, Colaborador{SenadorID: id, Tipo: tipo, Quantidade: q})
			}
		}
		if err := s.repo.SubstituirColaboradores(tipos, linhas, agora); err != nil {
			return res, fmt.Errorf("gravar colaboradores: %w", err)
		}
	}

	for n := range naoCasados {
		res.NaoCasados = append(res.NaoCasados, n)
	}
	sort.Strings(res.NaoCasados)
	slog.Info("sync de escritorios e colaboradores concluido", "escritorios", res.Escritorios,
		"terceirizados", res.Terceirizados, "estagiarios", res.Estagiarios, "nao_casados", len(res.NaoCasados))
	if len(res.NaoCasados) > 0 {
		// senadores fora de exercicio (licenciados, suplentes que sairam) caem aqui
		slog.Info("nomes da fonte sem senador em exercicio", "nomes", strings.Join(res.NaoCasados, "; "))
	}
	return res, errors.Join(erros...)
}
