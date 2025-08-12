# Instruções do GitHub Copilot - Projeto "Tô De Olho"

## 🎯 Visão do Projeto

O **"Tô De Olho"** é uma plataforma de transparência política que visa democratizar o acesso aos dados da Câmara dos Deputados, promovendo maior engajamento cidadão através de três núcleos fundamentais:

1. **Acessibilidade do Aplicativo**: Interface intuitiva e fácil acesso para todos os usuários
2. **Gestão Social**: Capacidade de participação cidadã nas decisões públicas
3. **Potencial de Ludificação**: Estratégias de gamificação para elevar o interesse pela gestão pública

### Características do Sistema

- **Linguagem oficial**: Português Brasileiro (pt-BR)
- **Dados oficiais**: API da Câmara dos Deputados + TSE
- **Interação cidadã**: Fórum e contato direto deputado-cidadão
- **Gamificação**: Sistema de pontos, conquistas e rankings

````

## 📊 Inteligência e Analytics Avançados

### Dashboard Interativo por Região

#### Visualizações Geográficas
- **Mapa do Brasil**: Visualização de dados por estado/região
- **Heatmap de Atividade**: Regiões mais/menos engajadas
- **Comparativos Regionais**: Performance parlamentar por área
- **Índice de Transparência**: Score por estado e deputado

#### Métricas Regionalizadas
```sql
-- Exemplo de view para métricas regionais
CREATE MATERIALIZED VIEW metricas_regionais AS
SELECT
    d.sigla_uf as estado,
    d.regiao,
    COUNT(d.id) as total_deputados,
    AVG(e.taxa_presenca) as presenca_media,
    SUM(desp.valor_total) as gastos_totais,
    COUNT(prop.id) as proposicoes_total,
    COUNT(v.id) as votacoes_participadas
FROM deputados d
LEFT JOIN estatisticas_deputado e ON d.id = e.deputado_id
LEFT JOIN despesas desp ON d.id = desp.deputado_id
LEFT JOIN proposicoes prop ON d.id = prop.autor_id
LEFT JOIN votos v ON d.id = v.deputado_id
WHERE d.ativo = true
GROUP BY d.sigla_uf, d.regiao;
````

### Sistema de Alertas Inteligentes

#### Alertas Automáticos

- **Gastos Suspeitos**: Despesas acima da média ou padrões anômalos
- **Mudança de Posição**: Deputado vota contra histórico
- **Baixa Presença**: Faltas excessivas em votações importantes
- **Nova Proposição**: Projetos que impactam sua região

#### Notificações Personalizadas

- **Por Interesse**: Temas específicos (educação, saúde, economia)
- **Por Região**: Apenas deputados da sua área
- **Por Deputado**: Acompanhar parlamentares específicos
- **Por Tipo**: Escolher tipos de atividade (votações, gastos, proposições)

## 🤖 Inteligência Artificial Generativa (Gemini SDK/MCP)

### Moderação de Conteúdo e Ética

#### Sistema de Moderação Automatizada

- **Filtro Anti-Toxicidade**: Detecção de discurso de ódio, ofensas e linguagem inadequada
- **Validação Ética**: Análise de conformidade com diretrizes de convivência democrática
- **Classificação de Sentimento**: Identificação de tom agressivo ou desrespeitoso
- **Detecção de Spam**: Identificação de conteúdo repetitivo ou malicioso

```go
// Exemplo de integração com Gemini para moderação
type ModerationService struct {
    geminiClient *genai.Client
    logger       *slog.Logger
}

type ModerationResult struct {
    IsApproved      bool                 `json:"is_approved"`
    ConfidenceScore float64              `json:"confidence_score"`
    Violations      []ViolationType      `json:"violations"`
    SuggestedEdit   string               `json:"suggested_edit,omitempty"`
    Reasoning       string               `json:"reasoning"`
}

type ViolationType string

const (
    ViolationToxicity       ViolationType = "toxicity"
    ViolationHateSpeech     ViolationType = "hate_speech"
    ViolationMisinformation ViolationType = "misinformation"
    ViolationSpam           ViolationType = "spam"
    ViolationOffTopic       ViolationType = "off_topic"
)
```

#### Funcionalidades de Moderação Inteligente

##### Análise em Tempo Real

- **Pré-moderação**: Análise antes da publicação de posts/comentários
- **Moderação Contínua**: Revisão de conteúdo já publicado
- **Escalação Automática**: Envio para moderação humana em casos duvidosos
- **Sugestões de Melhoria**: Propostas de reformulação para textos problemáticos

##### Sistema de Pontuação Ética

```sql
-- Tabela para tracking de comportamento dos usuários
CREATE TABLE usuario_comportamento (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID NOT NULL REFERENCES usuarios(id),
    score_civilidade DECIMAL(3,2) DEFAULT 5.00, -- 0.00 a 10.00
    total_posts INTEGER DEFAULT 0,
    posts_aprovados INTEGER DEFAULT 0,
    posts_rejeitados INTEGER DEFAULT 0,
    warnings_recebidos INTEGER DEFAULT 0,
    ultimo_warning TIMESTAMP,
    status_conta TEXT DEFAULT 'ativo', -- ativo, advertido, suspenso, banido
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### Assistente IA para Engajamento Cívico

#### Chatbot Educativo

- **Explicação de Termos**: Glossário político interativo
- **Orientação Cívica**: Como participar do processo democrático
- **Análise de Proposições**: Resumos simplificados de projetos de lei complexos
- **Comparação de Deputados**: Análises imparciais de performance parlamentar

#### Geração de Conteúdo Educativo

- **Resumos Automáticos**: Sínteses de sessões parlamentares e votações importantes
- **Relatórios Personalizados**: Análises específicas por região ou interesse
- **Explicações Contextuais**: Histórico e impacto de decisões políticas
- **Fact-Checking**: Verificação automática de informações políticas

```go
// Serviço de assistente IA educativo
// https://github.com/googleapis/go-genai
type EducationalAssistant struct {
    geminiClient  *genai.Client
    knowledgeBase *KnowledgeBaseService
    userProfile   *UserProfileService
}

func (ea *EducationalAssistant) ExplainProposition(ctx context.Context,
    propositionID uuid.UUID, userID uuid.UUID) (*ExplanationResponse, error) {

    // Buscar dados da proposição
    proposition, err := ea.knowledgeBase.GetProposition(ctx, propositionID)
    if err != nil {
        return nil, err
    }

    // Obter perfil do usuário para personalização
    profile, err := ea.userProfile.GetProfile(ctx, userID)
    if err != nil {
        return nil, err
    }

    // Gerar explicação personalizada via Gemini
    prompt := fmt.Sprintf(`
        Explique de forma simples e imparcial a proposição "%s" para um cidadão brasileiro.
        Nível de conhecimento político: %s
        Região de interesse: %s
        Área de atuação: %s

        Proposição: %s

        Forneça:
        1. Resumo em linguagem acessível
        2. Possíveis impactos práticos
        3. Argumentos pró e contra
        4. Relevância para a região do usuário
    `, proposition.Title, profile.PoliticalKnowledge,
       profile.Region, profile.Profession, proposition.Content)

    return ea.generateResponse(ctx, prompt)
}
```

### Ferramentas de Comparação

#### Comparativo de Deputados

- **Performance**: Presença, produtividade, gastos
- **Posicionamento**: Histórico de votações por tema
- **Evolução Temporal**: Mudanças ao longo do mandato
- **Ranking**: Posição entre pares da mesma região/partido

#### Análise Preditiva

- **Tendências de Voto**: Previsão baseada em histórico
- **Padrões de Gasto**: Projeção de despesas
- **Engajamento**: Previsão de participação em votações
- **Risco de Escândalo**: Identificação de padrões suspeitos

## 🤝 Funcionalidades Sociais Avançadas

### Networking Político

#### Grupos de Interesse

- **Por Tema**: Educação, saúde, meio ambiente, economia
- **Por Região**: Grupos estaduais e municipais
- **Por Idade**: Jovens, adultos, idosos
- **Por Profissão**: Professores, médicos, empresários

#### Eventos e Mobilização

- **Eventos Locais**: Encontros presenciais organizados via plataforma
- **Campanhas**: Mobilização para causas específicas
- **Petições**: Abaixo-assinados digitais com validação TSE
- **Transmissões**: Lives com deputados e especialistas

### Sistema de Mentoria Política

#### Educação Cívica

- **Cursos Interativos**: Como funciona o Congresso
- **Glossário Político**: Termos técnicos explicados de forma simples
- **Simuladores**: Como criar uma lei, processo legislativo
- **Quiz Educativo**: Gamificação do aprendizado político

#### Mentores Verificados

- **Especialistas**: Cientistas políticos, juristas
- **Ex-parlamentares**: Experiência prática
- **Jornalistas**: Cobertura política especializada
- **Ativistas**: Experiência em movimentos sociais

## 🛠️ Padrões de Desenvolvimento

### Stack Tecnológico

```
Backend:     Go 1.23+ (Gin framework)
Frontend:    Next.js 15 + TypeScript + Tailwind CSS
Database:    PostgreSQL 16 + Redis (cache)
Queue:       RabbitMQ (mensageria assíncrona)
AI/ML:       Google Gemini SDK + MCP (Model Context Protocol)
Monitoring:  Prometheus + Grafana
Security:    JWT + OAuth2 + Rate Limiting
```

### 🏗️ Arquitetura e Clean Code (2025-2026)

#### Princípios de Clean Architecture

##### Domain-Driven Design (DDD)
```go
// Estrutura baseada em domínios de negócio
/backend/services/deputados/
├── cmd/server/                  # Entry points
├── internal/
│   ├── domain/                  # Entities, Value Objects, Aggregates
│   │   ├── deputado.go         # Entity principal
│   │   ├── despesa.go          # Value Object
│   │   └── repository.go       # Interface do repositório
│   ├── application/             # Use Cases / Application Services
│   │   ├── usecases/           # Casos de uso do negócio
│   │   └── services/           # Serviços de aplicação
│   ├── infrastructure/          # Frameworks & Drivers
│   │   ├── repository/         # Implementação do repositório
│   │   ├── http/              # Handlers HTTP
│   │   └── grpc/              # Handlers gRPC
│   └── interfaces/             # Interface Adapters
├── pkg/                        # Código compartilhado público
└── tests/                      # Testes organizados por tipo
    ├── unit/                   # Testes unitários
    ├── integration/            # Testes de integração
    └── e2e/                    # Testes end-to-end
```

##### Dependency Injection & Inversion
```go
// Container de dependências
type ServiceContainer struct {
    // Repositories
    DeputadoRepo    domain.DeputadoRepository
    DespesaRepo     domain.DespesaRepository
    
    // Use Cases
    BuscarDeputado  *usecases.BuscarDeputadoUseCase
    ListarDeputados *usecases.ListarDeputadosUseCase
    
    // External Services
    CamaraAPI      *camara.Client
    EmailService   *email.Service
}

// Dependency injection via interfaces
func NewServiceContainer(cfg *config.Config) *ServiceContainer {
    // Repository layer
    deputadoRepo := postgres.NewDeputadoRepository(cfg.DB)
    despesaRepo := postgres.NewDespesaRepository(cfg.DB)
    
    // Use case layer
    buscarDeputado := usecases.NewBuscarDeputadoUseCase(deputadoRepo)
    
    return &ServiceContainer{
        DeputadoRepo:    deputadoRepo,
        BuscarDeputado:  buscarDeputado,
    }
}
```

#### SOLID Principles Implementation

##### Single Responsibility
```go
// ❌ Violação - classe fazendo muita coisa
type DeputadoService struct {
    // db operations, http calls, validation, logging...
}

// ✅ Responsabilidade única
type DeputadoRepository interface {
    Save(ctx context.Context, deputado *domain.Deputado) error
    FindByID(ctx context.Context, id uuid.UUID) (*domain.Deputado, error)
}

type DeputadoValidator interface {
    Validate(deputado *domain.Deputado) error
}

type DeputadoUseCase struct {
    repo      DeputadoRepository
    validator DeputadoValidator
    logger    *slog.Logger
}
```

##### Open/Closed Principle
```go
// Extensível sem modificação
type NotificationSender interface {
    Send(ctx context.Context, notification *Notification) error
}

// Implementações específicas
type EmailNotificationSender struct{}
type SMSNotificationSender struct{}
type PushNotificationSender struct{}

// Fácil adição de novos tipos sem alterar código existente
```

##### Interface Segregation
```go
// ❌ Interface "fat" - viola ISP
type DeputadoService interface {
    Create(deputado *Deputado) error
    Update(deputado *Deputado) error
    Delete(id uuid.UUID) error
    FindByID(id uuid.UUID) (*Deputado, error)
    FindAll() ([]*Deputado, error)
    SendEmail(email string) error
    GenerateReport() (*Report, error)
}

// ✅ Interfaces específicas e coesas
type DeputadoReader interface {
    FindByID(ctx context.Context, id uuid.UUID) (*Deputado, error)
    FindAll(ctx context.Context, filters *Filters) ([]*Deputado, error)
}

type DeputadoWriter interface {
    Save(ctx context.Context, deputado *Deputado) error
    Delete(ctx context.Context, id uuid.UUID) error
}

type DeputadoNotifier interface {
    SendEmail(ctx context.Context, email string) error
}
```

#### Clean Code Standards

##### Naming Conventions
```go
// ✅ Nomes expressivos e intencionais
func CalcularMediaGastosMensaisDeputado(gastos []domain.Despesa) decimal.Decimal {
    // Função faz exatamente o que o nome diz
}

// ✅ Constantes bem nomeadas
const (
    MaximoTentativasRequisicaoAPI = 3
    TimeoutPadraoHTTP            = 30 * time.Second
    LimiteDeputadosPorPagina     = 20
)

// ✅ Variáveis descritivas
var (
    ErrDeputadoNaoEncontrado     = errors.New("deputado não encontrado")
    ErrDadosDeputadoInvalidos    = errors.New("dados do deputado são inválidos")
    ErrPermissaoInsuficiente     = errors.New("usuário não tem permissão para esta operação")
)
```

##### Function Design
```go
// ✅ Funções pequenas com responsabilidade única
func ValidarCPFDeputado(cpf string) error {
    if len(cpf) != 11 {
        return ErrCPFTamanhoInvalido
    }
    
    if !regexp.MustCompile(`^\d{11}$`).MatchString(cpf) {
        return ErrCPFFormatoInvalido
    }
    
    return validarDigitosVerificadoresCPF(cpf)
}

// ✅ Evitar muitos parâmetros - usar structs
type CriarDeputadoParams struct {
    Nome            string    `json:"nome" validate:"required,min=2,max=100"`
    CPF             string    `json:"cpf" validate:"required,cpf"`
    DataNascimento  time.Time `json:"data_nascimento" validate:"required"`
    PartidoID       uuid.UUID `json:"partido_id" validate:"required"`
    EstadoUF        string    `json:"estado_uf" validate:"required,len=2"`
}

func CriarDeputado(ctx context.Context, params CriarDeputadoParams) (*domain.Deputado, error) {
    // Implementação focada e clara
}
```

##### Error Handling
```go
// ✅ Errors customizados com contexto
type DeputadoError struct {
    Op     string    // Operação que falhou
    ID     uuid.UUID // ID do deputado (se aplicável)
    Err    error     // Erro original
    Code   string    // Código do erro para client
}

func (e *DeputadoError) Error() string {
    return fmt.Sprintf("operação %s falhou para deputado %s: %v", e.Op, e.ID, e.Err)
}

func (e *DeputadoError) Unwrap() error {
    return e.Err
}

// ✅ Error wrapping com contexto
func (r *deputadoRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Deputado, error) {
    query := `SELECT id, nome, cpf, partido_id FROM deputados WHERE id = $1`
    
    var d domain.Deputado
    err := r.db.QueryRowContext(ctx, query, id).Scan(&d.ID, &d.Nome, &d.CPF, &d.PartidoID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, &DeputadoError{
                Op:   "FindByID",
                ID:   id,
                Err:  ErrDeputadoNaoEncontrado,
                Code: "DEPUTADO_NOT_FOUND",
            }
        }
        
        return nil, fmt.Errorf("erro ao buscar deputado %s: %w", id, err)
    }
    
    return &d, nil
}
```

### 🧪 Qualidade e Testes (Test-Driven Development)

#### Estratégia de Testing Pyramid

```
                 🔺 E2E Tests (5%)
               /              \
             🔺 Integration Tests (15%)
           /                        \
         🔺 Unit Tests (80%)
```

##### Unit Tests - Base da Pirâmide
```go
// Testes unitários com table-driven tests
func TestDeputadoValidator_Validate(t *testing.T) {
    tests := []struct {
        name      string
        deputado  *domain.Deputado
        wantError bool
        errorCode string
    }{
        {
            name: "deputado válido",
            deputado: &domain.Deputado{
                Nome:     "João Silva",
                CPF:      "12345678901",
                EstadoUF: "SP",
            },
            wantError: false,
        },
        {
            name: "CPF inválido",
            deputado: &domain.Deputado{
                Nome:     "João Silva",
                CPF:      "123",
                EstadoUF: "SP",
            },
            wantError: true,
            errorCode: "INVALID_CPF",
        },
    }
    
    validator := NewDeputadoValidator()
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.Validate(tt.deputado)
            
            if tt.wantError {
                assert.Error(t, err)
                var deputadoErr *DeputadoError
                assert.True(t, errors.As(err, &deputadoErr))
                assert.Equal(t, tt.errorCode, deputadoErr.Code)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

##### Integration Tests
```go
// Testes de integração com testcontainers
func TestDeputadoRepository_Integration(t *testing.T) {
    // Setup do container PostgreSQL para testes
    ctx := context.Background()
    
    postgresContainer, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:16"),
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
    )
    require.NoError(t, err)
    defer postgresContainer.Terminate(ctx)
    
    // Configurar migrations
    db := setupTestDatabase(t, postgresContainer)
    repo := postgres.NewDeputadoRepository(db)
    
    t.Run("deve salvar e recuperar deputado", func(t *testing.T) {
        deputado := &domain.Deputado{
            ID:       uuid.New(),
            Nome:     "Test Deputado",
            CPF:      "12345678901",
            EstadoUF: "SP",
        }
        
        err := repo.Save(ctx, deputado)
        assert.NoError(t, err)
        
        retrieved, err := repo.FindByID(ctx, deputado.ID)
        assert.NoError(t, err)
        assert.Equal(t, deputado.Nome, retrieved.Nome)
    })
}
```

##### E2E Tests
```go
// Testes end-to-end simulando cenários reais
func TestDeputadoAPI_E2E(t *testing.T) {
    // Setup completo da aplicação para testes
    app := setupTestApplication(t)
    defer app.Cleanup()
    
    client := app.HTTPClient()
    
    t.Run("jornada completa do usuário", func(t *testing.T) {
        // 1. Listar deputados (sem auth)
        resp, err := client.Get("/api/v1/deputados")
        assert.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)
        
        // 2. Fazer login como eleitor
        token := loginAsEleitor(t, client)
        
        // 3. Buscar deputado específico
        resp, err = client.Get("/api/v1/deputados/123", 
            withAuthHeader(token))
        assert.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)
        
        // 4. Comentar no perfil do deputado
        comment := map[string]string{
            "conteudo": "Excelente trabalho na comissão!",
        }
        resp, err = client.Post("/api/v1/deputados/123/comentarios", 
            comment, withAuthHeader(token))
        assert.NoError(t, err)
        assert.Equal(t, http.StatusCreated, resp.StatusCode)
    })
}
```

#### Test Utilities e Helpers
```go
// Factory para criação de dados de teste
type DeputadoFactory struct{}

func (f *DeputadoFactory) Build() *domain.Deputado {
    return &domain.Deputado{
        ID:             uuid.New(),
        Nome:           "Deputado Test",
        CPF:            f.generateValidCPF(),
        EstadoUF:       "SP",
        DataNascimento: time.Now().AddDate(-45, 0, 0),
        Status:         domain.StatusAtivo,
    }
}

func (f *DeputadoFactory) WithNome(nome string) *domain.Deputado {
    d := f.Build()
    d.Nome = nome
    return d
}

// Mocks com interface
type MockDeputadoRepository struct {
    mock.Mock
}

func (m *MockDeputadoRepository) Save(ctx context.Context, deputado *domain.Deputado) error {
    args := m.Called(ctx, deputado)
    return args.Error(0)
}

func (m *MockDeputadoRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Deputado, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*domain.Deputado), args.Error(1)
}
```

### 🚀 CI/CD Pipeline (GitHub Actions)

#### Workflow Principal
```yaml
# .github/workflows/ci-cd.yml
name: 🏛️ Tô De Olho - CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

env:
  GO_VERSION: "1.23"
  NODE_VERSION: "20"
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  # 🧪 Testes e Qualidade
  test:
    name: 🧪 Tests & Quality
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: testdb
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
      
      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379
    
    steps:
      - name: 📥 Checkout Code
        uses: actions/checkout@v4
        
      - name: 🔧 Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true
          
      - name: 📦 Download Dependencies
        run: go mod download
        
      - name: 🔍 Go Vet
        run: go vet ./...
        
      - name: 🧹 Go Fmt Check
        run: |
          if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
            echo "❌ Code is not formatted. Run 'gofmt -s -w .'"
            gofmt -s -l .
            exit 1
          fi
          
      - name: 🔒 Security Scan (gosec)
        uses: securecodewarrior/github-action-gosec@master
        with:
          args: './...'
          
      - name: 📊 Static Analysis (staticcheck)
        uses: dominikh/staticcheck-action@v1.3.0
        with:
          version: "2023.1.6"
          
      - name: 🧪 Unit Tests
        run: |
          go test -race -coverprofile=coverage.out -covermode=atomic ./...
          go tool cover -html=coverage.out -o coverage.html
          
      - name: 📈 Upload Coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
          flags: unittests
          
      - name: 🔧 Integration Tests
        run: go test -tags=integration ./tests/integration/...
        env:
          DATABASE_URL: postgres://postgres:postgres@localhost:5432/testdb?sslmode=disable
          REDIS_URL: redis://localhost:6379
          
      - name: 📱 Frontend Tests
        working-directory: ./frontend
        run: |
          npm ci
          npm run lint
          npm run type-check
          npm run test:coverage

  # 🏗️ Build e Security
  build:
    name: 🏗️ Build & Security
    runs-on: ubuntu-latest
    needs: test
    
    steps:
      - name: 📥 Checkout
        uses: actions/checkout@v4
        
      - name: 🔧 Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ env.GO_VERSION }}
          
      - name: 🏗️ Build Backend Services
        run: |
          # Build all microservices
          for service in deputados atividades despesas forum usuarios ingestao; do
            echo "Building $service service..."
            CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
              -ldflags '-extldflags "-static"' \
              -o ./bin/$service ./backend/services/$service/cmd/server
          done
          
      - name: 🔒 Vulnerability Scan (Trivy)
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          format: 'sarif'
          output: 'trivy-results.sarif'
          
      - name: 📤 Upload Trivy Results
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: 'trivy-results.sarif'

  # 🐳 Docker Build
  docker:
    name: 🐳 Docker Build & Push
    runs-on: ubuntu-latest
    needs: [test, build]
    if: github.event_name == 'push'
    
    steps:
      - name: 📥 Checkout
        uses: actions/checkout@v4
        
      - name: 🔑 Log in to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
          
      - name: 📋 Extract Metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=sha,prefix={{branch}}-
            
      - name: 🏗️ Build and Push
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  # 🚀 Deploy
  deploy:
    name: 🚀 Deploy to Staging
    runs-on: ubuntu-latest
    needs: [docker]
    if: github.ref == 'refs/heads/develop'
    environment: staging
    
    steps:
      - name: 📥 Checkout
        uses: actions/checkout@v4
        
      - name: ⚙️ Configure kubectl
        uses: azure/setup-kubectl@v3
        
      - name: 🚀 Deploy to Staging
        run: |
          # Deploy usando Helm ou kubectl
          kubectl set image deployment/to-de-olho-api \
            api=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}
          kubectl rollout status deployment/to-de-olho-api
          
      - name: 🧪 Health Check
        run: |
          # Verificar se a aplicação está respondendo
          curl -f http://staging.to-de-olho.com/health || exit 1
          
      - name: 🔔 Notify Success
        uses: 8398a7/action-slack@v3
        with:
          status: success
          text: "✅ Deploy para staging realizado com sucesso!"
        env:
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}
```

#### Quality Gates
```yaml
# .github/workflows/quality-gates.yml
name: 🛡️ Quality Gates

on:
  pull_request:
    branches: [main]

jobs:
  quality-check:
    name: 🛡️ Quality Gates
    runs-on: ubuntu-latest
    
    steps:
      - name: 📥 Checkout
        uses: actions/checkout@v4
        
      - name: 🔍 Code Coverage Check
        run: |
          coverage=$(go test -coverprofile=coverage.out ./... | grep "coverage:" | awk '{print $2}' | sed 's/%//')
          if (( $(echo "$coverage < 80" | bc -l) )); then
            echo "❌ Coverage ($coverage%) is below 80% threshold"
            exit 1
          fi
          echo "✅ Coverage: $coverage%"
          
      - name: 🔒 Security Score
        run: |
          # Scan with multiple tools and aggregate score
          gosec -quiet -fmt json -out gosec-report.json ./... || true
          # Parse and fail if critical issues found
          
      - name: 📊 Complexity Check
        run: |
          # Check cyclomatic complexity
          gocyclo -over 10 .
          
      - name: 🧹 Code Quality (SonarCloud)
        uses: SonarSource/sonarcloud-github-action@master
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
```

### 📋 Definition of Done (DoD)

#### Critérios para Finalização de Features
- [ ] **Código**: Seguir padrões de Clean Code e SOLID
- [ ] **Testes**: Cobertura mínima de 80% (unit + integration)
- [ ] **Documentação**: README atualizado e comentários no código
- [ ] **Performance**: Benchmarks dentro dos SLAs definidos
- [ ] **Security**: Scan de segurança sem vulnerabilidades críticas
- [ ] **Accessibility**: Conformidade WCAG 2.1 AA
- [ ] **Review**: Aprovação de pelo menos 2 desenvolvedores
- [ ] **CI/CD**: Pipeline passando em todos os stages

#### Code Review Checklist
```markdown
## 🔍 Code Review Checklist

### Arquitetura & Design
- [ ] Seguindo princípios SOLID
- [ ] Dependency injection implementada corretamente
- [ ] Interfaces bem definidas e coesas
- [ ] Separação clara de responsabilidades

### Qualidade do Código
- [ ] Nomes expressivos e intencionais
- [ ] Funções pequenas e focadas
- [ ] Tratamento adequado de erros
- [ ] Logs estruturados implementados

### Testes
- [ ] Testes unitários para business logic
- [ ] Testes de integração para APIs
- [ ] Mocks utilizados adequadamente
- [ ] Cobertura de casos de erro

### Performance & Security
- [ ] Queries otimizadas (sem N+1)
- [ ] Rate limiting implementado
- [ ] Validação de inputs
- [ ] Logs não exposem dados sensíveis

### Frontend (quando aplicável)
- [ ] Componentes reutilizáveis
- [ ] Accessibility attributes
- [ ] Error boundaries implementadas
- [ ] Loading states definidos
```

### 🏗️ Microsserviços

```
📋 deputados-service    → Gestão de parlamentares e perfis públicos
🗳️  atividades-service  → Proposições, votações, presença parlamentar
💰 despesas-service     → Análise de gastos e cota parlamentar
👥 usuarios-service     → Autenticação, perfis e gamificação
💬 forum-service        → Discussões cidadãs e interação deputado-público
🗳️ plebiscitos-service  → Sistema de votações e consultas populares
🔄 ingestao-service     → ETL dados Câmara/TSE (background jobs)
📊 analytics-service    → Métricas, rankings e insights regionais
🔍 search-service       → Busca inteligente de dados
🚨 alertas-service      → Notificações e alertas automáticos
🤖 ia-service          → Moderação, assistente educativo e análise preditiva
```

### Comunicação

- **API Gateway**: Ponto único de entrada com rate limiting
- **gRPC**: Comunicação interna entre microsserviços
- **Message Queue**: Processamento assíncrono de dados
- **WebSockets**: Notificações em tempo real
- **REST API**: Interface pública para frontend

## 📡 Dados da Câmara dos Deputados

### API Oficial: https://dadosabertos.camara.leg.br/api/v2/
**Versão**: 0.4.255 (Julho 2025) | **Limite**: 100 itens por requisição | **Padrão**: 15 itens

### 👥 Endpoints de Deputados

#### Dados Principais
- `GET /deputados` - Lista deputados com filtros avançados
  - Parâmetros: `idLegislatura`, `siglaUf`, `siglaPartido`, `siglaSexo`, `dataInicio`, `dataFim`
  - Retorna apenas deputados em exercício se não especificar tempo
- `GET /deputados/{id}` - Dados cadastrais completos do parlamentar

#### Atividades Parlamentares
- `GET /deputados/{id}/despesas` - **Cota parlamentar detalhada**
  - Filtros: mês, ano, legislatura, CNPJ/CPF fornecedor
  - Padrão: últimos 6 meses se não especificado
- `GET /deputados/{id}/discursos` - Pronunciamentos registrados
  - Padrão: últimos 7 dias se não especificado
- `GET /deputados/{id}/eventos` - Participação em eventos
  - Padrão: 5 dias (2 antes, 2 depois da requisição)
- `GET /deputados/{id}/orgaos` - **Comissões e cargos ocupados**
  - Inclui: presidente, vice-presidente, titular, suplente
  - Períodos de início e fim de ocupação

#### Histórico e Carreira
- `GET /deputados/{id}/historico` - **Mudanças no exercício parlamentar**
  - Mudanças de partido, nome parlamentar, licenças, afastamentos
- `GET /deputados/{id}/mandatosExternos` - Outros cargos eletivos (TSE)
- `GET /deputados/{id}/ocupacoes` - Atividades profissionais declaradas
- `GET /deputados/{id}/profissoes` - Formação e experiência profissional
- `GET /deputados/{id}/frentes` - Frentes parlamentares como membro

### 📜 Endpoints de Proposições

#### Gestão de Proposições
- `GET /proposicoes` - **Lista configurável de proposições**
  - Padrão: proposições dos últimos 30 dias
  - Filtros: `id`, `ano`, `dataApresentacaoInicio/Fim`, `idAutor`, `autor`
- `GET /proposicoes/{id}` - Detalhes completos da proposição
- `GET /proposicoes/{id}/autores` - **Autores e apoiadores**
  - Inclui: deputados, senadores, sociedade civil, outros poderes
- `GET /proposicoes/{id}/relacionadas` - Proposições relacionadas
- `GET /proposicoes/{id}/temas` - **Áreas temáticas oficiais**
- `GET /proposicoes/{id}/tramitacoes` - **Histórico completo de tramitação**
- `GET /proposicoes/{id}/votacoes` - Votações relacionadas

### 🗳️ Endpoints de Votações

#### Sistema de Votações
- `GET /votacoes` - Lista de votações
  - Padrão: últimos 30 dias, limitado ao mesmo ano
  - Filtros: órgãos, proposições, eventos
- `GET /votacoes/{id}` - Detalhes da votação específica
- `GET /votacoes/{id}/votos` - **Votos individuais dos deputados**
- `GET /votacoes/{id}/orientacoes` - **Orientação dos partidos/blocos**

### 📅 Endpoints de Eventos

#### Eventos e Reuniões
- `GET /eventos` - **Lista de eventos legislativos**
  - Padrão: 5 dias anteriores + 5 posteriores + hoje
  - Tipos: audiências públicas, reuniões, palestras
- `GET /eventos/{id}` - Detalhes do evento específico
- `GET /eventos/{id}/deputados` - **Participantes/presença**
- `GET /eventos/{id}/orgaos` - Órgãos organizadores
- `GET /eventos/{id}/pauta` - **Pauta deliberativa**
- `GET /eventos/{id}/votacoes` - Votações realizadas no evento

### 🏛️ Endpoints de Órgãos

#### Estrutura Organizacional
- `GET /orgaos` - **Comissões e órgãos legislativos**
  - Filtros: tipo, sigla, situação, período ativo
- `GET /orgaos/{id}` - Informações detalhadas do órgão
- `GET /orgaos/{id}/eventos` - Eventos realizados pelo órgão
- `GET /orgaos/{id}/membros` - **Membros e cargos ocupados**
- `GET /orgaos/{id}/votacoes` - Votações realizadas pelo órgão

### 🎭 Endpoints de Partidos e Blocos

#### Organizações Partidárias
- `GET /partidos` - **Partidos com representação na Câmara**
  - Filtros: legislatura, data, sigla
- `GET /partidos/{id}` - Detalhes do partido
- `GET /partidos/{id}/lideres` - **Líderes e vice-líderes**
- `GET /partidos/{id}/membros` - Deputados filiados

#### Blocos Partidários
- `GET /blocos` - **Blocos partidários ativos**
  - Existem apenas durante a legislatura de criação
- `GET /blocos/{id}` - Detalhes do bloco
- `GET /blocos/{id}/partidos` - Partidos integrantes

### 👥 Endpoints de Frentes e Grupos

#### Agrupamentos Temáticos
- `GET /frentes` - **Frentes parlamentares**
  - Agrupamentos oficiais por tema/proposta
  - Padrão: desde 2003 se não especificar legislatura
- `GET /frentes/{id}` - Detalhes da frente
- `GET /frentes/{id}/membros` - **Deputados participantes e papéis**

#### Cooperação Internacional
- `GET /grupos` - **Grupos interparlamentares**
  - Cooperação com parlamentares de outros países
- `GET /grupos/{id}` - Detalhes do grupo
- `GET /grupos/{id}/historico` - Variações ao longo do tempo
- `GET /grupos/{id}/membros` - Parlamentares integrantes

### 🏛️ Endpoints de Legislaturas

#### Períodos Parlamentares
- `GET /legislaturas` - **Períodos de mandatos parlamentares**
  - Identificadores sequenciais desde a primeira legislatura
- `GET /legislaturas/{id}` - Informações da legislatura específica
- `GET /legislaturas/{id}/lideres` - **Líderes da legislatura**
- `GET /legislaturas/{id}/mesa` - **Mesa Diretora da legislatura**

### 📚 Endpoints de Referências

#### Valores Válidos para Parâmetros
- `GET /referencias/deputados` - Todos os parâmetros válidos para deputados
- `GET /referencias/deputados/codSituacao` - **Situações parlamentares**
- `GET /referencias/deputados/siglaUF` - Estados e DF
- `GET /referencias/deputados/tipoDespesa` - **Tipos de cota parlamentar**
- `GET /referencias/proposicoes/siglaTipo` - **Tipos de proposições**
- `GET /referencias/proposicoes/codSituacao` - **Situações de tramitação**
- `GET /referencias/eventos/codTipoEvento` - **Tipos de eventos**
- `GET /referencias/orgaos/codTipoOrgao` - **Tipos de órgãos**

### Dados Essenciais para o Sistema

#### 1. Perfil Parlamentar

- Dados pessoais e mandato atual
- Histórico de mandatos e mudanças
- Formação acadêmica e profissional
- Comissões e cargos ocupados

#### 2. Performance Parlamentar

- **Presença**: Participação em sessões e eventos
- **Produtividade**: Proposições apresentadas e relatadas
- **Engajamento**: Discursos e pronunciamentos
- **Gastos**: Uso da cota parlamentar por categoria

#### 3. Posicionamento Político

- Histórico de votações por tema
- Alinhamento com partido/bloco
- Proposições de autoria
- Participação em frentes parlamentares

#### 4. Transparência Financeira

- Detalhamento de despesas por mês/ano
- Fornecedores mais utilizados
- Comparativo com outros deputados
- Evolução temporal dos gastos

## � Sistema de Usuários e Roles

### Tipos de Usuário

```go
const (
    RolePublico     = "publico"         // Acesso básico de leitura
    RoleEleitor     = "eleitor"         // Validado pelo TSE, pode participar do fórum
    RoleDeputado    = "deputado"        // Perfil oficial do parlamentar
    RoleModerador   = "moderador"       // Moderação do fórum
    RoleAdmin       = "admin"           // Administração do sistema
)
```

### Funcionalidades por Role

#### Público Geral

- Visualizar dados de deputados e atividades
- Consultar proposições e votações
- Ver rankings e estatísticas
- Acessar dados de transparência

#### Eleitor Validado (TSE)

- Todas as funcionalidades do público
- Participar do fórum de discussões
- Comentar em tópicos
- Sistema de gamificação (pontos, badges)
- Seguir deputados específicos

#### Deputado Verificado

- Perfil oficial verificado
- Responder diretamente aos cidadãos
- Criar tópicos no fórum
- Explicar votos e posicionamentos
- Acessar métricas do próprio desempenho
- Receber feedback direto dos eleitores

#### Moderador

- Moderar discussões do fórum
- Aplicar regras de convivência
- Gerenciar denúncias
- Validar contas de deputados

#### Administrador

- Gestão completa do sistema
- Configurações da plataforma
- Análise de métricas gerais
- Backup e manutenção

## 🎮 Sistema de Gamificação

### Elementos de Ludificação

#### Sistema de Pontos

- **Participação no Fórum**: Pontos por posts e comentários construtivos
- **Engajamento Cívico**: Pontos por acompanhar votações importantes
- **Conhecimento**: Pontos por acertar quiz sobre política
- **Transparência**: Pontos por usar ferramentas de fiscalização

#### Conquistas (Badges)

- 🏛️ **Fiscal Ativo**: Acompanha regularmente gastos de deputados
- 🗳️ **Eleitor Informado**: Conhece posicionamentos dos representantes
- 💬 **Voz Cidadã**: Participa ativamente das discussões
- 📊 **Analista**: Usa dados para fundamentar opiniões
- 🎯 **Vigilante**: Identifica inconsistências nos dados

#### Rankings

- **Cidadãos Mais Engajados**: Por pontuação acumulada
- **Deputados Mais Transparentes**: Por interação e dados atualizados
- **Estados Mais Participativos**: Por atividade dos usuários
- **Tópicos Mais Debatidos**: Por engajamento no fórum

### Mecânicas de Engajamento

#### Desafios Mensais

- "Conhece seu Deputado?": Quiz sobre o representante local
- "Fiscal do Mês": Acompanhar gastos e proposições
- "Debate Construtivo": Participar de discussões relevantes

#### Progressão

- **Nível Iniciante**: 0-100 pontos
- **Nível Cidadão**: 101-500 pontos
- **Nível Ativista**: 501-1000 pontos
- **Nível Especialista**: 1000+ pontos

#### Recompensas

- Acesso antecipado a relatórios especiais
- Badges exclusivos no perfil
- Reconhecimento na comunidade
- Participação em eventos especiais

## �️ Sistema de Participação Cidadã

### Plebiscitos e Consultas Populares

#### Tipos de Votação

- **Plebiscitos Locais**: Questões específicas por cidade/estado
- **Consultas Nacionais**: Temas de interesse geral
- **Enquetes Temáticas**: Posicionamento sobre proposições em tramitação
- **Avaliação de Deputados**: Feedback direto sobre performance parlamentar

#### Categorização Geográfica

```go
type Votacao struct {
    ID          uuid.UUID `json:"id"`
    Titulo      string    `json:"titulo"`
    Descricao   string    `json:"descricao"`
    Tipo        string    `json:"tipo"` // plebiscito, enquete, avaliacao
    Escopo      string    `json:"escopo"` // municipal, estadual, regional, nacional
    Estado      string    `json:"estado,omitempty"`
    Cidade      string    `json:"cidade,omitempty"`
    Regiao      string    `json:"regiao,omitempty"` // norte, nordeste, etc.
    DataInicio  time.Time `json:"data_inicio"`
    DataFim     time.Time `json:"data_fim"`
    Status      string    `json:"status"` // ativa, finalizada, rascunho
    Opcoes      []OpcaoVotacao `json:"opcoes"`
}

type OpcaoVotacao struct {
    ID       uuid.UUID `json:"id"`
    Texto    string    `json:"texto"`
    Votos    int       `json:"votos"`
    Detalhes string    `json:"detalhes,omitempty"`
}
```

#### Validação e Segurança

- **Eleitor Único**: Validação via CPF/TSE para evitar duplicatas
- **Verificação Regional**: Voto apenas em consultas da sua região
- **Auditoria**: Log completo de todas as votações
- **Anonimato**: Voto secreto com hash criptográfico

### Sistema de Propostas Cidadãs

#### Criação de Propostas

- **Cidadãos** podem propor plebiscitos locais
- **Deputados** podem criar consultas sobre seus projetos
- **Administradores** gerenciam propostas nacionais
- **Moderadores** validam propostas antes da publicação

#### Processo de Aprovação

```
1. Submissão da Proposta
   ├── Validação automática (spam, linguagem)
   ├── Revisão por moderadores
   └── Verificação de escopo geográfico

2. Período de Coleta de Apoio
   ├── Mínimo de apoiadores para ativação
   ├── Tempo limite para coleta
   └── Divulgação na plataforma

3. Votação Ativa
   ├── Período definido de votação
   ├── Notificações para eleitores elegíveis
   └── Acompanhamento em tempo real

4. Resultado e Ação
   ├── Publicação dos resultados
   ├── Encaminhamento para autoridades
   └── Acompanhamento de desdobramentos
```

### Estrutura de Projeto Go

```
/services/
├── deputados/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── domain/        # Entities e interfaces
│   │   ├── usecase/       # Business logic
│   │   ├── repository/    # Data access
│   │   └── handler/       # HTTP/gRPC handlers
│   ├── pkg/shared/        # Código compartilhado
│   └── deployments/       # Dockerfiles e K8s
```

### Convenções de Código

```go
// Naming: PascalCase para exports, camelCase para internal
type DeputadoService interface {
    BuscarPorID(ctx context.Context, id uuid.UUID) (*domain.Deputado, error)
    ListarAtivos(ctx context.Context, filtros *domain.FiltrosDeputado) ([]*domain.Deputado, error)
}

// Error handling com contexto
var (
    ErrDeputadoNaoEncontrado = errors.New("deputado não encontrado")
    ErrDadosInvalidos       = errors.New("dados do deputado inválidos")
)

// Logs estruturados
log.Info("deputado criado com sucesso",
    slog.String("id", deputado.ID.String()),
    slog.String("nome", deputado.Nome),
    slog.Duration("tempo", time.Since(start)))
```

### Frontend Next.js - Estrutura

```
/frontend/
├── app/                   # App Router (Next.js 15)
│   ├── (dashboard)/       # Route groups
│   ├── api/              # API routes
│   └── globals.css       # Tailwind + CSS vars
├── components/
│   ├── ui/               # Shadcn/ui components
│   ├── layout/           # Header, Footer, Sidebar
│   ├── features/         # Feature-specific components
│   └── charts/           # Gráficos com Recharts/D3
├── lib/
│   ├── api.ts            # API client (TanStack Query)
│   ├── auth.ts           # NextAuth.js setup
│   └── utils.ts          # Utilities + cn helper
└── types/                # TypeScript definitions
```

## 🎨 Diretrizes de UI/UX - Acessibilidade Universal

### Princípios de Design Inclusivo

#### Usabilidade Universal
- **Interface Intuitiva**: Design que funciona para todos os níveis de alfabetização digital
- **Linguagem Simples**: Evitar jargões técnicos, usar português claro e direto
- **Navegação Consistente**: Padrões familiares em toda a aplicação
- **Feedback Visual**: Confirmações claras para todas as ações do usuário

#### Acessibilidade (WCAG 2.1 AA)
```go
// Configurações de acessibilidade
type AccessibilityConfig struct {
    FontSizeMin      string `json:"font_size_min"`      // 16px mínimo
    ContrastRatio    string `json:"contrast_ratio"`     // 4.5:1 mínimo
    KeyboardNav      bool   `json:"keyboard_nav"`       // Navegação completa via teclado
    ScreenReader     bool   `json:"screen_reader"`      // Compatibilidade com leitores
    AltTextRequired  bool   `json:"alt_text_required"`  // Textos alternativos obrigatórios
}
```

#### Design Responsivo
- **Mobile First**: Priorizar experiência em dispositivos móveis
- **Progressive Enhancement**: Funcionalidades básicas em qualquer dispositivo
- **Touch Targets**: Botões com 44px mínimo (iOS/Android guidelines)
- **Zooom**: Suporte a zoom até 200% sem perda de funcionalidade

#### Simplificação da Interface
- **Hierarquia Visual Clara**: Títulos, subtítulos e conteúdo bem definidos
- **Cores Funcionais**: Sistema de cores que comunica significado
- **Iconografia Universal**: Ícones reconhecíveis internacionalmente
- **Carregamento Progressivo**: Skeleton screens e lazy loading

### Sistema de Comentários Sociais

#### Estrutura de Comentários (Estilo Instagram)
```sql
-- Sistema de comentários hierárquicos
CREATE TABLE comentarios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID NOT NULL REFERENCES usuarios(id),
    topico_id UUID REFERENCES topicos(id),
    comentario_pai_id UUID REFERENCES comentarios(id), -- Para respostas
    conteudo TEXT NOT NULL,
    total_likes INTEGER DEFAULT 0,
    total_respostas INTEGER DEFAULT 0,
    nivel_aninhamento INTEGER DEFAULT 0, -- Máximo 3 níveis
    is_moderado BOOLEAN DEFAULT false,
    moderacao_ia JSONB, -- Resultado da análise de IA
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

-- Sistema de likes/reactions
CREATE TABLE comentario_likes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    comentario_id UUID NOT NULL REFERENCES comentarios(id),
    usuario_id UUID NOT NULL REFERENCES usuarios(id),
    tipo_reacao TEXT DEFAULT 'like', -- like, dislike, love, angry
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(comentario_id, usuario_id)
);

-- Notificações para respostas
CREATE TABLE notificacoes_comentarios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_destinatario_id UUID NOT NULL REFERENCES usuarios(id),
    comentario_id UUID NOT NULL REFERENCES comentarios(id),
    tipo_notificacao TEXT NOT NULL, -- resposta, like, mencao
    lida BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);
```

#### Funcionalidades Sociais Avançadas
- **Menções**: @username para notificar usuários específicos
- **Hashtags**: #tema para categorizar discussões
- **Reações Emotivas**: Like, dislike, love, angry (estilo Facebook)
- **Threading**: Até 3 níveis de respostas aninhadas
- **Moderação em Tempo Real**: IA + moderação humana
- **Histórico de Edições**: Transparência nas alterações

## 🚀 Script de Inicialização - Bootstrap do Sistema

### Processo de Carga Inicial (Cold Start)

#### 1. Ingestão de Dados Históricos
```bash
#!/bin/bash
# scripts/bootstrap-inicial.sh

echo "🏛️ Iniciando bootstrap do sistema Tô De Olho..."

# 1. Carga de dados da Câmara (últimos 4 anos)
echo "📊 Carregando dados históricos da Câmara..."
go run cmd/bootstrap/main.go --mode=full-sync --years=4

# 2. Sincronização de deputados ativos
echo "👥 Sincronizando deputados ativos..."
go run cmd/sync/deputados.go --current-legislature

# 3. Carga de proposições relevantes
echo "📜 Carregando proposições em tramitação..."
go run cmd/sync/proposicoes.go --status=tramitando

# 4. Histórico de votações importantes
echo "🗳️ Sincronizando votações dos últimos 2 anos..."
go run cmd/sync/votacoes.go --period=24months

# 5. Dados de despesas (cota parlamentar)
echo "💰 Carregando dados de despesas..."
go run cmd/sync/despesas.go --full-sync

# 6. Criação de índices e otimizações
echo "⚡ Otimizando banco de dados..."
psql -f migrations/optimize-indexes.sql

# 7. Setup de dados demo para desenvolvimento
echo "🎮 Criando dados de demonstração..."
go run cmd/seed/demo-data.go

echo "✅ Bootstrap concluído com sucesso!"
```

#### 2. Pipeline de ETL Automatizado
```go
// cmd/bootstrap/main.go
type BootstrapService struct {
    camaraClient    *camara.Client
    tseClient       *tse.Client
    dbConn          *sql.DB
    logger          *slog.Logger
    progressTracker *ProgressTracker
}

type BootstrapOptions struct {
    FullSync        bool `json:"full_sync"`
    YearsBack       int  `json:"years_back"`
    CurrentOnly     bool `json:"current_only"`
    SkipValidation  bool `json:"skip_validation"`
    ParallelWorkers int  `json:"parallel_workers"`
}

func (bs *BootstrapService) ExecuteFullBootstrap(ctx context.Context, opts BootstrapOptions) error {
    // 1. Validar conectividade APIs
    if err := bs.validateAPIsConnectivity(ctx); err != nil {
        return fmt.Errorf("falha na conectividade: %w", err)
    }
    
    // 2. Executar ETL em paralelo com workers
    tasks := []BootstrapTask{
        {Name: "deputados", Priority: 1, Fn: bs.syncDeputados},
        {Name: "partidos", Priority: 1, Fn: bs.syncPartidos},
        {Name: "proposicoes", Priority: 2, Fn: bs.syncProposicoes},
        {Name: "votacoes", Priority: 3, Fn: bs.syncVotacoes},
        {Name: "despesas", Priority: 2, Fn: bs.syncDespesas},
    }
    
    return bs.executeTasksInParallel(ctx, tasks, opts.ParallelWorkers)
}
```

#### 3. Dados de Demonstração e Seed
```go
// cmd/seed/demo-data.go - Popular sistema para demonstrações
func SeedDemoData(db *sql.DB) error {
    // Usuários demo com diferentes roles
    demoUsers := []DemoUser{
        {Role: "publico", Username: "cidadao_demo", Region: "BA"},
        {Role: "eleitor", Username: "eleitor_bahia", CPF: "000.000.000-00"},
        {Role: "deputado", Username: "dep_oficial", DeputadoID: uuid.New()},
        {Role: "moderador", Username: "mod_forum", Permissions: []string{"moderate", "ban"}},
    }
    
    // Tópicos de discussão populares
    demoTopics := []Topic{
        {Title: "Orçamento da Educação 2025", Category: "educacao"},
        {Title: "Reforma Tributária - Impactos", Category: "economia"},
        {Title: "Meio Ambiente e Sustentabilidade", Category: "meio_ambiente"},
    }
    
    // Comentários e interações realísticas
    return seedInteractiveDemo(db, demoUsers, demoTopics)
}
```

### Estratégia de Cold Start - Ingestão Inteligente

#### 1. Priorização por Relevância e Volume
```go
// Pipeline de ingestão otimizada baseada na API oficial
type ColdStartPipeline struct {
    phases []IngestionPhase
    stats  *IngestionStats
}

// Fases priorizadas para cold start
var ColdStartPhases = []IngestionPhase{
    // FASE 1: Dados Estruturais (Rápido)
    {
        Name: "referencias",
        Priority: 1,
        Endpoints: []string{
            "/referencias/deputados/siglaUF",
            "/referencias/deputados/tipoDespesa", 
            "/referencias/proposicoes/siglaTipo",
            "/referencias/partidos",
        },
        EstimatedItems: 200,
        Description: "Tabelas de referência e lookup",
    },
    
    // FASE 2: Deputados Ativos (Crítico)
    {
        Name: "deputados_ativos",
        Priority: 2,
        Endpoints: []string{
            "/deputados", // Apenas legislatura atual
            "/deputados/{id}/orgaos",
            "/deputados/{id}/profissoes",
        },
        EstimatedItems: 513, // Total de deputados
        Description: "Deputados em exercício + cargos",
        Filters: map[string]string{
            "idLegislatura": "57", // Legislatura 2023-2027
        },
    },
    
    // FASE 3: Dados Históricos Essenciais (6 meses)
    {
        Name: "atividades_recentes",
        Priority: 3,
        Endpoints: []string{
            "/deputados/{id}/despesas",
            "/eventos", 
            "/proposicoes",
            "/votacoes",
        },
        EstimatedItems: 50000,
        Description: "Atividades dos últimos 6 meses",
        TimeFilter: "6months",
    },
    
    // FASE 4: Dados Históricos Completos (Opcional)
    {
        Name: "historico_completo",
        Priority: 4,
        Endpoints: []string{
            "/deputados/{id}/historico",
            "/deputados/{id}/mandatosExternos",
            "/proposicoes/{id}/tramitacoes",
        },
        EstimatedItems: 200000,
        Description: "Histórico completo para análises",
        Background: true, // Executar em background
    },
}
```

#### 2. Cache Inteligente e Otimizações
```go
// Sistema de cache hierárquico para cold start
type CacheStrategy struct {
    L1Cache *redis.Client     // Dados mais acessados (deputados, proposições)
    L2Cache *sql.DB          // Dados estruturados (PostgreSQL)
    L3Cache string           // Arquivos estáticos (JSON/parquet)
}

// Cache warming prioritário
func (cs *CacheStrategy) WarmupEssentialData() error {
    // 1. Cache de deputados ativos (acesso frequente)
    deputados, err := cs.fetchDeputadosAtivos()
    if err != nil {
        return err
    }
    
    // 2. Cache de proposições em tramitação
    proposicoes, err := cs.fetchProposicoesAtivas()
    if err != nil {
        return err
    }
    
    // 3. Cache de eventos da semana
    eventos, err := cs.fetchEventosSemana()
    if err != nil {
        return err
    }
    
    return cs.preloadToRedis(deputados, proposicoes, eventos)
}
```

#### 3. Monitoramento de Ingestão
```go
// Dashboard de progresso do cold start
type IngestionStats struct {
    TotalEndpoints    int                    `json:"total_endpoints"`
    CompletedPhases   []string              `json:"completed_phases"`
    CurrentPhase      string                `json:"current_phase"`
    ItemsProcessed    int64                 `json:"items_processed"`
    EstimatedRemaining int64                `json:"estimated_remaining"`
    ErrorRate         float64               `json:"error_rate"`
    AvgRequestTime    time.Duration         `json:"avg_request_time"`
    ETACompletion     time.Time             `json:"eta_completion"`
    
    // Por endpoint
    EndpointStats     map[string]EndpointStat `json:"endpoint_stats"`
}

type EndpointStat struct {
    URL            string        `json:"url"`
    RequestCount   int64         `json:"request_count"`
    SuccessCount   int64         `json:"success_count"`
    ErrorCount     int64         `json:"error_count"`
    AvgResponseTime time.Duration `json:"avg_response_time"`
    LastSync       time.Time     `json:"last_sync"`
    DataFreshness  string        `json:"data_freshness"` // fresh, stale, expired
}
```

#### 4. Recuperação de Falhas e Rate Limiting
```go
// Sistema robusto para lidar com limitações da API
type ResilientClient struct {
    httpClient   *http.Client
    rateLimiter  *rate.Limiter  // 100 req/min baseado na API
    retryPolicy  *RetryPolicy
    circuitBreaker *CircuitBreaker
}

type RetryPolicy struct {
    MaxRetries      int           `json:"max_retries"`
    InitialDelay    time.Duration `json:"initial_delay"`
    MaxDelay        time.Duration `json:"max_delay"`
    BackoffFactor   float64       `json:"backoff_factor"`
    RetryableErrors []int         `json:"retryable_errors"` // 429, 502, 503, 504
}

func (rc *ResilientClient) FetchWithResilience(endpoint string) (*http.Response, error) {
    // 1. Rate limiting (100 req/min)
    if err := rc.rateLimiter.Wait(context.Background()); err != nil {
        return nil, err
    }
    
    // 2. Circuit breaker para endpoints com falha
    if !rc.circuitBreaker.AllowRequest(endpoint) {
        return nil, ErrCircuitBreakerOpen
    }
    
    // 3. Retry com backoff exponencial
    return rc.retryPolicy.ExecuteWithRetry(func() (*http.Response, error) {
        return rc.httpClient.Get(endpoint)
    })
}
```

## 🔐 Segurança e Autenticação

### Sistema de Autenticação

```go
// JWT com refresh tokens
type TokenPair struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"`
}

// Rate limiting por usuário/IP
middleware.RateLimit(store.NewRedisStore(redisClient,
    ratelimit.WithRateLimit(100, time.Hour)))

// RBAC (Role-Based Access Control)
const (
    RolePublico    = "publico"
    RoleEleitor    = "eleitor_validado"
    RoleDeputado   = "deputado"
    RoleModerador  = "moderador"
    RoleAdmin      = "admin"
)
```

### Validação de Deputados

- Verificação via dados oficiais da Câmara
- Processo de validação manual inicial
- Badge de "Perfil Verificado"
- Acesso especial a funcionalidades do fórum

### Pipeline de Ingestão de Dados

```
Phase 1: Carga Inicial (Backfill)
├── Download de arquivos históricos (JSON/CSV)
├── Validação e limpeza de dados
├── Indexação no PostgreSQL
└── Cache inicial no Redis

Phase 2: Atualizações Contínuas
├── CronJobs diários da API
├── Processamento via message queue
├── Updates incrementais
└── Notificações de mudanças
```

## 🚀 Deploy e Infraestrutura

### Containerização

```dockerfile
# Build multi-stage para Go
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

### Kubernetes

```yaml
# Horizontal Pod Autoscaler
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: deputados-service-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: deputados-service
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
```

### CI/CD Pipeline

```yaml
name: Deploy
on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: "1.23"
      - name: Run Tests
        run: go test -race ./...
      - name: Security Scan
        run: gosec ./...

  deploy:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to Kubernetes
        run: kubectl rollout restart deployment/deputados-service
```

---

**🎯 Objetivo**: Criar uma plataforma funcional de transparência política que permita aos cidadãos fiscalizar e interagir com seus representantes na Câmara dos Deputados, promovendo maior engajamento democrático através de acessibilidade, gestão social e gamificação.

## 🌟 Diferenciais Competitivos - "Por que o Tô De Olho?"

### 🚀 Inovações Únicas no Mercado

#### 1. **IA Conversacional Educativa (Gemini Integration)**
- **Assistente Político Pessoal**: Chatbot que explica projetos de lei em linguagem simples
- **Análise Preditiva**: Previsão de como deputados votarão baseado em histórico
- **Fact-Checking Automático**: Verificação de informações políticas em tempo real
- **Resumos Inteligentes**: IA que transforma sessões de 3h em resumos de 3 minutos

#### 2. **Gamificação Cívica Inovadora**
- **RPG Democrático**: Sistema de níveis onde cidadãos "evoluem" seu conhecimento político
- **Conquistas Temáticas**: Badges por especialização (saúde, educação, economia)
- **Missões Cidadãs**: Desafios para engajar com deputados locais
- **Leaderboards Regionais**: Rankings que estimulam participação por estado/cidade

#### 3. **Democracia Participativa Digital**
- **Plebiscitos Hiperlocais**: Consultas por bairro/município com validação TSE
- **Simulador de Impacto**: "Se essa lei passar, como afetará sua região?"
- **Propostas Colaborativas**: Cidadãos co-criam projetos com deputados
- **Orçamento Participativo Digital**: Votação em prioridades orçamentárias

#### 4. **Transparência 360° com Social Media**
- **Instagram-Style Comments**: Sistema de comentários familiar e intuitivo
- **Stories Parlamentares**: Deputados explicam votos em formato story
- **Live Q&A**: Transmissões ao vivo deputado-cidadão
- **Feeds Personalizados**: Algoritmo que mostra política relevante para você

### 🎯 Comparativo com Concorrentes

| Diferencial | Tô De Olho | Concorrentes Atuais |
|-------------|------------|-------------------|
| **IA Educativa** | ✅ Gemini AI integrada | ❌ Apenas dados estáticos |
| **Gamificação** | ✅ Sistema RPG completo | ❌ No máximo badges simples |
| **Plebiscitos** | ✅ Validação TSE + regional | ❌ Enquetes não oficiais |
| **UX Social** | ✅ Instagram-style | ❌ Interfaces antigas |
| **Mobile-First** | ✅ App nativo futuro | ❌ Sites não responsivos |
| **Moderação IA** | ✅ Anti-toxicidade Gemini | ❌ Moderação manual |

### 🏆 Proposta de Valor Única

#### **"Política como Rede Social, Educação como Jogo"**

1. **Acessibilidade**: Qualquer pessoa, independente da escolaridade, consegue usar
2. **Engajamento**: Gamificação torna política viciante (no bom sentido)
3. **Educação**: IA ensina democracia de forma personalizada
4. **Participação**: Primeiro app que permite democracia direta digital
5. **Transparência**: Dados governamentais em formato humano

#### Casos de Uso Únicos
- **Jovens 16-25**: "TikTok da política" - aprende sem perceber
- **Cidadãos 30-50**: Acompanha deputados como segue influencers
- **Ativistas**: Ferramentas profissionais de mobilização
- **Deputados**: Canal direto com eleitores + analytics
- **Pesquisadores**: APIs e dados para estudos acadêmicos

### 🔮 Visão de Futuro (2026+)

#### Expansão Nacional
- **Câmara Municipal**: Transparência nas 5.570 cidades brasileiras
- **Senado Federal**: Mesma experiência para senadores
- **Assembleias Estaduais**: Política estadual gamificada
- **Judiciário**: Transparência do STF e tribunais

#### Tecnologias Emergentes
- **Blockchain**: Votações auditáveis e imutáveis
- **AR/VR**: Visitas virtuais ao Congresso
- **IoT**: Dados em tempo real de presença parlamentar
- **ML Avançado**: Predição de políticas públicas

### 💡 Mensagem Central

> **"Não é apenas outro site de transparência. É a primeira rede social que transforma cada brasileiro em um fiscal ativo da democracia, usando IA para educar, gamificação para engajar e tecnologia para empoderar."**
