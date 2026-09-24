"use client";

import { useQueries } from "@tanstack/react-query";
import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { AlertCircle, ExternalLink, Info } from "lucide-react";
import { getGabinete } from "@/lib/api";
import type { SenatorBasicProfile } from "@/contexts/comparator-context";
import type { GabineteResponse } from "@/types/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { ChartTooltipContent } from "@/components/ui/chart-tooltip";
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  NOTA_PRIVACIDADE,
  corVinculo,
  notaMesa,
  ordenarVinculos,
} from "@/components/senator/gabinete-cores";

interface CabinetTabProps {
  senators: SenatorBasicProfile[];
  year: number; // 0 = mandato: usa o ano mais recente de cada senador
}

const ROTULO_CURTO: Record<string, string> = {
  GABINETE: "Gabinete",
  ESCRITORIO: "Escritórios",
};

const plural = (n: number) => `${n.toLocaleString("pt-BR")} ${n === 1 ? "servidor" : "servidores"}`;

function primeiroNome(nome: string, max = 18): string {
  return nome.length > max ? `${nome.slice(0, max - 1)}…` : nome;
}

function totalLocal(g: GabineteResponse | undefined, local: string): number {
  return g?.locais.find((l) => l.local === local)?.total ?? 0;
}

export function CabinetTab({ senators, year }: CabinetTabProps) {
  const queries = useQueries({
    queries: senators.map((s) => ({
      queryKey: ["senador-gabinete", s.id, year],
      queryFn: () => getGabinete(s.id, year || undefined),
    })),
  });

  if (queries.some((q) => q.isLoading)) {
    return <Skeleton className="h-[500px] w-full rounded-lg" />;
  }
  if (queries.every((q) => q.isError)) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" aria-hidden="true" />
        <AlertTitle>Erro</AlertTitle>
        <AlertDescription>Não foi possível carregar a estrutura de gabinete.</AlertDescription>
      </Alert>
    );
  }

  const linhas = senators.map((s, i) => ({ senador: s, dados: queries[i].data }));
  const vinculos = ordenarVinculos(
    linhas.flatMap((l) => l.dados?.locais.flatMap((loc) => loc.vinculos.map((v) => v.vinculo)) ?? []),
  );

  // Uma barra por senador e local, empilhada por vínculo
  const dadosGrafico = linhas.flatMap(({ senador, dados }) =>
    ["GABINETE", "ESCRITORIO"].map((local) => {
      const loc = dados?.locais.find((l) => l.local === local);
      const linha: Record<string, string | number> = {
        rotulo: `${primeiroNome(senador.nome)} · ${ROTULO_CURTO[local]}`,
        titulo: `${senador.nome} · ${loc?.rotulo ?? ROTULO_CURTO[local]}${dados?.ano ? ` (${dados.ano})` : ""}`,
      };
      for (const v of loc?.vinculos ?? []) linha[v.vinculo] = v.quantidade;
      return linha;
    }),
  );
  const algumDado = linhas.some((l) => (l.dados?.total ?? 0) > 0);
  const anosDiferentes = new Set(linhas.map((l) => l.dados?.ano).filter((a) => a)).size > 1;
  const daMesa = linhas.filter((l) => l.dados?.cargo_mesa);
  const resumoGrafico = linhas
    .map(
      ({ senador, dados }) =>
        `${senador.nome}: gabinete ${totalLocal(dados, "GABINETE")}, escritórios ${totalLocal(dados, "ESCRITORIO")}`,
    )
    .join("; ");

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>
            Estrutura de Gabinete ({year === 0 ? "ano mais recente disponível" : year})
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {daMesa.length > 0 && (
            <Alert>
              <Info className="h-4 w-4" aria-hidden="true" />
              <AlertTitle>Integrantes da Mesa Diretora</AlertTitle>
              <AlertDescription>
                <ul className="list-disc space-y-1 pl-4">
                  {daMesa.map(({ senador, dados }) => (
                    <li key={senador.id}>
                      <strong>{senador.nome}</strong>: {notaMesa(dados!.cargo_mesa!)}
                    </li>
                  ))}
                </ul>
              </AlertDescription>
            </Alert>
          )}
          {anosDiferentes && (
            <p className="text-sm text-muted-foreground">
              Os senadores têm dados em anos diferentes; o ano de cada um aparece na tabela.
            </p>
          )}

          {!algumDado ? (
            <div className="flex h-40 items-center justify-center rounded-lg border-2 border-dashed text-muted-foreground">
              Sem dados de pessoal para {year === 0 ? "os senadores selecionados" : year}.
            </div>
          ) : (
            <figure>
              <div
                className="w-full"
                style={{ height: Math.max(220, dadosGrafico.length * 36 + 80) }}
                role="img"
                aria-label={`Gráfico de barras empilhadas por vínculo. ${resumoGrafico}.`}
              >
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={dadosGrafico} layout="vertical" margin={{ top: 8, right: 16, left: 8, bottom: 8 }}>
                    <CartesianGrid strokeDasharray="3 3" horizontal={false} stroke="var(--border)" />
                    <XAxis type="number" allowDecimals={false} tick={{ fill: "var(--muted-foreground)", fontSize: 12 }} />
                    <YAxis
                      dataKey="rotulo"
                      type="category"
                      width={190}
                      interval={0}
                      tick={{ fill: "var(--muted-foreground)", fontSize: 11 }}
                    />
                    <Tooltip
                      cursor={{ fill: "var(--muted)", opacity: 0.5 }}
                      content={({ active, payload, label }) => (
                        <ChartTooltipContent
                          active={active}
                          payload={payload}
                          label={label}
                          labelFormatter={(l, p) => (p[0]?.payload as { titulo?: string } | undefined)?.titulo ?? l}
                          valueFormatter={(v) => plural(Number(v) || 0)}
                        />
                      )}
                    />
                    <Legend
                      wrapperStyle={{ fontSize: 12 }}
                      formatter={(valor) => <span className="text-foreground">{valor}</span>}
                    />
                    {vinculos.map((v) => (
                      <Bar
                        key={v}
                        dataKey={v}
                        name={v}
                        stackId="pessoal"
                        fill={corVinculo(v)}
                        stroke="var(--card)"
                        strokeWidth={2}
                      />
                    ))}
                  </BarChart>
                </ResponsiveContainer>
              </div>
              <figcaption className="sr-only">Os mesmos números estão na tabela abaixo.</figcaption>
            </figure>
          )}

          <div className="overflow-x-auto">
            <Table>
              <TableCaption>Servidores por senador, local e vínculo (números agregados)</TableCaption>
              <TableHeader>
                <TableRow>
                  <TableHead scope="col">Senador</TableHead>
                  <TableHead scope="col" className="text-right">Ano</TableHead>
                  <TableHead scope="col" className="text-right">Gabinete</TableHead>
                  <TableHead scope="col" className="text-right">Escritórios</TableHead>
                  {vinculos.map((v) => (
                    <TableHead key={v} scope="col" className="text-right">
                      {v}
                    </TableHead>
                  ))}
                  <TableHead scope="col" className="text-right">Total</TableHead>
                  <TableHead scope="col">Fonte</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {linhas.map(({ senador, dados }, i) => (
                  <TableRow key={senador.id}>
                    <TableHead scope="row" className="font-medium text-foreground">
                      {senador.nome}
                    </TableHead>
                    {queries[i].isError || !dados || dados.total === 0 ? (
                      <TableCell colSpan={vinculos.length + 4} className="text-muted-foreground">
                        {queries[i].isError ? "Erro ao carregar" : "Sem dados na fonte oficial"}
                      </TableCell>
                    ) : (
                      <>
                        <TableCell className="text-right tabular-nums">{dados.ano}</TableCell>
                        <TableCell className="text-right tabular-nums">{totalLocal(dados, "GABINETE")}</TableCell>
                        <TableCell className="text-right tabular-nums">{totalLocal(dados, "ESCRITORIO")}</TableCell>
                        {vinculos.map((v) => (
                          <TableCell key={v} className="text-right tabular-nums">
                            {dados.por_vinculo.find((x) => x.vinculo === v)?.quantidade ?? 0}
                          </TableCell>
                        ))}
                        <TableCell className="text-right font-semibold tabular-nums">{dados.total}</TableCell>
                      </>
                    )}
                    <TableCell>
                      {dados?.fonte_url && (
                        <a
                          href={dados.fonte_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="inline-flex items-center gap-1 text-primary underline-offset-4 hover:underline"
                        >
                          Senado
                          <ExternalLink className="h-3.5 w-3.5" aria-hidden="true" />
                          <span className="sr-only">: detalhes de {senador.nome} (abre em nova aba)</span>
                        </a>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          <p className="text-sm text-muted-foreground">
            {NOTA_PRIVACIDADE} Fonte: API de Dados Abertos Administrativos do Senado Federal, atualizada
            semanalmente.
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
