package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Alzarus/to-de-olho/internal/emenda"
	"github.com/Alzarus/to-de-olho/internal/senador"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	apiKey := os.Getenv("TRANSPARENCIA_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("CHAVE_API_DADOS")
	}
	if apiKey == "" {
		log.Fatal("Defina TRANSPARENCIA_API_KEY ou CHAVE_API_DADOS no ambiente")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=todeolho port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Falha ao conectar ao banco:", err)
	}

	senadorRepo := senador.NewRepository(db)
	emendaRepo := emenda.NewRepository(db)
	syncService := emenda.NewSyncService(emendaRepo, senadorRepo, apiKey)

	anos := parseAnos(os.Getenv("EMENDAS_ANOS"))
	if len(anos) == 0 {
		anoAtual := time.Now().Year()
		anos = []int{anoAtual}
	}

	ctx := context.Background()
	for _, ano := range anos {
		log.Printf("Iniciando sync de emendas %d...", ano)
		if err := syncService.SyncAll(ctx, ano); err != nil {
			log.Printf("Erro no sync %d: %v", ano, err)
		}
	}

	log.Println("Sync concluído!")
}

func parseAnos(valor string) []int {
	valor = strings.TrimSpace(valor)
	if valor == "" {
		return nil
	}

	partes := strings.Split(valor, ",")
	anos := make([]int, 0, len(partes))
	for _, parte := range partes {
		anoStr := strings.TrimSpace(parte)
		if anoStr == "" {
			continue
		}
		ano, err := strconv.Atoi(anoStr)
		if err != nil {
			continue
		}
		anos = append(anos, ano)
	}
	return anos
}
