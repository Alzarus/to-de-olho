"use client";

import { useDespesas } from "@/hooks/use-senador";
import { Skeleton } from "@/components/ui/skeleton";
import { formatCurrency } from "@/lib/utils";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useCallback, useEffect, useState } from "react";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Search, ChevronLeft, ChevronRight, FileText } from "lucide-react";
import { useRouter, usePathname, useSearchParams } from "next/navigation";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";

export function CeapsTab({ id, ano }: { id: number; ano: number }) {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  // URL State
  const page = Number(searchParams.get("ceaps_page") ?? "1");
  const searchParam = searchParams.get("ceaps_q") ?? "";
  const tipo = searchParams.get("ceaps_type") ?? "todos";
  const sort = searchParams.get("ceaps_sort") ?? "data_desc";

  const [searchValue, setSearchValue] = useState(searchParam);

  const createQueryString = useCallback(
    (name: string, value: string) => {
      const params = new URLSearchParams(searchParams.toString());
      if (value && value !== "todos") {
          params.set(name, value);
      } else if (value === "todos" || value === "") {
          params.delete(name);
      }
      
      if (name !== "ceaps_page") {
          params.set("ceaps_page", "1");
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
            if (searchValue) params.set("ceaps_q", searchValue);
            else params.delete("ceaps_q");
            params.set("ceaps_page", "1");
            router.replace(`${pathname}?${params.toString()}`, { scroll: false });
        }
    }, 500);
    return () => clearTimeout(timer);
  }, [searchValue, searchParam, pathname, router, searchParams]);

  // Reset page logic on year change can stay if desired, but URL persistence typically handles it by default params
  useEffect(() => {
     // If ano changed from parent, we might want to reset the page to 1
     if (page !== 1) {
         // This is tricky because "ano" state is external. 
         // Let's assume user wants to reset paging on year switch.
         const params = new URLSearchParams(searchParams.toString());
         params.set("ceaps_page", "1");
         router.replace(`${pathname}?${params.toString()}`, { scroll: false });
     }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ano]);

  const { data, isLoading } = useDespesas(id, ano, page, 20, searchParam, tipo, sort);

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
      setSearchValue(e.target.value);
  };

  const setPage = (p: number) => updateUrl("ceaps_page", p.toString());
  const setTipo = (v: string) => updateUrl("ceaps_type", v);
  const setSort = (v: string) => updateUrl("ceaps_sort", v);

  const nextPage = () => setPage(page + 1);
  const prevPage = () => setPage(Math.max(1, page - 1));


  if (isLoading) {
    return <Skeleton className="h-[400px] w-full" />;
  }

  if (!data || !data.despesas) return null;

  const formatDate = (dateStr?: string) => {
      if (!dateStr) return "-";
      const date = new Date(dateStr);
      if (isNaN(date.getTime())) return "-";
      return date.toLocaleDateString("pt-BR");
  };

  return (
    <div className="space-y-6">
        <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
            <h3 className="text-lg font-semibold hidden sm:block">Detalhamento de Despesas</h3>
            <div className="relative w-full sm:w-72">
                <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input 
                    placeholder="Filtrar por fornecedor ou tipo..." 
                    className="pl-8"
                    value={searchValue}
                    onChange={handleSearch}
                />
            </div>
        </div>

        <div className="rounded-md border overflow-hidden">
          <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[100px]">Data</TableHead>
                    <TableHead>Fornecedor</TableHead>
                    <TableHead className="w-[200px]">Tipo</TableHead>
                    <TableHead className="text-right">Valor</TableHead>
                    <TableHead className="w-[50px]"></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data.despesas.length === 0 ? (
                      <TableRow>
                          <TableCell colSpan={5} className="text-center py-8 text-muted-foreground">
                              Nenhuma despesa encontrada.
                          </TableCell>
                      </TableRow>
                  ) : (
                      data.despesas.map((despesa, i) => (
                        <Dialog key={`${i}-${despesa.data_documento}`}>
                            <DialogTrigger asChild>
                                <TableRow className="cursor-pointer hover:bg-muted/50 group">
                                  <TableCell className="font-mono text-xs whitespace-nowrap">
                                      {formatDate(despesa.data_documento)}
                                  </TableCell>
                                  <TableCell className="max-w-[150px] sm:max-w-[200px] truncate" title={despesa.fornecedor}>
                                      {despesa.fornecedor}
                                  </TableCell>
                                  <TableCell>
                                      {/* Fix for cut-off text: removed max-w and allowed wrapping or using tooltip if needed. 
                                          Here using a Badge that wraps by default, but better truncated with title */}
                                      <div className="truncate max-w-[180px]" title={despesa.tipo_despesa}>
                                        <Badge variant="outline" className="text-[10px] font-normal truncate block w-full">
                                            {despesa.tipo_despesa}
                                        </Badge>
                                      </div>
                                  </TableCell>
                                  <TableCell className="text-right font-medium text-xs sm:text-sm">
                                      {formatCurrency(despesa.valor)}
                                  </TableCell>
                                  <TableCell>
                                      <FileText className="h-4 w-4 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" />
                                  </TableCell>
                                </TableRow>
                            </DialogTrigger>
                            <DialogContent>
                                <DialogHeader>
                                    <DialogTitle>Detalhes da Despesa</DialogTitle>
                                </DialogHeader>
                                <div className="space-y-4 py-4">
                                    <div className="grid grid-cols-2 gap-4">
                                        <div>
                                            <p className="text-xs font-medium text-muted-foreground">Data Emissão</p>
                                            <p className="text-sm font-mono">{formatDate(despesa.data_documento)}</p>
                                        </div>
                                        <div>
                                            <p className="text-xs font-medium text-muted-foreground">Valor Liquido</p>
                                            <p className="text-lg font-bold text-primary">{formatCurrency(despesa.valor)}</p>
                                        </div>
                                    </div>
                                    
                                    <div>
                                        <p className="text-xs font-medium text-muted-foreground">Fornecedor</p>
                                        <p className="text-base font-semibold">{despesa.fornecedor}</p>
                                    </div>

                                    <div>
                                        <p className="text-xs font-medium text-muted-foreground">Tipo de Despesa</p>
                                        <p className="text-sm bg-muted p-2 rounded-md">{despesa.tipo_despesa}</p>
                                    </div>

                                    {despesa.detalhe && (
                                        <div>
                                            <p className="text-xs font-medium text-muted-foreground">Detalhamento</p>
                                            <p className="text-sm">{despesa.detalhe}</p>
                                        </div>
                                    )}
                                </div>
                            </DialogContent>
                        </Dialog>
                      ))
                  )}
                </TableBody>
              </Table>
          </div>
        </div>

        {/* Pagination Controls */}
        {data.total > data.limit && (
           <div className="flex items-center justify-between">
               <div className="text-sm text-muted-foreground">
                   Página {data.page} de {data.total_pages}
               </div>
               <div className="flex items-center gap-2">
                   <Button 
                       variant="outline" 
                       size="sm" 
                       onClick={prevPage} 
                       disabled={page === 1}
                   >
                       <ChevronLeft className="h-4 w-4 mr-1" />
                       Anterior
                   </Button>
                   <Button 
                       variant="outline" 
                       size="sm" 
                       onClick={nextPage} 
                       disabled={page >= data.total_pages}
                   >
                       Próximo
                       <ChevronRight className="h-4 w-4 ml-1" />
                   </Button>
               </div>
           </div>
        )}
    </div>
  );
}
