# Script de Bootstrap - Cold Start do Sistema "Tô De Olho"
# Executa carga inicial de dados da Câmara dos Deputados

param(
    [Parameter(Mandatory=$false)]
    [string]$Mode = "development",
    
    [Parameter(Mandatory=$false)]
    [int]$Workers = 4,
    
    [Parameter(Mandatory=$false)]
    [switch]$SkipCache,
    
    [Parameter(Mandatory=$false)]
    [switch]$FullSync
)

Write-Host "🏛️ Iniciando Bootstrap do Sistema Tô De Olho..." -ForegroundColor Cyan
Write-Host "📊 Modo: $Mode | Workers: $Workers" -ForegroundColor Gray

# Verificar se Docker está rodando
Write-Host "🐳 Verificando Docker..." -ForegroundColor Yellow
$dockerRunning = docker info 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Docker não está rodando. Execute 'docker-compose -f docker-compose.dev.yml up -d'" -ForegroundColor Red
    exit 1
}

# Verificar se serviços estão saudáveis
Write-Host "🔍 Verificando saúde dos serviços..." -ForegroundColor Yellow
$services = @("todo-postgres", "todo-redis", "todo-rabbitmq")

foreach ($service in $services) {
    $health = docker inspect --format='{{.State.Health.Status}}' $service 2>$null
    if ($health -ne "healthy") {
        Write-Host "⚠️ Serviço $service não está saudável. Status: $health" -ForegroundColor Yellow
        Write-Host "🔄 Aguardando serviços ficarem prontos..." -ForegroundColor Yellow
        Start-Sleep -Seconds 10
    }
}

# Executar migrações
Write-Host "📋 Executando migrações do banco..." -ForegroundColor Green
& go run scripts/migrate.go up

# Cold start - ingestão de dados
Write-Host "🚀 Iniciando Cold Start - Ingestão de Dados..." -ForegroundColor Green

if ($FullSync) {
    Write-Host "📊 Modo: Sincronização Completa (4 anos de dados)" -ForegroundColor Cyan
    & go run cmd/bootstrap/main.go --mode=full-sync --years=4 --workers=$Workers
} else {
    Write-Host "⚡ Modo: Sincronização Essencial (6 meses)" -ForegroundColor Cyan
    & go run cmd/bootstrap/main.go --mode=essential --months=6 --workers=$Workers
}

# Popular dados demo
if ($Mode -eq "development") {
    Write-Host "🎮 Populando dados de demonstração..." -ForegroundColor Magenta
    & go run cmd/seed/demo-data.go
}

# Cache warmup
if (-not $SkipCache) {
    Write-Host "🔥 Aquecendo cache..." -ForegroundColor Yellow
    & go run scripts/cache-warmup.go
}

Write-Host "✅ Bootstrap concluído com sucesso!" -ForegroundColor Green
Write-Host ""
Write-Host "🌐 Acesse http://localhost:3000 para ver a aplicação" -ForegroundColor Cyan
Write-Host "📊 Grafana: http://localhost:3001 (admin:admin123)" -ForegroundColor Gray
Write-Host "🐰 RabbitMQ: http://localhost:15672 (admin:admin123)" -ForegroundColor Gray
