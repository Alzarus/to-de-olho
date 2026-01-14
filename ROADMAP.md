# Roadmap de Implementacao - To De Olho

> Ultima atualizacao: 14/01/2026 02:00  
> Deadline Entrega TCC: 15/01/2026  
> Prazo Projeto Completo: Ate a defesa (25/01 - 11/02/2026)  
> Status: **Em desenvolvimento**

---

## Cronograma Geral

| Fase | Periodo  | Foco                                   | Status    |
| ---- | -------- | -------------------------------------- | --------- |
| 1    | 12/01    | Fundacao (estrutura, senadores, CEAPS) | CONCLUIDA |
| 2    | 12-13/01 | Votacoes                               | CONCLUIDA |
| 3    | 13-14/01 | Comissoes + Proposicoes                | CONCLUIDA |
| 4    | 13/01    | Modulo de Ranking                      | CONCLUIDA |
| 5    | 14/01    | Frontend Next.js                       | EM PROG   |
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

## Fase 4: Modulo de Ranking (CONCLUIDA - 13/01)

- [x] Modelo `internal/ranking/model.go` (SenadorScore, ScoreDetalhes)
- [x] Service `internal/ranking/service.go` com calculo de scores
- [x] Handler `internal/ranking/handler.go` com endpoints
- [x] Formula: `Score = (Prod * 0.35) + (Pres * 0.25) + (Econ * 0.20) + (Com * 0.20)`
- [x] Endpoints: `GET /ranking`, `GET /ranking/metodologia`, `GET /senadores/:id/score`
- [x] Cache Redis (TTL 1h) com fallback
- [x] Ajustes finos: Pesos RQS/MOC (x0.5), Teto CEAPS por UF, Relatorias (pendente)

---

## Fase 5: Frontend Next.js (14/01 - EM ANDAMENTO)

- [x] Inicializar em `to-de-olho/frontend/` com Bun
- [x] Next.js 16.1.1 + React 19.2.3 + TypeScript 5
- [x] Tailwind CSS 4 + Shadcn/UI (9 componentes)
- [x] Recharts 3.6.0 + React Query 5.90.16
- [x] Configurar cores do projeto (senado blue/gold)
- [x] Layout base (header, footer, navegacao)
- [x] Paginas: `/`, `/ranking`, `/senador/[id]`, `/metodologia`
- [ ] Integrar API backend (React Query)
- [ ] Graficos Recharts (radar, barras, linha)
- [ ] Acessibilidade WCAG 2.1 AA

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

## Dados no Banco (14/01 01:00)

| Tabela           | Registros |
| ---------------- | --------- |
| senadores        | 81        |
| despesas_ceaps   | ~8.000+   |
| votacoes         | 85.877    |
| comissao_membros | 7.186     |
| proposicoes      | ~20.000+  |
