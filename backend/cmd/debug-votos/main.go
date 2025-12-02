package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"

	app "to-de-olho-backend/internal/application"
	"to-de-olho-backend/internal/config"
	"to-de-olho-backend/internal/infrastructure/cache"
	"to-de-olho-backend/internal/infrastructure/db"
	"to-de-olho-backend/internal/infrastructure/httpclient"
	"to-de-olho-backend/internal/infrastructure/repository"
)

func main() {
	godotenv.Load()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Erro ao carregar config: %v", err)
	}

	// Usar timeout maior para operações longas
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Hour)
	defer cancel()

	// Conectar ao banco
	pgPool, err := db.NewPostgresPoolFromConfig(ctx, &cfg.Database)
	if err != nil {
		log.Fatalf("Erro ao conectar ao Postgres: %v", err)
	}
	defer pgPool.Close()

	// Criar cliente da Câmara
	camaraClient := httpclient.NewCamaraClientFromConfig(&cfg.CamaraClient)

	// Criar repositórios
	votacaoRepo := repository.NewVotacaoRepository(pgPool)
	cacheClient := cache.NewFromConfig(&cfg.Redis)

	// Criar VotacoesService
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	votacoesService := app.NewVotacoesService(votacaoRepo, camaraClient, cacheClient)

	fmt.Println("=== BACKFILL COMPLETO DE VOTAÇÕES E VOTOS (2022-2025) ===")
	fmt.Println()

	// Verificar estado atual
	var votosIniciais int
	pgPool.QueryRow(ctx, "SELECT COUNT(*) FROM votos_deputados WHERE id_deputado > 0").Scan(&votosIniciais)
	fmt.Printf("Votos válidos antes do backfill: %d\n\n", votosIniciais)

	totalVotacoes := 0
	errosTotal := 0

	// Processar cada ano (do mais recente para o mais antigo para ter dados atuais primeiro)
	anos := []int{2025, 2024, 2023, 2022}

	for _, ano := range anos {
		fmt.Printf("\n=== Processando ano %d ===\n", ano)

		// Processar mês a mês para melhor controle
		for mes := time.January; mes <= time.December; mes++ {
			// Calcular último dia do mês
			primeiroDia := time.Date(ano, mes, 1, 0, 0, 0, 0, time.UTC)
			ultimoDia := primeiroDia.AddDate(0, 1, -1)
			ultimoDia = time.Date(ultimoDia.Year(), ultimoDia.Month(), ultimoDia.Day(), 23, 59, 59, 0, time.UTC)

			// Pular meses futuros
			if primeiroDia.After(time.Now()) {
				continue
			}

			fmt.Printf("  %s %d: ", mes.String()[:3], ano)

			// Criar contexto com timeout por mês (15 minutos max)
			mesCtx, mesCancel := context.WithTimeout(ctx, 15*time.Minute)

			processadas, err := votacoesService.SincronizarVotacoes(mesCtx, primeiroDia, ultimoDia)
			mesCancel()

			if err != nil {
				fmt.Printf("ERRO: %v\n", err)
				errosTotal++
				// Esperar mais tempo após erro para o circuit breaker resetar
				fmt.Println("    Aguardando 30s para circuit breaker resetar...")
				time.Sleep(30 * time.Second)
				continue
			}

			totalVotacoes += processadas
			fmt.Printf("%d votações\n", processadas)

			// Delay entre meses para não sobrecarregar a API
			time.Sleep(3 * time.Second)
		}
	}

	// Verificar no banco
	fmt.Println("\n=== RESUMO FINAL ===")

	var votacoesCount int
	err = pgPool.QueryRow(ctx, "SELECT COUNT(*) FROM votacoes").Scan(&votacoesCount)
	if err != nil {
		log.Printf("Erro ao contar votações: %v", err)
	}
	fmt.Printf("Total de votações no banco: %d\n", votacoesCount)

	var votosCount int
	err = pgPool.QueryRow(ctx, "SELECT COUNT(*) FROM votos_deputados WHERE id_deputado > 0").Scan(&votosCount)
	if err != nil {
		log.Printf("Erro ao contar votos: %v", err)
	}
	fmt.Printf("Total de votos válidos: %d\n", votosCount)

	// Estatísticas por ano
	fmt.Println("\nVotos por ano:")
	rows, err := pgPool.Query(ctx, `
		SELECT EXTRACT(YEAR FROM vot.data_votacao) as ano, COUNT(v.id) as votos
		FROM votos_deputados v
		JOIN votacoes vot ON v.id_votacao = vot.id
		WHERE v.id_deputado > 0
		GROUP BY EXTRACT(YEAR FROM vot.data_votacao)
		ORDER BY ano
	`)
	if err != nil {
		log.Printf("Erro: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var ano float64
			var votos int
			rows.Scan(&ano, &votos)
			fmt.Printf("  %d: %d votos\n", int(ano), votos)
		}
	}

	fmt.Printf("\n✅ Backfill concluído! Votações processadas: %d, Erros: %d\n", totalVotacoes, errosTotal)

	_ = logger
}
