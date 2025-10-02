'use client';

import { useState, useEffect } from 'react';
import axios from 'axios';
import { TrendingUp, Award, Users, DollarSign, RefreshCw, AlertCircle } from 'lucide-react';
import Tooltip from './Tooltip';
import { API_CONFIG, PAGINATION_CONFIG } from '../config/constants';

interface RankingItem {
  id: number;
  nome: string;
  partido: string;
  uf: string;
  total_gasto?: number;
  total_proposicoes?: number;
  percentual_presenca?: number;
  posicao: number;
}

// interface RankingResponse {
//   ano: number;
//   total_geral?: number;
//   media_gastos?: number;
//   deputados: RankingItem[];
// }

interface Insights {
  total_deputados: number;
  total_gasto_ano: number;
  total_proposicoes_ano: number;
  media_gastos_deputado: number;
  partido_maior_gasto: string;
  uf_maior_gasto: string;
  ultima_atualizacao: string;
}

const API_BASE_URL = API_CONFIG.BASE_URL;

export default function DashboardAnalytics() {
  const [rankings, setRankings] = useState<{
    gastos: RankingItem[];
    proposicoes: RankingItem[];
    presenca: RankingItem[];
  }>({ gastos: [], proposicoes: [], presenca: [] });
  
  const [insights, setInsights] = useState<Insights | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedRanking, setSelectedRanking] = useState<'gastos' | 'proposicoes' | 'presenca'>('gastos');

  useEffect(() => {
    fetchAnalytics();
  }, []);

  const fetchAnalytics = async () => {
    setLoading(true);
    setError(null);
    
    try {
      // Buscar rankings e insights em paralelo
      const [gastosRes, proposicoesRes, presencaRes, insightsRes] = await Promise.all([
        axios.get(`${API_BASE_URL}/analytics/rankings/gastos?limit=${PAGINATION_CONFIG.ANALYTICS_RANKING_SIZE}`),
        axios.get(`${API_BASE_URL}/analytics/rankings/proposicoes?limit=${PAGINATION_CONFIG.ANALYTICS_RANKING_SIZE}`),
        axios.get(`${API_BASE_URL}/analytics/rankings/presenca?limit=${PAGINATION_CONFIG.ANALYTICS_RANKING_SIZE}`),
        axios.get(`${API_BASE_URL}/analytics/insights`)
      ]);

      setRankings({
        gastos: gastosRes.data.data?.deputados || [],
        proposicoes: proposicoesRes.data.data?.deputados || [],
        presenca: presencaRes.data.data?.deputados || []
      });
      
      setInsights(insightsRes.data.data || null);
    } catch (err) {
      console.error('Erro ao buscar analytics:', err);
      setError('Erro ao carregar dados analíticos. Verifique se o backend está rodando.');
    } finally {
      setLoading(false);
    }
  };

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL'
    }).format(value);
  };

  const formatPercentage = (value: number) => {
    return `${value.toFixed(1)}%`;
  };

  if (loading) {
    return (
      <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
        <div className="flex items-center justify-center py-8" role="status" aria-live="polite">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-700" aria-hidden="true"></div>
          <span className="ml-3 text-lg text-gray-900">Carregando análises...</span>
          <span className="sr-only">Por favor aguarde, processando dados analíticos</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-6" role="alert">
        <div className="flex items-center">
          <AlertCircle className="h-6 w-6 text-red-600 mr-3" aria-hidden="true" />
          <div>
            <h3 className="text-lg font-medium text-red-800">Erro ao carregar análises</h3>
            <p className="text-red-700 mt-1 text-base">{error}</p>
            <button 
              onClick={fetchAnalytics}
              className="mt-3 bg-red-700 text-white text-base font-medium px-6 py-3 rounded 
                         hover:bg-red-800 focus:outline-none focus:ring-4 focus:ring-red-300
                         transition-colors duration-200"
              aria-label="Tentar carregar as análises novamente"
            >
              Tentar novamente
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header do Dashboard */}
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Transparência em Números</h2>
          <div className="text-lg text-gray-800 mt-1">
            <span>Análises dos dados parlamentares para você acompanhar de perto</span>
            <Tooltip 
              content="Estes dados são extraídos diretamente da Câmara dos Deputados e atualizados diariamente para garantir transparência total."
              trigger="text"
            >
              <span className="ml-1">Saiba mais</span>
            </Tooltip>
          </div>
        </div>
        <button
          onClick={fetchAnalytics}
          className="bg-blue-700 text-white text-base font-medium px-4 py-2 rounded-md 
                     hover:bg-blue-800 focus:outline-none focus:ring-4 focus:ring-blue-300
                     flex items-center transition-colors duration-200"
          aria-label="Atualizar dados analíticos"
        >
          <RefreshCw className="h-4 w-4 mr-2" aria-hidden="true" />
          Atualizar
        </button>
      </div>

      {/* Cards de Insights Gerais */}
      {insights && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="bg-blue-50 p-4 rounded-lg border border-blue-200">
            <div className="flex items-center">
              <Users className="h-8 w-8 text-blue-600 mr-3" aria-hidden="true" />
              <div>
                <p className="text-blue-800 text-sm font-medium">Total de Deputados</p>
                <p className="text-2xl font-bold text-blue-900">{insights.total_deputados}</p>
              </div>
            </div>
          </div>

          <div className="bg-green-50 p-4 rounded-lg border border-green-200">
            <div className="flex items-center">
              <DollarSign className="h-8 w-8 text-green-600 mr-3" aria-hidden="true" />
              <div>
                <span className="text-green-800 text-sm font-medium">
                  Gastos Totais
                  <Tooltip content="Soma de todos os gastos parlamentares registrados no período" />
                </span>
                <p className="text-2xl font-bold text-green-900">{formatCurrency(insights.total_gasto_ano)}</p>
              </div>
            </div>
          </div>

          <div className="bg-purple-50 p-4 rounded-lg border border-purple-200">
            <div className="flex items-center">
              <TrendingUp className="h-8 w-8 text-purple-600 mr-3" aria-hidden="true" />
              <div>
                <span className="text-purple-800 text-sm font-medium">
                  Proposições no Ano
                  <Tooltip content="Número de proposições (leis, projetos) em tramitação na Câmara" />
                </span>
                <p className="text-2xl font-bold text-purple-900">{insights.total_proposicoes_ano}</p>
              </div>
            </div>
          </div>

          <div className="bg-orange-50 p-4 rounded-lg border border-orange-200">
            <div className="flex items-center">
              <Award className="h-8 w-8 text-orange-600 mr-3" aria-hidden="true" />
              <div>
                <span className="text-orange-800 text-sm font-medium">
                  Média de Gastos por Deputado
                  <Tooltip content="Valor médio gasto por cada deputado federal" />
                </span>
                <p className="text-2xl font-bold text-orange-900">{formatCurrency(insights.media_gastos_deputado)}</p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Seletor de Rankings */}
      <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
        <div className="mb-4">
          <h3 className="text-xl font-semibold text-gray-900 mb-2">Rankings dos Deputados</h3>
          <div className="flex space-x-2" role="tablist">
            <button
              onClick={() => setSelectedRanking('gastos')}
              className={`px-4 py-2 rounded-md text-base font-medium transition-colors duration-200 
                         focus:outline-none focus:ring-4 focus:ring-blue-300 ${
                selectedRanking === 'gastos'
                  ? 'bg-blue-700 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
              role="tab"
              aria-selected={selectedRanking === 'gastos'}
              aria-label="Ver ranking por gastos parlamentares"
            >
              💰 Maiores Gastos
            </button>
            <button
              onClick={() => setSelectedRanking('proposicoes')}
              className={`px-4 py-2 rounded-md text-base font-medium transition-colors duration-200 
                         focus:outline-none focus:ring-4 focus:ring-blue-300 ${
                selectedRanking === 'proposicoes'
                  ? 'bg-blue-700 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
              role="tab"
              aria-selected={selectedRanking === 'proposicoes'}
              aria-label="Ver ranking por proposições apresentadas"
            >
              📋 Mais Proposições
            </button>
            <button
              onClick={() => setSelectedRanking('presenca')}
              className={`px-4 py-2 rounded-md text-base font-medium transition-colors duration-200 
                         focus:outline-none focus:ring-4 focus:ring-blue-300 ${
                selectedRanking === 'presenca'
                  ? 'bg-blue-700 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
              role="tab"
              aria-selected={selectedRanking === 'presenca'}
              aria-label="Ver ranking por presença nas votações"
            >
              ✅ Maior Presença
            </button>
          </div>
        </div>

        {/* Lista do Ranking Selecionado */}
        <div role="tabpanel">
          {rankings[selectedRanking].length > 0 ? (
            <div className="space-y-3">
              {rankings[selectedRanking].map((item, index) => (
                <div key={item.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-md">
                  <div className="flex items-center">
                    <span className="text-lg font-bold text-gray-500 w-8">#{index + 1}</span>
                    <div className="ml-3">
                      <h4 className="font-semibold text-gray-900 text-base">{item.nome}</h4>
                      <p className="text-gray-700 text-sm">{item.partido} - {item.uf}</p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="font-bold text-lg text-gray-900">
                      {selectedRanking === 'gastos' ? formatCurrency(item.total_gasto || 0) : 
                       selectedRanking === 'presenca' ? formatPercentage(item.percentual_presenca || 0) : 
                       `${item.total_proposicoes || 0} proposições`}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8">
              <TrendingUp className="h-16 w-16 text-gray-400 mx-auto mb-4" aria-hidden="true" />
              <p className="text-gray-800 text-lg">Nenhum dado disponível para este ranking</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}