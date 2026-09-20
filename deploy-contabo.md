# Deploy do Tô De Olho na VM Contabo

Runbook da migração do Google Cloud Run para a VM Contabo, onde o projeto passa
a coexistir com `neuroanalises`, `quem-votar` e `blazebot-license`.

## 1. Topologia de destino

A VM já opera com **Nginx Proxy Manager** (NPM) em `/opt/proxy` como único dono
das portas 80/443, emitindo certificados Let's Encrypt automaticamente. Todos os
projetos se conectam a ele pela rede Docker compartilhada `proxy_default`.

Por isso **não usamos Caddy**: ele tentaria assumir as mesmas portas e quebraria
os serviços já em produção.

Roteamento alvo para `todeolho.org` (DNS atrás da Cloudflare, origem na VM):

| Rota | Container de destino | Projeto |
|---|---|---|
| `/` | `todeolho-web:3000` | Tô De Olho (este repositório) |
| `/quemvotar` | `quemvotar-web:8080` | Quem Votar |

O backend **não** é publicado diretamente: ele fica apenas na rede
`todeolho_internal` e é alcançado pelo `rewrite` de `/api/*` do Next.js
(`frontend/next.config.ts` → `BACKEND_URL=http://api:8080`). Isso reduz a
superfície exposta.

```
Internet → Cloudflare → NPM (proxy-app-1, :80/:443)
                          ├── /           → todeolho-web:3000 ─┐
                          └── /quemvotar  → quemvotar-web:8080 │
                                                                │ rede todeolho_internal
                                                     todeolho-api:8080 → todeolho-db:5432
```

## 2. Secrets exigidos no repositório

Configurar em **Settings → Secrets and variables → Actions**, no environment
`production`:

| Secret | Conteúdo |
|---|---|
| `CONTABO_SSH_KEY` | Chave **privada** do par de deploy (ver passo 3) |
| `CONTABO_HOST` | IP da VM — mantido fora do código por causa da Cloudflare |
| `CONTABO_USER` | Usuário SSH de deploy |
| `POSTGRES_USER` | Usuário do banco |
| `POSTGRES_PASSWORD` | Senha forte, distinta da usada em desenvolvimento |
| `POSTGRES_DB` | Nome do banco (`todeolho`) |
| `TRANSPARENCIA_API_KEY` | Chave do Portal da Transparência (CGU) |
| `SYNC_SECRET` | `openssl rand -hex 32` |

## 3. Chave SSH dedicada ao CI

Não reaproveitar a chave pessoal. Gerar um par exclusivo:

```bash
ssh-keygen -t ed25519 -C "github-actions-todeolho" -f ~/.ssh/todeolho_deploy -N ""
```

Publicar a chave pública na VM e guardar a privada no secret:

```bash
ssh-copy-id -i ~/.ssh/todeolho_deploy.pub <user>@<host>
cat ~/.ssh/todeolho_deploy        # conteúdo → secret CONTABO_SSH_KEY
```

## 4. Configuração no Nginx Proxy Manager

O painel escuta apenas em `127.0.0.1:81`. Acessar por túnel SSH:

```bash
ssh -L 8181:127.0.0.1:81 <user>@<host>
# abrir http://localhost:8181
```

Hoje o proxy host de `todeolho.org` encaminha **tudo** para `quemvotar-web:8080`.
A ordem abaixo troca o roteamento **sem derrubar o Quem Votar**:

1. **Antes de mexer no `/`**, no proxy host `todeolho.org` → aba *Custom
   Locations*, adicionar:
   - Location: `/quemvotar`
   - Forward Hostname: `quemvotar-web`
   - Forward Port: `8080`
2. Salvar e confirmar que `https://todeolho.org/quemvotar/` continua servindo o
   app Flutter.
3. Só então, na aba *Details*, trocar o destino padrão:
   - Forward Hostname: `todeolho-web`
   - Forward Port: `3000`
4. Manter *Force SSL*, *HTTP/2* e *HSTS* ligados. O certificado existente
   (`npm-4`, CN=`todeolho.org`) continua válido — nada de DNS ou TLS muda.

## 5. Primeiro deploy

O push em `master` dispara `test-backend` e, se passar, o job `deploy`, que
sincroniza o código via `rsync`, gera `/opt/todeolho/.env` a partir dos secrets,
faz o build na VM e valida a saúde dos containers.

Para acompanhar:

```bash
ssh <user>@<host>
cd /opt/todeolho
docker compose -f docker-compose.contabo.yml ps
docker compose -f docker-compose.contabo.yml logs -f api
```

## 6. Carga inicial dos dados (backfill)

O banco sobe vazio. O schema é criado pelo `AutoMigrate`
(`backend/cmd/api/main.go`), mas a carga **não** roda sozinha no startup — é
disparada por HTTP (`backend/internal/scheduler/scheduler.go`).

Com os containers no ar, disparar de dentro da VM:

```bash
source /opt/todeolho/.env
curl -X POST http://127.0.0.1:5350/api/v1/sync/backfill \
     -H "X-Sync-Secret: $SYNC_SECRET"
```

Responde `202` na hora e roda em segundo plano, em 6 passos (senadores →
votações → metadados → CEAPS → …). Acompanhar por:

```bash
docker compose -f /opt/todeolho/docker-compose.contabo.yml logs -f api | grep -i backfill
```

O passo de votações é o mais demorado e depende das APIs do Senado; ele é pulado
automaticamente se já houver dados. Se o backfill for interrompido, basta
repetir a chamada — cada etapa verifica o que já existe antes de baixar.

## 7. Rollback

O volume `postgres_data` é preservado entre deploys. Para voltar a uma versão
anterior da aplicação, reverter o commit em `master` — o pipeline reconstrói a
imagem a partir do código revertido. Para reverter apenas o roteamento, basta
apontar o `location /` do NPM de volta para `quemvotar-web:8080`.
