# 🤖 GitHub Copilot - Instruções Core

## 🎯 Contexto Automático & Eficiência

### 📁 Referências Obrigatórias
**SEMPRE** que trabalhar com frontend ou backend:
- **Frontend**: Consulte automaticamente `#file:frontend` para estrutura, componentes e padrões
- **Backend**: Consulte automaticamente `#file:backend` para arquitetura, domínios e serviços
- **Documentação**: Utilize este `#file:copilot-instructions.md` como fonte da verdade

### ⚡ Eficiência de Tokens
- **Priorize**: Leitura de arquivos relevantes ao contexto específico
- **Evite**: Leituras desnecessárias ou redundantes
- **Use**: `semantic_search` para localizar implementações antes de criar
- **Aplique**: Padrões já existentes no projeto antes de criar novos

### 🔄 Workflow Inteligente
1. **Analise** o contexto atual (frontend/backend)
2. **Busque** referências nos diretórios relevantes
3. **Aplique** padrões e convenções estabelecidas
4. **Mantenha** consistência com o código existente

---

## 🎯 Visão do Projeto

O **"Tô De Olho"** é uma plataforma de transparência política que democratiza o acesso aos dados da Câmara dos Deputados, promovendo engajamento cidadão através de:

1. **Acessibilidade**: Interface intuitiva para todos os usuários
2. **Gestão Social**: Participação cidadã nas decisões públicas  
3. **Gamificação**: Sistema de pontos, conquistas e rankings

### Características Essenciais
- **Linguagem**: Português Brasileiro (pt-BR)
- **Dados**: API Câmara dos Deputados + TSE
- **Interação**: Fórum e contato deputado-cidadão
- **IA**: Google Gemini SDK para moderação e assistente educativo

## 🛠️ Stack Tecnológico (2025-2026)

```
DevEnv: Windows PowerShell + WSL2 + Docker Desktop
Backend:     Go 1.24+ (Clean Architecture + DDD)
Frontend:    Next.js 15 + TypeScript + Tailwind CSS
Database:    PostgreSQL 16 + Redis (cache)
Queue:       RabbitMQ (mensageria assíncrona)
AI/ML:       Google Gemini SDK + MCP
Monitoring:  Prometheus + Grafana
Security:    JWT + OAuth2 + Rate Limiting
Testing:     80% Coverage (Unit + Integration + E2E)
CI/CD:       GitHub Actions com Quality Gates
```

## 🏗️ Padrões de Arquitetura

### Clean Architecture + DDD
```go
// Estrutura por domínio de negócio
/backend/services/deputados/
├── cmd/server/                  # Entry points
├── internal/
│   ├── domain/                  # Entities, Value Objects, Aggregates
│   ├── application/             # Use Cases / Application Services
│   ├── infrastructure/          # Frameworks & Drivers
│   └── interfaces/             # Interface Adapters
├── pkg/                        # Código compartilhado público
└── tests/                      # Testes organizados por tipo
```

### Princípios SOLID Obrigatórios
- **Single Responsibility**: Uma classe, uma responsabilidade
- **Open/Closed**: Extensível sem modificação
- **Liskov Substitution**: Subtipos substituíveis
- **Interface Segregation**: Interfaces coesas e específicas
- **Dependency Inversion**: Dependa de abstrações

## 📋 Definition of Done (DoD)

### ✅ Critérios Obrigatórios
- [ ] **Clean Code**: Nomes expressivos, funções pequenas
- [ ] **Testes**: Cobertura mínima 80% (unit + integration)
- [ ] **SOLID**: Princípios implementados corretamente
- [ ] **Security**: Scan sem vulnerabilidades críticas
- [ ] **Performance**: Benchmarks dentro dos SLAs
- [ ] **Review**: Aprovação de 2+ desenvolvedores
- [ ] **CI/CD**: Pipeline verde em todos os stages

## 🧪 Estratégia de Testes

### Testing Pyramid (80/15/5)
```
🔺 E2E Tests (5%)        - Jornadas completas do usuário
🔺 Integration (15%)      - APIs + Database + Services  
🔺 Unit Tests (80%)       - Business Logic + Domains
```

### Padrões de Teste
```go
// Table-driven tests obrigatório
func TestDeputadoValidator_Validate(t *testing.T) {
    tests := []struct {
        name      string
        input     *domain.Deputado
        wantError bool
        errorCode string
    }{
        // casos de teste...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // implementação do teste
        })
    }
}
```

## 🚀 Microsserviços

```
📋 deputados-service    → Gestão de parlamentares
🗳️ atividades-service   → Proposições, votações, presença
💰 despesas-service     → Análise de gastos parlamentares
👥 usuarios-service     → Autenticação, perfis, gamificação
💬 forum-service        → Discussões e interação cidadã
🔄 ingestao-service     → ETL dados Câmara/TSE
🤖 ia-service          → Moderação Gemini e assistente
```

## 📝 Convenções de Código

### Naming (Go)
```go
// ✅ Funções exportadas - PascalCase
func BuscarDeputadoPorID(ctx context.Context, id uuid.UUID) (*domain.Deputado, error)

// ✅ Variáveis/funções internas - camelCase
func validarCPFDeputado(cpf string) error

// ✅ Constantes - PascalCase com prefixo
const (
    MaxTentativasRequisicaoAPI = 3
    TimeoutPadraoHTTP         = 30 * time.Second
)

// ✅ Errors - Err + descrição
var (
    ErrDeputadoNaoEncontrado = errors.New("deputado não encontrado")
    ErrDadosInvalidos       = errors.New("dados inválidos")
)
```

### Error Handling
```go
// ✅ Custom errors com contexto
type DeputadoError struct {
    Op   string    // Operação que falhou
    ID   uuid.UUID // ID do deputado
    Err  error     // Erro original
    Code string    // Código para client
}

// ✅ Error wrapping obrigatório
if err != nil {
    return fmt.Errorf("erro ao buscar deputado %s: %w", id, err)
}
```

## 🔒 Segurança & Performance

### Rate Limiting
```go
// Middleware obrigatório para todas as APIs
middleware.RateLimit(100, time.Hour) // 100 req/hora por IP
```

### Logs Estruturados
```go
// slog obrigatório para logs
log.Info("deputado criado com sucesso",
    slog.String("id", deputado.ID.String()),
    slog.String("nome", deputado.Nome),
    slog.Duration("tempo", time.Since(start)))
```

## 🎨 Frontend (Next.js 15)

### Estrutura
```
/frontend/
├── app/                   # App Router
├── components/
│   ├── ui/               # Shadcn/ui components
│   ├── features/         # Feature-specific
│   └── layout/           # Header, Footer, Sidebar
├── lib/
│   ├── api.ts            # TanStack Query
│   └── auth.ts           # NextAuth.js
└── types/                # TypeScript definitions
```

### Acessibilidade (WCAG 2.1 AA)
- Contraste mínimo 4.5:1
- Navegação completa via teclado
- Textos alternativos obrigatórios
- Suporte a leitores de tela

### Mobile-First (OBRIGATÓRIO)
- **Contexto**: 70% dos brasileiros acessam via smartphone
- **Breakpoints**: Mobile (375px) → Tablet (768px) → Desktop (1024px+)
- **Touch targets**: Mínimo 44px x 44px para botões/links
- **Typography**: Base 16px+ (evita zoom automático no mobile)
- **Performance**: Bundle <200KB, images WebP + lazy loading
- **Layout**: Grid responsivo (`grid-cols-1 md:grid-cols-2 lg:grid-cols-3`)

```tsx
// ✅ Pattern Mobile-First obrigatório
<button className="
  w-full py-3 px-4 text-base        // Mobile: botão full-width, touch-friendly
  md:w-auto md:px-6                 // Desktop: width auto, padding maior
  bg-blue-700 text-white rounded-lg // Core styles
  focus:ring-4 focus:ring-blue-300  // Acessibilidade
">
  Buscar Deputados
</button>

// ✅ Cards responsivos
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
  <DeputadoCard />
</div>

// ✅ Navigation mobile com drawer
<nav className="md:hidden">
  <button aria-label="Abrir menu">
    <Menu className="h-6 w-6" />
  </button>
</nav>
```

## 📊 Dados da Câmara

### API Base: `https://dadosabertos.camara.leg.br/api/v2/`

#### Endpoints Principais
- `GET /deputados` - Lista deputados (filtros: UF, partido, legislatura)
- `GET /deputados/{id}` - Dados cadastrais completos
- `GET /deputados/{id}/despesas` - Cota parlamentar detalhada
- `GET /proposicoes` - Proposições com filtros avançados
- `GET /votacoes` - Votações e votos individuais

#### Rate Limiting API
- **Limite**: 100 requisições/minuto
- **Implementar**: Circuit breaker + retry com backoff exponencial

---

## 📚 Documentação Adicional

Para detalhes específicos, consulte:
- **Arquitetura**: `.github/docs/architecture.md`
- **API Reference**: `.github/docs/api-reference.md`  
- **API Official Dados Abertos Camara**: `.github/docs/api-docs.json` 
- **Integração API Câmara**: `.github/docs/camara-api-integration.md`
- **Business Rules**: `.github/docs/business-rules.md`
- **Testing Guide**: `.github/docs/testing-guide.md`
- **CI/CD Pipeline**: `.github/docs/cicd-guide.md`

### 🔧 Arquitetura & Performance
- **`sistema-ultra-performance.md`**: Sistema de 6 camadas de otimização implementado
- **`security-performance-best-practices.md`**: Lições do Gemini Code Assist e correções aplicadas
- **`gcp-deployment-decision.md`**: Decisões de infraestrutura e deployment

### 📋 Desenvolvimento & Qualidade  
- **`testing-guide.md`**: Estratégias de teste e pyramid 80/15/5
- **`coverage-improvement-plan.md`**: Plano para alcançar 80% de cobertura
- **`environment-variables-best-practices.md`**: Gestão segura de configurações

---

> **🎯 Objetivo**: Código limpo, testável, escalável e seguro para democratizar a transparência política no Brasil através de tecnologia de ponta.
