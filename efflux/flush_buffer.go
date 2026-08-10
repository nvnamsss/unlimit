package efflux

import (
	"context"
	"sync"
	"time"

	"github.com/voidforge-studios/unlimit/logger"
)

// FlushFunc defines a function that processes a batch of items
type FlushFunc[T any] func(ctx context.Context, items []T) error

// FlushBuffer is a thread-safe buffer that automatically flushes data
// either when capacity is reached or after a time interval
type FlushBuffer[T any] struct {
	data          []T
	capacity      int
	flushInterval time.Duration
	flushFunc     FlushFunc[T]

	mu     sync.Mutex
	timer  *time.Timer
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewFlushBuffer creates a new flush buffer
// capacity: maximum items before auto-flush
// flushInterval: time between automatic flushes
// flushFunc: function called with buffered items
func NewFlushBuffer[T any](
	capacity int,
	flushInterval time.Duration,
	flushFunc FlushFunc[T],
) *FlushBuffer[T] {
	return &FlushBuffer[T]{
		data:          make([]T, 0, capacity),
		capacity:      capacity,
		flushInterval: flushInterval,
		flushFunc:     flushFunc,
	}
}

// Start begins the interval-based flushing goroutine
func (fb *FlushBuffer[T]) Start(ctx context.Context) {
	fb.ctx, fb.cancel = context.WithCancel(ctx)
	fb.timer = time.NewTimer(fb.flushInterval)

	fb.wg.Add(1)
	go fb.flushLoop()

	logger.Context(fb.ctx).Infof("flush buffer started (capacity: %d, interval: %v)",
		fb.capacity, fb.flushInterval)
}

// flushLoop runs in background and flushes on timer
func (fb *FlushBuffer[T]) flushLoop() {
	defer fb.wg.Done()

	for {
		select {
		case <-fb.timer.C:
			fb.mu.Lock()
			if len(fb.data) > 0 {
				logger.Context(fb.ctx).Debugf("interval flush triggered (%d items)", len(fb.data))
				fb.doFlush()
			}
			fb.timer.Reset(fb.flushInterval)
			fb.mu.Unlock()

		case <-fb.ctx.Done():
			logger.Context(fb.ctx).Infof("flush buffer stopping")
			return
		}
	}
}

// Add adds an item to the buffer
// Automatically flushes if capacity is reached
func (fb *FlushBuffer[T]) Add(item T) error {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	fb.data = append(fb.data, item)

	// Flush if capacity reached
	if len(fb.data) >= fb.capacity {
		logger.Context(fb.ctx).Debugf("capacity flush triggered (%d items)", len(fb.data))
		return fb.doFlush()
	}

	return nil
}

// AddBatch adds multiple items to the buffer
// May trigger multiple flushes if total size exceeds capacity
func (fb *FlushBuffer[T]) AddBatch(items []T) error {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	for _, item := range items {
		fb.data = append(fb.data, item)

		if len(fb.data) >= fb.capacity {
			if err := fb.doFlush(); err != nil {
				return err
			}
		}
	}

	return nil
}

// Flush manually flushes all buffered data
func (fb *FlushBuffer[T]) Flush() error {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	if len(fb.data) == 0 {
		return nil
	}

	logger.Context(fb.ctx).Debugf("manual flush triggered (%d items)", len(fb.data))
	return fb.doFlush()
}

// doFlush performs the actual flush operation
// Must be called with lock held
func (fb *FlushBuffer[T]) doFlush() error {
	if len(fb.data) == 0 {
		return nil
	}

	// Copy data to process
	items := make([]T, len(fb.data))
	copy(items, fb.data)

	// Clear buffer
	fb.data = fb.data[:0]

	// Reset timer
	if fb.timer != nil {
		fb.timer.Reset(fb.flushInterval)
	}

	// Process items (release lock during flush)
	fb.mu.Unlock()
	err := fb.flushFunc(fb.ctx, items)
	fb.mu.Lock()

	if err != nil {
		logger.Context(fb.ctx).Errorf("flush failed: %v", err)
		return err
	}

	logger.Context(fb.ctx).Debugf("flushed %d items successfully", len(items))
	return nil
}

// Stop gracefully stops the buffer and flushes remaining data
func (fb *FlushBuffer[T]) Stop() error {
	if fb.cancel != nil {
		fb.cancel()
	}

	if fb.timer != nil {
		fb.timer.Stop()
	}

	fb.wg.Wait()

	// Final flush
	fb.mu.Lock()
	defer fb.mu.Unlock()

	if len(fb.data) > 0 {
		logger.Context(fb.ctx).Infof("final flush (%d items)", len(fb.data))
		return fb.doFlush()
	}

	return nil
}

// Len returns current number of buffered items
func (fb *FlushBuffer[T]) Len() int {
	fb.mu.Lock()
	defer fb.mu.Unlock()
	return len(fb.data)
}

// Capacity returns maximum buffer capacity
func (fb *FlushBuffer[T]) Capacity() int {
	return fb.capacity
}
