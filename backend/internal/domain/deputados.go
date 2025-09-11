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
	Ano            int     `json:"ano"`
	Mes            int     `json:"mes"`
	TipoDespesa    string  `json:"tipoDespesa"`
	CodDocumento   int     `json:"codDocumento"`
	TipoDocumento  string  `json:"tipoDocumento"`
	CodTipoDoc     int     `json:"codTipoDocumento"`
	DataDocumento  string  `json:"dataDocumento"`
	NumDocumento   string  `json:"numDocumento"`
	ValorLiquido   float64 `json:"valorLiquido"`
	Fornecedor     string  `json:"nomeFornecedor"`
	CNPJFornecedor string  `json:"cnpjCpfFornecedor"`
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
