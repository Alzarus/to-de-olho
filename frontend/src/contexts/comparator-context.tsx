"use client";

import React, { createContext, useCallback, useContext, useEffect, useState } from "react";
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
  /** Troca a seleção inteira (ex.: vinda de ?ids= na URL) */
  replaceSelection: (senators: SenatorBasicProfile[]) => void;
  /** true depois de ler o localStorage: antes disso a seleção vazia não é real */
  isHydrated: boolean;
  isOpen: boolean; // Controls if the dock is manually expanded (mobile)
  setIsOpen: (open: boolean) => void;
}

const ComparatorContext = createContext<ComparatorContextProps | undefined>(
  undefined
);

export const MAX_SENATORS = 5;
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
  const [isHydrated, setIsHydrated] = useState(false);

  // Load from localStorage on mount
  useEffect(() => {
    let parsed: SenatorBasicProfile[] = [];
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        const valor: unknown = JSON.parse(stored);
        if (Array.isArray(valor)) parsed = valor.filter(isSenatorBasicProfile);
      }
    } catch (e) {
      console.error("Failed to parse stored selection", e);
    }
    // Defer state update to avoid synchronous render cascade
    queueMicrotask(() => {
      setSelectedSenators(parsed);
      setIsHydrated(true);
    });
  }, []);

  // Save to localStorage whenever selection changes (só depois de ler, para
  // não gravar a lista vazia do primeiro render por cima da salva)
  useEffect(() => {
    if (!isHydrated) return;
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(selectedSenators));
    } catch {
      // sem localStorage (modo privado restrito): a URL ainda guarda a seleção
    }
  }, [selectedSenators, isHydrated]);

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

  // Estável (useCallback): a página do comparador usa em efeito
  const replaceSelection = useCallback((senators: SenatorBasicProfile[]) => {
    setSelectedSenators(senators.slice(0, MAX_SENATORS));
  }, []);

  return (
    <ComparatorContext.Provider
      value={{
        selectedSenators,
        addSenator,
        removeSenator,
        clearSelection,
        replaceSelection,
        isHydrated,
        isOpen,
        setIsOpen,
      }}
    >
      {children}
    </ComparatorContext.Provider>
  );
}

function isSenatorBasicProfile(valor: unknown): valor is SenatorBasicProfile {
  if (typeof valor !== "object" || valor === null) return false;
  const v = valor as Record<string, unknown>;
  return typeof v.id === "number" && typeof v.nome === "string";
}

export function useComparator() {
  const context = useContext(ComparatorContext);
  if (context === undefined) {
    throw new Error("useComparator must be used within a ComparatorProvider");
  }
  return context;
}
