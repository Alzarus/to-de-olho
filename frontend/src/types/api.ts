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
}

export interface MetodologiaCriterio {
  nome: string;
  peso: string;
  descricao: string;
  normalizacao: string;
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
