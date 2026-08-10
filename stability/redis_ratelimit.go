package stability

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/voidforge-studios/unlimit/efflux"
)

const (
	defaultRedisRateLimiterPrefix   = "stability:ratelimit"
	defaultRedisRateLimiterName     = "global"
	defaultRedisOperationTimeout    = 200 * time.Millisecond
	defaultRedisRateLimiterStateTTL = 5 * time.Minute
)

const redisAllowScript = `
local capacity = tonumber(ARGV[1])
local leak_ns = tonumber(ARGV[2])
local n = tonumber(ARGV[3])
local ttl_ms = tonumber(ARGV[4])

local t = redis.call('TIME')
local now_ns = t[1] * 1000000000 + t[2] * 1000

local current = tonumber(redis.call('HGET', KEYS[1], 'current') or '0')
local last_ns = tonumber(redis.call('HGET', KEYS[1], 'last_ns') or tostring(now_ns))

if last_ns > now_ns then
	last_ns = now_ns
end

if leak_ns > 0 then
	local elapsed = now_ns - last_ns
	if elapsed >= leak_ns then
		local leaked = math.floor(elapsed / leak_ns)
		current = math.max(0, current - leaked)
		last_ns = last_ns + leaked * leak_ns
	end
end

local allowed = 0
local delay_ns = 0
if n <= (capacity - current) then
	current = current + n
	allowed = 1
else
	local needed = n - (capacity - current)
	delay_ns = needed * leak_ns
end

redis.call('HSET', KEYS[1], 'current', current, 'last_ns', last_ns)
if ttl_ms > 0 then
	redis.call('PEXPIRE', KEYS[1], ttl_ms)
end

return {allowed, delay_ns, current}
`

const redisReserveScript = `
local capacity = tonumber(ARGV[1])
local leak_ns = tonumber(ARGV[2])
local n = tonumber(ARGV[3])
local ttl_ms = tonumber(ARGV[4])

local t = redis.call('TIME')
local now_ns = t[1] * 1000000000 + t[2] * 1000

local current = tonumber(redis.call('HGET', KEYS[1], 'current') or '0')
local last_ns = tonumber(redis.call('HGET', KEYS[1], 'last_ns') or tostring(now_ns))

if last_ns > now_ns then
	last_ns = now_ns
end

if leak_ns > 0 then
	local elapsed = now_ns - last_ns
	if elapsed >= leak_ns then
		local leaked = math.floor(elapsed / leak_ns)
		current = math.max(0, current - leaked)
		last_ns = last_ns + leaked * leak_ns
	end
end

local delay_ns = 0
if n > (capacity - current) then
	local needed = n - (capacity - current)
	delay_ns = needed * leak_ns
end

redis.call('HSET', KEYS[1], 'current', current, 'last_ns', last_ns)
if ttl_ms > 0 then
	redis.call('PEXPIRE', KEYS[1], ttl_ms)
end

return {1, delay_ns}
`

const redisTokensScript = `
local capacity = tonumber(ARGV[1])
local leak_ns = tonumber(ARGV[2])
local ttl_ms = tonumber(ARGV[3])

local t = redis.call('TIME')
local now_ns = t[1] * 1000000000 + t[2] * 1000

local current = tonumber(redis.call('HGET', KEYS[1], 'current') or '0')
local last_ns = tonumber(redis.call('HGET', KEYS[1], 'last_ns') or tostring(now_ns))

if last_ns > now_ns then
	last_ns = now_ns
end

if leak_ns > 0 then
	local elapsed = now_ns - last_ns
	if elapsed >= leak_ns then
		local leaked = math.floor(elapsed / leak_ns)
		current = math.max(0, current - leaked)
		last_ns = last_ns + leaked * leak_ns
	end
end

local available = capacity - current
if available < 0 then
	available = 0
end

redis.call('HSET', KEYS[1], 'current', current, 'last_ns', last_ns)
if ttl_ms > 0 then
	redis.call('PEXPIRE', KEYS[1], ttl_ms)
end

return {available}
`

// RedisRateLimiterFailureMode controls behavior when Redis is unavailable.
type RedisRateLimiterFailureMode string

const (
	// RedisRateLimiterFailClosed denies requests when Redis cannot be reached.
	RedisRateLimiterFailClosed RedisRateLimiterFailureMode = "fail_closed"
	// RedisRateLimiterFailOpen allows requests when Redis cannot be reached.
	RedisRateLimiterFailOpen RedisRateLimiterFailureMode = "fail_open"
)

// RedisRateLimiterConfig configures a Redis-backed global rate limiter.
type RedisRateLimiterConfig struct {
	Limit            float64
	Burst            int
	KeyPrefix        string
	Name             string
	FailureMode      RedisRateLimiterFailureMode
	OperationTimeout time.Duration
	StateTTL         time.Duration
	PipelineEnabled  bool
	PipelineMaxWait  time.Duration
	PipelineMaxBatch int
}

// RedisRateLimiter is a distributed leaky bucket backed by Redis state.
type RedisRateLimiter struct {
	client  *redis.Client
	baseKey string

	limit            float64
	burst            int
	failureMode      RedisRateLimiterFailureMode
	operationTimeout time.Duration
	stateTTL         time.Duration

	pipelineEnabled  bool
	pipelineMaxWait  time.Duration
	pipelineMaxBatch int

	workerPool *efflux.ShardingWorkerPoolWR[*redisPipelineRequest, redisPipelineResult]
	pipelines  []*redisPipeline

	pipelineWg sync.WaitGroup
	closeOnce  sync.Once

	totalRequests   atomic.Int64
	allowedRequests atomic.Int64
}

// RedisRateLimiterStats exposes basic runtime counters.
type RedisRateLimiterStats struct {
	TotalRequests   int64
	AllowedRequests int64
}

type redisPipelineRequest struct {
	ctx      context.Context
	shardKey string
	redisKey string
	n        int
	response chan redisPipelineResult
}

func (r *redisPipelineRequest) GetID() string {
	if r.shardKey != "" {
		return r.shardKey
	}
	if r.redisKey != "" {
		return r.redisKey
	}
	return "default"
}

type redisPipelineResult struct {
	allowed bool
	delay   time.Duration
	err     error
}

var (
	errRedisPipelineClosed    = errors.New("redis pipeline is closed")
	errRedisPipelineQueueFull = errors.New("redis pipeline queue is full")
)

// NewRedisRateLimiter creates a distributed global leaky bucket backed by Redis.
func NewRedisRateLimiter(client *redis.Client, config RedisRateLimiterConfig) (*RedisRateLimiter, error) {
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
		prefix = defaultRedisRateLimiterPrefix
	}

	name := config.Name
	if name == "" {
		name = defaultRedisRateLimiterName
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
		opTimeout = defaultRedisOperationTimeout
	}

	stateTTL := config.StateTTL
	if stateTTL <= 0 {
		stateTTL = defaultRedisRateLimiterStateTTL
	}

	r := &RedisRateLimiter{
		client:           client,
		baseKey:          prefix + ":" + name,
		limit:            config.Limit,
		burst:            config.Burst,
		failureMode:      failureMode,
		operationTimeout: opTimeout,
		stateTTL:         stateTTL,
	}
	return r, nil
}

// Close stops the optional pipeline workers and releases related resources.
func (r *RedisRateLimiter) Close() {
	r.closeOnce.Do(func() {
		// v1 runs in non-pipeline mode.
	})
}

// Allow checks if one request can be admitted for the given key.
func (r *RedisRateLimiter) Allow(ctx context.Context, key string) bool {
	return r.AllowN(ctx, key, 1)
}

// AllowN checks if n requests can be admitted for the given key.
func (r *RedisRateLimiter) AllowN(ctx context.Context, key string, n int) bool {
	if n <= 0 {
		return true
	}

	r.totalRequests.Add(int64(n))

	allowed, _, err := r.allowWithDelay(ctx, key, n)
	if err != nil {
		if r.failureModeAllowsOnError() {
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
func (r *RedisRateLimiter) Limit() float64 {
	return r.limit
}

// Burst returns the configured bucket capacity.
func (r *RedisRateLimiter) Burst() int {
	return r.burst
}

// Stats returns a snapshot of internal counters.
func (r *RedisRateLimiter) Stats() RedisRateLimiterStats {
	return RedisRateLimiterStats{
		TotalRequests:   r.totalRequests.Load(),
		AllowedRequests: r.allowedRequests.Load(),
	}
}

func (r *RedisRateLimiter) allowWithDelay(ctx context.Context, key string, n int) (bool, time.Duration, error) {
	redisKey := r.redisKey(key)
	return r.allowWithDelayDirect(ctx, redisKey, n)
}

func (r *RedisRateLimiter) allowWithDelayDirect(ctx context.Context, redisKey string, n int) (bool, time.Duration, error) {
	if n > r.Burst() {
		return false, 0, nil
	}

	opCtx, cancel := r.operationContext(ctx)
	defer cancel()

	res, err := r.client.Eval(
		opCtx,
		redisAllowScript,
		[]string{redisKey},
		r.Burst(),
		r.leakDuration().Nanoseconds(),
		n,
		r.stateTTL.Milliseconds(),
	).Result()
	if err != nil {
		return false, 0, err
	}

	return parseAllowScriptResult(res)
}

func (r *RedisRateLimiter) allowWithDelayPipeline(ctx context.Context, key string, redisKey string, n int) (bool, time.Duration, error) {
	if n > r.Burst() {
		return false, 0, nil
	}

	opCtx, cancel := r.operationContext(ctx)
	defer cancel()

	req := &redisPipelineRequest{
		ctx:      opCtx,
		shardKey: key,
		redisKey: redisKey,
		n:        n,
		response: make(chan redisPipelineResult, 1),
	}

	_, err := r.workerPool.Submit(opCtx, req)
	if err != nil {
		if errors.Is(err, errRedisPipelineQueueFull) {
			return r.allowWithDelayDirect(opCtx, redisKey, n)
		}
		return false, 0, err
	}

	select {
	case result := <-req.response:
		return result.allowed, result.delay, result.err
	case <-opCtx.Done():
		return false, 0, opCtx.Err()
	}
}

func parseAllowScriptResult(res interface{}) (bool, time.Duration, error) {
	list, ok := res.([]interface{})
	if !ok || len(list) < 2 {
		return false, 0, errors.New("unexpected redis script response")
	}

	allowed, err := toInt64(list[0])
	if err != nil {
		return false, 0, err
	}
	delayNs, err := toInt64(list[1])
	if err != nil {
		return false, 0, err
	}

	if delayNs < 0 {
		delayNs = 0
	}

	return allowed == 1, time.Duration(delayNs), nil
}

func (r *RedisRateLimiter) redisKey(key string) string {
	if key == "" {
		return r.baseKey
	}

	return r.baseKey + ":" + key
}

func (r *RedisRateLimiter) leakDuration() time.Duration {
	return r.leakDurationLocked()
}

func (r *RedisRateLimiter) leakDurationLocked() time.Duration {
	if r.limit <= 0 {
		return time.Second
	}

	leak := time.Duration(float64(time.Second) / r.limit)
	if leak <= 0 {
		return time.Nanosecond
	}
	return leak
}

func (r *RedisRateLimiter) operationContext(parent context.Context) (context.Context, context.CancelFunc) {
	if r.operationTimeout <= 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, r.operationTimeout)
}

func (r *RedisRateLimiter) failureModeAllowsOnError() bool {
	return r.failureMode == RedisRateLimiterFailOpen
}

func toInt64(v interface{}) (int64, error) {
	switch val := v.(type) {
	case int64:
		return val, nil
	case int:
		return int64(val), nil
	case uint64:
		if val > uint64(^uint64(0)>>1) {
			return 0, errors.New("value overflows int64")
		}
		return int64(val), nil
	case float64:
		return int64(val), nil
	case string:
		parsed, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return 0, err
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("unsupported number type: %T", v)
	}
}

type redisPipeline struct {
	rrl *RedisRateLimiter

	pipelineMaxWait  time.Duration
	pipelineMaxBatch int
	pipelineQueue    chan *redisPipelineRequest
	pipelineStopCh   chan struct{}
}

func (r *redisPipeline) handle(ctx context.Context, req *redisPipelineRequest) (redisPipelineResult, error) {
	select {
	case r.pipelineQueue <- req:
	case <-ctx.Done():
		return redisPipelineResult{}, ctx.Err()
	case <-r.pipelineStopCh:
		return redisPipelineResult{}, errRedisPipelineClosed
	default:
		return redisPipelineResult{}, errRedisPipelineQueueFull
	}

	return redisPipelineResult{}, nil
}

func (r *redisPipeline) Start() {
	ticker := time.NewTicker(r.pipelineMaxWait)
	defer ticker.Stop()

	for {
		select {
		case <-r.pipelineStopCh:
			r.failPending(errRedisPipelineClosed)
			return
		case first := <-r.pipelineQueue:
			batch := []*redisPipelineRequest{first}
			r.collectBatch(&batch)
			r.executePipeline(batch)
		case <-ticker.C:
			batch := make([]*redisPipelineRequest, 0, r.pipelineMaxBatch)
			r.collectBatch(&batch)
			if len(batch) > 0 {
				r.executePipeline(batch)
			}
		}
	}
}

func (r *redisPipeline) collectBatch(batch *[]*redisPipelineRequest) {
	for len(*batch) < r.pipelineMaxBatch {
		select {
		case req := <-r.pipelineQueue:
			*batch = append(*batch, req)
		default:
			return
		}
	}
}

func (r *redisPipeline) executePipeline(batch []*redisPipelineRequest) {
	if len(batch) == 0 {
		return
	}

	ctx, cancel := r.rrl.operationContext(context.Background())
	defer cancel()

	pipe := r.rrl.client.Pipeline()
	cmds := make([]*redis.Cmd, len(batch))

	for i, req := range batch {
		if req.ctx != nil && req.ctx.Err() != nil {
			r.sendPipelineResult(req, redisPipelineResult{err: req.ctx.Err()})
			continue
		}

		cmds[i] = pipe.Eval(
			ctx,
			redisAllowScript,
			[]string{req.redisKey},
			r.rrl.Burst(),
			r.rrl.leakDuration().Nanoseconds(),
			req.n,
			r.rrl.stateTTL.Milliseconds(),
		)
	}

	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		for _, req := range batch {
			if req.ctx != nil && req.ctx.Err() != nil {
				r.sendPipelineResult(req, redisPipelineResult{err: req.ctx.Err()})
				continue
			}
			r.sendPipelineResult(req, redisPipelineResult{err: err})
		}
		return
	}

	for i, req := range batch {
		if cmds[i] == nil {
			continue
		}

		res, err := cmds[i].Result()
		if err != nil {
			r.sendPipelineResult(req, redisPipelineResult{err: err})
			continue
		}

		allowed, delay, parseErr := parseAllowScriptResult(res)
		r.sendPipelineResult(req, redisPipelineResult{
			allowed: allowed,
			delay:   delay,
			err:     parseErr,
		})
	}
}

func (r *redisPipeline) failPending(err error) {
	for {
		select {
		case req := <-r.pipelineQueue:
			if req == nil {
				continue
			}
			r.sendPipelineResult(req, redisPipelineResult{err: err})
		default:
			return
		}
	}
}

func (r *redisPipeline) sendPipelineResult(req *redisPipelineRequest, result redisPipelineResult) {
	select {
	case req.response <- result:
	default:
	}
}
