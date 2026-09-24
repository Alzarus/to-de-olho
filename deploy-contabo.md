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
| `CONTABO_KNOWN_HOSTS` | Chaves **públicas** do host (saída conferida do `ssh-keyscan`, ver passo 3) |
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

### Chave do host fixada (`CONTABO_KNOWN_HOSTS`)

O workflow não roda mais `ssh-keyscan` a cada deploy: aquilo aceitava qualquer
chave que o servidor apresentasse, inclusive a de um intermediário. A chave do
host fica fixada no secret `CONTABO_KNOWN_HOSTS` e o SSH usa
`StrictHostKeyChecking=yes`. Se o secret estiver vazio, o job falha com
mensagem explícita antes de conectar.

1. Numa máquina confiável, coletar as chaves públicas do host (use exatamente o
   mesmo valor de `CONTABO_HOST`, senão a linha não casa):

   ```bash
   ssh-keyscan -t ed25519,ecdsa,rsa <host> 2>/dev/null > known_hosts_todeolho
   ssh-keygen -lf known_hosts_todeolho      # fingerprints coletados
   ```

2. **Conferir** os fingerprints com os do próprio servidor, por um canal que não
   passe pela rede (console web da Contabo ou sessão já confiável):

   ```bash
   for f in /etc/ssh/ssh_host_*_key.pub; do ssh-keygen -lf "$f"; done
   ```

3. Se baterem, colar as três linhas de `known_hosts_todeolho` no secret
   `CONTABO_KNOWN_HOSTS` (environment `production`).

Se a VPS for reinstalada ou as chaves do host forem trocadas, o deploy passa a
falhar com `Host key verification failed` — é o comportamento desejado. Repetir
os passos acima para atualizar o secret.

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

## 5. Imagens no GHCR e primeiro deploy

Desde 24/09/2026 a VPS **não compila nada**. O build das imagens consumia CPU e
memória da mesma máquina que atende o público (num deploy, um JS de 1 MB levou
18 s para carregar). O fluxo agora é:

```
push em master
  └─ verificacao.yml (workflow reutilizável, o mesmo do CI de PR)
       ├─ backend: go mod verify, go vet, go test
       ├─ frontend: tsc --noEmit
       ├─ Trivy do repositório (CRITICAL)
       └─ imagens api e web: build no runner (Buildx + cache do Actions)
            → Trivy na imagem carregada (CRITICAL/HIGH com correção)
            → artefato da MESMA imagem escaneada
  └─ publicar: push em ghcr.io/alzarus/todeolho-{api,web}:<SHA>
  └─ deploy (VPS): docker login (token do job, via stdin) → pull api web
                   → up -d --no-build → health check → logout
  └─ marcar-producao: tag móvel :producao = o que está no ar
```

No pull request roda só a `verificacao` (build + Trivy, **sem** push). Se
qualquer etapa falhar, nada é publicado nem implantado.

O `IMAGE_TAG` (SHA do commit) é gravado no `/opt/todeolho/.env` e o
`docker-compose.contabo.yml` usa `ghcr.io/alzarus/todeolho-*:${IMAGE_TAG}`. O
compose de produção não tem `build:`; para construir localmente as mesmas
imagens, usar o override `docker-compose.contabo.build.yml`.

### Antes do primeiro deploy com GHCR

1. Criar o secret `CONTABO_KNOWN_HOSTS` (passo 3).
2. Em **Settings → Actions → General → Workflow permissions**, qualquer uma das
   opções serve: os workflows declaram `permissions` por job e só os jobs
   `publicar` e `marcar-producao` pedem `packages: write`. Se a organização/conta
   tiver restrição de pacotes, liberar a criação de pacotes pelo Actions.
3. O primeiro push cria os pacotes `todeolho-api` e `todeolho-web` na conta
   `Alzarus`, ligados ao repositório pela label `org.opencontainers.image.source`
   (o repositório recebe acesso de escrita automaticamente). Pacote novo nasce
   **privado**; o deploy e a varredura agendada fazem login com o
   `GITHUB_TOKEN`, então funciona assim. Como o código já é público, tornar os
   pacotes públicos (página do pacote → *Package settings* → *Change
   visibility*) é opcional e só facilita `docker pull` manual.

### Acompanhar

Na aba Actions, o run *Deploy to Contabo VPS* mostra no resumo a tag
implantada. Na VM:

```bash
ssh <user>@<host>
cd /opt/todeolho
grep IMAGE_TAG .env
docker compose -f docker-compose.contabo.yml ps
docker compose -f docker-compose.contabo.yml logs -f api
```

Resto do build antigo: `/opt/todeolho/backend` e `/opt/todeolho/frontend` não
são mais usados e podem ser apagados à mão. O cache de build do Docker na VM
também deixou de ser útil (o workflow não roda mais `docker builder prune`);
limpar uma vez com `docker builder prune` se o disco apertar, lembrando que o
cache é compartilhado com os outros projetos da VM.

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

O volume `postgres_data` é preservado entre deploys. Cada deploy publica as
imagens com o SHA completo do commit, então voltar a uma versão anterior não
exige rebuild:

1. Descobrir o SHA a restaurar: aba Actions → runs anteriores de *Deploy to
   Contabo VPS* (o resumo mostra a tag), ou `git log --format=%H` em `master`.
2. Actions → *Deploy to Contabo VPS* → **Run workflow**, preencher `tag` com o
   SHA completo (40 caracteres) e rodar.
3. O workflow pula testes, Trivy e publicação, grava o novo `IMAGE_TAG` no
   `.env`, faz `pull` + `up -d --no-build`, roda o health check e move a tag
   `:producao` para esse SHA.

Pelo terminal: `gh workflow run deploy.yml -f tag=<sha>`.

Cuidados:

- O rollback usa o `docker-compose.contabo.yml` e os secrets **atuais**. Se a
  versão nova mudou o compose ou variáveis de ambiente, conferir se a antiga
  continua compatível.
- O `AutoMigrate` não remove colunas nem tabelas, então o banco não volta junto:
  a versão antiga convive com o schema novo e não desfaz dados gravados pela
  nova. Coluna nova `NOT NULL` sem default pode quebrar inserts da versão
  antiga; conferir antes de reverter uma mudança de schema.
- O rollback é temporário: o próximo push em `master` implanta o novo commit.
  Para fixar a versão antiga, reverter o commit em `master`.

Para reverter apenas o roteamento, basta apontar o `location /` do NPM de volta
para `quemvotar-web:8080`.

## 8. Varreduras e monitoramento

| Workflow | Quando | O que faz |
|---|---|---|
| `seguranca-agendada.yml` | Segunda 09:00 UTC | Trivy do repositório e das imagens `:producao` (o que está no ar). Falha em CRITICAL/HIGH com correção. |
| `disponibilidade.yml` | A cada 15 min | `curl` em `/`, `/ranking`, `/api/v1/ranking`, `/quemvotar/` (200) e `/api/v1/sync/daily` (404 na borda), com 3 tentativas. |
| `.github/dependabot.yml` | Semanal | PRs agrupados para Go, Bun (frontend), imagens base e actions. |

Falha de workflow agendado gera e-mail para o dono do repositório. Quando a
varredura agendada acusar CVE na imagem base (ex.: `tzdata`, `openssl`), um
novo deploy já resolve: o build usa `pull: true` e sempre baixa a tag base mais
recente. Sem mudança de código, rodar *Deploy to Contabo VPS* sem `tag`.

Observação: o GitHub desativa workflows agendados de repositórios públicos após
60 dias sem commits; se isso acontecer, reativar em Actions.
