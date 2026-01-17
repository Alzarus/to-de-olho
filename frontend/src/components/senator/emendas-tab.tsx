"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from "@/components/ui/tooltip";
import { useEmendas } from "@/hooks/use-senador";
import { formatCurrency } from "@/lib/utils";
import { AlertCircle, Info } from "lucide-react";
import { BrazilMap } from "@/components/ui/brazil-map";

export function EmendasTab({ id, ano }: { id: number; ano: number }) {
  const { data, isLoading } = useEmendas(id, ano);

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="grid gap-4 md:grid-cols-3">
             <Skeleton className="h-32 w-full" />
             <Skeleton className="h-32 w-full" />
             <Skeleton className="h-32 w-full" />
        </div>
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (!data || data.emendas.length === 0) {
      return (
          <Card>
              <CardContent className="flex flex-col items-center justify-center py-12 text-muted-foreground">
                  <AlertCircle className="h-12 w-12 mb-4 opacity-20" />
                  <p>Nenhuma emenda encontrada para este período.</p>
                  <p className="text-xs mt-2">Os dados podem estar em processo de importação.</p>
              </CardContent>
          </Card>
      );
  }

    const { resumo, emendas } = data;
  
  // Filtrar PIX
    const emendasPix = emendas.filter((e) => isEmendaEspecial(e.tipo));
  const totalPix = emendasPix.reduce((acc, curr) => acc + curr.valor_pago, 0);

    const destinosPorUF = (resumo?.top_localidades || []).reduce<Record<string, number>>(
        (acc, item) => {
            const uf = extrairUF(item.localidade);
            if (!uf) return acc;
            acc[uf] = (acc[uf] || 0) + item.valor;
            return acc;
        },
        {},
    );
    const destinos = Object.entries(destinosPorUF).map(([uf, valor]) => ({ uf, valor }));
    const maxValor = destinos.reduce((max, item) => Math.max(max, item.valor), 0) || 1;

    return (
    <div className="space-y-6">
      {/* Resumo */}
            <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
                <Card>
                        <CardHeader className="pb-2">
                                <CardTitle className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                                    Total Pago
                                    <TooltipProvider delayDuration={0}>
                                        <Tooltip>
                                            <TooltipTrigger asChild>
                                                <button
                                                    type="button"
                                                    aria-label="Entenda o total pago"
                                                    className="text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                                                >
                                                    <Info className="h-4 w-4" />
                                                </button>
                                            </TooltipTrigger>
                                            <TooltipContent side="top" className="max-w-[220px] text-xs">
                                                Soma dos valores efetivamente pagos (inclui restos a pagar quitados).
                                            </TooltipContent>
                                        </Tooltip>
                                    </TooltipProvider>
                                </CardTitle>
                        </CardHeader>
                        <CardContent>
                                <div className="text-2xl font-bold">{formatCurrency(resumo?.total_pago || 0)}</div>
                                <p className="text-xs text-muted-foreground">de um total empenhado de {formatCurrency(resumo?.total_empenhado || 0)}</p>
                        </CardContent>
                </Card>
        <Card>
            <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">Emendas PIX (Pago)</CardTitle>
            </CardHeader>
            <CardContent>
                <div className="text-2xl font-bold text-orange-600">{formatCurrency(totalPix)}</div>
                <p className="text-xs text-muted-foreground">{emendasPix.length} emendas especiais</p>
            </CardContent>
        </Card>
        <Card>
            <CardHeader className="pb-2">
                                <CardTitle className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                                    Total Empenhado
                                    <TooltipProvider delayDuration={0}>
                                        <Tooltip>
                                            <TooltipTrigger asChild>
                                                <button
                                                    type="button"
                                                    aria-label="Entenda o total empenhado"
                                                    className="text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                                                >
                                                    <Info className="h-4 w-4" />
                                                </button>
                                            </TooltipTrigger>
                                            <TooltipContent side="top" className="max-w-[220px] text-xs">
                                                Soma dos valores reservados no orçamento (empenhos), mesmo que ainda não pagos.
                                            </TooltipContent>
                                        </Tooltip>
                                    </TooltipProvider>
                                </CardTitle>
            </CardHeader>
            <CardContent>
                                <div className="text-2xl font-bold">{formatCurrency(resumo?.total_empenhado || 0)}</div>
                                <p className="text-xs text-muted-foreground">{resumo?.quantidade || 0} emendas executadas</p>
            </CardContent>
        </Card>
      </div>

            {/* Mapa simplificado de destinos */}
            {resumo?.top_localidades && resumo.top_localidades.length > 0 && (
                <Card>
                    <CardHeader>
                        <CardTitle>Mapa de Destinos dos Recursos</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="grid grid-cols-1 gap-6 lg:grid-cols-[2fr_1fr]">
                        <div className="relative w-full h-full min-h-[300px] flex items-center justify-center bg-muted/20 rounded-lg border p-4">
                            <BrazilMap 
                                data={destinos.map(d => ({ uf: d.uf, value: d.valor }))}
                                className="w-full h-full"
                            />
                        </div>
                            <div className="space-y-4">
                                <div className="text-xs font-semibold uppercase text-muted-foreground">
                                  Cidades/UF com maior volume
                                </div>
                                {resumo.top_localidades.map((loc, i) => (
                                    <div key={i} className="flex items-center">
                                        <div className="w-full flex-1">
                                            <div className="flex items-center justify-between mb-1 gap-3">
                                                <span className="text-sm font-medium break-words">{loc.localidade}</span>
                                                <span className="text-sm font-bold whitespace-nowrap">
                                                  {formatCurrency(loc.valor)}
                                                </span>
                                            </div>
                                            <div className="h-2 w-full bg-muted rounded-full overflow-hidden">
                                                <div
                                                    className="h-full bg-blue-600 rounded-full"
                                                    style={{ width: `${(loc.valor / (resumo.top_localidades[0].valor)) * 100}%` }}
                                                />
                                            </div>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </CardContent>
                </Card>
            )}

            {/* Lista de Emendas */}
      <Card>
          <CardHeader>
              <CardTitle>Detalhamento das Emendas</CardTitle>
          </CardHeader>
          <CardContent>
              <div className="space-y-4">
                                    {emendas.slice(0, 50).map((emenda) => (
                                        <details key={emenda.id} className="group rounded-lg border p-3 sm:p-4">
                                            <summary className="flex cursor-pointer flex-col gap-3 sm:flex-row sm:items-start sm:justify-between list-none">
                                                <div className="min-w-0 space-y-1">
                                                    <div className="flex flex-wrap items-center gap-2">
                                                        <span className="font-mono text-xs text-muted-foreground">{emenda.numero}</span>
                                                        <Badge
                                                            variant={
                                                                isEmendaEspecial(emenda.tipo)
                                                                    ? "destructive"
                                                                    : "outline"
                                                            }
                                                            className="text-xs whitespace-normal h-auto text-left leading-tight"
                                                        >
                                                            {emenda.tipo}
                                                        </Badge>
                                                    </div>
                                                    <p className="font-medium text-sm break-words">{emenda.localidade}</p>
                                                    <p className="text-xs text-muted-foreground break-words">{emenda.funcional_programatica}</p>
                                                </div>
                                                <div className="text-left sm:text-right">
                                                    <p className="font-bold text-sm">{formatCurrency(emenda.valor_pago)}</p>
                                                    <p className="text-xs text-muted-foreground">Empenhado: {formatCurrency(emenda.valor_empenhado)}</p>
                                                </div>
                                            </summary>
                                            <div className="mt-4 grid gap-3 text-xs text-muted-foreground sm:grid-cols-2 sm:text-sm">
                                                <div>
                                                    <span className="font-medium text-foreground">Tipo:</span> {emenda.tipo}
                                                </div>
                                                <div>
                                                    <span className="font-medium text-foreground">Ano:</span> {emenda.ano}
                                                </div>
                                                <div>
                                                    <span className="font-medium text-foreground">Localidade:</span> {emenda.localidade}
                                                </div>
                                                <div>
                                                    <span className="font-medium text-foreground">Função:</span> {emenda.funcional_programatica}
                                                </div>
                                                <div>
                                                    <span className="font-medium text-foreground">Empenhado:</span> {formatCurrency(emenda.valor_empenhado)}
                                                </div>
                                                <div>
                                                    <span className="font-medium text-foreground">Pago:</span> {formatCurrency(emenda.valor_pago)}
                                                </div>
                                            </div>
                                        </details>
                                    ))}
                  {emendas.length > 50 && (
                      <p className="text-center text-sm text-muted-foreground pt-4">Exibindo as 50 maiores emendas de um total de {emendas.length}</p>
                  )}
              </div>
          </CardContent>
      </Card>
    </div>
  );
}

const UF_POSICOES: Record<string, { x: number; y: number }> = {
    AC: { x: 18, y: 58 },
    AL: { x: 78, y: 60 },
    AP: { x: 46, y: 14 },
    AM: { x: 30, y: 30 },
    BA: { x: 70, y: 58 },
    CE: { x: 78, y: 40 },
    DF: { x: 58, y: 52 },
    ES: { x: 74, y: 66 },
    GO: { x: 56, y: 54 },
    MA: { x: 64, y: 40 },
    MT: { x: 46, y: 52 },
    MS: { x: 48, y: 64 },
    MG: { x: 64, y: 62 },
    PA: { x: 48, y: 30 },
    PB: { x: 80, y: 50 },
    PR: { x: 50, y: 76 },
    PE: { x: 78, y: 54 },
    PI: { x: 68, y: 46 },
    RJ: { x: 68, y: 70 },
    RN: { x: 82, y: 44 },
    RS: { x: 50, y: 86 },
    RO: { x: 28, y: 52 },
    RR: { x: 30, y: 18 },
    SC: { x: 52, y: 80 },
    SP: { x: 58, y: 74 },
    SE: { x: 76, y: 62 },
    TO: { x: 56, y: 40 },
};

const UF_POR_NOME: Record<string, string> = {
    ACRE: "AC",
    ALAGOAS: "AL",
    AMAPA: "AP",
    AMAPÁ: "AP",
    AMAZONAS: "AM",
    BAHIA: "BA",
    CEARA: "CE",
    CEARÁ: "CE",
    DISTRITO_FEDERAL: "DF",
    ESPIRITO_SANTO: "ES",
    ESPÍRITO_SANTO: "ES",
    GOIAS: "GO",
    GOIÁS: "GO",
    MARANHAO: "MA",
    MARANHÃO: "MA",
    MATO_GROSSO: "MT",
    MATO_GROSSO_DO_SUL: "MS",
    MINAS_GERAIS: "MG",
    PARA: "PA",
    PARÁ: "PA",
    PARAIBA: "PB",
    PARAÍBA: "PB",
    PARANA: "PR",
    PARANÁ: "PR",
    PERNAMBUCO: "PE",
    PIAUI: "PI",
    PIAUÍ: "PI",
    RIO_DE_JANEIRO: "RJ",
    RIO_GRANDE_DO_NORTE: "RN",
    RIO_GRANDE_DO_SUL: "RS",
    RONDONIA: "RO",
    RONDÔNIA: "RO",
    RORAIMA: "RR",
    SANTA_CATARINA: "SC",
    SAO_PAULO: "SP",
    SÃO_PAULO: "SP",
    SERGIPE: "SE",
    TOCANTINS: "TO",
};

function extrairUF(localidade: string): string | null {
    if (!localidade) return null;
    const upper = localidade.toUpperCase();
    if (upper.includes("NACIONAL")) return null;

    const match = upper.match(/\b([A-Z]{2})\b/);
    if (match && match[1] !== "UF") {
        return match[1];
    }

    const normalizada = normalizarLocalidade(upper);
    for (const [nome, uf] of Object.entries(UF_POR_NOME)) {
        if (normalizada.includes(nome)) {
            return uf;
        }
    }
    return null;
}

function normalizarLocalidade(valor: string): string {
    return valor
        .normalize("NFD")
        .replace(/[\u0300-\u036f]/g, "")
        .replace(/[\s-]+/g, "_")
        .trim();
}

function isEmendaEspecial(tipo: string): boolean {
    const normalizado = tipo.toLowerCase();
    // Buscar por "especia" para capturar tanto "especial" quanto "especiais"
    return normalizado.includes("especia") || normalizado.includes("pix");
}
