package cache

import (
	"context"
	"testing"
	"time"

	"to-de-olho-backend/internal/config"
)

func TestNewFromConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *config.RedisConfig
	}{
		{
			name: "default configuration",
			config: &config.RedisConfig{
				Addr:         "localhost:6379",
				Password:     "",
				DB:           0,
				ReadTimeout:  500 * time.Millisecond,
				WriteTimeout: 500 * time.Millisecond,
				DialTimeout:  500 * time.Millisecond,
			},
		},
		{
			name: "custom configuration",
			config: &config.RedisConfig{
				Addr:         "redis.example.com:6380",
				Password:     "secret",
				DB:           1,
				ReadTimeout:  1 * time.Second,
				WriteTimeout: 1 * time.Second,
				DialTimeout:  1 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewFromConfig(tt.config)

			if cache == nil {
				t.Fatal("NewFromConfig returned nil")
			}

			// Verificar se o client foi inicializado
			if cache.rdb == nil {
				t.Error("Redis client não foi inicializado")
			}
		})
	}
}

func TestRedisCache_SetAndGet(t *testing.T) {
	// Use config padrão para testes
	cfg := &config.RedisConfig{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		ReadTimeout:  500 * time.Millisecond,
		WriteTimeout: 500 * time.Millisecond,
		DialTimeout:  500 * time.Millisecond,
	}

	cache := NewFromConfig(cfg)
	if cache == nil {
		t.Fatal("NewFromConfig returned nil")
	}

	ctx := context.Background()

	tests := []struct {
		name  string
		key   string
		value string
		ttl   time.Duration
	}{
		{
			name:  "basic set and get",
			key:   "test_key",
			value: "test_value",
			ttl:   5 * time.Minute,
		},
		{
			name:  "json value",
			key:   "json_key",
			value: `{"name":"test","id":123}`,
			ttl:   10 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Set
			cache.Set(ctx, tt.key, tt.value, tt.ttl)

			// Test Get
			result, found := cache.Get(ctx, tt.key)
			if !found {
				t.Skipf("Redis não disponível ou key não encontrada")
			}

			if result != tt.value {
				t.Errorf("Expected %q, got %q", tt.value, result)
			}
		})
	}
}

func TestRedisCache_SetAndGet_Basic(t *testing.T) {
	// Use config padrão para testes
	cfg := &config.RedisConfig{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		ReadTimeout:  500 * time.Millisecond,
		WriteTimeout: 500 * time.Millisecond,
		DialTimeout:  500 * time.Millisecond,
	}

	cache := NewFromConfig(cfg)
	if cache == nil {
		t.Fatal("NewFromConfig returned nil")
	}

	ctx := context.Background()
	key := "basic_test_key"
	value := "test_value"

	// Set value
	cache.Set(ctx, key, value, 5*time.Minute)

	// Verify it exists
	result, found := cache.Get(ctx, key)
	if !found {
		t.Skip("Redis não disponível")
	}

	if result != value {
		t.Errorf("Expected %q, got %q", value, result)
	}
}
