package cache

import (
	"context"
	"time"
)

// CacheInterface define a interface comum para todos os tipos de cache
type CacheInterface interface {
	Get(ctx context.Context, key string) (string, bool)
	Set(ctx context.Context, key, value string, ttl time.Duration)
}

// StatsProvider fornece estatísticas de cache (opcional)
type StatsProvider interface {
	Stats() map[string]interface{}
}

// CleanupProvider fornece limpeza de cache (opcional)
type CleanupProvider interface {
	CleanupExpired()
	StartCleanupWorker(ctx context.Context, interval time.Duration)
}

// Verificar se nossas implementações satisfazem as interfaces
var (
	_ CacheInterface  = (*Cache)(nil)
	_ CacheInterface  = (*MultiLevelCache)(nil)
	_ StatsProvider   = (*MultiLevelCache)(nil)
	_ CleanupProvider = (*MultiLevelCache)(nil)
)
