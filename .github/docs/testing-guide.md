# 🧪 Testing Guide - Estratégias de Teste

## 📊 Testing Pyramid (80/15/5)

```
        🔺 E2E Tests (5%)
       /                \
     🔺 Integration (15%)
   /                        \
 🔺 Unit Tests (80%)
```

### Cobertura Mínima por Camada
- **Domain Layer**: 95% (business logic crítica)
- **Application Layer**: 90% (casos de uso)
- **Infrastructure Layer**: 70% (adaptadores)
- **Interface Layer**: 60% (controllers simples)

## 🎯 Unit Tests (80%)

### Estrutura de Testes Unitários
```go
// test_helper.go
package domain_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "to-de-olho/internal/domain"
)

// Factories para criação de dados de teste
type DeputadoFactory struct{}

func (f *DeputadoFactory) Build() *domain.Deputado {
    return &domain.Deputado{
        ID:     uuid.New(),
        Nome:   "Deputado Test",
        CPF:    "12345678901",
        Estado: "SP",
        Status: domain.StatusAtivo,
    }
}

func (f *DeputadoFactory) WithNome(nome string) *domain.Deputado {
    d := f.Build()
    d.Nome = nome
    return d
}

func (f *DeputadoFactory) WithEstado(uf string) *domain.Deputado {
    d := f.Build()
    d.Estado = uf
    return d
}
```

### Table-Driven Tests (Obrigatório)
```go
func TestDeputadoValidator_Validate(t *testing.T) {
    validator := domain.NewDeputadoValidator()
    factory := &DeputadoFactory{}
    
    tests := []struct {
        name      string
        deputado  *domain.Deputado
        wantError bool
        errorCode string
        errorMsg  string
    }{
        {
            name:      "deputado válido",
            deputado:  factory.Build(),
            wantError: false,
        },
        {
            name:      "nome vazio deve falhar",
            deputado:  factory.WithNome(""),
            wantError: true,
            errorCode: "INVALID_NAME",
            errorMsg:  "nome é obrigatório",
        },
        {
            name:      "CPF inválido deve falhar",
            deputado:  factory.WithCPF("123"),
            wantError: true,
            errorCode: "INVALID_CPF",
            errorMsg:  "CPF inválido",
        },
        {
            name:      "estado inválido deve falhar",
            deputado:  factory.WithEstado("XX"),
            wantError: true,
            errorCode: "INVALID_STATE",
            errorMsg:  "estado inválido",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.Validate(tt.deputado)
            
            if tt.wantError {
                require.Error(t, err)
                
                var validationErr *domain.ValidationError
                require.True(t, errors.As(err, &validationErr))
                assert.Equal(t, tt.errorCode, validationErr.Code)
                assert.Contains(t, validationErr.Message, tt.errorMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Testes de Value Objects
```go
func TestCPF_NewCPF(t *testing.T) {
    tests := []struct {
        name      string
        input     string
        wantValid bool
        expected  string
    }{
        {
            name:      "CPF válido com formatação",
            input:     "123.456.789-01",
            wantValid: true,
            expected:  "12345678901",
        },
        {
            name:      "CPF válido sem formatação",
            input:     "12345678901",
            wantValid: true,
            expected:  "12345678901",
        },
        {
            name:      "CPF inválido - menos dígitos",
            input:     "123456789",
            wantValid: false,
        },
        {
            name:      "CPF inválido - dígitos verificadores",
            input:     "12345678900",
            wantValid: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cpf, err := domain.NewCPF(tt.input)
            
            if tt.wantValid {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, cpf.String())
            } else {
                assert.Error(t, err)
                assert.Nil(t, cpf)
            }
        })
    }
}
```

### Testes de Use Cases com Mocks
```go
func TestBuscarDeputadoUseCase_Execute(t *testing.T) {
    // Setup
    mockRepo := new(MockDeputadoRepository)
    mockValidator := new(MockValidator)
    logger := slog.New(slog.NewTextHandler(io.Discard, nil))
    
    uc := application.NewBuscarDeputadoUseCase(mockRepo, mockValidator, logger)
    factory := &DeputadoFactory{}
    
    tests := []struct {
        name      string
        input     application.BuscarDeputadoInput
        setupMock func()
        wantError bool
        errorCode string
    }{
        {
            name:  "deve buscar deputado com sucesso",
            input: application.BuscarDeputadoInput{ID: uuid.New()},
            setupMock: func() {
                deputado := factory.Build()
                mockValidator.On("ValidateStruct", mock.Anything).Return(nil)
                mockRepo.On("FindByID", mock.Anything, mock.Anything).Return(deputado, nil)
            },
            wantError: false,
        },
        {
            name:  "deve falhar com input inválido",
            input: application.BuscarDeputadoInput{ID: uuid.Nil},
            setupMock: func() {
                mockValidator.On("ValidateStruct", mock.Anything).Return(domain.ErrInputInvalido)
            },
            wantError: true,
            errorCode: "INVALID_INPUT",
        },
        {
            name:  "deve falhar quando deputado não existe",
            input: application.BuscarDeputadoInput{ID: uuid.New()},
            setupMock: func() {
                mockValidator.On("ValidateStruct", mock.Anything).Return(nil)
                mockRepo.On("FindByID", mock.Anything, mock.Anything).Return(nil, domain.ErrDeputadoNaoEncontrado)
            },
            wantError: true,
            errorCode: "DEPUTADO_NOT_FOUND",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mocks
            tt.setupMock()
            
            // Execute
            output, err := uc.Execute(context.Background(), tt.input)
            
            // Assert
            if tt.wantError {
                assert.Error(t, err)
                assert.Nil(t, output)
                
                var appErr *application.Error
                if errors.As(err, &appErr) {
                    assert.Equal(t, tt.errorCode, appErr.Code)
                }
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, output)
                assert.NotNil(t, output.Deputado)
            }
            
            // Verify mocks
            mockRepo.AssertExpectations(t)
            mockValidator.AssertExpectations(t)
        })
    }
}
```

## 🔗 Integration Tests (15%)

### Setup com Testcontainers
```go
package integration_test

import (
    "context"
    "database/sql"
    "testing"
    "time"
    
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    "github.com/testcontainers/testcontainers-go/modules/redis"
)

type TestSuite struct {
    DB            *sql.DB
    Redis         *redis.Client
    PostgresContainer testcontainers.Container
    RedisContainer    testcontainers.Container
}

func SetupTestSuite(t *testing.T) *TestSuite {
    ctx := context.Background()
    
    // Setup PostgreSQL container
    postgresContainer, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:16"),
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).
                WithStartupTimeout(5*time.Second),
        ),
    )
    require.NoError(t, err)
    
    // Setup Redis container
    redisContainer, err := redis.RunContainer(ctx,
        testcontainers.WithImage("redis:7"),
    )
    require.NoError(t, err)
    
    // Get connection strings
    postgresURL, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
    require.NoError(t, err)
    
    redisURL, err := redisContainer.ConnectionString(ctx)
    require.NoError(t, err)
    
    // Connect to databases
    db, err := sql.Open("postgres", postgresURL)
    require.NoError(t, err)
    
    redisClient := redis.NewClient(&redis.Options{
        Addr: redisURL,
    })
    
    // Run migrations
    err = runMigrations(db)
    require.NoError(t, err)
    
    // Cleanup function
    t.Cleanup(func() {
        db.Close()
        redisClient.Close()
        postgresContainer.Terminate(ctx)
        redisContainer.Terminate(ctx)
    })
    
    return &TestSuite{
        DB:                db,
        Redis:            redisClient,
        PostgresContainer: postgresContainer,
        RedisContainer:    redisContainer,
    }
}
```

### Testes de Repository
```go
func TestDeputadoRepository_Integration(t *testing.T) {
    suite := SetupTestSuite(t)
    
    repo := infrastructure.NewPostgresDeputadoRepository(suite.DB, slog.Default())
    factory := &DeputadoFactory{}
    
    t.Run("deve salvar e recuperar deputado", func(t *testing.T) {
        // Arrange
        ctx := context.Background()
        deputado := factory.Build()
        
        // Act - Save
        err := repo.Save(ctx, deputado)
        require.NoError(t, err)
        
        // Act - Find
        retrieved, err := repo.FindByID(ctx, deputado.ID)
        require.NoError(t, err)
        
        // Assert
        assert.Equal(t, deputado.ID, retrieved.ID)
        assert.Equal(t, deputado.Nome, retrieved.Nome)
        assert.Equal(t, deputado.CPF, retrieved.CPF)
        assert.Equal(t, deputado.Estado, retrieved.Estado)
    })
    
    t.Run("deve retornar erro quando deputado não existe", func(t *testing.T) {
        // Arrange
        ctx := context.Background()
        nonExistentID := uuid.New()
        
        // Act
        deputado, err := repo.FindByID(ctx, nonExistentID)
        
        // Assert
        assert.Error(t, err)
        assert.Nil(t, deputado)
        assert.True(t, errors.Is(err, domain.ErrDeputadoNaoEncontrado))
    })
    
    t.Run("deve listar deputados por estado", func(t *testing.T) {
        // Arrange
        ctx := context.Background()
        deputados := []*domain.Deputado{
            factory.WithEstado("SP"),
            factory.WithEstado("SP"),
            factory.WithEstado("RJ"),
        }
        
        for _, d := range deputados {
            err := repo.Save(ctx, d)
            require.NoError(t, err)
        }
        
        // Act
        deputadosSP, err := repo.FindByEstado(ctx, "SP")
        require.NoError(t, err)
        
        // Assert
        assert.Len(t, deputadosSP, 2)
        for _, d := range deputadosSP {
            assert.Equal(t, "SP", d.Estado)
        }
    })
}
```

### Testes de API
```go
func TestDeputadoAPI_Integration(t *testing.T) {
    suite := SetupTestSuite(t)
    
    // Setup da aplicação completa
    container := setupServiceContainer(suite.DB, suite.Redis)
    router := setupRouter(container)
    
    server := httptest.NewServer(router)
    defer server.Close()
    
    client := &http.Client{Timeout: 5 * time.Second}
    
    t.Run("GET /api/v1/deputados/{id} - deve retornar deputado", func(t *testing.T) {
        // Arrange - criar deputado no banco
        deputado := createTestDeputado(t, suite.DB)
        
        // Act
        resp, err := client.Get(fmt.Sprintf("%s/api/v1/deputados/%s", server.URL, deputado.ID))
        require.NoError(t, err)
        defer resp.Body.Close()
        
        // Assert
        assert.Equal(t, http.StatusOK, resp.StatusCode)
        
        var response map[string]interface{}
        err = json.NewDecoder(resp.Body).Decode(&response)
        require.NoError(t, err)
        
        assert.Equal(t, deputado.ID.String(), response["deputado"].(map[string]interface{})["id"])
        assert.Equal(t, deputado.Nome, response["deputado"].(map[string]interface{})["nome"])
    })
    
    t.Run("GET /api/v1/deputados/{id} - deve retornar 404 para deputado inexistente", func(t *testing.T) {
        // Arrange
        nonExistentID := uuid.New()
        
        // Act
        resp, err := client.Get(fmt.Sprintf("%s/api/v1/deputados/%s", server.URL, nonExistentID))
        require.NoError(t, err)
        defer resp.Body.Close()
        
        // Assert
        assert.Equal(t, http.StatusNotFound, resp.StatusCode)
    })
}
```

## 🌐 E2E Tests (5%)

### Setup com Playwright (Go)
```go
package e2e_test

import (
    "testing"
    "github.com/playwright-community/playwright-go"
)

func TestDeputadosE2E(t *testing.T) {
    // Setup
    pw, err := playwright.Run()
    require.NoError(t, err)
    defer pw.Stop()
    
    browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
        Headless: playwright.Bool(true),
    })
    require.NoError(t, err)
    defer browser.Close()
    
    page, err := browser.NewPage()
    require.NoError(t, err)
    
    baseURL := "http://localhost:3000"
    
    t.Run("jornada completa do usuário - buscar deputado", func(t *testing.T) {
        // 1. Navegar para a página inicial
        _, err := page.Goto(baseURL)
        require.NoError(t, err)
        
        // 2. Verificar se a página carregou
        title, err := page.Title()
        require.NoError(t, err)
        assert.Contains(t, title, "Tô De Olho")
        
        // 3. Buscar por um deputado
        searchInput := page.Locator("[data-testid=search-input]")
        err = searchInput.Fill("João Silva")
        require.NoError(t, err)
        
        searchButton := page.Locator("[data-testid=search-button]")
        err = searchButton.Click()
        require.NoError(t, err)
        
        // 4. Verificar resultados
        results := page.Locator("[data-testid=deputado-card]")
        count, err := results.Count()
        require.NoError(t, err)
        assert.Greater(t, count, 0)
        
        // 5. Clicar no primeiro resultado
        firstResult := results.First()
        err = firstResult.Click()
        require.NoError(t, err)
        
        // 6. Verificar página de detalhes
        deputadoName := page.Locator("[data-testid=deputado-name]")
        name, err := deputadoName.TextContent()
        require.NoError(t, err)
        assert.Contains(t, name, "João Silva")
        
        // 7. Verificar seções da página
        despesasSection := page.Locator("[data-testid=despesas-section]")
        assert.True(t, deputadoName.IsVisible())
        
        proposicoesSection := page.Locator("[data-testid=proposicoes-section]")
        assert.True(t, proposicoesSection.IsVisible())
    })
    
    t.Run("jornada do usuário autenticado - comentar", func(t *testing.T) {
        // 1. Fazer login
        err := loginAsEleitor(page, baseURL)
        require.NoError(t, err)
        
        // 2. Navegar para perfil de deputado
        deputadoURL := fmt.Sprintf("%s/deputados/%s", baseURL, testDeputadoID)
        _, err = page.Goto(deputadoURL)
        require.NoError(t, err)
        
        // 3. Adicionar comentário
        commentInput := page.Locator("[data-testid=comment-input]")
        err = commentInput.Fill("Excelente trabalho na comissão de educação!")
        require.NoError(t, err)
        
        submitButton := page.Locator("[data-testid=submit-comment]")
        err = submitButton.Click()
        require.NoError(t, err)
        
        // 4. Verificar comentário adicionado
        comments := page.Locator("[data-testid=comment-item]")
        count, err := comments.Count()
        require.NoError(t, err)
        assert.Greater(t, count, 0)
        
        lastComment := comments.Last()
        text, err := lastComment.TextContent()
        require.NoError(t, err)
        assert.Contains(t, text, "Excelente trabalho")
    })
}

func loginAsEleitor(page playwright.Page, baseURL string) error {
    loginURL := fmt.Sprintf("%s/login", baseURL)
    _, err := page.Goto(loginURL)
    if err != nil {
        return err
    }
    
    // Preencher formulário de login
    emailInput := page.Locator("[data-testid=email-input]")
    err = emailInput.Fill("eleitor@test.com")
    if err != nil {
        return err
    }
    
    passwordInput := page.Locator("[data-testid=password-input]")
    err = passwordInput.Fill("123456")
    if err != nil {
        return err
    }
    
    loginButton := page.Locator("[data-testid=login-button]")
    return loginButton.Click()
}
```

## 📊 Coverage e Quality Gates

### Script de Coverage
```bash
#!/bin/bash
# scripts/test-coverage.sh

echo "🧪 Executando testes com coverage..."

# Unit tests
go test -race -coverprofile=coverage.out -covermode=atomic ./...

# Verificar coverage mínimo
coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
threshold=80

if (( $(echo "$coverage < $threshold" | bc -l) )); then
    echo "❌ Coverage ($coverage%) está abaixo do mínimo ($threshold%)"
    exit 1
fi

echo "✅ Coverage: $coverage%"

# Gerar relatório HTML
go tool cover -html=coverage.out -o coverage.html
echo "📊 Relatório HTML gerado: coverage.html"

# Integration tests
echo "🔗 Executando testes de integração..."
go test -tags=integration ./tests/integration/...

# E2E tests (apenas se especificado)
if [ "$RUN_E2E" = "true" ]; then
    echo "🌐 Executando testes E2E..."
    go test -tags=e2e ./tests/e2e/...
fi

echo "✅ Todos os testes executados com sucesso!"
```

### Makefile para Testes
```makefile
# Makefile

.PHONY: test test-unit test-integration test-e2e test-coverage

# Testes unitários
test-unit:
	@echo "🧪 Executando testes unitários..."
	go test -race -short ./...

# Testes de integração
test-integration:
	@echo "🔗 Executando testes de integração..."
	go test -tags=integration ./tests/integration/...

# Testes E2E
test-e2e:
	@echo "🌐 Executando testes E2E..."
	go test -tags=e2e -timeout=10m ./tests/e2e/...

# Coverage completo
test-coverage:
	@./scripts/test-coverage.sh

# Executar todos os testes
test: test-unit test-integration
	@echo "✅ Todos os testes executados!"

# CI pipeline
test-ci: test-coverage
	@echo "🚀 Pipeline de testes CI executado!"
```

---

> **🎯 Objetivo**: Garantir qualidade através de testes automatizados, cobertura alta e feedback rápido no desenvolvimento.
