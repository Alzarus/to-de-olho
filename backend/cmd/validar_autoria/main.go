// validar_autoria confere, numa amostra de materias, se a posicao de autoria
// extraida do texto da listagem (/processo?codigoParlamentarAutor=) bate com a
// ordem oficial de /processo/{id} (autoriaIniciativa[].ordem).
//
// Criterio do PLANO-MIGRACAO.md §2.1: >= 300 materias e 100% de concordancia no
// primeiro autor. Abaixo disso, o sync deve usar so o detalhe.
//
// Uso: go run ./cmd/validar_autoria [-senadores 20] [-pls 15] [-seed 1]
// So faz GET na API do Senado; nao toca em banco.
package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Alzarus/to-de-olho/internal/proposicao"
	"github.com/Alzarus/to-de-olho/pkg/retry"
	senadoapi "github.com/Alzarus/to-de-olho/pkg/senado"
)

type amostra struct {
	senador string
	materia senadoapi.MateriaAPI
}

type resultado struct {
	amostra
	detalhe *senadoapi.ProcessoDetalhe
	err     error
}

func main() {
	nSenadores := flag.Int("senadores", 20, "senadores sorteados")
	nPLs := flag.Int("pls", 15, "PLs sorteados por senador, alem de todas as PECs")
	seed := flag.Int64("seed", 1, "semente do sorteio")
	flag.Parse()

	ctx := context.Background()
	client := senadoapi.NewLegisClient()
	rng := rand.New(rand.NewSource(*seed))

	parlamentares, err := client.ListarSenadoresAtuais(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "falha ao listar senadores:", err)
		os.Exit(1)
	}
	rng.Shuffle(len(parlamentares), func(i, j int) { parlamentares[i], parlamentares[j] = parlamentares[j], parlamentares[i] })
	if *nSenadores < len(parlamentares) {
		parlamentares = parlamentares[:*nSenadores]
	}

	// 1. Amostra: todas as PECs + nPLs PLs sorteados de cada senador
	var itens []amostra
	for _, p := range parlamentares {
		codigo, _ := strconv.Atoi(p.IdentificacaoParlamentar.CodigoParlamentar)
		nome := p.IdentificacaoParlamentar.NomeParlamentar
		var lista []senadoapi.MateriaAPI
		err := retry.WithRetry(ctx, 3, "lista "+nome, func() error {
			var err error
			lista, err = client.ListarProposicoesParlamentar(ctx, codigo)
			return err
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "falha na lista de", nome, err)
			os.Exit(1)
		}
		var pecs, pls []senadoapi.MateriaAPI
		for _, m := range lista {
			switch {
			case strings.HasPrefix(m.Identificacao, "PEC "):
				pecs = append(pecs, m)
			case strings.HasPrefix(m.Identificacao, "PL ") || strings.HasPrefix(m.Identificacao, "PLS "):
				pls = append(pls, m)
			}
		}
		rng.Shuffle(len(pls), func(i, j int) { pls[i], pls[j] = pls[j], pls[i] })
		if len(pls) > *nPLs {
			pls = pls[:*nPLs]
		}
		for _, m := range append(pecs, pls...) {
			itens = append(itens, amostra{senador: nome, materia: m})
		}
		fmt.Printf("%-28s lista=%4d  PECs=%3d  PLs sorteados=%2d\n", nome, len(lista), len(pecs), len(pls))
	}

	// 2. Detalhe de cada materia (4 em paralelo, com retry)
	resultados := make([]resultado, len(itens))
	var wg sync.WaitGroup
	fila := make(chan int)
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range fila {
				r := resultado{amostra: itens[i]}
				r.err = retry.WithRetry(ctx, 3, "processo", func() error {
					var err error
					r.detalhe, err = client.ObterProcesso(ctx, itens[i].materia.ID)
					return err
				})
				resultados[i] = r
			}
		}()
	}
	inicio := time.Now()
	for i := range itens {
		fila <- i
	}
	close(fila)
	wg.Wait()

	// 3. Comparacao, para cada autor parlamentar do detalhe cujo nome o texto resolve
	var (
		materias, materiasRecorte, errosDetalhe      int
		checagens, divergencias                      int
		primeiro, primeiroDiverge                    int
		senadorAmostrado, senadorTextoOK, senadorDiv int
		primeiroSenador, primeiroSenadorDiv          int
		exemplos                                     []string
	)
	recorte := "2023-02-01"
	for _, r := range resultados {
		if r.err != nil {
			errosDetalhe++
			continue
		}
		materias++
		if r.materia.DataApresentacao >= recorte {
			materiasRecorte++
		}
		for _, a := range r.detalhe.AutoriaIniciativa {
			if a.CodigoParlamentar == nil {
				continue
			}
			pos, _, ok := proposicao.PosicaoNoTexto(r.materia.Autoria, a.Autor)
			if !ok {
				continue
			}
			checagens++
			if pos != a.Ordem {
				divergencias++
			}
			if a.Ordem == 1 || pos == 1 {
				primeiro++
				if pos != a.Ordem {
					primeiroDiverge++
					exemplos = append(exemplos, fmt.Sprintf("  %s (processo %d): texto diz %d, detalhe diz %d para %s | %.120s",
						r.materia.Identificacao, r.materia.ID, pos, a.Ordem, a.Autor, r.materia.Autoria))
				}
			}
		}

		// A decisao que o sync toma de fato: posicao do senador dono da lista
		senadorAmostrado++
		pos, _, ok := proposicao.PosicaoNoTexto(r.materia.Autoria, r.senador)
		if !ok {
			continue // o sync cai no detalhe: nao ha o que comparar
		}
		senadorTextoOK++
		oficial := 0
		for _, a := range r.detalhe.AutoriaIniciativa {
			if a.CodigoParlamentar != nil && proposicao.MesmoNome(a.Autor, r.senador) {
				oficial = a.Ordem
			}
		}
		if pos != oficial {
			senadorDiv++
			exemplos = append(exemplos, fmt.Sprintf("  [senador da lista] %s (processo %d): %s texto=%d detalhe=%d",
				r.materia.Identificacao, r.materia.ID, r.senador, pos, oficial))
		}
		if pos == 1 || oficial == 1 {
			primeiroSenador++
			if pos != oficial {
				primeiroSenadorDiv++
			}
		}
	}

	sort.Strings(exemplos)
	fmt.Printf("\nDetalhes buscados em %s\n", time.Since(inicio).Round(time.Second))
	fmt.Printf("Materias na amostra: %d (%d apresentadas desde %s); detalhe indisponivel: %d\n", materias, materiasRecorte, recorte, errosDetalhe)
	fmt.Printf("\nA. Todos os autores parlamentares resolvidos pelo texto\n")
	fmt.Printf("   checagens: %d  divergencias: %d\n", checagens, divergencias)
	fmt.Printf("   envolvendo o primeiro autor: %d  divergencias: %d\n", primeiro, primeiroDiverge)
	fmt.Printf("\nB. Senador dono da lista (a decisao do sync)\n")
	fmt.Printf("   materias: %d  resolvidas pelo texto: %d (%.1f%%)  caem no detalhe: %d\n",
		senadorAmostrado, senadorTextoOK, 100*float64(senadorTextoOK)/float64(max(senadorAmostrado, 1)), senadorAmostrado-senadorTextoOK)
	fmt.Printf("   divergencias de posicao: %d  no primeiro autor: %d de %d\n", senadorDiv, primeiroSenadorDiv, primeiroSenador)

	if len(exemplos) > 0 {
		fmt.Println("\nDivergencias:")
		for i, e := range exemplos {
			if i == 30 {
				fmt.Printf("  ... e mais %d\n", len(exemplos)-30)
				break
			}
			fmt.Println(e)
		}
	}

	aprovado := materias >= 300 && primeiroDiverge == 0 && primeiroSenadorDiv == 0 && errosDetalhe == 0
	fmt.Printf("\nCRITERIO (>= 300 materias, 100%% no primeiro autor): %v\n", map[bool]string{true: "APROVADO", false: "REPROVADO"}[aprovado])
	if !aprovado {
		os.Exit(2)
	}
}
