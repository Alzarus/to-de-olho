#!/bin/bash

EXECUTION_CHECK_URL="http://api:3000/api/v1/execution-status"
BROKER_QUEUE="json-processor-queue"
CHECK_INTERVAL=1800 # 30 minutos em segundos

while true; do
    echo "[$(date)] Verificando o status de execução..."

    # Verifica o status no endpoint
    response=$(curl -s -X GET "$EXECUTION_CHECK_URL")

    # Extrai os valores necessários usando jq
    status=$(echo "$response" | jq -r '.status')
    executed_at=$(echo "$response" | jq -r '.executed_at')
    today=$(date +%Y-%m-%d)

    # Se já estiver em execução, aguarda
    if [[ "$status" == "RUNNING" ]]; then
        echo "[$(date)] Os crawlers ainda estão rodando. Aguardando..."
        sleep "$CHECK_INTERVAL"
        continue
    fi

    # Confirma se já foi executado com sucesso hoje
    if [[ "$status" == "COMPLETED" && "$executed_at" == "$today" ]]; then
        echo "[$(date)] Os crawlers já foram executados hoje. Verificando novamente em 30 minutos..."
        sleep "$CHECK_INTERVAL"
        continue
    fi

    echo "[$(date)] Iniciando execução dos crawlers..."

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
    
    # Aguarda a execução dos crawlers e captura erros
    if wait; then
        # Marca como COMPLETED se os crawlers rodarem corretamente
        COMPLETE_EXECUTION_JSON="{\"status\": \"COMPLETED\", \"executed_at\": \"$today\"}"
        curl -s -X POST -H "Content-Type: application/json" -d "$COMPLETE_EXECUTION_JSON" "$EXECUTION_CHECK_URL"

        # Envia mensagem para o RabbitMQ
        node /app/crawlers/broker.js "$BROKER_QUEUE" "Crawlers executados com sucesso"

        echo "[$(date)] Execução concluída com sucesso. Verificando novamente em 30 minutos..."
    else
        # Marca como FAILED se houver erro
        FAILED_EXECUTION_JSON="{\"status\": \"FAILED\", \"executed_at\": \"$today\"}"
        curl -s -X POST -H "Content-Type: application/json" -d "$FAILED_EXECUTION_JSON" "$EXECUTION_CHECK_URL"

        echo "[$(date)] Falha na execução dos crawlers. Verificando novamente em 30 minutos..."
    fi

    sleep "$CHECK_INTERVAL"
done
