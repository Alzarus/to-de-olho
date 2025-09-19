'use client';

import { useState, useEffect } from 'react';
import axios from 'axios';
import { Search, User, MapPin, Building2, Euro, AlertCircle, X } from 'lucide-react';
import DeputadoCard from './DeputadoCard';
import Tooltip from './Tooltip';

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
  
  // Estados para paginação
  const [pagination, setPagination] = useState<PaginationInfo>({
    page: 1,
    limit: 20,
    total: 0,
    total_pages: 1,
    has_next: false,
    has_prev: false
  });
  const [currentPage, setCurrentPage] = useState(1);

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
  }, [selectedUF, selectedPartido, currentPage]);

  const fetchDeputados = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const params = new URLSearchParams();
      if (selectedUF) params.append('uf', selectedUF);
      if (selectedPartido) params.append('partido', selectedPartido);
      if (searchTerm) params.append('nome', searchTerm);
      params.append('page', currentPage.toString());
      params.append('limit', '20');

      const response = await axios.get<APIResponse>(`${API_BASE_URL}/deputados?${params}`);
      
      // Sistema Ultra-Performance: dados em response.data.data.data e paginação em response.data.data.pagination
      const deputadosData = response.data?.data?.data;
      const paginationData = response.data?.data?.pagination;
      
      if (Array.isArray(deputadosData)) {
        setDeputados(deputadosData);
      } else {
        console.warn('Sistema Ultra-Performance: Dados de deputados não são um array');
        console.warn('Estrutura recebida:', response.data);
        console.warn('Esperado: response.data.data.data (Array) e response.data.data.pagination (Object)');
        setDeputados([]);
      }

      // Atualizar informações de paginação
      if (paginationData) {
        // Mapear estrutura do backend para estrutura do frontend
        setPagination({
          page: paginationData.page || currentPage,
          limit: paginationData.limit || 20,
          total: paginationData.total || 0,
          total_pages: paginationData.total_pages || 1,
          has_next: paginationData.has_next || false,
          has_prev: paginationData.has_prev || false
        });
      } else {
        // Fallback caso a API não retorne paginação
        setPagination({
          page: currentPage,
          limit: 20,
          total: deputadosData?.length || 0,
          total_pages: 1,
          has_next: false,
          has_prev: false
        });
      }
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
      
      // Garantir que sempre seja um array
      const despesasData = response.data?.data;
      if (Array.isArray(despesasData)) {
        setDespesas(despesasData);
      } else {
        console.warn('Dados de despesas não são um array:', despesasData);
        setDespesas([]);
      }
    } catch (err) {
      console.error('Erro ao buscar despesas:', err);
      setDespesasError('Erro ao carregar despesas. Tente novamente.');
      setDespesas([]); // Garantir que seja array em caso de erro
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
    setCurrentPage(1); // Reset para primeira página ao buscar
    fetchDeputados();
  };

  const handlePageChange = (newPage: number) => {
    if (newPage >= 1 && newPage <= pagination.total_pages) {
      setCurrentPage(newPage);
    }
  };

  const handleFilterChange = () => {
    setCurrentPage(1); // Reset para primeira página ao mudar filtros
  };

  if (loading) {
    return (
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <div className="flex items-center justify-center py-8" role="status" aria-live="polite">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-700" aria-hidden="true"></div>
          <span className="ml-3 text-lg text-gray-900">Carregando deputados...</span>
          <span className="sr-only">Por favor aguarde, buscando dados dos deputados federais</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <div className="bg-red-50 border border-red-200 rounded-lg p-6" role="alert">
          <div className="flex items-center">
            <AlertCircle className="h-6 w-6 text-red-600 mr-3" aria-hidden="true" />
            <div>
              <h3 className="text-lg font-medium text-red-800">Erro ao carregar dados</h3>
              <p className="text-red-700 mt-1 text-base">{error}</p>
              <button 
                onClick={fetchDeputados}
                className="mt-3 bg-red-700 text-white text-base font-medium px-6 py-3 rounded 
                           hover:bg-red-800 focus:outline-none focus:ring-4 focus:ring-red-300
                           transition-colors duration-200"
                aria-label="Tentar carregar os dados novamente"
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
          <div>
            <label 
              htmlFor="search-input"
              className="block text-base font-medium text-gray-900 mb-2"
            >
              <Search className="inline h-4 w-4 mr-1" aria-hidden="true" />
              Buscar por nome
            </label>
            <input
              id="search-input"
              type="text"
              placeholder="Digite o nome do deputado"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full text-base border border-gray-300 rounded-md px-4 py-3 
                         focus:outline-none focus:ring-4 focus:ring-blue-300 focus:border-blue-500
                         hover:border-gray-400 transition-colors"
              aria-describedby="search-help"
            />
            <div id="search-help" className="sr-only">
              Digite o nome do deputado para filtrar a lista de resultados
            </div>
          </div>

          <div>
            <label 
              htmlFor="uf-select"
              className="block text-base font-medium text-gray-900 mb-2"
            >
              <MapPin className="inline h-4 w-4 mr-1" aria-hidden="true" />
              Estado (UF)
              <Tooltip 
                content="Unidade Federativa do Brasil onde o deputado foi eleito. Cada estado tem uma quantidade específica de deputados baseada na população."
              />
            </label>
            <select
              id="uf-select"
              value={selectedUF}
              onChange={(e) => {
                setSelectedUF(e.target.value);
                handleFilterChange();
              }}
              className="w-full text-base border border-gray-300 rounded-md px-4 py-3 
                         focus:outline-none focus:ring-4 focus:ring-blue-300 focus:border-blue-500
                         hover:border-gray-400 transition-colors"
              aria-describedby="uf-help"
            >
              <option value="">Todos os estados</option>
              {estados.map(uf => (
                <option key={uf} value={uf}>{uf}</option>
              ))}
            </select>
            <div id="uf-help" className="sr-only">
              Selecione um estado para filtrar apenas deputados daquela unidade federativa
            </div>
          </div>

          <div>
            <label 
              htmlFor="partido-select"
              className="block text-base font-medium text-gray-900 mb-2"
            >
              <Building2 className="inline h-4 w-4 mr-1" aria-hidden="true" />
              Partido Político
              <Tooltip 
                content="Partidos políticos são organizações que representam diferentes ideologias e propostas para o país. Cada deputado é filiado a um partido."
              />
            </label>
            <select
              id="partido-select"
              value={selectedPartido}
              onChange={(e) => {
                setSelectedPartido(e.target.value);
                handleFilterChange();
              }}
              className="w-full text-base border border-gray-300 rounded-md px-4 py-3 
                         focus:outline-none focus:ring-4 focus:ring-blue-300 focus:border-blue-500
                         hover:border-gray-400 transition-colors"
              aria-describedby="partido-help"
            >
              <option value="">Todos os partidos</option>
              {partidos.map(partido => (
                <option key={partido} value={partido}>{partido}</option>
              ))}
            </select>
            <div id="partido-help" className="sr-only">
              Selecione um partido político para filtrar apenas deputados filiados a ele
            </div>
          </div>

          <div className="flex items-end">
            <button
              onClick={handleSearch}
              className="w-full bg-blue-700 text-white text-base font-medium px-6 py-3 rounded-md 
                         hover:bg-blue-800 focus:outline-none focus:ring-4 focus:ring-blue-300
                         transition-colors duration-200"
              aria-label="Buscar deputados com os filtros selecionados"
            >
              Buscar Deputados
            </button>
          </div>
        </div>
      </div>

      {/* Lista de Deputados */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">{/* resto do código... */}
          <div>
            <label 
              htmlFor="search-input"
              className="block text-base font-medium text-gray-900 mb-2"
            >
              <Search className="inline h-4 w-4 mr-1" aria-hidden="true" />
              Buscar por nome
            </label>
            <input
              id="search-input"
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Digite o nome do deputado..."
              className="w-full text-base border border-gray-300 rounded-md px-4 py-3 
                         focus:outline-none focus:ring-4 focus:ring-blue-300 focus:border-blue-500
                         hover:border-gray-400 transition-colors"
              aria-describedby="search-help"
            />
            <div id="search-help" className="sr-only">
              Digite o nome completo ou parcial do deputado federal que deseja encontrar
            </div>
          </div>
          
          <div>
            <label 
              htmlFor="uf-select"
              className="block text-base font-medium text-gray-900 mb-2"
            >
              <MapPin className="inline h-4 w-4 mr-1" aria-hidden="true" />
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
                         focus:outline-none focus:ring-4 focus:ring-blue-300 focus:border-blue-500
                         hover:border-gray-400 transition-colors"
              aria-describedby="uf-help"
            >
              <option value="">Todos os estados</option>
              {estados.map(uf => (
                <option key={uf} value={uf}>{uf}</option>
              ))}
            </select>
            <div id="uf-help" className="sr-only">
              Selecione um estado para filtrar apenas deputados daquela unidade federativa
            </div>
          </div>

          <div>
            <label 
              htmlFor="partido-select"
              className="block text-base font-medium text-gray-900 mb-2"
            >
              <Building2 className="inline h-4 w-4 mr-1" aria-hidden="true" />
              Partido Político
              <Tooltip 
                content="Partidos políticos são organizações que representam diferentes ideologias e propostas para o país. Cada deputado é filiado a um partido."
              />
            </label>
            <select
              id="partido-select"
              value={selectedPartido}
              onChange={(e) => {
                setSelectedPartido(e.target.value);
                handleFilterChange();
              }}
              className="w-full text-base border border-gray-300 rounded-md px-4 py-3 
                         focus:outline-none focus:ring-4 focus:ring-blue-300 focus:border-blue-500
                         hover:border-gray-400 transition-colors"
              aria-describedby="partido-help"
            >
              <option value="">Todos os partidos</option>
              {partidos.map(partido => (
                <option key={partido} value={partido}>{partido}</option>
              ))}
            </select>
            <div id="partido-help" className="sr-only">
              Selecione um partido político para filtrar apenas deputados filiados a ele
            </div>
          </div>

          <div className="flex items-end">
            <button
              onClick={handleSearch}
              className="w-full bg-blue-700 text-white text-base font-medium px-6 py-3 rounded-md 
                         hover:bg-blue-800 focus:outline-none focus:ring-4 focus:ring-blue-300
                         transition-colors duration-200"
              aria-label="Buscar deputados com os filtros selecionados"
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
        <div className="mt-8 flex items-center justify-between">
          <div className="text-sm text-gray-700">
            Mostrando página {pagination.page} de {pagination.total_pages} 
            ({pagination.total} deputados no total)
          </div>
          
          <div className="flex items-center space-x-2">
            <button
              onClick={() => handlePageChange(currentPage - 1)}
              disabled={!pagination.has_prev}
              className={`px-3 py-2 text-sm font-medium rounded-md ${
                pagination.has_prev
                  ? 'text-gray-700 bg-white border border-gray-300 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500'
                  : 'text-gray-400 bg-gray-100 border border-gray-200 cursor-not-allowed'
              }`}
              aria-label="Página anterior"
            >
              Anterior
            </button>

            {/* Números das páginas */}
            <div className="flex items-center space-x-1">
              {(() => {
                const pages = [];
                const startPage = Math.max(1, currentPage - 2);
                const endPage = Math.min(pagination.total_pages, currentPage + 2);

                if (startPage > 1) {
                  pages.push(
                    <button
                      key={1}
                      onClick={() => handlePageChange(1)}
                      className="px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      1
                    </button>
                  );
                  if (startPage > 2) {
                    pages.push(<span key="start-ellipsis" className="px-2 text-gray-500">...</span>);
                  }
                }

                for (let i = startPage; i <= endPage; i++) {
                  pages.push(
                    <button
                      key={i}
                      onClick={() => handlePageChange(i)}
                      className={`px-3 py-2 text-sm font-medium rounded-md ${
                        i === currentPage
                          ? 'text-blue-600 bg-blue-50 border border-blue-300'
                          : 'text-gray-700 bg-white border border-gray-300 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500'
                      }`}
                      aria-label={`Página ${i}`}
                      aria-current={i === currentPage ? 'page' : undefined}
                    >
                      {i}
                    </button>
                  );
                }

                if (endPage < pagination.total_pages) {
                  if (endPage < pagination.total_pages - 1) {
                    pages.push(<span key="end-ellipsis" className="px-2 text-gray-500">...</span>);
                  }
                  pages.push(
                    <button
                      key={pagination.total_pages}
                      onClick={() => handlePageChange(pagination.total_pages)}
                      className="px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      {pagination.total_pages}
                    </button>
                  );
                }

                return pages;
              })()}
            </div>

            <button
              onClick={() => handlePageChange(currentPage + 1)}
              disabled={!pagination.has_next}
              className={`px-3 py-2 text-sm font-medium rounded-md ${
                pagination.has_next
                  ? 'text-gray-700 bg-white border border-gray-300 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500'
                  : 'text-gray-400 bg-gray-100 border border-gray-200 cursor-not-allowed'
              }`}
              aria-label="Próxima página"
            >
              Próxima
            </button>
          </div>
        </div>
      )}

      {deputados.length === 0 && !loading && (
        <div className="text-center py-12" role="status" aria-live="polite">
          <User className="h-16 w-16 text-gray-400 mx-auto mb-4" aria-hidden="true" />
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
          role="dialog"
          aria-modal="true"
          aria-labelledby="deputy-modal-title"
          onClick={() => setSelectedDeputado(null)}
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
                  <User className="h-10 w-10 text-gray-400" aria-hidden="true" />
                </div>
              )}
              <div>
                <h2 id="deputy-modal-title" className="text-xl font-semibold text-gray-900">
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
                onClick={() => setSelectedDeputado(null)}
                className="flex-1 bg-gray-200 text-gray-900 text-base font-medium py-3 px-4 rounded-md 
                           hover:bg-gray-300 focus:outline-none focus:ring-4 focus:ring-gray-300
                           transition-colors duration-200"
                aria-label="Fechar janela de detalhes do deputado"
              >
                Fechar
              </button>
              <button 
                onClick={() => handleVerDespesas(selectedDeputado)}
                className="flex-1 bg-blue-700 text-white text-base font-medium py-3 px-4 rounded-md 
                           hover:bg-blue-800 focus:outline-none focus:ring-4 focus:ring-blue-300 
                           flex items-center justify-center transition-colors duration-200"
                aria-label={`Ver gastos parlamentares de ${selectedDeputado.nome}`}
              >
                <Euro className="h-4 w-4 mr-2" aria-hidden="true" />
                Ver Gastos Públicos
                <Tooltip 
                  content="Gastos parlamentares são recursos públicos utilizados pelos deputados para exercer suas funções, como passagens, hospedagem e material de escritório."
                />
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Modal de Despesas */}
      {showDespesas && (
        <div 
          className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50"
          role="dialog"
          aria-modal="true"
          aria-labelledby="expenses-modal-title"
        >
          <div className="bg-white rounded-lg max-w-4xl w-full max-h-[80vh] overflow-hidden">
            <div className="p-6 border-b border-gray-200">
              <div className="flex items-center justify-between">
                <div className="flex items-center">
                  <h2 id="expenses-modal-title" className="text-xl font-semibold text-gray-900">
                    Gastos Públicos do Deputado
                  </h2>
                  <Tooltip 
                    content="Estes são os gastos realizados com verba pública destinada ao exercício das funções parlamentares. Todos os valores são pagos pelos contribuintes brasileiros."
                  />
                </div>
                <button 
                  onClick={() => setShowDespesas(false)}
                  className="text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-gray-300 rounded"
                  aria-label="Fechar janela de gastos"
                >
                  <X className="h-6 w-6" />
                </button>
              </div>
              
              {/* Seletor de Ano */}
              <div className="mt-4">
                <label 
                  htmlFor="year-select"
                  className="block text-base font-medium text-gray-900 mb-2"
                >
                  Ano dos gastos:
                </label>
                <select
                  id="year-select"
                  value={selectedAno}
                  onChange={(e) => setSelectedAno(Number(e.target.value))}
                  className="border border-gray-300 rounded-md px-3 py-2 text-base 
                             focus:outline-none focus:ring-4 focus:ring-blue-300 focus:border-blue-500"
                  aria-describedby="year-help"
                >
                  {Array.from({ length: 5 }, (_, i) => new Date().getFullYear() - i).map(year => (
                    <option key={year} value={year}>{year}</option>
                  ))}
                </select>
                <div id="year-help" className="sr-only">
                  Selecione o ano para consultar os gastos parlamentares
                </div>
              </div>
            </div>

            <div className="p-6 overflow-y-auto max-h-[60vh]">
              {loadingDespesas ? (
                <div className="flex items-center justify-center py-8" role="status" aria-live="polite">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-700" aria-hidden="true"></div>
                  <span className="ml-2 text-lg text-gray-900">Carregando gastos...</span>
                  <span className="sr-only">Por favor aguarde, buscando dados dos gastos parlamentares</span>
                </div>
              ) : despesasError ? (
                <div className="text-center py-8" role="alert">
                  <p className="text-red-700 text-base">{despesasError}</p>
                  <button 
                    onClick={() => selectedDeputado && fetchDespesas(selectedDeputado.id, selectedAno)}
                    className="mt-2 text-blue-700 hover:text-blue-900 focus:outline-none focus:ring-2 focus:ring-blue-300 rounded text-base font-medium"
                    aria-label="Tentar carregar os gastos novamente"
                  >
                    Tentar novamente
                  </button>
                </div>
              ) : (!Array.isArray(despesas) || despesas.length === 0) ? (
                <div className="text-center py-8">
                  <Euro className="h-16 w-16 text-gray-400 mx-auto mb-4" aria-hidden="true" />
                  <p className="text-gray-800 text-lg">Nenhum gasto encontrado para este período.</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {Array.isArray(despesas) && despesas.map((despesa, index) => (
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
                        {formatCurrency(Array.isArray(despesas) ? despesas.reduce((sum, d) => sum + d.valorLiquido, 0) : 0)}
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
