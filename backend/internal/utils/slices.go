package utils

// NaoNulo devolve s, ou uma lista vazia quando s e nil. Um slice nil vira
// null no JSON, e o frontend espera [] quando nao ha dados.
func NaoNulo[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
