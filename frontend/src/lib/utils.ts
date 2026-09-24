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

// Presenca pode ser null: senador sem registro de votacao no periodo. Sem dado
// nao vira 0 (item 4 da auditoria).
export function formatPresenca(presenca: number | null | undefined, casas = 1): string {
  return presenca == null ? "—" : presenca.toFixed(casas);
}

export const SEM_DADOS_PRESENCA =
  "Sem registro de votação nominal no período: dados insuficientes para medir presença.";

// Primeiro ano com dados (início da 57ª legislatura)
export const ANO_INICIAL_DADOS = 2023;

/** Anos do seletor: do ano atual até 2023, sem lista fixa no código */
export function anosDisponiveis(hoje: Date = new Date()): number[] {
  const atual = Math.max(hoje.getFullYear(), ANO_INICIAL_DADOS);
  return Array.from({ length: atual - ANO_INICIAL_DADOS + 1 }, (_, i) => atual - i);
}
