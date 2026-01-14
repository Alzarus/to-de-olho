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

// === COMISSOES ===

// ComissoesResponse representa a resposta de /senador/{codigo}/comissoes
type ComissoesResponse struct {
	MembroComissaoParlamentar struct {
		Parlamentar struct {
			Codigo          string `json:"Codigo"`
			Nome            string `json:"Nome"`
			MembroComissoes struct {
				Comissao []ComissaoAPI `json:"Comissao"`
			} `json:"MembroComissoes"`
		} `json:"Parlamentar"`
	} `json:"MembroComissaoParlamentar"`
}

// ComissaoAPI representa uma comissao retornada pela API
type ComissaoAPI struct {
	IdentificacaoComissao struct {
		CodigoComissao    string `json:"CodigoComissao"`
		SiglaComissao     string `json:"SiglaComissao"`
		NomeComissao      string `json:"NomeComissao"`
		SiglaCasaComissao string `json:"SiglaCasaComissao"` // SF, CN
	} `json:"IdentificacaoComissao"`
	DescricaoParticipacao string `json:"DescricaoParticipacao"` // Titular, Suplente
	DataInicio            string `json:"DataInicio"`            // YYYY-MM-DD
	DataFim               string `json:"DataFim,omitempty"`     // YYYY-MM-DD
}

// ListarComissoesParlamentar busca comissoes de um parlamentar
func (c *LegisClient) ListarComissoesParlamentar(ctx context.Context, codigoParlamentar int) ([]ComissaoAPI, error) {
	url := fmt.Sprintf("%s/senador/%d/comissoes", c.baseURL, codigoParlamentar)

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

	var result ComissoesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("erro decodificando JSON: %w", err)
	}

	return result.MembroComissaoParlamentar.Parlamentar.MembroComissoes.Comissao, nil
}

// === PROPOSICOES ===

// MateriaAPI representa uma materia/proposicao retornada pela API
// A API retorna um array de objetos diretamente
type MateriaAPI struct {
	ID                    int    `json:"id"`
	CodigoMateria         int    `json:"codigoMateria"`
	Identificacao         string `json:"identificacao"`         // Ex: "PLS 4/2004", "PEC 5/2005"
	Objetivo              string `json:"objetivo"`              // Ex: "Iniciadora"
	CasaIdentificadora    string `json:"casaIdentificadora"`    // SF, CD
	Ementa                string `json:"ementa"`
	TipoDocumento         string `json:"tipoDocumento"`         // "Projeto de Lei Ordinária", "Proposta de Emenda à Constituição"
	DataApresentacao      string `json:"dataApresentacao"`      // YYYY-MM-DD
	Autoria               string `json:"autoria"`               // Nome do autor
	Tramitando            string `json:"tramitando"`            // "Sim", "Não"
	DataDeliberacao       string `json:"dataDeliberacao"`       // YYYY-MM-DD
	SiglaTipoDeliberacao  string `json:"siglaTipoDeliberacao"`  // ARQUIVADO_FIM_LEGISLATURA, APROVADA_NO_PLENARIO, etc.
	NormaGerada           string `json:"normaGerada,omitempty"` // "Lei nº 11.738 de 16/07/2008"
}

// ListarProposicoesParlamentar busca proposicoes de autoria de um parlamentar
// A API retorna um array de materias diretamente
func (c *LegisClient) ListarProposicoesParlamentar(ctx context.Context, codigoParlamentar int) ([]MateriaAPI, error) {
	url := fmt.Sprintf("%s/processo?codigoParlamentarAutor=%d", c.baseURL, codigoParlamentar)

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

	// A API retorna um array de materias diretamente
	var result []MateriaAPI
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("erro decodificando JSON: %w", err)
	}

	return result, nil
}
