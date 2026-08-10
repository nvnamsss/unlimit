package stability

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisRateLimiterBatchV4_RequestIDCorrelation(t *testing.T) {
	r := newTestRedisRateLimiterBatchV4(t, RedisRateLimiterBatchV4Config{
		Limit:            1_000_000,
		Burst:            1_000_000,
		PipelineMaxWait:  200 * time.Microsecond,
		PipelineMaxBatch: 128,
		PipelineShards:   8,
	})
	defer r.Close()

	const total = 300
	requestIDs := make(chan uint64, total)
	errs := make(chan error, total)

	var wg sync.WaitGroup
	for i := 0; i < total; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			allowed, requestID, err := r.allowNWithRequestID(context.Background(), fmt.Sprintf("k-%d", i%29), 1)
			if err != nil {
				errs <- err
				return
			}
			if !allowed {
				errs <- fmt.Errorf("request %d denied unexpectedly", i)
				return
			}
			if requestID == 0 {
				errs <- fmt.Errorf("request %d returned empty requestID", i)
				return
			}

			requestIDs <- requestID
		}(i)
	}

	wg.Wait()
	close(errs)
	close(requestIDs)

	for err := range errs {
		if err != nil {
			t.Fatalf("requestID correlation failed: %v", err)
		}
	}

	seen := make(map[uint64]struct{}, total)
	for id := range requestIDs {
		if _, exists := seen[id]; exists {
			t.Fatalf("duplicate requestID observed: %d", id)
		}
		seen[id] = struct{}{}
	}

	if len(seen) != total {
		t.Fatalf("expected %d unique requestIDs, got %d", total, len(seen))
	}

	stats := r.Stats()
	if stats.CorrelationMismatches != 0 {
		t.Fatalf("expected CorrelationMismatches=0, got %d", stats.CorrelationMismatches)
	}
}

func TestRedisRateLimiterBatchV4_BatchingStatsUnderLoad(t *testing.T) {
	r := newTestRedisRateLimiterBatchV4(t, RedisRateLimiterBatchV4Config{
		Limit:            1_000_000,
		Burst:            1_000_000,
		PipelineMaxWait:  500 * time.Microsecond,
		PipelineMaxBatch: 256,
		PipelineShards:   4,
	})
	defer r.Close()

	const total = 500
	var wg sync.WaitGroup

	for i := 0; i < total; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if !r.Allow(context.Background(), fmt.Sprintf("load-%d", i)) {
				t.Errorf("request %d denied unexpectedly", i)
			}
		}(i)
	}

	wg.Wait()

	stats := r.Stats()
	if stats.PipelineRequests != total {
		t.Fatalf("expected PipelineRequests=%d, got %d", total, stats.PipelineRequests)
	}
	if stats.PipelineBatches == 0 {
		t.Fatal("expected PipelineBatches > 0")
	}
	if stats.PipelineBatches >= stats.PipelineRequests {
		t.Fatalf("expected fewer batches than requests, got batches=%d requests=%d", stats.PipelineBatches, stats.PipelineRequests)
	}
}

func TestRedisRateLimiterBatchV4_StatsByUnits(t *testing.T) {
	r := newTestRedisRateLimiterBatchV4(t, RedisRateLimiterBatchV4Config{
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

func TestRedisRateLimiterBatchV4_InitializeTokenBucket_AllowsConfiguredTokens(t *testing.T) {
	r := newTestRedisRateLimiterBatchV4(t, RedisRateLimiterBatchV4Config{
		Limit:            1e-9,
		Burst:            10,
		PipelineMaxWait:  50 * time.Microsecond,
		PipelineMaxBatch: 32,
		PipelineShards:   4,
	})
	defer r.Close()

	if err := r.InitializeTokenBucket(context.Background(), "init-key", 3); err != nil {
		t.Fatalf("initialize token bucket failed: %v", err)
	}

	for i := 0; i < 3; i++ {
		if !r.Allow(context.Background(), "init-key") {
			t.Fatalf("expected allow %d to be true", i+1)
		}
	}

	if r.Allow(context.Background(), "init-key") {
		t.Fatal("expected 4th allow to be false")
	}
}

func TestRedisRateLimiterBatchV4_InitializeTokenBucket_ClampsToBurst(t *testing.T) {
	r := newTestRedisRateLimiterBatchV4(t, RedisRateLimiterBatchV4Config{
		Limit:            1e-9,
		Burst:            5,
		PipelineMaxWait:  50 * time.Microsecond,
		PipelineMaxBatch: 32,
		PipelineShards:   4,
	})
	defer r.Close()

	if err := r.InitializeTokenBucket(context.Background(), "clamp-key", 100); err != nil {
		t.Fatalf("initialize token bucket failed: %v", err)
	}

	for i := 0; i < 5; i++ {
		if !r.Allow(context.Background(), "clamp-key") {
			t.Fatalf("expected allow %d to be true", i+1)
		}
	}

	if r.Allow(context.Background(), "clamp-key") {
		t.Fatal("expected 6th allow to be false")
	}
}

func TestRedisRateLimiterBatchV4_CloseUnblocksPendingRequests(t *testing.T) {
	r := newTestRedisRateLimiterBatchV4(t, RedisRateLimiterBatchV4Config{
		Limit:            1_000_000,
		Burst:            1_000_000,
		PipelineMaxWait:  50 * time.Millisecond,
		PipelineMaxBatch: 512,
		PipelineShards:   4,
	})

	const workers = 64
	results := make(chan bool, workers)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			results <- r.Allow(ctx, fmt.Sprintf("pending-%d", i%8))
		}(i)
	}

	time.Sleep(3 * time.Millisecond)
	r.Close()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("pending requests blocked after Close")
	}

	close(results)
	for allowed := range results {
		if allowed {
			continue
		}
	}
}

func TestRedisRateLimiterBatchV4_FailureModeMirrorsV3(t *testing.T) {
	closedClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	t.Cleanup(func() { _ = closedClient.Close() })

	failClosed, err := NewRedisRateLimiterBatchV4(closedClient, RedisRateLimiterBatchV4Config{
		Limit:            100,
		Burst:            10,
		FailureMode:      RedisRateLimiterFailClosed,
		OperationTimeout: 20 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("new fail-closed limiter failed: %v", err)
	}
	defer failClosed.Close()

	if failClosed.Allow(context.Background(), "fail-closed") {
		t.Fatal("expected fail-closed limiter to deny on redis error")
	}

	failOpen, err := NewRedisRateLimiterBatchV4(closedClient, RedisRateLimiterBatchV4Config{
		Limit:            100,
		Burst:            10,
		FailureMode:      RedisRateLimiterFailOpen,
		OperationTimeout: 20 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("new fail-open limiter failed: %v", err)
	}
	defer failOpen.Close()

	if !failOpen.Allow(context.Background(), "fail-open") {
		t.Fatal("expected fail-open limiter to allow on redis error")
	}
}

func newTestRedisRateLimiterBatchV4(t *testing.T, cfg RedisRateLimiterBatchV4Config) *RedisRateLimiterBatchV4 {
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

	limiter, err := NewRedisRateLimiterBatchV4(client, cfg)
	if err != nil {
		t.Fatalf("new redis limiter batch v4 failed: %v", err)
	}
	return limiter
}
