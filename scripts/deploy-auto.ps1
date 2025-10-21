# 🚀 Script de Deploy Automático Inteligente - Tô De Olho (Windows)
# ==================================================================
# Autor: Sistema de Deploy Automatizado  
# Descrição: Deploy inteligente com controle automático de backfill

param(
    [switch]$Force = $false,
    [switch]$SkipBuild = $false,
    [int]$StartYear = 2022,
    [int]$EndYear = 0,
    [switch]$ForceBackfill = $false
)

# Configuração de cores para output
$Host.UI.RawUI.WindowTitle = "Tô De Olho - Deploy Inteligente"

Write-Host "🚀 Iniciando Deploy Automático Inteligente..." -ForegroundColor Green
Write-Host "🧠 Sistema verifica automaticamente se backfill é necessário" -ForegroundColor Cyan
Write-Host "📅 Período configurado: $StartYear-$(if($EndYear -eq 0){'atual'}else{$EndYear})" -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Yellow

# Definir ano final se não especificado
if ($EndYear -eq 0) {
    $EndYear = (Get-Date).Year
}

try {
    # 1. Verificar se Docker está rodando
    Write-Host "🔍 Verificando Docker..." -ForegroundColor Blue
    docker info *>$null
    if ($LASTEXITCODE -ne 0) {
        throw "❌ Docker não está rodando. Por favor, inicie o Docker Desktop e tente novamente."
    }

    # 2. Parar containers existentes (se houver)
    if ($Force) {
        Write-Host "🧹 Limpando ambiente existente (force mode)..." -ForegroundColor Yellow
        docker-compose down -v 2>$null
    }

    # 3. Rebuild das imagens (opcional)
    if (-not $SkipBuild) {
        Write-Host "🔨 Construindo imagens Docker..." -ForegroundColor Blue
        docker-compose build --no-cache
        if ($LASTEXITCODE -ne 0) { throw "Falha no build das imagens" }
    }

    # 4. Iniciar serviços de infraestrutura
    Write-Host "🗃️ Iniciando banco de dados e cache..." -ForegroundColor Blue
    docker-compose up -d postgres redis
    if ($LASTEXITCODE -ne 0) { throw "Falha ao iniciar infraestrutura" }

    # 5. Aguardar banco ficar ready
    Write-Host "⏳ Aguardando PostgreSQL ficar disponível..." -ForegroundColor Yellow
    $timeout = 60
    $elapsed = 0
    do {
        Start-Sleep -Seconds 2
        $elapsed += 2
        docker-compose exec -T postgres pg_isready -U postgres 2>$null
        $ready = ($LASTEXITCODE -eq 0)
    } while (-not $ready -and $elapsed -lt $timeout)
    
    if (-not $ready) { throw "PostgreSQL não ficou disponível em ${timeout}s" }

    # 6. Executar migrations
    Write-Host "📋 Executando migrations do banco..." -ForegroundColor Blue
    docker-compose run --rm backend ./server -migrate-only
    if ($LASTEXITCODE -ne 0) { throw "Falha nas migrations" }

    # 7. Iniciar backend para verificar backfill
    Write-Host "🔧 Iniciando backend temporariamente..." -ForegroundColor Blue
    docker-compose up -d backend
    Start-Sleep -Seconds 5

    # 8. Sistema Inteligente de Backfill
    Write-Host "" -ForegroundColor White
    Write-Host "🧠 Verificando necessidade de Backfill Histórico..." -ForegroundColor Green
    Write-Host "   📊 Período: $StartYear-$EndYear" -ForegroundColor Cyan
    Write-Host "   🎯 O sistema decide automaticamente se precisa rodar" -ForegroundColor Cyan
    Write-Host "" -ForegroundColor White

    # Definir variáveis de ambiente para o backfill
    $env:BACKFILL_START_YEAR = $StartYear
    $env:BACKFILL_END_YEAR = $EndYear
    $env:BACKFILL_TRIGGERED_BY = "deploy"
    if ($ForceBackfill) {
        $env:BACKFILL_FORCE = "true"
        Write-Host "⚠️ Forçando reexecução de backfill..." -ForegroundColor Yellow
    }

    # Executar ingestor inteligente
    Write-Host "🤖 Executando ingestor inteligente..." -ForegroundColor Blue
    
    # Capturar output do ingestor para verificar se executou backfill
    $ingestorOutput = docker-compose run --rm ingestor 2>&1
    $ingestorExitCode = $LASTEXITCODE

    Write-Host $ingestorOutput -ForegroundColor White

    if ($ingestorExitCode -ne 0) {
        # Verificar se falhou por backfill já executado
        if ($ingestorOutput -match "backfill.*não necessário|já foi executado|já em andamento") {
            Write-Host "✅ Backfill não necessário - dados já atualizados!" -ForegroundColor Green
        } else {
            throw "Falha no ingestor inteligente"
        }
    } else {
        Write-Host "✅ Ingestor executado com sucesso!" -ForegroundColor Green
    }

    # 9. Iniciar aplicação completa
    Write-Host "🚀 Iniciando aplicação completa..." -ForegroundColor Blue
    docker-compose up -d frontend scheduler
    if ($LASTEXITCODE -ne 0) { throw "Falha ao iniciar aplicação" }

    # 10. Aguardar aplicação ficar ready
    Write-Host "⏳ Aguardando aplicação ficar disponível..." -ForegroundColor Yellow
    $timeout = 60
    $elapsed = 0
    do {
        Start-Sleep -Seconds 2
        $elapsed += 2
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing -TimeoutSec 5
            $ready = ($response.StatusCode -eq 200)
        } catch {
            $ready = $false
        }
    } while (-not $ready -and $elapsed -lt $timeout)

    # 11. Verificar status do backfill
    Write-Host "" -ForegroundColor White
    Write-Host "📊 Verificando status do backfill..." -ForegroundColor Blue
    
    try {
        $backfillStatus = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/backfill/status" -TimeoutSec 10
        if ($backfillStatus) {
            Write-Host "🎯 Status do Backfill:" -ForegroundColor Cyan
            Write-Host "   • Execução: $($backfillStatus.execution_id)" -ForegroundColor White
            Write-Host "   • Status: $($backfillStatus.status)" -ForegroundColor White
            Write-Host "   • Progresso: $($backfillStatus.progress_percentage.ToString('F1'))%" -ForegroundColor White
            Write-Host "   • Deputados: $($backfillStatus.deputados_processados)" -ForegroundColor White
            Write-Host "   • Proposições: $($backfillStatus.proposicoes_processadas)" -ForegroundColor White
        }
    } catch {
        Write-Host "ℹ️ Status de backfill não disponível (normal se não executado)" -ForegroundColor DarkYellow
    }

    # 12. Status final
    Write-Host "" -ForegroundColor White
    Write-Host "✅ Deploy Inteligente concluído com sucesso!" -ForegroundColor Green
    Write-Host "================================================================" -ForegroundColor Yellow
    Write-Host "🌐 Frontend: http://localhost:3000" -ForegroundColor Cyan
    Write-Host "🔗 API Backend: http://localhost:8080" -ForegroundColor Cyan
    Write-Host "📊 Adminer (DB): http://localhost:8081" -ForegroundColor Cyan
    Write-Host "📈 Métricas: docker-compose logs scheduler" -ForegroundColor Cyan
    Write-Host "🤖 Backfill Status: curl http://localhost:8080/api/v1/backfill/status" -ForegroundColor Cyan
    Write-Host "" -ForegroundColor White
    Write-Host "🔄 Próxima sincronização: Diária às 6h (automática)" -ForegroundColor Yellow
    Write-Host "📋 Verificar dados: curl http://localhost:8080/api/v1/deputados" -ForegroundColor Yellow
    Write-Host "" -ForegroundColor White
    Write-Host "🎯 Sistema Inteligente ativo - Backfill automático conforme necessário!" -ForegroundColor Green

} catch {
    Write-Host "" -ForegroundColor White
    Write-Host "❌ Erro durante o deploy: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "" -ForegroundColor White
    Write-Host "🔧 Para debug, execute:" -ForegroundColor Yellow
    Write-Host "   docker-compose logs backend" -ForegroundColor Cyan
    Write-Host "   docker-compose logs ingestor" -ForegroundColor Cyan
    Write-Host "   docker-compose ps" -ForegroundColor Cyan
    exit 1
} finally {
    # Limpar variáveis de ambiente
    Remove-Item Env:BACKFILL_START_YEAR -ErrorAction SilentlyContinue
    Remove-Item Env:BACKFILL_END_YEAR -ErrorAction SilentlyContinue
    Remove-Item Env:BACKFILL_FORCE -ErrorAction SilentlyContinue
    Remove-Item Env:BACKFILL_TRIGGERED_BY -ErrorAction SilentlyContinue
}

# Perguntar se quer abrir o browser
$choice = Read-Host "`n🌐 Abrir aplicação no navegador? (s/N)"
if ($choice -eq 's' -or $choice -eq 'S') {
    Start-Process "http://localhost:3000"
}