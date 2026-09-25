#!/usr/bin/env bash
#
# Desativa no Nginx Proxy Manager os proxy hosts *.sslip.io.
#
# POR QUE
#   Na checagem de 24/09/2026 o NPM ainda servia tres enderecos sslip.io
#   (<IP>.sslip.io, neuroanalises.<IP>.sslip.io e
#   quemvotar.<IP>.sslip.io). Eles resolvem direto para o IP da VPS,
#   sem Cloudflare, e respondiam 200: o Quem Votar e o neuroanalises inteiros por
#   fora do WAF e do cache, e o proprio nome publica o IP da origem. Sao
#   enderecos da fase anterior aos dominios definitivos (todeolho.org/quemvotar
#   e neurofluxis.com). O firewall-cloudflare.sh se recusa a aplicar enquanto
#   eles existirem, porque cairiam com o filtro.
#
# COMO
#   "desativar" usa o endpoint /disable da API do NPM: o host continua salvo e
#   volta com "reativar". Antes, grava o JSON de cada host em
#   /opt/proxy/backup-sslip-<id>-<data>.json.
#
# USO (NA VPS, como root)
#   bash npm-sslip.sh status
#   bash npm-sslip.sh desativar
#   bash npm-sslip.sh reativar
#
set -euo pipefail

NPM_API="http://127.0.0.1:81/api"
BACKUP_DIR=/opt/proxy
FILTRO='sslip\.io$'

need() { command -v "$1" >/dev/null || { echo "faltando: $1" >&2; exit 1; }; }
need curl; need jq; need docker

# Mesmo helper do scripts/npm-routing.sh: o NPM nao tem login sem senha, mas os
# modulos dele emitem token para um usuario existente (imagem atual: /app/db.js).
token() {
  local t
  docker exec -i proxy-app-1 sh -c 'cat > /app/gen-token-fixed.mjs' <<'HELPER'
import "/app/db.js";
import userModel from "/app/models/user.js";
import tokenModel from "/app/internal/token.js";

async function run() {
  const user = await userModel.query().where("id", 1).andWhere("is_deleted", 0).first();
  if (!user) { console.error("usuario id=1 nao encontrado"); process.exit(1); }
  const res = await tokenModel.getTokenFromUser(user);
  console.log("TOKEN:" + res.token);
  process.exit(0);
}
run().catch(e => { console.error(e && e.message ? e.message : e); process.exit(1); });
HELPER
  t=$(docker exec proxy-app-1 node /app/gen-token-fixed.mjs 2>/dev/null | sed -n 's/^TOKEN://p' | tr -d '\r')
  [ -n "$t" ] || { echo "falha ao obter token da API do NPM" >&2; exit 1; }
  printf '%s' "$t"
}

hosts_sslip() { # JSON: lista dos proxy hosts com algum dominio sslip.io
  curl -fsS -H "Authorization: Bearer $1" "$NPM_API/nginx/proxy-hosts" \
    | jq --arg f "$FILTRO" '[.[] | select(any(.domain_names[]; test($f)))]'
}

cmd_status() {
  local t; t=$(token)
  hosts_sslip "$t" | jq -r '.[] | "id=\(.id) \(if .enabled == 1 or .enabled == true then "ATIVO   " else "desativado" end) \(.domain_names | join(" ")) -> \(.forward_host):\(.forward_port)"'
  hosts_sslip "$t" | jq -e 'length > 0' >/dev/null || echo "(nenhum proxy host sslip.io)"
}

alternar() { # alternar disable|enable
  local t acao="$1" id carimbo; t=$(token); carimbo=$(date -u +%Y%m%dT%H%M%SZ)
  for id in $(hosts_sslip "$t" | jq -r '.[].id'); do
    if [ "$acao" = disable ]; then
      curl -fsS -H "Authorization: Bearer $t" "$NPM_API/nginx/proxy-hosts/$id" > "$BACKUP_DIR/backup-sslip-$id-$carimbo.json"
    fi
    curl -fsS -X POST -H "Authorization: Bearer $t" "$NPM_API/nginx/proxy-hosts/$id/$acao" >/dev/null
    echo "[OK] proxy host $id: $acao"
  done
  cmd_status
}

case "${1:-status}" in
  status)    cmd_status ;;
  desativar) alternar disable; echo "Backups: $BACKUP_DIR/backup-sslip-*.json" ;;
  reativar)  alternar enable ;;
  *) sed -n '2,24p' "$0"; exit 1 ;;
esac
