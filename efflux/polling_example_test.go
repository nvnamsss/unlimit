package efflux_test

import (
	"context"
	"fmt"
	"time"

	"github.com/nvnamsss/unlimit/efflux"
)

// Example demonstrates basic usage of the polling framework.
func ExamplePollingManager() {
	// Create a simple poller that fetches tasks
	poller := efflux.PollerFunc[string](func(ctx context.Context, metrics efflux.PollingMetrics) ([]string, bool, error) {
		// Simulate fetching tasks from an external source
		return []string{"task1", "task2"}, false, nil
	})

	// Configure the polling manager
	config := efflux.PollingConfig{
		Interval:          5 * time.Second,
		Jitter:            time.Second,
		MaxRetries:        3,
		ResultChannelSize: 10,
	}

	// Create the polling manager
	pm := efflux.NewPollingManager[string]()

	// Run the poller
	resultCh, err := pm.RunPoller("task-poller", poller, config, efflux.PollingCallbacks[string]{})
	if err != nil {
		fmt.Printf("Failed to run poller: %v\n", err)
		return
	}

	// Process results from the channel
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		for results := range resultCh {
			for _, result := range results {
				fmt.Printf("Received: %s\n", result)
			}
		}
	}()

	<-ctx.Done()
	pm.StopPoller(context.Background(), "task-poller")
}

// ExamplePollingManager_withCallbacks demonstrates callback-based result handling.
func ExamplePollingManager_withCallbacks() {
	// Create a poller
	poller := efflux.PollerFunc[int](func(ctx context.Context, metrics efflux.PollingMetrics) ([]int, bool, error) {
		return []int{1, 2, 3}, false, nil
	})

	// Define callbacks
	callbacks := efflux.PollingCallbacks[int]{
		OnPollSuccess: func(ctx context.Context, pollerID string, results []int, duration time.Duration) {
			fmt.Printf("Poller %s polled %d items in %v\n", pollerID, len(results), duration)
		},
		OnPollError: func(ctx context.Context, pollerID string, err error, attempt int) {
			fmt.Printf("Poller %s error (attempt %d): %v\n", pollerID, attempt, err)
		},
	}

	// Configure without result channel (callback-only)
	config := efflux.PollingConfig{
		Interval:          time.Second,
		ResultChannelSize: 0, // Disable channel
	}

	// Create manager with callbacks
	pm := efflux.NewPollingManager[int]()

	// Run the poller
	_, err := pm.RunPoller("int-poller", poller, config, callbacks)
	if err != nil {
		fmt.Printf("Failed to run poller: %v\n", err)
		return
	}

	// Run for a short time
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	<-ctx.Done()
	pm.StopPoller(context.Background(), "int-poller")

	// Print stats
	metrics, _ := pm.GetPollerMetrics("int-poller")
	fmt.Printf("Polls: %d, Errors: %d, Results: %d\n", metrics.PollCount, metrics.ErrorCount, metrics.ResultCount)
}

// ExamplePollingManager_customBackoff demonstrates custom retry logic.
func ExamplePollingManager_customBackoff() {
	poller := efflux.PollerFunc[string](func(ctx context.Context, metrics efflux.PollingMetrics) ([]string, bool, error) {
		// Simulate occasional errors
		return nil, false, fmt.Errorf("temporary error")
	})

	// Custom backoff with linear increase
	customBackoff := func(attempt int) time.Duration {
		if attempt >= 5 {
			return 0 // Stop after 5 attempts
		}
		return time.Duration(attempt+1) * time.Second
	}

	config := efflux.PollingConfig{
		Interval:    10 * time.Second,
		MaxRetries:  5,
		BackoffFunc: customBackoff,
	}

	callbacks := efflux.PollingCallbacks[string]{
		OnRetry: func(ctx context.Context, pollerID string, attempt int, backoff time.Duration) {
			fmt.Printf("Poller %s retrying after %v (attempt %d)\n", pollerID, backoff, attempt)
		},
	}

	pm := efflux.NewPollingManager[string]()

	// Run the poller
	_, err := pm.RunPoller("retry-poller", poller, config, callbacks)
	if err != nil {
		fmt.Printf("Failed to run poller: %v\n", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	<-ctx.Done()
	pm.StopPoller(context.Background(), "retry-poller")
}
