"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Skeleton } from "@/components/ui/skeleton";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useSenadorScore } from "@/hooks/use-senador";

function SenadorSkeleton() {
  return (
    <div className="container mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      <Skeleton className="h-4 w-48 mb-8" />
      <div className="mb-12 flex flex-col gap-6 sm:flex-row sm:items-start sm:justify-between">
        <div className="flex items-start gap-6">
          <Skeleton className="h-24 w-24 rounded-2xl" />
          <div className="space-y-2">
            <Skeleton className="h-8 w-64" />
            <Skeleton className="h-5 w-48" />
            <Skeleton className="h-6 w-32" />
          </div>
        </div>
        <Skeleton className="h-32 w-48" />
      </div>
      <div className="mb-12 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {[...Array(4)].map((_, i) => (
          <Skeleton key={i} className="h-32" />
        ))}
      </div>
    </div>
  );
}

function SenadorError({ message }: { message: string }) {
  return (
    <div className="container mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      <div className="flex flex-col items-center justify-center py-24 text-center">
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
          Senador não encontrado
        </h3>
        <p className="mt-2 text-sm text-muted-foreground max-w-md">{message}</p>
        <Link
          href="/ranking"
          className="mt-6 text-primary hover:underline font-medium"
        >
          Voltar ao ranking
        </Link>
      </div>
    </div>
  );
}

export default function SenadorPage() {
  const params = useParams();
  const id = Number(params.id);
  const { data: senador, isLoading, error } = useSenadorScore(id);

  if (isLoading) {
    return <SenadorSkeleton />;
  }

  if (error || !senador) {
    return (
      <SenadorError
        message={
          error instanceof Error
            ? error.message
            : "Erro ao carregar dados do senador"
        }
      />
    );
  }

  return (
    <div className="container mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      {/* Breadcrumb */}
      <nav className="mb-8" aria-label="Breadcrumb">
        <ol className="flex items-center gap-2 text-sm text-muted-foreground">
          <li>
            <Link href="/" className="hover:text-foreground">
              Início
            </Link>
          </li>
          <li>/</li>
          <li>
            <Link href="/ranking" className="hover:text-foreground">
              Ranking
            </Link>
          </li>
          <li>/</li>
          <li className="text-foreground">{senador.nome}</li>
        </ol>
      </nav>

      {/* Header */}
      <div className="mb-12 flex flex-col gap-6 sm:flex-row sm:items-start sm:justify-between">
        <div className="flex items-start gap-6">
          <div className="flex h-24 w-24 items-center justify-center rounded-2xl bg-primary/10 text-primary">
            <span className="text-3xl font-bold">{senador.nome.charAt(0)}</span>
          </div>
          <div>
            <h1 className="text-3xl font-bold tracking-tight text-foreground">
              {senador.nome}
            </h1>
            <div className="mt-3 flex items-center gap-2">
              <Badge variant="default">{senador.partido}</Badge>
              <Badge variant="outline">{senador.uf}</Badge>
              <Badge variant="secondary">#{senador.posicao} no ranking</Badge>
            </div>
          </div>
        </div>

        {/* Score Card */}
        <Card className="w-full sm:w-auto sm:min-w-[200px]">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Score Total
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-4xl font-bold text-primary">
              {senador.score_final.toFixed(1)}
            </p>
            <p className="mt-1 text-sm text-muted-foreground">
              de 100 pontos possíveis
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Score Details */}
      <div className="mb-12 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Produtividade (35%)
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">
              {senador.produtividade.toFixed(1)}
            </p>
            <div className="mt-2 h-2 w-full overflow-hidden rounded-full bg-muted">
              <div
                className="h-full rounded-full bg-primary transition-all"
                style={{ width: `${Math.min(senador.produtividade, 100)}%` }}
              />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Presença (25%)
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{senador.presenca.toFixed(1)}</p>
            <div className="mt-2 h-2 w-full overflow-hidden rounded-full bg-muted">
              <div
                className="h-full rounded-full bg-primary transition-all"
                style={{ width: `${Math.min(senador.presenca, 100)}%` }}
              />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Economia (20%)
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">
              {senador.economia_cota.toFixed(1)}
            </p>
            <div className="mt-2 h-2 w-full overflow-hidden rounded-full bg-muted">
              <div
                className="h-full rounded-full bg-primary transition-all"
                style={{ width: `${Math.min(senador.economia_cota, 100)}%` }}
              />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Comissões (20%)
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{senador.comissoes.toFixed(1)}</p>
            <div className="mt-2 h-2 w-full overflow-hidden rounded-full bg-muted">
              <div
                className="h-full rounded-full bg-primary transition-all"
                style={{ width: `${Math.min(senador.comissoes, 100)}%` }}
              />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Detailed Tabs */}
      <Tabs defaultValue="proposicoes" className="w-full">
        <TabsList className="w-full justify-start">
          <TabsTrigger value="proposicoes">Proposições</TabsTrigger>
          <TabsTrigger value="votacoes">Votações</TabsTrigger>
          <TabsTrigger value="ceaps">CEAPS</TabsTrigger>
          <TabsTrigger value="comissoes">Comissões</TabsTrigger>
        </TabsList>

        <TabsContent value="proposicoes" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle>Produção Legislativa</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
                <div>
                  <p className="text-3xl font-bold text-foreground">
                    {senador.detalhes.total_proposicoes}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    Proposições apresentadas
                  </p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-green-600">
                    {senador.detalhes.proposicoes_aprovadas}
                  </p>
                  <p className="text-sm text-muted-foreground">Aprovadas</p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-yellow-600">
                    {senador.detalhes.transformadas_em_lei}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    Transformadas em lei
                  </p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-primary">
                    {senador.detalhes.pontuacao_proposicoes}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    Pontuação total
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="votacoes" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle>Presença em Votações</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
                <div>
                  <p className="text-3xl font-bold text-foreground">
                    {senador.detalhes.total_votacoes}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    Votações no período
                  </p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-green-600">
                    {senador.detalhes.votacoes_participadas}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    Votações participadas
                  </p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-primary">
                    {senador.detalhes.taxa_presenca_bruta.toFixed(1)}%
                  </p>
                  <p className="text-sm text-muted-foreground">
                    Taxa de presença
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="ceaps" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle>Cota para Exercício da Atividade Parlamentar</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
                <div>
                  <p className="text-3xl font-bold text-foreground">
                    R${" "}
                    {(senador.detalhes.gasto_ceaps / 1000).toLocaleString(
                      "pt-BR",
                      { minimumFractionDigits: 1 }
                    )}
                    k
                  </p>
                  <p className="text-sm text-muted-foreground">Gasto total</p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-muted-foreground">
                    R${" "}
                    {(senador.detalhes.teto_ceaps / 1000).toLocaleString(
                      "pt-BR",
                      { minimumFractionDigits: 1 }
                    )}
                    k
                  </p>
                  <p className="text-sm text-muted-foreground">Teto anual</p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-green-600">
                    {(
                      ((senador.detalhes.teto_ceaps -
                        senador.detalhes.gasto_ceaps) /
                        senador.detalhes.teto_ceaps) *
                      100
                    ).toFixed(1)}
                    %
                  </p>
                  <p className="text-sm text-muted-foreground">Economia</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="comissoes" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle>Participação em Comissões</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
                <div>
                  <p className="text-3xl font-bold text-foreground">
                    {senador.detalhes.comissoes_ativas}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    Comissões ativas
                  </p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-primary">
                    {senador.detalhes.comissoes_titular}
                  </p>
                  <p className="text-sm text-muted-foreground">Titularidades</p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-muted-foreground">
                    {senador.detalhes.comissoes_suplente}
                  </p>
                  <p className="text-sm text-muted-foreground">Suplências</p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-yellow-600">
                    {senador.detalhes.pontos_comissoes.toFixed(0)}
                  </p>
                  <p className="text-sm text-muted-foreground">Pontos</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
