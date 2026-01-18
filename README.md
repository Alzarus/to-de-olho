# Tô De Olho (Código-Fonte)

Este diretório contém o código-fonte completo da plataforma **Tô De Olho**, uma ferramenta de transparência legislativa desenvolvida como Trabalho de Conclusão de Curso (TCC).

O sistema monitora a atividade dos senadores brasileiros, consolidando dados de gastos, votações e emendas em um ranking de efetividade.

---

## 🛠️ Stack Tecnológico

A aplicação segue a arquitetura **Monolito Modular** com frontend desacoplado.

### Backend (`/backend`)

- **Linguagem**: Go 1.21+
- **Framework Web**: Gin (Performance HTTP)
- **Banco de Dados**: PostgreSQL 15 (Relacional)
- **ORM**: GORM (Object-Relational Mapping)
- **Cache**: Redis (Rankings e sessões)
- **Infraestrutura**: Docker (Multi-stage build)

### Frontend (`/frontend`)

- **Framework**: Next.js 15 (App Router)
- **Linguagem**: TypeScript 5
- **Estilização**: Tailwind CSS 4 + Shadcn/UI
- **Gráficos**: Recharts (SVG interativo)

---

## 🚀 Como Rodar Localmente

### Pré-requisitos

- [Go 1.21+](https://go.dev/)
- [Bun 1.0+](https://bun.sh/) (ou Node.js 20+)
- [Docker](https://www.docker.com/) (para banco/cache)

### 1. Banco de Dados

Na raiz desta pasta, inicie os serviços de infraestrutura (caso tenha docker-compose configurado ou suba manualmente):

```bash
# Exemplo manual:
docker run --name pg-todeolho -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres:15
docker run --name redis-todeolho -p 6379:6379 -d redis:7
```

### 2. Backend

```bash
cd backend

# Instalar dependências
go mod download

# Rodar migrações e servidor
# Padrão: localhost:8080
go run cmd/api/main.go
```

> **Nota**: O sistema iniciará o `Scheduler` em background para sincronizar dados das APIs do Senado.

### 3. Frontend

```bash
cd frontend

# Instalar dependências
bun install

# Rodar servidor de desenvolvimento
# Padrão: localhost:3000
bun run dev
```

Acesse **http://localhost:3000** no seu navegador.

---

## 📦 Deploy (Produção)

A infraestrutura foi desenhada para **Google Cloud Run** (Serverless Container).

### Pipeline de CI/CD

O arquivo `.github/workflows/ci.yml` automatiza o processo:

1.  **Testes**: Executa `go test` em cada push na branch `master`.
2.  **Build**: Gera container Docker otimizado (Distroless image).
3.  **Publish**: Envia para o Google Container Registry.
4.  **Deploy**: Atualiza o serviço no Cloud Run.

### Estratégia de Ingestão de Dados

O sistema opera em modo híbrido:

1.  **Backfill**: Carga inicial massiva (histórico).
2.  **Scheduler**: Sincronização diária (incremental) embutida no binário do backend.

---

## 📚 Documentação Adicional

Para detalhes arquiteturais, consulte a pasta `../docs`:

- `adr-arquitetura-backend.md`: Decisões técnicas do backend.
- `stack-frontend.md`: Decisões de UI/UX.
- `implementation_plan.md`: Plano de implementação detalhado.
