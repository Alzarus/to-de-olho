'use client';

import VotacoesAnalytics from '@/components/VotacoesAnalytics';
import VotacoesPage from '@/components/VotacoesPage';
import VotacoesRanking from '@/components/VotacoesRanking';

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

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 space-y-8">
        <VotacoesAnalytics />
        <VotacoesRanking limite={8} />
        <VotacoesPage />
      </div>
    </div>
  );
}
