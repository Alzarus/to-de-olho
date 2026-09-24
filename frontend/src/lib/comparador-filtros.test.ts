// Testes dos filtros do comparador com o runner nativo do Node (node --test).
import { test } from "node:test";
import assert from "node:assert/strict";
import {
  agruparSerie,
  casaBuscaFornecedor,
  casaLocal,
  categoriasPorVolume,
  classificarTipoEmenda,
  formatarDocumento,
  lerFiltrosDespesas,
  lerFiltrosEmendas,
  lerFiltrosFornecedores,
  lerFiltrosVotacoes,
  passoRampa,
  percentual,
  tetoMensal,
  ufDaLocalidade,
} from "./comparador-filtros.ts";

const q = (s: string) => new URLSearchParams(s);

test("lerFiltrosDespesas: padrões e valores inválidos", () => {
  assert.deepEqual(lerFiltrosDespesas(q("")), {
    categorias: null,
    mesDe: 1,
    mesAte: 12,
    granularidade: "mensal",
    unidade: "valor",
    evolucao: "total",
  });
  const f = lerFiltrosDespesas(q("dcat=A%2C+B&dcat=C&dcat=C&dde=9&date=3&dgran=trimestral&dunid=teto&devo=categorias"));
  assert.deepEqual(f.categorias, ["A, B", "C"]);
  assert.equal(f.mesDe, 3); // intervalo invertido é corrigido
  assert.equal(f.mesAte, 9);
  assert.equal(f.granularidade, "trimestral");
  assert.equal(f.unidade, "teto");
  const ruim = lerFiltrosDespesas(q("dde=0&date=13&dgran=semanal&dunid=x"));
  assert.equal(ruim.mesDe, 1);
  assert.equal(ruim.mesAte, 12);
  assert.equal(ruim.granularidade, "mensal");
  assert.equal(ruim.unidade, "valor");
});

test("categoriasPorVolume soma entre senadores", () => {
  const cats = categoriasPorVolume([
    [{ tipo_despesa: "A", total: 10 }, { tipo_despesa: "B", total: 5 }],
    undefined,
    [{ tipo_despesa: "B", total: 7 }, { tipo_despesa: "C", total: 1 }],
  ]);
  assert.deepEqual(cats, ["B", "A", "C"]);
});

test("agruparSerie: mensal, trimestral e acumulada no intervalo", () => {
  const meses = [
    { ano: 2025, mes: 1, total: 100 },
    { ano: 2025, mes: 2, total: 50 },
    { ano: 2025, mes: 4, total: 30 },
    { ano: 2024, mes: 12, total: 999 }, // outro ano: fora
  ];
  const mensal = agruparSerie(meses, "mensal", 2025, 1, 4);
  assert.deepEqual(mensal.map((p) => p.total), [100, 50, 0, 30]); // março sem gasto entra com zero
  assert.equal(mensal[0].rotulo, "jan/2025");

  const tri = agruparSerie(meses, "trimestral", 2025, 2, 6);
  assert.deepEqual(tri.map((p) => [p.chave, p.total, p.meses]), [
    ["2025-T1", 50, 2],
    ["2025-T2", 30, 3],
  ]);

  const acum = agruparSerie(meses, "acumulada", 2025, 1, 4);
  assert.deepEqual(acum.map((p) => [p.total, p.meses]), [
    [100, 1],
    [150, 2],
    [150, 3],
    [180, 4],
  ]);
});

test("agruparSerie: mandato completo preenche os meses entre o primeiro e o último", () => {
  const s = agruparSerie(
    [
      { ano: 2024, mes: 11, total: 1 },
      { ano: 2025, mes: 2, total: 2 },
    ],
    "mensal",
    0,
  );
  assert.deepEqual(s.map((p) => p.chave), ["2024-11", "2024-12", "2025-01", "2025-02"]);
  assert.deepEqual(agruparSerie([], "mensal", 0), []);
});

test("teto por UF e percentual", () => {
  assert.equal(tetoMensal("SE"), 53000);
  assert.equal(tetoMensal("df"), 36582.46);
  assert.equal(tetoMensal("XX"), 40000);
  assert.equal(tetoMensal(undefined), 40000);
  assert.equal(percentual(26500, 53000), 50);
  assert.equal(percentual(1, 3), 33.3);
  assert.equal(percentual(10, 0), null);
});

test("lerFiltrosFornecedores", () => {
  assert.deepEqual(lerFiltrosFornecedores(q("")), { categoria: null, top: 5, busca: "" });
  assert.deepEqual(lerFiltrosFornecedores(q("fcat=Passagens&ftop=20&fq=+posto+")), {
    categoria: "Passagens",
    top: 20,
    busca: "posto",
  });
  assert.equal(lerFiltrosFornecedores(q("ftop=7")).top, 5);
});

test("formatarDocumento: CNPJ com máscara, CPF parcialmente oculto", () => {
  assert.equal(formatarDocumento("11222333000144"), "11.222.333/0001-44");
  assert.equal(formatarDocumento("11.222.333/0001-44"), "11.222.333/0001-44");
  assert.equal(formatarDocumento("12345678901"), "***.456.789-**");
  assert.equal(formatarDocumento(""), "");
  assert.equal(formatarDocumento("EXTERIOR"), "EXTERIOR");
});

test("casaBuscaFornecedor: nome sem acento ou CNPJ por dígitos", () => {
  assert.ok(casaBuscaFornecedor("", "Qualquer", ""));
  assert.ok(casaBuscaFornecedor("aviacao", "TAM Aviação", "02.012.862/0001-60"));
  assert.ok(casaBuscaFornecedor("02.012.862", "TAM", "02.012.862/0001-60"));
  assert.ok(casaBuscaFornecedor("02012862", "TAM", "02.012.862/0001-60"));
  assert.ok(!casaBuscaFornecedor("99999", "TAM", "02.012.862/0001-60"));
  assert.ok(casaBuscaFornecedor("posto 1", "Posto 10", "")); // mistura letras e números: busca no nome
});

test("classificarTipoEmenda com os valores reais da API", () => {
  assert.equal(classificarTipoEmenda("Emenda Individual - Transferências Especiais"), "especial");
  assert.equal(classificarTipoEmenda("Emenda Individual - Transferências com Finalidade Definida"), "finalidade");
  assert.equal(classificarTipoEmenda("Emenda de Bancada"), "bancada");
  assert.equal(classificarTipoEmenda("Emenda de Relator"), "outros");
});

test("lerFiltrosEmendas", () => {
  assert.deepEqual(lerFiltrosEmendas(q("")), { tipo: "todos", metrica: "pago", local: "" });
  assert.deepEqual(lerFiltrosEmendas(q("etipo=especial&emet=empenhado&eloc=UF%3AAC")), {
    tipo: "especial",
    metrica: "empenhado",
    local: "UF:AC",
  });
  assert.equal(lerFiltrosEmendas(q("etipo=xpto")).tipo, "todos");
});

test("ufDaLocalidade e casaLocal", () => {
  assert.equal(ufDaLocalidade("ACRE (UF)"), "AC");
  assert.equal(ufDaLocalidade("SÃO PAULO (UF)"), "SP");
  assert.equal(ufDaLocalidade("RIO BRANCO - AC"), "AC");
  assert.equal(ufDaLocalidade("PLÁCIDO DE CASTRO - AC"), "AC");
  assert.equal(ufDaLocalidade("MÚLTIPLO"), null);
  assert.equal(ufDaLocalidade("Nacional"), null);
  assert.ok(casaLocal("", "MÚLTIPLO"));
  assert.ok(casaLocal("UF:AC", "RIO BRANCO - AC"));
  assert.ok(!casaLocal("UF:AC", "SÃO PAULO (UF)"));
  assert.ok(casaLocal("MÚLTIPLO", "MÚLTIPLO"));
});

test("lerFiltrosVotacoes", () => {
  assert.deepEqual(lerFiltrosVotacoes(q("")), { tipos: [], par: null });
  assert.deepEqual(lerFiltrosVotacoes(q("vtipo=pec,PL,PL,x y&vpar=9-3")), { tipos: ["PEC", "PL"], par: [3, 9] });
  assert.equal(lerFiltrosVotacoes(q("vpar=3-3")).par, null);
});

test("passoRampa", () => {
  assert.equal(passoRampa(0), 1); // abaixo do piso de 50%
  assert.equal(passoRampa(50), 1);
  assert.equal(passoRampa(72.9), 4);
  assert.equal(passoRampa(77.5), 4);
  assert.equal(passoRampa(82.5), 5);
  assert.equal(passoRampa(100), 7);
  assert.equal(passoRampa(140), 7);
});
