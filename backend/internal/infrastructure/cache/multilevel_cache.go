package cache

import (
	"container/heap"
	"context"
	"sync"
	"time"
)

// CacheItem representa um item no cache L1
type CacheItem struct {
	Value     string
	ExpiresAt time.Time
}

// IsExpired verifica se o item expirou
func (c *CacheItem) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// evictionItem representa um item para evicção no heap
type evictionItem struct {
	key       string
	expiresAt time.Time
}

// evictionHeap implementa heap.Interface para encontrar itens mais antigos
type evictionHeap []evictionItem

func (h evictionHeap) Len() int           { return len(h) }
func (h evictionHeap) Less(i, j int) bool { return h[i].expiresAt.Before(h[j].expiresAt) }
func (h evictionHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *evictionHeap) Push(x interface{}) {
	*h = append(*h, x.(evictionItem))
}

func (h *evictionHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

// MultiLevelCache implementa cache L1 (in-memory) + L2 (Redis)
type MultiLevelCache struct {
	l1Cache map[string]*CacheItem
	l1Mutex sync.RWMutex
	l2Cache *Cache // Redis cache

	// Configuração L1
	l1MaxSize    int
	l1DefaultTTL time.Duration
}

// NewMultiLevelCache cria uma nova instância do cache multi-level
func NewMultiLevelCache(redisCache *Cache, l1MaxSize int, l1DefaultTTL time.Duration) *MultiLevelCache {
	if l1MaxSize <= 0 {
		l1MaxSize = 1000 // default
	}
	if l1DefaultTTL <= 0 {
		l1DefaultTTL = 5 * time.Minute // default
	}

	return &MultiLevelCache{
		l1Cache:      make(map[string]*CacheItem),
		l2Cache:      redisCache,
		l1MaxSize:    l1MaxSize,
		l1DefaultTTL: l1DefaultTTL,
	}
}

// Get busca no cache L1 primeiro, depois L2
func (mlc *MultiLevelCache) Get(ctx context.Context, key string) (string, bool) {
	// 1. Tentar L1 primeiro (mais rápido)
	if value, found := mlc.getFromL1(key); found {
		return value, true
	}

	// 2. Tentar L2 (Redis)
	if mlc.l2Cache != nil {
		if value, found := mlc.l2Cache.Get(ctx, key); found {
			// Promover para L1 para próximas consultas
			mlc.setToL1(key, value, mlc.l1DefaultTTL)
			return value, true
		}
	}

	return "", false
}

// Set armazena em ambos os níveis
func (mlc *MultiLevelCache) Set(ctx context.Context, key, value string, ttl time.Duration) {
	// Armazenar em L1
	mlc.setToL1(key, value, ttl)

	// Armazenar em L2 (Redis) se disponível
	if mlc.l2Cache != nil {
		mlc.l2Cache.Set(ctx, key, value, ttl)
	}
}

// getFromL1 busca apenas no cache L1
func (mlc *MultiLevelCache) getFromL1(key string) (string, bool) {
	mlc.l1Mutex.RLock()
	defer mlc.l1Mutex.RUnlock()

	item, exists := mlc.l1Cache[key]
	if !exists {
		return "", false
	}

	if item.IsExpired() {
		// Remove item expirado em background
		go mlc.removeFromL1(key)
		return "", false
	}

	return item.Value, true
}

// setToL1 armazena apenas no cache L1
func (mlc *MultiLevelCache) setToL1(key, value string, ttl time.Duration) {
	mlc.l1Mutex.Lock()
	defer mlc.l1Mutex.Unlock()

	// Verificar limite de tamanho
	if len(mlc.l1Cache) >= mlc.l1MaxSize {
		mlc.evictOldestL1()
	}

	mlc.l1Cache[key] = &CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// removeFromL1 remove um item do cache L1
func (mlc *MultiLevelCache) removeFromL1(key string) {
	mlc.l1Mutex.Lock()
	defer mlc.l1Mutex.Unlock()
	delete(mlc.l1Cache, key)
}

// evictOldestL1 remove 10% dos itens mais antigos quando L1 atinge limite
// Otimizado: usa heap para O(N log k) onde k é a quantidade a remover
func (mlc *MultiLevelCache) evictOldestL1() {
	toRemove := mlc.l1MaxSize / 10
	if toRemove < 1 {
		toRemove = 1
	}

	// Use partial heap para encontrar os K menores elementos em O(N log k)
	// mais eficiente que ordenar tudo O(N log N)
	items := make(evictionHeap, 0, len(mlc.l1Cache))
	for key, item := range mlc.l1Cache {
		items = append(items, evictionItem{key: key, expiresAt: item.ExpiresAt})
	}

	// Heapify é O(N)
	heap.Init(&items)

	// Remover os K itens mais antigos em O(k log N)
	removedCount := 0
	for removedCount < toRemove && items.Len() > 0 {
		oldest := heap.Pop(&items).(evictionItem)
		delete(mlc.l1Cache, oldest.key)
		removedCount++
	}
}

// CleanupExpired remove itens expirados do L1 (deve ser chamado periodicamente)
func (mlc *MultiLevelCache) CleanupExpired() {
	mlc.l1Mutex.Lock()
	defer mlc.l1Mutex.Unlock()

	now := time.Now()
	for key, item := range mlc.l1Cache {
		if now.After(item.ExpiresAt) {
			delete(mlc.l1Cache, key)
		}
	}
}

// Stats retorna estatísticas do cache
func (mlc *MultiLevelCache) Stats() map[string]interface{} {
	mlc.l1Mutex.RLock()
	defer mlc.l1Mutex.RUnlock()

	return map[string]interface{}{
		"l1_size":        len(mlc.l1Cache),
		"l1_max_size":    mlc.l1MaxSize,
		"l1_usage":       float64(len(mlc.l1Cache)) / float64(mlc.l1MaxSize) * 100,
		"l1_default_ttl": mlc.l1DefaultTTL.String(),
	}
}

// StartCleanupWorker inicia worker em background para limpeza periódica
func (mlc *MultiLevelCache) StartCleanupWorker(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 10 * time.Minute // default
	}

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				mlc.CleanupExpired()
			}
		}
	}()
}
