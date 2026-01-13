# Roadmap de Implementacao - To De Olho

> Ultima atualizacao: 13/01/2026 02:15  
> Deadline Entrega TCC: 15/01/2026  
> Prazo Projeto Completo: Ate a defesa (25/01 - 11/02/2026)  
> Status: **Em desenvolvimento**

---

## Cronograma Geral

| Fase | Periodo  | Foco                                   | Status       |
| ---- | -------- | -------------------------------------- | ------------ |
| 1    | 12/01    | Fundacao (estrutura, senadores, CEAPS) | CONCLUIDA    |
| 2    | 12-13/01 | Votacoes                               | CONCLUIDA    |
| 3    | 13-14/01 | Comissoes + Proposicoes                | EM ANDAMENTO |
| 4    | 15-17/01 | Modulo de Ranking                      | PENDENTE     |
| 5    | 18-20/01 | Frontend Next.js                       | PENDENTE     |
| 6    | 21-24/01 | Testes, polimento, deploy              | PENDENTE     |

---

## Requisitos do Ranking (docs/metodologia-ranking.md)

| Criterio                  | Peso | Fonte de Dados                         | Status                      |
| ------------------------- | ---- | -------------------------------------- | --------------------------- |
| Produtividade Legislativa | 35%  | `/dadosabertos/processo`               | PENDENTE                    |
| Presenca em Votacoes      | 25%  | `/dadosabertos/votacao`                | CONCLUIDO (85.877 votacoes) |
| Economia CEAPS            | 20%  | API Administrativa                     | CONCLUIDO                   |
| Participacao em Comissoes | 20%  | `/dadosabertos/senador/{id}/comissoes` | PENDENTE                    |

**Progresso atual**: 45% do ranking pode ser calculado (CEAPS 20% + Votacoes 25%)

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

## Fase 3: Comissoes + Proposicoes (13-14/01)

### 3.1 Comissoes (Participacao - 20%)

- [ ] Modelo `internal/comissao/model.go`
- [ ] Client para `/dadosabertos/senador/{id}/comissoes`
- [ ] Sync de comissoes
- [ ] Endpoint `GET /api/v1/senadores/{id}/comissoes`

### 3.2 Proposicoes (Produtividade - 35%)

- [ ] Modelo `internal/proposicao/model.go`
- [ ] Client para `/dadosabertos/processo`
- [ ] Sync de proposicoes de autoria
- [ ] Endpoint `GET /api/v1/senadores/{id}/proposicoes`

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
│   └── votacao/{model,repository,handler,sync}.go
├── pkg/senado/{legis_client,adm_client}.go
├── docker-compose.yml
└── go.mod
```

---

## Dados no Banco (13/01 02:15)

| Tabela         | Registros |
| -------------- | --------- |
| senadores      | 81        |
| despesas_ceaps | ~8.000+   |
| votacoes       | 85.877    |
