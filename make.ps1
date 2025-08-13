param(
    [Parameter(Position=0)]
    [string]$Command = "help"
)

function Show-Help {
    Write-Host "Projeto To De Olho - Comandos Disponiveis:" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "DESENVOLVIMENTO LOCAL:" -ForegroundColor White
    Write-Host "  dev              Inicia ambiente local (sem Docker)" -ForegroundColor Green
    Write-Host "  test-api         Testa conexao com API da Camara" -ForegroundColor Green
    Write-Host ""
    Write-Host "DOCKER:" -ForegroundColor White  
    Write-Host "  dev-infra        Inicia infraestrutura (PostgreSQL + Redis)" -ForegroundColor Blue
    Write-Host "  docker-full      Inicia ambiente completo com Docker" -ForegroundColor Blue
    Write-Host "  stop             Para todos os servicos Docker" -ForegroundColor Yellow
    Write-Host "  clean            Remove containers e volumes" -ForegroundColor Red
    Write-Host ""
    Write-Host "UTILITARIOS:" -ForegroundColor White
    Write-Host "  status           Mostra status dos servicos" -ForegroundColor Gray
    Write-Host "  logs             Mostra logs dos containers" -ForegroundColor Gray
    Write-Host "  help             Mostra esta ajuda" -ForegroundColor Gray
}

function Start-Dev {
    Write-Host "Iniciando ambiente de desenvolvimento local..." -ForegroundColor Green
    & .\scripts\start-dev.ps1
}

function Start-DevInfra {
    Write-Host "Iniciando infraestrutura (PostgreSQL + Redis)..." -ForegroundColor Blue
    docker compose up -d postgres redis
    Write-Host "Aguardando servicos ficarem prontos..." -ForegroundColor Yellow
    Start-Sleep -Seconds 10
    Write-Host "Infraestrutura pronta!" -ForegroundColor Green
    Write-Host "PostgreSQL: localhost:5432" -ForegroundColor Cyan
    Write-Host "Redis: localhost:6379" -ForegroundColor Cyan
}

function Start-DockerFull {
    Write-Host "Iniciando ambiente completo com Docker..." -ForegroundColor Blue
    docker compose up -d --build
    Write-Host "Aguardando servicos ficarem prontos..." -ForegroundColor Yellow
    Start-Sleep -Seconds 15
    Write-Host "Ambiente Docker pronto!" -ForegroundColor Green
    Write-Host "Frontend: http://localhost:3000" -ForegroundColor Cyan
    Write-Host "Backend: http://localhost:8080" -ForegroundColor Cyan
}

function Stop-Services {
    Write-Host "Parando servicos..." -ForegroundColor Yellow
    docker compose down
    docker compose -f docker-compose.dev.yml down
    Write-Host "Servicos parados!" -ForegroundColor Green
}

function Clean-All {
    Write-Host "Limpando containers e volumes..." -ForegroundColor Red
    docker compose down -v --remove-orphans
    docker compose -f docker-compose.dev.yml down -v --remove-orphans
    Write-Host "Limpeza concluida!" -ForegroundColor Green
}

function Show-Status {
    Write-Host "Status dos servicos:" -ForegroundColor White
    Write-Host ""
    Write-Host "Containers Docker:" -ForegroundColor Cyan
    docker compose ps
}

function Show-Logs {
    Write-Host "Logs dos containers:" -ForegroundColor White
    docker compose logs --tail=50
}

function Test-API {
    Write-Host "Testando API da Camara dos Deputados..." -ForegroundColor Cyan
    node .\scripts\test-api.js
}

switch ($Command.ToLower()) {
    "help" { Show-Help }
    "dev" { Start-Dev }
    "dev-infra" { Start-DevInfra }
    "docker-full" { Start-DockerFull }
    "stop" { Stop-Services }
    "clean" { Clean-All }
    "status" { Show-Status }
    "logs" { Show-Logs }
    "test-api" { Test-API }
    default {
        Write-Host "Comando '$Command' nao reconhecido!" -ForegroundColor Red
        Write-Host ""
        Show-Help
        exit 1
    }
}
