# 🛣️ Roadmap de Desenvolvimento - "Tô De Olho"

> **Plataforma de Transparência Política - Câmara dos Deputados**
> 
> **Autor**: Pedro Batista de Almeida Filho | **Curso**: ADS - IFBA  
> **Status**: Setembro 2025 | **Progresso Geral**: 90% MVP Backend

## 📊 Status Atual do Projeto

| Componente | Status | Cobertura Testes | Próximo Marco |
|------------|--------|------------------|---------------|
| 🏗️ **Infraestrutura** | ✅ Completo | - | - |
| 🔧 **Backend Core** | ✅ MVP | ~85% | Funcionalidades |
| 🧪 **Testes & QA** | ✅ Estável | 85%+ | 90%+ |
| 🎨 **Frontend** | ✅ Básico | Manual | Expansão |
| 🐳 **Docker/Deploy** | ✅ Funcional | - | Produção |

---

## ✅ Marcos Recentes (Agosto-Setembro 2025)

### 🏆 **Conquistas Principais**

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
- **Sistema de migrações** PostgreSQL (✅ go:embed CI/CD blocker CORRIGIDO)
- **Health checks** implementados
- **Documentação técnica completa**
- **CI/CD Pipeline** desbloqueado (`go vet` e `go build` passando)

### 📊 **Cobertura de Testes Atual (ATUALIZADA - SETEMBRO 2025)**
```
✅ Domain Layer:           100.0% (business logic sólida)
✅ HTTP Handlers:          100.0% (REST endpoints)  
✅ Repository:             100.0% (data access)
✅ Cache (Redis):           95.7% (configuração robusta)
✅ Application Layer:       90.0% (use cases)
✅ Middleware:              84.6% (CORS/rate limiting)
🟡 HTTP Client (Câmara):    83.9% (external API)
🟡 Infrastructure/DB:       32.4% (básico funcionando)
❌ Config Package:           0.0% (não testado)
❌ CMD Entry Points:         0.0% (não testado)
❌ Migrations:               0.0% (não testado)

TOTAL REALISTA: ~85% (sem falhas ativas) ✅ TODOS OS TESTES PASSANDO
```

---

## 🎯 Próximas Prioridades (Setembro-Outubro 2025)

### **🚨 CRÍTICO - Situação Real Atual**

#### ✅ **RESOLVIDO: Pipeline CI/CD Desbloqueado** (SETEMBRO 2025)
- **Problema**: `internal/infrastructure/migrations/migrator.go:15:12: pattern *.sql: no matching files found`
- **Causa**: `go:embed *.sql` falhando em ambiente CI/CD (diferente do local)
- **Solução**: Migração de arquivos SQL embedidos para SQL inline no código
- **Validação**: ✅ `go vet ./...` e `go build ./...` passando sem erros
- **Status**: 🟢 **CI/CD FUNCIONAL**

#### ✅ **RESOLVIDO: Testes Corrigidos** (SETEMBRO 2025)
- **Problema**: 5 testes falhando (4 cache Redis + 1 database PostgreSQL)
- **Causa Cache**: Incompatibilidade entre `REDIS_ADDR` vs `REDIS_HOST`/`REDIS_PORT`
- **Causa DB**: Teste usando variáveis `DB_*` mas código usando `POSTGRES_*`
- **Solução Cache**: Método `New()` agora suporta ambas as configurações
- **Solução DB**: Teste corrigido para usar variáveis corretas e restaurar estado
- **Validação**: ✅ `go test ./...` - todos os testes passando
- **Status**: 🟢 **TESTES 100% FUNCIONAIS**

#### 1. **🏗️ Expandir Cobertura de Testes** (PRÓXIMA PRIORIDADE)
- Adicionar testes para `config` package (0% → 80%+)
- Testes básicos para `cmd` entry points (0% → 50%+)
- Testes para sistema de `migrations` (0% → 70%+)
- **Meta**: Atingir 90%+ de cobertura geral

#### 2. **📈 Funcionalidades Básicas** (MÉDIA PRIORIDADE)
- Sistema de ranking/gamificação básico
- Filtros avançados de busca
- Análise de despesas (gráficos simples)
- Sistema de favoritos do usuário

#### 3. **🚀 Preparação para Produção** (BAIXA PRIORIDADE)
- Docker multi-stage builds otimizados
- Configuração de ambiente de produção
- Scripts de deployment automatizado
- Monitoramento básico (logs + metrics)

---

## 📈 Roadmap de Médio Prazo (Outubro-Dezembro 2025)

### **🔥 Features Prioritárias**
1. **Autenticação & Usuários** (OAuth2 + JWT)
2. **Sistema de Comentários** (moderação IA)
3. **Dashboard Analítico** (métricas + visualizações)
4. **API Rate Limiting Avançado** (por usuário)
5. **Sistema de Notificações** (email + push)

### **🛠️ Melhorias Técnicas**
1. **Monitoramento Completo** (Prometheus + Grafana)
2. **Cache Inteligente** (invalidação automática)
3. **Otimização de Performance** (lazy loading, pagination)
4. **Segurança Avançada** (OWASP compliance)
5. **Documentação Interativa** (Swagger/OpenAPI)

### **📱 Expansão de Plataformas**
1. **PWA** (Progressive Web App)
2. **Mobile-First** optimizations
3. **API Pública** para desenvolvedores
4. **Integração TSE** (dados eleições)
5. **Webhooks** para notificações

---

## 💡 Inovações Futuras (2026+)

### **🤖 Inteligência Artificial**
- Análise de sentimento em proposições
- Predição de resultados de votações  
- Detecção automática de conflitos de interesse
- Assistente virtual para navegação

### **📊 Analytics Avançados**
- Machine Learning para padrões de gastos suspeitos
- Análise de redes de relacionamento político
- Predição de impacto de proposições
- Dashboard preditivo para cidadãos

### **🌐 Expansão Nacional**
- Integração com Senado Federal
- Dados de câmaras municipais
- Transparência de governos estaduais
- Portal unificado de transparência

---

## 📋 Checklist de Finalização MVP

### Backend Core ✅
- [x] API REST funcional
- [x] Integração Câmara dos Deputados
- [x] Sistema de cache (Redis)
- [x] Fallback database (PostgreSQL) 
- [x] Rate limiting
- [x] Health checks
- [x] CORS configurado
- [x] Clean Architecture
- [x] Testes automatizados (85%+)
- [x] CI/CD pipeline funcional

### Frontend Core ✅
- [x] Interface responsiva
- [x] Lista de deputados
- [x] Sistema de filtros
- [x] Modal de detalhes
- [x] Integração com backend
- [x] Loading states
- [x] Error handling

### Infrastructure ✅
- [x] Docker Compose
- [x] Scripts de automação
- [x] Migrações de database
- [x] Documentação técnica
- [x] Health monitoring

### Próximos Passos 🔄
- [ ] Testes para módulos sem cobertura
- [ ] Sistema de autenticação
- [ ] Dashboard de métricas
- [ ] Deploy em produção
- [ ] Monitoramento avançado

---

> **🎯 Status Atual**: MVP Backend 90% completo | Frontend básico funcionando | Infraestrutura sólida
> 
> **🚀 Próximo Marco**: Cobertura de testes 90%+ e funcionalidades básicas de usuário

**Última Atualização**: Setembro 9, 2025 | **Responsável**: Pedro Almeida
