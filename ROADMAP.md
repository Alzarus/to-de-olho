# Roadmap de Implementacao - To De Olho

> Ultima atualizacao: 13/01/2026 03:25  
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
| 4    | 15-17/01 | Modulo de Ranking                      | PENDENTE  |
| 5    | 18-20/01 | Frontend Next.js                       | PENDENTE  |
| 6    | 21-24/01 | Testes, polimento, deploy              | PENDENTE  |

---

## Requisitos do Ranking (docs/metodologia-ranking.md)

| Criterio                  | Peso | Fonte de Dados                         | Status                       |
| ------------------------- | ---- | -------------------------------------- | ---------------------------- |
| Produtividade Legislativa | 35%  | `/dadosabertos/processo`               | IMPLEMENTADO (ajuste parser) |
| Presenca em Votacoes      | 25%  | `/dadosabertos/votacao`                | CONCLUIDO (85.877 votacoes)  |
| Economia CEAPS            | 20%  | API Administrativa                     | CONCLUIDO                    |
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

## Fase 4: Modulo de Ranking (15-17/01)

- [ ] Modelo `internal/ranking/model.go` (SenadorScore)
- [ ] Service `internal/ranking/service.go`
- [ ] Formula: `Score = (Prod * 0.35) + (Pres * 0.25) + (Econ * 0.20) + (Com * 0.20)`
- [ ] Endpoint `GET /api/v1/ranking`
- [ ] Cache Redis (TTL 1h)

---

## Fase 5: Frontend Next.js (18-20/01)

- [ ] Inicializar em `to-de-olho/frontend/`
- [ ] Tailwind CSS 4 + Shadcn/UI
- [ ] Paginas: `/`, `/ranking`, `/senador/[id]`, `/metodologia`
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
│   └── proposicao/{model,repository,handler,sync}.go
├── pkg/senado/{legis_client,adm_client}.go
├── docker-compose.yml
└── go.mod
```

---

## Dados no Banco (13/01 03:25)

| Tabela           | Registros |
| ---------------- | --------- |
| senadores        | 81        |
| despesas_ceaps   | ~8.000+   |
| votacoes         | 85.877    |
| comissao_membros | 7.186     |
| proposicoes      | ajustando |
