"use client";

import React from "react";
import { useComparator } from "@/contexts/comparator-context";
import { Button } from "@/components/ui/button";
import { Search, X } from "lucide-react";

// Helper type to accept either full Score object or partial profile
interface SenatorLike {
  senador_id?: number;
  id?: number;
  nome: string;
  partido: string;
  uf: string;
  foto_url?: string;
  fotoUrl?: string;
}

export function CompareToggleButton({ 
  senator, 
  className = "" 
}: { 
  senator: SenatorLike, 
  className?: string 
}) {
  const { addSenator, removeSenator, selectedSenators } = useComparator();
  
  // Normalize ID
  const id = senator.senador_id || senator.id;
  if (!id) return null;

  const isSelected = selectedSenators.some(s => s.id === id);

  const toggle = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    
    if (isSelected) {
      removeSenator(id);
    } else {
      addSenator({
        id: id,
        nome: senator.nome,
        partido: senator.partido,
        uf: senator.uf,
        fotoUrl: senator.foto_url || senator.fotoUrl || ""
      });
    }
  };

  return (
    <Button
      variant={isSelected ? "secondary" : "outline"}
      size="sm"
      onClick={toggle}
      className={`gap-2 ${isSelected ? "bg-senado-gold-100 text-senado-gold-900 border-senado-gold-300 dark:bg-senado-gold-900/20 dark:text-senado-gold-200" : ""} ${className}`}
    >
      {isSelected ? (
        <>
          <X className="h-4 w-4" />
          Remover
        </>
      ) : (
        <>
          <Search className="h-4 w-4" />
          Comparar
        </>
      )}
    </Button>
  );
}
