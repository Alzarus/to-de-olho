package cache

import (
	"context"
	"testing"
	"time"
)

func TestMultiLevelCache_BasicOperations(t *testing.T) {
	// Setup
	redisCache := &Cache{rdb: nil} // Mock Redis
	mlCache := NewMultiLevelCache(redisCache, 100, 1*time.Minute)
	ctx := context.Background()

	// Test Set and Get
	mlCache.Set(ctx, "test-key", "test-value", 5*time.Minute)

	value, found := mlCache.Get(ctx, "test-key")
	if !found {
		t.Error("Item should be found in cache")
	}
	if value != "test-value" {
		t.Errorf("Expected 'test-value', got '%s'", value)
	}
}

func TestMultiLevelCache_L1Priority(t *testing.T) {
	// Setup
	redisCache := &Cache{rdb: nil}
	mlCache := NewMultiLevelCache(redisCache, 100, 1*time.Minute)
	ctx := context.Background()

	// Store in cache
	mlCache.Set(ctx, "priority-test", "l1-value", 5*time.Minute)

	// First Get should hit L1
	value, found := mlCache.Get(ctx, "priority-test")
	if !found || value != "l1-value" {
		t.Error("First get should hit L1 cache")
	}

	// Verify it's in L1
	l1Value, l1Found := mlCache.getFromL1("priority-test")
	if !l1Found || l1Value != "l1-value" {
		t.Error("Value should be in L1 cache")
	}
}

func TestMultiLevelCache_TTLExpiration(t *testing.T) {
	// Setup
	redisCache := &Cache{rdb: nil}
	mlCache := NewMultiLevelCache(redisCache, 100, 1*time.Minute)
	ctx := context.Background()

	// Store with very short TTL
	mlCache.Set(ctx, "expire-test", "expire-value", 1*time.Millisecond)

	// Wait for expiration
	time.Sleep(2 * time.Millisecond)

	// Should not be found
	_, found := mlCache.Get(ctx, "expire-test")
	if found {
		t.Error("Expired item should not be found")
	}
}

func TestMultiLevelCache_EvictionPolicy(t *testing.T) {
	// Setup with very small cache size
	redisCache := &Cache{rdb: nil}
	mlCache := NewMultiLevelCache(redisCache, 5, 1*time.Minute)
	ctx := context.Background()

	// Fill cache beyond capacity
	for i := 0; i < 10; i++ {
		key := string(rune('a' + i))
		mlCache.Set(ctx, key, "value"+key, 5*time.Minute)
	}

	// Verify cache size is within limits
	stats := mlCache.Stats()
	l1Size := stats["l1_size"].(int)
	if l1Size > 5 {
		t.Errorf("L1 cache size should not exceed 5, got %d", l1Size)
	}
}

func TestMultiLevelCache_CleanupExpired(t *testing.T) {
	// Setup
	redisCache := &Cache{rdb: nil}
	mlCache := NewMultiLevelCache(redisCache, 100, 1*time.Minute)
	ctx := context.Background()

	// Add items with different TTLs
	mlCache.Set(ctx, "short-ttl", "value1", 1*time.Millisecond)
	mlCache.Set(ctx, "long-ttl", "value2", 1*time.Hour)

	// Wait for short TTL to expire
	time.Sleep(2 * time.Millisecond)

	// Run cleanup
	mlCache.CleanupExpired()

	// Verify short TTL item is gone
	_, found := mlCache.getFromL1("short-ttl")
	if found {
		t.Error("Expired item should be removed by cleanup")
	}

	// Verify long TTL item remains
	_, found = mlCache.getFromL1("long-ttl")
	if !found {
		t.Error("Non-expired item should remain after cleanup")
	}
}

func TestMultiLevelCache_Stats(t *testing.T) {
	// Setup
	redisCache := &Cache{rdb: nil}
	mlCache := NewMultiLevelCache(redisCache, 100, 2*time.Minute)
	ctx := context.Background()

	// Add some items
	mlCache.Set(ctx, "stats1", "value1", 5*time.Minute)
	mlCache.Set(ctx, "stats2", "value2", 5*time.Minute)

	// Get stats
	stats := mlCache.Stats()

	// Verify stats structure
	if _, ok := stats["l1_size"]; !ok {
		t.Error("Stats should include l1_size")
	}
	if _, ok := stats["l1_max_size"]; !ok {
		t.Error("Stats should include l1_max_size")
	}
	if _, ok := stats["l1_usage"]; !ok {
		t.Error("Stats should include l1_usage")
	}

	// Verify values
	if stats["l1_size"].(int) != 2 {
		t.Errorf("Expected l1_size to be 2, got %v", stats["l1_size"])
	}
	if stats["l1_max_size"].(int) != 100 {
		t.Errorf("Expected l1_max_size to be 100, got %v", stats["l1_max_size"])
	}
}

func TestMultiLevelCache_DefaultValues(t *testing.T) {
	// Test with invalid parameters
	redisCache := &Cache{rdb: nil}
	mlCache := NewMultiLevelCache(redisCache, 0, 0)

	// Should use defaults
	stats := mlCache.Stats()
	if stats["l1_max_size"].(int) != 1000 {
		t.Errorf("Expected default max size 1000, got %v", stats["l1_max_size"])
	}
}

func TestMultiLevelCache_ConcurrentAccess(t *testing.T) {
	// Setup
	redisCache := &Cache{rdb: nil}
	mlCache := NewMultiLevelCache(redisCache, 100, 1*time.Minute)
	ctx := context.Background()

	// Concurrent writes and reads
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 50; i++ {
			key := "concurrent" + string(rune('0'+i%10))
			mlCache.Set(ctx, key, "value", 1*time.Minute)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 50; i++ {
			key := "concurrent" + string(rune('0'+i%10))
			mlCache.Get(ctx, key)
		}
		done <- true
	}()

	// Wait for completion
	<-done
	<-done

	// Should not panic or race
	t.Log("Concurrent access test completed successfully")
}

// Benchmark tests
func BenchmarkMultiLevelCache_Get_L1Hit(b *testing.B) {
	redisCache := &Cache{rdb: nil}
	mlCache := NewMultiLevelCache(redisCache, 1000, 10*time.Minute)
	ctx := context.Background()

	// Pre-populate L1
	mlCache.Set(ctx, "bench-key", "bench-value", 5*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mlCache.Get(ctx, "bench-key")
	}
}

func BenchmarkMultiLevelCache_Set(b *testing.B) {
	redisCache := &Cache{rdb: nil}
	mlCache := NewMultiLevelCache(redisCache, 1000, 10*time.Minute)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "bench-key-" + string(rune('a'+i%26))
		mlCache.Set(ctx, key, "bench-value", 5*time.Minute)
	}
}

func BenchmarkMultiLevelCache_GetMiss(b *testing.B) {
	redisCache := &Cache{rdb: nil}
	mlCache := NewMultiLevelCache(redisCache, 1000, 10*time.Minute)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "miss-key-" + string(rune('a'+i%26))
		mlCache.Get(ctx, key)
	}
}
