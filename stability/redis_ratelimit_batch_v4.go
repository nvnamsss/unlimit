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
)

const (
	defaultRedisRateLimiterBatchV4PipelineShards   = 64
	defaultRedisRateLimiterBatchV4QueueFactor      = 2
	defaultRedisRateLimiterBatchV4PipelineMaxWait  = 100 * time.Microsecond
	defaultRedisRateLimiterBatchV4PipelineMaxBatch = 1000
)

// RedisRateLimiterBatchV4Config configures a Redis token-bucket limiter with requestID-correlated load-balanced batching.
type RedisRateLimiterBatchV4Config struct {
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

// RedisRateLimiterBatchV4Stats exposes runtime counters.
type RedisRateLimiterBatchV4Stats struct {
	TotalRequests         int64
	AllowedRequests       int64
	PipelineBatches       int64
	PipelineRequests      int64
	QueueFullDenies       int64
	CorrelationMismatches int64
}

// RedisRateLimiterBatchV4 is a requestID-correlated sharded batched token-bucket limiter.
type RedisRateLimiterBatchV4 struct {
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

	loadBalancer *RoundRobinLoadBalancer[*redisBatchV4Pipeline]
	pipelines    []*redisBatchV4Pipeline

	pipelineWg sync.WaitGroup
	closeOnce  sync.Once
	closed     atomic.Bool

	totalRequests         atomic.Int64
	allowedRequests       atomic.Int64
	pipelineBatches       atomic.Int64
	pipelineRequests      atomic.Int64
	queueFullDenies       atomic.Int64
	correlationMismatches atomic.Int64
	requestSeq            atomic.Uint64
}

type redisBatchV4PipelineRequest struct {
	ctx        context.Context
	requestID  uint64
	shardKey   string
	redisKey   string
	n          int
	responseCh chan redisBatchV4PipelineResult
}

func (r *redisBatchV4PipelineRequest) GetID() string {
	if r.shardKey != "" {
		return r.shardKey
	}
	if r.redisKey != "" {
		return r.redisKey
	}
	return "default"
}

type redisBatchV4PipelineResult struct {
	requestID uint64
	allowed   bool
	err       error
}

var (
	errRedisBatchV4PipelineClosed      = errors.New("redis batch v4 pipeline is closed")
	errRedisBatchV4PipelineQueueFull   = errors.New("redis batch v4 pipeline queue is full")
	errRedisBatchV4CorrelationMismatch = errors.New("redis batch v4 correlation mismatch")
)

// NewRedisRateLimiterBatchV4 creates a new sharded batch token-bucket limiter.
func NewRedisRateLimiterBatchV4(client *redis.Client, config RedisRateLimiterBatchV4Config) (*RedisRateLimiterBatchV4, error) {
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
		pipelineMaxWait = defaultRedisRateLimiterBatchV4PipelineMaxWait
	}
	pipelineMaxBatch := config.PipelineMaxBatch
	if pipelineMaxBatch <= 0 {
		pipelineMaxBatch = defaultRedisRateLimiterBatchV4PipelineMaxBatch
	}
	pipelineShards := config.PipelineShards
	if pipelineShards <= 0 {
		pipelineShards = defaultRedisRateLimiterBatchV4PipelineShards
	}
	pipelineQueueSize := config.PipelineQueueSize
	if pipelineQueueSize <= 0 {
		pipelineQueueSize = pipelineMaxBatch * defaultRedisRateLimiterBatchV4QueueFactor
	}
	if pipelineQueueSize <= 0 {
		pipelineQueueSize = 1
	}

	r := &RedisRateLimiterBatchV4{
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
		loadBalancer:      NewRoundRobinLoadBalancer[*redisBatchV4Pipeline](),
	}

	r.pipelines = make([]*redisBatchV4Pipeline, 0, r.pipelineShards)

	for i := 0; i < r.pipelineShards; i++ {
		pipeline := &redisBatchV4Pipeline{
			rrl:              r,
			pipelineMaxWait:  r.pipelineMaxWait,
			pipelineMaxBatch: r.pipelineMaxBatch,
			pipelineQueue:    make(chan *redisBatchV4PipelineRequest, r.pipelineQueueSize),
			pipelineStopCh:   make(chan struct{}),
		}
		r.pipelines = append(r.pipelines, pipeline)
		if err := r.loadBalancer.AddTarget("pipeline-"+strconv.Itoa(i), pipeline, 1); err != nil {
			return nil, err
		}

		r.pipelineWg.Add(1)
		go func(p *redisBatchV4Pipeline) {
			defer r.pipelineWg.Done()
			p.Start()
		}(pipeline)
	}

	return r, nil
}

// Close stops batch pipeline shards.
func (r *RedisRateLimiterBatchV4) Close() {
	r.closeOnce.Do(func() {
		r.closed.Store(true)
		for _, p := range r.pipelines {
			close(p.pipelineStopCh)
		}
		r.pipelineWg.Wait()
	})
}

// Allow checks one request.
func (r *RedisRateLimiterBatchV4) Allow(ctx context.Context, key string) bool {
	return r.AllowN(ctx, key, 1)
}

// InitializeTokenBucket seeds token-bucket state for a key with an exact token count.
func (r *RedisRateLimiterBatchV4) InitializeTokenBucket(ctx context.Context, key string, tokens int) error {
	if r.closed.Load() {
		return errRedisBatchV4PipelineClosed
	}

	if tokens < 0 {
		tokens = 0
	}
	if tokens > r.burst {
		tokens = r.burst
	}

	opCtx, cancel := r.operationContext(ctx)
	defer cancel()

	redisKey := r.redisKey(key)
	now := time.Now().UnixNano()

	pipe := r.client.Pipeline()
	pipe.HSet(
		opCtx,
		redisKey,
		redisTokenBucketFieldTokens, tokens,
		redisTokenBucketFieldLastNS, now,
	)
	if r.stateTTL > 0 {
		pipe.PExpire(opCtx, redisKey, r.stateTTL)
	}

	_, err := pipe.Exec(opCtx)
	return err
}

// AllowN checks n requests.
func (r *RedisRateLimiterBatchV4) AllowN(ctx context.Context, key string, n int) bool {
	if n <= 0 {
		return true
	}

	r.totalRequests.Add(int64(n))

	allowed, _, err := r.allowNWithRequestID(ctx, key, n)
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

func (r *RedisRateLimiterBatchV4) allowNWithRequestID(ctx context.Context, key string, n int) (bool, uint64, error) {
	if n > r.burst {
		return false, 0, nil
	}
	if r.closed.Load() {
		return false, 0, errRedisBatchV4PipelineClosed
	}

	submitCtx, cancel := r.operationContext(ctx)
	defer cancel()

	requestID := r.requestSeq.Add(1)
	shardKey := key
	if shardKey == "" {
		shardKey = strconv.FormatUint(requestID, 10)
	}
	req := &redisBatchV4PipelineRequest{
		ctx:        ctx,
		requestID:  requestID,
		shardKey:   shardKey,
		redisKey:   r.redisKey(key),
		n:          n,
		responseCh: make(chan redisBatchV4PipelineResult, 1),
	}

	_, pipeline, err := r.loadBalancer.SelectTarget(submitCtx)
	if err != nil {
		return false, requestID, err
	}

	_, err = pipeline.handle(submitCtx, req)
	if err != nil {
		if errors.Is(err, errRedisBatchV4PipelineQueueFull) {
			r.queueFullDenies.Add(1)
			return false, requestID, nil
		}
		return false, requestID, err
	}

	select {
	case result := <-req.responseCh:
		if result.requestID != requestID {
			r.correlationMismatches.Add(1)
			return false, requestID, errRedisBatchV4CorrelationMismatch
		}
		return result.allowed, requestID, result.err
	case <-ctx.Done():
		return false, requestID, ctx.Err()
	}
}

// Stats returns a snapshot of internal counters.
func (r *RedisRateLimiterBatchV4) Stats() RedisRateLimiterBatchV4Stats {
	return RedisRateLimiterBatchV4Stats{
		TotalRequests:         r.totalRequests.Load(),
		AllowedRequests:       r.allowedRequests.Load(),
		PipelineBatches:       r.pipelineBatches.Load(),
		PipelineRequests:      r.pipelineRequests.Load(),
		QueueFullDenies:       r.queueFullDenies.Load(),
		CorrelationMismatches: r.correlationMismatches.Load(),
	}
}

func (r *RedisRateLimiterBatchV4) redisKey(key string) string {
	if key == "" {
		return r.baseKey
	}
	return r.baseKey + ":" + key
}

func (r *RedisRateLimiterBatchV4) operationContext(parent context.Context) (context.Context, context.CancelFunc) {
	if r.operationTimeout <= 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, r.operationTimeout)
}

type redisBatchV4Pipeline struct {
	rrl *RedisRateLimiterBatchV4

	pipelineMaxWait  time.Duration
	pipelineMaxBatch int
	pipelineQueue    chan *redisBatchV4PipelineRequest
	pipelineStopCh   chan struct{}
}

func (r *redisBatchV4Pipeline) handle(ctx context.Context, req *redisBatchV4PipelineRequest) (redisBatchV4PipelineResult, error) {
	select {
	case r.pipelineQueue <- req:
		return redisBatchV4PipelineResult{}, nil
	case <-ctx.Done():
		return redisBatchV4PipelineResult{}, ctx.Err()
	case <-r.pipelineStopCh:
		return redisBatchV4PipelineResult{}, errRedisBatchV4PipelineClosed
	default:
		return redisBatchV4PipelineResult{}, errRedisBatchV4PipelineQueueFull
	}
}

func (r *redisBatchV4Pipeline) Start() {
	ticker := time.NewTicker(r.pipelineMaxWait)
	defer ticker.Stop()

	for {
		select {
		case <-r.pipelineStopCh:
			r.failPending(errRedisBatchV4PipelineClosed)
			return
		case first := <-r.pipelineQueue:
			batch := []*redisBatchV4PipelineRequest{first}
			r.collectBatch(&batch)
			r.executePipeline(batch)
		case <-ticker.C:
			batch := make([]*redisBatchV4PipelineRequest, 0, r.pipelineMaxBatch)
			r.collectBatch(&batch)
			if len(batch) > 0 {
				r.executePipeline(batch)
			}
		}
	}
}

func (r *redisBatchV4Pipeline) collectBatch(batch *[]*redisBatchV4PipelineRequest) {
	for len(*batch) < r.pipelineMaxBatch {
		select {
		case req := <-r.pipelineQueue:
			*batch = append(*batch, req)
		default:
			return
		}
	}
}

func (r *redisBatchV4Pipeline) executePipeline(batch []*redisBatchV4PipelineRequest) {
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

		pipe.HIncrBy(ctx, redisKey, redisTokenBucketFieldTokens, refillByKey[redisKey])

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
			r.sendResult(req, redisBatchV4PipelineResult{requestID: req.requestID, err: err})
		}
		return
	}

	for i, req := range batch {
		if req.ctx != nil && req.ctx.Err() != nil {
			continue
		}

		if err := groupErr[req.redisKey]; err != nil {
			r.sendResult(req, redisBatchV4PipelineResult{requestID: req.requestID, err: err})
			continue
		}

		if consumeCmdByIndex[i] != nil {
			if _, err := consumeCmdByIndex[i].Result(); err != nil {
				r.sendResult(req, redisBatchV4PipelineResult{requestID: req.requestID, err: err})
				continue
			}
		}

		r.sendResult(req, redisBatchV4PipelineResult{requestID: req.requestID, allowed: allowByIndex[i]})
	}

}

func (r *redisBatchV4Pipeline) failPending(err error) {
	for {
		select {
		case req := <-r.pipelineQueue:
			if req == nil {
				continue
			}
			r.sendResult(req, redisBatchV4PipelineResult{requestID: req.requestID, err: err})
		default:
			return
		}
	}
}

func (r *redisBatchV4Pipeline) sendResult(req *redisBatchV4PipelineRequest, result redisBatchV4PipelineResult) {
	if req == nil {
		return
	}

	select {
	case req.responseCh <- result:
	default:
	}
}
