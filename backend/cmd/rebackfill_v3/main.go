// rebackfill_v3 recarrega votacoes e proposicoes com o schema v3 (itens 1, 2,
// 3, 4 e 9 da auditoria). Ver PLANO-MIGRACAO.md §6.
//
// Feito para rodar num banco de STAGING restaurado do dump de producao; as
// duas tabelas prontas sao depois trocadas em producao com pg_dump/pg_restore.
//
//  1. DDL: remove os indices unicos antigos e trunca as duas tabelas
//  2. AutoMigrate: cria colunas e indices novos (em tabela vazia, NOT NULL passa)
//  3. votos do recorte por intervalo de datas (1 chamada por mes, com retry)
//  4. proposicoes por senador em exercicio (retry por senador, autoria)
//  5. checagem de completude e queries de validacao (§7)
//  6. ranking recalculado e resumo (mediana, desvio, 100%, casos do plano)
//
// A tabela senadores NAO e sincronizada: os senador_id das tabelas novas tem
// de ser os mesmos de producao.
//
// Uso:
//
//	DATABASE_URL=postgres://...@127.0.0.1:5544/todeolho go run ./cmd/rebackfill_v3 -confirmar 127.0.0.1:5544
//	go run ./cmd/rebackfill_v3 -so-relatorio   # so as validacoes e o ranking
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"math"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/Alzarus/to-de-olho/internal/ceaps"
	"github.com/Alzarus/to-de-olho/internal/comissao"
	"github.com/Alzarus/to-de-olho/internal/proposicao"
	"github.com/Alzarus/to-de-olho/internal/ranking"
	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/internal/utils"
	"github.com/Alzarus/to-de-olho/internal/votacao"
	"github.com/Alzarus/to-de-olho/pkg/senado"
)

func main() {
	confirmar := flag.String("confirmar", "", "host:porta do banco alvo, repetido de proposito (protege contra rodar no banco errado)")
	soRelatorio := flag.Bool("so-relatorio", false, "nao recarrega; so roda as validacoes e o resumo do ranking")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	dsn := os.Getenv("DATABASE_URL")
	u, err := url.Parse(dsn)
	if dsn == "" || err != nil {
		sair("DATABASE_URL ausente ou invalido")
	}
	if !*soRelatorio && *confirmar != u.Host {
		sair(fmt.Sprintf("recusado: -confirmar %q nao bate com o host do DATABASE_URL (%s). Este comando TRUNCA votacoes e proposicoes.", *confirmar, u.Host))
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
	if err != nil {
		sair("falha ao conectar: " + err.Error())
	}
	ctx := context.Background()
	inicio := time.Now()
	fmt.Printf("banco: %s%s  recorte: desde %s\n\n", u.Host, u.Path, utils.InicioRecorte().Format("2006-01-02"))

	senadorRepo := senador.NewRepository(db)
	votacaoRepo := votacao.NewRepository(db)
	proposicaoRepo := proposicao.NewRepository(db)
	client := senado.NewLegisClient()

	if !*soRelatorio {
		// 1. DDL: sem os indices antigos o AutoMigrate cria os novos; com a
		// tabela vazia, as colunas NOT NULL novas entram sem default
		etapa("1. DDL e truncate")
		err := db.Transaction(func(tx *gorm.DB) error {
			for _, sql := range []string{
				"DROP INDEX IF EXISTS idx_votacao_unica",
				"DROP INDEX IF EXISTS idx_materia_senador",
				"TRUNCATE votacoes, proposicoes RESTART IDENTITY",
			} {
				fmt.Println("  ", sql)
				if err := tx.Exec(sql).Error; err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			sair("DDL: " + err.Error())
		}

		etapa("2. AutoMigrate")
		if err := db.AutoMigrate(&votacao.Votacao{}, &proposicao.Proposicao{}); err != nil {
			sair("AutoMigrate: " + err.Error())
		}

		etapa("3. Votos do recorte")
		t := time.Now()
		resumo, err := votacao.NewSyncService(votacaoRepo, senadorRepo, client).SyncPeriodo(ctx, utils.InicioRecorte(), time.Now())
		if err != nil {
			sair("votos: " + err.Error())
		}
		fmt.Printf("   votacoes=%d votos=%d ignorados(fora de senadores)=%d em %s\n", resumo.Votacoes, resumo.Votos, resumo.Ignorados, time.Since(t).Round(time.Second))

		etapa("4. Proposicoes")
		t = time.Now()
		if err := proposicao.NewSyncService(proposicaoRepo, senadorRepo, client).SyncFromAPI(ctx); err != nil {
			sair("proposicoes: " + err.Error())
		}
		fmt.Printf("   em %s\n", time.Since(t).Round(time.Second))
	}

	etapa("5. Validacao (PLANO-MIGRACAO.md §7)")
	ok := validar(db)

	etapa("6. Ranking")
	rk := ranking.NewService(senadorRepo, proposicaoRepo, votacaoRepo, ceaps.NewRepository(db), comissao.NewRepository(db))
	resp, err := rk.CalcularRanking(ctx, nil)
	if err != nil {
		sair("ranking: " + err.Error())
	}
	resumirRanking(db, resp)

	fmt.Printf("\nconcluido em %s\n", time.Since(inicio).Round(time.Second))
	if !ok {
		fmt.Println("VALIDACAO REPROVADA: nao levar estas tabelas para producao")
		os.Exit(2)
	}
	fmt.Println("VALIDACAO APROVADA")
}

// validar roda as queries do §7 e devolve false se alguma invariante quebrar
func validar(db *gorm.DB) bool {
	recorte := utils.InicioRecorte()
	ok := true
	checar := func(nome string, obtido, esperado int64) {
		marca := "ok"
		if obtido != esperado {
			marca, ok = "FALHOU", false
		}
		fmt.Printf("   %-62s %8d  (esperado %d) %s\n", nome, obtido, esperado, marca)
	}
	informar := func(nome string, obtido int64, referencia string) {
		fmt.Printf("   %-62s %8d  (referencia do plano: %s)\n", nome, obtido, referencia)
	}
	contar := func(sql string, args ...any) int64 {
		var n int64
		if err := db.Raw(sql, args...).Scan(&n).Error; err != nil {
			sair(sql + ": " + err.Error())
		}
		return n
	}

	fmt.Println("  item 3")
	informar("votacoes distintas no recorte", contar(`SELECT COUNT(DISTINCT codigo_votacao) FROM votacoes WHERE data >= ?`, recorte), "423 ate 03/09/2026")
	informar("votacoes da sessao 473484 (19/08/2025)", contar(`SELECT COUNT(DISTINCT codigo_votacao) FROM votacoes WHERE sessao_id = '473484'`), "24")
	informar("linhas de voto no recorte", contar(`SELECT COUNT(*) FROM votacoes WHERE data >= ?`, recorte), "~31.387")
	checar("linhas com (senador, votacao) repetido", contar(`SELECT COUNT(*) FROM (SELECT 1 FROM votacoes GROUP BY senador_id, codigo_votacao HAVING COUNT(*) > 1) x`), 0)

	fmt.Println("  item 4")
	checar("senadores em exercicio sem voto no recorte", contar(`
		SELECT COUNT(*) FROM senadores s WHERE s.em_exercicio
		AND NOT EXISTS (SELECT 1 FROM votacoes v WHERE v.senador_id = s.id AND v.data >= ?)`, recorte), 0)
	informar("votos do Marcelo Castro (742) no recorte", contar(`
		SELECT COUNT(*) FROM votacoes v JOIN senadores s ON s.id = v.senador_id
		WHERE s.codigo_parlamentar = 742 AND v.data >= ?`, recorte), "423")

	fmt.Println("  itens 1 e 9")
	informar("linhas em proposicoes", contar(`SELECT COUNT(*) FROM proposicoes`), "~57,7 mil")
	informar("materias distintas", contar(`SELECT COUNT(DISTINCT codigo_materia) FROM proposicoes`), "bem menos que as linhas")
	checar("linhas com (senador, materia) repetido", contar(`SELECT COUNT(*) FROM (SELECT 1 FROM proposicoes GROUP BY senador_id, codigo_materia HAVING COUNT(*) > 1) x`), 0)
	checar("materias com mais de um primeiro autor", contar(`SELECT COUNT(*) FROM (SELECT 1 FROM proposicoes WHERE posicao_autoria = 1 GROUP BY codigo_materia HAVING COUNT(*) > 1) x`), 0)
	informar("linhas sem posicao de autoria (senador fora de autoriaIniciativa)", contar(`SELECT COUNT(*) FROM proposicoes WHERE posicao_autoria IS NULL`), "0")
	checar("coautorias com pontuacao > 0", contar(`SELECT COUNT(*) FROM proposicoes WHERE COALESCE(posicao_autoria, 0) <> 1 AND pontuacao > 0`), 0)
	informar("proposicoes do Alan Rick (5672)", contar(`SELECT COUNT(*) FROM proposicoes p JOIN senadores s ON s.id = p.senador_id WHERE s.codigo_parlamentar = 5672`), "385")
	informar("proposicoes do Esperidiao Amin (22)", contar(`SELECT COUNT(*) FROM proposicoes p JOIN senadores s ON s.id = p.senador_id WHERE s.codigo_parlamentar = 22`), "sobe; deixa de perder ~64%")

	fmt.Println("  item 2: codigos de voto no recorte")
	var siglas []struct {
		SiglaVoto string
		Total     int64
	}
	db.Raw(`SELECT sigla_voto, COUNT(*) AS total FROM votacoes WHERE data >= ? GROUP BY 1 ORDER BY 2 DESC`, recorte).Scan(&siglas)
	desconhecidos := 0
	for _, s := range siglas {
		classe := map[votacao.Classe]string{votacao.Presente: "presente", votacao.Justificada: "justificada", votacao.Ausente: "ausente", votacao.NaoConta: "nao conta"}[votacao.Classificar(s.SiglaVoto)]
		fmt.Printf("     %-28s %7d  %s\n", s.SiglaVoto, s.Total, classe)
	}
	for _, s := range siglas {
		if s.SiglaVoto != "NA" && votacao.Classificar(s.SiglaVoto) == votacao.NaoConta {
			desconhecidos++
		}
	}
	checar("codigos de voto fora do dicionario", int64(desconhecidos), 0)
	return ok
}

// resumirRanking imprime os numeros que o plano (§7) usa como referencia
func resumirRanking(db *gorm.DB, resp *ranking.RankingResponse) {
	var bs, as []float64
	cem, cemA := 0, 0
	for _, s := range resp.Ranking {
		b := *s.Presenca
		bs = append(bs, b)
		as = append(as, s.Detalhes.TaxaPresencaBruta)
		if b >= 100 {
			cem++
		}
		if s.Detalhes.TaxaPresencaBruta >= 100 {
			cemA++
		}
	}
	fmt.Printf("   ordenados: %d   sem dados: %d\n", len(resp.Ranking), len(resp.SemDados))
	for _, s := range resp.SemDados {
		fmt.Printf("     sem dados: %s (%d registros)\n", s.Nome, s.Detalhes.TotalVotacoes)
	}
	fmt.Printf("   %-14s %8s %8s %8s\n", "presenca", "mediana", "desvio", "com 100")
	fmt.Printf("   %-14s %8.1f %8.1f %5d/%d   (plano: 89,8 / 11,0 / 2/81)\n", "A bruta", mediana(as), desvio(as), cemA, len(as))
	fmt.Printf("   %-14s %8.1f %8.1f %5d/%d   (plano: 94,3 / 9,8 / 7/81; hoje 58/81)\n", "B ajustada", mediana(bs), desvio(bs), cem, len(bs))

	fmt.Println("\n   casos do plano (registros / presentes / AP / justificadas / A / B / posicao)")
	for _, nome := range []string{"Marcelo Castro", "Giordano", "Jader Barbalho", "Irajá"} {
		for _, s := range resp.Ranking {
			if s.Nome == nome {
				d := s.Detalhes
				fmt.Printf("     %-16s %4d %4d %4d %4d %6.1f %6.1f  %2dº\n", nome, d.TotalVotacoes, d.Presentes, d.AusenciasAP, d.AusenciasJustificadas, d.TaxaPresencaBruta, *s.Presenca, s.Posicao)
			}
		}
	}
	fmt.Println("     plano: Marcelo Castro 423 402 0 14 96,6 100,0 | Giordano 423 233 166 17 55,1 57,4")
	fmt.Println("            Jader Barbalho 423 253 145 8 59,8 61,0 | Irajá 423 273 127 23 64,5 68,2")

	// correlacao entre ordem alfabetica e produtividade (hoje -0,33; deve ir a ~0)
	todos := append(append([]ranking.SenadorScore{}, resp.Ranking...), resp.SemDados...)
	sort.Slice(todos, func(i, j int) bool { return strings.ToLower(todos[i].Nome) < strings.ToLower(todos[j].Nome) })
	var x, y []float64
	for i, s := range todos {
		x = append(x, float64(i))
		y = append(y, s.Produtividade)
	}
	fmt.Printf("\n   correlacao ordem alfabetica x produtividade: %.2f   (antes: -0,33; esperado ~0)\n", pearson(x, y))

	fmt.Println("\n   top 10")
	for _, s := range resp.Ranking[:min(10, len(resp.Ranking))] {
		fmt.Printf("     %2dº %-28s score %5.1f  prod %5.1f  pres %5.1f  econ %5.1f  com %5.1f\n", s.Posicao, s.Nome, s.ScoreFinal, s.Produtividade, *s.Presenca, s.EconomiaCota, s.Comissoes)
	}
}

func mediana(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	c := append([]float64{}, v...)
	sort.Float64s(c)
	if len(c)%2 == 1 {
		return c[len(c)/2]
	}
	return (c[len(c)/2-1] + c[len(c)/2]) / 2
}

func desvio(v []float64) float64 {
	if len(v) < 2 {
		return 0
	}
	var m, s float64
	for _, x := range v {
		m += x
	}
	m /= float64(len(v))
	for _, x := range v {
		s += (x - m) * (x - m)
	}
	return math.Sqrt(s / float64(len(v)-1))
}

func pearson(x, y []float64) float64 {
	n := float64(len(x))
	var sx, sy, sxy, sxx, syy float64
	for i := range x {
		sx += x[i]
		sy += y[i]
		sxy += x[i] * y[i]
		sxx += x[i] * x[i]
		syy += y[i] * y[i]
	}
	den := math.Sqrt((n*sxx - sx*sx) * (n*syy - sy*sy))
	if den == 0 {
		return 0
	}
	return (n*sxy - sx*sy) / den
}

func etapa(nome string) { fmt.Printf("\n== %s ==\n", nome) }

func sair(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
