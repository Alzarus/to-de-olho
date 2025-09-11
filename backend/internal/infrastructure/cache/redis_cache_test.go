package cache

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	// Salvar variáveis originais
	originalHost := os.Getenv("REDIS_HOST")
	originalPort := os.Getenv("REDIS_PORT")

	// Limpar variáveis para testar defaults
	os.Unsetenv("REDIS_HOST")
	os.Unsetenv("REDIS_PORT")

	defer func() {
		// Restaurar variáveis originais
		if originalHost != "" {
			os.Setenv("REDIS_HOST", originalHost)
		}
		if originalPort != "" {
			os.Setenv("REDIS_PORT", originalPort)
		}
	}()

	tests := []struct {
		name         string
		host         string
		port         string
		expectedAddr string
	}{
		{
			name:         "valores padrão",
			host:         "",
			port:         "",
			expectedAddr: "localhost:6379",
		},
		{
			name:         "host customizado",
			host:         "redis.example.com",
			port:         "",
			expectedAddr: "redis.example.com:6379",
		},
		{
			name:         "porta customizada",
			host:         "",
			port:         "6380",
			expectedAddr: "localhost:6380",
		},
		{
			name:         "host e porta customizados",
			host:         "custom.redis.com",
			port:         "6381",
			expectedAddr: "custom.redis.com:6381",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Configurar variáveis de ambiente
			if tt.host != "" {
				os.Setenv("REDIS_HOST", tt.host)
			} else {
				os.Unsetenv("REDIS_HOST")
			}

			if tt.port != "" {
				os.Setenv("REDIS_PORT", tt.port)
			} else {
				os.Unsetenv("REDIS_PORT")
			}

			cache := New()

			if cache == nil {
				t.Error("New() deveria retornar uma instância válida")
				return
			}

			if cache.rdb == nil {
				t.Error("cliente Redis não foi inicializado")
				return
			}

			// Verificar se o endereço foi configurado corretamente
			// Nota: Não podemos acessar diretamente o addr do cliente Redis,
			// então testamos indiretamente
			options := cache.rdb.Options()
			if options.Addr != tt.expectedAddr {
				t.Errorf("endereço esperado: %s, recebido: %s", tt.expectedAddr, options.Addr)
			}
		})
	}
}

func TestCache_GetSet(t *testing.T) {
	tests := []struct {
		name        string
		cache       *Cache
		key         string
		value       string
		ttl         time.Duration
		shouldExist bool
		description string
	}{
		{
			name:        "cache nil - operações seguras",
			cache:       nil,
			key:         "test-key",
			value:       "test-value",
			ttl:         1 * time.Minute,
			shouldExist: false,
			description: "Cache nil deveria ser tratado graciosamente",
		},
		{
			name:        "cache com client nil",
			cache:       &Cache{rdb: nil},
			key:         "test-key",
			value:       "test-value",
			ttl:         1 * time.Minute,
			shouldExist: false,
			description: "Cache com client nil deveria ser tratado graciosamente",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Test Set (não deveria causar panic)
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Set() causou panic: %v - %s", r, tt.description)
					}
				}()

				tt.cache.Set(ctx, tt.key, tt.value, tt.ttl)
			}()

			// Test Get (não deveria causar panic)
			var result string
			var exists bool

			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Get() causou panic: %v - %s", r, tt.description)
					}
				}()

				result, exists = tt.cache.Get(ctx, tt.key)
			}()

			if exists != tt.shouldExist {
				t.Errorf("existência esperada: %v, recebida: %v - %s",
					tt.shouldExist, exists, tt.description)
			}

			if tt.shouldExist && result != tt.value {
				t.Errorf("valor esperado: %s, recebido: %s - %s",
					tt.value, result, tt.description)
			}
		})
	}
}

func TestCache_GetNotFound(t *testing.T) {
	// Teste com cache válido mas sem Redis real
	cache := &Cache{rdb: nil}
	ctx := context.Background()

	value, exists := cache.Get(ctx, "key-inexistente")

	if exists {
		t.Error("chave inexistente deveria retornar false")
	}

	if value != "" {
		t.Errorf("valor deveria ser string vazia, recebeu: %s", value)
	}
}

func TestCache_SetWithDifferentTTL(t *testing.T) {
	cache := &Cache{rdb: nil}
	ctx := context.Background()

	tests := []struct {
		name string
		ttl  time.Duration
	}{
		{"TTL zero", 0},
		{"TTL negativo", -1 * time.Second},
		{"TTL positivo", 10 * time.Minute},
		{"TTL muito longo", 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Não deveria causar panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Set com TTL %v causou panic: %v", tt.ttl, r)
				}
			}()

			cache.Set(ctx, "test-key", "test-value", tt.ttl)
		})
	}
}

func TestCache_EdgeCases(t *testing.T) {
	cache := &Cache{rdb: nil}
	ctx := context.Background()

	// Teste com chave vazia
	cache.Set(ctx, "", "valor", 1*time.Minute)
	value, exists := cache.Get(ctx, "")
	if exists {
		t.Error("chave vazia não deveria existir")
	}

	// Teste com valor vazio
	cache.Set(ctx, "chave", "", 1*time.Minute)
	value, exists = cache.Get(ctx, "chave")
	if exists && value != "" {
		t.Error("valor vazio deveria ser retornado corretamente")
	}

	// Teste com chave muito longa
	longKey := string(make([]byte, 1000))
	cache.Set(ctx, longKey, "valor", 1*time.Minute)
	_, exists = cache.Get(ctx, longKey)
	// Não deveria causar erro
}

func TestCache_ContextCancellation(t *testing.T) {
	cache := &Cache{rdb: nil}

	// Contexto já cancelado
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Operações com contexto cancelado não deveriam causar panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("operações com contexto cancelado causaram panic: %v", r)
		}
	}()

	cache.Set(ctx, "test", "value", 1*time.Minute)
	_, _ = cache.Get(ctx, "test")
}

// Benchmark para operações de cache
func BenchmarkCache_Set(b *testing.B) {
	cache := &Cache{rdb: nil}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(ctx, "benchmark-key", "benchmark-value", 1*time.Minute)
	}
}

func BenchmarkCache_Get(b *testing.B) {
	cache := &Cache{rdb: nil}
	ctx := context.Background()

	// Setup
	cache.Set(ctx, "benchmark-key", "benchmark-value", 1*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(ctx, "benchmark-key")
	}
}
