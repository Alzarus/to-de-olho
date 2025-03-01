#!/bin/bash

set -e  # Interrompe a execução em caso de erro não tratado
mkdir -p /logs # Certifica-se que a pasta logs esta criada
EXECUTION_CHECK_URL="http://api:3000/api/v1/execution-status"
BROKER_QUEUE="json-processor-queue"
CHECK_INTERVAL=1800  # 30 minutos
MAX_RETRIES=3  # Número máximo de tentativas para cada crawler

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

update_status() {
    local status=$1
    local executed_at=$(date +%Y-%m-%d)
    local payload="{\"status\": \"$status\", \"executed_at\": \"$executed_at\"}"
    curl -s -X POST -H "Content-Type: application/json" -d "$payload" "$EXECUTION_CHECK_URL"
}

run_crawler_with_retries() {
    local crawler_name=$1
    local attempt=1
    local success=false

    while [[ $attempt -le $MAX_RETRIES ]]; do
        log "Executando o crawler '$crawler_name' (tentativa $attempt de $MAX_RETRIES)..."
        if npm run "$crawler_name" > "/logs/${crawler_name}.log" 2>&1; then
            log "Crawler '$crawler_name' concluído com sucesso."
            success=true
            break
        else
            log "Erro ao executar o crawler '$crawler_name' (tentativa $attempt). Verifique /logs/${crawler_name}.log para mais detalhes."
            attempt=$((attempt + 1))
            sleep 5  # Espera 5 segundos antes de tentar novamente
        fi
    done

    if [[ "$success" == false ]]; then
        log "Falha permanente no crawler '$crawler_name' após $MAX_RETRIES tentativas."
        return 1
    fi
}

log "Iniciando o monitoramento dos crawlers..."

while true; do
    log "Verificando o status de execução..."

    response=$(curl -s -X GET "$EXECUTION_CHECK_URL")
    status=$(echo "$response" | jq -r '.status')
    executed_at=$(echo "$response" | jq -r '.executed_at')
    today=$(date +%Y-%m-%d)

    if [[ "$status" == "RUNNING" ]]; then
        log "Os crawlers ainda estão rodando. Aguardando..."
        sleep "$CHECK_INTERVAL"
        continue
    fi

    if [[ "$status" == "COMPLETED" && "$executed_at" == "$today" ]]; then
        log "Os crawlers já foram executados hoje. Verificando novamente em 30 minutos..."
        sleep "$CHECK_INTERVAL"
        continue
    fi

    log "Iniciando execução dos crawlers..."
    update_status "RUNNING"

    declare -a crawlers=(
        "start-contract"
        "start-councilor"
        "start-frequency"
        "start-general-productivity"
        "start-proposition"
        "start-proposition-productivity"
        "start-travel-expenses"
    )

    errors=0
    for crawler in "${crawlers[@]}"; do
        if ! run_crawler_with_retries "$crawler"; then
            errors=$((errors + 1))
        fi
    done

    if [[ $errors -eq 0 ]]; then
        update_status "COMPLETED"
        node /app/crawlers/broker.js "$BROKER_QUEUE" "Crawlers executados com sucesso"
        log "Execução concluída com sucesso. Verificando novamente em 30 minutos..."
    else
        update_status "FAILED"
        log "Falha na execução de alguns crawlers. Verifique os logs em /logs. Verificando novamente em 30 minutos..."
    fi

    sleep "$CHECK_INTERVAL"
done
