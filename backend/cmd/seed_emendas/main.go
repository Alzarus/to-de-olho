package main

import (
	"log"
	"os"

	"github.com/Alzarus/to-de-olho/internal/emenda"
	"github.com/Alzarus/to-de-olho/internal/senador"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=todeolho port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Falha ao conectar ao banco:", err)
	}

	caminhoCSV := os.Getenv("EMENDAS_CSV_PATH")
	if len(os.Args) > 1 {
		caminhoCSV = os.Args[1]
	}

	if caminhoCSV == "" {
		log.Fatal("Informe o caminho do CSV via EMENDAS_CSV_PATH ou argumento CLI")
	}

	senadorRepo := senador.NewRepository(db)
	emendaRepo := emenda.NewRepository(db)
	service := emenda.NewService(emendaRepo, senadorRepo)

	log.Println("Iniciando importacao CSV de emendas...")
	if err := service.ImportarCSV(caminhoCSV); err != nil {
		log.Fatal("Falha ao importar emendas:", err)
	}
	log.Println("Importacao de emendas concluida!")
}
