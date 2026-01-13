package senado

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	BaseURLLegis = "https://legis.senado.leg.br/dadosabertos"
)

// LegisClient consome a API Legislativa do Senado
type LegisClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewLegisClient cria um novo client
func NewLegisClient() *LegisClient {
	return &LegisClient{
		baseURL: BaseURLLegis,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// === TIPOS DE RESPOSTA DA API ===

// ListaParlamentarResponse representa a resposta de /senador/lista/atual
type ListaParlamentarResponse struct {
	ListaParlamentarEmExercicio struct {
		Parlamentares struct {
			Parlamentar []ParlamentarAPI `json:"Parlamentar"`
		} `json:"Parlamentares"`
	} `json:"ListaParlamentarEmExercicio"`
}

// ParlamentarAPI representa um parlamentar retornado pela API
type ParlamentarAPI struct {
	IdentificacaoParlamentar struct {
		CodigoParlamentar       string `json:"CodigoParlamentar"`
		NomeParlamentar         string `json:"NomeParlamentar"`
		NomeCompletoParlamentar string `json:"NomeCompletoParlamentar"`
		SiglaPartidoParlamentar string `json:"SiglaPartidoParlamentar"`
		UfParlamentar           string `json:"UfParlamentar"`
		UrlFotoParlamentar      string `json:"UrlFotoParlamentar"`
		EmailParlamentar        string `json:"EmailParlamentar"`
	} `json:"IdentificacaoParlamentar"`
}

// === METODOS DO CLIENT ===

// ListarSenadoresAtuais retorna lista de senadores em exercicio
func (c *LegisClient) ListarSenadoresAtuais(ctx context.Context) ([]ParlamentarAPI, error) {
	url := fmt.Sprintf("%s/senador/lista/atual", c.baseURL)

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

	var result ListaParlamentarResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("erro decodificando JSON: %w", err)
	}

	return result.ListaParlamentarEmExercicio.Parlamentares.Parlamentar, nil
}

// DetalhesParlamentarResponse representa a resposta de /senador/{codigo}
type DetalhesParlamentarResponse struct {
	DetalheParlamentar struct {
		Parlamentar struct {
			IdentificacaoParlamentar struct {
				CodigoParlamentar       string `json:"CodigoParlamentar"`
				NomeParlamentar         string `json:"NomeParlamentar"`
				NomeCompletoParlamentar string `json:"NomeCompletoParlamentar"`
				SexoParlamentar         string `json:"SexoParlamentar"`
				FormaTratamento         string `json:"FormaTratamento"`
				UrlFotoParlamentar      string `json:"UrlFotoParlamentar"`
				UrlPaginaParlamentar    string `json:"UrlPaginaParlamentar"`
				EmailParlamentar        string `json:"EmailParlamentar"`
				SiglaPartidoParlamentar string `json:"SiglaPartidoParlamentar"`
				UfParlamentar           string `json:"UfParlamentar"`
			} `json:"IdentificacaoParlamentar"`
			DadosBasicosParlamentar struct {
				DataNascimento      string `json:"DataNascimento"`
				Naturalidade        string `json:"Naturalidade"`
				UfNaturalidade      string `json:"UfNaturalidade"`
			} `json:"DadosBasicosParlamentar"`
		} `json:"Parlamentar"`
	} `json:"DetalheParlamentar"`
}

// DetalhesSenador busca detalhes de um senador especifico
func (c *LegisClient) DetalhesSenador(ctx context.Context, codigo int) (*DetalhesParlamentarResponse, error) {
	url := fmt.Sprintf("%s/senador/%d", c.baseURL, codigo)

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

	var result DetalhesParlamentarResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("erro decodificando JSON: %w", err)
	}

	return &result, nil
}

// === VOTACOES ===

// VotacaoResponse representa a resposta de /votacao
// A API retorna um array de objetos de votacao no nivel top
type VotacaoResponse []VotacaoSessaoAPI

// VotacaoSessaoAPI representa uma sessao de votacao
type VotacaoSessaoAPI struct {
	Ano              int               `json:"ano"`
	CodigoSessao     int               `json:"codigoSessao"`
	DataSessao       string            `json:"dataSessao"`
	DescricaoVotacao string            `json:"descricaoVotacao"`
	VotacaoSecreta   string            `json:"votacaoSecreta"`
	TotalVotosSim    int               `json:"totalVotosSim"`
	TotalVotosNao    int               `json:"totalVotosNao"`
	Votos            []VotoParlamentar `json:"votos"`
}

// VotoParlamentar representa o voto de um parlamentar
type VotoParlamentar struct {
	CodigoParlamentar int    `json:"codigoParlamentar"`
	NomeParlamentar   string `json:"nomeParlamentar"`
	SiglaVoto         string `json:"siglaVotoParlamentar"` // Votou, Sim, Nao, Abstencao, NCom
	DescricaoVoto     string `json:"descricaoVotoParlamentar"`
}

// ListarVotacoesParlamentar busca votacoes de um parlamentar
// NOTA: A API nao aceita parametros dataInicio/dataFim, retorna todas as votacoes
// Retorna lista de sessoes, cada uma com votos dos parlamentares
func (c *LegisClient) ListarVotacoesParlamentar(ctx context.Context, codigoParlamentar int) ([]VotacaoSessaoAPI, error) {
	url := fmt.Sprintf("%s/votacao?codigoParlamentar=%d", c.baseURL, codigoParlamentar)

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

	// A API retorna um array de sessoes diretamente
	var result []VotacaoSessaoAPI
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("erro decodificando JSON: %w", err)
	}

	return result, nil
}
