// Types espelhando modelos do backend Go

export interface ScoreDetalhes {
  // Produtividade
  total_proposicoes: number;
  proposicoes_aprovadas: number;
  transformadas_em_lei: number;
  pontuacao_proposicoes: number;

  // Presenca
  total_votacoes: number;
  votacoes_participadas: number;
  taxa_presenca_bruta: number;

  // Economia CEAPS
  gasto_ceaps: number;
  teto_ceaps: number;

  // Comissoes
  comissoes_ativas: number;
  comissoes_titular: number;
  comissoes_suplente: number;
  pontos_comissoes: number;
}

export interface SenadorScore {
  senador_id: number;
  nome: string;
  partido: string;
  uf: string;
  foto_url?: string;

  // Scores individuais normalizados (0-100)
  produtividade: number;
  presenca: number;
  economia_cota: number;
  comissoes: number;

  // Score final ponderado (0-100)
  score_final: number;
  posicao: number;

  // Detalhes para transparencia
  detalhes: ScoreDetalhes;
  calculado_em: string;
}

export interface RankingResponse {
  ranking: SenadorScore[];
  total: number;
  calculado_em: string;
  metodologia: string;
}

export interface Mandato {
  id: number;
  legislatura: number;
  inicio: string;
  fim?: string;
  tipo: string;
}

export interface Senador {
  id: number;
  codigo_parlamentar: number;
  nome: string;
  nome_completo: string;
  partido: string;
  uf: string;
  foto_url?: string;
  email?: string;
  telefone?: string;
  mandatos?: Mandato[];
}

export interface MetodologiaCriterio {
  nome: string;
  peso: string;
  descricao: string;
  normalizacao: string;
  formula_detalhada?: string;
}

export interface MetodologiaResponse {
  titulo: string;
  versao: string;
  referencia: string;
  formula: string;
  criterios: MetodologiaCriterio[];
  escala: string;
}

export interface VotosPorTipo {
  voto: string;
  total: number;
}

export interface VotosPorTipoResponse {
  senador_id: number;
  por_tipo: VotosPorTipo[];
}

export interface Despesa {
  ano: number;
  mes: number;
  senador_id: number;
  tipo_despesa: string;
  valor_reembolsado: number;
  data_documento: string;
  detalhe: string;
  fornecedor: string;
}

export interface DespesaAgregado {
  tipo: string;
  total: number;
}

export interface DespesasResponse {
  senador_id: number;
  total: number;
  despesas: Despesa[];
}

export interface DespesasAgregadoResponse {
  senador_id: number;
  total_geral: number;
  por_tipo: DespesaAgregado[];
}

// Emendas (RF08-RF10)
export interface LocalidadeValor {
  localidade: string;
  valor: number;
}

export interface ResumoEmendas {
  total_empenhado: number;
  total_pago: number;
  quantidade: number;
  top_localidades: LocalidadeValor[];
}

export interface Emenda {
  id: number;
  senador_id: number;
  ano: number;
  numero: string;
  tipo: string;
  funcional_programatica: string;
  localidade: string;
  valor_empenhado: number;
  valor_pago: number;
  data_ultima_atualizacao: string;
}

export interface EmendasResponse {
  emendas: Emenda[];
  resumo?: ResumoEmendas;
}
