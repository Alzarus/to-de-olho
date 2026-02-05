"use client";

import { useState, useEffect, Suspense } from "react";
import { useParams, useSearchParams } from "next/navigation";
import Link from "next/link";
import { format } from "date-fns";
import { ptBR } from "date-fns/locale";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ArrowLeft, Calendar, FileText, CheckCircle2, XCircle, MinusCircle, AlertCircle, HelpCircle, Search, X } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

import { getVotacaoById, VotacaoDetail } from "@/services/votacaoService";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

const VOTE_DESCRIPTIONS: Record<string, string> = {
  AP: "Atividade Parlamentar",
  LP: "Licença Particular",
  LS: "Licença Saúde",
  LG: "Licença Gestante",
  LC: "Licença Conjunta",
  MIS: "Missão Oficial",
  NCom: "Não Compareceu",
  "P-NR": "Presidente (Não Votou)",
  "P-NRV": "Presidente (Não Registrou Voto)",
  "P-OD": "Presidente (Obstrução)",
};

const VoteBadge = ({ voto }: { voto: string }) => {
  switch (voto) {
    case "Sim":
      return (
        <Badge className="bg-green-100 text-green-800 hover:bg-green-100 dark:bg-green-900/30 dark:text-green-400">
          <CheckCircle2 className="mr-1 h-3 w-3" /> Sim
        </Badge>
      );
    case "Nao":
      return (
        <Badge className="bg-red-100 text-red-800 hover:bg-red-100 dark:bg-red-900/30 dark:text-red-400">
          <XCircle className="mr-1 h-3 w-3" /> Não
        </Badge>
      );
    case "Abstencao":
      return (
        <Badge variant="secondary">
          <MinusCircle className="mr-1 h-3 w-3" /> Abstenção
        </Badge>
      );
    case "Obstrucao":
      return (
        <Badge variant="outline" className="border-yellow-500 text-yellow-500">
          <AlertCircle className="mr-1 h-3 w-3" /> Obstrução
        </Badge>
      );
    default: // NCom or other
      const description = VOTE_DESCRIPTIONS[voto] || "Outros / Sem descrição";
      return (
        <TooltipProvider delayDuration={0}>
          <Tooltip>
             <TooltipTrigger asChild>
                <Badge variant="outline" className="text-muted-foreground cursor-help">
                  <HelpCircle className="mr-1 h-3 w-3" /> {voto}
                </Badge>
             </TooltipTrigger>
             <TooltipContent>
                <p>{description}</p>
             </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      );
  }
};

function VotacaoDetalheContent() {
  const params = useParams();
  const searchParams = useSearchParams();
  const backUrl = searchParams.get("backUrl") || "/votacoes";

  const id = params?.id as string;
  const [data, setData] = useState<VotacaoDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  // Filters
  const [filterNome, setFilterNome] = useState("");
  const [filterVoto, setFilterVoto] = useState("");
  const [filterPartido, setFilterPartido] = useState("");
  const [filterUF, setFilterUF] = useState("");

  useEffect(() => {
    if (!id) return;

    const fetchData = async () => {
      setLoading(true);
      try {
        const res = await getVotacaoById(id);
        setData(res);
      } catch (err) {
        console.error("Failed to fetch votacao detail", err);
        setError("Não foi possível carregar os detalhes da votação.");
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [id]);

  if (loading) {
    return (
      <div className="container mx-auto max-w-7xl px-4 py-12">
        <Skeleton className="h-8 w-64 mb-4" />
        <Skeleton className="h-32 w-full mb-8" />
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {[...Array(6)].map((_, i) => (
                <Skeleton key={i} className="h-24 w-full" />
            ))}
        </div>
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="container mx-auto max-w-7xl px-4 py-12 flex flex-col items-center justify-center min-h-[50vh]">
        <AlertCircle className="h-12 w-12 text-destructive mb-4" />
        <h2 className="text-xl font-semibold mb-2">Erro</h2>
        <p className="text-muted-foreground mb-4">{error || "Votação não encontrada"}</p>
        <Button asChild>
          <Link href="/votacoes">Voltar para Votações</Link>
        </Button>
      </div>
    );
  }

  const { votacao, votos } = data;

  // Derived filter options
  const partidos = Array.from(new Set(votos.map(v => v.senador_partido).filter(Boolean))).sort();
  const ufs = Array.from(new Set(votos.map(v => v.senador_uf).filter(Boolean))).sort();

  // Filter Logic
  const filteredVotos = votos.filter(v => {
      const matchNome = v.senador_nome.toLowerCase().includes(filterNome.toLowerCase());
      const matchVoto = filterVoto ? (
          filterVoto === "Outros" 
            ? !["Sim", "Nao", "Abstencao", "Obstrucao"].includes(v.voto) 
            : v.voto === filterVoto
      ) : true;
      const matchPartido = filterPartido ? v.senador_partido === filterPartido : true;
      const matchUF = filterUF ? v.senador_uf === filterUF : true;

      return matchNome && matchVoto && matchPartido && matchUF;
  });

  // Stats
  const sim = votos.filter(v => v.voto === "Sim").length;
  const nao = votos.filter(v => v.voto === "Nao").length;
  const outros = votos.length - sim - nao;

// ... (inside return)
  return (
    <div className="container mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      {/* Back Link */}
      <div className="mb-6">
        <Link 
            href={backUrl}
            className="inline-flex items-center text-sm font-medium text-muted-foreground hover:text-primary"
        >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Voltar para lista
        </Link>
      </div>

      {/* Header Info */}
      <div className="mb-8">
        <div className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
            <div>
                 <Badge variant="outline" className="mb-2">{votacao.codigo_sessao}</Badge>
                 <h1 className="text-2xl font-bold tracking-tight text-foreground sm:text-3xl mb-2">
                    {votacao.materia || "Sem Matéria Vinculada"}
                 </h1>
                 <p className="text-lg text-muted-foreground max-w-4xl">
                    {votacao.descricao_votacao}
                 </p>
            </div>
            <div className="mt-4 sm:mt-0 text-right">
                <p className="text-sm font-medium text-muted-foreground">Data da Votação</p>
                <p className="text-lg font-semibold">
                    {format(new Date(votacao.data), "dd/MM/yyyy", { locale: ptBR })}
                </p>
                <p className="text-sm text-muted-foreground mt-1">
                    {format(new Date(votacao.data), "HH:mm", { locale: ptBR })}
                </p>
            </div>
        </div>
      </div>

      {/* Scoreboard */}
      <Card className="mb-8 bg-muted/30">
        <CardContent className="p-4 sm:p-6">
            <div className="grid grid-cols-3 gap-2 sm:gap-4 text-center">
                <div className="flex flex-col items-center border-r border-border">
                    <span className="text-2xl sm:text-3xl font-bold text-green-600 dark:text-green-400">{sim}</span>
                    <span className="text-xs sm:text-sm font-medium text-muted-foreground uppercase tracking-wider">Sim</span>
                </div>
                <div className="flex flex-col items-center border-r border-border">
                    <span className="text-2xl sm:text-3xl font-bold text-red-600 dark:text-red-400">{nao}</span>
                    <span className="text-xs sm:text-sm font-medium text-muted-foreground uppercase tracking-wider">Nao</span>
                </div>
                <div className="flex flex-col items-center">
                    <span className="text-2xl sm:text-3xl font-bold text-gray-600 dark:text-gray-400">{outros}</span>
                    <span className="text-xs sm:text-sm font-medium text-muted-foreground uppercase tracking-wider">
                      Outros
                    </span>
                </div>
            </div>
            {/* Progress Bar */}
            <div className="mt-4 sm:mt-6 flex h-3 sm:h-4 w-full overflow-hidden rounded-full bg-muted">
                <div style={{ width: `${(sim / votos.length) * 100}%` }} className="bg-green-500" />
                <div style={{ width: `${(nao / votos.length) * 100}%` }} className="bg-red-500" />
                <div style={{ width: `${(outros / votos.length) * 100}%` }} className="bg-gray-400" />
            </div>
            {/* Legenda de Siglas */}
            <p className="mt-3 text-[10px] sm:text-xs text-muted-foreground/80 text-center leading-relaxed">
              <strong>Outros:</strong> Abstencao, Obstrucao, NCom (Nao Compareceu), AP (Ausente por Missao), 
              P-NRV (Presente Nao Votou), LS (Licenca Saude), MIS (Missao), LP (Licenca Particular), Presidente (art. 51 RISF).
            </p>
        </CardContent>
      </Card>

      {/* Filters */}
      <Card className="mb-8">
        <CardContent className="p-4 sm:p-6">
            <h3 className="text-sm font-medium text-muted-foreground mb-3">Filtrar Votos</h3>
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4">
                {/* Search Name */}
                <div className="relative">
                    <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                    <Input 
                        placeholder="Nome do Senador..." 
                        className="pl-9 pr-8" 
                        value={filterNome}
                        onChange={(e) => setFilterNome(e.target.value)}
                    />
                    {filterNome && (
                        <button
                            type="button"
                            onClick={() => setFilterNome("")}
                            className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                            aria-label="Limpar filtro de nome"
                        >
                            <X className="h-4 w-4" />
                        </button>
                    )}
                </div>

                {/* Filter Vote */}
                <Select
                    value={filterVoto}
                    onValueChange={(val) => setFilterVoto(val === "all" ? "" : val)}
                >
                    <SelectTrigger className="w-full sm:min-w-[140px]">
                        <SelectValue placeholder="Todos os Votos" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="all">Todos os Votos</SelectItem>
                        <SelectItem value="Sim">Sim</SelectItem>
                        <SelectItem value="Nao">Não</SelectItem>
                        <SelectItem value="Abstencao">Abstenção</SelectItem>
                        <SelectItem value="Obstrucao">Obstrução</SelectItem>
                        <SelectItem value="Outros">Outros</SelectItem>
                    </SelectContent>
                </Select>

                {/* Filter Party */}
                <Select
                    value={filterPartido}
                    onValueChange={(val) => setFilterPartido(val === "all" ? "" : val)}
                >
                    <SelectTrigger className="w-full sm:min-w-[140px]">
                        <SelectValue placeholder="Todos os Partidos" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="all">Todos os Partidos</SelectItem>
                        {partidos.map((p) => (
                            <SelectItem key={p} value={p}>
                                {p}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>

                {/* Filter UF */}
                <Select
                    value={filterUF}
                    onValueChange={(val) => setFilterUF(val === "all" ? "" : val)}
                >
                    <SelectTrigger className="w-full sm:min-w-[140px]">
                        <SelectValue placeholder="Todas as UFs" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="all">Todas as UFs</SelectItem>
                        {ufs.map((uf) => (
                            <SelectItem key={uf} value={uf}>
                                {uf}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
            </div>
            {/* Active Filters count or Clear */}
            {(filterNome || filterVoto || filterPartido || filterUF) && (
                <div className="mt-4 flex justify-end">
                    <Button variant="ghost" size="sm" onClick={() => {
                        setFilterNome("");
                        setFilterVoto("");
                        setFilterPartido("");
                        setFilterUF("");
                    }} className="text-muted-foreground hover:text-foreground">
                        Limpar Filtros
                    </Button>
                </div>
            )}
        </CardContent>
      </Card>

      {/* List of Votes */}
      <h2 className="text-xl font-bold mb-4">Votos dos Senadores ({filteredVotos.length})</h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
        {filteredVotos.map((voto) => (
            <Card key={voto.senador_id} className="overflow-hidden hover:bg-muted/30 transition-colors">
                <CardContent className="p-4 flex items-center gap-3">
                    {/* Photo */}
                     {voto.senador_foto ? (
                        /* eslint-disable-next-line @next/next/no-img-element */
                        <img
                          src={voto.senador_foto}
                          alt={voto.senador_nome}
                          loading="lazy"
                          className="h-12 w-12 rounded-full object-cover border border-border"
                        />
                      ) : (
                        <div className="h-12 w-12 rounded-full bg-primary/10 flex items-center justify-center text-primary font-bold">
                          {voto.senador_nome.charAt(0)}
                        </div>
                      )}
                    
                    <div className="flex-1 min-w-0">
                        <Link href={`/senador/${voto.senador_id}`} className="block truncate font-medium hover:underline">
                            {voto.senador_nome}
                        </Link>
                        <p className="text-xs text-muted-foreground">
                            {voto.senador_partido} - {voto.senador_uf}
                        </p>
                        <div className="mt-2">
                             <VoteBadge voto={voto.voto} />
                        </div>
                    </div>
                </CardContent>
            </Card>
        ))}
      </div>
      {filteredVotos.length === 0 && (
          <div className="text-center py-12 text-muted-foreground">
              Nenhum senador encontrado com os filtros selecionados.
          </div>
      )}
    </div>
  );
}

export default function VotacaoDetalhePage() {
  return (
    <Suspense fallback={<div className="container py-12 text-center">Carregando detalhes...</div>}>
      <VotacaoDetalheContent />
    </Suspense>
  );
}
