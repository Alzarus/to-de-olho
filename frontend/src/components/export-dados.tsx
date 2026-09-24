"use client";

import { useState } from "react";
import { Download, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  baixarArquivo,
  MIME_CSV,
  MIME_JSON,
  toCSV,
  toJSON,
  type ColunaCSV,
  type ValorCelula,
} from "@/lib/export";
import { cn } from "@/lib/utils";

/** Avisa o progresso ao buscar páginas: (página atual, total de páginas) */
export type AoProgredir = (atual: number, total: number) => void;

interface ExportarDadosProps<T> {
  /** O que está sendo exportado, para o menu e o JSON (ex.: "Despesas CEAPS") */
  descricao: string;
  /** Nome do arquivo sem extensão */
  nomeArquivo: string;
  colunas: readonly ColunaCSV<T>[];
  /** Busca o conjunto completo (todas as páginas, não só a visível) */
  obter: (aoProgredir: AoProgredir) => Promise<T[]>;
  /** Filtro aplicado, escrito no fim do CSV e nos parâmetros do JSON */
  observacao?: string;
  parametros?: Record<string, ValorCelula>;
  /** Texto de apoio no menu (ex.: "1.234 lançamentos, todas as páginas") */
  detalhe?: string;
  className?: string;
}

/** Botão "Exportar" com CSV (planilha) e JSON (pesquisa) e progresso anunciado */
export function ExportarDados<T>({
  descricao,
  nomeArquivo,
  colunas,
  obter,
  observacao,
  parametros,
  detalhe,
  className,
}: ExportarDadosProps<T>) {
  const [progresso, setProgresso] = useState<string | null>(null);
  const ocupado = progresso !== null;

  async function exportar(formato: "csv" | "json") {
    setProgresso("Preparando arquivo…");
    try {
      const dados = await obter((atual, total) => {
        if (total > 1) setProgresso(`Buscando página ${atual} de ${total}…`);
      });
      if (formato === "csv") {
        baixarArquivo(`${nomeArquivo}.csv`, toCSV(dados, colunas, { observacao }), MIME_CSV);
      } else {
        baixarArquivo(
          `${nomeArquivo}.json`,
          toJSON(dados, {
            descricao,
            parametros: { ...parametros, ...(observacao ? { filtro: observacao } : {}) },
          }),
          MIME_JSON,
        );
      }
      toast.success(`Arquivo gerado com ${dados.length} ${dados.length === 1 ? "linha" : "linhas"}.`);
    } catch (erro) {
      console.error(erro);
      toast.error("Não foi possível gerar o arquivo. Tente de novo em instantes.");
    } finally {
      setProgresso(null);
    }
  }

  return (
    <div className={cn("no-print flex items-center gap-2", className)}>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" size="sm" disabled={ocupado} aria-busy={ocupado}>
            {ocupado ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
            ) : (
              <Download className="mr-2 h-4 w-4" aria-hidden="true" />
            )}
            Exportar
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-64">
          <DropdownMenuLabel className="font-normal">
            <span className="block font-medium">{descricao}</span>
            {detalhe && <span className="block text-xs text-muted-foreground">{detalhe}</span>}
          </DropdownMenuLabel>
          <DropdownMenuItem onSelect={() => exportar("csv")}>Planilha (CSV)</DropdownMenuItem>
          <DropdownMenuItem onSelect={() => exportar("json")}>Dados para pesquisa (JSON)</DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <span role="status" aria-live="polite" className="text-xs text-muted-foreground">
        {progresso}
      </span>
    </div>
  );
}
