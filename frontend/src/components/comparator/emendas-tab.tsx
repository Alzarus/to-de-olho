"use client";

import { useQueries } from "@tanstack/react-query";
import { getEmendas } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, Legend, CartesianGrid, Cell } from "recharts";
import type { SenatorBasicProfile } from "@/contexts/comparator-context";
import type { Emenda } from "@/types/api";
import { ChartTooltipContent } from "@/components/ui/chart-tooltip";
import { useIsMobile } from "@/hooks/use-mobile";
import {
  casaLocal,
  classificarTipoEmenda,
  corSerie,
  lerFiltrosEmendas,
  PARAMS_EMENDAS,
  ROTULO_TIPO_EMENDA,
  ufDaLocalidade,
  type MetricaEmenda,
  type TipoEmenda,
} from "@/lib/comparador-filtros";
import {
  BarraFiltros,
  eixoReais,
  FiltroSegmentado,
  FiltroSelect,
  formatarReais,
  MarcadorSerie,
  TabelaDados,
  truncar,
  useFiltrosUrl,
} from "@/components/comparator/filtros";

interface EmendasTabProps {
  senators: SenatorBasicProfile[];
  year: number;
}

const TODAS = "__todas__";
const ORDEM_TIPOS: TipoEmenda[] = ["especial", "finalidade", "bancada", "outros"];
const ROTULO_METRICA: Record<MetricaEmenda, string> = { pago: "Pago", empenhado: "Empenhado" };

type LinhaGrafico = Record<string, string | number>;
const chaveSerie = (id: number) => `s${id}`;

const valorDe = (e: Emenda, metrica: MetricaEmenda) => (metrica === "pago" ? e.valor_pago : e.valor_empenhado);

export function EmendasTab({ senators, year }: EmendasTabProps) {
  const isMobile = useIsMobile();
  const { params, definir } = useFiltrosUrl();
  const filtros = lerFiltrosEmendas(params);

  const queries = useQueries({
    queries: senators.map((s) => ({
      queryKey: ["senador-emendas", s.id, year],
      queryFn: () => getEmendas(s.id, year),
    })),
  });

  if (queries.some((q) => q.isLoading)) return <Skeleton className="h-[500px] w-full" />;

  const todas = senators.map((_, i) => queries[i].data?.emendas ?? []);

  // Opções dos filtros a partir dos dados (a base só tem emendas individuais hoje)
  const tiposPresentes = new Set(todas.flat().map((e) => classificarTipoEmenda(e.tipo)));
  const opcoesTipo = [
    { valor: "todos", rotulo: "Todos os tipos" },
    ...ORDEM_TIPOS.filter((t) => tiposPresentes.has(t) || t === filtros.tipo).map((t) => ({
      valor: t,
      rotulo: ROTULO_TIPO_EMENDA[t],
    })),
  ];

  const somaPorLocal = new Map<string, number>();
  const somaPorUf = new Map<string, number>();
  for (const e of todas.flat()) {
    const v = valorDe(e, filtros.metrica);
    const uf = ufDaLocalidade(e.localidade);
    if (uf) somaPorUf.set(uf, (somaPorUf.get(uf) ?? 0) + v);
    else somaPorLocal.set(e.localidade, (somaPorLocal.get(e.localidade) ?? 0) + v);
  }
  const ordenar = (m: Map<string, number>) => [...m.entries()].sort((a, b) => b[1] - a[1]).map(([k]) => k);
  const opcoesLocal = [
    { valor: TODAS, rotulo: "Todas as localidades" },
    ...ordenar(somaPorUf).map((uf) => ({ valor: `UF:${uf}`, rotulo: `${uf} (estado e municípios)` })),
    ...ordenar(somaPorLocal).map((l) => ({ valor: l, rotulo: l })),
  ];
  if (filtros.local && !opcoesLocal.some((o) => o.valor === filtros.local)) {
    opcoesLocal.push({ valor: filtros.local, rotulo: filtros.local.replace(/^UF:/, "") });
  }

  const filtradas = todas.map((lista) =>
    lista.filter(
      (e) =>
        (filtros.tipo === "todos" || classificarTipoEmenda(e.tipo) === filtros.tipo) && casaLocal(filtros.local, e.localidade),
    ),
  );

  const rotuloMetrica = ROTULO_METRICA[filtros.metrica];
  const yearLabel = year === 0 ? "Mandato completo" : year.toString();
  const recorte = [
    filtros.tipo !== "todos" && ROTULO_TIPO_EMENDA[filtros.tipo],
    filtros.local && `localidade ${filtros.local.replace(/^UF:/, "UF ")}`,
  ]
    .filter(Boolean)
    .join(" · ");

  // 1. Total por senador na métrica escolhida
  const porSenador = senators.map((s, i) => ({
    id: s.id,
    nome: s.nome,
    cor: corSerie(i),
    valor: filtradas[i].reduce((soma, e) => soma + valorDe(e, filtros.metrica), 0),
    quantidade: filtradas[i].length,
    pago: filtradas[i].reduce((soma, e) => soma + e.valor_pago, 0),
    empenhado: filtradas[i].reduce((soma, e) => soma + e.valor_empenhado, 0),
  }));

  // 2. Por tipo de emenda (categorias no eixo, senadores nas barras)
  const tiposNoGrafico = ORDEM_TIPOS.filter((t) => filtradas.some((l) => l.some((e) => classificarTipoEmenda(e.tipo) === t)));
  const porTipo: LinhaGrafico[] = tiposNoGrafico.map((t) => {
    const linha: LinhaGrafico = { tipo: ROTULO_TIPO_EMENDA[t] };
    senators.forEach((s, i) => {
      linha[chaveSerie(s.id)] = filtradas[i]
        .filter((e) => classificarTipoEmenda(e.tipo) === t)
        .reduce((soma, e) => soma + valorDe(e, filtros.metrica), 0);
    });
    return linha;
  });

  // 3. Principais localidades de cada senador
  const localidades = senators.map((s, i) => {
    const m = new Map<string, number>();
    for (const e of filtradas[i]) m.set(e.localidade, (m.get(e.localidade) ?? 0) + valorDe(e, filtros.metrica));
    return {
      id: s.id,
      nome: s.nome,
      cor: corSerie(i),
      top: [...m.entries()].sort((a, b) => b[1] - a[1]).slice(0, 5),
    };
  });

  const nomePorChave = new Map(senators.map((s) => [chaveSerie(s.id), s.nome]));

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <BarraFiltros rotulo="Filtros de emendas">
        <FiltroSelect
          rotulo="Tipo de emenda"
          valor={filtros.tipo}
          largura="w-[300px]"
          opcoes={opcoesTipo}
          aoMudar={(v) => definir({ [PARAMS_EMENDAS.tipo]: v === "todos" ? null : v })}
        />
        <FiltroSegmentado
          rotulo="Valor"
          valor={filtros.metrica}
          opcoes={[
            { valor: "pago", rotulo: "Pago" },
            { valor: "empenhado", rotulo: "Empenhado" },
          ]}
          aoMudar={(v) => definir({ [PARAMS_EMENDAS.metrica]: v === "pago" ? null : v })}
        />
        <FiltroSelect
          rotulo="UF / localidade"
          valor={filtros.local || TODAS}
          largura="w-[240px]"
          opcoes={opcoesLocal}
          aoMudar={(v) => definir({ [PARAMS_EMENDAS.local]: v === TODAS ? null : v })}
        />
      </BarraFiltros>

      <Card>
        <CardHeader>
          <CardTitle>
            Valor {rotuloMetrica.toLowerCase()} em emendas ({yearLabel})
          </CardTitle>
          <CardDescription>
            Soma das emendas de cada senador{recorte ? ` · ${recorte}` : ""}. Empenhado é o valor reservado no
            orçamento; pago é o que chegou a ser transferido.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div
            className="h-[320px]"
            role="figure"
            aria-label={`Gráfico de barras: valor ${rotuloMetrica.toLowerCase()} em emendas por senador. Os valores estão na tabela abaixo.`}
          >
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={porSenador} layout="vertical" margin={{ top: 8, right: 24, left: 8, bottom: 8 }}>
                <CartesianGrid strokeDasharray="3 3" horizontal={false} stroke="var(--border)" />
                <XAxis type="number" tickFormatter={eixoReais} tick={{ fill: "var(--muted-foreground)", fontSize: 11 }} />
                <YAxis
                  dataKey="nome"
                  type="category"
                  width={isMobile ? 100 : 160}
                  tick={{ fill: "var(--muted-foreground)", fontSize: isMobile ? 11 : 12 }}
                  tickFormatter={(v: string) => truncar(v, isMobile ? 14 : 22)}
                />
                <Tooltip
                  cursor={{ fill: "var(--muted)", opacity: 0.5 }}
                  content={({ active, payload, label }) => (
                    <ChartTooltipContent
                      active={active}
                      payload={payload}
                      label={label}
                      valueFormatter={formatarReais}
                      colorFormatter={(entry) => (entry.payload as { cor?: string } | undefined)?.cor}
                      extra={(p) => {
                        const linha = p[0]?.payload as (typeof porSenador)[number] | undefined;
                        return linha ? `${linha.quantidade} ${linha.quantidade === 1 ? "emenda" : "emendas"}` : null;
                      }}
                    />
                  )}
                />
                <Bar dataKey="valor" name={rotuloMetrica} radius={[0, 4, 4, 0]} barSize={18}>
                  {porSenador.map((s) => (
                    <Cell key={s.id} fill={s.cor} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </div>
          <TabelaDados
            titulo={`Emendas por senador, ${yearLabel}`}
            linhas={porSenador}
            chave={(l) => String(l.id)}
            colunas={[
              { cabecalho: "Senador", valor: (l) => l.nome },
              { cabecalho: "Emendas", numerica: true, valor: (l) => l.quantidade.toLocaleString("pt-BR") },
              { cabecalho: "Empenhado", numerica: true, valor: (l) => formatarReais(l.empenhado) },
              { cabecalho: "Pago", numerica: true, valor: (l) => formatarReais(l.pago) },
            ]}
          />
        </CardContent>
      </Card>

      {porTipo.length > 1 && (
        <Card>
          <CardHeader>
            <CardTitle>Por tipo de emenda ({yearLabel})</CardTitle>
            <CardDescription>
              Transferência especial (“emenda PIX”) vai direto ao caixa do ente, sem projeto definido; a de
              finalidade definida é vinculada a um programa. Valor {rotuloMetrica.toLowerCase()}.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div
              style={{ height: Math.max(220, porTipo.length * (senators.length * 18 + 32) + 90) }}
              role="figure"
              aria-label="Gráfico de barras agrupadas: emendas por tipo e senador. Os valores estão na tabela abaixo."
            >
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={porTipo} layout="vertical" margin={{ top: 8, right: 24, left: 8, bottom: 8 }} barGap={2}>
                  <CartesianGrid strokeDasharray="3 3" horizontal={false} stroke="var(--border)" />
                  <XAxis type="number" tickFormatter={eixoReais} tick={{ fill: "var(--muted-foreground)", fontSize: 11 }} />
                  <YAxis
                    dataKey="tipo"
                    type="category"
                    width={isMobile ? 110 : 220}
                    tick={{ fill: "var(--muted-foreground)", fontSize: 11 }}
                    tickFormatter={(v: string) => truncar(v, isMobile ? 16 : 34)}
                    interval={0}
                  />
                  <Tooltip
                    cursor={{ fill: "var(--muted)", opacity: 0.5 }}
                    content={({ active, payload, label }) => (
                      <ChartTooltipContent
                        active={active}
                        payload={payload}
                        label={label}
                        valueFormatter={formatarReais}
                        nameFormatter={(n, entry) => nomePorChave.get(String(entry.dataKey)) ?? String(n)}
                      />
                    )}
                  />
                  <Legend
                    wrapperStyle={{ paddingTop: 12, fontSize: isMobile ? 11 : 12 }}
                    formatter={(valor: string) => <span className="text-foreground">{truncar(valor, 32)}</span>}
                  />
                  {senators.map((s, i) => (
                    <Bar key={s.id} dataKey={chaveSerie(s.id)} name={s.nome} fill={corSerie(i)} radius={[0, 4, 4, 0]} maxBarSize={18} />
                  ))}
                </BarChart>
              </ResponsiveContainer>
            </div>
            <TabelaDados
              titulo={`Emendas por tipo, ${yearLabel}`}
              linhas={porTipo}
              chave={(l) => String(l.tipo)}
              colunas={[
                { cabecalho: "Tipo", valor: (l) => String(l.tipo) },
                ...senators.map((s) => ({
                  cabecalho: s.nome,
                  numerica: true,
                  valor: (l: LinhaGrafico) => formatarReais(l[chaveSerie(s.id)]),
                })),
              ]}
            />
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Principais destinos ({yearLabel})</CardTitle>
          <CardDescription>
            As 5 localidades com maior valor {rotuloMetrica.toLowerCase()} de cada senador
            {recorte ? ` · ${recorte}` : ""}. “MÚLTIPLO” reúne emendas para vários municípios.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {localidades.map((s) => (
              <section key={s.id} aria-labelledby={`destinos-${s.id}`}>
                <h3 id={`destinos-${s.id}`} className="mb-2 flex items-center gap-2 text-sm font-semibold">
                  <MarcadorSerie cor={s.cor} />
                  <span className="truncate">{s.nome}</span>
                </h3>
                {s.top.length === 0 ? (
                  <p className="text-sm text-muted-foreground">Sem emendas no recorte.</p>
                ) : (
                  <ol className="space-y-1.5 text-sm">
                    {s.top.map(([local, valor]) => (
                      <li key={local} className="flex justify-between gap-3">
                        <span className="min-w-0 truncate" title={local}>
                          {local}
                        </span>
                        <span className="shrink-0 tabular-nums font-medium">{formatarReais(valor)}</span>
                      </li>
                    ))}
                  </ol>
                )}
              </section>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
