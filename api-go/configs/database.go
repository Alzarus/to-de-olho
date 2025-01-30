package configs

import (
	"fmt"
	"log"
	"os"
	"to-de-olho-api/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_DATABASE")

	dsn := fmt.Sprintf("host=%s user=%s dbname=%s password=%s port=%s sslmode=disable",
		dbHost, dbUser, dbName, dbPassword, dbPort)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Falha ao conectar ao banco de dados:", err)
	}

	err = database.AutoMigrate(
		&models.Frequency{},
		&models.Contract{},
		&models.Councilor{},
		&models.Proposition{},
		&models.GeneralProductivity{},
		&models.PropositionProductivity{},
		&models.TravelExpense{},
		&models.ExecutionStatus{},
	)
	if err != nil {
		log.Fatal("Falha ao rodar as migrações:", err)
	}

	DB = database
}
