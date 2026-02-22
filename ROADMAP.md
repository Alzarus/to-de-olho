# Roadmap de Implementação - Tô De Olho

> Ultima atualizacao: 22/02/2026 05:00  
> Status: **Aprovado (Pós-Defesa)** / Fase 8

---

## Cronograma Atualizado

| Fase | Foco principal                         | Situacao       |
| ---- | -------------------------------------- | -------------- |
| 1-5  | Ingestao, API, Ranking, Frontend Base  | **CONCLUIDAS** |
| 6    | **Essenciais TCC** (Emendas, Votacoes) | **CONCLUIDA**  |
| 7    | Polimento, Testes e Deploy             | **CONCLUIDA**  |
| 8    | Expansão Pós-Defesa (Backlog)          | FUTURO         |

---

### Phase 7: Polishing, Testing & Deploy (✅ CONCLUIDA)

- [x] **Deploy**: Setup Cloud Run + Firebase Hosting
- [x] **Tests**: Playwright E2E tests configurados e em pipeline
- [x] **GitHub Actions**: CI/CD automatizado para Cloud Run com Scaling
- [x] **Cloud Computing**:
  - Backend: Cloud Run (Go 1.21+) Escalado
  - Frontend: Cloud Run (Next.js 15) Escalado
  - Proxy/Edge: Firebase Hosting (p/ Custom Domain + SSL)
- [x] **Dominio**: todeolho.org + api.todeolho.org mapeados e rodando
- [x] **SEO Completo**: Meta tags, Open Graph, Sitemap dinamico (Next.js)

### Prioridade 2: Qualidade (✅ Mapeada)

- [x] **Deploy Producao**: Infraestrutura dimensionada e backfill atestado com milhões de inserts.
- [x] **Testes**: Especiais para API vital e validação do fluxo do Senador no E2E.

---

## Fase 6: Finalizacao & Essenciais (CONCLUIDA)

---

## Entregas Realizadas (Resumo Fases 1-5)

Todas as funcionalidades base já estão operacionais no ambiente de desenvolvimento.

### Ranking & Dados (Backend)

- [x] **Ingestão**: Sincronização diária de Senadores, Votações e Comissões.
- [x] **CEAPS**: Dados financeiros de 2023-2026 com cálculo de economia e teto por UF.
- [x] **Algoritmo de Ranking**: `Score = (Prod*0.35) + (Pres*0.25) + (Econ*0.20) + (Com*0.20)`.
- [x] **API**: Endpoints RESTful rápidos (Cache Redis) e seguros.
- [x] **Correções**: Idempotência no banco, limpeza de duplicatas, constraints de integridade.

### Interface (Frontend)

- [x] **Stack Moderno**: Next.js 15, React 19, Tailwind 4, Shadcn/UI.
- [x] **Design**: Identidade visual do Senado (Azul/Dourado), Dark mode, Mobile-first.
- [x] **Features**: Lista de Ranking, Filtros por Ano/Mandato, Página de Detalhe do Senador.

---

## Backlog / Futuro (Pós-MVP)

Funcionalidades listadas no TCC como "Desejáveis" ou para trabalhos futuros, caso falte tempo antes da defesa.

- [ ] **Módulo de Gabinete (RF17, RF18)**: Lista de servidores e folha de pagamento.
- [ ] **Transparência Fornecedores (RF06, RF07, RF20)**: Ranking de recebedores e alertas de suspeita.
- [ ] **Atividade Legislativa Expandida (RF14, RF15, RF16)**: Discursos, Agenda, Redes Sociais.
- [ ] **Relatorias**: Bônus de pontuação para relatores de matérias complexas.

---

## Health Check (Métricas do Sistema)

| Domínio de Dados  | Volume / Status (Pós-Backfill 2023+) | Status Sync (Cloud Run)       |
| ----------------- | ------------------------------------ | ----------------------------- |
| Senadores         | 82 (Ativos e Suplentes)              | ✅ OK (Diário Completo)       |
| Votos Registrados | 33.797 votos                         | ✅ OK (Sincronia Incremental) |
| Despesas CEAPS    | R$ 93.3 Milhões liquidados           | ✅ OK (Diário)                |
| Verbas de Emendas | 1.281 repasses                       | ✅ OK (Sincronia Incremental) |
| Proposições       | Filtrado (Ano >= AnoAtual-2)         | ✅ OK (Sincronia Incremental) |
| Comissões (Vagas) | Participações ativas                 | ✅ OK (Diário)                |
