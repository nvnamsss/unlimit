package stability

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	defaultRedisRateLimiterV3Prefix   = "stability:ratelimit"
	defaultRedisRateLimiterV3Name     = "global"
	defaultRedisRateLimiterV3Timeout  = 200 * time.Millisecond
	defaultRedisRateLimiterV3StateTTL = 5 * time.Minute
)

const (
	redisTokenBucketFieldTokens = "tokens"
	redisTokenBucketFieldLastNS = "last_ns"
)

// RedisRateLimiterV3Config configures a Redis-backed token-bucket limiter.
type RedisRateLimiterV3Config struct {
	Limit            float64
	Burst            int
	KeyPrefix        string
	Name             string
	FailureMode      RedisRateLimiterFailureMode
	OperationTimeout time.Duration
	StateTTL         time.Duration
}

// RedisRateLimiterV3Stats exposes v3 runtime counters.
type RedisRateLimiterV3Stats struct {
	TotalRequests   int64
	AllowedRequests int64
}

// RedisRateLimiterV3 is a Redis-backed token-bucket rate limiter.
type RedisRateLimiterV3 struct {
	client  *redis.Client
	baseKey string

	limit            float64
	burst            int
	failureMode      RedisRateLimiterFailureMode
	operationTimeout time.Duration
	stateTTL         time.Duration

	totalRequests   atomic.Int64
	allowedRequests atomic.Int64
}

// NewRedisRateLimiterV3 creates a new token-bucket limiter backed by Redis.
func NewRedisRateLimiterV3(client *redis.Client, config RedisRateLimiterV3Config) (*RedisRateLimiterV3, error) {
	if client == nil {
		return nil, errors.New("redis client is nil")
	}
	if config.Limit <= 0 {
		return nil, errors.New("limit must be greater than zero")
	}
	if config.Burst <= 0 {
		return nil, errors.New("burst must be greater than zero")
	}

	prefix := config.KeyPrefix
	if prefix == "" {
		prefix = defaultRedisRateLimiterV3Prefix
	}

	name := config.Name
	if name == "" {
		name = defaultRedisRateLimiterV3Name
	}

	failureMode := config.FailureMode
	if failureMode == "" {
		failureMode = RedisRateLimiterFailClosed
	}
	if failureMode != RedisRateLimiterFailClosed && failureMode != RedisRateLimiterFailOpen {
		return nil, fmt.Errorf("unsupported failure mode: %s", failureMode)
	}

	opTimeout := config.OperationTimeout
	if opTimeout <= 0 {
		opTimeout = defaultRedisRateLimiterV3Timeout
	}

	stateTTL := config.StateTTL
	if stateTTL <= 0 {
		stateTTL = defaultRedisRateLimiterV3StateTTL
	}

	return &RedisRateLimiterV3{
		client:           client,
		baseKey:          prefix + ":" + name,
		limit:            config.Limit,
		burst:            config.Burst,
		failureMode:      failureMode,
		operationTimeout: opTimeout,
		stateTTL:         stateTTL,
	}, nil
}

// Allow checks if one request can be admitted for the given key.
func (r *RedisRateLimiterV3) Allow(ctx context.Context, key string) bool {
	return r.AllowN(ctx, key, 1)
}

// AllowN checks if n requests can be admitted for the given key.
func (r *RedisRateLimiterV3) AllowN(ctx context.Context, key string, n int) bool {
	if n <= 0 {
		return true
	}

	r.totalRequests.Add(int64(n))
	opCtx, cancel := r.operationContext(ctx)

	allowed, err := consumeTokenBucketDirect(
		opCtx,
		cancel,
		r.client,
		r.redisKey(key),
		r.limit,
		r.burst,
		n,
		r.stateTTL,
	)
	if err != nil {
		if r.failureMode == RedisRateLimiterFailOpen {
			r.allowedRequests.Add(int64(n))
			return true
		}
		return false
	}

	if allowed {
		r.allowedRequests.Add(int64(n))
	}
	return allowed
}

// Limit returns the configured requests per second.
func (r *RedisRateLimiterV3) Limit() float64 {
	return r.limit
}

// Burst returns the configured bucket capacity.
func (r *RedisRateLimiterV3) Burst() int {
	return r.burst
}

// Stats returns a snapshot of internal counters.
func (r *RedisRateLimiterV3) Stats() RedisRateLimiterV3Stats {
	return RedisRateLimiterV3Stats{
		TotalRequests:   r.totalRequests.Load(),
		AllowedRequests: r.allowedRequests.Load(),
	}
}

func (r *RedisRateLimiterV3) redisKey(key string) string {
	if key == "" {
		return r.baseKey
	}
	return r.baseKey + ":" + key
}

func (r *RedisRateLimiterV3) operationContext(parent context.Context) (context.Context, context.CancelFunc) {
	if r.operationTimeout <= 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, r.operationTimeout)
}

func consumeTokenBucketDirect(
	ctx context.Context,
	cancel context.CancelFunc,
	client *redis.Client,
	redisKey string,
	limit float64,
	burst int,
	n int,
	stateTTL time.Duration,
) (bool, error) {
	defer cancel()

	if n > burst {
		return false, nil
	}

	const maxRetry = 8
	for i := 0; i < maxRetry; i++ {
		allowed := false
		err := client.Watch(ctx, func(tx *redis.Tx) error {
			now := time.Now().UnixNano()

			vals, err := tx.HMGet(ctx, redisKey, redisTokenBucketFieldTokens, redisTokenBucketFieldLastNS).Result()
			if err != nil {
				return err
			}

			tokens := int64(burst)
			lastNS := now
			initialized := false
			if len(vals) >= 2 {
				if vals[0] != nil {
					parsed, parseErr := toInt64(vals[0])
					if parseErr != nil {
						return parseErr
					}
					tokens = parsed
					initialized = true
				}
				if vals[1] != nil {
					parsed, parseErr := toInt64(vals[1])
					if parseErr != nil {
						return parseErr
					}
					lastNS = parsed
				}
			}

			if lastNS > now {
				lastNS = now
			}

			refill := calculateRefillTokens(now, lastNS, limit)
			if refill > 0 {
				tokens += refill
				if tokens > int64(burst) {
					tokens = int64(burst)
				}
				lastNS = now
			}

			if !initialized {
				lastNS = now
			}

			if tokens < int64(n) {
				_, pipeErr := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
					if !initialized {
						pipe.HSet(ctx, redisKey,
							redisTokenBucketFieldTokens, tokens,
							redisTokenBucketFieldLastNS, lastNS,
						)
					} else if refill > 0 {
						pipe.HSet(ctx, redisKey,
							redisTokenBucketFieldTokens, tokens,
							redisTokenBucketFieldLastNS, lastNS,
						)
					}
					if stateTTL > 0 {
						pipe.PExpire(ctx, redisKey, stateTTL)
					}
					return nil
				})
				if pipeErr != nil {
					return pipeErr
				}
				allowed = false
				return nil
			}

			_, pipeErr := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				if refill > 0 {
					pipe.HIncrBy(ctx, redisKey, redisTokenBucketFieldTokens, refill)
				}
				// Consume token units using HINCR as requested.
				pipe.HIncrBy(ctx, redisKey, redisTokenBucketFieldTokens, -int64(n))
				pipe.HSet(ctx, redisKey, redisTokenBucketFieldLastNS, lastNS)
				if stateTTL > 0 {
					pipe.PExpire(ctx, redisKey, stateTTL)
				}
				return nil
			})
			if pipeErr != nil {
				return pipeErr
			}

			allowed = true
			return nil
		}, redisKey)
		if err == nil {
			return allowed, nil
		}
		if errors.Is(err, redis.TxFailedErr) {
			continue
		}
		return false, err
	}

	return false, errors.New("token bucket transaction retries exceeded")
}

func calculateRefillTokens(nowNS int64, lastNS int64, limit float64) int64 {
	if limit <= 0 {
		return 0
	}
	if nowNS <= lastNS {
		return 0
	}
	elapsed := nowNS - lastNS
	refillFloat := (float64(elapsed) * limit) / float64(time.Second)
	if refillFloat <= 0 {
		return 0
	}
	refill := int64(math.Floor(refillFloat))
	if refill < 0 {
		return 0
	}
	return refill
}
