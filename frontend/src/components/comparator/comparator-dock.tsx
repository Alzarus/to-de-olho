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
        className="fixed bottom-4 left-0 right-0 z-50 flex justify-center pointer-events-none"
        style={{ paddingRight: "var(--removed-body-scroll-bar-size)" }}
      >
        <div className="pointer-events-auto w-[95%] max-w-fit flex flex-col items-center gap-2 rounded-2xl border border-border bg-background p-2 shadow-sm dark:bg-popover">
             
          {/* Header / Collapsed View Controls */}
          <div className="flex items-center gap-2 px-1 sm:gap-3 sm:px-2">
            
            <button
                onClick={() => setIsExpanded(!isExpanded)}
                className="flex min-w-0 items-center gap-1.5 text-sm font-medium text-foreground hover:opacity-80 sm:gap-2"
            >
                <div className="flex shrink-0 items-center justify-center rounded-full bg-primary/10 p-1">
                    {isExpanded ? <ChevronDown size={14} /> : <ChevronUp size={14} />}
                </div>
                <span className="truncate">{selectedSenators.length} <span className="hidden xs:inline">selecionado(s)</span></span>
            </button>

            <div className="h-4 w-px shrink-0 bg-border" />

            {selectedSenators.length >= 2 ? (
              <Link
                href="/comparar"
                className="flex shrink-0 items-center gap-1.5 rounded-full bg-primary px-2.5 py-1.5 text-xs font-bold text-primary-foreground transition-all hover:bg-primary/90 active:scale-95 sm:gap-2 sm:px-3"
              >
                <Users size={14} />
                <span>Comparar</span>
              </Link>
            ) : (
              <span className="shrink-0 text-[10px] text-muted-foreground sm:text-xs">
                Selecione +1
              </span>
            )}

            <button 
                onClick={clearSelection}
                className="ml-0.5 rounded-full p-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground sm:ml-1"
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
