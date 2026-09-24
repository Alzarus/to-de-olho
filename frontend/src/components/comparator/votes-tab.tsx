"use client";

import { useState } from "react";
import Link from "next/link";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { AlertCircle } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import type { SenatorBasicProfile } from "@/contexts/comparator-context";
import { getAlinhamento, type Divergencia, type ParAlinhamento } from "@/lib/alinhamento";
import { getVotacoesFacetas } from "@/services/votacaoService";
import { corSerie, lerFiltrosVotacoes, passoRampa, PARAMS_VOTACOES, PISO_RAMPA } from "@/lib/comparador-filtros";
import {
  BarraFiltros,
  FiltroMulti,
  FiltroSelect,
  formatarPct,
  MarcadorSerie,
  useFiltrosUrl,
} from "@/components/comparator/filtros";
import { cn } from "@/lib/utils";

interface VotesTabProps {
  senators: SenatorBasicProfile[];
  year: number;
}

const QUALQUER_PAR = "__qualquer__";
const POR_PAGINA = 20;
const PASSOS = [1, 2, 3, 4, 5, 6, 7] as const;

const chavePar = (a: number, b: number) => `${Math.min(a, b)}-${Math.max(a, b)}`;

function votoDe(d: Divergencia, id: number): string | undefined {
  return d.votos.find((v) => v.senador_id === id)?.voto;
}

/** O par votou (Sim/Não/Abstenção) e divergiu nesta votação. */
function parDivergiu(d: Divergencia, [a, b]: [number, number]): boolean {
  const va = votoDe(d, a);
  const vb = votoDe(d, b);
  return va !== undefined && vb !== undefined && va !== vb;
}

function formatarData(iso: string): string {
  const d = new Date(iso);
  return Number.isNaN(d.getTime()) ? "" : d.toLocaleDateString("pt-BR", { timeZone: "UTC" });
}

export function VotesTab({ senators, year }: VotesTabProps) {
  const { params, definir } = useFiltrosUrl();
  const filtros = lerFiltrosVotacoes(params);
  const ids = senators.map((s) => s.id);
  const apiYear = year === 0 ? undefined : year;
  const [visiveis, setVisiveis] = useState(POR_PAGINA);

  const facetas = useQuery({
    queryKey: ["votacoes-facetas", apiYear],
    queryFn: () => getVotacoesFacetas(apiYear),
  });

  const alinhamento = useQuery({
    queryKey: ["votacoes-alinhamento", [...ids].sort((a, b) => a - b).join(","), apiYear, filtros.tipos.join(",")],
    queryFn: () => getAlinhamento(ids, apiYear, filtros.tipos),
    enabled: ids.length >= 2,
    placeholderData: keepPreviousData,
    retry: 1,
  });

  if (senators.length < 2) {
    return (
      <Card>
        <CardContent className="py-12 text-center text-muted-foreground">
          Selecione pelo menos dois senadores para comparar como votam.
        </CardContent>
      </Card>
    );
  }

  const opcoesTipo = (facetas.data?.tipos ?? []).map((t) => ({
    valor: t.valor,
    rotulo: `${t.valor} (${t.total.toLocaleString("pt-BR")} ${t.total === 1 ? "votação" : "votações"})`,
  }));
  for (const t of filtros.tipos) {
    if (!opcoesTipo.some((o) => o.valor === t)) opcoesTipo.push({ valor: t, rotulo: t });
  }

  const pares = senators.flatMap((a, i) => senators.slice(i + 1).map((b) => [a, b] as const));
  const opcoesPar = [
    { valor: QUALQUER_PAR, rotulo: "Qualquer par" },
    ...pares.map(([a, b]) => ({ valor: chavePar(a.id, b.id), rotulo: `${a.nome} × ${b.nome}` })),
  ];
  const parSelecionado =
    filtros.par && ids.includes(filtros.par[0]) && ids.includes(filtros.par[1]) ? filtros.par : null;

  const yearLabel = year === 0 ? "todos os anos" : String(year);
  const recorteTipos = filtros.tipos.length > 0 ? ` · matérias: ${filtros.tipos.join(", ")}` : "";

  const barra = (
    <BarraFiltros rotulo="Filtros de votações">
      <FiltroMulti
        rotulo="Tipo de matéria"
        opcoes={opcoesTipo}
        selecionados={filtros.tipos}
        resumo={filtros.tipos.length === 0 ? "Todos os tipos" : undefined}
        aoMudar={(valores) => {
          setVisiveis(POR_PAGINA);
          definir({ [PARAMS_VOTACOES.tipos]: valores.length > 0 ? valores.join(",") : null });
        }}
        aoRestaurar={() => definir({ [PARAMS_VOTACOES.tipos]: null })}
        rotuloRestaurar="Todos os tipos"
      />
      <FiltroSelect
        rotulo="Divergências do par"
        valor={parSelecionado ? chavePar(...parSelecionado) : QUALQUER_PAR}
        largura="w-[300px]"
        opcoes={opcoesPar}
        aoMudar={(v) => {
          setVisiveis(POR_PAGINA);
          definir({ [PARAMS_VOTACOES.par]: v === QUALQUER_PAR ? null : v });
        }}
      />
    </BarraFiltros>
  );

  if (alinhamento.isLoading) {
    return (
      <div className="space-y-6">
        {barra}
        <Skeleton className="h-[480px] w-full rounded-lg" />
      </div>
    );
  }

  if (alinhamento.isError || !alinhamento.data) {
    return (
      <div className="space-y-6">
        {barra}
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Erro</AlertTitle>
          <AlertDescription>Não foi possível calcular o alinhamento de votos.</AlertDescription>
        </Alert>
      </div>
    );
  }

  const dados = alinhamento.data;
  const porPar = new Map<string, ParAlinhamento>(dados.pares.map((p) => [chavePar(p.senador_a, p.senador_b), p]));
  const divergencias = parSelecionado ? dados.divergencias.filter((d) => parDivergiu(d, parSelecionado)) : dados.divergencias;
  const truncadas = dados.total_divergencias > dados.divergencias.length;
  const nomeSenador = new Map(senators.map((s) => [s.id, s.nome]));

  return (
    <div className={cn("space-y-6", alinhamento.isFetching && "opacity-70 transition-opacity")}>
      {barra}

      <Card>
        <CardHeader>
          <CardTitle>Alinhamento de votos ({yearLabel})</CardTitle>
          <CardDescription>
            Percentual de votações nominais abertas em que os dois votaram igual{recorteTipos}. Contam só as votações
            em que ambos votaram Sim, Não ou Abstenção; votações secretas e presenças sem voto ficam de fora.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="border-separate border-spacing-[2px] text-sm">
              <caption className="sr-only">
                Matriz de alinhamento: cada célula traz o percentual de votos iguais entre o senador da linha e o da
                coluna e o número de votações em comum.
              </caption>
              <thead>
                <tr>
                  <td />
                  {senators.map((s, i) => (
                    <th key={s.id} scope="col" className="min-w-[96px] max-w-[140px] px-2 pb-2 align-bottom text-xs font-semibold">
                      <span className="flex items-center justify-center gap-1.5">
                        <MarcadorSerie cor={corSerie(i)} />
                        <span className="line-clamp-2">{s.nome}</span>
                      </span>
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {senators.map((linha, i) => (
                  <tr key={linha.id}>
                    <th scope="row" className="max-w-[160px] pr-3 text-left text-xs font-semibold">
                      <span className="flex items-center gap-1.5">
                        <MarcadorSerie cor={corSerie(i)} />
                        <span className="line-clamp-2">{linha.nome}</span>
                      </span>
                    </th>
                    {senators.map((coluna) => {
                      if (coluna.id === linha.id) {
                        return (
                          <td key={coluna.id} className="h-16 rounded-md bg-muted/40 text-center text-muted-foreground">
                            <span aria-label="mesmo senador">—</span>
                          </td>
                        );
                      }
                      const p = porPar.get(chavePar(linha.id, coluna.id));
                      if (!p || p.percentual === null) {
                        return (
                          <td key={coluna.id} className="h-16 rounded-md border border-dashed text-center text-xs text-muted-foreground">
                            sem votações em comum
                          </td>
                        );
                      }
                      const passo = passoRampa(p.percentual);
                      return (
                        <td
                          key={coluna.id}
                          className="h-16 rounded-md px-2 text-center"
                          style={{
                            backgroundColor: `var(--seq-${passo})`,
                            color: passo <= 4 ? "var(--seq-texto-baixo)" : "var(--seq-texto-alto)",
                          }}
                        >
                          <span className="block text-base font-bold tabular-nums">{formatarPct(p.percentual)}</span>
                          <span className="block text-[11px] tabular-nums">
                            {p.votos_iguais} de {p.votacoes_comuns}
                          </span>
                        </td>
                      );
                    })}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="mt-4 flex items-center gap-2 text-xs text-muted-foreground" aria-hidden="true">
            <span>{PISO_RAMPA}% ou menos</span>
            <span className="flex">
              {PASSOS.map((n) => (
                <span key={n} className="h-3 w-6 first:rounded-l last:rounded-r" style={{ backgroundColor: `var(--seq-${n})` }} />
              ))}
            </span>
            <span>100% de votos iguais</span>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Votações em que divergiram</CardTitle>
          <CardDescription>
            {parSelecionado
              ? `Votações abertas em que ${nomeSenador.get(parSelecionado[0])} e ${nomeSenador.get(parSelecionado[1])} votaram diferente`
              : "Votações abertas em que ao menos dois dos senadores votaram diferente"}
            {recorteTipos}. {divergencias.length.toLocaleString("pt-BR")}{" "}
            {divergencias.length === 1 ? "votação" : "votações"}
            {truncadas && ` (das ${dados.divergencias.length} mais recentes de ${dados.total_divergencias.toLocaleString("pt-BR")})`}.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {divergencias.length === 0 ? (
            <p className="py-8 text-center text-sm text-muted-foreground">Nenhuma divergência no recorte.</p>
          ) : (
            <>
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <caption className="sr-only">Votações com divergência e o voto de cada senador</caption>
                  <thead>
                    <tr className="border-b text-left">
                      <th scope="col" className="px-2 py-2 font-semibold">Data</th>
                      <th scope="col" className="px-2 py-2 font-semibold">Votação</th>
                      {senators.map((s, i) => (
                        <th key={s.id} scope="col" className="px-2 py-2 font-semibold">
                          <span className="flex items-center gap-1.5 whitespace-nowrap">
                            <MarcadorSerie cor={corSerie(i)} />
                            {s.nome}
                          </span>
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {divergencias.slice(0, visiveis).map((d) => (
                      <tr key={d.codigo_votacao} className="border-b align-top last:border-0">
                        <td className="whitespace-nowrap px-2 py-2 tabular-nums text-muted-foreground">{formatarData(d.data)}</td>
                        <th scope="row" className="px-2 py-2 text-left font-normal">
                          <Link href={`/votacoes/${d.codigo_votacao}`} className="font-medium text-primary hover:underline">
                            {d.materia || d.sigla_materia || `Votação ${d.codigo_votacao}`}
                          </Link>
                          {d.descricao_votacao && (
                            <p className="mt-0.5 line-clamp-2 text-xs text-muted-foreground">{d.descricao_votacao}</p>
                          )}
                        </th>
                        {senators.map((s) => {
                          const voto = votoDe(d, s.id);
                          const destaque = parSelecionado?.includes(s.id);
                          return (
                            <td key={s.id} className={cn("px-2 py-2 whitespace-nowrap", destaque && "font-semibold")}>
                              {voto ?? <span className="text-muted-foreground" title="Sem voto Sim/Não/Abstenção">—</span>}
                            </td>
                          );
                        })}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              {divergencias.length > visiveis && (
                <div className="mt-4 flex justify-center">
                  <Button variant="outline" onClick={() => setVisiveis((v) => v + POR_PAGINA)}>
                    Mostrar mais {Math.min(POR_PAGINA, divergencias.length - visiveis)}
                  </Button>
                </div>
              )}
              <p className="mt-3 text-xs text-muted-foreground">“—”: sem voto Sim, Não ou Abstenção nessa votação.</p>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
