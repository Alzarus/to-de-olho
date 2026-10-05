package senado

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Alzarus/to-de-olho/pkg/retry"
)

// === ESCRITORIOS DE APOIO, TERCEIRIZADOS E ESTAGIARIOS (retrato atual) ===
//
// As tres rotas trazem a situacao de hoje, sem historico por ano. O sync usa
// so o endereco dos escritorios (informacao institucional) e a contagem de
// terceirizados e estagiarios por gabinete: nomes de pessoas nao sao gravados.

// EscritorioApoioAPI e um escritorio de apoio de /senadores/escritorios-apoio
type EscritorioApoioAPI struct {
	Parlamentar struct {
		Nome   string `json:"nome"`   // nome parlamentar em maiusculas, ex.: "ALAN RICK"
		Estado string `json:"estado"` // UF
	} `json:"parlamentar"`
	Setor struct {
		Nome     string `json:"nome"` // "Escritório de Apoio nº 1 do Senador Alan Rick"
		Telefone string `json:"telefone"`
		Endereco string `json:"endereco"`
	} `json:"setor"`
}

// TerceirizadoAPI e um contrato de /contratacoes/terceirizados (so os campos usados)
type TerceirizadoAPI struct {
	Situacao string `json:"situacao"` // "Ativo"
	Lotacao  struct {
		Sigla string `json:"sigla"`
		Nome  string `json:"nome"` // "GABINETE DO SENADOR CLEITINHO"
	} `json:"lotacao"`
}

// EstagiarioAPI e um estagiario de /colaboradores/estagiarios (so a lotacao)
type EstagiarioAPI struct {
	SiglaOrgao string `json:"siglaOrgao"`
	NomeOrgao  string `json:"nomeOrgao"`
}

type envelopeAdm[T any] struct {
	Data []T `json:"data"`
}

// EscritoriosApoio lista os escritorios de apoio de todos os senadores
func (c *AdmClient) EscritoriosApoio(ctx context.Context) ([]EscritorioApoioAPI, error) {
	var corpo envelopeAdm[EscritorioApoioAPI]
	err := c.getJSONAdm(ctx, "/api/v1/senadores/escritorios-apoio", &corpo)
	return corpo.Data, err
}

// Terceirizados lista os terceirizados do Senado (a rota devolve uma lista direta)
func (c *AdmClient) Terceirizados(ctx context.Context) ([]TerceirizadoAPI, error) {
	var lista []TerceirizadoAPI
	err := c.getJSONAdm(ctx, "/api/v1/contratacoes/terceirizados", &lista)
	return lista, err
}

// Estagiarios lista os estagiarios do Senado
func (c *AdmClient) Estagiarios(ctx context.Context) ([]EstagiarioAPI, error) {
	var corpo envelopeAdm[EstagiarioAPI]
	err := c.getJSONAdm(ctx, "/api/v1/colaboradores/estagiarios", &corpo)
	return corpo.Data, err
}

// getJSONAdm faz o GET com novas tentativas (pkg/retry); as rotas antigas
// (/senadores/escritorios, /servidores/estagiarios) redirecionam para estas
func (c *AdmClient) getJSONAdm(ctx context.Context, caminho string, out any) error {
	endpoint := c.baseURL + caminho
	return retry.WithRetry(ctx, c.tentativas, caminho, func() error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return fmt.Errorf("erro criando request: %w", err)
		}
		req.Header.Set("Accept", "application/json")
		resp, err := c.recursosClient.Do(req)
		if err != nil {
			return fmt.Errorf("erro na requisicao %s: %w", caminho, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("status inesperado %d em %s", resp.StatusCode, caminho)
		}
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("erro decodificando %s: %w", caminho, err)
		}
		return nil
	})
}
