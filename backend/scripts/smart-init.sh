#!/bin/sh
# Script de inicialização inteligente que roda automaticamente após o backend subir

echo "🤖 Iniciando verificação inteligente de backfill..."

# Aguardar backend estar disponível
echo "⏳ Aguardando backend ficar disponível..."
until curl -f http://backend:8080/api/v1/health >/dev/null 2>&1; do
    echo "   Backend não está pronto ainda, aguardando..."
    sleep 2
done

echo "✅ Backend disponível! Executando ingestor inteligente..."

# Executar ingestor em modo automático
exec ./ingestor -mode auto