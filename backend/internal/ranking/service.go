package ranking

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"

	"github.com/Alzarus/to-de-olho/internal/ceaps"
	"github.com/Alzarus/to-de-olho/internal/comissao"
	"github.com/Alzarus/to-de-olho/internal/proposicao"
	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/internal/votacao"
)

// Service gerencia o calculo de ranking de senadores
type Service struct {
	senadorRepo    *senador.Repository
	proposicaoRepo *proposicao.Repository
	votacaoRepo    *votacao.Repository
	ceapsRepo      *ceaps.Repository
	comissaoRepo   *comissao.Repository
}

// NewService cria um novo servico de ranking
func NewService(
	senadorRepo *senador.Repository,
	proposicaoRepo *proposicao.Repository,
	votacaoRepo *votacao.Repository,
	ceapsRepo *ceaps.Repository,
	comissaoRepo *comissao.Repository,
) *Service {
	return &Service{
		senadorRepo:    senadorRepo,
		proposicaoRepo: proposicaoRepo,
		votacaoRepo:    votacaoRepo,
		ceapsRepo:      ceapsRepo,
		comissaoRepo:   comissaoRepo,
	}
}

// CalcularRanking calcula o ranking de todos os senadores
func (s *Service) CalcularRanking(ctx context.Context, ano *int) (*RankingResponse, error) {
	// 1. Tentar buscar do cache (Memória Local)
	// [COST-SAVING] Substituicao do Redis por cache em memoria local
	cacheKey := "ranking:v2:geral"
	if ano != nil {
		cacheKey = fmt.Sprintf("ranking:v2:%d", *ano)
	}

	// Tenta pegar do cache local
	if cached := localCache.Get(cacheKey); cached != nil {
		slog.Info("ranking retornado do cache local (RAM)", "key", cacheKey)
		return cached, nil
	}

	slog.Info("iniciando calculo de ranking (cache miss)", "ano", ano)

	// Buscar todos os senadores
	senadores, err := s.senadorRepo.FindAll(false)
	if err != nil {
		return nil, err
	}

	// Primeiro, coletar dados brutos de todos para normalizacao
	var maxPontuacaoProd float64
	var maxPontosComissoes float64

	dadosBrutos := make(map[int]*dadosBrutosSenador)

	for _, sen := range senadores {
		dados := s.coletarDadosBrutos(sen.ID, ano)
		dadosBrutos[sen.ID] = dados

		if dados.pontuacaoProposicoes > maxPontuacaoProd {
			maxPontuacaoProd = dados.pontuacaoProposicoes
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

	// Calcular scores normalizados; quem nao tem dado sai da ordenacao
	var scores, semDados []SenadorScore
	for _, sen := range senadores {
		dados := dadosBrutos[sen.ID]
		score := s.calcularScoreNormalizado(sen, dados, maxPontuacaoProd, maxPontosComissoes, ano)
		if score.DadosInsuficientes {
			semDados = append(semDados, score)
			continue
		}
		scores = append(scores, score)
	}
	ordenar(scores)

	slog.Info("ranking calculado", "total_senadores", len(scores), "sem_dados", len(semDados))

	metodologia := "Score = (Produtividade * 0.35) + (Presenca * 0.25) + (Economia * 0.20) + (Comissoes * 0.20)"
	if ano != nil {
		metodologia = fmt.Sprintf("Score (Ano %d) = (Produtividade * 0.35) + (Presenca * 0.25) + (Economia * 0.20) + (Comissoes * 0.20)", *ano)
	}

	response := &RankingResponse{
		Ranking:     scores,
		SemDados:    semDados,
		Total:       len(scores),
		CalculadoEm: time.Now(),
		Metodologia: metodologia,
	}

	// Salvar no cache local (TTL 24 horas)
	localCache.Set(cacheKey, response, 24*time.Hour)

	return response, nil
}

// InvalidateCache invalida todo o cache de ranking
func (s *Service) InvalidateCache() {
	slog.Info("invalidando cache de ranking")
	localCache.InvalidateAll()
}

// CalcularScoreSenador calcula o score de um senador especifico
func (s *Service) CalcularScoreSenador(ctx context.Context, senadorID int, ano *int) (*SenadorScore, error) {
	// Reutilizar o calculo do ranking completo para garantir consistencia da posicao
	// Como o ranking tem cache, isso e eficiente
	ranking, err := s.CalcularRanking(ctx, ano)
	if err != nil {
		return nil, err
	}

	// Buscar o senador no ranking ou entre os sem dados
	for _, lista := range [][]SenadorScore{ranking.Ranking, ranking.SemDados} {
		for _, score := range lista {
			if score.SenadorID == senadorID {
				return &score, nil
			}
		}
	}

	return nil, fmt.Errorf("senador nao encontrado no ranking")
}


// dadosBrutosSenador armazena dados brutos antes da normalizacao
type dadosBrutosSenador struct {
	// Proposicoes
	totalProposicoes     int
	totalCoautorias      int
	proposicoesAprovadas int
	transformadasEmLei   int
	pontuacaoProposicoes float64

	// Votacoes
	votacoes *votacao.VotacaoStats // nil se a consulta falhou

	// CEAPS
	gastoAnual float64

	// Comissoes
	comissoesAtivas   int
	comissoesTitular  int
	comissoesSuplente int
	pontosComissoes   float64
}

// coletarDadosBrutos busca dados de todos os modulos para um senador
func (s *Service) coletarDadosBrutos(senadorID int, ano *int) *dadosBrutosSenador {
	dados := &dadosBrutosSenador{}

	// Proposicoes
	var propStats *proposicao.ProposicaoStats
	var err error
	if ano != nil {
		propStats, err = s.proposicaoRepo.GetStatsByAno(senadorID, *ano)
	} else {
		propStats, err = s.proposicaoRepo.GetStats(senadorID)
	}

	if err == nil {
		dados.totalProposicoes = propStats.TotalProposicoes
		dados.totalCoautorias = propStats.TotalCoautorias
		dados.proposicoesAprovadas = propStats.AprovadosPlenario
		dados.transformadasEmLei = propStats.TransformadasEmLei
		dados.pontuacaoProposicoes = propStats.PontuacaoTotal
	}

	// Votacoes
	var votStats *votacao.VotacaoStats
	if ano != nil {
		votStats, err = s.votacaoRepo.GetStatsByAno(senadorID, *ano)
	} else {
		votStats, err = s.votacaoRepo.GetStats(senadorID)
	}

	if err == nil {
		dados.votacoes = votStats
	} else {
		slog.Error("falha ao buscar stats de votacao", "senador", senadorID, "error", err)
	}

	// CEAPS
	if ano != nil {
		if gastoAnual, err := s.ceapsRepo.GetTotalByAno(senadorID, *ano); err == nil {
			dados.gastoAnual = gastoAnual
		}
	} else {
		// Mandato: Soma de todos os anos
		gastoTotal, err := s.ceapsRepo.GetTotal(senadorID)
		if err == nil {
			dados.gastoAnual = gastoTotal
		}
	}

	// Comissoes
	var comStats *comissao.ComissaoStats
	if ano != nil {
		comStats, err = s.comissaoRepo.GetStatsByAno(senadorID, *ano)
	} else {
		comStats, err = s.comissaoRepo.GetStats(senadorID)
	}

	if err == nil {
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
	ano *int,
) SenadorScore {
	// Normalizar Produtividade (0-100) com Logaritmo para suavizar outliers
	produtividade := (math.Log1p(dados.pontuacaoProposicoes) / math.Log1p(maxPontuacaoProd)) * 100

	// Presenca: metrica B (ajustada), 0-100. Sem registro que conte no
	// periodo nao ha presenca a medir: fica nula e o senador sai da ordenacao
	// em vez de levar 0 (item 4).
	var presenca *float64
	var presencaNoScore float64
	vs := dados.votacoes
	if vs == nil {
		vs = &votacao.VotacaoStats{}
	}
	if vs.DadosSuficientes {
		p := arredondar(vs.PresencaAjustada)
		presenca = &p
		presencaNoScore = vs.PresencaAjustada
	}

	// Economia CEAPS (0-100)
	// Quanto menos gasta, maior o score
	// Buscar teto da UF, se nao houver usa media
	tetoMensal, ok := TetoCEAPSPorUF[sen.UF]
	if !ok {
		tetoMensal = 40000.0 // Fallback seguro
	}
	var tetoPeriodo float64
	
	if ano != nil {
		// Teto anual para um ano especifico
		tetoPeriodo = tetoMensal * 12
	} else {
		// Teto acumulado do mandato (desde Fev/2023 ate agora)
		inicioMandato := time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)
		mesesCorridos := time.Since(inicioMandato).Hours() / 24 / 30
		if mesesCorridos < 1 {
			mesesCorridos = 1
		}
		tetoPeriodo = tetoMensal * mesesCorridos
	}

	economia := (1 - (dados.gastoAnual / tetoPeriodo)) * 100
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
		(presencaNoScore * PesoPresenca) +
		(economia * PesoEconomia) +
		(comissoes * PesoComissoes)

	insuficiente := presenca == nil
	if insuficiente {
		scoreFinal = 0
	}

	return SenadorScore{
		SenadorID:     sen.ID,
		Nome:          sen.Nome,
		Partido:       sen.Partido,
		UF:            sen.UF,
		FotoURL:       sen.FotoURL,
		Cargo:         sen.Cargo,
		Titular:       sen.Titular,
		Produtividade: arredondar(produtividade),
		Presenca:      presenca,
		EconomiaCota:  arredondar(economia),
		Comissoes:     arredondar(comissoes),
		ScoreFinal:    arredondar(scoreFinal),
		DadosInsuficientes: insuficiente,
		CalculadoEm:   time.Now(),
		Detalhes: ScoreDetalhes{
			TotalProposicoes:     dados.totalProposicoes,
			TotalCoautorias:      dados.totalCoautorias,
			ProposicoesAprovadas: dados.proposicoesAprovadas,
			TransformadasEmLei:   dados.transformadasEmLei,
			PontuacaoProposicoes: dados.pontuacaoProposicoes,
			TotalVotacoes:         vs.TotalVotacoes,
			VotacoesParticipadas:  vs.VotosRegistrados,
			Presentes:             vs.Presentes,
			AusenciasAP:           vs.AusenciasAP,
			NaoCompareceu:         vs.NaoCompareceu,
			AusenciasJustificadas: vs.AusenciasJustificadas,
			TaxaPresencaBruta:     arredondar(vs.PresencaBruta),
			TaxaPresencaAjustada:  arredondar(vs.PresencaAjustada),
			GastoCEAPS:           arredondar(dados.gastoAnual),
			TetoCEAPS:            tetoPeriodo,
			ComissoesAtivas:      dados.comissoesAtivas,
			ComissoesTitular:     dados.comissoesTitular,
			ComissoesSuplente:    dados.comissoesSuplente,
			PontosComissoes:      arredondar(dados.pontosComissoes),
		},
	}
}

// ordenar ordena por score final decrescente, com desempate por nome para o
// resultado nao depender da ordem de leitura, e atribui as posicoes
func ordenar(scores []SenadorScore) {
	sort.SliceStable(scores, func(i, j int) bool {
		if scores[i].ScoreFinal != scores[j].ScoreFinal {
			return scores[i].ScoreFinal > scores[j].ScoreFinal
		}
		return scores[i].Nome < scores[j].Nome
	})
	for i := range scores {
		scores[i].Posicao = i + 1
	}
}

// arredondar arredonda para 2 casas decimais
func arredondar(valor float64) float64 {
	return float64(int(valor*100+0.5)) / 100
}
