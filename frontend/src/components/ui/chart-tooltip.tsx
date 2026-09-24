"use client";

import * as React from "react";
import type { DefaultTooltipContentProps } from "recharts";

import { cn } from "@/lib/utils";

export type TooltipPayloadEntry = NonNullable<
  DefaultTooltipContentProps<number | string, number | string>["payload"]
>[number];
export type ChartTooltipPayload = ReadonlyArray<TooltipPayloadEntry>;

export interface ChartTooltipContentProps {
  /** Injetados pelo Recharts quando usado em `<Tooltip content={...} />`. */
  active?: boolean;
  payload?: ChartTooltipPayload;
  label?: React.ReactNode;
  /** Formata o título (categoria/eixo) do tooltip. */
  labelFormatter?: (label: React.ReactNode, payload: ChartTooltipPayload) => React.ReactNode;
  /** Formata o valor de cada série. Padrão: número em pt-BR. */
  valueFormatter?: (value: unknown, entry: TooltipPayloadEntry) => React.ReactNode;
  /** Formata o nome de cada série. */
  nameFormatter?: (name: unknown, entry: TooltipPayloadEntry) => React.ReactNode;
  /** Cor do marcador de cada item (ex.: barras com `<Cell>` de cores próprias). */
  colorFormatter?: (entry: TooltipPayloadEntry) => string | undefined;
  /** Conteúdo extra opcional, exibido abaixo dos itens (texto secundário). */
  extra?: (payload: ChartTooltipPayload) => React.ReactNode;
  hideLabel?: boolean;
  className?: string;
}

function defaultValueFormatter(value: unknown): React.ReactNode {
  if (typeof value === "number") {
    return value.toLocaleString("pt-BR", { maximumFractionDigits: 1 });
  }
  return value == null ? "—" : String(value);
}

function entryColor(entry: TooltipPayloadEntry): string | undefined {
  // Em gráficos de pizza a cor costuma vir apenas no item de dados (payload.fill/color)
  const data = entry.payload as { color?: string; fill?: string } | undefined;
  return entry.color ?? entry.fill ?? data?.color ?? data?.fill;
}

/**
 * Conteúdo de tooltip acessível para gráficos Recharts.
 * O texto sempre usa as cores do tema (popover), nunca a cor da série,
 * garantindo contraste AA nos temas claro e escuro. A cor da série aparece
 * apenas no marcador quadrado (decorativo, aria-hidden).
 */
export function ChartTooltipContent({
  active,
  payload,
  label,
  labelFormatter,
  valueFormatter = defaultValueFormatter,
  nameFormatter,
  colorFormatter = entryColor,
  extra,
  hideLabel = false,
  className,
}: ChartTooltipContentProps) {
  if (!active || !payload || payload.length === 0) return null;

  const items = payload.filter((entry) => entry.type !== "none" && !entry.hide);
  const formattedLabel = labelFormatter ? labelFormatter(label, payload) : label;
  const showLabel = !hideLabel && formattedLabel != null && formattedLabel !== "";
  const extraContent = extra?.(payload);

  return (
    <div
      className={cn(
        "rounded-lg border border-border bg-popover text-popover-foreground px-3 py-2 text-xs shadow-md max-w-[300px]",
        className,
      )}
    >
      {showLabel && (
        <p className="font-semibold text-sm text-popover-foreground mb-1.5 break-words leading-tight">
          {formattedLabel}
        </p>
      )}
      <ul className="space-y-1">
        {items.map((entry, index) => {
          const name = nameFormatter ? nameFormatter(entry.name, entry) : entry.name;
          return (
            <li
              key={`${String(entry.dataKey ?? entry.name)}-${index}`}
              className="flex items-center justify-between gap-4"
            >
              <span className="flex min-w-0 items-center gap-1.5 text-popover-foreground">
                <span
                  aria-hidden="true"
                  className="h-2.5 w-2.5 shrink-0 rounded-[2px]"
                  style={{ backgroundColor: colorFormatter(entry) }}
                />
                <span className="break-words">{name}</span>
              </span>
              <span className="font-semibold tabular-nums text-popover-foreground">
                {valueFormatter(entry.value, entry)}
              </span>
            </li>
          );
        })}
      </ul>
      {extraContent != null && extraContent !== false && (
        <div className="mt-2 border-t border-border pt-2 text-muted-foreground">
          {extraContent}
        </div>
      )}
    </div>
  );
}
