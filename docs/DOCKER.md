# 🐳 DOCKER - Como Usar

## 🚀 Início Rápido

### Opção 1: Ambiente Local (Recomendado para desenvolvimento)
```bash
# Subir apenas a infraestrutura (PostgreSQL + Redis)
.\make.ps1 dev-infra

# Em seguida, rodar backend e frontend localmente
.\make.ps1 dev
```

### Opção 2: Ambiente Completo com Docker
```bash
# Subir tudo com Docker (Backend + Frontend + Banco + Redis)
.\make.ps1 docker-full
```

### Opção 3: Ambiente de Desenvolvimento Completo
```bash
# Inclui monitoramento (Grafana, Prometheus, RabbitMQ)
docker compose -f docker-compose.dev.yml up -d
```

## 📋 Comandos Disponíveis

| Comando | Descrição |
|---------|-----------|
| `.\make.ps1 help` | Mostra todos os comandos |
| `.\make.ps1 dev-infra` | Sobe PostgreSQL + Redis |
| `.\make.ps1 docker-full` | Sobe ambiente completo com Docker |
| `.\make.ps1 status` | Mostra status dos containers |
| `.\make.ps1 logs` | Mostra logs dos containers |
| `.\make.ps1 stop` | Para todos os containers |
| `.\make.ps1 clean` | Remove containers e volumes |

## 🔧 Configuração Manual

### 1. Verificar Docker
```bash
docker --version
docker compose --version
```

### 2. Subir Infraestrutura
```bash
docker compose up -d postgres redis
```

### 3. Verificar se está funcionando
```bash
docker compose ps
docker compose logs postgres
```

### 4. Testar conexão com banco
```bash
docker exec -it todo-postgres psql -U postgres -d todo_dev -c "SELECT version();"
```

## 🌐 URLs de Acesso

### Ambiente Local
- **Frontend**: http://localhost:3000
- **Backend**: http://localhost:8080
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379

### Ambiente Completo de Desenvolvimento
- **Grafana**: http://localhost:3001 (admin/admin123)
- **Prometheus**: http://localhost:9090
- **RabbitMQ**: http://localhost:15672 (admin/admin123)

## 🗃️ Dados do Banco

### Configurações Padrão
- **Host**: localhost
- **Porta**: 5432
- **Banco**: todo_dev
- **Usuário**: postgres
- **Senha**: postgres

### Conectar via linha de comando
```bash
docker exec -it todo-postgres psql -U postgres -d todo_dev
```

## 🔄 Volumes Persistentes

Os dados são mantidos em volumes Docker:
- `postgres_data` - Dados do PostgreSQL
- `redis_data` - Dados do Redis
- `grafana_data` - Configurações do Grafana
- `prometheus_data` - Métricas do Prometheus

Para limpar todos os dados:
```bash
.\make.ps1 clean
```

## 🚨 Troubleshooting

### Docker Desktop não está rodando
```bash
# Verificar se está rodando
docker info

# Se não estiver, inicie o Docker Desktop manualmente
```

### Porta já está em uso
```bash
# Verificar quais portas estão em uso
netstat -ano | findstr :3000
netstat -ano | findstr :8080
netstat -ano | findstr :5432

# Parar processos se necessário
taskkill /PID <PID> /F
```

### Limpar tudo e recomeçar
```bash
.\make.ps1 clean
docker system prune -a --volumes
.\make.ps1 dev-infra
```

## 📊 Monitoramento

### Logs em tempo real
```bash
docker compose logs -f
docker compose logs -f postgres
docker compose logs -f backend
```

### Métricas dos containers
```bash
docker stats
```

### Health checks
```bash
docker compose ps
```

## 🎯 Dicas de Performance

1. **Use dev-infra** para desenvolvimento local (mais rápido)
2. **Use docker-full** apenas para testar o ambiente completo
3. **Monitore recursos** com `docker stats`
4. **Limpe volumes** regularmente com `.\make.ps1 clean`

---

✅ **O ambiente Docker está pronto para uso!** 🚀
