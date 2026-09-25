package senador

import (
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"gorm.io/gorm"
)

// hostDoSenado diz se o host é do Senado (senado.leg.br ou subdomínio).
func hostDoSenado(host string) bool {
	host = strings.ToLower(host)
	return host == "senado.leg.br" || strings.HasSuffix(host, ".senado.leg.br")
}

// NormalizarFotoURL troca http:// por https:// nas fotos do Senado.
//
// A API de dados abertos devolve UrlFotoParlamentar como
// http://www.senado.leg.br/...; num site servido em https isso é conteúdo
// misto (o navegador bloqueia ou avisa) e o next/image só aceita o host em
// https. O Senado serve as mesmas fotos em https (verificado em 24/09/2026:
// https://www.senado.leg.br/senadores/img/fotos-oficiais/senador5895.jpg
// redireciona para legis.senado.leg.br em https e devolve image/jpeg).
// URLs de outros hosts ficam como estão: não sabemos se aceitam https.
func NormalizarFotoURL(bruta string) string {
	u, err := url.Parse(strings.TrimSpace(bruta))
	// Porta explícita (http://host:80/...) fica como está: trocar só o
	// esquema apontaria o https para a porta do http.
	if err != nil || !strings.EqualFold(u.Scheme, "http") || u.Port() != "" || !hostDoSenado(u.Hostname()) {
		return bruta
	}
	u.Scheme = "https"
	return u.String()
}

// NormalizarFotosHTTPS corrige as fotos gravadas antes da normalização no
// sync. Idempotente: depois da primeira execução não sobra linha que case com
// o filtro.
//
// SQL direto, e não Model().Update, para não mexer em updated_at: essa coluna
// alimenta a "última atualização" exibida no site, e trocar o esquema da URL
// não é dado novo do Senado.
func NormalizarFotosHTTPS(db *gorm.DB) error {
	res := db.Exec(`
		UPDATE senadores
		SET foto_url = 'https://' || substr(foto_url, 8)
		WHERE foto_url ~* '^http://([a-z0-9-]+\.)*senado\.leg\.br(/|$)'`)
	if res.Error != nil {
		return fmt.Errorf("normalizar foto_url para https: %w", res.Error)
	}
	if res.RowsAffected > 0 {
		slog.Info("senadores: foto_url normalizada para https", "linhas", res.RowsAffected)
	}
	return nil
}
