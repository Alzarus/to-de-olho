"use client";

import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";

import { getRanking } from "@/lib/api";
import { EmendasTab } from "@/components/senator/emendas-tab";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Input } from "@/components/ui/input";
import { Search, X } from "lucide-react";
import type { SenadorScore } from "@/types/api";

const ANOS_DISPONIVEIS = [0, 2026, 2025, 2024, 2023];

export default function EmendasPage() {
  const [search, setSearch] = useState("");
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [selectedUf, setSelectedUf] = useState("TODOS");
  const [selectedPartido, setSelectedPartido] = useState("TODOS");
  const [ano, setAno] = useState<number>(2024);

  const { data, isLoading } = useQuery({
    queryKey: ["ranking-for-emendas"],
    queryFn: () => getRanking(100),
    staleTime: 1000 * 60 * 60,
  });

  const senadores = useMemo(() => data?.ranking || [], [data?.ranking]);

  const uniqueUfs = useMemo(() => {
    const ufs = new Set<string>();
    senadores.forEach((sen) => {
      if (sen.uf) ufs.add(sen.uf);
    });
    return Array.from(ufs).sort();
  }, [senadores]);

  const uniquePartidos = useMemo(() => {
    const partidos = new Set<string>();
    senadores.forEach((sen) => {
      if (sen.partido) partidos.add(sen.partido);
    });
    return Array.from(partidos).sort();
  }, [senadores]);

  const filteredSenadores = useMemo(() => {
    const trimmed = search.trim().toLowerCase();
    return senadores
      .filter((sen) => (trimmed ? sen.nome.toLowerCase().includes(trimmed) : true))
      .filter((sen) => (selectedUf === "TODOS" ? true : sen.uf === selectedUf))
      .filter((sen) => (selectedPartido === "TODOS" ? true : sen.partido === selectedPartido));
  }, [senadores, search, selectedUf, selectedPartido]);

  const effectiveSelectedId = selectedId ?? filteredSenadores[0]?.senador_id ?? null;
  const selectedSenador: SenadorScore | undefined = filteredSenadores.find(
    (sen) => sen.senador_id === effectiveSelectedId,
  );

  return (
    <div className="container mx-auto max-w-7xl px-4 py-8 pb-24 sm:px-6 sm:py-12 sm:pb-12 lg:px-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold tracking-tight">Emendas Parlamentares</h1>
        <p className="mt-2 text-muted-foreground max-w-3xl">
          Consulte a execução de emendas por senador, com destaque para as transferências especiais (PIX).
        </p>
      </div>

      <Card className="mb-8">
        <CardHeader>
          <CardTitle>Filtrar por senador</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <div className="space-y-2">
              <label htmlFor="senador-input" className="text-sm font-medium text-muted-foreground">
                Buscar e selecionar senador
              </label>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  id="senador-input"
                  placeholder="Digite o nome do senador..."
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  className="pl-9"
                />
                {search && (
                  <button
                    type="button"
                    onClick={() => setSearch("")}
                    className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                    aria-label="Limpar busca"
                  >
                    <X className="h-4 w-4" />
                  </button>
                )}
              </div>
            </div>

            <div className="space-y-2">
              <label htmlFor="uf-select" className="text-sm font-medium text-muted-foreground">
                Estado
              </label>
              <select
                id="uf-select"
                value={selectedUf}
                onChange={(e) => setSelectedUf(e.target.value)}
                className="h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
              >
                <option value="TODOS">Todos os estados</option>
                {uniqueUfs.map((uf) => (
                  <option key={uf} value={uf}>
                    {uf}
                  </option>
                ))}
              </select>
            </div>

            <div className="space-y-2">
              <label htmlFor="partido-select" className="text-sm font-medium text-muted-foreground">
                Partido
              </label>
              <select
                id="partido-select"
                value={selectedPartido}
                onChange={(e) => setSelectedPartido(e.target.value)}
                className="h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
              >
                <option value="TODOS">Todos os partidos</option>
                {uniquePartidos.map((partido) => (
                  <option key={partido} value={partido}>
                    {partido}
                  </option>
                ))}
              </select>
            </div>

            <div className="space-y-2">
              <label htmlFor="ano-select" className="text-sm font-medium text-muted-foreground">
                Ano
              </label>
              <select
                id="ano-select"
                value={ano}
                onChange={(e) => setAno(Number(e.target.value))}
                className="h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
              >
                {ANOS_DISPONIVEIS.map((anoOption) => (
                  <option key={anoOption} value={anoOption}>
                    {anoOption === 0 ? "Mandato (todos os anos)" : anoOption}
                  </option>
                ))}
              </select>
            </div>
          </div>

          {isLoading ? (
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {[...Array(6)].map((_, i) => (
                <Skeleton key={i} className="h-24 w-full" />
              ))}
            </div>
          ) : filteredSenadores.length === 0 ? (
            <div className="rounded-md border border-dashed p-6 text-center text-sm text-muted-foreground">
              Nenhum senador encontrado com esse filtro.
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
              {filteredSenadores.slice(0, 12).map((senador) => {
                const isSelected = senador.senador_id === effectiveSelectedId;
                return (
                  <button
                    key={senador.senador_id}
                    type="button"
                    onClick={() => setSelectedId(senador.senador_id)}
                    aria-pressed={isSelected}
                    aria-label={`Selecionar ${senador.nome} (${senador.partido}-${senador.uf})`}
                    className={
                      "flex flex-col items-center rounded-lg border-2 p-3 text-center transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring " +
                      (isSelected
                        ? "border-primary bg-primary/5 shadow-sm"
                        : "border-border hover:border-muted-foreground/40")
                    }
                  >
                    {senador.foto_url ? (
                      /* eslint-disable-next-line @next/next/no-img-element */
                      <img
                        src={senador.foto_url}
                        alt=""
                        className="mb-2 h-14 w-14 rounded-full object-cover"
                      />
                    ) : (
                      <div className="mb-2 flex h-14 w-14 items-center justify-center rounded-full bg-muted text-sm font-semibold">
                        {senador.nome.charAt(0)}
                      </div>
                    )}
                    <span className="text-sm font-semibold text-foreground line-clamp-2">
                      {senador.nome}
                    </span>
                    <span className="mt-1 text-xs text-muted-foreground">
                      {senador.partido}-{senador.uf}
                    </span>
                  </button>
                );
              })}
              {filteredSenadores.length > 12 && (
                <div className="rounded-lg border border-dashed p-3 text-xs text-muted-foreground">
                  Refine a busca para ver mais resultados.
                </div>
              )}
            </div>
          )}

        </CardContent>
      </Card>

      {selectedSenador ? (
        <EmendasTab id={selectedSenador.senador_id} ano={ano} />
      ) : (
        <Card>
          <CardContent className="py-10 text-center text-sm text-muted-foreground">
            Selecione um senador para visualizar as emendas.
          </CardContent>
        </Card>
      )}
    </div>
  );
}
