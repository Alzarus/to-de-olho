package ranking

import (
	"testing"

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

func TestSemVotosFicaForaDaOrdenacao(t *testing.T) {
	s := &Service{}
	sen := senador.Senador{ID: 1, Nome: "Fulano", UF: "PI"}

	// consulta sem registro que conte: presenca nula, nao zero
	semDados := s.calcularScoreNormalizado(sen, &dadosBrutosSenador{votacoes: &votacao.VotacaoStats{}}, 1, 1, nil)
	if !semDados.DadosInsuficientes || semDados.Presenca != nil || semDados.ScoreFinal != 0 {
		t.Errorf("sem votos deveria ficar sem presenca e fora da ordenacao: %+v", semDados)
	}
	falhou := s.calcularScoreNormalizado(sen, &dadosBrutosSenador{}, 1, 1, nil)
	if !falhou.DadosInsuficientes {
		t.Error("stats que falharam (nil) tambem sao dados insuficientes")
	}

	com := s.calcularScoreNormalizado(sen, &dadosBrutosSenador{votacoes: &votacao.VotacaoStats{
		DadosSuficientes: true, PresencaAjustada: 100, PresencaBruta: 96.63, AusenciasJustificadas: 14,
	}}, 1, 1, nil)
	if com.DadosInsuficientes || com.Presenca == nil || *com.Presenca != 100 || com.Detalhes.TaxaPresencaBruta != 96.63 {
		t.Errorf("presenca B deveria entrar no score e A nos detalhes: %+v", com)
	}
	if com.ScoreFinal < 25 {
		t.Errorf("presenca 100 vale 25 pontos no score, obtido %v", com.ScoreFinal)
	}
}

func TestOrdenarDesempataPorNome(t *testing.T) {
	scores := []SenadorScore{{Nome: "Zeca", ScoreFinal: 50}, {Nome: "Ana", ScoreFinal: 50}, {Nome: "Bia", ScoreFinal: 70}}
	ordenar(scores)
	if scores[0].Nome != "Bia" || scores[1].Nome != "Ana" || scores[2].Nome != "Zeca" || scores[2].Posicao != 3 {
		t.Errorf("ordem errada: %+v", scores)
	}
}
