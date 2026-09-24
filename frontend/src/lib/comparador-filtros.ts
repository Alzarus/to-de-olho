// Filtros e agregações das abas do comparador. Funções puras, sem
// dependências: rodam no navegador e no `node --test` (os testes importam
// este arquivo direto). Todo filtro vive na URL, junto de ?ids= e ?ano=.

/** Interface mínima de URLSearchParams (também atende ReadonlyURLSearchParams do Next). */
export interface ParametrosURL {
  get(chave: string): string | null;
  getAll(chave: string): string[];
}

// ---------------------------------------------------------------------------
// Cores das séries: a cor segue o senador (posição na seleção), nunca o rank.
// As variáveis --serie-N ficam em globals.css, com passos próprios para o
// tema escuro (paleta validada: CVD e contraste; ver comentário lá).
// ---------------------------------------------------------------------------

export const TOTAL_CORES_SERIE = 5;

export function corSerie(indice: number): string {
  return `var(--serie-${(indice % TOTAL_CORES_SERIE) + 1})`;
}

// ---------------------------------------------------------------------------
// Teto da CEAPS
// ---------------------------------------------------------------------------

/**
 * Valor mensal da cota por UF. Cópia de `TetoCEAPSPorUF` em
 * backend/internal/ranking/model.go (fonte: Senado Federal, Ato da Comissão
 * Diretora, referência março de 2025, atualizado em 10/01/2026). Se mudar lá,
 * mude aqui.
 */
export const TETO_CEAPS_POR_UF: Readonly<Record<string, number>> = {
  AC: 50426.26, AL: 44500.0, AM: 52798.82, AP: 51103.82,
  BA: 45000.0, CE: 48245.57, DF: 36582.46, ES: 42000.0,
  GO: 36582.46, MA: 47500.0, MG: 40000.0, MS: 42000.0,
  MT: 44500.0, PA: 48207.3, PB: 45000.0, PE: 46000.0,
  PI: 49000.0, PR: 43000.0, RJ: 42000.0, RN: 46000.0,
  RO: 44000.0, RR: 51500.0, RS: 45500.0, SC: 42000.0,
  SE: 53000.0, SP: 40000.0, TO: 36582.46,
};

/** Mesmo valor padrão do backend (`TetoCEAPSMedia / 12`) para UF desconhecida. */
export const TETO_CEAPS_MENSAL_PADRAO = 40000;

export function tetoMensal(uf: string | undefined | null): number {
  if (!uf) return TETO_CEAPS_MENSAL_PADRAO;
  return TETO_CEAPS_POR_UF[uf.toUpperCase()] ?? TETO_CEAPS_MENSAL_PADRAO;
}

/** Percentual com uma casa; null quando a base é zero. */
export function percentual(parte: number, base: number): number | null {
  if (!(base > 0)) return null;
  return Math.round((parte / base) * 1000) / 10;
}

// ---------------------------------------------------------------------------
// Despesas
// ---------------------------------------------------------------------------

export type Granularidade = "mensal" | "trimestral" | "acumulada";
export type UnidadeDespesa = "valor" | "teto";
export type EvolucaoDe = "total" | "categorias";

export interface FiltrosDespesas {
  /** null: padrão (as 5 maiores do recorte) */
  categorias: string[] | null;
  mesDe: number; // 1-12
  mesAte: number; // 1-12
  granularidade: Granularidade;
  unidade: UnidadeDespesa;
  evolucao: EvolucaoDe;
}

export const PARAMS_DESPESAS = {
  categorias: "dcat",
  mesDe: "dde",
  mesAte: "date",
  granularidade: "dgran",
  unidade: "dunid",
  evolucao: "devo",
} as const;

function mesValido(valor: string | null, padrao: number): number {
  const n = Number(valor);
  return Number.isInteger(n) && n >= 1 && n <= 12 ? n : padrao;
}

function umDe<T extends string>(valor: string | null, opcoes: readonly T[], padrao: T): T {
  return (opcoes as readonly string[]).includes(valor ?? "") ? (valor as T) : padrao;
}

export function lerFiltrosDespesas(p: ParametrosURL): FiltrosDespesas {
  const cats = p
    .getAll(PARAMS_DESPESAS.categorias)
    .map((c) => c.trim())
    .filter((c) => c !== "");
  let mesDe = mesValido(p.get(PARAMS_DESPESAS.mesDe), 1);
  let mesAte = mesValido(p.get(PARAMS_DESPESAS.mesAte), 12);
  if (mesDe > mesAte) [mesDe, mesAte] = [mesAte, mesDe];
  return {
    categorias: cats.length > 0 ? [...new Set(cats)] : null,
    mesDe,
    mesAte,
    granularidade: umDe(p.get(PARAMS_DESPESAS.granularidade), ["mensal", "trimestral", "acumulada"] as const, "mensal"),
    unidade: umDe(p.get(PARAMS_DESPESAS.unidade), ["valor", "teto"] as const, "valor"),
    evolucao: umDe(p.get(PARAMS_DESPESAS.evolucao), ["total", "categorias"] as const, "total"),
  };
}

/** O intervalo de meses cobre o ano todo (sem recorte). */
export function anoInteiro(f: Pick<FiltrosDespesas, "mesDe" | "mesAte">): boolean {
  return f.mesDe === 1 && f.mesAte === 12;
}

export interface CategoriaTotal {
  tipo_despesa: string;
  total: number;
}

/** Categorias somadas entre os senadores, da maior para a menor. */
export function categoriasPorVolume(porSenador: ReadonlyArray<ReadonlyArray<CategoriaTotal> | undefined>): string[] {
  const soma = new Map<string, number>();
  for (const lista of porSenador) {
    for (const c of lista ?? []) soma.set(c.tipo_despesa, (soma.get(c.tipo_despesa) ?? 0) + c.total);
  }
  return [...soma.entries()].sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0])).map(([cat]) => cat);
}

export interface MesTotal {
  ano: number;
  mes: number;
  total: number;
}

export interface PontoSerie {
  /** chave de ordenação: "2025-03", "2025-T1" */
  chave: string;
  rotulo: string;
  rotuloCurto: string;
  /** valor no ponto (acumulado quando a granularidade é "acumulada") */
  total: number;
  /** meses que o ponto cobre (base do teto: teto mensal x meses) */
  meses: number;
}

const MESES_ABREV = ["jan", "fev", "mar", "abr", "mai", "jun", "jul", "ago", "set", "out", "nov", "dez"];

/**
 * Agrupa o gasto mensal na granularidade pedida, dentro do intervalo de meses.
 * `ano` 0 (mandato completo): todos os meses, sem recorte de mês.
 * Meses sem gasto entram com zero, para que as séries dos senadores se alinhem.
 */
export function agruparSerie(
  meses: ReadonlyArray<MesTotal>,
  granularidade: Granularidade,
  ano: number,
  mesDe = 1,
  mesAte = 12,
): PontoSerie[] {
  const porMes = new Map<string, number>();
  for (const m of meses) {
    if (ano > 0 && (m.ano !== ano || m.mes < mesDe || m.mes > mesAte)) continue;
    const k = `${m.ano}-${String(m.mes).padStart(2, "0")}`;
    porMes.set(k, (porMes.get(k) ?? 0) + m.total);
  }

  // Grade de meses: o intervalo do ano, ou do primeiro ao último mês com dado
  const grade: Array<{ ano: number; mes: number }> = [];
  if (ano > 0) {
    for (let m = mesDe; m <= mesAte; m++) grade.push({ ano, mes: m });
  } else if (porMes.size > 0) {
    const chaves = [...porMes.keys()].sort();
    let [a, m] = chaves[0].split("-").map(Number);
    const [aFim, mFim] = chaves[chaves.length - 1].split("-").map(Number);
    while (a < aFim || (a === aFim && m <= mFim)) {
      grade.push({ ano: a, mes: m });
      m++;
      if (m > 12) {
        m = 1;
        a++;
      }
    }
  }

  const valorMes = (a: number, m: number) => porMes.get(`${a}-${String(m).padStart(2, "0")}`) ?? 0;

  if (granularidade === "trimestral") {
    const trimestres = new Map<string, PontoSerie>();
    for (const { ano: a, mes: m } of grade) {
      const t = Math.floor((m - 1) / 3) + 1;
      const chave = `${a}-T${t}`;
      const p = trimestres.get(chave) ?? {
        chave,
        rotulo: `${t}º tri/${a}`,
        rotuloCurto: `T${t}/${String(a).slice(2)}`,
        total: 0,
        meses: 0,
      };
      p.total += valorMes(a, m);
      p.meses += 1;
      trimestres.set(chave, p);
    }
    return [...trimestres.values()];
  }

  let acumulado = 0;
  return grade.map(({ ano: a, mes: m }, i) => {
    const v = valorMes(a, m);
    acumulado += v;
    const acum = granularidade === "acumulada";
    return {
      chave: `${a}-${String(m).padStart(2, "0")}`,
      rotulo: `${MESES_ABREV[m - 1]}/${a}`,
      rotuloCurto: `${MESES_ABREV[m - 1]}/${String(a).slice(2)}`,
      total: acum ? acumulado : v,
      meses: acum ? i + 1 : 1,
    };
  });
}

export function nomeMes(mes: number): string {
  return ["janeiro", "fevereiro", "março", "abril", "maio", "junho", "julho", "agosto", "setembro", "outubro", "novembro", "dezembro"][mes - 1] ?? "";
}

// ---------------------------------------------------------------------------
// Fornecedores
// ---------------------------------------------------------------------------

export const OPCOES_TOP = [5, 10, 20] as const;
export type TopN = (typeof OPCOES_TOP)[number];

export interface FiltrosFornecedores {
  categoria: string | null;
  top: TopN;
  busca: string;
}

export const PARAMS_FORNECEDORES = { categoria: "fcat", top: "ftop", busca: "fq" } as const;

export function lerFiltrosFornecedores(p: ParametrosURL): FiltrosFornecedores {
  const top = Number(p.get(PARAMS_FORNECEDORES.top));
  return {
    categoria: p.get(PARAMS_FORNECEDORES.categoria)?.trim() || null,
    top: (OPCOES_TOP as readonly number[]).includes(top) ? (top as TopN) : 5,
    busca: (p.get(PARAMS_FORNECEDORES.busca) ?? "").trim().slice(0, 100),
  };
}

export function somenteDigitos(valor: string): string {
  return valor.replace(/\D/g, "");
}

/**
 * CNPJ com máscara (00.000.000/0000-00). CPF de pessoa física aparece
 * parcialmente oculto (***.000.000-**), como no Portal da Transparência.
 * Outros formatos voltam como vieram.
 */
export function formatarDocumento(doc: string | null | undefined): string {
  if (!doc) return "";
  const d = somenteDigitos(doc);
  if (d.length === 14) return `${d.slice(0, 2)}.${d.slice(2, 5)}.${d.slice(5, 8)}/${d.slice(8, 12)}-${d.slice(12)}`;
  if (d.length === 11) return `***.${d.slice(3, 6)}.${d.slice(6, 9)}-**`;
  return doc.trim();
}

function semAcento(s: string): string {
  return s.normalize("NFD").replace(/[̀-ͯ]/g, "").toLowerCase();
}

/** Busca por nome (sem acento/caixa) ou por CNPJ/CPF (só dígitos, 3 ou mais). */
export function casaBuscaFornecedor(busca: string, nome: string, doc: string): boolean {
  const b = busca.trim();
  if (!b) return true;
  const digitos = somenteDigitos(b);
  if (digitos.length >= 3 && digitos.length === b.replace(/[\s./-]/g, "").length) {
    return somenteDigitos(doc).includes(digitos);
  }
  return semAcento(nome).includes(semAcento(b));
}

// ---------------------------------------------------------------------------
// Emendas
// ---------------------------------------------------------------------------

export type TipoEmenda = "especial" | "finalidade" | "bancada" | "outros";
export type FiltroTipoEmenda = "todos" | TipoEmenda;
export type MetricaEmenda = "pago" | "empenhado";

export const ROTULO_TIPO_EMENDA: Readonly<Record<TipoEmenda, string>> = {
  especial: "Individual – transferência especial (“PIX”)",
  finalidade: "Individual – finalidade definida",
  bancada: "Bancada",
  outros: "Outros tipos",
};

/**
 * Valores reais da API (produção, set/2026): "Emenda Individual -
 * Transferências Especiais" e "Emenda Individual - Transferências com
 * Finalidade Definida". Bancada ainda não aparece na base, mas o filtro já a
 * reconhece se vier.
 */
export function classificarTipoEmenda(tipo: string): TipoEmenda {
  const t = semAcento(tipo);
  if (t.includes("bancada")) return "bancada";
  if (t.includes("especia") || t.includes("pix")) return "especial";
  if (t.includes("finalidade")) return "finalidade";
  return "outros";
}

export interface FiltrosEmendas {
  tipo: FiltroTipoEmenda;
  metrica: MetricaEmenda;
  /** "" todas; "UF:XX" uma UF; senão a localidade exata */
  local: string;
}

export const PARAMS_EMENDAS = { tipo: "etipo", metrica: "emet", local: "eloc" } as const;

export function lerFiltrosEmendas(p: ParametrosURL): FiltrosEmendas {
  return {
    tipo: umDe(p.get(PARAMS_EMENDAS.tipo), ["todos", "especial", "finalidade", "bancada", "outros"] as const, "todos"),
    metrica: umDe(p.get(PARAMS_EMENDAS.metrica), ["pago", "empenhado"] as const, "pago"),
    local: (p.get(PARAMS_EMENDAS.local) ?? "").trim(),
  };
}

const UF_POR_NOME: Readonly<Record<string, string>> = {
  acre: "AC", alagoas: "AL", amapa: "AP", amazonas: "AM", bahia: "BA", ceara: "CE",
  "distrito federal": "DF", "espirito santo": "ES", goias: "GO", maranhao: "MA",
  "mato grosso": "MT", "mato grosso do sul": "MS", "minas gerais": "MG", para: "PA",
  paraiba: "PB", parana: "PR", pernambuco: "PE", piaui: "PI", "rio de janeiro": "RJ",
  "rio grande do norte": "RN", "rio grande do sul": "RS", rondonia: "RO", roraima: "RR",
  "santa catarina": "SC", "sao paulo": "SP", sergipe: "SE", tocantins: "TO",
};

const SIGLAS_UF = new Set(Object.values(UF_POR_NOME));

/**
 * UF de uma localidade de emenda: "ACRE (UF)" -> AC, "RIO BRANCO - AC" -> AC.
 * "MÚLTIPLO", "Nacional" e afins não têm UF (null).
 */
export function ufDaLocalidade(localidade: string): string | null {
  const l = localidade.trim();
  const municipio = /\s-\s([A-Z]{2})$/.exec(l);
  if (municipio && SIGLAS_UF.has(municipio[1])) return municipio[1];
  const estado = /^(.+?)\s*\(UF\)$/i.exec(l);
  if (estado) return UF_POR_NOME[semAcento(estado[1]).trim()] ?? null;
  return null;
}

export function casaLocal(filtro: string, localidade: string): boolean {
  if (!filtro) return true;
  if (filtro.startsWith("UF:")) return ufDaLocalidade(localidade) === filtro.slice(3);
  return localidade === filtro;
}

// ---------------------------------------------------------------------------
// Votações (alinhamento)
// ---------------------------------------------------------------------------

export interface FiltrosVotacoes {
  tipos: string[];
  /** par da lista de divergências: [menor id, maior id] */
  par: [number, number] | null;
}

export const PARAMS_VOTACOES = { tipos: "vtipo", par: "vpar" } as const;

export function lerFiltrosVotacoes(p: ParametrosURL): FiltrosVotacoes {
  const tipos = (p.get(PARAMS_VOTACOES.tipos) ?? "")
    .split(",")
    .map((t) => t.trim().toUpperCase())
    .filter((t) => /^[A-Z0-9-]{1,20}$/.test(t));
  const par = /^(\d+)-(\d+)$/.exec(p.get(PARAMS_VOTACOES.par) ?? "");
  let parLido: [number, number] | null = null;
  if (par) {
    const a = Number(par[1]);
    const b = Number(par[2]);
    if (a !== b && a > 0 && b > 0) parLido = [Math.min(a, b), Math.max(a, b)];
  }
  return { tipos: [...new Set(tipos)], par: parLido };
}

/**
 * Piso da escala de cor do alinhamento. Senadores concordam na maior parte
 * das votações abertas (70% a 90% é o comum): uma escala de 0 a 100 pintaria
 * quase tudo no mesmo passo. Abaixo do piso, o passo mais claro.
 */
export const PISO_RAMPA = 50;

/** Passo da rampa sequencial (1 a 7) para um percentual entre o piso e 100. */
export function passoRampa(pct: number): number {
  const p = Math.max(PISO_RAMPA, Math.min(100, pct));
  return Math.min(7, Math.floor(((p - PISO_RAMPA) / (100 - PISO_RAMPA)) * 7) + 1);
}
