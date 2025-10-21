'use client';

import React, { useState, useEffect, useCallback, useRef } from 'react';
import axios from 'axios';
import { Search, User, MapPin, Building2, Euro, AlertCircle, X } from 'lucide-react';
import DeputadoCard from './DeputadoCard';
import Tooltip from './Tooltip';
import { API_CONFIG, PAGINATION_CONFIG, TIMING_CONFIG, BRASIL_CONFIG, ANALYTICS_CONFIG } from '../config/constants';

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

interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

interface APIResponse {
  data?: {
    data?: Deputado[] | null;
    pagination?: PaginationInfo;
  };
  meta?: {
    cache_hit: boolean;
    process_time: number;
    timestamp: number;
  };
}

interface DespesasAPIResponse {
  ano?: string;
  data?: Despesa[] | null;
  source?: string;
  total?: number;
  valor_total?: number;
}

export default function DeputadosPage() {
  const [deputados, setDeputados] = useState<Deputado[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [debouncedSearchTerm, setDebouncedSearchTerm] = useState('');
  const [selectedUF, setSelectedUF] = useState('');
  const [selectedPartido, setSelectedPartido] = useState('');
  const [selectedDeputado, setSelectedDeputado] = useState<Deputado | null>(null);
  const [showDespesas, setShowDespesas] = useState(false);
  const [despesas, setDespesas] = useState<Despesa[]>([]);
  const [loadingDespesas, setLoadingDespesas] = useState(false);
  const [despesasError, setDespesasError] = useState<string | null>(null);
  const [selectedAno, setSelectedAno] = useState(new Date().getFullYear());
  
  // Ref para controle de debouncing
  const debounceTimer = useRef<NodeJS.Timeout | null>(null);
  
  // Estados para paginação
  const [pagination, setPagination] = useState<PaginationInfo>({
    page: 1,
    limit: PAGINATION_CONFIG.DEFAULT_PAGE_SIZE,
    total: 0,
    total_pages: 1,
    has_next: false,
    has_prev: false
  });
  const [currentPage, setCurrentPage] = useState(1);

  const estados = BRASIL_CONFIG.ESTADOS;
  
  const partidos = BRASIL_CONFIG.PARTIDOS;

  useEffect(() => {
    fetchDeputados();
  }, [currentPage]);

  // Debounce effect para o searchTerm
  useEffect(() => {
    if (debounceTimer.current) {
      clearTimeout(debounceTimer.current);
    }
    
    debounceTimer.current = setTimeout(() => {
      setDebouncedSearchTerm(searchTerm);
    }, TIMING_CONFIG.SEARCH_DEBOUNCE_MS); // Debounce configurável via env

    return () => {
      if (debounceTimer.current) {
        clearTimeout(debounceTimer.current);
      }
    };
  }, [searchTerm]);

  const fetchDeputados = useCallback(async () => {
    setLoading(true);
    setError(null);
    
    try {
      const params = new URLSearchParams();
      if (selectedUF) params.append('uf', selectedUF);
      if (selectedPartido) params.append('partido', selectedPartido);
      if (debouncedSearchTerm) params.append('nome', debouncedSearchTerm);
      params.append('page', currentPage.toString());
      params.append('limit', PAGINATION_CONFIG.DEFAULT_PAGE_SIZE.toString());

      const response = await axios.get<APIResponse>(`${API_CONFIG.BASE_URL}/deputados?${params}`);
      
      const deputadosData = response.data?.data?.data;
      const paginationData = response.data?.data?.pagination;
      
      if (Array.isArray(deputadosData)) {
        setDeputados(deputadosData);
      } else {
        console.warn('Sistema Ultra-Performance: Dados de deputados não são um array');
        setDeputados([]);
      }

      if (paginationData) {
        setPagination({
          page: paginationData.page || currentPage,
          limit: paginationData.limit || PAGINATION_CONFIG.DEFAULT_PAGE_SIZE,
          total: paginationData.total || 0,
          total_pages: paginationData.total_pages || 1,
          has_next: paginationData.has_next || false,
          has_prev: paginationData.has_prev || false
        });
      }
    } catch (err) {
      console.error('Erro ao buscar deputados:', err);
      setError('Erro ao carregar deputados. Verifique se o backend está rodando.');
    } finally {
      setLoading(false);
    }
  }, [selectedUF, selectedPartido, debouncedSearchTerm, currentPage]);

  useEffect(() => {
    fetchDeputados();
  }, [fetchDeputados]);

  const handleSearch = () => {
    setCurrentPage(1);
    fetchDeputados();
  };

  const handlePageChange = (newPage: number) => {
    if (newPage >= 1 && newPage <= pagination.total_pages) {
      setCurrentPage(newPage);
    }
  };

  const handleFilterChange = () => {
    setCurrentPage(1);
  };

  const handleVerDespesas = (deputado: Deputado) => {
    setSelectedDeputado(deputado);
    setShowDespesas(true);
    fetchDespesas(deputado.id, selectedAno);
  };

  const fetchDespesas = async (deputadoId: number, ano: number) => {
    setLoadingDespesas(true);
    setDespesasError(null);
    
    try {
      const url = `${API_CONFIG.BASE_URL}/deputados/${deputadoId}/despesas?ano=${ano}`;
      const response = await axios.get<DespesasAPIResponse>(url);
      
      const despesasData = response.data?.data;
      if (Array.isArray(despesasData)) {
        setDespesas(despesasData);
      } else {
        setDespesas([]);
      }
    } catch (err) {
      console.error('Erro ao buscar despesas:', err);
      setDespesasError('Erro ao carregar despesas do deputado');
    } finally {
      setLoadingDespesas(false);
    }
  };

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL'
    }).format(value);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('pt-BR');
  };

  if (loading) {
    return (
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <div className="flex items-center justify-center py-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-700"></div>
          <span className="ml-2 text-lg text-gray-900">Carregando deputados...</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <div className="text-center py-8">
          <AlertCircle className="h-16 w-16 text-red-500 mx-auto mb-4" />
          <p className="text-red-700 text-lg">{error}</p>
          <button 
            onClick={fetchDeputados}
            className="mt-4 bg-blue-700 text-white px-4 py-2 rounded-md hover:bg-blue-800"
          >
            Tentar novamente
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Informações e filtros */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <div className="mb-6">
          <div className="text-lg text-gray-800">
            <span>
              Mostrando {deputados.length} de {pagination.total} deputados federais. 
              Acompanhe seus dados de transparência.
            </span>
            <Tooltip 
              content="Os deputados federais são eleitos para representar a população brasileira na Câmara dos Deputados, onde criam e votam leis que afetam todo o país."
              trigger="text"
            >
              <span className="ml-1">Saiba mais sobre deputados</span>
            </Tooltip>
          </div>
        </div>

        {/* Filtros */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="relative">
            <label htmlFor="search-input" className="block text-base font-medium text-gray-900 mb-2">
              <Search className="inline h-4 w-4 mr-1" />
              Buscar por nome
              {searchTerm !== debouncedSearchTerm && (
                <span className="ml-2 text-sm text-blue-600">⏳ Buscando...</span>
              )}
            </label>
            <input
              id="search-input"
              type="text"
              placeholder={`Digite o nome do deputado (aguarda ${TIMING_CONFIG.SEARCH_DEBOUNCE_MS}ms)`}
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full text-base border border-gray-300 rounded-md px-4 py-3 
                         focus:outline-none focus:ring-4 focus:ring-blue-300 focus:border-blue-500"
            />
            {searchTerm && (
              <button
                onClick={() => {
                  setSearchTerm('');
                  setDebouncedSearchTerm('');
                }}
                className="absolute right-3 top-11 text-gray-400 hover:text-gray-600"
                aria-label="Limpar busca"
              >
                <X className="h-4 w-4" />
              </button>
            )}
          </div>

          <div>
            <label htmlFor="uf-select" className="block text-base font-medium text-gray-900 mb-2">
              <MapPin className="inline h-4 w-4 mr-1" />
              Estado (UF)
            </label>
            <select
              id="uf-select"
              value={selectedUF}
              onChange={(e) => {
                setSelectedUF(e.target.value);
                handleFilterChange();
              }}
              className="w-full text-base border border-gray-300 rounded-md px-4 py-3 
                         focus:outline-none focus:ring-4 focus:ring-blue-300 focus:border-blue-500"
            >
              <option value="">Todos os estados</option>
              {estados.map(uf => (
                <option key={uf} value={uf}>{uf}</option>
              ))}
            </select>
          </div>

          <div>
            <label htmlFor="partido-input" className="block text-base font-medium text-gray-900 mb-2">
              <Building2 className="inline h-4 w-4 mr-1" />
              Partido Político
            </label>
            <input
              id="partido-input"
              list="partidos-list"
              value={selectedPartido}
              onChange={(e) => {
                setSelectedPartido(e.target.value);
                handleFilterChange();
              }}
              placeholder="Digite ou selecione um partido"
              className="w-full text-base border border-gray-300 rounded-md px-4 py-3 
                         focus:outline-none focus:ring-4 focus:ring-blue-300 focus:border-blue-500"
              aria-label="Filtrar por partido político"
            />
            <datalist id="partidos-list">
              <option value="">Todos os partidos</option>
              {partidos.map(partido => (
                <option key={partido} value={partido}>{partido}</option>
              ))}
            </datalist>
          </div>

          <div className="flex items-end">
            <button
              onClick={handleSearch}
              className="w-full bg-blue-700 text-white text-base font-medium px-6 py-3 rounded-md 
                         hover:bg-blue-800 focus:outline-none focus:ring-4 focus:ring-blue-300"
            >
              Buscar Deputados
            </button>
          </div>
        </div>
      </div>

      {/* Lista de Deputados */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {Array.isArray(deputados) && deputados.map((deputado) => (
          <DeputadoCard
            key={deputado.id}
            deputado={deputado}
            onClick={() => setSelectedDeputado(deputado)}
            onVerDespesas={() => handleVerDespesas(deputado)}
          />
        ))}
      </div>

      {/* Paginação */}
      {pagination.total_pages > 1 && (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
          <div className="flex items-center justify-between">
            <div className="text-sm text-gray-700">
              Página {pagination.page} de {pagination.total_pages} 
              ({pagination.total} deputados no total)
            </div>
            
            <div className="flex items-center space-x-2">
              <button
                onClick={() => handlePageChange(currentPage - 1)}
                disabled={!pagination.has_prev}
                className={`px-3 py-2 text-sm font-medium rounded-md ${
                  pagination.has_prev
                    ? 'text-gray-700 bg-white border border-gray-300 hover:bg-gray-50'
                    : 'text-gray-400 bg-gray-100 border border-gray-200 cursor-not-allowed'
                }`}
              >
                Anterior
              </button>

              <span className="text-sm text-gray-500">
                {pagination.page} / {pagination.total_pages}
              </span>

              <button
                onClick={() => handlePageChange(currentPage + 1)}
                disabled={!pagination.has_next}
                className={`px-3 py-2 text-sm font-medium rounded-md ${
                  pagination.has_next
                    ? 'text-gray-700 bg-white border border-gray-300 hover:bg-gray-50'
                    : 'text-gray-400 bg-gray-100 border border-gray-200 cursor-not-allowed'
                }`}
              >
                Próxima
              </button>
            </div>
          </div>
        </div>
      )}

      {deputados.length === 0 && !loading && (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-8 text-center">
          <User className="h-16 w-16 text-gray-400 mx-auto mb-4" />
          <h3 className="text-xl font-medium text-gray-900 mb-2">
            Nenhum deputado encontrado
          </h3>
          <p className="text-lg text-gray-800">
            Tente ajustar os filtros de busca ou verifique se digitou o nome corretamente
          </p>
        </div>
      )}

      {/* Modal de Detalhes do Deputado */}
      {selectedDeputado && (
        <div 
          className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50"
          onClick={() => {
            setSelectedDeputado(null);
            setDespesas([]);
            setDespesasError(null);
          }}
        >
          <div 
            className="bg-white rounded-lg max-w-md w-full p-6"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center mb-4">
              {selectedDeputado.urlFoto ? (
                <img
                  src={selectedDeputado.urlFoto}
                  alt={`Foto de ${selectedDeputado.nome}`}
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
                <p className="text-lg text-gray-800">
                  {selectedDeputado.siglaPartido} - {selectedDeputado.siglaUf}
                </p>
              </div>
            </div>
            
            <div className="space-y-3 mb-6 text-base">
              <p><strong>ID:</strong> {selectedDeputado.id}</p>
              <p><strong>Situação:</strong> {selectedDeputado.condicaoEleitoral}</p>
              {selectedDeputado.email && (
                <p><strong>Email:</strong> {selectedDeputado.email}</p>
              )}
            </div>

            <div className="flex space-x-3">
              <button 
                onClick={() => {
                  setSelectedDeputado(null);
                  setDespesas([]);
                  setDespesasError(null);
                }}
                className="flex-1 bg-gray-200 text-gray-900 text-base font-medium py-3 px-4 rounded-md 
                           hover:bg-gray-300"
              >
                Fechar
              </button>
              <button 
                onClick={() => handleVerDespesas(selectedDeputado)}
                className="flex-1 bg-blue-700 text-white text-base font-medium py-3 px-4 rounded-md 
                           hover:bg-blue-800 flex items-center justify-center"
              >
                <Euro className="h-4 w-4 mr-2" />
                Ver Gastos
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Modal de Despesas - Simplificado */}
      {showDespesas && selectedDeputado && (
        <div 
          className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50"
        >
          <div className="bg-white rounded-lg max-w-4xl w-full max-h-[80vh] overflow-hidden">
            <div className="p-6 border-b border-gray-200">
              <div className="flex items-center justify-between">
                <h2 className="text-xl font-semibold text-gray-900">
                  Gastos Públicos - {selectedDeputado.nome}
                </h2>
                <button 
                  onClick={() => {
                    setShowDespesas(false);
                    setDespesas([]);
                    setDespesasError(null);
                  }}
                  className="text-gray-400 hover:text-gray-600"
                >
                  <X className="h-6 w-6" />
                </button>
              </div>
              
              <div className="mt-4">
                <select
                  value={selectedAno}
                  onChange={(e) => {
                    setSelectedAno(Number(e.target.value));
                    fetchDespesas(selectedDeputado.id, Number(e.target.value));
                  }}
                  className="border border-gray-300 rounded-md px-3 py-2"
                >
                  {ANALYTICS_CONFIG.ANOS_DISPONIVEIS.map(year => (
                    <option key={year} value={year}>{year}</option>
                  ))}
                </select>
              </div>
            </div>

            <div className="p-6 overflow-y-auto max-h-[60vh]">
              {loadingDespesas ? (
                <div className="flex items-center justify-center py-8">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-700"></div>
                  <span className="ml-2">Carregando gastos...</span>
                </div>
              ) : despesasError ? (
                <div className="text-center py-8">
                  <p className="text-red-700">{despesasError}</p>
                  <button 
                    onClick={() => selectedDeputado && fetchDespesas(selectedDeputado.id, selectedAno)}
                    className="mt-2 text-blue-600 hover:text-blue-800"
                  >
                    Tentar novamente
                  </button>
                </div>
              ) : despesas.length === 0 ? (
                <div className="text-center py-8">
                  <Euro className="h-16 w-16 text-gray-400 mx-auto mb-4" />
                  <p className="text-gray-800">Nenhum gasto encontrado para este período.</p>
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
                        <p className="text-sm text-gray-600">
                          <strong>Fornecedor:</strong> {despesa.nomeFornecedor}
                        </p>
                      )}
                      
                      <div className="text-xs text-gray-500 mt-2">
                        {formatDate(despesa.dataDocumento)}
                      </div>
                    </div>
                  ))}
                  
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