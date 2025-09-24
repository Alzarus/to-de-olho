# 📊 Plano de Melhoria da Cobertura de Testes - Backend

## 🎯 Objetivo: Elevar cobertura geral para 70%

**Status Atual Analisado (18/09/2025):**

## 📈 Situação Atual por Módulo

| Módulo | Cobertura Atual | Meta 70% | Gap | Prioridade |
|--------|----------------|----------|-----|------------|
| **application** | 91.2% | ✅ 70% | +21.2% | ✅ Mantido |
| **config** | 96.3% | ✅ 70% | +26.3% | ✅ Mantido |
| **domain** | 69.9% | ⚠️ 70% | -0.1% | 🔥 CRÍTICO |
| **infrastructure/background** | 0.0% | ❌ 70% | -70% | 🔥 URGENTE |
| **infrastructure/cache** | 77.9% | ✅ 70% | +7.9% | ✅ Mantido |
| **infrastructure/db** | 89.2% | ✅ 70% | +19.2% | ✅ Mantido |
| **infrastructure/httpclient** | 86.5% | ✅ 70% | +16.5% | ✅ Mantido |
| **infrastructure/ingestor** | 12.1% | ❌ 70% | -57.9% | 🔥 URGENTE |
| **infrastructure/migrations** | 25.0% | ❌ 70% | -45% | 🔥 ALTA |
| **infrastructure/repository** | 46.1% | ❌ 70% | -23.9% | 🔥 ALTA |
| **infrastructure/resilience** | 85.2% | ✅ 70% | +15.2% | ✅ Mantido |
| **interfaces/http** | 49.1% | ❌ 70% | -20.9% | 🔥 ALTA |
| **interfaces/http/middleware** | 36.1% | ❌ 70% | -33.9% | 🔥 ALTA |

---

## 🚀 Estratégia de Execução

### Fase 1: Emergência (Semana 1-2) - Módulos 0-25%
**Objetivo: Subir módulos críticos para pelo menos 50%**

#### 🔥 **infrastructure/background** (0.0% → 70%)
**Análise:** Módulo sem nenhum teste - URGENTE

**Ações:**
- [ ] **Criar estrutura básica de testes**
  - Setup de testes unitários para `handlers.go` e `processor.go`
  - Mocking das dependencies (DB, message queues)
  
- [ ] **Testes prioritários (35% cobertura):**
  ```go
  // handlers_test.go
  func TestBackgroundHandlers_ProcessDeputado(t *testing.T)
  func TestBackgroundHandlers_ProcessProposicao(t *testing.T)
  func TestBackgroundHandlers_HandleErrors(t *testing.T)
  
  // processor_test.go  
  func TestProcessor_Start(t *testing.T)
  func TestProcessor_Stop(t *testing.T)
  func TestProcessor_ProcessMessage(t *testing.T)
  ```

- [ ] **Testes complementares (70% cobertura):**
  - Error handling scenarios
  - Retry logic 
  - Circuit breaker integration
  - Message validation

**Estimativa:** 16 horas | **Responsável:** Dev Backend Sr.

---

#### 🔥 **infrastructure/ingestor** (12.1% → 70%)
**Análise:** Módulo crítico com lógica complexa de ETL

**Ações:**
- [ ] **Priorizar `BackfillManager` (30% cobertura):**
  ```go
  // backfill_manager_test.go (melhorar)
  func TestBackfillManager_CreateCheckpoint(t *testing.T)
  func TestBackfillManager_UpdateProgress(t *testing.T)
  func TestBackfillManager_ResumeFromCheckpoint(t *testing.T)
  func TestBackfillManager_HandleFailures(t *testing.T)
  ```

- [ ] **Melhorar `incremental_sync.go` (25% cobertura):**
  ```go
  // incremental_sync_test.go (expandir)
  func TestIncrementalSync_SyncDeputados(t *testing.T)
  func TestIncrementalSync_SyncProposicoes(t *testing.T)
  func TestIncrementalSync_DeltaDetection(t *testing.T)
  func TestIncrementalSync_ConflictResolution(t *testing.T)
  ```

- [ ] **Adicionar `strategic_backfill.go` (15% cobertura):**
  ```go
  // strategic_backfill_test.go (expandir)
  func TestStrategicBackfill_PrioritizeData(t *testing.T)
  func TestStrategicBackfill_BatchProcessing(t *testing.T)
  func TestStrategicBackfill_PerformanceOptimization(t *testing.T)
  ```

**Estimativa:** 24 horas | **Responsável:** Dev Backend Sr. + Dev Mid

---

#### 🔥 **infrastructure/migrations** (25.0% → 70%)
**Análise:** Módulo fundamental para DB, precisa ser confiável

**Ações:**
- [ ] **Implementar técnica Interface Wrapper (como documentado no testing-guide.md):**
  ```go
  // migrator_mockable.go
  type MigratorWithMocks struct {
      *Migrator
      CreateMigrationsTableFunc func(ctx context.Context) error
      GetAppliedMigrationsFunc  func(ctx context.Context) (map[int]bool, error)
      ApplyMigrationFunc        func(ctx context.Context, migration Migration) error
  }
  ```

- [ ] **Expandir testes existentes:**
  ```go
  // migrator_test.go (melhorar de 25% → 70%)
  func TestMigrator_Run_Success(t *testing.T)
  func TestMigrator_Run_PartialFailure(t *testing.T)
  func TestMigrator_Run_Recovery(t *testing.T)
  func TestMigrator_GetAppliedMigrations_Error(t *testing.T)
  func TestMigrator_ApplyMigration_Rollback(t *testing.T)
  ```

**Meta:** 25.0% → 70% (+45%) ✅
**Estimativa:** 12 horas | **Responsável:** Dev Backend Mid

---

### Fase 2: Consolidação (Semana 3-4) - Módulos 25-50%

#### 🔥 **infrastructure/repository** (46.1% → 70%)
**Análise:** Repository layer é crítico para data integrity

**Ações:**
- [ ] **Melhorar testes existentes:**
  - `deputado_repository_test.go` - cenários de erro
  - `proposicao_repository_test.go` - edge cases
  - `despesa_repository.go` - novo arquivo sem testes

- [ ] **Adicionar testes de integration:**
  ```go
  // integration_repository_test.go
  func TestRepository_TransactionHandling(t *testing.T)
  func TestRepository_ConcurrencyControl(t *testing.T)  
  func TestRepository_ConnectionPool(t *testing.T)
  ```

- [ ] **Testes de performance:**
  ```go
  // performance_benchmark_test.go (expandir)
  func BenchmarkRepository_BulkInsert(b *testing.B)
  func BenchmarkRepository_ComplexQueries(b *testing.B)
  func TestRepository_QueryOptimization(t *testing.T)
  ```

**Estimativa:** 16 horas | **Responsável:** Dev Backend Sr.

---

#### 🔥 **interfaces/http** (49.1% → 70%)
**Análise:** HTTP layer precisa cobrir todos os cenários de entrada

**Ações:**
- [ ] **Expandir testes de controllers:**
  ```go
  // deputados_controller_test.go
  func TestDeputadosController_GetAll_Pagination(t *testing.T)
  func TestDeputadosController_GetByID_NotFound(t *testing.T)
  func TestDeputadosController_Create_ValidationErrors(t *testing.T)
  func TestDeputadosController_Update_Concurrency(t *testing.T)
  ```

- [ ] **Adicionar testes de error handling:**
  ```go
  // error_handling_test.go
  func TestHTTP_ValidationErrors(t *testing.T)
  func TestHTTP_InternalServerErrors(t *testing.T)
  func TestHTTP_TimeoutHandling(t *testing.T)
  ```

**Estimativa:** 14 horas | **Responsável:** Dev Backend Mid

---

#### 🔥 **interfaces/http/middleware** (36.1% → 70%)
**Análise:** Middleware é camada de segurança crítica

**Ações:**
- [ ] **Testes de autenticação:**
  ```go
  // auth_middleware_test.go
  func TestAuthMiddleware_ValidToken(t *testing.T)
  func TestAuthMiddleware_ExpiredToken(t *testing.T)
  func TestAuthMiddleware_InvalidToken(t *testing.T)
  func TestAuthMiddleware_MissingToken(t *testing.T)
  ```

- [ ] **Testes de rate limiting:**
  ```go
  // rate_limit_middleware_test.go
  func TestRateLimit_WithinLimit(t *testing.T)
  func TestRateLimit_ExceedsLimit(t *testing.T)
  func TestRateLimit_DifferentIPs(t *testing.T)
  ```

- [ ] **Testes de CORS e security:**
  ```go
  // security_middleware_test.go
  func TestCORS_Middleware(t *testing.T)
  func TestSecurity_Headers(t *testing.T)
  func TestCSRF_Protection(t *testing.T)
  ```

**Estimativa:** 12 horas | **Responsável:** Dev Backend Mid

---

### Fase 3: Refinamento (Semana 5) - Módulos próximos de 70%

#### ⚠️ **domain** (69.9% → 75%)
**Análise:** Já muito próximo da meta, pequenos ajustes

**Ações:**
- [ ] **Identificar funções não cobertas:**
  ```bash
  go test -coverprofile=domain_coverage.out ./internal/domain
  go tool cover -html=domain_coverage.out
  ```

- [ ] **Adicionar testes para edge cases:**
  - Validações de Value Objects
  - Regras de negócio complexas
  - Error scenarios específicos

**Estimativa:** 4 horas | **Responsável:** Qualquer dev

---

## 🛠️ Ferramentas e Scripts de Apoio

### Script de Monitoramento Contínuo
```bash
#!/bin/bash
# scripts/coverage-monitor.sh

echo "🔍 Monitoramento de Coverage - Target: 70%"

modules=(
    "internal/application"
    "internal/config" 
    "internal/domain"
    "internal/infrastructure/background"
    "internal/infrastructure/cache"
    "internal/infrastructure/db"
    "internal/infrastructure/httpclient"
    "internal/infrastructure/ingestor"
    "internal/infrastructure/migrations"
    "internal/infrastructure/repository"
    "internal/infrastructure/resilience"
    "internal/interfaces/http"
    "internal/interfaces/http/middleware"
)

for module in "${modules[@]}"; do
    coverage=$(go test -cover ./$module 2>/dev/null | grep "coverage:" | awk '{print $4}' | sed 's/%//')
    
    if (( $(echo "$coverage >= 70" | bc -l) )); then
        echo "✅ $module: $coverage%"
    elif (( $(echo "$coverage >= 50" | bc -l) )); then
        echo "⚠️  $module: $coverage%"
    else
        echo "❌ $module: $coverage%"
    fi
done
```

### Makefile para Automação
```makefile
# Makefile - adicionar ao projeto

.PHONY: test-coverage-critical test-coverage-target

# Testar apenas módulos críticos (< 50%)
test-coverage-critical:
	@echo "🔥 Testando módulos críticos..."
	go test -cover ./internal/infrastructure/background
	go test -cover ./internal/infrastructure/ingestor  
	go test -cover ./internal/infrastructure/migrations
	go test -cover ./internal/infrastructure/repository
	go test -cover ./internal/interfaces/http
	go test -cover ./internal/interfaces/http/middleware

# Verificar se meta 70% foi atingida
test-coverage-target:
	@./scripts/coverage-monitor.sh
	@echo ""
	@echo "🎯 Meta: 70% de cobertura geral"
```

---

## 📊 Métricas e Acompanhamento

### Definition of Done para Cobertura
- [ ] **Módulo atingiu 70% de cobertura**
- [ ] **Testes passam em CI/CD**
- [ ] **Code review aprovado**
- [ ] **Documentação atualizada**
- [ ] **Performance benchmarks OK**

### Timeline de Execução

| Semana | Fase | Módulos | Meta Coverage | Esforço |
|--------|------|---------|---------------|---------|
| **1-2** | Emergência | background, ingestor, migrations | 0→50% | 52h |
| **3-4** | Consolidação | repository, http, middleware | 30→70% | 42h |
| **5** | Refinamento | domain, adjustments | 69→75% | 8h |
| **Total** | - | **Todos** | **70%+** | **102h** |

### Estimativa de Recursos
- **Dev Backend Sr.**: 56 horas (55%)
- **Dev Backend Mid**: 42 horas (41%) 
- **Any Dev**: 4 horas (4%)

---

## 🎯 Resultado Esperado

### Cobertura Final Projetada (Conservadora)

| Módulo | Atual | Meta | Projetado |
|--------|-------|------|-----------|
| application | 91.2% | 91% | 91.2% |
| config | 96.3% | 96% | 96.3% |
| domain | 69.9% | 75% | 75.0% |
| background | 0.0% | 70% | 72.0% |
| cache | 77.9% | 77% | 77.9% |
| db | 89.2% | 89% | 89.2% |
| httpclient | 86.5% | 86% | 86.5% |
| ingestor | 12.1% | 70% | 71.0% |
| migrations | 25.0% | 70% | 73.0% |
| repository | 46.1% | 70% | 72.0% |
| resilience | 85.2% | 85% | 85.2% |
| http | 49.1% | 70% | 71.0% |
| middleware | 36.1% | 70% | 72.0% |

**🎯 Cobertura Geral Projetada: ~76.8%** *(Meta: 70%)*

---

## 🚨 Riscos e Mitigações

### Riscos Identificados
1. **Complexidade do módulo ingestor** - ETL com muitas dependencies
   - *Mitigação:* Usar interface wrappers e mocking avançado

2. **Testes de background jobs** - Assíncronos e complexos
   - *Mitigação:* Testcontainers para ambiente isolado

3. **Performance dos testes** - Pode ficar lento
   - *Mitigação:* Paralelização e otimização de setup

### Pontos de Atenção
- **Não sacrificar qualidade por speed** - Testes mal escritos são piores que sem testes
- **Manter balance de testing pyramid** - Focar em unit tests (80%)
- **Documentation as you go** - Documentar patterns de teste para o time

---

> **✅ Com este plano, projetamos atingir 76.8% de cobertura geral em 5 semanas, superando a meta de 70%**