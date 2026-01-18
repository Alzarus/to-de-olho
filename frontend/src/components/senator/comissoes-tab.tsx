"use client";

import { PaginationWithInput } from "@/components/ui/pagination-with-input";

import { useComissoes } from "@/hooks/use-senador";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { CalendarDays, Users, Search, ChevronLeft, ChevronRight } from "lucide-react";
import { useRouter, usePathname, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";

export function ComissoesTab({ id }: { id: number }) {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  // URL State
  const page = Number(searchParams.get("com_page") ?? "1");
  const searchParam = searchParams.get("com_q") ?? "";
  const status = searchParams.get("com_status") ?? "todos";
  const participacao = searchParams.get("com_part") ?? "todos";
  const limit = 20;

  const [searchValue, setSearchValue] = useState(searchParam);

  const createQueryString = useCallback(
    (name: string, value: string) => {
      const params = new URLSearchParams(searchParams.toString());
      if (value && value !== "todos") {
          params.set(name, value);
      } else if (value === "todos" || value === "") {
          params.delete(name);
      }
      
      if (name !== "com_page") {
          params.set("com_page", "1");
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
            if (searchValue) params.set("com_q", searchValue);
            else params.delete("com_q");
            params.set("com_page", "1");
            router.replace(`${pathname}?${params.toString()}`, { scroll: false });
        }
    }, 500);
    return () => clearTimeout(timer);
  }, [searchValue, searchParam, pathname, router, searchParams]);
  
  const statusParam = status === "todos" ? "" : status;
  const participacaoParam = participacao === "todos" ? "" : participacao;

  const { data, isLoading } = useComissoes(id, page, limit, searchParam, statusParam, participacaoParam);

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
      setSearchValue(e.target.value);
  };

  const setPage = (p: number) => updateUrl("com_page", p.toString());
  const setStatus = (v: string) => updateUrl("com_status", v);
  const setParticipacao = (v: string) => updateUrl("com_part", v);

  const nextPage = () => setPage(page + 1);
  const prevPage = () => setPage(Math.max(1, page - 1));

  if (isLoading) {
    return <Skeleton className="h-[200px] w-full" />;
  }

  if (!data) return null;

  const formatDate = (dateStr?: string) => {
      if (!dateStr) return "Atual";
      return new Date(dateStr).toLocaleDateString("pt-BR");
  };

  return (
    <div className="space-y-6">
        <div className="flex flex-col gap-4">
            <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
                <h3 className="text-lg font-semibold hidden sm:block">Participação em Comissões</h3>
                <div className="relative w-full sm:w-72">
                    <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                    <Input 
                        placeholder="Buscar comissão..." 
                        className="pl-8"
                        value={searchValue}
                        onChange={handleSearch}
                    />
                </div>
            </div>
            
            <div className="flex flex-wrap items-center gap-2">
                <Select value={status} onValueChange={setStatus}>
                    <SelectTrigger className="w-full sm:w-[140px]">
                        <SelectValue placeholder="Status" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="todos">Todos</SelectItem>
                        <SelectItem value="ativa">Ativas</SelectItem>
                        <SelectItem value="inativa">Encerradas</SelectItem>
                    </SelectContent>
                </Select>

                <Select value={participacao} onValueChange={setParticipacao}>
                    <SelectTrigger className="w-full sm:w-[140px]">
                        <SelectValue placeholder="Participação" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="todos">Todos</SelectItem>
                        <SelectItem value="Titular">Titular</SelectItem>
                        <SelectItem value="Suplente">Suplente</SelectItem>
                    </SelectContent>
                </Select>
            </div>
        </div>

      <div className="grid gap-4 sm:grid-cols-1 lg:grid-cols-2">
        {data.comissoes.map((comissao) => (
          <Card key={comissao.id} className="hover:bg-muted/50 transition-colors">
            <CardContent className="p-4 sm:p-6">
              <div className="flex flex-col gap-4">
                <div className="flex items-start justify-between gap-4">
                    <div>
                        <div className="flex items-center gap-2 mb-1">
                            <Badge variant="outline">{comissao.sigla_casa_comissao}</Badge>
                            <h3 className="font-semibold">{comissao.sigla_comissao}</h3>
                        </div>
                        <p className="text-sm text-muted-foreground line-clamp-2">
                            {comissao.nome_comissao}
                        </p>
                    </div>
                    <Badge variant={comissao.data_fim ? "secondary" : "default"}>
                        {comissao.descricao_participacao}
                    </Badge>
                </div>

                <div className="flex items-center gap-4 text-xs text-muted-foreground">
                    <div className="flex items-center gap-1">
                        <CalendarDays className="h-3 w-3" />
                        {formatDate(comissao.data_inicio)} - {formatDate(comissao.data_fim)}
                    </div>
                    {!comissao.data_fim && (
                        <div className="flex items-center gap-1 text-green-600">
                             <Users className="h-3 w-3" /> Ativa
                        </div>
                    )}
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
        {data.comissoes.length === 0 && (
             <p className="col-span-full text-center py-8 text-muted-foreground">
                 Nenhuma participação em comissão registrada.
             </p>
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
