"use client";

import Link from "next/link";
import { useComparator } from "@/contexts/comparator-context";
import { X, Users } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";

export function ComparatorDock() {
  const { selectedSenators, removeSenator, clearSelection } = useComparator();

  if (selectedSenators.length === 0) {
    return null;
  }

  return (
    <AnimatePresence>
      <motion.div
        initial={{ y: 100, opacity: 0 }}
        animate={{ y: 0, opacity: 1 }}
        exit={{ y: 100, opacity: 0 }}
        className="fixed bottom-4 left-1/2 z-50 -translate-x-1/2"
      >
        <div className="flex items-center gap-4 rounded-full border border-senado-blue-100 bg-white/95 p-2 px-4 shadow-xl backdrop-blur-sm dark:border-senado-blue-800 dark:bg-senado-blue-950/90">
          <div className="flex -space-x-3">
            {selectedSenators.map((senator) => (
              <div key={senator.id} className="relative group">
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src={senator.fotoUrl}
                  alt={senator.nome}
                  className="h-10 w-10 rounded-full border-2 border-white object-cover dark:border-senado-blue-950"
                />
                <button
                  onClick={() => removeSenator(senator.id)}
                  className="absolute -right-1 -top-1 hidden h-4 w-4 items-center justify-center rounded-full bg-red-500 text-white group-hover:flex"
                  aria-label={`Remover ${senator.nome}`}
                >
                  <X size={10} />
                </button>
              </div>
            ))}
          </div>

          <div className="h-8 w-px bg-gray-200 dark:bg-gray-700" />

          <div className="flex items-center gap-2">
            <span className="text-sm font-medium text-senado-blue-900 dark:text-white">
              {selectedSenators.length} selecionado(s)
            </span>
            {selectedSenators.length >= 2 ? (
              <Link
                href="/comparar"
                className="flex items-center gap-2 rounded-full bg-senado-gold-500 px-4 py-2 text-sm font-bold text-senado-blue-950 transition-colors hover:bg-senado-gold-400"
              >
                <Users size={16} />
                Comparar
              </Link>
            ) : (
              <span className="text-xs text-muted-foreground">
                Selecione mais um
              </span>
            )}
            
            <button 
                onClick={clearSelection}
                className="ml-2 rounded-full p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-senado-blue-800"
                title="Limpar seleção"
            >
                <X size={16} />
            </button>
          </div>
        </div>
      </motion.div>
    </AnimatePresence>
  );
}
