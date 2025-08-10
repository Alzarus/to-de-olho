# 🏛️ Tô De Olho - Plataforma de Transparência Política

> **TCC - Análise e Desenvolvimento de Sistemas**  
> **Autor**: Pedro Batista de Almeida Filho  
> **IFBA - Campus Salvador** | **Agosto 2025**

## 🎯 Visão Geral

O **"Tô De Olho"** é uma plataforma inovadora de transparência política que democratiza o acesso aos dados da Câmara dos Deputados, promovendo maior engajamento cidadão através de três núcleos fundamentais:

- 🌐 **Acessibilidade Universal**: Interface intuitiva para todos os usuários
- 👥 **Gestão Social**: Participação cidadã nas decisões públicas  
- 🎮 **Gamificação**: Sistema de pontos e conquistas para engajar usuários

## 🚀 Status do Projeto

| Fase | Status | Progresso |
|------|--------|-----------|
| 🏗️ **Setup Inicial** | 🔄 Em Andamento | 60% |
| 🏛️ **Core Backend** | ⏳ Pendente | 0% |
| 🎨 **Frontend Base** | ⏳ Pendente | 0% |
| 🤖 **IA & Analytics** | ⏳ Pendente | 0% |

## 📋 Próximos Passos (Setembro 2025)

### 1. 🛠️ Inicialização Rápida

```powershell
# Clonar e configurar o projeto
git clone https://github.com/alzarus/to-de-olho.git
cd to-de-olho

# Iniciar ambiente de desenvolvimento
make dev

# Executar bootstrap (primeira vez)
make bootstrap
```

### 2. 🏗️ Tarefas Prioritárias

#### ✅ **Concluído**
- [x] Estrutura de pastas do monorepo
- [x] Docker Compose para desenvolvimento
- [x] Configuração inicial Go modules
- [x] Scripts de bootstrap e automação
- [x] Makefile com comandos úteis

#### 🔄 **Em Andamento**
- [ ] **Setup dos Microsserviços Go**
  - [ ] `deputados-service` - Gestão de parlamentares
  - [ ] `atividades-service` - Proposições e votações
  - [ ] `despesas-service` - Análise de gastos
  - [ ] `usuarios-service` - Autenticação e perfis
  - [ ] `forum-service` - Discussões cidadãs

#### ⏳ **Próximas**
- [ ] **Integração API Câmara**
  - [ ] Client HTTP resiliente
  - [ ] Sistema de rate limiting (100 req/min)
  - [ ] Cache inteligente Redis
  - [ ] Jobs background para sync
- [ ] **Database Schema**
  - [ ] Migrações PostgreSQL
  - [ ] Seed de dados demo
  - [ ] Índices otimizados
- [ ] **Frontend Next.js 15**
  - [ ] Setup TypeScript + Tailwind
  - [ ] Componentes Shadcn/ui
  - [ ] Sistema de autenticação

## 🛠️ Stack Tecnológica

### Backend
- **Go 1.23+** - Microsserviços com Gin Framework
- **PostgreSQL 16** - Banco principal com particionamento
- **Redis 7** - Cache e sessões
- **RabbitMQ** - Mensageria assíncrona
- **Docker** - Containerização

### Frontend  
- **Next.js 15** - App Router + TypeScript
- **Tailwind CSS** - Styling responsivo
- **Shadcn/ui** - Componentes acessíveis
- **TanStack Query** - Estado e cache

### Integrações
- **Google Gemini AI** - Moderação e assistente educativo
- **API Câmara v2** - Dados oficiais deputados
- **TSE** - Validação de eleitores
- **Prometheus + Grafana** - Monitoramento

## 🏃‍♂️ Comandos Úteis

```powershell
# Desenvolvimento
make dev              # Inicia ambiente completo
make bootstrap        # Primeira inicialização
make logs            # Ver logs dos serviços

# Build e Deploy
make build-backend   # Compila microsserviços
make build-frontend  # Build Next.js
make test           # Executa todos os testes

# Banco de Dados
make migrate-up     # Aplica migrações
make seed          # Popula dados demo
make backup        # Backup do banco

# Monitoramento
make monitoring    # Abre dashboards
make check-health  # Verifica serviços
```

## 📁 Estrutura do Projeto

```
to-de-olho/
├── backend/
│   ├── services/          # Microsserviços Go
│   │   ├── deputados/     # Gestão parlamentares
│   │   ├── atividades/    # Proposições/votações  
│   │   ├── despesas/      # Análise gastos
│   │   ├── usuarios/      # Auth/perfis
│   │   └── forum/         # Discussões
│   └── shared/            # Código compartilhado
├── frontend/              # Next.js 15 app
├── infrastructure/        # Docker, K8s, monitoring
├── scripts/              # Automação e bootstrap
└── docs/                 # Documentação
```

## 🔗 Links Importantes

- 📖 [Roadmap Detalhado](./ROADMAP.md)
- 🤖 [Instruções IA](./copilot-instructions.md)  
- 📊 [API Docs](./api-docs.json)
- 🏛️ [API Câmara](https://dadosabertos.camara.leg.br/api/v2/)

## 🎓 Contexto Acadêmico

Este projeto é desenvolvido como Trabalho de Conclusão de Curso (TCC) para o curso de **Análise e Desenvolvimento de Sistemas** do **IFBA - Campus Salvador**.

**Objetivos Acadêmicos:**
- Aplicar conhecimentos de arquitetura de software
- Implementar sistema distribuído em microsserviços
- Integrar tecnologias modernas (Go, Next.js, IA)
- Promover impacto social através da tecnologia

---

**🌟 "Transformando dados políticos em engajamento cidadão"**
