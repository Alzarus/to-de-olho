"use client";

import { useState, useEffect, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Search, ChevronLeft, ChevronRight, ArrowUp, ArrowDown } from "lucide-react";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

import { getVotacoes, Votacao } from "@/services/votacaoService";

function VotacoesContent() {
  const router = useRouter();
  const searchParams = useSearchParams();

  // Ler estado da URL (ou defaults)
  const page = Number(searchParams.get("page")) || 1;
  const anoParam = searchParams.get("ano");
  const ano = anoParam ? Number(anoParam) : null;
  const search = searchParams.get("search") || "";
  const sortDir = searchParams.get("ordem") || "desc";

  const [data, setData] = useState<Votacao[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  
  // Input local para busca (debounce)
  const [localSearch, setLocalSearch] = useState(search);
  const limit = 20;

  // Atualizar URL helper
  const updateUrl = (newParams: Record<string, string | number | null>) => {
    const params = new URLSearchParams(searchParams.toString());
    Object.entries(newParams).forEach(([key, value]) => {
      if (value === null || value === "" || value === 0) {
        params.delete(key);
      } else {
        params.set(key, String(value));
      }
    });
    router.push(`/votacoes?${params.toString()}`);
  };

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
  }, [localSearch, search]); // eslint-disable-line react-hooks/exhaustive-deps

  // Fetch Data
  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        const res = await getVotacoes(page, limit, ano ?? undefined, search);
        setData(res.data);
        setTotal(res.total);
      } catch (error) {
        console.error("Failed to fetch votacoes", error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [page, ano, search]);

  // Aplicar ordenação client-side
  const sortedData = [...data].sort((a, b) => {
    const dateA = new Date(a.data).getTime();
    const dateB = new Date(b.data).getTime();
    return sortDir === "desc" ? dateB - dateA : dateA - dateB;
  });

  const totalPages = Math.ceil(total / limit);

  const toggleSort = () => {
    updateUrl({ ordem: sortDir === "desc" ? "asc" : "desc" });
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
            Acompanhe como votam os senadores nas principais matérias legislativas.
          </p>
        </div>

        {/* Seletor de Ano */}
        <div className="flex items-center gap-2">
          <label
            htmlFor="ano-select"
            className="text-sm font-medium text-muted-foreground whitespace-nowrap"
          >
            Ano:
          </label>
          <select
            id="ano-select"
            value={ano ?? ""}
            onChange={(e) => updateUrl({ ano: e.target.value ? Number(e.target.value) : null, page: 1 })}
            className="h-10 rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
          >
            <option value="">Todos</option>
            {[2026, 2025, 2024, 2023].map((y) => (
              <option key={y} value={y}>
                {y}
              </option>
            ))}
          </select>
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
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" aria-hidden="true" />
              <Input
                placeholder="Buscar por matéria (PEC, PL...) ou descrição..."
                className="pl-9 h-9"
                value={localSearch}
                onChange={(e) => setLocalSearch(e.target.value)}
                aria-label="Buscar votação por matéria ou descrição"
              />
            </div>
          </div>
        </div>
        
        <CardContent className="p-0 overflow-x-auto">
          <Table role="table" aria-label="Lista de votações nominais" className="min-w-[600px]">
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
                      <ArrowDown className="h-3 w-3" aria-label="Ordenado por mais recentes" />
                    ) : (
                      <ArrowUp className="h-3 w-3" aria-label="Ordenado por mais antigas" />
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
                    <TableCell><Skeleton className="h-4 w-20" /></TableCell>
                    <TableCell><Skeleton className="h-4 w-24" /></TableCell>
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
                    key={votacao.sessao_id} 
                    className="hover:bg-muted/50 cursor-pointer group"
                    onClick={() => router.push(`/votacoes/${votacao.sessao_id}?backUrl=${encodeURIComponent(`/votacoes?${searchParams.toString()}`)}`)}
                    role="row"
                    tabIndex={0}
                    onKeyDown={(e) => e.key === "Enter" && router.push(`/votacoes/${votacao.sessao_id}`)}
                  >
                    <TableCell className="font-medium whitespace-nowrap">
                       {new Date(votacao.data).getUTCDate().toString().padStart(2, '0')}/
                       {(new Date(votacao.data).getUTCMonth() + 1).toString().padStart(2, '0')}/
                       {new Date(votacao.data).getUTCFullYear()}
                    </TableCell>
                    <TableCell>
                       <Badge variant="outline" className="group-hover:border-primary/50 transition-colors">
                        {votacao.codigo_sessao}
                       </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-col gap-1 min-w-[250px]">
                        {votacao.materia && (
                          <span className="font-semibold text-primary block group-hover:text-primary/80 transition-colors">
                            {votacao.materia}
                          </span>
                        )}
                        <span className="text-sm text-muted-foreground line-clamp-2">
                          {votacao.descricao_votacao}
                        </span>
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
        <nav className="mt-6 flex items-center justify-center gap-2" aria-label="Navegação de páginas">
          <Button
            variant="outline"
            size="icon"
            onClick={() => updateUrl({ page: Math.max(1, page - 1) })}
            disabled={page === 1}
            aria-label="Página anterior"
          >
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <span className="text-sm font-medium" aria-current="page">
            Página {page} de {totalPages}
          </span>
          <Button
            variant="outline"
            size="icon"
            onClick={() => updateUrl({ page: Math.min(totalPages, page + 1) })}
            disabled={page === totalPages}
            aria-label="Próxima página"
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
        </nav>
      )}
    </div>
  );
}

export default function VotacoesPage() {
  return (
    <Suspense fallback={<div className="container py-12 text-center">Carregando...</div>}>
      <VotacoesContent />
    </Suspense>
  );
}
