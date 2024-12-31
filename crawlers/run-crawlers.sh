#!/bin/bash

LOCKFILE="/tmp/crawlers.lock"
EXECUTION_CHECK_URL="http://api:3000/api/v1/execution-status"
BROKER_QUEUE="json-processor-queue"

# Verifica se já está rodando
if [ -f "$LOCKFILE" ]; then
    echo "[$(date)] Crawlers já estão em execução."
    exit 1
fi

# Cria o lockfile
touch "$LOCKFILE"

# Verifica o status no endpoint
response=$(curl -s -X GET "$EXECUTION_CHECK_URL")

# Extrai status e data usando grep e awk
status=$(echo "$response" | grep -o '"status":"[^"]*' | awk -F':' '{print $2}' | tr -d '"')
executed_at=$(echo "$response" | grep -o '"executed_at":"[^"]*' | awk -F':' '{print $2}' | tr -d '"')
today=$(date +%Y-%m-%d)

# Se já executou hoje, finaliza
if [[ "$status" == "RUNNING" && "$executed_at" == "$today" ]]; then
    echo "[$(date)] Os crawlers já foram executados hoje."
    rm -f "$LOCKFILE"
    exit 0
fi

# Atualiza o status para RUNNING
START_EXECUTION_JSON="{\"status\": \"RUNNING\", \"executed_at\": \"$today\"}"
curl -s -X POST -H "Content-Type: application/json" -d "$START_EXECUTION_JSON" "$EXECUTION_CHECK_URL"

# Executa os crawlers em paralelo
npm run start-contract &
npm run start-councilor &
npm run start-frequency &
npm run start-general-productivity &
npm run start-proposition &
npm run start-proposition-productivity &
npm run start-travel-expenses &

wait

# Marca a execução como concluída
COMPLETE_EXECUTION_JSON="{\"status\": \"READY\", \"executed_at\": \"$today\"}"
curl -s -X POST -H "Content-Type: application/json" -d "$COMPLETE_EXECUTION_JSON" "$EXECUTION_CHECK_URL"

# Envia mensagem para o RabbitMQ
node /app/crawlers/broker.js "$BROKER_QUEUE" "Crawlers executados com sucesso"

echo "[$(date)] Execução concluída."
rm -f "$LOCKFILE"
