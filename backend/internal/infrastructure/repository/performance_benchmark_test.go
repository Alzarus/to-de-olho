package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"to-de-olho-backend/internal/domain"
)

// createTestDespesas cria despesas de teste
func createTestDespesas(count int) []domain.Despesa {
	despesas := make([]domain.Despesa, count)
	for i := 0; i < count; i++ {
		despesas[i] = domain.Despesa{
			Ano:            2025,
			Mes:            1,
			TipoDespesa:    "COMBUSTÍVEIS E LUBRIFICANTES",
			CodDocumento:   1000 + i, // int, não string
			TipoDocumento:  "Nota Fiscal",
			NumDocumento:   fmt.Sprintf("NF%d", i),
			ValorDocumento: 100.50 + float64(i),
			ValorLiquido:   100.50 + float64(i),
			ValorBruto:     100.50 + float64(i),
			NomeFornecedor: fmt.Sprintf("Fornecedor %d", i),
		}
	}
	return despesas
}

// BenchmarkCacheOperations testa performance de operações de cache
func BenchmarkCacheOperations(b *testing.B) {
	// Simular diferentes tipos de cache
	benchmarkCases := []struct {
		name  string
		cache interface{}
	}{
		{"Redis", &MockCache{}},
		{"MultiLevel", &MockMultiLevelCache{}},
	}

	for _, bc := range benchmarkCases {
		b.Run(bc.name, func(b *testing.B) {
			ctx := context.Background()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("benchmark_key_%d", i%1000) // Reutilizar algumas chaves
				value := fmt.Sprintf("benchmark_value_%d", i)

				// Simular operações de cache
				switch c := bc.cache.(type) {
				case *MockCache:
					c.Set(ctx, key, value, 5*time.Minute)
					c.Get(ctx, key)
				case *MockMultiLevelCache:
					c.Set(ctx, key, value, 5*time.Minute)
					c.Get(ctx, key)
				}
			}
		})
	}
}

// BenchmarkCacheHitRates compara taxa de acerto entre diferentes caches
func BenchmarkCacheHitRates(b *testing.B) {
	ctx := context.Background()

	// Configurar caches
	redisCache := &MockCache{}
	multiCache := &MockMultiLevelCache{}

	// Pre-popular com dados quentes
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("hot_key_%d", i)
		value := fmt.Sprintf("hot_value_%d", i)
		redisCache.Set(ctx, key, value, 10*time.Minute)
		multiCache.Set(ctx, key, value, 10*time.Minute)
	}

	b.Run("Redis_HitRate", func(b *testing.B) {
		hits := 0
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("hot_key_%d", i%100) // 100% hit rate esperado
			if _, found := redisCache.Get(ctx, key); found {
				hits++
			}
		}
		b.ReportMetric(float64(hits)/float64(b.N)*100, "%_hit_rate")
	})

	b.Run("MultiLevel_HitRate", func(b *testing.B) {
		hits := 0
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("hot_key_%d", i%100) // 100% hit rate esperado
			if _, found := multiCache.Get(ctx, key); found {
				hits++
			}
		}
		b.ReportMetric(float64(hits)/float64(b.N)*100, "%_hit_rate")
	})
}

// Mocks para benchmarks

type MockCache struct {
	data map[string]string
}

func (m *MockCache) Get(ctx context.Context, key string) (string, bool) {
	if m.data == nil {
		m.data = make(map[string]string)
	}
	value, exists := m.data[key]
	return value, exists
}

func (m *MockCache) Set(ctx context.Context, key, value string, ttl time.Duration) {
	if m.data == nil {
		m.data = make(map[string]string)
	}
	m.data[key] = value
}

type MockMultiLevelCache struct {
	l1Data map[string]string
	l2Data map[string]string
}

func (m *MockMultiLevelCache) Get(ctx context.Context, key string) (string, bool) {
	if m.l1Data == nil {
		m.l1Data = make(map[string]string)
	}
	if m.l2Data == nil {
		m.l2Data = make(map[string]string)
	}

	// Simular L1 primeiro (mais rápido)
	if value, exists := m.l1Data[key]; exists {
		return value, true
	}

	// Depois L2 (mais lento)
	if value, exists := m.l2Data[key]; exists {
		// Promover para L1
		m.l1Data[key] = value
		return value, true
	}

	return "", false
}

func (m *MockMultiLevelCache) Set(ctx context.Context, key, value string, ttl time.Duration) {
	if m.l1Data == nil {
		m.l1Data = make(map[string]string)
	}
	if m.l2Data == nil {
		m.l2Data = make(map[string]string)
	}
	m.l1Data[key] = value
	m.l2Data[key] = value
}
