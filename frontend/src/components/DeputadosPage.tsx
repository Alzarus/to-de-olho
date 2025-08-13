'use client';

import { useState, useEffect } from 'react';
import axios from 'axios';
import { Search, User, MapPin, Building2, Euro, AlertCircle } from 'lucide-react';

interface Deputado {
  id: number;
  nome: string;
  siglaPartido: string;
  siglaUf: string;
  urlFoto: string;
  condicaoEleitoral: string;
  email: string;
}

interface APIResponse {
  data: Deputado[];
  total: number;
  source: string;
}

const API_BASE_URL = 'http://localhost:8080/api/v1';

export default function DeputadosPage() {
  const [deputados, setDeputados] = useState<Deputado[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedUF, setSelectedUF] = useState('');
  const [selectedPartido, setSelectedPartido] = useState('');
  const [selectedDeputado, setSelectedDeputado] = useState<Deputado | null>(null);

  const estados = [
    'AC', 'AL', 'AP', 'AM', 'BA', 'CE', 'DF', 'ES', 'GO', 'MA', 
    'MT', 'MS', 'MG', 'PA', 'PB', 'PR', 'PE', 'PI', 'RJ', 'RN', 
    'RS', 'RO', 'RR', 'SC', 'SP', 'SE', 'TO'
  ];

  const partidos = [
    'PT', 'PL', 'UNIÃO', 'PP', 'MDB', 'PSD', 'REPUBLICANOS', 'PSDB', 
    'PDT', 'PODE', 'PSOL', 'PSB', 'CIDADANIA', 'AVANTE', 'SOLIDARIEDADE'
  ];

  useEffect(() => {
    fetchDeputados();
  }, [selectedUF, selectedPartido]);

  const fetchDeputados = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const params = new URLSearchParams();
      if (selectedUF) params.append('uf', selectedUF);
      if (selectedPartido) params.append('partido', selectedPartido);
      if (searchTerm) params.append('nome', searchTerm);

      const response = await axios.get<APIResponse>(`${API_BASE_URL}/deputados?${params}`);
      setDeputados(response.data.data || []);
    } catch (err) {
      console.error('Erro ao buscar deputados:', err);
      setError('Erro ao carregar deputados. Verifique se o backend está rodando.');
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = () => {
    fetchDeputados();
  };

  const filteredDeputados = deputados.filter(deputado =>
    deputado.nome.toLowerCase().includes(searchTerm.toLowerCase())
  );

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-8">
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
          <span className="ml-3 text-gray-600">Carregando deputados...</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-8">
        <div className="bg-red-50 border border-red-200 rounded-lg p-6">
          <div className="flex items-center">
            <AlertCircle className="h-6 w-6 text-red-600 mr-3" />
            <div>
              <h3 className="text-lg font-medium text-red-800">Erro ao carregar dados</h3>
              <p className="text-red-600 mt-1">{error}</p>
              <button 
                onClick={fetchDeputados}
                className="mt-3 bg-red-600 text-white px-4 py-2 rounded hover:bg-red-700"
              >
                Tentar novamente
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-4xl font-bold text-gray-900 mb-2">
          Deputados Federais
        </h1>
        <p className="text-gray-600">
          Explore os {filteredDeputados.length} deputados federais e seus dados de transparência
        </p>
      </div>

      {/* Filtros */}
      <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-200 mb-8">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              <Search className="inline h-4 w-4 mr-1" />
              Buscar por nome
            </label>
            <input
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Digite o nome do deputado..."
              className="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              <MapPin className="inline h-4 w-4 mr-1" />
              Estado (UF)
            </label>
            <select
              value={selectedUF}
              onChange={(e) => setSelectedUF(e.target.value)}
              className="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">Todos os estados</option>
              {estados.map(uf => (
                <option key={uf} value={uf}>{uf}</option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              <Building2 className="inline h-4 w-4 mr-1" />
              Partido
            </label>
            <select
              value={selectedPartido}
              onChange={(e) => setSelectedPartido(e.target.value)}
              className="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">Todos os partidos</option>
              {partidos.map(partido => (
                <option key={partido} value={partido}>{partido}</option>
              ))}
            </select>
          </div>

          <div className="flex items-end">
            <button
              onClick={handleSearch}
              className="w-full bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              Buscar
            </button>
          </div>
        </div>
      </div>

      {/* Lista de Deputados */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredDeputados.map((deputado) => (
          <div
            key={deputado.id}
            className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-md transition-shadow cursor-pointer"
            onClick={() => setSelectedDeputado(deputado)}
          >
            <div className="flex items-center mb-4">
              {deputado.urlFoto ? (
                <img
                  src={deputado.urlFoto}
                  alt={deputado.nome}
                  className="w-16 h-16 rounded-full object-cover mr-4"
                  onError={(e) => {
                    e.currentTarget.style.display = 'none';
                    e.currentTarget.nextElementSibling?.classList.remove('hidden');
                  }}
                />
              ) : null}
              <div className="w-16 h-16 bg-gray-200 rounded-full flex items-center justify-center mr-4 hidden">
                <User className="h-8 w-8 text-gray-400" />
              </div>
              <div>
                <h3 className="font-semibold text-gray-900 text-sm leading-tight">
                  {deputado.nome}
                </h3>
                <div className="flex items-center text-sm text-gray-600 mt-1">
                  <Building2 className="h-3 w-3 mr-1" />
                  {deputado.siglaPartido}
                  <MapPin className="h-3 w-3 ml-2 mr-1" />
                  {deputado.siglaUf}
                </div>
              </div>
            </div>
            
            <div className="text-sm text-gray-600">
              <p><strong>Situação:</strong> {deputado.condicaoEleitoral}</p>
              {deputado.email && (
                <p className="truncate"><strong>Email:</strong> {deputado.email}</p>
              )}
            </div>
            
            <button className="mt-4 w-full bg-blue-50 text-blue-700 py-2 px-4 rounded-md hover:bg-blue-100 transition-colors flex items-center justify-center">
              <Euro className="h-4 w-4 mr-2" />
              Ver Despesas
            </button>
          </div>
        ))}
      </div>

      {filteredDeputados.length === 0 && !loading && (
        <div className="text-center py-12">
          <User className="h-16 w-16 text-gray-300 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">
            Nenhum deputado encontrado
          </h3>
          <p className="text-gray-600">
            Tente ajustar os filtros de busca
          </p>
        </div>
      )}

      {/* Modal de Detalhes do Deputado */}
      {selectedDeputado && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg max-w-md w-full p-6">
            <div className="flex items-center mb-4">
              {selectedDeputado.urlFoto ? (
                <img
                  src={selectedDeputado.urlFoto}
                  alt={selectedDeputado.nome}
                  className="w-20 h-20 rounded-full object-cover mr-4"
                />
              ) : (
                <div className="w-20 h-20 bg-gray-200 rounded-full flex items-center justify-center mr-4">
                  <User className="h-10 w-10 text-gray-400" />
                </div>
              )}
              <div>
                <h2 className="text-xl font-semibold text-gray-900">
                  {selectedDeputado.nome}
                </h2>
                <p className="text-gray-600">
                  {selectedDeputado.siglaPartido} - {selectedDeputado.siglaUf}
                </p>
              </div>
            </div>
            
            <div className="space-y-3 mb-6">
              <p><strong>ID:</strong> {selectedDeputado.id}</p>
              <p><strong>Situação:</strong> {selectedDeputado.condicaoEleitoral}</p>
              {selectedDeputado.email && (
                <p><strong>Email:</strong> {selectedDeputado.email}</p>
              )}
            </div>

            <div className="flex space-x-3">
              <button 
                onClick={() => setSelectedDeputado(null)}
                className="flex-1 bg-gray-200 text-gray-800 py-2 px-4 rounded-md hover:bg-gray-300"
              >
                Fechar
              </button>
              <button className="flex-1 bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 flex items-center justify-center">
                <Euro className="h-4 w-4 mr-2" />
                Ver Despesas
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
