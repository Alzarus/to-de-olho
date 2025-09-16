'use client';

import { useState, useEffect } from 'react';
import axios from 'axios';
import { Search, User, MapPin, Building2, Euro, AlertCircle, X } from 'lucide-react';
import DeputadoCard from './DeputadoCard';

export interface Deputado {
  id: number;
  nome: string;
  siglaPartido: string;
  siglaUf: string;
  urlFoto: string;
  condicaoEleitoral: string;
  email: string;
}

export interface Despesa {
  id?: number;
  ano: number;
  mes: number;
  tipoDespesa: string;
  codDocumento: number;
  tipoDocumento: string;
  codTipoDocumento: number;
  dataDocumento: string;
  numDocumento: string;
  valorDocumento: number;
  urlDocumento: string;
  nomeFornecedor: string;
  cnpjCpfFornecedor: string;
  valorLiquido: number;
  valorBruto: number;
  valorGlosa: number;
  numRessarcimento?: string;
  codLote: number;
  parcela?: number;
}

interface APIResponse {
  data: Deputado[];
  total: number;
  source: string;
}

interface DespesasAPIResponse {
  data: Despesa[];
  total: number;
  source: string;
}

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

export default function DeputadosPage() {
  const [deputados, setDeputados] = useState<Deputado[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedUF, setSelectedUF] = useState('');
  const [selectedPartido, setSelectedPartido] = useState('');
  const [selectedDeputado, setSelectedDeputado] = useState<Deputado | null>(null);
  const [showDespesas, setShowDespesas] = useState(false);
  const [despesas, setDespesas] = useState<Despesa[]>([]);
  const [loadingDespesas, setLoadingDespesas] = useState(false);
  const [despesasError, setDespesasError] = useState<string | null>(null);
  const [selectedAno, setSelectedAno] = useState(new Date().getFullYear());

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

  const fetchDespesas = async (deputadoId: number, ano: number) => {
    setLoadingDespesas(true);
    setDespesasError(null);
    
    try {
      const url = `${API_BASE_URL}/deputados/${deputadoId}/despesas?ano=${ano}`;
      const response = await axios.get<DespesasAPIResponse>(url);
      setDespesas(response.data.data || []);
    } catch (err) {
      console.error('Erro ao buscar despesas:', err);
      setDespesasError('Erro ao carregar despesas. Tente novamente.');
    } finally {
      setLoadingDespesas(false);
    }
  };

  const handleVerDespesas = (deputado: Deputado) => {
    setSelectedDeputado(deputado);
    setShowDespesas(true);
    fetchDespesas(deputado.id, selectedAno);
  };

  const closeDespesasModal = () => {
    setShowDespesas(false);
    setDespesas([]);
    setDespesasError(null);
  };

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL'
    }).format(value);
  };

  const formatDate = (dateString: string) => {
    try {
      const date = new Date(dateString);
      return date.toLocaleDateString('pt-BR');
    } catch {
      return dateString;
    }
  };

  const handleSearch = () => {
    fetchDeputados();
  };

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
          Explore os {deputados.length} deputados federais e seus dados de transparência
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
        {deputados.map((deputado) => (
          <DeputadoCard
            key={deputado.id}
            deputado={deputado}
            onClick={() => setSelectedDeputado(deputado)}
            onVerDespesas={() => handleVerDespesas(deputado)}
          />
        ))}
      </div>

      {deputados.length === 0 && !loading && (
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
              <button 
                onClick={() => handleVerDespesas(selectedDeputado)}
                className="flex-1 bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 flex items-center justify-center"
              >
                <Euro className="h-4 w-4 mr-2" />
                Ver Despesas
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Modal de Despesas */}
      {showDespesas && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg max-w-4xl w-full max-h-[80vh] overflow-hidden">
            <div className="p-6 border-b border-gray-200">
              <div className="flex items-center justify-between">
                <h2 className="text-xl font-semibold text-gray-900">
                  Despesas do Deputado
                </h2>
                <button 
                  onClick={() => setShowDespesas(false)}
                  className="text-gray-400 hover:text-gray-600"
                >
                  <X className="h-6 w-6" />
                </button>
              </div>
              
              {/* Seletor de Ano */}
              <div className="mt-4">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Ano:
                </label>
                <select
                  value={selectedAno}
                  onChange={(e) => setSelectedAno(Number(e.target.value))}
                  className="border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  {Array.from({ length: 5 }, (_, i) => new Date().getFullYear() - i).map(year => (
                    <option key={year} value={year}>{year}</option>
                  ))}
                </select>
              </div>
            </div>

            <div className="p-6 overflow-y-auto max-h-[60vh]">
              {loadingDespesas ? (
                <div className="flex items-center justify-center py-8">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
                  <span className="ml-2 text-gray-600">Carregando despesas...</span>
                </div>
              ) : despesasError ? (
                <div className="text-center py-8">
                  <p className="text-red-600">{despesasError}</p>
                  <button 
                    onClick={() => window.location.reload()}
                    className="mt-2 text-blue-600 hover:text-blue-800"
                  >
                    Tentar novamente
                  </button>
                </div>
              ) : despesas.length === 0 ? (
                <div className="text-center py-8">
                  <Euro className="h-16 w-16 text-gray-300 mx-auto mb-4" />
                  <p className="text-gray-600">Nenhuma despesa encontrada para este período.</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {despesas.map((despesa, index) => (
                    <div key={index} className="border border-gray-200 rounded-lg p-4">
                      <div className="flex justify-between items-start mb-2">
                        <h3 className="font-medium text-gray-900">{despesa.tipoDespesa}</h3>
                        <span className="text-lg font-semibold text-green-600">
                          {formatCurrency(despesa.valorLiquido)}
                        </span>
                      </div>
                      
                      {despesa.nomeFornecedor && (
                        <p className="text-sm text-gray-600 mb-1">
                          <strong>Fornecedor:</strong> {despesa.nomeFornecedor}
                        </p>
                      )}
                      
                      {despesa.cnpjCpfFornecedor && (
                        <p className="text-sm text-gray-600 mb-1">
                          <strong>CNPJ/CPF:</strong> {despesa.cnpjCpfFornecedor}
                        </p>
                      )}
                      
                      <div className="flex justify-between items-center text-xs text-gray-500 mt-2">
                        <span>{formatDate(despesa.dataDocumento)}</span>
                        {despesa.numDocumento && (
                          <span>Doc: {despesa.numDocumento}</span>
                        )}
                      </div>
                      
                      {despesa.valorBruto !== despesa.valorLiquido && (
                        <div className="text-xs text-gray-500 mt-1">
                          <span>Valor bruto: {formatCurrency(despesa.valorBruto)}</span>
                          {despesa.valorGlosa > 0 && (
                            <span className="ml-2">Glosa: {formatCurrency(despesa.valorGlosa)}</span>
                          )}
                        </div>
                      )}
                    </div>
                  ))}
                  
                  {/* Resumo Total */}
                  <div className="border-t border-gray-200 pt-4 mt-6">
                    <div className="flex justify-between items-center">
                      <span className="text-lg font-semibold text-gray-900">Total:</span>
                      <span className="text-xl font-bold text-green-600">
                        {formatCurrency(despesas.reduce((sum, d) => sum + d.valorLiquido, 0))}
                      </span>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
