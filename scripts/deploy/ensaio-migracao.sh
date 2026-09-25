#!/usr/bin/env bash
# Ensaio da migração, antes de trocar a versão no ar (deploy.yml, VPS).
#
# Por quê: em 25/09/2026 o Dependabot #50 levou o GORM a 1.31.2, e o
# AutoMigrate novo tentou "DROP CONSTRAINT uni_despesas_ceaps_id_origem", que
# não existe no banco de produção. O CI testa com banco vazio e não pegou. O
# deploy trocou a API, ela entrou em loop, e o site ficou 20 min com 500.
#
# O que faz, sem tocar no banco de produção nem nos contêineres no ar:
#   1. copia o banco de produção para um banco temporário no mesmo Postgres
#      (pg_dump | psql; CREATE DATABASE ... TEMPLATE exigiria derrubar as
#      conexões da API, o que tiraria o site do ar);
#   2. sobe a imagem NOVA da API num contêiner descartável apontando para a
#      cópia, sem portas publicadas e com o scheduler desligado;
#   3. exige /health 200 dentro do prazo (a API só abre a porta depois do
#      AutoMigrate e das rotinas idempotentes de subida);
#   4. apaga contêiner e cópia sempre, com sucesso, falha ou interrupção.
#
# Uso (na VPS): bash ensaio-migracao.sh <sha de 40>
# Saída 0 = pode subir; qualquer outra = não suba esta versão.
set -euo pipefail

tag="${1:?uso: ensaio-migracao.sh <sha de 40>}"
dir="${TODEOLHO_DIR:-/opt/todeolho}"
prazo="${ENSAIO_PRAZO_S:-360}"
banco_ct="${TODEOLHO_DB_CT:-todeolho-db}"
imagem="${TODEOLHO_REGISTRO:-ghcr.io/alzarus}/todeolho-api:$tag"
copia=todeolho_ensaio
conteiner=todeolho-ensaio

ler_env() { grep -m1 "^$1=" "$dir/.env" | cut -d= -f2-; }
# Senha na URL do Postgres: @, :, / e % quebrariam o parse sem codificar.
urlenc() {
  local s="$1" i c saida=""
  for ((i = 0; i < ${#s}; i++)); do
    c=${s:i:1}
    case $c in
      [a-zA-Z0-9._~-]) saida+=$c ;;
      *) printf -v c '%%%02X' "'$c"; saida+=$c ;;
    esac
  done
  printf '%s' "$saida"
}
usuario=$(ler_env POSTGRES_USER)
senha=$(ler_env POSTGRES_PASSWORD)
banco=$(ler_env POSTGRES_DB)

# </dev/null: o script chega pelo stdin do ssh; nenhum comando pode consumi-lo.
sql() { docker exec "$banco_ct" psql -U "$usuario" -d postgres -v ON_ERROR_STOP=1 -Atqc "$1" </dev/null; }

envs=""
limpar() {
  docker rm -f "$conteiner" >/dev/null 2>&1 || true
  # WITH (FORCE) derruba conexões que o contêiner tenha deixado abertas.
  sql "DROP DATABASE IF EXISTS $copia WITH (FORCE)" >/dev/null 2>&1 \
    || echo "[AVISO] não consegui apagar o banco $copia; apague a mão"
  [ -z "$envs" ] || rm -f "$envs"
}
# Sem o trap de sinais, um cancelamento do job (ssh cai, bash recebe HUP)
# mataria o script sem passar pelo EXIT e deixaria a cópia no disco.
trap limpar EXIT
trap 'exit 130' INT TERM HUP
# Sobra de um ensaio anterior interrompido de forma que nem o trap rodou.
limpar
envs=$(mktemp)

falhar() {
  echo "::error::ensaio da migração: $1"
  echo "[FALHA] a versão $tag NÃO foi implantada; a produção segue intacta."
  exit 1
}

# Espaço: a cópia ocupa o mesmo que o banco (dados + índices). Exige o dobro
# livre, para o ensaio não encher o disco de que o banco de produção depende.
tamanho=$(sql "select pg_database_size('$banco')")
livre=$(( $(docker exec "$banco_ct" df -Pk /var/lib/postgresql/data </dev/null | awk 'NR==2 {print $4}') * 1024 ))
echo "[INFO] banco $banco: $((tamanho / 1048576)) MB; livre no volume: $((livre / 1048576)) MB"
[ "$livre" -gt $((tamanho * 2)) ] || falhar "espaço livre insuficiente para copiar o banco"

inicio=$SECONDS
sql "CREATE DATABASE $copia"
# O dump é um snapshot MVCC: não bloqueia a API, que segue lendo e gravando.
docker exec "$banco_ct" sh -c \
  'set -o pipefail; pg_dump -U "$1" -d "$2" --no-owner --no-privileges | psql -U "$1" -d "$3" -q -v ON_ERROR_STOP=1 >/dev/null' \
  sh "$usuario" "$banco" "$copia" </dev/null \
  || falhar "não consegui copiar o banco de produção"
echo "[OK] cópia do banco em $((SECONDS - inicio)) s"

# Mesma rede do banco; nenhuma porta publicada. Sem SYNC_SECRET, as rotas de
# sync ficam fechadas (falha fechada em internal/api/sync_auth.go); sem a
# chave do Portal da Transparência e com SCHEDULER_DESATIVADO, nenhuma chamada
# externa sai daqui. Credenciais por arquivo, não na linha de comando (ps).
rede=$(docker inspect -f '{{range $k, $v := .NetworkSettings.Networks}}{{$k}} {{end}}' "$banco_ct" | awk '{print $1}')
cat > "$envs" <<EOF
DATABASE_URL=postgres://$(urlenc "$usuario"):$(urlenc "$senha")@$banco_ct:5432/$copia?sslmode=disable
PORT=8080
GIN_MODE=release
SCHEDULER_DESATIVADO=true
DB_MAX_OPEN_CONNS=5
DB_MAX_IDLE_CONNS=1
EOF
docker run -d --name "$conteiner" --network "$rede" --env-file "$envs" \
  --memory 1g --restart no "$imagem" >/dev/null </dev/null

inicio=$SECONDS
while :; do
  if [ "$(docker inspect -f '{{.State.Running}}' "$conteiner")" != true ]; then
    docker logs --tail 40 "$conteiner" 2>&1 | cut -c1-400
    falhar "a API nova encerrou durante a subida contra a cópia do banco (veja o log acima)"
  fi
  # A imagem é distroless: o próprio binário faz o GET em /health.
  if docker exec "$conteiner" /server healthcheck </dev/null 2>/dev/null; then
    echo "[OK] ensaio: a API $tag subiu sobre a cópia do banco de produção em $((SECONDS - inicio)) s"
    exit 0
  fi
  if [ $((SECONDS - inicio)) -ge "$prazo" ]; then
    docker logs --tail 40 "$conteiner" 2>&1 | cut -c1-400
    falhar "a API nova não respondeu /health em ${prazo} s"
  fi
  sleep 5
done
