package middleware

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	limiter  *tokenBucket
	lastSeen time.Time
}

type tokenBucket struct {
	capacity   int
	remaining  int
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

func newBucket(capacity int, per time.Duration) *tokenBucket {
	return &tokenBucket{
		capacity:   capacity,
		remaining:  capacity,
		refillRate: float64(capacity) / per.Seconds(),
		lastRefill: time.Now(),
	}
}

func (b *tokenBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	// refill
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	refilled := int(elapsed * b.refillRate)
	if refilled > 0 {
		b.remaining = min(b.capacity, b.remaining+refilled)
		b.lastRefill = now
	}
	if b.remaining <= 0 {
		return false
	}
	b.remaining--
	return true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RateLimitPerIP simple middleware: capacity tokens per window duration
func RateLimitPerIP(capacity int, window time.Duration) gin.HandlerFunc {
	visitors := make(map[string]*visitor)
	mu := sync.Mutex{}
	cleanupTicker := time.NewTicker(5 * time.Minute)
	go func() {
		for range cleanupTicker.C {
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > 10*time.Minute {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := clientIP(c.Request)
		mu.Lock()
		v, ok := visitors[ip]
		if !ok {
			v = &visitor{limiter: newBucket(capacity, window), lastSeen: time.Now()}
			visitors[ip] = v
		}
		v.lastSeen = time.Now()
		mu.Unlock()

		if !v.limiter.allow() {
			c.Header("Retry-After", strconv.Itoa(int(window.Seconds())))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}

func clientIP(r *http.Request) string {
	// Trust X-Forwarded-For if present; else RemoteAddr
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
