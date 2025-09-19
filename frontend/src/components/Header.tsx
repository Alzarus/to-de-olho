'use client';

import { Eye, Github, Shield } from 'lucide-react';
import Tooltip from './Tooltip';

export default function Header() {
  return (
    <header className="bg-white shadow-sm border-b border-gray-200">
      <div className="max-w-7xl mx-auto px-4 py-4">
        <div className="flex items-center justify-between">
          {/* Logo e Título */}
          <div className="flex items-center">
            <div className="bg-blue-700 p-2 rounded-lg mr-3">
              <Eye className="h-6 w-6 text-white" aria-hidden="true" />
            </div>
            <div>
              <h1 className="text-2xl font-bold text-gray-900">Tô De Olho</h1>
              <p className="text-sm text-gray-600">Transparência Política para Todos</p>
            </div>
          </div>

          {/* Info do Projeto */}
          <div className="flex items-center space-x-4">
            <div className="flex items-center text-sm text-gray-600">
              <Shield className="h-4 w-4 mr-1" aria-hidden="true" />
              <span>Dados Oficiais da Câmara</span>
              <Tooltip 
                content="Todos os dados são extraídos diretamente da API oficial da Câmara dos Deputados, garantindo autenticidade e atualização."
              />
            </div>
            
            <a
              href="https://github.com/Alzarus/to-de-olho"
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center text-sm text-gray-600 hover:text-blue-600 transition-colors
                         focus:outline-none focus:ring-2 focus:ring-blue-300 rounded p-1"
              aria-label="Ver código fonte no GitHub"
            >
              <Github className="h-4 w-4 mr-1" aria-hidden="true" />
              <span>Código Aberto</span>
            </a>
          </div>
        </div>
      </div>
    </header>
  );
}