package efflux

import (
	"context"
	"sync"
)

// A simple worker pool receiving tasks and processing them concurrently
type SimpleWorkerPool[T any] struct {
	tasks             []T
	concurrentWorkers int
	results           []error
	handler           TaskHandler[T]
}

// Start processes the tasks with at most concurrentWorkers goroutines
func (s *SimpleWorkerPool[T]) Start(ctx context.Context) []error {
	// Initialize results slice
	s.results = make([]error, len(s.tasks))

	// Create a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Start worker goroutines
	// apply round-robin for assigning task
	for i := 0; i < s.concurrentWorkers; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			for i := index; i < len(s.tasks); i += s.concurrentWorkers {
				// Execute the task and store the result
				err := s.handler(ctx, s.tasks[i])
				// No locking needed, as each goroutine writes to its own index
				s.results[i] = err
			}
		}(i)
	}

	// Wait for all workers to finish
	wg.Wait()

	// filter errors
	var filtered []error
	for _, err := range s.results {
		if err != nil {
			filtered = append(filtered, err)
		}
	}

	return filtered
}

// GetResults returns the results of all tasks
func (s *SimpleWorkerPool[T]) GetResults() []error {
	return s.results
}

func NewSimpleWorker[T any](tasks []T, concurrentWorkers int) *SimpleWorkerPool[T] {
	// Ensure at least 1 worker
	if concurrentWorkers < 1 {
		concurrentWorkers = 1
	}

	// Cap workers to number of tasks
	if concurrentWorkers > len(tasks) && len(tasks) > 0 {
		concurrentWorkers = len(tasks)
	}

	return &SimpleWorkerPool[T]{
		tasks:             tasks,
		concurrentWorkers: concurrentWorkers,
	}
}
