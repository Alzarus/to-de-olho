package envutils

import (
	"strconv"
	"strings"
)

// IsEnabled retorna o valor booleano de uma feature flag baseada em variável de ambiente.
// Quando a variável não está definida, retorna o valor padrão informado.
// Aceita valores "true"/"false" (case insensitive) e trata entradas inválidas
// retornando o valor padrão e registrando um aviso.
func IsEnabled(value string, defaultValue bool) bool {
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseBool(strings.TrimSpace(strings.ToLower(value)))
	if err != nil {
		return defaultValue
	}

	return parsed
}
