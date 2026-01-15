"use client";

import { useQuery } from "@tanstack/react-query";
import { getSenador, getSenadorScore } from "@/lib/api";

export function useSenador(id: number) {
  return useQuery({
    queryKey: ["senador", id],
    queryFn: () => getSenador(id),
    enabled: id > 0,
  });
}

export function useSenadorScore(id: number) {
  return useQuery({
    queryKey: ["senador-score", id],
    queryFn: () => getSenadorScore(id),
    enabled: id > 0,
  });
}
