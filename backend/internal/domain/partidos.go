package domain

import (
	"errors"
	"strings"
	"time"
)

// Partido representa um partido político retornado pela API da Câmara
type Partido struct {
	ID        int64                  `json:"id"`
	Sigla     string                 `json:"sigla"`
	Nome      string                 `json:"nome"`
	URI       string                 `json:"uri"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	CreatedAt time.Time              `json:"created_at,omitempty"`
	UpdatedAt time.Time              `json:"updated_at,omitempty"`
}

// Validate valida campos essenciais do partido
func (p *Partido) Validate() error {
	if p.ID <= 0 {
		return errors.New("id do partido inválido")
	}
	if strings.TrimSpace(p.Sigla) == "" && strings.TrimSpace(p.Nome) == "" {
		return errors.New("sigla ou nome do partido é obrigatório")
	}
	return nil
}
