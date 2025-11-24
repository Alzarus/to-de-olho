# 🤖 GitHub Copilot - Instruções Core

## ⚡ TL;DR Operacional

- [ ] Identifique se a tarefa é de **backend**, **frontend** ou **documentação** e leia os arquivos de referência antes de agir.
- [ ] Utilize `semantic_search` ou `file_search` para localizar implementações existentes e evitar duplicidade.
- [ ] Reaproveite padrões estabelecidos, mantenha o idioma em pt-BR e aplique princípios SOLID.
- [ ] Cubra mudanças com testes (table-driven no Go, testing library no Next.js) e garanta 80% de cobertura do módulo afetado.
- [ ] Finalize com checklist de DoD, incluindo atualização de documentação e verificação do pipeline.

## 1. Contexto Automático e Eficiência

### 1.1 Arquivos Essenciais
- **Frontend**: consulte o diretório `frontend/` para layout, componentes compartilhados e convenções de UI.
- **Backend**: consulte o diretório `backend/` para domínios, serviços, infraestrutura e integrações.
- `.github/copilot-instructions.md`: este guia principal.
- `.github/docs/api-docs.json`: documentação oficial da API da Câmara.

### 1.2 Ferramentas Suportadas
- `semantic_search`: localizar funções, testes ou padrões antes de criar algo novo.
- `file_search`: quando souber o nome do arquivo ou símbolo que deseja localizar.
- `mcp_docker`: utilize para acesso direto aos contêineres, verificação de logs e estado dos serviços (obrigatório para monitoramento).
- `microsoft/playwright-mcp`: utilize para testes end-to-end e validação visual do frontend.
- Ferramenta externa: #upstash/context7 para buscar documentação (ex.: Next.js, Go, Gemini SDK). Fluxo mínimo: `resolve-library-id` → `get-library-docs`, sempre filtrando pelo tópico necessário.

### 1.3 Workflow Inteligente
1. **Contexto Primeiro**: Antes de qualquer alteração, analise o contexto do projeto. Leia arquivos relacionados em `.github/` e no diretório de trabalho. Não assuma nada; verifique.
2. Analise a tarefa, classifique o escopo (backend, frontend, dados, docs).
3. Leia os arquivos já existentes e reutilize o estilo do projeto.
4. Implemente em passos pequenos, adicionando testes e comentários apenas quando necessários para clareza.
5. Rode ou descreva testes relevantes; não deixe lacunas sem justificar.
6. Documente mudanças e deixe próximos passos claros.

## 2. Visão do Projeto

O **Tô De Olho** democratiza dados da Câmara dos Deputados com foco em:
- **Acessibilidade**: interface inclusiva e mobile-first.
- **Gestão social**: fórum e engajamento cidadão.
- **Gamificação**: pontos, conquistas e rankings para incentivar participação.
- **IA aplicada**: Google Gemini para moderação e assistência educativa.

## 3. Arquitetura e Domínios

### 3.1 Clean Architecture + DDD
```go
// Estrutura por domínio de negócio
/backend/services/deputados/
├── cmd/server/                  # Entry points
├── internal/
│   ├── domain/                  # Entities, Value Objects, Aggregates
│   ├── application/             # Use Cases / Application Services
│   ├── infrastructure/          # Frameworks & Drivers
│   └── interfaces/              # Interface Adapters
├── pkg/                         # Código compartilhado público
└── tests/                       # Testes organizados por tipo
```

### 3.2 Princípios SOLID Obrigatórios
- Single Responsibility: cada componente com uma responsabilidade clara.
- Open/Closed: estender sem modificar comportamento estável.
- Liskov Substitution: subtipos substituíveis sem efeitos colaterais.
- Interface Segregation: contratos pequenos e coesos.
- Dependency Inversion: dependa de abstrações, injete implementações.

### 3.3 Microsserviços

```
📋 deputados-service    → Gestão de parlamentares
🗳️ atividades-service   → Proposições, votações, presença
💰 despesas-service     → Análise de gastos parlamentares
👥 usuarios-service     → Autenticação, perfis, gamificação
� forum-service        → Discussões e interação cidadã
� ingestao-service     → ETL dados Câmara/TSE
🤖 ia-service           → Moderação Gemini e assistente
```

## 4. Fluxos por Tipo de Tarefa

### 4.1 Implementação
- Leia requisitos, identifique camadas impactadas (domain, application, infra, UI).
- Use `semantic_search` para encontrar padrões similares em vez de criar do zero.
- Crie testes table-driven para Go ou spec focados em comportamento no frontend.
- Valide contratos (DTOs, interfaces) e atualize mocks/fakes.

### 4.2 Review e Correções
- Priorize bugs críticos, inconsistências com regras de negócio e regressões.
- Cite arquivo e linha ao apontar problemas; sugerir correções quando viável.
- Rodar testes relacionados ou explicar por que não foi possível.

### 4.3 Documentação & Pesquisa
- Atualize README/arquivos de docs quando o comportamento público muda.
- Recorra a #upstash/context7 quando precisar de documentação oficial ou exemplos externos.
- Mantenha linguagem em pt-BR e exemplos aderentes ao projeto.

## 5. Convenções de Código

### 5.1 Backend (Go)
```go
// Funções exportadas - PascalCase
func BuscarDeputadoPorID(ctx context.Context, id uuid.UUID) (*domain.Deputado, error)

// Funções/variáveis internas - camelCase
func validarCPFDeputado(cpf string) error

// Constantes - PascalCase com prefixo
const (
  MaxTentativasRequisicaoAPI = 3
  TimeoutPadraoHTTP         = 30 * time.Second
)

// Errors - prefixo Err + descrição em pt-BR
var (
  ErrDeputadoNaoEncontrado = errors.New("deputado não encontrado")
  ErrDadosInvalidos        = errors.New("dados inválidos")
)
```

#### Tratamento de Erros
```go
type DeputadoError struct {
  Op   string    // Operação que falhou
  ID   uuid.UUID // ID relacionado
  Err  error     // Erro original
  Code string    // Código para client
}

if err != nil {
  return fmt.Errorf("erro ao buscar deputado %s: %w", id, err)
}
```

### 5.2 Segurança & Performance
```go
// Rate limiting obrigatório em todas as APIs
middleware.RateLimit(100, time.Hour)

// Logs estruturados via slog
log.Info("deputado criado com sucesso",
  slog.String("id", deputado.ID.String()),
  slog.String("nome", deputado.Nome),
  slog.Duration("tempo", time.Since(start)))
```

#### Resiliência HTTP
- Configure `http.Client{Timeout: ...}` para limitar o tempo total de requisições externas; defina `ReadTimeout`/`WriteTimeout` em servidores HTTP.
- Propague `context.WithTimeout` a partir dos handlers e encerre rotinas internas quando `ctx.Done()` for disparado.
- Classifique erros transitórios via `errors.Is`/`Timeout()`/`Temporary()` para aplicar retries com backoff e circuit breakers.
- Ajuste `http.Transport` (por exemplo `MaxIdleConns`, `IdleConnTimeout`) ao lidar com alto throughput ou múltiplas integrações.

### 5.3 Frontend (Next.js 15)

```
/frontend/
├── app/                   # App Router
├── components/
│   ├── ui/                # shadcn/ui
│   ├── features/          # componentes por domínio
│   └── layout/            # Header, Footer, Sidebar
├── lib/
│   ├── api.ts             # TanStack Query
│   └── auth.ts            # NextAuth.js
└── types/                 # Tipagens compartilhadas
```

#### Acessibilidade e Mobile-First
- Contraste mínimo 4.5:1 e foco visível em todos os elementos clicáveis.
- Navegação por teclado obrigatória (aria-labels, roles).
- Touch targets >= 44px, fonte base 16px.
- Performance: bundle <200 KB, imagens WebP e lazy loading.

```tsx
<button className="
  w-full py-3 px-4 text-base
  md:w-auto md:px-6
  bg-blue-700 text-white rounded-lg
  focus:ring-4 focus:ring-blue-300
">
  Buscar Deputados
</button>

<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
  <DeputadoCard />
</div>
```

#### Data fetching e caching (App Router)
- Prefira componentes servidor `async` com `fetch` e ajuste o cache conforme o caso: `cache: 'force-cache'` (estático), `cache: 'no-store'` (dinâmico) ou `next: { revalidate: <segundos> }` para revalidação automática.
- Exporte `revalidate` ou `dynamic` em cada segmento quando precisar forçar comportamento estático/dinâmico global.
- Após mutações via Server Actions, chame `revalidatePath('/rota')` ou `revalidateTag('tag')` para manter a UI consistente.
- Para scripts globais, utilize `next/script` no layout raiz; isso garante carregamento único e evita bloqueio de renderização.

## 6. Dados da Câmara

### 6.1 API Base
`https://dadosabertos.camara.leg.br/api/v2/`

### 6.2 Endpoints Principais
- `GET /deputados`: lista e filtros (UF, partido, legislatura).
- `GET /deputados/{id}`: dados cadastrais e mandatos.
- `GET /deputados/{id}/despesas`: cota parlamentar detalhada.
- `GET /proposicoes`: proposições com filtros avançados.
- `GET /votacoes`: votações e votos individuais.

### 6.3 Resiliência
- Limite de 100 requisições/minuto.
- Implementar circuit breaker, retry exponencial com jitter, cache agressivo quando possível.

## 7. Qualidade e Testes

### 7.1 Definition of Done
- [ ] Clean Code: nomes claros, funções pequenas.
- [ ] Testes: cobertura mínima 80% no escopo alterado.
- [ ] SOLID aplicado nas camadas relevantes.
- [ ] Segurança: sem vulnerabilidades críticas.
- [ ] Performance: benchmarks dentro dos SLAs.
- [ ] Review: aprovação de 2 mantenedores.
- [ ] CI/CD: pipeline completo em verde.

### 7.2 Pirâmide de Testes (80/15/5)
```
🔺 E2E Tests (5%)        - Jornadas completas do usuário
🔺 Integration (15%)     - APIs + Database + Services
🔺 Unit Tests (80%)      - Business Logic + Domains
```

### 7.3 Padrões de Teste Go
```go
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

## 8. Documentação e Recursos

- `.github/docs/architecture.md`: arquitetura detalhada e padrões.
- `.github/docs/api-reference.md`: contratos das APIs internas.
- `.github/docs/camara-api-integration.md`: estratégias de integração.
- `.github/docs/business-rules.md`: regras de negócio consolidadas.
- `.github/docs/testing-guide.md`: padrão de testes e metas.
- `.github/docs/cicd-guide.md`: pipeline e quality gates.
- `sistema-ultra-performance.md`: estratégia de otimização em 6 camadas.
- `security-performance-best-practices.md`: lições de segurança e performance.
- `gcp-deployment-decision.md`: decisões de deploy e infraestrutura.
- `coverage-improvement-plan.md`: plano para atingir e manter 80% de cobertura.
- `environment-variables-best-practices.md`: gestão segura de configurações.
- `gemini-code-review.md`: boas práticas ao usar assistentes IA.

---

> 🎯 Objetivo: entregar código limpo, testável, escalável e seguro para democratizar a transparência política no Brasil.
