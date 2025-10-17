# Tô De Olho - Plataforma de Transparência Política

Trabalho de Conclusão de Curso – Tecnologia em Análise e Desenvolvimento de Sistemas (IFBA, Campus Salvador). Autor: Pedro Batista de Almeida Filho. Atualização em outubro de 2025.

## Visão geral

O projeto Tô De Olho tem como objetivo ampliar o acesso público aos dados da Câmara dos Deputados. A plataforma reúne pipeline de ingestão, motores de analytics e uma interface web responsiva para consulta e comparação de informações parlamentares.

## Componentes principais

- **Backend**: Go 1.24+, arquitetura limpa e domínio orientado a contexto, PostgreSQL 16, Redis 7, fila de sincronização e políticas de resiliência (rate limiting, circuit breaker, retries exponenciais).
- **Frontend**: Next.js 15, TypeScript, Tailwind CSS, princípios mobile-first e conformidade com WCAG 2.1 AA.
- **Ingestão de dados**: backfill histórico inteligente, sincronização incremental diária, checkpoints por entidade e métricas de monitoramento.

## Execução com Docker Compose

```bash
git clone https://github.com/alzarus/to-de-olho.git
cd to-de-olho

# Configurar variáveis de ambiente (OBRIGATÓRIO)
cp .env.example .env
# Edite o arquivo .env com suas configurações específicas

# Iniciar os serviços
docker compose up -d --build
```

O conjunto de serviços disponibiliza:

- Frontend: http://localhost:3000
- API backend: http://localhost:8080
- Interface administrativa do banco: http://localhost:8081
- Verificação de saúde da API: http://localhost:8080/health

## Situação atual do pipeline de dados (out/2025)

- O histórico de despesas encontra-se temporariamente indisponível porque a tabela `despesas` exige as colunas `cod_tipo_documento` e `valor_documento`. A migration `014_alter_despesas_add_columns.sql` e o `DespesaRepository` atualizado corrigem o problema.
- Após aplicar a migration, execute novamente os serviços `ingestor` e `scheduler` para repopular as tabelas.
- Demais entidades (deputados, proposições, votações, partidos) estão preparadas para ingestão e analytics, aguardando o restabelecimento das despesas para completar os dashboards.

## Monitoramento e diagnóstico

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

## Personalização opcional

As variáveis de ambiente podem ser copiadas de `.env.example`. Exemplos frequentes:

- `BACKFILL_START_YEAR`: ano inicial do backfill histórico.
- `CAMARA_CLIENT_RPS` e `CAMARA_CLIENT_BURST`: limites de requisições à API da Câmara.
- `SCHEDULER_PARALLEL_WORKERS`: número de workers simultâneos na sincronização incremental.

## Desenvolvimento local sem Docker

```bash
# Backend
cd backend
go mod tidy
go run cmd/server/main.go

# Frontend
cd frontend
npm install
npm run dev
```

## Testes

```bash
# Backend
cd backend
go test ./...

# Frontend
cd frontend
npm run test
```

## Documentação complementar

- [.github/docs/architecture.md](.github/docs/architecture.md) – arquitetura e padrões de projeto.
- [.github/docs/api-reference.md](.github/docs/api-reference.md) – descrição das APIs internas.
- [.github/docs/business-rules.md](.github/docs/business-rules.md) – regras de negócio consolidadas.
- [.github/docs/testing-guide.md](.github/docs/testing-guide.md) – estratégia de testes e metas de cobertura.

## Próximos marcos

Consulte o arquivo `ROADMAP.md` para o planejamento detalhado. Em resumo, as próximas entregas incluem: ajuste do pipeline de despesas, analytics de votações agregados, melhorias na experiência do usuário, preparação para deploy GCP e implementação do assistente IA Gemini.

## Licença

Projeto distribuído sob a licença MIT. Consulte o arquivo [LICENSE](LICENSE) para informações detalhadas.