"use client";

import { useRanking } from "@/hooks/use-ranking";
import { ComparatorRadarChart } from "./radar-chart";
import { Skeleton } from "@/components/ui/skeleton";
import { AlertCircle } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { SenadorScore } from "@/types/api";

interface OverviewTabProps {
  selectedIds: number[];
}

const COLORS = [
  "#3b82f6", // Blue
  "#22c55e", // Green
  "#eab308", // Yellow
  "#ef4444", // Red
  "#a855f7", // Purple
];

export function OverviewTab({ selectedIds }: OverviewTabProps) {
  const { data, isLoading, error } = useRanking();

  if (isLoading) {
    return <Skeleton className="h-[400px] w-full rounded-lg" />;
  }

  if (error || !data?.ranking) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertTitle>Erro</AlertTitle>
        <AlertDescription>
          Não foi possível carregar os dados para comparação.
        </AlertDescription>
      </Alert>
    );
  }

  // Filter and map senators with colors
  const comparisonData = selectedIds
    .map((id, index) => {
        const senator = data.ranking.find(s => s.senador_id === id);
        if (!senator) return null;
        return {
            ...senator,
            color: COLORS[index % COLORS.length]
        };
    })
    .filter((s): s is SenadorScore & { color: string } => s !== null);

  if (comparisonData.length === 0) {
     return (
        <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Atenção</AlertTitle>
            <AlertDescription>
                Nenhum dado encontrado para os senadores selecionados.
            </AlertDescription>
        </Alert>
     )
  }

  return (
    <div className="grid gap-6 lg:grid-cols-3">
        {/* Radar Chart Section */}
        <div className="lg:col-span-2">
            <ComparatorRadarChart senators={comparisonData} />
        </div>

        {/* Metrics Summary Section */}
        <div className="space-y-4">
            <Card>
                <CardHeader>
                    <CardTitle className="text-lg">Destaques</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                    {comparisonData.map(senator => (
                        <div key={senator.senador_id} className="flex items-center justify-between border-b pb-2 last:border-0 last:pb-0">
                             <div className="flex items-center gap-2">
                                <span className="h-3 w-3 rounded-full" style={{ backgroundColor: senator.color }} />
                                <span className="font-medium text-sm">{senator.nome}</span>
                             </div>
                             <div className="text-right">
                                <div className="font-bold text-lg">{senator.score_final.toFixed(1)}</div>
                                <div className="text-xs text-muted-foreground">Score Final</div>
                             </div>
                        </div>
                    ))}
                </CardContent>
            </Card>

            <Card>
                 <CardHeader>
                    <CardTitle className="text-lg">Ranking Geral</CardTitle>
                </CardHeader>
                <CardContent>
                    <p className="text-sm text-muted-foreground mb-4">
                        Posição relativa entre todos os senadores.
                    </p>
                     {comparisonData.map(senator => (
                        <div key={senator.senador_id} className="flex items-center justify-between mb-2">
                             <span className="text-sm font-medium">{senator.nome}</span>
                             <span className="text-sm font-bold">#{data.ranking.indexOf(senator) + 1}º</span>
                        </div>
                    ))}
                </CardContent>
            </Card>
        </div>
    </div>
  );
}
