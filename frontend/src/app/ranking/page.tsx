"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import Link from "next/link";
import { useRanking } from "@/hooks/use-ranking";
import type { SenadorScore } from "@/types/api";

function RankingTable({ data }: { data: SenadorScore[] }) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full">
        <thead>
          <tr className="border-b border-border">
            <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">
              Posição
            </th>
            <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">
              Senador
            </th>
            <th className="hidden px-4 py-3 text-center text-sm font-medium text-muted-foreground sm:table-cell">
              Produtividade
            </th>
            <th className="hidden px-4 py-3 text-center text-sm font-medium text-muted-foreground md:table-cell">
              Presença
            </th>
            <th className="hidden px-4 py-3 text-center text-sm font-medium text-muted-foreground lg:table-cell">
              Economia
            </th>
            <th className="hidden px-4 py-3 text-center text-sm font-medium text-muted-foreground lg:table-cell">
              Comissões
            </th>
            <th className="px-4 py-3 text-right text-sm font-medium text-muted-foreground">
              Score Total
            </th>
          </tr>
        </thead>
        <tbody>
          {data.map((senador) => (
            <tr
              key={senador.senador_id}
              className="border-b border-border transition-colors hover:bg-muted/50"
            >
              <td className="px-4 py-4">
                <div
                  className={`inline-flex h-8 w-8 items-center justify-center rounded-full text-sm font-bold ${
                    senador.posicao === 1
                      ? "bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400"
                      : senador.posicao === 2
                      ? "bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-300"
                      : senador.posicao === 3
                      ? "bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400"
                      : "bg-muted text-muted-foreground"
                  }`}
                >
                  {senador.posicao}
                </div>
              </td>
              <td className="px-4 py-4">
                <Link
                  href={`/senador/${senador.senador_id}`}
                  className="group flex items-center gap-3"
                >
                  {senador.foto_url ? (
                    /* eslint-disable-next-line @next/next/no-img-element */
                    <img
                      src={senador.foto_url}
                      alt={senador.nome}
                      className="h-10 w-10 rounded-full object-cover"
                    />
                  ) : (
                    <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary/10 text-primary">
                      <span className="text-sm font-medium">
                        {senador.nome.charAt(0)}
                      </span>
                    </div>
                  )}
                  <div>
                    <p className="font-medium text-foreground group-hover:text-primary transition-colors">
                      {senador.nome}
                    </p>
                    <p className="text-sm text-muted-foreground">
                      <Badge variant="secondary" className="mr-1">
                        {senador.partido}
                      </Badge>
                      {senador.uf}
                    </p>
                  </div>
                </Link>
              </td>
              <td className="hidden px-4 py-4 text-center sm:table-cell">
                <span className="text-sm font-medium">
                  {senador.produtividade.toFixed(1)}
                </span>
              </td>
              <td className="hidden px-4 py-4 text-center md:table-cell">
                <span className="text-sm font-medium">
                  {senador.presenca.toFixed(1)}
                </span>
              </td>
              <td className="hidden px-4 py-4 text-center lg:table-cell">
                <span className="text-sm font-medium">
                  {senador.economia_cota.toFixed(1)}
                </span>
              </td>
              <td className="hidden px-4 py-4 text-center lg:table-cell">
                <span className="text-sm font-medium">
                  {senador.comissoes.toFixed(1)}
                </span>
              </td>
              <td className="px-4 py-4 text-right">
                <span className="text-lg font-bold text-primary">
                  {senador.score_final.toFixed(1)}
                </span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function RankingTableSkeleton() {
  return (
    <div className="space-y-4">
      {[...Array(10)].map((_, i) => (
        <div key={i} className="flex items-center gap-4 p-4">
          <Skeleton className="h-8 w-8 rounded-full" />
          <Skeleton className="h-10 w-10 rounded-full" />
          <div className="flex-1 space-y-2">
            <Skeleton className="h-4 w-48" />
            <Skeleton className="h-3 w-24" />
          </div>
          <Skeleton className="h-6 w-16" />
        </div>
      ))}
    </div>
  );
}

function RankingError({ message }: { message: string }) {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      <div className="rounded-full bg-destructive/10 p-4">
        <svg
          className="h-8 w-8 text-destructive"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
          />
        </svg>
      </div>
      <h3 className="mt-4 text-lg font-semibold text-foreground">
        Erro ao carregar dados
      </h3>
      <p className="mt-2 text-sm text-muted-foreground max-w-md">{message}</p>
      <p className="mt-4 text-xs text-muted-foreground">
        Verifique se o backend está rodando em localhost:8080
      </p>
    </div>
  );
}

export default function RankingPage() {
  const { data, isLoading, error } = useRanking();

  return (
    <div className="container mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      {/* Header */}
      <div className="mb-12">
        <h1 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
          Ranking de Senadores
        </h1>
        <p className="mt-4 max-w-3xl text-lg text-muted-foreground">
          Avaliação objetiva dos 81 senadores brasileiros baseada em 4
          critérios: produtividade legislativa (35%), presença em votações
          (25%), economia de recursos (20%) e participação em comissões (20%).
        </p>
      </div>

      {/* Criteria Cards */}
      <div className="mb-12 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Produtividade
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-primary">35%</p>
            <p className="mt-1 text-xs text-muted-foreground">
              Proposições apresentadas e aprovadas
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Presença
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-primary">25%</p>
            <p className="mt-1 text-xs text-muted-foreground">
              Participação em votações
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Economia
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-primary">20%</p>
            <p className="mt-1 text-xs text-muted-foreground">
              Uso responsável do CEAPS
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Comissões
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-primary">20%</p>
            <p className="mt-1 text-xs text-muted-foreground">
              Atuação em comissões
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Ranking Table */}
      <Card>
        <CardHeader>
          <CardTitle>
            Classificação Geral
            {data && (
              <span className="ml-2 text-sm font-normal text-muted-foreground">
                ({data.total} senadores)
              </span>
            )}
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {isLoading && <RankingTableSkeleton />}
          {error && (
            <RankingError
              message={
                error instanceof Error
                  ? error.message
                  : "Erro desconhecido ao carregar ranking"
              }
            />
          )}
          {data && <RankingTable data={data.ranking} />}
        </CardContent>
      </Card>

      {/* Methodology Link */}
      <div className="mt-8 text-center">
        <p className="text-sm text-muted-foreground">
          Quer entender como calculamos os scores?{" "}
          <Link
            href="/metodologia"
            className="font-medium text-primary hover:underline"
          >
            Consulte nossa metodologia
          </Link>
        </p>
      </div>
    </div>
  );
}
