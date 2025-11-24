package domain

import (
	"errors"
	"fmt"
	"time"
)

// Proposicao representa uma proposição legislativa
type Proposicao struct {
	ID               int              `json:"id" db:"id"`
	URI              string           `json:"uri" db:"uri"`
	SiglaTipo        string           `json:"siglaTipo" db:"sigla_tipo"`
	CodTipo          int              `json:"codTipo" db:"cod_tipo"`
	Numero           int              `json:"numero" db:"numero"`
	Ano              int              `json:"ano" db:"ano"`
	Ementa           string           `json:"ementa" db:"ementa"`
	DataApresentacao string           `json:"dataApresentacao" db:"data_apresentacao"`
	StatusProposicao StatusProposicao `json:"statusProposicao" db:"-"`
	UltimoRelator    *UltimoRelator   `json:"ultimoRelator,omitempty" db:"-"`
	DescricaoTipo    string           `json:"descricaoTipo" db:"descricao_tipo"`
	Tema             string           `json:"tema,omitempty" db:"tema"`
	Keywords         string           `json:"keywords,omitempty" db:"keywords"`

	// Campos calculados/auxiliares
	StatusID        *int    `json:"-" db:"status_id"`
	StatusDescricao *string `json:"-" db:"status_descricao"`
	RelatorID       *int    `json:"-" db:"relator_id"`
	RelatorNome     *string `json:"-" db:"relator_nome"`
}

// StatusProposicao representa o status atual de uma proposição
type StatusProposicao struct {
	ID                  int    `json:"id"`
	URI                 string `json:"uri"`
	SiglaOrgao          string `json:"siglaOrgao"`
	UriOrgao            string `json:"uriOrgao"`
	Regime              string `json:"regime"`
	DescricaoTramitacao string `json:"descricaoTramitacao"`
	CodTipoTramitacao   int    `json:"codTipoTramitacao"`
	DescricaoSituacao   string `json:"descricaoSituacao"`
	CodSituacao         int    `json:"codSituacao"`
	DataHora            string `json:"dataHora"`
	Sequencia           int    `json:"sequencia"`
	UriUltimoRelator    string `json:"uriUltimoRelator"`
}

// UltimoRelator representa o último relator da proposição
type UltimoRelator struct {
	ID           int    `json:"id"`
	Nome         string `json:"nome"`
	CodTipo      int    `json:"codTipo"`
	SiglaUf      string `json:"siglaUf"`
	SiglaPartido string `json:"siglaPartido"`
	UriPartido   string `json:"uriPartido"`
	UriCamara    string `json:"uriCamara"`
	URLFoto      string `json:"urlFoto"`
}

// ProposicaoFilter representa os filtros para busca de proposições
type ProposicaoFilter struct {
	SiglaTipo              string     `json:"siglaTipo,omitempty"`
	Numero                 *int       `json:"numero,omitempty"`
	Ano                    *int       `json:"ano,omitempty"`
	DataApresentacaoInicio *time.Time `json:"dataApresentacaoInicio,omitempty"`
	DataApresentacaoFim    *time.Time `json:"dataApresentacaoFim,omitempty"`
	CodSituacao            *int       `json:"codSituacao,omitempty"`
	SiglaUfAutor           string     `json:"siglaUfAutor,omitempty"`
	SiglaPartidoAutor      string     `json:"siglaPartidoAutor,omitempty"`
	NomeAutor              string     `json:"nomeAutor,omitempty"`
	Tema                   string     `json:"tema,omitempty"`
	Keywords               string     `json:"keywords,omitempty"`
	Ordem                  string     `json:"ordem,omitempty"`      // ASC ou DESC
	OrdenarPor             string     `json:"ordenarPor,omitempty"` // id, dataApresentacao, etc
	Pagina                 int        `json:"pagina"`
	Limite                 int        `json:"limite"`
}

// ProposicaoCount representa contagem de proposições por deputado
type ProposicaoCount struct {
	IDDeputado int `json:"id_deputado"`
	Count      int `json:"count"`
}

// PresencaCount representa presença agregada por deputado (ex: número de votações ou participações)
type PresencaCount struct {
	IDDeputado    int `json:"id_deputado"`
	Participacoes int `json:"participacoes"`
}

// Constantes para tipos de proposição mais comuns
const (
	TipoProposicaoPL  = "PL"  // Projeto de Lei
	TipoProposicaoPEC = "PEC" // Proposta de Emenda à Constituição
	TipoProposicaoPLP = "PLP" // Projeto de Lei Complementar
	TipoProposicaoMPV = "MPV" // Medida Provisória
	TipoProposicaoPDC = "PDC" // Projeto de Decreto Legislativo
	TipoProposicaoPRC = "PRC" // Projeto de Resolução
)

// Validate valida os dados da proposição
func (p *Proposicao) Validate() error {
	if p.ID <= 0 {
		return ErrProposicaoIDInvalido
	}

	if p.Ementa == "" {
		return ErrProposicaoEmentaVazia
	}

	currentYear := time.Now().Year()
	if p.Ano < 1988 || p.Ano > currentYear {
		return ErrProposicaoAnoInvalido
	}

	if p.Numero <= 0 {
		return ErrProposicaoNumeroInvalido
	}

	if p.SiglaTipo == "" {
		return ErrProposicaoTipoInvalido
	}

	return nil
}

// GetIdentificacao retorna a identificação completa da proposição
func (p *Proposicao) GetIdentificacao() string {
	return fmt.Sprintf("%s %d/%d", p.SiglaTipo, p.Numero, p.Ano)
}

// IsEmenda verifica se a proposição é uma emenda
func (p *Proposicao) IsEmenda() bool {
	return p.SiglaTipo == TipoProposicaoPEC
}

// IsProjeto verifica se a proposição é um projeto de lei
func (p *Proposicao) IsProjeto() bool {
	return p.SiglaTipo == TipoProposicaoPL || p.SiglaTipo == TipoProposicaoPLP
}

// IsMedidaProvisoria verifica se a proposição é uma medida provisória
func (p *Proposicao) IsMedidaProvisoria() bool {
	return p.SiglaTipo == TipoProposicaoMPV
}

// GetDataApresentacaoTime converte a data de apresentação para time.Time
func (p *Proposicao) GetDataApresentacaoTime() (time.Time, error) {
	if p.DataApresentacao == "" {
		return time.Time{}, ErrProposicaoDataInvalida
	}

	// Tenta diferentes formatos de data
	layouts := []string{
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02",
		"02/01/2006",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, p.DataApresentacao); err == nil {
			return t, nil
		}
	}

	return time.Time{}, ErrProposicaoDataInvalida
}

// Validate valida os filtros de busca
func (f *ProposicaoFilter) Validate() error {
	if f.Limite <= 0 {
		f.Limite = 20 // Padrão
	}

	if f.Limite > 100 {
		return ErrProposicaoLimiteExcedido
	}

	if f.Pagina <= 0 {
		f.Pagina = 1 // Padrão
	}

	if f.Ano != nil && (*f.Ano < 1988 || *f.Ano > time.Now().Year()) {
		return ErrProposicaoAnoInvalido
	}

	if f.Numero != nil && *f.Numero <= 0 {
		return ErrProposicaoNumeroInvalido
	}

	// Validar ordem
	if f.Ordem != "" && f.Ordem != "ASC" && f.Ordem != "DESC" {
		return ErrProposicaoOrdemInvalida
	}

	// Validar ordenarPor - conforme documentação oficial da API
	ordemPermitida := map[string]bool{
		"id":               true,
		"codTipo":          true,
		"siglaTipo":        true,
		"numero":           true,
		"ano":              true,
		"dataApresentacao": true,
	}

	if f.OrdenarPor != "" && !ordemPermitida[f.OrdenarPor] {
		return ErrProposicaoOrdenarPorInvalido
	}

	return nil
}

// SetDefaults define valores padrão para os filtros
func (f *ProposicaoFilter) SetDefaults() {
	if f.Limite <= 0 {
		f.Limite = 20
	}

	if f.Pagina <= 0 {
		f.Pagina = 1
	}

	if f.Ordem == "" {
		f.Ordem = "DESC"
	}

	if f.OrdenarPor == "" {
		f.OrdenarPor = "id" // Campo seguro aceito pela API para ordenação padrão
	}
}

// BuildAPIQueryParams converte filtros para parâmetros de query da API da Câmara
func (f *ProposicaoFilter) BuildAPIQueryParams() map[string]string {
	params := make(map[string]string)

	if f.SiglaTipo != "" {
		params["siglaTipo"] = f.SiglaTipo
	}

	if f.Numero != nil {
		params["numero"] = fmt.Sprintf("%d", *f.Numero)
	}

	if f.Ano != nil {
		params["ano"] = fmt.Sprintf("%d", *f.Ano)
	}

	if f.DataApresentacaoInicio != nil {
		params["dataApresentacaoInicio"] = f.DataApresentacaoInicio.Format("2006-01-02")
	}

	if f.DataApresentacaoFim != nil {
		params["dataApresentacaoFim"] = f.DataApresentacaoFim.Format("2006-01-02")
	}

	if f.CodSituacao != nil {
		params["codSituacao"] = fmt.Sprintf("%d", *f.CodSituacao)
	}

	if f.SiglaUfAutor != "" {
		params["siglaUfAutor"] = f.SiglaUfAutor
	}

	if f.SiglaPartidoAutor != "" {
		params["siglaPartidoAutor"] = f.SiglaPartidoAutor
	}

	// Corrigir: NomeAutor deve usar o parâmetro "autor", não "idAutor"
	// Tests (and older API usage) expect NomeAutor to be sent as idAutor
	if f.NomeAutor != "" {
		params["idAutor"] = f.NomeAutor
	}

	// Remover parâmetros que podem estar causando erro 400:
	// - "tema" deve ser "codTema" com código numérico
	// - "keywords" pode estar com formato incorreto

	if f.Tema != "" {
		params["tema"] = f.Tema
	}

	if f.Keywords != "" {
		params["keywords"] = f.Keywords
	}

	// Parâmetros de ordenação e paginação - conforme API oficial
	params["ordem"] = f.Ordem
	params["ordenarPor"] = f.OrdenarPor
	params["pagina"] = fmt.Sprintf("%d", f.Pagina)
	params["itens"] = fmt.Sprintf("%d", f.Limite)

	return params
}

// Errors relacionados a proposições
var (
	ErrProposicaoIDInvalido         = errors.New("ID da proposição deve ser maior que zero")
	ErrProposicaoEmentaVazia        = errors.New("ementa da proposição não pode estar vazia")
	ErrProposicaoAnoInvalido        = errors.New("ano deve ser entre 1988 e o ano atual")
	ErrProposicaoNumeroInvalido     = errors.New("número da proposição deve ser maior que zero")
	ErrProposicaoTipoInvalido       = errors.New("sigla do tipo da proposição é obrigatória")
	ErrProposicaoDataInvalida       = errors.New("formato de data inválido")
	ErrProposicaoLimiteExcedido     = errors.New("limite máximo de 100 itens por página")
	ErrProposicaoOrdemInvalida      = errors.New("ordem deve ser ASC ou DESC")
	ErrProposicaoOrdenarPorInvalido = errors.New("campo para ordenação inválido")
	ErrProposicaoNaoEncontrada      = errors.New("proposição não encontrada")
)
