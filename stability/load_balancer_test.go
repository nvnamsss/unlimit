package stability

import (
	"context"
	"errors"
	"sync"
	"testing"
)

// TestRoundRobinLoadBalancer_New tests the constructor
func TestRoundRobinLoadBalancer_New(t *testing.T) {
	lb := NewRoundRobinLoadBalancer[string]()

	if lb == nil {
		t.Error("NewRoundRobinLoadBalancer() should return a non-nil instance")
	}

	if len(lb.targets) != 0 {
		t.Errorf("Expected empty targets map, got %d targets", len(lb.targets))
	}

	if len(lb.targetOrder) != 0 {
		t.Errorf("Expected empty target order, got %d items", len(lb.targetOrder))
	}

	if lb.currentIndex != 0 {
		t.Errorf("Expected currentIndex to be 0, got %d", lb.currentIndex)
	}
}

// TestRoundRobinLoadBalancer_AddTarget tests adding targets
func TestRoundRobinLoadBalancer_AddTarget(t *testing.T) {
	lb := NewRoundRobinLoadBalancer[string]()

	err := lb.AddTarget("target1", "data1", 1)
	if err != nil {
		t.Errorf("AddTarget() should not return error, got %v", err)
	}

	targets := lb.GetTargets()
	if len(targets) != 1 {
		t.Errorf("Expected 1 target, got %d", len(targets))
	}

	target, exists := targets["target1"]
	if !exists {
		t.Error("Target 'target1' should exist")
	}

	if target.Data != "data1" {
		t.Errorf("Expected target data 'data1', got %s", target.Data)
	}

	if !target.Available {
		t.Error("New target should be available")
	}
}

// TestRoundRobinLoadBalancer_AddTarget_Duplicate tests adding duplicate targets
func TestRoundRobinLoadBalancer_AddTarget_Duplicate(t *testing.T) {
	lb := NewRoundRobinLoadBalancer[string]()

	lb.AddTarget("target1", "data1", 1)
	err := lb.AddTarget("target1", "data2", 2)

	if err == nil {
		t.Error("AddTarget() should return an error for duplicate target")
	}
	if !errors.Is(err, ErrTargetExists) {
		t.Errorf("Expected ErrTargetExists, got %v", err)
	}
}

// TestRoundRobinLoadBalancer_RemoveTarget tests removing targets
func TestRoundRobinLoadBalancer_RemoveTarget(t *testing.T) {
	lb := NewRoundRobinLoadBalancer[string]()

	lb.AddTarget("target1", "data1", 1)
	lb.AddTarget("target2", "data2", 1)

	err := lb.RemoveTarget("target1")
	if err != nil {
		t.Errorf("RemoveTarget() should not return error, got %v", err)
	}

	targets := lb.GetTargets()
	if len(targets) != 1 {
		t.Errorf("Expected 1 target after removal, got %d", len(targets))
	}

	if _, exists := targets["target1"]; exists {
		t.Error("Target 'target1' should be removed")
	}
}

// TestRoundRobinLoadBalancer_RemoveTarget_NotFound tests removing non-existent target
func TestRoundRobinLoadBalancer_RemoveTarget_NotFound(t *testing.T) {
	lb := NewRoundRobinLoadBalancer[string]()

	err := lb.RemoveTarget("nonexistent")

	if err == nil {
		t.Error("RemoveTarget() should return an error for non-existent target")
	}
	if !errors.Is(err, ErrTargetNotFound) {
		t.Errorf("Expected ErrTargetNotFound, got %v", err)
	}
}

// TestRoundRobinLoadBalancer_SelectTarget tests target selection
func TestRoundRobinLoadBalancer_SelectTarget(t *testing.T) {
	lb := NewRoundRobinLoadBalancer[string]()

	lb.AddTarget("target1", "data1", 1)
	lb.AddTarget("target2", "data2", 1)
	lb.AddTarget("target3", "data3", 1)

	ctx := context.Background()

	// Test round-robin selection
	selections := make(map[string]int)
	for i := 0; i < 6; i++ {
		id, data, err := lb.SelectTarget(ctx)
		if err != nil {
			t.Errorf("SelectTarget() should not return error, got %v", err)
		}
		selections[id]++

		// Verify data matches ID
		expectedData := "data" + id[len(id)-1:]
		if data != expectedData {
			t.Errorf("Expected data %s, got %s", expectedData, data)
		}
	}

	// Each target should be selected twice in 6 rounds
	for targetID, count := range selections {
		if count != 2 {
			t.Errorf("Target %s selected %d times, expected 2", targetID, count)
		}
	}
}

// TestRoundRobinLoadBalancer_SelectTarget_EmptyPool tests selection from empty pool
func TestRoundRobinLoadBalancer_SelectTarget_EmptyPool(t *testing.T) {
	lb := NewRoundRobinLoadBalancer[string]()

	ctx := context.Background()
	_, _, err := lb.SelectTarget(ctx)

	if err == nil {
		t.Error("SelectTarget() should return an error for empty pool")
	}
	if !errors.Is(err, ErrNoTargetsAvailable) {
		t.Errorf("Expected ErrNoTargetsAvailable, got %v", err)
	}
}

// TestRoundRobinLoadBalancer_SetTargetAvailability tests setting target availability
func TestRoundRobinLoadBalancer_SetTargetAvailability(t *testing.T) {
	lb := NewRoundRobinLoadBalancer[string]()

	lb.AddTarget("target1", "data1", 1)

	err := lb.SetTargetAvailability("target1", false)
	if err != nil {
		t.Errorf("SetTargetAvailability() should not return error, got %v", err)
	}

	available := lb.GetAvailableTargets()
	if len(available) != 0 {
		t.Errorf("Expected 0 available targets, got %d", len(available))
	}

	// Test selecting from unavailable targets
	ctx := context.Background()
	_, _, err = lb.SelectTarget(ctx)
	if err == nil {
		t.Error("SelectTarget() should return error when no targets available")
	}
}

// TestRoundRobinLoadBalancer_SetTargetAvailability_NotFound tests setting availability for non-existent target
func TestRoundRobinLoadBalancer_SetTargetAvailability_NotFound(t *testing.T) {
	lb := NewRoundRobinLoadBalancer[string]()

	err := lb.SetTargetAvailability("nonexistent", false)

	if err == nil {
		t.Error("SetTargetAvailability() should return an error for non-existent target")
	}
	if !errors.Is(err, ErrTargetNotFound) {
		t.Errorf("Expected ErrTargetNotFound, got %v", err)
	}
}

// TestRoundRobinLoadBalancer_GetStats tests statistics retrieval
func TestRoundRobinLoadBalancer_GetStats(t *testing.T) {
	lb := NewRoundRobinLoadBalancer[string]()

	lb.AddTarget("target1", "data1", 1)
	lb.AddTarget("target2", "data2", 1)
	lb.SetTargetAvailability("target2", false)

	// Make some selections to update stats
	ctx := context.Background()
	lb.SelectTarget(ctx)
	lb.SelectTarget(ctx)

	stats := lb.GetStats()

	if stats.TotalServers != 2 {
		t.Errorf("Expected 2 total servers, got %d", stats.TotalServers)
	}

	if stats.HealthyServers != 1 {
		t.Errorf("Expected 1 healthy server, got %d", stats.HealthyServers)
	}

	if stats.TotalRequests != 2 {
		t.Errorf("Expected 2 total requests, got %d", stats.TotalRequests)
	}
}

// TestRoundRobinLoadBalancer_Reset tests resetting the load balancer
func TestRoundRobinLoadBalancer_Reset(t *testing.T) {
	lb := NewRoundRobinLoadBalancer[string]()

	lb.AddTarget("target1", "data1", 1)

	ctx := context.Background()
	lb.SelectTarget(ctx)

	lb.Reset()

	stats := lb.GetStats()
	if stats.TotalRequests != 0 {
		t.Errorf("Expected 0 total requests after reset, got %d", stats.TotalRequests)
	}

	targets := lb.GetTargets()
	target := targets["target1"]
	if target.Requests != 0 {
		t.Errorf("Expected 0 target requests after reset, got %d", target.Requests)
	}

	if !target.LastUsed.IsZero() {
		t.Error("Expected LastUsed to be zero time after reset")
	}
}

// TestRoundRobinLoadBalancer_DifferentTypes tests with different data types
func TestRoundRobinLoadBalancer_DifferentTypes(t *testing.T) {
	// Test with int
	lbInt := NewRoundRobinLoadBalancer[int]()
	lbInt.AddTarget("target1", 42, 1)

	ctx := context.Background()
	_, data, err := lbInt.SelectTarget(ctx)
	if err != nil {
		t.Errorf("SelectTarget() should not return error, got %v", err)
	}
	if data != 42 {
		t.Errorf("Expected data 42, got %d", data)
	}

	// Test with struct
	type TestStruct struct {
		Name string
		Port int
	}

	lbStruct := NewRoundRobinLoadBalancer[TestStruct]()
	testData := TestStruct{Name: "server1", Port: 8080}
	lbStruct.AddTarget("target1", testData, 1)

	_, structData, err := lbStruct.SelectTarget(ctx)
	if err != nil {
		t.Errorf("SelectTarget() should not return error, got %v", err)
	}
	if structData.Name != "server1" || structData.Port != 8080 {
		t.Errorf("Expected struct data {server1 8080}, got %+v", structData)
	}
}

// TestRoundRobinLoadBalancer_ConcurrentOperations tests concurrent operations
func TestRoundRobinLoadBalancer_ConcurrentOperations(t *testing.T) {
	lb := NewRoundRobinLoadBalancer[string]()

	// Add targets
	for i := 0; i < 5; i++ {
		lb.AddTarget(string(rune('A'+i)), "data"+string(rune('A'+i)), 1)
	}

	const numGoroutines = 10
	const selectionsPerGoroutine = 100

	var wg sync.WaitGroup
	ctx := context.Background()
	results := make(chan string, numGoroutines*selectionsPerGoroutine)

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < selectionsPerGoroutine; j++ {
				id, _, err := lb.SelectTarget(ctx)
				if err != nil {
					t.Errorf("SelectTarget() should not return error, got %v", err)
					return
				}
				results <- id
			}
		}()
	}

	wg.Wait()
	close(results)

	// Count selections
	selections := make(map[string]int)
	for id := range results {
		selections[id]++
	}

	// Verify all targets were selected
	if len(selections) != 5 {
		t.Errorf("Expected 5 different targets selected, got %d", len(selections))
	}

	// Verify reasonable distribution (within 20% of expected)
	expected := numGoroutines * selectionsPerGoroutine / 5
	tolerance := expected / 5

	for targetID, count := range selections {
		if count < expected-tolerance || count > expected+tolerance {
			t.Errorf("Target %s selected %d times, expected around %d (±%d)",
				targetID, count, expected, tolerance)
		}
	}

	// Verify stats integrity
	stats := lb.GetStats()
	expectedTotal := int64(numGoroutines * selectionsPerGoroutine)
	if stats.TotalRequests != expectedTotal {
		t.Errorf("Expected %d total requests, got %d", expectedTotal, stats.TotalRequests)
	}
}

// TestRoundRobinLoadBalancer_PartiallyAvailableTargets tests selection with some unavailable targets
func TestRoundRobinLoadBalancer_PartiallyAvailableTargets(t *testing.T) {
	lb := NewRoundRobinLoadBalancer[string]()

	lb.AddTarget("target1", "data1", 1)
	lb.AddTarget("target2", "data2", 1)
	lb.AddTarget("target3", "data3", 1)

	// Make target2 unavailable
	lb.SetTargetAvailability("target2", false)

	ctx := context.Background()
	selections := make(map[string]int)

	// Test multiple selections
	for i := 0; i < 10; i++ {
		id, _, err := lb.SelectTarget(ctx)
		if err != nil {
			t.Errorf("SelectTarget() should not return error, got %v", err)
		}
		selections[id]++
	}

	// Should only select available targets
	if len(selections) != 2 {
		t.Errorf("Expected 2 different targets selected, got %d", len(selections))
	}

	if _, exists := selections["target2"]; exists {
		t.Error("Unavailable target2 should not be selected")
	}

	// Both available targets should be selected
	if selections["target1"] == 0 || selections["target3"] == 0 {
		t.Error("Both available targets should be selected")
	}
}
