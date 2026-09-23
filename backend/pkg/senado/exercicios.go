package senado

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

// umOuVarios decodifica campos que a API (convertida de XML) devolve como
// objeto quando ha um item e como lista quando ha varios.
type umOuVarios[T any] []T

func (u *umOuVarios[T]) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		return nil
	}
	if b[0] == '[' {
		var lista []T
		if err := json.Unmarshal(b, &lista); err != nil {
			return err
		}
		*u = lista
		return nil
	}
	var item T
	if err := json.Unmarshal(b, &item); err != nil {
		return err
	}
	*u = []T{item}
	return nil
}

type mandatosResponse struct {
	MandatoParlamentar struct {
		Parlamentar struct {
			Mandatos struct {
				Mandato umOuVarios[mandatoAPI] `json:"Mandato"`
			} `json:"Mandatos"`
		} `json:"Parlamentar"`
	} `json:"MandatoParlamentar"`
}

type mandatoAPI struct {
	DescricaoParticipacao        string `json:"DescricaoParticipacao"` // Titular, 1º Suplente...
	PrimeiraLegislaturaDoMandato struct {
		NumeroLegislatura string `json:"NumeroLegislatura"`
	} `json:"PrimeiraLegislaturaDoMandato"`
	Exercicios struct {
		Exercicio umOuVarios[struct {
			DataInicio string `json:"DataInicio"`
			DataFim    string `json:"DataFim"`
		}] `json:"Exercicio"`
	} `json:"Exercicios"`
}

// Exercicio e um periodo em que o parlamentar ocupou a cadeira. Fim vazio:
// em exercicio ate hoje.
type Exercicio struct {
	Legislatura  int
	Participacao string
	Inicio       string // AAAA-MM-DD
	Fim          string // AAAA-MM-DD, inclusivo; vazio se em aberto
}

// ListarExercicios busca os periodos de exercicio de todos os mandatos.
// Endpoint: /senador/{codigo}/mandatos
func (c *LegisClient) ListarExercicios(ctx context.Context, codigoParlamentar int) ([]Exercicio, error) {
	var resp mandatosResponse
	if err := c.getJSON(ctx, fmt.Sprintf("%s/senador/%d/mandatos", c.baseURL, codigoParlamentar), &resp); err != nil {
		return nil, err
	}
	var out []Exercicio
	for _, m := range resp.MandatoParlamentar.Parlamentar.Mandatos.Mandato {
		leg, _ := strconv.Atoi(m.PrimeiraLegislaturaDoMandato.NumeroLegislatura)
		for _, e := range m.Exercicios.Exercicio {
			if e.DataInicio == "" {
				continue
			}
			out = append(out, Exercicio{Legislatura: leg, Participacao: m.DescricaoParticipacao, Inicio: e.DataInicio, Fim: e.DataFim})
		}
	}
	return out, nil
}
