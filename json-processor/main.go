package main

import (
	"json-processor/broker"
	"json-processor/processing"
	"json-processor/utils"
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
