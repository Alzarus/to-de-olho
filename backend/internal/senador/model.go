package senador

import "time"

// Senador representa um senador da republica
type Senador struct {
	ID                int       `gorm:"primaryKey" json:"id"`
	CodigoParlamentar int       `gorm:"uniqueIndex;not null" json:"codigo_parlamentar"`
	Nome              string    `gorm:"not null" json:"nome"`
	NomeCompleto      string    `json:"nome_completo,omitempty"`
	Partido           string    `json:"partido"`
	UF                string    `gorm:"size:2" json:"uf"`
	FotoURL           string    `json:"foto_url,omitempty"`
	Email             string    `json:"email,omitempty"`
	EmExercicio       bool      `gorm:"default:true" json:"em_exercicio"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`

	// Relacionamentos
	Mandatos []Mandato `gorm:"foreignKey:SenadorID" json:"mandatos,omitempty"`
}

// Mandato representa um mandato de um senador
type Mandato struct {
	ID          int        `gorm:"primaryKey" json:"id"`
	SenadorID   int        `gorm:"index;not null" json:"senador_id"`
	Legislatura int        `json:"legislatura"`
	Inicio      time.Time  `json:"inicio"`
	Fim         *time.Time `json:"fim,omitempty"`
	Tipo        string     `json:"tipo"` // Titular, Suplente, etc.
	CreatedAt   time.Time  `json:"created_at"`
}

// TableName define o nome da tabela para Senador
func (Senador) TableName() string {
	return "senadores"
}

// TableName define o nome da tabela para Mandato
func (Mandato) TableName() string {
	return "mandatos"
}
