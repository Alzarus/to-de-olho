package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"to-de-olho-backend/internal/domain"

	"github.com/gin-gonic/gin"
)

// Benchmark response sem otimizações
func BenchmarkResponseBaseline(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	// Handler básico
	router := gin.New()
	router.GET("/deputados", func(c *gin.Context) {
		deputados := make([]domain.Deputado, 513)
		for i := 0; i < 513; i++ {
			deputados[i] = domain.Deputado{
				ID:       i + 1,
				Nome:     fmt.Sprintf("Deputado %d", i+1),
				Partido:  "PT",
				UF:       "SP",
				Situacao: "Exercício",
			}
		}
		c.JSON(http.StatusOK, gin.H{"data": deputados})
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/deputados", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// Benchmark compressão gzip
func BenchmarkCompressionGzip(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	// Dados grandes para testar compressão
	largeData := make([]domain.Deputado, 513)
	for i := 0; i < 513; i++ {
		largeData[i] = domain.Deputado{
			ID:       i + 1,
			Nome:     fmt.Sprintf("Deputado com Nome Muito Longo %d", i+1),
			Partido:  "PARTIDO",
			UF:       "SP",
			Situacao: "Exercício",
			Email:    "deputado.muito.longo@email.com.br",
		}
	}

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": largeData})
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// Benchmark paginação
func BenchmarkPagination(b *testing.B) {
	// Benchmark para diferentes tamanhos de página
	pageSizes := []int{10, 20, 50, 100}

	for _, pageSize := range pageSizes {
		b.Run(fmt.Sprintf("PageSize%d", pageSize), func(b *testing.B) {
			paginationReq := &domain.PaginationRequest{
				Page:   1,
				Limit:  pageSize,
				SortBy: "nome",
				Order:  "asc",
			}

			// Simular dados grandes
			allData := make([]domain.Deputado, 1000)
			for i := 0; i < 1000; i++ {
				allData[i] = domain.Deputado{
					ID:   i + 1,
					Nome: fmt.Sprintf("Deputado %d", i+1),
				}
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = domain.BuildPagination(paginationReq, int64(len(allData)), allData[:pageSize])
			}
		})
	}
}

// Benchmark serialização JSON
func BenchmarkJSONSerialization(b *testing.B) {
	// Dados grandes
	largeData := make([]domain.Deputado, 513)
	for i := 0; i < 513; i++ {
		largeData[i] = domain.Deputado{
			ID:       i + 1,
			Nome:     fmt.Sprintf("Deputado %d", i+1),
			Partido:  "PT",
			UF:       "SP",
			Situacao: "Exercício",
			Email:    "deputado@email.com",
		}
	}

	b.Run("StandardJSON", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			encoder := json.NewEncoder(&buf)
			_ = encoder.Encode(largeData)
		}
	})

	b.Run("DirectMarshal", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(largeData)
		}
	})
}

// Test de performance end-to-end
func TestResponsePerformance(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/deputados", func(c *gin.Context) {
		deputados := make([]domain.Deputado, 513)
		for i := 0; i < 513; i++ {
			deputados[i] = domain.Deputado{
				ID:   i + 1,
				Nome: fmt.Sprintf("Deputado %d", i+1),
			}
		}
		c.JSON(http.StatusOK, gin.H{"data": deputados})
	})

	// Test response time
	start := time.Now()
	req := httptest.NewRequest(http.MethodGet, "/deputados", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	responseTime := time.Since(start)

	t.Logf("Response time: %v", responseTime)
	t.Logf("Response size: %d bytes", w.Body.Len())
	t.Logf("Status code: %d", w.Code)

	// Verificar se está dentro do SLA (< 100ms)
	if responseTime > 100*time.Millisecond {
		t.Errorf("Response time %v exceeds SLA of 100ms", responseTime)
	}

	// Verificar se retornou dados
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
