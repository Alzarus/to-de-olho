# 🏛️ Integração API Câmara dos Deputados

## 📊 Mapeamento da API Oficial

### Base URL
```
https://dadosabertos.camara.leg.br/api/v2/
```

### Rate Limiting
- **Limite**: 100 requisições/minuto
- **Headers de controle**: `X-RateLimit-*`
- **Strategy**: Circuit breaker + backoff exponencial

## 📋 Endpoints Principais

### 1. Deputados

#### Listar Deputados
```http
GET /deputados?idLegislatura=57&siglaUf=SP&siglaPartido=PT&siglaSexo=M&itens=100&ordem=nome
```

**Parâmetros Oficiais:**
- `idLegislatura` (int): ID da legislatura (57 = atual)
- `siglaUf` (string): UF (AC, AL, AP, AM, BA, CE, DF, ES, GO, MA, MT, MS, MG, PA, PB, PR, PE, PI, RJ, RN, RS, RO, RR, SC, SP, SE, TO)
- `siglaPartido` (string): Sigla do partido
- `siglaSexo` (string): M ou F
- `dataInicio` (date): AAAA-MM-DD
- `dataFim` (date): AAAA-MM-DD
- `ordem` (string): id, nome, siglaUf, siglaPartido
- `ordenarPor` (string): ASC, DESC
- `itens` (int): 1-100 (default: 15)
- `pagina` (int): Página da consulta

**Response Structure:**
```json
{
  "dados": [
    {
      "id": 220595,
      "uri": "https://dadosabertos.camara.leg.br/api/v2/deputados/220595",
      "nome": "Aécio Neves",
      "siglaPartido": "PSDB",
      "uriPartido": "https://dadosabertos.camara.leg.br/api/v2/partidos/36835",
      "siglaUf": "MG",
      "idLegislatura": 57,
      "urlFoto": "https://www.camara.leg.br/internet/deputado/bandep/220595.jpg",
      "email": "dep.aecioneves@camara.leg.br"
    }
  ],
  "links": [
    {
      "rel": "self",
      "href": "https://dadosabertos.camara.leg.br/api/v2/deputados"
    },
    {
      "rel": "first",
      "href": "https://dadosabertos.camara.leg.br/api/v2/deputados?pagina=1&itens=15"
    },
    {
      "rel": "last", 
      "href": "https://dadosabertos.camara.leg.br/api/v2/deputados?pagina=35&itens=15"
    }
  ]
}
```

#### Deputado por ID
```http
GET /deputados/{id}
```

**Response:**
```json
{
  "dados": {
    "id": 220595,
    "uri": "https://dadosabertos.camara.leg.br/api/v2/deputados/220595",
    "nomeCivil": "Aécio Neves da Cunha",
    "ultimoStatus": {
      "id": 220595,
      "uri": "https://dadosabertos.camara.leg.br/api/v2/deputados/220595",
      "nome": "Aécio Neves",
      "siglaPartido": "PSDB",
      "uriPartido": "https://dadosabertos.camara.leg.br/api/v2/partidos/36835",
      "siglaUf": "MG",
      "idLegislatura": 57,
      "urlFoto": "https://www.camara.leg.br/internet/deputado/bandep/220595.jpg",
      "data": "2023-02-01",
      "nomeEleitoral": "Aécio Neves",
      "gabinete": {
        "nome": "401",
        "predio": "4",
        "sala": "401",
        "andar": "4",
        "telefone": "3215-5401",
        "email": "dep.aecioneves@camara.leg.br"
      },
      "situacao": "Exercício",
      "condicaoEleitoral": "Titular",
      "descricaoStatus": "Deputado em exercício de mandato eletivo."
    },
    "cpf": "61926477015",
    "sexo": "M",
    "urlWebsite": "",
    "redeSocial": [
      "https://www.facebook.com/aecionevesoficial",
      "https://twitter.com/AecioNeves",
      "https://www.instagram.com/aecioneves/"
    ],
    "dataNascimento": "1960-03-10",
    "dataFalecimento": null,
    "ufNascimento": "MG",
    "municipioNascimento": "Belo Horizonte",
    "escolaridade": "Superior"
  }
}
```

### 2. Despesas

#### Despesas do Deputado
```http
GET /deputados/{id}/despesas?ano=2024&mes=10&itens=100&ordem=dataDocumento
```

**Parâmetros:**
- `ano` (int): Ano da despesa
- `mes` (int): Mês da despesa (1-12)
- `cnpjCpfFornecedor` (string): CNPJ/CPF do fornecedor
- `ordem` (string): ano, mes, tipoDespesa, valorDocumento, valorGlosa, valorLiquido, dataDocumento
- `ordenarPor` (string): ASC, DESC

**Response:**
```json
{
  "dados": [
    {
      "ano": 2024,
      "mes": 10,
      "tipoDespesa": "PASSAGEM AÉREA",
      "codDocumento": 12345,
      "tipoDocumento": "Nota Fiscal",
      "codTipoDocumento": 2,
      "dataDocumento": "2024-10-15",
      "numDocumento": "123456",
      "valorDocumento": 1850.50,
      "urlDocumento": "https://www.camara.leg.br/cota-parlamentar/documentos/publ/123/2024/12345.pdf",
      "nomeFornecedor": "TAM LINHAS AEREAS S/A",
      "cnpjCpfFornecedor": "02012862000160",
      "valorLiquido": 1850.50,
      "valorGlosa": 0,
      "numRessarcimento": null,
      "codLote": 123456,
      "parcela": 1
    }
  ],
  "links": []
}
```

### 3. Proposições

#### Listar Proposições
```http
GET /proposicoes?siglaTipo=PL&numero=1234&ano=2024&dataInicio=2024-01-01&dataFim=2024-12-31
```

**Parâmetros:**
- `siglaTipo` (string): PL, PEC, PDC, etc.
- `numero` (int): Número da proposição
- `ano` (int): Ano da proposição
- `dataInicio` (date): Data início tramitação
- `dataFim` (date): Data fim tramitação
- `idSituacao` (int): ID da situação
- `keywords` (string): Palavras-chave na ementa

### 4. Votações

#### Listar Votações
```http
GET /votacoes?dataInicio=2024-01-01&dataFim=2024-12-31&idOrgao=180
```

#### Votos de uma Votação
```http
GET /votacoes/{id}/votos
```

**Response:**
```json
{
  "dados": [
    {
      "deputado_": {
        "id": 220595,
        "uri": "https://dadosabertos.camara.leg.br/api/v2/deputados/220595",
        "nome": "Aécio Neves",
        "siglaPartido": "PSDB",
        "siglaUf": "MG",
        "idLegislatura": 57
      },
      "tipoVoto": "Sim",
      "dataRegistroVoto": "2024-03-15T14:30:00"
    }
  ]
}
```

## 🔄 Mapeamento para Domínio Interno

### Deputado Entity Mapping
```go
// API da Câmara → Domínio Interno
type DeputadoAPIResponse struct {
    ID               int    `json:"id"`                // Câmara ID
    NomeCivil        string `json:"nomeCivil"`
    UltimoStatus     Status `json:"ultimoStatus"`
    CPF              string `json:"cpf"`
    Sexo             string `json:"sexo"`
    DataNascimento   string `json:"dataNascimento"`
    UFNascimento     string `json:"ufNascimento"`
    Escolaridade     string `json:"escolaridade"`
    URLWebsite       string `json:"urlWebsite"`
    RedeSocial       []string `json:"redeSocial"`
}

func (api *DeputadoAPIResponse) ToDomain() *domain.Deputado {
    return &domain.Deputado{
        ID:             uuid.New(), // Gerar UUID interno
        CamaraID:       api.ID,     // Manter referência oficial
        Nome:           api.UltimoStatus.Nome,
        NomeCivil:      api.NomeCivil,
        CPF:            domain.NewCPF(api.CPF),
        Sexo:           domain.Sexo(api.Sexo),
        DataNascimento: parseDate(api.DataNascimento),
        // ... outros mapeamentos
    }
}
```

### Despesa Entity Mapping
```go
type DespesaAPIResponse struct {
    Ano                 int     `json:"ano"`
    Mes                 int     `json:"mes"`
    TipoDespesa         string  `json:"tipoDespesa"`
    DataDocumento       string  `json:"dataDocumento"`
    ValorDocumento      float64 `json:"valorDocumento"`
    NomeFornecedor      string  `json:"nomeFornecedor"`
    CNPJCPFFornecedor   string  `json:"cnpjCpfFornecedor"`
    URLDocumento        string  `json:"urlDocumento"`
}

func (api *DespesaAPIResponse) ToDomain(deputadoID uuid.UUID) *domain.Despesa {
    return &domain.Despesa{
        ID:           uuid.New(),
        DeputadoID:   deputadoID,
        Tipo:         domain.TipoDespesa(api.TipoDespesa),
        Valor:        decimal.NewFromFloat(api.ValorDocumento),
        Data:         parseDate(api.DataDocumento),
        Fornecedor:   domain.NewFornecedor(api.NomeFornecedor, api.CNPJCPFFornecedor),
        URLDocumento: api.URLDocumento,
    }
}
```

## 🛡️ Estratégias de Resilência

### Circuit Breaker
```go
type CamaraAPIClient struct {
    client    *http.Client
    breaker   *gobreaker.CircuitBreaker
    cache     cache.Cache
    limiter   *rate.Limiter
}

func (c *CamaraAPIClient) GetDeputados(ctx context.Context, params DeputadosParams) (*DeputadosResponse, error) {
    // Rate limiting
    if err := c.limiter.Wait(ctx); err != nil {
        return nil, err
    }
    
    // Circuit breaker
    result, err := c.breaker.Execute(func() (interface{}, error) {
        return c.makeRequest(ctx, "/deputados", params)
    })
    
    if err != nil {
        // Tentar cache em caso de erro
        if cached := c.cache.Get(cacheKey); cached != nil {
            return cached.(*DeputadosResponse), nil
        }
        return nil, err
    }
    
    // Cache resultado
    c.cache.Set(cacheKey, result, 5*time.Minute)
    
    return result.(*DeputadosResponse), nil
}
```

### Retry com Backoff
```go
func (c *CamaraAPIClient) makeRequestWithRetry(ctx context.Context, endpoint string, params interface{}) (*http.Response, error) {
    backoff := &backoff.ExponentialBackOff{
        InitialInterval:     1 * time.Second,
        RandomizationFactor: 0.5,
        Multiplier:          2,
        MaxInterval:         30 * time.Second,
        MaxElapsedTime:      2 * time.Minute,
        Clock:               backoff.SystemClock,
    }
    
    operation := func() (*http.Response, error) {
        resp, err := c.client.Do(req)
        if err != nil {
            return nil, err
        }
        
        // Retry em caso de rate limit ou erro temporário
        if resp.StatusCode == 429 || resp.StatusCode >= 500 {
            return nil, fmt.Errorf("temporary error: %d", resp.StatusCode)
        }
        
        return resp, nil
    }
    
    return backoff.RetryWithData(operation, backoff)
}
```

## 🔄 ETL - Ingestão de Dados

### Pipeline de Sincronização
```go
type IngestaoService struct {
    camaraAPI     CamaraAPIClient
    deputadoRepo  domain.DeputadoRepository
    despesaRepo   domain.DespesaRepository
    eventBus      events.EventBus
}

func (s *IngestaoService) SincronizarDeputados(ctx context.Context) error {
    // 1. Buscar deputados da API
    deputadosAPI, err := s.camaraAPI.GetAllDeputados(ctx)
    if err != nil {
        return fmt.Errorf("erro ao buscar deputados: %w", err)
    }
    
    // 2. Converter para domínio
    for _, deputadoAPI := range deputadosAPI {
        deputado := deputadoAPI.ToDomain()
        
        // 3. Verificar se já existe
        existing, _ := s.deputadoRepo.FindByCamaraID(ctx, deputadoAPI.ID)
        if existing != nil {
            // Atualizar dados
            existing.AtualizarDados(deputado)
            deputado = existing
        }
        
        // 4. Salvar no banco
        if err := s.deputadoRepo.Save(ctx, deputado); err != nil {
            return fmt.Errorf("erro ao salvar deputado %d: %w", deputadoAPI.ID, err)
        }
        
        // 5. Publicar evento
        s.eventBus.Publish(events.DeputadoSincronizado{
            DeputadoID: deputado.ID,
            CamaraID:   deputadoAPI.ID,
            Timestamp:  time.Now(),
        })
    }
    
    return nil
}
```

## 📊 Monitoramento e Métricas

### Métricas da API
```go
var (
    camaraAPIRequests = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "camara_api_requests_total",
            Help: "Total de requisições para API da Câmara",
        },
        []string{"endpoint", "status"},
    )
    
    camaraAPILatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "camara_api_request_duration_seconds",
            Help: "Latência das requisições para API da Câmara",
        },
        []string{"endpoint"},
    )
    
    camaraAPIErrors = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "camara_api_errors_total",
            Help: "Total de erros na API da Câmara",
        },
        []string{"endpoint", "error_type"},
    )
)
```

## 🚀 Próximos Passos

1. **Implementar client HTTP robusto** com circuit breaker
2. **Criar jobs de sincronização** com scheduler
3. **Configurar cache Redis** para dados frequentes
4. **Implementar webhooks** para notificações de mudanças
5. **Criar dashboard** de monitoramento da ingestão
6. **Testes de integração** com API sandbox
