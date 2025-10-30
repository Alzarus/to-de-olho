'use client';

import React, { useEffect, useMemo, useState } from "react";

interface VotacaoStatsResponse {
  totalVotacoes: number;
  votacoesAprovadas: number;
  votacoesRejeitadas: number;
  mediaParticipacao: number;
  votacoesPorMes: number[];
  votacoesPorRelevancia: Record<string, number>;
}

interface VotacoesAnalyticsProps {
  periodo?: string;
  className?: string;
}

const RELEVANCIA_LABELS: Record<string, string> = {
  alta: "Alta",
  "alta relevancia": "Alta",
  altaRelevancia: "Alta",
  media: "Média",
  média: "Média",
  baixa: "Baixa",
};

const monthFormatter = new Intl.DateTimeFormat("pt-BR", { month: "short" });

const VotacoesAnalytics: React.FC<VotacoesAnalyticsProps> = ({ periodo, className = "" }) => {
  const [stats, setStats] = useState<VotacaoStatsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const controller = new AbortController();

    const fetchStats = async () => {
      try {
        setLoading(true);
        setError(null);

        const params = new URLSearchParams();
        params.set("periodo", periodo || new Date().getFullYear().toString());

        const response = await fetch(`/api/v1/analytics/votacoes/stats?${params.toString()}`, {
          signal: controller.signal,
          cache: "no-store",
        });

        if (!response.ok) {
          throw new Error("Não foi possível carregar as estatísticas de votações");
        }

        const payload = await response.json();
        if (!payload?.data) {
          throw new Error("Resposta da API não contém dados de estatísticas");
        }

        setStats(payload.data as VotacaoStatsResponse);
      } catch (err) {
        if (controller.signal.aborted) {
          return;
        }
        setError(err instanceof Error ? err.message : "Erro desconhecido ao carregar estatísticas");
      } finally {
        if (!controller.signal.aborted) {
          setLoading(false);
        }
      }
    };

    fetchStats();

    return () => {
      controller.abort();
    };
  }, [periodo]);

  const aprovacaoPercentual = useMemo(() => {
    if (!stats || stats.totalVotacoes === 0) {
      return { aprovadas: 0, rejeitadas: 0 };
    }
    const aprovadas = (stats.votacoesAprovadas / stats.totalVotacoes) * 100;
    const rejeitadas = (stats.votacoesRejeitadas / stats.totalVotacoes) * 100;
    return {
      aprovadas: Number(aprovadas.toFixed(1)),
      rejeitadas: Number(rejeitadas.toFixed(1)),
    };
  }, [stats]);

  const relevanciaDistribuicao = useMemo(() => {
    if (!stats) {
      return [] as Array<{ label: string; valor: number }>;
    }

    const entries = Object.entries(stats.votacoesPorRelevancia || {});
    return entries.map(([key, valor]) => ({
      label: RELEVANCIA_LABELS[key.toLowerCase()] || key,
      valor,
    }));
  }, [stats]);

  const mesesSeries = useMemo(() => {
    if (!stats) {
      return [] as Array<{ mes: string; quantidade: number }>;
    }

    return stats.votacoesPorMes.map((quantidade, index) => {
      const data = new Date();
      data.setMonth(index);
      return {
        mes: monthFormatter.format(data),
        quantidade,
      };
    });
  }, [stats]);

  return (
    <section className={`space-y-6 ${className}`} aria-labelledby="analytics-votacoes">
      <header className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
        <div>
          <h2 id="analytics-votacoes" className="text-xl font-semibold text-gray-900">
            Panorama das votações
          </h2>
          <p className="text-sm text-gray-600">
            Dados consolidados diretamente do backfill histórico e sincronizações diárias da Câmara dos Deputados.
          </p>
        </div>
        <span className="inline-flex items-center gap-2 text-sm text-gray-500 bg-gray-100 border border-gray-200 px-3 py-1 rounded-full">
          📅 Período analisado: {periodo || new Date().getFullYear()}
        </span>
      </header>

      {loading && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4" role="status" aria-live="polite">
          {Array.from({ length: 4 }).map((_, index) => (
            <div key={index} className="h-28 bg-gray-100 border border-gray-200 rounded-lg animate-pulse" />
          ))}
        </div>
      )}

      {error && !loading && (
        <div className="p-4 bg-red-50 border border-red-200 text-red-700 rounded-lg">
          <p className="font-semibold">Erro ao carregar estatísticas</p>
          <p className="text-sm mt-1">{error}</p>
        </div>
      )}

      {!loading && !error && stats && (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <article className="bg-white border border-gray-200 rounded-lg p-4 shadow-sm" aria-label="Total de votações">
              <p className="text-sm text-gray-500">Total de votações</p>
              <p className="mt-2 text-3xl font-bold text-gray-900">{stats.totalVotacoes.toLocaleString("pt-BR")}</p>
              <p className="text-xs text-gray-500 mt-1">Inclui plenário e sessões extraordinárias</p>
            </article>

            <article className="bg-white border border-gray-200 rounded-lg p-4 shadow-sm" aria-label="Aprovação das matérias">
              <p className="text-sm text-gray-500">Aprovação</p>
              <p className="mt-2 text-3xl font-bold text-green-600">{stats.votacoesAprovadas.toLocaleString("pt-BR")}</p>
              <p className="text-xs text-gray-500 mt-1">{aprovacaoPercentual.aprovadas}% das votações foram aprovadas</p>
            </article>

            <article className="bg-white border border-gray-200 rounded-lg p-4 shadow-sm" aria-label="Matérias rejeitadas">
              <p className="text-sm text-gray-500">Rejeições</p>
              <p className="mt-2 text-3xl font-bold text-red-600">{stats.votacoesRejeitadas.toLocaleString("pt-BR")}</p>
              <p className="text-xs text-gray-500 mt-1">{aprovacaoPercentual.rejeitadas}% das matérias foram rejeitadas</p>
            </article>

            <article className="bg-white border border-gray-200 rounded-lg p-4 shadow-sm" aria-label="Média de participação">
              <p className="text-sm text-gray-500">Participação média</p>
              <p className="mt-2 text-3xl font-bold text-blue-600">{stats.mediaParticipacao.toFixed(1)}%</p>
              <p className="text-xs text-gray-500 mt-1">Percentual médio de votos registrados por votação</p>
            </article>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <section className="bg-white border border-gray-200 rounded-lg p-6 shadow-sm" aria-label="Votações por mês">
              <h3 className="text-lg font-semibold text-gray-900">Volume mensal</h3>
              <p className="text-sm text-gray-600 mt-1">Distribuição das votações ao longo dos meses</p>

              <div className="mt-4 space-y-3">
                {mesesSeries.map(({ mes, quantidade }) => (
                  <div key={mes} className="flex items-center gap-4" aria-label={`${quantidade} votações em ${mes}`}>
                    <span className="w-12 text-xs font-semibold text-gray-500 uppercase">{mes}</span>
                    <div className="flex-1 h-3 bg-gray-100 rounded-full overflow-hidden">
                      <div
                        className="h-full bg-blue-500"
                        style={{ width: `${stats.totalVotacoes === 0 ? 0 : Math.round((quantidade / stats.totalVotacoes) * 100)}%` }}
                      ></div>
                    </div>
                    <span className="w-10 text-sm text-gray-700 text-right">{quantidade}</span>
                  </div>
                ))}
              </div>
            </section>

            <section className="bg-white border border-gray-200 rounded-lg p-6 shadow-sm" aria-label="Relevância das matérias">
              <h3 className="text-lg font-semibold text-gray-900">Relevância das pautas</h3>
              <p className="text-sm text-gray-600 mt-1">Como o plenário classificou a importância das matérias analisadas</p>

              <div className="mt-4 space-y-4">
                {relevanciaDistribuicao.map(({ label, valor }) => (
                  <div key={label}>
                    <div className="flex items-center justify-between text-sm text-gray-600">
                      <span>{label}</span>
                      <span className="font-medium text-gray-900">{valor}</span>
                    </div>
                    <div className="mt-2 h-3 bg-gray-100 rounded-full overflow-hidden">
                      <div
                        className="h-full bg-green-500"
                        style={{ width: `${stats.totalVotacoes === 0 ? 0 : Math.round((valor / stats.totalVotacoes) * 100)}%` }}
                      ></div>
                    </div>
                  </div>
                ))}

                {relevanciaDistribuicao.length === 0 && (
                  <p className="text-sm text-gray-500">Nenhum dado de relevância disponível para o período selecionado.</p>
                )}
              </div>
            </section>
          </div>
        </>
      )}
    </section>
  );
};

export default VotacoesAnalytics;
