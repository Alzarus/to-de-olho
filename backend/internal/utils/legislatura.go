package utils

import (
	"log/slog"
	"os"
	"time"
)

// MesesTransicaoLegislatura e o tempo, contado da posse, em que o ranking do
// mandato continua mostrando a legislatura anterior (item 14). Coincide com o
// piso de exercicio do ranking (6 meses): antes disso, todo senador da nova
// legislatura ficaria em "dados insuficientes" e o ranking sairia vazio.
const MesesTransicaoLegislatura = 6

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

// recorteForcado le RECORTE_INICIO (AAAA-MM-DD), que sobrescreve o recorte
// para recalcular com outro periodo sem deploy.
func recorteForcado() (time.Time, bool) {
	v := os.Getenv("RECORTE_INICIO")
	if v == "" {
		return time.Time{}, false
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		slog.Warn("RECORTE_INICIO invalido, usando a posse da legislatura", "valor", v)
		return time.Time{}, false
	}
	return t, true
}

// periodoDoMandatoEm e o periodo do ranking do mandato visto na data agora.
//
// Normalmente vai da posse da legislatura atual ate agora (decisao D1). Nos
// primeiros MesesTransicaoLegislatura meses de uma legislatura, e a
// legislatura anterior inteira, ja encerrada (encerrada = true).
func periodoDoMandatoEm(agora time.Time) (inicio, fim time.Time, encerrada bool) {
	if r, ok := recorteForcado(); ok {
		return r, agora, false
	}
	posse := InicioLegislatura(agora)
	if agora.Before(posse.AddDate(0, MesesTransicaoLegislatura, 0)) {
		return posse.AddDate(-4, 0, 0), posse, true
	}
	return posse, agora, false
}

// PeriodoDoMandato e [inicio, fim): o periodo do ranking do mandato e das
// estatisticas "do mandato" das fichas. Ver periodoDoMandatoEm.
func PeriodoDoMandato() (inicio, fim time.Time) {
	inicio, fim, _ = periodoDoMandatoEm(time.Now())
	return inicio, fim
}

// MandatoEncerrado indica que o ranking do mandato mostra a legislatura
// anterior (transicao, item 14).
func MandatoEncerrado() bool {
	_, _, encerrada := periodoDoMandatoEm(time.Now())
	return encerrada
}

// InicioRecorte e o inicio do periodo do mandato: a partir dele as cargas de
// votacoes e proposicoes trazem dados. Na transicao, e a posse da legislatura
// anterior, para que a carga cubra tambem o periodo que o ranking mostra.
func InicioRecorte() time.Time {
	inicio, _ := PeriodoDoMandato()
	return inicio
}

// PeriodoDoAno e o ano-calendario cortado pela posse da legislatura a que o
// ano pertence e por agora. Janeiro de um ano de posse (2023, 2027...) e da
// legislatura anterior e fica de fora; no ano corrente, o fim e hoje. Todas
// as fontes do ranking anual usam este mesmo intervalo.
func PeriodoDoAno(ano int) (inicio, fim time.Time) {
	inicio = time.Date(ano, time.January, 1, 0, 0, 0, 0, time.UTC)
	fim = time.Date(ano+1, time.January, 1, 0, 0, 0, 0, time.UTC)
	corte, ok := recorteForcado()
	if !ok {
		corte = InicioLegislatura(fim.AddDate(0, 0, -1))
	}
	if inicio.Before(corte) {
		inicio = corte
	}
	if agora := time.Now(); fim.After(agora) {
		fim = agora
	}
	if fim.Before(inicio) {
		fim = inicio
	}
	return inicio, fim
}
