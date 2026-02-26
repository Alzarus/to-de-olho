"use client";

import { useVotosPorTipo, useVotacoes } from "@/hooks/use-senador";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { X } from "lucide-react";
import { useRouter, usePathname, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { VotosPieChart } from "@/components/votos-pie-chart";
import { PaginationWithInput } from "@/components/ui/pagination-with-input";

const VOTE_LABELS: Record<string, string> = {
  Sim: "Sim",
  Nao: "Não",
  Abstencao: "Abstenção",
  Obstrucao: "Obstrução",
  NCom: "Não Compareceu",
};

export function VotacoesTab({ id }: { id: number }) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const pathname = usePathname();

  // URL State
  const page = Number(searchParams.get("vot_page") ?? "1");
  const filteredVoto = searchParams.get("vot_type") ?? "";
  const limit = 20;

  // Pie Chart Data
  const { data: chartData, isLoading: isChartLoading } = useVotosPorTipo(id);
  
  // List Data
  const { data: votacoesData, isLoading: isListLoading } = useVotacoes(id, page, limit, filteredVoto);

  const createQueryString = useCallback(
    (name: string, value: string) => {
      const params = new URLSearchParams(searchParams.toString());
      if (value) {
          params.set(name, value);
      } else {
          params.delete(name);
      }
      
      if (name !== "vot_page") {
          params.set("vot_page", "1");
      }
      return params.toString();
    },
    [searchParams]
  );

  const updateUrl = useCallback((name: string, value: string) => {
      router.replace(`${pathname}?${createQueryString(name, value)}`, { scroll: false });
  }, [router, pathname, createQueryString]);

  const setPage = (p: number) => updateUrl("vot_page", p.toString());
  const setFilter = (v: string) => updateUrl("vot_type", v);

  const handleSliceClick = (voteType: string) => {
      setFilter(voteType === filteredVoto ? "" : voteType);
  };

  if (isChartLoading) {
      return <Skeleton className="h-[400px] w-full" />;
  }

  if (!chartData || !chartData.por_tipo) return null;

  return (
    <div className="grid gap-6 lg:grid-cols-2 w-full min-w-0">
        <Card className="h-fit w-full max-w-full overflow-hidden">
            <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle>Distribuição de Votos</CardTitle>
                {filteredVoto && (
                    <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setFilter("")}
                        className="text-muted-foreground"
                    >
                        <X className="mr-1 h-4 w-4" />
                        Limpar filtro
                    </Button>
                )}
            </CardHeader>
            <CardContent>
                <VotosPieChart data={chartData.por_tipo} onSliceClick={handleSliceClick} />
            </CardContent>
        </Card>

        {/* List Section */}
        <Card className="h-fit w-full max-w-full overflow-hidden">
            <CardHeader>
                <CardTitle>
                    {filteredVoto 
                        ? `Votos: ${VOTE_LABELS[filteredVoto] || filteredVoto}`
                        : "Todas as Votações"
                    }
                </CardTitle>
            </CardHeader>
            <CardContent>
                {isListLoading ? (
                     <div className="space-y-4">
                        {[...Array(5)].map((_, i) => (
                            <Skeleton key={i} className="h-16 w-full" />
                        ))}
                    </div>
                ) : !votacoesData || votacoesData.votacoes.length === 0 ? (
                    <div className="text-center py-8 text-muted-foreground">
                        Nenhuma votação encontrada.
                    </div>
                ) : (
                    <div className="space-y-4 min-h-[600px] flex flex-col">
                        <div className="space-y-2 flex-1">
                            {votacoesData.votacoes.map((v) => (
                                <Link
                                    key={v.id}
                                    href={`/votacoes/${v.sessao_id}?backUrl=${encodeURIComponent(pathname + "?" + searchParams.toString())}`}
                                    className="block p-3 rounded-lg border hover:bg-muted/50 transition-colors"
                                >
                                    <div className="flex items-start justify-between gap-2">
                                        <div className="flex-1 min-w-0">
                                            <div className="flex flex-wrap items-center gap-2 mb-1">
                                                <Badge 
                                                    variant={
                                                        v.voto === "Sim" ? "default" :
                                                        v.voto === "Nao" ? "destructive" :
                                                        "secondary"
                                                    } 
                                                    className="text-xs"
                                                >
                                                    {VOTE_LABELS[v.voto] || v.voto}
                                                </Badge>
                                                <span className="text-xs text-muted-foreground">
                                                    {new Date(v.data).toLocaleDateString("pt-BR")}
                                                </span>
                                            </div>
                                            <p className="font-medium text-sm truncate">
                                                {v.materia || "Sem matéria"}
                                            </p>
                                            <p className="text-xs text-muted-foreground line-clamp-1">
                                                {v.descricao_votacao}
                                            </p>
                                        </div>
                                    </div>
                                </Link>
                            ))}
                        </div>

                         <PaginationWithInput 
                            currentPage={votacoesData.page} 
                            totalPages={votacoesData.total_pages} 
                            onPageChange={setPage} 
                            className="border-t pt-4"
                        />
                    </div>
                )}
            </CardContent>
        </Card>
    </div>
  );
}
