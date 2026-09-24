package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// cachePublico vale para as leituras públicas da API.
//
// Os dados só mudam no sync diário (06:00 UTC), então servir uma cópia de
// alguns minutos não mostra nada errado a ninguém e tira da origem quase todo
// o tráfego repetido:
//   - max-age=60: o navegador reaproveita por 1 min (troca de aba, voltar).
//     Curto de propósito: depois do sync, quem está com o site aberto vê o
//     dado novo logo, sem depender de purge.
//   - s-maxage=600: a Cloudflare (Cache Rule criada à parte) guarda 10 min.
//     É aqui que está o ganho: uma rajada de visitantes, ou de robôs, na
//     mesma URL vira uma requisição à origem a cada 10 min.
//   - stale-while-revalidate=300: vencida a cópia, a borda ainda responde
//     com ela por até 5 min enquanto busca a nova, e ninguém espera a
//     origem. No pior caso o dado novo do sync aparece ~15 min depois.
const cachePublico = "public, max-age=60, s-maxage=600, stale-while-revalidate=300"

// cacheStats vale para /api/v1/stats, que traz o contador de acessos. O total
// muda a cada visitante novo, então a cópia é mais curta que a das outras
// leituras; mas as consultas dessa rota (COUNT e SUM nas maiores tabelas) são
// as mais caras da home, e um minuto de borda já as tira do caminho de cada
// visita. Um contador até 1 min atrasado não engana ninguém.
const cacheStats = "public, max-age=30, s-maxage=60, stale-while-revalidate=30"

// politicaDeCache escolhe o Cache-Control de uma rota GET. Vazio: não marca.
//
// Ficam de fora /health (o healthcheck precisa da resposta viva, não de uma
// cópia) e /api/v1/sync/* (jobs de ingestão, nunca cacheáveis). Rotas que já
// definem Cache-Control próprio (como no-store em /api/v1/acessos) também não
// são sobrescritas: ver escritorComCache.aplicar.
func politicaDeCache(path string) string {
	switch {
	case !strings.HasPrefix(path, "/api/v1/"):
		return ""
	case isSyncPath(path):
		return ""
	case path == "/api/v1/stats":
		return cacheStats
	default:
		return cachePublico
	}
}

// cacheDeLeitura marca as respostas 200 de GET com Cache-Control.
//
// Só o 200 recebe o header. Erro (4xx, 5xx, 429 do limite) em cache
// espalharia uma falha passageira para todo mundo durante minutos.
func cacheDeLeitura() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}
		politica := politicaDeCache(c.Request.URL.Path)
		if politica == "" {
			c.Next()
			return
		}
		c.Writer = &escritorComCache{ResponseWriter: c.Writer, politica: politica}
		c.Next()
	}
}

// escritorComCache decide o Cache-Control no último instante possível, quando
// o status já é conhecido e os headers ainda não saíram. Depois de c.Next()
// seria tarde: o corpo (e com ele os headers) já foi escrito.
type escritorComCache struct {
	gin.ResponseWriter
	politica string
	aplicado bool
}

func (w *escritorComCache) aplicar() {
	if w.aplicado {
		return
	}
	w.aplicado = true
	if w.Status() == http.StatusOK && w.Header().Get("Cache-Control") == "" {
		w.Header().Set("Cache-Control", w.politica)
	}
}

func (w *escritorComCache) WriteHeaderNow() {
	w.aplicar()
	w.ResponseWriter.WriteHeaderNow()
}

func (w *escritorComCache) Write(dados []byte) (int, error) {
	w.aplicar()
	return w.ResponseWriter.Write(dados)
}

func (w *escritorComCache) WriteString(s string) (int, error) {
	w.aplicar()
	return w.ResponseWriter.WriteString(s)
}
