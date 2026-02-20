package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/Alzarus/to-de-olho/internal/ceaps"
	"github.com/Alzarus/to-de-olho/internal/comissao"
	"github.com/Alzarus/to-de-olho/internal/proposicao"
	"github.com/Alzarus/to-de-olho/internal/ranking"
	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/internal/votacao"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Carregar envs
	_ = godotenv.Load(".env")

	// Conectar ao Banco
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=todeolho port=5432 sslmode=disable"
		slog.Info("DATABASE_URL vazia, usando fallback local (igual main.go)", "dsn", dsn)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		slog.Error("falha ao conectar ao banco", "error", err)
		os.Exit(1)
	}

	// Repositorios
	senadorRepo := senador.NewRepository(db)
	proposicaoRepo := proposicao.NewRepository(db)
	votacaoRepo := votacao.NewRepository(db)
	ceapsRepo := ceaps.NewRepository(db)
	comissaoRepo := comissao.NewRepository(db)

	// Servico de Ranking
	rankingService := ranking.NewService(
		senadorRepo,
		proposicaoRepo,
		votacaoRepo,
		ceapsRepo,
		comissaoRepo,
	)

	ctx := context.Background()

	slog.Info(">>> FORCANDO CALCULO DE RANKING (AGORA) <<<")
	
	// Calcular Geral
	_, err = rankingService.CalcularRanking(ctx, nil)
	if err != nil {
		slog.Error("Erro ao calcular ranking geral", "error", err)
	} else {
		slog.Info("Ranking Geral calculado com sucesso!")
	}

	// Calcular por Ano (2023, 2024, 2025)
	anos := []int{2023, 2024, 2025}
	for _, ano := range anos {
		anoVal := ano
		_, err = rankingService.CalcularRanking(ctx, &anoVal)
		if err != nil {
			slog.Error("Erro ao calcular ranking anual", "ano", ano, "error", err)
		} else {
			slog.Info("Ranking Anual calculado", "ano", ano)
		}
	}

	slog.Info(">>> CONCLUIDO! PODE GRAVAR O VIDEO <<<")
}
