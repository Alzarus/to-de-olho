package transparencia

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Alzarus/to-de-olho/pkg/retry"
)

const BaseURL = "https://api.portaldatransparencia.gov.br/api-de-dados"

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // Timeout maior
		},
	}
}

type EmendaDTO struct {
	CodigoEmenda      string `json:"codigoEmenda"`
	Ano               int    `json:"ano"`
	TipoEmenda        string `json:"tipoEmenda"`
	Autor             string `json:"autor"`
	NomeAutor         string `json:"nomeAutor"`
	NumeroEmenda      string `json:"numeroEmenda"`
	LocalidadeDoGasto string `json:"localidadeDoGasto"`
	Funcao            string `json:"funcao"`
	Subfuncao         string `json:"subfuncao"`
	ValorEmpenhado    string `json:"valorEmpenhado"`
	ValorLiquidado    string `json:"valorLiquidado"`
	ValorPago         string `json:"valorPago"`
	ValorRestoPago    string `json:"valorRestoPago"`
}

func ParseMoney(s string) float64 {
	if s == "" {
		return 0
	}
	s = strings.ReplaceAll(s, "R$", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", ".")
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return val
}

func (c *Client) GetEmendas(ano int, nomeAutor string, pagina int) ([]EmendaDTO, error) {
	return c.GetEmendasWithCtx(context.Background(), ano, nomeAutor, pagina)
}

func (c *Client) GetEmendasWithCtx(ctx context.Context, ano int, nomeAutor string, pagina int) ([]EmendaDTO, error) {
	params := url.Values{}
	params.Add("ano", fmt.Sprintf("%d", ano))
	params.Add("pagina", fmt.Sprintf("%d", pagina))
	if nomeAutor != "" {
		params.Add("nomeAutor", nomeAutor)
	}

	reqURL := fmt.Sprintf("%s/emendas?%s", BaseURL, params.Encode())

	var emendas []EmendaDTO

	err := retry.WithRetry(ctx, 3, fmt.Sprintf("GetEmendas(ano=%d, autor=%s, pag=%d)", ano, nomeAutor, pagina), func() error {
		req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
		if err != nil {
			return err
		}

		req.Header.Set("chave-api-dados", c.apiKey)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) ToDeOlho/1.0")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return err // Erro de rede, retry
		}
		defer resp.Body.Close()

		// Nao fazer retry em erros 4xx (exceto 429)
		if resp.StatusCode == http.StatusTooManyRequests {
			return fmt.Errorf("API rate limit (429)")
		}
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return fmt.Errorf("API returned client error: %d (nao recuperavel)", resp.StatusCode)
		}
		if resp.StatusCode >= 500 {
			return fmt.Errorf("API returned server error: %d", resp.StatusCode)
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("API returned status: %d", resp.StatusCode)
		}

		// Verificar se body esta vazio
		if resp.ContentLength == 0 {
			emendas = []EmendaDTO{}
			return nil
		}

		if err := json.NewDecoder(resp.Body).Decode(&emendas); err != nil {
			return fmt.Errorf("erro decode: %w", err)
		}

		return nil
	})

	return emendas, err
}

