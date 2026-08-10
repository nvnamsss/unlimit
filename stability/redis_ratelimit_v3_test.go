package stability

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisRateLimiterV3_StatsByUnits(t *testing.T) {
	r := newTestRedisRateLimiterV3(t, RedisRateLimiterV3Config{
		Limit: 100,
		Burst: 2,
	})

	if !r.AllowN(context.Background(), "k1", 2) {
		t.Fatal("expected allow for initial tokens")
	}
	if r.Allow(context.Background(), "k1") {
		t.Fatal("expected deny when tokens are exhausted")
	}

	stats := r.Stats()
	if stats.TotalRequests != 3 {
		t.Fatalf("expected TotalRequests=3, got %d", stats.TotalRequests)
	}
	if stats.AllowedRequests != 2 {
		t.Fatalf("expected AllowedRequests=2, got %d", stats.AllowedRequests)
	}
}

func newTestRedisRateLimiterV3(t *testing.T, cfg RedisRateLimiterV3Config) *RedisRateLimiterV3 {
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

	limiter, err := NewRedisRateLimiterV3(client, cfg)
	if err != nil {
		t.Fatalf("new redis limiter v3 failed: %v", err)
	}
	return limiter
}
