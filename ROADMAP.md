# Roadmap de Implementação - Tô De Olho

> Última atualização: 17/01/2026 08:40  
> Deadline Entrega TCC: 15/01/2026 (Soft) / Defesa: 25/01 - 11/02  
> Status: **Fase 6 - Finalização & Polimento**

---

## Cronograma Atualizado

| Fase | Foco principal                         | Situação                           |
| ---- | -------------------------------------- | ---------------------------------- |
| 1-5  | Ingestão, API, Ranking, Frontend Base  | **CONCLUÍDAS**                     |
| 6    | **Essenciais TCC** (Emendas, Votações) | **EM ANDAMENTO** (Prioridade Alta) |
| 7    | Polimento, Testes e Deploy             | PENDENTE                           |
| 8    | Backlog (Gabinete, Fornecedores)       | FUTURO                             |

---

## Fase 6: Finalização & Essenciais (Prioridade TCC)

Foco em cobrir os Requisitos Funcionais (RF) explícitos no texto do TCC que ainda faltam.

### Prioridade 1: Funcionalidades Essenciais

- [x] **Tela de Votações (RF11)**
  - Tabela paginada com filtros (Ano, Busca textual).
  - Detalhe: votação nominal com filtros (Nome, Voto, Partido, UF).
  - _Concluída: Lista de votações do senador expandida (500 itens) e consistente com gráfico._
- [x] **Página do Senador e UX (RF21)**
  - [x] Exibição de Mandato (Início/Fim) e Badges de status.
  - [x] Gráfico de Votos Refatorado: Agrupamento "Outros", Tooltip rico e Cores acessíveis.
  - [x] Navegação Contextual: Botão "Voltar" preserva filtros e abas.
  - [x] Correção Sync: Ajuste de timeout para volumes grandes (ex: Magno Malta).
- [ ] **Módulo de Emendas (RF08, RF09, RF10)**
  - [x] Backfill CSV do Portal da Transparência (ingestão streaming + normalização de autor).
  - [x] Resumo anual por senador e endpoint `/senadores/:id/emendas`.
  - [x] Destaque para Emendas PIX (Transferências Especiais).
  - [ ] Mapa de distribuição geográfica simples.
- [x] **Comparador de Senadores (RF19)**
  - [x] Seleção de até 5 senadores via dock flutuante ou página dedicada.
  - [x] Abas: Visão Geral (Radar Chart), Despesas (Gráficos), Fornecedores.
  - [x] Filtros por UF, Partido e busca textual.
- [x] **Visualização do Ranking (RF23)**
  - Gráfico Radar (Recharts) na página do Senador (4 eixos: Produtividade, Presença, Economia, Comissões).

### Prioridade 2: Qualidade & Deploy

- [ ] **Deploy Produção**: Dockerfile otimizado + Cloud Run.
- [ ] **SEO Completo**: Meta tags, Open Graph, Sitemap (Next.js).
- [ ] **Performance**: Cache headers, otimização de imagens (Bun).
- [ ] **Testes**: Garantir cobertura mínima nos serviços críticos (Ranking/Sync).

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
