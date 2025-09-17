package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type Deputado struct {
	ID       int    `json:"id"`
	Nome     string `json:"nome"`
	Partido  string `json:"siglaPartido"`
	UF       string `json:"siglaUf"`
	URLFoto  string `json:"urlFoto"`
	Situacao string `json:"condicaoEleitoral"`
	Email    string `json:"email"`
}

// Validate valida os dados básicos do deputado
func (d *Deputado) Validate() error {
	if d.ID <= 0 {
		return errors.New("ID deve ser maior que zero")
	}

	if strings.TrimSpace(d.Nome) == "" {
		return errors.New("nome é obrigatório")
	}

	if strings.TrimSpace(d.Partido) == "" {
		return errors.New("partido é obrigatório")
	}

	if !isValidUF(d.UF) {
		return errors.New("UF inválida")
	}

	return nil
}

// GetNomeCompleto retorna nome formatado com partido e UF
func (d *Deputado) GetNomeCompleto() string {
	return fmt.Sprintf("%s (%s/%s)", d.Nome, d.Partido, d.UF)
}

// isValidUF verifica se a UF é válida
func isValidUF(uf string) bool {
	validUFs := []string{
		"AC", "AL", "AP", "AM", "BA", "CE", "DF", "ES", "GO", "MA",
		"MT", "MS", "MG", "PA", "PB", "PR", "PE", "PI", "RJ", "RN",
		"RS", "RO", "RR", "SC", "SP", "SE", "TO",
	}

	for _, validUF := range validUFs {
		if uf == validUF {
			return true
		}
	}
	return false
}

type Despesa struct {
	Ano               int     `json:"ano"`
	Mes               int     `json:"mes"`
	TipoDespesa       string  `json:"tipoDespesa"`
	CodDocumento      int     `json:"codDocumento"`
	TipoDocumento     string  `json:"tipoDocumento"`
	CodTipoDocumento  int     `json:"codTipoDocumento"`
	DataDocumento     string  `json:"dataDocumento"`
	NumDocumento      string  `json:"numDocumento"`
	ValorDocumento    float64 `json:"valorDocumento"`
	URLDocumento      string  `json:"urlDocumento"`
	NomeFornecedor    string  `json:"nomeFornecedor"`
	CNPJCPFFornecedor string  `json:"cnpjCpfFornecedor"`
	ValorLiquido      float64 `json:"valorLiquido"`
	ValorBruto        float64 `json:"valorBruto,omitempty"`
	ValorGlosa        float64 `json:"valorGlosa"`
	NumRessarcimento  string  `json:"numRessarcimento,omitempty"`
	CodLote           int     `json:"codLote"`
	Parcela           int     `json:"parcela,omitempty"`
}

// Validate valida os dados da despesa
func (d *Despesa) Validate() error {
	currentYear := time.Now().Year()

	if d.Ano < 2000 || d.Ano > currentYear {
		return errors.New("ano deve ser entre 2000 e ano atual")
	}

	if d.Mes < 1 || d.Mes > 12 {
		return errors.New("mês deve ser entre 1 e 12")
	}

	if d.ValorLiquido < 0 {
		return errors.New("valor deve ser positivo")
	}

	return nil
}

// GetMesNome retorna o nome do mês por extenso
func (d *Despesa) GetMesNome() string {
	meses := []string{
		"", "Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho",
		"Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro",
	}

	if d.Mes < 1 || d.Mes > 12 {
		return "Mês Inválido"
	}

	return meses[d.Mes]
}

// Erros do domínio deputados
var (
	ErrDeputadoIDInvalido      = errors.New("ID do deputado inválido")
	ErrDeputadoNomeVazio       = errors.New("nome do deputado é obrigatório")
	ErrDeputadoUFInvalida      = errors.New("UF inválida")
	ErrDeputadoPartidoInvalido = errors.New("partido inválido")
	ErrDeputadoNaoEncontrado   = errors.New("deputado não encontrado")
)

// DeputadoFilter representa os filtros para busca de deputados
type DeputadoFilter struct {
	Nome    string `json:"nome,omitempty"`
	UF      string `json:"uf,omitempty"`
	Partido string `json:"partido,omitempty"`
	Limite  int    `json:"limite,omitempty"`
	Pagina  int    `json:"pagina,omitempty"`
}

// SetDefaults aplica valores padrão aos filtros
func (f *DeputadoFilter) SetDefaults() {
	if f.Limite <= 0 {
		f.Limite = 20
	}
	if f.Pagina <= 0 {
		f.Pagina = 1
	}
}
