package efflux

import (
	"context"
	"fmt"
	"time"
)

// ExampleOneTimePollingManager demonstrates basic usage of one-time polling.
func ExampleOneTimePollingManager() {
	manager := NewOneTimePollingManager[string](100*time.Millisecond, 5*time.Second)

	pollCount := 0
	poller := PollerFunc[string](func(ctx context.Context, metrics PollingMetrics) ([]string, bool, error) {
		pollCount++
		if pollCount >= 3 {
			return []string{"final"}, true, nil
		}
		return []string{fmt.Sprintf("data-%d", pollCount)}, false, nil
	})

	ctx := context.Background()
	result, err := manager.Run(ctx, poller)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Total polls: %d\n", result.TotalPolls)
	fmt.Printf("Stopped by poller: %v\n", result.StoppedByPoller)
	fmt.Printf("Results collected: %d batches\n", len(result.Results))
	// Output:
	// Total polls: 3
	// Stopped by poller: true
	// Results collected: 3 batches
}

// ExampleOneTimePollingManager_RunSimple demonstrates the simplified API.
func ExampleOneTimePollingManager_RunSimple() {
	manager := NewOneTimePollingManager[int](50*time.Millisecond, 2*time.Second)

	count := 0
	poller := PollerFunc[int](func(ctx context.Context, metrics PollingMetrics) ([]int, bool, error) {
		count++
		if count >= 3 {
			return []int{count}, true, nil
		}
		return []int{count}, false, nil
	})

	ctx := context.Background()
	results, err := manager.RunSimple(ctx, poller)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Collected %d results\n", len(results))
	// Output:
	// Collected 3 results
}

// ExampleOneTimePollingManager_timeout demonstrates timeout behavior.
func ExampleOneTimePollingManager_timeout() {
	manager := NewOneTimePollingManager[string](100*time.Millisecond, 250*time.Millisecond)

	poller := PollerFunc[string](func(ctx context.Context, metrics PollingMetrics) ([]string, bool, error) {
		// Never signals to stop, will timeout
		return []string{"data"}, false, nil
	})

	ctx := context.Background()
	result, err := manager.Run(ctx, poller)

	fmt.Printf("Timed out: %v\n", result.TimedOut)
	fmt.Printf("Error: %v\n", err)
	fmt.Printf("Polls before timeout: %d\n", result.TotalPolls)
	// Output:
	// Timed out: true
	// Error: context deadline exceeded
	// Polls before timeout: 3
}
