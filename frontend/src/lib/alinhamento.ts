import { fetcher } from "@/lib/api";

// GET /api/v1/votacoes/alinhamento: concordância de votos entre senadores em
// votações nominais abertas (só Sim, Não e Abstenção dos dois contam).

export interface ParAlinhamento {
  senador_a: number;
  senador_b: number;
  votacoes_comuns: number;
  votos_iguais: number;
  /** votos_iguais / votacoes_comuns * 100; null sem votação em comum */
  percentual: number | null;
}

export interface VotoSenador {
  senador_id: number;
  voto: string; // Sim, Não ou Abstenção
}

export interface Divergencia {
  codigo_votacao: number;
  data: string;
  materia: string;
  sigla_materia: string;
  descricao_votacao: string;
  resultado: string;
  votos: VotoSenador[];
}

export interface AlinhamentoResponse {
  ano: number;
  tipos: string[];
  senadores: number[];
  pares: ParAlinhamento[];
  divergencias: Divergencia[];
  total_divergencias: number;
}

export async function getAlinhamento(
  ids: readonly number[],
  ano?: number,
  tipos?: readonly string[],
): Promise<AlinhamentoResponse> {
  const params = new URLSearchParams({ ids: ids.join(",") });
  if (ano) params.set("ano", String(ano));
  if (tipos && tipos.length > 0) params.set("tipo", tipos.join(","));
  return fetcher<AlinhamentoResponse>(`/api/v1/votacoes/alinhamento?${params.toString()}`);
}
