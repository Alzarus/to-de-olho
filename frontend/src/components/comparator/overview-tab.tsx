"use client";

import { useRanking } from "@/hooks/use-ranking";
import { ComparatorRadarChart } from "./radar-chart";
import { Skeleton } from "@/components/ui/skeleton";
import { AlertCircle } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import Link from "next/link";
import type { SenadorScore } from "@/types/api";
import type { SenatorBasicProfile } from "@/contexts/comparator-context";
import { corSerie } from "@/lib/comparador-filtros";

interface OverviewTabProps {
  senators: SenatorBasicProfile[];
  year: number;
}

interface ForaDoRanking {
  id: number;
  nome: string;
  motivo: string;
}


export function OverviewTab({ senators, year }: OverviewTabProps) {
  const selectedIds = senators.map((s) => s.id);
  const { data, isLoading, error } = useRanking(undefined, year === 0 ? undefined : year);

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
            color: corSerie(index)
        };
    })
    .filter((s): s is SenadorScore & { color: string } => s !== null);

  const yearLabel = year === 0 ? "Mandato Completo" : year.toString();

  // Quem não está na ordenação do período: em sem_dados (com o motivo da API)
  // ou sem registro nenhum. Antes sumiam do gráfico sem explicação.
  const foraDoRanking: ForaDoRanking[] = senators
    .filter((s) => !data.ranking.some((r) => r.senador_id === s.id))
    .map((s) => {
      const semDados = data.sem_dados?.find((r) => r.senador_id === s.id);
      return {
        id: s.id,
        nome: s.nome,
        motivo:
          semDados?.motivo ||
          (semDados
            ? "Dados insuficientes no período (menos de 6 meses em exercício ou sem registro de votação)."
            : "Sem registro no ranking deste período (pode não ter exercido o mandato nele)."),
      };
    });

  const avisoForaDoRanking = foraDoRanking.length > 0 && (
    <Alert role="status" className="lg:col-span-3">
      <AlertCircle className="h-4 w-4" aria-hidden="true" />
      <AlertTitle>
        {foraDoRanking.length === 1
          ? "1 senador ficou fora da comparação do ranking"
          : `${foraDoRanking.length} senadores ficaram fora da comparação do ranking`}{" "}
        ({yearLabel})
      </AlertTitle>
      <AlertDescription>
        <ul className="mt-1 list-disc space-y-1 pl-5">
          {foraDoRanking.map((s) => (
            <li key={s.id}>
              <Link href={`/senador/${s.id}`} className="font-medium underline-offset-2 hover:underline">
                {s.nome}
              </Link>
              : {s.motivo}
            </li>
          ))}
        </ul>
        <p className="mt-2">
          As abas de despesas, emendas e fornecedores continuam mostrando os dados desses
          senadores. Veja os critérios na{" "}
          <Link href="/metodologia" className="font-medium underline underline-offset-2">
            metodologia
          </Link>
          .
        </p>
      </AlertDescription>
    </Alert>
  );

  if (comparisonData.length === 0) {
     return (
        <div className="space-y-4">
            {avisoForaDoRanking || (
                <Alert>
                    <AlertCircle className="h-4 w-4" />
                    <AlertTitle>Atenção</AlertTitle>
                    <AlertDescription>
                        Nenhum dado encontrado para os senadores selecionados.
                    </AlertDescription>
                </Alert>
            )}
        </div>
     )
  }

  return (
    <div className="grid gap-6 lg:grid-cols-3">
        {avisoForaDoRanking}
        {/* Radar Chart Section */}
        <div className="lg:col-span-2">
            <ComparatorRadarChart senators={comparisonData} year={year} />
        </div>

        {/* Metrics Summary Section */}
        <div className="space-y-4">
            <Card>
                <CardHeader>
                    <CardTitle className="text-lg">Destaques - {yearLabel}</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                    {comparisonData.map(senator => (
                        <div key={senator.senador_id} className="flex items-center justify-between border-b pb-2 last:border-0 last:pb-0">
                             <Link href={`/senador/${senator.senador_id}`} className="flex items-center gap-2 hover:underline cursor-pointer group">
                                <span className="h-3 w-3 rounded-full group-hover:scale-110 transition-transform" style={{ backgroundColor: senator.color }} />
                                <span className="font-medium text-sm">{senator.nome}</span>
                             </Link>
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
                    <CardTitle className="text-lg">Ranking Geral - {yearLabel}</CardTitle>
                </CardHeader>
                <CardContent>
                    <p className="text-sm text-muted-foreground mb-4">
                        Posição relativa entre todos os senadores ({yearLabel}).
                    </p>
                     {comparisonData.map(senator => (
                        <div key={senator.senador_id} className="flex items-center justify-between mb-2">
                             <Link href={`/senador/${senator.senador_id}`} className="text-sm font-medium hover:underline cursor-pointer">
                                {senator.nome}
                             </Link>
                             <span className="text-sm font-bold">#{data.ranking.findIndex(s => s.senador_id === senator.senador_id) + 1}º</span>
                        </div>
                    ))}
                </CardContent>
            </Card>
        </div>
    </div>
  );
}
