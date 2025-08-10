# 🛣️ Roadmap de Desenvolvimento - Projeto "Tô De Olho"

> **Plataforma de Transparência Política da Câmara dos Deputados**
> 
> **Autor**: Pedro Batista de Almeida Filho  
> **Curso**: Análise e Desenvolvimento de Sistemas - IFBA  
> **Data de Início**: Agosto/2025

## 📋 Status Geral do Projeto

| Fase | Status | Progresso | Previsão de Conclusão |
|------|--------|-----------|----------------------|
| 🏗️ **Planejamento** | ✅ Concluído | 100% | Agosto/2025 |
| 🔧 **Setup Inicial** | 🔄 Em Andamento | 65% | Setembro/2025 |
| 🏛️ **Core Backend** | ⏳ Pendente | 0% | Outubro/2025 |
| 🎨 **Frontend Base** | ⏳ Pendente | 0% | Novembro/2025 |
| 🤖 **IA & Analytics** | ⏳ Pendente | 0% | Dezembro/2025 |
| 🎮 **Gamificação** | ⏳ Pendente | 0% | Janeiro/2026 |
| 🚀 **Deploy & Testes** | ⏳ Pendente | 0% | Fevereiro/2026 |

---

## 🎯 Objetivos Principais

### 📊 Três Núcleos Fundamentais
- [x] **Acessibilidade**: Interface intuitiva para todos os usuários
- [x] **Gestão Social**: Participação cidadã nas decisões públicas  
- [x] **Ludificação**: Gamificação para elevar interesse pela gestão pública

### 🌟 Características Principais
- [x] Linguagem oficial: Português Brasileiro (pt-BR)
- [x] Dados oficiais: API da Câmara dos Deputados + TSE
- [x] Interação cidadã: Fórum e contato direto deputado-cidadão
- [x] Sistema de pontos, conquistas e rankings

---

## ✅ **STATUS ATUAL - Agosto 2025**

### 🎉 **Concluído Hoje (10/08/2025)**

#### ✅ **Infraestrutura Base - 100% Concluída**
- ✅ Estrutura completa do monorepo criada
- ✅ Docker Compose configurado (PostgreSQL 16 + Redis 7 + RabbitMQ)
- ✅ Scripts de automação (PowerShell + Makefile)
- ✅ Go modules configurado com dependências
- ✅ Package.json do frontend Next.js 15
- ✅ Prometheus + Grafana para monitoramento
- ✅ README.md atualizado com instruções

#### ✅ **Arquivos Criados**
```
✅ docker-compose.dev.yml      # Infraestrutura de desenvolvimento
✅ go.mod                      # Dependências Go do workspace
✅ frontend/package.json       # Configuração Next.js 15
✅ scripts/bootstrap.ps1       # Script de inicialização
✅ scripts/init-databases.sh   # Setup múltiplos bancos
✅ Makefile                    # Comandos de automação
✅ Estrutura de pastas completa
```

### 🔄 **Próximos Passos Imediatos (Esta Semana)**

#### 1. **Testar Ambiente de Desenvolvimento**
```powershell
# Executar para validar setup
make dev
make bootstrap
```

#### 2. **Criar Primeiro Microsserviço** 
- 🎯 **deputados-service** (Prioridade 1)
- Estrutura hexagonal (domain, usecase, repository, handler)
- API REST básica para CRUD deputados
- Testes unitários

#### 3. **Setup Frontend Next.js**
- Configuração TypeScript + Tailwind CSS
- Componentes base Shadcn/ui
- Roteamento e layout principal

### 📊 **Progresso Geral Atualizado**

| Componente | Status | Progresso |
|------------|--------|-----------|
| **Infraestrutura** | ✅ Concluído | 100% |
| **Scripts Automação** | ✅ Concluído | 100% |
| **Documentação** | ✅ Concluído | 90% |
| **Microsserviços** | ⏳ Próximo | 0% |
| **Frontend Base** | ⏳ Próximo | 0% |
| **Integração API** | ⏳ Futuro | 0% |

### 🎯 **Meta da Semana (11-17 Agosto)**
**Objetivo**: Ter o primeiro microsserviço funcionando com dados reais da API da Câmara

---

## 🏗️ Fases de Desenvolvimento

### **FASE 1: Setup e Infraestrutura Inicial** 📅 Agosto-Setembro/2025

#### 🔧 Configuração do Ambiente
- [x] **Setup do Repositório**
  - [x] Estrutura de monorepo
  - [ ] Configuração do Git (branches, hooks)
  - [ ] Setup do GitHub Actions (CI/CD)
  - [x] Documentação inicial

- [x] **Infraestrutura Base**
  - [x] Docker Compose para desenvolvimento
  - [x] PostgreSQL 16 setup
  - [x] Redis para cache
  - [x] RabbitMQ para mensageria

- [x] **Script de Bootstrap (Cold Start)**
  - [x] Script de inicialização automática
  - [ ] Sincronização inicial da API Câmara (513 deputados)
  - [ ] Carga priorizada: Referências → Deputados → Atividades → Histórico
  - [ ] Sistema de cache hierárquico (Redis + PostgreSQL)
  - [ ] Rate limiting e recuperação de falhas
  - [ ] Monitoramento de progresso em tempo real
  - [ ] Seed de dados demo para desenvolvimento

#### 📦 Stack Tecnológico
- [x] **Backend**: Go 1.23+ com Gin Framework
- [x] **Frontend**: Next.js 15 + TypeScript + Tailwind CSS
- [x] **Database**: PostgreSQL 16 + Redis
- [x] **Queue**: RabbitMQ
- [ ] **AI**: Google Gemini SDK + MCP
- [x] **Monitoring**: Prometheus + Grafana

---

### **FASE 2: Core Backend Services** 📅 Setembro-Outubro/2025

#### 🏛️ Microsserviços Principais

##### 1. **deputados-service** 
- [ ] Estrutura base do serviço
- [ ] Models e domínio
- [ ] Repository layer (PostgreSQL)
- [ ] Business logic (use cases)
- [ ] HTTP handlers (REST API)
- [ ] Testes unitários

##### 2. **atividades-service**
- [ ] Gestão de proposições
- [ ] Sistema de votações
- [ ] Controle de presença parlamentar
- [ ] Integração com API da Câmara

##### 3. **despesas-service**
- [ ] Análise de gastos públicos
- [ ] Cota parlamentar
- [ ] Relatórios de transparência
- [ ] Detecção de anomalias

##### 4. **usuarios-service**
- [ ] Autenticação JWT + OAuth2
- [ ] Sistema de roles (RBAC)
- [ ] Perfis de usuário
- [ ] Validação TSE para eleitores

#### 🔗 Integrações Externas
- [ ] **API Câmara dos Deputados (v2)**
  - [ ] Client HTTP resiliente com retry e circuit breaker
  - [ ] Rate limiting (100 req/min)
  - [ ] Cache inteligente de dados frequentes
  - [ ] Sync incremental e background jobs
  - [ ] Monitoramento de health da API
  - [ ] Fallback para dados cached em caso de indisponibilidade

- [ ] **Endpoints Prioritários da Câmara**
  - [ ] `/deputados` - Lista completa de deputados ativos
  - [ ] `/deputados/{id}/despesas` - Gastos detalhados (últimos 6 meses)
  - [ ] `/deputados/{id}/eventos` - Presença em eventos (5 dias)
  - [ ] `/proposicoes` - Proposições dos últimos 30 dias
  - [ ] `/votacoes` - Votações dos últimos 30 dias
  - [ ] `/referencias/*` - Tabelas de lookup e validação

- [ ] **API TSE** (Validação de Eleitores)
  - [ ] Verificação de CPF válido
  - [ ] Validação regional por estado
  - [ ] Sistema anti-fraude para votações
  - [ ] Cache de validações frequentes

---

### **FASE 3: Frontend e Interface** 📅 Outubro-Novembro/2025

#### 🎨 Interface Base (Design Universal)
- [ ] **Setup Next.js 15**
  - [ ] App Router configuration
  - [ ] TypeScript setup completo
  - [ ] Tailwind CSS + design system
  - [ ] Shadcn/ui components

- [ ] **Acessibilidade Universal (WCAG 2.1 AA)**
  - [ ] Navegação por teclado completa
  - [ ] Compatibilidade com leitores de tela
  - [ ] Contraste mínimo 4.5:1
  - [ ] Fonte mínima 16px
  - [ ] Zoom até 200% sem perda de funcionalidade

- [ ] **Design Mobile-First**
  - [ ] Touch targets 44px mínimo
  - [ ] Progressive enhancement
  - [ ] Interface intuitiva para todos os níveis
  - [ ] Linguagem simples sem jargões

#### 📱 Páginas Principais
- [ ] **Dashboard Principal**
  - [ ] Visão geral dos deputados
  - [ ] Métricas regionais
  - [ ] Últimas atividades

- [ ] **Perfil do Deputado**
  - [ ] Dados pessoais e mandato
  - [ ] Performance parlamentar
  - [ ] Histórico de votações
  - [ ] Análise de gastos

- [ ] **Sistema de Busca**
  - [ ] Busca inteligente
  - [ ] Filtros avançados
  - [ ] Autocomplete
  - [ ] Resultados paginados

- [ ] **Área do Usuário**
  - [ ] Login/Registro
  - [ ] Perfil personalizado
  - [ ] Deputados favoritos
  - [ ] Histórico de atividades

#### 📊 Visualizações de Dados
- [ ] **Charts e Gráficos**
  - [ ] Recharts/D3.js integration
  - [ ] Gráficos interativos
  - [ ] Mapas do Brasil (regiões)
  - [ ] Heatmaps de atividade

---

### **FASE 4: Funcionalidades Sociais** 📅 Novembro-Dezembro/2025

#### 💬 Sistema de Fórum (Instagram-Style)
- [ ] **forum-service**
  - [ ] Estrutura de tópicos e threads
  - [ ] Sistema de moderação IA + humana
  - [ ] Notificações em tempo real
  - [ ] WebSockets para chat

- [ ] **Sistema de Comentários Sociais**
  - [ ] Comentários hierárquicos (3 níveis)
  - [ ] Sistema de likes/reactions
  - [ ] Menções @username
  - [ ] Hashtags #tema
  - [ ] Notificações para respostas
  - [ ] Histórico de edições

- [ ] **Interação Deputado-Cidadão**
  - [ ] Canal direto de comunicação
  - [ ] Q&A sessions
  - [ ] Explicação de votos
  - [ ] Feedback dos eleitores
  - [ ] Stories parlamentares

#### 🗳️ Plebiscitos e Consultas
- [ ] **plebiscitos-service**
  - [ ] Sistema de votação seguro
  - [ ] Validação por região
  - [ ] Auditoria completa
  - [ ] Resultados em tempo real

- [ ] **Tipos de Consulta**
  - [ ] Plebiscitos locais
  - [ ] Consultas nacionais
  - [ ] Enquetes temáticas
  - [ ] Avaliação de deputados

---

### **FASE 5: IA e Analytics Avançados** 📅 Dezembro/2025-Janeiro/2026

#### 🤖 Integração com Gemini AI
- [ ] **ia-service**
  - [ ] SDK do Google Gemini
  - [ ] Sistema de moderação automática
  - [ ] Assistente educativo
  - [ ] Análise preditiva

#### 🛡️ Moderação Inteligente
- [ ] **Sistema Anti-Toxicidade**
  - [ ] Detecção de discurso de ódio
  - [ ] Filtro de spam
  - [ ] Classificação de sentimento
  - [ ] Sugestões de melhoria

#### 📈 Analytics e Insights
- [ ] **analytics-service**
  - [ ] Dashboard regional interativo
  - [ ] Métricas em tempo real
  - [ ] Alertas automáticos
  - [ ] Relatórios personalizados

#### 🔍 Sistema de Alertas
- [ ] **alertas-service**
  - [ ] Gastos suspeitos
  - [ ] Mudanças de posição
  - [ ] Baixa presença parlamentar
  - [ ] Novas proposições relevantes

---

### **FASE 6: Gamificação e Engajamento** 📅 Janeiro/2026

#### 🎮 Sistema de Pontos
- [ ] **Mecânicas de Ludificação**
  - [ ] Sistema de pontos por atividade
  - [ ] Badges e conquistas
  - [ ] Rankings por categoria
  - [ ] Progressão de níveis

#### 🏆 Elementos Gamificados
- [ ] **Conquistas (Badges)**
  - [ ] 🏛️ Fiscal Ativo
  - [ ] 🗳️ Eleitor Informado
  - [ ] 💬 Voz Cidadã
  - [ ] 📊 Analista
  - [ ] 🎯 Vigilante

- [ ] **Desafios e Eventos**
  - [ ] Desafios mensais
  - [ ] Quiz educativo
  - [ ] Competições regionais
  - [ ] Eventos especiais

---

### **FASE 7: Deploy e Otimização** 📅 Fevereiro/2026

#### 🚀 Infraestrutura de Produção
- [ ] **Containerização**
  - [ ] Dockerfiles otimizados
  - [ ] Docker Compose production
  - [ ] Multi-stage builds
  - [ ] Health checks

- [ ] **Kubernetes Setup**
  - [ ] Deployment manifests
  - [ ] Services e Ingress
  - [ ] ConfigMaps e Secrets
  - [ ] Horizontal Pod Autoscaler

#### 🔍 Monitoring e Observabilidade
- [ ] **Métricas e Logs**
  - [ ] Prometheus setup
  - [ ] Grafana dashboards
  - [ ] Structured logging
  - [ ] Distributed tracing

#### 🧪 Testes e Qualidade
- [ ] **Cobertura de Testes**
  - [ ] Testes unitários (>80%)
  - [ ] Testes de integração
  - [ ] Testes end-to-end
  - [ ] Performance testing

#### 🔐 Segurança
- [ ] **Security Hardening**
  - [ ] HTTPS/TLS configurado
  - [ ] Rate limiting
  - [ ] Input validation
  - [ ] Security headers
  - [ ] Vulnerability scanning

---

## 📊 Estimativas de Volume de Dados (API Câmara)

### 🏛️ Dados Principais da Câmara dos Deputados

| Tipo de Dado | Volume Estimado | Frequência | Endpoint Principal |
|--------------|-----------------|------------|-------------------|
| **Deputados Ativos** | ~513 registros | Estático | `/deputados` |
| **Proposições/Mês** | ~1.500 novas | Diária | `/proposicoes` |
| **Votações/Mês** | ~200-300 | Semanal | `/votacoes` |
| **Eventos/Semana** | ~50-100 | Diária | `/eventos` |
| **Despesas/Deputado/Mês** | ~20-50 itens | Mensal | `/deputados/{id}/despesas` |
| **Discursos/Deputado/Semana** | ~5-10 | Semanal | `/deputados/{id}/discursos` |

### ⚡ Estratégia de Cold Start

#### **Fase 1: Estrutura Base (< 1 minuto)**
- Tabelas de referência (~200 registros)
- Estados, tipos de despesa, tipos de proposição
- Cache warming inicial

#### **Fase 2: Deputados Ativos (< 5 minutos)**
- 513 deputados da legislatura atual
- Dados cadastrais + órgãos + profissões
- ~1.500 requisições total

#### **Fase 3: Dados Recentes (< 30 minutos)**
- Despesas dos últimos 6 meses (~15.000 registros)
- Proposições dos últimos 30 dias (~1.500 registros)
- Votações dos últimos 30 dias (~300 registros)
- Eventos da semana (~100 registros)

#### **Fase 4: Histórico Completo (Background - 2-4 horas)**
- Dados históricos completos dos deputados
- Tramitações de proposições
- Histórico de mandatos externos
- Total estimado: ~200.000 registros

### 🚨 Limitações da API
- **Rate Limit**: 100 requisições/minuto
- **Itens por página**: Máximo 100, padrão 15
- **Dados por ano**: Algumas consultas limitadas ao mesmo ano
- **Timeout**: Requisições podem demorar em horários de pico

---

### 🎯 KPIs Técnicos
| Métrica | Meta | Status Atual |
|---------|------|--------------|
| **Cobertura de Testes** | >80% | 0% |
| **Performance API** | <200ms | - |
| **Uptime** | >99.5% | - |
| **Dados Atualizados** | Daily | - |

### 👥 KPIs de Negócio (Futuro)
| Métrica | Meta | Status |
|---------|------|--------|
| **Usuários Ativos** | 1000+ | - |
| **Deputados Verificados** | 50+ | - |
| **Consultas Realizadas** | 100+ | - |
| **Engajamento Médio** | 15min/sessão | - |

---

## 🚨 Riscos e Mitigações

### ⚠️ Riscos Técnicos
| Risco | Probabilidade | Impacto | Mitigação |
|-------|---------------|---------|-----------|
| **API Câmara Indisponível** | Média | Alto | Cache extensivo + fallback |
| **Sobrecarga de Dados** | Alta | Médio | Paginação + rate limiting |
| **Performance Frontend** | Média | Médio | Code splitting + CDN |
| **Segurança** | Baixa | Alto | Security reviews + audits |

### 📅 Riscos de Cronograma
| Risco | Probabilidade | Impacto | Mitigação |
|-------|---------------|---------|-----------|
| **Complexidade IA** | Alta | Alto | MVP simplificado primeiro |
| **Integração TSE** | Média | Médio | Validação manual temporária |
| **Testes Extensivos** | Média | Médio | Testes paralelos ao desenvolvimento |

---

## 📝 Notas de Desenvolvimento

### � **Comandos Disponíveis (Criados Hoje)**

```powershell
# === DESENVOLVIMENTO ===
make dev              # Inicia ambiente completo (Docker)
make bootstrap        # Bootstrap inicial com dados demo
make bootstrap-full   # Bootstrap completo (4 anos de dados)
make stop            # Para todos os serviços
make clean           # Remove containers e volumes

# === BUILD & TEST ===
make build-backend   # Compila todos os microsserviços
make build-frontend  # Build Next.js production
make test           # Executa todos os testes
make test-coverage  # Testes com relatório de cobertura
make lint           # Linting Go + TypeScript
make format         # Formata código

# === BANCO DE DADOS ===
make migrate-up     # Executa migrações
make migrate-down   # Desfaz última migração
make migrate-reset  # Reset completo do banco
make seed          # Popula dados demo
make backup        # Backup do PostgreSQL

# === MONITORAMENTO ===
make monitoring    # Abre dashboards (Grafana, Prometheus, RabbitMQ)
make check-health  # Verifica saúde dos serviços
make logs         # Logs dos serviços

# === UTILITÁRIOS ===
make install-tools # Instala ferramentas de dev
make git-hooks    # Configura git hooks
make docs         # Gera documentação da API
```

### 🌐 **URLs do Ambiente Local**
```
🚀 Aplicação Principal:  http://localhost:3000
📊 Grafana (Monitoring): http://localhost:3001 (admin:admin123)
🔥 Prometheus:           http://localhost:9090
🐰 RabbitMQ Management:  http://localhost:15672 (admin:admin123)
🗄️ PostgreSQL:           localhost:5432 (postgres:postgres)
⚡ Redis:                localhost:6379
```

### �📚 Recursos de Estudo
- [ ] API Câmara dos Deputados - Documentação completa
- [ ] Go best practices - Clean Architecture
- [ ] Next.js 15 - App Router patterns
- [ ] Google Gemini SDK - Documentation
- [ ] PostgreSQL optimization
- [ ] Kubernetes basics

### 🔧 Ferramentas de Desenvolvimento
- [x] VSCode + Go extension
- [x] Docker Desktop
- [ ] Postman/Insomnia (API testing)
- [ ] pgAdmin (PostgreSQL)
- [ ] Redis CLI
- [ ] kubectl

### 🎯 **PRÓXIMAS TAREFAS PRIORITÁRIAS**

#### **Semana 1 (11-17 Agosto): Primeiro Microsserviço**
```
🏛️ deputados-service
├── 📁 backend/services/deputados/
├── 📄 cmd/server/main.go          # Entry point
├── 📄 internal/domain/            # Entities
├── 📄 internal/usecase/           # Business logic  
├── 📄 internal/repository/        # Data access
├── 📄 internal/handler/           # HTTP/gRPC
└── 📄 internal/config/            # Configurações
```

#### **Semana 2 (18-24 Agosto): API Câmara Integration**
- Client HTTP resiliente
- Rate limiting (100 req/min)
- Cache Redis para dados frequentes
- Jobs background para sincronização

#### **Semana 3 (25-31 Agosto): Frontend Base**
- Setup Next.js 15 completo
- Componentes Shadcn/ui
- Layout principal
- Primeira página funcional

#### **Semana 4 (01-07 Setembro): Integração E2E**
- Frontend consumindo API
- Autenticação básica
- Deploy de desenvolvimento

---

## 📅 Cronograma Detalhado

```mermaid
gantt
    title Cronograma de Desenvolvimento - Tô De Olho
    dateFormat  YYYY-MM-DD
    section Setup
    Planejamento           :done, plan, 2025-08-01, 2025-08-31
    Infraestrutura Base    :infra, 2025-09-01, 2025-09-30
    
    section Backend
    Core Services          :backend, 2025-09-15, 2025-10-31
    Integrações Externas   :apis, 2025-10-15, 2025-11-15
    
    section Frontend
    Interface Base         :frontend, 2025-10-01, 2025-11-30
    Visualizações         :charts, 2025-11-01, 2025-11-30
    
    section Features
    Sistema Social        :social, 2025-11-15, 2025-12-31
    IA e Analytics        :ai, 2025-12-01, 2026-01-31
    Gamificação           :game, 2026-01-01, 2026-01-31
    
    section Deploy
    Produção              :deploy, 2026-02-01, 2026-02-28
```

---

## ✅ Checklist Geral

### 🏗️ Infraestrutura
- [ ] Repositório configurado
- [ ] CI/CD pipeline
- [ ] Ambiente de desenvolvimento
- [ ] Database setup
- [ ] Message queue

### 🔧 Backend Services
- [ ] deputados-service
- [ ] atividades-service  
- [ ] despesas-service
- [ ] usuarios-service
- [ ] forum-service
- [ ] plebiscitos-service
- [ ] analytics-service
- [ ] ia-service
- [ ] alertas-service

### 🎨 Frontend
- [ ] Next.js setup
- [ ] Design system
- [ ] Páginas principais
- [ ] Componentes reutilizáveis
- [ ] Charts e visualizações

### 🤖 Funcionalidades Avançadas
- [ ] IA Gemini integrada
- [ ] Sistema de moderação
- [ ] Analytics regionais
- [ ] Gamificação completa

### 🚀 Deploy
- [ ] Containerização
- [ ] Kubernetes
- [ ] Monitoring
- [ ] Segurança
- [ ] Testes de produção

---

## 🌟 Diferenciais Competitivos

### 🚀 Por que o "Tô De Olho" é Único?

#### **1. IA Conversacional Educativa**
- Assistente político pessoal com Gemini AI
- Explicação de projetos em linguagem simples
- Fact-checking automático
- Análise preditiva de votações

#### **2. Gamificação Cívica**
- RPG democrático com níveis de conhecimento
- Badges temáticas por especialização
- Missões cidadãs e desafios mensais
- Rankings regionais de participação

#### **3. Democracia Digital**
- Plebiscitos hiperlocais com validação TSE
- Simulador de impacto de leis
- Propostas colaborativas cidadão-deputado
- Orçamento participativo digital

#### **4. UX Social Media**
- Sistema de comentários estilo Instagram
- Stories parlamentares
- Live Q&A deputado-cidadão
- Feeds personalizados

### 🎯 Proposta de Valor

> **"Política como Rede Social, Educação como Jogo"**

**Não é apenas outro site de transparência. É a primeira rede social que transforma cada brasileiro em um fiscal ativo da democracia.**

---

**📧 Contato**: Pedro Batista de Almeida Filho - IFBA  
**📅 Última Atualização**: 10 de Agosto de 2025 - 21:30  
**🔄 Próxima Revisão**: 17 de Agosto de 2025 (Semana 1 concluída)  
**✅ Status Atual**: Setup Inicial 65% Concluído

---

> 🎯 **Objetivo**: Desenvolver uma plataforma completa de transparência política que democratize o acesso aos dados da Câmara dos Deputados, promovendo maior engajamento democrático através de tecnologia, gamificação e participação social.

> 🚀 **Progresso Hoje**: Infraestrutura base 100% configurada, monorepo estruturado, Docker Compose funcional, scripts de automação criados. **Próximo**: Implementar primeiro microsserviço (deputados-service).
