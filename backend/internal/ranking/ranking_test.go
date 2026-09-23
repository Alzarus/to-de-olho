package ranking

import (
	"testing"

	"github.com/Alzarus/to-de-olho/internal/comissao"
	"github.com/Alzarus/to-de-olho/internal/proposicao"
	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/internal/votacao"
)

// TestArredondar verifica se a funcao auxiliar de arredondamento funciona como esperado
func TestArredondar(t *testing.T) {
	testes := []struct {
		entrada  float64
		esperado float64
	}{
		{10.555, 10.56},
		{10.554, 10.55},
		{99.999, 100.00},
		{0.0, 0.0},
		{33.3333, 33.33},
	}

	for _, teste := range testes {
		resultado := arredondar(teste.entrada)
		if resultado != teste.esperado {
			t.Errorf("arredondar(%f) = %f; esperado %f", teste.entrada, resultado, teste.esperado)
		}
	}
}

// TestCalculoPesosSimples verifica se a logica basica dos pesos esta correta
// Este é um "white-box test" simplificado para validar a metodologia
func TestCalculoPesosSimples(t *testing.T) {
	// Ponderacao oficial
	// (Produtividade * 0.35) + (Presenca * 0.25) + (Economia * 0.20) + (Comissoes * 0.20)

	produtividade := 100.0
	presenca := 100.0
	economia := 100.0
	comissoes := 100.0

	scoreFinal := (produtividade * 0.35) +
		(presenca * 0.25) +
		(economia * 0.20) +
		(comissoes * 0.20)

	if scoreFinal != 100.0 {
		t.Errorf("Score maximo esperado 100.0, obtido %f", scoreFinal)
	}

	// Cenaria onde so tem presenca
	produtividade = 0
	presenca = 100
	economia = 0
	comissoes = 0

	scoreFinal = (produtividade * 0.35) +
		(presenca * 0.25) +
		(economia * 0.20) +
		(comissoes * 0.20)

	if scoreFinal != 25.0 {
		t.Errorf("Score apenas com presenca esperado 25.0, obtido %f", scoreFinal)
	}
}

func dados(meses float64, vot *votacao.VotacaoStats) *dadosBrutosSenador {
	return &dadosBrutosSenador{
		prop: &proposicao.ProposicaoStats{}, vot: vot, com: &comissao.ComissaoStats{}, meses: meses,
	}
}

func TestDadosInsuficientesFicamForaDaOrdenacao(t *testing.T) {
	s := &Service{}
	sen := senador.Senador{ID: 1, Nome: "Fulano", UF: "PI"}
	cheio := &votacao.VotacaoStats{DadosSuficientes: true, PresencaAjustada: 100, PresencaBruta: 96.63}

	casos := []struct {
		nome   string
		d      *dadosBrutosSenador
		motivo string
	}{
		{"sem votos que contem", dados(40, &votacao.VotacaoStats{}), "sem registro de votacao no periodo"},
		{"menos de 6 meses em exercicio", dados(1.4, cheio), "menos de 6 meses em exercicio no periodo"},
		{"consulta falhou", &dadosBrutosSenador{}, "falha ao consultar os dados"},
	}
	for _, c := range casos {
		sc := s.calcularScoreNormalizado(sen, c.d, 1, 1)
		if !sc.DadosInsuficientes || sc.Motivo != c.motivo || sc.ScoreFinal != 0 {
			t.Errorf("%s: %+v", c.nome, sc)
		}
	}
	semVoto := s.calcularScoreNormalizado(sen, casos[0].d, 1, 1)
	if semVoto.Presenca != nil {
		t.Error("sem votos, presenca tem de ser nula, nao zero")
	}

	com := s.calcularScoreNormalizado(sen, dados(40, cheio), 1, 1)
	if com.DadosInsuficientes || com.Presenca == nil || *com.Presenca != 100 || com.Detalhes.TaxaPresencaBruta != 96.63 {
		t.Errorf("presenca B deveria entrar no score e A nos detalhes: %+v", com)
	}
}

func TestTetoProporcionalAosMesesEmExercicio(t *testing.T) {
	s := &Service{}
	sen := senador.Senador{ID: 1, Nome: "Fulano", UF: "PI"} // teto mensal 49.000
	vot := &votacao.VotacaoStats{DadosSuficientes: true, PresencaAjustada: 90}

	// 10 meses em exercicio, gastou 245 mil = metade do teto do periodo
	d := dados(10, vot)
	d.gasto = 245000
	sc := s.calcularScoreNormalizado(sen, d, 1, 1)
	if sc.EconomiaCota != 50 || sc.Detalhes.TetoCEAPS != 490000 || sc.Detalhes.MesesExercicio != 10 {
		t.Errorf("economia com teto proporcional: %+v", sc.Detalhes)
	}

	// o mesmo gasto em 44 meses no mandato inteiro daria economia perto de 90:
	// o criterio antigo media a data de posse, nao a economia (item 8)
	d.meses = 44
	if sc := s.calcularScoreNormalizado(sen, d, 1, 1); sc.EconomiaCota < 88 {
		t.Errorf("com 44 meses: %v", sc.EconomiaCota)
	}
}

func TestOrdenarDesempataPorNome(t *testing.T) {
	scores := []SenadorScore{{Nome: "Zeca", ScoreFinal: 50}, {Nome: "Ana", ScoreFinal: 50}, {Nome: "Bia", ScoreFinal: 70}}
	ordenar(scores)
	if scores[0].Nome != "Bia" || scores[1].Nome != "Ana" || scores[2].Nome != "Zeca" || scores[2].Posicao != 3 {
		t.Errorf("ordem errada: %+v", scores)
	}
}
