"use client";

import React, { createContext, useContext, useEffect, useState } from "react";
import { toast } from "sonner";

export interface SenatorBasicProfile {
  id: number;
  nome: string;
  partido: string;
  uf: string;
  fotoUrl: string;
}

interface ComparatorContextProps {
  selectedSenators: SenatorBasicProfile[];
  addSenator: (senator: SenatorBasicProfile) => void;
  removeSenator: (id: number) => void;
  clearSelection: () => void;
  isOpen: boolean; // Controls if the dock is manually expanded (mobile)
  setIsOpen: (open: boolean) => void;
}

const ComparatorContext = createContext<ComparatorContextProps | undefined>(
  undefined
);

const MAX_SENATORS = 5;
const STORAGE_KEY = "todeolho:comparator:selected";

export function ComparatorProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  const [selectedSenators, setSelectedSenators] = useState<
    SenatorBasicProfile[]
  >([]);
  const [isOpen, setIsOpen] = useState(true);

  // Load from localStorage on mount
  useEffect(() => {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      try {
        const parsed = JSON.parse(stored);
        // Defer state update to avoid synchronous render cascade
        queueMicrotask(() => {
          setSelectedSenators(parsed);
        });
      } catch (e) {
        console.error("Failed to parse stored selection", e);
      }
    }
  }, []);

  // Save to localStorage whenever selection changes
  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(selectedSenators));
  }, [selectedSenators]);

  const addSenator = (senator: SenatorBasicProfile) => {
    if (selectedSenators.some((s) => s.id === senator.id)) {
      toast.info("Senador já adicionado à comparação.");
      return;
    }

    if (selectedSenators.length >= MAX_SENATORS) {
      toast.warning(`Limite de ${MAX_SENATORS} senadores atingido.`);
      return;
    }

    setSelectedSenators((prev) => [...prev, senator]);
    setIsOpen(true);
    toast.success(`${senator.nome} adicionado!`);
  };

  const removeSenator = (id: number) => {
    setSelectedSenators((prev) => prev.filter((s) => s.id !== id));
  };

  const clearSelection = () => {
    setSelectedSenators([]);
  };

  return (
    <ComparatorContext.Provider
      value={{
        selectedSenators,
        addSenator,
        removeSenator,
        clearSelection,
        isOpen,
        setIsOpen,
      }}
    >
      {children}
    </ComparatorContext.Provider>
  );
}

export function useComparator() {
  const context = useContext(ComparatorContext);
  if (context === undefined) {
    throw new Error("useComparator must be used within a ComparatorProvider");
  }
  return context;
}
