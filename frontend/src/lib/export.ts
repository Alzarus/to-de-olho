// Exportação de dados (CSV e JSON) para quem quer levar os números para a
// planilha ou para a pesquisa. Sem dependências: roda no navegador e no
// `node --test` (os testes importam este arquivo direto).
//
// Formato do CSV pensado para o Excel em pt-BR:
// - BOM UTF-8 no início (sem ele o Excel abre "Ã§" no lugar de "ç");
// - separador ";" (a vírgula é o separador decimal);
// - todos os campos entre aspas, com aspas internas dobradas;
// - número com vírgula decimal e sem separador de milhar;
// - quebra de linha CRLF (RFC 4180);
// - última linha com a fonte e a data de extração.

export type ValorCelula = string | number | boolean | null | undefined;

export interface ColunaCSV<T> {
  /** Cabeçalho legível, em pt-BR */
  cabecalho: string;
  valor: (linha: T) => ValorCelula;
  /** Casas decimais para números não inteiros (padrão: 2) */
  casas?: number;
}

export const FONTE_DADOS =
  "Fonte: Senado Federal / Portal da Transparência, via Tô De Olho (todeolho.org)";

const BOM = "﻿";
const SEPARADOR = ";";
const QUEBRA = "\r\n";

function doisDigitos(n: number): string {
  return n.toString().padStart(2, "0");
}

/** dd/mm/aaaa no fuso local */
export function formatarDataBR(data: Date): string {
  return `${doisDigitos(data.getDate())}/${doisDigitos(data.getMonth() + 1)}/${data.getFullYear()}`;
}

/** Número com vírgula decimal e sem separador de milhar ("577444,53") */
export function formatarNumeroCSV(valor: number, casas = 2): string {
  if (!Number.isFinite(valor)) return "";
  if (Number.isInteger(valor)) return valor.toString();
  return valor.toFixed(casas).replace(".", ",");
}

// Texto que começa com =, +, - ou @ vira fórmula no Excel (injeção de CSV).
// Números não passam por aqui, então "-12,5" continua número.
function neutralizarFormula(texto: string): string {
  return /^[=+\-@\t\r]/.test(texto) ? `'${texto}` : texto;
}

function formatarCelula(valor: ValorCelula, casas?: number): string {
  if (valor === null || valor === undefined) return "";
  if (typeof valor === "number") return formatarNumeroCSV(valor, casas);
  if (typeof valor === "boolean") return valor ? "Sim" : "Não";
  return neutralizarFormula(valor);
}

function aspas(campo: string): string {
  return `"${campo.replace(/"/g, '""')}"`;
}

function linhaCSV(campos: string[]): string {
  return campos.map(aspas).join(SEPARADOR);
}

export interface OpcoesCSV {
  /** Data da extração (padrão: agora) */
  extraidoEm?: Date;
  /** Observação extra antes da fonte (ex.: filtro aplicado) */
  observacao?: string;
}

export function linhaFonte(extraidoEm: Date = new Date()): string {
  return `${FONTE_DADOS}. Extraído em ${formatarDataBR(extraidoEm)}`;
}

export function toCSV<T>(
  linhas: readonly T[],
  colunas: readonly ColunaCSV<T>[],
  opcoes: OpcoesCSV = {},
): string {
  const saida: string[] = [linhaCSV(colunas.map((c) => c.cabecalho))];
  for (const linha of linhas) {
    saida.push(linhaCSV(colunas.map((c) => formatarCelula(c.valor(linha), c.casas))));
  }
  if (opcoes.observacao) saida.push(linhaCSV([opcoes.observacao]));
  saida.push(linhaCSV([linhaFonte(opcoes.extraidoEm)]));
  return BOM + saida.join(QUEBRA) + QUEBRA;
}

export interface OpcoesJSON {
  descricao: string;
  extraidoEm?: Date;
  /** Filtros e parâmetros usados na consulta */
  parametros?: Record<string, ValorCelula>;
}

/** JSON para pesquisa: os dados crus da API com fonte e data de extração */
export function toJSON<T>(dados: readonly T[], opcoes: OpcoesJSON): string {
  const extraidoEm = opcoes.extraidoEm ?? new Date();
  return JSON.stringify(
    {
      descricao: opcoes.descricao,
      fonte: FONTE_DADOS,
      extraido_em: extraidoEm.toISOString(),
      parametros: opcoes.parametros ?? {},
      total: dados.length,
      dados,
    },
    null,
    2,
  );
}

/** Trecho seguro para nome de arquivo: sem acento, minúsculo, com hífen */
export function slugArquivo(texto: string): string {
  return texto
    .normalize("NFD")
    .replace(/[̀-ͯ]/g, "")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

/** "Mandato" quando o ano é 0 (mandato completo) */
export function rotuloAnoArquivo(ano: number | undefined): string {
  return ano ? ano.toString() : "mandato";
}

export const MIME_CSV = "text/csv;charset=utf-8";
export const MIME_JSON = "application/json;charset=utf-8";

export function baixarArquivo(nome: string, conteudo: string, mime: string): void {
  const blob = new Blob([conteudo], { type: mime });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = nome;
  link.style.display = "none";
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  // Revoga depois: alguns navegadores ainda leem o blob após o click
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}

/**
 * Busca todas as páginas de uma lista paginada, em sequência, avisando o
 * progresso. Para exportar o conjunto completo, não só a página visível.
 */
export async function buscarTodasPaginas<P, I>(
  buscarPagina: (pagina: number) => Promise<P>,
  extrair: (resposta: P) => { itens: I[]; totalPaginas: number },
  aoProgredir?: (paginaAtual: number, totalPaginas: number) => void,
  limitePaginas = 500,
): Promise<I[]> {
  const itens: I[] = [];
  let pagina = 1;
  let totalPaginas = 1;
  do {
    const resposta = extrair(await buscarPagina(pagina));
    itens.push(...resposta.itens);
    totalPaginas = Math.min(resposta.totalPaginas, limitePaginas);
    aoProgredir?.(pagina, Math.max(totalPaginas, 1));
    pagina += 1;
  } while (pagina <= totalPaginas);
  return itens;
}
