// Colunas de exportação compartilhadas (ranking, comparador, ficha do senador).
import type { ColunaCSV } from "@/lib/export";
import type {
  Despesa,
  Emenda,
  FornecedorAgregado,
  SenadorScore,
  VotacaoItem,
} from "@/types/api";

/** Linha do ranking com a situação (classificado ou fora da ordenação) */
export interface LinhaRanking extends SenadorScore {
  foraDaOrdenacao: boolean;
}

export function linhasRanking(
  ranking: readonly SenadorScore[],
  semDados: readonly SenadorScore[] = [],
): LinhaRanking[] {
  return [
    ...ranking.map((s) => ({ ...s, foraDaOrdenacao: false })),
    ...semDados.map((s) => ({ ...s, foraDaOrdenacao: true })),
  ];
}

export const COLUNAS_RANKING: ColunaCSV<LinhaRanking>[] = [
  { cabecalho: "Senador", valor: (s) => s.nome },
  { cabecalho: "Partido", valor: (s) => s.partido },
  { cabecalho: "UF", valor: (s) => s.uf },
  {
    cabecalho: "Situação no ranking",
    valor: (s) => (s.foraDaOrdenacao ? "Dados insuficientes" : "Classificado"),
  },
  { cabecalho: "Posição", valor: (s) => (s.foraDaOrdenacao || !s.posicao ? null : s.posicao) },
  { cabecalho: "Nota final (0-100)", valor: (s) => (s.foraDaOrdenacao ? null : s.score_final) },
  { cabecalho: "Produtividade (0-100)", valor: (s) => s.produtividade },
  { cabecalho: "Presença (0-100)", valor: (s) => s.presenca },
  { cabecalho: "Economia da cota (0-100)", valor: (s) => s.economia_cota },
  { cabecalho: "Comissões (0-100)", valor: (s) => s.comissoes },
  { cabecalho: "Proposições (autoria principal)", valor: (s) => s.detalhes.total_proposicoes },
  { cabecalho: "Proposições aprovadas", valor: (s) => s.detalhes.proposicoes_aprovadas },
  { cabecalho: "Transformadas em lei", valor: (s) => s.detalhes.transformadas_em_lei },
  // Presentes e total em colunas separadas: "5/12" o Excel lê como data
  { cabecalho: "Votações com participação", valor: (s) => s.detalhes.votacoes_participadas },
  { cabecalho: "Votações no período", valor: (s) => s.detalhes.total_votacoes },
  { cabecalho: "Gasto CEAPS (R$)", valor: (s) => s.detalhes.gasto_ceaps },
  { cabecalho: "Teto CEAPS no período (R$)", valor: (s) => s.detalhes.teto_ceaps },
  { cabecalho: "Comissões ativas", valor: (s) => s.detalhes.comissoes_ativas },
  { cabecalho: "Motivo (fora da ordenação)", valor: (s) => s.motivo },
];

/** Nome do senador em cada linha, para planilhas com vários senadores */
export interface ComSenador {
  senador: string;
}

export const COLUNAS_DESPESA: ColunaCSV<Despesa>[] = [
  { cabecalho: "Data de emissão", valor: (d) => formatarDataISO(d.data_emissao) },
  { cabecalho: "Ano", valor: (d) => d.ano },
  { cabecalho: "Mês", valor: (d) => d.mes },
  { cabecalho: "Tipo de despesa", valor: (d) => d.tipo_despesa },
  { cabecalho: "Fornecedor", valor: (d) => d.fornecedor },
  { cabecalho: "Detalhamento", valor: (d) => d.detalhe },
  { cabecalho: "Valor (R$)", valor: (d) => d.valor },
];

export const COLUNAS_FORNECEDOR: ColunaCSV<FornecedorAgregado & ComSenador>[] = [
  { cabecalho: "Senador", valor: (f) => f.senador },
  { cabecalho: "Fornecedor", valor: (f) => f.fornecedor },
  { cabecalho: "CNPJ/CPF", valor: (f) => f.cnpj_cpf },
  { cabecalho: "Lançamentos", valor: (f) => f.quantidade },
  { cabecalho: "Total (R$)", valor: (f) => f.total },
];

export const COLUNAS_EMENDA: ColunaCSV<Emenda & ComSenador>[] = [
  { cabecalho: "Senador", valor: (e) => e.senador },
  { cabecalho: "Ano", valor: (e) => e.ano },
  { cabecalho: "Número", valor: (e) => e.numero },
  { cabecalho: "Tipo", valor: (e) => e.tipo },
  { cabecalho: "Funcional programática", valor: (e) => e.funcional_programatica },
  { cabecalho: "Localidade", valor: (e) => e.localidade },
  { cabecalho: "Valor empenhado (R$)", valor: (e) => e.valor_empenhado },
  { cabecalho: "Valor pago (R$)", valor: (e) => e.valor_pago },
  { cabecalho: "Última atualização", valor: (e) => formatarDataISO(e.data_ultima_atualizacao) },
];

const ROTULO_VOTO: Record<string, string> = {
  Sim: "Sim",
  Nao: "Não",
  Abstencao: "Abstenção",
  Obstrucao: "Obstrução",
  NCom: "Não compareceu",
};

export const COLUNAS_VOTACAO: ColunaCSV<VotacaoItem>[] = [
  { cabecalho: "Data", valor: (v) => formatarDataISO(v.data) },
  { cabecalho: "Código da votação", valor: (v) => v.codigo_votacao },
  { cabecalho: "Matéria", valor: (v) => v.materia },
  { cabecalho: "Descrição", valor: (v) => v.descricao_votacao },
  { cabecalho: "Voto", valor: (v) => ROTULO_VOTO[v.voto] ?? v.voto },
  { cabecalho: "Sigla do voto", valor: (v) => v.sigla_voto },
];

/** "2024-03-15T00:00:00Z" -> "15/03/2024", sem passar pelo fuso */
export function formatarDataISO(iso: string | null | undefined): string {
  if (!iso) return "";
  const m = /^(\d{4})-(\d{2})-(\d{2})/.exec(iso);
  return m ? `${m[3]}/${m[2]}/${m[1]}` : iso;
}

export const NOME_MES = [
  "janeiro",
  "fevereiro",
  "março",
  "abril",
  "maio",
  "junho",
  "julho",
  "agosto",
  "setembro",
  "outubro",
  "novembro",
  "dezembro",
];
