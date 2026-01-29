"use client";

import { Badge } from "@/components/ui/badge";
import { Clock } from "lucide-react";
import { useEffect, useState } from "react";

export function UpdateBadge() {
  const [lastSync, setLastSync] = useState<string | null>(null);

  useEffect(() => {
    fetch("/api/v1/metadata/last-sync")
      .then((res) => res.json())
      .then((data) => {
        if (data.last_sync) {
           const date = new Date(data.last_sync);
           setLastSync(date.toLocaleDateString("pt-BR", {
             day: "2-digit",
             month: "2-digit",
             year: "numeric",
             hour: "2-digit",
             minute: "2-digit"
           }));
        }
      })
      .catch((err) => console.error("Falha ao buscar last-sync", err));
  }, []);

  if (!lastSync) return null;

  return (
    <Badge variant="outline" className="gap-1.5 text-[10px] sm:text-xs font-normal border-muted-foreground/30 text-muted-foreground">
      <Clock className="w-3 h-3" />
      <span className="hidden sm:inline">Dados atualizados em:</span>
      <span className="sm:hidden">Att:</span>
      <span className="font-medium text-foreground">{lastSync}</span>
    </Badge>
  );
}
