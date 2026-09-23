package ranking

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"

	"github.com/Alzarus/to-de-olho/internal/ceaps"
	"github.com/Alzarus/to-de-olho/internal/comissao"
	"github.com/Alzarus/to-de-olho/internal/proposicao"
	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/internal/utils"
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

// CalcularRanking calcula o ranking de todos os senadores em exercicio.
//
// Todas as fontes usam o mesmo periodo: o mandato (desde a posse da
// legislatura) ou um ano-calendario cortado pelo recorte e por hoje.
func (s *Service) CalcularRanking(ctx context.Context, ano *int) (*RankingResponse, error) {
	cacheKey := "ranking:v3:geral"
	if ano != nil {
		cacheKey = fmt.Sprintf("ranking:v3:%d", *ano)
	}
	if cached := localCache.Get(cacheKey); cached != nil {
		slog.Info("ranking retornado do cache local (RAM)", "key", cacheKey)
		return cached, nil
	}

	slog.Info("iniciando calculo de ranking (cache miss)", "ano", ano)

	senadores, err := s.senadorRepo.FindAll(false)
	if err != nil {
		return nil, err
	}

	inicio, fim := periodo(ano)
	dadosBrutos := make(map[int]*dadosBrutosSenador, len(senadores))
	for _, sen := range senadores {
		dadosBrutos[sen.ID] = s.coletarDadosBrutos(sen.ID, inicio, fim)
	}

	// A escala (maior pontuacao da Casa) sai so de quem entra na ordenacao
	maxPontuacaoProd, maxPontosComissoes := 1.0, 1.0
	for _, d := range dadosBrutos {
		if motivoInsuficiente(d) != "" {
			continue
		}
		maxPontuacaoProd = math.Max(maxPontuacaoProd, d.prop.PontuacaoTotal)
		maxPontosComissoes = math.Max(maxPontosComissoes, float64(d.com.Pontos))
	}

	var scores, semDados []SenadorScore
	for _, sen := range senadores {
		score := s.calcularScoreNormalizado(sen, dadosBrutos[sen.ID], maxPontuacaoProd, maxPontosComissoes)
		if score.DadosInsuficientes {
			semDados = append(semDados, score)
			continue
		}
		scores = append(scores, score)
	}
	ordenar(scores)
	sort.Slice(semDados, func(i, j int) bool { return semDados[i].Nome < semDados[j].Nome })

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
	localCache.Set(cacheKey, response, 24*time.Hour)
	return response, nil
}

// periodo do ranking: mandato ou ano (ver utils.PeriodoDoAno)
func periodo(ano *int) (inicio, fim time.Time) {
	if ano != nil {
		return utils.PeriodoDoAno(*ano)
	}
	return utils.PeriodoDoMandato()
}

// InvalidateCache invalida todo o cache de ranking
func (s *Service) InvalidateCache() {
	slog.Info("invalidando cache de ranking")
	localCache.InvalidateAll()
}

// CalcularScoreSenador calcula o score de um senador especifico
func (s *Service) CalcularScoreSenador(ctx context.Context, senadorID int, ano *int) (*SenadorScore, error) {
	// Reutiliza o ranking completo (em cache) para a posicao ser consistente
	ranking, err := s.CalcularRanking(ctx, ano)
	if err != nil {
		return nil, err
	}
	for _, lista := range [][]SenadorScore{ranking.Ranking, ranking.SemDados} {
		for _, score := range lista {
			if score.SenadorID == senadorID {
				return &score, nil
			}
		}
	}
	return nil, fmt.Errorf("senador nao encontrado no ranking")
}

// dadosBrutosSenador guarda o que cada fonte devolveu para o periodo
type dadosBrutosSenador struct {
	prop  *proposicao.ProposicaoStats
	vot   *votacao.VotacaoStats
	com   *comissao.ComissaoStats
	gasto float64 // CEAPS no periodo
	meses float64 // meses em exercicio no periodo
	erro  error   // alguma consulta falhou
}

// coletarDadosBrutos busca dados de todos os modulos para um senador
func (s *Service) coletarDadosBrutos(senadorID int, inicio, fim time.Time) *dadosBrutosSenador {
	d := &dadosBrutosSenador{}
	var erros []error
	var err error
	if d.prop, err = s.proposicaoRepo.GetStatsPeriodo(senadorID, inicio, fim); err != nil {
		erros = append(erros, err)
	}
	if d.vot, err = s.votacaoRepo.GetStatsPeriodo(senadorID, inicio, fim); err != nil {
		erros = append(erros, err)
	}
	if d.com, err = s.comissaoRepo.GetStatsPeriodo(senadorID, inicio, fim); err != nil {
		erros = append(erros, err)
	}
	if d.gasto, err = s.ceapsRepo.GetTotalPeriodo(senadorID, inicio, fim); err != nil {
		erros = append(erros, err)
	}
	if d.meses, err = s.senadorRepo.MesesEmExercicio(senadorID, inicio, fim); err != nil {
		erros = append(erros, err)
	}
	if len(erros) > 0 {
		d.erro = errors.Join(erros...)
		slog.Error("falha ao coletar dados do ranking", "senador", senadorID, "error", d.erro)
	}
	return d
}

// motivoInsuficiente devolve por que o senador fica fora da ordenacao, ou ""
func motivoInsuficiente(d *dadosBrutosSenador) string {
	switch {
	case d.erro != nil || d.prop == nil || d.vot == nil || d.com == nil:
		return "falha ao consultar os dados"
	case d.meses < PisoMesesExercicio:
		return fmt.Sprintf("menos de %.0f meses em exercicio no periodo", PisoMesesExercicio)
	case !d.vot.DadosSuficientes:
		return "sem registro de votacao no periodo"
	}
	return ""
}

// calcularScoreNormalizado calcula o score final normalizado
func (s *Service) calcularScoreNormalizado(
	sen senador.Senador,
	d *dadosBrutosSenador,
	maxPontuacaoProd float64,
	maxPontosComissoes float64,
) SenadorScore {
	prop, vot, com := d.prop, d.vot, d.com
	if prop == nil {
		prop = &proposicao.ProposicaoStats{}
	}
	if vot == nil {
		vot = &votacao.VotacaoStats{}
	}
	if com == nil {
		com = &comissao.ComissaoStats{}
	}

	// Produtividade (0-100), escala logaritmica para suavizar outliers
	produtividade := math.Min(100, math.Log1p(prop.PontuacaoTotal)/math.Log1p(maxPontuacaoProd)*100)

	// Presenca: metrica B (ajustada). Sem registro que conte no periodo nao
	// ha presenca a medir: fica nula (item 4)
	var presenca *float64
	if vot.DadosSuficientes {
		p := arredondar(vot.PresencaAjustada)
		presenca = &p
	}

	// Economia (0-100): teto proporcional aos meses em exercicio (item 8)
	tetoMensal, ok := TetoCEAPSPorUF[sen.UF]
	if !ok {
		tetoMensal = TetoCEAPSMedia / 12
	}
	tetoPeriodo := tetoMensal * d.meses
	economia := 0.0
	if tetoPeriodo > 0 {
		economia = math.Max(0, math.Min(100, (1-d.gasto/tetoPeriodo)*100))
	}

	// Comissoes (0-100)
	comissoes := math.Min(100, float64(com.Pontos)/maxPontosComissoes*100)

	motivo := motivoInsuficiente(d)
	scoreFinal := 0.0
	if motivo == "" {
		scoreFinal = produtividade*PesoProdutividade + vot.PresencaAjustada*PesoPresenca +
			economia*PesoEconomia + comissoes*PesoComissoes
	}

	return SenadorScore{
		SenadorID:          sen.ID,
		Nome:               sen.Nome,
		Partido:            sen.Partido,
		UF:                 sen.UF,
		FotoURL:            sen.FotoURL,
		Cargo:              sen.Cargo,
		Titular:            sen.Titular,
		Produtividade:      arredondar(produtividade),
		Presenca:           presenca,
		EconomiaCota:       arredondar(economia),
		Comissoes:          arredondar(comissoes),
		ScoreFinal:         arredondar(scoreFinal),
		DadosInsuficientes: motivo != "",
		Motivo:             motivo,
		CalculadoEm:        time.Now(),
		Detalhes: ScoreDetalhes{
			TotalProposicoes:      prop.TotalProposicoes,
			TotalCoautorias:       prop.TotalCoautorias,
			TotalSemPontos:        prop.TotalSemPontos,
			ProposicoesAprovadas:  prop.AprovadosPlenario,
			TransformadasEmLei:    prop.TransformadasEmLei,
			PontuacaoProposicoes:  prop.PontuacaoTotal,
			TotalVotacoes:         vot.TotalVotacoes,
			VotacoesParticipadas:  vot.VotosRegistrados,
			Presentes:             vot.Presentes,
			AusenciasAP:           vot.AusenciasAP,
			NaoCompareceu:         vot.NaoCompareceu,
			AusenciasJustificadas: vot.AusenciasJustificadas,
			TaxaPresencaBruta:     arredondar(vot.PresencaBruta),
			TaxaPresencaAjustada:  arredondar(vot.PresencaAjustada),
			GastoCEAPS:            arredondar(d.gasto),
			TetoCEAPS:             arredondar(tetoPeriodo),
			MesesExercicio:        arredondar(d.meses),
			ComissoesAtivas:       com.ComissoesAtivas,
			ComissoesTitular:      com.ComissoesTitular,
			ComissoesSuplente:     com.ComissoesSuplente,
			ComissoesForaDaConta:  com.ForaDaConta,
			PontosComissoes:       float64(com.Pontos),
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
