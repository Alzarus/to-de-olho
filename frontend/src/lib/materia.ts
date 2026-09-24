// Nome popular e descrição breve das matérias (PEC, PL, PLP...).
// O apelido vem do Senado (campo oficial) ou da curadoria do projeto, sempre
// com fonte; nada aqui é gerado por IA.

// Nome por extenso das siglas de matéria que aparecem nas votações e proposições
export const NOMES_TIPO: Record<string, string> = {
  PEC: "Proposta de Emenda à Constituição",
  MSF: "Mensagem (indicação de autoridades)",
  OFS: "Ofício",
  PLP: "Projeto de Lei Complementar",
  PL: "Projeto de Lei",
  MPV: "Medida Provisória",
  PDL: "Projeto de Decreto Legislativo",
  RQS: "Requerimento",
  REQ: "Requerimento",
  PRS: "Projeto de Resolução do Senado",
  PLS: "Projeto de Lei do Senado",
  PLC: "Projeto de Lei da Câmara",
  SCD: "Substitutivo da Câmara dos Deputados",
  PLN: "Projeto de Lei do Congresso Nacional",
  PDS: "Projeto de Decreto Legislativo do Senado",
};

export const nomeTipo = (sigla: string) => NOMES_TIPO[sigla] ?? sigla;

export type FonteApelido = "oficial" | "curadoria";

/** Campos de nome popular que a API junta às votações e proposições. */
export interface CamposMateria {
  apelido?: string | null;
  apelido_fonte?: FonteApelido | null;
  apelido_fonte_url?: string | null; // só na curadoria
  explicacao_ementa?: string | null;
  temas?: string[] | null;
}

// Matérias cuja votação já se descreve pela própria descrição (indicação de
// autoridade, ofício): não usam ementa nem nome popular.
const SIGLAS_DESCRICAO_VOTACAO = new Set(["MSF", "OFS"]);

export const usaDescricaoVotacao = (sigla?: string | null) =>
  !!sigla && SIGLAS_DESCRICAO_VOTACAO.has(sigla.toUpperCase());

/** "PL 2338/2023 (Substitutivo-CD)" → "Projeto de Lei 2338/2023 (Substitutivo-CD)" */
export function identificacaoPorExtenso(identificacao: string): string {
  const texto = identificacao.trim();
  const [sigla, ...resto] = texto.split(/\s+/);
  if (!sigla || !NOMES_TIPO[sigla.toUpperCase()]) return texto;
  return [nomeTipo(sigla.toUpperCase()), ...resto].join(" ");
}

/** Título da matéria: o apelido, senão o tipo por extenso e o número. */
export function tituloMateria(m: CamposMateria, identificacao: string): string {
  const apelido = m.apelido?.trim();
  if (apelido) return apelido;
  return identificacaoPorExtenso(identificacao);
}

/** Descrição breve: a explicação da ementa do Senado, senão a ementa. */
export function descricaoMateria(m: CamposMateria, ementa?: string | null): string {
  return m.explicacao_ementa?.trim() || ementa?.trim() || "";
}
