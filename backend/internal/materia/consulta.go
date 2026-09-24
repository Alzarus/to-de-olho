package materia

import "fmt"

// Fragmentos SQL para as listas de votacoes e proposicoes trazerem o nome
// popular. O apelido oficial tem prioridade; o da curadoria so aparece quando
// o Senado nao atribui nenhum. Os aliases mat/matc nao colidem com as colunas
// das tabelas que fazem o JOIN.

// ColunasSelect sao as colunas lidas pelo LEFT JOIN de Join.
const ColunasSelect = `COALESCE(mat.apelido, matc.apelido) AS apelido,
	CASE WHEN mat.apelido IS NOT NULL THEN 'oficial'
	     WHEN matc.apelido IS NOT NULL THEN 'curadoria' END AS apelido_fonte,
	CASE WHEN mat.apelido IS NULL THEN matc.fonte_url END AS apelido_fonte_url,
	mat.explicacao_ementa AS explicacao_ementa,
	mat.temas AS temas`

// Join devolve os LEFT JOIN de materias e da curadoria pela expressao que da
// o codigo da materia (inteiro), por exemplo "v.codigo_materia".
func Join(expr string) string {
	return fmt.Sprintf(`LEFT JOIN materias mat ON mat.codigo_materia = %[1]s
	LEFT JOIN materias_apelidos_curados matc ON matc.codigo_materia = %[1]s`, expr)
}

// CondicaoApelido filtra por apelido (oficial ou curado) com ILIKE, sem JOIN:
// "<expr> IN (...)". Recebe o mesmo padrao duas vezes.
func CondicaoApelido(expr string) string {
	return expr + ` IN (SELECT codigo_materia FROM materias WHERE apelido ILIKE ?
		UNION SELECT codigo_materia FROM materias_apelidos_curados WHERE apelido ILIKE ?)`
}

// CodigoTextoParaInt converte uma coluna texto de codigo (proposicoes guarda
// codigo_materia como string) em inteiro, nulo se nao for numerica.
func CodigoTextoParaInt(coluna string) string {
	return fmt.Sprintf("(CASE WHEN %[1]s ~ '^[0-9]{1,9}$' THEN %[1]s::int END)", coluna)
}
