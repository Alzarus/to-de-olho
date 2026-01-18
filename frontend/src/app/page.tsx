"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ArrowRight, BarChart3, Coins, Users, Activity, ExternalLink, BookOpen } from "lucide-react";

// ... existing code ...

            <div className="flex flex-col w-full sm:flex-row items-center justify-center gap-4 mt-8">
              <Button asChild size="lg" className="w-full sm:w-auto px-8 h-12 text-base shadow-lg shadow-primary/20">
                <Link href="/ranking">
                  Explorar Ranking
                  <ArrowRight className="ml-2 h-4 w-4" />
                </Link>
              </Button>
              <Button asChild variant="outline" size="lg" className="w-full sm:w-auto px-8 h-12 text-base hover:bg-muted/50">
                <Link href="/comparar">
                  Comparar Senadores
                </Link>
              </Button>
               <Button asChild variant="outline" size="lg" className="w-full sm:w-auto px-8 h-12 text-base hover:bg-muted/50">
                <Link href="/metodologia">
                  Entenda o Cálculo
                </Link>
              </Button>
            </div>
import { useRanking } from "@/hooks/use-ranking";
import { Skeleton } from "@/components/ui/skeleton";

export default function Home() {
  // Fetch Top 3 Senators for the podium
  const { data: rankingData, isLoading } = useRanking(3);

  // Stats - mocked for now, but could be calculated from full list
  const stats = [
    { 
      label: "Senadores Monitorados", 
      value: "81", 
      icon: Users,
      description: "Cobertura completa"
    },
    { 
      label: "Votos Registrados", 
      value: "170k+", 
      icon: Activity,
      description: "Desde 2019"
    },
    { 
      label: "Economia Identificada", 
      value: "R$ 14M", 
      icon: Coins,
      description: "Em cotas parlamentares"
    },
    { 
      label: "Dados Processados", 
      value: "100%", 
      icon: BarChart3,
      description: "Fontes oficiais"
    },
  ];

  return (
    <div className="flex flex-col min-h-screen">
      {/* Hero Section with Mesh Gradient */}
      <section className="relative overflow-hidden py-20 sm:py-32 lg:pb-32 xl:pb-36">
        {/* Mesh Gradient Background */}
        <div className="absolute inset-0 -z-10 h-full w-full bg-background">
          <div className="absolute top-0 z-[-2] h-screen w-screen bg-[radial-gradient(ellipse_80%_80%_at_50%_-20%,rgba(59,130,246,0.3),rgba(255,255,255,0))] dark:bg-[radial-gradient(ellipse_80%_80%_at_50%_-20%,rgba(30,58,138,0.5),rgba(0,0,0,0))]"></div>
          <div className="absolute bottom-0 right-0 z-[-2] h-[500px] w-[500px] rounded-full bg-yellow-500/10 blur-[100px] dark:bg-yellow-500/5"></div>
        </div>

        <div className="container mx-auto px-4 md:px-6">
          <div className="flex flex-col items-center space-y-4 text-center">

            
            <h1 className="text-4xl font-bold tracking-tighter sm:text-5xl md:text-6xl/none">
                <span className="bg-clip-text text-transparent bg-gradient-to-r from-foreground to-foreground/70">Transparência no</span> <span className="text-primary">Senado Federal</span>
            </h1>
            
            <p className="mx-auto max-w-[700px] text-muted-foreground md:text-xl leading-relaxed">
              Acompanhe a atuação parlamentar com métricas objetivas. 
              Ranking de produtividade, transparência fiscal e análise detalhada de votos.
            </p>
            
            <div className="flex flex-col w-full sm:flex-row items-center justify-center gap-4 mt-8">
              <Button asChild size="lg" className="w-full sm:w-auto px-8 h-12 text-base shadow-lg shadow-primary/20">
                <Link href="/ranking">
                  Explorar Ranking
                  <ArrowRight className="ml-2 h-4 w-4" />
                </Link>
              </Button>
              <Button asChild variant="outline" size="lg" className="w-full sm:w-auto px-8 h-12 text-base hover:bg-muted/50">
                <Link href="/comparar">
                  Comparar Senadores
                </Link>
              </Button>
               <Button asChild variant="ghost" size="lg" className="w-full sm:w-auto px-8 h-12 text-base">
                <Link href="/metodologia">
                  Entenda o Cálculo
                </Link>
              </Button>
            </div>
          </div>
        </div>
      </section>

      {/* Stats Grid */}
      <section className="container mx-auto px-4 md:px-6 -mt-12 mb-20 relative z-10">
        <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
          {stats.map((stat, index) => (
            <Card key={index} className="bg-background/60 backdrop-blur-sm border-muted/40 shadow-sm hover:shadow-md transition-all duration-300 hover:-translate-y-1">
              <CardContent className="p-6 flex flex-col items-center text-center space-y-2">
                <div className="p-3 rounded-full bg-primary/10 text-primary mb-2">
                    <stat.icon size={24} />
                </div>
                <h3 className="text-2xl font-bold tracking-tight">{stat.value}</h3>
                <p className="text-sm font-medium text-muted-foreground">{stat.label}</p>
                <p className="text-xs text-muted-foreground/60">{stat.description}</p>
              </CardContent>
            </Card>
          ))}
        </div>
      </section>

      {/* Top Ranking Preview */}
      <section className="container mx-auto px-4 md:px-6 mb-24">
        <div className="flex flex-col md:flex-row items-end justify-between mb-10 gap-4">
            <div>
                <h2 className="text-3xl font-bold tracking-tight mb-2">Destaques do Ranking</h2>
                <p className="text-muted-foreground">Os senadores com maior pontuação geral no mandato atual.</p>
            </div>
            <Button variant="ghost" className="hidden md:flex gap-1" asChild>
                <Link href="/ranking">
                    Ver lista completa <ArrowRight size={16} />
                </Link>
            </Button>
        </div>

        <div className="grid gap-6 md:grid-cols-3">
            {isLoading ? (
                Array(3).fill(0).map((_, i) => (
                    <Card key={i} className="overflow-hidden">
                        <CardHeader className="pb-2">
                            <Skeleton className="h-12 w-12 rounded-full" />
                            <Skeleton className="h-4 w-1/2 mt-2" />
                        </CardHeader>
                        <CardContent>
                            <Skeleton className="h-32 w-full" />
                        </CardContent>
                    </Card>
                ))
            ) : (
                rankingData?.ranking.slice(0, 3).map((senator, index) => (
                    <Card key={senator.senador_id} className={`overflow-hidden border-t-4 transition-all hover:shadow-lg ${
                        index === 0 ? "border-t-yellow-500" :
                        index === 1 ? "border-t-gray-400" :
                        "border-t-amber-700"
                    }`}>
                        <CardContent className="p-6">
                            <div className="flex items-start justify-between mb-6">
                                <div className="flex gap-4">
                                     {/* eslint-disable-next-line @next/next/no-img-element */}
                                    <img 
                                        src={senator.foto_url} 
                                        alt={senator.nome}
                                        className="h-16 w-16 rounded-full object-cover border-2 border-background shadow-md" 
                                    />
                                    <div>
                                        <div className="flex items-center gap-2 mb-1">
                                            <Badge variant="secondary" className="text-xs">{index + 1}º Lugar</Badge>
                                        </div>
                                        <h3 className="font-bold text-lg leading-tight">{senator.nome}</h3>
                                        <p className="text-sm text-muted-foreground">{senator.partido} • {senator.uf}</p>
                                    </div>
                                </div>
                                <div className="text-right">
                                    <span className="block text-3xl font-bold tracking-tighter text-primary">
                                        {senator.score_final.toFixed(1)}
                                    </span>
                                    <span className="text-xs font-medium text-muted-foreground">Pontos</span>
                                </div>
                            </div>
                            
                            <div className="space-y-3">
                                <div className="space-y-1">
                                    <div className="flex justify-between text-xs">
                                        <span className="text-muted-foreground">Produtividade</span>
                                        <span className="font-medium">{senator.produtividade.toFixed(1)}</span>
                                    </div>
                                    <div className="h-1.5 w-full bg-secondary rounded-full overflow-hidden">
                                        <div className="h-full bg-blue-500 rounded-full" style={{ width: `${senator.produtividade}%` }}></div>
                                    </div>
                                </div>
                                <div className="space-y-1">
                                    <div className="flex justify-between text-xs">
                                        <span className="text-muted-foreground">Presença</span>
                                        <span className="font-medium">{senator.presenca.toFixed(1)}</span>
                                    </div>
                                     <div className="h-1.5 w-full bg-secondary rounded-full overflow-hidden">
                                        <div className="h-full bg-green-500 rounded-full" style={{ width: `${senator.presenca}%` }}></div>
                                    </div>
                                </div>
                                <div className="space-y-1">
                                    <div className="flex justify-between text-xs">
                                        <span className="text-muted-foreground">Economia</span>
                                        <span className="font-medium">{senator.economia_cota?.toFixed(1) || '0.0'}</span>
                                    </div>
                                     <div className="h-1.5 w-full bg-secondary rounded-full overflow-hidden">
                                        <div className="h-full bg-amber-500 rounded-full" style={{ width: `${senator.economia_cota || 0}%` }}></div>
                                    </div>
                                </div>
                                <div className="space-y-1">
                                    <div className="flex justify-between text-xs">
                                        <span className="text-muted-foreground">Comissões</span>
                                        <span className="font-medium">{senator.comissoes?.toFixed(1) || '0.0'}</span>
                                    </div>
                                     <div className="h-1.5 w-full bg-secondary rounded-full overflow-hidden">
                                        <div className="h-full bg-violet-500 rounded-full" style={{ width: `${senator.comissoes || 0}%` }}></div>
                                    </div>
                                </div>
                            </div>

                            <Button className="w-full mt-6" variant="outline" size="sm" asChild>
                                <Link href={`/senador/${senator.senador_id}`}>
                                    Ver Detalhes <ExternalLink className="ml-2 h-3 w-3" />
                                </Link>
                            </Button>
                        </CardContent>
                    </Card>
                ))
            )}
        </div>
        
        <div className="mt-8 text-center md:hidden">
            <Button variant="ghost" asChild>
                <Link href="/ranking">
                    Ver todos os senadores
                </Link>
            </Button>
        </div>
      </section>
    </div>
  );
}
