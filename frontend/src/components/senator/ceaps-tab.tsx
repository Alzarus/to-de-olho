"use client";

import { PaginationWithInput } from "@/components/ui/pagination-with-input";

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
import { Search, ChevronLeft, ChevronRight, FileText, ArrowDown, ArrowUp, ArrowUpDown } from "lucide-react";
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

// Helper component for sortable headers
function SortableHeader({ 
  label, 
  sortKey, 
  currentSort, 
  onSort,
  className = ""
}: { 
  label: string; 
  sortKey: string; 
  currentSort: string; 
  onSort: (key: string) => void;
  className?: string;
}) {
  // Extract direction from currentSort (e.g., "data_desc" -> key="data", dir="desc")
  const [currentKey, currentDir] = currentSort.split("_");
  const isActive = currentKey === sortKey;
  
  return (
    <TableHead 
      className={`cursor-pointer hover:text-foreground transition-colors select-none ${className}`}
      onClick={() => onSort(sortKey)}
    >
      <span className="flex items-center gap-1">
        {label}
        {isActive ? (
          currentDir === "desc" ? <ArrowDown className="h-3 w-3" /> : <ArrowUp className="h-3 w-3" />
        ) : (
          <ArrowUpDown className="h-3 w-3 opacity-40" />
        )}
      </span>
    </TableHead>
  );
}

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

  // Reset page logic on year change
  useEffect(() => {
     if (page !== 1) {
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
  
  // Sorting Link
  const handleSort = (key: string) => {
      const [currentKey, currentDir] = sort.split("_");
      let newDir = "desc";
      
      if (currentKey === key) {
          newDir = currentDir === "desc" ? "asc" : "desc";
      }
      
      updateUrl("ceaps_sort", `${key}_${newDir}`);
  };

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
        <div className="flex flex-col gap-4">
            <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
                <h3 className="text-lg font-semibold hidden sm:block">Detalhamento de Despesas</h3>
                <div className="relative w-full sm:w-72">
                    <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                    <Input 
                        placeholder="Filtrar por fornecedor..." 
                        className="pl-8"
                        value={searchValue}
                        onChange={handleSearch}
                    />
                </div>
            </div>

            <div className="flex flex-wrap items-center gap-2">
                <Select value={tipo} onValueChange={setTipo}>
                    <SelectTrigger className="w-full sm:w-[280px]">
                        <SelectValue placeholder="Filtrar por Tipo" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="todos">Todos os Tipos</SelectItem>
                        <SelectItem value="Aluguel de imóveis para escritório político, compreendendo despesas concernentes a eles.">Aluguel de Imóveis</SelectItem>
                        <SelectItem value="Divulgação da atividade parlamentar">Divulgação da Atividade</SelectItem>
                        <SelectItem value="Locomoção, hospedagem, alimentação, combustíveis e lubrificantes">Locomoção/Hospedagem/Combustível</SelectItem>
                        <SelectItem value="Passagens aéreas, aquáticas e terrestres nacionais">Passagens</SelectItem>
                        <SelectItem value="Aquisição de material de consumo para uso no escritório político, inclusive aquisição ou locação de software, despesas postais, aquisição de publicações, locação de móveis e de equipamentos.">Material de Consumo/Escritório</SelectItem>
                        <SelectItem value="Contratação de consultorias, assessorias, pesquisas, trabalhos técnicos e outros serviços de apoio ao exercício do mandato parlamentar">Consultorias e Assessorias</SelectItem>
                        <SelectItem value="Serviços de Segurança Privada">Segurança Privada</SelectItem>
                    </SelectContent>
                </Select>
            </div>
        </div>

        <div className="rounded-md border overflow-hidden">
          <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <SortableHeader 
                        label="Data" 
                        sortKey="data" 
                        currentSort={sort} 
                        onSort={handleSort} 
                        className="w-[100px]"
                    />
                    <SortableHeader 
                        label="Fornecedor" 
                        sortKey="fornecedor" 
                        currentSort={sort} 
                        onSort={handleSort} 
                    />
                    <TableHead className="w-[200px]">Tipo</TableHead>
                    <SortableHeader 
                        label="Valor" 
                        sortKey="valor" 
                        currentSort={sort} 
                        onSort={handleSort} 
                        className="text-right justify-end flex"
                    />
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
                        <Dialog key={`${i}-${despesa.data_emissao}`}>
                            <DialogTrigger asChild>
                                <TableRow className="cursor-pointer hover:bg-muted/50 group">
                                  <TableCell className="font-mono text-xs whitespace-nowrap">
                                      {formatDate(despesa.data_emissao)}
                                  </TableCell>
                                  <TableCell className="max-w-[150px] sm:max-w-[200px] truncate" title={despesa.fornecedor}>
                                      {despesa.fornecedor}
                                  </TableCell>
                                  <TableCell>
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
                                            <p className="text-sm font-mono">{formatDate(despesa.data_emissao)}</p>
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
        <PaginationWithInput 
            currentPage={data.page} 
            totalPages={data.total_pages} 
            onPageChange={setPage} 
            className="border-t pt-4"
        />
    </div>
  );
}
