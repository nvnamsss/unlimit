package efflux

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/voidforge-studios/unlimit/logger"
)

// PollingMetrics contains current polling statistics.
type PollingMetrics struct {
	PollCount   int64
	ErrorCount  int64
	ResultCount int64
}

// Poller defines the interface for implementing custom polling logic.
// Implementations should be thread-safe if used with multiple workers.
type Poller[T any] interface {
	// Poll executes a single poll operation and returns results, a stop signal, or an error.
	// The context can be used to cancel long-running poll operations.
	// If shouldStop is true, the polling manager will stop this poller.
	Poll(ctx context.Context, metrics PollingMetrics) (results []T, shouldStop bool, err error)
}

// PollerFunc is a function adapter for the Poller interface.
// It receives current metrics and returns results, a stop signal, and an error.
type PollerFunc[T any] func(ctx context.Context, metrics PollingMetrics) ([]T, bool, error)

// Poll implements the Poller interface for PollerFunc.
func (f PollerFunc[T]) Poll(ctx context.Context, metrics PollingMetrics) ([]T, bool, error) {
	return f(ctx, metrics)
}

// PollingCallbacks provides hooks for monitoring polling lifecycle events.
// All callbacks are optional and called synchronously from the poller goroutine.
type PollingCallbacks[T any] struct {
	// OnPollStart is called before each poll operation.
	OnPollStart func(ctx context.Context, pollerID string)

	// OnPollSuccess is called after a successful poll with the results.
	OnPollSuccess func(ctx context.Context, pollerID string, results []T, duration time.Duration)

	// OnPollError is called when a poll operation fails.
	OnPollError func(ctx context.Context, pollerID string, err error, attempt int)

	// OnRetry is called before retrying a failed poll operation.
	OnRetry func(ctx context.Context, pollerID string, attempt int, backoff time.Duration)
}

// PollingConfig contains configuration for a poller.
type PollingConfig struct {
	// Interval is the base duration between poll operations.
	Interval time.Duration

	// Jitter adds randomness to the interval to prevent thundering herd.
	// The actual interval will be: Interval + random(0, Jitter)
	// Set to 0 to disable jitter.
	Jitter time.Duration

	// BackoffFunc determines the wait duration before retrying after errors.
	// If nil, uses exponential backoff with a 5-minute cap.
	// Return 0 to stop retrying.
	BackoffFunc func(attempt int) time.Duration

	// MaxRetries is the maximum number of retry attempts per poll cycle.
	// Set to 0 to disable retries (fail immediately).
	// Set to -1 for unlimited retries (not recommended).
	MaxRetries int

	// ResultChannelSize is the buffer size for the result channel.
	// Set to 0 to disable channel-based result delivery.
	// If > 0, results are sent to the channel in addition to callbacks.
	ResultChannelSize int
}

// pollerState holds the state of a running poller.
type pollerState[T any] struct {
	poller    Poller[T]
	config    PollingConfig
	callbacks PollingCallbacks[T]

	// Lifecycle
	ctx       context.Context
	cancel    context.CancelFunc
	done      chan struct{}
	startTime time.Time

	// Results
	resultCh chan []T

	// Metrics
	pollCount   atomic.Int64
	errorCount  atomic.Int64
	resultCount atomic.Int64
}

// PollerInfo contains information about a running poller.
type PollerInfo struct {
	ID        string
	StartTime time.Time
	Metrics   PollingMetrics
}

// PollingManager manages multiple pollers concurrently.
// Each poller runs independently with its own configuration and lifecycle.
type PollingManager[T any] struct {
	pollers   map[string]*pollerState[T]
	pollersMu sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewPollingManager creates a new polling manager.
func NewPollingManager[T any]() *PollingManager[T] {
	ctx, cancel := context.WithCancel(context.Background())
	return &PollingManager[T]{
		pollers: make(map[string]*pollerState[T]),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// RunPoller starts a new independent poller with the given configuration.
// It returns immediately after spawning the poller goroutine.
// The poller runs until StopPoller is called or it signals to stop.
//
// Parameters:
//   - pollerID: unique identifier for this poller
//   - poller: the Poller implementation
//   - config: configuration for this specific poller
//   - callbacks: callbacks for this specific poller (optional)
//
// Returns:
//   - resultCh: channel to receive results (nil if ResultChannelSize is 0)
//   - error: error if pollerID already exists or invalid config
func (pm *PollingManager[T]) RunPoller(
	pollerID string,
	poller Poller[T],
	config PollingConfig,
	callbacks PollingCallbacks[T],
) (<-chan []T, error) {
	if pollerID == "" {
		return nil, errors.New("pollerID cannot be empty")
	}

	// Set config defaults
	if config.Interval <= 0 {
		config.Interval = time.Second
	}
	if config.BackoffFunc == nil {
		config.BackoffFunc = defaultBackoff
	}

	pm.pollersMu.Lock()
	defer pm.pollersMu.Unlock()

	// Check if already exists
	if _, exists := pm.pollers[pollerID]; exists {
		return nil, fmt.Errorf("poller %q already exists", pollerID)
	}

	// Create poller state
	ctx, cancel := context.WithCancel(pm.ctx)
	state := &pollerState[T]{
		poller:    poller,
		config:    config,
		callbacks: callbacks,
		ctx:       ctx,
		cancel:    cancel,
		done:      make(chan struct{}),
		startTime: time.Now(),
	}

	// Initialize result channel if configured
	if config.ResultChannelSize > 0 {
		state.resultCh = make(chan []T, config.ResultChannelSize)
	}

	// Register poller
	pm.pollers[pollerID] = state

	// Start poller goroutine
	go pm.runPoller(pollerID, state)

	logger.Infof("started poller: %s with interval: %v", pollerID, config.Interval)
	return state.resultCh, nil
}

// StopPoller stops a specific poller by its ID.
// It cancels the poller's context and waits for it to finish.
func (pm *PollingManager[T]) StopPoller(ctx context.Context, pollerID string) error {
	pm.pollersMu.RLock()
	state, exists := pm.pollers[pollerID]
	pm.pollersMu.RUnlock()

	if !exists {
		return fmt.Errorf("poller %q not found", pollerID)
	}

	// Cancel the poller's context
	state.cancel()

	// Wait for completion with timeout
	select {
	case <-state.done:
		// Poller has cleaned up itself (removed from registry and closed channel in runPoller)
		logger.Infof("poller %q stopped gracefully", pollerID)
		return nil
	case <-ctx.Done():
		logger.Warnf("poller %q stop timed out", pollerID)

		// Force remove from registry even on timeout
		pm.pollersMu.Lock()
		delete(pm.pollers, pollerID)
		pm.pollersMu.Unlock()

		return ctx.Err()
	}
}

// StopAll stops all active pollers.
// It attempts to stop them gracefully within the context timeout.
func (pm *PollingManager[T]) StopAll(ctx context.Context) error {
	pm.pollersMu.RLock()
	pollerIDs := make([]string, 0, len(pm.pollers))
	for id := range pm.pollers {
		pollerIDs = append(pollerIDs, id)
	}
	pm.pollersMu.RUnlock()

	var errs []error
	for _, id := range pollerIDs {
		if err := pm.StopPoller(ctx, id); err != nil {
			errs = append(errs, fmt.Errorf("poller %q: %w", id, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors stopping pollers: %v", errs)
	}

	return nil
}

// Close stops all pollers and cleans up the manager.
// This should be called when the manager is no longer needed.
func (pm *PollingManager[T]) Close(ctx context.Context) error {
	pm.cancel()
	return pm.StopAll(ctx)
}

// ListPollers returns information about all active pollers.
func (pm *PollingManager[T]) ListPollers() []PollerInfo {
	pm.pollersMu.RLock()
	defer pm.pollersMu.RUnlock()

	infos := make([]PollerInfo, 0, len(pm.pollers))
	for id, state := range pm.pollers {
		infos = append(infos, PollerInfo{
			ID:        id,
			StartTime: state.startTime,
			Metrics: PollingMetrics{
				PollCount:   state.pollCount.Load(),
				ErrorCount:  state.errorCount.Load(),
				ResultCount: state.resultCount.Load(),
			},
		})
	}

	return infos
}

// GetPollerMetrics returns metrics for a specific poller.
func (pm *PollingManager[T]) GetPollerMetrics(pollerID string) (PollingMetrics, error) {
	pm.pollersMu.RLock()
	state, exists := pm.pollers[pollerID]
	pm.pollersMu.RUnlock()

	if !exists {
		return PollingMetrics{}, fmt.Errorf("poller %q not found", pollerID)
	}

	return PollingMetrics{
		PollCount:   state.pollCount.Load(),
		ErrorCount:  state.errorCount.Load(),
		ResultCount: state.resultCount.Load(),
	}, nil
}

// IsRunning checks if a poller with the given ID is currently running.
func (pm *PollingManager[T]) IsRunning(pollerID string) bool {
	pm.pollersMu.RLock()
	defer pm.pollersMu.RUnlock()

	_, exists := pm.pollers[pollerID]
	return exists
}

// runPoller is the main loop for a single poller.
func (pm *PollingManager[T]) runPoller(pollerID string, state *pollerState[T]) {
	defer close(state.done)
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("poller %q recovered from panic: %v\n%s", pollerID, r, debug.Stack())
		}
		// Clean up the poller from registry when it exits
		pm.pollersMu.Lock()
		delete(pm.pollers, pollerID)
		pm.pollersMu.Unlock()

		// Close result channel
		if state.resultCh != nil {
			close(state.resultCh)
		}
	}()

	ticker := time.NewTicker(calculateInterval(state.config))
	defer ticker.Stop()

	for {
		select {
		case <-state.ctx.Done():
			return
		case <-ticker.C:
			shouldStop := pm.pollWithRetry(state.ctx, pollerID, state)
			if shouldStop {
				logger.Infof("poller %q stopping as requested by poller", pollerID)
				return
			}
			ticker.Reset(calculateInterval(state.config))
		}
	}
}

// pollWithRetry executes a poll operation with retry logic.
func (pm *PollingManager[T]) pollWithRetry(ctx context.Context, pollerID string, state *pollerState[T]) bool {
	attempt := 0

	for {
		// Check max retries
		if state.config.MaxRetries >= 0 && attempt > state.config.MaxRetries {
			return false
		}

		// Callback: OnPollStart
		if state.callbacks.OnPollStart != nil {
			state.callbacks.OnPollStart(ctx, pollerID)
		}

		startTime := time.Now()
		metrics := PollingMetrics{
			PollCount:   state.pollCount.Load(),
			ErrorCount:  state.errorCount.Load(),
			ResultCount: state.resultCount.Load(),
		}
		results, shouldStop, err := state.poller.Poll(ctx, metrics)
		duration := time.Since(startTime)

		state.pollCount.Add(1)

		if err != nil {
			state.errorCount.Add(1)

			// Callback: OnPollError
			if state.callbacks.OnPollError != nil {
				state.callbacks.OnPollError(ctx, pollerID, err, attempt)
			}

			// Check if we should retry
			if state.config.MaxRetries < 0 || attempt < state.config.MaxRetries {
				backoff := state.config.BackoffFunc(attempt)
				if backoff <= 0 {
					return false
				}

				// Callback: OnRetry
				if state.callbacks.OnRetry != nil {
					state.callbacks.OnRetry(ctx, pollerID, attempt, backoff)
				}

				// Wait for backoff with context cancellation
				select {
				case <-time.After(backoff):
					attempt++
					continue
				case <-ctx.Done():
					return false
				}
			}

			return false
		}

		// Success
		if len(results) > 0 {
			state.resultCount.Add(int64(len(results)))

			// Callback: OnPollSuccess
			if state.callbacks.OnPollSuccess != nil {
				state.callbacks.OnPollSuccess(ctx, pollerID, results, duration)
			}

			// Send to channel if configured (non-blocking)
			if state.resultCh != nil {
				select {
				case state.resultCh <- results:
				case <-ctx.Done():
					return false
				default:
					logger.Warnf("poller %q: result channel full, dropping %d results", pollerID, len(results))
				}
			}
		} else {
			// Empty results still count as success
			if state.callbacks.OnPollSuccess != nil {
				state.callbacks.OnPollSuccess(ctx, pollerID, results, duration)
			}
		}

		return shouldStop
	}
}

// calculateInterval returns the next poll interval with optional jitter.
func calculateInterval(config PollingConfig) time.Duration {
	interval := config.Interval
	if config.Jitter > 0 {
		jitter := time.Duration(rand.Int63n(int64(config.Jitter)))
		interval += jitter
	}
	return interval
}

// defaultBackoff provides exponential backoff with a 5-minute cap.
func defaultBackoff(attempt int) time.Duration {
	if attempt < 0 {
		return 0
	}
	duration := time.Second << uint(attempt)
	if duration > 5*time.Minute {
		return 5 * time.Minute
	}
	return duration
}
