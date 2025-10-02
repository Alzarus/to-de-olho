import React, { useState, useEffect, useCallback } from 'react';

interface VotacaoPartido {
  partido: string;
  orientacao: string;
  votaramFavor: number;
  votaramContra: number;
  votaramAbstencao: number;
  totalMembros: number;
  disciplina: number; // % que seguiu orientação
}

interface VotacaoDetalhada {
  id: number;
  titulo: string;
  ementa: string;
  data: string;
  aprovacao: string;
  placarSim: number;
  placarNao: number;
  placarAbstencao: number;
  orientacoes: VotacaoPartido[];
}

interface VotacaoAnalysisProps {
  votacaoId?: number;
  className?: string;
}

const VotacaoAnalysis: React.FC<VotacaoAnalysisProps> = ({ 
  votacaoId = 1, // Default para PEC da Blindagem (exemplo)
  className = ""
}) => {
  const [votacao, setVotacao] = useState<VotacaoDetalhada | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filtroPartido, setFiltroPartido] = useState('');

  const fetchVotacaoDetalhada = useCallback(async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/v1/votacoes/${votacaoId}/completa`);
      
      if (!response.ok) {
        throw new Error('Erro ao carregar dados da votação');
      }
      
      const data = await response.json();
      setVotacao(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro desconhecido');
    } finally {
      setLoading(false);
    }
  }, [votacaoId]);

  useEffect(() => {
    fetchVotacaoDetalhada();
  }, [fetchVotacaoDetalhada]);

  const getCorOrientacao = (orientacao: string): string => {
    switch (orientacao.toLowerCase()) {
      case 'sim':
      case 'favorável':
        return 'bg-green-100 text-green-800 border border-green-200';
      case 'não':
      case 'contrário':
        return 'bg-red-100 text-red-800 border border-red-200';
      case 'liberado':
        return 'bg-gray-100 text-gray-800 border border-gray-200';
      case 'abstenção':
        return 'bg-yellow-100 text-yellow-800 border border-yellow-200';
      default:
        return 'bg-gray-100 text-gray-600 border border-gray-200';
    }
  };

  const getCorResultado = (aprovacao: string): string => {
    return aprovacao === 'Aprovada' 
      ? 'bg-green-50 border-green-200' 
      : 'bg-red-50 border-red-200';
  };

  const partidosFiltrados = votacao?.orientacoes.filter(partido =>
    filtroPartido === '' || 
    partido.partido.toLowerCase().includes(filtroPartido.toLowerCase())
  ) || [];

  if (loading) {
    return (
      <div className={`p-6 bg-white rounded-lg shadow-sm border ${className}`}>
        <div className="animate-pulse space-y-4">
          <div className="h-6 bg-gray-200 rounded w-3/4"></div>
          <div className="space-y-3">
            {[...Array(8)].map((_, i) => (
              <div key={i} className="h-4 bg-gray-200 rounded"></div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`p-6 bg-red-50 border border-red-200 rounded-lg ${className}`}>
        <h3 className="text-lg font-semibold text-red-800 mb-2">Erro ao Carregar Votação</h3>
        <p className="text-red-600">{error}</p>
        <button 
          onClick={fetchVotacaoDetalhada}
          className="mt-4 px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 transition-colors"
        >
          Tentar Novamente
        </button>
      </div>
    );
  }

  if (!votacao) {
    return (
      <div className={`p-6 bg-gray-50 border border-gray-200 rounded-lg ${className}`}>
        <p className="text-gray-600">Nenhuma votação encontrada.</p>
      </div>
    );
  }

  return (
    <div className={`bg-white rounded-lg shadow-sm border ${className}`}>
      {/* Cabeçalho da Votação */}
      <div className={`p-6 border-b ${getCorResultado(votacao.aprovacao)}`}>
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
          <div className="flex-1">
            <h2 className="text-xl font-bold text-gray-900 mb-2">
              {votacao.titulo}
            </h2>
            <p className="text-gray-700 text-sm mb-3 leading-relaxed">
              {votacao.ementa}
            </p>
            <div className="flex flex-wrap gap-3 text-sm">
              <span className="text-gray-600">
                📅 {new Date(votacao.data).toLocaleDateString('pt-BR')}
              </span>
              <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                votacao.aprovacao === 'Aprovada' 
                  ? 'bg-green-100 text-green-800' 
                  : 'bg-red-100 text-red-800'
              }`}>
                {votacao.aprovacao}
              </span>
            </div>
          </div>

          {/* Placar Geral */}
          <div className="bg-white rounded-lg p-4 border shadow-sm min-w-[200px]">
            <h3 className="font-semibold text-gray-900 mb-3 text-center">Placar Final</h3>
            <div className="space-y-2">
              <div className="flex justify-between items-center">
                <span className="text-green-600 font-medium">✅ Favoráveis:</span>
                <span className="font-bold text-green-600">{votacao.placarSim}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-red-600 font-medium">❌ Contrários:</span>
                <span className="font-bold text-red-600">{votacao.placarNao}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-yellow-600 font-medium">⚪ Abstenções:</span>
                <span className="font-bold text-yellow-600">{votacao.placarAbstencao}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Filtro por Partido */}
      <div className="p-4 bg-gray-50 border-b">
        <div className="flex flex-col sm:flex-row gap-3 items-start sm:items-center">
          <label htmlFor="filtro-partido" className="text-sm font-medium text-gray-700">
            Filtrar Partido:
          </label>
          <input
            id="filtro-partido"
            type="text"
            value={filtroPartido}
            onChange={(e) => setFiltroPartido(e.target.value)}
            placeholder="Digite o nome do partido..."
            className="px-3 py-2 border border-gray-300 rounded-md text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          />
          {filtroPartido && (
            <button
              onClick={() => setFiltroPartido('')}
              className="text-sm text-gray-500 hover:text-gray-700"
            >
              Limpar filtro
            </button>
          )}
        </div>
      </div>

      {/* Lista de Partidos */}
      <div className="p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">
          Como Cada Partido Votou
          <span className="text-sm text-gray-500 ml-2">
            ({partidosFiltrados.length} partidos)
          </span>
        </h3>

        <div className="space-y-3">
          {partidosFiltrados.map((partido) => (
            <div 
              key={partido.partido}
              className="border border-gray-200 rounded-lg p-4 hover:shadow-sm transition-shadow"
            >
              <div className="flex flex-col lg:flex-row lg:items-center gap-4">
                {/* Info do Partido */}
                <div className="flex-shrink-0 lg:w-48">
                  <h4 className="font-bold text-gray-900 text-lg">{partido.partido}</h4>
                  <div className="flex items-center gap-2 mt-1">
                    <span className={`px-2 py-1 rounded text-xs font-medium ${getCorOrientacao(partido.orientacao)}`}>
                      Orientou: {partido.orientacao}
                    </span>
                    <span className="text-xs text-gray-500">
                      {partido.disciplina.toFixed(1)}% disciplina
                    </span>
                  </div>
                </div>

                {/* Distribuição de Votos */}
                <div className="flex-1">
                  <div className="grid grid-cols-1 sm:grid-cols-4 gap-3">
                    <div className="text-center">
                      <div className="text-lg font-bold text-green-600">{partido.votaramFavor}</div>
                      <div className="text-xs text-gray-500">Favoráveis</div>
                    </div>
                    <div className="text-center">
                      <div className="text-lg font-bold text-red-600">{partido.votaramContra}</div>
                      <div className="text-xs text-gray-500">Contrários</div>
                    </div>
                    <div className="text-center">
                      <div className="text-lg font-bold text-yellow-600">{partido.votaramAbstencao}</div>
                      <div className="text-xs text-gray-500">Abstenções</div>
                    </div>
                    <div className="text-center">
                      <div className="text-lg font-bold text-gray-600">{partido.totalMembros}</div>
                      <div className="text-xs text-gray-500">Total</div>
                    </div>
                  </div>

                  {/* Barra de Progresso Visual */}
                  <div className="mt-3 bg-gray-200 rounded-full h-3 overflow-hidden">
                    <div className="h-full flex">
                      <div 
                        className="bg-green-500"
                        style={{ width: `${(partido.votaramFavor / partido.totalMembros) * 100}%` }}
                      ></div>
                      <div 
                        className="bg-red-500"
                        style={{ width: `${(partido.votaramContra / partido.totalMembros) * 100}%` }}
                      ></div>
                      <div 
                        className="bg-yellow-500"
                        style={{ width: `${(partido.votaramAbstencao / partido.totalMembros) * 100}%` }}
                      ></div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>

        {partidosFiltrados.length === 0 && filtroPartido && (
          <div className="text-center py-8 text-gray-500">
            <p>Nenhum partido encontrado com o filtro &ldquo;{filtroPartido}&rdquo;</p>
          </div>
        )}
      </div>

      {/* Rodapé com Insights */}
      <div className="p-4 bg-gray-50 border-t rounded-b-lg">
        <div className="text-sm text-gray-600 space-y-1">
          <p>💡 <strong>Disciplina partidária:</strong> Percentual de deputados que seguiram a orientação do partido</p>
          <p>📊 <strong>Dados atualizados:</strong> Informações obtidas diretamente da API da Câmara dos Deputados</p>
        </div>
      </div>
    </div>
  );
};

export default VotacaoAnalysis;