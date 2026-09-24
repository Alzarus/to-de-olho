// Types espelhando modelos do backend Go

export interface ScoreDetalhes {
  // Produtividade
  total_proposicoes: number; // autoria principal (primeiro autor)
  total_coautorias: number; // coautorias: aparecem na ficha, nao pontuam
  total_sem_pontos?: number; // vetos, autoria como deputado ou institucional
  proposicoes_aprovadas: number;
  transformadas_em_lei: number;
  pontuacao_proposicoes: number;

  // Presenca
  total_votacoes: number;
  votacoes_participadas: number;
  presentes: number;
  ausencias_ap: number; // "Atividade parlamentar": conta como falta
  nao_compareceu: number;
  ausencias_justificadas: number; // licencas e missoes: fora do denominador
  taxa_presenca_bruta: number; // A
  taxa_presenca_ajustada: number; // B (a que entra no score)

  // Economia CEAPS
  gasto_ceaps: number;
  teto_ceaps: number; // teto mensal da UF x meses em exercicio
  meses_exercicio?: number; // no periodo

  // Comissoes
  comissoes_ativas: number;
  comissoes_titular: number;
  comissoes_suplente: number;
  pontos_comissoes: number;
  comissoes_fora_da_conta?: number; // frentes, grupos de amizade, honrarias
}

export interface SenadorScore {
  senador_id: number;
  nome: string;
  partido: string;
  uf: string;
  foto_url?: string;
  cargo?: string;
  titular?: string;

  // Scores individuais normalizados (0-100)
  produtividade: number;
  presenca: number | null; // null: sem registro de votacao no periodo
  economia_cota: number;
  comissoes: number;

  // Score final ponderado (0-100)
  score_final: number;
  posicao: number; // 0 quando fora da ordenacao
  dados_insuficientes: boolean;
  motivo?: string; // por que ficou fora da ordenacao

  // Detalhes para transparencia
  detalhes: ScoreDetalhes;
  calculado_em: string;
}

export interface RankingResponse {
  ranking: SenadorScore[];
  sem_dados?: SenadorScore[]; // fora da ordenacao: "dados insuficientes"
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
  cargo?: string;
  titular?: string;
  em_exercicio?: boolean;
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
  valor: number;
  data_emissao: string;
  detalhe: string;
  fornecedor: string;
}

export interface DespesaAgregado {
  tipo_despesa: string;
  total: number;
}

export interface DespesasResponse {
  senador_id: number;
  total: number;
  limit: number;
  page: number;
  total_pages: number;
  despesas: Despesa[];
}

export interface DespesasAgregadoResponse {
  senador_id: number;
  total_geral: number;
  por_tipo: DespesaAgregado[];
}

export interface GastoMensal {
  ano: number;
  mes: number;
  total: number;
}

export interface DespesasMensalResponse {
  senador_id: number;
  meses: GastoMensal[];
}

export interface FornecedorAgregado {
  fornecedor: string;
  cnpj_cpf: string;
  total: number;
  quantidade: number;
}

export interface DespesasFornecedoresResponse {
  senador_id: number;
  fornecedores: FornecedorAgregado[];
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

// Proposicoes
export interface Proposicao {
  id: number;
  senador_id: number;
  codigo_materia: string;
  sigla_subtipo_materia: string;
  numero_materia: string;
  ano_materia: number;
  descricao_identificacao: string;
  ementa: string;
  situacao_atual: string;
  data_apresentacao?: string;
  estagio_tramitacao: string;
  pontuacao: number;
  posicao_autoria?: number | null; // 1 = primeiro autor; null = autoria institucional
  tipo_autor?: string; // SENADOR, LIDER, PRESIDENTE_SF, DEPUTADO
  total_autores?: number | null;
  autoria?: string;
}

export interface ProposicaoResponse {
  senador_id: number;
  total: number;
  limit: number;
  page: number;
  total_pages: number;
  proposicoes: Proposicao[];
}

// Comissoes
export interface ComissaoMembro {
  id: number;
  senador_id: number;
  codigo_comissao: string;
  sigla_comissao: string;
  nome_comissao: string;
  sigla_casa_comissao: string;
  descricao_participacao: string;
  data_inicio?: string;
  data_fim?: string;
}

export interface ComissoesResponse {
  senador_id: number;
  total: number;
  limit: number;
  page: number;
  total_pages: number;
  comissoes: ComissaoMembro[];
}
// Votacoes
export interface VotacaoItem {
  id: number;
  codigo_votacao: number;
  sessao_id: string;
  sigla_voto: string;
  data: string;
  voto: string;
  materia: string;
  descricao_votacao: string;
}

export interface VotacoesResponse {
  senador_id: number;
  total: number;
  limit: number;
  page: number;
  total_pages: number;
  votacoes: VotacaoItem[];
}

// Stats da plataforma (home page)
export interface StatsResponse {
  total_senadores: number;
  total_votos: number; // votos individuais (um por senador por votação)
  total_votacoes: number; // votações nominais distintas
  votacoes_desde: string | null;
  total_despesas_ceaps: number;
  ceaps_ano_inicio: number;
  ceaps_ano_fim: number;
  total_emendas: number;
  total_acessos: number; // visitantes únicos por dia, somados
  acessos_desde: string | null; // início da contagem real
  ultima_atualizacao: string;
}
