"use client";

import { Badge } from "@/components/ui/badge";
import { useLastSync } from "@/hooks/use-metadata";
import { Clock } from "lucide-react";
import { useEffect, useState } from "react";

// compacto: só a data; o rótulo fica para leitores de tela e na dica do mouse
export function UpdateBadge({ compacto = false }: { compacto?: boolean }) {
  const { data } = useLastSync();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => setMounted(true), 0);
    return () => clearTimeout(timer);
  }, []);

  if (!mounted || !data?.last_sync) return null;

  const date = new Date(data.last_sync);
  const lastSyncFormatted = date.toLocaleDateString("pt-BR", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit"
  });

  if (!lastSyncFormatted) return null;

  return (
    <Badge
      variant="outline"
      className="w-full justify-center sm:w-auto inline-flex gap-1 sm:gap-1.5 text-[10px] sm:text-xs font-normal border-muted-foreground/30 text-muted-foreground py-1 whitespace-nowrap"
      title={compacto ? `Dados atualizados em ${lastSyncFormatted}` : undefined}
    >
      <Clock className="w-3 h-3 sm:w-3 sm:h-3" aria-hidden="true" />
      <span className={compacto ? "sr-only" : undefined}>Dados atualizados em:</span>
      <span className="font-medium text-foreground">{lastSyncFormatted}</span>
    </Badge>
  );
}
