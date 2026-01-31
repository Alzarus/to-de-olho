package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Alzarus/to-de-olho/internal/proposicao"
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

	var props []proposicao.Proposicao
	if err := db.Find(&props).Error; err != nil {
		log.Fatalf("erro ao buscar proposicoes: %v", err)
	}

	fmt.Printf("Encontradas %d proposicoes. Iniciando recalculo...\n", len(props))
	
	count := 0
	updated := 0
	
	// Batch update para performance
	batchSize := 100
	batch := make([]proposicao.Proposicao, 0, batchSize)

	for _, p := range props {
		oldScore := p.Pontuacao
		newScore := p.CalcularPontuacao()
		
		if oldScore != newScore {
			p.Pontuacao = newScore
			batch = append(batch, p)
			count++
		}
		
		if len(batch) >= batchSize {
			saveBatch(db, batch)
			updated += len(batch)
			batch = batch[:0]
			fmt.Printf("\rProcessados: %d/%d", updated, len(props))
		}
	}
	
	// Salvar restantes
	if len(batch) > 0 {
		saveBatch(db, batch)
		updated += len(batch)
	}

	fmt.Printf("\nConcluido! %d proposicoes tiveram seus scores atualizados.\n", count)
}

func saveBatch(db *gorm.DB, batch []proposicao.Proposicao) {
	// GORM upsert simples ou update individual
	// Para garantir que atualize o campo Pontuacao, vamos fazer um loop de updates por simplicidade
	// (Upsert em massa requer cuidado com conflitos)
	for _, p := range batch {
		db.Model(&p).Update("pontuacao", p.Pontuacao)
	}
}
