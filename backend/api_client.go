package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	BaseURLCamara = "https://dadosabertos.camara.leg.br/api/v2"
	UserAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

// Cliente HTTP customizado com timeout
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

// fetchDeputadosFromAPI busca deputados da API oficial da Câmara
func fetchDeputadosFromAPI(partido, uf, nome string) ([]Deputado, error) {
	url := fmt.Sprintf("%s/deputados", BaseURLCamara)

	// Construir query parameters
	params := "?ordem=ASC&ordenarPor=nome"
	if partido != "" {
		params += "&siglaPartido=" + partido
	}
	if uf != "" {
		params += "&siglaUf=" + uf
	}
	if nome != "" {
		params += "&nome=" + nome
	}

	fullURL := url + params

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro na requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API retornou status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %w", err)
	}

	var apiResp APIResponseDeputados
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON: %w", err)
	}

	return apiResp.Dados, nil
}

// fetchDeputadoByIDFromAPI busca um deputado específico por ID
func fetchDeputadoByIDFromAPI(id string) (*Deputado, error) {
	url := fmt.Sprintf("%s/deputados/%s", BaseURLCamara, id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro na requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API retornou status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %w", err)
	}

	var apiResp struct {
		Dados Deputado `json:"dados"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON: %w", err)
	}

	return &apiResp.Dados, nil
}

// fetchDespesasFromAPI busca despesas de um deputado
func fetchDespesasFromAPI(deputadoID, ano string) ([]Despesa, error) {
	url := fmt.Sprintf("%s/deputados/%s/despesas", BaseURLCamara, deputadoID)
	params := fmt.Sprintf("?ano=%s&ordem=DESC&ordenarPor=dataDocumento", ano)

	fullURL := url + params

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro na requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API retornou status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %w", err)
	}

	var apiResp APIResponseDespesas
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON: %w", err)
	}

	return apiResp.Dados, nil
}
