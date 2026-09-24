#!/usr/bin/env bash
#
# Backup diario da VPS para o Cloudflare R2, criptografado com age.
#
# O QUE ENTRA
#   - Postgres do To De Olho: pg_dump -Fc (formato custom; restaura por tabela).
#   - SQLite do Quem Votar: copia online pela API de backup do better-sqlite3
#     (o banco esta em WAL; copiar o arquivo com a API rodando pode dar copia
#     inconsistente).
#   - Nginx Proxy Manager: dados (proxy hosts, custom locations, o bloqueio de
#     /api/v1/sync) e certificados Let's Encrypt.
#   - MANIFESTO com sha256 de cada parte e as imagens em execucao.
#   Os .env NAO entram: os segredos vivem nos secrets do GitHub e o deploy
#   regrava o .env a cada execucao.
#
# CRIPTOGRAFIA
#   A VPS guarda so a chave PUBLICA do age (AGE_DESTINATARIO). A privada fica
#   fora da VPS (na sua maquina e no gerenciador de senhas). Quem invadir a VPS
#   consegue gerar backups novos, mas nao le os antigos.
#
# DESTINO
#   r2:<bucket>/diario/AAAA-MM-DD/...  (todo dia)
#   r2:<bucket>/mensal/AAAA-MM/...     (dia 1 de cada mes)
#   A retencao e feita por regras de ciclo de vida no proprio R2 (ver README),
#   nao por este script: o token da VPS nao precisa apagar nada.
#
# CONFIGURACAO: /etc/todeolho-backup.env (modo 600), criado pelo instalar-backup.sh.
# EXECUCAO: timer systemd todeolho-backup.timer (diario, depois do sync das 06:00 UTC).
#
set -euo pipefail
umask 077

CONFIG=/etc/todeolho-backup.env
# shellcheck source=/dev/null
. "$CONFIG"
: "${AGE_DESTINATARIO:?AGE_DESTINATARIO ausente em $CONFIG}"
: "${R2_BUCKET:?R2_BUCKET ausente em $CONFIG}"
: "${RCLONE_CONFIG_R2_ENDPOINT:?RCLONE_CONFIG_R2_ENDPOINT ausente em $CONFIG}"
export RCLONE_CONFIG_R2_TYPE=s3 RCLONE_CONFIG_R2_PROVIDER=Cloudflare RCLONE_CONFIG_R2_NO_CHECK_BUCKET=true
export RCLONE_CONFIG_R2_ACCESS_KEY_ID RCLONE_CONFIG_R2_SECRET_ACCESS_KEY RCLONE_CONFIG_R2_ENDPOINT

LOCAL_DIR="${LOCAL_DIR:-/var/backups/todeolho}"
MANTER_LOCAL="${MANTER_LOCAL:-3}"
TDO_ENV="${TDO_ENV:-/opt/todeolho/.env}"
NPM_DIR="${NPM_DIR:-/opt/proxy}"
# Tamanho minimo do dump do Postgres. O banco tem centenas de MB; um dump
# minusculo significa banco vazio ou pg_dump que falhou sem erro.
MIN_PG_BYTES="${MIN_PG_BYTES:-1000000}"

DATA=$(date -u +%F)
CARIMBO=$(date -u +%Y%m%dT%H%M%SZ)
TMP=$(mktemp -d /var/tmp/todeolho-backup.XXXXXX)
trap 'rm -rf "$TMP"' EXIT

log() { echo "[backup] $*"; }
falha() { echo "[backup][ERRO] $*" >&2; avisar_healthcheck fail "$*"; exit 1; }

# Ping opcional para healthchecks.io (ou similar): sem ping em 25 h, voce
# recebe um e-mail. E o que transforma "backup falhou em silencio" em alerta.
avisar_healthcheck() {
  [ -n "${HC_URL:-}" ] || return 0
  local sufixo=""; [ "${1:-}" = fail ] && sufixo="/fail"; [ "${1:-}" = start ] && sufixo="/start"
  curl -fsS -m 10 --retry 3 --data-raw "${2:-}" "${HC_URL}${sufixo}" >/dev/null || true
}

valor_env() { grep -E "^$1=" "$TDO_ENV" | head -1 | cut -d= -f2-; }

backup_postgres() {
  local u d arq="$TMP/todeolho-postgres.dump"
  u=$(valor_env POSTGRES_USER); d=$(valor_env POSTGRES_DB)
  [ -n "$u" ] && [ -n "$d" ] || falha "POSTGRES_USER/POSTGRES_DB nao encontrados em $TDO_ENV"
  docker exec todeolho-db pg_dump -U "$u" -d "$d" -Fc -Z 6 > "$arq" || falha "pg_dump falhou"
  [ "$(stat -c %s "$arq")" -ge "$MIN_PG_BYTES" ] || falha "dump do Postgres pequeno demais ($(stat -c %s "$arq") bytes)"
  # Le o sumario do dump: pega arquivo truncado ou corrompido agora, nao no dia
  # em que for preciso restaurar.
  docker exec -i todeolho-db pg_restore --list < "$arq" > /dev/null || falha "dump do Postgres ilegivel"
  log "postgres: $(du -h "$arq" | cut -f1)"
}

backup_sqlite_quemvotar() {
  local tmp_ct="/data/.backup-${CARIMBO}.db"
  [ "$(docker ps -q --filter name=^quemvotar-api$)" ] || { log "quemvotar-api fora do ar: SQLite pulado"; return 0; }
  # A imagem e distroless (sem shell nem sqlite3): usa o node da propria imagem.
  docker exec -w /app quemvotar-api /nodejs/bin/node -e "
    const D=require('better-sqlite3');
    const db=new D(process.env.DATABASE_PATH,{readonly:true,fileMustExist:true});
    db.backup('${tmp_ct}').then(()=>process.exit(0)).catch(e=>{console.error(e.message);process.exit(1)});
  " || falha "backup do SQLite do Quem Votar falhou"
  docker cp "quemvotar-api:${tmp_ct}" "$TMP/quemvotar.db" >/dev/null
  docker exec -w /app quemvotar-api /nodejs/bin/node -e "require('fs').unlinkSync('${tmp_ct}')" || true
  log "quemvotar sqlite: $(du -h "$TMP/quemvotar.db" | cut -f1)"
}

backup_npm() {
  local partes=()
  for p in data letsencrypt; do [ -d "$NPM_DIR/$p" ] && partes+=("$p"); done
  [ ${#partes[@]} -gt 0 ] || { log "NPM: nada em $NPM_DIR/{data,letsencrypt}; pulado"; return 0; }
  # Logs do NPM ficam de fora: grandes e sem valor para restaurar.
  tar -C "$NPM_DIR" --exclude='data/logs' -czf "$TMP/npm.tar.gz" "${partes[@]}"
  log "npm: $(du -h "$TMP/npm.tar.gz" | cut -f1)"
}

escrever_manifesto() {
  {
    echo "backup ${CARIMBO} host=$(hostname)"
    echo "--- sha256"; (cd "$TMP" && sha256sum -- *.dump *.db *.tar.gz 2>/dev/null)
    echo "--- imagens em execucao"; docker ps --format '{{.Names}} {{.Image}}' | sort
    echo "--- imagem->digest"; docker inspect --format '{{.Name}} {{.Image}}' $(docker ps -q) 2>/dev/null | sort
  } > "$TMP/MANIFESTO.txt"
}

empacotar_e_enviar() {
  local nome="todeolho-backup-${CARIMBO}.tar.age" final
  mkdir -p "$LOCAL_DIR"; final="$LOCAL_DIR/$nome"
  tar -C "$TMP" -cf - MANIFESTO.txt $(cd "$TMP" && ls *.dump *.db *.tar.gz 2>/dev/null) \
    | age -r "$AGE_DESTINATARIO" -o "$final"
  rclone copyto --s3-no-check-bucket --retries 5 "$final" "r2:${R2_BUCKET}/diario/${DATA}/${nome}" || falha "envio ao R2 falhou"
  if [ "$(date -u +%d)" = "01" ]; then
    rclone copyto --s3-no-check-bucket --retries 5 "$final" "r2:${R2_BUCKET}/mensal/$(date -u +%Y-%m)/${nome}" || falha "envio mensal ao R2 falhou"
  fi
  # Confere que o objeto remoto tem o mesmo tamanho do local.
  local remoto; remoto=$(rclone size --json "r2:${R2_BUCKET}/diario/${DATA}/${nome}" | grep -o '"bytes":[0-9]*' | cut -d: -f2)
  [ "$remoto" = "$(stat -c %s "$final")" ] || falha "tamanho no R2 ($remoto) difere do local"
  ls -1t "$LOCAL_DIR"/todeolho-backup-*.tar.age | tail -n +"$((MANTER_LOCAL + 1))" | xargs -r rm -f
  log "enviado: diario/${DATA}/${nome} ($(du -h "$final" | cut -f1))"
}

exec 9>/run/todeolho-backup.lock
flock -n 9 || { log "outro backup em andamento; saindo"; exit 0; }
avisar_healthcheck start
backup_postgres
backup_sqlite_quemvotar
backup_npm
escrever_manifesto
empacotar_e_enviar
avisar_healthcheck ok "ok ${CARIMBO}"
log "concluido"
