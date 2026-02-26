"use client";

import { useQuery } from "@tanstack/react-query";
import { getRanking, getMetodologia } from "@/lib/api";

export function useRanking(limite?: number, ano?: number, inativos?: boolean) {
  return useQuery({
    queryKey: ["ranking", limite, ano, inativos],
    queryFn: () => getRanking(limite, ano, inativos),
  });
}

export function useMetodologia() {
  return useQuery({
    queryKey: ["metodologia"],
    queryFn: getMetodologia,
  });
}
