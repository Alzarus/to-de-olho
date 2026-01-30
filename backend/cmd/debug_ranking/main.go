package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Alzarus/to-de-olho/internal/ceaps"
	"github.com/Alzarus/to-de-olho/internal/comissao"
	"github.com/Alzarus/to-de-olho/internal/proposicao"
	"github.com/Alzarus/to-de-olho/internal/ranking"
	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/internal/votacao"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=todeolho port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	senadorRepo := senador.NewRepository(db)
	proposicaoRepo := proposicao.NewRepository(db)
	votacaoRepo := votacao.NewRepository(db)
	ceapsRepo := ceaps.NewRepository(db)
	comissaoRepo := comissao.NewRepository(db)

	rankingService := ranking.NewService(
		senadorRepo,
		proposicaoRepo,
		votacaoRepo,
		ceapsRepo,
		comissaoRepo,
		nil, // no redis
	)

	ctx := context.Background()
	start := time.Now()
	fmt.Println("Calculando ranking...")
	
	resp, err := rankingService.CalcularRanking(ctx, nil)
	if err != nil {
		log.Fatalf("erro ao calcular ranking: %v", err)
	}

	fmt.Printf("Ranking calculado em %v. Total: %d\n", time.Since(start), resp.Total)
	fmt.Println("Top 10 Senadores (Score Final | Produtividade | Pontos Raw | Total Props)")
	fmt.Println("-------------------------------------------------------------------------")

	for i := 0; i < 10 && i < len(resp.Ranking); i++ {
		s := resp.Ranking[i]
		fmt.Printf("#%d %-20s | Final: %.2f | Prod: %.2f | RawPts: %.2f | TotalProps: %d\n",
			s.Posicao,
			truncate(s.Nome, 20),
			s.ScoreFinal,
			s.Produtividade,
			s.Detalhes.PontuacaoProposicoes,
			s.Detalhes.TotalProposicoes,
		)
	}
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n-3] + "..."
	}
	return s
}
