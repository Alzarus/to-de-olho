"use client";

import * as React from "react";
import { Search, Check, X } from "lucide-react";
import { cn } from "@/lib/utils";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { getRanking } from "@/lib/api";
import { useQuery } from "@tanstack/react-query";
import { useComparator } from "@/contexts/comparator-context";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";

interface SenatorSelectorProps {
  className?: string;
}

export function SenatorSelector({ className }: SenatorSelectorProps) {
  const [search, setSearch] = React.useState("");
  const [selectedUf, setSelectedUf] = React.useState("TODOS");
  const [selectedPartido, setSelectedPartido] = React.useState("TODOS");
  
  const { addSenator, removeSenator, selectedSenators } = useComparator();
  const selectedIds = new Set(selectedSenators.map(s => s.id));
  
  // Use getRanking which is confirmed working (returns { ranking: SenadorScore[] })
  const { data, isLoading } = useQuery({
    queryKey: ["ranking-for-selector"],
    queryFn: () => getRanking(100), // Get all senators
    staleTime: 1000 * 60 * 60, // 1 hour
  });
  
   
  const senadoresData = React.useMemo(() => data?.ranking || [], [data?.ranking]);

  const { uniqueUfs, uniquePartidos } = React.useMemo(() => {
    const ufs = new Set<string>();
    const partidos = new Set<string>();
    
    senadoresData.forEach(s => {
        if (s.uf) ufs.add(s.uf);
        if (s.partido) partidos.add(s.partido);
    });

    return {
        uniqueUfs: Array.from(ufs).sort(),
        uniquePartidos: Array.from(partidos).sort()
    };
  }, [senadoresData]);

  const filteredSenators = React.useMemo(() => {
    return senadoresData
      .filter(s => s.nome.toLowerCase().includes(search.toLowerCase()))
      .filter(s => selectedUf === "TODOS" || s.uf === selectedUf)
      .filter(s => selectedPartido === "TODOS" || s.partido === selectedPartido);
  }, [senadoresData, search, selectedUf, selectedPartido]);

  const handleToggle = (senator: { senador_id: number; nome: string; partido: string; uf: string; foto_url?: string }) => {
    if (selectedIds.has(senator.senador_id)) {
      removeSenator(senator.senador_id);
    } else if (selectedSenators.length < 5) {
      addSenator({
        id: senator.senador_id,
        nome: senator.nome,
        partido: senator.partido,
        uf: senator.uf,
        fotoUrl: senator.foto_url || "",
      });
    }
  };

  if (isLoading) {
    return (
      <div className={cn("space-y-4", className)}>
        <div className="flex gap-2 flex-wrap">
          <Skeleton className="h-10 w-[250px]" />
          <Skeleton className="h-10 w-[120px]" />
          <Skeleton className="h-10 w-[150px]" />
        </div>
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3">
          {[...Array(10)].map((_, i) => (
            <Skeleton key={i} className="h-32 rounded-lg" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className={cn("space-y-4", className)}>
      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Buscar por nome..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
          />
          {search && (
            <button 
              onClick={() => setSearch("")}
              className="absolute right-2 top-2.5 text-muted-foreground hover:text-foreground"
            >
              <X className="h-4 w-4" />
            </button>
          )}
        </div>

        <Select 
          value={selectedUf}
          onValueChange={(value) => setSelectedUf(value)}
        >
          <SelectTrigger className="w-full sm:w-[180px]">
            <SelectValue placeholder="Selecione o estado" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="TODOS">Todos os Estados</SelectItem>
            {uniqueUfs.map(uf => (
              <SelectItem key={uf} value={uf}>{uf}</SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select 
          value={selectedPartido}
          onValueChange={(value) => setSelectedPartido(value)}
        >
          <SelectTrigger className="w-full sm:w-[180px]">
             <SelectValue placeholder="Selecione o partido" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="TODOS">Todos os Partidos</SelectItem>
            {uniquePartidos.map(p => (
              <SelectItem key={p} value={p}>{p}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Selection Info */}
      <div className="flex items-center justify-between text-sm text-muted-foreground">
        <span>
          {filteredSenators.length} senador{filteredSenators.length !== 1 ? 'es' : ''} encontrado{filteredSenators.length !== 1 ? 's' : ''}
        </span>
        <span className={cn(
          "font-medium",
          selectedSenators.length >= 5 ? "text-amber-600 dark:text-amber-400" : ""
        )}>
          {selectedSenators.length}/5 selecionados
        </span>
      </div>

      {/* Senator Grid */}
      {filteredSenators.length === 0 ? (
        <div className="py-12 text-center text-muted-foreground">
          Nenhum senador encontrado com os filtros selecionados.
        </div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3">
          {filteredSenators.map((senator) => {
            const isSelected = selectedIds.has(senator.senador_id);
            const isDisabled = !isSelected && selectedSenators.length >= 5;
            
            return (
              <button
                key={senator.senador_id}
                onClick={() => handleToggle(senator)}
                disabled={isDisabled}
                className={cn(
                  "relative flex flex-col items-center p-4 rounded-lg border-2 transition-all text-center",
                  "hover:shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
                  isSelected 
                    ? "border-primary bg-primary/5 shadow-sm" 
                    : "border-border hover:border-muted-foreground/50",
                  isDisabled && "opacity-50 cursor-not-allowed hover:shadow-none hover:border-border"
                )}
              >
                {/* Selection Indicator */}
                {isSelected && (
                  <div className="absolute top-2 right-2 h-5 w-5 rounded-full bg-primary flex items-center justify-center">
                    <Check className="h-3 w-3 text-primary-foreground" />
                  </div>
                )}
                
                {/* Photo */}
                {senator.foto_url ? (
                  /* eslint-disable-next-line @next/next/no-img-element */
                  <img 
                    src={senator.foto_url} 
                    alt="" 
                    className="h-16 w-16 rounded-full object-cover mb-2 border border-border"
                  />
                ) : (
                  <div className="h-16 w-16 rounded-full bg-muted flex items-center justify-center text-xl font-bold mb-2">
                    {senator.nome.charAt(0)}
                  </div>
                )}
                
                {/* Info */}
                <span className="font-medium text-sm leading-tight line-clamp-2">
                  {senator.nome}
                </span>
                <div className="flex items-center gap-1 mt-1">
                  <Badge variant="secondary" className="text-xs px-1.5 py-0">
                    {senator.partido}
                  </Badge>
                  <span className="text-xs text-muted-foreground">{senator.uf}</span>
                </div>
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
