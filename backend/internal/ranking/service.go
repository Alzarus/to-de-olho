package ranking

import (
	"context"
	"encoding/json"
	"log/slog"
	"sort"
	"time"

	"github.com/pedroalmeida/to-de-olho/internal/ceaps"
	"github.com/pedroalmeida/to-de-olho/internal/comissao"
	"github.com/pedroalmeida/to-de-olho/internal/proposicao"
	"github.com/pedroalmeida/to-de-olho/internal/senador"
	"github.com/pedroalmeida/to-de-olho/internal/votacao"
	"github.com/redis/go-redis/v9"
)

// Service gerencia o calculo de ranking de senadores
type Service struct {
	senadorRepo    *senador.Repository
	proposicaoRepo *proposicao.Repository
	votacaoRepo    *votacao.Repository
	ceapsRepo      *ceaps.Repository
	comissaoRepo   *comissao.Repository
	redisClient    *redis.Client
}

// NewService cria um novo servico de ranking
func NewService(
	senadorRepo *senador.Repository,
	proposicaoRepo *proposicao.Repository,
	votacaoRepo *votacao.Repository,
	ceapsRepo *ceaps.Repository,
	comissaoRepo *comissao.Repository,
	redisClient *redis.Client,
) *Service {
	return &Service{
		senadorRepo:    senadorRepo,
		proposicaoRepo: proposicaoRepo,
		votacaoRepo:    votacaoRepo,
		ceapsRepo:      ceapsRepo,
		comissaoRepo:   comissaoRepo,
		redisClient:    redisClient,
	}
}

// CalcularRanking calcula o ranking de todos os senadores
func (s *Service) CalcularRanking(ctx context.Context) (*RankingResponse, error) {
	// 1. Tentar buscar do cache
	cacheKey := "ranking:geral"
	if s.redisClient != nil {
		val, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var response RankingResponse
			if err := json.Unmarshal([]byte(val), &response); err == nil {
				slog.Info("ranking retornado do cache")
				return &response, nil
			}
		}
	}

	slog.Info("iniciando calculo de ranking (cache miss)")

	// Buscar todos os senadores
	senadores, err := s.senadorRepo.FindAll()
	if err != nil {
		return nil, err
	}

	// Primeiro, coletar dados brutos de todos para normalizacao
	var maxPontuacaoProd float64
	var maxPontosComissoes float64

	dadosBrutos := make(map[int]*dadosBrutosSenador)

	for _, sen := range senadores {
		dados := s.coletarDadosBrutos(sen.ID)
		dadosBrutos[sen.ID] = dados

		if float64(dados.pontuacaoProposicoes) > maxPontuacaoProd {
			maxPontuacaoProd = float64(dados.pontuacaoProposicoes)
		}
		if dados.pontosComissoes > maxPontosComissoes {
			maxPontosComissoes = dados.pontosComissoes
		}
	}

	// Garantir minimos para evitar divisao por zero
	if maxPontuacaoProd == 0 {
		maxPontuacaoProd = 1
	}
	if maxPontosComissoes == 0 {
		maxPontosComissoes = 1
	}

	// Calcular scores normalizados
	var scores []SenadorScore
	anoAtual := time.Now().Year()

	for _, sen := range senadores {
		dados := dadosBrutos[sen.ID]
		score := s.calcularScoreNormalizado(sen, dados, maxPontuacaoProd, maxPontosComissoes, anoAtual)
		scores = append(scores, score)
	}

	// Ordenar por score final (decrescente)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].ScoreFinal > scores[j].ScoreFinal
	})

	// Atribuir posicoes
	for i := range scores {
		scores[i].Posicao = i + 1
	}

	slog.Info("ranking calculado", "total_senadores", len(scores))

	response := &RankingResponse{
		Ranking:     scores,
		Total:       len(scores),
		CalculadoEm: time.Now(),
		Metodologia: "Score = (Produtividade * 0.35) + (Presenca * 0.25) + (Economia * 0.20) + (Comissoes * 0.20)",
	}

	// Salvar no cache (TTL 1 hora)
	if s.redisClient != nil {
		if data, err := json.Marshal(response); err == nil {
			s.redisClient.Set(ctx, cacheKey, data, 1*time.Hour)
		}
	}

	return response, nil
}

// CalcularScoreSenador calcula o score de um senador especifico
func (s *Service) CalcularScoreSenador(ctx context.Context, senadorID int) (*SenadorScore, error) {
	sen, err := s.senadorRepo.FindByID(senadorID)
	if err != nil {
		return nil, err
	}

	// Precisamos dos maximos para normalizacao
	// Em producao, isso poderia vir de cache
	senadores, err := s.senadorRepo.FindAll()
	if err != nil {
		return nil, err
	}

	var maxPontuacaoProd float64
	var maxPontosComissoes float64

	for _, s2 := range senadores {
		dados := s.coletarDadosBrutos(s2.ID)
		if float64(dados.pontuacaoProposicoes) > maxPontuacaoProd {
			maxPontuacaoProd = float64(dados.pontuacaoProposicoes)
		}
		if dados.pontosComissoes > maxPontosComissoes {
			maxPontosComissoes = dados.pontosComissoes
		}
	}

	if maxPontuacaoProd == 0 {
		maxPontuacaoProd = 1
	}
	if maxPontosComissoes == 0 {
		maxPontosComissoes = 1
	}

	dados := s.coletarDadosBrutos(senadorID)
	anoAtual := time.Now().Year()
	score := s.calcularScoreNormalizado(*sen, dados, maxPontuacaoProd, maxPontosComissoes, anoAtual)

	return &score, nil
}

// dadosBrutosSenador armazena dados brutos antes da normalizacao
type dadosBrutosSenador struct {
	// Proposicoes
	totalProposicoes     int
	proposicoesAprovadas int
	transformadasEmLei   int
	pontuacaoProposicoes int

	// Votacoes
	totalVotacoes     int
	votosRegistrados  int
	taxaPresencaBruta float64

	// CEAPS
	gastoAnual float64

	// Comissoes
	comissoesAtivas   int
	comissoesTitular  int
	comissoesSuplente int
	pontosComissoes   float64
}

// coletarDadosBrutos busca dados de todos os modulos para um senador
func (s *Service) coletarDadosBrutos(senadorID int) *dadosBrutosSenador {
	dados := &dadosBrutosSenador{}

	// Proposicoes
	if propStats, err := s.proposicaoRepo.GetStats(senadorID); err == nil {
		dados.totalProposicoes = propStats.TotalProposicoes
		dados.proposicoesAprovadas = propStats.AprovadosPlenario
		dados.transformadasEmLei = propStats.TransformadasEmLei
		dados.pontuacaoProposicoes = propStats.PontuacaoTotal
	}

	// Votacoes
	if votStats, err := s.votacaoRepo.GetStats(senadorID); err == nil {
		dados.totalVotacoes = votStats.TotalVotacoes
		dados.votosRegistrados = votStats.VotosRegistrados
		dados.taxaPresencaBruta = votStats.TaxaPresenca
	}

	// CEAPS
	anoAtual := time.Now().Year()
	if gastoAnual, err := s.ceapsRepo.GetTotalByAno(senadorID, anoAtual); err == nil {
		dados.gastoAnual = gastoAnual
	}

	// Comissoes
	if comStats, err := s.comissaoRepo.GetStats(senadorID); err == nil {
		dados.comissoesAtivas = comStats.ComissoesAtivas
		dados.comissoesTitular = comStats.ComissoesTitular
		dados.comissoesSuplente = comStats.ComissoesSuplente
		// Pontuacao: Titular = 2 pts, Suplente = 1 pt, Ativa = 1 pt bonus
		dados.pontosComissoes = float64(comStats.ComissoesTitular*2 + comStats.ComissoesSuplente + comStats.ComissoesAtivas)
	}

	return dados
}

// calcularScoreNormalizado calcula o score final normalizado
func (s *Service) calcularScoreNormalizado(
	sen senador.Senador,
	dados *dadosBrutosSenador,
	maxPontuacaoProd float64,
	maxPontosComissoes float64,
	anoAtual int,
) SenadorScore {
	// Normalizar Produtividade (0-100)
	produtividade := (float64(dados.pontuacaoProposicoes) / maxPontuacaoProd) * 100

	// Presenca ja vem normalizada (0-100)
	presenca := dados.taxaPresencaBruta

	// Economia CEAPS (0-100)
	// Quanto menos gasta, maior o score
	// Buscar teto da UF, se nao houver usa media
	tetoMensal, ok := TetoCEAPSPorUF[sen.UF]
	if !ok {
		tetoMensal = 40000.0 // Fallback seguro
	}
	tetoAnual := tetoMensal * 12

	economia := (1 - (dados.gastoAnual / tetoAnual)) * 100
	if economia < 0 {
		economia = 0 // Se gastou mais que o teto, score 0
	}
	if economia > 100 {
		economia = 100 // Cap em 100
	}

	// Comissoes (0-100)
	comissoes := (dados.pontosComissoes / maxPontosComissoes) * 100

	// Score final ponderado
	scoreFinal := (produtividade * PesoProdutividade) +
		(presenca * PesoPresenca) +
		(economia * PesoEconomia) +
		(comissoes * PesoComissoes)

	return SenadorScore{
		SenadorID:     sen.ID,
		Nome:          sen.Nome,
		Partido:       sen.Partido,
		UF:            sen.UF,
		FotoURL:       sen.FotoURL,
		Produtividade: arredondar(produtividade),
		Presenca:      arredondar(presenca),
		EconomiaCota:  arredondar(economia),
		Comissoes:     arredondar(comissoes),
		ScoreFinal:    arredondar(scoreFinal),
		CalculadoEm:   time.Now(),
		Detalhes: ScoreDetalhes{
			TotalProposicoes:     dados.totalProposicoes,
			ProposicoesAprovadas: dados.proposicoesAprovadas,
			TransformadasEmLei:   dados.transformadasEmLei,
			PontuacaoProposicoes: dados.pontuacaoProposicoes,
			TotalVotacoes:        dados.totalVotacoes,
			VotacoesParticipadas: dados.votosRegistrados,
			TaxaPresencaBruta:    arredondar(dados.taxaPresencaBruta),
			GastoCEAPS:           arredondar(dados.gastoAnual),
			TetoCEAPS:            tetoAnual,
			ComissoesAtivas:      dados.comissoesAtivas,
			ComissoesTitular:     dados.comissoesTitular,
			ComissoesSuplente:    dados.comissoesSuplente,
			PontosComissoes:      arredondar(dados.pontosComissoes),
		},
	}
}

// arredondar arredonda para 2 casas decimais
func arredondar(valor float64) float64 {
	return float64(int(valor*100+0.5)) / 100
}
