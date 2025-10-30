'use client';

import React, { useEffect, useMemo, useState } from 'react';

interface RankingDeputado {
  idDeputado: number;
  nome: string;
  siglaPartido: string;
  siglaUf: string;
  urlFoto?: string;
  totalVotacoes: number;
  votosFavoraveis: number;
  votosContrarios: number;
  abstencoes: number;
  taxaAprovacao: number;
}

interface VotacoesRankingProps {
  ano?: number | string;
  limite?: number;
  className?: string;
}

const FOTO_PLACEHOLDER =
  'https://www.camara.leg.br/internet/deputado/img/deputado-sem-foto.jpg';

const VotacoesRanking: React.FC<VotacoesRankingProps> = ({
  ano,
  limite = 5,
  className = '',
}) => {
  const [ranking, setRanking] = useState<RankingDeputado[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const periodo = useMemo(() => {
    if (typeof ano === 'number') {
      return ano;
    }
    if (typeof ano === 'string' && ano.trim().length > 0) {
      return Number.parseInt(ano, 10);
    }
    return new Date().getFullYear();
  }, [ano]);

  useEffect(() => {
    const controller = new AbortController();

    const fetchRanking = async () => {
      try {
        setLoading(true);
        setError(null);

        const params = new URLSearchParams();
        params.set('ano', periodo.toString());
        params.set('limite', limite.toString());

        const response = await fetch(
          `/api/v1/analytics/votacoes/ranking?${params.toString()}`,
          {
            signal: controller.signal,
            cache: 'no-store',
          },
        );

        if (!response.ok) {
          throw new Error('Não foi possível carregar o ranking de deputados');
        }

        const payload = await response.json();
        const dados = Array.isArray(payload?.data) ? payload.data : [];

        setRanking(dados as RankingDeputado[]);
      } catch (err) {
        if (controller.signal.aborted) {
          return;
        }
        setError(
          err instanceof Error
            ? err.message
            : 'Erro desconhecido ao carregar ranking',
        );
      } finally {
        if (!controller.signal.aborted) {
          setLoading(false);
        }
      }
    };

    fetchRanking();

    return () => controller.abort();
  }, [limite, periodo]);

  return (
    <section
      className={`bg-white border border-gray-200 rounded-xl shadow-sm p-6 ${className}`}
      aria-labelledby="ranking-votacoes"
    >
      <header className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 mb-4">
        <div>
          <h2 id="ranking-votacoes" className="text-xl font-semibold text-gray-900">
            Ranking de atuação em plenário
          </h2>
          <p className="text-sm text-gray-600">
            Os deputados mais presentes e com maior volume de votos registrados no período selecionado.
          </p>
        </div>
        <span className="inline-flex items-center gap-2 text-sm text-gray-500 bg-gray-100 border border-gray-200 px-3 py-1 rounded-full">
          🗓️ Ano analisado: {periodo}
        </span>
      </header>

      {loading && (
        <div className="space-y-4" role="status" aria-live="polite">
          {Array.from({ length: limite }).map((_, index) => (
            <div
              key={index}
              className="h-20 bg-gray-100 border border-gray-200 rounded-lg animate-pulse"
            />
          ))}
        </div>
      )}

      {error && !loading && (
        <div className="p-4 bg-red-50 border border-red-200 text-red-700 rounded-lg">
          <p className="font-semibold">Erro ao carregar ranking</p>
          <p className="text-sm mt-1">{error}</p>
        </div>
      )}

      {!loading && !error && ranking.length === 0 && (
        <p className="text-sm text-gray-500">
          Ainda não há dados suficientes para montar o ranking deste período.
        </p>
      )}

      {!loading && !error && ranking.length > 0 && (
        <ol className="space-y-4">
          {ranking.map((deputado, index) => {
            const taxaAprovacao = Number.isFinite(deputado.taxaAprovacao)
              ? deputado.taxaAprovacao
              : 0;
            const foto = deputado.urlFoto?.trim()
              ? deputado.urlFoto
              : FOTO_PLACEHOLDER;

            return (
              <li key={deputado.idDeputado}>
                <article className="flex flex-col sm:flex-row sm:items-center gap-4 p-4 border border-gray-200 rounded-lg hover:border-blue-300 transition-colors">
                  <span className="text-2xl font-bold text-blue-600">#{index + 1}</span>
                  <div className="flex items-center gap-3 flex-1">
                    <img
                      src={foto}
                      alt={`Foto do deputado ${deputado.nome}`}
                      className="w-14 h-14 rounded-full object-cover border border-gray-200"
                      loading="lazy"
                    />
                    <div>
                      <h3 className="text-lg font-semibold text-gray-900 leading-tight">
                        {deputado.nome}
                      </h3>
                      <p className="text-sm text-gray-600">
                        {deputado.siglaPartido} · {deputado.siglaUf}
                      </p>
                    </div>
                  </div>

                  <dl className="grid grid-cols-2 sm:grid-cols-3 gap-3 w-full sm:w-auto text-sm text-gray-600">
                    <div className="text-center sm:text-right">
                      <dt className="uppercase text-xs tracking-wide text-gray-500">Votações</dt>
                      <dd className="text-base font-semibold text-gray-900">
                        {deputado.totalVotacoes}
                      </dd>
                    </div>
                    <div className="text-center sm:text-right">
                      <dt className="uppercase text-xs tracking-wide text-gray-500">Aprovação</dt>
                      <dd className="text-base font-semibold text-green-600">
                        {taxaAprovacao.toFixed(1)}%
                      </dd>
                    </div>
                    <div className="text-center sm:text-right">
                      <dt className="uppercase text-xs tracking-wide text-gray-500">Saldo</dt>
                      <dd className="text-base font-semibold text-blue-600">
                        {deputado.votosFavoraveis - deputado.votosContrarios}
                      </dd>
                    </div>
                  </dl>
                </article>
              </li>
            );
          })}
        </ol>
      )}
    </section>
  );
};

export default VotacoesRanking;
