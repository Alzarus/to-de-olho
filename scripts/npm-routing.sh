#!/usr/bin/env bash
#
# Troca o roteamento de todeolho.org no Nginx Proxy Manager.
#
# Estado atual : todeolho.org/*          -> quemvotar-web:8080
# Estado alvo  : todeolho.org/           -> todeolho-web:3000
#                todeolho.org/quemvotar  -> quemvotar-web:8080
#
# A ordem importa: a fase 1 cria a Custom Location /quemvotar enquanto a raiz
# ainda aponta para o quemvotar (nada muda para o usuario). So a fase 2 move a
# raiz. Se a fase 2 for aplicada antes do container todeolho-web existir, o site
# responde 502 -- por isso a fase 2 checa o container antes de agir.
#
# Uso (executar NA VM, como root):
#   bash npm-routing.sh status     # mostra o roteamento atual, nao altera nada
#   bash npm-routing.sh fase1      # adiciona /quemvotar como Custom Location
#   bash npm-routing.sh fase2      # move a raiz para todeolho-web:3000
#   bash npm-routing.sh rollback   # devolve a raiz para quemvotar-web:8080
#
set -euo pipefail

DOMAIN="todeolho.org"
NPM_API="http://127.0.0.1:81/api"

need() { command -v "$1" >/dev/null || { echo "faltando: $1" >&2; exit 1; }; }
need curl; need jq; need docker

# O NPM nao expoe um endpoint de login sem senha, mas seus proprios modulos
# sabem emitir um token para um usuario existente. Injetamos um helper minimo
# no container e o executamos ali dentro. O /opt/proxy/gen-token.mjs que ja
# existia aponta para /app/lib/db.js, caminho de uma versao antiga do NPM --
# nesta imagem o modulo esta em /app/db.js.
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
  t=$(docker exec proxy-app-1 node /app/gen-token-fixed.mjs 2>/dev/null \
      | sed -n 's/^TOKEN://p' | tr -d '\r')
  [ -n "$t" ] || { echo "falha ao obter token da API do NPM" >&2; exit 1; }
  printf '%s' "$t"
}

host_json() {
  curl -fsS -H "Authorization: Bearer $1" "$NPM_API/nginx/proxy-hosts" \
    | jq --arg d "$DOMAIN" '.[] | select(.domain_names | index($d))'
}

cmd_status() {
  local t; t=$(token)
  host_json "$t" | jq '{
    id, domain_names,
    raiz: ("\(.forward_scheme)://\(.forward_host):\(.forward_port)"),
    custom_locations: [.locations[]? | {path, destino: "\(.forward_scheme)://\(.forward_host):\(.forward_port)"}],
    ssl_forcado: .ssl_forced, hsts: .hsts_enabled, certificado: .certificate_id
  }'
}

# Reenvia o objeto inteiro: a API do NPM espera PUT completo, nao patch parcial.
apply() {
  local t="$1" id="$2" body="$3"
  curl -fsS -X PUT -H "Authorization: Bearer $t" -H "Content-Type: application/json" \
       -d "$body" "$NPM_API/nginx/proxy-hosts/$id" > /dev/null
  echo "[ok] proxy host $id atualizado"
}

payload() {
  jq '{domain_names, forward_scheme, forward_host, forward_port, access_list_id,
       certificate_id, ssl_forced, hsts_enabled, hsts_subdomains, http2_support,
       block_exploits, caching_enabled, allow_websocket_upgrade, advanced_config,
       meta, locations}'
}

cmd_fase1() {
  local t id cur new; t=$(token); cur=$(host_json "$t"); id=$(jq -r .id <<<"$cur")

  if jq -e '.locations[]? | select(.path=="/quemvotar")' <<<"$cur" >/dev/null; then
    echo "[skip] /quemvotar ja existe"; return 0
  fi

  new=$(jq '.locations += [{
        path: "/quemvotar",
        forward_scheme: "http",
        forward_host: "quemvotar-web",
        forward_port: 8080,
        advanced_config: ""
      }]' <<<"$cur" | payload)

  apply "$t" "$id" "$new"
  echo "[!] valide https://$DOMAIN/quemvotar/ ANTES de rodar a fase2"
}

cmd_fase2() {
  local t id cur new; t=$(token); cur=$(host_json "$t"); id=$(jq -r .id <<<"$cur")

  docker ps --format '{{.Names}}' | grep -qx todeolho-web \
    || { echo "[abortado] container todeolho-web nao esta rodando" >&2; exit 1; }

  jq -e '.locations[]? | select(.path=="/quemvotar")' <<<"$cur" >/dev/null \
    || { echo "[abortado] rode a fase1 primeiro, ou o quemvotar sai do ar" >&2; exit 1; }

  new=$(jq '.forward_host="todeolho-web" | .forward_port=3000 | .forward_scheme="http"' \
        <<<"$cur" | payload)
  apply "$t" "$id" "$new"
}

cmd_rollback() {
  local t id cur new; t=$(token); cur=$(host_json "$t"); id=$(jq -r .id <<<"$cur")
  new=$(jq '.forward_host="quemvotar-web" | .forward_port=8080 | .forward_scheme="http"' \
        <<<"$cur" | payload)
  apply "$t" "$id" "$new"
  echo "[ok] raiz devolvida ao quemvotar-web:8080"
}

case "${1:-status}" in
  status)   cmd_status ;;
  fase1)    cmd_fase1 ;;
  fase2)    cmd_fase2 ;;
  rollback) cmd_rollback ;;
  *) echo "uso: $0 {status|fase1|fase2|rollback}" >&2; exit 1 ;;
esac
