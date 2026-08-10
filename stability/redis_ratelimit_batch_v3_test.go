package stability

import (
	"context"
	"sync"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisRateLimiterBatchV3_SameKeyConcurrentIsolation(t *testing.T) {
	r := newTestRedisRateLimiterBatchV3(t, RedisRateLimiterBatchV3Config{
		Limit:            0.000001,
		Burst:            2,
		PipelineMaxWait:  200 * time.Microsecond,
		PipelineMaxBatch: 128,
		PipelineShards:   4,
	})
	defer r.Close()

	const requests = 12
	results := make(chan bool, requests)

	var wg sync.WaitGroup
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- r.Allow(context.Background(), "same-key")
		}()
	}

	wg.Wait()
	close(results)

	allowed := 0
	for ok := range results {
		if ok {
			allowed++
		}
	}

	if allowed != 2 {
		t.Fatalf("expected exactly 2 allowed requests, got %d", allowed)
	}
}

func TestRedisRateLimiterBatchV3_MultiKeyIsolation(t *testing.T) {
	r := newTestRedisRateLimiterBatchV3(t, RedisRateLimiterBatchV3Config{
		Limit:            0.000001,
		Burst:            1,
		PipelineMaxWait:  200 * time.Microsecond,
		PipelineMaxBatch: 128,
		PipelineShards:   4,
	})
	defer r.Close()

	keys := []string{"k1", "k2", "k3"}
	perKeyAllowed := make(map[string]int, len(keys))

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, key := range keys {
		for i := 0; i < 2; i++ {
			wg.Add(1)
			k := key
			go func() {
				defer wg.Done()
				ok := r.Allow(context.Background(), k)
				if !ok {
					return
				}
				mu.Lock()
				perKeyAllowed[k]++
				mu.Unlock()
			}()
		}
	}

	wg.Wait()

	for _, key := range keys {
		if perKeyAllowed[key] != 1 {
			t.Fatalf("expected key %s to have exactly 1 allowed request, got %d", key, perKeyAllowed[key])
		}
	}
}

func TestRedisRateLimiterBatchV3_StatsByUnits(t *testing.T) {
	r := newTestRedisRateLimiterBatchV3(t, RedisRateLimiterBatchV3Config{
		Limit:            100,
		Burst:            2,
		PipelineMaxWait:  50 * time.Microsecond,
		PipelineMaxBatch: 32,
		PipelineShards:   4,
	})
	defer r.Close()

	if !r.Allow(context.Background(), "k1") {
		t.Fatal("expected first request to be allowed")
	}
	if !r.Allow(context.Background(), "k1") {
		t.Fatal("expected second request to be allowed")
	}
	if r.Allow(context.Background(), "k1") {
		t.Fatal("expected third request to be denied")
	}

	stats := r.Stats()
	if stats.TotalRequests != 3 {
		t.Fatalf("expected TotalRequests=3, got %d", stats.TotalRequests)
	}
	if stats.AllowedRequests != 2 {
		t.Fatalf("expected AllowedRequests=2, got %d", stats.AllowedRequests)
	}
}

func newTestRedisRateLimiterBatchV3(t *testing.T, cfg RedisRateLimiterBatchV3Config) *RedisRateLimiterBatchV3 {
	t.Helper()

	srv, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis failed: %v", err)
	}
	t.Cleanup(srv.Close)

	client := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("redis ping failed: %v", err)
	}

	limiter, err := NewRedisRateLimiterBatchV3(client, cfg)
	if err != nil {
		t.Fatalf("new redis limiter batch v3 failed: %v", err)
	}
	return limiter
}
