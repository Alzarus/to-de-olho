package cache

import (
	"context"
	"os"
	"time"

	"to-de-olho-backend/internal/config"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	rdb *redis.Client
}

func New() *Cache {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	password := os.Getenv("REDIS_PASSWORD")

	c := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           0,
		ReadTimeout:  500 * time.Millisecond,
		WriteTimeout: 500 * time.Millisecond,
		DialTimeout:  500 * time.Millisecond,
	})
	return &Cache{rdb: c}
}

// NewFromConfig creates a new cache instance from config
func NewFromConfig(cfg *config.RedisConfig) *Cache {
	c := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		DialTimeout:  cfg.DialTimeout,
	})
	return &Cache{rdb: c}
}

func (c *Cache) Get(ctx context.Context, key string) (string, bool) {
	if c == nil || c.rdb == nil {
		return "", false
	}
	v, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", false
	}
	return v, true
}

func (c *Cache) Set(ctx context.Context, key, value string, ttl time.Duration) {
	if c == nil || c.rdb == nil {
		return
	}
	_ = c.rdb.Set(ctx, key, value, ttl).Err()
}
