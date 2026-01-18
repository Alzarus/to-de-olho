"use client";

import { useQuery } from "@tanstack/react-query";
import {
  getSenador,
  getSenadorScore,
  getVotosPorTipo,
  getEmendas,
  getProposicoes,
  getComissoes,
  getDespesas,
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

export function useProposicoes(
  id: number,
  page: number = 1,
  limit: number = 20,
  search: string = "",
  ano?: number,
  sigla: string = "",
  status: string = "",
  sort: string = "",
) {
  return useQuery({
    queryKey: [
      "senador-proposicoes",
      id,
      page,
      limit,
      search,
      ano,
      sigla,
      status,
      sort,
    ],
    queryFn: () =>
      getProposicoes(id, page, limit, search, ano, sigla, status, sort),
    enabled: id > 0,
    placeholderData: (previousData) => previousData,
  });
}

export function useComissoes(
  id: number,
  page: number = 1,
  limit: number = 20,
  search: string = "",
  status: string = "",
  participacao: string = "",
) {
  return useQuery({
    queryKey: [
      "senador-comissoes",
      id,
      page,
      limit,
      search,
      status,
      participacao,
    ],
    queryFn: () => getComissoes(id, page, limit, search, status, participacao),
    enabled: id > 0,
    placeholderData: (previousData) => previousData,
  });
}

export function useDespesas(
  id: number,
  ano?: number,
  page: number = 1,
  limit: number = 20,
  search: string = "",
  tipo: string = "",
  sort: string = "",
) {
  return useQuery({
    queryKey: ["senador-despesas", id, ano, page, limit, search, tipo, sort],
    queryFn: () => getDespesas(id, ano, page, limit, search, tipo, sort),
    enabled: id > 0,
    placeholderData: (previousData) => previousData,
  });
}
