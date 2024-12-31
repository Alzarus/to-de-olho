package main

import (
	"json-processor-go/broker"
	"json-processor-go/processing"
	"json-processor-go/utils"
	"os"
)

func main() {
	utils.InitializeLogger()

	brokerURL := os.Getenv("BROKER_URL")
	if brokerURL == "" {
		utils.Log.Fatal("BROKER_URL is not set")
	}

	go broker.ListenForMessages(brokerURL, processing.ProcessJsonFiles)
	select {}
}
