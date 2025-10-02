'use client';

import VotacoesPage from '@/components/VotacoesPage';

export default function Votacoes() {
  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header principal */}
      <div className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="text-center">
            <h1 className="text-3xl font-bold text-gray-900 mb-2">
              🗳️ Transparência nas Votações
            </h1>
            <p className="text-lg text-gray-600 max-w-3xl mx-auto">
              Acompanhe como os deputados federais votaram nas principais proposições da Câmara. 
              Dados atualizados diretamente da API oficial para máxima transparência.
            </p>
          </div>
        </div>
      </div>

      {/* Estatísticas rápidas */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <div className="bg-white rounded-lg shadow-sm border p-6 text-center">
            <div className="text-2xl font-bold text-green-600 mb-1">1.247</div>
            <div className="text-sm text-gray-600">Votações Sincronizadas</div>
          </div>
          
          <div className="bg-white rounded-lg shadow-sm border p-6 text-center">
            <div className="text-2xl font-bold text-blue-600 mb-1">67%</div>
            <div className="text-sm text-gray-600">Taxa de Aprovação</div>
          </div>
          
          <div className="bg-white rounded-lg shadow-sm border p-6 text-center">
            <div className="text-2xl font-bold text-purple-600 mb-1">513</div>
            <div className="text-sm text-gray-600">Deputados Ativos</div>
          </div>
          
          <div className="bg-white rounded-lg shadow-sm border p-6 text-center">
            <div className="text-2xl font-bold text-orange-600 mb-1">28</div>
            <div className="text-sm text-gray-600">Partidos Políticos</div>
          </div>
        </div>

        {/* Componente principal de votações */}
        <VotacoesPage />
      </div>
    </div>
  );
}