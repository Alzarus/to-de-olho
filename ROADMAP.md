# Roadmap de Implementacao - To De Olho

> Ultima atualizacao: 16/01/2026 02:45  
> Deadline Entrega TCC: 15/01/2026  
> Prazo Projeto Completo: Ate a defesa (25/01 - 11/02/2026)  
> Status: **Frontend concluido, aguardando deploy**

---

## Cronograma Geral

| Fase | Periodo  | Foco                                   | Status    |
| ---- | -------- | -------------------------------------- | --------- |
| 1    | 12/01    | Fundacao (estrutura, senadores, CEAPS) | CONCLUIDA |
| 2    | 12-13/01 | Votacoes                               | CONCLUIDA |
| 3    | 13-14/01 | Comissoes + Proposicoes                | CONCLUIDA |
| 4    | 13/01    | Modulo de Ranking                      | CONCLUIDA |
| 5    | 14-16/01 | Frontend Next.js                       | CONCLUIDA |
| 6    | 21-24/01 | Testes, polimento, deploy              | PENDENTE  |

---

## Requisitos do Ranking (docs/metodologia-ranking.md)

| Criterio                  | Peso | Fonte de Dados                         | Status                       |
| ------------------------- | ---- | -------------------------------------- | ---------------------------- |
| Produtividade Legislativa | 35%  | `/dadosabertos/processo`               | CONCLUIDO (pesos ajustados)  |
| Presenca em Votacoes      | 25%  | `/dadosabertos/votacao`                | CONCLUIDO (85.877 votacoes)  |
| Economia CEAPS            | 20%  | API Administrativa                     | CONCLUIDO (teto variavel UF) |
| Participacao em Comissoes | 20%  | `/dadosabertos/senador/{id}/comissoes` | CONCLUIDO (7.186 registros)  |

**Progresso atual**: 100% dos modulos implementados. Ranking pronto para fase 4.

---

## Auditoria de Requisitos (tcc-escrita/sections/requisitos.tex)

### Requisitos Funcionais (RF)

**Modulo de Senadores:**
| RF | Descricao | Status |
|------|----------------------------------------------------|-------------|
| RF01 | Lista 81 senadores com foto, partido, UF | OK |
| RF02 | Busca por nome, partido, UF | PENDENTE |
| RF03 | Perfil com abas (Visao Geral, Gastos, Gabinete, Votacoes, Emendas) | PARCIAL (falta Gabinete, Emendas) |

**Modulo de Transparencia Financeira (CEAPS):**
| RF | Descricao | Status |
|------|----------------------------------------------------|-------------|
| RF04 | Importar lancamentos CEAPS via APIs | OK |
| RF05 | Visualizar gasto por tipo de despesa | OK |
| RF06 | Fornecedores que mais receberam recursos | PENDENTE |
| RF07 | Alertas para despesas atipicas | PENDENTE |

**Modulo de Emendas e Orcamento:**
| RF | Descricao | Status |
|------|----------------------------------------------------|-------------|
| RF08 | Integrar Portal da Transparencia (emendas) | PENDENTE |
| RF09 | Destacar Transferencias Especiais (emendas PIX) | PENDENTE |
| RF10 | Mapas interativos de distribuicao de emendas | PENDENTE |

**Modulo de Atividade Legislativa:**
| RF | Descricao | Status |
|------|----------------------------------------------------|-------------|
| RF11 | Listar votacoes nominais com voto de cada senador | OK (dados) / PENDENTE (tela) |
| RF12 | Participacao em comissoes com cargo | OK |
| RF13 | Proposicoes com tipo e tramitacao | OK |
| RF14 | Discursos em plenario | PENDENTE |
| RF15 | Agenda de reunioes de comissoes | PENDENTE |
| RF16 | Links para redes sociais | PENDENTE |

**Modulo de Gabinete:**
| RF | Descricao | Status |
|------|----------------------------------------------------|-------------|
| RF17 | Lista de servidores do gabinete | PENDENTE |
| RF18 | Folha de pagamento do gabinete | PENDENTE |

**Modulo de Comparacao e Analise:**
| RF | Descricao | Status |
|------|----------------------------------------------------|-------------|
| RF19 | Comparar 2-5 senadores lado a lado | PENDENTE |
| RF20 | Ranking de fornecedores com cruzamento sancoes | PENDENTE |
| RF21 | Indicadores de confianca (ultima sync, completude) | PARCIAL |

**Modulo de Ranking e Score:**
| RF | Descricao | Status |
|------|----------------------------------------------------|-------------|
| RF22 | Calcular e exibir Score de efetividade | OK |
| RF23 | Grafico radar com 4 dimensoes | PENDENTE |
| RF24 | Ordenar/filtrar por criterio individual | PENDENTE |

### Requisitos Nao-Funcionais (RNF)

**Desempenho:**
| RNF | Descricao | Status |
|-------|---------------------------------------------------|-------------|
| RNF01 | Resposta em ate 2 segundos | OK (cache Redis) |
| RNF02 | Escalabilidade horizontal | OK (Cloud Run) |

**Usabilidade e Acessibilidade:**
| RNF | Descricao | Status |
|-------|---------------------------------------------------|-------------|
| RNF03 | Desktop e mobile | OK |
| RNF04 | Mobile-first | OK |
| RNF05 | WCAG 2.1 AA | PARCIAL |

**Confiabilidade:**
| RNF | Descricao | Status |
|-------|---------------------------------------------------|-------------|
| RNF06 | Sync diario com APIs oficiais | PARCIAL (manual) |
| RNF07 | Disponibilidade 99% | PENDENTE (deploy) |

**Seguranca e Privacidade:**
| RNF | Descricao | Status |
|-------|---------------------------------------------------|-------------|
| RNF08 | HTTPS/TLS | PENDENTE (deploy) |
| RNF09 | LGPD | OK (dados publicos) |

**Manutenibilidade:**
| RNF | Descricao | Status |
|-------|---------------------------------------------------|-------------|
| RNF10 | Arquitetura modular | OK |
| RNF11 | Linting e documentacao | PARCIAL |
| RNF12 | CI/CD | PENDENTE |

### Resumo

| Categoria | Total | OK  | Parcial | Pendente |
| --------- | ----- | --- | ------- | -------- |
| **RF**    | 24    | 8   | 3       | 13       |
| **RNF**   | 12    | 6   | 3       | 3        |
| **Total** | 36    | 14  | 6       | 16       |

---

## Fase 1: Fundacao (CONCLUIDA - 12/01)

- [x] Estrutura do projeto Go (`to-de-olho/backend/`)
- [x] Docker Compose (PostgreSQL 15 + Redis 7)
- [x] Modelos GORM: Senador, Mandato, DespesaCEAPS
- [x] Client API Legislativa (`pkg/senado/legis_client.go`)
- [x] Client API Administrativa (`pkg/senado/adm_client.go`)
- [x] Sync de senadores (81 registros)
- [x] Sync de despesas CEAPS 2024
- [x] Endpoints: `/health`, `/senadores`, `/senadores/{id}`, `/despesas`

---

## Fase 2: Votacoes (CONCLUIDA - 12-13/01)

- [x] Modelo `internal/votacao/model.go`
- [x] Repository com calculo de stats
- [x] Handler com 3 endpoints
- [x] Client para `/dadosabertos/votacao`
- [x] Sync de votacoes (85.877 registros)
- [x] Endpoints: `/votacoes`, `/votacoes/stats`, `/votacoes/tipos`

---

## Fase 3: Comissoes + Proposicoes (CONCLUIDA - 13/01)

### 3.1 Comissoes (Participacao - 20%)

- [x] Modelo `internal/comissao/model.go`
- [x] Client para `/dadosabertos/senador/{id}/comissoes`
- [x] Sync de comissoes (7.186 registros)
- [x] Endpoints: `/comissoes`, `/comissoes/ativas`, `/comissoes/stats`, `/comissoes/casas`

### 3.2 Proposicoes (Produtividade - 35%)

- [x] Modelo `internal/proposicao/model.go`
- [x] Client para `/dadosabertos/processo`
- [x] Sistema de pontuacao (1-16 por estagio, x3 PEC, x2 PLP)
- [x] Endpoints: `/proposicoes`, `/proposicoes/stats`, `/proposicoes/tipos`
- [x] Ajustar parsing da API (formato JSON corrigido - array direto)

---

## Fase 4: Modulo de Ranking (CONCLUIDA - 14/01)

- [x] Modelo `internal/ranking/model.go` (SenadorScore, ScoreDetalhes)
- [x] Service `internal/ranking/service.go` com calculo de scores
- [x] Handler `internal/ranking/handler.go` com endpoints
- [x] Formula: `Score = (Prod * 0.35) + (Pres * 0.25) + (Econ * 0.20) + (Com * 0.20)`
- [x] Endpoints: `GET /ranking`, `GET /ranking/metodologia`, `GET /senadores/:id/score`
- [x] Cache Redis (TTL 1h) com fallback
- [x] Ajustes finos: Pesos RQS/MOC (x0.5), Teto CEAPS por UF, Relatorias (pendente)
- [x] Filtro por ano (`?ano=2025`) implementado no backend

---

## Fase 5: Frontend Next.js (CONCLUIDA - 16/01)

- [x] Inicializar em `to-de-olho/frontend/` com Bun
- [x] Next.js 16.1.1 + React 19.2.3 + TypeScript 5
- [x] Tailwind CSS 4 + Shadcn/UI (9 componentes)
- [x] Recharts 3.6.0 + React Query 5.90.16
- [x] Configurar cores do projeto (senado blue/gold)
- [x] Layout base + identidade visual (Logo/Favicon)
- [x] Paginas: `/`, `/ranking`, `/senador/[id]`, `/metodologia`
- [x] Integrar API backend (React Query) com filtro de ano
- [x] Seletor "Mandato Completo" no ranking e pagina do senador
- [x] Tabs com estado controlado (nao reseta ao mudar ano)
- [ ] Acessibilidade WCAG 2.1 AA (parcial)
- [ ] Graficos Recharts (radar, barras, linha)

---

## Fase 6: Finalizacao (21-24/01)

- [ ] Testes (cobertura 60%+)
- [ ] Dockerfile producao
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
├── src/lib/utils.ts   # Utilitarios
├── package.json       # Bun + dependencias
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

## Melhorias e Correcoes (Atualizado 16/01)

### Pendentes

- [ ] **Testes**: Aumentar cobertura e validar correcoes de ingestao.
- [ ] **Deploy**: Configurar pipeline CI/CD final.
- [ ] **Graficos**: Adicionar graficos Recharts (radar, barras, linha).
- [ ] **Backfill Automatico**: Criar `cmd/backfill/main.go` conforme `docs/estrategia-ingestao-dados.md`.
- [ ] **Docker Compose**: Adicionar variavel `BACKFILL_START_YEAR` para auto-sync no primeiro deploy.
- [ ] **Normalizacao por Mandato**: Diferenciar senadores com mandato 2019-2026 vs 2023-2030, normalizando metricas por meses de exercicio efetivo.
- [ ] **SEO Completo**: Otimizacao de SEO em todas as paginas:
  - Meta tags dinamicas por pagina (`title`, `description`, `og:*`, `twitter:*`)
  - Structured data (JSON-LD) para senadores e ranking
  - Sitemap.xml automatico via Next.js
  - robots.txt configurado
  - URLs canonicas
  - Open Graph images dinamicas para compartilhamento
  - Alt text descritivo em todas imagens
  - Semantic HTML (`<article>`, `<section>`, `<nav>`)
- [ ] **Tela de Votacoes** (`/votacoes`) - RF11 do TCC:
  - **Backend**: Endpoint `GET /api/v1/votacoes` com filtros (ano, mes, tema, senador)
  - **Listagem**: Tabela paginada com votacoes nominais recentes
  - **Detalhes por votacao**: Modal ou pagina com descricao, data, resultado, e voto de cada senador
  - **Filtros**:
    - Periodo (data inicio/fim ou ano)
    - Tema/Materia (PEC, PL, PLP, etc.)
    - Senador especifico (como votou em cada materia)
    - Resultado (aprovadas, rejeitadas, obstrucao)
  - **Visualizacao**:
    - Lista de votos: Sim (verde), Nao (vermelho), Abstencao (amarelo), Ausente (cinza)
    - Grafico de distribuicao de votos por votacao
    - Estatisticas de alinhamento partidario
  - **Mobile-first**: Cards compactos no mobile, tabela no desktop
- [ ] **Modulo de Relatorias** - Bonus para Produtividade Legislativa:
  - **Backend**:
    - Endpoint: `GET /api/v1/relatorias?senador={id}&ano={ano}`
    - Client para `/dadosabertos/processo/relatoria?codigoParlamentar={codigo}`
    - Model `Relatoria` com campos: senador_id, materia, tipo, data_inicio, data_fim, comissao
    - Sync service para ingestao de relatorias
  - **Integracao Ranking**:
    - Bonus conforme `docs/metodologia-ranking.md`: PEC +4pts, PLP/PL +2pts, Comissao +1pt
    - Somar pontos de relatorias ao score de Produtividade Legislativa
  - **Frontend**:
    - Exibir relatorias na pagina do senador (tab ou secao)
    - Badge com total de relatorias no card do ranking
- [ ] **Modulo de Gabinete** (RF17, RF18) - Transparencia de Custos:
  - **Backend**:
    - Endpoint: `GET /api/v1/senadores/{id}/gabinete`
    - Client para API Administrativa: `/api/v1/servidores/servidores?lotacaoEquals={sigla}`
    - Client para `/api/v1/servidores/remuneracoes/{ano}/{mes}`
    - Model `ServidorGabinete`: nome, cargo, vinculo, remuneracao_bruta, remuneracao_liquida
    - Mapeamento: Senador -> Lotacao -> Servidores
  - **Frontend**:
    - Tab "Gabinete" na pagina do senador
    - Lista de servidores com cargo e salario
    - Total mensal da folha do gabinete
    - Comparativo com media do Senado
  - **Valor**: Quem trabalha para o senador e quanto custa
- [ ] **Modulo de Emendas PIX** (RF08, RF09, RF10) - Transparencia Orcamentaria:
  - **Backend**:
    - Endpoint: `GET /api/v1/senadores/{id}/emendas`
    - Client Portal da Transparencia: `/api-de-dados/emendas`
    - Backfill CSV: `download-de-dados/emendas-parlamentares/UNICO`
    - Model `Emenda`: codigo, ano, tipo, valor_empenhado, valor_pago, municipio, uf
    - Filtro especifico para Transferencias Especiais (Emendas PIX)
  - **Frontend**:
    - Tab "Emendas" na pagina do senador
    - Mapa interativo (Leaflet) com destinos geograficos
    - Cards: Total empenhado, Total pago, % Emendas PIX
    - Filtros: ano, tipo, municipio, status
  - **Valor**: Para onde vai o dinheiro das emendas do senador
- [ ] **Filtros do Ranking**: Adicionar opcoes de filtragem e ordenacao:
  - Filtro por partido (ex: PT, PL, MDB)
  - Filtro por UF/região
  - Filtro por ciclo eleitoral (mandato 2019 vs 2023)
  - Ordenação por critério individual (Produtividade, Presença, Economia, Comissões)
  - Busca por nome do senador
  - Paginação com opção de itens por página (10, 25, 50)

### Realizadas (16/01)

- [x] **Fix Votacoes**: Corrigido parsing de datas em `sync.go` com fallback para campo `ano`.
- [x] **Fix Votacoes DB**: Corrigidos 85.771 registros com ano invalido usando ano do `sessao_id`.
- [x] **Re-sync Votacoes**: Re-sincronizadas todas as votacoes (70.432 no mandato).
- [x] **Sync CEAPS**: Sincronizados dados de 2023, 2024, 2025 e 2026 (11 registros).
- [x] **Integridade de Dados**: Adicionadas constraints unicas compostas em `Votacao` e `DespesaCEAPS`.
- [x] **Idempotencia**: Corrigido Upsert para usar `FirstOrCreate` conforme documentado no ADR.
- [x] **Limpeza Duplicatas**: Removidas 223.047 votacoes duplicadas do banco.
- [x] **Auditoria TCC x docs x codigo**: Verificado alinhamento completo (formula, pesos, endpoints).
- [x] **Frontend Ranking**: Adicionada opcao "Mandato Completo" no seletor de ano.
- [x] **Frontend Senador**: Tabs agora usa estado controlado (nao reseta ao mudar ano).
- [x] **Cache Redis**: Limpeza de cache para refletir novos dados.

### Realizadas (15/01)

- [x] **Backend CEAPS**: Correcao do calculo de total do mandato.
- [x] **Frontend Layout**: Padronizacao do icone do rodape.
- [x] **Frontend**: Opcao "Mandato Completo" no seletor de ano.
