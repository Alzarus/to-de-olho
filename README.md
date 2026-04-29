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

### Modo Integrado (Docker Compose)

A forma recomendada de subir o ambiente completo é usando o orquestrador nativo do projeto, presente na raiz (`docker-compose.yml`), passando o arquivo de ambiente desejado (ex: `.env.gsort`):

```bash
docker compose --env-file .env.gsort up -d
```

Isso levantará o PostgreSQL, a API, o Frontend Next.js na porta `3000`, e o daemon de rede da Cloudflare simultaneamente.

### Modo Desenvolvimento Individual

Caso prefira rodar as peças soltas:
```bash
# Exemplo manual de banco de dados:
docker run --name pg-todeolho -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres:15
# Redis não é mais obrigatório na arquitetura local simplificada
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

## 📦 Deploy (Produção / GSORT)

A infraestrutura atual foi desenhada para a arquitetura **on-premise** via **Docker Compose**, otimizada para o laboratório Gsort (IFBA), desativando os serviços antigos da Google Cloud. O roteamento externo é provido nativamente através de um túnel seguro **Cloudflare Zero Trust** (`cloudflared`), contornando eficientemente a falta de IP estático público e bloqueios de roteador/firewall (NAT) institucionais.

### Passos de Deploy

O ambiente consolida Banco de Dados, Backend, Frontend e Daemon de Túnel em um orquestrador unificado.

1.  Preencha as chaves de API necessárias e o token gerado pela Cloudflare no arquivo `.env.gsort`.
2.  Compile e suba todos os serviços em background a partir da pasta raiz:
    ```bash
    docker compose --env-file .env.gsort up -d --build
    ```
3.  Sendo a primeira inicialização local com banco zerado, engatilhe o mapeamento dos dados oficiais (*backfill*) rodando a rotina interna:
    ```bash
    docker exec todeolho-api /force_sync
    ```

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
