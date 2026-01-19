"use client";

import * as React from "react";
import { Search, X } from "lucide-react";
import { cn } from "@/lib/utils";
import { Input } from "@/components/ui/input";
import { getSenadores } from "@/lib/api";
import { useQuery } from "@tanstack/react-query";
import { useComparator } from "@/contexts/comparator-context";

export function SenatorSearch() {
  const [isOpen, setIsOpen] = React.useState(false);
  const [search, setSearch] = React.useState("");
  const [selectedUf, setSelectedUf] = React.useState("TODOS");
  const [selectedPartido, setSelectedPartido] = React.useState("TODOS");
  
  const containerRef = React.useRef<HTMLDivElement>(null);
  const { addSenator, selectedSenators } = useComparator();
  
  /* eslint-disable @next/next/no-img-element */
  const { data: senadoresData = [] } = useQuery({
    queryKey: ["senadores-search"],
    queryFn: getSenadores,
    staleTime: 1000 * 60 * 60, // 1 hour
  });

  const { uniqueUfs, uniquePartidos } = React.useMemo(() => {
    const ufs = new Set<string>();
    const partidos = new Set<string>();
    const list = Array.isArray(senadoresData) ? senadoresData : [];
    
    list.forEach(s => {
        if (s.uf) ufs.add(s.uf);
        if (s.partido) partidos.add(s.partido);
    });

    return {
        uniqueUfs: Array.from(ufs).sort(),
        uniquePartidos: Array.from(partidos).sort()
    };
  }, [senadoresData]);

  // Filter out already selected senators
  const availableSenators = React.useMemo(() => {
    const list = Array.isArray(senadoresData) ? senadoresData : [];
    const selectedIds = new Set(selectedSenators.map(s => s.id));
    return list
      .filter(s => !selectedIds.has(s.id))
      .filter(s => s.nome.toLowerCase().includes(search.toLowerCase()))
      .filter(s => selectedUf === "TODOS" || s.uf === selectedUf)
      .filter(s => selectedPartido === "TODOS" || s.partido === selectedPartido);
  }, [senadoresData, selectedSenators, search, selectedUf, selectedPartido]);

  const handleSelect = (senator: { id: number; nome: string; partido: string; uf: string; foto_url?: string }) => {
    addSenator({
        id: senator.id,
        nome: senator.nome,
        partido: senator.partido,
        uf: senator.uf,
        fotoUrl: senator.foto_url || "",
    });
    setSearch("");
    setIsOpen(false);
  };

  // Close on click outside
  React.useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  return (
    <div className="relative w-full sm:w-auto flex flex-col sm:flex-row gap-2" ref={containerRef}>
      <div className="relative flex-1">
        <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
        <Input
          placeholder="Nome do senador..."
          value={search}
          onChange={(e) => {
            setSearch(e.target.value);
            if (!isOpen) setIsOpen(true);
          }}
          onFocus={() => setIsOpen(true)}
          className="pl-9 w-full sm:w-[250px]"
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

      <select 
        value={selectedUf}
        onChange={(e) => setSelectedUf(e.target.value)}
        className="h-10 rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
      >
        <option value="TODOS">Todos UFs</option>
        {uniqueUfs.map(uf => (
            <option key={uf} value={uf}>{uf}</option>
        ))}
      </select>

      <select 
        value={selectedPartido}
        onChange={(e) => setSelectedPartido(e.target.value)}
        className="h-10 rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
      >
        <option value="TODOS">Todos Partidos</option>
        {uniquePartidos.map(p => (
            <option key={p} value={p}>{p}</option>
        ))}
      </select>

      {isOpen && (
        <div className="absolute top-full left-0 z-50 mt-1 max-h-[300px] w-full sm:w-[450px] overflow-auto rounded-md border border-border bg-popover p-1 text-popover-foreground shadow-md">
            {availableSenators.length === 0 ? (
                <div className="py-6 text-center text-sm text-muted-foreground">
                    {search || selectedUf !== "TODOS" || selectedPartido !== "TODOS" ? "Nenhum senador encontrado." : "Utilize os filtros para buscar."}
                </div>
            ) : (
                <div className="space-y-1">
                    {availableSenators.map((senator) => (
                        <button
                            key={senator.id}
                            onClick={() => handleSelect(senator)}
                            className={cn(
                                "flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-none hover:bg-accent hover:text-accent-foreground",
                                "cursor-pointer"
                            )}
                        >
                            {senator.foto_url ? (
                                 
                                <img src={senator.foto_url} alt="" className="h-6 w-6 rounded-full object-cover" />
                            ) : (
                                <div className="h-6 w-6 rounded-full bg-muted flex items-center justify-center text-[10px]">
                                    {senator.nome.charAt(0)}
                                </div>
                            )}
                            <div className="flex flex-col items-start">
                                <span className="font-medium text-left">{senator.nome}</span>
                                <span className="text-xs text-muted-foreground">{senator.partido} - {senator.uf}</span>
                            </div>
                        </button>
                    ))}
                </div>
            )}
        </div>
      )}
    </div>
  );
}
