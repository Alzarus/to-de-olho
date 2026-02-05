package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
    _ = godotenv.Load() // Ignore error in debug script

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=todeolho port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	var lastSenadorUpdate time.Time
	var lastVotacaoCreate time.Time
	var lastProposicaoUpdate time.Time

	// Check Senadores
	db.Raw("SELECT MAX(updated_at) FROM senadores").Scan(&lastSenadorUpdate)
	
	// Check Votacoes
	db.Raw("SELECT MAX(created_at) FROM votacoes").Scan(&lastVotacaoCreate)
	
	// Check Proposicoes
	db.Raw("SELECT MAX(updated_at) FROM proposicoes").Scan(&lastProposicaoUpdate)

	fmt.Printf("=== DB Timestamps ===\n")
	fmt.Printf("Last Senador Update:   %v\n", lastSenadorUpdate)
	fmt.Printf("Last Votacao Create:   %v\n", lastVotacaoCreate)
	fmt.Printf("Last Proposicao Update: %v\n", lastProposicaoUpdate)
	fmt.Printf("Current Time:          %v\n", time.Now())

    // Check API Key Env Var (safely, just present or not)
    apiKey := os.Getenv("TRANSPARENCIA_API_KEY")
    if apiKey == "" {
        fmt.Printf("\n[WARNING] TRANSPARENCIA_API_KEY is NOT set in this environment.\n")
    } else {
        fmt.Printf("\n[OK] TRANSPARENCIA_API_KEY is set (len=%d).\n", len(apiKey))
    }
}
