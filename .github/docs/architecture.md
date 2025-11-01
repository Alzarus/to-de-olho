# 🏗️ Arquitetura - Clean Architecture + DDD

## 📐 Visão Geral da Arquitetura

### Princípios Fundamentais
- **Clean Architecture**: Independência de frameworks, UI, banco de dados
- **Domain-Driven Design**: Modelagem baseada no domínio de negócio
- **SOLID Principles**: Aplicados em todas as camadas
- **Hexagonal Architecture**: Portas e adaptadores para isolamento

## 🏛️ Estrutura por Microsserviço

```
/backend/services/{service-name}/
├── cmd/
│   └── server/
│       └── main.go              # Entry point da aplicação
├── internal/
│   ├── domain/                  # 🎯 Camada de Domínio (Business Rules)
│   │   ├── entities/           # Entidades do negócio
│   │   ├── valueobjects/       # Value Objects
│   │   ├── aggregates/         # Agregados DDD
│   │   ├── repositories/       # Interfaces dos repositórios
│   │   └── services/           # Serviços de domínio
│   ├── application/             # 🎮 Camada de Aplicação (Use Cases)
│   │   ├── usecases/           # Casos de uso específicos
│   │   ├── dtos/               # Data Transfer Objects
│   │   └── ports/              # Interfaces para infraestrutura
│   ├── infrastructure/          # 🔧 Camada de Infraestrutura
│   │   ├── database/           # PostgreSQL, Redis
│   │   ├── http/               # Handlers HTTP/REST
│   │   ├── grpc/               # Handlers gRPC
│   │   ├── queue/              # RabbitMQ
│   │   └── external/           # APIs externas (Câmara, TSE)
│   └── interfaces/              # 🌐 Camada de Interface
│       ├── rest/               # Controllers REST
│       ├── graphql/            # Resolvers GraphQL
│       └── cli/                # Comandos CLI
├── pkg/                        # 📦 Código público reutilizável
│   ├── logger/                 # Logging estruturado
│   ├── validator/              # Validação de dados
│   └── metrics/                # Métricas e observabilidade
└── tests/                      # 🧪 Testes organizados
    ├── unit/                   # Testes unitários
    ├── integration/            # Testes de integração
    └── e2e/                    # Testes end-to-end
```

## 🎯 Camadas da Clean Architecture

### 1. Domain Layer (Núcleo)
```go
// Entidade principal
type Deputado struct {
    ID           uuid.UUID
    Nome         string
    CPF          CPF           // Value Object
    Estado       Estado        // Value Object
    Partido      Partido       // Aggregate
    Mandatos     []Mandato     // Collection
    createdAt    time.Time
    updatedAt    time.Time
}

// Interface do repositório (porta)
type DeputadoRepository interface {
    Save(ctx context.Context, deputado *Deputado) error
    FindByID(ctx context.Context, id uuid.UUID) (*Deputado, error)
    FindByEstado(ctx context.Context, uf string) ([]*Deputado, error)
}

// Serviço de domínio
type DeputadoService struct {
    repo DeputadoRepository
}

func (s *DeputadoService) ValidarElegibilidade(deputado *Deputado) error {
    // Regras de negócio específicas do domínio
    if deputado.Idade() < 21 {
        return ErrIdadeInsuficiente
    }
    return nil
}
```

### 2. Application Layer (Casos de Uso)
```go
// Caso de uso específico
type BuscarDeputadoUseCase struct {
    repo      domain.DeputadoRepository
    validator domain.DeputadoValidator
    logger    *slog.Logger
}

type BuscarDeputadoInput struct {
    ID uuid.UUID `validate:"required"`
}

type BuscarDeputadoOutput struct {
    Deputado *domain.Deputado `json:"deputado"`
    Meta     *MetaInfo        `json:"meta"`
}

func (uc *BuscarDeputadoUseCase) Execute(ctx context.Context, input BuscarDeputadoInput) (*BuscarDeputadoOutput, error) {
    // 1. Validar input
    if err := uc.validator.ValidateStruct(input); err != nil {
        return nil, fmt.Errorf("input inválido: %w", err)
    }
    
    // 2. Buscar deputado
    deputado, err := uc.repo.FindByID(ctx, input.ID)
    if err != nil {
        return nil, fmt.Errorf("erro ao buscar deputado: %w", err)
    }
    
    // 3. Montar resposta
    return &BuscarDeputadoOutput{
        Deputado: deputado,
        Meta:     &MetaInfo{Timestamp: time.Now()},
    }, nil
}
```

### 3. Infrastructure Layer (Implementações)
```go
// Implementação do repositório (adaptador)
type PostgresDeputadoRepository struct {
    db     *sql.DB
    logger *slog.Logger
}

func (r *PostgresDeputadoRepository) Save(ctx context.Context, deputado *domain.Deputado) error {
    query := `
        INSERT INTO deputados (id, nome, cpf, estado_uf, partido_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (id) DO UPDATE SET
            nome = EXCLUDED.nome,
            updated_at = EXCLUDED.updated_at
    `
    
    _, err := r.db.ExecContext(ctx, query,
        deputado.ID,
        deputado.Nome,
        deputado.CPF.String(),
        deputado.Estado.UF,
        deputado.Partido.ID,
        deputado.CreatedAt,
        time.Now(),
    )
    
    if err != nil {
        return fmt.Errorf("erro ao salvar deputado: %w", err)
    }
    
    return nil
}
```

### 4. Interface Layer (Controllers)
```go
// Handler HTTP (adaptador)
type DeputadoHandler struct {
    buscarDeputadoUC *application.BuscarDeputadoUseCase
    logger           *slog.Logger
}

func (h *DeputadoHandler) BuscarDeputado(c *gin.Context) {
    // 1. Extrair parâmetros
    idParam := c.Param("id")
    deputadoID, err := uuid.Parse(idParam)
    if err != nil {
        c.JSON(400, gin.H{"error": "ID inválido"})
        return
    }
    
    // 2. Executar caso de uso
    input := application.BuscarDeputadoInput{ID: deputadoID}
    output, err := h.buscarDeputadoUC.Execute(c.Request.Context(), input)
    if err != nil {
        h.logger.Error("erro ao buscar deputado", 
            slog.String("id", deputadoID.String()),
            slog.String("error", err.Error()))
        c.JSON(500, gin.H{"error": "Erro interno"})
        return
    }
    
    // 3. Retornar resposta
    c.JSON(200, output)
}
```

## 🔗 Dependency Injection

### Container de Dependências
```go
type ServiceContainer struct {
    // Infrastructure
    DB     *sql.DB
    Redis  *redis.Client
    Logger *slog.Logger
    
    // Repositories
    DeputadoRepo    domain.DeputadoRepository
    ProposicaoRepo  domain.ProposicaoRepository
    
    // Use Cases
    BuscarDeputadoUC  *application.BuscarDeputadoUseCase
    ListarDeputadosUC *application.ListarDeputadosUseCase
    
    // Handlers
    DeputadoHandler *interfaces.DeputadoHandler
}

func NewServiceContainer(cfg *config.Config) *ServiceContainer {
    // 1. Infrastructure layer
    db := postgresql.NewConnection(cfg.DatabaseURL)
    redis := redis.NewClient(cfg.RedisURL)
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    
    // 2. Repository layer
    deputadoRepo := infrastructure.NewPostgresDeputadoRepository(db, logger)
    
    // 3. Use case layer
    buscarDeputadoUC := application.NewBuscarDeputadoUseCase(deputadoRepo, logger)
    
    // 4. Handler layer
    deputadoHandler := interfaces.NewDeputadoHandler(buscarDeputadoUC, logger)
    
    return &ServiceContainer{
        DB:               db,
        Redis:           redis,
        Logger:          logger,
        DeputadoRepo:    deputadoRepo,
        BuscarDeputadoUC: buscarDeputadoUC,
        DeputadoHandler: deputadoHandler,
    }
}
```

## 🔄 Comunicação Entre Microsserviços

### 1. Comunicação Síncrona (gRPC)
```protobuf
// deputados.proto
syntax = "proto3";

package deputados.v1;

service DeputadosService {
    rpc GetDeputado(GetDeputadoRequest) returns (GetDeputadoResponse);
    rpc ListDeputados(ListDeputadosRequest) returns (ListDeputadosResponse);
}

message GetDeputadoRequest {
    string id = 1;
}

message GetDeputadoResponse {
    Deputado deputado = 1;
}
```

### 2. Comunicação Assíncrona (Message Queue)
```go
// Event publishing
type DeputadoCriadoEvent struct {
    DeputadoID uuid.UUID `json:"deputado_id"`
    Nome       string    `json:"nome"`
    Estado     string    `json:"estado"`
    Timestamp  time.Time `json:"timestamp"`
}

func (uc *CriarDeputadoUseCase) Execute(ctx context.Context, input CriarDeputadoInput) error {
    // 1. Criar deputado
    deputado, err := uc.criarDeputado(input)
    if err != nil {
        return err
    }
    
    // 2. Persistir
    if err := uc.repo.Save(ctx, deputado); err != nil {
        return err
    }
    
    // 3. Publicar evento
    event := DeputadoCriadoEvent{
        DeputadoID: deputado.ID,
        Nome:       deputado.Nome,
        Estado:     deputado.Estado.UF,
        Timestamp:  time.Now(),
    }
    
    return uc.eventPublisher.Publish("deputado.criado", event)
}
```

## 📊 Observabilidade

### Métricas (Prometheus)
```go
// Métricas customizadas
var (
    deputadosProcessados = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "deputados_processados_total",
            Help: "Total de deputados processados",
        },
        []string{"operacao", "status"},
    )
    
    tempoProcessamento = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "deputados_tempo_processamento_seconds",
            Help: "Tempo de processamento das operações",
        },
        []string{"operacao"},
    )
)

func (uc *BuscarDeputadoUseCase) Execute(ctx context.Context, input BuscarDeputadoInput) (*BuscarDeputadoOutput, error) {
    start := time.Now()
    defer func() {
        tempoProcessamento.WithLabelValues("buscar_deputado").Observe(time.Since(start).Seconds())
    }()
    
    // Lógica do caso de uso...
    
    deputadosProcessados.WithLabelValues("buscar", "sucesso").Inc()
    return output, nil
}
```

### Logs Estruturados
```go
// Context com trace ID
func (h *DeputadoHandler) BuscarDeputado(c *gin.Context) {
    traceID := uuid.New().String()
    ctx := context.WithValue(c.Request.Context(), "trace_id", traceID)
    
    logger := h.logger.With(
        slog.String("trace_id", traceID),
        slog.String("operation", "buscar_deputado"),
        slog.String("deputado_id", c.Param("id")),
    )
    
    logger.Info("iniciando busca de deputado")
    
    // Executar caso de uso com contexto...
    
    logger.Info("busca de deputado concluída com sucesso")
}
```

## � Melhores Práticas Go + Microsserviços (2025)

### Go (Implementação)
- **Contexto em toda a jornada**: propague `context.Context` desde os handlers até repositórios para deadlines, cancelamentos e trace IDs consistentes.
- **Interfaces pequenas e explícitas**: siga o princípio “aceite interfaces, retorne structs” evitando acoplamento acidental entre casos de uso e infraestrutura.
- **Erros enriquecidos**: envolva (`fmt.Errorf("...: %w", err)`) e classifique erros com códigos semânticos para permitir `errors.Is/As` e respostas HTTP previsíveis.
- **Logs estruturados por pares chave/valor**: utilizando `slog` ou adaptadores Go Kit para manter rastreabilidade uniforme em todos os serviços.
- **Telemetry-first**: exponha métricas customizadas e tracing distribuído diretamente nas camadas de aplicação; mantenha exporters (Prometheus, OTEL) plugáveis via ports/adapters.
- **Qualidade contínua**: aplique `gofmt`, `golangci-lint` e testes table-driven como etapa obrigatória do pipeline (">make test"), garantindo rigor antes do deploy.

### Arquitetura de Microsserviços
- **Contratos versionados**: estabeleça versionamento explícito (ex.: `/v1/deputados`) e testes de contrato para REST/gRPC antes de promover mudanças entre serviços.
- **Resiliência aplicada**: padronize políticas de retry exponencial, timeout e circuit breaker (ex.: `go-resiliency`, `hystrix-go`) encapsuladas em middlewares compartilhados.
- **Backpressure & rate limiting**: mantenha limites por consumidor usando o `pkg/ratelimiter` e combine com filas RabbitMQ para suavizar picos.
- **Comunicação orientada a eventos**: sempre que possível preferir eventos idempotentes com schemas versionados (Avro/JSON Schema) para evitar acoplamento temporal.
- **Configuração 12-factor**: centralize secrets e feature flags via config server/SSM e injete por variáveis de ambiente + configuração tipada (`internal/config`).
- **Observabilidade completa**: correlacione logs, métricas e traces utilizando IDs compartilhados e dashboards Grafana com alertas pró-ativos.
- **Entrega contínua segura**: pipelines com gates (lint, testes, security scan) e deploy canário/blue-green, reduzindo blast radius em releases frequentes.

## �🧪 Testabilidade

### Mocks e Stubs
```go
// Mock do repositório para testes
type MockDeputadoRepository struct {
    mock.Mock
}

func (m *MockDeputadoRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Deputado, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*domain.Deputado), args.Error(1)
}

// Teste do caso de uso
func TestBuscarDeputadoUseCase_Execute(t *testing.T) {
    // Arrange
    mockRepo := new(MockDeputadoRepository)
    uc := application.NewBuscarDeputadoUseCase(mockRepo, slog.Default())
    
    deputadoEsperado := &domain.Deputado{
        ID:   uuid.New(),
        Nome: "João Silva",
    }
    
    mockRepo.On("FindByID", mock.Anything, deputadoEsperado.ID).Return(deputadoEsperado, nil)
    
    // Act
    input := application.BuscarDeputadoInput{ID: deputadoEsperado.ID}
    output, err := uc.Execute(context.Background(), input)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, deputadoEsperado, output.Deputado)
    mockRepo.AssertExpectations(t)
}
```

## 🔄 Arquitetura de Ingestão de Dados

### **Componentes de Ingestão**

#### 1. **Strategic Backfill** (`cmd/ingestor`)
- **Execução**: Uma única vez no deploy inicial
- **Período**: 2022-2025 (dados históricos + atuais)
- **Estratégia**: Prioridade deputados → proposições → despesas
- **Resilência**: Checkpoints, retry exponencial, resumable

#### 2. **Incremental Sync** (`cmd/scheduler`) 
- **Execução**: Diária às 6h (cron job)
- **Objetivo**: Manter dados atualizados
- **Escopo**: Delta sync + analytics pre-computados

#### 3. **Pipeline de Dados**
```
API Câmara → Strategic Backfill → PostgreSQL → Cache Redis → Frontend
     ↓              ↓                  ↓           ↓
Rate Limit    Checkpoints     Transações    TTL 1h    Rankings
(100/min)     Resumable       ACID          Cache     Analytics
```

### **Tabelas de Dados**

#### **Deputados** (Base fundamental)
```sql
CREATE TABLE deputados_cache (
    id INT PRIMARY KEY,
    payload JSONB NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW()
);
```

#### **Proposições** (Por ano 2022-2025)
```sql  
CREATE TABLE proposicoes_cache (
    id INTEGER PRIMARY KEY,
    payload JSONB NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### **Despesas** (Analytics críticos)
```sql
CREATE TABLE despesas (
    id BIGSERIAL PRIMARY KEY,
    deputado_id INTEGER NOT NULL,
    ano INTEGER NOT NULL,
    mes INTEGER NOT NULL,
    tipo_despesa VARCHAR(100) NOT NULL,
    valor_liquido DECIMAL(15,2) NOT NULL,
    valor_bruto DECIMAL(15,2),
    fornecedor VARCHAR(255),
    documento_url TEXT,
    payload JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### **Configuração Auto-Deploy**
- **Trigger**: `docker-compose up` executa backfill automático
- **Dados**: 2022-2025 completo (4 anos de histórico)
- **Failsafe**: Checkpoints permitem resumar em caso de falha
- **Monitoramento**: Métricas em `sync_metrics` table

---

> **🎯 Benefícios**: Código testável, manutenível, escalável e independente de frameworks externos.
