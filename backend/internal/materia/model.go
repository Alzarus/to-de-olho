// Package materia guarda o nome popular (apelido) e a descricao breve das
// materias votadas ou apresentadas pelos senadores.
//
// Fontes, em ordem de prioridade:
//  1. apelido oficial do Senado: campo `apelido` de /processo/{id};
//  2. curadoria do projeto (materias_apelidos_curados): so nomes verificados
//     numa pagina oficial, cuja URL fica gravada em fonte_url.
//
// Nada aqui e gerado por IA: a descricao e a explicacao da ementa publicada
// pelo Senado, ou a propria ementa.
package materia

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Materia e o detalhe de uma materia, vindo de /processo/{id}.
type Materia struct {
	CodigoMateria    int       `gorm:"primaryKey;autoIncrement:false" json:"codigo_materia"`
	IdProcesso       int       `gorm:"uniqueIndex:idx_materia_id_processo;not null" json:"id_processo"`
	Identificacao    string    `gorm:"not null" json:"identificacao"` // "PL 2338/2023"
	Sigla            string    `gorm:"size:20" json:"sigla"`
	Apelido          *string   `json:"apelido"` // nulo quando o Senado nao atribui
	Ementa           string    `json:"ementa"`
	ExplicacaoEmenta *string   `json:"explicacao_ementa"` // nulo quando o Senado nao publica
	Temas            Temas     `gorm:"type:jsonb" json:"temas"`
	UrlDocumento     string    `json:"url_documento"`
	UpdatedAt        time.Time `gorm:"index" json:"updated_at"`
}

// TableName define o nome da tabela
func (Materia) TableName() string { return "materias" }

// ApelidoCurado e um nome popular atribuido pelo projeto a uma materia sem
// apelido oficial. So entra com fonte oficial verificada (fonte_url).
type ApelidoCurado struct {
	CodigoMateria int       `gorm:"primaryKey;autoIncrement:false" json:"codigo_materia"`
	Identificacao string    `gorm:"not null" json:"identificacao"`
	Apelido       string    `gorm:"not null" json:"apelido"`
	FonteURL      string    `gorm:"column:fonte_url;not null" json:"fonte_url"`
	Observacao    string    `json:"observacao,omitempty"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName define o nome da tabela
func (ApelidoCurado) TableName() string { return "materias_apelidos_curados" }

// Temas sao as classificacoes da materia (classificacoes[].descricao), gravadas
// como array JSON.
type Temas []string

// GormDataType diz ao GORM que Temas e uma coluna (jsonb), nao uma relacao
func (Temas) GormDataType() string { return "jsonb" }

// Value grava os temas como JSON (nulo quando vazio)
func (t Temas) Value() (driver.Value, error) {
	if len(t) == 0 {
		return nil, nil
	}
	b, err := json.Marshal([]string(t))
	if err != nil {
		return nil, fmt.Errorf("temas: %w", err)
	}
	return string(b), nil
}

// Scan le os temas do JSON (nulo vira lista vazia)
func (t *Temas) Scan(v any) error {
	var b []byte
	switch x := v.(type) {
	case nil:
		*t = nil
		return nil
	case []byte:
		b = x
	case string:
		b = []byte(x)
	default:
		return fmt.Errorf("temas: tipo inesperado %T", v)
	}
	var out []string
	if err := json.Unmarshal(b, &out); err != nil {
		return fmt.Errorf("temas: %w", err)
	}
	*t = out
	return nil
}

// NormalizarApelido fica com a primeira linha, sem espacos nas pontas. A API
// ja mandou "Lei da Reciprocidade\nLei da Reciprocidade Econômica" e
// "PL Antifacção ". Vazio vira nulo.
func NormalizarApelido(s string) *string {
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		s = s[:i]
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

// textoOuNulo devolve nulo para texto vazio
func textoOuNulo(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}
