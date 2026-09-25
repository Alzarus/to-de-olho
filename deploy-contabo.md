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
       ├─ imagens api e web: build no runner (Buildx + cache do Actions)
       │    → Trivy na imagem carregada (CRITICAL/HIGH com correção)
       │    → artefato da MESMA imagem escaneada
       └─ fumaça: db + api + web com o docker-compose.contabo.yml, 200 em
            /, /ranking, /senador/1 e /api/v1/ranking, e o gráfico do senador
            no navegador (Playwright: svg do Recharts, tooltip, console limpo)
  └─ publicar: push em ghcr.io/alzarus/todeolho-{api,web}:<SHA>
  └─ deploy (VPS): docker login (token do job, via stdin)
                   → pull api web <SHA> e a reserva :producao → logout
                   → ensaio da migração numa cópia do banco (seção 7.1)
                   → up -d --no-build → health check (API, loopback e
                     caminho público pela 443 do NPM)
                   → falhou? rollback automático para :producao (seção 7.2)
  └─ marcar-producao: tag móvel :producao = última versão saudável
```

No pull request roda só a `verificacao` (build + Trivy, **sem** push). Se
qualquer etapa falhar, nada é publicado nem implantado.

O `IMAGE_TAG` (SHA do commit) é gravado no `/opt/todeolho/.env` pelo
`scripts/deploy/implantar.sh`, só depois do ensaio, e o
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
3. O workflow pula testes, Trivy e publicação e segue o mesmo caminho de um
   deploy normal: `pull`, ensaio, `up -d --no-build`, health check (com
   rollback automático) e, se ficar saudável, move a tag `:producao` para esse
   SHA.

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

### 7.1 Ensaio da migração (automático, antes da troca)

Em 25/09/2026 o GORM 1.31.2 (Dependabot #50) passou no CI, que testa com banco
vazio, e na produção o `AutoMigrate` tentou `DROP CONSTRAINT
uni_despesas_ceaps_id_origem`, que não existe. A API entrou em loop e o site
ficou 20 min com 500.

Desde então, antes de trocar a versão, o `scripts/deploy/ensaio-migracao.sh`:

1. copia o banco de produção para `todeolho_ensaio`, no mesmo Postgres
   (`pg_dump | psql`: snapshot MVCC, sem bloquear a API; `CREATE DATABASE ...
   TEMPLATE` exigiria derrubar as conexões dela). Exige livre no disco o dobro
   do tamanho do banco;
2. sobe a imagem **nova** da API em `todeolho-ensaio`, na rede interna, sem
   porta publicada, com `SCHEDULER_DESATIVADO=true` e sem `SYNC_SECRET` nem
   chave do Portal (nenhum sync, nenhuma chamada externa);
3. exige `/health` 200 em até 6 min; se a API encerrar antes (migração
   quebrada), falha na hora, com o log;
4. apaga contêiner e cópia sempre (trap em EXIT, HUP e TERM; um ensaio
   interrompido de outro jeito é limpo pelo seguinte).

Se o ensaio falhar, o job para ali: **a produção não é tocada** e o `.env`
continua apontando a versão no ar. Contra o schema de produção, o ensaio barra
a imagem com GORM 1.31.2 (`constraint "uni_despesas_ceaps_id_origem" ... does
not exist`) e aprova a atual.

O que ele não pega: migração que passa mas deixa o dado errado, e erro que só
aparece com tráfego. Para isso existem o health check e o rollback.

### 7.2 Rollback automático

O `scripts/deploy/implantar.sh` faz o `up` e o health check:

- API `/health` pela rede interna (até 6 min; se o contêiner reiniciar 3
  vezes, desiste na hora, porque loop de reinício não se recupera sozinho);
- web no loopback `127.0.0.1:5350`;
- `/` e `/api/v1/ranking` **pelo caminho público**: porta 443 do NPM com o
  SNI e o Host `todeolho.org` (`curl --resolve todeolho.org:443:127.0.0.1`),
  como o Quem Votar faz desde o incidente do 502 (quem-votar #20).

Se falhar, ele reimplanta as imagens da tag `:producao` (a última versão que
passou no health check), baixadas no passo de pull **antes** da troca, para
não depender da rede da VPS na hora do rollback. Confere a saúde delas e
termina o job em **falha**, com o motivo e a versão restaurada no resumo do
run. Casos-limite:

| situação | o que acontece |
|---|---|
| não existe `:producao` (primeiro deploy) | falha sem rollback; o resumo indica o rollback manual |
| `:producao` é a mesma imagem que falhou | falha sem rollback (o problema está na VPS: banco, rede, NPM) |
| a `:producao` também falha | uma única tentativa: falha com "ROLLBACK FALHOU", sem loop; intervenção manual |
| pull da `:producao` falha por rede | o deploy para antes da troca: sem reserva, a versão no ar não é trocada |

O rollback troca imagens; o schema fica como a versão nova deixou (ver os
cuidados acima). O `marcar-producao` só roda com o deploy saudável, então a
`:producao` nunca aponta uma versão que falhou.

### 7.3 Merges e Dependabot

Combinado desde 25/09/2026, depois de três quedas no mesmo dia:

- **Um merge por vez, esperando o deploy terminar** (verde, ou com o rollback
  concluído) antes do próximo. Cada deploy publica o `master` inteiro, e não
  só o PR: em 25/09 os merges em lote levaram juntos o Node 25 (#51) e o GORM
  quebrado (#50), e um PR independente (#59) quase republicou o GORM porque o
  `master` ainda o continha.
- **Versão maior exige revisão.** O Dependabot nunca agrupa major; o PR dela
  recebe o rótulo `revisar` e um roteiro (`dependabot-revisao.yml`). Exemplos:
  lucide-react 1.x removeu os ícones de marcas (#61); recharts 3.10 mudou o
  tipo do tooltip (#53).
- Ignorados de propósito no `dependabot.yml`: major do `node` nas imagens (a
  troca é manual e só para LTS), major do `@types/node` e `gorm.io/*` (até a
  atualização testada com o ensaio).
- O `master` é protegido: merge só com os checks do CI verdes
  (`verificacao / ...`, inclusive a fumaça).

## 8. Varreduras e monitoramento

| Workflow | Quando | O que faz |
|---|---|---|
| `seguranca-agendada.yml` | Segunda 09:00 UTC | Trivy do repositório e das imagens `:producao` (o que está no ar). Falha em CRITICAL/HIGH com correção. |
| `.github/dependabot.yml` | Semanal | PRs agrupados (minor e patch) para Go, Bun (frontend), imagens base e actions; major em PR próprio, com o rótulo `revisar`. |

Falha de workflow agendado gera e-mail para o dono do repositório. Quando a
varredura agendada acusar CVE na imagem base (ex.: `tzdata`, `openssl`), um
novo deploy já resolve: o build usa `pull: true` e sempre baixa a tag base mais
recente. Sem mudança de código, rodar *Deploy to Contabo VPS* sem `tag`.

Observação: o GitHub desativa workflows agendados de repositórios públicos após
60 dias sem commits; se isso acontecer, reativar em Actions.

### Disponibilidade: monitor externo

O antigo `disponibilidade.yml` (cron de 15 min no Actions) foi removido em
25/09/2026. Ele não servia: o cron rodou 2 vezes em quase 10 h, e nas duas a
Cloudflare devolveu **403** ao runner do GitHub em todas as rotas. Nunca
avisaria de uma queda de verdade e só gerava alarme falso.

O monitor é um serviço externo, fora da VPS e do GitHub:

- **UptimeRobot** (plano grátis, intervalo de 5 min, alerta por e-mail e/ou
  Telegram), com um monitor HTTP(S) por rota: `https://todeolho.org/`,
  `https://todeolho.org/api/v1/ranking` e `https://todeolho.org/quemvotar/`,
  esperando 200. Opcional: monitor de *keyword* (`Tô De Olho`) na home, para
  pegar página de erro servida com 200.
- **healthchecks.io** (grátis) para o backup: o `backup.sh` faz ping em
  `HC_URL` ao terminar; sem ping no prazo, o serviço avisa.
