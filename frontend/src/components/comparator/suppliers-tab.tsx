"use client";

import { useEffect, useId, useState } from "react";
import { keepPreviousData, useQueries } from "@tanstack/react-query";
import { getSenadorScore, getDespesasAgregado, getDespesasFornecedores } from "@/lib/api";
import { Skeleton } from "@/components/ui/skeleton";
import { AlertCircle, Building2, Search } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import Link from "next/link";
import {
  casaBuscaFornecedor,
  categoriasPorVolume,
  corSerie,
  formatarDocumento,
  lerFiltrosFornecedores,
  OPCOES_TOP,
  PARAMS_FORNECEDORES,
} from "@/lib/comparador-filtros";
import {
  BarraFiltros,
  FiltroSelect,
  formatarReais,
  MarcadorSerie,
  useFiltrosUrl,
} from "@/components/comparator/filtros";

interface SuppliersTabProps {
  selectedIds: number[];
  year: number;
}

const TODAS = "__todas__";

interface Fornecedor {
  chave: string;
  nome: string;
  doc: string;
  total: number;
  quantidade: number;
}

export function SuppliersTab({ selectedIds, year }: SuppliersTabProps) {
  const apiYear = year === 0 ? undefined : year;
  const { params, definir } = useFiltrosUrl();
  const filtros = lerFiltrosFornecedores(params);
  const tipos = filtros.categoria ? [filtros.categoria] : undefined;

  // Busca: digita livre e grava na URL depois de uma pausa
  const idBusca = useId();
  const [busca, setBusca] = useState(filtros.busca);
  useEffect(() => {
    if (busca.trim() === filtros.busca) return;
    const t = setTimeout(() => definir({ [PARAMS_FORNECEDORES.busca]: busca.trim() || null }), 350);
    return () => clearTimeout(t);
  }, [busca, filtros.busca, definir]);

  const scoreQueries = useQueries({
    queries: selectedIds.map((id) => ({
      queryKey: ["senador-score", id, apiYear],
      queryFn: () => getSenadorScore(id, apiYear),
    })),
  });

  // Categorias do filtro: as mesmas da aba Despesas (cache compartilhado)
  const agregadoQueries = useQueries({
    queries: selectedIds.map((id) => ({
      queryKey: ["senador-despesas-agregado", id, apiYear],
      queryFn: () => getDespesasAgregado(id, apiYear),
    })),
  });

  // Total por fornecedor, somado no backend com todos os lançamentos
  const fornecedorQueries = useQueries({
    queries: selectedIds.map((id) => ({
      queryKey: tipos
        ? ["senador-despesas-fornecedores", id, apiYear, "tipos", ...tipos]
        : ["senador-despesas-fornecedores", id, apiYear],
      queryFn: () => getDespesasFornecedores(id, apiYear, tipos),
      placeholderData: keepPreviousData,
    })),
  });

  if (scoreQueries.some((q) => q.isLoading) || fornecedorQueries.some((q) => q.isLoading)) {
    return <Skeleton className="h-[600px] w-full rounded-lg" />;
  }

  if (fornecedorQueries.every((q) => q.isError)) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertTitle>Erro</AlertTitle>
        <AlertDescription>Não foi possível carregar os dados de fornecedores.</AlertDescription>
      </Alert>
    );
  }

  const categorias = categoriasPorVolume(agregadoQueries.map((q) => q.data?.por_tipo));
  if (filtros.categoria && !categorias.includes(filtros.categoria)) categorias.unshift(filtros.categoria);

  const porSenador = selectedIds.map((id, index) => {
    const nome = scoreQueries[index].data?.nome || `Senador ${id}`;
    // Chave: CNPJ/CPF, que casa o mesmo fornecedor escrito de formas diferentes
    const fornecedores = new Map<string, Fornecedor>();
    fornecedorQueries[index].data?.fornecedores.forEach((f) => {
      const chave = f.cnpj_cpf || f.fornecedor || "NÃO INFORMADO";
      fornecedores.set(chave, {
        chave,
        nome: f.fornecedor || "NÃO INFORMADO",
        doc: f.cnpj_cpf ?? "",
        total: f.total,
        quantidade: f.quantidade,
      });
    });
    return { id, nome, cor: corSerie(index), fornecedores };
  });

  const casa = (f: Fornecedor) => casaBuscaFornecedor(filtros.busca, f.nome, f.doc);

  // Fornecedores em comum: pagos por 2 ou mais senadores
  const todas = new Map<string, Fornecedor>();
  porSenador.forEach((s) => s.fornecedores.forEach((f, k) => !todas.has(k) && todas.set(k, f)));
  const emComum =
    selectedIds.length < 2
      ? []
      : [...todas.values()]
          .filter((f) => porSenador.filter((s) => s.fornecedores.has(f.chave)).length >= 2 && casa(f))
          .map((f) => ({
            ...f,
            total: porSenador.reduce((soma, s) => soma + (s.fornecedores.get(f.chave)?.total ?? 0), 0),
          }))
          .sort((a, b) => b.total - a.total);

  const topPorSenador = porSenador.map((s) => {
    const lista = [...s.fornecedores.values()].filter(casa).sort((a, b) => b.total - a.total);
    return { ...s, encontrados: lista.length, top: lista.slice(0, filtros.top) };
  });

  const yearLabel = year === 0 ? "Mandato completo" : year.toString();
  const recorte = [filtros.categoria, filtros.busca && `busca “${filtros.busca}”`].filter(Boolean).join(" · ");
  const linkCeaps = (id: number) => `/senador/${id}?tab=ceaps${year > 0 ? `&ano=${year}` : ""}`;

  return (
    <div className="space-y-6">
      <BarraFiltros rotulo="Filtros de fornecedores">
        <FiltroSelect
          rotulo="Categoria"
          valor={filtros.categoria ?? TODAS}
          largura="w-[260px]"
          opcoes={[{ valor: TODAS, rotulo: "Todas as categorias" }, ...categorias.map((c) => ({ valor: c, rotulo: c }))]}
          aoMudar={(v) => definir({ [PARAMS_FORNECEDORES.categoria]: v === TODAS ? null : v })}
        />
        <FiltroSelect
          rotulo="Mostrar"
          valor={String(filtros.top)}
          largura="w-[130px]"
          opcoes={OPCOES_TOP.map((n) => ({ valor: String(n), rotulo: `Top ${n}` }))}
          aoMudar={(v) => definir({ [PARAMS_FORNECEDORES.top]: v === "5" ? null : v })}
        />
        <div className="flex flex-col gap-1">
          <label htmlFor={idBusca} className="text-xs font-medium text-muted-foreground">
            Buscar fornecedor
          </label>
          <div className="relative">
            <Search className="pointer-events-none absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" aria-hidden="true" />
            <Input
              id={idBusca}
              type="search"
              value={busca}
              onChange={(e) => setBusca(e.target.value)}
              placeholder="Nome ou CNPJ"
              className="h-9 w-[240px] pl-8"
              autoComplete="off"
            />
          </div>
        </div>
      </BarraFiltros>

      {selectedIds.length > 1 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Building2 className="h-5 w-5" aria-hidden="true" />
              Fornecedores em comum ({yearLabel})
            </CardTitle>
            <CardDescription>
              Pagos com a cota por dois ou mais dos senadores comparados, somados.
              {recorte && ` Recorte: ${recorte}.`}
            </CardDescription>
          </CardHeader>
          <CardContent>
            {emComum.length === 0 ? (
              <p className="text-sm text-muted-foreground">Nenhum fornecedor em comum no recorte.</p>
            ) : (
              <ul className="divide-y">
                {emComum.slice(0, filtros.top).map((f) => (
                  <li key={f.chave} className="flex flex-col gap-2 py-2.5 sm:flex-row sm:items-center sm:justify-between">
                    <div className="min-w-0">
                      <p className="font-medium text-sm sm:text-base">{f.nome}</p>
                      {f.doc && <p className="font-mono text-xs text-muted-foreground">{formatarDocumento(f.doc)}</p>}
                    </div>
                    <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
                      <ul className="flex flex-wrap gap-x-3 gap-y-1 text-xs" aria-label="Pago por senador">
                        {porSenador
                          .filter((s) => s.fornecedores.has(f.chave))
                          .map((s) => (
                            <li key={s.id} className="flex items-center gap-1">
                              <MarcadorSerie cor={s.cor} />
                              <Link href={linkCeaps(s.id)} className="hover:underline">
                                {s.nome.split(" ")[0]}
                              </Link>
                              <span className="tabular-nums text-muted-foreground">
                                {formatarReais(s.fornecedores.get(f.chave)?.total ?? 0)}
                              </span>
                            </li>
                          ))}
                      </ul>
                      <span className="font-bold text-sm tabular-nums">{formatarReais(f.total)}</span>
                    </div>
                  </li>
                ))}
              </ul>
            )}
            {emComum.length > filtros.top && (
              <p className="mt-2 text-xs text-muted-foreground">
                Mostrando {filtros.top} de {emComum.length}. Aumente o “Mostrar” ou refine a busca.
              </p>
            )}
          </CardContent>
        </Card>
      )}

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {topPorSenador.map((senador) => (
          <Card key={senador.id} className="overflow-hidden">
            <div className="h-1.5 w-full" style={{ backgroundColor: senador.cor }} aria-hidden="true" />
            <CardHeader>
              <CardTitle className="text-base truncate" title={senador.nome}>
                <Link href={linkCeaps(senador.id)} className="hover:underline">
                  {senador.nome}
                </Link>
              </CardTitle>
              <CardDescription>
                {senador.encontrados === 0
                  ? "Nenhum fornecedor no recorte."
                  : `Top ${Math.min(filtros.top, senador.encontrados)} de ${senador.encontrados} fornecedores`}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <ol className="space-y-3">
                {senador.top.map((f) => (
                  <li key={f.chave} className="flex items-start justify-between gap-3 text-sm">
                    <div className="min-w-0">
                      <p className="line-clamp-2" title={f.nome}>
                        {f.nome}
                      </p>
                      {f.doc && <p className="font-mono text-xs text-muted-foreground">{formatarDocumento(f.doc)}</p>}
                    </div>
                    <div className="shrink-0 text-right">
                      <p className="font-medium tabular-nums">{formatarReais(f.total)}</p>
                      <p className="text-xs text-muted-foreground">
                        {f.quantidade} {f.quantidade === 1 ? "lançamento" : "lançamentos"}
                      </p>
                    </div>
                  </li>
                ))}
              </ol>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
