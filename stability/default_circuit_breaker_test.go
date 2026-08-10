package stability

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// TestDefaultCircuitBreaker_New tests circuit breaker creation and initialization
func TestDefaultCircuitBreaker_New(t *testing.T) {
	config := Config{
		Name:        "test-breaker",
		MaxRequests: 3,
		Interval:    time.Minute,
		Timeout:     30 * time.Second,
	}

	cb := NewCircuitBreaker(config)
	if cb == nil {
		t.Fatal("NewCircuitBreaker should not return nil")
	}

	// Verify initial state
	if !cb.IsClosed() {
		t.Error("Circuit breaker should start in closed state")
	}
	if cb.IsOpen() || cb.IsHalfOpen() {
		t.Error("Circuit breaker should not be open or half-open initially")
	}

	// Verify configuration
	gotConfig := cb.GetConfig()
	if gotConfig.Name != config.Name {
		t.Errorf("Expected name %s, got %s", config.Name, gotConfig.Name)
	}
	if gotConfig.MaxRequests != config.MaxRequests {
		t.Errorf("Expected MaxRequests %d, got %d", config.MaxRequests, gotConfig.MaxRequests)
	}
}

// TestDefaultCircuitBreaker_NewWithDefaults tests creation with default values
func TestDefaultCircuitBreaker_NewWithDefaults(t *testing.T) {
	config := Config{Name: "test"}
	cb := NewCircuitBreaker(config)

	gotConfig := cb.GetConfig()
	if gotConfig.MaxRequests == 0 {
		t.Error("MaxRequests should have default value")
	}
	if gotConfig.Interval == 0 {
		t.Error("Interval should have default value")
	}
	if gotConfig.Timeout == 0 {
		t.Error("Timeout should have default value")
	}
	if gotConfig.ReadyToTrip == nil {
		t.Error("ReadyToTrip should have default value")
	}
	if gotConfig.IsSuccessful == nil {
		t.Error("IsSuccessful should have default value")
	}
}

// TestDefaultCircuitBreaker_Execute tests basic execution functionality
func TestDefaultCircuitBreaker_Execute(t *testing.T) {
	cb := NewCircuitBreaker(DefaultConfig())

	// Test successful execution
	result, err := cb.Execute(func() (interface{}, error) {
		return "success", nil
	})
	if err != nil {
		t.Errorf("Execute should succeed, got error: %v", err)
	}
	if result != "success" {
		t.Errorf("Expected result 'success', got %v", result)
	}

	// Test failed execution
	testErr := errors.New("test error")
	result, err = cb.Execute(func() (interface{}, error) {
		return nil, testErr
	})
	if err != testErr {
		t.Errorf("Expected error %v, got %v", testErr, err)
	}
	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}
}

// TestDefaultCircuitBreaker_ExecuteWithContext tests context-aware execution
func TestDefaultCircuitBreaker_ExecuteWithContext(t *testing.T) {
	cb := NewCircuitBreaker(DefaultConfig())

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Test successful execution with context
	result, err := cb.ExecuteWithContext(ctx, func(ctx context.Context) (interface{}, error) {
		return "context-success", nil
	})
	if err != nil {
		t.Errorf("ExecuteWithContext should succeed, got error: %v", err)
	}
	if result != "context-success" {
		t.Errorf("Expected result 'context-success', got %v", result)
	}

	// Test context cancellation
	ctx, cancel = context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = cb.ExecuteWithContext(ctx, func(ctx context.Context) (interface{}, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
			return "should not reach", nil
		}
	})
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}

// TestDefaultCircuitBreaker_Call tests simplified execution without return values
func TestDefaultCircuitBreaker_Call(t *testing.T) {
	cb := NewCircuitBreaker(DefaultConfig())

	// Test successful call
	err := cb.Call(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Call should succeed, got error: %v", err)
	}

	// Test failed call
	testErr := errors.New("call error")
	err = cb.Call(func() error {
		return testErr
	})
	if err != testErr {
		t.Errorf("Expected error %v, got %v", testErr, err)
	}
}

// TestDefaultCircuitBreaker_CallWithContext tests context-aware simplified execution
func TestDefaultCircuitBreaker_CallWithContext(t *testing.T) {
	cb := NewCircuitBreaker(DefaultConfig())

	ctx := context.Background()
	err := cb.CallWithContext(ctx, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("CallWithContext should succeed, got error: %v", err)
	}
}

// TestDefaultCircuitBreaker_StateTransitions tests state machine transitions
func TestDefaultCircuitBreaker_StateTransitions(t *testing.T) {
	config := Config{
		Name:        "state-test",
		MaxRequests: 2,
		Interval:    time.Minute,
		Timeout:     100 * time.Millisecond,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	}
	cb := NewCircuitBreaker(config)

	// Initial state should be closed
	if !cb.IsClosed() {
		t.Error("Circuit breaker should start closed")
	}

	// Cause failures to trip the circuit
	testErr := errors.New("test failure")
	for i := 0; i < 3; i++ {
		cb.Execute(func() (interface{}, error) {
			return nil, testErr
		})
	}

	// Should now be open
	if !cb.IsOpen() {
		t.Error("Circuit breaker should be open after failures")
	}

	// Verify execution is blocked
	result, err := cb.Execute(func() (interface{}, error) {
		return "should not execute", nil
	})
	if err == nil {
		t.Error("Execute should fail when circuit is open")
	}
	if !IsCircuitBreakerError(err) {
		t.Error("Error should be CircuitBreakerError")
	}
	if result != nil {
		t.Error("Result should be nil when circuit is open")
	}

	// Wait for timeout to transition to half-open
	time.Sleep(150 * time.Millisecond)

	// Should transition to half-open
	if !cb.IsHalfOpen() {
		t.Error("Circuit breaker should be half-open after timeout")
	}

	// Test successful requests in half-open state
	for i := 0; i < int(config.MaxRequests); i++ {
		_, err := cb.Execute(func() (interface{}, error) {
			return "success", nil
		})
		if err != nil {
			t.Errorf("Execute should succeed in half-open state: %v", err)
		}
	}

	// Should transition back to closed
	if !cb.IsClosed() {
		t.Error("Circuit breaker should be closed after successful requests")
	}
}

// TestDefaultCircuitBreaker_Reset tests manual reset functionality
func TestDefaultCircuitBreaker_Reset(t *testing.T) {
	config := Config{
		Name: "reset-test",
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures >= 1
		},
	}
	cb := NewCircuitBreaker(config)

	// Trip the circuit
	cb.Execute(func() (interface{}, error) {
		return nil, errors.New("failure")
	})

	if !cb.IsOpen() {
		t.Error("Circuit breaker should be open after failure")
	}

	// Reset the circuit
	cb.Reset()

	if !cb.IsClosed() {
		t.Error("Circuit breaker should be closed after reset")
	}

	// Verify counts are reset
	counts := cb.GetCounts()
	if counts.Requests != 0 || counts.TotalFailures != 0 || counts.ConsecutiveFailures != 0 {
		t.Error("Counts should be reset after Reset()")
	}
}

// TestDefaultCircuitBreaker_Trip tests manual trip functionality
func TestDefaultCircuitBreaker_Trip(t *testing.T) {
	cb := NewCircuitBreaker(DefaultConfig())

	if !cb.IsClosed() {
		t.Error("Circuit breaker should start closed")
	}

	// Manually trip the circuit
	cb.Trip()

	if !cb.IsOpen() {
		t.Error("Circuit breaker should be open after Trip()")
	}

	// Verify execution is blocked
	_, err := cb.Execute(func() (interface{}, error) {
		return "should not execute", nil
	})
	if err == nil {
		t.Error("Execute should fail when circuit is manually tripped")
	}
}

// TestDefaultCircuitBreaker_CanExecute tests execution permission checking
func TestDefaultCircuitBreaker_CanExecute(t *testing.T) {
	config := Config{
		Name: "canexecute-test",
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures >= 1
		},
	}
	cb := NewCircuitBreaker(config)

	// Should allow execution initially
	if !cb.CanExecute() {
		t.Error("CanExecute should return true when closed")
	}

	// Trip the circuit
	cb.Execute(func() (interface{}, error) {
		return nil, errors.New("failure")
	})

	// Should not allow execution when open
	if cb.CanExecute() {
		t.Error("CanExecute should return false when open")
	}
}

// TestDefaultCircuitBreaker_GetCounts tests statistics tracking
func TestDefaultCircuitBreaker_GetCounts(t *testing.T) {
	cb := NewCircuitBreaker(DefaultConfig())

	// Initial counts should be zero
	counts := cb.GetCounts()
	if counts.Requests != 0 || counts.TotalSuccesses != 0 || counts.TotalFailures != 0 {
		t.Error("Initial counts should be zero")
	}

	// Execute successful requests
	for i := 0; i < 3; i++ {
		cb.Execute(func() (interface{}, error) {
			return "success", nil
		})
	}

	counts = cb.GetCounts()
	if counts.Requests != 3 || counts.TotalSuccesses != 3 || counts.ConsecutiveSuccesses != 3 {
		t.Errorf("Expected 3 successful requests, got: %+v", counts)
	}

	// Execute failed requests
	for i := 0; i < 2; i++ {
		cb.Execute(func() (interface{}, error) {
			return nil, errors.New("failure")
		})
	}

	counts = cb.GetCounts()
	if counts.Requests != 5 || counts.TotalFailures != 2 || counts.ConsecutiveFailures != 2 {
		t.Errorf("Expected 5 total requests with 2 failures, got: %+v", counts)
	}
}

// TestDefaultCircuitBreaker_UpdateConfig tests configuration updates
func TestDefaultCircuitBreaker_UpdateConfig(t *testing.T) {
	cb := NewCircuitBreaker(DefaultConfig())

	newConfig := Config{
		Name:        "updated",
		MaxRequests: 5,
		Interval:    2 * time.Minute,
		Timeout:     2 * time.Minute,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures >= 10
		},
	}

	err := cb.UpdateConfig(newConfig)
	if err != nil {
		t.Errorf("UpdateConfig should succeed, got error: %v", err)
	}

	gotConfig := cb.GetConfig()
	if gotConfig.Name != newConfig.Name {
		t.Errorf("Expected name %s, got %s", newConfig.Name, gotConfig.Name)
	}
	if gotConfig.MaxRequests != newConfig.MaxRequests {
		t.Errorf("Expected MaxRequests %d, got %d", newConfig.MaxRequests, gotConfig.MaxRequests)
	}
}

// TestDefaultCircuitBreaker_UpdateConfigValidation tests configuration validation
func TestDefaultCircuitBreaker_UpdateConfigValidation(t *testing.T) {
	cb := NewCircuitBreaker(DefaultConfig())

	// Test invalid MaxRequests
	err := cb.UpdateConfig(Config{MaxRequests: 0})
	if err == nil {
		t.Error("UpdateConfig should fail with MaxRequests = 0")
	}

	// Test invalid Interval
	err = cb.UpdateConfig(Config{
		MaxRequests: 1,
		Interval:    0,
	})
	if err == nil {
		t.Error("UpdateConfig should fail with Interval = 0")
	}

	// Test invalid Timeout
	err = cb.UpdateConfig(Config{
		MaxRequests: 1,
		Interval:    time.Minute,
		Timeout:     0,
	})
	if err == nil {
		t.Error("UpdateConfig should fail with Timeout = 0")
	}
}

// TestDefaultCircuitBreaker_StateChangeNotification tests state change listeners
func TestDefaultCircuitBreaker_StateChangeNotification(t *testing.T) {
	var stateChanges []struct {
		from State
		to   State
	}
	var mu sync.Mutex

	config := Config{
		Name: "notification-test",
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures >= 1
		},
		OnStateChange: func(name string, from State, to State) {
			mu.Lock()
			stateChanges = append(stateChanges, struct {
				from State
				to   State
			}{from, to})
			mu.Unlock()
		},
	}

	cb := NewCircuitBreaker(config)

	// Trip the circuit
	cb.Execute(func() (interface{}, error) {
		return nil, errors.New("failure")
	})

	// Wait a bit for goroutine to execute
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	if len(stateChanges) != 1 {
		t.Errorf("Expected 1 state change, got %d", len(stateChanges))
	}
	if len(stateChanges) > 0 && stateChanges[0].from != StateClosed || stateChanges[0].to != StateOpen {
		t.Errorf("Expected Closed->Open transition, got %v->%v", stateChanges[0].from, stateChanges[0].to)
	}
	mu.Unlock()
}

// TestDefaultCircuitBreaker_Subscribe tests listener subscription
func TestDefaultCircuitBreaker_Subscribe(t *testing.T) {
	cb := NewCircuitBreaker(DefaultConfig())

	var called bool
	listener := StateChangeFunc(func(name string, from State, to State) {
		called = true
	})

	cb.Subscribe(listener)
	cb.Trip() // Should trigger state change

	// Wait a bit for goroutine to execute
	time.Sleep(10 * time.Millisecond)

	if !called {
		t.Error("Subscribed listener should have been called")
	}
}

// TestDefaultCircuitBreaker_ConcurrentOperations tests thread safety
func TestDefaultCircuitBreaker_ConcurrentOperations(t *testing.T) {
	cb := NewCircuitBreaker(DefaultConfig())

	var wg sync.WaitGroup
	const goroutines = 10
	const operationsPerGoroutine = 100

	wg.Add(goroutines)

	// Run concurrent operations
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				cb.Execute(func() (interface{}, error) {
					return "success", nil
				})
			}
		}()
	}

	wg.Wait()

	// Verify final state
	counts := cb.GetCounts()
	expectedRequests := uint64(goroutines * operationsPerGoroutine)
	if counts.Requests != expectedRequests {
		t.Errorf("Expected %d requests, got %d", expectedRequests, counts.Requests)
	}
	if counts.TotalSuccesses != expectedRequests {
		t.Errorf("Expected %d successes, got %d", expectedRequests, counts.TotalSuccesses)
	}
}

// TestDefaultCircuitBreaker_ErrorHandling tests error handling and propagation
func TestDefaultCircuitBreaker_ErrorHandling(t *testing.T) {
	cb := NewCircuitBreaker(DefaultConfig())

	// Test that errors are properly propagated
	customErr := errors.New("custom error")
	_, err := cb.Execute(func() (interface{}, error) {
		return nil, customErr
	})

	if err != customErr {
		t.Errorf("Expected custom error to be propagated, got %v", err)
	}

	// Test circuit breaker error when open
	cb.Trip()
	_, err = cb.Execute(func() (interface{}, error) {
		return "should not execute", nil
	})

	if !IsCircuitBreakerError(err) {
		t.Error("Expected CircuitBreakerError when circuit is open")
	}

	cbErr, ok := err.(*CircuitBreakerError)
	if !ok {
		t.Error("Error should be *CircuitBreakerError")
	} else {
		if cbErr.State != StateOpen {
			t.Errorf("Expected State to be Open, got %v", cbErr.State)
		}
	}
}

// TestDefaultCircuitBreaker_CompareWithManualImplementation tests behavior comparison
func TestDefaultCircuitBreaker_CompareWithManualImplementation(t *testing.T) {
	config := Config{
		Name:        "comparison-test",
		MaxRequests: 2,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures >= 2
		},
	}

	cb := NewCircuitBreaker(config)

	// Manual state tracking for comparison
	var manualState State = StateClosed
	var manualFailures uint64

	testCases := []struct {
		name        string
		shouldFail  bool
		expectAllow bool
	}{
		{"success1", false, true},
		{"success2", false, true},
		{"failure1", true, true},
		{"failure2", true, true},  // This should trip the circuit
		{"blocked", false, false}, // Should be blocked
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Update manual state
			if manualState == StateClosed && tc.shouldFail {
				manualFailures++
				if manualFailures >= 2 {
					manualState = StateOpen
				}
			}

			// Test circuit breaker
			canExecute := cb.CanExecute()
			if canExecute != tc.expectAllow {
				t.Errorf("Expected CanExecute=%v, got %v", tc.expectAllow, canExecute)
			}

			if tc.expectAllow {
				var err error
				if tc.shouldFail {
					_, err = cb.Execute(func() (interface{}, error) {
						return nil, errors.New("test failure")
					})
				} else {
					_, err = cb.Execute(func() (interface{}, error) {
						return "success", nil
					})
				}

				// Error should match expectation
				if tc.shouldFail && err == nil {
					t.Error("Expected error for failing operation")
				}
				if !tc.shouldFail && err != nil {
					t.Errorf("Unexpected error for successful operation: %v", err)
				}
			}

			// Verify state matches manual tracking
			expectedOpen := (manualState == StateOpen)
			actualOpen := cb.IsOpen()
			if expectedOpen != actualOpen {
				t.Errorf("State mismatch: expected open=%v, got open=%v", expectedOpen, actualOpen)
			}
		})
	}
}

// TestMetricsCollector_StateHistory tests metrics collection
func TestMetricsCollector_StateHistory(t *testing.T) {
	mc := NewMetricsCollector()
	if mc == nil {
		t.Fatal("NewMetricsCollector should not return nil")
	}

	// Test initial state
	history := mc.GetStateHistory()
	if len(history) != 0 {
		t.Error("Initial state history should be empty")
	}

	// Record state changes
	mc.OnStateChange("test", StateClosed, StateOpen)
	mc.OnStateChange("test", StateOpen, StateHalfOpen)

	history = mc.GetStateHistory()
	if len(history) != 2 {
		t.Errorf("Expected 2 state changes, got %d", len(history))
	}

	if history[0].From != StateClosed || history[0].To != StateOpen {
		t.Errorf("Expected Closed->Open, got %v->%v", history[0].From, history[0].To)
	}
}

// TestMetricsCollector_LatencyStats tests latency statistics
func TestMetricsCollector_LatencyStats(t *testing.T) {
	mc := NewMetricsCollector()

	// Test empty stats
	stats := mc.GetLatencyStats()
	if stats.Count != 0 {
		t.Error("Empty metrics should have Count=0")
	}

	// Record some latencies
	latencies := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
	}

	for _, lat := range latencies {
		mc.RecordLatency(lat)
	}

	stats = mc.GetLatencyStats()
	if stats.Count != 3 {
		t.Errorf("Expected Count=3, got %d", stats.Count)
	}
	if stats.Min != 10*time.Millisecond {
		t.Errorf("Expected Min=10ms, got %v", stats.Min)
	}
	if stats.Max != 30*time.Millisecond {
		t.Errorf("Expected Max=30ms, got %v", stats.Max)
	}
	if stats.Mean != 20*time.Millisecond {
		t.Errorf("Expected Mean=20ms, got %v", stats.Mean)
	}
}

// TestMetricsCollector_Reset tests metrics reset
func TestMetricsCollector_Reset(t *testing.T) {
	mc := NewMetricsCollector()

	// Add some data
	mc.OnStateChange("test", StateClosed, StateOpen)
	mc.RecordLatency(10 * time.Millisecond)

	// Verify data exists
	if len(mc.GetStateHistory()) == 0 {
		t.Error("Should have state history before reset")
	}
	if mc.GetLatencyStats().Count == 0 {
		t.Error("Should have latency stats before reset")
	}

	// Reset and verify
	mc.Reset()
	if len(mc.GetStateHistory()) != 0 {
		t.Error("State history should be empty after reset")
	}
	if mc.GetLatencyStats().Count != 0 {
		t.Error("Latency stats should be empty after reset")
	}
}
