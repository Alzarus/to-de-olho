package utils

import "time"

// GetInicioLegislaturaAtual retorna o ano de inicio da legislatura atual (ex: 2023 para 2023-2026)
func GetInicioLegislaturaAtual() int {
	year := time.Now().Year()
	// As legislaturas no Senado comecam nos anos seguintes as eleicoes (ex: 2015, 2019, 2023, 2027)
	return year - ((year - 3) % 4)
}
