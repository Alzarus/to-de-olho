import { fetcher } from "@/lib/api";

export interface Votacao {
  id: number;
  senador_id: number;
  sessao_id: string;
  codigo_sessao: string;
  data: string;
  voto: string;
  descricao_votacao: string;
  materia: string;
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
  materia?: string
): Promise<VotacaoResponse> => {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString(),
  });

  if (ano) params.append("ano", ano.toString());
  if (materia) params.append("materia", materia);

  return fetcher<VotacaoResponse>(`/api/v1/votacoes?${params.toString()}`);
};

export const getVotacaoById = async (id: string): Promise<VotacaoDetail> => {
  return fetcher<VotacaoDetail>(`/api/v1/votacoes/${id}`);
};
