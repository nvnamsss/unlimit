package stability

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/voidforge-studios/unlimit/efflux"
)

const (
	defaultRedisRateLimiterBatchV3PipelineShards   = 64
	defaultRedisRateLimiterBatchV3QueueFactor      = 2
	defaultRedisRateLimiterBatchV3PipelineMaxWait  = 100 * time.Microsecond
	defaultRedisRateLimiterBatchV3PipelineMaxBatch = 1000
)

// RedisRateLimiterBatchV3Config configures a Redis token-bucket limiter with sharded pipeline batching.
type RedisRateLimiterBatchV3Config struct {
	Limit             float64
	Burst             int
	KeyPrefix         string
	Name              string
	FailureMode       RedisRateLimiterFailureMode
	OperationTimeout  time.Duration
	StateTTL          time.Duration
	PipelineMaxWait   time.Duration
	PipelineMaxBatch  int
	PipelineShards    int
	PipelineQueueSize int
}

// RedisRateLimiterBatchV3Stats exposes runtime counters.
type RedisRateLimiterBatchV3Stats struct {
	TotalRequests    int64
	AllowedRequests  int64
	PipelineBatches  int64
	PipelineRequests int64
	QueueFullDenies  int64
}

// RedisRateLimiterBatchV3 is a sharded batched token-bucket limiter.
type RedisRateLimiterBatchV3 struct {
	client  *redis.Client
	baseKey string

	limit            float64
	burst            int
	failureMode      RedisRateLimiterFailureMode
	operationTimeout time.Duration
	stateTTL         time.Duration

	pipelineMaxWait   time.Duration
	pipelineMaxBatch  int
	pipelineShards    int
	pipelineQueueSize int

	workerPool *efflux.ShardingWorkerPoolWR[*redisBatchV3PipelineRequest, redisBatchV3PipelineResult]
	pipelines  []*redisBatchV3Pipeline

	pipelineWg sync.WaitGroup
	closeOnce  sync.Once
	closed     atomic.Bool

	totalRequests    atomic.Int64
	allowedRequests  atomic.Int64
	pipelineBatches  atomic.Int64
	pipelineRequests atomic.Int64
	queueFullDenies  atomic.Int64
	requestSeq       atomic.Uint64
}

type redisBatchV3PipelineRequest struct {
	ctx       context.Context
	requestID uint64
	shardKey  string
	redisKey  string
	n         int
	release   chan struct{}
	owner     *redisBatchV3Pipeline
}

func (r *redisBatchV3PipelineRequest) GetID() string {
	if r.shardKey != "" {
		return r.shardKey
	}
	if r.redisKey != "" {
		return r.redisKey
	}
	return "default"
}

type redisBatchV3PipelineResult struct {
	allowed bool
	err     error
}

var (
	errRedisBatchV3PipelineClosed    = errors.New("redis batch v3 pipeline is closed")
	errRedisBatchV3PipelineQueueFull = errors.New("redis batch v3 pipeline queue is full")
	errRedisBatchV3ResultMissing     = errors.New("redis batch v3 result missing")
)

// NewRedisRateLimiterBatchV3 creates a new sharded batch token-bucket limiter.
func NewRedisRateLimiterBatchV3(client *redis.Client, config RedisRateLimiterBatchV3Config) (*RedisRateLimiterBatchV3, error) {
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

	pipelineMaxWait := config.PipelineMaxWait
	if pipelineMaxWait <= 0 {
		pipelineMaxWait = defaultRedisRateLimiterBatchV3PipelineMaxWait
	}
	pipelineMaxBatch := config.PipelineMaxBatch
	if pipelineMaxBatch <= 0 {
		pipelineMaxBatch = defaultRedisRateLimiterBatchV3PipelineMaxBatch
	}
	pipelineShards := config.PipelineShards
	if pipelineShards <= 0 {
		pipelineShards = defaultRedisRateLimiterBatchV3PipelineShards
	}
	pipelineQueueSize := config.PipelineQueueSize
	if pipelineQueueSize <= 0 {
		pipelineQueueSize = pipelineMaxBatch * defaultRedisRateLimiterBatchV3QueueFactor
	}
	if pipelineQueueSize <= 0 {
		pipelineQueueSize = 1
	}

	r := &RedisRateLimiterBatchV3{
		client:            client,
		baseKey:           prefix + ":" + name,
		limit:             config.Limit,
		burst:             config.Burst,
		failureMode:       failureMode,
		operationTimeout:  opTimeout,
		stateTTL:          stateTTL,
		pipelineMaxWait:   pipelineMaxWait,
		pipelineMaxBatch:  pipelineMaxBatch,
		pipelineShards:    pipelineShards,
		pipelineQueueSize: pipelineQueueSize,
	}

	workers := make([]efflux.WorkerWR[*redisBatchV3PipelineRequest, redisBatchV3PipelineResult], 0, r.pipelineShards)
	r.pipelines = make([]*redisBatchV3Pipeline, 0, r.pipelineShards)

	for i := 0; i < r.pipelineShards; i++ {
		pipeline := &redisBatchV3Pipeline{
			rrl:              r,
			pipelineMaxWait:  r.pipelineMaxWait,
			pipelineMaxBatch: r.pipelineMaxBatch,
			pipelineQueue:    make(chan *redisBatchV3PipelineRequest, r.pipelineQueueSize),
			pipelineStopCh:   make(chan struct{}),
			results:          make(map[uint64]redisBatchV3PipelineResult),
		}
		r.pipelines = append(r.pipelines, pipeline)
		workers = append(workers, efflux.NewBasicWorkerWR[*redisBatchV3PipelineRequest, redisBatchV3PipelineResult](pipeline.handle))

		r.pipelineWg.Add(1)
		go func(p *redisBatchV3Pipeline) {
			defer r.pipelineWg.Done()
			p.Start()
		}(pipeline)
	}

	r.workerPool = efflux.NewShardingWorkerPoolWR[*redisBatchV3PipelineRequest, redisBatchV3PipelineResult](workers)
	return r, nil
}

// Close stops batch pipeline shards.
func (r *RedisRateLimiterBatchV3) Close() {
	r.closeOnce.Do(func() {
		r.closed.Store(true)
		for _, p := range r.pipelines {
			close(p.pipelineStopCh)
		}
		r.pipelineWg.Wait()
	})
}

// Allow checks one request.
func (r *RedisRateLimiterBatchV3) Allow(ctx context.Context, key string) bool {
	return r.AllowN(ctx, key, 1)
}

// AllowN checks n requests.
func (r *RedisRateLimiterBatchV3) AllowN(ctx context.Context, key string, n int) bool {
	if n <= 0 {
		return true
	}

	r.totalRequests.Add(int64(n))

	allowed, err := r.allowN(ctx, key, n)
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

func (r *RedisRateLimiterBatchV3) allowN(ctx context.Context, key string, n int) (bool, error) {
	if n > r.burst {
		return false, nil
	}
	if r.closed.Load() {
		return false, errRedisBatchV3PipelineClosed
	}

	opCtx, cancel := r.operationContext(ctx)
	defer cancel()

	req := &redisBatchV3PipelineRequest{
		ctx:       opCtx,
		requestID: r.requestSeq.Add(1),
		shardKey:  key,
		redisKey:  r.redisKey(key),
		n:         n,
		release:   make(chan struct{}, 1),
	}

	_, err := r.workerPool.Submit(opCtx, req)
	if err != nil {
		if errors.Is(err, errRedisBatchV3PipelineQueueFull) {
			r.queueFullDenies.Add(1)
			return false, nil
		}
		return false, err
	}

	select {
	case <-req.release:
		if req.owner == nil {
			return false, errRedisBatchV3ResultMissing
		}
		result, ok := req.owner.takeResult(req.requestID)
		if !ok {
			return false, errRedisBatchV3ResultMissing
		}
		return result.allowed, result.err
	case <-opCtx.Done():
		if req.owner != nil {
			req.owner.deleteResult(req.requestID)
		}
		return false, opCtx.Err()
	}
}

// Stats returns a snapshot of internal counters.
func (r *RedisRateLimiterBatchV3) Stats() RedisRateLimiterBatchV3Stats {
	return RedisRateLimiterBatchV3Stats{
		TotalRequests:    r.totalRequests.Load(),
		AllowedRequests:  r.allowedRequests.Load(),
		PipelineBatches:  r.pipelineBatches.Load(),
		PipelineRequests: r.pipelineRequests.Load(),
		QueueFullDenies:  r.queueFullDenies.Load(),
	}
}

func (r *RedisRateLimiterBatchV3) redisKey(key string) string {
	if key == "" {
		return r.baseKey
	}
	return r.baseKey + ":" + key
}

func (r *RedisRateLimiterBatchV3) operationContext(parent context.Context) (context.Context, context.CancelFunc) {
	if r.operationTimeout <= 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, r.operationTimeout)
}

type redisBatchV3Pipeline struct {
	rrl *RedisRateLimiterBatchV3

	pipelineMaxWait  time.Duration
	pipelineMaxBatch int
	pipelineQueue    chan *redisBatchV3PipelineRequest
	pipelineStopCh   chan struct{}

	resultsMu sync.Mutex
	results   map[uint64]redisBatchV3PipelineResult
}

func (r *redisBatchV3Pipeline) handle(ctx context.Context, req *redisBatchV3PipelineRequest) (redisBatchV3PipelineResult, error) {
	req.owner = r

	select {
	case r.pipelineQueue <- req:
		return redisBatchV3PipelineResult{}, nil
	case <-ctx.Done():
		return redisBatchV3PipelineResult{}, ctx.Err()
	case <-r.pipelineStopCh:
		return redisBatchV3PipelineResult{}, errRedisBatchV3PipelineClosed
	default:
		return redisBatchV3PipelineResult{}, errRedisBatchV3PipelineQueueFull
	}
}

func (r *redisBatchV3Pipeline) Start() {
	ticker := time.NewTicker(r.pipelineMaxWait)
	defer ticker.Stop()

	for {
		select {
		case <-r.pipelineStopCh:
			r.failPending(errRedisBatchV3PipelineClosed)
			return
		case first := <-r.pipelineQueue:
			batch := []*redisBatchV3PipelineRequest{first}
			r.collectBatch(&batch)
			r.executePipeline(batch)
		case <-ticker.C:
			batch := make([]*redisBatchV3PipelineRequest, 0, r.pipelineMaxBatch)
			r.collectBatch(&batch)
			if len(batch) > 0 {
				r.executePipeline(batch)
			}
		}
	}
}

func (r *redisBatchV3Pipeline) collectBatch(batch *[]*redisBatchV3PipelineRequest) {
	for len(*batch) < r.pipelineMaxBatch {
		select {
		case req := <-r.pipelineQueue:
			*batch = append(*batch, req)
		default:
			return
		}
	}
}

func (r *redisBatchV3Pipeline) executePipeline(batch []*redisBatchV3PipelineRequest) {
	if len(batch) == 0 {
		return
	}

	r.rrl.pipelineBatches.Add(1)
	r.rrl.pipelineRequests.Add(int64(len(batch)))

	ctx, cancel := r.rrl.operationContext(context.Background())
	defer cancel()

	group := make(map[string][]int)
	for i, req := range batch {
		group[req.redisKey] = append(group[req.redisKey], i)
	}

	allowByIndex := make([]bool, len(batch))
	consumeCmdByIndex := make([]*redis.IntCmd, len(batch))
	groupErr := make(map[string]error)
	refillByKey := make(map[string]int64, len(group))

	for redisKey, indexes := range group {
		now := time.Now().UnixNano()
		vals, err := r.rrl.client.HMGet(ctx, redisKey, redisTokenBucketFieldTokens, redisTokenBucketFieldLastNS).Result()
		if err != nil {
			groupErr[redisKey] = err
			continue
		}

		tokens := int64(0)
		lastNS := now
		initialRefill := int64(0)
		if len(vals) >= 2 {
			if vals[0] != nil {
				parsed, parseErr := toInt64(vals[0])
				if parseErr != nil {
					groupErr[redisKey] = parseErr
					continue
				}
				tokens = parsed
			} else {
				// Missing token field starts at full bucket capacity.
				initialRefill = int64(r.rrl.burst)
			}
			if vals[1] != nil {
				parsed, parseErr := toInt64(vals[1])
				if parseErr != nil {
					groupErr[redisKey] = parseErr
					continue
				}
				lastNS = parsed
			}
		}

		if lastNS > now {
			lastNS = now
		}

		refill := calculateRefillTokens(now, lastNS, r.rrl.limit)
		if refill > 0 {
			maxRefill := int64(r.rrl.burst) - tokens
			if maxRefill < 0 {
				maxRefill = 0
			}
			if refill > maxRefill {
				refill = maxRefill
			}
		}
		refillByKey[redisKey] = initialRefill + refill

		remaining := tokens + initialRefill + refill
		if remaining > int64(r.rrl.burst) {
			remaining = int64(r.rrl.burst)
		}

		for _, idx := range indexes {
			reqN := int64(batch[idx].n)
			if reqN <= remaining {
				allowByIndex[idx] = true
				remaining -= reqN
			}
		}
	}

	pipe := r.rrl.client.Pipeline()
	for redisKey, indexes := range group {
		if err := groupErr[redisKey]; err != nil {
			continue
		}

		// One refill command for this key group.
		pipe.HIncrBy(ctx, redisKey, redisTokenBucketFieldTokens, refillByKey[redisKey])

		// k consume commands where k is number of requests for this key group.
		for _, idx := range indexes {
			delta := int64(0)
			if allowByIndex[idx] {
				delta = -int64(batch[idx].n)
			}
			consumeCmdByIndex[idx] = pipe.HIncrBy(ctx, redisKey, redisTokenBucketFieldTokens, delta)
		}

		pipe.HSet(ctx, redisKey, redisTokenBucketFieldLastNS, time.Now().UnixNano())
		if r.rrl.stateTTL > 0 {
			pipe.PExpire(ctx, redisKey, r.rrl.stateTTL)
		}
	}

	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		for _, req := range batch {
			r.storeResultAndRelease(req, redisBatchV3PipelineResult{err: err})
		}
		return
	}

	for i, req := range batch {
		if req.ctx != nil && req.ctx.Err() != nil {
			continue
		}

		if err := groupErr[req.redisKey]; err != nil {
			r.storeResultAndRelease(req, redisBatchV3PipelineResult{err: err})
			continue
		}

		if consumeCmdByIndex[i] != nil {
			if _, err := consumeCmdByIndex[i].Result(); err != nil {
				r.storeResultAndRelease(req, redisBatchV3PipelineResult{err: err})
				continue
			}
		}

		r.storeResultAndRelease(req, redisBatchV3PipelineResult{allowed: allowByIndex[i]})
	}
}

func (r *redisBatchV3Pipeline) failPending(err error) {
	for {
		select {
		case req := <-r.pipelineQueue:
			if req == nil {
				continue
			}
			r.storeResultAndRelease(req, redisBatchV3PipelineResult{err: err})
		default:
			return
		}
	}
}

func (r *redisBatchV3Pipeline) storeResultAndRelease(req *redisBatchV3PipelineRequest, result redisBatchV3PipelineResult) {
	if req == nil {
		return
	}

	r.resultsMu.Lock()
	r.results[req.requestID] = result
	r.resultsMu.Unlock()

	r.releaseRequest(req)
}

func (r *redisBatchV3Pipeline) releaseRequest(req *redisBatchV3PipelineRequest) {
	if req == nil {
		return
	}

	select {
	case req.release <- struct{}{}:
	default:
	}
}

func (r *redisBatchV3Pipeline) takeResult(requestID uint64) (redisBatchV3PipelineResult, bool) {
	r.resultsMu.Lock()
	defer r.resultsMu.Unlock()

	result, ok := r.results[requestID]
	if !ok {
		return redisBatchV3PipelineResult{}, false
	}

	delete(r.results, requestID)
	return result, true
}

func (r *redisBatchV3Pipeline) deleteResult(requestID uint64) {
	r.resultsMu.Lock()
	defer r.resultsMu.Unlock()

	delete(r.results, requestID)
}
