"use client";

import { useQuery } from "@tanstack/react-query";
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
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
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

function formatarData(iso: string | null): string {
  if (!iso) return "—";
  return new Date(iso).toLocaleDateString("pt-BR", {
    day: "2-digit",
    month: "long",
    year: "numeric",
  });
}

const plural = (n: number) => `${n.toLocaleString("pt-BR")} ${n === 1 ? "servidor" : "servidores"}`;

export function GabineteTab({ id, ano }: { id: number; ano: number }) {
  const { data, isLoading, isError } = useQuery({
    queryKey: ["senador-gabinete", id, ano],
    queryFn: () => getGabinete(id, ano || undefined),
    enabled: id > 0,
  });

  if (isLoading) {
    return <Skeleton className="h-[420px] w-full rounded-lg" />;
  }

  if (isError || !data) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" aria-hidden="true" />
        <AlertTitle>Erro</AlertTitle>
        <AlertDescription>Não foi possível carregar a estrutura de gabinete.</AlertDescription>
      </Alert>
    );
  }

  const semDados = data.total === 0;
  const vinculos = ordenarVinculos(data.locais.flatMap((l) => l.vinculos.map((v) => v.vinculo)));
  const dadosGrafico = data.locais.map((l) => {
    const linha: Record<string, string | number> = { local: l.rotulo };
    for (const v of l.vinculos) linha[v.vinculo] = v.quantidade;
    return linha;
  });
  const resumoGrafico = data.locais
    .map((l) => `${l.rotulo}: ${plural(l.total)} (${l.vinculos.map((v) => `${v.quantidade} ${v.vinculo.toLowerCase()}`).join(", ")})`)
    .join("; ");

  return (
    <div className="space-y-6">
      {ano === 0 && data.ano > 0 && (
        <p className="text-sm text-muted-foreground">
          A estrutura de gabinete é informada por ano; mostrando {data.ano}, o ano mais recente disponível.
        </p>
      )}

      {data.cargo_mesa && (
        <Alert>
          <Info className="h-4 w-4" aria-hidden="true" />
          <AlertTitle>Integrante da Mesa Diretora</AlertTitle>
          <AlertDescription>{notaMesa(data.cargo_mesa)}</AlertDescription>
        </Alert>
      )}

      {semDados ? (
        <Card>
          <CardContent className="p-6 text-center text-muted-foreground">
            Sem dados de pessoal para {ano > 0 ? ano : "este senador"} na fonte oficial.
          </CardContent>
        </Card>
      ) : (
        <>
          <div className="grid gap-4 sm:grid-cols-3">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  Total ({data.ano})
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-3xl font-bold">{data.total.toLocaleString("pt-BR")}</p>
                <p className="mt-1 text-sm text-muted-foreground">servidores lotados</p>
              </CardContent>
            </Card>
            {data.locais.map((l) => (
              <Card key={l.local}>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm font-medium text-muted-foreground">{l.rotulo}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-3xl font-bold">{l.total.toLocaleString("pt-BR")}</p>
                  <ul className="mt-2 space-y-1 text-sm">
                    {l.vinculos.map((v) => (
                      <li key={v.vinculo} className="flex items-center justify-between gap-2">
                        <span className="flex items-center gap-1.5">
                          <span
                            aria-hidden="true"
                            className="h-2.5 w-2.5 rounded-[2px]"
                            style={{ backgroundColor: corVinculo(v.vinculo) }}
                          />
                          {v.vinculo}
                        </span>
                        <span className="font-semibold tabular-nums">{v.quantidade}</span>
                      </li>
                    ))}
                  </ul>
                </CardContent>
              </Card>
            ))}
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Servidores por local e vínculo ({data.ano})</CardTitle>
            </CardHeader>
            <CardContent>
              <figure>
                <div className="h-[200px] w-full" role="img" aria-label={`Gráfico de barras empilhadas. ${resumoGrafico}.`}>
                  <ResponsiveContainer width="100%" height="100%">
                    <BarChart data={dadosGrafico} layout="vertical" margin={{ top: 8, right: 16, left: 8, bottom: 8 }}>
                      <CartesianGrid strokeDasharray="3 3" horizontal={false} stroke="var(--border)" />
                      <XAxis type="number" allowDecimals={false} tick={{ fill: "var(--muted-foreground)", fontSize: 12 }} />
                      <YAxis
                        dataKey="local"
                        type="category"
                        width={140}
                        tick={{ fill: "var(--muted-foreground)", fontSize: 12 }}
                      />
                      <Tooltip
                        cursor={{ fill: "var(--muted)", opacity: 0.5 }}
                        content={({ active, payload, label }) => (
                          <ChartTooltipContent
                            active={active}
                            payload={payload}
                            label={label}
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

              <Table className="mt-4">
                <TableCaption>Servidores por local e vínculo em {data.ano} (números agregados)</TableCaption>
                <TableHeader>
                  <TableRow>
                    <TableHead scope="col">Local</TableHead>
                    {vinculos.map((v) => (
                      <TableHead key={v} scope="col" className="text-right">
                        {v}
                      </TableHead>
                    ))}
                    <TableHead scope="col" className="text-right">
                      Total
                    </TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data.locais.map((l) => (
                    <TableRow key={l.local}>
                      <TableHead scope="row" className="font-medium text-foreground">
                        {l.rotulo}
                      </TableHead>
                      {vinculos.map((v) => (
                        <TableCell key={v} className="text-right tabular-nums">
                          {l.vinculos.find((x) => x.vinculo === v)?.quantidade ?? 0}
                        </TableCell>
                      ))}
                      <TableCell className="text-right font-semibold tabular-nums">{l.total}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </>
      )}

      {data.beneficios.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Benefícios ({data.ano})</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="grid gap-3 sm:grid-cols-2">
              {data.beneficios.map((b) => {
                const usou = b.utilizacao.trim().toLowerCase() === "utilizou";
                return (
                  <li key={b.tipo} className="flex items-center justify-between gap-2 rounded-lg border p-3">
                    <span className="font-medium">{b.tipo}</span>
                    <Badge variant={usou ? "default" : "secondary"}>{b.utilizacao || "Não informado"}</Badge>
                  </li>
                );
              })}
            </ul>
          </CardContent>
        </Card>
      )}

      <div className="space-y-2 text-sm text-muted-foreground">
        <p>{NOTA_PRIVACIDADE}</p>
        <p>
          Fonte: API de Dados Abertos Administrativos do Senado Federal (recursos utilizados pelo senador).
          Atualizado em {formatarData(data.atualizado_em)}.{" "}
          <a
            href={data.fonte_url}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-1 font-medium text-primary underline-offset-4 hover:underline"
          >
            Ver detalhes na página oficial
            <ExternalLink className="h-3.5 w-3.5" aria-hidden="true" />
            <span className="sr-only">(abre em nova aba)</span>
          </a>
        </p>
      </div>
    </div>
  );
}
