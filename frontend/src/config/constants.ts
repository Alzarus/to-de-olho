/**
 * ⚙️ Constantes de Configuração - Frontend
 * 
 * Centralizando todos os valores configuráveis para evitar hardcoding
 * Seguindo as diretrizes do ROADMAP.md - Problema 2
 */

// 🌐 API e URLs
export const API_CONFIG = {
  BASE_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1',
  TIMEOUT: parseInt(process.env.NEXT_PUBLIC_API_TIMEOUT || '30000'), // 30s
} as const;

// 📄 Paginação
export const PAGINATION_CONFIG = {
  DEFAULT_PAGE_SIZE: parseInt(process.env.NEXT_PUBLIC_PAGE_SIZE || '20'),
  MAX_PAGE_SIZE: parseInt(process.env.NEXT_PUBLIC_MAX_PAGE_SIZE || '100'),
  ANALYTICS_RANKING_SIZE: parseInt(process.env.NEXT_PUBLIC_ANALYTICS_RANKING_SIZE || '10'),
} as const;

// ⏱️ Timing e Performance
export const TIMING_CONFIG = {
  // Debounce para filtros (em ms)
  SEARCH_DEBOUNCE_MS: parseInt(process.env.NEXT_PUBLIC_SEARCH_DEBOUNCE_MS || '500'),
  
  // Auto-refresh intervals (em ms)
  ANALYTICS_REFRESH_INTERVAL: parseInt(process.env.NEXT_PUBLIC_ANALYTICS_REFRESH_MS || '300000'), // 5min
  
  // Cache TTL para requests (em ms)  
  CACHE_TTL_MS: parseInt(process.env.NEXT_PUBLIC_CACHE_TTL_MS || '60000'), // 1min
} as const;

// 🎨 UI/UX
export const UI_CONFIG = {
  // Acessibilidade - tamanhos mínimos para touch targets
  MIN_TOUCH_TARGET_SIZE: '44px',
  
  // Breakpoints responsivos (seguindo Tailwind)
  BREAKPOINTS: {
    SM: '640px',
    MD: '768px', 
    LG: '1024px',
    XL: '1280px',
  },
  
  // Contraste e cores (WCAG 2.1 AA)
  MIN_CONTRAST_RATIO: 4.5,
} as const;

// 📊 Analytics e Métricas
export const ANALYTICS_CONFIG = {
  // Quantos itens mostrar nos rankings
  TOP_DEPUTADOS_COUNT: PAGINATION_CONFIG.ANALYTICS_RANKING_SIZE,
  TOP_PROPOSICOES_COUNT: PAGINATION_CONFIG.ANALYTICS_RANKING_SIZE,
  TOP_GASTOS_COUNT: PAGINATION_CONFIG.ANALYTICS_RANKING_SIZE,
  
  // Anos disponíveis para filtros (dinâmico baseado no ano atual)
  ANOS_DISPONIVEIS: (() => {
    const anoAtual = new Date().getFullYear();
    const anoInicial = parseInt(process.env.NEXT_PUBLIC_ANO_INICIAL || '2019');
    return Array.from({ length: anoAtual - anoInicial + 1 }, (_, i) => anoAtual - i);
  })(),
} as const;

// 🏛️ Dados Brasileiros
export const BRASIL_CONFIG = {
  ESTADOS: [
    'AC', 'AL', 'AP', 'AM', 'BA', 'CE', 'DF', 'ES', 'GO', 'MA', 
    'MT', 'MS', 'MG', 'PA', 'PB', 'PR', 'PE', 'PI', 'RJ', 'RN', 
    'RS', 'RO', 'RR', 'SC', 'SP', 'SE', 'TO'
  ],
  
  PARTIDOS: [
    'PT', 'PL', 'PP', 'MDB', 'PSD', 'REPUBLICANOS', 'PSB', 'UNIÃO', 
    'PSDB', 'PDT', 'SOLIDARIEDADE', 'PSOL', 'PODE', 'AVANTE', 'PC do B',
    'PV', 'CIDADANIA', 'PATRIOTA', 'PROS', 'PMB', 'AGIR', 'REDE'
  ],
} as const;

// 🔒 Validações
export const VALIDATION_CONFIG = {
  // Tamanhos mínimos/máximos para campos de busca
  MIN_SEARCH_LENGTH: parseInt(process.env.NEXT_PUBLIC_MIN_SEARCH_LENGTH || '2'),
  MAX_SEARCH_LENGTH: parseInt(process.env.NEXT_PUBLIC_MAX_SEARCH_LENGTH || '100'),
  
  // Regex patterns
  NOME_PATTERN: /^[a-zA-ZÀ-ÿ\s]+$/,
  ANO_PATTERN: /^\d{4}$/,
} as const;

// 🚀 Performance
export const PERFORMANCE_CONFIG = {
  // Lazy loading
  ENABLE_LAZY_LOADING: process.env.NEXT_PUBLIC_ENABLE_LAZY_LOADING !== 'false',
  
  // Service Worker cache
  ENABLE_SW_CACHE: process.env.NEXT_PUBLIC_ENABLE_SW_CACHE !== 'false',
  
  // Image optimization
  IMAGE_QUALITY: parseInt(process.env.NEXT_PUBLIC_IMAGE_QUALITY || '85'),
  
  // Bundle size limits (em KB)
  MAX_BUNDLE_SIZE_KB: parseInt(process.env.NEXT_PUBLIC_MAX_BUNDLE_SIZE_KB || '200'),
} as const;

// 🔧 Debug e Desenvolvimento
export const DEBUG_CONFIG = {
  ENABLE_LOGGING: process.env.NODE_ENV === 'development',
  LOG_LEVEL: process.env.NEXT_PUBLIC_LOG_LEVEL || 'info',
  ENABLE_PERFORMANCE_METRICS: process.env.NEXT_PUBLIC_ENABLE_PERF_METRICS === 'true',
} as const;