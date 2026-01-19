"use client";

import { useState, useEffect, Suspense } from "react";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import Link from "next/link";
import { useParams, useRouter, useSearchParams, usePathname } from "next/navigation";
import { useSenadorScore, useSenador } from "@/hooks/use-senador";
import { formatCurrency } from "@/lib/utils";
import { VotosPieChart } from "@/components/votos-pie-chart";
import { useVotosPorTipo } from "@/hooks/use-senador";
import { fetcher } from "@/lib/api";
import { X } from "lucide-react";
import { CompareToggleButton } from "@/components/comparator/compare-toggle-button";
import { SenatorRadarChart } from "@/components/senator/radar-chart";
import { EmendasTab } from "@/components/senator/emendas-tab";
import { ProposicoesTab } from "@/components/senator/proposicoes-tab";
import { ComissoesTab } from "@/components/senator/comissoes-tab";
import { CeapsTab } from "@/components/senator/ceaps-tab";




const VOTE_LABELS: Record<string, string> = {
  Sim: "Sim",
  Nao: "Não",
  Abstencao: "Abstenção",
  Obstrucao: "Obstrução",
  NCom: "Não Compareceu",
};

import { VotacoesResponse, VotacaoItem } from "@/types/api";

function VotosChartWrapper({ id }: { id: number }) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const pathname = usePathname();
  const { data, isLoading } = useVotosPorTipo(id);
  
  const selectedVoteType = searchParams.get("voto");
  
  const [votacoes, setVotacoes] = useState<VotacaoItem[]>([]);
  const [votacoesLoading, setVotacoesLoading] = useState(false);

  const updateVoteType = (type: string | null) => {
    const params = new URLSearchParams(searchParams.toString());
    if (type) {
      params.set("voto", type);
    } else {
      params.delete("voto");
    }
    router.replace(`${pathname}?${params.toString()}`, { scroll: false });
  };
  
  // Buscar votações quando tipo selecionado
  useEffect(() => {
    if (!selectedVoteType) {
      setVotacoes([]);
      return;
    }
    
    const fetchVotacoes = async () => {
      setVotacoesLoading(true);
      try {
        const res = await fetcher<VotacoesResponse>(`/api/v1/senadores/${id}/votacoes?limit=1000`);
        
        const filtered = res.votacoes.filter(v => {
          if (selectedVoteType === "Outros") {
            const mainTypes = ["Sim", "Nao", "Abstencao", "Obstrucao"];
            return !mainTypes.includes(v.voto);
          }
          return v.voto === selectedVoteType;
        });
        
        setVotacoes(filtered);
      } catch (error) {
        console.error("Erro ao buscar votações:", error);
      } finally {
        setVotacoesLoading(false);
      }
    };
    
    fetchVotacoes();
  }, [id, selectedVoteType]);

  if (isLoading) return <Skeleton className="h-[300px] w-full" />;
  
  if (!data || !data.por_tipo) return null;

  const handleSliceClick = (voteType: string) => {
    updateVoteType(voteType === selectedVoteType ? null : voteType);
  };

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle>Distribuição de Votos</CardTitle>
        {selectedVoteType && (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => updateVoteType(null)}
            className="text-muted-foreground"
          >
            <X className="mr-1 h-4 w-4" />
            Limpar filtro
          </Button>
        )}
      </CardHeader>
      <CardContent className="space-y-4">
        <VotosPieChart data={data.por_tipo} onSliceClick={handleSliceClick} />
        
        {/* Lista de votações filtradas */}
        {selectedVoteType && (
          <div className="mt-4 border-t pt-4">
            <h4 className="text-sm font-semibold mb-3">
              Votações com voto &quot;{VOTE_LABELS[selectedVoteType] || selectedVoteType}&quot;
              {votacoes.length > 0 && (
                <Badge variant="secondary" className="ml-2">
                  {votacoes.length}
                </Badge>
              )}
            </h4>
            
            {votacoesLoading ? (
              <div className="space-y-2">
                {[...Array(3)].map((_, i) => (
                  <Skeleton key={i} className="h-12 w-full" />
                ))}
              </div>
            ) : votacoes.length === 0 ? (
              <p className="text-sm text-muted-foreground text-center py-4">
                Nenhuma votação encontrada para este tipo.
              </p>
            ) : (
              <div className="space-y-2 max-h-[300px] overflow-y-auto">
                {votacoes.map((v) => (
                  <Link
                    key={v.id}
                    href={`/votacoes/${v.sessao_id}?backUrl=${encodeURIComponent(pathname + "?" + searchParams.toString())}`}
                    className="block p-3 rounded-lg border hover:bg-muted/50 transition-colors"
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div className="flex-1 min-w-0">
                        <p className="font-medium text-sm truncate">
                          {v.materia || "Sem matéria"}
                        </p>
                        <p className="text-xs text-muted-foreground line-clamp-1">
                          {v.descricao_votacao}
                        </p>
                      </div>
                      <Badge variant="outline" className="shrink-0 text-xs">
                        {new Date(v.data).toLocaleDateString("pt-BR")}
                      </Badge>
                    </div>
                  </Link>
                ))}
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

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

function SenadorContent() {
  const params = useParams();
  const router = useRouter();
  const searchParams = useSearchParams();
  const pathname = usePathname();
  
  const id = Number(params.id);
  const [ano, setAno] = useState<number>(0);
  
  // Tab control via URL
  const activeTab = searchParams.get("tab") || "proposicoes";
  const setActiveTab = (tab: string) => {
    const newParams = new URLSearchParams(searchParams.toString());
    newParams.set("tab", tab);
    router.replace(`${pathname}?${newParams.toString()}`, { scroll: false });
  };

  const { data: senador, isLoading, error } = useSenadorScore(id, ano);
  const { data: senadorDetalhes } = useSenador(id);

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

  // Encontrar mandato atual
  const mandatoAtual = senadorDetalhes?.mandatos?.find(m => !m.fim || new Date(m.fim) > new Date()) 
    || senadorDetalhes?.mandatos?.[0];

  const formatDate = (dateString?: string) => {
    if (!dateString) return "Atual";
    return new Date(dateString).toLocaleDateString("pt-BR", { month: 'short', year: 'numeric' });
  };

  return (
    <div className="container mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      {/* Breadcrumb e Seletor de Ano */}
      <div className="mb-8 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <nav aria-label="Breadcrumb">
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

        {/* Year Selector */}
        <div className="flex items-center gap-2">
          <label htmlFor="ano-select" className="text-sm font-medium text-muted-foreground">
            Ano:
          </label>
          <Select
            value={ano.toString()}
            onValueChange={(value) => setAno(Number(value))}
          >
            <SelectTrigger id="ano-select" className="w-[180px]">
              <SelectValue placeholder="Selecione o ano" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="0">Mandato (Todos os anos)</SelectItem>
              <SelectItem value="2026">2026</SelectItem>
              <SelectItem value="2025">2025</SelectItem>
              <SelectItem value="2024">2024</SelectItem>
              <SelectItem value="2023">2023</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>
      
      {/* Header */}
      <div className="mb-12 flex flex-col gap-6 lg:flex-row lg:items-start lg:justify-between">
        <div className="flex flex-col sm:flex-row items-start gap-6">
          {senador.foto_url ? (
            /* eslint-disable-next-line @next/next/no-img-element */
            <img
              src={senador.foto_url}
              alt={senador.nome}
              className="h-24 w-24 rounded-2xl object-cover bg-muted shrink-0"
            />
          ) : (
            <div className="flex h-24 w-24 items-center justify-center rounded-2xl bg-primary/10 text-primary shrink-0">
              <span className="text-3xl font-bold">{senador.nome.charAt(0)}</span>
            </div>
          )}
          <div className="min-w-0">
            <h1 className="text-3xl font-bold tracking-tight text-foreground break-words">
              {senador.nome}
            </h1>
            <div className="mt-3 flex flex-wrap items-center gap-2">
              <Badge variant="default">{senador.partido}</Badge>
              <Badge variant="outline">{senador.uf}</Badge>
              <Badge className="bg-[#d4af37] text-white hover:bg-[#d4af37]/90 whitespace-nowrap">
                #{senador.posicao > 0 ? senador.posicao : "-"} no ranking
              </Badge>
              {mandatoAtual && (
                <Badge variant="secondary" className="whitespace-nowrap">
                  Mandato: {formatDate(mandatoAtual.inicio)} - {formatDate(mandatoAtual.fim)}
                </Badge>
              )}
            </div>
            <div className="mt-4">
              <CompareToggleButton 
                  senator={{
                    id: senador.senador_id,
                    nome: senador.nome,
                    partido: senador.partido,
                    uf: senador.uf,
                    fotoUrl: senador.foto_url
                  }} 
              />
            </div>
          </div>
        </div>

        {/* Score Card */}
        <Card className="w-full lg:w-auto lg:min-w-[200px]">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Score Total ({ano === 0 ? "Mandato" : ano})
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
      {/* Score Details & Radar */}
      <div className="mb-12 grid gap-6 lg:grid-cols-3">
        {/* Radar Chart */}
        <div className="lg:col-span-1 h-[400px] lg:h-auto">
             <SenatorRadarChart score={senador} />
        </div>

        {/* Metrics Cards */}
        <div className="lg:col-span-2 grid gap-4 sm:grid-cols-2 content-start">
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
      </div>

      {/* Detailed Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <div className="w-full overflow-x-auto pb-1 no-scrollbar">
          <TabsList className="w-full justify-start inline-flex min-w-max">
            <TabsTrigger value="proposicoes">Proposições</TabsTrigger>
            <TabsTrigger value="votacoes">Votações</TabsTrigger>
            <TabsTrigger value="ceaps">CEAPS</TabsTrigger>
            <TabsTrigger value="comissoes">Comissões</TabsTrigger>
            <TabsTrigger value="emendas">Emendas</TabsTrigger>
          </TabsList>
        </div>

        <TabsContent value="proposicoes" className="mt-6">
          <div className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Produção Legislativa ({ano === 0 ? "Mandato" : ano})</CardTitle>
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
            <ProposicoesTab id={id} />
          </div>
        </TabsContent>

        <TabsContent value="votacoes" className="mt-6">
          <div className="grid gap-6 lg:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Presença em Votações ({ano === 0 ? "Mandato" : ano})</CardTitle>
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

            <VotosChartWrapper id={id} />
          </div>
        </TabsContent>

        <TabsContent value="ceaps" className="mt-6">
          <div className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Cota para Exercício da Atividade Parlamentar ({ano === 0 ? "Mandato" : ano})</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
                  <div>
                    <p className="text-3xl font-bold text-foreground">
                      {formatCurrency(senador.detalhes.gasto_ceaps)}
                    </p>
                    <p className="text-sm text-muted-foreground">Gasto total</p>
                  </div>
                  <div>
                    <p className="text-3xl font-bold text-muted-foreground">
                      {formatCurrency(senador.detalhes.teto_ceaps)}
                    </p>
                    <p className="text-sm text-muted-foreground">Teto {ano === 0 ? "no periodo" : "anual"}</p>
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
            <CeapsTab id={id} ano={ano} />
          </div>
        </TabsContent>

        <TabsContent value="comissoes" className="mt-6">
          <div className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Participação em Comissões ({ano === 0 ? "Mandato" : ano})</CardTitle>
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
            <ComissoesTab id={id} />
          </div>
        </TabsContent>

        <TabsContent value="emendas" className="mt-6">
          <h2 className="text-2xl font-bold tracking-tight mb-4">
            Emendas Parlamentares ({ano === 0 ? "Mandato" : ano})
          </h2>
          <p className="text-muted-foreground mb-4">
            Recursos destinados através de emendas individuais, de bancada, comissão, relator e transferências especiais (PIX).
            Dados do Portal da Transparência.
          </p>
          <EmendasTab id={id} ano={ano} />
        </TabsContent>
      </Tabs>
    </div>
  );
}

export default function SenadorPage() {
  return (
    <Suspense fallback={<SenadorSkeleton />}>
      <SenadorContent />
    </Suspense>
  );
}
