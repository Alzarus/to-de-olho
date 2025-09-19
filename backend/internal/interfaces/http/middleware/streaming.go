package middleware

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// StreamingResponse estrutura para response streamável
type StreamingResponse struct {
	Data       interface{} `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
	Meta       *Meta       `json:"meta,omitempty"`
}

// Pagination informações de paginação
type Pagination struct {
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	Total      int64  `json:"total"`
	TotalPages int    `json:"total_pages"`
	HasNext    bool   `json:"has_next"`
	HasPrev    bool   `json:"has_prev"`
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
}

// Meta metadados da response
type Meta struct {
	RequestID   string `json:"request_id"`
	ProcessTime string `json:"process_time"`
	CacheHit    bool   `json:"cache_hit"`
	Compressed  bool   `json:"compressed"`
}

// StreamingMiddleware middleware para responses streamáveis
func StreamingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verificar se deve usar streaming
		if shouldStream(c.Request) {
			c.Header("Transfer-Encoding", "chunked")
			c.Header("Content-Type", "application/json")
			c.Header("Cache-Control", "no-cache")
			c.Header("Connection", "keep-alive")
		}

		c.Next()
	}
}

// shouldStream determina se deve usar streaming
func shouldStream(req *http.Request) bool {
	// Streaming para datasets grandes
	limitStr := req.URL.Query().Get("limit")
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 100 {
			return true
		}
	}

	// Streaming para exports
	if req.URL.Query().Get("format") == "stream" {
		return true
	}

	// Streaming para análises em tempo real
	if req.URL.Path == "/api/v1/analytics/realtime" {
		return true
	}

	return false
}

// WriteStreamingJSON escreve JSON de forma streamável
func WriteStreamingJSON(c *gin.Context, data interface{}) {
	c.Header("Content-Type", "application/json")

	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao serializar response",
		})
		return
	}

	// Força flush do buffer
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

// WriteChunkedJSON escreve JSON em chunks para grandes volumes
func WriteChunkedJSON(c *gin.Context, items []interface{}, chunkSize int) {
	c.Header("Content-Type", "application/json")
	c.Writer.WriteString(`{"data":[`)

	for i, item := range items {
		if i > 0 {
			c.Writer.WriteString(",")
		}

		// Escrever item
		data, err := json.Marshal(item)
		if err != nil {
			continue
		}
		c.Writer.Write(data)

		// Flush a cada chunk
		if (i+1)%chunkSize == 0 {
			if flusher, ok := c.Writer.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}

	c.Writer.WriteString(`]}`)

	// Flush final
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}
