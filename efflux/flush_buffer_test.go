package efflux

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestFlushBuffer_CapacityFlush(t *testing.T) {
	var flushCount atomic.Int32
	var totalItems atomic.Int32

	flushFunc := func(ctx context.Context, items []int) error {
		flushCount.Add(1)
		totalItems.Add(int32(len(items)))
		return nil
	}

	buffer := NewFlushBuffer[int](5, time.Minute, flushFunc)
	ctx := context.Background()
	buffer.Start(ctx)
	defer buffer.Stop()

	// Add 12 items (should trigger 2 flushes at capacity 5)
	for i := 0; i < 12; i++ {
		if err := buffer.Add(i); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// Wait a bit for flushes to complete
	time.Sleep(100 * time.Millisecond)

	if flushCount.Load() != 2 {
		t.Errorf("expected 2 flushes, got %d", flushCount.Load())
	}

	if totalItems.Load() != 10 {
		t.Errorf("expected 10 items flushed, got %d", totalItems.Load())
	}

	if buffer.Len() != 2 {
		t.Errorf("expected 2 items remaining, got %d", buffer.Len())
	}
}

func TestFlushBuffer_IntervalFlush(t *testing.T) {
	var flushCount atomic.Int32
	var totalItems atomic.Int32

	flushFunc := func(ctx context.Context, items []int) error {
		flushCount.Add(1)
		totalItems.Add(int32(len(items)))
		return nil
	}

	buffer := NewFlushBuffer[int](100, 200*time.Millisecond, flushFunc)
	ctx := context.Background()
	buffer.Start(ctx)
	defer buffer.Stop()

	// Add 5 items (below capacity)
	for i := 0; i < 5; i++ {
		if err := buffer.Add(i); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// Wait for interval flush
	time.Sleep(300 * time.Millisecond)

	if flushCount.Load() != 1 {
		t.Errorf("expected 1 flush, got %d", flushCount.Load())
	}

	if totalItems.Load() != 5 {
		t.Errorf("expected 5 items flushed, got %d", totalItems.Load())
	}

	if buffer.Len() != 0 {
		t.Errorf("expected 0 items remaining, got %d", buffer.Len())
	}
}

func TestFlushBuffer_ManualFlush(t *testing.T) {
	var flushCount atomic.Int32
	var totalItems atomic.Int32

	flushFunc := func(ctx context.Context, items []int) error {
		flushCount.Add(1)
		totalItems.Add(int32(len(items)))
		return nil
	}

	buffer := NewFlushBuffer[int](100, time.Minute, flushFunc)
	ctx := context.Background()
	buffer.Start(ctx)
	defer buffer.Stop()

	// Add items
	for i := 0; i < 7; i++ {
		if err := buffer.Add(i); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// Manual flush
	if err := buffer.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	if flushCount.Load() != 1 {
		t.Errorf("expected 1 flush, got %d", flushCount.Load())
	}

	if totalItems.Load() != 7 {
		t.Errorf("expected 7 items flushed, got %d", totalItems.Load())
	}

	if buffer.Len() != 0 {
		t.Errorf("expected 0 items remaining, got %d", buffer.Len())
	}
}

func TestFlushBuffer_AddBatch(t *testing.T) {
	var flushCount atomic.Int32
	var totalItems atomic.Int32

	flushFunc := func(ctx context.Context, items []int) error {
		flushCount.Add(1)
		totalItems.Add(int32(len(items)))
		return nil
	}

	buffer := NewFlushBuffer[int](5, time.Minute, flushFunc)
	ctx := context.Background()
	buffer.Start(ctx)
	defer buffer.Stop()

	// Add batch of 12 items (should trigger 2 flushes)
	items := make([]int, 12)
	for i := range items {
		items[i] = i
	}

	if err := buffer.AddBatch(items); err != nil {
		t.Fatalf("AddBatch failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if flushCount.Load() != 2 {
		t.Errorf("expected 2 flushes, got %d", flushCount.Load())
	}

	if totalItems.Load() != 10 {
		t.Errorf("expected 10 items flushed, got %d", totalItems.Load())
	}

	if buffer.Len() != 2 {
		t.Errorf("expected 2 items remaining, got %d", buffer.Len())
	}
}

func TestFlushBuffer_FinalFlush(t *testing.T) {
	var totalItems atomic.Int32

	flushFunc := func(ctx context.Context, items []int) error {
		totalItems.Add(int32(len(items)))
		return nil
	}

	buffer := NewFlushBuffer[int](100, time.Minute, flushFunc)
	ctx := context.Background()
	buffer.Start(ctx)

	// Add items
	for i := 0; i < 7; i++ {
		if err := buffer.Add(i); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// Stop should trigger final flush
	if err := buffer.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if totalItems.Load() != 7 {
		t.Errorf("expected 7 items flushed on stop, got %d", totalItems.Load())
	}
}

func TestFlushBuffer_ConcurrentAdd(t *testing.T) {
	var totalItems atomic.Int32

	flushFunc := func(ctx context.Context, items []int) error {
		totalItems.Add(int32(len(items)))
		return nil
	}

	buffer := NewFlushBuffer[int](50, 100*time.Millisecond, flushFunc)
	ctx := context.Background()
	buffer.Start(ctx)
	defer buffer.Stop()

	// Concurrent adds
	var wg sync.WaitGroup
	numGoroutines := 10
	itemsPerGoroutine := 20

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			for j := 0; j < itemsPerGoroutine; j++ {
				buffer.Add(start + j)
			}
		}(i * itemsPerGoroutine)
	}

	wg.Wait()
	time.Sleep(200 * time.Millisecond)

	// Flush remaining
	buffer.Flush()

	expected := int32(numGoroutines * itemsPerGoroutine)
	if totalItems.Load() != expected {
		t.Errorf("expected %d items flushed, got %d", expected, totalItems.Load())
	}
}

func TestFlushBuffer_EmptyFlush(t *testing.T) {
	var flushCount atomic.Int32

	flushFunc := func(ctx context.Context, items []int) error {
		flushCount.Add(1)
		return nil
	}

	buffer := NewFlushBuffer[int](10, time.Minute, flushFunc)
	ctx := context.Background()
	buffer.Start(ctx)
	defer buffer.Stop()

	// Flush with no items
	if err := buffer.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	if flushCount.Load() != 0 {
		t.Errorf("expected 0 flushes for empty buffer, got %d", flushCount.Load())
	}
}

func TestFlushBuffer_Capacity(t *testing.T) {
	buffer := NewFlushBuffer[int](42, time.Minute, nil)

	if buffer.Capacity() != 42 {
		t.Errorf("expected capacity 42, got %d", buffer.Capacity())
	}
}
