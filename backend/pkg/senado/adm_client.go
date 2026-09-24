package senado

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Alzarus/to-de-olho/pkg/retry"
)

const (
	BaseURLAdm = "https://adm.senado.gov.br/adm-dadosabertos"
)

// timeoutRecursos e o limite de cada chamada a recursos-utilizados: a rota
// agrega varios sistemas do Senado e ja levou dezenas de segundos por senador.
const timeoutRecursos = 120 * time.Second

// AdmClient consome a API Administrativa do Senado
type AdmClient struct {
	baseURL    string
	httpClient *http.Client
	// recursosClient tem timeout proprio (timeoutRecursos)
	recursosClient *http.Client
	// tentativas por chamada de RecursosUtilizados (pkg/retry)
	tentativas int
}

// NewAdmClient cria um novo client
func NewAdmClient() *AdmClient {
	return &AdmClient{
		baseURL: BaseURLAdm,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // Timeout maior para downloads
		},
		recursosClient: &http.Client{Timeout: timeoutRecursos},
		tentativas:     3,
	}
}

// NewAdmClientURL cria um client apontando para outra base (testes com httptest)
func NewAdmClientURL(baseURL string, tentativas int) *AdmClient {
	c := NewAdmClient()
	c.baseURL = baseURL
	if tentativas > 0 {
		c.tentativas = tentativas
	}
	return c
}

// === TIPOS DE RESPOSTA DA API ===

// DespesaCEAPSAPI representa uma despesa retornada pela API
type DespesaCEAPSAPI struct {
	ID             int     `json:"id"` // unico por lancamento
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

// === RECURSOS UTILIZADOS (gabinete, beneficios, cota) ===

// ErrSenadorNaoEncontrado e devolvido quando a API responde 404 para o codigo
var ErrSenadorNaoEncontrado = errors.New("senador nao encontrado na API administrativa")

// ValorRecursoAPI e uma linha de despesa agregada (cota ou gastos nao inclusos)
type ValorRecursoAPI struct {
	Recurso string  `json:"recurso"`
	Valor   float64 `json:"valor"`
}

// GrupoRecursosAPI agrupa despesas com o total
type GrupoRecursosAPI struct {
	Despesas   []ValorRecursoAPI `json:"despesas"`
	TotalValor float64           `json:"totalValor"`
}

// BeneficioAPI indica o uso de auxilio-moradia ou imovel funcional
type BeneficioAPI struct {
	Beneficio  string `json:"beneficio"`  // ex.: "Auxílio-Moradia", "Imóvel Funcional"
	Utilizacao string `json:"utilizacao"` // ex.: "Utilizou", "Não utilizou"
}

// VinculoPessoalAPI e a quantidade de servidores por tipo de vinculo
type VinculoPessoalAPI struct {
	Vinculo    string `json:"vinculo"` // ex.: "Comissionado", "Efetivo", "Requisitado"
	Quantidade int    `json:"quantidade"`
}

// PessoalLocalAPI e a lotacao agregada num local (gabinete ou escritorios de apoio)
type PessoalLocalAPI struct {
	Local                     string              `json:"local"` // "Gabinete" ou "Escritório(s) de Apoio"
	Quantidade                int                 `json:"quantidade"`
	Vinculos                  []VinculoPessoalAPI `json:"vinculos"`
	QuantidadeTotalEscritorio int                 `json:"quantidadeTotalEscritorio"`
}

// RecursosUtilizadosAPI e o item de data[] de /senadores/{codigo}/recursos-utilizados.
// A fonte so traz numeros agregados de pessoal, sem nomes nem salarios.
type RecursosUtilizadosAPI struct {
	Parlamentar struct {
		Nome    string `json:"nome"`
		Partido string `json:"partido"`
		Estado  string `json:"estado"`
	} `json:"parlamentar"`
	Ano               int               `json:"ano"`
	Cotas             GrupoRecursosAPI  `json:"cotas"`
	GastosNaoInclusos GrupoRecursosAPI  `json:"gastosNaoInclusos"`
	Beneficios        []BeneficioAPI    `json:"beneficios"`
	Pessoal           []PessoalLocalAPI `json:"pessoal"`
}

type recursosUtilizadosResponse struct {
	StatusCode int                     `json:"statusCode"`
	Msg        string                  `json:"msg"`
	Data       []RecursosUtilizadosAPI `json:"data"`
}

// RecursosUtilizados busca pessoal agregado, beneficios e cota de um senador
// num ano. Devolve (nil, nil) quando a API responde sem dados e
// ErrSenadorNaoEncontrado no 404 (sem novas tentativas). Outras falhas sao
// repetidas com backoff (pkg/retry).
func (c *AdmClient) RecursosUtilizados(ctx context.Context, codigo, ano int) (*RecursosUtilizadosAPI, error) {
	q := url.Values{}
	q.Set("ano", strconv.Itoa(ano))
	endpoint := fmt.Sprintf("%s/api/v1/senadores/%d/recursos-utilizados?%s", c.baseURL, codigo, q.Encode())

	var resultado *RecursosUtilizadosAPI
	var naoEncontrado bool
	err := retry.WithRetry(ctx, c.tentativas, fmt.Sprintf("recursos-utilizados %d/%d", codigo, ano), func() error {
		r, status, err := c.buscarRecursos(ctx, endpoint)
		if status == http.StatusNotFound {
			naoEncontrado = true
			return nil
		}
		if err != nil {
			return err
		}
		resultado = r
		return nil
	})
	if err != nil {
		return nil, err
	}
	if naoEncontrado {
		return nil, fmt.Errorf("codigo %d: %w", codigo, ErrSenadorNaoEncontrado)
	}
	return resultado, nil
}

func (c *AdmClient) buscarRecursos(ctx context.Context, endpoint string) (*RecursosUtilizadosAPI, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("erro criando request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.recursosClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("erro na requisicao de recursos utilizados: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, resp.StatusCode, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("status inesperado %d em %s", resp.StatusCode, endpoint)
	}

	var corpo recursosUtilizadosResponse
	if err := json.NewDecoder(resp.Body).Decode(&corpo); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("erro decodificando recursos utilizados: %w", err)
	}
	if len(corpo.Data) == 0 {
		return nil, resp.StatusCode, nil
	}
	return &corpo.Data[0], resp.StatusCode, nil
}
