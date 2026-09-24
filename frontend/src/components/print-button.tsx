"use client";

import { useEffect, useState } from "react";
import { flushSync } from "react-dom";
import { Printer } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

/** Abre a caixa de impressão do navegador, onde também dá para salvar em PDF */
export function PrintButton({ className }: { className?: string }) {
  return (
    <Button
      type="button"
      variant="outline"
      size="sm"
      onClick={() => window.print()}
      className={cn("no-print", className)}
    >
      <Printer className="mr-2 h-4 w-4" aria-hidden="true" />
      Imprimir / Salvar PDF
    </Button>
  );
}

function agoraFormatado(): string {
  return new Date().toLocaleString("pt-BR", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

/**
 * Só aparece no papel: endereço da página e data da impressão no topo, para
 * quem recebe o PDF saber de onde veio e de quando é o dado.
 */
export function PrintInfo() {
  const [info, setInfo] = useState<{ url: string; data: string } | null>(null);

  useEffect(() => {
    // flushSync: o navegador monta a página de impressão logo após o evento,
    // antes do próximo ciclo normal de renderização do React
    const atualizar = () =>
      flushSync(() => setInfo({ url: window.location.href, data: agoraFormatado() }));
    window.addEventListener("beforeprint", atualizar);
    return () => window.removeEventListener("beforeprint", atualizar);
  }, []);

  return (
    <div className="print-only mb-4 border-b px-4 pb-2 text-xs" aria-hidden="true">
      <p className="font-semibold">Tô De Olho — transparência no Senado (todeolho.org)</p>
      {info && (
        <p>
          {info.url} · impresso em {info.data}
        </p>
      )}
    </div>
  );
}
