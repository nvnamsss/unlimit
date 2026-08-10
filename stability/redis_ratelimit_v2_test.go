package stability

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisRateLimiterV2_AllowPipeline(t *testing.T) {
	r := newTestRedisRateLimiterV2(t, RedisRateLimiterV2Config{
		Limit:            100000,
		Burst:            1000,
		PipelineEnabled:  true,
		PipelineMaxWait:  25 * time.Microsecond,
		PipelineMaxBatch: 128,
		PipelineShards:   8,
	})

	ctx := context.Background()
	if !r.Allow(ctx, "user:1") {
		t.Fatal("expected first request to be allowed")
	}
}

func TestRedisRateLimiterV2_AllowNExceedBurst(t *testing.T) {
	r := newTestRedisRateLimiterV2(t, RedisRateLimiterV2Config{
		Limit:           100,
		Burst:           5,
		PipelineEnabled: false,
	})

	if r.AllowN(context.Background(), "user:1", 6) {
		t.Fatal("expected request above burst to be denied")
	}
}

func TestRedisRateLimiterV2_StatsAllowedRequests(t *testing.T) {
	r := newTestRedisRateLimiterV2(t, RedisRateLimiterV2Config{
		Limit:           100,
		Burst:           5,
		PipelineEnabled: false,
	})

	if !r.AllowN(context.Background(), "user:1", 2) {
		t.Fatal("expected allow for n=2")
	}

	stats := r.Stats()
	if stats.AllowedRequests != 2 {
		t.Fatalf("expected AllowedRequests=2, got %d", stats.AllowedRequests)
	}
	if stats.TotalRequests != 2 {
		t.Fatalf("expected TotalRequests=2, got %d", stats.TotalRequests)
	}
}

func TestRedisPipelineV2_HandleQueueFull(t *testing.T) {
	p := &redisPipelineV2{
		pipelineQueue:  make(chan *redisPipelineV2Request, 1),
		pipelineStopCh: make(chan struct{}),
	}

	p.pipelineQueue <- &redisPipelineV2Request{}

	_, err := p.handle(context.Background(), &redisPipelineV2Request{})
	if err != errRedisPipelineV2QueueFull {
		t.Fatalf("expected queue full error, got: %v", err)
	}
}

func newTestRedisRateLimiterV2(t *testing.T, cfg RedisRateLimiterV2Config) *RedisRateLimiterV2 {
	t.Helper()

	srv, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis failed: %v", err)
	}
	t.Cleanup(srv.Close)

	client := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("redis ping failed: %v", err)
	}

	limiter, err := NewRedisRateLimiterV2(client, cfg)
	if err != nil {
		t.Fatalf("new redis limiter v2 failed: %v", err)
	}
	t.Cleanup(limiter.Close)

	return limiter
}
