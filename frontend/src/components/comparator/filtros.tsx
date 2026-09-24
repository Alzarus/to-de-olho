"use client";

// Controles de filtro das abas do comparador. Todo filtro vive na URL (junto
// de ?ids=, ?ano= e ?tab=): o link compartilhado reproduz a mesma visão.
// Controles nativos ou Radix, sempre com rótulo visível e operáveis por teclado.

import { useCallback, useId, type ReactNode } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { ChevronDown } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";

export type ValorParametro = string | number | readonly string[] | null;

/** Lê os parâmetros da URL e altera alguns deles sem poluir o histórico. */
export function useFiltrosUrl() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();

  const definir = useCallback(
    (mudancas: Record<string, ValorParametro>) => {
      const params = new URLSearchParams(searchParams.toString());
      for (const [chave, valor] of Object.entries(mudancas)) {
        params.delete(chave);
        if (valor === null || valor === "") continue;
        if (Array.isArray(valor)) {
          for (const v of valor as readonly string[]) params.append(chave, v);
        } else {
          params.set(chave, String(valor));
        }
      }
      const query = params.toString();
      router.replace(query ? `${pathname}?${query}` : pathname, { scroll: false });
    },
    [searchParams, router, pathname],
  );

  return { params: searchParams, definir };
}

/** Linha de filtros acima dos gráficos que eles recortam. */
export function BarraFiltros({ rotulo, children, className }: { rotulo: string; children: ReactNode; className?: string }) {
  return (
    <div
      role="group"
      aria-label={rotulo}
      className={cn("no-print flex flex-wrap items-end gap-x-4 gap-y-3 rounded-lg border bg-card p-3 sm:p-4", className)}
    >
      {children}
    </div>
  );
}

export interface Opcao {
  valor: string;
  rotulo: string;
}

export function FiltroSelect({
  rotulo,
  valor,
  opcoes,
  aoMudar,
  largura = "w-[180px]",
  desabilitado = false,
}: {
  rotulo: string;
  valor: string;
  opcoes: readonly Opcao[];
  aoMudar: (valor: string) => void;
  largura?: string;
  desabilitado?: boolean;
}) {
  const id = useId();
  return (
    <div className="flex flex-col gap-1">
      <label htmlFor={id} className="text-xs font-medium text-muted-foreground">
        {rotulo}
      </label>
      <Select value={valor} onValueChange={aoMudar} disabled={desabilitado}>
        <SelectTrigger id={id} className={cn("h-9", largura)}>
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {opcoes.map((o) => (
            <SelectItem key={o.valor} value={o.valor}>
              {o.rotulo}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}

/** Grupo de opções exclusivas (rádios nativos: setas do teclado alternam). */
export function FiltroSegmentado({
  rotulo,
  valor,
  opcoes,
  aoMudar,
}: {
  rotulo: string;
  valor: string;
  opcoes: readonly Opcao[];
  aoMudar: (valor: string) => void;
}) {
  const nome = useId();
  return (
    <fieldset className="flex flex-col gap-1">
      <legend className="mb-1 text-xs font-medium text-muted-foreground">{rotulo}</legend>
      <div className="inline-flex h-9 items-center rounded-md border bg-muted/40 p-0.5">
        {opcoes.map((o) => {
          const id = `${nome}-${o.valor}`;
          const ativo = o.valor === valor;
          return (
            <span key={o.valor} className="relative">
              <input
                type="radio"
                id={id}
                name={nome}
                value={o.valor}
                checked={ativo}
                onChange={() => aoMudar(o.valor)}
                className="peer sr-only"
              />
              <label
                htmlFor={id}
                className={cn(
                  "flex h-8 cursor-pointer items-center rounded-[5px] px-3 text-sm transition-colors",
                  "peer-focus-visible:ring-[3px] peer-focus-visible:ring-ring/50",
                  ativo
                    ? "bg-background font-semibold text-foreground shadow-sm"
                    : "text-muted-foreground hover:text-foreground",
                )}
              >
                {o.rotulo}
              </label>
            </span>
          );
        })}
      </div>
    </fieldset>
  );
}

/** Seleção múltipla num menu com caixas de seleção (Radix: setas, espaço, Esc). */
export function FiltroMulti({
  rotulo,
  opcoes,
  selecionados,
  aoMudar,
  aoRestaurar,
  rotuloRestaurar = "Restaurar padrão",
  resumo,
}: {
  rotulo: string;
  opcoes: readonly Opcao[];
  selecionados: readonly string[];
  aoMudar: (valores: string[]) => void;
  aoRestaurar?: () => void;
  rotuloRestaurar?: string;
  resumo?: string;
}) {
  const idRotulo = useId();
  const idBotao = useId();
  const marcados = new Set(selecionados);
  const texto =
    resumo ??
    (selecionados.length === 0
      ? "Nenhuma"
      : selecionados.length === 1
        ? (opcoes.find((o) => o.valor === selecionados[0])?.rotulo ?? selecionados[0])
        : `${selecionados.length} selecionadas`);

  const alternar = (valor: string) => {
    // Mantém a ordem das opções, para a legenda não pular
    const novo = marcados.has(valor)
      ? selecionados.filter((v) => v !== valor)
      : opcoes.map((o) => o.valor).filter((v) => v === valor || marcados.has(v));
    aoMudar(novo);
  };

  return (
    <div className="flex flex-col gap-1">
      <span id={idRotulo} className="text-xs font-medium text-muted-foreground">
        {rotulo}
      </span>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            id={idBotao}
            variant="outline"
            aria-labelledby={`${idRotulo} ${idBotao}`}
            className="h-9 w-[240px] justify-between font-normal"
          >
            <span className="truncate">{texto}</span>
            <ChevronDown className="h-4 w-4 opacity-60" aria-hidden="true" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" className="max-h-[360px] w-[320px] overflow-y-auto">
          <DropdownMenuLabel>{rotulo}</DropdownMenuLabel>
          {aoRestaurar && (
            <>
              <DropdownMenuItem onSelect={aoRestaurar}>{rotuloRestaurar}</DropdownMenuItem>
              <DropdownMenuSeparator />
            </>
          )}
          {opcoes.map((o) => (
            <DropdownMenuCheckboxItem
              key={o.valor}
              checked={marcados.has(o.valor)}
              onCheckedChange={() => alternar(o.valor)}
              // Mantém o menu aberto para marcar várias
              onSelect={(e) => e.preventDefault()}
              className="whitespace-normal"
            >
              {o.rotulo}
            </DropdownMenuCheckboxItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}

export interface ColunaTabela<T> {
  cabecalho: string;
  valor: (linha: T) => ReactNode;
  numerica?: boolean;
}

/** Tabela com os dados do gráfico, recolhida por padrão (alternativa ao visual). */
export function TabelaDados<T>({
  titulo,
  colunas,
  linhas,
  chave,
}: {
  titulo: string;
  colunas: readonly ColunaTabela<T>[];
  linhas: readonly T[];
  chave: (linha: T, indice: number) => string;
}) {
  return (
    <details className="group mt-4 rounded-md border">
      <summary className="cursor-pointer select-none rounded-md px-3 py-2 text-sm font-medium text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50">
        Ver dados em tabela
      </summary>
      <div className="overflow-x-auto px-3 pb-3">
        <table className="w-full text-sm">
          <caption className="sr-only">{titulo}</caption>
          <thead>
            <tr className="border-b">
              {colunas.map((c) => (
                <th
                  key={c.cabecalho}
                  scope="col"
                  className={cn("px-2 py-2 font-semibold", c.numerica ? "text-right" : "text-left")}
                >
                  {c.cabecalho}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {linhas.map((linha, i) => (
              <tr key={chave(linha, i)} className="border-b last:border-0">
                {colunas.map((c, j) =>
                  j === 0 ? (
                    <th key={c.cabecalho} scope="row" className="px-2 py-1.5 text-left font-normal">
                      {c.valor(linha)}
                    </th>
                  ) : (
                    <td key={c.cabecalho} className={cn("px-2 py-1.5", c.numerica && "text-right tabular-nums")}>
                      {c.valor(linha)}
                    </td>
                  ),
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </details>
  );
}

/** Marcador da cor do senador ao lado do nome (decorativo: o nome identifica). */
export function MarcadorSerie({ cor, className }: { cor: string; className?: string }) {
  return (
    <span
      aria-hidden="true"
      className={cn("inline-block h-2.5 w-2.5 shrink-0 rounded-[2px]", className)}
      style={{ backgroundColor: cor }}
    />
  );
}

const REAIS = new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" });

/** Valor completo, para tabela e tooltip: R$ 12.345,67 */
export function formatarReais(valor: unknown): string {
  const n = typeof valor === "number" ? valor : Number(valor);
  return REAIS.format(Number.isFinite(n) ? n : 0);
}

/** Percentual com uma casa: 83,4% (— sem dado) */
export function formatarPct(valor: unknown): string {
  if (valor === null || valor === undefined || valor === "") return "—";
  const n = typeof valor === "number" ? valor : Number(valor);
  if (!Number.isFinite(n)) return "—";
  return `${n.toLocaleString("pt-BR", { maximumFractionDigits: 1 })}%`;
}

/** Rótulo curto de eixo: R$ 45 mil, R$ 1,2 mi */
export function eixoReais(valor: number): string {
  const abs = Math.abs(valor);
  if (abs >= 1_000_000) return `R$ ${(valor / 1_000_000).toLocaleString("pt-BR", { maximumFractionDigits: 1 })} mi`;
  if (abs >= 1_000) return `R$ ${(valor / 1_000).toLocaleString("pt-BR", { maximumFractionDigits: 0 })} mil`;
  return `R$ ${valor.toLocaleString("pt-BR", { maximumFractionDigits: 0 })}`;
}

export function truncar(texto: string, tamanho: number): string {
  return texto.length > tamanho ? `${texto.slice(0, tamanho - 1)}…` : texto;
}
