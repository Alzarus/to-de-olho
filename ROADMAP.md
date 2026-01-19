# Roadmap de Implementação - Tô De Olho

> Ultima atualizacao: 18/01/2026 22:30  
> Deadline Entrega TCC: 15/01/2026 (Soft) / Defesa: 25/01 - 11/02  
> Status: **Fase 7 - Deploy & Producao**

---

## Cronograma Atualizado

| Fase | Foco principal                         | Situacao         |
| ---- | -------------------------------------- | ---------------- |
| 1-5  | Ingestao, API, Ranking, Frontend Base  | **CONCLUIDAS**   |
| 6    | **Essenciais TCC** (Emendas, Votacoes) | **CONCLUIDA**    |
| 7    | Polimento, Testes e Deploy             | **EM PROGRESSO** |
| 8    | Backlog (Gabinete, Fornecedores)       | FUTURO           |

---

### Phase 7: Polishing, Testing & Deploy (⏳ Em Progresso)

- [x] **Deploy**: Setup Cloud Run + Firebase Hosting
- [ ] **Tests**: Playwright E2E tests
- [ ] **Domain**: Mapeamento final de DNS
- [x] **GitHub Actions**: CI/CD automatizado para Cloud Run
- [x] **Cloud Computing**:
  - Backend: Cloud Run (Go 1.23)
  - Frontend: Cloud Run (Next.js 15)
  - Proxy/Edge: Firebase Hosting (p/ Custom Domain + SSL)
- [ ] **Dominio**: todeolho.org + api.todeolho.org (no mesmo dominio via rewrite)

### Prioridade 2: Qualidade

- [x] **Deploy Producao**: Dockerfile otimizado com backfill e sync diario + Cloud Run.
- [ ] **SEO Completo**: Meta tags, Open Graph, Sitemap (Next.js).
- [ ] **Performance**: Cache headers, otimizacao de imagens (Bun).
- [ ] **Testes**: Garantir cobertura minima nos servicos criticos (Ranking/Sync).

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

| Tabela            | Registros Totais | Mandato Atual (2023+) | Status Sync |
| ----------------- | ---------------- | --------------------- | ----------- |
| Senadores         | 81               | 81                    | ✅ OK       |
| Votações          | ~170k            | 70.432                | ✅ OK       |
| Despesas CEAPS    | ~26k             | 26.542                | ✅ OK       |
| Proposições       | ~9k              | 5.008                 | ✅ OK       |
| Comissões (Vagas) | ~7k              | 3.686                 | ✅ OK       |
