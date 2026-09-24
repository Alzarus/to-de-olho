package senado

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// CargoMesaAPI e um integrante atual da Mesa Diretora do Senado
type CargoMesaAPI struct {
	CodigoParlamentar int
	Cargo             string // ex.: "PRESIDENTE", "1º VICE-PRESIDENTE", "1ª SECRETÁRIA"
}

// listaOuItem aceita um valor JSON que a API ora manda como objeto, ora como lista
type listaOuItem[T any] []T

func (l *listaOuItem[T]) UnmarshalJSON(b []byte) error {
	b = []byte(strings.TrimSpace(string(b)))
	if len(b) > 0 && b[0] == '[' {
		var itens []T
		if err := json.Unmarshal(b, &itens); err != nil {
			return err
		}
		*l = itens
		return nil
	}
	var item T
	if err := json.Unmarshal(b, &item); err != nil {
		return err
	}
	*l = []T{item}
	return nil
}

type mesaSFResponse struct {
	MesaSenado struct {
		Colegiados struct {
			Colegiado listaOuItem[struct {
				SiglaColegiado string `json:"SiglaColegiado"`
				Cargos         struct {
					Cargo listaOuItem[struct {
						Cargo listaOuItem[string] `json:"Cargo"`
						Http  string              `json:"Http"` // codigo do parlamentar
					}] `json:"Cargo"`
				} `json:"Cargos"`
			}] `json:"Colegiado"`
		} `json:"Colegiados"`
	} `json:"MesaSenado"`
}

// ComposicaoMesaSF lista os integrantes atuais da Mesa do Senado
// (/composicao/mesaSF). Parte da equipe de quem ocupa cargo na Mesa fica
// lotada nos orgaos da Mesa, fora do gabinete.
func (c *LegisClient) ComposicaoMesaSF(ctx context.Context) ([]CargoMesaAPI, error) {
	var resp mesaSFResponse
	if err := c.getJSON(ctx, c.baseURL+"/composicao/mesaSF.json", &resp); err != nil {
		return nil, fmt.Errorf("composicao da mesa: %w", err)
	}
	var cargos []CargoMesaAPI
	for _, col := range resp.MesaSenado.Colegiados.Colegiado {
		for _, cg := range col.Cargos.Cargo {
			codigo, err := strconv.Atoi(strings.TrimSpace(cg.Http))
			if err != nil || codigo == 0 || len(cg.Cargo) == 0 {
				continue
			}
			cargos = append(cargos, CargoMesaAPI{CodigoParlamentar: codigo, Cargo: strings.TrimSpace(cg.Cargo[0])})
		}
	}
	return cargos, nil
}
