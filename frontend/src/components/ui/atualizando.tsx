import { AlertCircle, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";

/**
 * Envolve uma lista que mantém os dados anteriores (placeholderData) enquanto
 * o filtro ou a página nova carrega: esmaece, bloqueia cliques e mostra um
 * aviso fixo no topo. Sem isso, a troca de filtro parecia não fazer nada.
 */
export function Atualizando({
  ativo,
  children,
  className,
}: {
  ativo: boolean;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <div className={cn("relative", className)} aria-busy={ativo}>
      {ativo && (
        <div className="pointer-events-none sticky top-20 z-10 flex h-0 justify-center">
          <span
            role="status"
            className="mt-2 flex h-fit items-center gap-2 rounded-full border bg-background px-3 py-1 text-xs text-muted-foreground shadow-sm"
          >
            <Loader2 className="h-3 w-3 animate-spin" aria-hidden="true" />
            Atualizando...
          </span>
        </div>
      )}
      <div className={cn("transition-opacity", ativo && "pointer-events-none opacity-50")}>{children}</div>
    </div>
  );
}

/** Falha ao carregar: no lugar da aba em branco, um aviso com "tentar de novo" */
export function ErroCarregamento({
  mensagem = "Não foi possível carregar os dados.",
  aoTentar,
}: {
  mensagem?: string;
  aoTentar?: () => void;
}) {
  return (
    <div
      role="alert"
      className="flex flex-col items-center gap-3 rounded-lg border border-dashed py-10 text-center text-sm text-muted-foreground"
    >
      <AlertCircle className="h-8 w-8 opacity-40" aria-hidden="true" />
      <p>{mensagem}</p>
      {aoTentar && (
        <button
          type="button"
          onClick={aoTentar}
          className="rounded-md border px-3 py-1.5 text-foreground hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          Tentar de novo
        </button>
      )}
    </div>
  );
}
