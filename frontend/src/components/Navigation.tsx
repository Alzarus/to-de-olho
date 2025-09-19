'use client';

import { useState } from 'react';
import { Users, FileText, BarChart3, Menu, X } from 'lucide-react';

interface NavigationProps {
  currentSection: string;
  onSectionChange: (section: string) => void;
}

const navigationItems = [
  { id: 'dashboard', label: 'Dashboard', icon: BarChart3, description: 'Visão geral dos dados' },
  { id: 'deputados', label: 'Deputados', icon: Users, description: 'Lista e perfis dos deputados' },
  { id: 'proposicoes', label: 'Proposições', icon: FileText, description: 'Leis e projetos em tramitação' },
  { id: 'analytics', label: 'Análises', icon: BarChart3, description: 'Relatórios e estatísticas' },
];

export default function Navigation({ currentSection, onSectionChange }: NavigationProps) {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  return (
    <>
      {/* Desktop Navigation */}
      <nav className="hidden md:flex bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4">
          <div className="flex space-x-8">
            {navigationItems.map((item) => {
              const Icon = item.icon;
              const isActive = currentSection === item.id;
              
              return (
                <button
                  key={item.id}
                  onClick={() => onSectionChange(item.id)}
                  className={`flex items-center px-3 py-4 border-b-2 text-sm font-medium transition-colors duration-200 ${
                    isActive
                      ? 'border-blue-500 text-blue-600'
                      : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                  }`}
                  aria-current={isActive ? 'page' : undefined}
                >
                  <Icon className="h-4 w-4 mr-2" aria-hidden="true" />
                  {item.label}
                </button>
              );
            })}
          </div>
        </div>
      </nav>

      {/* Mobile Navigation */}
      <nav className="md:hidden bg-white shadow-sm border-b border-gray-200">
        <div className="px-4 py-3">
          <div className="flex items-center justify-center">
            <button
              onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
              className="text-gray-500 hover:text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
              aria-label="Abrir menu de navegação"
            >
              {isMobileMenuOpen ? (
                <X className="h-6 w-6" />
              ) : (
                <Menu className="h-6 w-6" />
              )}
            </button>
          </div>
          
          {/* Mobile Menu Dropdown */}
          {isMobileMenuOpen && (
            <div className="mt-3 space-y-1">
              {navigationItems.map((item) => {
                const Icon = item.icon;
                const isActive = currentSection === item.id;
                
                return (
                  <button
                    key={item.id}
                    onClick={() => {
                      onSectionChange(item.id);
                      setIsMobileMenuOpen(false);
                    }}
                    className={`w-full flex items-center px-3 py-2 rounded-md text-sm font-medium transition-colors duration-200 ${
                      isActive
                        ? 'bg-blue-50 text-blue-700'
                        : 'text-gray-600 hover:text-gray-900 hover:bg-gray-50'
                    }`}
                  >
                    <Icon className="h-4 w-4 mr-3" aria-hidden="true" />
                    <div className="text-left">
                      <div>{item.label}</div>
                      <div className="text-xs text-gray-500">{item.description}</div>
                    </div>
                  </button>
                );
              })}
            </div>
          )}
        </div>
      </nav>
    </>
  );
}