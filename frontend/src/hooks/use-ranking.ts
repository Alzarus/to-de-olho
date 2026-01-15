"use client";

import { useQuery } from "@tanstack/react-query";
import { getRanking, getMetodologia } from "@/lib/api";

export function useRanking(limite?: number, ano?: number) {
  return useQuery({
    queryKey: ["ranking", limite, ano],
    queryFn: () => getRanking(limite, ano),
  });
}

export function useMetodologia() {
  return useQuery({
    queryKey: ["metodologia"],
    queryFn: getMetodologia,
  });
}
