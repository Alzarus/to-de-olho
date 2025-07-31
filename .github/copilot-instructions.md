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

```

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
```

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
Monitoring:  Prometheus + Grafana
Security:    JWT + OAuth2 + Rate Limiting
```

### Microsserviços
```
📋 deputados-service    → Gestão de parlamentares e perfis públicos
🗳️  atividades-service  → Proposições, votações, presença parlamentar
💰 despesas-service     → Análise de gastos e cota parlamentar
👥 usuarios-service     → Autenticação, perfis e gamificação
💬 forum-service        → Discussões cidadãs e interação deputado-público
� plebiscitos-service  → Sistema de votações e consultas populares
�🔄 ingestao-service     → ETL dados Câmara/TSE (background jobs)
� analytics-service    → Métricas, rankings e insights regionais
🔍 search-service       → Busca inteligente de dados
🚨 alertas-service      → Notificações e alertas automáticos
```

### Comunicação
- **API Gateway**: Ponto único de entrada com rate limiting
- **gRPC**: Comunicação interna entre microsserviços
- **Message Queue**: Processamento assíncrono de dados
- **WebSockets**: Notificações em tempo real
- **REST API**: Interface pública para frontend

## 📡 Dados da Câmara dos Deputados

### Endpoints Principais da API (https://dadosabertos.camara.leg.br/api/v2)

#### Deputados
- `GET /deputados` - Lista deputados com filtros
- `GET /deputados/{id}` - Dados detalhados do deputado
- `GET /deputados/{id}/despesas` - Gastos com cota parlamentar
- `GET /deputados/{id}/discursos` - Pronunciamentos registrados
- `GET /deputados/{id}/eventos` - Participação em eventos
- `GET /deputados/{id}/historico` - Mudanças no mandato
- `GET /deputados/{id}/orgaos` - Comissões e órgãos
- `GET /deputados/{id}/profissoes` - Formação e experiência

#### Atividades Legislativas
- `GET /proposicoes` - Lista de proposições (PLs, PECs, etc.)
- `GET /proposicoes/{id}` - Detalhes da proposição
- `GET /proposicoes/{id}/autores` - Autores da proposição
- `GET /proposicoes/{id}/tramitacoes` - Histórico de tramitação
- `GET /proposicoes/{id}/votacoes` - Votações relacionadas

#### Votações
- `GET /votacoes` - Lista de votações
- `GET /votacoes/{id}` - Detalhes da votação
- `GET /votacoes/{id}/votos` - Votos individuais dos deputados
- `GET /votacoes/{id}/orientacoes` - Orientação dos partidos

#### Eventos e Presenças
- `GET /eventos` - Reuniões, sessões e audiências
- `GET /eventos/{id}/deputados` - Presença em eventos
- `GET /eventos/{id}/pauta` - Pauta deliberativa

#### Órgãos e Partidos
- `GET /orgaos` - Comissões e órgãos da Câmara
- `GET /partidos` - Partidos políticos
- `GET /blocos` - Blocos partidários

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
          go-version: '1.23'
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