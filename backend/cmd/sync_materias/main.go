// sync_materias carrega a tabela materias (nome popular e descricao breve) sem
// esperar o sync diario: grava a curadoria, completa codigo_materia das
// votacoes que o banco consegue derivar e busca /processo/{id} das pendentes.
//
// Uso: DATABASE_URL=... go run ./cmd/sync_materias [-limite 0] [-completar]
//
//	-limite     materias por rodada (0: todas)
//	-completar  recarrega as votacoes do recorte se faltar codigo_materia
//	            (uma chamada por mes; precisa da tabela senadores populada)
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/Alzarus/to-de-olho/internal/materia"
	"github.com/Alzarus/to-de-olho/internal/proposicao"
	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/internal/votacao"
	senadoapi "github.com/Alzarus/to-de-olho/pkg/senado"
)

func main() {
	limite := flag.Int("limite", 0, "materias por rodada (0: todas)")
	completar := flag.Bool("completar", false, "recarregar votacoes do recorte sem codigo_materia")
	flag.Parse()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL nao definido")
		os.Exit(2)
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
	if err != nil {
		fmt.Fprintln(os.Stderr, "falha ao conectar:", err)
		os.Exit(1)
	}
	if err := run(db, *limite, *completar); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(db *gorm.DB, limite int, completar bool) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := db.AutoMigrate(&votacao.Votacao{}, &proposicao.Proposicao{}, &materia.Materia{}, &materia.ApelidoCurado{}); err != nil {
		return fmt.Errorf("auto-migrate: %w", err)
	}
	if err := materia.SemearApelidosCurados(db); err != nil {
		return err
	}
	if err := votacao.PreencherCodigoMateria(db); err != nil {
		return err
	}
	client := senadoapi.NewLegisClient()
	if completar {
		sync := votacao.NewSyncService(votacao.NewRepository(db), senador.NewRepository(db), client)
		if err := sync.CompletarCodigoMateria(ctx); err != nil {
			return err
		}
	}
	resumo, err := materia.NewSyncService(db, client).Sync(ctx, limite)
	if err != nil {
		return err
	}
	slog.Info("fim", "resumo", fmt.Sprintf("%+v", resumo))
	return nil
}
