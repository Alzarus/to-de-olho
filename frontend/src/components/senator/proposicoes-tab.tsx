"use client";

import { PaginationWithInput } from "@/components/ui/pagination-with-input";

import { useProposicoes } from "@/hooks/use-senador";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import type { Proposicao } from "@/types/api";
import { descricaoMateria, nomeTipo, tituloMateria } from "@/lib/materia";
import { DescricaoExpansivel, MarcaCuradoria, TemasChips } from "@/components/materia/materia-info";
import { Skeleton } from "@/components/ui/skeleton";
import { CheckCircle2, Gavel, Loader2, Search, X } from "lucide-react";
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

// "PL 2338/2023": a identificacao da API, senao montada dos campos
const identificacao = (p: Proposicao) =>
  p.descricao_identificacao?.trim() || `${p.sigla_subtipo_materia} ${p.numero_materia}/${p.ano_materia}`;

// Por que a materia nao pontua (espelha Proposicao.AutoriaPrincipal no backend)
const TIPOS_SENADOR = ["SENADOR", "LIDER", "PRESIDENTE_SF"];
function rotuloSemPontos(prop: Proposicao): string | null {
  if (prop.sigla_subtipo_materia === "VET") return "Veto";
  if (prop.posicao_autoria == null) return "Autoria institucional";
  if (prop.tipo_autor && !TIPOS_SENADOR.includes(prop.tipo_autor)) return "Autoria como deputado";
  if (prop.posicao_autoria !== 1) return "Coautoria";
  return null;
}

// Rótulo das siglas no filtro de tipo: separa os dois requerimentos, que
// NOMES_TIPO chama só de "Requerimento"
const ROTULO_SIGLA: Record<string, string> = {
  RQS: "Requerimento ao Plenário",
  REQ: "Requerimento de comissão",
  INS: "Indicação",
  MOC: "Moção",
  PET: "Petição",
  PRN: "Projeto de Resolução do Congresso",
  VET: "Veto",
};
const rotuloSigla = (sigla: string) => ROTULO_SIGLA[sigla] ?? nomeTipo(sigla);

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
  const autoria = searchParams.get("prop_autoria") ?? "todos";
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
  const autoriaParam = autoria !== "todos" ? autoria : "";

  // isFetching: a lista anterior fica na tela (placeholderData) enquanto o
  // filtro novo carrega; sem indicador, parecia que o filtro não funcionava
  const { data, isLoading, isFetching } = useProposicoes(id, page, limit, searchParam, undefined, siglaParam, statusParam, sort, autoriaParam);
  const atualizando = isFetching && !isLoading;

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
      setSearchValue(e.target.value);
  };

  const setPage = (p: number) => updateUrl("prop_page", p.toString());
  const setSigla = (v: string) => updateUrl("prop_type", v);
  const setStatus = (v: string) => updateUrl("prop_status", v);
  const setSort = (v: string) => updateUrl("prop_sort", v);
  const setAutoria = (v: string) => updateUrl("prop_autoria", v);

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

  // Opções de tipo a partir dos dados; a sigla escolhida fica na lista mesmo
  // sem proposições com a autoria atual
  const tipos = [...(data.tipos ?? [])];
  if (siglaParam && !tipos.some((t) => t.tipo === siglaParam)) {
      tipos.push({ tipo: siglaParam, total: 0 });
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4">
          <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
              <h3 className="text-lg font-semibold hidden sm:flex items-center gap-2">
                  Proposições
                  {atualizando && <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" aria-hidden="true" />}
              </h3>
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
              <Select value={autoria} onValueChange={setAutoria}>
                  <SelectTrigger className="w-full sm:w-[220px]" aria-label="Filtrar por autoria">
                      <SelectValue placeholder="Autoria" />
                  </SelectTrigger>
                  <SelectContent>
                      <SelectItem value="todos">Todas as autorias</SelectItem>
                      <SelectItem value="principal">Só autoria principal</SelectItem>
                      <SelectItem value="coautoria">Só coautorias</SelectItem>
                  </SelectContent>
              </Select>

              <Select value={sigla} onValueChange={setSigla}>
                  <SelectTrigger className="w-full sm:w-[240px]" aria-label="Filtrar por tipo">
                      <SelectValue placeholder="Filtrar por Tipo" />
                  </SelectTrigger>
                  <SelectContent>
                      <SelectItem value="todos">Todos os Tipos</SelectItem>
                      {tipos.map((t) => (
                          <SelectItem key={t.tipo} value={t.tipo} title={rotuloSigla(t.tipo)}>
                              {t.tipo} - {rotuloSigla(t.tipo)} ({t.total})
                          </SelectItem>
                      ))}
                  </SelectContent>
              </Select>

              <Select value={status} onValueChange={setStatus}>
                  <SelectTrigger className="w-full sm:w-[180px]" aria-label="Filtrar por situação">
                      <SelectValue placeholder="Status" />
                  </SelectTrigger>
                  <SelectContent>
                      <SelectItem value="todos">Todos os Status</SelectItem>
                      <SelectItem value="Apresentado">Apresentada</SelectItem>
                      <SelectItem value="EmComissao">Em Comissão</SelectItem>
                      <SelectItem value="AprovadoComissao">Aprovado na Comissão</SelectItem>
                      <SelectItem value="AprovadoPlenario">Aprovado no Plenário</SelectItem>
                      <SelectItem value="TransformadoLei">Transformado em Lei</SelectItem>
                      <SelectItem value="PREJUDICADO">Prejudicado</SelectItem>
                      <SelectItem value="RETIRADO_PELO_AUTOR">Retirado pelo autor</SelectItem>
                  </SelectContent>
              </Select>

              <Select value={sort} onValueChange={setSort}>
                  <SelectTrigger className="w-full sm:w-[180px]" aria-label="Ordenação">
                      <SelectValue placeholder="Ordenação" />
                  </SelectTrigger>
                  <SelectContent>
                      <SelectItem value="data_desc">Mais recentes</SelectItem>
                      <SelectItem value="data_asc">Mais antigas</SelectItem>
                  </SelectContent>
              </Select>
          </div>
      </div>

      <div
          className={`grid gap-4 transition-opacity ${atualizando ? "pointer-events-none opacity-50" : ""}`}
          aria-busy={atualizando}
      >
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
                                {rotuloSemPontos(prop) && (
                                  <Badge
                                    variant="secondary"
                                    className="text-xs"
                                    title="Não pontua no ranking: só conta o primeiro autor, como senador"
                                  >
                                    {rotuloSemPontos(prop)}
                                  </Badge>
                                )}
                                <span className="text-xs text-muted-foreground">
                                    {prop.data_apresentacao 
                                        ? new Date(prop.data_apresentacao).toLocaleDateString("pt-BR") 
                                        : "Data n/d"}
                                </span>
                              </div>
                              <h3 className="font-semibold leading-tight group-hover:text-primary transition-colors">
                                {tituloMateria(prop, identificacao(prop))}
                              </h3>
                              {prop.apelido_fonte === "curadoria" && <MarcaCuradoria />}
                              <DescricaoExpansivel texto={descricaoMateria(prop, prop.ementa)} />
                              <TemasChips temas={prop.temas} />
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
                             {tituloMateria(prop, identificacao(prop))}
                        </DialogTitle>
                        <div>
                            <Badge variant="outline" className="font-mono text-xs">{identificacao(prop)}</Badge>
                        </div>
                        {prop.apelido_fonte === "curadoria" && (
                            <MarcaCuradoria fonteUrl={prop.apelido_fonte_url} comLink />
                        )}
                        <DialogDescription className="pt-2">
                             Apresentado em {new Date(prop.data_apresentacao || "").toLocaleDateString("pt-BR")}
                             {prop.autoria && <><br />Autoria: {prop.autoria}</>}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        {prop.explicacao_ementa && (
                            <div className="space-y-2">
                                <h4 className="text-sm font-medium text-muted-foreground">Explicação da ementa (Senado)</h4>
                                <p className="text-base">{prop.explicacao_ementa}</p>
                            </div>
                        )}
                        <div className="space-y-2">
                            <h4 className="text-sm font-medium text-muted-foreground">Ementa</h4>
                            <p className="text-base">{prop.ementa}</p>
                        </div>
                        {prop.temas && prop.temas.length > 0 && (
                            <div className="space-y-2">
                                <h4 className="text-sm font-medium text-muted-foreground">Temas</h4>
                                <TemasChips temas={prop.temas} max={10} />
                            </div>
                        )}
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
