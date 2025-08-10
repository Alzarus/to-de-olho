# Makefile para automação do projeto "Tô De Olho"

.PHONY: help dev build test clean bootstrap

# Configurações
DOCKER_COMPOSE_DEV = docker-compose -f docker-compose.dev.yml
DOCKER_COMPOSE_PROD = docker-compose -f docker-compose.prod.yml

help: ## Mostra ajuda
	@echo "🏛️ Projeto Tô De Olho - Comandos Disponíveis:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: ## Inicia ambiente de desenvolvimento
	@echo "🚀 Iniciando ambiente de desenvolvimento..."
	$(DOCKER_COMPOSE_DEV) up -d
	@echo "⏳ Aguardando serviços ficarem prontos..."
	@sleep 10
	@echo "✅ Ambiente pronto!"
	@echo "🌐 Aplicação: http://localhost:3000"
	@echo "📊 Grafana: http://localhost:3001"
	@echo "🐰 RabbitMQ: http://localhost:15672"

stop: ## Para todos os serviços
	@echo "🛑 Parando serviços..."
	$(DOCKER_COMPOSE_DEV) down

clean: ## Remove containers e volumes
	@echo "🧹 Limpando containers e volumes..."
	$(DOCKER_COMPOSE_DEV) down -v --remove-orphans
	docker system prune -f

bootstrap: ## Executa bootstrap completo do sistema
	@echo "🏗️ Executando bootstrap..."
	@powershell -ExecutionPolicy Bypass -File scripts/bootstrap.ps1

bootstrap-full: ## Bootstrap com dados completos (4 anos)
	@echo "📊 Bootstrap completo (pode demorar 30+ minutos)..."
	@powershell -ExecutionPolicy Bypass -File scripts/bootstrap.ps1 -FullSync

logs: ## Mostra logs dos serviços
	$(DOCKER_COMPOSE_DEV) logs -f

build-backend: ## Builda todos os microsserviços
	@echo "🔨 Building backend services..."
	@cd backend && go mod tidy
	@cd backend/services/deputados && go build -o bin/deputados ./cmd/server
	@cd backend/services/atividades && go build -o bin/atividades ./cmd/server
	@cd backend/services/despesas && go build -o bin/despesas ./cmd/server
	@cd backend/services/usuarios && go build -o bin/usuarios ./cmd/server
	@cd backend/services/forum && go build -o bin/forum ./cmd/server

build-frontend: ## Builda o frontend Next.js
	@echo "🎨 Building frontend..."
	@cd frontend && npm ci && npm run build

test: ## Executa todos os testes
	@echo "🧪 Executando testes..."
	@cd backend && go test -race -v ./...
	@cd frontend && npm test

test-coverage: ## Executa testes com coverage
	@echo "📊 Executando testes com coverage..."
	@cd backend && go test -race -coverprofile=coverage.out ./...
	@cd backend && go tool cover -html=coverage.out -o coverage.html

lint: ## Executa linting
	@echo "🔍 Executando linting..."
	@cd backend && golangci-lint run
	@cd frontend && npm run lint

format: ## Formata código
	@echo "✨ Formatando código..."
	@cd backend && go fmt ./...
	@cd frontend && npm run format

migrate-up: ## Executa migrações do banco
	@echo "📋 Executando migrações..."
	@go run scripts/migrate.go up

migrate-down: ## Desfaz última migração
	@echo "📋 Desfazendo migração..."
	@go run scripts/migrate.go down

migrate-reset: ## Reset completo do banco
	@echo "🔄 Reset do banco de dados..."
	@go run scripts/migrate.go reset

seed: ## Popula banco com dados demo
	@echo "🎮 Populando dados demo..."
	@go run cmd/seed/demo-data.go

check-health: ## Verifica saúde dos serviços
	@echo "🔍 Verificando saúde dos serviços..."
	@docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

install-tools: ## Instala ferramentas de desenvolvimento
	@echo "🛠️ Instalando ferramentas..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/air-verse/air@latest
	@go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

git-hooks: ## Configura git hooks
	@echo "🪝 Configurando git hooks..."
	@cp scripts/pre-commit .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit

docs: ## Gera documentação da API
	@echo "📚 Gerando documentação..."
	@swag init -g backend/cmd/api-gateway/main.go -o docs/swagger

deploy-dev: ## Deploy para ambiente de desenvolvimento
	@echo "🚀 Deploy para desenvolvimento..."
	$(DOCKER_COMPOSE_PROD) -f docker-compose.staging.yml up -d

deploy-prod: ## Deploy para produção
	@echo "🏭 Deploy para produção..."
	$(DOCKER_COMPOSE_PROD) up -d

backup: ## Backup do banco de dados
	@echo "💾 Fazendo backup..."
	@docker exec todo-postgres pg_dumpall -U postgres > backup_$(shell date +%Y%m%d_%H%M%S).sql

restore: ## Restaura backup (Usage: make restore BACKUP=backup_file.sql)
	@echo "🔄 Restaurando backup..."
	@docker exec -i todo-postgres psql -U postgres < $(BACKUP)

monitoring: ## Abre dashboards de monitoramento
	@echo "📊 Abrindo dashboards..."
	@start http://localhost:3001  # Grafana
	@start http://localhost:9090  # Prometheus
	@start http://localhost:15672 # RabbitMQ

api-test: ## Testa APIs com collection do Postman/Insomnia
	@echo "🔌 Testando APIs..."
	@newman run docs/postman-collection.json

performance-test: ## Executa testes de performance
	@echo "⚡ Testando performance..."
	@artillery run tests/load/load-test.yml
