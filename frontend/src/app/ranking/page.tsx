"use client";

import { useState, useMemo, Suspense } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import {
  Search,
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  X,
  ChevronDown,
  ChevronUp,
} from "lucide-react";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { useRanking } from "@/hooks/use-ranking";
import { usePersistentYear } from "@/hooks/use-persistent-year";
import type { SenadorScore } from "@/types/api";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { BRAZIL_STATES } from "@/components/ui/brazil-map-data";

const UF_MAP = Object.fromEntries(
  BRAZIL_STATES.map((state) => [state.id, state.name]),
);

const UFS = [
  "AC",
  "AL",
  "AP",
  "AM",
  "BA",
  "CE",
  "DF",
  "ES",
  "GO",
  "MA",
  "MT",
  "MS",
  "MG",
  "PA",
  "PB",
  "PR",
  "PE",
  "PI",
  "RJ",
  "RN",
  "RS",
  "RO",
  "RR",
  "SC",
  "SP",
  "SE",
  "TO",
];

// Componente de card expandível para mobile
function MobileRankingCard({
  senador,
  index,
}: {
  senador: SenadorScore;
  index: number;
}) {
  const [expanded, setExpanded] = useState(false);

  return (
    <div
      className="border-b border-border p-4 transition-colors hover:bg-muted/50"
      role="article"
      aria-label={`Senador ${senador.nome}, posição ${index + 1}`}
    >
      <div className="flex items-center gap-3">
        <Link
          href={`/senador/${senador.senador_id}`}
          className="flex min-w-0 flex-1 items-center gap-3"
        >
          <div
            className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-sm font-bold ${
              index === 0
                ? "bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400"
                : index === 1
                  ? "bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-300"
                  : index === 2
                    ? "bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400"
                    : "bg-muted text-muted-foreground"
            }`}
            aria-label={`Posição ${index + 1}`}
          >
            {index + 1}
          </div>
          {senador.foto_url ? (
            /* eslint-disable-next-line @next/next/no-img-element */
            <img
              src={senador.foto_url}
              alt=""
              aria-hidden="true"
              className="h-10 w-10 shrink-0 rounded-full object-cover"
            />
          ) : (
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary/10 text-primary">
              <span className="text-sm font-medium" aria-hidden="true">
                {senador.nome.charAt(0)}
              </span>
            </div>
          )}
          <div className="flex min-w-0 flex-col py-0.5">
            <p className="truncate font-medium leading-normal text-foreground">
              {senador.nome}
            </p>
            <div className="flex flex-wrap items-center gap-1.5 pt-0.5">
              <Badge
                variant="secondary"
                className="max-w-fit px-1.5 py-0 text-[10px] uppercase leading-relaxed"
              >
                {senador.partido}
              </Badge>
              <TooltipProvider delayDuration={0}>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <span className="text-xs text-muted-foreground border-b border-dotted cursor-help">
                      {senador.uf}
                    </span>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p className="text-xs font-medium">
                      {UF_MAP[senador.uf] || senador.uf}
                    </p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
          </div>
        </Link>
        <div className="flex shrink-0 items-center gap-1.5">
          <span
            className="text-lg font-bold text-primary"
            aria-label={`Score total: ${senador.score_final.toFixed(1)}`}
          >
            {senador.score_final.toFixed(1)}
          </span>
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setExpanded(!expanded)}
            aria-expanded={expanded}
            aria-label={expanded ? "Recolher detalhes" : "Expandir detalhes"}
            className="h-8 w-8"
          >
            {expanded ? (
              <ChevronUp className="h-4 w-4" />
            ) : (
              <ChevronDown className="h-4 w-4" />
            )}
          </Button>
        </div>
      </div>

      {expanded && (
        <div
          className="mt-4 grid grid-cols-2 gap-3 pt-3 border-t border-border"
          role="list"
          aria-label="Detalhes do score"
        >
          <div className="text-center p-2 rounded bg-muted/50" role="listitem">
            <p className="text-xs text-muted-foreground">Produtividade</p>
            <p className="text-sm font-semibold">
              {senador.produtividade.toFixed(1)}
            </p>
          </div>
          <div className="text-center p-2 rounded bg-muted/50" role="listitem">
            <p className="text-xs text-muted-foreground">Presença</p>
            <p className="text-sm font-semibold">
              {senador.presenca.toFixed(1)}
            </p>
          </div>
          <div className="text-center p-2 rounded bg-muted/50" role="listitem">
            <p className="text-xs text-muted-foreground">Economia</p>
            <p className="text-sm font-semibold">
              {senador.economia_cota.toFixed(1)}
            </p>
          </div>
          <div className="text-center p-2 rounded bg-muted/50" role="listitem">
            <p className="text-xs text-muted-foreground">Comissões</p>
            <p className="text-sm font-semibold">
              {senador.comissoes.toFixed(1)}
            </p>
          </div>
        </div>
      )}
    </div>
  );
}

// Cabeçalho de tabela clicável para ordenação
function SortableHeader({
  label,
  sortKey,
  currentSort,
  sortDir,
  onSort,
  className = "",
}: {
  label: string;
  sortKey: string;
  currentSort: string;
  sortDir: string;
  onSort: (key: string) => void;
  className?: string;
}) {
  const isActive = currentSort === sortKey;

  return (
    <th
      className={`px-4 py-3 text-sm font-medium text-muted-foreground cursor-pointer hover:text-foreground transition-colors select-none ${className}`}
      onClick={() => onSort(sortKey)}
      role="columnheader"
      aria-sort={
        isActive ? (sortDir === "desc" ? "descending" : "ascending") : "none"
      }
      tabIndex={0}
      onKeyDown={(e) => e.key === "Enter" && onSort(sortKey)}
    >
      <span className="inline-flex items-center gap-1">
        {label}
        {isActive ? (
          sortDir === "desc" ? (
            <ArrowDown className="h-3 w-3" />
          ) : (
            <ArrowUp className="h-3 w-3" />
          )
        ) : (
          <ArrowUpDown className="h-3 w-3 opacity-40" />
        )}
      </span>
    </th>
  );
}

function RankingTable({
  data,
  sortBy,
  sortDir,
  onSort,
}: {
  data: SenadorScore[];
  sortBy: string;
  sortDir: string;
  onSort: (key: string) => void;
}) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full" role="table" aria-label="Ranking de senadores">
        <thead>
          <tr className="border-b border-border">
            <th
              className="px-4 py-3 text-left text-sm font-medium text-muted-foreground"
              scope="col"
            >
              Posição
            </th>
            <th
              className="px-4 py-3 text-left text-sm font-medium text-muted-foreground"
              scope="col"
            >
              Senador
            </th>
            <SortableHeader
              label="Produtividade"
              sortKey="produtividade"
              currentSort={sortBy}
              sortDir={sortDir}
              onSort={onSort}
              className="hidden text-center sm:table-cell"
            />
            <SortableHeader
              label="Presença"
              sortKey="presenca"
              currentSort={sortBy}
              sortDir={sortDir}
              onSort={onSort}
              className="hidden text-center md:table-cell"
            />
            <SortableHeader
              label="Economia"
              sortKey="economia_cota"
              currentSort={sortBy}
              sortDir={sortDir}
              onSort={onSort}
              className="hidden text-center lg:table-cell"
            />
            <SortableHeader
              label="Comissões"
              sortKey="comissoes"
              currentSort={sortBy}
              sortDir={sortDir}
              onSort={onSort}
              className="hidden text-center lg:table-cell"
            />
            <SortableHeader
              label="Score Total"
              sortKey="score_final"
              currentSort={sortBy}
              sortDir={sortDir}
              onSort={onSort}
              className="text-right"
            />
          </tr>
        </thead>
        <tbody>
          {data.map((senador, index) => (
            <tr
              key={senador.senador_id}
              className="border-b border-border transition-colors hover:bg-muted/50"
            >
              <td className="px-4 py-4">
                <div
                  className={`inline-flex h-8 w-8 items-center justify-center rounded-full text-sm font-bold ${
                    index === 0
                      ? "bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400"
                      : index === 1
                        ? "bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-300"
                        : index === 2
                          ? "bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400"
                          : "bg-muted text-muted-foreground"
                  }`}
                >
                  {index + 1}
                </div>
              </td>
              <td className="px-4 py-4">
                <Link
                  href={`/senador/${senador.senador_id}`}
                  className="group flex items-center gap-3"
                >
                  {senador.foto_url ? (
                    /* eslint-disable-next-line @next/next/no-img-element */
                    <img
                      src={senador.foto_url}
                      alt=""
                      className="h-10 w-10 rounded-full object-cover"
                    />
                  ) : (
                    <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary/10 text-primary">
                      <span className="text-sm font-medium">
                        {senador.nome.charAt(0)}
                      </span>
                    </div>
                  )}
                  <div>
                    <p className="font-medium text-foreground group-hover:text-primary transition-colors">
                      {senador.nome}
                    </p>
                    <p className="text-sm text-muted-foreground flex items-center gap-1">
                      <Badge variant="secondary" className="mr-1">
                        {senador.partido}
                      </Badge>
                      <TooltipProvider delayDuration={0}>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <span className="border-b border-dotted cursor-help">
                              {senador.uf}
                            </span>
                          </TooltipTrigger>
                          <TooltipContent>
                            <p className="text-xs font-medium">
                              {UF_MAP[senador.uf] || senador.uf}
                            </p>
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    </p>
                  </div>
                </Link>
              </td>
              <td className="hidden px-4 py-4 text-center sm:table-cell">
                <span className="text-sm font-medium">
                  {senador.produtividade.toFixed(1)}
                </span>
              </td>
              <td className="hidden px-4 py-4 text-center md:table-cell">
                <span className="text-sm font-medium">
                  {senador.presenca.toFixed(1)}
                </span>
              </td>
              <td className="hidden px-4 py-4 text-center lg:table-cell">
                <span className="text-sm font-medium">
                  {senador.economia_cota.toFixed(1)}
                </span>
              </td>
              <td className="hidden px-4 py-4 text-center lg:table-cell">
                <span className="text-sm font-medium">
                  {senador.comissoes.toFixed(1)}
                </span>
              </td>
              <td className="px-4 py-4 text-right">
                <span className="text-lg font-bold text-primary">
                  {senador.score_final.toFixed(1)}
                </span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function RankingTableSkeleton() {
  return (
    <div className="space-y-4 p-4">
      {[...Array(10)].map((_, i) => (
        <div key={i} className="flex items-center gap-4">
          <Skeleton className="h-8 w-8 rounded-full" />
          <Skeleton className="h-10 w-10 rounded-full" />
          <div className="flex-1 space-y-2">
            <Skeleton className="h-4 w-48" />
            <Skeleton className="h-3 w-24" />
          </div>
          <Skeleton className="h-6 w-16" />
        </div>
      ))}
    </div>
  );
}

function RankingError({ message }: { message: string }) {
  return (
    <div
      className="flex flex-col items-center justify-center py-12 text-center"
      role="alert"
    >
      <div className="rounded-full bg-destructive/10 p-4">
        <svg
          className="h-8 w-8 text-destructive"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
          />
        </svg>
      </div>
      <h3 className="mt-4 text-lg font-semibold text-foreground">
        Erro ao carregar dados
      </h3>
      <p className="mt-2 text-sm text-muted-foreground max-w-md">{message}</p>
      <p className="mt-4 text-xs text-muted-foreground">
        Tente recarregar a página em alguns instantes.
      </p>
    </div>
  );
}

function RankingContent() {
  const router = useRouter();
  const searchParams = useSearchParams();

  // Ler filtros da URL
  const anoParam = searchParams.get("ano");
  const ano = anoParam ? Number(anoParam) : 0;
  const partido = searchParams.get("partido") || "";
  const uf = searchParams.get("uf") || "";
  const sortBy = searchParams.get("ordenar") || "score_final";
  const sortDir = searchParams.get("direcao") || "desc";
  const search = searchParams.get("busca") || "";
  const [localSearch, setLocalSearch] = useState(search);

  // Persist year selection
  usePersistentYear("ranking");

  const updateUrl = (newParams: Record<string, string | number | null>) => {
    const params = new URLSearchParams(searchParams.toString());
    Object.entries(newParams).forEach(([key, value]) => {
      if (value === null || value === "") {
        params.delete(key);
      } else {
        params.set(key, String(value));
      }
    });
    router.push(`/ranking?${params.toString()}`);
  };

  const handleSort = (key: string) => {
    if (sortBy === key) {
      updateUrl({ direcao: sortDir === "desc" ? "asc" : "desc" });
    } else {
      updateUrl({ ordenar: key, direcao: "desc" });
    }
  };

  const { data, isLoading, error } = useRanking(
    undefined,
    ano === 0 ? undefined : ano,
  );

  // Extrair lista de partidos dos dados
  const partidos = useMemo(() => {
    if (!data?.ranking) return [];
    const unique = [...new Set(data.ranking.map((s) => s.partido))];
    return unique.sort();
  }, [data]);

  // Aplicar filtros e ordenação client-side
  const filteredData = useMemo(() => {
    if (!data?.ranking) return [];

    let result = [...data.ranking];

    if (partido) {
      result = result.filter((s) => s.partido === partido);
    }

    if (uf) {
      result = result.filter((s) => s.uf === uf);
    }

    if (search) {
      const searchLower = search.toLowerCase();
      result = result.filter((s) => s.nome.toLowerCase().includes(searchLower));
    }

    const sortKey = sortBy as keyof SenadorScore;
    result.sort((a, b) => {
      const aVal = a[sortKey] as number;
      const bVal = b[sortKey] as number;
      return sortDir === "desc" ? bVal - aVal : aVal - bVal;
    });

    return result;
  }, [data, partido, uf, search, sortBy, sortDir]);

  const clearFilters = () => {
    router.push("/ranking");
    setLocalSearch("");
  };

  const hasActiveFilters = partido || uf || search;

  return (
    <div className="container mx-auto max-w-7xl px-4 py-8 sm:py-12 sm:px-6 lg:px-8">
      {/* Header */}
      <header className="mb-6 sm:mb-8 sm:flex sm:items-center sm:justify-between">
        <div className="mb-4 sm:mb-0">
          <h1 className="text-2xl font-bold tracking-tight text-foreground sm:text-3xl lg:text-4xl">
            Ranking de Senadores
          </h1>
          <p className="mt-1 text-base text-muted-foreground sm:mt-2 sm:text-lg">
            Avaliação objetiva baseada em produtividade, presença, economia e
            participação.
          </p>
        </div>

        {/* Seletor de Ano */}
        <div className="flex items-center gap-2">
          <label
            htmlFor="ano-select"
            className="text-sm font-medium text-muted-foreground"
          >
            Ano:
          </label>
          <Select
            value={ano.toString()}
            onValueChange={(value) => updateUrl({ ano: Number(value) })}
          >
            <SelectTrigger className="w-full sm:w-[180px]">
              <SelectValue placeholder="Selecione o ano" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="0">Mandato Completo</SelectItem>
              <SelectItem value="2026">2026</SelectItem>
              <SelectItem value="2025">2025</SelectItem>
              <SelectItem value="2024">2024</SelectItem>
              <SelectItem value="2023">2023</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </header>

      {/* Cards de critérios - apenas desktop */}
      <div
        className="mb-6 hidden gap-4 sm:grid sm:grid-cols-2 lg:grid-cols-4"
        role="region"
        aria-label="Critérios de avaliação"
      >
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Produtividade
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-primary">35%</p>
            <p className="mt-1 text-xs text-muted-foreground">
              Proposições apresentadas e aprovadas
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Presença
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-primary">25%</p>
            <p className="mt-1 text-xs text-muted-foreground">
              Participação em votações
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Economia
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-primary">20%</p>
            <p className="mt-1 text-xs text-muted-foreground">
              Uso responsável do CEAPS
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Comissões
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-primary">20%</p>
            <p className="mt-1 text-xs text-muted-foreground">
              Atuação em comissões
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Link mobile para metodologia */}
      <div className="mb-4 sm:hidden">
        <Link
          href="/metodologia"
          className="text-sm text-primary hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring rounded"
        >
          Ver critérios de avaliação (Produtividade 35%, Presença 25%, Economia
          20%, Comissões 20%)
        </Link>
      </div>

      {/* Tabela/Cards de Ranking */}
      <Card>
        <CardHeader className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-2">
            <CardTitle>
              Classificação Geral - {ano === 0 ? "Mandato Completo" : ano}
            </CardTitle>
            <Badge variant="outline" className="font-normal">
              {filteredData.length} senadores
            </Badge>
          </div>
        </CardHeader>

        {/* Barra de Filtros - dentro do card, perto da tabela */}
        <div className="border-t border-b border-border bg-muted/30 px-4 py-3">
          <div className="flex flex-wrap items-center gap-2 sm:gap-3">
            {/* Busca */}
            <div className="relative flex-1 min-w-[150px] sm:min-w-[200px] sm:max-w-xs">
              <Search
                className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground"
                aria-hidden="true"
              />
              <Input
                placeholder="Buscar por nome..."
                className="pl-9 h-9"
                value={localSearch}
                onChange={(e) => {
                  setLocalSearch(e.target.value);
                  updateUrl({ busca: e.target.value || null });
                }}
                aria-label="Buscar senador por nome"
              />
            </div>

            {/* Partido */}
            <Select
              value={partido}
              onValueChange={(value) =>
                updateUrl({ partido: value === "all" ? null : value })
              }
            >
              <SelectTrigger className="h-9 w-full sm:w-[180px]">
                <SelectValue placeholder="Todos os Partidos" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Todos os Partidos</SelectItem>
                {partidos.map((p) => (
                  <SelectItem key={p} value={p}>
                    {p}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            {/* UF */}
            <Select
              value={uf}
              onValueChange={(value) =>
                updateUrl({ uf: value === "all" ? null : value })
              }
            >
              <SelectTrigger className="h-9 w-full sm:w-[140px]">
                <SelectValue placeholder="Todas as UFs" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Todas as UFs</SelectItem>
                {UFS.map((u) => (
                  <SelectItem key={u} value={u}>
                    {u}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            {/* Limpar filtros */}
            {hasActiveFilters && (
              <Button
                variant="ghost"
                size="sm"
                onClick={clearFilters}
                className="h-9 text-muted-foreground"
              >
                <X className="mr-1 h-4 w-4" aria-hidden="true" />
                Limpar
              </Button>
            )}
          </div>
        </div>

        <CardContent className="p-0">
          {isLoading && <RankingTableSkeleton />}
          {error && (
            <RankingError
              message={
                error instanceof Error
                  ? error.message
                  : "Erro desconhecido ao carregar ranking"
              }
            />
          )}
          {data && filteredData.length === 0 && !isLoading && (
            <div className="py-12 text-center text-muted-foreground">
              Nenhum senador encontrado com os filtros selecionados.
            </div>
          )}

          {/* Desktop: Tabela */}
          {data && filteredData.length > 0 && (
            <div className="hidden sm:block">
              <RankingTable
                data={filteredData}
                sortBy={sortBy}
                sortDir={sortDir}
                onSort={handleSort}
              />
            </div>
          )}

          {/* Mobile: Cards expansíveis */}
          {data && filteredData.length > 0 && (
            <div
              className="sm:hidden"
              role="list"
              aria-label="Lista de senadores"
            >
              {filteredData.map((senador, index) => (
                <MobileRankingCard
                  key={senador.senador_id}
                  senador={senador}
                  index={index}
                />
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Link para metodologia */}
      <div className="mt-6 text-center">
        <p className="text-sm text-muted-foreground">
          Quer entender como calculamos os scores?{" "}
          <Link
            href="/metodologia"
            className="font-medium text-primary hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring rounded"
          >
            Consulte nossa metodologia
          </Link>
        </p>
      </div>
    </div>
  );
}

export default function RankingPage() {
  return (
    <Suspense
      fallback={
        <div className="container py-12 text-center">Carregando...</div>
      }
    >
      <RankingContent />
    </Suspense>
  );
}
