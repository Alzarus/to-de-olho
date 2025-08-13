# 🐳 Script para verificar e iniciar Docker - Projeto "Tô De Olho"

Write-Host "🐳 Verificando Docker..." -ForegroundColor Cyan

# Verificar se Docker está instalado
try {
    $dockerVersion = docker --version
    Write-Host "✅ Docker encontrado: $dockerVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ Docker não está instalado!" -ForegroundColor Red
    Write-Host "📥 Instale o Docker Desktop em: https://www.docker.com/products/docker-desktop/" -ForegroundColor Yellow
    exit 1
}

# Verificar se Docker está rodando
try {
    docker info | Out-Null
    Write-Host "✅ Docker está rodando!" -ForegroundColor Green
} catch {
    Write-Host "❌ Docker não está rodando!" -ForegroundColor Red
    Write-Host "🚀 Tentando iniciar Docker Desktop..." -ForegroundColor Yellow
    
    # Tentar iniciar Docker Desktop
    try {
        Start-Process "C:\Program Files\Docker\Docker\Docker Desktop.exe" -WindowStyle Hidden
        Write-Host "⏳ Aguardando Docker Desktop inicializar..." -ForegroundColor Yellow
        
        # Aguardar até 60 segundos
        $timeout = 60
        $elapsed = 0
        
        while ($elapsed -lt $timeout) {
            try {
                docker info | Out-Null
                Write-Host "✅ Docker Desktop iniciado com sucesso!" -ForegroundColor Green
                break
            } catch {
                Start-Sleep -Seconds 2
                $elapsed += 2
                Write-Host "." -NoNewline
            }
        }
        
        if ($elapsed -ge $timeout) {
            Write-Host "`n❌ Timeout: Docker Desktop não iniciou em $timeout segundos" -ForegroundColor Red
            Write-Host "🔧 Inicie o Docker Desktop manualmente e execute este script novamente" -ForegroundColor Yellow
            exit 1
        }
        
    } catch {
        Write-Host "❌ Não foi possível iniciar Docker Desktop automaticamente" -ForegroundColor Red
        Write-Host "🔧 Inicie o Docker Desktop manualmente e execute este script novamente" -ForegroundColor Yellow
        exit 1
    }
}

Write-Host "`n🐳 Iniciando containers do Tô De Olho..." -ForegroundColor Cyan
Write-Host "=" * 50 -ForegroundColor Gray

# Opções para o usuário
Write-Host "`n📋 Escolha o ambiente:" -ForegroundColor White
Write-Host "1. 🚀 Completo (Backend + Frontend + Banco + Redis)" -ForegroundColor Yellow
Write-Host "2. 🗄️ Apenas Infraestrutura (Banco + Redis)" -ForegroundColor Yellow
Write-Host "3. 🔧 Desenvolvimento (Infraestrutura + Monitoramento)" -ForegroundColor Yellow
Write-Host "4. ❌ Cancelar" -ForegroundColor Red

$choice = Read-Host "`nDigite sua escolha (1-4)"

switch ($choice) {
    "1" {
        Write-Host "`n🚀 Subindo ambiente completo..." -ForegroundColor Green
        docker compose up -d --build
    }
    "2" {
        Write-Host "`n🗄️ Subindo apenas infraestrutura..." -ForegroundColor Green  
        docker compose up -d postgres redis
    }
    "3" {
        Write-Host "`n🔧 Subindo ambiente de desenvolvimento..." -ForegroundColor Green
        docker compose -f docker-compose.dev.yml up -d
    }
    "4" {
        Write-Host "❌ Operação cancelada" -ForegroundColor Red
        exit 0
    }
    default {
        Write-Host "❌ Opção inválida!" -ForegroundColor Red
        exit 1
    }
}

# Aguardar containers subirem
Write-Host "`n⏳ Aguardando containers iniciarem..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

# Verificar status dos containers
Write-Host "`n📊 Status dos containers:" -ForegroundColor White
docker compose ps

# Verificar health checks
Write-Host "`n🏥 Verificando health checks..." -ForegroundColor White

$containers = docker compose ps --format json | ConvertFrom-Json
foreach ($container in $containers) {
    $health = docker inspect $container.Name --format='{{.State.Health.Status}}' 2>$null
    if ($health) {
        $status = if ($health -eq "healthy") { "✅" } else { "⚠️" }
        Write-Host "   $status $($container.Name): $health" -ForegroundColor $(if ($health -eq "healthy") { "Green" } else { "Yellow" })
    } else {
        Write-Host "   ℹ️ $($container.Name): sem health check" -ForegroundColor Gray
    }
}

# Informações de acesso
Write-Host "`n" + "=" * 50 -ForegroundColor Gray
Write-Host "🎉 CONTAINERS INICIADOS!" -ForegroundColor Cyan  
Write-Host "=" * 50 -ForegroundColor Gray

Write-Host "`n🌐 ACESSE OS SERVIÇOS:" -ForegroundColor White
Write-Host "   Frontend:    http://localhost:3000" -ForegroundColor Cyan
Write-Host "   Backend API: http://localhost:8080" -ForegroundColor Cyan
Write-Host "   PostgreSQL:  localhost:5432" -ForegroundColor Gray
Write-Host "   Redis:       localhost:6379" -ForegroundColor Gray

if ($choice -eq "3") {
    Write-Host "   Grafana:     http://localhost:3001 (admin/admin123)" -ForegroundColor Gray
    Write-Host "   Prometheus:  http://localhost:9090" -ForegroundColor Gray
    Write-Host "   RabbitMQ:    http://localhost:15672 (admin/admin123)" -ForegroundColor Gray
}

Write-Host "`n🔧 COMANDOS ÚTEIS:" -ForegroundColor White
Write-Host "   docker compose logs -f           # Ver logs em tempo real" -ForegroundColor Gray
Write-Host "   docker compose ps                # Status dos containers" -ForegroundColor Gray
Write-Host "   docker compose down              # Parar containers" -ForegroundColor Gray
Write-Host "   docker compose down -v           # Parar e remover volumes" -ForegroundColor Gray

Write-Host "`n✅ Ambiente Docker configurado com sucesso!" -ForegroundColor Green
