package ranking

import (
	"sync"
	"time"
)

// InMemoryCache provides a simple thread-safe cache for ranking data
type InMemoryCache struct {
	data  map[string]cacheItem
	mutex sync.RWMutex
}

type cacheItem struct {
	Response *RankingResponse
	Expiry   time.Time
}

// Global cache instance (package-level or injected)
var localCache = &InMemoryCache{
	data: make(map[string]cacheItem),
}

func (c *InMemoryCache) Get(key string) *RankingResponse {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return nil
	}

	if time.Now().After(item.Expiry) {
		return nil
	}

	return item.Response
}

func (c *InMemoryCache) Set(key string, response *RankingResponse, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = cacheItem{
		Response: response,
		Expiry:   time.Now().Add(ttl),
	}
}
