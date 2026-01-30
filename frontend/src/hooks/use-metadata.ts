"use client";

import { useQuery } from "@tanstack/react-query";
import { getLastSync } from "@/lib/api";

export function useLastSync() {
  return useQuery({
    queryKey: ["metadata-last-sync"],
    queryFn: getLastSync,
    // Manter cache por 5 minutos no frontend para evitar requests em navegação
    staleTime: 5 * 60 * 1000,
    refetchOnWindowFocus: false,
    placeholderData: (previousData) => previousData,
  });
}
