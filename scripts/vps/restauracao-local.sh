#!/usr/bin/env bash
#
# Lado LOCAL do backup: gera a chave age e testa a restauracao de um backup.
# Roda na sua maquina (Git Bash no Windows serve); so precisa de Docker.
# Backup que nunca foi restaurado e so uma esperanca: rode o "testar" ao
# instalar e depois uma vez por mes.
#
# USO
#   bash restauracao-local.sh chave
#       Gera o par age em ~/.config/todeolho/backup-age.key (fora do OneDrive)
#       e imprime a chave PUBLICA para o /etc/todeolho-backup.env da VPS.
#       Guarde tambem a chave privada no gerenciador de senhas: sem ela,
#       nenhum backup pode ser aberto.
#
#   bash restauracao-local.sh testar <arquivo.tar.age>
#       Arquivo baixado do painel do R2 (bucket > diario/AAAA-MM-DD/).
#   bash restauracao-local.sh testar r2
#       Baixa o mais recente; pede as credenciais R2 (so leitura basta).
#
#   Descriptografa, restaura o Postgres num conteiner descartavel, confere o
#   SQLite e o pacote do NPM, compara com o MANIFESTO e apaga tudo no fim.
#
set -euo pipefail
export MSYS_NO_PATHCONV=1

CHAVE="${AGE_CHAVE:-$HOME/.config/todeolho/backup-age.key}"
TRAB=$(mktemp -d)
PG=tdo-restauracao-teste
trap 'docker rm -f "$PG" >/dev/null 2>&1 || true; rm -rf "$TRAB"' EXIT

# Caminho que o Docker Desktop entende (C:/... no Git Bash).
caminho_docker() { (cd "$1" && pwd -W 2>/dev/null || pwd); }

ferramentas() { # roda age/rclone/sqlite numa alpine descartavel
  docker run --rm -i -v "$(caminho_docker "$TRAB"):/t" -v "$(caminho_docker "$(dirname "$CHAVE")"):/k:ro" \
    -e RCLONE_CONFIG_R2_TYPE=s3 -e RCLONE_CONFIG_R2_PROVIDER=Cloudflare \
    -e RCLONE_CONFIG_R2_ACCESS_KEY_ID -e RCLONE_CONFIG_R2_SECRET_ACCESS_KEY -e RCLONE_CONFIG_R2_ENDPOINT \
    alpine:3.22 sh -c "apk add -q age rclone sqlite >/dev/null && $1"
}

cmd_chave() {
  [ -s "$CHAVE" ] && { echo "[OK] chave ja existe em $CHAVE"; grep -o 'age1[0-9a-z]*' "$CHAVE"; return; }
  mkdir -p "$(dirname "$CHAVE")"
  # Gera num temporario: se o docker falhar, nao sobra um arquivo vazio que
  # a proxima execucao tomaria por uma chave existente.
  docker run --rm alpine:3.22 sh -c 'apk add -q age >/dev/null && age-keygen 2>/dev/null' > "$CHAVE.tmp" \
    && grep -q '^AGE-SECRET-KEY-' "$CHAVE.tmp" || { rm -f "$CHAVE.tmp"; echo "[ERRO] age-keygen falhou (o Docker esta rodando?)" >&2; exit 1; }
  mv "$CHAVE.tmp" "$CHAVE"
  chmod 600 "$CHAVE"
  echo "[OK] chave privada em $CHAVE (copie para o gerenciador de senhas)"
  echo "Chave PUBLICA (vai em AGE_DESTINATARIO na VPS):"
  grep -o 'age1[0-9a-z]*' "$CHAVE"
}

obter_arquivo() {
  if [ "$1" != r2 ]; then cp "$1" "$TRAB/backup.tar.age"; return; fi
  # Usa o ambiente se ja vier preenchido (set -a; . ~/.config/todeolho/backup.env).
  [ -n "${R2_BUCKET:-}" ] || read -rp "Bucket R2: " R2_BUCKET
  [ -n "${RCLONE_CONFIG_R2_ENDPOINT:-}" ] || read -rp "Endpoint (https://<conta>.r2.cloudflarestorage.com): " RCLONE_CONFIG_R2_ENDPOINT
  [ -n "${RCLONE_CONFIG_R2_ACCESS_KEY_ID:-}" ] || read -rp "Access key ID: " RCLONE_CONFIG_R2_ACCESS_KEY_ID
  [ -n "${RCLONE_CONFIG_R2_SECRET_ACCESS_KEY:-}" ] || { read -rsp "Secret access key: " RCLONE_CONFIG_R2_SECRET_ACCESS_KEY; echo; }
  export RCLONE_CONFIG_R2_ENDPOINT RCLONE_CONFIG_R2_ACCESS_KEY_ID RCLONE_CONFIG_R2_SECRET_ACCESS_KEY
  ferramentas "f=\$(rclone lsf -R --files-only --s3-no-check-bucket r2:${R2_BUCKET}/diario | sort | tail -1) \
    && echo \"mais recente: \$f\" && rclone copyto --s3-no-check-bucket r2:${R2_BUCKET}/diario/\$f /t/backup.tar.age"
}

conferir_manifesto() {
  ferramentas "cd /t/x && sed -n '/--- sha256/,/---/p' MANIFESTO.txt | grep -v '^---' | sha256sum -c -"
}

restaurar_postgres() {
  docker run -d --name "$PG" -e POSTGRES_PASSWORD=teste postgres:15-alpine >/dev/null
  for _ in $(seq 1 30); do docker exec "$PG" pg_isready -U postgres >/dev/null 2>&1 && break; sleep 1; done
  docker exec "$PG" createdb -U postgres todeolho
  docker cp "$(caminho_docker "$TRAB/x")/todeolho-postgres.dump" "$PG:/tmp/d.dump" >/dev/null
  # --no-owner: o usuario de producao nao existe no conteiner de teste.
  docker exec "$PG" pg_restore -U postgres -d todeolho --no-owner --exit-on-error /tmp/d.dump
  echo "--- Postgres restaurado: maiores tabelas"
  docker exec "$PG" psql -U postgres -d todeolho -Atc \
    "select relname||': '||n_live_tup from pg_stat_user_tables order by n_live_tup desc limit 12" 2>/dev/null \
    || true
  docker exec "$PG" psql -U postgres -d todeolho -Atc "analyze" >/dev/null
  docker exec "$PG" psql -U postgres -d todeolho -Atc \
    "select 'senadores: '||count(*) from senadores" || echo "[ERRO] tabela senadores ausente"
}

cmd_testar() {
  [ -f "$CHAVE" ] || { echo "[ERRO] chave privada nao encontrada em $CHAVE" >&2; exit 1; }
  obter_arquivo "${1:?informe o arquivo .tar.age ou 'r2'}"
  ferramentas "mkdir -p /t/x && age -d -i /k/$(basename "$CHAVE") /t/backup.tar.age | tar -xf - -C /t/x && ls -la /t/x"
  conferir_manifesto && echo "[OK] sha256 conferem com o MANIFESTO"
  restaurar_postgres
  # Backups a partir de 26/09 trazem o SQLite em gzip; os anteriores, cru.
  [ -f "$TRAB/x/quemvotar.db.gz" ] && ferramentas "gunzip /t/x/quemvotar.db.gz"
  if [ -f "$TRAB/x/quemvotar.db" ]; then
    ferramentas "sqlite3 /t/x/quemvotar.db 'pragma integrity_check;' && sqlite3 /t/x/quemvotar.db \"select 'tabelas: '||count(*) from sqlite_master where type='table'\""
  fi
  [ -f "$TRAB/x/npm.tar.gz" ] && echo "npm: $(tar -tzf "$TRAB/x/npm.tar.gz" | wc -l) arquivos"
  echo "[OK] restauracao testada; conteineres e arquivos temporarios serao apagados"
}

case "${1:-}" in
  chave)  cmd_chave ;;
  testar) cmd_testar "${2:-}" ;;
  *) sed -n '2,22p' "$0"; exit 1 ;;
esac
