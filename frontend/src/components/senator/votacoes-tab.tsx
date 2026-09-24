"use client";

import { useVotosPorTipo, useVotacoes, useSenador } from "@/hooks/use-senador";
import { getVotacoes } from "@/lib/api";
import { buscarTodasPaginas, rotuloAnoArquivo, slugArquivo } from "@/lib/export";
import { COLUNAS_VOTACAO } from "@/lib/export-colunas";
import { ExportarDados } from "@/components/export-dados";
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
import { descricaoMateria, tituloMateria, usaDescricaoVotacao } from "@/lib/materia";
import { DescricaoExpansivel, MarcaCuradoria, TemasChips } from "@/components/materia/materia-info";

const VOTE_LABELS: Record<string, string> = {
  Sim: "Sim",
  Nao: "Não",
  Abstencao: "Abstenção",
  Obstrucao: "Obstrução",
  NCom: "Não Compareceu",
};

export function VotacoesTab({ id, ano }: { id: number; ano?: number }) {
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
  const { data: votacoesData, isLoading: isListLoading } = useVotacoes(id, page, limit, filteredVoto, ano === 0 ? undefined : ano);
  const { data: senador } = useSenador(id);
  const anoApi = ano === 0 ? undefined : ano;

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
                <VotosPieChart data={chartData.por_tipo} onSliceClick={handleSliceClick} activeType={filteredVoto} />
            </CardContent>
        </Card>

        {/* List Section */}
        <Card className="h-fit w-full max-w-full overflow-hidden">
            <CardHeader className="flex flex-row flex-wrap items-center justify-between gap-2">
                <CardTitle>
                    {filteredVoto 
                        ? `Votos: ${VOTE_LABELS[filteredVoto] || filteredVoto}`
                        : "Todas as Votações"
                    }
                </CardTitle>
                {votacoesData && votacoesData.total > 0 && (
                    <ExportarDados
                        descricao={`Votações${anoApi ? ` de ${anoApi}` : ""}`}
                        detalhe={`${votacoesData.total} votações, todas as páginas${filteredVoto ? ", com o filtro atual" : ""}`}
                        nomeArquivo={`todeolho_${slugArquivo(senador?.nome ?? `senador-${id}`)}_votacoes_${rotuloAnoArquivo(anoApi)}`}
                        colunas={COLUNAS_VOTACAO}
                        observacao={
                            filteredVoto
                                ? `Filtro aplicado: voto "${VOTE_LABELS[filteredVoto] || filteredVoto}"`
                                : undefined
                        }
                        parametros={{ senador_id: id, ano: anoApi ?? "todos" }}
                        obter={(aoProgredir) =>
                            // limit 100 é o máximo da API de votações
                            buscarTodasPaginas(
                                (p) => getVotacoes(id, p, 100, filteredVoto, anoApi),
                                (r) => ({ itens: r.votacoes ?? [], totalPaginas: r.total_pages }),
                                aoProgredir,
                            )
                        }
                    />
                )}
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
                            {votacoesData.votacoes.map((v) => {
                                const href = `/votacoes/${v.codigo_votacao}?backUrl=${encodeURIComponent(pathname + "?" + searchParams.toString())}`;
                                const soDescricao = usaDescricaoVotacao(v.sigla_materia);
                                return (
                                // O link cobre o cabeçalho; a descrição fica fora dele para o
                                // botão "ver mais" não ficar aninhado num elemento interativo
                                <div
                                    key={v.id}
                                    className="p-3 rounded-lg border hover:bg-muted/50 transition-colors"
                                >
                                    <Link href={href} className="block rounded-sm focus-visible:outline-2 focus-visible:outline-ring">
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
                                            {v.materia && (
                                                <Badge variant="outline" className="text-xs font-mono">
                                                    {v.materia}
                                                </Badge>
                                            )}
                                            <span className="text-xs text-muted-foreground">
                                                {new Date(v.data).toLocaleDateString("pt-BR")}
                                            </span>
                                        </div>
                                        <p className="font-medium text-sm">
                                            {v.materia ? tituloMateria(v, v.materia) : "Sem matéria"}
                                        </p>
                                    </Link>
                                    {v.apelido_fonte === "curadoria" && <MarcaCuradoria className="mt-1" />}
                                    {soDescricao ? (
                                        <p className="text-xs text-muted-foreground line-clamp-2 mt-1">
                                            {v.descricao_votacao}
                                        </p>
                                    ) : (
                                        <>
                                            <DescricaoExpansivel texto={descricaoMateria(v, v.ementa)} className="mt-1" />
                                            {v.descricao_votacao && (
                                                <p className="text-xs text-muted-foreground line-clamp-1 mt-1">
                                                    {v.descricao_votacao}
                                                </p>
                                            )}
                                            <TemasChips temas={v.temas} className="mt-1.5" />
                                        </>
                                    )}
                                </div>
                                );
                            })}
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
