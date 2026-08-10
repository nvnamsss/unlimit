package efflux

import (
	"context"
	"fmt"
	"time"
)

// OneTimePollingManager executes polling operations until completion or timeout.
// Unlike PollingManager which manages persistent background pollers, this is designed
// for transient polling scenarios where you need to poll until a condition is met.
type OneTimePollingManager[T any] struct {
	// Interval is the duration between poll operations.
	Interval time.Duration

	// Timeout is the maximum duration for the entire polling session.
	// If zero, polling continues indefinitely until shouldStop or context cancellation.
	Timeout time.Duration

	// MaxRetries is the maximum number of retry attempts per poll cycle.
	// Set to 0 to disable retries (fail immediately).
	// Set to -1 for unlimited retries (not recommended).
	MaxRetries int

	// BackoffFunc determines the wait duration before retrying after errors.
	// If nil, uses exponential backoff with a 5-minute cap.
	BackoffFunc func(attempt int) time.Duration
}

// OneTimePollingResult contains the results and metadata from a one-time polling session.
type OneTimePollingResult[T any] struct {
	// Results contains all batches of results collected during polling.
	Results [][]T

	// TotalPolls is the total number of successful poll operations.
	TotalPolls int

	// Errors contains all errors encountered during polling (after retries).
	Errors []error

	// Duration is the total time spent polling.
	Duration time.Duration

	// StoppedByPoller indicates the polling stopped because the poller returned shouldStop=true.
	StoppedByPoller bool

	// TimedOut indicates the polling stopped because the timeout was reached.
	TimedOut bool
}

// NewOneTimePollingManager creates a new one-time polling manager with the given configuration.
func NewOneTimePollingManager[T any](interval time.Duration, timeout time.Duration) *OneTimePollingManager[T] {
	return &OneTimePollingManager[T]{
		Interval:    interval,
		Timeout:     timeout,
		MaxRetries:  3,
		BackoffFunc: defaultBackoff,
	}
}

// Run executes polling operations using the provided poller function.
// It polls immediately on the first call, then at the configured interval until:
// - The poller returns shouldStop=true
// - The timeout is reached (if configured)
// - The context is cancelled
//
// The poller function receives current metrics and returns results, a stop signal, and an error.
func (opm *OneTimePollingManager[T]) Run(ctx context.Context, pollerFunc PollerFunc[T]) (*OneTimePollingResult[T], error) {
	if pollerFunc == nil {
		return nil, fmt.Errorf("pollerFunc cannot be nil")
	}

	result := &OneTimePollingResult[T]{
		Results: make([][]T, 0),
		Errors:  make([]error, 0),
	}

	startTime := time.Now()
	defer func() {
		result.Duration = time.Since(startTime)
	}()

	// Create timeout context if configured
	pollingCtx := ctx
	if opm.Timeout > 0 {
		var cancel context.CancelFunc
		pollingCtx, cancel = context.WithTimeout(ctx, opm.Timeout)
		defer cancel()
	}

	metrics := PollingMetrics{}

	// First poll happens immediately
	results, shouldStop, err := opm.executePoll(pollingCtx, pollerFunc, metrics)
	if err != nil {
		result.Errors = append(result.Errors, err)
	} else {
		result.TotalPolls++
		metrics.PollCount++
		if len(results) > 0 {
			result.Results = append(result.Results, results)
			metrics.ResultCount += int64(len(results))
		}
	}

	// Check if we should stop after first poll
	if shouldStop {
		result.StoppedByPoller = true
		return result, nil
	}

	// Continue polling at intervals
	ticker := time.NewTicker(opm.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-pollingCtx.Done():
			// Check if it's timeout vs context cancellation
			if ctx.Err() == nil && opm.Timeout > 0 {
				result.TimedOut = true
			}
			return result, pollingCtx.Err()

		case <-ticker.C:
			results, shouldStop, err := opm.executePoll(pollingCtx, pollerFunc, metrics)
			if err != nil {
				result.Errors = append(result.Errors, err)
				metrics.ErrorCount++
			} else {
				result.TotalPolls++
				metrics.PollCount++
				if len(results) > 0 {
					result.Results = append(result.Results, results)
					metrics.ResultCount += int64(len(results))
				}
			}

			if shouldStop {
				result.StoppedByPoller = true
				return result, nil
			}
		}
	}
}

// executePoll executes a single poll operation with retry logic.
func (opm *OneTimePollingManager[T]) executePoll(
	ctx context.Context,
	pollerFunc PollerFunc[T],
	metrics PollingMetrics,
) ([]T, bool, error) {
	attempt := 0

	for {
		// Check max retries
		if opm.MaxRetries >= 0 && attempt > opm.MaxRetries {
			return nil, false, fmt.Errorf("max retries (%d) exceeded", opm.MaxRetries)
		}

		results, shouldStop, err := pollerFunc(ctx, metrics)

		if err != nil {
			// Check if we should retry
			if opm.MaxRetries < 0 || attempt < opm.MaxRetries {
				backoff := opm.BackoffFunc(attempt)
				if backoff <= 0 {
					return nil, false, err
				}

				// Wait for backoff with context cancellation
				select {
				case <-time.After(backoff):
					attempt++
					continue
				case <-ctx.Done():
					return nil, false, ctx.Err()
				}
			}

			return nil, false, err
		}

		// Success
		return results, shouldStop, nil
	}
}

// RunSimple is a convenience method that returns only the results, discarding metadata.
// It flattens all result batches into a single slice.
func (opm *OneTimePollingManager[T]) RunSimple(ctx context.Context, pollerFunc PollerFunc[T]) ([]T, error) {
	result, err := opm.Run(ctx, pollerFunc)
	if err != nil && err != context.DeadlineExceeded && err != context.Canceled {
		return nil, err
	}

	// Flatten all batches into a single slice
	var allResults []T
	for _, batch := range result.Results {
		allResults = append(allResults, batch...)
	}

	return allResults, nil
}

// RunFlattened runs polling and returns a flattened slice of all results.
// It returns the result object for metadata and any error encountered.
func (opm *OneTimePollingManager[T]) RunFlattened(ctx context.Context, pollerFunc PollerFunc[T]) ([]T, *OneTimePollingResult[T], error) {
	result, err := opm.Run(ctx, pollerFunc)
	if err != nil && err != context.DeadlineExceeded && err != context.Canceled {
		return nil, result, err
	}

	// Flatten all batches into a single slice
	var allResults []T
	for _, batch := range result.Results {
		allResults = append(allResults, batch...)
	}

	return allResults, result, nil
}
