package efflux

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// mockJobHandler implements JobHandler for testing
type mockJobHandler struct {
	name      string
	processFn func(ctx context.Context) (int, error)
	interval  time.Duration
}

func (m *mockJobHandler) Name() string { return m.name }
func (m *mockJobHandler) Handle(ctx context.Context) (int, error) {
	if m.processFn != nil {
		return m.processFn(ctx)
	}
	return 1, nil
}
func (m *mockJobHandler) Interval() time.Duration { return m.interval }

func TestJob_New(t *testing.T) {
	job := NewJob()
	if job == nil {
		t.Fatal("NewJob() returned nil")
	}
	if len(job.handlers) != 0 {
		t.Errorf("New job should have no handlers, got %d", len(job.handlers))
	}
}

func TestJob_AddHandler(t *testing.T) {
	job := NewJob()
	handler := &mockJobHandler{name: "test", interval: time.Millisecond * 10}
	cfg := &HandlerConfig{Handler: handler}
	job.AddHandler(cfg)
	if len(job.handlers) != 1 {
		t.Errorf("Expected 1 handler, got %d", len(job.handlers))
	}
	if job.handlers[0].Handler.Name() != "test" {
		t.Errorf("Handler name mismatch, got %s", job.handlers[0].Handler.Name())
	}
}

func TestJob_StartAndCancel(t *testing.T) {
	var called int32
	job := NewJob()
	k := 10
	interval := time.Millisecond * 20
	handler := &mockJobHandler{
		name:     "periodic",
		interval: interval,
		processFn: func(ctx context.Context) (int, error) {
			atomic.AddInt32(&called, 1)
			return 1, nil
		},
	}
	cfg := &HandlerConfig{Handler: handler}
	job.AddHandler(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := job.Start(ctx)
	if err != nil {
		t.Fatalf("Job.Start() returned error: %v", err)
	}

	// Let the handler run a few times
	time.Sleep(interval * time.Duration(k))
	job.cancel()
	job.wg.Wait()

	if atomic.LoadInt32(&called) == 0 {
		t.Error("Handler was not called")
	}

	// measure if the handler was called approximately k times by checking the count in range [k-1, k+1]
	if atomic.LoadInt32(&called) < int32(k-1) || atomic.LoadInt32(&called) > int32(k+1) {
		t.Errorf("Handler was called %d times, expected between %d and %d", atomic.LoadInt32(&called), k-1, k+1)
	}
}

func TestJob_HandlerError(t *testing.T) {
	var called int32
	job := NewJob()
	handler := &mockJobHandler{
		name:     "error",
		interval: time.Millisecond * 10,
		processFn: func(ctx context.Context) (int, error) {
			atomic.AddInt32(&called, 1)
			return 0, errors.New("handler error")
		},
	}
	cfg := &HandlerConfig{Handler: handler}
	job.AddHandler(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = job.Start(ctx)
	time.Sleep(time.Millisecond * 30)
	job.cancel()
	job.wg.Wait()

	if atomic.LoadInt32(&called) == 0 {
		t.Error("Error handler was not called")
	}
}

func TestJob_ConcurrentHandlers(t *testing.T) {
	job := NewJob()
	var h1Called, h2Called int32

	handler1 := &mockJobHandler{
		name:     "h1",
		interval: time.Millisecond * 10,
		processFn: func(ctx context.Context) (int, error) {
			atomic.AddInt32(&h1Called, 1)
			return 1, nil
		},
	}
	handler2 := &mockJobHandler{
		name:     "h2",
		interval: time.Millisecond * 10,
		processFn: func(ctx context.Context) (int, error) {
			atomic.AddInt32(&h2Called, 1)
			return 1, nil
		},
	}
	job.AddHandler(&HandlerConfig{Handler: handler1})
	job.AddHandler(&HandlerConfig{Handler: handler2})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = job.Start(ctx)
	time.Sleep(time.Millisecond * 30)
	job.cancel()
	job.wg.Wait()

	if atomic.LoadInt32(&h1Called) == 0 || atomic.LoadInt32(&h2Called) == 0 {
		t.Error("Both handlers should have been called at least once")
	}
}
