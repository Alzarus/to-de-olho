package main

import (
	"fmt"
	"log"
	
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Emenda struct {
	ID         uint
	Localidade string
    ValorPago float64
}

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=todeolho port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	var localidades []string
	// Get top 20 distinct localidade
	db.Model(&Emenda{}).Distinct("localidade").Limit(20).Find(&localidades)
	
	fmt.Println("Distinct localidade samples:")
	for _, l := range localidades {
		fmt.Printf("'%s'\n", l)
	}

    var emendas []Emenda
    db.Limit(5).Find(&emendas)
    fmt.Println("\nEmenda samples:")
    for _, e := range emendas {
        fmt.Printf("Loc: '%s' Val: %f\n", e.Localidade, e.ValorPago)
    }
}
