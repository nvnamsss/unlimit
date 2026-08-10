package stability

import (
	"context"
	"time"
)

// CircuitBreaker defines the interface for circuit breaker implementations
type CircuitBreaker interface {
	// Execute executes the given function if the circuit breaker allows it
	// Returns the result of the function or an error if the circuit is open
	Execute(fn func() (interface{}, error)) (interface{}, error)

	// ExecuteWithContext executes the function with context support
	ExecuteWithContext(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error)

	// Call is a simplified version of Execute for functions without return values
	Call(fn func() error) error

	// CallWithContext is a simplified version of ExecuteWithContext for functions without return values
	CallWithContext(ctx context.Context, fn func(ctx context.Context) error) error

	// CanExecute checks if the circuit breaker would allow execution without actually executing
	CanExecute() bool

	// State returns the current state of the circuit breaker
	State() State

	// IsOpen returns true if the circuit breaker is in the open state
	IsOpen() bool

	// IsHalfOpen returns true if the circuit breaker is in the half-open state
	IsHalfOpen() bool

	// IsClosed returns true if the circuit breaker is in the closed state
	IsClosed() bool

	// Reset manually resets the circuit breaker to the closed state
	Reset()

	// Trip manually trips the circuit breaker to the open state
	Trip()

	// GetCounts returns the current failure and success counts
	GetCounts() Counts

	// GetConfig returns the current configuration
	GetConfig() Config

	// UpdateConfig updates the circuit breaker configuration
	UpdateConfig(config Config) error

	// Subscribe adds a listener for state change events
	Subscribe(listener StateChangeListener)

	// Unsubscribe removes a state change listener
	Unsubscribe(listener StateChangeListener)
}

// State represents the circuit breaker state
type State int

const (
	// StateClosed indicates the circuit breaker is closed (normal operation)
	StateClosed State = iota
	// StateOpen indicates the circuit breaker is open (failing fast)
	StateOpen
	// StateHalfOpen indicates the circuit breaker is half-open (testing recovery)
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// Counts holds the statistics for the circuit breaker
type Counts struct {
	// Requests is the total number of requests
	Requests uint64
	// TotalSuccesses is the total number of successful requests
	TotalSuccesses uint64
	// TotalFailures is the total number of failed requests
	TotalFailures uint64
	// ConsecutiveSuccesses is the number of consecutive successful requests
	ConsecutiveSuccesses uint64
	// ConsecutiveFailures is the number of consecutive failed requests
	ConsecutiveFailures uint64
}

// SuccessRate returns the success rate as a value between 0 and 1
func (c Counts) SuccessRate() float64 {
	if c.Requests == 0 {
		return 1.0
	}
	return float64(c.TotalSuccesses) / float64(c.Requests)
}

// FailureRate returns the failure rate as a value between 0 and 1
func (c Counts) FailureRate() float64 {
	if c.Requests == 0 {
		return 0.0
	}
	return float64(c.TotalFailures) / float64(c.Requests)
}

// Config holds the configuration for a circuit breaker
type Config struct {
	// Name is an identifier for the circuit breaker
	Name string

	// MaxRequests is the maximum number of requests allowed to pass through
	// when the circuit breaker is half-open
	MaxRequests uint32

	// Interval is the cyclic period of the closed state for the circuit breaker
	// to clear the internal counts
	Interval time.Duration

	// Timeout is the period of the open state, after which the state becomes half-open
	Timeout time.Duration

	// ReadyToTrip is called with a copy of Counts whenever a request fails in the closed state
	// If ReadyToTrip returns true, the circuit breaker will be placed into the open state
	ReadyToTrip func(counts Counts) bool

	// OnStateChange is called whenever the state of the circuit breaker changes
	OnStateChange func(name string, from State, to State)

	// IsSuccessful determines whether the given error should be counted as a success or failure
	// If nil, any non-nil error is considered a failure
	IsSuccessful func(err error) bool
}

// StateChangeListener is called when the circuit breaker state changes
type StateChangeListener interface {
	OnStateChange(name string, from State, to State)
}

// StateChangeFunc is a function adapter for StateChangeListener
type StateChangeFunc func(name string, from State, to State)

func (f StateChangeFunc) OnStateChange(name string, from State, to State) {
	f(name, from, to)
}

// CircuitBreakerError represents an error from the circuit breaker
type CircuitBreakerError struct {
	State      State
	RetryAfter time.Duration
	Message    string
	InnerError error
}

func (e *CircuitBreakerError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "circuit breaker is " + e.State.String()
}

func (e *CircuitBreakerError) Unwrap() error {
	return e.InnerError
}

// KeyedCircuitBreaker manages multiple circuit breakers by key
type KeyedCircuitBreaker interface {
	// Execute executes the function for the given key
	Execute(key string, fn func() (interface{}, error)) (interface{}, error)

	// ExecuteWithContext executes the function with context for the given key
	ExecuteWithContext(ctx context.Context, key string, fn func(ctx context.Context) (interface{}, error)) (interface{}, error)

	// Call executes the function for the given key without return value
	Call(key string, fn func() error) error

	// CallWithContext executes the function with context for the given key without return value
	CallWithContext(ctx context.Context, key string, fn func(ctx context.Context) error) error

	// GetCircuitBreaker returns the circuit breaker for the given key
	GetCircuitBreaker(key string) CircuitBreaker

	// RemoveCircuitBreaker removes the circuit breaker for the given key
	RemoveCircuitBreaker(key string)

	// Keys returns all currently tracked keys
	Keys() []string

	// Count returns the number of active circuit breakers
	Count() int

	// Reset resets all circuit breakers
	Reset()

	// ResetKey resets the circuit breaker for the given key
	ResetKey(key string)

	// Cleanup removes inactive circuit breakers
	Cleanup()
}

// MultiLevelCircuitBreaker manages circuit breakers at multiple levels
type MultiLevelCircuitBreaker interface {
	// Execute executes the function if all levels allow it
	Execute(levels []string, fn func() (interface{}, error)) (interface{}, error)

	// ExecuteWithContext executes the function with context if all levels allow it
	ExecuteWithContext(ctx context.Context, levels []string, fn func(ctx context.Context) (interface{}, error)) (interface{}, error)

	// AddLevel adds a circuit breaker at the specified level
	AddLevel(level string, breaker CircuitBreaker)

	// RemoveLevel removes the circuit breaker at the specified level
	RemoveLevel(level string)

	// GetLevel returns the circuit breaker at the specified level
	GetLevel(level string) CircuitBreaker

	// Levels returns all configured levels
	Levels() []string

	// CanExecute checks if all levels would allow execution
	CanExecute(levels []string) bool
}

// AdaptiveCircuitBreaker can adjust its configuration based on system conditions
type AdaptiveCircuitBreaker interface {
	CircuitBreaker

	// Adapt adjusts the circuit breaker configuration based on the given metric
	Adapt(metric float64)

	// SetAdaptationRule sets the rule for how the configuration should adapt
	SetAdaptationRule(rule CircuitBreakerAdaptationRule)

	// GetAdaptationRule returns the current adaptation rule
	GetAdaptationRule() CircuitBreakerAdaptationRule

	// IsAdaptive returns whether adaptive behavior is enabled
	IsAdaptive() bool

	// SetAdaptive enables or disables adaptive behavior
	SetAdaptive(adaptive bool)
}

// CircuitBreakerAdaptationRule defines how circuit breaker configuration should adapt
type CircuitBreakerAdaptationRule interface {
	// Adapt returns a new configuration based on the current config and metric
	Adapt(currentConfig Config, metric float64) Config

	// ShouldAdapt returns whether adaptation should occur given the metric
	ShouldAdapt(metric float64) bool
}

// CircuitBreakerMetrics provides metrics about circuit breaker performance
type CircuitBreakerMetrics interface {
	// GetCounts returns the current counts
	GetCounts() Counts

	// GetStateHistory returns the history of state changes
	GetStateHistory() []StateChange

	// GetLatencyStats returns latency statistics
	GetLatencyStats() LatencyStats

	// Reset resets all metrics
	Reset()
}

// StateChange represents a state change event
type StateChange struct {
	Timestamp time.Time
	From      State
	To        State
	Reason    string
}

// LatencyStats holds latency statistics
type LatencyStats struct {
	Mean  time.Duration
	P50   time.Duration
	P90   time.Duration
	P95   time.Duration
	P99   time.Duration
	Min   time.Duration
	Max   time.Duration
	Count uint64
}

// CircuitBreakerFactory creates circuit breakers
type CircuitBreakerFactory interface {
	// Create creates a new circuit breaker with the given configuration
	Create(config Config) CircuitBreaker

	// CreateKeyed creates a new keyed circuit breaker
	CreateKeyed(config Config) KeyedCircuitBreaker

	// CreateMultiLevel creates a new multi-level circuit breaker
	CreateMultiLevel() MultiLevelCircuitBreaker

	// CreateAdaptive creates a new adaptive circuit breaker
	CreateAdaptive(config Config, rule CircuitBreakerAdaptationRule) AdaptiveCircuitBreaker

	// CreateWithMetrics creates a circuit breaker with metrics collection
	CreateWithMetrics(config Config) (CircuitBreaker, CircuitBreakerMetrics)
}

// Bulkhead provides resource isolation by limiting concurrent executions
type Bulkhead interface {
	// Execute executes the function if a slot is available
	Execute(fn func() (interface{}, error)) (interface{}, error)

	// ExecuteWithContext executes the function with context if a slot is available
	ExecuteWithContext(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error)

	// TryExecute attempts to execute the function without blocking
	// Returns false if no slot is available
	TryExecute(fn func() (interface{}, error)) (interface{}, bool, error)

	// AvailableSlots returns the number of available execution slots
	AvailableSlots() int

	// MaxSlots returns the maximum number of execution slots
	MaxSlots() int

	// ActiveExecutions returns the number of currently active executions
	ActiveExecutions() int
}

// ResiliencePattern combines multiple stability patterns
type ResiliencePattern interface {
	// Execute executes the function with all configured patterns applied
	Execute(fn func() (interface{}, error)) (interface{}, error)

	// ExecuteWithContext executes the function with context and all patterns applied
	ExecuteWithContext(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error)

	// WithCircuitBreaker adds circuit breaker protection
	WithCircuitBreaker(breaker CircuitBreaker) ResiliencePattern

	// WithRetry adds retry logic
	WithRetry(policy RetryPolicy) ResiliencePattern

	// WithTimeout adds timeout protection
	WithTimeout(timeout time.Duration) ResiliencePattern

	// WithBulkhead adds bulkhead isolation
	WithBulkhead(bulkhead Bulkhead) ResiliencePattern

	// WithFallback adds fallback logic
	WithFallback(fallback func(error) (interface{}, error)) ResiliencePattern
}

// RetryPolicy defines retry behavior
type RetryPolicy interface {
	// ShouldRetry determines if a retry should be attempted
	ShouldRetry(attempt int, err error) bool

	// NextDelay returns the delay before the next retry attempt
	NextDelay(attempt int) time.Duration

	// MaxAttempts returns the maximum number of retry attempts
	MaxAttempts() int
}

// DefaultConfig returns a default circuit breaker configuration
func DefaultConfig() Config {
	return Config{
		Name:        "default",
		MaxRequests: 1,
		Interval:    time.Minute,
		Timeout:     time.Minute,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures > 5
		},
		IsSuccessful: func(err error) bool {
			return err == nil
		},
	}
}

// IsCircuitBreakerError checks if an error is a circuit breaker error
func IsCircuitBreakerError(err error) bool {
	_, ok := err.(*CircuitBreakerError)
	return ok
}

// NewCircuitBreakerError creates a new circuit breaker error
func NewCircuitBreakerError(state State, message string) *CircuitBreakerError {
	return &CircuitBreakerError{
		State:   state,
		Message: message,
	}
}
