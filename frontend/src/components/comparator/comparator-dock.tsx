"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { useComparator } from "@/contexts/comparator-context";
import { X, Users, ChevronUp, ChevronDown } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";

export function ComparatorDock() {
  const { selectedSenators, removeSenator, clearSelection } = useComparator();
  const [isExpanded, setIsExpanded] = useState(false);

  // Auto-expand when adding the first few, but maybe keep collapsed if many?
  // Let's default to collapsed to save space as requested, or expand only on first add?
  // User said "seria bom um expand para não ocupar tanto espaço inicialmente".
  // So default is collapsed (false).
  
  if (selectedSenators.length === 0) {
    return null;
  }

  return (
    <AnimatePresence>
      <motion.div
        initial={{ y: 100, opacity: 0 }}
        animate={{ y: 0, opacity: 1 }}
        exit={{ y: 100, opacity: 0 }}
        className="fixed bottom-4 left-1/2 z-50 w-[95%] max-w-fit -translate-x-1/2"
      >
        <div className="flex flex-col items-center gap-2 rounded-2xl border border-border/50 bg-background/80 p-2 shadow-2xl backdrop-blur-md dark:bg-popover/80">
             
          {/* Header / Collapsed View Controls */}
          <div className="flex items-center gap-3 px-2">
            
            <button
                onClick={() => setIsExpanded(!isExpanded)}
                className="flex items-center gap-2 text-sm font-medium text-foreground hover:opacity-80"
            >
                <div className="flex items-center justify-center rounded-full bg-primary/10 p-1">
                    {isExpanded ? <ChevronDown size={14} /> : <ChevronUp size={14} />}
                </div>
                <span>{selectedSenators.length} selecionado(s)</span>
            </button>

            <div className="h-4 w-px bg-border" />

            {selectedSenators.length >= 2 ? (
              <Link
                href="/comparar"
                className="flex items-center gap-2 rounded-full bg-primary px-3 py-1.5 text-xs font-bold text-primary-foreground transition-colors hover:bg-primary/90"
              >
                <Users size={14} />
                Comparar
              </Link>
            ) : (
              <span className="text-xs text-muted-foreground">
                Selecione +1
              </span>
            )}

            <button 
                onClick={clearSelection}
                className="ml-1 rounded-full p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-senado-blue-800"
                title="Limpar seleção"
            >
                <X size={14} />
            </button>
          </div>

          {/* Expanded List */}
          <AnimatePresence>
            {isExpanded && (
                <motion.div
                    initial={{ height: 0, opacity: 0 }}
                    animate={{ height: "auto", opacity: 1 }}
                    exit={{ height: 0, opacity: 0 }}
                    className="overflow-hidden"
                >
                    <div className="flex gap-2 p-2 pt-3">
                        {selectedSenators.map((senator) => (
                        <div key={senator.id} className="relative group">
                            {/* eslint-disable-next-line @next/next/no-img-element */}
                            <img
                            src={senator.fotoUrl}
                            alt={senator.nome}
                            className="h-10 w-10 rounded-full border-2 border-white object-cover shadow-sm dark:border-senado-blue-950"
                            />
                            <button
                            onClick={() => {
                                removeSenator(senator.id);
                            }}
                            className="absolute -right-1.5 -top-1.5 z-20 flex h-5 w-5 items-center justify-center rounded-full bg-destructive text-white shadow ring-2 ring-white hover:bg-destructive/90"
                            aria-label={`Remover ${senator.nome}`}
                            >
                            <X size={12} strokeWidth={3} />
                            </button>
                        </div>
                        ))}
                    </div>
                </motion.div>
            )}
          </AnimatePresence>

        </div>
      </motion.div>
    </AnimatePresence>
  );
}
