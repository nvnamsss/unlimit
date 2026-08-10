package efflux

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// mockPoller is a test implementation of the Poller interface.
type mockPoller struct {
	pollFunc    func(ctx context.Context, metrics PollingMetrics) ([]string, bool, error)
	pollCount   atomic.Int64
	shouldError atomic.Bool
	errorCount  atomic.Int64
}

func (m *mockPoller) Poll(ctx context.Context, metrics PollingMetrics) ([]string, bool, error) {
	m.pollCount.Add(1)

	if m.shouldError.Load() {
		m.errorCount.Add(1)
		return nil, false, errors.New("mock poll error")
	}

	if m.pollFunc != nil {
		return m.pollFunc(ctx, metrics)
	}

	return []string{"result1", "result2"}, false, nil
}

func TestNewPollingManager(t *testing.T) {
	pm := NewPollingManager[string]()

	if pm == nil {
		t.Fatal("expected non-nil polling manager")
	}
	if pm.pollers == nil {
		t.Error("expected pollers map to be initialized")
	}
	if pm.ctx == nil {
		t.Error("expected context to be initialized")
	}
}

func TestPollingManager_RunPoller(t *testing.T) {
	t.Run("starts poller successfully", func(t *testing.T) {
		pm := NewPollingManager[string]()
		poller := &mockPoller{}
		config := PollingConfig{
			Interval:          50 * time.Millisecond,
			ResultChannelSize: 10,
		}

		resultCh, err := pm.RunPoller("test-poller", poller, config, PollingCallbacks[string]{})
		if err != nil {
			t.Fatalf("failed to run poller: %v", err)
		}
		if resultCh == nil {
			t.Error("expected result channel to be non-nil")
		}

		// Let it run a bit
		time.Sleep(150 * time.Millisecond)

		// Stop the poller
		stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := pm.StopPoller(stopCtx, "test-poller"); err != nil {
			t.Errorf("failed to stop poller: %v", err)
		}

		// Verify polling happened
		if poller.pollCount.Load() == 0 {
			t.Error("expected at least one poll")
		}
	})

	t.Run("rejects duplicate poller ID", func(t *testing.T) {
		pm := NewPollingManager[string]()
		poller := &mockPoller{}
		config := PollingConfig{Interval: 100 * time.Millisecond}

		_, err1 := pm.RunPoller("test-poller", poller, config, PollingCallbacks[string]{})
		if err1 != nil {
			t.Fatalf("first RunPoller failed: %v", err1)
		}

		_, err2 := pm.RunPoller("test-poller", poller, config, PollingCallbacks[string]{})
		if err2 == nil {
			t.Error("expected error for duplicate poller ID")
		}

		stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		pm.StopPoller(stopCtx, "test-poller")
	})

	t.Run("rejects empty poller ID", func(t *testing.T) {
		pm := NewPollingManager[string]()
		poller := &mockPoller{}
		config := PollingConfig{Interval: 100 * time.Millisecond}

		_, err := pm.RunPoller("", poller, config, PollingCallbacks[string]{})
		if err == nil {
			t.Error("expected error for empty poller ID")
		}
	})

	t.Run("applies default config", func(t *testing.T) {
		pm := NewPollingManager[string]()
		poller := &mockPoller{}
		config := PollingConfig{} // Empty config

		_, err := pm.RunPoller("test-poller", poller, config, PollingCallbacks[string]{})
		if err != nil {
			t.Fatalf("failed to run poller: %v", err)
		}

		// Verify defaults were applied
		pm.pollersMu.RLock()
		state := pm.pollers["test-poller"]
		pm.pollersMu.RUnlock()

		if state.config.Interval != time.Second {
			t.Errorf("expected default interval 1s, got %v", state.config.Interval)
		}
		if state.config.BackoffFunc == nil {
			t.Error("expected default backoff function")
		}

		stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		pm.StopPoller(stopCtx, "test-poller")
	})
}

func TestPollingManager_StopPoller(t *testing.T) {
	t.Run("stops poller gracefully", func(t *testing.T) {
		pm := NewPollingManager[string]()
		poller := &mockPoller{}
		config := PollingConfig{Interval: 50 * time.Millisecond}

		_, err := pm.RunPoller("test-poller", poller, config, PollingCallbacks[string]{})
		if err != nil {
			t.Fatalf("failed to run poller: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err = pm.StopPoller(stopCtx, "test-poller")
		if err != nil {
			t.Errorf("failed to stop poller: %v", err)
		}

		// Verify poller is removed
		if pm.IsRunning("test-poller") {
			t.Error("poller should not be running after stop")
		}
	})

	t.Run("returns error for non-existent poller", func(t *testing.T) {
		pm := NewPollingManager[string]()

		stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := pm.StopPoller(stopCtx, "non-existent")
		if err == nil {
			t.Error("expected error for non-existent poller")
		}
	})
}

func TestPollingManager_StopAll(t *testing.T) {
	pm := NewPollingManager[string]()
	poller1 := &mockPoller{}
	poller2 := &mockPoller{}
	config := PollingConfig{Interval: 50 * time.Millisecond}

	_, err := pm.RunPoller("poller-1", poller1, config, PollingCallbacks[string]{})
	if err != nil {
		t.Fatalf("failed to run poller-1: %v", err)
	}

	_, err = pm.RunPoller("poller-2", poller2, config, PollingCallbacks[string]{})
	if err != nil {
		t.Fatalf("failed to run poller-2: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = pm.StopAll(stopCtx)
	if err != nil {
		t.Errorf("failed to stop all pollers: %v", err)
	}

	if pm.IsRunning("poller-1") || pm.IsRunning("poller-2") {
		t.Error("all pollers should be stopped")
	}
}

func TestPollingManager_ListPollers(t *testing.T) {
	pm := NewPollingManager[string]()
	poller1 := &mockPoller{}
	poller2 := &mockPoller{}
	config := PollingConfig{Interval: 100 * time.Millisecond}

	pm.RunPoller("poller-1", poller1, config, PollingCallbacks[string]{})
	pm.RunPoller("poller-2", poller2, config, PollingCallbacks[string]{})

	infos := pm.ListPollers()
	if len(infos) != 2 {
		t.Errorf("expected 2 pollers, got %d", len(infos))
	}

	ids := make(map[string]bool)
	for _, info := range infos {
		ids[info.ID] = true
	}

	if !ids["poller-1"] || !ids["poller-2"] {
		t.Error("expected both poller IDs in list")
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	pm.StopAll(stopCtx)
}

func TestPollingManager_GetPollerMetrics(t *testing.T) {
	pm := NewPollingManager[string]()
	poller := &mockPoller{}
	config := PollingConfig{Interval: 50 * time.Millisecond}

	pm.RunPoller("test-poller", poller, config, PollingCallbacks[string]{})
	time.Sleep(150 * time.Millisecond)

	metrics, err := pm.GetPollerMetrics("test-poller")
	if err != nil {
		t.Errorf("failed to get metrics: %v", err)
	}

	if metrics.PollCount == 0 {
		t.Error("expected non-zero poll count")
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	pm.StopPoller(stopCtx, "test-poller")
}

func TestPollingManager_IsRunning(t *testing.T) {
	pm := NewPollingManager[string]()
	poller := &mockPoller{}
	config := PollingConfig{Interval: 100 * time.Millisecond}

	if pm.IsRunning("test-poller") {
		t.Error("poller should not be running initially")
	}

	pm.RunPoller("test-poller", poller, config, PollingCallbacks[string]{})

	if !pm.IsRunning("test-poller") {
		t.Error("poller should be running after RunPoller")
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	pm.StopPoller(stopCtx, "test-poller")

	if pm.IsRunning("test-poller") {
		t.Error("poller should not be running after StopPoller")
	}
}

func TestPollingManager_Callbacks(t *testing.T) {
	t.Run("calls OnPollStart", func(t *testing.T) {
		pm := NewPollingManager[string]()
		poller := &mockPoller{}
		config := PollingConfig{Interval: 50 * time.Millisecond}

		var startCalled atomic.Bool
		var calledPollerID string
		callbacks := PollingCallbacks[string]{
			OnPollStart: func(ctx context.Context, pollerID string) {
				startCalled.Store(true)
				calledPollerID = pollerID
			},
		}

		pm.RunPoller("test-poller", poller, config, callbacks)
		time.Sleep(100 * time.Millisecond)

		if !startCalled.Load() {
			t.Error("expected OnPollStart to be called")
		}
		if calledPollerID != "test-poller" {
			t.Errorf("expected pollerID 'test-poller', got %q", calledPollerID)
		}

		stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		pm.StopPoller(stopCtx, "test-poller")
	})

	t.Run("calls OnPollSuccess", func(t *testing.T) {
		pm := NewPollingManager[string]()
		poller := &mockPoller{}
		config := PollingConfig{Interval: 50 * time.Millisecond}

		var successCalled atomic.Bool
		callbacks := PollingCallbacks[string]{
			OnPollSuccess: func(ctx context.Context, pollerID string, results []string, duration time.Duration) {
				successCalled.Store(true)
			},
		}

		pm.RunPoller("test-poller", poller, config, callbacks)
		time.Sleep(100 * time.Millisecond)

		if !successCalled.Load() {
			t.Error("expected OnPollSuccess to be called")
		}

		stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		pm.StopPoller(stopCtx, "test-poller")
	})

	t.Run("calls OnPollError and OnRetry", func(t *testing.T) {
		pm := NewPollingManager[string]()
		poller := &mockPoller{}
		poller.shouldError.Store(true)
		config := PollingConfig{
			Interval:   50 * time.Millisecond,
			MaxRetries: 2,
		}

		var errorCalled atomic.Bool
		var retryCalled atomic.Bool
		callbacks := PollingCallbacks[string]{
			OnPollError: func(ctx context.Context, pollerID string, err error, attempt int) {
				errorCalled.Store(true)
			},
			OnRetry: func(ctx context.Context, pollerID string, attempt int, backoff time.Duration) {
				retryCalled.Store(true)
			},
		}

		pm.RunPoller("test-poller", poller, config, callbacks)
		time.Sleep(200 * time.Millisecond)

		if !errorCalled.Load() {
			t.Error("expected OnPollError to be called")
		}
		if !retryCalled.Load() {
			t.Error("expected OnRetry to be called")
		}

		stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		pm.StopPoller(stopCtx, "test-poller")
	})
}

func TestPollingManager_ResultChannel(t *testing.T) {
	pm := NewPollingManager[string]()
	poller := &mockPoller{}
	config := PollingConfig{
		Interval:          50 * time.Millisecond,
		ResultChannelSize: 10,
	}

	resultCh, err := pm.RunPoller("test-poller", poller, config, PollingCallbacks[string]{})
	if err != nil {
		t.Fatalf("failed to run poller: %v", err)
	}

	// Collect results
	var results []string
	timeout := time.After(200 * time.Millisecond)
	collecting := true

	for collecting {
		select {
		case res := <-resultCh:
			results = append(results, res...)
		case <-timeout:
			collecting = false
		}
	}

	if len(results) == 0 {
		t.Error("expected to receive results")
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	pm.StopPoller(stopCtx, "test-poller")
}

func TestPollingManager_ShouldStop(t *testing.T) {
	pm := NewPollingManager[string]()

	// Create poller that signals to stop after first poll
	pollCount := 0
	poller := &mockPoller{
		pollFunc: func(ctx context.Context, metrics PollingMetrics) ([]string, bool, error) {
			pollCount++
			shouldStop := pollCount >= 1
			return []string{"result"}, shouldStop, nil
		},
	}

	config := PollingConfig{Interval: 50 * time.Millisecond}

	pm.RunPoller("test-poller", poller, config, PollingCallbacks[string]{})

	// Wait for poller to stop itself (give it time to complete the stop sequence)
	time.Sleep(300 * time.Millisecond)

	// Poller should have stopped itself and been cleaned up
	if pm.IsRunning("test-poller") {
		t.Error("poller should have stopped itself")
	}
}

func TestPollingManager_MultiplePollers(t *testing.T) {
	pm := NewPollingManager[string]()

	poller1 := &mockPoller{}
	poller2 := &mockPoller{}
	poller3 := &mockPoller{}

	config := PollingConfig{Interval: 50 * time.Millisecond}

	pm.RunPoller("poller-1", poller1, config, PollingCallbacks[string]{})
	pm.RunPoller("poller-2", poller2, config, PollingCallbacks[string]{})
	pm.RunPoller("poller-3", poller3, config, PollingCallbacks[string]{})

	time.Sleep(150 * time.Millisecond)

	// All should be running
	if !pm.IsRunning("poller-1") || !pm.IsRunning("poller-2") || !pm.IsRunning("poller-3") {
		t.Error("all pollers should be running")
	}

	// All should have polled
	if poller1.pollCount.Load() == 0 || poller2.pollCount.Load() == 0 || poller3.pollCount.Load() == 0 {
		t.Error("all pollers should have executed polls")
	}

	// Stop one poller
	stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	pm.StopPoller(stopCtx, "poller-2")

	// Others should still be running
	if !pm.IsRunning("poller-1") || !pm.IsRunning("poller-3") {
		t.Error("other pollers should still be running")
	}
	if pm.IsRunning("poller-2") {
		t.Error("poller-2 should be stopped")
	}

	pm.StopAll(stopCtx)
}

func TestPollerFunc(t *testing.T) {
	t.Run("adapts function to Poller interface", func(t *testing.T) {
		called := false
		pollerFunc := PollerFunc[int](func(ctx context.Context, metrics PollingMetrics) ([]int, bool, error) {
			called = true
			return []int{1, 2, 3}, false, nil
		})

		results, shouldStop, err := pollerFunc.Poll(context.Background(), PollingMetrics{})

		if !called {
			t.Error("expected function to be called")
		}
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if shouldStop {
			t.Error("expected shouldStop to be false")
		}
		if len(results) != 3 {
			t.Errorf("expected 3 results, got %d", len(results))
		}
	})
}

func TestPollingDefaultBackoff(t *testing.T) {
	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{10, 5 * time.Minute}, // Cap at 5 minutes
		{-1, 0},               // Invalid attempt
	}

	for _, tt := range tests {
		result := defaultBackoff(tt.attempt)
		if result != tt.expected {
			t.Errorf("attempt %d: expected %v, got %v", tt.attempt, tt.expected, result)
		}
	}
}
