package stability

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestLeakyBucket_New(t *testing.T) {
	lb := NewLeakyBucket(10, 5)
	if lb == nil {
		t.Fatal("NewLeakyBucket should not return nil")
	}
	if lb.Burst() != 10 {
		t.Errorf("Expected burst 10, got %d", lb.Burst())
	}
	if lb.Limit() != 5 {
		t.Errorf("Expected limit 5, got %f", lb.Limit())
	}
}

func TestLeakyBucket_Allow(t *testing.T) {
	lb := NewLeakyBucket(3, 1000) // Fast leak for test
	for i := 0; i < 3; i++ {
		if !lb.Allow() {
			t.Errorf("Allow should succeed for burst capacity")
		}
	}
	if lb.Allow() {
		t.Errorf("Allow should fail when bucket is full")
	}
}

func TestLeakyBucket_AllowN(t *testing.T) {
	lb := NewLeakyBucket(5, 1000)
	if !lb.AllowN(5) {
		t.Errorf("AllowN should succeed for burst capacity")
	}
	if lb.AllowN(1) {
		t.Errorf("AllowN should fail when bucket is full")
	}
}

func TestLeakyBucket_Wait(t *testing.T) {
	lb := NewLeakyBucket(1, 10) // 10 req/sec
	if !lb.Allow() {
		t.Fatal("First Allow should succeed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	if err := lb.Wait(ctx); err != nil {
		t.Errorf("Wait should succeed, got error: %v", err)
	}
	elapsed := time.Since(start)
	if elapsed < 90*time.Millisecond {
		t.Errorf("Wait should block for at least 100ms, got %v", elapsed)
	}
}

func TestLeakyBucket_WaitN(t *testing.T) {
	lb := NewLeakyBucket(2, 5) // 5 req/sec
	if !lb.AllowN(2) {
		t.Fatal("AllowN should succeed for burst capacity")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	start := time.Now()
	if err := lb.WaitN(ctx, 2); err != nil {
		t.Errorf("WaitN should succeed, got error: %v", err)
	}
	elapsed := time.Since(start)
	if elapsed < 390*time.Millisecond {
		t.Errorf("WaitN should block for at least 400ms, got %v", elapsed)
	}
}

func TestLeakyBucket_Reserve(t *testing.T) {
	lb := NewLeakyBucket(1, 2)
	lb.Allow() // Fill bucket
	res := lb.Reserve()
	if !res.OK() {
		t.Error("Reserve should be OK")
	}
	if res.Delay() < 490*time.Millisecond {
		t.Errorf("Reserve delay should be at least 500ms, got %v", res.Delay())
	}
}

func TestLeakyBucket_Reset(t *testing.T) {
	lb := NewLeakyBucket(2, 1000)
	lb.AllowN(2)
	lb.Reset()
	if !lb.AllowN(2) {
		t.Error("AllowN should succeed after Reset")
	}
}

func TestLeakyBucket_SetLimitAndBurst(t *testing.T) {
	lb := NewLeakyBucket(2, 1)
	lb.SetLimit(10)
	if lb.Limit() != 10 {
		t.Errorf("Expected limit 10, got %f", lb.Limit())
	}
	lb.SetBurst(5)
	if lb.Burst() != 5 {
		t.Errorf("Expected burst 5, got %d", lb.Burst())
	}
}

func TestLeakyBucket_Tokens(t *testing.T) {
	lb := NewLeakyBucket(3, 1000)
	if lb.Tokens() != 3 {
		t.Errorf("Expected 3 tokens, got %f", lb.Tokens())
	}
	lb.AllowN(2)
	if lb.Tokens() != 1 {
		t.Errorf("Expected 1 token, got %f", lb.Tokens())
	}
}

func TestLeakyBucket_ConcurrentOperations(t *testing.T) {
	lb := NewLeakyBucket(10, 1000)
	var wg sync.WaitGroup
	wg.Add(20)
	results := make([]bool, 20)
	for i := 0; i < 20; i++ {
		go func(idx int) {
			defer wg.Done()
			results[idx] = lb.Allow()
		}(i)
	}
	wg.Wait()
	allowed := 0
	for _, ok := range results {
		if ok {
			allowed++
		}
	}
	if allowed != 10 {
		t.Errorf("Expected 10 allowed, got %d", allowed)
	}
}

func TestLeakyBucket_CompareWithTokenBucket(t *testing.T) {
	// Simple comparative test: Leaky bucket should not allow burst after initial fill
	lb := NewLeakyBucket(2, 1)
	lb.AllowN(2)
	if lb.Allow() {
		t.Error("LeakyBucket should not allow burst after full")
	}
	// Simulate a token bucket (for comparison)
	tbTokens := 2
	tbRate := time.Second
	lastRefill := time.Now()
	allowTB := func() bool {
		if tbTokens > 0 {
			tbTokens--
			return true
		}
		if time.Since(lastRefill) > tbRate {
			tbTokens = 2
			lastRefill = time.Now()
			tbTokens--
			return true
		}
		return false
	}
	if !allowTB() {
		t.Error("TokenBucket should allow after refill")
	}
}
