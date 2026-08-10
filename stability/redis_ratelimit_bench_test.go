package stability

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

var redisRateLimiterAllowSink bool
var redisRateLimiterTokenSink float64

const benchmarkTargetRateLimit = 100000.0

type redisBenchmarkMetrics struct {
	requests         int64
	pipelineRequests int64
	pipelineCommands int64
	pipelineTotalNs  int64
	pipelineMaxNs    int64
}

func (m *redisBenchmarkMetrics) snapshot() (requests int64, pipelineRequests int64, pipelineCommands int64, pipelineTotalNs int64, pipelineMaxNs int64) {
	return atomic.LoadInt64(&m.requests), atomic.LoadInt64(&m.pipelineRequests), atomic.LoadInt64(&m.pipelineCommands), atomic.LoadInt64(&m.pipelineTotalNs), atomic.LoadInt64(&m.pipelineMaxNs)
}

func (m *redisBenchmarkMetrics) updatePipelineMax(durationNs int64) {
	for {
		current := atomic.LoadInt64(&m.pipelineMaxNs)
		if durationNs <= current {
			return
		}
		if atomic.CompareAndSwapInt64(&m.pipelineMaxNs, current, durationNs) {
			return
		}
	}
}

type redisBenchmarkHook struct {
	metrics *redisBenchmarkMetrics
}

func (h redisBenchmarkHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (h redisBenchmarkHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		atomic.AddInt64(&h.metrics.requests, 1)
		return next(ctx, cmd)
	}
}

func (h redisBenchmarkHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		start := time.Now()
		atomic.AddInt64(&h.metrics.pipelineRequests, 1)
		atomic.AddInt64(&h.metrics.pipelineCommands, int64(len(cmds)))
		err := next(ctx, cmds)
		durationNs := time.Since(start).Nanoseconds()
		atomic.AddInt64(&h.metrics.pipelineTotalNs, durationNs)
		h.metrics.updatePipelineMax(durationNs)
		return err
	}
}

func BenchmarkRedisRateLimiter_Allow(b *testing.B) {
	for _, burst := range []int{1, 10, 1000} {
		b.Run(fmt.Sprintf("Burst_%d", burst), func(b *testing.B) {
			limiter := newBenchmarkRedisRateLimiter(b, RedisRateLimiterConfig{
				Limit: 1000000,
				Burst: burst,
				Name:  fmt.Sprintf("allow_%d", burst),
			})
			ctx := context.Background()

			b.ReportAllocs()
			for b.Loop() {
				// Keep setup out of timing so this benchmark isolates Allow.
				b.StopTimer()
				b.StartTimer()

				redisRateLimiterAllowSink = limiter.Allow(ctx, "bench")
			}

			if !redisRateLimiterAllowSink && !testing.Short() {
				b.Fatal("unexpected deny in allow benchmark")
			}
		})
	}
}

func BenchmarkRedisRateLimiter_AllowN(b *testing.B) {
	for _, n := range []int{1, 4, 16} {
		b.Run(fmt.Sprintf("N_%d", n), func(b *testing.B) {
			limiter := newBenchmarkRedisRateLimiter(b, RedisRateLimiterConfig{
				Limit: 1000000,
				Burst: 100,
				Name:  fmt.Sprintf("allow_n_%d", n),
			})
			ctx := context.Background()

			b.ReportAllocs()
			for b.Loop() {
				b.StopTimer()
				limiter.Reset()
				b.StartTimer()

				redisRateLimiterAllowSink = limiter.AllowN(ctx, "bench", n)
			}

			if !redisRateLimiterAllowSink && !testing.Short() {
				b.Fatal("unexpected deny in allowN benchmark")
			}
		})
	}
}

func BenchmarkRedisRateLimiter_Tokens(b *testing.B) {
	limiter := newBenchmarkRedisRateLimiter(b, RedisRateLimiterConfig{
		Limit: 1000000,
		Burst: 1000,
		Name:  "tokens",
	})

	b.ReportAllocs()
	for b.Loop() {
		redisRateLimiterTokenSink = limiter.Tokens()
	}

	if redisRateLimiterTokenSink < 0 && !testing.Short() {
		b.Fatal("unexpected negative token count")
	}
}

func BenchmarkRedisRateLimiter_WaitImmediate(b *testing.B) {
	limiter := newBenchmarkRedisRateLimiter(b, RedisRateLimiterConfig{
		Limit: 1000000,
		Burst: 1,
		Name:  "wait_immediate",
	})
	ctx := context.Background()

	b.ReportAllocs()
	for b.Loop() {
		b.StopTimer()
		limiter.Reset()
		b.StartTimer()

		if err := limiter.Wait(ctx); err != nil {
			b.Fatalf("wait failed: %v", err)
		}
	}
}

func BenchmarkRedisRateLimiter_AllowParallel(b *testing.B) {
	limiter := newBenchmarkRedisRateLimiter(b, RedisRateLimiterConfig{
		Limit: 1000000,
		Burst: 1000000000,
		Name:  "allow_parallel",
	})
	ctx := context.Background()

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			redisRateLimiterAllowSink = limiter.Allow(ctx, "parallel")
		}
	})
}

func BenchmarkRedisRateLimiter_AllowModeComparison(b *testing.B) {
	b.Run("Direct", func(b *testing.B) {
		limiter, metrics := newBenchmarkRedisRateLimiterWithMetrics(b, RedisRateLimiterConfig{
			Limit: benchmarkTargetRateLimit,
			Burst: 1000000000,
			Name:  "mode_direct",
		})
		baseReq, basePipelineReq, basePipelineCmd, _, _ := metrics.snapshot()

		var ops int64
		start := time.Now()
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			localOps := 0
			for pb.Next() {
				redisRateLimiterAllowSink = limiter.Allow(context.Background(), "direct")
				localOps++
			}
			atomic.AddInt64(&ops, int64(localOps))
		})
		elapsed := time.Since(start)
		if elapsed > 0 {
			b.ReportMetric(float64(ops)/elapsed.Seconds(), "req/s")
		}
		reqNow, pipelineReqNow, pipelineCmdNow, _, _ := metrics.snapshot()
		redisReqTotal := (reqNow - baseReq) + (pipelineReqNow - basePipelineReq)
		pipelineReqTotal := pipelineReqNow - basePipelineReq
		pipelineCmdTotal := pipelineCmdNow - basePipelineCmd
		b.ReportMetric(float64(redisReqTotal), "redis-req-total")
		b.ReportMetric(float64(pipelineReqTotal), "redis-pipeline-req-total")
		b.ReportMetric(float64(pipelineCmdTotal), "redis-pipeline-cmd-total")
		if redisReqTotal > 0 {
			b.ReportMetric(float64(ops)/float64(redisReqTotal), "allow-per-redis-req")
		}
		b.ReportMetric(benchmarkTargetRateLimit, "target-req/s")
		b.ReportMetric(float64(ops), "req-total")
	})

	b.Run("Pipeline", func(b *testing.B) {
		limiter, metrics := newBenchmarkRedisRateLimiterWithMetrics(b, RedisRateLimiterConfig{
			Limit:            benchmarkTargetRateLimit,
			Burst:            1000000000,
			Name:             "mode_pipeline",
			PipelineEnabled:  true,
			PipelineMaxWait:  10 * time.Microsecond,
			PipelineMaxBatch: 1000,
		})
		baseReq, basePipelineReq, basePipelineCmd, basePipelineTotalNs, basePipelineMaxNs := metrics.snapshot()

		var ops int64
		start := time.Now()
		b.ReportAllocs()
		for b.Loop() {
			redisRateLimiterAllowSink = limiter.Allow(context.Background(), "pipeline")
			atomic.AddInt64(&ops, 1)
		}

		// b.RunParallel(func(pb *testing.PB) {
		// 	localOps := 0
		// 	for pb.Next() {
		// 		redisRateLimiterAllowSink = limiter.Allow()
		// 		localOps++
		// 	}
		// 	atomic.AddInt64(&ops, int64(localOps))
		// })
		elapsed := time.Since(start)
		if elapsed > 0 {
			b.ReportMetric(float64(ops)/elapsed.Seconds(), "req/s")
		}
		reqNow, pipelineReqNow, pipelineCmdNow, pipelineTotalNsNow, pipelineMaxNsNow := metrics.snapshot()
		redisReqTotal := (reqNow - baseReq) + (pipelineReqNow - basePipelineReq)
		pipelineReqTotal := pipelineReqNow - basePipelineReq
		pipelineCmdTotal := pipelineCmdNow - basePipelineCmd
		pipelineTotalNs := pipelineTotalNsNow - basePipelineTotalNs
		pipelineMaxNs := pipelineMaxNsNow
		if pipelineMaxNs < basePipelineMaxNs {
			pipelineMaxNs = basePipelineMaxNs
		}
		b.ReportMetric(float64(redisReqTotal), "redis-req-total")
		b.ReportMetric(float64(pipelineReqTotal), "redis-pipeline-req-total")
		b.ReportMetric(float64(pipelineCmdTotal), "redis-pipeline-cmd-total")
		if pipelineReqTotal > 0 {
			b.ReportMetric(float64(pipelineTotalNs)/float64(pipelineReqTotal)/1000.0, "pipeline-avg-us")
			b.ReportMetric(float64(pipelineMaxNs)/1000.0, "pipeline-max-us")
			if elapsed > 0 {
				b.ReportMetric(float64(pipelineReqTotal)/elapsed.Seconds(), "pipeline-rtt/s")
			}
		}
		if redisReqTotal > 0 {
			b.ReportMetric(float64(ops)/float64(redisReqTotal), "allow-per-redis-req")
		}
		b.ReportMetric(benchmarkTargetRateLimit, "target-req/s")
		b.ReportMetric(float64(ops), "req-total")
	})
}

func newBenchmarkRedisRateLimiter(b *testing.B, cfg RedisRateLimiterConfig) *RedisRateLimiter {
	limiter, _ := newBenchmarkRedisRateLimiterWithMetrics(b, cfg)
	return limiter
}

func newBenchmarkRedisRateLimiterWithMetrics(b *testing.B, cfg RedisRateLimiterConfig) (*RedisRateLimiter, *redisBenchmarkMetrics) {
	b.Helper()

	srv, err := miniredis.Run()
	if err != nil {
		b.Fatalf("start miniredis failed: %v", err)
	}
	b.Cleanup(srv.Close)

	client := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	b.Cleanup(func() {
		_ = client.Close()
	})

	metrics := &redisBenchmarkMetrics{}
	client.AddHook(redisBenchmarkHook{metrics: metrics})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		b.Fatalf("redis ping failed: %v", err)
	}

	limiter, err := NewRedisRateLimiter(client, cfg)
	if err != nil {
		b.Fatalf("new redis limiter failed: %v", err)
	}
	b.Cleanup(limiter.Close)

	return limiter, metrics
}
