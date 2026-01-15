import type {
  RankingResponse,
  SenadorScore,
  Senador,
  MetodologiaResponse,
} from "@/types/api";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

async function fetcher<T>(endpoint: string): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${endpoint}`, {
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
export async function getRanking(limite?: number): Promise<RankingResponse> {
  const params = limite ? `?limite=${limite}` : "";
  return fetcher<RankingResponse>(`/api/v1/ranking${params}`);
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

export async function getSenadorScore(id: number): Promise<SenadorScore> {
  return fetcher<SenadorScore>(`/api/v1/senadores/${id}/score`);
}
