# Tô De Olho - Plataforma de Transparência Política

Trabalho de Conclusão de Curso – Tecnologia em Análise e Desenvolvimento de Sistemas (IFBA, Campus Salvador). Autor: Pedro Batista de Almeida Filho. Atualização em outubro de 2025.

## Sumário

- [Visão Geral](#visão-geral)
- [Arquitetura em Alto Nível](#arquitetura-em-alto-nível)
- [Stack Tecnológica](#stack-tecnológica)
- [Pré-requisitos](#pré-requisitos)
- [Setup Rápido](#setup-rápido)
- [Estrutura do Repositório](#estrutura-do-repositório)
- [Fluxo de Desenvolvimento](#fluxo-de-desenvolvimento)
- [Qualidade e Conformidade](#qualidade-e-conformidade)
- [Monitoramento e Observabilidade](#monitoramento-e-observabilidade)
- [Documentação Complementar](#documentação-complementar)
- [Próximos Marcos](#próximos-marcos)
- [Licença](#licença)

## Visão Geral

O **Tô De Olho** amplia o acesso público aos dados da Câmara dos Deputados combinando ingestão contínua, analytics e uma interface web responsiva. Seus pilares são:

- Transparência: consolidação de dados parlamentares legíveis para o cidadão.
- Engajamento social: ranking, gamificação e fóruns que estimulam participação.
- Escalabilidade: microsserviços isolados, filas assíncronas e camadas de cache.

## Arquitetura em Alto Nível

- **Ingestão inteligente**: combina backfill histórico, sincronização incremental diária e checkpoints por entidade.
- **Motores de analytics**: serviços Go processam votações, presenças e despesas, com reprocessamento seletivo sob demanda.
- **Experiência web**: frontend em Next.js 15 com foco mobile-first, acessibilidade e uso de TanStack Query.
- **Observabilidade**: métricas com Prometheus/Grafana e logs estruturados.

O diagrama completo está em `.github/docs/architecture.md`.

## Stack Tecnológica

| Camada | Tecnologias | Observações |
| --- | --- | --- |
| Backend | Go 1.24+, PostgreSQL 16, Redis 7, RabbitMQ | Clean Architecture + DDD, policies de resiliência (rate limiting, circuit breaker, retry exponencial). |
| Frontend | Next.js 15, TypeScript, Tailwind CSS | Acessibilidade WCAG 2.1 AA, mobile-first, shadcn/ui. |
| Dados | API Câmara, API TSE | Uso de ETL dedicado e smart backfill. |
| Observabilidade | Prometheus, Grafana, structured logs via slog | Alertas a partir de SLOs de ingestão e API. |
| AI & Moderação | Google Gemini SDK, MCP | Moderação de conteúdo e assistente educativo. |

## Pré-requisitos

- Docker Desktop 4.30+ e Docker Compose v2.
- Go 1.24+ instalado localmente (desenvolvimento sem Docker).
- Node.js 20 LTS + npm 10 (frontend).
- Make, Git e acesso à internet para APIs públicas da Câmara/TSE.

## Setup Rápido

### Variáveis de Ambiente

```bash
cp .env.example .env
# Ajuste segredos e limites de consumo da API da Câmara antes de iniciar os serviços
```

Variáveis importantes:

- `BACKFILL_START_YEAR`: define o recorte inicial do ETL histórico.
- `CAMARA_CLIENT_RPS` e `CAMARA_CLIENT_BURST`: limites praticados pela API da Câmara.
- `SCHEDULER_PARALLEL_WORKERS`: controla o nível de paralelismo do scheduler.

### Execução com Docker Compose

```bash
git clone https://github.com/alzarus/to-de-olho.git
cd to-de-olho
docker compose up -d --build
```

Serviços expostos:

- Frontend: http://localhost:3000
- API backend: http://localhost:8080
- Banco/adminer: http://localhost:8081
- Health check: http://localhost:8080/health

### Desenvolvimento Local sem Docker

```bash
# Backend
cd backend
go mod tidy
go run cmd/server/main.go

# Frontend
cd frontend
npm install
npm run dev

# Ambiente integrado (watch)
cd ..
make dev
```

## Estrutura do Repositório

```
backend/       # Serviços Go (domínios, infraestrutura, interfaces)
frontend/      # Aplicação Next.js (App Router, componentes reutilizáveis)
infrastructure/# Manifestos Prometheus/Grafana e observabilidade
scripts/       # Automação de deploy e inicialização de bancos
.github/docs/  # Documentação de arquitetura, APIs, testes e boas práticas
```

Consulte `ROADMAP.md` para visão macro do produto.

## Fluxo de Desenvolvimento

1. Abra um branch a partir de `dev`.
2. Execute `make dev` para subir backend, frontend e infraestrutura de apoio.
3. Implemente seguindo os padrões descritos em `.github/copilot-instructions.md`.
4. Cubra mudanças com testes (unidade, integração ou e2e conforme impacto).
5. Rode `go test ./...` no backend e `npm run test` no frontend antes de abrir PR.
6. Atualize documentação relevante e inclua migrações/seed quando necessário.

## Qualidade e Conformidade

- Cobertura mínima: 80% (unitária + integração) conforme `.github/docs/testing-guide.md`.
- Revisão obrigatória de dois mantenedores e pipeline verde no GitHub Actions.
- Segurança: siga `.github/docs/environment-variables-best-practices.md` e relatórios de scan.
- Performance: valide SLAs definidos em `sistema-ultra-performance.md`.

## Monitoramento e Observabilidade

```bash
# Status da ingestão inteligente
curl http://localhost:8080/api/v1/backfill/status
curl http://localhost:8080/api/v1/scheduler/status

# Histórico de execuções
curl http://localhost:8080/api/v1/backfill/executions
curl http://localhost:8080/api/v1/scheduler/executions

# Logs dos pipelines
docker compose logs -f ingestor
docker compose logs -f scheduler
```

Dashboards e alertas residem em `infrastructure/`.

## Documentação Complementar

- [.github/docs/architecture.md](.github/docs/architecture.md) – arquitetura e padrões de projeto.
- [.github/docs/api-reference.md](.github/docs/api-reference.md) – descrição das APIs internas.
- [.github/docs/business-rules.md](.github/docs/business-rules.md) – regras de negócio consolidadas.
- [.github/docs/testing-guide.md](.github/docs/testing-guide.md) – estratégia de testes e metas de cobertura.
- `gemini-code-review.md` – lições aprendidas com assistente IA.

## Próximos Marcos

Planejamento detalhado está em `ROADMAP.md`. Destaques: evolução do pipeline de despesas, analytics avançados de votações, melhoria contínua da UX, preparação para deploy na GCP e integração total do assistente IA Gemini.

## Licença

Projeto distribuído sob a licença MIT. Consulte o arquivo [LICENSE](LICENSE) para informações detalhadas.
