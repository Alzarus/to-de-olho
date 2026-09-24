package proposicao

import (
	"testing"

	"github.com/Alzarus/to-de-olho/internal/testdb"
)

func principal(sigla, estagio string) Proposicao {
	um := 1
	return Proposicao{SiglaSubtipoMateria: sigla, EstagioTramitacao: estagio, PosicaoAutoria: &um, TipoAutor: "SENADOR"}
}

func TestCalcularPontuacaoPorSigla(t *testing.T) {
	casos := []struct {
		sigla, estagio string
		pontos         float64
	}{
		{"PEC", "Apresentado", 3},
		{"PEC", "TransformadoLei", 48},
		{"PLP", "AprovadoPlenario", 16},
		{"PL", "EmComissao", 2},
		{"PLS", "Apresentado", 1},
		{"PDL", "AprovadoComissao", 4},
		{"PRS", "Apresentado", 1},
		{"RQS", "Apresentado", 0.5},
		{"PFS", "Apresentado", 0.5},
		{"REQ", "Apresentado", 0.1},
		{"INS", "Apresentado", 0.1}, // antes da v2.2 valia 1, como um PL
		{"RDH", "Apresentado", 0.1},
		{"ECD", "TransformadoLei", 0}, // emenda da Camara: a materia ja pontua pelo projeto original
		{"OFS", "Apresentado", 0},
		{"XYZ", "Apresentado", 0}, // sigla fora da tabela nao pontua
	}
	for _, c := range casos {
		p := principal(c.sigla, c.estagio)
		if got := p.CalcularPontuacao(); got != c.pontos {
			t.Errorf("%s %s: %v pontos, esperado %v", c.sigla, c.estagio, got, c.pontos)
		}
	}
}

// Todas as siglas com autoria de senador em producao (23/09/2026) tem peso
// decidido. Sigla nova cai no log do sync e vale zero ate entrar na tabela.
func TestSiglasDeProducaoTemPeso(t *testing.T) {
	for _, s := range []string{
		"RQS", "REQ", "PL", "PLS", "RDH", "INS", "PRS", "PEC", "PDL", "PLP", "RQN", "RAS", "RCE",
		"R.S", "RQJ", "RMA", "RCT", "RQE", "RQI", "RDR", "PDS", "PFS", "RRA", "RRE", "OFS", "RFF",
		"PCE", "SCD", "PET", "RTG", "ATS", "PRN", "R.C", "ECD", "DEN", "CON", "SIN", "DIV", "MSG",
		"OFN", "RQR", "PRM", "MOC",
	} {
		if _, ok := PesoTipo(s); !ok {
			t.Errorf("sigla %s sem peso definido", s)
		}
	}
}

func TestRecalcularPontuacaoIdempotente(t *testing.T) {
	db := testdb.Abrir(t, &Proposicao{})
	um, dois := 1, 2
	linhas := []Proposicao{
		{SenadorID: 1, CodigoMateria: "1", SiglaSubtipoMateria: "INS", EstagioTramitacao: "Apresentado", PosicaoAutoria: &um, TipoAutor: "SENADOR", Pontuacao: 1},  // regra antiga
		{SenadorID: 1, CodigoMateria: "2", SiglaSubtipoMateria: "PEC", EstagioTramitacao: "Apresentado", PosicaoAutoria: &um, TipoAutor: "SENADOR", Pontuacao: 3},  // ja certa
		{SenadorID: 1, CodigoMateria: "3", SiglaSubtipoMateria: "PL", EstagioTramitacao: "Apresentado", PosicaoAutoria: &dois, TipoAutor: "SENADOR", Pontuacao: 0}, // coautoria
	}
	if err := db.Create(&linhas).Error; err != nil {
		t.Fatal(err)
	}

	n, err := RecalcularPontuacao(db)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("primeira execucao: %d linhas alteradas, esperado 1", n)
	}
	var ins Proposicao
	db.Where("codigo_materia = ?", "1").First(&ins)
	if ins.Pontuacao != 0.1 {
		t.Errorf("INS deveria valer 0,1; vale %v", ins.Pontuacao)
	}
	if n, _ = RecalcularPontuacao(db); n != 0 {
		t.Errorf("segunda execucao alterou %d linhas; deveria ser 0", n)
	}
}
