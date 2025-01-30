package main

import (
	"log"
	"to-de-olho-api/configs"
	"to-de-olho-api/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	configs.ConnectDatabase()

	routes.SetupRoutes(r)

	if err := r.Run(":3000"); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %s", err)
	}
}
