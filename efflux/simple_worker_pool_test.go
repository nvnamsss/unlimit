package efflux

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
)

func TestSimpleWorkerPool_New(t *testing.T) {
	tasks := []int{1, 2, 3}
	pool := NewSimpleWorker(tasks, 2)
	if pool == nil {
		t.Fatal("NewSimpleWorker returned nil")
	}
	if pool.concurrentWorkers != 2 {
		t.Errorf("Expected 2 workers, got %d", pool.concurrentWorkers)
	}
	if len(pool.tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(pool.tasks))
	}
}

func TestSimpleWorkerPool_Start(t *testing.T) {
	tasks := []int{1, 2, 3, 4}
	pool := NewSimpleWorker(tasks, 2)
	called := int32(0)
	pool.handler = func(ctx context.Context, task int) error {
		atomic.AddInt32(&called, 1)
		return nil
	}
	errs := pool.Start(context.Background())
	if len(errs) != 0 {
		t.Errorf("Expected no errors, got %d", len(errs))
	}
	if atomic.LoadInt32(&called) != int32(len(tasks)) {
		t.Errorf("Expected %d tasks processed, got %d", len(tasks), called)
	}
}

func TestSimpleWorkerPool_ErrorHandling(t *testing.T) {
	tasks := []int{1, 2, 3}
	pool := NewSimpleWorker(tasks, 2)
	pool.handler = func(ctx context.Context, task int) error {
		if task%2 == 0 {
			return errors.New("even error")
		}
		return nil
	}
	errs := pool.Start(context.Background())
	if len(errs) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errs))
	}
}

func TestSimpleWorkerPool_DifferentTypes(t *testing.T) {
	type myTask struct{ ID int }
	tasks := []myTask{{1}, {2}}
	pool := NewSimpleWorker(tasks, 1)
	pool.handler = func(ctx context.Context, task myTask) error {
		if task.ID == 2 {
			return errors.New("fail")
		}
		return nil
	}
	errs := pool.Start(context.Background())
	if len(errs) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errs))
	}
}

func TestSimpleWorkerPool_ConcurrentOperations(t *testing.T) {
	const numTasks = 100
	tasks := make([]int, numTasks)
	for i := range tasks {
		tasks[i] = i
	}
	pool := NewSimpleWorker(tasks, 8)
	var processed int32
	pool.handler = func(ctx context.Context, task int) error {
		atomic.AddInt32(&processed, 1)
		return nil
	}
	errs := pool.Start(context.Background())
	if len(errs) != 0 {
		t.Errorf("Expected no errors, got %d", len(errs))
	}
	if atomic.LoadInt32(&processed) != numTasks {
		t.Errorf("Expected %d tasks processed, got %d", numTasks, processed)
	}
}

func TestSimpleWorkerPool_EmptyOperations(t *testing.T) {
	pool := NewSimpleWorker([]int{}, 4)
	pool.handler = func(ctx context.Context, task int) error {
		return nil
	}
	errs := pool.Start(context.Background())
	if len(errs) != 0 {
		t.Errorf("Expected no errors for empty pool, got %d", len(errs))
	}
	if len(pool.GetResults()) != 0 {
		t.Errorf("Expected no results for empty pool, got %d", len(pool.GetResults()))
	}
}
