import React, { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { PAGINATION_CONFIG } from '../config/constants';

interface VotacaoResumo {
  id: number;
  titulo: string;
  ementa: string;
  data: string;
  aprovacao: string;
  placarSim: number;
  placarNao: number;
  placarAbstencao: number;
  totalVotos: number;
  tipoProposicao: string;
  numeroProposicao: string;
  relevancia: 'alta' | 'média' | 'baixa';
}

interface VotacoesPageProps {
  className?: string;
}

const VotacoesPage: React.FC<VotacoesPageProps> = ({ className = "" }) => {
  const [votacoes, setVotacoes] = useState<VotacaoResumo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filtros, setFiltros] = useState({
    busca: '',
    ano: '',
    aprovacao: '',
    relevancia: ''
  });
  const [paginacao, setPaginacao] = useState({
    pagina: 1,
    totalPaginas: 1,
    total: 0
  });

  const fetchVotacoes = useCallback(async () => {
    try {
      setLoading(true);
      const params = new URLSearchParams({
        pagina: paginacao.pagina.toString(),
        limite: PAGINATION_CONFIG.DEFAULT_PAGE_SIZE.toString(),
        ...(filtros.busca && { busca: filtros.busca }),
        ...(filtros.ano && { ano: filtros.ano }),
        ...(filtros.aprovacao && { aprovacao: filtros.aprovacao }),
        ...(filtros.relevancia && { relevancia: filtros.relevancia })
      });

      const response = await fetch(`/api/v1/votacoes?${params}`);
      
      if (!response.ok) {
        throw new Error('Erro ao carregar votações');
      }
      
      const data = await response.json();
      setVotacoes(data.votacoes || []);
      setPaginacao(prev => ({
        ...prev,
        totalPaginas: data.totalPaginas || 1,
        total: data.total || 0
      }));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro desconhecido');
    } finally {
      setLoading(false);
    }
  }, [filtros, paginacao.pagina]);

  useEffect(() => {
    fetchVotacoes();
  }, [fetchVotacoes]);

  const handleFiltroChange = (campo: string, valor: string) => {
    setFiltros(prev => ({ ...prev, [campo]: valor }));
    setPaginacao(prev => ({ ...prev, pagina: 1 })); // Reset para primeira página
  };

  const getCorAprovacao = (aprovacao: string): string => {
    return aprovacao === 'Aprovada' 
      ? 'bg-green-100 text-green-800 border border-green-200' 
      : 'bg-red-100 text-red-800 border border-red-200';
  };

  const getCorRelevancia = (relevancia: string): string => {
    switch (relevancia) {
      case 'alta':
        return 'bg-red-100 text-red-800 border border-red-200';
      case 'média':
        return 'bg-yellow-100 text-yellow-800 border border-yellow-200';
      case 'baixa':
        return 'bg-gray-100 text-gray-600 border border-gray-200';
      default:
        return 'bg-gray-100 text-gray-600 border border-gray-200';
    }
  };

  const calcularPorcentagem = (votos: number, total: number): number => {
    return total > 0 ? (votos / total) * 100 : 0;
  };

  const anosDisponiveis = Array.from(new Set(
    votacoes.map(v => new Date(v.data).getFullYear())
  )).sort((a, b) => b - a);

  if (error) {
    return (
      <div className={`p-6 bg-red-50 border border-red-200 rounded-lg ${className}`}>
        <h2 className="text-xl font-bold text-red-800 mb-2">Erro ao Carregar Votações</h2>
        <p className="text-red-600 mb-4">{error}</p>
        <button 
          onClick={fetchVotacoes}
          className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 transition-colors"
        >
          Tentar Novamente
        </button>
      </div>
    );
  }

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Cabeçalho */}
      <div className="bg-white rounded-lg shadow-sm border p-6">
        <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">
              🗳️ Votações da Câmara dos Deputados
            </h1>
            <p className="text-gray-600 mt-1">
              Acompanhe como os deputados votaram nas principais proposições
            </p>
            {paginacao.total > 0 && (
              <p className="text-sm text-gray-500 mt-2">
                {paginacao.total} votações encontradas
              </p>
            )}
          </div>
          
          <div className="flex flex-col sm:flex-row gap-3">
            <Link 
              href="/votacoes/analise-completa"
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-center"
            >
              📊 Análise Completa
            </Link>
            <Link 
              href="/votacoes/rankings"
              className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors text-center"
            >
              🏆 Rankings
            </Link>
          </div>
        </div>
      </div>

      {/* Filtros */}
      <div className="bg-white rounded-lg shadow-sm border p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Filtros</h3>
        
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {/* Busca por texto */}
          <div>
            <label htmlFor="busca" className="block text-sm font-medium text-gray-700 mb-1">
              Buscar Votação
            </label>
            <input
              id="busca"
              type="text"
              value={filtros.busca}
              onChange={(e) => handleFiltroChange('busca', e.target.value)}
              placeholder="Título ou ementa..."
              className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            />
          </div>

          {/* Filtro por ano */}
          <div>
            <label htmlFor="ano" className="block text-sm font-medium text-gray-700 mb-1">
              Ano
            </label>
            <select
              id="ano"
              value={filtros.ano}
              onChange={(e) => handleFiltroChange('ano', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="">Todos os anos</option>
              {anosDisponiveis.map(ano => (
                <option key={ano} value={ano}>{ano}</option>
              ))}
            </select>
          </div>

          {/* Filtro por resultado */}
          <div>
            <label htmlFor="aprovacao" className="block text-sm font-medium text-gray-700 mb-1">
              Resultado
            </label>
            <select
              id="aprovacao"
              value={filtros.aprovacao}
              onChange={(e) => handleFiltroChange('aprovacao', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="">Todas</option>
              <option value="Aprovada">Aprovadas</option>
              <option value="Rejeitada">Rejeitadas</option>
            </select>
          </div>

          {/* Filtro por relevância */}
          <div>
            <label htmlFor="relevancia" className="block text-sm font-medium text-gray-700 mb-1">
              Relevância
            </label>
            <select
              id="relevancia"
              value={filtros.relevancia}
              onChange={(e) => handleFiltroChange('relevancia', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="">Todas</option>
              <option value="alta">Alta Relevância</option>
              <option value="média">Média Relevância</option>
              <option value="baixa">Baixa Relevância</option>
            </select>
          </div>
        </div>

        {/* Botão limpar filtros */}
        {(filtros.busca || filtros.ano || filtros.aprovacao || filtros.relevancia) && (
          <div className="mt-4">
            <button
              onClick={() => setFiltros({ busca: '', ano: '', aprovacao: '', relevancia: '' })}
              className="text-sm text-gray-500 hover:text-gray-700"
            >
              🗑️ Limpar todos os filtros
            </button>
          </div>
        )}
      </div>

      {/* Lista de Votações */}
      <div className="space-y-4">
        {loading ? (
          // Loading skeleton
          Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="bg-white rounded-lg shadow-sm border p-6 animate-pulse">
              <div className="space-y-3">
                <div className="h-5 bg-gray-200 rounded w-3/4"></div>
                <div className="h-4 bg-gray-200 rounded w-full"></div>
                <div className="h-4 bg-gray-200 rounded w-2/3"></div>
                <div className="flex gap-4">
                  <div className="h-8 bg-gray-200 rounded w-20"></div>
                  <div className="h-8 bg-gray-200 rounded w-24"></div>
                  <div className="h-8 bg-gray-200 rounded w-16"></div>
                </div>
              </div>
            </div>
          ))
        ) : votacoes.length === 0 ? (
          <div className="bg-white rounded-lg shadow-sm border p-8 text-center">
            <div className="text-gray-400 text-6xl mb-4">🗳️</div>
            <h3 className="text-lg font-semibold text-gray-900 mb-2">
              Nenhuma votação encontrada
            </h3>
            <p className="text-gray-600">
              Tente ajustar os filtros ou aguarde a sincronização dos dados.
            </p>
          </div>
        ) : (
          votacoes.map((votacao) => (
            <div key={votacao.id} className="bg-white rounded-lg shadow-sm border hover:shadow-md transition-shadow">
              <div className="p-6">
                <div className="flex flex-col lg:flex-row lg:items-start gap-4">
                  {/* Conteúdo Principal */}
                  <div className="flex-1">
                    <div className="flex flex-wrap items-center gap-2 mb-3">
                      <span className={`px-2 py-1 rounded text-xs font-medium ${getCorAprovacao(votacao.aprovacao)}`}>
                        {votacao.aprovacao}
                      </span>
                      <span className={`px-2 py-1 rounded text-xs font-medium ${getCorRelevancia(votacao.relevancia)}`}>
                        {votacao.relevancia} relevância
                      </span>
                      <span className="text-xs text-gray-500">
                        {votacao.tipoProposicao} {votacao.numeroProposicao}
                      </span>
                      <span className="text-xs text-gray-500">
                        📅 {new Date(votacao.data).toLocaleDateString('pt-BR')}
                      </span>
                    </div>

                    <h3 className="text-lg font-semibold text-gray-900 mb-2 leading-tight">
                      {votacao.titulo}
                    </h3>

                    <p className="text-gray-700 text-sm mb-4 leading-relaxed line-clamp-2">
                      {votacao.ementa}
                    </p>

                    {/* Estatísticas da Votação */}
                    <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 text-center">
                      <div>
                        <div className="text-lg font-bold text-green-600">{votacao.placarSim}</div>
                        <div className="text-xs text-gray-500">
                          Favoráveis ({calcularPorcentagem(votacao.placarSim, votacao.totalVotos).toFixed(1)}%)
                        </div>
                      </div>
                      <div>
                        <div className="text-lg font-bold text-red-600">{votacao.placarNao}</div>
                        <div className="text-xs text-gray-500">
                          Contrários ({calcularPorcentagem(votacao.placarNao, votacao.totalVotos).toFixed(1)}%)
                        </div>
                      </div>
                      <div>
                        <div className="text-lg font-bold text-yellow-600">{votacao.placarAbstencao}</div>
                        <div className="text-xs text-gray-500">
                          Abstenções ({calcularPorcentagem(votacao.placarAbstencao, votacao.totalVotos).toFixed(1)}%)
                        </div>
                      </div>
                      <div>
                        <div className="text-lg font-bold text-gray-600">{votacao.totalVotos}</div>
                        <div className="text-xs text-gray-500">Total de votos</div>
                      </div>
                    </div>
                  </div>

                  {/* Ações */}
                  <div className="flex-shrink-0">
                    <Link 
                      href={`/votacoes/${votacao.id}`}
                      className="block w-full sm:w-auto px-4 py-2 bg-blue-600 text-white text-center rounded hover:bg-blue-700 transition-colors"
                    >
                      📊 Ver Análise
                    </Link>
                  </div>
                </div>

                {/* Barra de Progresso Visual */}
                <div className="mt-4 bg-gray-200 rounded-full h-2 overflow-hidden">
                  <div className="h-full flex">
                    <div 
                      className="bg-green-500"
                      style={{ width: `${calcularPorcentagem(votacao.placarSim, votacao.totalVotos)}%` }}
                    ></div>
                    <div 
                      className="bg-red-500"
                      style={{ width: `${calcularPorcentagem(votacao.placarNao, votacao.totalVotos)}%` }}
                    ></div>
                    <div 
                      className="bg-yellow-500"
                      style={{ width: `${calcularPorcentagem(votacao.placarAbstencao, votacao.totalVotos)}%` }}
                    ></div>
                  </div>
                </div>
              </div>
            </div>
          ))
        )}
      </div>

      {/* Paginação */}
      {paginacao.totalPaginas > 1 && (
        <div className="bg-white rounded-lg shadow-sm border p-4">
          <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
            <div className="text-sm text-gray-600">
              Página {paginacao.pagina} de {paginacao.totalPaginas}
            </div>
            
            <div className="flex gap-2">
              <button
                onClick={() => setPaginacao(prev => ({ ...prev, pagina: Math.max(1, prev.pagina - 1) }))}
                disabled={paginacao.pagina === 1}
                className="px-3 py-2 text-sm border border-gray-300 rounded disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50 transition-colors"
              >
                ← Anterior
              </button>
              
              {/* Números das páginas */}
              {Array.from({ length: Math.min(5, paginacao.totalPaginas) }, (_, i) => {
                const pageNum = Math.max(1, paginacao.pagina - 2) + i;
                if (pageNum > paginacao.totalPaginas) return null;
                
                return (
                  <button
                    key={pageNum}
                    onClick={() => setPaginacao(prev => ({ ...prev, pagina: pageNum }))}
                    className={`px-3 py-2 text-sm border rounded transition-colors ${
                      pageNum === paginacao.pagina
                        ? 'bg-blue-600 text-white border-blue-600'
                        : 'border-gray-300 hover:bg-gray-50'
                    }`}
                  >
                    {pageNum}
                  </button>
                );
              })}
              
              <button
                onClick={() => setPaginacao(prev => ({ ...prev, pagina: Math.min(prev.totalPaginas, prev.pagina + 1) }))}
                disabled={paginacao.pagina === paginacao.totalPaginas}
                className="px-3 py-2 text-sm border border-gray-300 rounded disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50 transition-colors"
              >
                Próxima →
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default VotacoesPage;