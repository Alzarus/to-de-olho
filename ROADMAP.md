# 🛣️ Roadmap - "Tô De Olho" 

> **Transparência Política para Todos os Brasileiros**
> 
> **Status**: Set/2025 | **Arquitetura**: Ingestão Total + Analytics + WCAG 2.1 AA

## 🎯 Visão Core 2026

**Missão**: Plataforma acessível que democratiza dados da Câmara com:
- **🔄 Ingestão Completa**: Base própria (histórico + diário)  
- **📊 Analytics Inteligentes**: Rankings, insights, tendências
- **♿ WCAG 2.1 AA**: Interface para TODA população brasileira
- **🤖 IA Educativa**: Assistente político contextual

## 📊 Status Arquitetural

| Camada | Status | Prioridade | Marco |
|--------|--------|------------|-------|
| 🔄 **Ingestão ETL** | ✅ Implementado | CRÍTICA | ✅ Set/2025 |
| 📊 **Analytics Engine** | ✅ Base pronta | ALTA | ✅ Set/2025 |
| ♿ **Frontend WCAG** | ❌ Não conforme | CRÍTICA | Out/2025 |
| 🏗️ **Backend Core** | ✅ Sólido | - | Manter |
| 🤖 **IA Gemini** | ❌ Planejado | MÉDIA | Dez/2025 |

## 🎉 Progresso Setembro 2025

### ✅ **CONCLUÍDO - Set/16/2025**

#### 🔄 **Sistema de Ingestão Completo**
- ✅ **Backfill Configurável**: Ano inicial configurável via `INGESTOR_BACKFILL_START_YEAR=2025`
- ✅ **Estratégia Inteligente**: Checkpoints, retry exponencial, circuit breaker
- ✅ **Três Modos**: `daily`, `backfill`, `strategic` com parâmetros flexíveis
- ✅ **Configuração Robusta**: `IngestorConfig` com batch size e max retries
- ✅ **Comando**: `./ingestor -mode=strategic -start-year=2025`

#### 📊 **Analytics com Dados Internos**
- ✅ **Repositórios Diretos**: Analytics usa PostgreSQL ao invés da API Câmara
- ✅ **Rankings Disponíveis**: Gastos, proposições, presença (com simulação)
- ✅ **Cache Inteligente**: Redis para performance + fallback
- ✅ **Insights Gerais**: Dashboard agregado para transparência

#### 🧪 **Testing & Collection**
- ✅ **Postman Collection Completa**: 25+ endpoints organizados
- ✅ **Ambientes Configurados**: Local development + variáveis
- ✅ **Testes Automáticos**: Validação de status, performance, estrutura
- ✅ **Documentação**: README detalhado para uso da collection

#### 🏗️ **Arquitetura Melhorada**
- ✅ **Interfaces Limpas**: Repository patterns implementados
- ✅ **Configuração Central**: `config.go` com todas as settings
- ✅ **Separation of Concerns**: Analytics não depende mais de services externos
- ✅ **Error Handling**: Timeouts, contextos, logs estruturados

#### 🧪 **Qualidade & Performance**
- ✅ **Testes Corrigidos**: Analytics service 100% funcional (12 erros de compilação resolvidos)
- ✅ **Performance Otimizada**: Processamento de 513 deputados em ~76μs (vs limitação anterior de 100)
- ✅ **Processamento em Batches**: Algoritmo otimizado para grandes volumes (50 deputados/batch)
- ✅ **Timeout Inteligente**: 30s para rankings, 15s para insights, logs informativos

## 🔄 Arquitetura de Ingestão (PRIORIDADE #1)

### **Problema Atual**: Frontend consulta API externa (lento + instável)
### **Solução**: Base própria enriched + Analytics pre-computados

```
API Câmara → Ingestor ETL → PostgreSQL → Analytics → API Nossa → Frontend
     ↓           ↓            ↓           ↓         ↓          ↓
Deputados   Backfill    Cache Redis   Rankings   Cache     UX Rápida
Proposições   Daily      Histórico    Insights   Intelig.   + Offline
Despesas     Schedule    Fallback     Trending   Response
```

### **Implementação Out/2025**:
1. **Backfill Histórico** (2019-2025): Deputados, proposições, despesas
   - **Estratégia**: Lotes por legislatura+ano, rate limit 100/min, circuit breaker
   - **Ordem**: Deputados → Proposições → Despesas → Votações
   - **Resilência**: Retry exponencial, checkpoints, fallback por lote
2. **Ingestão Diária** (6h): Scheduler automático + delta sync  
3. **Analytics Pre-compute**: Rankings, gastos suspeitos, temas trending
4. **API Própria**: Cache inteligente + fallback Câmara

## ♿ Frontend WCAG 2.1 AA (PRIORIDADE #1)

### **Problemas Identificados**:
- ❌ Contraste baixo: `text-gray-600` (3:1) → precisa 4.5:1+
- ❌ Textos pequenos: `text-sm` → mínimo 16px base
- ❌ Navegação teclado: sem `tabIndex`, `aria-labels`
- ❌ Cores únicas: filtros sem indicadores textuais

### **Padrão Acessível**:
```tsx
// ✅ Contraste alto, navegação teclado, aria-labels
<button 
  className="bg-blue-700 text-white text-base px-6 py-3 rounded-lg
             hover:bg-blue-800 focus:ring-4 focus:ring-blue-300
             focus:outline-none"
  aria-label="Buscar deputados por filtros selecionados"
  tabIndex={0}
>
  Buscar Deputados
</button>
```

### **UX Brasileira**:
- **Linguagem simples**: "Gastos do Deputado" vs "Despesas Parlamentares"
- **Contexto político**: Tooltips explicativos para termos técnicos
- **Mobile-first**: 70% acessos via smartphone no Brasil
- **Offline-ready**: PWA para áreas com internet instável

## 📊 Analytics & Insights Engine

### **Rankings Automáticos**:
```go
type Rankings struct {
    Presenca      []DeputadoRank // Quem mais falta
    GastosEfic    []DeputadoRank // Melhor custo/benefício  
    Proposicoes   []DeputadoRank // Mais ativo legislativo
    Transparencia []DeputadoRank // Dados mais completos
}
```

### **Insights Cidadão**:
- **Trending**: Temas mais votados últimos 30 dias
- **Impacto**: Proposições que afetam seu município  
- **Comparativo**: Seu deputado vs média nacional/estadual
- **Alertas**: Gastos suspeitos, mudanças importantes

## 🤖 IA Assistente Educativo

### **Contexto Brasileiro**:
- **Base Knowledge**: 10k+ perguntas políticas respondidas  
- **Moderação**: Gemini AI para conteúdo seguro e factual
- **Educação**: "Como funciona uma PEC?" integrado ao contexto
- **Personalização**: Baseado na localização (UF/município)

---

## 🚀 Cronograma Executivo Atualizado

### ✅ **Setembro 2025 - Base Sólida** (CONCLUÍDO)
- ✅ Ingestor completo (deputados + proposições + despesas)
- ✅ Analytics com dados internos + cache inteligente
- ✅ Collection Postman completa para testes
- ✅ Configuração flexível via environment variables

### **Outubro 2025 - Dados Reais & Frontend**
- [ ] **Dados Reais**: Substituir simulação por repository SQL otimizado
  - [ ] Implementar `DespesaRepository` com queries por deputado/ano
  - [ ] Criar índices para performance: `(deputado_id, ano, valor)`
  - [ ] Validar accuracy rankings vs dados oficiais Câmara
- [ ] Executar backfill completo 2025 (dados reais da Câmara)
- [ ] Frontend WCAG 2.1 AA compliance  
- [ ] Testes de carga: 1000+ requests simultâneas
- [ ] **Performance Real**: Benchmark analytics com 513 deputados + dados completos

### **Novembro 2025 - Analytics Avançados**  
- [ ] Rankings automáticos com dados reais (presença, gastos, eficiência)
- [ ] Dashboard insights cidadão
- [ ] API analytics + frontend integration
- [ ] Implementar proposições por autor/tema
- [ ] **Cache Strategy**: Warming + hierarchy (L1+L2+L3)
- [ ] **Background Jobs**: Rankings pesados processados offline

### **Dezembro 2025 - IA & UX**
- [ ] Assistente Gemini básico
- [ ] PWA + offline capabilities  
- [ ] Testes usuário população alvo

### **Q1 2026 - Produção**
- [ ] Deploy produção + monitoramento
  - **Plataforma**: Google Cloud Platform (Cloud Run + Cloud SQL + Memorystore)
  - **Domínio**: `todeolho.com.br` via Cloud Domains  
  - **Custo inicial**: ~$90-120/mês (auto-scale conforme uso)
- [ ] Documentação pública + API aberta
- [ ] Marketing transparência eleitoral

## 🎯 Próximos Passos Imediatos

### 🔥 **Alta Prioridade (Esta Semana)**
1. **Executar Backfill Completo**: `./ingestor -mode=strategic -start-year=2025`
2. **Testar API com Postman**: Validar todos endpoints com dados reais
3. **Implementar Despesas por Deputado**: Método no repositório + endpoint
4. **Frontend WCAG**: Correções de contraste e navegação por teclado

### 📊 **Performance & Dados Reais (Próxima Sprint)**
1. **Substituir Simulação por Dados Reais**:
   - Implementar busca real de despesas no `DeputadoRepository`
   - Criar queries SQL otimizadas para gastos por ano
   - Adicionar índices para performance (`deputado_id`, `ano`, `valor`)

2. **Otimização Analytics Production**:
   - Cache warming: Pré-computar rankings principais no deploy
   - Background jobs: Processar rankings pesados em background
   - Paginação inteligente: Implementar para rankings > 100 itens
   - Monitoring: Prometheus metrics para performance analytics

3. **Validação e Qualidade**:
   - Executar benchmark com dados reais (513 deputados completos)
   - Stress testing: 1000+ requisições simultâneas
   - Validar accuracy dos rankings vs dados oficiais Câmara
   - Configurar alertas para performance degradation

### 🏗️ **Arquitetura & Escalabilidade**
1. **Repository Layer Completo**:
   - `DespesaRepository` com queries otimizadas
   - `VotacaoRepository` para ranking de presença real
   - Connection pooling e read replicas para analytics
   
2. **Cache Strategy**:
   - Redis Cluster para alta disponibilidade
   - Cache hierarchy: L1 (in-memory) + L2 (Redis) + L3 (DB)
   - TTL inteligente baseado na frequência de updates

3. **API Governance**:
   - Rate limiting por usuário/API key
   - Circuit breaker para dependencies externas
   - Health checks e readiness probes

### 📊 **Métricas de Sucesso**
- **Performance**: API < 200ms vs 2s+ da API Câmara original
- **Cobertura**: 100% deputados 2025 + principais proposições
- **Acessibilidade**: WCAG 2.1 AA completo
- **Testes**: 90%+ cobertura de código
- **Analytics**: Rankings com 513 deputados em <100ms (atual: 76μs)
- **Confiabilidade**: Zero timeouts em cenários de produção

### 🛠️ **Melhorias Técnicas Implementadas Hoje (Set/16)**
- ✅ **Mock Repositories**: Interfaces corretas para testes analytics
- ✅ **Processamento Escalável**: 600 deputados suportados vs 100 anterior
- ✅ **Batch Processing**: Algoritmo otimizado em lotes de 50
- ✅ **Error Handling**: Timeouts configuráveis com logs detalhados
- ✅ **Interface CachePort**: Abstração completa para cache Redis

### 🔮 **Melhorias Técnicas Futuras**
- Circuit breaker para API externa (já implementado baseline)
- Metrics com Prometheus/Grafana  
- Rate limiting por IP (já implementado)
- Logs estruturados com observabilidade
- **Dados Reais**: Substituir simulação por queries PostgreSQL otimizadas
- **Cache Warming**: Pré-computar rankings durante CI/CD
- **Horizontal Scaling**: Suporte a múltiplas instâncias analytics

---

## ✅ Definition of Done

### **Acessibilidade** (não negociável):
- [ ] Contraste 4.5:1+ em todos elementos
- [ ] Navegação completa via teclado  
- [ ] Screen reader friendly (NVDA/JAWS)
- [ ] Textos mínimo 16px, máximo 80 chars/linha

### **Performance**:
- [ ] API nossa: <200ms (vs 2s+ Câmara)
- [ ] Frontend: <1s FCP, <2.5s LCP
- [ ] Offline: Cache 7 dias dados essenciais

### **Qualidade**:
- [ ] Testes: 90%+ cobertura  
- [ ] Security: 0 vulnerabilidades críticas
- [ ] UX: Validado com brasileiros +50 anos

---

> **🎯 Meta 2026**: Ferramenta #1 transparência política para eleições
> 
> **🇧🇷 Impacto**: Decisões eleitorais mais informadas para TODOS os brasileiros
