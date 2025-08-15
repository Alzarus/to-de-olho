package domain

type Deputado struct {
	ID       int    `json:"id"`
	Nome     string `json:"nome"`
	Partido  string `json:"siglaPartido"`
	UF       string `json:"siglaUf"`
	URLFoto  string `json:"urlFoto"`
	Situacao string `json:"condicaoEleitoral"`
	Email    string `json:"email"`
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
