"use client";

import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Download, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  getDespesasAgregado,
  getDespesasFornecedores,
  getDespesasMensal,
  getEmendas,
} from "@/lib/api";
import {
  baixarArquivo,
  MIME_CSV,
  MIME_JSON,
  rotuloAnoArquivo,
  toCSV,
  toJSON,
  type ColunaCSV,
} from "@/lib/export";
import {
  COLUNAS_EMENDA,
  COLUNAS_FORNECEDOR,
  COLUNAS_RANKING,
  NOME_MES,
  linhasRanking,
  type ComSenador,
  type LinhaRanking,
} from "@/lib/export-colunas";
import type { SenatorBasicProfile } from "@/contexts/comparator-context";
import type {
  DespesasAgregadoResponse,
  DespesasFornecedoresResponse,
  DespesasMensalResponse,
  EmendasResponse,
  RankingResponse,
} from "@/types/api";

interface ExportMenuProps {
  senators: SenatorBasicProfile[];
  /** 0 = mandato completo */
  year: number;
  rankingData?: RankingResponse;
}

interface LinhaMensal extends ComSenador {
  ano: number;
  mes: number;
  total: number;
}

interface LinhaCategoria extends ComSenador {
  tipo: string;
  total: number;
}

const COLUNAS_MENSAL: ColunaCSV<LinhaMensal>[] = [
  { cabecalho: "Senador", valor: (l) => l.senador },
  { cabecalho: "Ano", valor: (l) => l.ano },
  { cabecalho: "Mês", valor: (l) => l.mes },
  { cabecalho: "Nome do mês", valor: (l) => NOME_MES[l.mes - 1] ?? "" },
  { cabecalho: "Gasto CEAPS (R$)", valor: (l) => l.total },
];

const COLUNAS_CATEGORIA: ColunaCSV<LinhaCategoria>[] = [
  { cabecalho: "Senador", valor: (l) => l.senador },
  { cabecalho: "Tipo de despesa", valor: (l) => l.tipo },
  { cabecalho: "Total (R$)", valor: (l) => l.total },
];

type Conjunto = "resumo" | "mensal" | "categorias" | "fornecedores" | "emendas" | "json";

/**
 * Exporta cada aba do comparador com o conjunto completo (agregados do
 * backend e lista inteira de emendas), não só o que aparece no gráfico.
 * Reaproveita o cache do React Query das abas (mesmas chaves).
 */
export function ComparatorExportMenu({ senators, year, rankingData }: ExportMenuProps) {
  const queryClient = useQueryClient();
  const [progresso, setProgresso] = useState<string | null>(null);
  const apiYear = year === 0 ? undefined : year;
  const sufixo = rotuloAnoArquivo(apiYear);
  const periodo = apiYear ? `ano ${apiYear}` : "mandato completo";

  // Busca um dado por senador, em sequência, dizendo em qual está
  async function porSenador<T>(
    rotulo: string,
    chave: string,
    buscar: (id: number) => Promise<T>,
    anoChave: number | undefined = apiYear,
  ): Promise<{ senador: SenatorBasicProfile; dados: T }[]> {
    const saida: { senador: SenatorBasicProfile; dados: T }[] = [];
    for (const [i, s] of senators.entries()) {
      setProgresso(`${rotulo}: ${i + 1} de ${senators.length} senadores…`);
      const dados = await queryClient.fetchQuery({
        queryKey: [chave, s.id, anoChave],
        queryFn: () => buscar(s.id),
      });
      saida.push({ senador: s, dados });
    }
    return saida;
  }

  function linhasResumo(): { linhas: LinhaRanking[]; faltando: string[] } {
    const todas = linhasRanking(rankingData?.ranking ?? [], rankingData?.sem_dados ?? []);
    const linhas: LinhaRanking[] = [];
    const faltando: string[] = [];
    for (const s of senators) {
      const linha = todas.find((l) => l.senador_id === s.id);
      if (linha) linhas.push(linha);
      else faltando.push(s.nome);
    }
    return { linhas, faltando };
  }

  const mensal = () =>
    porSenador<DespesasMensalResponse>("Despesas por mês", "senador-despesas-mensal", (id) =>
      getDespesasMensal(id, apiYear),
    );
  const categorias = () =>
    porSenador<DespesasAgregadoResponse>(
      "Despesas por categoria",
      "senador-despesas-agregado",
      (id) => getDespesasAgregado(id, apiYear),
    );
  const fornecedores = () =>
    porSenador<DespesasFornecedoresResponse>(
      "Fornecedores",
      "senador-despesas-fornecedores",
      (id) => getDespesasFornecedores(id, apiYear),
    );
  // A aba de emendas usa o ano cru (0 = mandato) na chave
  const emendas = () =>
    porSenador<EmendasResponse>("Emendas", "senador-emendas", (id) => getEmendas(id, year), year);

  async function exportar(conjunto: Conjunto) {
    try {
      switch (conjunto) {
        case "resumo": {
          const { linhas, faltando } = linhasResumo();
          const observacao = faltando.length
            ? `Sem dados no ranking (${periodo}): ${faltando.join(", ")}`
            : undefined;
          baixarArquivo(
            `todeolho_comparacao_resumo_${sufixo}.csv`,
            toCSV(linhas, COLUNAS_RANKING, { observacao }),
            MIME_CSV,
          );
          break;
        }
        case "mensal": {
          const dados = await mensal();
          const linhas: LinhaMensal[] = dados.flatMap(({ senador, dados: d }) =>
            (d.meses ?? []).map((m) => ({ senador: senador.nome, ano: m.ano, mes: m.mes, total: m.total })),
          );
          baixarArquivo(
            `todeolho_comparacao_despesas_mensais_${sufixo}.csv`,
            toCSV(linhas, COLUNAS_MENSAL),
            MIME_CSV,
          );
          break;
        }
        case "categorias": {
          const dados = await categorias();
          const linhas: LinhaCategoria[] = dados.flatMap(({ senador, dados: d }) =>
            (d.por_tipo ?? []).map((t) => ({ senador: senador.nome, tipo: t.tipo_despesa, total: t.total })),
          );
          baixarArquivo(
            `todeolho_comparacao_despesas_categorias_${sufixo}.csv`,
            toCSV(linhas, COLUNAS_CATEGORIA),
            MIME_CSV,
          );
          break;
        }
        case "fornecedores": {
          const dados = await fornecedores();
          const linhas = dados.flatMap(({ senador, dados: d }) =>
            (d.fornecedores ?? []).map((f) => ({ ...f, senador: senador.nome })),
          );
          baixarArquivo(
            `todeolho_comparacao_fornecedores_${sufixo}.csv`,
            toCSV(linhas, COLUNAS_FORNECEDOR),
            MIME_CSV,
          );
          break;
        }
        case "emendas": {
          const dados = await emendas();
          const linhas = dados.flatMap(({ senador, dados: d }) =>
            (d.emendas ?? []).map((e) => ({ ...e, senador: senador.nome })),
          );
          baixarArquivo(
            `todeolho_comparacao_emendas_${sufixo}.csv`,
            toCSV(linhas, COLUNAS_EMENDA),
            MIME_CSV,
          );
          break;
        }
        case "json": {
          const { linhas } = linhasResumo();
          const [m, c, f, e] = [await mensal(), await categorias(), await fornecedores(), await emendas()];
          const dados = senators.map((s, i) => ({
            senador_id: s.id,
            nome: s.nome,
            partido: s.partido,
            uf: s.uf,
            ranking: linhas.find((l) => l.senador_id === s.id) ?? null,
            despesas_mensais: m[i].dados.meses ?? [],
            despesas_por_categoria: c[i].dados.por_tipo ?? [],
            fornecedores: f[i].dados.fornecedores ?? [],
            emendas: e[i].dados.emendas ?? [],
          }));
          baixarArquivo(
            `todeolho_comparacao_${sufixo}.json`,
            toJSON(dados, {
              descricao: "Comparação de senadores: ranking, despesas CEAPS e emendas",
              parametros: { ano: apiYear ?? "mandato", ids: senators.map((s) => s.id).join(",") },
            }),
            MIME_JSON,
          );
          break;
        }
      }
      toast.success("Arquivo gerado.");
    } catch (erro) {
      console.error(erro);
      toast.error("Não foi possível gerar o arquivo. Tente de novo em instantes.");
    } finally {
      setProgresso(null);
    }
  }

  const ocupado = progresso !== null;

  return (
    <div className="flex flex-1 items-center gap-2 sm:flex-none">
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="outline"
            size="sm"
            className="flex-1 sm:flex-none"
            disabled={senators.length === 0 || ocupado}
            aria-busy={ocupado}
          >
            {ocupado ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
            ) : (
              <Download className="mr-2 h-4 w-4" aria-hidden="true" />
            )}
            Exportar
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-72">
          <DropdownMenuLabel>Planilha (CSV) — {periodo}</DropdownMenuLabel>
          <DropdownMenuItem onSelect={() => exportar("resumo")} disabled={!rankingData}>
            Resumo do ranking
          </DropdownMenuItem>
          <DropdownMenuItem onSelect={() => exportar("mensal")}>Despesas por mês</DropdownMenuItem>
          <DropdownMenuItem onSelect={() => exportar("categorias")}>
            Despesas por categoria
          </DropdownMenuItem>
          <DropdownMenuItem onSelect={() => exportar("fornecedores")}>
            Fornecedores (lista completa)
          </DropdownMenuItem>
          <DropdownMenuItem onSelect={() => exportar("emendas")}>
            Emendas (lista completa)
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuLabel>Para pesquisa</DropdownMenuLabel>
          <DropdownMenuItem onSelect={() => exportar("json")}>
            Todos os dados (JSON)
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <span role="status" aria-live="polite" className="text-xs text-muted-foreground">
        {progresso}
      </span>
    </div>
  );
}
