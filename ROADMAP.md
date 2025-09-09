# 🛣️ Roadmap de Desenvolvimento - "Tô De Olho"

> **Plataforma de Transparência Política - Câmara dos Deputados**
> 
> **Autor**: Pedro Batista de Almeida Filho | **Curso**: ADS - IFBA  
> **Status**: Setembro 2025 | **Progresso Geral**: 80% MVP Backend

## � Status Atual do Projeto

| Componente | Status | Cobertura Testes | Próximo Marco |
|------------|--------|------------------|---------------|
| 🏗️ **Infraestrutura** | ✅ Completo | 100% | - |
| 🔧 **Backend Core** | ✅ Funcional | ~85% | Otimizações |
| 🧪 **Testes & QA** | 🟡 Refinando | 85%+ | 90%+ target |
| 🎨 **Frontend** | ✅ MVP | Manual | Expansão |
| 🐳 **Docker/Deploy** | ✅ Funcional | - | Prod ready |

---

## ✅ Marcos Recentes (Agosto-Setembro 2025)

### � **Conquistas Principais**

#### ✅ **Arquitetura & Configuração (COMPLETO)**
- **Clean Architecture** implementada (Domain/Application/Infrastructure/Interfaces)
- **Sistema de configuração centralizada** com validação automática
- **Melhores práticas de env vars** documentadas e implementadas
- **Rate limiting configurável** (100 req/min padrão)
- **Cache Redis + PostgreSQL fallback** funcionando

#### ✅ **API Backend (FUNCIONAL)**
- **Endpoints REST funcionais**:
  - `GET /api/v1/health`
  - `GET /api/v1/deputados` (filtros: UF, partido, nome)
  - `GET /api/v1/deputados/:id`
  - `GET /api/v1/deputados/:id/despesas`
- **Cliente API Câmara resiliente** (retry, backoff, circuit breaker)
- **Fallback de dados** via PostgreSQL quando API externa falhar
- **CORS** configurado para frontend

#### ✅ **Frontend MVP (OPERACIONAL)**
- **Next.js 15 + TypeScript** configurado
- **Interface responsiva** com Tailwind CSS
- **Lista de deputados funcional** com fotos e dados
- **Sistema de filtros** (UF, partido, busca)
- **Modal de detalhes** do deputado
- **Integração backend** via Axios

#### ✅ **Infrastructure & DevOps**
- **Docker Compose funcional** (backend, frontend, PostgreSQL, Redis)
- **Scripts de automação** Windows/PowerShell
- **Sistema de migrações** PostgreSQL
- **Health checks** implementados
- **Documentação técnica completa**

### 📊 **Cobertura de Testes Atual**
```
✅ Domain Layer:           100.0% (entities/business logic)
✅ HTTP Handlers:          100.0% (REST endpoints)  
✅ Repository:             100.0% (data layer)
✅ Application Layer:       90.0% (use cases)
🟡 HTTP Client (Câmara):    83.9% (external integration)
🟡 Middleware:              84.6% (rate limiting/CORS)
🟡 Cache (Redis):           ~80%  (config migration)
⚠️ Infrastructure/DB:       51.4% (needs improvement)
❌ CMD Entry Points:         0.0% (integration tests pending)
```

---

## 🎯 Próximas Prioridades (Setembro-Outubro 2025)

### **🚨 CRÍTICO - Próximas 2 Semanas**

#### 1. **🔧 Estabilização de Testes** (URGENTE)
- Corrigir testes falhando em `cache` e `db`
- Migrar testes legados para nova configuração
- Atingir 90%+ cobertura consistente
- CI/CD pipeline com quality gates

#### 2. **📈 Observabilidade** (ALTA PRIORIDADE)
- Métricas Prometheus (`/metrics` endpoint)
- Logs estruturados com slog
- Health checks avançados
- Dashboards Grafana básicos

#### 3. **🔄 Pipeline de Ingestão** (ALTA PRIORIDADE)
- Sistema de jobs background (RabbitMQ/Go routines)
- Sincronização automática diária
- Backfill histórico configurável
- Monitoramento de ingestão

### **📅 Médio Prazo (Outubro-Novembro 2025)**

#### 4. **🎨 Frontend Expansion**
- Dashboard interativo com charts
- Busca avançada com Elasticsearch/PostgreSQL FTS
- Sistema de favoritos
- PWA básico

#### 5. **🔐 Autenticação & Usuários**
- OAuth2 (Google/GitHub)
- Sistema de perfis
- API de usuários
- Rate limiting por usuário

#### 6. **📊 Analytics & IA**
- Integração Google Gemini SDK
- Análises de gastos com anomalias
- Chatbot educativo básico
- Relatórios automatizados

---

## 🏗️ Arquitetura Atual

### **Backend (Go + Clean Architecture)**
```
cmd/
├── server/     # API REST (Gin)
└── ingestor/   # ETL jobs

internal/
├── domain/     # Business entities
├── application/# Use cases
├── infrastructure/
│   ├── cache/  # Redis
│   ├── db/     # PostgreSQL
│   ├── httpclient/  # API Câmara
│   └── repository/  # Data access
└── interfaces/
    ├── http/   # REST handlers
    └── middleware/  # Cross-cutting
```

### **Frontend (Next.js 15)**
```
src/
├── app/        # App Router
├── components/ # React components
└── lib/        # Utilities
```

### **Serviços (Docker)**
```bash
Backend API    → localhost:8080
Frontend Web   → localhost:3000  
PostgreSQL 16  → localhost:5432
Redis 7        → localhost:6379
```

---

## 📋 Checklist de Features

### ✅ **MVP Funcional (COMPLETO)**
- [x] API REST básica funcionando
- [x] Listagem de deputados com filtros
- [x] Cache Redis + fallback PostgreSQL  
- [x] Frontend responsivo
- [x] Docker Compose ambiente
- [x] Sistema de configuração
- [x] Documentação técnica

### 🔄 **Em Desenvolvimento**
- [ ] Pipeline de ingestão automática
- [ ] Métricas e observabilidade
- [ ] Cobertura de testes 90%+
- [ ] Dashboard frontend expandido

### ⏳ **Planejado**
- [ ] Sistema de autenticação
- [ ] IA/Analytics avançados
- [ ] Mobile PWA
- [ ] Deploy produção

---

## 🚀 Comandos de Desenvolvimento

### **Ambiente Completo**
```bash
# Iniciar todos os serviços
docker compose up -d

# Verificar status
docker compose ps

# Logs em tempo real
docker compose logs -f backend
```

### **Desenvolvimento Local**
```bash
# Apenas infraestrutura (BD + Cache)
docker compose up postgres redis -d

# Backend local
cd backend && go run cmd/server/main.go

# Frontend local  
cd frontend && npm run dev
```

### **Testes e QA**
```bash
# Testes com cobertura
cd backend && go test ./... -cover

# Build de produção
go build -o bin/server ./cmd/server
```

---

## 🎯 Métricas de Sucesso

### **Técnicas (Atuais)**
| Métrica | Meta | Status Atual |
|---------|------|--------------|
| **Cobertura Testes** | 90% | ~85% ✅ |
| **Performance API** | <200ms | <100ms ✅ |
| **Uptime** | 99.5% | 100% ✅ |
| **Build Time** | <2min | ~30s ✅ |

### **Funcionais (Próximas)**
| Feature | Status | Prioridade |
|---------|--------|------------|
| **513 deputados carregados** | ✅ | - |
| **Filtros funcionais** | ✅ | - |
| **Cache efetivo** | ✅ | - |
| **Ingestão automática** | 🔄 | Alta |
| **Dashboard analytics** | ⏳ | Média |

---

## 🔍 Riscos e Mitigações

| Risco | Probabilidade | Mitigação |
|-------|---------------|-----------|
| **API Câmara instável** | Média | Cache extensivo + fallback ✅ |
| **Complexidade testes** | Baixa | Refatoração contínua 🔄 |
| **Performance frontend** | Baixa | Code splitting + lazy loading |
| **Deploy produção** | Média | Kubernetes + CI/CD pipeline |

---

## 📚 Recursos Técnicos

### **Documentação Disponível**
- [Environment Variables Best Practices](.github/docs/environment-variables-best-practices.md)
- [Architecture Guide](.github/docs/architecture.md)
- [API Reference](.github/docs/api-reference.md)
- [Business Rules](.github/docs/business-rules.md)

### **URLs de Desenvolvimento**
- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080/api/v1  
- **Health Check**: http://localhost:8080/api/v1/health
- **Deputados**: http://localhost:8080/api/v1/deputados?siglaUf=SP

---

**📅 Última Atualização**: 8 de Setembro de 2025  
**🎯 Status**: MVP Backend funcional, Frontend básico operacional, Testes em refinamento  
**⚡ Próximo Marco**: Pipeline de ingestão + Observabilidade completa

---

## 📝 Notas Técnicas

### **Stack Tecnológico Atual**
- **Backend**: Go 1.24 + Gin + Clean Architecture
- **Frontend**: Next.js 15 + TypeScript + Tailwind CSS
- **Database**: PostgreSQL 16 + Redis 7
- **Infrastructure**: Docker Compose + Scripts PowerShell
- **APIs**: Câmara dos Deputados v2 (rate limited)

### **Comandos Essenciais**
```bash
# Ambiente completo
docker compose up -d

# Apenas desenvolvimento
docker compose up postgres redis -d
cd backend && go run cmd/server/main.go  # Terminal 1
cd frontend && npm run dev              # Terminal 2

# Testes
cd backend && go test ./... -cover
```

### **URLs Principais**
- **Frontend**: http://localhost:3000
- **API**: http://localhost:8080/api/v1
- **Health**: http://localhost:8080/api/v1/health
- **Deputados**: http://localhost:8080/api/v1/deputados?siglaUf=SP&itens=10

### **Objetivos Finais**
> **Visão**: Democratizar acesso aos dados da Câmara dos Deputados através de interface intuitiva, gamificação cívica e participação social.

**Núcleos**: Acessibilidade Universal + Gestão Social + Ludificação Democrática
