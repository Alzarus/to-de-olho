import { fetcher } from "@/lib/api";

export interface Votacao {
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
  resultado?: string;
  created_at: string;
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
): Promise<VotacaoResponse> => {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString(),
  });

  if (ano) params.append("ano", ano.toString());
  if (materia) params.append("materia", materia);
  if (ordem) params.append("ordem", ordem);
  if (sessao) params.append("sessao", sessao);

  return fetcher<VotacaoResponse>(`/api/v1/votacoes?${params.toString()}`);
};

export const getVotacaoById = async (id: string): Promise<VotacaoDetail> => {
  return fetcher<VotacaoDetail>(`/api/v1/votacoes/${id}`);
};
