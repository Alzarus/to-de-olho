# 🚀 Comandos Simples - Tô De Olho

## ⚡ COMANDOS QUE FUNCIONAM (Docker Direto)

### 🐳 Comandos Essenciais

```bash
# Iniciar ambiente de desenvolvimento
docker-compose -f docker-compose.dev.yml up -d

# Ver containers rodando
docker ps

# Ver logs
docker-compose -f docker-compose.dev.yml logs -f

# Parar tudo
docker-compose -f docker-compose.dev.yml down

# Reiniciar tudo
docker-compose -f docker-compose.dev.yml restart
```

### 🌐 URLs (após iniciar)

- **📱 Aplicação**: http://localhost:3000
- **📊 Grafana**: http://localhost:3001 (admin:admin123)  
- **🔥 Prometheus**: http://localhost:9090
- **🐰 RabbitMQ**: http://localhost:15672 (admin:admin123)

### 🔧 Comandos Úteis

```bash
# Acessar PostgreSQL
docker exec -it todo-postgres psql -U postgres

# Acessar Redis
docker exec -it todo-redis redis-cli

# Ver logs específicos
docker logs todo-postgres
docker logs todo-redis

# Limpar containers antigos
docker system prune -f

# Ver estatísticas
docker stats
```

### 🎯 Para Começar

1. **Execute**: `docker-compose -f docker-compose.dev.yml up -d`
2. **Aguarde**: ~30 segundos para tudo iniciar
3. **Acesse**: http://localhost:3001 (Grafana)
4. **Próximo**: Implementar microsserviços

---

**✨ Simples assim! Sem scripts complicados.**
