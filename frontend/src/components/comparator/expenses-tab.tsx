"use client";

import { keepPreviousData, useQueries } from "@tanstack/react-query";
import { getSenadorScore, getDespesasAgregado, getDespesasMensal } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";
import { AlertCircle } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  LineChart,
  Line,
  ReferenceLine,
  Cell,
} from "recharts";
import { ChartTooltipContent } from "@/components/ui/chart-tooltip";
import { useIsMobile } from "@/hooks/use-mobile";
import type { SenatorBasicProfile } from "@/contexts/comparator-context";
import {
  agruparSerie,
  anoInteiro,
  categoriasPorVolume,
  corSerie,
  lerFiltrosDespesas,
  nomeMes,
  PARAMS_DESPESAS,
  percentual,
  tetoMensal,
  type PontoSerie,
} from "@/lib/comparador-filtros";
import {
  BarraFiltros,
  eixoReais,
  FiltroMulti,
  FiltroSegmentado,
  FiltroSelect,
  formatarPct,
  formatarReais,
  TabelaDados,
  truncar,
  useFiltrosUrl,
  type ColunaTabela,
} from "@/components/comparator/filtros";

interface ExpensesTabProps {
  senators: SenatorBasicProfile[];
  year: number;
}

const CATEGORIAS_PADRAO = 5;
const COR_TETO = "var(--serie-ref)";

const OPCOES_MES = Array.from({ length: 12 }, (_, i) => ({
  valor: String(i + 1),
  rotulo: nomeMes(i + 1).replace(/^./, (c) => c.toUpperCase()),
}));

// Uma linha por senador, com as colunas do gráfico de uso da cota
interface LinhaCota {
  id: number;
  nome: string;
  cor: string;
  gasto: number;
  teto: number;
  pct: number | null;
}

type LinhaGrafico = Record<string, string | number | null>;

const chaveSerie = (id: number) => `s${id}`;

export function ExpensesTab({ senators, year }: ExpensesTabProps) {
  const isMobile = useIsMobile();
  const { params, definir } = useFiltrosUrl();
  const filtros = lerFiltrosDespesas(params);
  const apiYear = year === 0 ? undefined : year;
  // O recorte de meses só vale com um ano escolhido
  const recorte = year > 0 && !anoInteiro(filtros);
  const meses = { de: filtros.mesDe, ate: filtros.mesAte };
  const emPct = filtros.unidade === "teto";

  const scoreQueries = useQueries({
    queries: senators.map((s) => ({
      queryKey: ["senador-score", s.id, apiYear],
      queryFn: () => getSenadorScore(s.id, apiYear),
    })),
  });

  const agregadoQueries = useQueries({
    queries: senators.map((s) => ({
      queryKey: recorte
        ? ["senador-despesas-agregado", s.id, apiYear, meses.de, meses.ate]
        : ["senador-despesas-agregado", s.id, apiYear],
      queryFn: () => getDespesasAgregado(s.id, apiYear, recorte ? meses : undefined),
      placeholderData: keepPreviousData,
    })),
  });

  const mensalQueries = useQueries({
    queries: senators.map((s) => ({
      queryKey: ["senador-despesas-mensal", s.id, apiYear],
      queryFn: () => getDespesasMensal(s.id, apiYear),
    })),
  });

  // Categorias disponíveis no recorte e as selecionadas (padrão: as 5 maiores)
  const categoriasDisponiveis = categoriasPorVolume(agregadoQueries.map((q) => q.data?.por_tipo));
  const categorias = filtros.categorias ?? categoriasDisponiveis.slice(0, CATEGORIAS_PADRAO);
  const evolucaoPorCategoria = filtros.evolucao === "categorias" && categorias.length > 0;

  const mensalCategoriaQueries = useQueries({
    queries: senators.map((s) => ({
      queryKey: ["senador-despesas-mensal", s.id, apiYear, "tipos", ...categorias],
      queryFn: () => getDespesasMensal(s.id, apiYear, categorias),
      enabled: evolucaoPorCategoria,
      placeholderData: keepPreviousData,
    })),
  });

  const carregando =
    scoreQueries.some((q) => q.isLoading) ||
    agregadoQueries.some((q) => q.isLoading) ||
    mensalQueries.some((q) => q.isLoading);

  if (carregando) {
    return <Skeleton className="h-[600px] w-full rounded-lg" />;
  }

  if (scoreQueries.every((q) => q.isError) && agregadoQueries.every((q) => q.isError)) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertTitle>Erro</AlertTitle>
        <AlertDescription>Não foi possível carregar os dados de despesas.</AlertDescription>
      </Alert>
    );
  }

  const periodo =
    year === 0
      ? "Mandato completo"
      : recorte
        ? `${nomeMes(filtros.mesDe)} a ${nomeMes(filtros.mesAte)} de ${year}`
        : String(year);
  const mesesNoRecorte = filtros.mesAte - filtros.mesDe + 1;

  // --- 1. Uso da cota ---
  // Ano inteiro ou mandato: gasto e teto do score (teto proporcional aos meses
  // em exercício). Com recorte: soma dos meses e teto mensal da UF x meses.
  const cota: LinhaCota[] = senators.map((s, i) => {
    const detalhes = scoreQueries[i].data?.detalhes;
    let gasto = detalhes?.gasto_ceaps ?? 0;
    let teto = detalhes?.teto_ceaps ?? 0;
    if (recorte) {
      gasto = (mensalQueries[i].data?.meses ?? [])
        .filter((m) => m.ano === year && m.mes >= filtros.mesDe && m.mes <= filtros.mesAte)
        .reduce((soma, m) => soma + m.total, 0);
      teto = tetoMensal(s.uf) * mesesNoRecorte;
    } else if (!(teto > 0)) {
      teto = tetoMensal(s.uf) * (year === 0 ? 0 : 12);
    }
    return { id: s.id, nome: s.nome, cor: corSerie(i), gasto, teto, pct: percentual(gasto, teto) };
  });
  const tetoPorSenador = new Map(cota.map((c) => [c.id, c.teto]));

  // --- 2. Categorias ---
  const dadosCategorias: LinhaGrafico[] = categorias.map((cat) => {
    const linha: LinhaGrafico = { categoria: cat };
    senators.forEach((s, i) => {
      const total = agregadoQueries[i].data?.por_tipo?.find((t) => t.tipo_despesa === cat)?.total ?? 0;
      linha[chaveSerie(s.id)] = emPct ? percentual(total, tetoPorSenador.get(s.id) ?? 0) : total;
    });
    return linha;
  });

  // --- 3. Evolução ---
  const fonteEvolucao = evolucaoPorCategoria ? mensalCategoriaQueries : mensalQueries;
  const seriesEvolucao = senators.map((s, i) =>
    agruparSerie(fonteEvolucao[i].data?.meses ?? [], filtros.granularidade, year, filtros.mesDe, filtros.mesAte),
  );
  const pontos = new Map<string, LinhaGrafico>();
  const rotulos = new Map<string, PontoSerie>();
  seriesEvolucao.forEach((serie) => serie.forEach((p) => rotulos.set(p.chave, p)));
  const chavesOrdenadas = [...rotulos.keys()].sort();
  for (const chave of chavesOrdenadas) {
    const p = rotulos.get(chave)!;
    pontos.set(chave, { chave, rotulo: p.rotulo, rotuloCurto: p.rotuloCurto });
  }
  senators.forEach((s, i) => {
    const teto = tetoMensal(s.uf);
    for (const p of seriesEvolucao[i]) {
      const linha = pontos.get(p.chave)!;
      linha[chaveSerie(s.id)] = emPct ? percentual(p.total, teto * p.meses) : p.total;
    }
  });
  const dadosEvolucao = [...pontos.values()];

  const formatarValor = (v: unknown) => (emPct ? formatarPct(v) : formatarReais(v));
  const eixoValor = (v: number) => (emPct ? `${v.toLocaleString("pt-BR")}%` : eixoReais(v));
  const nomePorChave = new Map(senators.map((s) => [chaveSerie(s.id), s.nome]));

  const legendaSenadores = (
    <Legend
      wrapperStyle={{ paddingTop: 12, fontSize: isMobile ? 11 : 12 }}
      formatter={(valor: string) => <span className="text-foreground">{truncar(valor, isMobile ? 18 : 32)}</span>}
    />
  );

  const colunasPorSenador = (rotuloPrimeira: string, primeira: (l: LinhaGrafico) => string): ColunaTabela<LinhaGrafico>[] => [
    { cabecalho: rotuloPrimeira, valor: primeira },
    ...senators.map((s) => ({
      cabecalho: s.nome,
      numerica: true,
      valor: (l: LinhaGrafico) => formatarValor(l[chaveSerie(s.id)]),
    })),
  ];

  const textoTeto = recorte
    ? `Teto = cota mensal da UF × ${mesesNoRecorte} ${mesesNoRecorte === 1 ? "mês" : "meses"} do recorte.`
    : "Teto = cota mensal da UF × meses em exercício no período (o mesmo do ranking).";

  return (
    <div className="space-y-6">
      <BarraFiltros rotulo="Filtros de despesas">
        <FiltroMulti
          rotulo="Categorias"
          opcoes={categoriasDisponiveis.map((c) => ({ valor: c, rotulo: c }))}
          selecionados={categorias}
          aoMudar={(valores) => definir({ [PARAMS_DESPESAS.categorias]: valores.length > 0 ? valores : null })}
          aoRestaurar={() => definir({ [PARAMS_DESPESAS.categorias]: null })}
          rotuloRestaurar={`Padrão: as ${CATEGORIAS_PADRAO} maiores`}
          resumo={
            filtros.categorias === null
              ? `As ${Math.min(CATEGORIAS_PADRAO, categoriasDisponiveis.length)} maiores`
              : undefined
          }
        />
        <FiltroSelect
          rotulo="De"
          valor={String(filtros.mesDe)}
          opcoes={OPCOES_MES}
          desabilitado={year === 0}
          largura="w-[140px]"
          aoMudar={(v) => {
            const de = Number(v);
            definir({
              [PARAMS_DESPESAS.mesDe]: de === 1 ? null : de,
              [PARAMS_DESPESAS.mesAte]: de > filtros.mesAte ? de : filtros.mesAte === 12 ? null : filtros.mesAte,
            });
          }}
        />
        <FiltroSelect
          rotulo="Até"
          valor={String(filtros.mesAte)}
          opcoes={OPCOES_MES}
          desabilitado={year === 0}
          largura="w-[140px]"
          aoMudar={(v) => {
            const ate = Number(v);
            definir({
              [PARAMS_DESPESAS.mesAte]: ate === 12 ? null : ate,
              [PARAMS_DESPESAS.mesDe]: ate < filtros.mesDe ? ate : filtros.mesDe === 1 ? null : filtros.mesDe,
            });
          }}
        />
        <FiltroSegmentado
          rotulo="Unidade"
          valor={filtros.unidade}
          opcoes={[
            { valor: "valor", rotulo: "R$" },
            { valor: "teto", rotulo: "% do teto" },
          ]}
          aoMudar={(v) => definir({ [PARAMS_DESPESAS.unidade]: v === "valor" ? null : v })}
        />
        {year === 0 && (
          <p className="basis-full text-xs text-muted-foreground">
            Escolha um ano no seletor de período para recortar meses.
          </p>
        )}
      </BarraFiltros>

      {/* 1. Uso da cota */}
      <Card>
        <CardHeader>
          <CardTitle>Uso da cota ({periodo})</CardTitle>
          <CardDescription>
            {emPct
              ? "Gasto como percentual do teto de cada senador: compara estados com cotas diferentes."
              : "Gasto com a CEAPS e teto disponível no período."}{" "}
            {textoTeto}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="h-[320px]" role="figure" aria-label={`Gráfico de barras: uso da cota por senador, ${periodo}. Os valores estão na tabela abaixo.`}>
            <ResponsiveContainer width="100%" height="100%">
              <BarChart
                data={cota}
                layout="vertical"
                margin={{ top: emPct ? 24 : 8, right: emPct ? 48 : 24, left: 8, bottom: 8 }}
                barGap={2}
              >
                <CartesianGrid strokeDasharray="3 3" horizontal={false} stroke="var(--border)" />
                <XAxis
                  type="number"
                  tickFormatter={eixoValor}
                  tick={{ fill: "var(--muted-foreground)", fontSize: 12 }}
                  domain={emPct ? [0, (max: number) => Math.max(100, Math.ceil(max / 10) * 10)] : [0, "auto"]}
                />
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
                      valueFormatter={(v) => formatarValor(v)}
                      colorFormatter={(entry) =>
                        entry.dataKey === "teto" ? COR_TETO : (entry.payload as LinhaCota | undefined)?.cor
                      }
                      extra={
                        emPct
                          ? (p) => {
                              const linha = p[0]?.payload as LinhaCota | undefined;
                              return linha ? `${formatarReais(linha.gasto)} de ${formatarReais(linha.teto)}` : null;
                            }
                          : undefined
                      }
                    />
                  )}
                />
                {emPct ? (
                  <>
                    <ReferenceLine
                      x={100}
                      stroke="var(--muted-foreground)"
                      strokeDasharray="4 4"
                      label={{ value: "Teto (100%)", position: "top", fill: "var(--muted-foreground)", fontSize: 11 }}
                    />
                    <Bar dataKey="pct" name="% do teto usado" radius={[0, 4, 4, 0]} barSize={18}>
                      {cota.map((c) => (
                        <Cell key={c.id} fill={c.cor} />
                      ))}
                    </Bar>
                  </>
                ) : (
                  <>
                    <Bar dataKey="gasto" name="Gasto" radius={[0, 4, 4, 0]} barSize={14}>
                      {cota.map((c) => (
                        <Cell key={c.id} fill={c.cor} />
                      ))}
                    </Bar>
                    <Bar dataKey="teto" name="Teto" fill={COR_TETO} radius={[0, 4, 4, 0]} barSize={14} />
                  </>
                )}
              </BarChart>
            </ResponsiveContainer>
          </div>
          <p className="mt-2 text-xs text-muted-foreground">
            {emPct
              ? "Cada barra usa a cor do senador; a linha tracejada marca 100% do teto."
              : "Barra colorida: gasto (cor do senador). Barra cinza: teto."}
          </p>
          <TabelaDados
            titulo={`Uso da cota por senador, ${periodo}`}
            linhas={cota}
            chave={(l) => String(l.id)}
            colunas={[
              { cabecalho: "Senador", valor: (l) => l.nome },
              { cabecalho: "Gasto", numerica: true, valor: (l) => formatarReais(l.gasto) },
              { cabecalho: "Teto", numerica: true, valor: (l) => formatarReais(l.teto) },
              { cabecalho: "% do teto", numerica: true, valor: (l) => formatarPct(l.pct) },
            ]}
          />
        </CardContent>
      </Card>

      {/* 2. Categorias */}
      <Card>
        <CardHeader>
          <CardTitle>Gasto por categoria ({periodo})</CardTitle>
          <CardDescription>
            {filtros.categorias === null
              ? `As ${Math.min(CATEGORIAS_PADRAO, categoriasDisponiveis.length)} categorias com maior gasto somado entre os senadores. Use o filtro “Categorias” para escolher outras.`
              : `${categorias.length} ${categorias.length === 1 ? "categoria escolhida" : "categorias escolhidas"}.`}
            {emPct && " Valores em % do teto de cada senador no período."}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {categorias.length === 0 ? (
            <p className="py-12 text-center text-sm text-muted-foreground">Nenhuma categoria selecionada.</p>
          ) : (
            <>
              <div
                style={{ height: Math.max(260, categorias.length * (senators.length * 16 + 28) + 90) }}
                role="figure"
                aria-label={`Gráfico de barras agrupadas: gasto por categoria e senador, ${periodo}. Os valores estão na tabela abaixo.`}
              >
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={dadosCategorias} layout="vertical" margin={{ top: 8, right: 24, left: 8, bottom: 8 }} barGap={2}>
                    <CartesianGrid strokeDasharray="3 3" horizontal={false} stroke="var(--border)" />
                    <XAxis type="number" tickFormatter={eixoValor} tick={{ fill: "var(--muted-foreground)", fontSize: 11 }} />
                    <YAxis
                      dataKey="categoria"
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
                          valueFormatter={(v) => formatarValor(v)}
                        />
                      )}
                    />
                    {legendaSenadores}
                    {senators.map((s, i) => (
                      <Bar
                        key={s.id}
                        dataKey={chaveSerie(s.id)}
                        name={s.nome}
                        fill={corSerie(i)}
                        radius={[0, 4, 4, 0]}
                        maxBarSize={16}
                      />
                    ))}
                  </BarChart>
                </ResponsiveContainer>
              </div>
              <TabelaDados
                titulo={`Gasto por categoria, ${periodo}`}
                linhas={dadosCategorias}
                chave={(l) => String(l.categoria)}
                colunas={colunasPorSenador("Categoria", (l) => String(l.categoria))}
              />
            </>
          )}
        </CardContent>
      </Card>

      {/* 3. Evolução */}
      <Card>
        <CardHeader className="gap-4">
          <div>
            <CardTitle>Evolução dos gastos ({periodo})</CardTitle>
            <CardDescription>
              {evolucaoPorCategoria
                ? `Soma das ${categorias.length} categorias selecionadas.`
                : "Total da cota (todas as categorias)."}{" "}
              {filtros.granularidade === "acumulada" && "Valores acumulados desde o início do período. "}
              {emPct && "Em % do teto mensal da UF de cada senador multiplicado pelos meses de cada ponto."}
            </CardDescription>
          </div>
          <div role="group" aria-label="Opções da evolução" className="no-print flex flex-wrap items-end gap-4">
            <FiltroSegmentado
              rotulo="Granularidade"
              valor={filtros.granularidade}
              opcoes={[
                { valor: "mensal", rotulo: "Mensal" },
                { valor: "trimestral", rotulo: "Trimestral" },
                { valor: "acumulada", rotulo: "Acumulada" },
              ]}
              aoMudar={(v) => definir({ [PARAMS_DESPESAS.granularidade]: v === "mensal" ? null : v })}
            />
            <FiltroSelect
              rotulo="Evolução de"
              valor={filtros.evolucao}
              largura="w-[220px]"
              opcoes={[
                { valor: "total", rotulo: "Total da cota" },
                { valor: "categorias", rotulo: "Categorias selecionadas" },
              ]}
              aoMudar={(v) => definir({ [PARAMS_DESPESAS.evolucao]: v === "total" ? null : v })}
            />
          </div>
        </CardHeader>
        <CardContent>
          {dadosEvolucao.length === 0 ? (
            <p className="py-12 text-center text-sm text-muted-foreground">Sem despesas no período.</p>
          ) : (
            <>
              <div
                className="h-[380px]"
                role="figure"
                aria-label={`Gráfico de linhas: evolução dos gastos por senador, ${periodo}. Os valores estão na tabela abaixo.`}
              >
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={dadosEvolucao} margin={{ top: 16, right: 16, left: 8, bottom: 8 }}>
                    <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="var(--border)" />
                    <XAxis
                      dataKey={isMobile ? "rotuloCurto" : "rotulo"}
                      tick={{ fill: "var(--muted-foreground)", fontSize: 11 }}
                      minTickGap={12}
                    />
                    <YAxis tickFormatter={eixoValor} width={isMobile ? 56 : 76} tick={{ fill: "var(--muted-foreground)", fontSize: 11 }} />
                    <Tooltip
                      cursor={{ stroke: "var(--muted-foreground)", strokeWidth: 1 }}
                      content={({ active, payload, label }) => (
                        <ChartTooltipContent
                          active={active}
                          payload={payload}
                          label={label}
                          valueFormatter={(v) => formatarValor(v)}
                          nameFormatter={(n, entry) => nomePorChave.get(String(entry.dataKey)) ?? String(n)}
                        />
                      )}
                    />
                    {legendaSenadores}
                    {emPct && <ReferenceLine y={100} stroke="var(--muted-foreground)" strokeDasharray="4 4" />}
                    {senators.map((s, i) => (
                      <Line
                        key={s.id}
                        type="monotone"
                        dataKey={chaveSerie(s.id)}
                        name={s.nome}
                        stroke={corSerie(i)}
                        strokeWidth={2}
                        dot={{ r: 4, strokeWidth: 2, fill: "var(--card)" }}
                        activeDot={{ r: 6, stroke: "var(--card)", strokeWidth: 2 }}
                        connectNulls
                      />
                    ))}
                  </LineChart>
                </ResponsiveContainer>
              </div>
              {emPct && (
                <p className="mt-2 text-xs text-muted-foreground">A linha tracejada marca 100% do teto.</p>
              )}
              <TabelaDados
                titulo={`Evolução dos gastos, ${periodo}`}
                linhas={dadosEvolucao}
                chave={(l) => String(l.chave)}
                colunas={colunasPorSenador(filtros.granularidade === "trimestral" ? "Trimestre" : "Mês", (l) => String(l.rotulo))}
              />
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
