"use client";

import { useQuery } from "@tanstack/react-query";
import {
  getSenador,
  getSenadorScore,
  getVotosPorTipo,
  getEmendas,
} from "@/lib/api";

export function useSenador(id: number) {
  return useQuery({
    queryKey: ["senador", id],
    queryFn: () => getSenador(id),
    enabled: id > 0,
  });
}

export function useSenadorScore(id: number, ano?: number) {
  return useQuery({
    queryKey: ["senador-score", id, ano],
    queryFn: () => getSenadorScore(id, ano),
    enabled: id > 0,
  });
}

export function useVotosPorTipo(id: number) {
  return useQuery({
    queryKey: ["senador-votos-tipo", id],
    queryFn: () => getVotosPorTipo(id),
    enabled: id > 0,
  });
}

export function useEmendas(id: number, ano?: number) {
  return useQuery({
    queryKey: ["senador-emendas", id, ano],
    queryFn: () => getEmendas(id, ano),
    enabled: id > 0,
  });
}
