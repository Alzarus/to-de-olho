// Package gabinete guarda a estrutura de pessoal dos senadores em numeros
// agregados (quantidade por local e vinculo) e o uso de beneficios.
//
// Decisao de privacidade (LGPD, fase 1): nenhum nome, matricula ou salario
// individual e gravado. Quem quiser o detalhe segue o link para a pagina
// oficial de transparencia do Senado.
package gabinete

import "time"

// Locais de lotacao normalizados
const (
	LocalGabinete   = "GABINETE"
	LocalEscritorio = "ESCRITORIO"
)

// Recurso e a quantidade de servidores de um senador num ano, por local e vinculo.
// Chave natural: (senador_id, ano, local, vinculo).
type Recurso struct {
	ID           int       `gorm:"primaryKey" json:"-"`
	SenadorID    int       `gorm:"uniqueIndex:idx_gabinete_recurso_chave,priority:1;not null" json:"senador_id"`
	Ano          int       `gorm:"uniqueIndex:idx_gabinete_recurso_chave,priority:2;not null" json:"ano"`
	Local        string    `gorm:"uniqueIndex:idx_gabinete_recurso_chave,priority:3;size:20;not null" json:"local"`
	Vinculo      string    `gorm:"uniqueIndex:idx_gabinete_recurso_chave,priority:4;size:80;not null" json:"vinculo"`
	Quantidade   int       `gorm:"not null" json:"quantidade"`
	AtualizadoEm time.Time `gorm:"not null" json:"atualizado_em"`
}

// TableName define o nome da tabela
func (Recurso) TableName() string { return "gabinete_recursos" }

// Beneficio e o uso de um beneficio (auxilio-moradia, imovel funcional) no ano,
// como a fonte informa ("Utilizou", "Não utilizou").
// Chave natural: (senador_id, ano, tipo).
type Beneficio struct {
	ID           int       `gorm:"primaryKey" json:"-"`
	SenadorID    int       `gorm:"uniqueIndex:idx_gabinete_beneficio_chave,priority:1;not null" json:"senador_id"`
	Ano          int       `gorm:"uniqueIndex:idx_gabinete_beneficio_chave,priority:2;not null" json:"ano"`
	Tipo         string    `gorm:"uniqueIndex:idx_gabinete_beneficio_chave,priority:3;size:120;not null" json:"tipo"`
	Utilizacao   string    `gorm:"size:120" json:"utilizacao"`
	AtualizadoEm time.Time `gorm:"not null" json:"atualizado_em"`
}

// TableName define o nome da tabela
func (Beneficio) TableName() string { return "gabinete_beneficios" }

// CargoMesa e o cargo atual de um senador na Mesa Diretora. Parte da equipe de
// quem ocupa a Mesa (sobretudo a Presidencia) fica lotada nos orgaos da Mesa e
// nao aparece no gabinete: a ficha mostra uma nota.
type CargoMesa struct {
	SenadorID    int       `gorm:"primaryKey;autoIncrement:false" json:"senador_id"`
	Cargo        string    `gorm:"size:80;not null" json:"cargo"`
	AtualizadoEm time.Time `gorm:"not null" json:"atualizado_em"`
}

// TableName define o nome da tabela
func (CargoMesa) TableName() string { return "gabinete_mesa" }

// Escritorio e um escritorio de apoio do senador no estado (retrato atual da
// fonte, sem historico). Endereco e telefone sao do escritorio, institucionais.
type Escritorio struct {
	ID           int       `gorm:"primaryKey" json:"-"`
	SenadorID    int       `gorm:"index;not null" json:"senador_id"`
	Nome         string    `gorm:"size:200;not null" json:"nome"`
	Endereco     string    `gorm:"size:400" json:"endereco"`
	Telefone     string    `gorm:"size:80" json:"telefone"`
	AtualizadoEm time.Time `gorm:"not null" json:"atualizado_em"`
}

// TableName define o nome da tabela
func (Escritorio) TableName() string { return "gabinete_escritorios" }

// Tipos de colaborador contados a parte: nao entram em recursos-utilizados,
// mas a pagina de transparencia do Senado os lista no gabinete
const (
	ColaboradorTerceirizado = "Terceirizado"
	ColaboradorEstagiario   = "Estagiário"
)

// Colaborador e a quantidade atual de terceirizados ou estagiarios lotados no
// gabinete do senador (so a contagem, sem nomes). Chave: (senador_id, tipo).
type Colaborador struct {
	SenadorID    int       `gorm:"primaryKey;autoIncrement:false" json:"senador_id"`
	Tipo         string    `gorm:"primaryKey;size:40" json:"tipo"`
	Quantidade   int       `gorm:"not null" json:"quantidade"`
	AtualizadoEm time.Time `gorm:"not null" json:"atualizado_em"`
}

// TableName define o nome da tabela
func (Colaborador) TableName() string { return "gabinete_colaboradores" }

// Modelos lista as tabelas do pacote para o AutoMigrate
func Modelos() []any {
	return []any{&Recurso{}, &Beneficio{}, &CargoMesa{}, &Escritorio{}, &Colaborador{}}
}

// === RESPOSTA DA API ===

// VinculoResumo e a quantidade de um vinculo num local
type VinculoResumo struct {
	Vinculo    string `json:"vinculo"`
	Quantidade int    `json:"quantidade"`
}

// LocalResumo agrega os vinculos de um local
type LocalResumo struct {
	Local    string          `json:"local"`  // GABINETE | ESCRITORIO
	Rotulo   string          `json:"rotulo"` // "Gabinete em Brasília" | "Escritórios no estado"
	Total    int             `json:"total"`
	Vinculos []VinculoResumo `json:"vinculos"`
}

// BeneficioResumo e um beneficio na resposta
type BeneficioResumo struct {
	Tipo       string `json:"tipo"`
	Utilizacao string `json:"utilizacao"`
}

// EscritorioResumo e um escritorio de apoio na resposta
type EscritorioResumo struct {
	Nome     string `json:"nome"`
	Endereco string `json:"endereco"`
	Telefone string `json:"telefone"`
}

// Resumo e a resposta de GET /senadores/:id/gabinete
type Resumo struct {
	SenadorID       int               `json:"senador_id"`
	Ano             int               `json:"ano"` // 0 quando nao ha dados
	AnosDisponiveis []int             `json:"anos_disponiveis"`
	Total           int               `json:"total"`
	Locais          []LocalResumo     `json:"locais"`
	PorVinculo      []VinculoResumo   `json:"por_vinculo"`
	Beneficios      []BeneficioResumo `json:"beneficios"`
	CargoMesa       *string           `json:"cargo_mesa"`
	// Retrato atual (a fonte nao tem historico): vale para hoje, qualquer
	// que seja o ano escolhido
	Escritorios   []EscritorioResumo `json:"escritorios"`
	Colaboradores []VinculoResumo    `json:"colaboradores"`
	AtualizadoEm    *time.Time        `json:"atualizado_em"`
	FonteURL        string            `json:"fonte_url"`
}
