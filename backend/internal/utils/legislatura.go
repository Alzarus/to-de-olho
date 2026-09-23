package utils

import (
	"log/slog"
	"os"
	"time"
)

// InicioLegislatura retorna a data de posse (1o de fevereiro) da legislatura
// em curso na data t. Legislaturas comecam em 2019, 2023, 2027...; em janeiro
// de um ano de posse a legislatura anterior ainda esta em curso.
func InicioLegislatura(t time.Time) time.Time {
	ano := t.Year() - ((t.Year() - 3) % 4)
	inicio := time.Date(ano, time.February, 1, 0, 0, 0, 0, time.UTC)
	if t.Before(inicio) {
		inicio = inicio.AddDate(-4, 0, 0)
	}
	return inicio
}

// InicioRecorte e a data a partir da qual votacoes e proposicoes entram no
// ranking: a posse da legislatura atual (decisao D1). RECORTE_INICIO
// (AAAA-MM-DD) sobrescreve, para recalcular com outro recorte sem deploy.
func InicioRecorte() time.Time {
	if v := os.Getenv("RECORTE_INICIO"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			return t
		}
		slog.Warn("RECORTE_INICIO invalido, usando a posse da legislatura atual", "valor", v)
	}
	return InicioLegislatura(time.Now())
}

// PeriodoDoMandato e [inicio do recorte, agora): o periodo do ranking do mandato.
func PeriodoDoMandato() (inicio, fim time.Time) {
	return InicioRecorte(), time.Now()
}

// PeriodoDoAno e o ano-calendario cortado pelo recorte e por agora. Em 2023,
// janeiro fica de fora (legislatura anterior); no ano corrente, o fim e hoje.
// Todas as fontes do ranking anual usam este mesmo intervalo.
func PeriodoDoAno(ano int) (inicio, fim time.Time) {
	inicio = time.Date(ano, time.January, 1, 0, 0, 0, 0, time.UTC)
	fim = time.Date(ano+1, time.January, 1, 0, 0, 0, 0, time.UTC)
	if r := InicioRecorte(); inicio.Before(r) {
		inicio = r
	}
	if agora := time.Now(); fim.After(agora) {
		fim = agora
	}
	if fim.Before(inicio) {
		fim = inicio
	}
	return inicio, fim
}
