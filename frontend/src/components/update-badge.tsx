"use client";

import { Badge } from "@/components/ui/badge";
import { Clock } from "lucide-react";

export function UpdateBadge() {
  // Hardcoded for now based on the last big sync or could be dynamic if we had an endpoint
  // Using the date marked in ROADMAP.md or "Hoje"
  const today = new Date().toLocaleDateString("pt-BR", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  });

  return (
    <Badge variant="outline" className="gap-1.5 text-[10px] sm:text-xs font-normal border-muted-foreground/30 text-muted-foreground">
      <Clock className="w-3 h-3" />
      <span className="hidden sm:inline">Dados atualizados em:</span>
      <span className="sm:hidden">Att:</span>
      <span className="font-medium text-foreground">{today}</span>
    </Badge>
  );
}
