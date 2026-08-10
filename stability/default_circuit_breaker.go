package stability

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Exported sentinel errors for configuration validation
var (
	// ErrInvalidMaxRequests is returned when MaxRequests is zero or negative
	ErrInvalidMaxRequests = errors.New("maxRequests must be greater than 0")

	// ErrInvalidInterval is returned when Interval is zero or negative
	ErrInvalidInterval = errors.New("interval must be greater than 0")

	// ErrInvalidTimeout is returned when Timeout is zero or negative
	ErrInvalidTimeout = errors.New("timeout must be greater than 0")
)

// DefaultCircuitBreaker implements the CircuitBreaker interface
type DefaultCircuitBreaker struct {
	mu            sync.RWMutex
	config        Config
	counts        Counts
	state         State
	openTime      time.Time
	lastClearTime time.Time
	listeners     []StateChangeListener
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(config Config) CircuitBreaker {
	cb := &DefaultCircuitBreaker{
		config:        config,
		state:         StateClosed,
		lastClearTime: time.Now(),
		listeners:     make([]StateChangeListener, 0),
	}

	// Set default values if not provided
	if cb.config.MaxRequests == 0 {
		cb.config.MaxRequests = 1
	}
	if cb.config.Interval == 0 {
		cb.config.Interval = time.Minute
	}
	if cb.config.Timeout == 0 {
		cb.config.Timeout = time.Minute
	}
	if cb.config.ReadyToTrip == nil {
		cb.config.ReadyToTrip = func(counts Counts) bool {
			return counts.ConsecutiveFailures > 5
		}
	}
	if cb.config.IsSuccessful == nil {
		cb.config.IsSuccessful = func(err error) bool {
			return err == nil
		}
	}

	return cb
}

// Execute executes the given function if the circuit breaker allows it
func (cb *DefaultCircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	return cb.ExecuteWithContext(context.Background(), func(ctx context.Context) (interface{}, error) {
		return fn()
	})
}

// ExecuteWithContext executes the function with context support
func (cb *DefaultCircuitBreaker) ExecuteWithContext(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	if !cb.allowRequest() {
		return nil, NewCircuitBreakerError(cb.State(), "circuit breaker is open")
	}

	start := time.Now()
	result, err := fn(ctx)
	duration := time.Since(start)

	cb.recordResult(err, duration)
	return result, err
}

// Call is a simplified version of Execute for functions without return values
func (cb *DefaultCircuitBreaker) Call(fn func() error) error {
	_, err := cb.Execute(func() (interface{}, error) {
		return nil, fn()
	})
	return err
}

// CallWithContext is a simplified version of ExecuteWithContext for functions without return values
func (cb *DefaultCircuitBreaker) CallWithContext(ctx context.Context, fn func(ctx context.Context) error) error {
	_, err := cb.ExecuteWithContext(ctx, func(ctx context.Context) (interface{}, error) {
		return nil, fn(ctx)
	})
	return err
}

// CanExecute checks if the circuit breaker would allow execution without actually executing
func (cb *DefaultCircuitBreaker) CanExecute() bool {
	return cb.allowRequest()
}

// State returns the current state of the circuit breaker
func (cb *DefaultCircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.currentState()
}

// IsOpen returns true if the circuit breaker is in the open state
func (cb *DefaultCircuitBreaker) IsOpen() bool {
	return cb.State() == StateOpen
}

// IsHalfOpen returns true if the circuit breaker is in the half-open state
func (cb *DefaultCircuitBreaker) IsHalfOpen() bool {
	return cb.State() == StateHalfOpen
}

// IsClosed returns true if the circuit breaker is in the closed state
func (cb *DefaultCircuitBreaker) IsClosed() bool {
	return cb.State() == StateClosed
}

// Reset manually resets the circuit breaker to the closed state
func (cb *DefaultCircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	oldState := cb.state
	cb.state = StateClosed
	cb.counts = Counts{}
	cb.openTime = time.Time{}
	cb.lastClearTime = time.Now()

	if oldState != StateClosed {
		cb.notifyStateChange(oldState, StateClosed)
	}
}

// Trip manually trips the circuit breaker to the open state
func (cb *DefaultCircuitBreaker) Trip() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	oldState := cb.state
	cb.state = StateOpen
	cb.openTime = time.Now()

	if oldState != StateOpen {
		cb.notifyStateChange(oldState, StateOpen)
	}
}

// GetCounts returns the current failure and success counts
func (cb *DefaultCircuitBreaker) GetCounts() Counts {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.counts
}

// GetConfig returns the current configuration
func (cb *DefaultCircuitBreaker) GetConfig() Config {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.config
}

// UpdateConfig updates the circuit breaker configuration
func (cb *DefaultCircuitBreaker) UpdateConfig(config Config) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Validate configuration
	if config.MaxRequests == 0 {
		return ErrInvalidMaxRequests
	}
	if config.Interval <= 0 {
		return ErrInvalidInterval
	}
	if config.Timeout <= 0 {
		return ErrInvalidTimeout
	}

	cb.config = config
	return nil
}

// Subscribe adds a listener for state change events
func (cb *DefaultCircuitBreaker) Subscribe(listener StateChangeListener) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.listeners = append(cb.listeners, listener)
}

// Unsubscribe removes a state change listener
func (cb *DefaultCircuitBreaker) Unsubscribe(listener StateChangeListener) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	for i, l := range cb.listeners {
		// Compare function pointers or use reflection for interface comparison
		if &l == &listener {
			cb.listeners = append(cb.listeners[:i], cb.listeners[i+1:]...)
			break
		}
	}
}

// allowRequest determines if a request should be allowed
func (cb *DefaultCircuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	state := cb.currentState()

	switch state {
	case StateClosed:
		return true
	case StateOpen:
		return false
	case StateHalfOpen:
		return cb.counts.Requests < uint64(cb.config.MaxRequests)
	default:
		return false
	}
}

// currentState returns the current state, handling transitions
func (cb *DefaultCircuitBreaker) currentState() State {
	now := time.Now()

	switch cb.state {
	case StateClosed:
		// Check if we should clear counts based on interval
		if cb.config.Interval > 0 && now.Sub(cb.lastClearTime) > cb.config.Interval {
			cb.counts = Counts{}
			cb.lastClearTime = now
		}
		return StateClosed

	case StateOpen:
		// Check if timeout has passed to transition to half-open
		if now.Sub(cb.openTime) > cb.config.Timeout {
			oldState := cb.state
			cb.state = StateHalfOpen
			cb.counts = Counts{}
			cb.lastClearTime = now
			cb.notifyStateChange(oldState, StateHalfOpen)
			return StateHalfOpen
		}
		return StateOpen

	case StateHalfOpen:
		return StateHalfOpen

	default:
		return StateClosed
	}
}

// recordResult records the result of a request execution
func (cb *DefaultCircuitBreaker) recordResult(err error, duration time.Duration) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.counts.Requests++

	if cb.config.IsSuccessful(err) {
		cb.onSuccess()
	} else {
		cb.onFailure()
	}
}

// onSuccess handles a successful request
func (cb *DefaultCircuitBreaker) onSuccess() {
	cb.counts.TotalSuccesses++
	cb.counts.ConsecutiveSuccesses++
	cb.counts.ConsecutiveFailures = 0

	if cb.state == StateHalfOpen {
		// Transition back to closed if we have enough successful requests
		if cb.counts.ConsecutiveSuccesses >= uint64(cb.config.MaxRequests) {
			oldState := cb.state
			cb.state = StateClosed
			cb.counts = Counts{}
			cb.lastClearTime = time.Now()
			cb.notifyStateChange(oldState, StateClosed)
		}
	}
}

// onFailure handles a failed request
func (cb *DefaultCircuitBreaker) onFailure() {
	cb.counts.TotalFailures++
	cb.counts.ConsecutiveFailures++
	cb.counts.ConsecutiveSuccesses = 0

	if cb.state == StateClosed {
		// Check if we should trip to open
		if cb.config.ReadyToTrip(cb.counts) {
			oldState := cb.state
			cb.state = StateOpen
			cb.openTime = time.Now()
			cb.notifyStateChange(oldState, StateOpen)
		}
	} else if cb.state == StateHalfOpen {
		// Any failure in half-open state trips back to open
		oldState := cb.state
		cb.state = StateOpen
		cb.openTime = time.Now()
		cb.notifyStateChange(oldState, StateOpen)
	}
}

// notifyStateChange notifies all listeners of a state change
func (cb *DefaultCircuitBreaker) notifyStateChange(from, to State) {
	// Call config callback if set
	if cb.config.OnStateChange != nil {
		cb.config.OnStateChange(cb.config.Name, from, to)
	}

	// Notify all subscribed listeners
	for _, listener := range cb.listeners {
		// Use goroutine to avoid blocking
		go listener.OnStateChange(cb.config.Name, from, to)
	}
}

// MetricsCollector provides metrics collection for circuit breakers
type MetricsCollector struct {
	mu           sync.RWMutex
	stateHistory []StateChange
	latencies    []time.Duration
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		stateHistory: make([]StateChange, 0),
		latencies:    make([]time.Duration, 0),
	}
}

// OnStateChange implements StateChangeListener
func (mc *MetricsCollector) OnStateChange(name string, from State, to State) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.stateHistory = append(mc.stateHistory, StateChange{
		Timestamp: time.Now(),
		From:      from,
		To:        to,
		Reason:    "state transition",
	})
}

// RecordLatency records request latency
func (mc *MetricsCollector) RecordLatency(duration time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.latencies = append(mc.latencies, duration)
}

// GetStateHistory returns the history of state changes
func (mc *MetricsCollector) GetStateHistory() []StateChange {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	history := make([]StateChange, len(mc.stateHistory))
	copy(history, mc.stateHistory)
	return history
}

// GetLatencyStats returns latency statistics
func (mc *MetricsCollector) GetLatencyStats() LatencyStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if len(mc.latencies) == 0 {
		return LatencyStats{}
	}

	// Calculate basic statistics
	var sum time.Duration
	min := mc.latencies[0]
	max := mc.latencies[0]

	for _, lat := range mc.latencies {
		sum += lat
		if lat < min {
			min = lat
		}
		if lat > max {
			max = lat
		}
	}

	mean := sum / time.Duration(len(mc.latencies))

	return LatencyStats{
		Mean:  mean,
		Min:   min,
		Max:   max,
		Count: uint64(len(mc.latencies)),
		// For percentiles, would need sorted data - simplified for now
		P50: mean,
		P90: mean,
		P95: mean,
		P99: mean,
	}
}

// Reset resets all collected metrics
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.stateHistory = mc.stateHistory[:0]
	mc.latencies = mc.latencies[:0]
}
