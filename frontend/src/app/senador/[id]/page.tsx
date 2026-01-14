import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import Link from "next/link";

// Mock data - will be replaced with API call
const mockSenador = {
  id: 1,
  nome: "Senador Exemplo",
  nome_completo: "Senador Exemplo da Silva",
  partido: "PART",
  uf: "BA",
  foto_url: "",
  email: "senador@senado.leg.br",
  telefone: "(61) 3303-0000",
  score: {
    total: 92.5,
    posicao: 1,
    detalhes: {
      produtividade: 95,
      presenca: 88,
      economia: 90,
      comissoes: 95,
    },
  },
  proposicoes: {
    total: 45,
    aprovadas: 12,
    em_tramitacao: 28,
    arquivadas: 5,
  },
  votacoes: {
    total: 450,
    presentes: 396,
    ausentes: 54,
    percentual_presenca: 88,
  },
  ceaps: {
    gasto_total: 125000,
    teto: 150000,
    economia: 25000,
    percentual_economia: 16.7,
  },
  comissoes: {
    total: 8,
    titularidades: 3,
    suplencias: 5,
    presidencias: 1,
  },
};

export default async function SenadorPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const senador = mockSenador; // Will fetch from API using id

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
            <p className="mt-1 text-lg text-muted-foreground">
              {senador.nome_completo}
            </p>
            <div className="mt-3 flex items-center gap-2">
              <Badge variant="default">{senador.partido}</Badge>
              <Badge variant="outline">{senador.uf}</Badge>
              <Badge variant="secondary">
                #{senador.score.posicao} no ranking
              </Badge>
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
              {senador.score.total.toFixed(1)}
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
              {senador.score.detalhes.produtividade.toFixed(1)}
            </p>
            <div className="mt-2 h-2 w-full overflow-hidden rounded-full bg-muted">
              <div
                className="h-full rounded-full bg-primary transition-all"
                style={{ width: `${senador.score.detalhes.produtividade}%` }}
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
            <p className="text-2xl font-bold">
              {senador.score.detalhes.presenca.toFixed(1)}
            </p>
            <div className="mt-2 h-2 w-full overflow-hidden rounded-full bg-muted">
              <div
                className="h-full rounded-full bg-primary transition-all"
                style={{ width: `${senador.score.detalhes.presenca}%` }}
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
              {senador.score.detalhes.economia.toFixed(1)}
            </p>
            <div className="mt-2 h-2 w-full overflow-hidden rounded-full bg-muted">
              <div
                className="h-full rounded-full bg-primary transition-all"
                style={{ width: `${senador.score.detalhes.economia}%` }}
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
            <p className="text-2xl font-bold">
              {senador.score.detalhes.comissoes.toFixed(1)}
            </p>
            <div className="mt-2 h-2 w-full overflow-hidden rounded-full bg-muted">
              <div
                className="h-full rounded-full bg-primary transition-all"
                style={{ width: `${senador.score.detalhes.comissoes}%` }}
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
                    {senador.proposicoes.total}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    Proposições apresentadas
                  </p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-green-600">
                    {senador.proposicoes.aprovadas}
                  </p>
                  <p className="text-sm text-muted-foreground">Aprovadas</p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-yellow-600">
                    {senador.proposicoes.em_tramitacao}
                  </p>
                  <p className="text-sm text-muted-foreground">Em tramitação</p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-muted-foreground">
                    {senador.proposicoes.arquivadas}
                  </p>
                  <p className="text-sm text-muted-foreground">Arquivadas</p>
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
              <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
                <div>
                  <p className="text-3xl font-bold text-foreground">
                    {senador.votacoes.total}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    Votações no período
                  </p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-green-600">
                    {senador.votacoes.presentes}
                  </p>
                  <p className="text-sm text-muted-foreground">Presenças</p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-red-600">
                    {senador.votacoes.ausentes}
                  </p>
                  <p className="text-sm text-muted-foreground">Ausências</p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-primary">
                    {senador.votacoes.percentual_presenca}%
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
              <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
                <div>
                  <p className="text-3xl font-bold text-foreground">
                    R$ {(senador.ceaps.gasto_total / 1000).toFixed(0)}k
                  </p>
                  <p className="text-sm text-muted-foreground">Gasto total</p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-muted-foreground">
                    R$ {(senador.ceaps.teto / 1000).toFixed(0)}k
                  </p>
                  <p className="text-sm text-muted-foreground">Teto anual</p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-green-600">
                    R$ {(senador.ceaps.economia / 1000).toFixed(0)}k
                  </p>
                  <p className="text-sm text-muted-foreground">Economia</p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-primary">
                    {senador.ceaps.percentual_economia.toFixed(1)}%
                  </p>
                  <p className="text-sm text-muted-foreground">
                    Abaixo do teto
                  </p>
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
                    {senador.comissoes.total}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    Comissões totais
                  </p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-primary">
                    {senador.comissoes.titularidades}
                  </p>
                  <p className="text-sm text-muted-foreground">Titularidades</p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-muted-foreground">
                    {senador.comissoes.suplencias}
                  </p>
                  <p className="text-sm text-muted-foreground">Suplências</p>
                </div>
                <div>
                  <p className="text-3xl font-bold text-yellow-600">
                    {senador.comissoes.presidencias}
                  </p>
                  <p className="text-sm text-muted-foreground">Presidências</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
