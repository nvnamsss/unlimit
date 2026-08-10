package efflux

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestTask implements the Task interface for testing
type TestTask struct {
	id         string
	retries    int
	maxRetries int
	executeAt  time.Time
	err        error
	mu         sync.Mutex
}

func (t *TestTask) GetID() string { return t.id }

func (t *TestTask) GetRetries() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.retries
}

func (t *TestTask) SetRetries(retries int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.retries = retries
}

func (t *TestTask) GetMaxRetries() int { return t.maxRetries }

func (t *TestTask) GetExecuteAt() time.Time {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.executeAt
}

func (t *TestTask) SetExecuteAt(executeAt time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.executeAt = executeAt
}

func (t *TestTask) GetError() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.err
}

func (t *TestTask) SetError(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.err = err
}

// TestWorkerPool_New verifies WorkerPool creation and initialization
func TestWorkerPool_New(t *testing.T) {
	handler := func(ctx context.Context, task *TestTask) error {
		return nil
	}
	callbacks := WorkerCallbacks[*TestTask]{}

	t.Run("ValidConfiguration", func(t *testing.T) {
		pool := NewWorkerPool[*TestTask](3, 10, handler, callbacks)

		if pool == nil {
			t.Fatal("expected non-nil WorkerPool")
		}
		if pool.numWorkers != 3 {
			t.Errorf("expected 3 workers, got %d", pool.numWorkers)
		}
		if cap(pool.taskQueue) != 10 {
			t.Errorf("expected queue capacity 10, got %d", cap(pool.taskQueue))
		}
	})

	t.Run("WithBackoff", func(t *testing.T) {
		customBackoff := func(retry int) time.Duration {
			return time.Millisecond * 100
		}

		pool := NewWorkerPool[*TestTask](3, 10, handler, callbacks).WithBackoff(customBackoff)

		if pool.backoffFn == nil {
			t.Error("expected custom backoff function to be set")
		}
	})
}

// TestWorkerPool_Start verifies worker pool startup
func TestWorkerPool_Start(t *testing.T) {
	handler := func(ctx context.Context, task *TestTask) error {
		return nil
	}
	callbacks := WorkerCallbacks[*TestTask]{}

	t.Run("StartsWorkers", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		pool := NewWorkerPool[*TestTask](3, 10, handler, callbacks)
		pool.Start(ctx)

		// Give workers time to start
		time.Sleep(50 * time.Millisecond)

		if pool.ctx == nil {
			t.Error("expected context to be set")
		}

		pool.Stop()
	})
}

// TestWorkerPool_Submit verifies task submission
func TestWorkerPool_Submit(t *testing.T) {
	processed := make(chan string, 10)
	handler := func(ctx context.Context, task *TestTask) error {
		processed <- task.GetID()
		return nil
	}
	callbacks := WorkerCallbacks[*TestTask]{}

	t.Run("SubmitTaskSuccess", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		pool := NewWorkerPool[*TestTask](2, 10, handler, callbacks)
		pool.Start(ctx)

		task := &TestTask{id: "task-1", maxRetries: 3}
		if !pool.Submit(task) {
			t.Error("expected Submit to return true")
		}

		select {
		case id := <-processed:
			if id != "task-1" {
				t.Errorf("expected task-1, got %s", id)
			}
		case <-time.After(500 * time.Millisecond):
			t.Error("task was not processed")
		}

		pool.Stop()
	})

	t.Run("SubmitAfterStop", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		pool := NewWorkerPool[*TestTask](2, 10, handler, callbacks)
		pool.Start(ctx)
		pool.Stop()

		task := &TestTask{id: "task-2", maxRetries: 3}
		if pool.Submit(task) {
			t.Error("expected Submit to return false after Stop")
		}
	})

	t.Run("SubmitBeforeStart", func(t *testing.T) {
		pool := NewWorkerPool[*TestTask](2, 10, handler, callbacks)

		task := &TestTask{id: "task-3", maxRetries: 3}
		// Should handle submission before Start
		pool.Submit(task)
	})
}

// TestWorkerPool_TrySubmit verifies non-blocking task submission
func TestWorkerPool_TrySubmit(t *testing.T) {
	blockChan := make(chan struct{})
	handler := func(ctx context.Context, task *TestTask) error {
		<-blockChan // Block until released
		return nil
	}
	callbacks := WorkerCallbacks[*TestTask]{}

	t.Run("TrySubmitWhenFull", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Small queue that fills quickly
		pool := NewWorkerPool[*TestTask](1, 2, handler, callbacks)
		pool.Start(ctx)

		// Fill the queue
		for i := 0; i < 3; i++ {
			task := &TestTask{id: fmt.Sprintf("task-%d", i), maxRetries: 3}
			pool.TrySubmit(task)
		}

		// Try to submit when full
		task := &TestTask{id: "overflow", maxRetries: 3}
		if pool.TrySubmit(task) {
			t.Error("expected TrySubmit to return false when queue is full")
		}

		close(blockChan)
		pool.Stop()
	})
}

// TestWorkerPool_TaskCompletion verifies successful task execution
func TestWorkerPool_TaskCompletion(t *testing.T) {
	var startCalled, completeCalled atomic.Int32
	handler := func(ctx context.Context, task *TestTask) error {
		return nil
	}
	callbacks := WorkerCallbacks[*TestTask]{
		OnStart: func(ctx context.Context, workerID int, task *TestTask) {
			startCalled.Add(1)
		},
		OnComplete: func(ctx context.Context, workerID int, task *TestTask) {
			completeCalled.Add(1)
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	pool := NewWorkerPool[*TestTask](2, 10, handler, callbacks)
	pool.Start(ctx)

	task := &TestTask{id: "task-1", maxRetries: 3}
	pool.Submit(task)

	time.Sleep(100 * time.Millisecond)
	pool.Stop()

	if startCalled.Load() != 1 {
		t.Errorf("expected OnStart to be called 1 time, got %d", startCalled.Load())
	}
	if completeCalled.Load() != 1 {
		t.Errorf("expected OnComplete to be called 1 time, got %d", completeCalled.Load())
	}
}

// TestWorkerPool_TaskFailure verifies task error handling
func TestWorkerPool_TaskFailure(t *testing.T) {
	testErr := errors.New("test error")
	handler := func(ctx context.Context, task *TestTask) error {
		return testErr
	}

	var failedCalled atomic.Int32
	callbacks := WorkerCallbacks[*TestTask]{
		OnFailed: func(ctx context.Context, workerID int, task *TestTask, err error) {
			if !errors.Is(err, testErr) {
				t.Errorf("expected test error, got %v", err)
			}
			failedCalled.Add(1)
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool := NewWorkerPool[*TestTask](2, 10, handler, callbacks)
	pool.Start(ctx)

	task := &TestTask{id: "task-fail", maxRetries: 0}
	pool.Submit(task)

	time.Sleep(100 * time.Millisecond)
	pool.Stop()

	if failedCalled.Load() != 1 {
		t.Errorf("expected OnFailed to be called 1 time, got %d", failedCalled.Load())
	}
}

// TestWorkerPool_Retry verifies retry mechanism
func TestWorkerPool_Retry(t *testing.T) {
	var attempts atomic.Int32
	handler := func(ctx context.Context, task *TestTask) error {
		attempts.Add(1)
		return errors.New("temporary error")
	}

	callbacks := WorkerCallbacks[*TestTask]{}

	// Use fast backoff for testing
	fastBackoff := func(retry int) time.Duration {
		return 10 * time.Millisecond
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool := NewWorkerPool[*TestTask](1, 10, handler, callbacks).WithBackoff(fastBackoff)
	pool.Start(ctx)

	task := &TestTask{id: "retry-task", maxRetries: 2}
	pool.Submit(task)

	// Wait for all retries
	time.Sleep(500 * time.Millisecond)
	pool.Stop()

	// Should execute: initial + 2 retries = 3 times
	if attempts.Load() != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts.Load())
	}
}

// TestWorkerPool_Concurrency verifies concurrent task processing
func TestWorkerPool_Concurrency(t *testing.T) {
	var processedTasks atomic.Int32
	var activeTasks atomic.Int32
	var maxActiveTasks atomic.Int32

	handler := func(ctx context.Context, task *TestTask) error {
		current := activeTasks.Add(1)

		// Track max concurrent tasks
		for {
			max := maxActiveTasks.Load()
			if current <= max {
				break
			}
			if maxActiveTasks.CompareAndSwap(max, current) {
				break
			}
		}

		time.Sleep(50 * time.Millisecond)
		activeTasks.Add(-1)
		processedTasks.Add(1)
		return nil
	}

	callbacks := WorkerCallbacks[*TestTask]{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	numWorkers := 5
	queueSize := 20
	pool := NewWorkerPool[*TestTask](numWorkers, queueSize, handler, callbacks)
	pool.Start(ctx)

	numTasks := 15
	for i := 0; i < numTasks; i++ {
		task := &TestTask{id: fmt.Sprintf("task-%d", i), maxRetries: 3}
		pool.Submit(task)
	}

	// Wait for all tasks
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if processedTasks.Load() == int32(numTasks) {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	pool.Stop()

	if processedTasks.Load() != int32(numTasks) {
		t.Errorf("expected %d tasks processed, got %d", numTasks, processedTasks.Load())
	}
	if maxActiveTasks.Load() > int32(numWorkers) {
		t.Errorf("max active tasks %d exceeded worker limit %d", maxActiveTasks.Load(), numWorkers)
	}
	if maxActiveTasks.Load() <= int32(numWorkers/2) {
		t.Errorf("expected multiple workers to be used concurrently, got max %d", maxActiveTasks.Load())
	}
}

// TestWorkerPool_WorkerLimit verifies worker count limit
func TestWorkerPool_WorkerLimit(t *testing.T) {
	numWorkers := 100000
	taskStarted := make(chan struct{}, numWorkers*2)
	taskComplete := make(chan struct{})

	handler := func(ctx context.Context, task *TestTask) error {
		taskStarted <- struct{}{}
		<-taskComplete
		return nil
	}

	callbacks := WorkerCallbacks[*TestTask]{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := NewWorkerPool[*TestTask](numWorkers, numWorkers, handler, callbacks)
	pool.Start(ctx)

	// Submit more tasks than workers
	for i := 0; i < numWorkers+2; i++ {
		task := &TestTask{id: fmt.Sprintf("task-%d", i), maxRetries: 3}
		pool.Submit(task)
	}

	// Verify exactly numWorkers tasks start
	for i := 0; i < numWorkers; i++ {
		select {
		case <-taskStarted:
			// Expected
		case <-time.After(time.Second):
			t.Fatalf("expected %d tasks to start, got %d", numWorkers, i)
		}
	}

	// Verify no more tasks start (all workers busy)
	select {
	case <-taskStarted:
		t.Fatal("no more tasks should start, all workers are busy")
	case <-time.After(200 * time.Millisecond):
		// Expected
	}

	// Release one task
	taskComplete <- struct{}{}

	// Verify another task starts
	select {
	case <-taskStarted:
		// Expectedls
	case <-time.After(time.Second):
		t.Fatal("expected another task to start after one completed")
	}

	close(taskComplete)
	pool.Stop()
}

// TestWorkerPool_Stop verifies graceful shutdown
func TestWorkerPool_Stop(t *testing.T) {
	var completed atomic.Int32
	handler := func(ctx context.Context, task *TestTask) error {
		time.Sleep(50 * time.Millisecond)
		completed.Add(1)
		return nil
	}

	callbacks := WorkerCallbacks[*TestTask]{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := NewWorkerPool[*TestTask](3, 10, handler, callbacks)
	pool.Start(ctx)

	// Submit tasks
	for i := 0; i < 5; i++ {
		task := &TestTask{id: fmt.Sprintf("task-%d", i), maxRetries: 3}
		pool.Submit(task)
	}

	// Stop should wait for all tasks
	pool.Stop()

	if completed.Load() != 5 {
		t.Errorf("expected 5 tasks completed, got %d", completed.Load())
	}
}

// TestWorkerPool_QueueOperations verifies queue metrics
func TestWorkerPool_QueueOperations(t *testing.T) {
	handler := func(ctx context.Context, task *TestTask) error {
		return nil
	}
	callbacks := WorkerCallbacks[*TestTask]{}

	pool := NewWorkerPool[*TestTask](2, 5, handler, callbacks)

	if pool.QueueCapacity() != 5 {
		t.Errorf("expected queue capacity 5, got %d", pool.QueueCapacity())
	}

	task := &TestTask{id: "task-1", maxRetries: 3}
	pool.Submit(task)

	if pool.QueueLength() != 1 {
		t.Errorf("expected queue length 1, got %d", pool.QueueLength())
	}
}

// TestDefaultBackoff verifies exponential backoff calculation
func TestDefaultBackoff(t *testing.T) {
	tests := []struct {
		retry    int
		expected time.Duration
	}{
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{10, 5 * time.Minute}, // Capped at 5 minutes
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Retry%d", tt.retry), func(t *testing.T) {
			result := DefaultBackoff(tt.retry)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestWorkerPool_PanicRecovery verifies panic handling in workers
func TestWorkerPool_PanicRecovery(t *testing.T) {
	var panicTaskProcessed, normalTaskProcessed atomic.Bool

	handler := func(ctx context.Context, task *TestTask) error {
		if task.GetID() == "panic-task" {
			panicTaskProcessed.Store(true)
			panic("test panic")
		}
		normalTaskProcessed.Store(true)
		return nil
	}

	callbacks := WorkerCallbacks[*TestTask]{}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool := NewWorkerPool[*TestTask](2, 10, handler, callbacks)
	pool.Start(ctx)

	// Submit panic task
	pool.Submit(&TestTask{id: "panic-task", maxRetries: 0})
	time.Sleep(100 * time.Millisecond)

	// Submit normal task - should still work
	pool.Submit(&TestTask{id: "normal-task", maxRetries: 0})
	time.Sleep(100 * time.Millisecond)

	pool.Stop()

	if !panicTaskProcessed.Load() {
		t.Error("panic task should have been processed")
	}
	if !normalTaskProcessed.Load() {
		t.Error("normal task should have been processed after panic")
	}
}
