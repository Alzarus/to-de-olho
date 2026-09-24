import { fetcher } from "@/lib/api";
import type { CamposMateria } from "@/lib/materia";

export interface Votacao extends CamposMateria {
  id: number;
  senador_id: number;
  codigo_votacao: number; // id da votacao (codigoSessaoVotacao do Senado)
  sessao_id: string; // codigo da sessao: agrupa as votacoes do dia
  codigo_sessao: string;
  sequencial_votacao?: number | null;
  data: string;
  voto: string; // rotulo: Sim, Nao, Abstencao, Obstrucao ou a sigla
  sigla_voto: string; // codigo bruto da API
  descricao_votacao: string;
  materia: string;
  ementa?: string;
  resultado?: string; // A (aprovada), R (rejeitada)
  sigla_materia?: string; // tipo da matéria: PEC, MSF, PLP...
  secreta?: boolean | null; // votação secreta: o voto individual não é publicado
  created_at: string;
}

export interface FiltrosVotacoes {
  tipos?: string[]; // siglas da matéria
  secreta?: boolean; // undefined: abertas e secretas
  resultado?: string; // A ou R
}

export interface Faceta {
  valor: string;
  total: number;
}

export interface FacetasVotacoes {
  tipos: Faceta[];
  secreta: Faceta[]; // valor "true" ou "false"
  resultados: Faceta[];
  total: number;
}

export interface VotacaoResponse {
  data: Votacao[];
  total: number;
  page: number;
  limit: number;
}

export interface VotacaoDetail {
  votacao: Votacao;
  votos: (Votacao & {
    senador_nome: string;
    senador_partido: string;
    senador_uf: string;
    senador_foto: string;
  })[];
}

export const getVotacoes = async (
  page = 1,
  limit = 20,
  ano?: number,
  materia?: string,
  ordem?: string,
  sessao?: string,
  filtros: FiltrosVotacoes = {},
): Promise<VotacaoResponse> => {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString(),
  });

  if (ano) params.append("ano", ano.toString());
  if (materia) params.append("materia", materia);
  if (ordem) params.append("ordem", ordem);
  if (sessao) params.append("sessao", sessao);
  if (filtros.tipos?.length) params.append("tipo", filtros.tipos.join(","));
  if (filtros.secreta !== undefined)
    params.append("secreta", String(filtros.secreta));
  if (filtros.resultado) params.append("resultado", filtros.resultado);

  return fetcher<VotacaoResponse>(`/api/v1/votacoes?${params.toString()}`);
};

export const getVotacoesFacetas = async (
  ano?: number,
): Promise<FacetasVotacoes> => {
  const params = new URLSearchParams();
  if (ano) params.append("ano", ano.toString());
  return fetcher<FacetasVotacoes>(
    `/api/v1/votacoes/facetas?${params.toString()}`,
  );
};

export const getVotacaoById = async (id: string): Promise<VotacaoDetail> => {
  return fetcher<VotacaoDetail>(`/api/v1/votacoes/${id}`);
};
