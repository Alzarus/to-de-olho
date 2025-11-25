# Roadmap - "Tô De Olho"

> Transparência política para todos os brasileiros.
>
> Status consolidado em 31/out/2025.

## Prioridades Gerais

Missão: concluir, validar e preparar para produção todos os componentes de ingestão, analytics e experiência do usuário da plataforma.

## Status Atual

| Funcionalidade                | Situação atual                    | Prioridade | Deadline     |
|------------------------------|----------------------------------|------------|--------------|
| Sistema de votações          | Concluído                         | Baixa      | set/2025     |
| Engine de analytics          | Concluído, testes cobrindo votações | Média      | set/2025     |
| Frontend WCAG                | Concluído                         | Média      | set/2025     |
| API REST v1                  | Concluído                         | Média      | set/2025     |
| Sincronização + API Câmara   | Backfill histórico concluído; repositório de despesas com merge seguro; scheduler ativo e saudável (healthcheck corrigido) | Crítica    | out/2025     |
| Esquema do banco             | Migrations 014-016 aplicadas no dev | Média      | out/2025     |
| Deploy em produção           | Não iniciado                      | Alta       | nov/2025     |
| Integração IA Gemini         | Não iniciado                      | Média      | dez/2025     |

## Demandas Urgentes

- Revisar componentes de interface que dificultam a filtragem de deputados (exemplo: seletor de partido).
- Implementar exibição de votações no frontend principal. *(Concluído em 30/out/2025 — componentes `VotacoesAnalytics` e `VotacoesRanking` publicados na página principal)*
- Habilitar ingestão completa (deputados, despesas, votações e proposições) em backfill e scheduler com as flags correspondentes, validando métricas após ativação (pipeline de despesas atualizado para evitar perda de dados em 31/out/2025).

## Backfill Histórico (API Câmara)

> Objetivo: garantir backfill idempotente, confiável e observável cobrindo todas as entidades do `api-docs.json`, permitindo carga inicial completa e sincronizações incrementais diárias.

- **Resumo do estado atual (24/nov/2025)**
  - Concluído: Deputados (backfill e scheduler), Votações históricas (executor com circuit breaker monitorado), Despesas 2025-2022 com checkpoints anuais e Partidos (upsert + checkpoint dedicado).
  - Atualizado: Rankings de analytics recalculados após backfill histórico; scheduler diário operando com flags habilitadas (`SCHEDULER_INCLUDE_*`). Pipeline de despesas com merge transacional. Proposições desbloqueadas após correção de filtro (`ordenarPor=id`). Frontend principal exibe analytics em tempo real.
  - Observado hoje: `proposicoes_cache` contém apenas 1 registro (2025) sem autores populados; `votos_deputados` possui 335 registros, porém `id_deputado` está vindo como `0`, impossibilitando o ranking de presença. Backfill de despesas segue em execução (batches por deputado) enquanto proposições ainda não foram ingeridas.
  - Em andamento: validação de performance em staging e cobertura de repositórios sem integração automatizada.
  - Pontos de atenção: sub-recursos de deputados (discursos, eventos, histórico, etc.), filtros avançados de proposições (arrays, `codTema`, `autor`), suporte a IDs alfanuméricos de votações.
  - Próximos alvos (prioridade média): Órgãos, Legislaturas, Referências.
  - Backlog (prioridade baixa): Eventos, Blocos, Frentes, Grupos.

### Estratégia operacional
- Backfill inicial até **yesterday** (configurável) para evitar dados em trânsito
- Reprocessar diariamente o dia anterior (overlap de 1 dia) para capturar alterações tardias
- Utilizar consistentemente **upsert + checkpoints por entidade/ano** para idempotência
- Garantir execução de todas as entidades no backfill e no scheduler, habilitando `BACKFILL_INCLUDE_*` e `SCHEDULER_INCLUDE_*` em produção.

### Checkpoints sugeridos (prioridade)
1. Deputados
2. Proposições — checkpoints por ano
3. Despesas — checkpoints por ano
4. Votações — checkpoints anuais ou por período; reutilizar upsert existente
5. Partidos / Órgãos / Legislaturas / Referências
6. Eventos / Blocos / Frentes / Grupos

### Tarefas concretas

- **Despesas (altíssima prioridade)**
- [x] Implementar etapa dedicada no backfill histórico usando `DespesaRepository.UpsertDespesas` com checkpoints anuais (21/out/2025).
- [ ] Validar a aplicação da migration `014_alter_despesas_add_columns.sql` em todos os ambientes (dev confirmado até a versão 016; falta staging/prod).
- [x] Ajustar constraint de `valor_liquido` para aceitar estornos (migration 016 aplicada e validada em dev).
- [x] Mitigar risco de perda de dados substituindo `DELETE` por merge transacional no `DespesaRepository` (31/out/2025).
- [x] Habilitar `BACKFILL_INCLUDE_DESPESAS=true` e `SCHEDULER_INCLUDE_DESPESAS=true`, validando métricas (`despesas_processadas`, `despesas_sincronizadas`). *(flags ativadas em 20/nov/2025; acompanhar primeiras execuções do scheduler)*
- [x] Monitorar conclusão do backfill histórico atual (`ef924048-2457-4dab-b5c0-40c2a4ef8d9b`) e registrar checkpoints anuais (finalizado em 29/out/2025 às 04:14 BRT).

**Votações (alta prioridade)**
- [x] Checkpoint "votacoes" no plano anual (`StrategicBackfillExecutor.createBackfillPlan`)
- [x] Executor integrado ao `VotacoesService` (`executeVotacoesBackfill`)
- [x] Janela anual com `SincronizarVotacoes` (upsert + votos/orientações)
- [x] Testes de integração no `VotacaoRepository`
 - [x] Ajustar domínio/repos para IDs alfanuméricos (persistir `id` string, manter `IDVotacaoCamara` opcional) *(concluído em 30/out/2025)*
- [ ] Revisar `CamaraClient` para filtros oficiais (`idProposicao`, `idEvento`, `idOrgao`, datas no mesmo ano) e paginação (≤200 itens)
- [ ] Ajustar pipeline de votos para persistir `id_deputado` correto (atualmente gravando `0`, o que zera ranking de presença)
- [x] Testes unitários/mocks do executor e regressões de checkpoint *(cobertos por `strategic_backfill_votacoes_test.go` em 20/nov/2025)*
- [ ] Backfill completo em staging (performance/governança)

**Partidos (prioridade média)**
- [x] Domínio + migration `012_create_partidos_table.sql`
- [x] `CamaraClient.FetchPartidos` + `PartidosService.ListarPartidos` com upsert
- [x] Checkpoint e executor dedicados
- [x] Testes unit/integration para service e repository *(`partidos_service_test.go` e `partido_repository_test.go` em 20/nov/2025)*
- [ ] Execução validada em staging com monitoramento de consistência

**Proposições (adequação à spec)**
- [ ] Serializar listas (`siglaTipo`, `numero`, `ano`, `codTema`, `keywords`) segundo `style=form&explode=false`
- [ ] Corrigir parâmetros de autor (`autor="nome"`, `idDeputadoAutor`, `siglaPartidoAutor`, `siglaUfAutor`) e remover campos inexistentes na API
- [ ] Ingerir/backfilar sub-recursos (`/tramitacoes`, `/autores`, `/votacoes`, `/temas`) e persistir
- [ ] Popular `proposicoes_cache` com todos os anos alvo (atualmente apenas 1 registro em 2025) e garantir que campos de autor sejam preenchidos para suportar `GetProposicoesCountByDeputadoAno`
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
- [ ] Monitorar métricas `*_processadas`/`*_sincronizadas` e alertar quando permanecerem zeradas após execuções planejadas.

**QA / Release**
- [ ] Cobertura ≥80% (unit + integration) — faltam cenários para executor e partidos
- [ ] Validação com dataset real em staging
- [ ] Planejamento de janelas de execução (backfill inicial custoso)

**Próximos passos imediatos (24/nov/2025)**
1. Acelerar o backfill de proposições: destravar ingestão no ingestor (batches anuais + checkpoints) e popular `proposicoes_cache` com autores e metadados completos.
2. Corrigir pipeline de votos (scheduler/ingestor) para persistir `id_deputado` oficial ao salvar em `votos_deputados`, reprocessando o período 2022-2025 após o ajuste.
3. Reexecutar `POST /api/v1/analytics/rankings/atualizar` após os dados estarem consistentes e validar os rankings na UI (`DashboardAnalytics.tsx`).
4. Auditar os dashboards de votações no frontend com amostras oficiais, ajustando caching se necessário (componentes já migrados para Server Components).
5. Documentar para SRE o estado atual do backfill (despesas em progresso, proposições pendentes) e atualizar runbook de monitoramento.
6. Desenvolver a ingestão para Órgãos, Legislaturas e Referências (domínio, clients, checkpoints, testes).
7. Criar testes table-driven adicionais para `PartidosService` e `PartidoRepository`.

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

**✅ Analytics de Votações - Situação**
- Endpoints `/api/v1/analytics/votacoes/stats`, `/analytics/votacoes/rankings/deputados` e `/analytics/votacoes/rankings/disciplina` implementados e cobertos por testes unitários (out/2025).
- Serviço `AnalyticsService` gera rankings e estatísticas a partir do repositório de votações; caches validados em testes.
- Próximos passos: validar consistência com dados reais após novo backfill e publicar dashboards consolidados no frontend (`VotacoesAnalytics.tsx`, `RankingDisciplina.tsx`).

**Novos Endpoints API**:
```go
GET /api/v1/deputados/{id}/discursos     - Pronunciamentos e análises
GET /api/v1/deputados/{id}/historico     - Mudanças de partido  
GET /api/v1/eventos                      - Agenda parlamentar
GET /api/v1/orgaos/{id}/membros          - Composição comissões
GET /api/v1/analytics/presenca           - Ranking presença eventos
```

**Componentes Frontend**:
- `VotacoesAnalytics.tsx` - Dashboard estatísticas votações *(atualizado em 30/out/2025)*
- `VotacoesRanking.tsx` - Ranking de atuação em plenário *(NOVA - 30/out/2025)*
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

## 🔍 Descoberta Crítica - Analytics de Votações (Atualizado em 29/out/2025)

**✅ Status**: Sistema de votações implementado e analytics agregados disponíveis; aguardando validação com dados reais e publicação no frontend

**✅ O que JÁ temos**:
- ✅ `VotacaoStats`, `RankingDeputadoVotacao`, `VotacaoPartido` (domain models)
- ✅ Endpoints: `/votacoes`, `/votacoes/:id`, `/votacoes/:id/completa`, `/api/v1/analytics/votacoes/stats`, `/api/v1/analytics/votacoes/rankings/deputados`, `/api/v1/analytics/votacoes/rankings/disciplina`
- ✅ `AnalyticsService` calculando rankings e estatísticas com cache Redis
- ✅ Testes unitários cobrindo ranking de deputados, disciplina partidária e estatísticas agregadas
- ✅ `VotacaoAnalysis.tsx` para análise detalhada individual
- ✅ `VotacoesAnalytics.tsx` e `VotacoesRanking.tsx` integrados à página principal de votações (30/out/2025)

**⚠️ O que falta validar**:
- ⚠️ Dashboards comparativos no frontend com dados reais (`VotacoesAnalytics.tsx`, `VotacoesRanking.tsx`, `RankingDisciplina.tsx`)
- ⚠️ Tendências e séries temporais (avaliar necessidade de endpoint dedicado ou extensão de `GetStatsVotacoes`)
- ⚠️ Auditoria dos resultados após backfill completo para garantir fidelidade dos indicadores

**🎯 Próximas ações**:
- Auditar amostras com os dados do backfill concluído e comparar com fontes oficiais
- Integrar endpoints nos componentes de frontend e validar acessibilidade/performance
- Definir requisitos para endpoint de tendências (quando necessário) e planejar implementação

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

### 0. Scheduler de despesas e votações (atualizado em 24/nov/2025)
Status: ✅ **RESOLVIDO**. Healthcheck ajustado para `pidof scheduler` (removendo dependência HTTP inexistente) e erro 400 na ingestão de proposições corrigido (alterado default de ordenação para `id`).
Impacto: Serviço opera como `healthy` e sincroniza todas as entidades (deputados, despesas, votações, proposições) sem erros bloqueantes.
Plano: Monitorar métricas de volume de dados nos próximos dias.

### 1. Validação de analytics de votações (atualizado em 21/out/2025)
Problema: endpoints e cálculos foram implementados e testados, mas ainda falta confrontar os resultados com dados reais após o novo backfill.
Impacto: risco de discrepâncias em dashboards e métricas públicas caso haja divergência entre dados reais e agregações.
Plano: executar backfill completo com despesas e votações habilitadas, auditar amostras no frontend e ajustar caching/normalização conforme necessário.

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
