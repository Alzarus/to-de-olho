"use client";

import { useState, useEffect, Suspense } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { format } from "date-fns";
import { ptBR } from "date-fns/locale";
import { Search, ChevronLeft, ChevronRight } from "lucide-react";

import { Card, CardContent } from "@/components/ui/card";
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

// Componente interno para envelopar no Suspense
function VotacoesContent() {
  const router = useRouter();
  const searchParams = useSearchParams();

  // Ler estado da URL (ou defaults)
  const page = Number(searchParams.get("page")) || 1;
  const currentYear = new Date().getFullYear();
  const anoParam = searchParams.get("ano");
  const ano = anoParam ? Number(anoParam) : currentYear;
  const search = searchParams.get("search") || "";

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
        const res = await getVotacoes(page, limit, ano, search);
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

  const totalPages = Math.ceil(total / limit);

  return (
    <div className="container mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      {/* Header */}
      <div className="mb-8 sm:flex sm:items-center sm:justify-between">
        <div className="mb-4 sm:mb-0">
          <h1 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
            Votações Nominais
          </h1>
          <p className="mt-2 text-lg text-muted-foreground max-w-3xl">
            Acompanhe como votam os senadores nas principais matérias legislativas.
          </p>
        </div>

        {/* Year Selector */}
        <div className="flex items-center gap-2">
            <label
                htmlFor="ano-select"
                className="text-sm font-medium text-muted-foreground whitespace-nowrap"
            >
                Ano:
            </label>
            <select
                id="ano-select"
                value={ano}
                onChange={(e) => updateUrl({ ano: Number(e.target.value), page: 1 })}
                className="h-10 rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            >
                {[2026, 2025, 2024, 2023].map((y) => (
                    <option key={y} value={y}>
                        {y}
                    </option>
                ))}
            </select>
        </div>
      </div>

       {/* Stats - Visual Balance */}
       <div className="mb-8 grid gap-4 sm:grid-cols-3">
          <Card>
             <CardContent className="p-6">
                <div className="flex flex-col gap-1">
                   <span className="text-sm font-medium text-muted-foreground">Total Votações (Ano)</span>
                   <span className="text-2xl font-bold">{total}</span>
                </div>
             </CardContent>
          </Card>
          <Card className="sm:col-span-2">
             <CardContent className="p-6 flex items-end h-full">
                 <div className="w-full">
                    <div className="relative w-full">
                        <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                        <Input
                            placeholder="Buscar por matéria (PEC, PL...) ou descrição..."
                            className="pl-9 w-full"
                            value={localSearch}
                            onChange={(e) => setLocalSearch(e.target.value)}
                        />
                    </div>
                 </div>
             </CardContent>
          </Card>
       </div>

      {/* Table */}
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[100px]">Data</TableHead>
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
              ) : data.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={3} className="h-24 text-center">
                    Nenhuma votação encontrada.
                  </TableCell>
                </TableRow>
              ) : (
                data.map((votacao) => (
                  <TableRow 
                    key={votacao.sessao_id} 
                    className="hover:bg-muted/50 cursor-pointer group"
                    onClick={() => router.push(`/votacoes/${votacao.sessao_id}?backUrl=${encodeURIComponent(`/votacoes?${searchParams.toString()}`)}`)}
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
                      <div className="flex flex-col gap-1 max-w-[200px] sm:max-w-[400px] md:max-w-none">
                        {votacao.materia && (
                          <span className="font-semibold text-primary truncate block group-hover:text-primary/80 transition-colors">
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

      {/* Pagination */}
      {!loading && totalPages > 1 && (
        <div className="mt-8 flex items-center justify-center gap-2">
          <Button
            variant="outline"
            size="icon"
            onClick={() => updateUrl({ page: Math.max(1, page - 1) })}
            disabled={page === 1}
          >
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <span className="text-sm font-medium">
            Página {page} de {totalPages}
          </span>
          <Button
            variant="outline"
            size="icon"
            onClick={() => updateUrl({ page: Math.min(totalPages, page + 1) })}
            disabled={page === totalPages}
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      )}
    </div>
  );
}

export default function VotacoesPage() {
  return (
    <Suspense fallback={<div className="container py-12 text-center">Carregando filtros...</div>}>
      <VotacoesContent />
    </Suspense>
  );
}
