package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/nvnamsss/unlimit/efflux"
	"github.com/nvnamsss/unlimit/logger"
)

// SimpleTask implements the efflux.Task interface
type SimpleTask struct {
	id         string
	retries    int
	maxRetries int
	executeAt  time.Time
	err        error
	data       string
}

func (t *SimpleTask) GetID() string             { return t.id }
func (t *SimpleTask) GetRetries() int           { return t.retries }
func (t *SimpleTask) SetRetries(retries int)    { t.retries = retries }
func (t *SimpleTask) GetMaxRetries() int        { return t.maxRetries }
func (t *SimpleTask) GetExecuteAt() time.Time   { return t.executeAt }
func (t *SimpleTask) SetExecuteAt(t2 time.Time) { t.executeAt = t2 }
func (t *SimpleTask) SetError(err error)        { t.err = err }
func (t *SimpleTask) GetError() error           { return t.err }

// taskHandler simulates task processing
func taskHandler(ctx context.Context, task *SimpleTask) error {
	// Simulate work with variable duration
	workDuration := time.Duration(rand.Intn(100)) * time.Millisecond

	select {
	case <-time.After(workDuration):
		// Simulate random failures (10% failure rate)
		if rand.Float32() < 0.1 {
			return fmt.Errorf("simulated task failure for task %s", task.GetID())
		}

		// Simulate memory allocation
		_ = make([]byte, rand.Intn(1024*10)) // Allocate up to 10KB

		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// panicTaskHandler simulates a task handler that panics
func panicTaskHandler(ctx context.Context, task *SimpleTask) error {
	// Simulate work before panic
	workDuration := time.Duration(rand.Intn(50)) * time.Millisecond
	time.Sleep(workDuration)

	// Panic on specific tasks to test panic recovery
	if task.data == "panic" {
		panic(fmt.Sprintf("intentional panic for task %s", task.GetID()))
	}

	return nil
}

// testPanickedWorker creates tasks that will cause workers to panic
func testPanickedWorker(pool *efflux.WorkerPool[*SimpleTask]) {
	logger.Infof("testing panicked worker...")

	// Submit normal task
	normalTask := &SimpleTask{
		id:         "normal-task",
		maxRetries: 3,
		executeAt:  time.Now(),
		data:       "normal",
	}
	pool.Submit(normalTask)
	logger.Infof("submitted normal task")

	time.Sleep(100 * time.Millisecond)

	// Submit panic-inducing task
	panicTask := &SimpleTask{
		id:         "panic-task",
		maxRetries: 3,
		executeAt:  time.Now(),
		data:       "panic",
	}
	pool.Submit(panicTask)
	logger.Infof("submitted panic task")

	time.Sleep(200 * time.Millisecond)

	// Submit another normal task to verify worker pool still works
	normalTask2 := &SimpleTask{
		id:         "normal-task-2",
		maxRetries: 3,
		executeAt:  time.Now(),
		data:       "normal",
	}
	pool.Submit(normalTask2)
	logger.Infof("submitted second normal task to verify worker pool recovery")
}

func main() {
	// Command line flags
	workers := flag.Int("workers", 10, "number of worker goroutines")
	queueSize := flag.Int("queue", 100, "task queue size")
	taskCount := flag.Int("tasks", 1000, "number of tasks to process")
	pprofAddr := flag.String("pprof", ":6060", "pprof server address")
	duration := flag.Int("duration", 60, "benchmark duration in seconds (0 for task-based)")
	testPanic := flag.Bool("test-panic", false, "test worker panic recovery")
	flag.Parse()

	// Start pprof server for profiling
	go func() {
		logger.Infof("starting pprof server on %s", *pprofAddr)
		logger.Infof("CPU profile: curl http://localhost%s/debug/pprof/profile?seconds=30 > cpu.prof", *pprofAddr)
		logger.Infof("Heap profile: curl http://localhost%s/debug/pprof/heap > heap.prof", *pprofAddr)
		logger.Infof("Goroutine profile: curl http://localhost%s/debug/pprof/goroutine > goroutine.prof", *pprofAddr)
		if err := http.ListenAndServe(*pprofAddr, nil); err != nil {
			logger.Errorf("pprof server failed: %v", err)
		}
	}()

	// Create worker pool with callbacks
	callbacks := efflux.WorkerCallbacks[*SimpleTask]{
		OnStart: func(ctx context.Context, workerID int, task *SimpleTask) {
			logger.Debugf("worker %d started task %s", workerID, task.GetID())
		},
		OnComplete: func(ctx context.Context, workerID int, task *SimpleTask) {
			logger.Debugf("worker %d completed task %s", workerID, task.GetID())
		},
		OnFailed: func(ctx context.Context, workerID int, task *SimpleTask, err error) {
			logger.Warnf("worker %d failed task %s: %v", workerID, task.GetID(), err)
		},
	}

	// Select task handler based on test mode
	handler := taskHandler
	if *testPanic {
		handler = panicTaskHandler
		logger.Infof("panic testing mode enabled")
	}

	pool := efflux.NewWorkerPool[*SimpleTask](
		*workers,
		*queueSize,
		handler,
		callbacks,
	)

	// Start worker pool
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool.Start(ctx)
	logger.Infof("worker pool started with %d workers, queue size %d", *workers, *queueSize)

	// Print initial memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	logger.Infof("initial memory: Alloc=%v MB, TotalAlloc=%v MB, Sys=%v MB, NumGC=%v",
		m.Alloc/1024/1024, m.TotalAlloc/1024/1024, m.Sys/1024/1024, m.NumGC)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Run panic test if enabled
	if *testPanic {
		go func() {
			time.Sleep(1 * time.Second) // Wait for pool to be ready
			testPanickedWorker(pool)
		}()
	}

	// Task submission goroutine
	go func() {
		if *duration > 0 {
			// Time-based benchmark
			logger.Infof("running time-based benchmark for %d seconds", *duration)
			deadline := time.After(time.Duration(*duration) * time.Second)
			taskID := 0

			for {
				select {
				case <-deadline:
					logger.Infof("benchmark duration reached, submitted %d tasks", taskID)
					return
				case <-ctx.Done():
					return
				default:
					task := &SimpleTask{
						id:         fmt.Sprintf("task-%d", taskID),
						maxRetries: 3,
						executeAt:  time.Now(),
						data:       fmt.Sprintf("data-%d", taskID),
					}

					if pool.Submit(task) {
						taskID++
						// Small delay to avoid overwhelming the queue
						time.Sleep(time.Millisecond)
					} else {
						logger.Warnf("failed to submit task %s", task.GetID())
					}
				}
			}
		} else {
			// Task count-based benchmark
			logger.Infof("submitting %d tasks", *taskCount)
			for i := 0; i < *taskCount; i++ {
				task := &SimpleTask{
					id:         fmt.Sprintf("task-%d", i),
					maxRetries: 3,
					executeAt:  time.Now(),
					data:       fmt.Sprintf("data-%d", i),
				}

				if !pool.Submit(task) {
					logger.Warnf("failed to submit task %s", task.GetID())
				}
			}
			logger.Infof("all %d tasks submitted", *taskCount)
		}
	}()

	// Metrics reporting goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				runtime.ReadMemStats(&m)
				logger.Infof("metrics: QueueLen=%d/%d, Goroutines=%d, Memory: Alloc=%v MB, TotalAlloc=%v MB, Sys=%v MB, NumGC=%v",
					pool.QueueLength(), pool.QueueCapacity(),
					runtime.NumGoroutine(),
					m.Alloc/1024/1024, m.TotalAlloc/1024/1024, m.Sys/1024/1024, m.NumGC)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	logger.Infof("received shutdown signal, stopping worker pool...")

	// Stop the pool gracefully
	pool.Stop()

	// Print final memory stats
	runtime.ReadMemStats(&m)
	logger.Infof("final memory: Alloc=%v MB, TotalAlloc=%v MB, Sys=%v MB, NumGC=%v",
		m.Alloc/1024/1024, m.TotalAlloc/1024/1024, m.Sys/1024/1024, m.NumGC)

	logger.Infof("worker pool stopped gracefully")
}
