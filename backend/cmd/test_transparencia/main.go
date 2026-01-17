package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	apiKey := os.Getenv("TRANSPARENCIA_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("CHAVE_API_DADOS")
	}
	if apiKey == "" {
		log.Fatal("Defina TRANSPARENCIA_API_KEY ou CHAVE_API_DADOS no ambiente")
	}

	url := "https://api.portaldatransparencia.gov.br/api-de-dados/emendas?ano=2024&pagina=1"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("chave-api-dados", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}
