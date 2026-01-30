"use client";

import { PaginationWithInput } from "@/components/ui/pagination-with-input";

import { useProposicoes } from "@/hooks/use-senador";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { CheckCircle2, Gavel, Search, ChevronLeft, ChevronRight, Filter, X } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { useRouter, usePathname, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";

export function ProposicoesTab({ id }: { id: number }) {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  // URL State
  const page = Number(searchParams.get("prop_page") ?? "1");
  const searchParam = searchParams.get("prop_q") ?? "";
  const sigla = searchParams.get("prop_type") ?? "todos";
  const status = searchParams.get("prop_status") ?? "todos";
  const sort = searchParams.get("prop_sort") ?? "data_desc";
  const limit = 20;

  // Local state for input debounce
  const [searchValue, setSearchValue] = useState(searchParam);

  const createQueryString = useCallback(
    (name: string, value: string) => {
      const params = new URLSearchParams(searchParams.toString());
      if (value && value !== "todos") {
          params.set(name, value);
      } else if (value === "todos" || value === "") {
          params.delete(name);
      }
      
      // Reset page on filter change
      if (name !== "prop_page") {
          params.set("prop_page", "1");
      }
      return params.toString();
    },
    [searchParams]
  );

  const updateUrl = useCallback((name: string, value: string) => {
      router.replace(`${pathname}?${createQueryString(name, value)}`, { scroll: false });
  }, [router, pathname, createQueryString]);

  // Sync debounce
  useEffect(() => {
      const timer = setTimeout(() => {
          if (searchValue !== searchParam) {
              const params = new URLSearchParams(searchParams.toString());
              if (searchValue) params.set("prop_q", searchValue);
              else params.delete("prop_q");
              params.set("prop_page", "1");
              router.replace(`${pathname}?${params.toString()}`, { scroll: false });
          }
      }, 500);
      return () => clearTimeout(timer);
  }, [searchValue, searchParam, pathname, router, searchParams]);

  const siglaParam = sigla !== "todos" ? sigla : "";
  const statusParam = status !== "todos" ? status : "";

  const { data, isLoading } = useProposicoes(id, page, limit, searchParam, undefined, siglaParam, statusParam, sort);

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
      setSearchValue(e.target.value);
  };

  const setPage = (p: number) => updateUrl("prop_page", p.toString());
  const setSigla = (v: string) => updateUrl("prop_type", v);
  const setStatus = (v: string) => updateUrl("prop_status", v);
  const setSort = (v: string) => updateUrl("prop_sort", v);

  const nextPage = () => setPage(page + 1);
  const prevPage = () => setPage(Math.max(1, page - 1));

  if (isLoading) {
    return (
      <div className="space-y-6">
          <div className="flex gap-4">
               <div className="h-10 w-full max-w-sm bg-muted rounded-md animate-pulse" />
               <div className="h-10 w-32 bg-muted rounded-md animate-pulse" />
          </div>
          <div className="space-y-4">
            {[...Array(3)].map((_, i) => (
                <Skeleton key={i} className="h-32" />
            ))}
          </div>
      </div>
    );
  }

  if (!data) return null;

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4">
          <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
              <h3 className="text-lg font-semibold hidden sm:block">Proposições</h3>
              <div className="relative w-full sm:w-72">
                <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  type="search"
                  placeholder="Buscar por ementa ou código..."
                  className="pl-8 pr-8"
                  value={searchValue}
                  onChange={handleSearch}
                />
                {searchValue && (
                    <button
                        type="button"
                        onClick={() => setSearchValue("")}
                        className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                        aria-label="Limpar busca"
                    >
                        <X className="h-4 w-4" />
                    </button>
                )}
              </div>
          </div>
          
          <div className="flex flex-wrap items-center gap-2">
              <Select value={sigla} onValueChange={setSigla}>
                  <SelectTrigger className="w-full sm:w-[180px]">
                      <SelectValue placeholder="Filtrar por Tipo" />
                  </SelectTrigger>
                  <SelectContent>
                      <SelectItem value="todos">Todos os Tipos</SelectItem>
                      <SelectItem value="PEC" title="Proposta de Emenda à Constituição">PEC - Emenda Constitucional</SelectItem>
                      <SelectItem value="PLP" title="Projeto de Lei Complementar">PLP - Lei Complementar</SelectItem>
                      <SelectItem value="PL" title="Projeto de Lei">PL - Projeto de Lei</SelectItem>
                      <SelectItem value="PDL" title="Projeto de Decreto Legislativo">PDL - Decreto Legislativo</SelectItem>
                      <SelectItem value="PRS" title="Projeto de Resolução do Senado">PRS - Resolução do Senado</SelectItem>
                      <SelectItem value="REQ" title="Requerimento">REQ - Requerimento</SelectItem>
                  </SelectContent>
              </Select>

              <Select value={status} onValueChange={setStatus}>
                  <SelectTrigger className="w-full sm:w-[180px]">
                      <SelectValue placeholder="Status" />
                  </SelectTrigger>
                  <SelectContent>
                      <SelectItem value="todos">Todos os Status</SelectItem>
                      <SelectItem value="Apresentado">Em Tramitação</SelectItem>
                      <SelectItem value="EmComissao">Em Comissão</SelectItem>
                      <SelectItem value="AprovadoComissao">Aprovado na Comissão</SelectItem>
                      <SelectItem value="AprovadoPlenario">Aprovado no Plenário</SelectItem>
                      <SelectItem value="TransformadoLei">Transformado em Lei</SelectItem>
                      <SelectItem value="ARQUIVADA">Arquivado</SelectItem>
                      <SelectItem value="PREJUDICADO">Prejudicado</SelectItem>
                  </SelectContent>
              </Select>

              <Select value={sort} onValueChange={setSort}>
                  <SelectTrigger className="w-full sm:w-[180px]">
                      <SelectValue placeholder="Ordenação" />
                  </SelectTrigger>
                  <SelectContent>
                      <SelectItem value="data_desc">Mais recentes</SelectItem>
                      <SelectItem value="data_asc">Mais antigas</SelectItem>
                  </SelectContent>
              </Select>
          </div>
      </div>

      <div className="grid gap-4">
          {data.proposicoes.length === 0 ? (
              <div className="text-center py-12 border rounded-lg bg-muted/10">
                  <p className="text-muted-foreground">Nenhuma proposição encontrada.</p>
              </div>
          ) : (
            data.proposicoes.map((prop) => (
            <Dialog key={prop.id}>
                <DialogTrigger asChild>
                    <Card className="hover:bg-muted/50 transition-colors cursor-pointer group">
                      <CardContent className="p-4 sm:p-6">
                        <div className="flex flex-col gap-2">
                          <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4">
                            <div className="space-y-1.5 flex-1">
                              <div className="flex flex-wrap items-center gap-2">
                                <Badge variant="outline" className="font-mono text-xs">
                                  {prop.sigla_subtipo_materia} {prop.numero_materia}/{prop.ano_materia}
                                </Badge>
                                <span className="text-xs text-muted-foreground">
                                    {prop.data_apresentacao 
                                        ? new Date(prop.data_apresentacao).toLocaleDateString("pt-BR") 
                                        : "Data n/d"}
                                </span>
                              </div>
                              <h3 className="font-semibold leading-tight group-hover:text-primary transition-colors">
                                {prop.ementa}
                              </h3>
                              {prop.descricao_identificacao && (
                                  <p className="text-sm text-muted-foreground line-clamp-2">
                                      {prop.descricao_identificacao}
                                  </p>
                              )}
                            </div>
                            <div className="flex flex-row sm:flex-col items-center sm:items-end justify-between sm:justify-start gap-2 shrink-0">
                                <Badge 
                                    className="whitespace-nowrap"
                                    variant={
                                        prop.situacao_atual?.includes("Aprovad") || prop.situacao_atual?.includes("Lei") 
                                        ? "default" 
                                        : "secondary"
                                    }
                                >
                                    {prop.situacao_atual || prop.estagio_tramitacao}
                                </Badge>
                                <div className="text-[10px] text-muted-foreground font-mono bg-muted px-1.5 py-0.5 rounded">
                                    {prop.codigo_materia}
                                </div>
                            </div>
                          </div>
                          
                          <div className="mt-2 flex items-center gap-4 text-sm text-muted-foreground">
                             {prop.estagio_tramitacao === "TransformadoLei" && (
                                 <div className="flex items-center gap-1.5 text-green-600 font-medium text-xs bg-green-50 px-2 py-1 rounded-full">
                                     <Gavel className="h-3 w-3" /> Lei Sancionada
                                 </div>
                             )}
                             {prop.estagio_tramitacao === "AprovadoPlenario" && (
                                 <div className="flex items-center gap-1.5 text-blue-600 font-medium text-xs bg-blue-50 px-2 py-1 rounded-full">
                                     <CheckCircle2 className="h-3 w-3" /> Aprovado no Plenário
                                 </div>
                             )}
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                </DialogTrigger>
                <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
                    <DialogHeader>
                        <DialogTitle className="leading-snug pr-8">
                             {prop.sigla_subtipo_materia} {prop.numero_materia}/{prop.ano_materia}
                        </DialogTitle>
                        <DialogDescription className="pt-2">
                             Apresentado em {new Date(prop.data_apresentacao || "").toLocaleDateString("pt-BR")}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div className="space-y-2">
                            <h4 className="text-sm font-medium text-muted-foreground">Ementa</h4>
                            <p className="text-base">{prop.ementa}</p>
                        </div>
                         {prop.descricao_identificacao && (
                            <div className="space-y-2">
                                <h4 className="text-sm font-medium text-muted-foreground">Identificação</h4>
                                <p className="text-sm">{prop.descricao_identificacao}</p>
                            </div>
                        )}
                        <div className="grid grid-cols-2 gap-4 pt-4 border-t">
                             <div>
                                 <h4 className="text-xs font-medium text-muted-foreground mb-1">Situação Atual</h4>
                                 <p className="text-sm font-medium">{prop.situacao_atual}</p>
                             </div>
                             <div>
                                 <h4 className="text-xs font-medium text-muted-foreground mb-1">Estágio</h4>
                                 <p className="text-sm">{prop.estagio_tramitacao}</p>
                             </div>
                             <div>
                                 <h4 className="text-xs font-medium text-muted-foreground mb-1">Código Matéria</h4>
                                 <p className="text-sm font-mono">{prop.codigo_materia}</p>
                             </div>
                        </div>
                        <div className="pt-4 flex justify-end">
                            <Button asChild variant="outline" size="sm">
                                <a 
                                    href={`https://www25.senado.leg.br/web/atividade/materias/-/materia/${prop.codigo_materia}`} 
                                    target="_blank" 
                                    rel="noopener noreferrer"
                                >
                                    Ver no Site do Senado
                                </a>
                            </Button>
                        </div>
                    </div>
                </DialogContent>
            </Dialog>
          ))
        )}
      </div>
      
      {/* Pagination Controls */}
      <PaginationWithInput 
            currentPage={data.page} 
            totalPages={data.total_pages} 
            onPageChange={setPage} 
            className="border-t pt-4"
      />
    </div>
  );
}
