# Roadmap de Implementação - Tô De Olho

> Última atualização: 16/01/2026 02:45  
> Deadline Entrega TCC: 15/01/2026  
> Prazo Projeto Completo: Até a defesa (25/01 - 11/02/2026)  
> Status: **Frontend concluído, aguardando deploy**

---

## Cronograma Geral

| Fase | Período  | Foco                                   | Status    |
| ---- | -------- | -------------------------------------- | --------- |
| 1    | 12/01    | Fundação (estrutura, senadores, CEAPS) | CONCLUÍDA |
| 2    | 12-13/01 | Votações                               | CONCLUÍDA |
| 3    | 13-14/01 | Comissões + Proposições                | CONCLUÍDA |
| 4    | 13/01    | Módulo de Ranking                      | CONCLUÍDA |
| 5    | 14-16/01 | Frontend Next.js                       | CONCLUÍDA |
| 6    | 21-24/01 | Testes, polimento, deploy              | PENDENTE  |

---

## Requisitos do Ranking (docs/metodologia-ranking.md)

| Critério                  | Peso | Fonte de Dados                         | Status                       |
| ------------------------- | ---- | -------------------------------------- | ---------------------------- |
| Produtividade Legislativa | 35%  | `/dadosabertos/processo`               | CONCLUÍDO (pesos ajustados)  |
| Presença em Votações      | 25%  | `/dadosabertos/votacao`                | CONCLUÍDO (85.877 votações)  |
| Economia CEAPS            | 20%  | API Administrativa                     | CONCLUÍDO (teto variável UF) |
| Participação em Comissões | 20%  | `/dadosabertos/senador/{id}/comissoes` | CONCLUÍDO (7.186 registros)  |

**Progresso atual**: 100% dos módulos implementados. Ranking pronto para fase 4.

---

## Auditoria de Requisitos (tcc-escrita/sections/requisitos.tex)

### Requisitos Funcionais (RF)

**Módulo de Senadores:**
| RF | Descrição | Status |
|------|----------------------------------------------------|-------------|
| RF01 | Lista 81 senadores com foto, partido, UF | OK |
| RF02 | Busca por nome, partido, UF | PENDENTE |
| RF03 | Perfil com abas (Visão Geral, Gastos, Gabinete, Votações, Emendas) | PARCIAL (falta Gabinete, Emendas) |

**Módulo de Transparência Financeira (CEAPS):**
| RF | Descrição | Status |
|------|----------------------------------------------------|-------------|
| RF04 | Importar lançamentos CEAPS via APIs | OK |
| RF05 | Visualizar gasto por tipo de despesa | OK |
| RF06 | Fornecedores que mais receberam recursos | PENDENTE |
| RF07 | Alertas para despesas atípicas | PENDENTE |

**Módulo de Emendas e Orçamento:**
| RF | Descrição | Status |
|------|----------------------------------------------------|-------------|
| RF08 | Integrar Portal da Transparência (emendas) | PENDENTE |
| RF09 | Destacar Transferências Especiais (emendas PIX) | PENDENTE |
| RF10 | Mapas interativos de distribuição de emendas | PENDENTE |

**Módulo de Atividade Legislativa:**
| RF | Descrição | Status |
|------|----------------------------------------------------|-------------|
| RF11 | Listar votações nominais com voto de cada senador | OK (dados) / PENDENTE (tela) |
| RF12 | Participação em comissões com cargo | OK |
| RF13 | Proposições com tipo e tramitação | OK |
| RF14 | Discursos em plenário | PENDENTE |
| RF15 | Agenda de reuniões de comissões | PENDENTE |
| RF16 | Links para redes sociais | PENDENTE |

**Módulo de Gabinete:**
| RF | Descrição | Status |
|------|----------------------------------------------------|-------------|
| RF17 | Lista de servidores do gabinete | PENDENTE |
| RF18 | Folha de pagamento do gabinete | PENDENTE |

**Módulo de Comparação e Análise:**
| RF | Descrição | Status |
|------|----------------------------------------------------|-------------|
| RF19 | Comparar 2-5 senadores lado a lado | PENDENTE |
| RF20 | Ranking de fornecedores com cruzamento sanções | PENDENTE |
| RF21 | Indicadores de confiança (última sync, completude) | PARCIAL |

**Módulo de Ranking e Score:**
| RF | Descrição | Status |
|------|----------------------------------------------------|-------------|
| RF22 | Calcular e exibir Score de efetividade | OK |
| RF23 | Gráfico radar com 4 dimensões | PENDENTE |
| RF24 | Ordenar/filtrar por critério individual | PENDENTE |

### Requisitos Não-Funcionais (RNF)

**Desempenho:**
| RNF | Descrição | Status |
|-------|---------------------------------------------------|-------------|
| RNF01 | Resposta em até 2 segundos | OK (cache Redis) |
| RNF02 | Escalabilidade horizontal | OK (Cloud Run) |

**Usabilidade e Acessibilidade:**
| RNF | Descrição | Status |
|-------|---------------------------------------------------|-------------|
| RNF03 | Desktop e mobile | OK |
| RNF04 | Mobile-first | OK |
| RNF05 | WCAG 2.1 AA | PARCIAL |

**Confiabilidade:**
| RNF | Descrição | Status |
|-------|---------------------------------------------------|-------------|
| RNF06 | Sync diário com APIs oficiais | PARCIAL (manual) |
| RNF07 | Disponibilidade 99% | PENDENTE (deploy) |

**Segurança e Privacidade:**
| RNF | Descrição | Status |
|-------|---------------------------------------------------|-------------|
| RNF08 | HTTPS/TLS | PENDENTE (deploy) |
| RNF09 | LGPD | OK (dados públicos) |

**Manutenibilidade:**
| RNF | Descrição | Status |
|-------|---------------------------------------------------|-------------|
| RNF10 | Arquitetura modular | OK |
| RNF11 | Linting e documentação | PARCIAL |
| RNF12 | CI/CD | PENDENTE |

### Resumo

| Categoria | Total | OK  | Parcial | Pendente |
| --------- | ----- | --- | ------- | -------- |
| **RF**    | 24    | 8   | 3       | 13       |
| **RNF**   | 12    | 6   | 3       | 3        |
| **Total** | 36    | 14  | 6       | 16       |

---

## Fase 1: Fundação (CONCLUÍDA - 12/01)

- [x] Estrutura do projeto Go (`to-de-olho/backend/`)
- [x] Docker Compose (PostgreSQL 15 + Redis 7)
- [x] Modelos GORM: Senador, Mandato, DespesaCEAPS
- [x] Client API Legislativa (`pkg/senado/legis_client.go`)
- [x] Client API Administrativa (`pkg/senado/adm_client.go`)
- [x] Sync de senadores (81 registros)
- [x] Sync de despesas CEAPS 2024
- [x] Endpoints: `/health`, `/senadores`, `/senadores/{id}`, `/despesas`

---

## Fase 2: Votações (CONCLUÍDA - 12-13/01)

- [x] Modelo `internal/votacao/model.go`
- [x] Repository com cálculo de stats
- [x] Handler com 3 endpoints
- [x] Client para `/dadosabertos/votacao`
- [x] Sync de votações (85.877 registros)
- [x] Endpoints: `/votacoes`, `/votacoes/stats`, `/votacoes/tipos`

---

## Fase 3: Comissões + Proposições (CONCLUÍDA - 13/01)

### 3.1 Comissões (Participação - 20%)

- [x] Modelo `internal/comissao/model.go`
- [x] Client para `/dadosabertos/senador/{id}/comissoes`
- [x] Sync de comissões (7.186 registros)
- [x] Endpoints: `/comissoes`, `/comissoes/ativas`, `/comissoes/stats`, `/comissoes/casas`

### 3.2 Proposições (Produtividade - 35%)

- [x] Modelo `internal/proposicao/model.go`
- [x] Client para `/dadosabertos/processo`
- [x] Sistema de pontuação (1-16 por estágio, x3 PEC, x2 PLP)
- [x] Endpoints: `/proposicoes`, `/proposicoes/stats`, `/proposicoes/tipos`
- [x] Ajustar parsing da API (formato JSON corrigido - array direto)

---

## Fase 4: Módulo de Ranking (CONCLUÍDA - 14/01)

- [x] Modelo `internal/ranking/model.go` (SenadorScore, ScoreDetalhes)
- [x] Service `internal/ranking/service.go` com cálculo de scores
- [x] Handler `internal/ranking/handler.go` com endpoints
- [x] Fórmula: `Score = (Prod * 0.35) + (Pres * 0.25) + (Econ * 0.20) + (Com * 0.20)`
- [x] Endpoints: `GET /ranking`, `GET /ranking/metodologia`, `GET /senadores/:id/score`
- [x] Cache Redis (TTL 1h) com fallback
- [x] Ajustes finos: Pesos RQS/MOC (x0.5), Teto CEAPS por UF, Relatorias (pendente)
- [x] Filtro por ano (`?ano=2025`) implementado no backend

---

## Fase 5: Frontend Next.js (CONCLUÍDA - 16/01)

- [x] Inicializar em `to-de-olho/frontend/` com Bun
- [x] Next.js 16.1.1 + React 19.2.3 + TypeScript 5
- [x] Tailwind CSS 4 + Shadcn/UI (9 componentes)
- [x] Recharts 3.6.0 + React Query 5.90.16
- [x] Configurar cores do projeto (senado blue/gold)
- [x] Layout base + identidade visual (Logo/Favicon)
- [x] Páginas: `/`, `/ranking`, `/senador/[id]`, `/metodologia`
- [x] Integrar API backend (React Query) com filtro de ano
- [x] Seletor "Mandato Completo" no ranking e página do senador
- [x] Tabs com estado controlado (não reseta ao mudar ano)
- [ ] Acessibilidade WCAG 2.1 AA (parcial)
- [ ] Gráficos Recharts (radar, barras, linha)

---

## Fase 6: Finalização (21-24/01)

- [ ] Testes (cobertura 60%+)
- [ ] Dockerfile produção
- [ ] CI/CD GitHub Actions
- [ ] Deploy Cloud Run

---

## Arquivos Criados

```
to-de-olho/backend/
├── cmd/api/main.go
├── internal/
│   ├── api/router.go
│   ├── senador/{model,repository,handler,sync}.go
│   ├── ceaps/{model,repository,handler,sync}.go
│   ├── votacao/{model,repository,handler,sync}.go
│   ├── comissao/{model,repository,handler,sync}.go
│   ├── proposicao/{model,repository,handler,sync}.go
│   └── ranking/{model,service,handler}.go
├── pkg/senado/{legis_client,adm_client}.go
├── docker-compose.yml
└── go.mod

to-de-olho/frontend/
├── src/app/           # App Router (Next.js 16)
├── src/components/ui/ # 9 componentes Shadcn
├── src/lib/utils.ts   # Utilitários
├── package.json       # Bun + dependências
└── bun.lock
```

---

## Dados no Banco (16/01 21:45)

| Tabela           | Registros | Mandato 2023-2026 |
| ---------------- | --------- | ----------------- |
| senadores        | 81        | 81                |
| despesas_ceaps   | 26.542    | 26.542            |
| votacoes         | ~170.000  | 70.432            |
| comissao_membros | 7.287     | 3.686             |
| proposicoes      | ~9.000    | 5.008             |

---

## Melhorias e Correções (Atualizado 16/01)

### Pendentes

- [ ] **Testes**: Aumentar cobertura e validar correções de ingestão.
- [ ] **Deploy**: Configurar pipeline CI/CD final.
- [ ] **Gráficos**: Adicionar gráficos Recharts (radar, barras, linha).
- [ ] **Backfill Automático**: Criar `cmd/backfill/main.go` conforme `docs/estrategia-ingestao-dados.md`.
- [ ] **Docker Compose**: Adicionar variável `BACKFILL_START_YEAR` para auto-sync no primeiro deploy.
- [ ] **Normalização por Mandato**: Diferenciar senadores com mandato 2019-2026 vs 2023-2030, normalizando métricas por meses de exercício efetivo.
- [ ] **SEO Completo**: Otimização de SEO em todas as páginas:
  - Meta tags dinâmicas por página (`title`, `description`, `og:*`, `twitter:*`)
  - Structured data (JSON-LD) para senadores e ranking
  - Sitemap.xml automático via Next.js
  - robots.txt configurado
  - URLs canônicas
  - Open Graph images dinâmicas para compartilhamento
  - Alt text descritivo em todas imagens
  - Semantic HTML (`<article>`, `<section>`, `<nav>`)
- [ ] **Tela de Votações** (`/votacoes`) - RF11 do TCC:
  - **Backend**: Endpoint `GET /api/v1/votacoes` com filtros (ano, mês, tema, senador)
  - **Listagem**: Tabela paginada com votações nominais recentes
  - **Detalhes por votação**: Modal ou página com descrição, data, resultado, e voto de cada senador
  - **Filtros**:
    - Período (data início/fim ou ano)
    - Tema/Matéria (PEC, PL, PLP, etc.)
    - Senador específico (como votou em cada matéria)
    - Resultado (aprovadas, rejeitadas, obstrução)
  - **Visualização**:
    - Lista de votos: Sim (verde), Não (vermelho), Abstenção (amarelo), Ausente (cinza)
    - Gráfico de distribuição de votos por votação
    - Estatísticas de alinhamento partidário
  - **Mobile-first**: Cards compactos no mobile, tabela no desktop
- [ ] **Módulo de Relatorias** - Bônus para Produtividade Legislativa:
  - **Backend**:
    - Endpoint: `GET /api/v1/relatorias?senador={id}&ano={ano}`
    - Client para `/dadosabertos/processo/relatoria?codigoParlamentar={codigo}`
    - Model `Relatoria` com campos: senador_id, matéria, tipo, data_inicio, data_fim, comissão
    - Sync service para ingestão de relatorias
  - **Integração Ranking**:
    - Bônus conforme `docs/metodologia-ranking.md`: PEC +4pts, PLP/PL +2pts, Comissão +1pt
    - Somar pontos de relatorias ao score de Produtividade Legislativa
  - **Frontend**:
    - Exibir relatorias na página do senador (tab ou seção)
    - Badge com total de relatorias no card do ranking
- [ ] **Módulo de Gabinete** (RF17, RF18) - Transparência de Custos:
  - **Backend**:
    - Endpoint: `GET /api/v1/senadores/{id}/gabinete`
    - Client para API Administrativa: `/api/v1/servidores/servidores?lotacaoEquals={sigla}`
    - Client para `/api/v1/servidores/remuneracoes/{ano}/{mes}`
    - Model `ServidorGabinete`: nome, cargo, vínculo, remuneracao_bruta, remuneracao_liquida
    - Mapeamento: Senador -> Lotação -> Servidores
  - **Frontend**:
    - Tab "Gabinete" na página do senador
    - Lista de servidores com cargo e salário
    - Total mensal da folha do gabinete
    - Comparativo com média do Senado
  - **Valor**: Quem trabalha para o senador e quanto custa
- [ ] **Módulo de Emendas PIX** (RF08, RF09, RF10) - Transparência Orçamentária:
  - **Backend**:
    - Endpoint: `GET /api/v1/senadores/{id}/emendas`
    - Client Portal da Transparência: `/api-de-dados/emendas`
    - Backfill CSV: `download-de-dados/emendas-parlamentares/UNICO`
    - Model `Emenda`: código, ano, tipo, valor_empenhado, valor_pago, município, uf
    - Filtro específico para Transferências Especiais (Emendas PIX)
  - **Frontend**:
    - Tab "Emendas" na página do senador
    - Mapa interativo (Leaflet) com destinos geográficos
    - Cards: Total empenhado, Total pago, % Emendas PIX
    - Filtros: ano, tipo, município, status
  - **Valor**: Para onde vai o dinheiro das emendas do senador
- [ ] **Filtros do Ranking**: Adicionar opções de filtragem e ordenação:
  - Filtro por partido (ex: PT, PL, MDB)
  - Filtro por UF/região
  - Filtro por ciclo eleitoral (mandato 2019 vs 2023)
  - Ordenação por critério individual (Produtividade, Presença, Economia, Comissões)
  - Busca por nome do senador
  - Paginação com opção de itens por página (10, 25, 50)

### Realizadas (16/01)

- [x] **Fix Votações**: Corrigido parsing de datas em `sync.go` com fallback para campo `ano`.
- [x] **Fix Votações DB**: Corrigidos 85.771 registros com ano inválido usando ano do `sessao_id`.
- [x] **Re-sync Votações**: Re-sincronizadas todas as votações (70.432 no mandato).
- [x] **Sync CEAPS**: Sincronizados dados de 2023, 2024, 2025 e 2026 (11 registros).
- [x] **Integridade de Dados**: Adicionadas constraints únicas compostas em `Votacao` e `DespesaCEAPS`.
- [x] **Idempotência**: Corrigido Upsert para usar `FirstOrCreate` conforme documentado no ADR.
- [x] **Limpeza Duplicatas**: Removidas 223.047 votações duplicadas do banco.
- [x] **Auditoria TCC x docs x código**: Verificado alinhamento completo (fórmula, pesos, endpoints).
- [x] **Frontend Ranking**: Adicionada opção "Mandato Completo" no seletor de ano.
- [x] **Frontend Senador**: Tabs agora usa estado controlado (não reseta ao mudar ano).
- [x] **Cache Redis**: Limpeza de cache para refletir novos dados.

### Realizadas (15/01)

- [x] **Backend CEAPS**: Correção do cálculo de total do mandato.
- [x] **Frontend Layout**: Padronização do ícone do rodapé.
- [x] **Frontend**: Opção "Mandato Completo" no seletor de ano.
