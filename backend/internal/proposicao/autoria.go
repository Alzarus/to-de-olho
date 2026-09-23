package proposicao

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Posicao do senador na autoria de uma materia (item 9 da auditoria: so o
// primeiro autor pontua).
//
// A listagem /processo?codigoParlamentarAutor= traz a autoria so como texto:
//
//	"Senador Izalci Lucas (PL/DF), Senadora Damares Alves (REPUBLICANOS/DF)"
//	"Senador Magno Malta (PL/ES) e outros."
//
// PosicaoNoTexto extrai a posicao desse texto. Quando o texto nao tem a forma
// esperada, ou o nome nao aparece exatamente uma vez, devolve ok=false e o
// sync consulta o detalhe (/processo/{id}, campo autoriaIniciativa[].ordem).

// entradaAutor casa "Senador Fulano (PARTIDO/UF)" seguido de virgula ou fim.
var entradaAutor = regexp.MustCompile(`^\s*([^,()]+?)\s*\([^()]*\)\s*(?:,|$)`)

var sufixoOutros = regexp.MustCompile(`(?i)\s+e\s+outros\.?\s*$`)

// PosicaoNoTexto devolve a posicao (1 = primeiro autor) de nomeParlamentar no
// texto de autoria e o total de autores. total = 0 quando o texto termina em
// "e outros" (total desconhecido).
func PosicaoNoTexto(autoria, nomeParlamentar string) (posicao, total int, ok bool) {
	texto := strings.TrimSpace(autoria)
	eOutros := false
	if loc := sufixoOutros.FindStringIndex(texto); loc != nil {
		texto = texto[:loc[0]]
		eOutros = true
	}
	texto = strings.TrimSuffix(strings.TrimSpace(texto), ".")
	if texto == "" {
		return 0, 0, false
	}

	var nomes []string
	for resto := texto; strings.TrimSpace(resto) != ""; {
		m := entradaAutor.FindStringSubmatchIndex(resto)
		if m == nil {
			return 0, 0, false // estrutura nao reconhecida (ex.: "Comissao de Constituicao, Justica e ...")
		}
		nomes = append(nomes, normalizarNome(resto[m[2]:m[3]]))
		resto = resto[m[1]:]
	}

	alvo := normalizarNome(nomeParlamentar)
	for i, nome := range nomes {
		if nome != alvo {
			continue
		}
		if posicao != 0 {
			return 0, 0, false // nome repetido: ambiguo
		}
		posicao = i + 1
	}
	if posicao == 0 {
		return 0, 0, false
	}
	if !eOutros {
		total = len(nomes)
	}
	return posicao, total, true
}

// normalizarNome tira acento, caixa, espacos repetidos e o tratamento
// (Senador, Senadora, Deputado, Deputada).
func normalizarNome(nome string) string {
	var b strings.Builder
	for _, r := range norm.NFD.String(nome) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(unicode.ToLower(r))
	}
	palavras := strings.Fields(b.String())
	if len(palavras) > 1 {
		switch palavras[0] {
		case "senador", "senadora", "deputado", "deputada":
			palavras = palavras[1:]
		}
	}
	return strings.Join(palavras, " ")
}

// MesmoNome compara dois nomes com a mesma normalizacao de PosicaoNoTexto.
func MesmoNome(a, b string) bool {
	return normalizarNome(a) == normalizarNome(b)
}
