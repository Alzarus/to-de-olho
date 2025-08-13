# 🚀 Script para subir o ambiente completo do "Tô De Olho"

Write-Host "🏛️ Iniciando Tô De Olho - Plataforma de Transparência Política" -ForegroundColor Cyan
Write-Host "=" * 60 -ForegroundColor Gray

# Verificar se estamos na pasta correta
if (-not (Test-Path "backend") -or -not (Test-Path "frontend")) {
    Write-Host "❌ Execute este script na pasta raiz do projeto!" -ForegroundColor Red
    exit 1
}

# Função para iniciar processo em background
function Start-BackgroundProcess {
    param($Name, $WorkingDirectory, $Command, $Arguments)
    
    Write-Host "🚀 Iniciando $Name..." -ForegroundColor Green
    
    $processInfo = New-Object System.Diagnostics.ProcessStartInfo
    $processInfo.FileName = $Command
    $processInfo.Arguments = $Arguments
    $processInfo.WorkingDirectory = $WorkingDirectory
    $processInfo.UseShellExecute = $false
    $processInfo.CreateNoWindow = $false
    
    $process = [System.Diagnostics.Process]::Start($processInfo)
    return $process
}

try {
    # 1. Iniciar Backend (Go)
    Write-Host "`n1️⃣ Subindo Backend (Go + Gin)..." -ForegroundColor Yellow
    $backendPath = Join-Path $PWD "backend"
    $backendProcess = Start-BackgroundProcess "Backend" $backendPath "go" "run ."
    
    Write-Host "   📊 Backend rodando em: http://localhost:8080" -ForegroundColor Green
    Write-Host "   🏥 Health check: http://localhost:8080/api/v1/health" -ForegroundColor Green
    
    # Aguardar backend inicializar
    Start-Sleep -Seconds 3
    
    # 2. Iniciar Frontend (Next.js)
    Write-Host "`n2️⃣ Subindo Frontend (Next.js 15)..." -ForegroundColor Yellow
    $frontendPath = Join-Path $PWD "frontend"
    $frontendProcess = Start-BackgroundProcess "Frontend" $frontendPath "npm" "run dev"
    
    Write-Host "   🎨 Frontend rodando em: http://localhost:3000" -ForegroundColor Green
    
    # Aguardar frontend inicializar
    Start-Sleep -Seconds 5
    
    # 3. Testar APIs
    Write-Host "`n3️⃣ Testando conexões..." -ForegroundColor Yellow
    
    try {
        $healthCheck = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/health" -Method GET -TimeoutSec 5
        Write-Host "   ✅ Backend: $($healthCheck.message)" -ForegroundColor Green
    } catch {
        Write-Host "   ⚠️ Backend ainda não está respondendo (pode estar subindo)" -ForegroundColor Yellow
    }
    
    # 4. Informações importantes
    Write-Host "`n" + "=" * 60 -ForegroundColor Gray
    Write-Host "🎉 TÔ DE OLHO ESTÁ RODANDO!" -ForegroundColor Cyan
    Write-Host "=" * 60 -ForegroundColor Gray
    
    Write-Host "`n📱 ACESSE A APLICAÇÃO:" -ForegroundColor White
    Write-Host "   🌐 Frontend: http://localhost:3000" -ForegroundColor Cyan
    Write-Host "   🔧 API Backend: http://localhost:8080/api/v1" -ForegroundColor Cyan
    
    Write-Host "`n🔍 ENDPOINTS DISPONÍVEIS:" -ForegroundColor White
    Write-Host "   🏥 Health Check: http://localhost:8080/api/v1/health" -ForegroundColor Gray
    Write-Host "   👥 Deputados: http://localhost:8080/api/v1/deputados" -ForegroundColor Gray
    Write-Host "   💰 Despesas: http://localhost:8080/api/v1/deputados/{id}/despesas" -ForegroundColor Gray
    
    Write-Host "`n⚡ COMANDOS ÚTEIS:" -ForegroundColor White
    Write-Host "   Ctrl+C para parar os serviços" -ForegroundColor Yellow
    Write-Host "   Ou feche esta janela do PowerShell" -ForegroundColor Yellow
    
    Write-Host "`n🔄 LOGS EM TEMPO REAL:" -ForegroundColor White
    Write-Host "Os logs dos serviços aparecerão abaixo..." -ForegroundColor Gray
    Write-Host "-" * 60 -ForegroundColor Gray
    
    # Aguardar indefinidamente ou até Ctrl+C
    Write-Host "⏳ Serviços rodando... Pressione Ctrl+C para parar" -ForegroundColor Green
    
    # Aguardar processos
    while ($backendProcess -and !$backendProcess.HasExited -and $frontendProcess -and !$frontendProcess.HasExited) {
        Start-Sleep -Seconds 2
    }
    
} catch {
    Write-Host "`n❌ Erro ao iniciar serviços: $($_.Exception.Message)" -ForegroundColor Red
} finally {
    # Cleanup
    Write-Host "`n🛑 Parando serviços..." -ForegroundColor Yellow
    
    if ($backendProcess -and !$backendProcess.HasExited) {
        Write-Host "   Parando Backend..." -ForegroundColor Gray
        $backendProcess.Kill()
    }
    
    if ($frontendProcess -and !$frontendProcess.HasExited) {
        Write-Host "   Parando Frontend..." -ForegroundColor Gray
        $frontendProcess.Kill()
    }
    
    Write-Host "✅ Serviços parados!" -ForegroundColor Green
}
