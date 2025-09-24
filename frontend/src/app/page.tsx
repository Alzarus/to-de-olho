'use client';

import { useState } from 'react';
import DeputadosPage from '@/components/DeputadosPage'
import DashboardAnalytics from '@/components/DashboardAnalytics'
import Header from '@/components/Header'
import Navigation from '@/components/Navigation'

export default function Home() {
  const [currentSection, setCurrentSection] = useState('dashboard');

  const renderCurrentSection = () => {
    switch (currentSection) {
      case 'dashboard':
        return <DashboardAnalytics />;
      case 'deputados':
        return <DeputadosPage />;
      case 'proposicoes':
        return (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-8 text-center">
            <h2 className="text-2xl font-bold text-gray-900 mb-4">Proposições</h2>
            <p className="text-gray-600">Seção em desenvolvimento. Em breve você poderá acompanhar projetos de lei e proposições.</p>
          </div>
        );
      case 'analytics':
        return (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-8 text-center">
            <h2 className="text-2xl font-bold text-gray-900 mb-4">Análises Avançadas</h2>
            <p className="text-gray-600">Seção em desenvolvimento. Relatórios detalhados e análises estatísticas estarão disponíveis em breve.</p>
          </div>
        );
      default:
        return <DashboardAnalytics />;
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <Header />
      <Navigation currentSection={currentSection} onSectionChange={setCurrentSection} />
      
      <main className="max-w-7xl mx-auto px-4 py-8">
        {renderCurrentSection()}
      </main>
    </div>
  )
}
