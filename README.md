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

```bash
# Clonar e configurar o projeto
git clone https://github.com/alzarus/to-de-olho.git
cd to-de-olho

# Iniciar ambiente de desenvolvimento
docker-compose -f docker-compose.dev.yml up -d

# Verificar se está funcionando
docker ps
```

### 2. 🏗️ Tarefas Prioritárias

#### ✅ **Concluído**
- [x] Estrutura de pastas do monorepo
- [x] Docker Compose para desenvolvimento
- [x] Configuração inicial Go modules
- [x] Scripts de bootstrap e automação
- [x] Makefile com comandos úteis
- [x] **Sistema de configuração centralizada**
- [x] **Melhores práticas de variáveis de ambiente**
- [x] **Sistema de migrações de banco**
- [x] **API básica de deputados funcionando**
- [x] **Frontend com integração à API**
- [x] **Cache Redis implementado**
- [x] **Rate limiting configurável**

#### 🔄 **Em Andamento**
- [ ] **Microsserviços Avançados**
  - [ ] `atividades-service` - Proposições e votações
  - [ ] `despesas-service` - Análise de gastos
  - [ ] `usuarios-service` - Autenticação e perfis
  - [ ] `forum-service` - Discussões cidadãs

#### ⏳ **Próximas**
- [ ] **Integração API Câmara Avançada**
  - [ ] Jobs background para sync completo
  - [ ] Webhook handlers
  - [ ] Data enrichment com IA
- [ ] **Features Frontend**
  - [ ] Dashboard interativo
  - [ ] Sistema de busca avançada
  - [ ] Componentes de visualização
- [ ] **Sistema de Autenticação**
  - [ ] OAuth2 com Google/GitHub
  - [ ] Perfis de usuário
  - [ ] Sistema de gamificação

## 🛠️ Stack Tecnológica

### Backend
- **Go 1.24+** - Microsserviços com Gin Framework
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

```bash
# Desenvolvimento
docker-compose up -d                              # Inicia ambiente completo
docker-compose ps                                 # Status dos serviços
docker-compose logs -f                            # Ver logs dos serviços
docker-compose down                               # Parar serviços

# Acesso direto aos serviços
docker exec -it todeolho-postgres psql -U admin -d todeolho
docker exec -it todeolho-redis redis-cli
```

## ⚙️ Configuração

### Variáveis de Ambiente

O projeto utiliza configuração centralizada e tipada. Copie o template:

```bash
# Backend
cp backend/.env.example backend/.env

# Frontend
cp frontend/.env.example frontend/.env
```

#### Variáveis Principais

```bash
# Servidor
PORT=8080
GIN_MODE=release
RATE_LIMIT_RPS=100

# Banco PostgreSQL (OBRIGATÓRIO)
POSTGRES_PASSWORD=sua_senha_segura

# API Câmara dos Deputados
CAMARA_CLIENT_RPS=2        # Requests por segundo (max: 100/min)
CAMARA_CLIENT_TIMEOUT=30s  # Timeout das requisições

# Redis Cache
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=            # Deixar vazio para desenvolvimento
```

**📋 Documentação Completa**: [Environment Variables Best Practices](.github/docs/environment-variables-best-practices.md)

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
