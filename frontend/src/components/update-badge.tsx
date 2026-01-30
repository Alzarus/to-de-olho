"use client";

import { Badge } from "@/components/ui/badge";
import { useLastSync } from "@/hooks/use-metadata";
import { Clock } from "lucide-react";
import { useEffect, useState } from "react";

export function UpdateBadge() {
  const { data } = useLastSync();
  const [lastSyncFormatted, setLastSyncFormatted] = useState<string | null>(null);

  useEffect(() => {
    if (data?.last_sync) {
       const date = new Date(data.last_sync);
       setLastSyncFormatted(date.toLocaleDateString("pt-BR", {
         day: "2-digit",
         month: "2-digit",
         year: "numeric",
         hour: "2-digit",
         minute: "2-digit"
       }));
    }
  }, [data]);

  if (!lastSyncFormatted) return null;

  return (
    <Badge variant="outline" className="inline-flex gap-1 sm:gap-1.5 text-[9px] sm:text-xs font-normal border-muted-foreground/30 text-muted-foreground">
      <Clock className="w-2.5 h-2.5 sm:w-3 sm:h-3" />
      <span className="hidden sm:inline">Dados atualizados em:</span>
      <span className="sm:hidden">Att:</span>
      <span className="font-medium text-foreground">{lastSyncFormatted}</span>
    </Badge>
  );
}
