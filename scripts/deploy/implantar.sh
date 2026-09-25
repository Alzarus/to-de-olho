#!/usr/bin/env bash
# Troca a versão no ar, confere a saúde e, se falhar, volta sozinho para a
# versão da tag `producao` (deploy.yml, VPS).
#
# Por quê: em 25/09/2026 o deploy do #53 falhou no health check e deixou a
# versão quebrada no ar; o site ficou 20 min com 500 até o rollback manual.
#
# Contrato:
#   - saída 0: a versão nova está no ar e saudável;
#   - saída 1: a versão nova falhou. Se havia uma `producao` diferente dela,
#     ela foi restaurada (e a linha RESULTADO diz se ficou saudável). Nunca há
#     segunda tentativa de rollback: se a própria reserva falhar, o script
#     para e pede intervenção, em vez de alternar versões em loop.
#   - Linhas "RESULTADO:" vão para o resumo do job (deploy.yml).
#
# A reserva vem das imagens locais `:producao`, baixadas no passo de pull
# (que falha se não conseguir baixá-las, exceto quando a tag não existe).
# Assim o rollback não depende da rede da VPS, que já falhou no pull em 25/09.
#
# Limite conhecido: o rollback troca imagens, não desfaz migração de schema.
# Coluna ou tabela nova que o AutoMigrate criou continua lá; a versão antiga
# as ignora. Migração destrutiva é barrada antes, pelo ensaio-migracao.sh.
#
# Uso (na VPS): bash implantar.sh <sha de 40>
set -euo pipefail

nova="${1:?uso: implantar.sh <sha de 40>}"
dir="${TODEOLHO_DIR:-/opt/todeolho}"
registro="${TODEOLHO_REGISTRO:-ghcr.io/alzarus}"
# A API só abre a porta depois do AutoMigrate e do recálculo da pontuação; uma
# migração pesada já levou mais de 2 min (deploys do #40 e #41).
prazo_api="${PRAZO_API_S:-360}"
host_publico="${HOST_PUBLICO:-todeolho.org}"

compose=(docker compose -f "$dir/docker-compose.contabo.yml" --env-file "$dir/.env")

# IMAGE_TAG no .env: o compose lê de lá, e um `docker compose logs` ou
# `restart` manual depois enxerga a versão que ficou no ar.
subir() {
  sed -i '/^IMAGE_TAG=/d' "$dir/.env"
  printf 'IMAGE_TAG=%s\n' "$1" >> "$dir/.env"
  # IMAGE_TAG também no ambiente do comando: variável de ambiente vale mais
  # que o .env para o compose, e uma IMAGE_TAG herdada da sessão faria o
  # rollback subir de novo a versão que acabou de falhar.
  # </dev/null: o script chega pelo stdin do ssh; nenhum comando pode consumi-lo.
  IMAGE_TAG="$1" "${compose[@]}" up -d --no-build --remove-orphans </dev/null
}

id_imagem() { docker image inspect -f '{{.Id}}' "$registro/todeolho-$1:$2" 2>/dev/null; }

# Imprime a tag a restaurar: o SHA que as duas imagens `:producao` locais
# também carregam (fica legível no log e no .env) ou, sem ele, a própria
# `producao`. Vazio se não houver reserva.
resolver_reserva() {
  local web sha
  id_imagem api producao >/dev/null || return 0
  web=$(id_imagem web producao) || return 0
  for sha in $(docker image inspect -f '{{join .RepoTags " "}}' "$registro/todeolho-api:producao" \
      | tr ' ' '\n' | sed -n 's|.*:\([0-9a-f]\{40\}\)$|\1|p'); do
    if [ "$(id_imagem web "$sha")" = "$web" ]; then echo "$sha"; return 0; fi
  done
  echo producao
}

logs() { "${compose[@]}" logs --tail=40 "$1" </dev/null 2>&1 | cut -c1-400; }

api_saudavel() {
  local inicio=$SECONDS
  # Pela rede interna, a partir do contêiner web (a API não publica porta).
  until docker exec todeolho-web node -e \
      "require('http').get('http://api:8080/health',r=>process.exit(r.statusCode===200?0:1)).on('error',()=>process.exit(1))" \
      </dev/null 2>/dev/null; do
    # API em loop de reinício (o caso do GORM em 25/09) não vai se recuperar:
    # não faz sentido esperar os 6 min com o site fora do ar.
    reinicios=$(docker inspect -f '{{.RestartCount}}' todeolho-api 2>/dev/null || echo 0)
    if [ "$reinicios" -ge 3 ]; then
      echo "[FALHA] API reiniciou $reinicios vezes desde a subida"; logs api; return 1
    fi
    if [ $((SECONDS - inicio)) -ge "$prazo_api" ]; then
      echo "[FALHA] API sem /health 200 em ${prazo_api} s"; logs api; return 1
    fi
    sleep 5
  done
  echo "[OK] API respondendo /health em $((SECONDS - inicio)) s"
}

# GET com até 6 tentativas (30 s): cobre a subida do Next e o NPM reabrindo
# a conexão com o upstream novo.
http_200() {
  local rotulo="$1" code="" _; shift
  for _ in 1 2 3 4 5 6; do
    code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 15 "$@" || true)
    if [ "$code" = 200 ]; then echo "[OK] $rotulo -> 200"; return 0; fi
    sleep 5
  done
  echo "[FALHA] $rotulo -> HTTP $code"; return 1
}

# Pelo mesmo caminho do público: porta 443 do Nginx Proxy Manager, com o SNI
# e o Host de produção. A porta de loopback sozinha não basta: no incidente
# do Quem Votar (25/09) ela dava 200 enquanto o NPM devolvia 502.
publico() {
  http_200 "https://$host_publico$1 (NPM, 443)" -k \
    --resolve "$host_publico:443:127.0.0.1" "https://$host_publico$1"
}

saudavel() {
  api_saudavel || return 1
  http_200 "web no loopback (127.0.0.1:5350)" http://127.0.0.1:5350/ || { logs web; return 1; }
  publico / || { logs web; return 1; }
  publico /api/v1/ranking || { logs api; return 1; }
}

mesma_versao() {
  [ "$1" = "$2" ] && return 0
  [ "$(id_imagem api "$1")" = "$(id_imagem api "$2")" ] \
    && [ "$(id_imagem web "$1")" = "$(id_imagem web "$2")" ]
}

reserva=$(resolver_reserva)
echo "[INFO] versão nova: $nova; reserva (producao): ${reserva:-nenhuma}"

subir "$nova"
if saudavel; then
  echo "RESULTADO: versão $nova no ar e saudável"
  exit 0
fi

echo "::error::a versão $nova falhou no health check"
if [ -z "$reserva" ]; then
  echo "RESULTADO: FALHA sem rollback. Não existe tag producao; a versão $nova segue no ar quebrada. Rollback manual: gh workflow run deploy.yml -f tag=<sha bom>"
  exit 1
fi
if mesma_versao "$nova" "$reserva"; then
  echo "RESULTADO: FALHA sem rollback. A producao ($reserva) é a mesma versão que acabou de falhar; nada a restaurar. Investigue a VPS (banco, rede, NPM)."
  exit 1
fi

echo "[ROLLBACK] restaurando a producao: $reserva"
subir "$reserva"
if saudavel; then
  echo "RESULTADO: ROLLBACK. A versão $nova falhou; a producao $reserva foi restaurada e está saudável. O job termina em falha de propósito."
else
  echo "RESULTADO: ROLLBACK FALHOU. A versão $nova falhou, e a producao $reserva também não ficou saudável. Nenhuma nova tentativa; intervenção manual."
fi
exit 1
