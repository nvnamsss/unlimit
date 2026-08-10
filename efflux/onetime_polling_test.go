package efflux

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestOneTimePollingManager_Run_ShouldStop(t *testing.T) {
	manager := NewOneTimePollingManager[string](100*time.Millisecond, 5*time.Second)

	pollCount := 0
	poller := PollerFunc[string](func(ctx context.Context, metrics PollingMetrics) ([]string, bool, error) {
		pollCount++
		if pollCount >= 3 {
			return []string{"final"}, true, nil
		}
		return []string{"data"}, false, nil
	})

	ctx := context.Background()
	result, err := manager.Run(ctx, poller)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.StoppedByPoller {
		t.Error("expected StoppedByPoller to be true")
	}

	if result.TimedOut {
		t.Error("expected TimedOut to be false")
	}

	if result.TotalPolls != 3 {
		t.Errorf("expected 3 polls, got %d", result.TotalPolls)
	}

	if len(result.Results) != 3 {
		t.Errorf("expected 3 result batches, got %d", len(result.Results))
	}

	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got %d", len(result.Errors))
	}
}

func TestOneTimePollingManager_Run_Timeout(t *testing.T) {
	manager := NewOneTimePollingManager[string](50*time.Millisecond, 200*time.Millisecond)

	poller := PollerFunc[string](func(ctx context.Context, metrics PollingMetrics) ([]string, bool, error) {
		return []string{"data"}, false, nil
	})

	ctx := context.Background()
	result, err := manager.Run(ctx, poller)

	if err != context.DeadlineExceeded {
		t.Fatalf("expected context.DeadlineExceeded, got: %v", err)
	}

	if !result.TimedOut {
		t.Error("expected TimedOut to be true")
	}

	if result.StoppedByPoller {
		t.Error("expected StoppedByPoller to be false")
	}

	// Should have polled multiple times before timeout
	if result.TotalPolls < 2 {
		t.Errorf("expected at least 2 polls before timeout, got %d", result.TotalPolls)
	}
}

func TestOneTimePollingManager_Run_ErrorRetry(t *testing.T) {
	manager := NewOneTimePollingManager[string](100*time.Millisecond, 5*time.Second)
	manager.MaxRetries = 2

	attemptCount := 0
	pollCycle := 0
	poller := PollerFunc[string](func(ctx context.Context, metrics PollingMetrics) ([]string, bool, error) {
		attemptCount++
		// First poll cycle: all attempts fail
		if pollCycle == 0 {
			if attemptCount <= 3 { // Initial + 2 retries
				return nil, false, errors.New("poll cycle 1 error")
			}
			pollCycle++
			attemptCount = 0
		}
		// Second poll cycle: succeeds
		return []string{"success"}, true, nil
	})

	ctx := context.Background()
	result, err := manager.Run(ctx, poller)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.StoppedByPoller {
		t.Error("expected StoppedByPoller to be true")
	}

	if result.TotalPolls != 1 {
		t.Errorf("expected 1 successful poll, got %d", result.TotalPolls)
	}

	// First poll cycle should have recorded an error after max retries
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error (from first failed poll cycle), got %d", len(result.Errors))
	}
}

func TestOneTimePollingManager_Run_ContextCancellation(t *testing.T) {
	manager := NewOneTimePollingManager[string](100*time.Millisecond, 0)

	poller := PollerFunc[string](func(ctx context.Context, metrics PollingMetrics) ([]string, bool, error) {
		return []string{"data"}, false, nil
	})

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after a short delay
	go func() {
		time.Sleep(250 * time.Millisecond)
		cancel()
	}()

	result, err := manager.Run(ctx, poller)

	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}

	if result.TimedOut {
		t.Error("expected TimedOut to be false for context cancellation")
	}

	if result.TotalPolls < 1 {
		t.Errorf("expected at least 1 poll before cancellation, got %d", result.TotalPolls)
	}
}

func TestOneTimePollingManager_Run_ImmediateFirstPoll(t *testing.T) {
	manager := NewOneTimePollingManager[string](1*time.Second, 5*time.Second)

	firstPollTime := time.Time{}
	poller := PollerFunc[string](func(ctx context.Context, metrics PollingMetrics) ([]string, bool, error) {
		if firstPollTime.IsZero() {
			firstPollTime = time.Now()
		}
		return []string{"data"}, true, nil // Stop after first poll
	})

	ctx := context.Background()
	startTime := time.Now()
	result, err := manager.Run(ctx, poller)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	firstPollDelay := firstPollTime.Sub(startTime)
	if firstPollDelay > 100*time.Millisecond {
		t.Errorf("first poll should be immediate, but took %v", firstPollDelay)
	}

	if result.TotalPolls != 1 {
		t.Errorf("expected 1 poll, got %d", result.TotalPolls)
	}
}

func TestOneTimePollingManager_RunSimple(t *testing.T) {
	manager := NewOneTimePollingManager[int](50*time.Millisecond, 500*time.Millisecond)

	pollCount := 0
	poller := PollerFunc[int](func(ctx context.Context, metrics PollingMetrics) ([]int, bool, error) {
		pollCount++
		if pollCount >= 3 {
			return []int{30, 40}, true, nil
		}
		return []int{pollCount * 10}, false, nil
	})

	ctx := context.Background()
	results, err := manager.RunSimple(ctx, poller)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []int{10, 20, 30, 40}
	if len(results) != len(expected) {
		t.Fatalf("expected %d results, got %d", len(expected), len(results))
	}

	for i, v := range expected {
		if results[i] != v {
			t.Errorf("expected results[%d]=%d, got %d", i, v, results[i])
		}
	}
}

func TestOneTimePollingManager_RunFlattened(t *testing.T) {
	manager := NewOneTimePollingManager[string](50*time.Millisecond, 500*time.Millisecond)

	pollCount := 0
	poller := PollerFunc[string](func(ctx context.Context, metrics PollingMetrics) ([]string, bool, error) {
		pollCount++
		if pollCount >= 2 {
			return []string{"b"}, true, nil
		}
		return []string{"a"}, false, nil
	})

	ctx := context.Background()
	results, result, err := manager.RunFlattened(ctx, poller)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 flattened results, got %d", len(results))
	}

	if results[0] != "a" || results[1] != "b" {
		t.Errorf("expected [a, b], got %v", results)
	}

	if result.TotalPolls != 2 {
		t.Errorf("expected 2 polls, got %d", result.TotalPolls)
	}

	if !result.StoppedByPoller {
		t.Error("expected StoppedByPoller to be true")
	}
}

func TestOneTimePollingManager_NilPollerFunc(t *testing.T) {
	manager := NewOneTimePollingManager[string](100*time.Millisecond, 1*time.Second)

	ctx := context.Background()
	_, err := manager.Run(ctx, nil)

	if err == nil {
		t.Fatal("expected error for nil pollerFunc")
	}

	expectedMsg := "pollerFunc cannot be nil"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}
}
