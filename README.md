# Tô De Olho

Plataforma de transparência sobre os senadores brasileiros, desenvolvida como Trabalho de Conclusão de Curso (TCC) em Análise e Desenvolvimento de Sistemas no IFBA.

Em produção: **[todeolho.org](https://todeolho.org)**

O sistema consolida dados abertos do Senado Federal e do Portal da Transparência. Proposições, presença em votações, gastos da cota parlamentar (CEAPS) e participação em comissões formam um ranking de efetividade ([metodologia](./METODOLOGIA.md)). Emendas parlamentares e a estrutura de gabinete são exibidas, mas não entram no ranking.

---

## O que o site oferece

- **Ranking** dos senadores em exercício, por mandato ou por ano, com a nota de cada critério.
- **Ficha do senador**: proposições (com nome popular das matérias), votações, gastos da cota, comissões, emendas e gabinete em números agregados (sem nomes nem salários).
- **Votações** nominais do Plenário, com filtros por tipo de matéria, votação secreta e resultado.
- **Comparador** de senadores, com gráficos filtráveis, alinhamento de votos e link compartilhável (`?ids=`).
- **Exportação** em CSV (pronto para o Excel em português) e JSON, e versão para impressão/PDF.
- **Metodologia** pública em [`/metodologia`](https://todeolho.org/metodologia) e em [`METODOLOGIA.md`](./METODOLOGIA.md), com o histórico de versões do cálculo.
- Link para o **[Quem Votar](https://todeolho.org/quemvotar)**, app irmão para as eleições de 2026.

## Fontes de dados

| Fonte | Uso |
|---|---|
| API de Dados Abertos do Senado (legislativo) | senadores, proposições, votações, comissões, Mesa Diretora |
| API de Dados Abertos do Senado (administrativo) | cota parlamentar (CEAPS) e estrutura de gabinete |
| Portal da Transparência (CGU) | emendas parlamentares |

---

## Stack

Monolito modular em Go com frontend Next.js desacoplado.

### Backend (`/backend`)

- **Go 1.27** com **Gin**
- **PostgreSQL 15** via **GORM** (tabelas criadas por `AutoMigrate` na subida)
- Módulos em `internal/`: `senador`, `proposicao`, `materia`, `votacao`, `ceaps`, `comissao`, `emenda`, `gabinete`, `ranking`, `acesso`, `scheduler`
- Clientes das APIs externas em `pkg/` (com novas tentativas em `pkg/retry`)
- Imagem final *distroless*

### Frontend (`/frontend`)

- **Next.js 16** (App Router) e **React 19**
- **TypeScript 5**
- **Tailwind CSS 4** + shadcn/ui
- **Recharts 3** para os gráficos

---

## Como rodar localmente

### Pré-requisitos

- [Go 1.27+](https://go.dev/)
- [Bun](https://bun.sh/) (ou Node.js 20+)
- [Docker](https://www.docker.com/) para o banco

### 1. Banco de dados

```bash
docker run --name pg-todeolho -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres:15-alpine
```

### 2. Backend

Variáveis de ambiente (ou `backend/.env`):

| Variável | Uso |
|---|---|
| `DATABASE_URL` | conexão com o Postgres, ex.: `postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable` |
| `TRANSPARENCIA_API_KEY` | chave do Portal da Transparência (emendas) |
| `SYNC_SECRET` | segredo das rotas `/api/v1/sync/*` (header `X-Sync-Secret`) |
| `PORT` | porta da API (padrão `8080`) |

```bash
cd backend
go mod download
go run cmd/api/main.go   # localhost:8080
```

Na subida, a API cria as tabelas e inicia o scheduler. Para carregar os dados com o banco vazio, rode o backfill:

```bash
curl -X POST -H "X-Sync-Secret: $SYNC_SECRET" http://localhost:8080/api/v1/sync/backfill
```

### 3. Frontend

```bash
cd frontend
bun install
BACKEND_URL=http://localhost:8080 bun run dev   # localhost:3000
```

O Next.js encaminha `/api/*` para o backend definido em `BACKEND_URL`. O padrão (`http://api:8080`) é o nome do serviço no Docker Compose.

### Testes

```bash
cd backend && go test ./...          # testes de repositório rodam só com TEST_DATABASE_URL (Postgres descartável)
cd frontend && bunx tsc --noEmit && bun run build
```

---

## Deploy

Produção numa VM Contabo, com Docker Compose (`docker-compose.contabo.yml`) atrás do Nginx Proxy Manager e da Cloudflare. O backend não é exposto: o frontend o alcança pela rede interna do Compose.

- **CI** (`.github/workflows/ci.yml`): testes e `go vet` do backend, e varredura de dependências e das imagens com Trivy.
- **Deploy** (`.github/workflows/deploy.yml`): a cada push no `master`, o workflow gera o `.env` a partir dos secrets do repositório, sobe os containers na VM e confere a saúde da API e do frontend.

Runbook completo em [`deploy-contabo.md`](./deploy-contabo.md). As chaves exigidas estão em [`.env.contabo.example`](./.env.contabo.example). O deploy anterior, no laboratório GSORT do IFBA, está documentado em [`deploy-gsort.md`](./deploy-gsort.md).

### Atualização dos dados

- **Backfill**: carga completa da legislatura atual (desde 01/02/2023), via `POST /api/v1/sync/backfill`.
- **Sync diário**: roda dentro da API às 03:00 (Brasília) e atualiza senadores, votações, cota, emendas, comissões, proposições e matérias, e recalcula o ranking. A estrutura de gabinete é recarregada uma vez por semana.

---

## Documentação

- [`METODOLOGIA.md`](./METODOLOGIA.md): cálculo do ranking, fórmulas, códigos de voto e histórico de versões.
- [`ROADMAP.md`](./ROADMAP.md): próximos passos.
- [`deploy-contabo.md`](./deploy-contabo.md): runbook de produção.
