// Testes do toCSV com o runner nativo do Node (node --test), sem dependência
// nova: `npm test` (Node 22.6+ com --experimental-strip-types).
import { test } from "node:test";
import assert from "node:assert/strict";
import {
  toCSV,
  toJSON,
  formatarNumeroCSV,
  slugArquivo,
  buscarTodasPaginas,
  type ColunaCSV,
} from "./export.ts";

interface Linha {
  nome: string;
  gasto: number;
  presentes: number;
  total: number;
  obs?: string | null;
}

const colunas: ColunaCSV<Linha>[] = [
  { cabecalho: "Senador", valor: (l) => l.nome },
  { cabecalho: "Gasto CEAPS (R$)", valor: (l) => l.gasto },
  { cabecalho: "Votações presentes", valor: (l) => l.presentes },
  { cabecalho: "Votações no período", valor: (l) => l.total },
  { cabecalho: "Observação", valor: (l) => l.obs },
];

const DATA = new Date(2026, 8, 23, 10, 0, 0);

function linhas(csv: string): string[] {
  return csv.replace(/^﻿/, "").split("\r\n");
}

test("começa com BOM UTF-8 e usa CRLF", () => {
  const csv = toCSV([], colunas, { extraidoEm: DATA });
  assert.equal(csv.charCodeAt(0), 0xfeff);
  assert.ok(csv.endsWith("\r\n"));
  assert.ok(!/[^\r]\n/.test(csv), "toda quebra deve ser CRLF");
});

test("cabeçalho legível, separador ; e aspas em todos os campos", () => {
  const csv = toCSV(
    [{ nome: "Fulano", gasto: 577444.53, presentes: 5, total: 12 }],
    colunas,
    { extraidoEm: DATA },
  );
  const [cab, l1] = linhas(csv);
  assert.equal(
    cab,
    '"Senador";"Gasto CEAPS (R$)";"Votações presentes";"Votações no período";"Observação"',
  );
  // 577444,53: vírgula decimal, sem milhar; 5 e 12 em colunas separadas (não vira data)
  assert.equal(l1, '"Fulano";"577444,53";"5";"12";""');
});

test("escapa aspas e mantém ; e quebras dentro do campo", () => {
  const csv = toCSV(
    [{ nome: 'Empresa "X"; Ltda\nfilial', gasto: 1, presentes: 0, total: 0 }],
    colunas,
    { extraidoEm: DATA },
  );
  assert.ok(csv.includes('"Empresa ""X""; Ltda\nfilial";"1"'));
});

test("linha final com fonte e data dd/mm/aaaa", () => {
  const csv = toCSV([], colunas, { extraidoEm: DATA });
  const todas = linhas(csv).filter(Boolean);
  assert.equal(
    todas[todas.length - 1],
    '"Fonte: Senado Federal / Portal da Transparência, via Tô De Olho (todeolho.org). Extraído em 23/09/2026"',
  );
});

test("observação vem antes da fonte", () => {
  const csv = toCSV([], colunas, { extraidoEm: DATA, observacao: "Filtro: PIX" });
  const todas = linhas(csv).filter(Boolean);
  assert.equal(todas[todas.length - 2], '"Filtro: PIX"');
});

test("números: inteiros sem casas, negativos, casas configuráveis e não finitos vazios", () => {
  assert.equal(formatarNumeroCSV(1234567), "1234567");
  assert.equal(formatarNumeroCSV(-12.5), "-12,50");
  assert.equal(formatarNumeroCSV(87.3333, 1), "87,3");
  assert.equal(formatarNumeroCSV(Number.NaN), "");
  assert.equal(formatarNumeroCSV(Number.POSITIVE_INFINITY), "");
});

test("null/undefined viram campo vazio e texto-fórmula é neutralizado", () => {
  const csv = toCSV(
    [{ nome: "=HYPERLINK(\"x\")", gasto: -3, presentes: 1, total: 2, obs: null }],
    colunas,
    { extraidoEm: DATA },
  );
  const [, l1] = linhas(csv);
  assert.equal(l1, '"\'=HYPERLINK(""x"")";"-3";"1";"2";""');
});

test("toJSON inclui fonte, data e dados", () => {
  const json = JSON.parse(
    toJSON([{ a: 1 }], { descricao: "teste", extraidoEm: DATA, parametros: { ano: 2026 } }),
  ) as { fonte: string; total: number; dados: unknown[]; parametros: { ano: number } };
  assert.match(json.fonte, /Senado Federal/);
  assert.equal(json.total, 1);
  assert.deepEqual(json.dados, [{ a: 1 }]);
  assert.equal(json.parametros.ano, 2026);
});

test("slugArquivo remove acento e espaço", () => {
  assert.equal(slugArquivo("Davi Alcolumbre — Emendas São Paulo"), "davi-alcolumbre-emendas-sao-paulo");
});

test("buscarTodasPaginas percorre todas as páginas e avisa o progresso", async () => {
  const progresso: string[] = [];
  const itens = await buscarTodasPaginas(
    async (p: number) => ({ lista: [p * 10, p * 10 + 1], paginas: 3 }),
    (r) => ({ itens: r.lista, totalPaginas: r.paginas }),
    (atual, total) => progresso.push(`${atual}/${total}`),
  );
  assert.deepEqual(itens, [10, 11, 20, 21, 30, 31]);
  assert.deepEqual(progresso, ["1/3", "2/3", "3/3"]);
});
