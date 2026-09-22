package api

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// SyncSecretHeader e o header que carrega o segredo compartilhado dos jobs de sync.
const SyncSecretHeader = "X-Sync-Secret"

// syncSecret le o segredo do ambiente. Variavel de funcao para permitir teste.
var syncSecret = func() string {
	return strings.TrimSpace(os.Getenv("SYNC_SECRET"))
}

// requireSyncSecret protege os endpoints /api/v1/sync/*.
//
// Falha fechada: se SYNC_SECRET nao estiver configurado, nenhum sync roda.
// A versao anterior liberava tudo nesse caso, o que deixava os jobs de
// ingestao abertos para qualquer pessoa na internet.
func requireSyncSecret() gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := syncSecret()
		if secret == "" {
			slog.Error("SYNC_SECRET nao configurado: endpoints de sync desativados",
				"path", c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "sync indisponivel",
			})
			return
		}

		informado := c.GetHeader(SyncSecretHeader)
		if subtle.ConstantTimeCompare([]byte(informado), []byte(secret)) != 1 {
			slog.Warn("tentativa de sync sem credencial valida",
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
				"ip", c.ClientIP())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "acesso negado",
			})
			return
		}

		c.Next()
	}
}

// isSyncPath identifica as rotas de ingestao, que nao sao para navegador.
func isSyncPath(path string) bool {
	return path == "/api/v1/sync" || strings.HasPrefix(path, "/api/v1/sync/")
}
