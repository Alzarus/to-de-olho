import { API_CONFIG } from '@/config/constants';

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

const RANKING_REVALIDATE_SECONDS = 300;
export const VOTACOES_RANKING_TAG = 'analytics:votacoes:ranking';

function buildApiUrl(path: string): URL {
  const base = API_CONFIG.BASE_URL.endsWith('/')
    ? API_CONFIG.BASE_URL
    : `${API_CONFIG.BASE_URL}/`;
  const normalizedPath = path.startsWith('/') ? path.slice(1) : path;
  return new URL(normalizedPath, base);
}

async function fetchRanking(ano: number, limite: number): Promise<RankingDeputado[]> {
  const url = buildApiUrl('analytics/votacoes/ranking');
  url.searchParams.set('ano', ano.toString());
  url.searchParams.set('limite', limite.toString());

  const response = await fetch(url.toString(), {
    headers: {
      Accept: 'application/json',
    },
    next: {
      revalidate: RANKING_REVALIDATE_SECONDS,
      tags: [VOTACOES_RANKING_TAG],
    },
  });

  if (!response.ok) {
    throw new Error('Não foi possível carregar o ranking de deputados');
  }

  const payload = await response.json();
  const data = payload?.data;

  if (!Array.isArray(data)) {
    throw new Error('Resposta da API não contém o ranking de deputados');
  }

  return data as RankingDeputado[];
}

function parseAno(ano?: number | string) {
  if (typeof ano === 'number' && Number.isFinite(ano)) {
    return ano;
  }
  if (typeof ano === 'string' && ano.trim().length > 0) {
    const parsed = Number.parseInt(ano, 10);
    if (!Number.isNaN(parsed)) {
      return parsed;
    }
  }
  return new Date().getFullYear();
}

export function VotacoesRankingSkeleton({
  limite = 5,
  className = '',
}: {
  limite?: number;
  className?: string;
}) {
  return (
    <div className={`space-y-4 ${className}`} role="status" aria-live="polite">
      {Array.from({ length: limite }).map((_, index) => (
        <div
          key={index}
          className="h-20 bg-gray-100 border border-gray-200 rounded-lg animate-pulse"
        />
      ))}
    </div>
  );
}

export default async function VotacoesRanking({
  ano,
  limite = 5,
  className = '',
}: VotacoesRankingProps) {
  const anoSelecionado = parseAno(ano);

  let ranking: RankingDeputado[] = [];
  let errorMessage: string | null = null;

  try {
    ranking = await fetchRanking(anoSelecionado, limite);
  } catch (error) {
    errorMessage =
      error instanceof Error
        ? error.message
        : 'Erro desconhecido ao carregar ranking de votações';
  }

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
          🗓️ Ano analisado: {anoSelecionado}
        </span>
      </header>

      {errorMessage && (
        <div className="p-4 bg-red-50 border border-red-200 text-red-700 rounded-lg">
          <p className="font-semibold">Erro ao carregar ranking</p>
          <p className="text-sm mt-1">{errorMessage}</p>
        </div>
      )}

      {!errorMessage && ranking.length === 0 && (
        <p className="text-sm text-gray-500">
          Ainda não há dados suficientes para montar o ranking deste período.
        </p>
      )}

      {!errorMessage && ranking.length > 0 && (
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
}
