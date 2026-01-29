import type {
  RankingResponse,
  SenadorScore,
  Senador,
  MetodologiaResponse,
  VotosPorTipoResponse,
  DespesasResponse,
  DespesasAgregadoResponse,
  EmendasResponse,
  ProposicaoResponse,
  ComissoesResponse,
  VotacoesResponse,
} from "@/types/api";

// Em producao, usamos rewrites do Next.js para /api/* -> backend
const API_BASE_URL = "https://todeolho.org";

export async function fetcher<T>(endpoint: string): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${endpoint}`, {
    // ISR: Revalidar a cada 1 hora (3600s)
    next: { revalidate: 3600 },
    headers: {
      "Content-Type": "application/json",
    },
  });

  if (!res.ok) {
    throw new Error(`API Error: ${res.status} ${res.statusText}`);
  }

  return res.json();
}

// Ranking
export async function getRanking(
  limite?: number,
  ano?: number,
): Promise<RankingResponse> {
  const params = new URLSearchParams();
  if (limite) params.append("limite", limite.toString());
  if (ano) params.append("ano", ano.toString());

  const queryString = params.toString() ? `?${params.toString()}` : "";
  return fetcher<RankingResponse>(`/api/v1/ranking${queryString}`);
}

export async function getMetodologia(): Promise<MetodologiaResponse> {
  return fetcher<MetodologiaResponse>("/api/v1/ranking/metodologia");
}

// Senadores
export async function getSenadores(): Promise<Senador[]> {
  return fetcher<Senador[]>("/api/v1/senadores");
}

export async function getSenador(id: number): Promise<Senador> {
  return fetcher<Senador>(`/api/v1/senadores/${id}`);
}

export async function getSenadorScore(
  id: number,
  ano?: number,
): Promise<SenadorScore> {
  const params = new URLSearchParams();
  if (ano) params.append("ano", ano.toString());

  const queryString = params.toString() ? `?${params.toString()}` : "";
  return fetcher<SenadorScore>(`/api/v1/senadores/${id}/score${queryString}`);
}

export async function getVotosPorTipo(
  id: number,
): Promise<VotosPorTipoResponse> {
  return fetcher<VotosPorTipoResponse>(
    `/api/v1/senadores/${id}/votacoes/tipos`,
  );
}

export async function getVotacoes(
  id: number,
  page: number = 1,
  limit: number = 20,
  voto: string = "",
): Promise<VotacoesResponse> {
  const params = new URLSearchParams();
  params.append("page", page.toString());
  params.append("limit", limit.toString());
  if (voto) params.append("voto", voto);

  return fetcher<VotacoesResponse>(
    `/api/v1/senadores/${id}/votacoes?${params.toString()}`,
  );
}

export async function getDespesas(
  id: number,
  ano?: number,
  page: number = 1,
  limit: number = 20,
  search: string = "",
  tipo: string = "",
  sort: string = "",
): Promise<DespesasResponse> {
  const params = new URLSearchParams();
  if (ano) params.append("ano", ano.toString());
  params.append("page", page.toString());
  params.append("limit", limit.toString());
  if (search) params.append("q", search);
  if (tipo && tipo !== "todos") params.append("tipo", tipo);
  if (sort) params.append("sort", sort);

  const query = params.toString() ? `?${params.toString()}` : "";
  return fetcher<DespesasResponse>(`/api/v1/senadores/${id}/despesas${query}`);
}

export async function getDespesasAgregado(
  id: number,
  ano?: number,
): Promise<DespesasAgregadoResponse> {
  const params = new URLSearchParams();
  if (ano) params.append("ano", ano.toString());
  const query = params.toString() ? `?${params.toString()}` : "";
  return fetcher<DespesasAgregadoResponse>(
    `/api/v1/senadores/${id}/despesas/agregado${query}`,
  );
}

export async function getEmendas(
  id: number,
  ano?: number,
): Promise<EmendasResponse> {
  const params = new URLSearchParams();
  if (ano) params.append("ano", ano.toString());
  const query = params.toString() ? `?${params.toString()}` : "";
  return fetcher<EmendasResponse>(`/api/v1/senadores/${id}/emendas${query}`);
}

export async function getProposicoes(
  id: number,
  page: number = 1,
  limit: number = 20,
  search: string = "",
  ano?: number,
  sigla: string = "",
  status: string = "",
  sort: string = "",
): Promise<ProposicaoResponse> {
  const params = new URLSearchParams();
  params.append("page", page.toString());
  params.append("limit", limit.toString());
  if (search) params.append("q", search);
  if (ano) params.append("ano", ano.toString());
  if (sigla) params.append("sigla", sigla);
  if (status) params.append("status", status);
  if (sort) params.append("sort", sort);

  const query = params.toString() ? `?${params.toString()}` : "";
  return fetcher<ProposicaoResponse>(
    `/api/v1/senadores/${id}/proposicoes${query}`,
  );
}

// Comissoes
export async function getComissoes(
  id: number,
  page: number = 1,
  limit: number = 20,
  search: string = "",
  status: string = "",
  participacao: string = "",
): Promise<ComissoesResponse> {
  const params = new URLSearchParams();
  params.append("page", page.toString());
  params.append("limit", limit.toString());
  if (search) params.append("q", search);
  if (status) params.append("status", status);
  if (participacao) params.append("participacao", participacao);

  const query = params.toString() ? `?${params.toString()}` : "";
  return fetcher<ComissoesResponse>(
    `/api/v1/senadores/${id}/comissoes${query}`,
  );
}
