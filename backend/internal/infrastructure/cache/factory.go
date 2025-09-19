package cache

import (
	"context"
	"time"
	"to-de-olho-backend/internal/config"
)

// CacheConfig configura qual tipo de cache usar
type CacheConfig struct {
	Type            string        `env:"CACHE_TYPE" envDefault:"redis"`           // "redis", "multilevel"
	L1MaxSize       int           `env:"CACHE_L1_MAX_SIZE" envDefault:"1000"`     // Tamanho máximo L1
	L1DefaultTTL    time.Duration `env:"CACHE_L1_DEFAULT_TTL" envDefault:"5m"`    // TTL padrão L1
	CleanupInterval time.Duration `env:"CACHE_CLEANUP_INTERVAL" envDefault:"10m"` // Intervalo limpeza
}

// NewOptimizedCache cria a instância de cache mais adequada
func NewOptimizedCache(cacheConfig *CacheConfig, redisConfig *config.RedisConfig) (CacheInterface, error) {
	// Sempre criar Redis como base
	redisCache := NewFromConfig(redisConfig)

	switch cacheConfig.Type {
	case "multilevel":
		// Cache multi-level (L1 + L2)
		mlCache := NewMultiLevelCache(
			redisCache,
			cacheConfig.L1MaxSize,
			cacheConfig.L1DefaultTTL,
		)

		// Iniciar worker de limpeza em background
		go mlCache.StartCleanupWorker(context.Background(), cacheConfig.CleanupInterval)

		return mlCache, nil

	case "redis":
		fallthrough
	default:
		// Apenas Redis (L2)
		return redisCache, nil
	}
}

// GetDefaultCacheConfig retorna configuração padrão otimizada
func GetDefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		Type:            "multilevel", // Usar multi-level por padrão para máxima performance
		L1MaxSize:       2000,         // Mais espaço para dados frequentes
		L1DefaultTTL:    5 * time.Minute,
		CleanupInterval: 10 * time.Minute,
	}
}
