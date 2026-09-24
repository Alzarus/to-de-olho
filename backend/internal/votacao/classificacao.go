package votacao

import "log/slog"

// Classe e o papel de um codigo de voto no calculo de presenca (item 2).
type Classe int

const (
	// NaoConta sai do numerador e do denominador ("NA": dispositivo nao citado)
	NaoConta Classe = iota
	// Presente estava no plenario, tenha votado ou nao
	Presente
	// Justificada e ausencia com licenca ou missao: sai do denominador na metrica B
	Justificada
	// Ausente conta como falta: "AP" (atividade parlamentar, declarada pelo
	// proprio senador), "LP" (licenca particular) e "NCom" (nao compareceu)
	Ausente
)

// classes e o dicionario completo dos codigos que a API devolve em
// siglaVotoParlamentar. Um codigo fora daqui e logado e nao conta.
var classes = map[string]Classe{
	"Sim":                       Presente,
	"Não":                       Presente,
	"Abstenção":                 Presente,
	"Votou":                     Presente, // votacao secreta
	"P-NRV":                     Presente, // presente, nao registrou voto
	"Presidente (art. 51 RISF)": Presente,

	// Obstrucao fica fora da conta (metodologia v2.1): o TCC diz que nao conta
	// como presenca, e ela tampouco e falta, por ser estrategia regimental.
	"P-OD":      NaoConta, // presente, obstrucao declarada
	"Obstrução": NaoConta,

	"LS":  Justificada, // licenca saude
	"LAP": Justificada, // licenca para atividade parlamentar
	"MIS": Justificada, // missao oficial
	"LC":  Justificada,
	"LG":  Justificada, // licenca gestante
	"REP": Justificada, // representacao da Casa
	"LAN": Justificada,

	// LP conta como falta (metodologia v2.1): o TCC so justifica licenca
	// medica e missao oficial; licenca particular e escolha do senador, como AP.
	"AP":   Ausente,
	"LP":   Ausente,
	"NCom": Ausente,

	"NA": NaoConta,
}

// Classificar devolve a classe de um codigo de voto da API.
func Classificar(sigla string) Classe {
	c, ok := classes[sigla]
	if !ok {
		slog.Warn("codigo de voto desconhecido, fora do calculo de presenca", "sigla", sigla)
		return NaoConta
	}
	return c
}

// calcularStats monta as estatisticas a partir da contagem de votos por sigla.
//
//	A (bruta)    = presentes / (total - NA)
//	B (ajustada) = presentes / (total - NA - justificadas)   <- taxa_presenca, vai para o ranking
//
// B e a metrica descrita no TCC (decisao D3). Sem nenhum registro no
// denominador, DadosSuficientes = false: o ranking nao pode tratar isso como 0.
func calcularStats(senadorID int, porSigla map[string]int) *VotacaoStats {
	st := &VotacaoStats{SenadorID: senadorID}
	var presentes, justificadas, ausentes int
	for sigla, n := range porSigla {
		st.TotalVotacoes += n
		switch Classificar(sigla) {
		case Presente:
			presentes += n
		case Justificada:
			justificadas += n
		case Ausente:
			ausentes += n
		}
		switch sigla {
		case "Sim", "Não", "Abstenção":
			st.VotosRegistrados += n
		case "Obstrução", "P-OD":
			st.Obstrucoes += n
		case "AP":
			st.AusenciasAP += n
		case "NCom":
			st.NaoCompareceu += n
		}
	}
	st.Presentes = presentes
	st.AusenciasJustificadas = justificadas
	st.Ausencias = ausentes

	if base := presentes + justificadas + ausentes; base > 0 {
		st.PresencaBruta = pct(presentes, base)
	}
	if base := presentes + ausentes; base > 0 {
		st.PresencaAjustada = pct(presentes, base)
		st.TaxaParticipacao = pct(st.VotosRegistrados, base)
		st.DadosSuficientes = true
	}
	st.TaxaPresenca = st.PresencaAjustada
	return st
}

func pct(parte, todo int) float64 {
	return float64(parte) / float64(todo) * 100
}
