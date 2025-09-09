# 🔧 Melhores Práticas - Variáveis de Ambiente

## 📋 Resumo da Implementação

Este documento descreve as melhores práticas de variáveis de ambiente implementadas no projeto **"Tô De Olho"**, seguindo padrões de segurança, Clean Architecture e configuração centralizada.

## 🏗️ Arquitetura de Configuração

### Configuração Centralizada

- **Arquivo**: `internal/config/config.go`
- **Padrão**: Struct unificada com validação
- **Vantagens**: Type safety, defaults, validação automática

```go
type Config struct {
    Server       ServerConfig
    Database     DatabaseConfig
    Redis        RedisConfig
    CamaraClient CamaraClientConfig
    App          AppConfig
}
```

### Estruturas por Domínio

Cada serviço possui sua própria configuração estruturada:

- **ServerConfig**: Porta, timeouts, rate limiting
- **DatabaseConfig**: Conexão PostgreSQL, pool settings
- **RedisConfig**: Cache, timeouts de rede
- **CamaraClientConfig**: API externa, rate limiting, retries
- **AppConfig**: Environment, logs, versão

## 🛡️ Segurança Implementada

### ✅ Práticas Aplicadas

1. **Não exposição no Docker**
   - ❌ Removido `COPY .env` dos Dockerfiles
   - ✅ Variables definidas via docker-compose.yml
   - ✅ Build-time args para frontend

2. **Validação Obrigatória**
   ```go
   func (c *Config) Validate() error {
       if c.Database.Password == "" {
           return fmt.Errorf("POSTGRES_PASSWORD is required")
       }
       // ... outras validações críticas
   }
   ```

3. **Defaults Seguros**
   - Rate limiting: 100 req/min
   - SSL Mode: disable (desenvolvimento)
   - Timeouts configuráveis
   - Connection pools otimizados

### 🔐 Variáveis Sensíveis

**Obrigatórias** (sem default):
- `POSTGRES_PASSWORD`
- `REDIS_PASSWORD` (produção)

**Opcionais com defaults seguros**:
- Todas as outras variáveis possuem valores padrão apropriados

## 📂 Organização dos Arquivos

### `.env.example` (Template Completo)
```bash
# ═══════════════════════════════════════
# 🚀 CONFIGURAÇÃO TO DE OLHO API
# ═══════════════════════════════════════

# ┌─────────────────────────────────────┐
# │           🌐 SERVIDOR               │
# └─────────────────────────────────────┘
PORT=8080
GIN_MODE=release
RATE_LIMIT_RPS=100
```

### `.env` (Valores Específicos do Ambiente)
- Contém apenas valores diferentes dos defaults
- Nunca incluído no controle de versão
- Ignorado no .gitignore

## 🔄 Padrão de Inicialização

### Fluxo Unificado (main.go)

```go
func main() {
    // 1. Carregar configuração centralizada
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Falha ao carregar configuração: %v", err)
    }

    // 2. Inicializar serviços com config
    pgPool, err := db.NewPostgresPoolFromConfig(ctx, &cfg.Database)
    client := httpclient.NewCamaraClientFromConfig(&cfg.CamaraClient)
    cache := cache.NewFromConfig(&cfg.Redis)
    
    // 3. Usar configuração estruturada
    r.Use(middleware.RateLimitPerIP(cfg.Server.RateLimit, time.Minute))
}
```

### Construtores FromConfig

Cada serviço possui um construtor que recebe configuração tipada:

```go
func NewPostgresPoolFromConfig(ctx context.Context, cfg *config.DatabaseConfig) (*pgxpool.Pool, error)
func NewFromConfig(cfg *config.RedisConfig) *RedisCache
func NewCamaraClientFromConfig(cfg *config.CamaraClientConfig) *CamaraClient
```

## 📊 Benefícios Implementados

### ✅ Type Safety
- Configuração tipada em Go
- Compilação falha se configuração inválida
- IntelliSense e autocomplete

### ✅ Validação Automática
- Validação na inicialização
- Erro claro se configuração inválida
- Defaults apropriados

### ✅ Facilidade de Manutenção
- Configuração centralizada
- Um local para adicionar novas configs
- Documentação integrada

### ✅ Ambiente Específico
- Development: defaults otimizados
- Production: overrides via environment
- Testing: mocks configuráveis

## 🚀 Frontend (Next.js)

### Build-time Variables

```dockerfile
# Dockerfile
ARG NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL

# docker-compose.yml
args:
  NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL:-http://localhost:8080}
```

### Runtime vs Build-time

- **NEXT_PUBLIC_***: Disponível no browser (build-time)
- **Outras**: Apenas server-side (runtime)

## 📋 Checklist de Implementação

### ✅ Backend (Go)

- [x] Configuração centralizada (`config.go`)
- [x] Validação de configs críticas
- [x] Construtores `FromConfig` para todos os serviços
- [x] Remoção de `os.Getenv` direto nos serviços
- [x] Defaults seguros para desenvolvimento
- [x] Rate limiting configurável
- [x] Timeouts configuráveis
- [x] Pool de conexões configurável

### ✅ Frontend (Next.js)

- [x] Build-time environment injection
- [x] Dockerfile com ARG/ENV apropriados
- [x] docker-compose com build args
- [x] Variáveis NEXT_PUBLIC_ documentadas

### ✅ Infrastructure

- [x] docker-compose.yml com variáveis
- [x] .env.example abrangente
- [x] .gitignore atualizado
- [x] Secrets removidos dos Dockerfiles

### ✅ Documentação

- [x] README com setup instructions
- [x] .env.example com comentários
- [x] Este documento de melhores práticas

## 🎯 Próximos Passos

1. **Teste em Produção**: Validar configuração em ambiente real
2. **Secrets Management**: Implementar HashiCorp Vault ou AWS Secrets
3. **Monitoring**: Logs estruturados de configuração
4. **Health Checks**: Validação contínua de configurações

---

**📝 Nota**: Esta implementação segue os padrões do GitHub Copilot Instructions e Clean Architecture, garantindo código maintível, seguro e escalável.
