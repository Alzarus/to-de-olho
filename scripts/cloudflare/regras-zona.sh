#!/usr/bin/env bash
#
# Regras da zona todeolho.org na Cloudflare, num unico lugar.
#
# POR QUE UM SCRIPT SO
#   A fase de cache (http_request_cache_settings) e a de limite de taxa
#   (http_ratelimit) guardam TODAS as regras da zona num unico ruleset. Um PUT
#   substitui a lista inteira. O Quem Votar tinha o seu proprio script
#   (quem-votar/deploy/cloudflare-cache-rule.sh) que fazia exatamente isso: se
#   o To De Olho criasse outro, um apagaria as regras do outro. Este script e a
#   fonte unica das regras dos dois projetos, e regras que ele nao conhece
#   (criadas a mao no painel) sao PRESERVADAS no fim da lista, nunca apagadas.
#
# O QUE ELE APLICA
#   1. Configuracoes TLS da zona: SSL "Full (strict)" (a origem tem certificado
#      Let's Encrypt valido para todeolho.org e www; conferido em 24/09/2026),
#      Always Use HTTPS, TLS minimo 1.2, TLS 1.3, Automatic HTTPS Rewrites.
#      Antes de ligar o strict, confere o certificado da origem de cada registro
#      DNS proxied da zona; se algum falhar, o strict nao e aplicado.
#   2. Cache de borda (detalhes nos comentarios de cada regra abaixo).
#   3. Uma regra de limite de taxa por IP em /api/ (o plano gratuito permite uma).
#
# TOKEN (escopo minimo; My Profile > API Tokens > Create Custom Token)
#   Zone > Zone: Read          Zone > DNS: Read
#   Zone > Zone Settings: Edit Zone > Cache Rules: Edit
#   Zone > Zone WAF: Edit      Zone > Cache Purge: Purge
#   Zone Resources: Include > Specific zone > todeolho.org
#   O token e lido do ambiente e nunca impresso nem gravado.
#
# USO (na sua maquina, Git Bash; precisa de curl, jq e openssl)
#   read -rs CF_API_TOKEN && export CF_API_TOKEN     # cola o token, Enter
#   bash scripts/cloudflare/regras-zona.sh status    # so le e compara
#   bash scripts/cloudflare/regras-zona.sh aplicar   # aplica tudo
#   bash scripts/cloudflare/regras-zona.sh purgar    # esvazia o cache da zona
#   unset CF_API_TOKEN
#
set -euo pipefail

ZONA="${CF_ZONA:-todeolho.org}"
HOST="${CF_HOSTNAME:-todeolho.org}"
API="https://api.cloudflare.com/client/v4"

need() { command -v "$1" >/dev/null || { echo "[ERRO] faltando: $1" >&2; exit 1; }; }
need curl; need jq; need openssl
: "${CF_API_TOKEN:?defina CF_API_TOKEN (read -rs CF_API_TOKEN && export CF_API_TOKEN)}"

# Chamada a API. Em falha, mostra so as mensagens de erro (a resposta completa
# pode trazer identificadores da conta) e encerra.
cf() {
  local metodo="$1" caminho="$2" corpo="${3:-}" resposta
  local args=(-sS -X "$metodo" -H "Authorization: Bearer ${CF_API_TOKEN}" -H "Content-Type: application/json")
  [ -n "$corpo" ] && args+=(--data "$corpo")
  resposta=$(curl "${args[@]}" "${API}${caminho}")
  if [ "$(jq -r '.success' <<<"$resposta")" != "true" ]; then
    echo "[ERRO] $metodo $caminho:" >&2
    jq -r '.errors[]? | "  \(.code): \(.message)"' <<<"$resposta" >&2 || true
    return 1
  fi
  printf '%s' "$resposta"
}

zona_id() {
  cf GET "/zones?name=${ZONA}" | jq -r '.result[0].id // empty'
}

# ---------------------------------------------------------------------------
# Regras de cache gerenciadas. A "description" e a chave: e por ela que o
# script reconhece as regras que sao suas na hora de mesclar com as do painel.
# ---------------------------------------------------------------------------
regras_cache() {
  jq -n --arg h "$HOST" '[
    # --- Quem Votar (copiadas de quem-votar/deploy/cloudflare-cache-rule.sh,
    # --- mesma descricao e mesmos parametros; a justificativa esta la).
    {
      description: "Quem Votar - midias de urna imutaveis",
      expression: "(http.host eq \"\($h)\" and starts_with(http.request.uri.path, \"/quemvotar/api/arquivo/\"))",
      action: "set_cache_settings",
      action_parameters: {
        cache: true,
        edge_ttl: { mode: "override_origin", default: 2592000 },
        browser_ttl: { mode: "override_origin", default: 31536000 },
        cache_key: { ignore_query_strings_order: true }
      }
    },
    {
      description: "Quem Votar - rotas de dados eleitorais",
      expression: "(http.host eq \"\($h)\" and starts_with(http.request.uri.path, \"/quemvotar/api/\") and not starts_with(http.request.uri.path, \"/quemvotar/api/arquivo/\"))",
      action: "set_cache_settings",
      action_parameters: {
        cache: true,
        edge_ttl: { mode: "bypass_by_default" },
        browser_ttl: { mode: "respect_origin" },
        cache_key: { ignore_query_strings_order: true }
      }
    },
    # --- To De Olho: API publica. Obedece o Cache-Control da API (s-maxage);
    # --- sem cabecalho, nao cacheia (mesma escolha do Quem Votar: na duvida,
    # --- perder desempenho, nunca servir dado velho). Ficam de fora o sync
    # --- (ja bloqueado no NPM), o contador de acessos e /stats.
    {
      description: "To De Olho - API publica",
      expression: "(http.host eq \"\($h)\" and starts_with(http.request.uri.path, \"/api/v1/\") and not starts_with(http.request.uri.path, \"/api/v1/sync\") and not starts_with(http.request.uri.path, \"/api/v1/acessos\") and not starts_with(http.request.uri.path, \"/api/v1/stats\"))",
      action: "set_cache_settings",
      action_parameters: {
        cache: true,
        edge_ttl: { mode: "bypass_by_default" },
        browser_ttl: { mode: "respect_origin" },
        cache_key: { ignore_query_strings_order: true }
      }
    },
    # --- To De Olho: imagens otimizadas pelo Next (/_next/image, fotos dos
    # --- senadores). O caminho nao tem extensao, entao a Cloudflare o trata
    # --- como DYNAMIC. O Next ja emite Cache-Control (minimumCacheTTL 24 h).
    {
      description: "To De Olho - imagens otimizadas",
      expression: "(http.host eq \"\($h)\" and http.request.uri.path eq \"/_next/image\")",
      action: "set_cache_settings",
      action_parameters: {
        cache: true,
        edge_ttl: { mode: "respect_origin" },
        browser_ttl: { mode: "respect_origin" },
        cache_key: { ignore_query_strings_order: true }
      }
    },
    # --- To De Olho: paginas HTML. E a regra que protege a VPS num pico: sem
    # --- ela, cada visita renderiza na origem. TTL de borda SOBREPOSTO em 2 min
    # --- porque o Next emite s-maxage=31536000 nas paginas estaticas; obedecer
    # --- isso prenderia o HTML por um ano. 2 min tambem limita a janela em que,
    # --- apos um deploy, um HTML antigo aponta para chunks /_next/static que ja
    # --- nao existem (o deploy pode chamar "purgar" para zerar essa janela).
    # --- Navegacao do App Router (query _rsc ou cabecalho RSC) nao e cacheada: a
    # --- mesma URL devolve HTML ou payload RSC conforme cabecalhos que a
    # --- Cloudflare nao poe na chave de cache (ela ignora Vary).
    # --- As fichas (/senador/[id]) saem do Next com "private, no-store" so por
    # --- serem dinamicas; nao ha cookie nem conteudo por usuario (conferido no
    # --- PR #48), entao o TTL sobreposto as cacheia de proposito: sao as paginas
    # --- mais caras de renderizar.
    # --- Erros 4xx/5xx nunca ficam na borda.
    {
      description: "To De Olho - paginas HTML",
      expression: "(http.host eq \"\($h)\" and http.request.method eq \"GET\" and not starts_with(http.request.uri.path, \"/api/\") and not starts_with(http.request.uri.path, \"/quemvotar\") and not starts_with(http.request.uri.path, \"/_next/\") and not http.request.uri.query contains \"_rsc=\" and not any(lower(http.request.headers.names[*])[*] == \"rsc\"))",
      action: "set_cache_settings",
      action_parameters: {
        cache: true,
        edge_ttl: {
          mode: "override_origin",
          default: 120,
          status_code_ttl: [ { status_code_range: { from: 400, to: 599 }, value: -1 } ]
        },
        browser_ttl: { mode: "respect_origin" },
        cache_key: { ignore_query_strings_order: true }
      }
    }
  ]'
}

# Limite de taxa (plano gratuito: 1 regra, janela de 10 s, bloqueio de 10 s,
# contagem por IP + data center). Valor GENEROSO de proposito: no Brasil boa
# parte dos celulares sai por CGNAT, com muitos usuarios atras do mesmo IP.
# 100 requisicoes em 10 s por IP ainda corta raspagem e inundacao, mas nao pune
# uma operadora inteira. A API Go tem o seu proprio limite por IP como segunda
# camada.
regras_limite() {
  jq -n '[
    {
      description: "To De Olho - limite por IP na API",
      expression: "(starts_with(http.request.uri.path, \"/api/\"))",
      action: "block",
      ratelimit: {
        characteristics: ["cf.colo.id", "ip.src"],
        period: 10,
        requests_per_period: 100,
        mitigation_timeout: 10
      }
    }
  ]'
}

# Mescla: regras gerenciadas primeiro, depois as do painel que o script nao
# conhece (sem os campos somente leitura). Imprime o JSON do PUT.
mesclar() {
  local fase="$1" gerenciadas="$2" z="$3" atuais
  atuais=$(cf GET "/zones/${z}/rulesets/phases/${fase}/entrypoint" 2>/dev/null | jq '.result.rules // []' || echo '[]')
  jq -n --argjson g "$gerenciadas" --argjson a "$atuais" '
    ($g | map(.description)) as $nomes
    | { rules: ($g + [ $a[] | select(.description as $d | $nomes | index($d) | not)
                       | del(.id, .version, .last_updated, .ref, .logging) ]) }'
}

descrever_fase() {
  local fase="$1" z="$2" gerenciadas="$3" atuais
  echo "-- fase ${fase}"
  atuais=$(cf GET "/zones/${z}/rulesets/phases/${fase}/entrypoint" 2>/dev/null | jq '.result.rules // []' || echo '[]')
  jq -r --argjson g "$gerenciadas" '
    ($g | map(.description)) as $nomes
    | if length == 0 then "   (nenhuma regra)" else
        .[] | "   [\(if (.description as $d | $nomes | index($d)) then "gerenciada" else "do painel " end)] \(.description) \(if .enabled == false then "(DESLIGADA)" else "" end)"
      end' <<<"$atuais"
  jq -r --argjson a "$atuais" '.[] | .description as $d
    | select([$a[].description] | index($d) | not) | "   [FALTA   ] \($d)"' <<<"$gerenciadas"
}

configuracoes_alvo='{"ssl":"strict","always_use_https":"on","min_tls_version":"1.2","tls_1_3":"on","automatic_https_rewrites":"on"}'

descrever_configuracoes() {
  local z="$1" chave atual alvo
  echo "-- configuracoes da zona"
  for chave in $(jq -r 'keys[]' <<<"$configuracoes_alvo"); do
    atual=$(cf GET "/zones/${z}/settings/${chave}" | jq -r '.result.value')
    alvo=$(jq -r --arg k "$chave" '.[$k]' <<<"$configuracoes_alvo")
    printf '   %-26s atual=%-8s alvo=%s %s\n' "$chave" "$atual" "$alvo" "$([ "$atual" = "$alvo" ] && echo '[OK]' || echo '[MUDA]')"
  done
}

# Para cada registro A/AAAA/CNAME proxied, abre TLS direto na origem com o SNI
# do nome e exige certificado valido. So assim o "Full (strict)" nao derruba
# nenhum site da zona.
origens_com_certificado_valido() {
  local z="$1" falhou=0 nome conteudo tipo
  echo "-- certificado da origem de cada registro proxied"
  while IFS=$'\t' read -r tipo nome conteudo; do
    [ "$tipo" = "CNAME" ] && { echo "   [PULADO] $nome (CNAME para $conteudo: conferir a mao)"; continue; }
    local alvo="$conteudo"; [ "$tipo" = "AAAA" ] && alvo="[$conteudo]"
    if curl -sS -o /dev/null --max-time 10 --resolve "${nome}:443:${alvo}" "https://${nome}/" 2>/dev/null; then
      echo "   [OK]    $nome"
    else
      echo "   [FALHA] $nome: a origem nao apresenta certificado valido para esse nome"; falhou=1
    fi
  done < <(cf GET "/zones/${z}/dns_records?per_page=500" | jq -r '.result[] | select(.proxied and (.type=="A" or .type=="AAAA" or .type=="CNAME")) | [.type,.name,.content] | @tsv')
  return $falhou
}

cmd_status() {
  local z; z=$(zona_id); [ -n "$z" ] || { echo "[ERRO] zona ${ZONA} nao encontrada para este token" >&2; exit 1; }
  echo "Zona ${ZONA}"
  descrever_configuracoes "$z"
  descrever_fase http_request_cache_settings "$z" "$(regras_cache)"
  descrever_fase http_ratelimit "$z" "$(regras_limite)"
  origens_com_certificado_valido "$z" || true
}

aplicar_configuracoes() {
  local z="$1" chave valor
  for chave in $(jq -r 'keys[]' <<<"$configuracoes_alvo"); do
    valor=$(jq -r --arg k "$chave" '.[$k]' <<<"$configuracoes_alvo")
    if [ "$chave" = "ssl" ] && ! origens_com_certificado_valido "$z"; then
      echo "[AVISO] ssl=strict NAO aplicado: corrija os certificados acima e rode de novo."; continue
    fi
    cf PATCH "/zones/${z}/settings/${chave}" "$(jq -nc --arg v "$valor" '{value:$v}')" >/dev/null \
      && echo "[OK] ${chave}=${valor}" || echo "[AVISO] ${chave} nao aplicado (ver erro acima)"
  done
}

aplicar_fase() {
  local fase="$1" gerenciadas="$2" z="$3" corpo
  corpo=$(mesclar "$fase" "$gerenciadas" "$z")
  cf PUT "/zones/${z}/rulesets/phases/${fase}/entrypoint" "$corpo" >/dev/null \
    && echo "[OK] ${fase}: $(jq '.rules | length' <<<"$corpo") regra(s), $(jq length <<<"$gerenciadas") gerenciada(s)"
}

cmd_aplicar() {
  local z; z=$(zona_id); [ -n "$z" ] || { echo "[ERRO] zona ${ZONA} nao encontrada para este token" >&2; exit 1; }
  aplicar_configuracoes "$z"
  aplicar_fase http_request_cache_settings "$(regras_cache)" "$z"
  # O limite de taxa pode ser recusado pelo plano; isso nao desfaz o resto.
  aplicar_fase http_ratelimit "$(regras_limite)" "$z" || echo "[AVISO] limite de taxa nao aplicado (ver erro acima)"
  cat <<EOF

Conferir (a 1a chamada e MISS, as seguintes HIT):
  curl -sI https://${HOST}/            | grep -i 'cf-cache-status\|cache-control'
  curl -sI https://${HOST}/api/v1/ranking | grep -i 'cf-cache-status\|cache-control'
A API so vira HIT depois do deploy que passa a emitir Cache-Control (PR fix/api-resiliencia).
EOF
}

cmd_purgar() {
  local z; z=$(zona_id)
  cf POST "/zones/${z}/purge_cache" '{"purge_everything":true}' >/dev/null && echo "[OK] cache da zona ${ZONA} esvaziado"
}

case "${1:-}" in
  status)  cmd_status ;;
  aplicar) cmd_aplicar ;;
  purgar)  cmd_purgar ;;
  regras)  jq -n --argjson c "$(regras_cache)" --argjson l "$(regras_limite)" '{cache:$c, limite:$l}' ;;
  *) sed -n '2,40p' "$0"; exit 1 ;;
esac
