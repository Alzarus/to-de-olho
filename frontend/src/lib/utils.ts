import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/**
 * Formata valores monetarios de forma legivel.
 * Usa notacao k para milhares e M para milhoes.
 * Ex: 150000 -> "R$ 150k", 1500000 -> "R$ 1,5M"
 */
export function formatCurrency(value: number): string {
  if (value >= 1_000_000) {
    const millions = value / 1_000_000;
    return `R$ ${millions.toLocaleString("pt-BR", {
      maximumFractionDigits: 1,
    })}M`;
  }
  if (value >= 1_000) {
    const thousands = value / 1_000;
    return `R$ ${thousands.toLocaleString("pt-BR", {
      maximumFractionDigits: 0,
    })}k`;
  }
  return `R$ ${value.toLocaleString("pt-BR", { maximumFractionDigits: 0 })}`;
}
