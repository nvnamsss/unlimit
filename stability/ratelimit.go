package stability

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// RateLimiter defines the interface for rate limiting implementations
type RateLimiter interface {
	// Allow checks if a request is allowed under the current rate limit
	// Returns true if the request should be allowed, false otherwise
	Allow() bool

	// AllowN checks if n requests are allowed under the current rate limit
	// Returns true if all n requests should be allowed, false otherwise
	AllowN(n int) bool

	// Wait blocks until a request can be allowed under the rate limit
	// Returns an error if the context is canceled before the request can proceed
	Wait(ctx context.Context) error

	// WaitN blocks until n requests can be allowed under the rate limit
	// Returns an error if the context is canceled before the requests can proceed
	WaitN(ctx context.Context, n int) error

	// Reserve reserves a request and returns a Reservation that indicates
	// how long the caller must wait before the request can proceed
	Reserve() Reservation

	// ReserveN reserves n requests and returns a Reservation that indicates
	// how long the caller must wait before the requests can proceed
	ReserveN(n int) Reservation

	// Limit returns the current rate limit (requests per second)
	Limit() float64

	// Burst returns the current burst size (maximum requests that can be made instantly)
	Burst() int

	// SetLimit changes the rate limit (requests per second)
	SetLimit(limit float64)

	// SetBurst changes the burst size
	SetBurst(burst int)

	// Tokens returns the current number of available tokens
	Tokens() float64

	// Reset resets the rate limiter to its initial state
	Reset()
}

// Reservation represents a reservation for rate-limited requests
type Reservation interface {
	// OK returns whether the reservation is valid
	OK() bool

	// Delay returns the duration to wait before the reserved requests can proceed
	Delay() time.Duration

	// DelayFrom returns the duration to wait from the given time
	DelayFrom(now time.Time) time.Duration

	// Cancel cancels the reservation, returning tokens to the rate limiter
	Cancel()

	// CancelAt cancels the reservation at the specified time
	CancelAt(now time.Time)
}

// KeyedRateLimiter defines the interface for rate limiting with keys
// Useful for per-user, per-IP, or per-API-key rate limiting
type KeyedRateLimiter interface {
	// Allow checks if a request is allowed for the given key
	Allow(key string) bool

	// AllowN checks if n requests are allowed for the given key
	AllowN(key string, n int) bool

	// Wait blocks until a request can be allowed for the given key
	Wait(ctx context.Context, key string) error

	// WaitN blocks until n requests can be allowed for the given key
	WaitN(ctx context.Context, key string, n int) error

	// Reserve reserves a request for the given key
	Reserve(key string) Reservation

	// ReserveN reserves n requests for the given key
	ReserveN(key string, n int) Reservation

	// GetLimiter returns the rate limiter for the given key
	// Creates a new one if it doesn't exist
	GetLimiter(key string) RateLimiter

	// RemoveLimiter removes the rate limiter for the given key
	RemoveLimiter(key string)

	// Keys returns all currently tracked keys
	Keys() []string

	// Count returns the number of active rate limiters
	Count() int

	// Reset resets all rate limiters
	Reset()

	// Cleanup removes inactive rate limiters (implementation dependent)
	Cleanup()
}

// HierarchicalRateLimiter defines the interface for hierarchical rate limiting
// Useful for global + per-user limits, or service + endpoint limits
type HierarchicalRateLimiter interface {
	// Allow checks if a request is allowed at all levels
	Allow(levels ...string) bool

	// AllowN checks if n requests are allowed at all levels
	AllowN(n int, levels ...string) bool

	// Wait blocks until a request can be allowed at all levels
	Wait(ctx context.Context, levels ...string) error

	// WaitN blocks until n requests can be allowed at all levels
	WaitN(ctx context.Context, n int, levels ...string) error

	// AddLevel adds a new rate limiting level
	AddLevel(level string, limiter RateLimiter)

	// RemoveLevel removes a rate limiting level
	RemoveLevel(level string)

	// GetLevel returns the rate limiter for the given level
	GetLevel(level string) RateLimiter

	// Levels returns all configured levels
	Levels() []string
}

// AdaptiveRateLimiter defines the interface for adaptive rate limiting
// Can adjust limits based on system conditions or feedback
type AdaptiveRateLimiter interface {
	RateLimiter

	// Adapt adjusts the rate limit based on the given metric
	// metric could be error rate, latency, CPU usage, etc.
	Adapt(metric float64)

	// SetAdaptationRule sets the rule for how limits should adapt
	SetAdaptationRule(rule AdaptationRule)

	// GetAdaptationRule returns the current adaptation rule
	GetAdaptationRule() AdaptationRule

	// IsAdaptive returns whether adaptive behavior is enabled
	IsAdaptive() bool

	// SetAdaptive enables or disables adaptive behavior
	SetAdaptive(adaptive bool)
}

// AdaptationRule defines how rate limits should adapt based on metrics
type AdaptationRule interface {
	// Calculate returns the new limit based on current limit and metric
	Calculate(currentLimit float64, metric float64) float64

	// ShouldAdapt returns whether adaptation should occur given the metric
	ShouldAdapt(metric float64) bool
}

// RateLimiterConfig holds configuration for rate limiter creation
type RateLimiterConfig struct {
	// Limit is the rate limit in requests per second
	Limit float64

	// Burst is the maximum number of requests that can be made instantly
	Burst int

	// Algorithm specifies the rate limiting algorithm to use
	Algorithm string

	// Window is the time window for window-based algorithms
	Window time.Duration

	// CleanupInterval for keyed rate limiters
	CleanupInterval time.Duration

	// TTL for automatic cleanup of inactive limiters
	TTL time.Duration
}

// RateLimiterFactory creates rate limiters based on configuration
type RateLimiterFactory interface {
	// Create creates a new rate limiter with the given configuration
	Create(config RateLimiterConfig) (RateLimiter, error)

	// CreateKeyed creates a new keyed rate limiter
	CreateKeyed(config RateLimiterConfig) (KeyedRateLimiter, error)

	// CreateHierarchical creates a new hierarchical rate limiter
	CreateHierarchical() HierarchicalRateLimiter

	// CreateAdaptive creates a new adaptive rate limiter
	CreateAdaptive(config RateLimiterConfig, rule AdaptationRule) (AdaptiveRateLimiter, error)

	// SupportedAlgorithms returns the list of supported algorithms
	SupportedAlgorithms() []string
}

// RateLimitedError is returned when a request is rate limited
type RateLimitedError struct {
	// RetryAfter indicates when the client can retry
	RetryAfter time.Duration

	// Limit is the current rate limit
	Limit float64

	// Remaining is the number of remaining requests in the current window
	Remaining int

	// ResetTime is when the rate limit window resets
	ResetTime time.Time

	// Key is the rate limiting key (if applicable)
	Key string
}

func (e *RateLimitedError) Error() string {
	if e.RetryAfter > 0 {
		return "rate limit exceeded, retry after " + e.RetryAfter.String()
	}
	return "rate limit exceeded"
}

// RateLimitStatus provides information about current rate limit status
type RateLimitStatus struct {
	// Allowed indicates if the last request was allowed
	Allowed bool

	// Limit is the current rate limit
	Limit float64

	// Remaining is the number of remaining requests
	Remaining int

	// ResetTime is when the rate limit resets
	ResetTime time.Time

	// RetryAfter indicates when to retry if not allowed
	RetryAfter time.Duration
}

// LeakyBucket implements a leaky bucket rate limiter
// The bucket has a fixed capacity and leaks at a constant rate
type LeakyBucket struct {
	mu          sync.RWMutex
	capacity    int             // Maximum capacity of the bucket
	leak        time.Duration   // Time between each leak (1/rate)
	lastLeak    time.Time       // Last time the bucket leaked
	current     int             // Current level in the bucket
	closed      bool            // Whether the limiter is closed
	leakTicker  *time.Ticker    // Ticker for continuous leaking
	stopCh      chan struct{}   // Channel to stop the background goroutine
	waitQueue   []chan struct{} // Queue for waiting requests
	waitQueueMu sync.Mutex      // Mutex for wait queue
}

// NewLeakyBucket creates a new leaky bucket rate limiter
func NewLeakyBucket(capacity int, rate float64) *LeakyBucket {
	if capacity <= 0 {
		capacity = 1
	}
	if rate <= 0 {
		rate = 1
	}

	leak := time.Duration(float64(time.Second) / rate)

	lb := &LeakyBucket{
		capacity:  capacity,
		leak:      leak,
		lastLeak:  time.Now(),
		current:   0,
		stopCh:    make(chan struct{}),
		waitQueue: make([]chan struct{}, 0),
	}

	// Start background goroutine for continuous leaking
	go lb.startLeaking()

	return lb
}

// startLeaking runs the background goroutine that continuously leaks from the bucket
func (lb *LeakyBucket) startLeaking() {
	lb.leakTicker = time.NewTicker(lb.leak)
	defer lb.leakTicker.Stop()

	for {
		select {
		case <-lb.leakTicker.C:
			lb.doLeak()
		case <-lb.stopCh:
			return
		}
	}
}

// doLeak performs a single leak operation
func (lb *LeakyBucket) doLeak() {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if lb.current > 0 {
		lb.current--
		lb.lastLeak = time.Now()

		// Notify one waiting request if any
		lb.waitQueueMu.Lock()
		if len(lb.waitQueue) > 0 {
			ch := lb.waitQueue[0]
			lb.waitQueue = lb.waitQueue[1:]
			close(ch)
		}
		lb.waitQueueMu.Unlock()
	}
}

// Allow checks if a request is allowed under the current rate limit
func (lb *LeakyBucket) Allow() bool {
	return lb.AllowN(1)
}

// AllowN checks if n requests are allowed under the current rate limit
func (lb *LeakyBucket) AllowN(n int) bool {
	if n <= 0 {
		return true
	}

	lb.mu.Lock()
	defer lb.mu.Unlock()

	if lb.closed {
		return false
	}

	// Check if we have enough capacity
	if lb.current+n <= lb.capacity {
		lb.current += n
		return true
	}

	return false
}

// Wait blocks until a request can be allowed under the rate limit
func (lb *LeakyBucket) Wait(ctx context.Context) error {
	return lb.WaitN(ctx, 1)
}

// WaitN blocks until n requests can be allowed under the rate limit
func (lb *LeakyBucket) WaitN(ctx context.Context, n int) error {
	if n <= 0 {
		return nil
	}

	// Try immediate allow first
	if lb.AllowN(n) {
		return nil
	}

	// Need to wait for space in bucket
	for i := 0; i < n; i++ {
		if err := lb.waitForSpace(ctx); err != nil {
			return err
		}
	}

	return nil
}

// waitForSpace waits for a single space in the bucket
func (lb *LeakyBucket) waitForSpace(ctx context.Context) error {
	lb.waitQueueMu.Lock()
	waitCh := make(chan struct{})
	lb.waitQueue = append(lb.waitQueue, waitCh)
	lb.waitQueueMu.Unlock()

	select {
	case <-waitCh:
		// Space became available, try to add to bucket
		lb.mu.Lock()
		if lb.current < lb.capacity && !lb.closed {
			lb.current++
			lb.mu.Unlock()
			return nil
		}
		lb.mu.Unlock()
		// If we couldn't add (bucket full or closed), try again
		return lb.waitForSpace(ctx)
	case <-ctx.Done():
		// Remove from wait queue
		lb.waitQueueMu.Lock()
		for i, ch := range lb.waitQueue {
			if ch == waitCh {
				lb.waitQueue = append(lb.waitQueue[:i], lb.waitQueue[i+1:]...)
				break
			}
		}
		lb.waitQueueMu.Unlock()
		return ctx.Err()
	}
}

// Reserve reserves a request and returns a Reservation
func (lb *LeakyBucket) Reserve() Reservation {
	return lb.ReserveN(1)
}

// ReserveN reserves n requests and returns a Reservation
func (lb *LeakyBucket) ReserveN(n int) Reservation {
	if n <= 0 {
		return &leakyBucketReservation{ok: true, delay: 0}
	}

	lb.mu.RLock()
	defer lb.mu.RUnlock()

	if lb.closed {
		return &leakyBucketReservation{ok: false}
	}

	// Calculate delay needed
	availableSpace := lb.capacity - lb.current

	if n <= availableSpace {
		// Can be accommodated immediately
		return &leakyBucketReservation{
			ok:       true,
			delay:    0,
			lb:       lb,
			tokens:   n,
			reserved: true,
		}
	}

	// Calculate how long we need to wait
	tokensNeeded := n - availableSpace
	delay := time.Duration(tokensNeeded) * lb.leak

	return &leakyBucketReservation{
		ok:     true,
		delay:  delay,
		lb:     lb,
		tokens: n,
	}
}

// Limit returns the current rate limit (requests per second)
func (lb *LeakyBucket) Limit() float64 {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return float64(time.Second) / float64(lb.leak)
}

// Burst returns the current burst size (maximum requests that can be made instantly)
func (lb *LeakyBucket) Burst() int {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return lb.capacity
}

// SetLimit changes the rate limit (requests per second)
func (lb *LeakyBucket) SetLimit(limit float64) {
	if limit <= 0 {
		return
	}

	lb.mu.Lock()
	defer lb.mu.Unlock()

	newLeak := time.Duration(float64(time.Second) / limit)
	lb.leak = newLeak

	// Restart the leaking process with new rate
	if lb.leakTicker != nil {
		lb.leakTicker.Stop()
		lb.leakTicker = time.NewTicker(lb.leak)
	}
}

// SetBurst changes the burst size
func (lb *LeakyBucket) SetBurst(burst int) {
	if burst <= 0 {
		return
	}

	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.capacity = burst
	// If current level exceeds new capacity, adjust it
	if lb.current > lb.capacity {
		lb.current = lb.capacity
	}
}

// Tokens returns the current number of available tokens
func (lb *LeakyBucket) Tokens() float64 {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return float64(lb.capacity - lb.current)
}

// Reset resets the rate limiter to its initial state
func (lb *LeakyBucket) Reset() {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.current = 0
	lb.lastLeak = time.Now()

	// Notify all waiting requests
	lb.waitQueueMu.Lock()
	for _, ch := range lb.waitQueue {
		close(ch)
	}
	lb.waitQueue = lb.waitQueue[:0]
	lb.waitQueueMu.Unlock()
}

// Close stops the leaky bucket and releases resources
func (lb *LeakyBucket) Close() {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if lb.closed {
		return
	}

	lb.closed = true
	close(lb.stopCh)

	// Notify all waiting requests
	lb.waitQueueMu.Lock()
	for _, ch := range lb.waitQueue {
		close(ch)
	}
	lb.waitQueue = lb.waitQueue[:0]
	lb.waitQueueMu.Unlock()
}

// leakyBucketReservation implements the Reservation interface for leaky bucket
type leakyBucketReservation struct {
	ok       bool
	delay    time.Duration
	lb       *LeakyBucket
	tokens   int
	reserved bool
	used     int32 // atomic flag to track if reservation was used
}

// OK returns whether the reservation is valid
func (r *leakyBucketReservation) OK() bool {
	return r.ok
}

// Delay returns the duration to wait before the reserved requests can proceed
func (r *leakyBucketReservation) Delay() time.Duration {
	return r.delay
}

// DelayFrom returns the duration to wait from the given time
func (r *leakyBucketReservation) DelayFrom(now time.Time) time.Duration {
	return r.delay
}

// Cancel cancels the reservation, returning tokens to the rate limiter
func (r *leakyBucketReservation) Cancel() {
	r.CancelAt(time.Now())
}

// CancelAt cancels the reservation at the specified time
func (r *leakyBucketReservation) CancelAt(now time.Time) {
	if !r.ok || r.lb == nil || !atomic.CompareAndSwapInt32(&r.used, 0, 1) {
		return
	}

	if r.reserved {
		r.lb.mu.Lock()
		r.lb.current -= r.tokens
		if r.lb.current < 0 {
			r.lb.current = 0
		}
		r.lb.mu.Unlock()
	}
}

func NewLeakyBucketReservation(capacity int, rate float64) Reservation {
	return &leakyBucketReservation{
		ok:       true,
		delay:    0,
		lb:       NewLeakyBucket(capacity, rate),
		tokens:   0,
		reserved: false,
	}
}
