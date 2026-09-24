package api

import (
	"log/slog"
	"math"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// rotaAcessos é o contador de visitas da home. Tem limite próprio, bem mais
// duro, porque cada POST pode virar uma linha nova no banco e somar no total
// exibido: sem teto, qualquer um inflava o contador trocando o User-Agent.
const rotaAcessos = "/api/v1/acessos"

// chaveRedeInterna agrupa as requisições sem CF-Connecting-IP vindas de rede
// privada. Em produção isso é o SSR do Next chamando a API pela rede do
// Docker (ou o curl de manutenção no próprio host).
const chaveRedeInterna = "rede-interna"

// intervaloLimpeza é de quanto em quanto tempo os baldes ociosos são varridos.
const intervaloLimpeza = time.Minute

// Taxa é um token bucket: PorSegundo fichas repostas por segundo, até Rajada
// guardadas. PorSegundo <= 0 desliga o limite daquele grupo.
type Taxa struct {
	PorSegundo float64
	Rajada     int
}

func (t Taxa) ativa() bool { return t.PorSegundo > 0 && t.Rajada > 0 }

// ConfigLimite reúne os limites por grupo de origem.
type ConfigLimite struct {
	// Visitante vale por IP de quem abriu o site (CF-Connecting-IP).
	Visitante Taxa
	// Interno é um balde único para a rede privada sem CF-Connecting-IP.
	// Fica separado e generoso para o SSR não disputar ficha com visitante:
	// todas as páginas renderizadas no servidor chegam do mesmo contêiner.
	Interno Taxa
	// Acessos vale só para POST /api/v1/acessos, por visitante.
	Acessos Taxa
	// MaxVisitantes limita quantos IPs são lembrados ao mesmo tempo. Acima
	// disso os IPs novos dividem um balde de excedente (ver baldes.reservar).
	MaxVisitantes int
}

// ConfigLimitePadrao são os valores usados quando o ambiente não diz nada.
//
//   - Visitante: 10 req/s com rajada de 40. Abrir a ficha de um senador ou o
//     comparador com cinco senadores dispara de 10 a 30 GETs quase juntos; a
//     exportação percorre páginas em sequência, bem abaixo de 10/s. Sobra
//     margem para algumas pessoas atrás do mesmo IP (CGNAT das operadoras).
//   - Interno: 50 req/s com rajada de 200. O Next revalida as páginas por
//     ISR, então o SSR raramente passa de poucas req/s; o teto só existe para
//     que quem alcance a origem sem Cloudflare (e sem o header) não tenha
//     vazão ilimitada.
//   - Acessos: 1 por minuto com rajada de 10. O navegador avisa uma vez por
//     aba; 10 abas seguidas passam, um script em loop não.
func ConfigLimitePadrao() ConfigLimite {
	return ConfigLimite{
		Visitante:     Taxa{PorSegundo: 10, Rajada: 40},
		Interno:       Taxa{PorSegundo: 50, Rajada: 200},
		Acessos:       Taxa{PorSegundo: 1.0 / 60, Rajada: 10},
		MaxVisitantes: 100_000,
	}
}

// ConfigLimiteDoAmbiente parte do padrão e aplica as variáveis RATE_LIMIT_*
// (documentadas em backend/.env.example). Valor inválido é ignorado com aviso
// no log, para um erro de digitação não derrubar a API.
func ConfigLimiteDoAmbiente() ConfigLimite {
	cfg := ConfigLimitePadrao()
	taxa := func(nome string, destino *float64, escala float64) {
		if v, ok := numeroDoAmbiente(nome); ok {
			*destino = v / escala
		}
	}
	inteiro := func(nome string, destino *int) {
		if v, ok := numeroDoAmbiente(nome); ok {
			*destino = int(v)
		}
	}
	taxa("RATE_LIMIT_RPS", &cfg.Visitante.PorSegundo, 1)
	inteiro("RATE_LIMIT_BURST", &cfg.Visitante.Rajada)
	taxa("RATE_LIMIT_INTERNO_RPS", &cfg.Interno.PorSegundo, 1)
	inteiro("RATE_LIMIT_INTERNO_BURST", &cfg.Interno.Rajada)
	taxa("RATE_LIMIT_ACESSOS_POR_MINUTO", &cfg.Acessos.PorSegundo, 60)
	inteiro("RATE_LIMIT_ACESSOS_BURST", &cfg.Acessos.Rajada)
	inteiro("RATE_LIMIT_MAX_VISITANTES", &cfg.MaxVisitantes)
	return cfg
}

// numeroDoAmbiente lê um número não negativo. ok = false quando a variável
// está vazia ou inválida (e aí o chamador mantém o padrão).
func numeroDoAmbiente(nome string) (float64, bool) {
	v := strings.TrimSpace(os.Getenv(nome))
	if v == "" {
		return 0, false
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil || n < 0 || math.IsNaN(n) || math.IsInf(n, 0) {
		slog.Warn("variável de limite inválida, usando o padrão", "variavel", nome, "valor", v)
		return 0, false
	}
	return n, true
}

// balde guarda o token bucket de uma origem e quando ela apareceu por último.
type balde struct {
	lim   *rate.Limiter
	visto time.Time
}

// baldes é o conjunto de token buckets de um grupo, um por chave, em memória.
//
// Em memória basta: a API roda num contêiner só. Com mais réplicas cada uma
// teria o próprio balde e o limite efetivo seria multiplicado.
type baldes struct {
	mu            sync.Mutex
	taxa          Taxa
	porChave      map[string]*balde
	max           int
	excedente     *rate.Limiter
	ociosidade    time.Duration
	ultimaLimpeza time.Time
}

func novosBaldes(t Taxa, maxChaves int, agora time.Time) *baldes {
	// Um balde parado há mais tempo que o necessário para encher de novo
	// equivale a um balde novo, então pode ser apagado sem dar ficha extra a
	// ninguém. Apagar antes disso deixaria quem espera um pouco recomeçar
	// com a rajada cheia antes da hora.
	ociosidade := time.Duration(float64(t.Rajada) / t.PorSegundo * float64(time.Second))
	return &baldes{
		taxa:          t,
		porChave:      make(map[string]*balde),
		max:           maxChaves,
		excedente:     rate.NewLimiter(rate.Limit(t.PorSegundo), t.Rajada),
		ociosidade:    max(ociosidade, intervaloLimpeza),
		ultimaLimpeza: agora,
	}
}

// reservar consome uma ficha da chave. Devolve 0 se liberou ou, se não há
// ficha, quanto falta para a próxima (vai no Retry-After).
func (b *baldes) reservar(chave string, agora time.Time) time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()

	// A limpeza roda aqui, no máximo uma vez por intervalo, em vez de numa
	// goroutine própria: não precisa de ciclo de vida para parar nos testes
	// e só custa quando há tráfego, que é quando há o que limpar.
	if agora.Sub(b.ultimaLimpeza) >= intervaloLimpeza {
		b.limpar(agora)
	}

	var lim *rate.Limiter
	switch e, ok := b.porChave[chave]; {
	case ok:
		e.visto = agora
		lim = e.lim
	case len(b.porChave) >= b.max:
		// Tabela cheia: quem gira IPs (ou forja CF-Connecting-IP batendo
		// direto na origem) para ganhar um balde novo a cada requisição cai
		// num balde compartilhado. A memória fica limitada e a enxurrada
		// continua limitada à taxa de um visitante só.
		lim = b.excedente
	default:
		lim = rate.NewLimiter(rate.Limit(b.taxa.PorSegundo), b.taxa.Rajada)
		b.porChave[chave] = &balde{lim: lim, visto: agora}
	}

	r := lim.ReserveN(agora, 1)
	if !r.OK() {
		return intervaloLimpeza
	}
	if espera := r.DelayFrom(agora); espera > 0 {
		// Devolve a ficha: quem foi recusado não deve empurrar a fila dos
		// próximos pedidos para mais longe.
		r.CancelAt(agora)
		return espera
	}
	return 0
}

// limpar apaga os baldes ociosos. Chamada com b.mu travado.
func (b *baldes) limpar(agora time.Time) {
	for chave, e := range b.porChave {
		if agora.Sub(e.visto) > b.ociosidade {
			delete(b.porChave, chave)
		}
	}
	b.ultimaLimpeza = agora
}

func (b *baldes) tamanho() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.porChave)
}

// origemDaRequisicao decide em que balde a requisição cai.
//
// Com CF-Connecting-IP válido, a chave é o IP do visitante: a Cloudflare
// sobrescreve esse header, então pelo caminho normal ele não é forjável. Sem
// o header, a chave é o IP da conexão (c.RemoteIP), nunca X-Forwarded-For,
// que o cliente escreve como quiser. Conexão de rede privada ou loopback sem
// o header é o SSR do Next e vai para o balde interno.
//
// Quem alcançar a origem sem passar pela Cloudflare pode forjar o header;
// o estrago fica contido pelo teto de baldes (MaxVisitantes) e se encerra
// quando o firewall aceitar só os IPs da Cloudflare.
func origemDaRequisicao(c *gin.Context) (chave string, interna bool) {
	if ip := net.ParseIP(strings.TrimSpace(c.GetHeader("CF-Connecting-IP"))); ip != nil {
		return chaveDoIP(ip), false
	}
	remoto := net.ParseIP(c.RemoteIP())
	if remoto == nil {
		return c.Request.RemoteAddr, false
	}
	if remoto.IsPrivate() || remoto.IsLoopback() {
		return chaveRedeInterna, true
	}
	return chaveDoIP(remoto), false
}

// chaveDoIP agrupa IPv6 por /64. Um único acesso doméstico recebe um /64
// inteiro, então limitar por endereço deixaria a mesma pessoa trocar de IP à
// vontade e ganhar um balde novo a cada troca.
func chaveDoIP(ip net.IP) string {
	if v4 := ip.To4(); v4 != nil {
		return v4.String()
	}
	return ip.Mask(net.CIDRMask(64, 128)).String() + "/64"
}

// limitarRequisicoes aplica os limites por origem.
//
// /health fica de fora porque o HEALTHCHECK do Docker bate nela a cada 30 s
// pelo loopback e não pode ser recusado por causa de tráfego alheio.
func limitarRequisicoes(cfg ConfigLimite, relogio func() time.Time) gin.HandlerFunc {
	agora := relogio()
	var visitantes, interno, acessos *baldes
	if cfg.Visitante.ativa() {
		visitantes = novosBaldes(cfg.Visitante, cfg.MaxVisitantes, agora)
	}
	if cfg.Interno.ativa() {
		interno = novosBaldes(cfg.Interno, 1, agora)
	}
	if cfg.Acessos.ativa() {
		acessos = novosBaldes(cfg.Acessos, cfg.MaxVisitantes, agora)
	}

	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		chave, interna := origemDaRequisicao(c)
		grupo := visitantes
		switch {
		case c.Request.Method == http.MethodPost && c.Request.URL.Path == rotaAcessos:
			// Vale também para a rede interna (chave única): o navegador
			// sempre chega com CF-Connecting-IP, então POST sem o header é
			// alguém batendo direto na origem.
			grupo = acessos
		case interna:
			grupo = interno
		}
		if grupo == nil {
			c.Next()
			return
		}

		if espera := grupo.reservar(chave, relogio()); espera > 0 {
			c.Header("Retry-After", strconv.Itoa(segundosParaRetry(espera)))
			// 429 não pode ficar em cache (nem na Cloudflare nem no navegador).
			c.Header("Cache-Control", "no-store")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "muitas requisições; tente de novo em instantes",
			})
			return
		}
		c.Next()
	}
}

// segundosParaRetry arredonda para cima: Retry-After só aceita segundos
// inteiros, e 0 faria o cliente repetir na hora e ser recusado de novo.
func segundosParaRetry(d time.Duration) int {
	s := int(math.Ceil(d.Seconds()))
	if s < 1 {
		return 1
	}
	return s
}
