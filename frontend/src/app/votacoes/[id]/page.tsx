'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import VotacaoAnalysis from '@/components/VotacaoAnalysis';

export default function VotacaoPage() {
  const params = useParams();
  const router = useRouter();
  const [votacaoId, setVotacaoId] = useState<number | null>(null);

  useEffect(() => {
    if (params?.id) {
      const id = parseInt(params.id as string, 10);
      if (isNaN(id)) {
        router.push('/votacoes');
        return;
      }
      setVotacaoId(id);
    }
  }, [params?.id, router]);

  if (votacaoId === null) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Carregando votação...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header da página */}
      <div className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center gap-4">
            <button
              onClick={() => router.back()}
              className="flex items-center gap-2 text-gray-600 hover:text-gray-900 transition-colors"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
              </svg>
              Voltar
            </button>
            
            <div className="flex-1">
              <h1 className="text-xl font-semibold text-gray-900">
                Análise de Votação #{votacaoId}
              </h1>
              <p className="text-sm text-gray-600 mt-1">
                Veja como os deputados e partidos votaram nesta proposição
              </p>
            </div>

            <div className="flex gap-2">
              <button
                onClick={() => window.print()}
                className="px-3 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50 transition-colors"
              >
                🖨️ Imprimir
              </button>
              
              <button
                onClick={() => {
                  if (navigator.share) {
                    navigator.share({
                      title: `Análise de Votação #${votacaoId}`,
                      text: 'Veja como os deputados votaram nesta proposição',
                      url: window.location.href,
                    });
                  } else {
                    navigator.clipboard.writeText(window.location.href);
                    alert('Link copiado para a área de transferência!');
                  }
                }}
                className="px-3 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
              >
                📤 Compartilhar
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Conteúdo principal */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <VotacaoAnalysis 
          votacaoId={votacaoId}
          className="mb-6"
        />
        
        {/* Links relacionados */}
        <div className="bg-white rounded-lg shadow-sm border p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">
            🔗 Links Úteis
          </h3>
          
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <a
              href={`/votacoes`}
              className="p-4 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors block"
            >
              <div className="font-medium text-gray-900 mb-1">📋 Todas as Votações</div>
              <div className="text-sm text-gray-600">
                Explore outras votações da Câmara dos Deputados
              </div>
            </a>

            <a
              href={`/deputados`}
              className="p-4 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors block"
            >
              <div className="font-medium text-gray-900 mb-1">👥 Deputados</div>
              <div className="text-sm text-gray-600">
                Veja o perfil e histórico dos deputados
              </div>
            </a>

            <a
              href={`/analytics`}
              className="p-4 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors block"
            >
              <div className="font-medium text-gray-900 mb-1">📊 Analytics</div>
              <div className="text-sm text-gray-600">
                Análises avançadas e estatísticas
              </div>
            </a>
          </div>
        </div>

        {/* Informações sobre transparência */}
        <div className="mt-6 bg-blue-50 border border-blue-200 rounded-lg p-6">
          <h3 className="text-lg font-semibold text-blue-900 mb-3">
            🔍 Sobre a Transparência Política
          </h3>
          
          <div className="text-sm text-blue-800 space-y-2">
            <p>
              <strong>Fonte dos Dados:</strong> Os dados das votações são obtidos diretamente da 
              API oficial da Câmara dos Deputados, garantindo veracidade e atualização.
            </p>
            
            <p>
              <strong>Orientação Partidária:</strong> A orientação representa a posição oficial 
              do partido comunicada aos seus deputados antes da votação.
            </p>
            
            <p>
              <strong>Disciplina Partidária:</strong> Percentual de deputados que seguiram a 
              orientação oficial do seu partido na votação.
            </p>
            
            <p>
              <strong>Missão do Projeto:</strong> Democratizar o acesso à informação política 
              e promover maior engajamento cidadão nas decisões públicas.
            </p>
          </div>
          
          <div className="mt-4 pt-4 border-t border-blue-200">
            <div className="flex flex-col sm:flex-row gap-3">
              <a
                href="https://dadosabertos.camara.leg.br/swagger/api.html"
                target="_blank"
                rel="noopener noreferrer"
                className="text-blue-700 hover:text-blue-900 text-sm underline"
              >
                📡 API Dados Abertos - Câmara
              </a>
              
              <a
                href="/sobre"
                className="text-blue-700 hover:text-blue-900 text-sm underline"
              >
                ℹ️ Sobre o Projeto Tô De Olho
              </a>
              
              <a
                href="/contato"
                className="text-blue-700 hover:text-blue-900 text-sm underline"
              >
                📧 Entre em Contato
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}