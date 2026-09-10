package stability

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nvnamsss/unlimit/efflux"
	"github.com/redis/go-redis/v9"
)

const (
	defaultRedisRateLimiterV2PipelineShards   = 64
	defaultRedisRateLimiterV2QueueFactor      = 2
	defaultRedisRateLimiterV2PipelineMaxWait  = 100 * time.Millisecond
	defaultRedisRateLimiterV2PipelineMaxBatch = 100
)

// RedisRateLimiterV2Config configures a Redis-backed keyed rate limiter with sharded pipelining.
type RedisRateLimiterV2Config struct {
	Limit             float64
	Burst             int
	KeyPrefix         string
	Name              string
	FailureMode       RedisRateLimiterFailureMode
	OperationTimeout  time.Duration
	StateTTL          time.Duration
	PipelineEnabled   bool
	PipelineMaxWait   time.Duration
	PipelineMaxBatch  int
	PipelineShards    int
	PipelineQueueSize int
}

// RedisRateLimiterV2Stats exposes basic runtime counters.
type RedisRateLimiterV2Stats struct {
	PipelineBatches  int64
	PipelineRequests int64
	QueueFullDenies  int64
	AllowedRequests  int64
	TotalRequests    int64
}

// RedisRateLimiterV2 is a Redis-backed leaky bucket with optional sharded batched pipelines.
type RedisRateLimiterV2 struct {
	client  *redis.Client
	baseKey string

	limit            float64
	burst            int
	failureMode      RedisRateLimiterFailureMode
	operationTimeout time.Duration
	stateTTL         time.Duration

	pipelineEnabled   bool
	pipelineMaxWait   time.Duration
	pipelineMaxBatch  int
	pipelineShards    int
	pipelineQueueSize int

	workerPool *efflux.ShardingWorkerPoolWR[*redisPipelineV2Request, redisPipelineV2Result]
	pipelines  []*redisPipelineV2

	pipelineWg sync.WaitGroup
	closeOnce  sync.Once
	closed     atomic.Bool

	pipelineBatches  atomic.Int64
	pipelineRequests atomic.Int64
	queueFullDenies  atomic.Int64
	allowedRequests  atomic.Int64
	totalRequests    atomic.Int64
}

type redisPipelineV2Request struct {
	ctx      context.Context
	shardKey string
	redisKey string
	n        int
	response chan redisPipelineV2Result
}

func (r *redisPipelineV2Request) GetID() string {
	if r.shardKey != "" {
		return r.shardKey
	}
	if r.redisKey != "" {
		return r.redisKey
	}
	return "default"
}

type redisPipelineV2Result struct {
	allowed bool
	delay   time.Duration
	err     error
}

var (
	errRedisPipelineV2Closed    = errors.New("redis pipeline v2 is closed")
	errRedisPipelineV2QueueFull = errors.New("redis pipeline v2 queue is full")
)

// NewRedisRateLimiterV2 creates a new sharded Redis-backed leaky bucket limiter.
func NewRedisRateLimiterV2(client *redis.Client, config RedisRateLimiterV2Config) (*RedisRateLimiterV2, error) {
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

	pipelineMaxWait := config.PipelineMaxWait
	if pipelineMaxWait <= 0 {
		pipelineMaxWait = defaultRedisRateLimiterV2PipelineMaxWait
	}

	pipelineMaxBatch := config.PipelineMaxBatch
	if pipelineMaxBatch <= 0 {
		pipelineMaxBatch = defaultRedisRateLimiterV2PipelineMaxBatch
	}

	pipelineShards := config.PipelineShards
	if pipelineShards <= 0 {
		pipelineShards = defaultRedisRateLimiterV2PipelineShards
	}

	pipelineQueueSize := config.PipelineQueueSize
	if pipelineQueueSize <= 0 {
		pipelineQueueSize = pipelineMaxBatch * defaultRedisRateLimiterV2QueueFactor
	}
	if pipelineQueueSize <= 0 {
		pipelineQueueSize = 1
	}

	r := &RedisRateLimiterV2{
		client:            client,
		baseKey:           prefix + ":" + name,
		limit:             config.Limit,
		burst:             config.Burst,
		failureMode:       failureMode,
		operationTimeout:  opTimeout,
		stateTTL:          stateTTL,
		pipelineEnabled:   config.PipelineEnabled,
		pipelineMaxWait:   pipelineMaxWait,
		pipelineMaxBatch:  pipelineMaxBatch,
		pipelineShards:    pipelineShards,
		pipelineQueueSize: pipelineQueueSize,
	}

	if !r.pipelineEnabled {
		return r, nil
	}

	workers := make([]efflux.WorkerWR[*redisPipelineV2Request, redisPipelineV2Result], 0, r.pipelineShards)
	r.pipelines = make([]*redisPipelineV2, 0, r.pipelineShards)

	for i := 0; i < r.pipelineShards; i++ {
		pipeline := &redisPipelineV2{
			rrl:              r,
			pipelineMaxWait:  r.pipelineMaxWait,
			pipelineMaxBatch: r.pipelineMaxBatch,
			pipelineQueue:    make(chan *redisPipelineV2Request, r.pipelineQueueSize),
			pipelineStopCh:   make(chan struct{}),
		}

		r.pipelines = append(r.pipelines, pipeline)
		workers = append(workers, efflux.NewBasicWorkerWR[*redisPipelineV2Request, redisPipelineV2Result](pipeline.handle))

		r.pipelineWg.Add(1)
		go func(p *redisPipelineV2) {
			defer r.pipelineWg.Done()
			p.Start()
		}(pipeline)
	}

	r.workerPool = efflux.NewShardingWorkerPoolWR[*redisPipelineV2Request, redisPipelineV2Result](workers)
	return r, nil
}

// Close stops all pipeline shards and drains pending requests with close errors.
func (r *RedisRateLimiterV2) Close() {
	r.closeOnce.Do(func() {
		r.closed.Store(true)
		if !r.pipelineEnabled {
			return
		}

		for _, pipeline := range r.pipelines {
			close(pipeline.pipelineStopCh)
		}

		r.pipelineWg.Wait()
	})
}

// Allow checks if one request can be admitted for the given key.
func (r *RedisRateLimiterV2) Allow(ctx context.Context, key string) bool {
	return r.AllowN(ctx, key, 1)
}

// AllowN checks if n requests can be admitted for the given key.
func (r *RedisRateLimiterV2) AllowN(ctx context.Context, key string, n int) bool {
	if n <= 0 {
		return true
	}

	r.totalRequests.Add(int64(n))

	allowed, _, err := r.AllowWithDelayN(ctx, key, n)
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

// AllowWithDelayN returns whether request is allowed and the computed delay from Redis.
func (r *RedisRateLimiterV2) AllowWithDelayN(ctx context.Context, key string, n int) (bool, time.Duration, error) {
	redisKey := r.redisKey(key)
	if r.pipelineEnabled {
		return r.allowWithDelayPipeline(ctx, key, redisKey, n)
	}

	return r.allowWithDelayDirect(ctx, redisKey, n)
}

// Limit returns the configured requests per second.
func (r *RedisRateLimiterV2) Limit() float64 {
	return r.limit
}

// Burst returns the configured bucket capacity.
func (r *RedisRateLimiterV2) Burst() int {
	return r.burst
}

// Stats returns a snapshot of internal counters.
func (r *RedisRateLimiterV2) Stats() RedisRateLimiterV2Stats {
	return RedisRateLimiterV2Stats{
		PipelineBatches:  r.pipelineBatches.Load(),
		PipelineRequests: r.pipelineRequests.Load(),
		QueueFullDenies:  r.queueFullDenies.Load(),
		AllowedRequests:  r.allowedRequests.Load(),
		TotalRequests:    r.totalRequests.Load(),
	}
}

func (r *RedisRateLimiterV2) allowWithDelayDirect(ctx context.Context, redisKey string, n int) (bool, time.Duration, error) {
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

func (r *RedisRateLimiterV2) allowWithDelayPipeline(ctx context.Context, key string, redisKey string, n int) (bool, time.Duration, error) {
	if n > r.Burst() {
		return false, 0, nil
	}
	if r.closed.Load() {
		return false, 0, errRedisPipelineV2Closed
	}

	opCtx, cancel := r.operationContext(ctx)
	defer cancel()

	req := &redisPipelineV2Request{
		ctx:      opCtx,
		shardKey: key,
		redisKey: redisKey,
		n:        n,
		response: make(chan redisPipelineV2Result, 1),
	}

	_, err := r.workerPool.Submit(opCtx, req)
	if err != nil {
		if errors.Is(err, errRedisPipelineV2QueueFull) {
			r.queueFullDenies.Add(1)
			return false, 0, nil
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

func (r *RedisRateLimiterV2) redisKey(key string) string {
	if key == "" {
		return r.baseKey
	}

	return r.baseKey + ":" + key
}

func (r *RedisRateLimiterV2) leakDuration() time.Duration {
	if r.limit <= 0 {
		return time.Second
	}

	leak := time.Duration(float64(time.Second) / r.limit)
	if leak <= 0 {
		return time.Nanosecond
	}
	return leak
}

func (r *RedisRateLimiterV2) operationContext(parent context.Context) (context.Context, context.CancelFunc) {
	if r.operationTimeout <= 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, r.operationTimeout)
}

func (r *RedisRateLimiterV2) failureModeAllowsOnError() bool {
	return r.failureMode == RedisRateLimiterFailOpen
}

type redisPipelineV2 struct {
	rrl *RedisRateLimiterV2

	pipelineMaxWait  time.Duration
	pipelineMaxBatch int
	pipelineQueue    chan *redisPipelineV2Request
	pipelineStopCh   chan struct{}
}

func (r *redisPipelineV2) handle(ctx context.Context, req *redisPipelineV2Request) (redisPipelineV2Result, error) {
	select {
	case r.pipelineQueue <- req:
	case <-ctx.Done():
		return redisPipelineV2Result{}, ctx.Err()
	case <-r.pipelineStopCh:
		return redisPipelineV2Result{}, errRedisPipelineV2Closed
	default:
		return redisPipelineV2Result{}, errRedisPipelineV2QueueFull
	}

	return redisPipelineV2Result{}, nil
}

func (r *redisPipelineV2) Start() {
	ticker := time.NewTicker(r.pipelineMaxWait)
	defer ticker.Stop()

	for {
		select {
		case <-r.pipelineStopCh:
			r.failPending(errRedisPipelineV2Closed)
			return
		case first := <-r.pipelineQueue:
			batch := []*redisPipelineV2Request{first}
			r.collectBatch(&batch)
			r.executePipeline(batch)
		case <-ticker.C:
			batch := make([]*redisPipelineV2Request, 0, r.pipelineMaxBatch)
			r.collectBatch(&batch)
			if len(batch) > 0 {
				r.executePipeline(batch)
			}
		}
	}
}

func (r *redisPipelineV2) collectBatch(batch *[]*redisPipelineV2Request) {
	for len(*batch) < r.pipelineMaxBatch {
		select {
		case req := <-r.pipelineQueue:
			*batch = append(*batch, req)
		default:
			return
		}
	}
}

func (r *redisPipelineV2) executePipeline(batch []*redisPipelineV2Request) {
	if len(batch) == 0 {
		return
	}

	r.rrl.pipelineBatches.Add(1)
	r.rrl.pipelineRequests.Add(int64(len(batch)))

	ctx, cancel := r.rrl.operationContext(context.Background())
	defer cancel()

	pipe := r.rrl.client.Pipeline()
	cmds := make([]*redis.Cmd, len(batch))

	for i, req := range batch {
		if req.ctx != nil && req.ctx.Err() != nil {
			r.sendPipelineResult(req, redisPipelineV2Result{err: req.ctx.Err()})
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
				r.sendPipelineResult(req, redisPipelineV2Result{err: req.ctx.Err()})
				continue
			}
			r.sendPipelineResult(req, redisPipelineV2Result{err: err})
		}
		return
	}

	for i, req := range batch {
		if cmds[i] == nil {
			continue
		}

		res, err := cmds[i].Result()
		if err != nil {
			r.sendPipelineResult(req, redisPipelineV2Result{err: err})
			continue
		}

		allowed, delay, parseErr := parseAllowScriptResult(res)
		r.sendPipelineResult(req, redisPipelineV2Result{
			allowed: allowed,
			delay:   delay,
			err:     parseErr,
		})
	}
}

func (r *redisPipelineV2) failPending(err error) {
	for {
		select {
		case req := <-r.pipelineQueue:
			if req == nil {
				continue
			}
			r.sendPipelineResult(req, redisPipelineV2Result{err: err})
		default:
			return
		}
	}
}

func (r *redisPipelineV2) sendPipelineResult(req *redisPipelineV2Request, result redisPipelineV2Result) {
	select {
	case req.response <- result:
	default:
	}
}
