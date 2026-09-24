"use client";

import { useState, useEffect, Suspense, useCallback } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Search, ArrowUp, ArrowDown, X, Lock } from "lucide-react";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { PaginationWithInput } from "@/components/ui/pagination-with-input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

import {
  getVotacoes,
  getVotacoesFacetas,
  FacetasVotacoes,
  Votacao,
} from "@/services/votacaoService";
import { usePersistentYear } from "@/hooks/use-persistent-year";
import {
  descricaoMateria,
  NOMES_TIPO,
  nomeTipo,
  tituloMateria,
  usaDescricaoVotacao,
} from "@/lib/materia";
import {
  DescricaoExpansivel,
  MarcaCuradoria,
  TemasChips,
} from "@/components/materia/materia-info";


const NOMES_RESULTADO: Record<string, string> = {
  A: "Aprovada",
  R: "Rejeitada",
};

const nomeResultado = (r: string) => NOMES_RESULTADO[r] ?? r;

// Valor neutro dos selects (o Radix não aceita value vazio)
const TODAS = "todas";

function VotacoesContent() {
  const router = useRouter();
  const searchParams = useSearchParams();

  // Ler estado da URL (ou defaults)
  const page = Number(searchParams.get("page")) || 1;
  const anoParam = searchParams.get("ano");
  const ano = anoParam ? Number(anoParam) : 0;
  const search = searchParams.get("search") || "";
  const sortDir = searchParams.get("ordem") || "desc";
  // Votacoes de uma sessao (destino dos links antigos /votacoes/NNNNNN_AAAA)
  const sessao = searchParams.get("sessao") || "";
  // Filtros: tipo=PEC,MSF · secreta=true|false · resultado=A|R
  const tipoParam = searchParams.get("tipo") || "";
  const secretaParam = searchParams.get("secreta") || "";
  const resultado = searchParams.get("resultado") || "";
  const secreta =
    secretaParam === "true"
      ? true
      : secretaParam === "false"
        ? false
        : undefined;
  const filtrosAtivos = Boolean(tipoParam || secretaParam || resultado);

  const [data, setData] = useState<Votacao[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [facetas, setFacetas] = useState<FacetasVotacoes | null>(null);

  // Input local para busca (debounce)
  const [localSearch, setLocalSearch] = useState(search);
  const limit = 20;

  // Persist year
  usePersistentYear("votacoes");

  // Atualizar URL helper - wrapped in useCallback
  const updateUrl = useCallback(
    (newParams: Record<string, string | number | null>) => {
      const params = new URLSearchParams(searchParams.toString());
      Object.entries(newParams).forEach(([key, value]) => {
        if (value === null || value === "") {
          params.delete(key);
        } else {
          params.set(key, String(value));
        }
      });
      router.push(`/votacoes?${params.toString()}`, { scroll: false });
    },
    [searchParams, router],
  );

  // Sync initial localSearch if URL changes externally
  useEffect(() => {
    setLocalSearch(search);
  }, [search]);

  // Handle Search Debounce
  useEffect(() => {
    const timer = setTimeout(() => {
      if (localSearch !== search) {
        updateUrl({ search: localSearch, page: 1 });
      }
    }, 500);
    return () => clearTimeout(timer);
  }, [localSearch, search, updateUrl]);

  const anoConsulta = ano === 0 || sessao ? undefined : ano;

  // Contagens dos filtros (dependem só do ano)
  useEffect(() => {
    let ativo = true;
    getVotacoesFacetas(anoConsulta)
      .then((f) => ativo && setFacetas(f))
      .catch((error) => console.error("Failed to fetch facetas", error));
    return () => {
      ativo = false;
    };
  }, [anoConsulta]);

  // Fetch Data
  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        const res = await getVotacoes(
          page,
          limit,
          anoConsulta,
          search,
          sortDir,
          sessao || undefined,
          {
            tipos: tipoParam ? tipoParam.split(",") : undefined,
            secreta,
            resultado: resultado || undefined,
          },
        );
        setData(res.data);
        setTotal(res.total);
      } catch (error) {
        console.error("Failed to fetch votacoes", error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [
    page,
    anoConsulta,
    search,
    sessao,
    sortDir,
    tipoParam,
    secreta,
    resultado,
  ]);

  // A ordem vem do backend (data, sessão e sequencial dentro do dia). Ordenar
  // aqui só invertia a página atual em vez de buscar as votações mais antigas.
  const sortedData = data;

  const totalPages = Math.ceil(total / limit);

  const tiposSelecionados = tipoParam ? tipoParam.split(",") : [];

  // Tipos com contagem; os marcados continuam visíveis mesmo sem votação no ano
  const opcoesTipo = [
    ...(facetas?.tipos ?? []),
    ...tiposSelecionados
      .filter((t) => !facetas?.tipos.some((f) => f.valor === t))
      .map((t) => ({ valor: t, total: 0 })),
  ];

  const contagem = (
    lista: { valor: string; total: number }[] | undefined,
    valor: string,
  ) => lista?.find((f) => f.valor === valor)?.total ?? 0;

  const alternarTipo = (sigla: string) => {
    const novos = tiposSelecionados.includes(sigla)
      ? tiposSelecionados.filter((t) => t !== sigla)
      : [...tiposSelecionados, sigla];
    updateUrl({ tipo: novos.join(","), page: 1 });
  };

  const limparFiltros = () =>
    updateUrl({ tipo: null, secreta: null, resultado: null, page: 1 });

  const toggleSort = () => {
    updateUrl({ ordem: sortDir === "desc" ? "asc" : "desc", page: 1 });
  };

  return (
    <div className="container mx-auto max-w-7xl px-4 py-8 sm:py-12 sm:px-6 lg:px-8">
      {/* Header */}
      <header className="mb-6 sm:mb-8 sm:flex sm:items-center sm:justify-between">
        <div className="mb-4 sm:mb-0">
          <h1 className="text-2xl font-bold tracking-tight text-foreground sm:text-3xl lg:text-4xl">
            Votações Nominais
          </h1>
          <p className="mt-1 text-base text-muted-foreground sm:mt-2 sm:text-lg">
            Acompanhe como votam os senadores nas principais matérias
            legislativas.
          </p>
          {sessao && (
            <p className="mt-2 text-sm text-muted-foreground" role="status">
              Mostrando as votações da sessão {sessao}.{" "}
              <button
                type="button"
                onClick={() => updateUrl({ sessao: null, page: 1 })}
                className="font-medium text-primary hover:underline"
              >
                Ver todas
              </button>
            </p>
          )}
        </div>

        {/* Seletor de Ano */}
        <div className="flex items-center gap-2">
          <label
            htmlFor="ano-select"
            className="text-sm font-medium text-muted-foreground whitespace-nowrap"
          >
            Ano:
          </label>
          <Select
            value={ano.toString()}
            onValueChange={(value) =>
              updateUrl({ ano: Number(value), page: 1 })
            }
          >
            <SelectTrigger id="ano-select" className="w-full sm:w-[180px]">
              <SelectValue placeholder="Selecione o ano" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="0">Todos</SelectItem>
              {[2026, 2025, 2024, 2023].map((y) => (
                <SelectItem key={y} value={y.toString()}>
                  {y}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </header>

      {/* Tabela */}
      <Card>
        <CardHeader className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-2">
            <CardTitle>Votações do Período</CardTitle>
            <Badge variant="outline" className="font-normal">
              {total.toLocaleString("pt-BR")} votações
            </Badge>
          </div>
        </CardHeader>

        {/* Barra de Filtros - dentro do card, perto da tabela */}
        <div className="border-t border-b border-border bg-muted/30 px-4 py-3">
          <div className="flex flex-wrap items-center gap-2 sm:gap-3">
            {/* Busca */}
            <div className="relative flex-1 min-w-[180px] sm:max-w-md">
              <Search
                className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground"
                aria-hidden="true"
              />
              <Input
                placeholder="Buscar por matéria, descrição ou código da sessão..."
                className="pl-9 pr-8 h-9"
                value={localSearch}
                onChange={(e) => setLocalSearch(e.target.value)}
                aria-label="Buscar votação por matéria, descrição ou código"
              />
              {localSearch && (
                <button
                  type="button"
                  onClick={() => setLocalSearch("")}
                  className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  aria-label="Limpar busca"
                >
                  <X className="h-4 w-4" aria-hidden="true" />
                </button>
              )}
            </div>

            {/* Aberta / secreta */}
            <div className="flex items-center gap-2">
              <label
                htmlFor="secreta-select"
                className="text-sm font-medium text-muted-foreground whitespace-nowrap"
              >
                Votação:
              </label>
              <Select
                value={secretaParam || TODAS}
                onValueChange={(v) =>
                  updateUrl({ secreta: v === TODAS ? null : v, page: 1 })
                }
              >
                <SelectTrigger id="secreta-select" className="h-9 w-[190px]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={TODAS}>
                    Abertas e secretas
                    {facetas
                      ? ` (${facetas.total.toLocaleString("pt-BR")})`
                      : ""}
                  </SelectItem>
                  <SelectItem value="false">
                    Abertas
                    {facetas
                      ? ` (${contagem(facetas.secreta, "false").toLocaleString("pt-BR")})`
                      : ""}
                  </SelectItem>
                  <SelectItem value="true">
                    Secretas
                    {facetas
                      ? ` (${contagem(facetas.secreta, "true").toLocaleString("pt-BR")})`
                      : ""}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Resultado */}
            <div className="flex items-center gap-2">
              <label
                htmlFor="resultado-select"
                className="text-sm font-medium text-muted-foreground whitespace-nowrap"
              >
                Resultado:
              </label>
              <Select
                value={resultado || TODAS}
                onValueChange={(v) =>
                  updateUrl({ resultado: v === TODAS ? null : v, page: 1 })
                }
              >
                <SelectTrigger id="resultado-select" className="h-9 w-[170px]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={TODAS}>
                    Todos
                    {facetas
                      ? ` (${facetas.total.toLocaleString("pt-BR")})`
                      : ""}
                  </SelectItem>
                  {(facetas?.resultados ?? []).map((f) => (
                    <SelectItem key={f.valor} value={f.valor}>
                      {nomeResultado(f.valor)} (
                      {f.total.toLocaleString("pt-BR")})
                    </SelectItem>
                  ))}
                  {resultado &&
                    !facetas?.resultados.some((f) => f.valor === resultado) && (
                      <SelectItem value={resultado}>
                        {nomeResultado(resultado)} (0)
                      </SelectItem>
                    )}
                </SelectContent>
              </Select>
            </div>

            {filtrosAtivos && (
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={limparFiltros}
                className="h-9"
              >
                <X className="h-4 w-4" aria-hidden="true" />
                Limpar filtros
              </Button>
            )}
          </div>

          {/* Tipo de matéria */}
          <fieldset className="mt-3">
            <legend className="mb-2 text-sm font-medium text-muted-foreground">
              Tipo de matéria
              {tiposSelecionados.length > 0 &&
                ` (${tiposSelecionados.length} ${tiposSelecionados.length === 1 ? "selecionado" : "selecionados"})`}
            </legend>
            <div className="flex flex-wrap gap-2">
              {opcoesTipo.length === 0 && (
                <span className="text-sm text-muted-foreground">
                  {facetas ? "Nenhum tipo no período." : "Carregando tipos..."}
                </span>
              )}
              {opcoesTipo.map((f) => {
                const marcado = tiposSelecionados.includes(f.valor);
                return (
                  <label
                    key={f.valor}
                    title={nomeTipo(f.valor)}
                    className={`inline-flex cursor-pointer items-center gap-2 rounded-md border px-2.5 py-1 text-sm transition-colors has-[:focus-visible]:ring-2 has-[:focus-visible]:ring-ring ${
                      marcado
                        ? "border-primary bg-primary/10 text-foreground"
                        : "border-border bg-background text-foreground hover:bg-muted"
                    }`}
                  >
                    <input
                      type="checkbox"
                      className="h-4 w-4 accent-primary"
                      checked={marcado}
                      onChange={() => alternarTipo(f.valor)}
                    />
                    <span>
                      <span className="font-semibold">{f.valor}</span>
                      {NOMES_TIPO[f.valor] && (
                        <span className="text-muted-foreground">
                          {" "}
                          · {NOMES_TIPO[f.valor]}
                        </span>
                      )}{" "}
                      <span className="text-muted-foreground">
                        ({f.total.toLocaleString("pt-BR")})
                      </span>
                    </span>
                  </label>
                );
              })}
            </div>
          </fieldset>

          <p className="mt-3 flex items-start gap-1.5 text-xs text-muted-foreground">
            <Lock className="mt-0.5 h-3.5 w-3.5 shrink-0" aria-hidden="true" />
            Em votação secreta (como a maioria das indicações de autoridades), o
            Senado publica só quem votou, não o voto de cada senador.
          </p>
        </div>

        <CardContent className="p-0 overflow-x-auto">
          <Table
            role="table"
            aria-label="Lista de votações nominais"
            className="min-w-[600px]"
          >
            <TableHeader>
              <TableRow>
                <TableHead
                  className="w-[100px] cursor-pointer hover:text-foreground transition-colors select-none"
                  onClick={toggleSort}
                  role="columnheader"
                  aria-sort={sortDir === "desc" ? "descending" : "ascending"}
                  tabIndex={0}
                  onKeyDown={(e) => e.key === "Enter" && toggleSort()}
                >
                  <span className="inline-flex items-center gap-1">
                    Data
                    {sortDir === "desc" ? (
                      <ArrowDown
                        className="h-3 w-3"
                        aria-label="Ordenado por mais recentes"
                      />
                    ) : (
                      <ArrowUp
                        className="h-3 w-3"
                        aria-label="Ordenado por mais antigas"
                      />
                    )}
                  </span>
                </TableHead>
                <TableHead className="w-[120px]">Sessão</TableHead>
                <TableHead>Matéria / Descrição</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                [...Array(5)].map((_, i) => (
                  <TableRow key={i}>
                    <TableCell>
                      <Skeleton className="h-4 w-20" />
                    </TableCell>
                    <TableCell>
                      <Skeleton className="h-4 w-24" />
                    </TableCell>
                    <TableCell>
                      <Skeleton className="h-4 w-full mb-2" />
                      <Skeleton className="h-3 w-1/2" />
                    </TableCell>
                  </TableRow>
                ))
              ) : sortedData.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={3} className="h-24 text-center">
                    Nenhuma votação encontrada.
                  </TableCell>
                </TableRow>
              ) : (
                sortedData.map((votacao) => (
                  <TableRow
                    key={votacao.codigo_votacao}
                    className="hover:bg-muted/50 cursor-pointer group"
                    onClick={() =>
                      router.push(
                        `/votacoes/${votacao.codigo_votacao}?backUrl=${encodeURIComponent(`/votacoes?${searchParams.toString()}`)}`,
                      )
                    }
                    role="row"
                    tabIndex={0}
                    onKeyDown={(e) =>
                      e.key === "Enter" &&
                      router.push(`/votacoes/${votacao.codigo_votacao}`)
                    }
                  >
                    <TableCell className="font-medium whitespace-nowrap">
                      {new Date(votacao.data)
                        .getUTCDate()
                        .toString()
                        .padStart(2, "0")}
                      /
                      {(new Date(votacao.data).getUTCMonth() + 1)
                        .toString()
                        .padStart(2, "0")}
                      /{new Date(votacao.data).getUTCFullYear()}
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant="outline"
                        className="group-hover:border-primary/50 transition-colors"
                      >
                        {votacao.codigo_sessao}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-col gap-1 min-w-[250px]">
                        {(votacao.sigla_materia ||
                          votacao.secreta ||
                          votacao.resultado) && (
                          <div className="flex flex-wrap items-center gap-1.5">
                            {votacao.sigla_materia && (
                              <Badge
                                variant="secondary"
                                title={nomeTipo(votacao.sigla_materia)}
                              >
                                {votacao.sigla_materia}
                                <span className="sr-only">
                                  {" "}
                                  ({nomeTipo(votacao.sigla_materia)})
                                </span>
                              </Badge>
                            )}
                            {votacao.secreta && (
                              <Badge variant="outline">
                                <Lock aria-hidden="true" />
                                Secreta
                              </Badge>
                            )}
                            {votacao.resultado && (
                              <Badge variant="outline">
                                {nomeResultado(votacao.resultado)}
                              </Badge>
                            )}
                          </div>
                        )}
                        {votacao.materia && (
                          <span className="flex flex-wrap items-center gap-1.5">
                            <span className="font-semibold text-primary group-hover:text-primary/80 transition-colors">
                              {tituloMateria(votacao, votacao.materia)}
                            </span>
                            {votacao.apelido_fonte === "curadoria" && (
                              <MarcaCuradoria />
                            )}
                            <Badge variant="outline" className="font-mono">
                              {votacao.materia}
                            </Badge>
                          </span>
                        )}
                        {usaDescricaoVotacao(votacao.sigla_materia) ? (
                          <span className="text-sm text-muted-foreground line-clamp-2">
                            {votacao.descricao_votacao}
                          </span>
                        ) : (
                          <>
                            <DescricaoExpansivel
                              texto={descricaoMateria(votacao, votacao.ementa)}
                            />
                            {votacao.descricao_votacao && (
                              <span className="text-xs text-muted-foreground line-clamp-1">
                                {votacao.descricao_votacao}
                              </span>
                            )}
                            <TemasChips temas={votacao.temas} />
                          </>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Paginação */}
      {!loading && totalPages > 1 && (
        <PaginationWithInput
          currentPage={page}
          totalPages={totalPages}
          onPageChange={(p) => updateUrl({ page: p })}
          className="mt-6"
        />
      )}
    </div>
  );
}

export default function VotacoesPage() {
  return (
    <Suspense
      fallback={
        <div className="container py-12 text-center">Carregando...</div>
      }
    >
      <VotacoesContent />
    </Suspense>
  );
}
