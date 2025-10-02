# Roadmap - "Tô De Olho"

> Transparência política para todos os brasileiros.
>
> Status consolidado em 02/out/2025.

## Prioridades Gerais

Missão: concluir, validar e preparar para produção todos os componentes de ingestão, analytics e experiência do usuário da plataforma.

## Status Atual

| Funcionalidade                | Situação atual                    | Prioridade | Deadline     |
|------------------------------|----------------------------------|------------|--------------|
| Sistema de votações          | Concluído                         | Baixa      | set/2025     |
| Sincronização + API Câmara   | Em ajuste (ingestão de despesas)  | Crítica    | out/2025     |
| Engine de analytics          | Concluído, aguardando dados reais | Média      | set/2025     |
| Frontend WCAG                | Concluído                         | Média      | set/2025     |
| API REST v1                  | Concluído                         | Média      | set/2025     |
| Esquema do banco             | Em ajuste (migration 014)         | Crítica    | out/2025     |
| Deploy em produção           | Não iniciado                      | Alta       | nov/2025     |
| Integração IA Gemini         | Não iniciado                      | Média      | dez/2025     |

## Demandas Urgentes

- Revisar componentes de interface que dificultam a filtragem de deputados (exemplo: seletor de partido).
- Implementar exibição de votações no frontend principal.

## Backfill Histórico (API Câmara)

> Objetivo: garantir backfill idempotente, confiável e observável cobrindo todas as entidades do `api-docs.json`, permitindo carga inicial completa e sincronizações incrementais diárias.

### Resumo do estado atual
- Concluído: Deputados, Proposições, Despesas, Votações (checkpoints + executor via `VotacoesService.SincronizarVotacoes`), Partidos (upsert + checkpoint dedicado).
- Em andamento: testes unitários do executor de votações, validação de performance em staging, cobertura de repositórios sem integração automatizada.
- Pontos de atenção: sub-recursos de deputados (discursos, eventos, histórico, etc.), filtros avançados de proposições (arrays, `codTema`, `autor`), suporte a IDs alfanuméricos de votações.
- Próximos alvos (prioridade média): Órgãos, Legislaturas, Referências.
- Backlog (prioridade baixa): Eventos, Blocos, Frentes, Grupos.

### Estratégia operacional
- Backfill inicial até **yesterday** (configurável) para evitar dados em trânsito
- Reprocessar diariamente o dia anterior (overlap de 1 dia) para capturar alterações tardias
- Utilizar consistentemente **upsert + checkpoints por entidade/ano** para idempotência

### Checkpoints sugeridos (prioridade)
1. Deputados
2. Proposições — checkpoints por ano
3. Despesas — checkpoints por ano
4. Votações — checkpoints anuais ou por período; reutilizar upsert existente
5. Partidos / Órgãos / Legislaturas / Referências
6. Eventos / Blocos / Frentes / Grupos

### Tarefas concretas

**Votações (alta prioridade)**
- [x] Checkpoint "votacoes" no plano anual (`StrategicBackfillExecutor.createBackfillPlan`)
- [x] Executor integrado ao `VotacoesService` (`executeVotacoesBackfill`)
- [x] Janela anual com `SincronizarVotacoes` (upsert + votos/orientações)
- [x] Testes de integração no `VotacaoRepository`
- [ ] Ajustar domínio/repos para IDs alfanuméricos (persistir `id` string, manter `IDVotacaoCamara` opcional)
- [ ] Revisar `CamaraClient` para filtros oficiais (`idProposicao`, `idEvento`, `idOrgao`, datas no mesmo ano) e paginação (≤200 itens)
- [ ] Testes unitários/mocks do executor e regressões de checkpoint
- [ ] Backfill completo em staging (performance/governança)

**Partidos (prioridade média)**
- [x] Domínio + migration `012_create_partidos_table.sql`
- [x] `CamaraClient.FetchPartidos` + `PartidosService.ListarPartidos` com upsert
- [x] Checkpoint e executor dedicados
- [ ] Testes unit/integration para service e repository
- [ ] Execução validada em staging com monitoramento de consistência

**Proposições (adequação à spec)**
- [ ] Serializar listas (`siglaTipo`, `numero`, `ano`, `codTema`, `keywords`) segundo `style=form&explode=false`
- [ ] Corrigir parâmetros de autor (`autor="nome"`, `idDeputadoAutor`, `siglaPartidoAutor`, `siglaUfAutor`) e remover campos inexistentes na API
- [ ] Ingerir/backfilar sub-recursos (`/tramitacoes`, `/autores`, `/votacoes`, `/temas`) e persistir
- [ ] Cobrir mudanças com testes table-driven e atualizar caches/repos

**Órgãos / Legislaturas / Referências (prioridade média)**
- [ ] Modelagem de domínio + migrations
- [ ] Clients + repositórios com upsert
- [ ] Checkpoints e executores específicos
- [ ] Testes e validação

**Eventos / Blocos / Frentes / Grupos (prioridade baixa)**
- [ ] Mesma abordagem (model + migration + upsert + executor)
- [ ] Avaliar particionamento/processamento por período para grandes volumes

**Observabilidade e operação**
- [ ] Padronizar logs estruturados por checkpoint (substituir `log.Printf` por `slog`)
- [ ] Exportar métricas Prometheus (usar `pkg/metrics`)
- [ ] Dashboards Grafana + alertas

**QA / Release**
- [ ] Cobertura ≥80% (unit + integration) — faltam cenários para executor e partidos
- [ ] Validação com dataset real em staging
- [ ] Planejamento de janelas de execução (backfill inicial custoso)

**Próximos passos imediatos**
1. Aplicar a migration `014_alter_despesas_add_columns.sql`, implantar o `DespesaRepository` atualizado e reprocessar o backfill de despesas.
2. Executar testes unitários do executor de votações e validar desempenho em ambiente de staging.
3. Desenvolver a ingestão para Órgãos, Legislaturas e Referências (domínio, clients, checkpoints, testes).
4. Criar testes table-driven adicionais para `PartidosService` e `PartidoRepository`.

### 1. Deploy GCP (crítico - nov/2025)
**Objetivo**: Colocar plataforma no ar para uso público

**Necessário Implementar**:
- Cloud Run containers (backend)
- Cloud SQL PostgreSQL (dados)
- Memorystore Redis (cache)  
- Load Balancer + SSL
- Domínio `todeolho.com.br`

**Configurações**:
```yaml
# docker-compose.prod.yml
services:
  backend:
    image: gcr.io/todeolho/backend:latest
    environment:
      - POSTGRES_HOST=10.x.x.x
      - REDIS_ADDR=10.x.x.x:6379
```

### 2. Expansão de analytics (alta - nov/2025)
**Objetivo**: Ampliar funcionalidades de análise baseadas na API da Câmara

**Funcionalidades Prioritárias**:
- **� Analytics de Votações**: Rankings e estatísticas agregadas (DESCOBERTO - Set/24/2025)
- **�🗣️ Central de Discursos**: Análise de pronunciamentos (/deputados/{id}/discursos)
- **🏛️ Monitor de Comissões**: Participação em órgãos (/deputados/{id}/orgaos)  
- **📅 Agenda Parlamentar**: Eventos próximos (/eventos)
- **📈 Rankings Avançados**: Presença, participação, histórico
- **🔄 Histórico Político**: Mudanças de partido e carreira

**⚠️ Analytics de Votações - AÇÃO NECESSÁRIA**:
```go
// Status: Infraestrutura completa, faltam endpoints analytics
// Temos: VotacaoStats, VotacaoAnalysis.tsx, dados da API
// Falta: Implementar no AnalyticsService

GET /api/v1/analytics/votacoes/stats              - Estatísticas gerais
GET /api/v1/analytics/votacoes/rankings/deputados - Ranking participação
GET /api/v1/analytics/votacoes/rankings/disciplina - Disciplina partidária  
GET /api/v1/analytics/votacoes/tendencias         - Análise temporal
```

**Novos Endpoints API**:
```go
GET /api/v1/deputados/{id}/discursos     - Pronunciamentos e análises
GET /api/v1/deputados/{id}/historico     - Mudanças de partido  
GET /api/v1/eventos                      - Agenda parlamentar
GET /api/v1/orgaos/{id}/membros          - Composição comissões
GET /api/v1/analytics/presenca           - Ranking presença eventos
```

**Componentes Frontend**:
- `VotacoesAnalytics.tsx` - Dashboard estatísticas votações *(NOVA - Set/24/2025)*
- `RankingDisciplina.tsx` - Disciplina partidária *(NOVA - Set/24/2025)*  
- `EventosProximos.tsx` - Agenda de reuniões e sessões
- `HistoricoParlamentar.tsx` - Timeline de mudanças
- `AnaliseDiscursos.tsx` - Análise de pronunciamentos
- `MonitorComissoes.tsx` - Dashboard de órgãos

### 3. PWA e suporte offline (média - nov/2025)
**Objetivo**: App funcionar offline para áreas com internet instável

**Implementar**:
- Service Workers para cache
- Manifest.json para instalação
- Cache estratégico de dados essenciais
- Sync em background quando online

### 4. IA Gemini (baixa - dez/2025)
**Objetivo**: Assistente educativo para explicar processos políticos

**Funcionalidades**:
- Chat explicativo sobre votações
- Glossário político interativo
- Resumos automáticos de proposições
- Moderação de comentários

## 🔄 Integrações Pendentes

### **✅ Sistema de Sincronização Completo** 
**Status**: ✅ **IMPLEMENTADO** - Votações incluídas no scheduler diário

**Funcionalidades Ativas**:
- ✅ Sync diário de votações (últimas 7 dias)
- ✅ Votos individuais dos deputados
- ✅ Orientações partidárias oficiais
- ✅ Cache Redis implementado
- ✅ API da Câmara v2 integrada

## 🔍 Descoberta Crítica - Analytics de Votações (Set/24/2025)

**⚠️ Status**: Sistema de votações implementado, mas **analytics agregadas incompletas**

**✅ O que JÁ temos**:
- ✅ `VotacaoStats`, `RankingDeputadoVotacao`, `VotacaoPartido` (domain models)
- ✅ Endpoints: `/votacoes`, `/votacoes/:id`, `/votacoes/:id/completa`  
- ✅ `VotacaoAnalysis.tsx` - Análise detalhada individual
- ✅ API integration completa (votos + orientações partidárias)
- ✅ Repository patterns e cache Redis

**❌ O que está FALTANDO**:
- ❌ Rankings agregados (disciplina partidária, participação deputados)
- ❌ Endpoints `/analytics/votacoes/*` (não existem no AnalyticsService)
- ❌ Dashboard comparativo no frontend
- ❌ Estatísticas temporais e tendências

**🎯 Ação Necessária** (ALTA prioridade):
```go
// Implementar no AnalyticsService:
func (s *AnalyticsService) GetRankingDeputadosVotacao(ctx context.Context, ano int) 
func (s *AnalyticsService) GetRankingPartidosDisciplina(ctx context.Context, ano int)
func (s *AnalyticsService) GetStatsVotacoes(ctx context.Context, periodo string)
```

## 🎯 Cronograma Realista

### **✅ Outubro 2025 - Sistema Completo (FINALIZADO)**
- [x] **Migration 007**: ✅ Tabelas criadas e funcionando
- [x] **HTTP Handlers**: ✅ Endpoints REST para votações implementados
- [x] **API Câmara**: ✅ Client completo para dados de votações
- [x] **Sync Integration**: ✅ Votações no processo diário
- [x] **Testing**: ✅ Endpoints validados e funcionando

### **Novembro 2025 - PWA & Deploy**
- [ ] **Analytics Votações**: Completar rankings e estatísticas *(Semana 1 - PRIORIDADE)*
- [ ] **Service Workers**: Cache offline *(Semana 2)*
- [ ] **GCP Setup**: Configurar infraestrutura *(Semana 3)*  
- [ ] **CI/CD Pipeline**: GitHub Actions para deploy *(Semana 4)*

### **Dezembro 2025 - IA & Refinamentos**
- [ ] **Deploy Produção**: Primeira versão live *(Semana 1 - movido de Nov)*
- [ ] **Assistente Gemini**: Chat educativo básico *(Semana 2)*
- [ ] **Monitoramento**: Métricas e alertas *(Semana 3)*
- [ ] **Performance**: Otimizações baseadas em uso real *(Semana 4)*
- [ ] **Documentação**: API pública e guias *(Semana 4)*

## Bloqueadores Identificados

### 0. Ingestão de despesas indisponível (registrado em 02/out/2025)
Problema: a tabela `despesas` foi criada sem as colunas `cod_tipo_documento` e `valor_documento`, requisitadas pelo `DespesaRepository`. O `COPY FROM` falha e a transação é abortada.
Impacto: nenhuma despesa é persistida; as tabelas `despesas`, `scheduler_executions` e `sync_metrics` permanecem vazias e o scheduler repete a operação em loop.
Plano: aplicar a migration `014_alter_despesas_add_columns.sql`, confirmar o sucesso nos logs do backend, reiniciar o scheduler e validar a inserção de dados.

### 1. Analytics de votações incompletos (registrado em 24/set/2025)
Problema: a infraestrutura de coleta está disponível, porém falta implementação de métodos agregadores no `AnalyticsService`.
Impacto: dashboards sem indicadores de disciplina partidária e participação global.
Plano: implementar métodos agregadores e expor endpoints REST correspondentes; revisar componentes frontend.

### 2. Alinhamento com dados reais de votação
Problema: possíveis diferenças entre a especificação e a estrutura retornada pela API da Câmara.
Plano: validar respostas reais antes de consolidar filtros e parâmetros no client.

### 3. Limitador de taxa em produção
Problema: limite de 100 requisições por minuto na API oficial.
Plano: reforçar cache, mecanismos de retry e janelas de sincronização para evitar bloqueios.

### 4. Custo de infraestrutura GCP
Problema: projeção atual de custo (USD 90-120/mês) pode variar com o tráfego.
Plano: configurar alertas de faturamento e parâmetros de escalonamento controlado antes do go-live.

## ✅ Critérios de Sucesso

### **Funcional**:
- [x] ✅ Sistema de votações completo (GET /api/v1/votacoes)
- [x] ✅ Rankings de deputados funcionam com dados reais
- [x] ✅ API responde <50ms em 95% das requisições
- [ ] App funciona offline por 7 dias
- [ ] Usuário pode comentar em votações

### **Técnico**:
- [x] ✅ Database schema completo e otimizado
- [x] ✅ Logs estruturados com slog
- [ ] Zero downtime durante deploys
- [ ] Backups automáticos diários
- [ ] SSL A+ rating

### **Negócio**:
- [ ] Domínio `todeolho.com.br` acessível
- [x] ✅ 100% dados 2025 sincronizados  
- [x] ✅ Sistema pronto para eleições 2026

---

## 🎯 Objetivo Final

Meta: disponibilizar a plataforma em 30/nov/2025 com:
- Sistema de votações concluído e validado.
- Módulo de analytics com rankings e estatísticas consolidadas.
- Interface em conformidade com WCAG 2.1 AA.
- API REST v1 estabilizada.
- Esquema de banco otimizado e completo.
- Deploy em produção concluído.
- Recursos PWA com suporte offline básico.

Impacto esperado: oferecer ferramenta de transparência política em operação antes do ciclo eleitoral de 2026.
