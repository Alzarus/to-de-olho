"use client";

import { useQuery } from "@tanstack/react-query";
import { getRanking, getMetodologia } from "@/lib/api";

export function useRanking(limite?: number) {
  return useQuery({
    queryKey: ["ranking", limite],
    queryFn: () => getRanking(limite),
  });
}

export function useMetodologia() {
  return useQuery({
    queryKey: ["metodologia"],
    queryFn: getMetodologia,
  });
}
