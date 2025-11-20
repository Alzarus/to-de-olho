import { API_CONFIG } from '@/config/constants';

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
  alta: 'Alta',
  'alta relevancia': 'Alta',
  altaRelevancia: 'Alta',
  media: 'Média',
  média: 'Média',
  baixa: 'Baixa',
};

const ANALYTICS_REVALIDATE_SECONDS = 300;
export const VOTACOES_ANALYTICS_TAG = 'analytics:votacoes:stats';

const monthFormatter = new Intl.DateTimeFormat('pt-BR', { month: 'short' });

function buildApiUrl(path: string): URL {
  const base = API_CONFIG.BASE_URL.endsWith('/')
    ? API_CONFIG.BASE_URL
    : `${API_CONFIG.BASE_URL}/`;
  const normalizedPath = path.startsWith('/') ? path.slice(1) : path;
  return new URL(normalizedPath, base);
}

async function fetchVotacoesStats(periodo: string): Promise<VotacaoStatsResponse> {
  const url = buildApiUrl('analytics/votacoes/stats');
  url.searchParams.set('periodo', periodo);

  const response = await fetch(url.toString(), {
    headers: {
      Accept: 'application/json',
    },
    next: {
      revalidate: ANALYTICS_REVALIDATE_SECONDS,
      tags: [VOTACOES_ANALYTICS_TAG],
    },
  });

  if (!response.ok) {
    throw new Error('Não foi possível carregar as estatísticas de votações');
  }

  const payload = await response.json();
  const data = payload?.data;

  if (!data || typeof data !== 'object') {
    throw new Error('Resposta da API não contém dados de estatísticas');
  }

  return data as VotacaoStatsResponse;
}

function calcularAprovacao(stats: VotacaoStatsResponse | null) {
  if (!stats || stats.totalVotacoes === 0) {
    return { aprovadas: 0, rejeitadas: 0 };
  }
  const aprovadas = (stats.votacoesAprovadas / stats.totalVotacoes) * 100;
  const rejeitadas = (stats.votacoesRejeitadas / stats.totalVotacoes) * 100;
  return {
    aprovadas: Number(aprovadas.toFixed(1)),
    rejeitadas: Number(rejeitadas.toFixed(1)),
  };
}

function mapRelevancia(stats: VotacaoStatsResponse | null) {
  if (!stats) {
    return [] as Array<{ label: string; valor: number }>;
  }

  return Object.entries(stats.votacoesPorRelevancia ?? {}).map(([key, valor]) => ({
    label: RELEVANCIA_LABELS[key.toLowerCase()] || key,
    valor,
  }));
}

function mapMeses(stats: VotacaoStatsResponse | null) {
  if (!stats) {
    return [] as Array<{ mes: string; quantidade: number }>;
  }

  return (stats.votacoesPorMes ?? []).map((quantidade, index) => {
    const data = new Date(Date.UTC(2020, index, 1));
    return {
      mes: monthFormatter.format(data),
      quantidade,
    };
  });
}

export function VotacoesAnalyticsSkeleton({ className = '' }: { className?: string }) {
  return (
    <div
      className={`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 ${className}`}
      role="status"
      aria-live="polite"
    >
      {Array.from({ length: 4 }).map((_, index) => (
        <div key={index} className="h-28 bg-gray-100 border border-gray-200 rounded-lg animate-pulse" />
      ))}
    </div>
  );
}

export default async function VotacoesAnalytics({
  periodo,
  className = '',
}: VotacoesAnalyticsProps) {
  const periodoSelecionado = periodo ?? new Date().getFullYear().toString();

  let stats: VotacaoStatsResponse | null = null;
  let errorMessage: string | null = null;

  try {
    stats = await fetchVotacoesStats(periodoSelecionado);
  } catch (error) {
    errorMessage =
      error instanceof Error
        ? error.message
        : 'Erro desconhecido ao carregar estatísticas de votações';
  }

  const aprovacaoPercentual = calcularAprovacao(stats);
  const relevanciaDistribuicao = mapRelevancia(stats);
  const mesesSeries = mapMeses(stats);

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
          📅 Período analisado: {periodoSelecionado}
        </span>
      </header>

      {errorMessage && (
        <div className="p-4 bg-red-50 border border-red-200 text-red-700 rounded-lg">
          <p className="font-semibold">Erro ao carregar estatísticas</p>
          <p className="text-sm mt-1">{errorMessage}</p>
        </div>
      )}

      {!errorMessage && stats && (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <article className="bg-white border border-gray-200 rounded-lg p-4 shadow-sm" aria-label="Total de votações">
              <p className="text-sm text-gray-500">Total de votações</p>
              <p className="mt-2 text-3xl font-bold text-gray-900">{stats.totalVotacoes.toLocaleString('pt-BR')}</p>
              <p className="text-xs text-gray-500 mt-1">Inclui plenário e sessões extraordinárias</p>
            </article>

            <article className="bg-white border border-gray-200 rounded-lg p-4 shadow-sm" aria-label="Aprovação das matérias">
              <p className="text-sm text-gray-500">Aprovação</p>
              <p className="mt-2 text-3xl font-bold text-green-600">{stats.votacoesAprovadas.toLocaleString('pt-BR')}</p>
              <p className="text-xs text-gray-500 mt-1">{aprovacaoPercentual.aprovadas}% das votações foram aprovadas</p>
            </article>

            <article className="bg-white border border-gray-200 rounded-lg p-4 shadow-sm" aria-label="Matérias rejeitadas">
              <p className="text-sm text-gray-500">Rejeições</p>
              <p className="mt-2 text-3xl font-bold text-red-600">{stats.votacoesRejeitadas.toLocaleString('pt-BR')}</p>
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
                        style={{
                          width: `${stats.totalVotacoes === 0 ? 0 : Math.round((quantidade / stats.totalVotacoes) * 100)}%`,
                        }}
                      ></div>
                    </div>
                    <span className="w-10 text-sm text-gray-700 text-right">{quantidade}</span>
                  </div>
                ))}

                {mesesSeries.length === 0 && (
                  <p className="text-sm text-gray-500">Nenhum dado mensal disponível para o período selecionado.</p>
                )}
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
                        style={{
                          width: `${stats.totalVotacoes === 0 ? 0 : Math.round((valor / stats.totalVotacoes) * 100)}%`,
                        }}
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
}
