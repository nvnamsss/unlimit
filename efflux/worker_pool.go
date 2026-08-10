package efflux

import (
	"context"
	"log"
	"runtime/debug"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/voidforge-studios/unlimit/logger"
)

type Backoff func(retry int) time.Duration

// Callbacks for worker events
type WorkerCallbacks[T Task] struct {
	OnStart    func(ctx context.Context, workerID int, task T)
	OnComplete func(ctx context.Context, workerID int, task T)
	OnFailed   func(ctx context.Context, workerID int, task T, err error)
}

// WorkerPool manages a pool of workers for processing tasks concurrently
type WorkerPool[T Task] struct {
	numWorkers int
	taskQueue  chan T
	callbacks  WorkerCallbacks[T]
	workers    []Worker[T]
	handler    TaskHandler[T]
	backoffFn  Backoff
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.RWMutex
	started    bool
	stopped    bool
}

// NewWorkerPool creates a worker pool with fixed number of workers
func NewWorkerPool[T Task](
	numWorkers int,
	queueSize int,
	handler TaskHandler[T],
	callbacks WorkerCallbacks[T],
) *WorkerPool[T] {
	workers := make([]Worker[T], numWorkers)
	for i := 0; i < numWorkers; i++ {
		workers[i] = NewBasicWorker(uuid.New().String(), handler)
	}

	return &WorkerPool[T]{
		taskQueue:  make(chan T, queueSize),
		callbacks:  callbacks,
		handler:    handler,
		backoffFn:  DefaultBackoff,
		numWorkers: numWorkers,
	}
}

// WithBackoff sets custom backoff strategy
func (wp *WorkerPool[T]) WithBackoff(backoff Backoff) *WorkerPool[T] {
	wp.backoffFn = backoff
	return wp
}

// Start spawns worker goroutines that process tasks from the queue
func (wp *WorkerPool[T]) Start(ctx context.Context) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.started || wp.stopped {
		return
	}

	wp.ctx, wp.cancel = context.WithCancel(ctx)

	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
	wp.started = true

	logger.Context(wp.ctx).Infof("started %d workers", wp.numWorkers)
}

// worker is the main worker loop
func (wp *WorkerPool[T]) worker(workerID int) {
	defer wp.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			logger.Infof("Recovered from worker %v: %v", workerID, r)
			log.Printf("panic:%s", debug.Stack())
		}
	}()

	for {
		select {
		case task, ok := <-wp.taskQueue:
			if !ok {
				logger.Context(wp.ctx).Infof("worker %d stopping: queue closed", workerID)
				return
			}

			// handle the error later
			wp.workers[workerID].Process(wp.ctx, task)
			// wp.processTask(workerID, task)

		case <-wp.ctx.Done():
			logger.Context(wp.ctx).Infof("worker %d stopping: context cancelled", workerID)
			return
		}
	}
}

// processTask executes a single task with retry logic
// func (wp *WorkerPool[T]) processTask(workerID int, task T) {
// 	ctx := wp.ctx
// 	logger.Context(ctx).Infof("worker %d processing task %s", workerID, task.GetID())

// 	// OnStart callback
// 	if wp.callbacks.OnStart != nil {
// 		wp.callbacks.OnStart(ctx, workerID, task)
// 	}

// 	// Execute task
// 	// worker pre task
// 	err := wp.handler(ctx, task)
// 	// worker post task
// 	task.SetError(err)

// 	if err != nil {
// 		logger.Context(ctx).Errorf("task %s failed: %v", task.GetID(), err)

// 		// Handle retry logic
// 		if task.GetRetries() < task.GetMaxRetries() {
// 			retryCount := task.GetRetries() + 1
// 			task.SetRetries(retryCount)
// 			backoffDuration := wp.backoffFn(retryCount)

// 			logger.Context(ctx).Infof("task %s will retry in %v (attempt %d/%d)",
// 				task.GetID(), backoffDuration, retryCount, task.GetMaxRetries())

// 			// Resubmit task after backoff delay
// 			go func() {
// 				time.Sleep(backoffDuration)
// 				select {
// 				case wp.taskQueue <- task:
// 					logger.Context(ctx).Debugf("task %s resubmitted for retry", task.GetID())
// 				case <-ctx.Done():
// 					logger.Context(ctx).Warnf("failed to resubmit task %s: context cancelled", task.GetID())
// 				}
// 			}()
// 		} else {
// 			logger.Context(ctx).Errorf("task %s exhausted all retries", task.GetID())
// 		}

// 		// OnFailed callback
// 		if wp.callbacks.OnFailed != nil {
// 			wp.callbacks.OnFailed(ctx, workerID, task, err)
// 		}
// 	} else {
// 		logger.Context(ctx).Infof("worker %d completed task %s", workerID, task.GetID())

// 		// OnComplete callback
// 		if wp.callbacks.OnComplete != nil {
// 			wp.callbacks.OnComplete(ctx, workerID, task)
// 		}
// 	}
// }

// Submit adds a task to the queue
// Returns false if the pool is stopped or context is cancelled
func (wp *WorkerPool[T]) Submit(task T) bool {
	defer func() {
		if recover() != nil {
			// Queue may be concurrently closed during shutdown.
		}
	}()

	wp.mu.RLock()
	stopped := wp.stopped
	ctx := wp.ctx
	wp.mu.RUnlock()

	if stopped {
		return false
	}

	if ctx == nil {
		// Pool not started yet, try non-blocking submit
		select {
		case wp.taskQueue <- task:
			return true
		default:
			return false
		}
	}

	select {
	case wp.taskQueue <- task:
		return true
	case <-ctx.Done():
		return false
	}
}

// TrySubmit attempts to add a task without blocking
// Returns true if submitted, false if queue is full
func (wp *WorkerPool[T]) TrySubmit(task T) bool {
	defer func() {
		if recover() != nil {
			// Queue may be concurrently closed during shutdown.
		}
	}()

	wp.mu.RLock()
	stopped := wp.stopped
	wp.mu.RUnlock()

	if stopped {
		return false
	}

	select {
	case wp.taskQueue <- task:
		return true
	default:
		return false
	}
}

// Stop gracefully shuts down the worker pool and waits for all workers to finish
func (wp *WorkerPool[T]) Stop() {
	wp.mu.Lock()
	if !wp.started || wp.stopped {
		wp.mu.Unlock()
		return
	}
	wp.stopped = true
	wp.mu.Unlock()

	close(wp.taskQueue)
	wp.wg.Wait()
}

// QueueLength returns current number of tasks in queue
func (wp *WorkerPool[T]) QueueLength() int {
	return len(wp.taskQueue)
}

// QueueCapacity returns maximum queue capacity
func (wp *WorkerPool[T]) QueueCapacity() int {
	return cap(wp.taskQueue)
}

// DefaultBackoff provides exponential backoff: 1s, 2s, 4s, 8s, 16s...
func DefaultBackoff(retry int) time.Duration {
	if retry <= 0 {
		return time.Second
	}
	duration := time.Second * time.Duration(1<<uint(retry))
	// Cap at 5 minutes to prevent excessive delays
	if duration > 5*time.Minute {
		return 5 * time.Minute
	}
	return duration
}
