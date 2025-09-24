package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// gzipWriter implementa compressão gzip para responses
type gzipWriter struct {
	gin.ResponseWriter
	writer io.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	return g.writer.Write([]byte(s))
}

// GzipMiddleware middleware para compressão automática de responses
func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verificar se cliente suporta gzip
		if !shouldCompress(c.Request) {
			c.Next()
			return
		}

		// Verificar tamanho mínimo para compressão (1KB)
		if c.Writer.Size() > 0 && c.Writer.Size() < 1024 {
			c.Next()
			return
		}

		// Configurar headers de compressão
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		// Criar writer gzip
		gz := gzip.NewWriter(c.Writer)
		defer gz.Close()

		// Substituir writer
		c.Writer = &gzipWriter{
			ResponseWriter: c.Writer,
			writer:         gz,
		}

		c.Next()
	}
}

// shouldCompress verifica se deve comprimir a response
func shouldCompress(req *http.Request) bool {
	// Verificar Accept-Encoding
	encoding := req.Header.Get("Accept-Encoding")
	if !strings.Contains(encoding, "gzip") {
		return false
	}

	// Não comprimir já comprimidos
	contentType := req.Header.Get("Content-Type")
	skipTypes := []string{
		"image/",
		"video/",
		"audio/",
		"application/zip",
		"application/gzip",
		"application/octet-stream",
	}

	for _, skip := range skipTypes {
		if strings.HasPrefix(contentType, skip) {
			return false
		}
	}

	return true
}

// CompressionStatsMiddleware coleta métricas de compressão
func CompressionStatsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		originalSize := c.Writer.Size()

		c.Next()

		compressedSize := c.Writer.Size()
		if originalSize > 0 && compressedSize > 0 {
			ratio := float64(compressedSize) / float64(originalSize) * 100
			c.Header("X-Compression-Ratio", string(rune(int(ratio))))
		}
	}
}
