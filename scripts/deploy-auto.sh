#!/bin/bash
# 🚀 Script de Deploy Automático - Tô De Olho
# ============================================
# Autor: Sistema de Deploy Automatizado
# Descrição: Configura e executa ingestão completa de dados (2022+)

set -e  # Exit on any error

echo "🚀 Iniciando Deploy Automático do Tô De Olho..."
echo "📅 Período de Ingestão: 2022-$(date +%Y) (dados históricos + atuais)"
echo "=================================================="

# 1. Verificar se Docker está rodando
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker não está rodando. Por favor, inicie o Docker e tente novamente."
    exit 1
fi

# 2. Parar containers existentes (se houver)
echo "🧹 Limpando ambiente existente..."
docker-compose down -v 2>/dev/null || true

# 3. Rebuild das imagens
echo "🔨 Construindo imagens Docker..."
docker-compose build --no-cache

# 4. Iniciar serviços de infraestrutura
echo "🗃️ Iniciando banco de dados e cache..."
docker-compose up -d postgres redis

# 5. Aguardar banco ficar ready
echo "⏳ Aguardando PostgreSQL ficar disponível..."
timeout 60 bash -c 'until docker-compose exec -T postgres pg_isready -U postgres; do sleep 2; done'

# 6. Executar migrations
echo "📋 Executando migrations do banco..."
docker-compose run --rm backend ./server -migrate-only

# 7. Executar backfill estratégico (2022+)
echo "🔄 Iniciando Backfill Estratégico (2022-$(date +%Y))..."
echo "   📊 Ordem: Deputados → Proposições → Despesas"
echo "   🎯 Volume esperado: ~50k proposições + ~500k despesas"
echo "   ⏱️ Tempo estimado: 15-30 minutos"
echo ""
# Exportar flag para indicar que um backfill está rodando. Isso evita que o scheduler
# dispare sincronizações iniciais enquanto o backfill estiver em progresso.
export BACKFILL_RUNNING=true
docker-compose run --rm ingestor
# Limpar flag após o backfill terminar
unset BACKFILL_RUNNING

# 8. Iniciar aplicação completa
echo "🚀 Iniciando aplicação completa..."
docker-compose up -d backend frontend scheduler

# 9. Aguardar aplicação ficar ready
echo "⏳ Aguardando aplicação ficar disponível..."
timeout 60 bash -c 'until curl -f http://localhost:8080/health > /dev/null 2>&1; do sleep 2; done'

# 10. Status final
echo ""
echo "✅ Deploy concluído com sucesso!"
echo "=================================================="
echo "🌐 Frontend: http://localhost:3000"
echo "🔗 API Backend: http://localhost:8080"
echo "📊 Adminer (DB): http://localhost:8081"
echo "📈 Métricas: docker-compose logs scheduler"
echo ""
echo "🔄 Próxima sincronização: Diária às 6h (automática)"
echo "📋 Para verificar dados: curl http://localhost:8080/api/v1/deputados"
echo ""
echo "🎯 Sistema pronto para uso com dados atualizados!"