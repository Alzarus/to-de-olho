package senado

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	BaseURLAdm = "https://adm.senado.gov.br/adm-dadosabertos"
)

// AdmClient consome a API Administrativa do Senado
type AdmClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAdmClient cria um novo client
func NewAdmClient() *AdmClient {
	return &AdmClient{
		baseURL: BaseURLAdm,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // Timeout maior para downloads
		},
	}
}

// === TIPOS DE RESPOSTA DA API ===

// DespesaCEAPSAPI representa uma despesa retornada pela API
type DespesaCEAPSAPI struct {
	Ano            int     `json:"ano"`
	Mes            int     `json:"mes"`
	CodSenador     int     `json:"codSenador"`
	NomeSenador    string  `json:"nomeSenador"`
	TipoDespesa    string  `json:"tipoDespesa"`
	Fornecedor     string  `json:"fornecedor"`
	CNPJCPF        string  `json:"cpfCnpj"`
	Documento      string  `json:"documento"`
	Data           string  `json:"data"` // formato: "YYYY-MM-DD" ou "DD/MM/YYYY"
	ValorReembolso float64 `json:"valorReembolsado"`
}

// ListaDespesasResponse representa a resposta da API de despesas
type ListaDespesasResponse struct {
	Despesas []DespesaCEAPSAPI `json:"despesas"`
}

// === METODOS DO CLIENT ===

// ListarDespesasCEAPS busca despesas de um ano especifico
func (c *AdmClient) ListarDespesasCEAPS(ctx context.Context, ano int) ([]DespesaCEAPSAPI, error) {
	url := fmt.Sprintf("%s/api/v1/senadores/despesas_ceaps/%d", c.baseURL, ano)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro criando request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro na requisicao: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status inesperado: %d", resp.StatusCode)
	}

	var result ListaDespesasResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// Tentar decodificar como array direto (formato alternativo)
		resp.Body.Close()

		// Refazer request
		resp2, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("erro na requisicao retry: %w", err)
		}
		defer resp2.Body.Close()

		var despesas []DespesaCEAPSAPI
		if err := json.NewDecoder(resp2.Body).Decode(&despesas); err != nil {
			return nil, fmt.Errorf("erro decodificando JSON: %w", err)
		}
		return despesas, nil
	}

	return result.Despesas, nil
}
