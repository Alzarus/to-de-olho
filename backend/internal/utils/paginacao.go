package utils

import "strconv"

// LimiteMaximoPadrao é o maior limit aceito nas listas paginadas da API. O
// front nunca pede mais que isso (a exportação percorre páginas de 100), e
// sem teto um único ?limit=100000 obrigava o banco a montar e serializar a
// tabela inteira numa resposta só.
const LimiteMaximoPadrao = 100

// OffsetMaximo limita a profundidade da paginação. OFFSET alto faz o Postgres
// ler e descartar todas as linhas anteriores; sem teto, ?page=99999999 custava
// uma varredura completa e, perto do limite do int, (page-1)*limit estourava e
// virava OFFSET negativo (erro 500). 50 mil cobre a exportação do front (até
// 500 páginas de 100) com folga para as listas atuais.
const OffsetMaximo = 50_000

// Paginacao é o resultado já validado de ?limit= e ?page=.
type Paginacao struct {
	Limit  int
	Page   int
	Offset int
}

// LerPaginacao interpreta limit e page vindos da query string.
//
// Valor ausente, não numérico ou menor que 1 vira o padrão (limit) ou 1
// (page). limit acima de maximo é truncado para maximo, em vez de voltar ao
// padrão, para quem pede "o máximo possível" receber o máximo permitido. page
// é truncada para que o offset não passe de OffsetMaximo: além disso a
// resposta seria vazia de qualquer forma nas listas atuais.
func LerPaginacao(limitStr, pageStr string, padrao, maximo int) Paginacao {
	limit := padrao
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = min(l, maximo)
	}

	page := 1
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		// Compara antes de multiplicar: page enorme não pode estourar o int.
		page = min(p, OffsetMaximo/limit+1)
	}

	return Paginacao{Limit: limit, Page: page, Offset: (page - 1) * limit}
}
