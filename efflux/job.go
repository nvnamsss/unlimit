package efflux

import (
	"context"
	"sync"
	"time"

	"github.com/nvnamsss/unlimit/logger"
)

type JobHandler interface {
	Name() string
	Handle(ctx context.Context) (processed int, err error)
	Interval() time.Duration
}

type HandlerConfig struct {
	Handler JobHandler
}

// Job manages a set of periodic job handlers and coordinates their execution.
// It allows adding multiple JobHandler instances, each with its own interval,
// and runs them concurrently using goroutines. The Job type provides methods
// to start all handlers, stop them via context cancellation, and wait for their
// completion. This is useful for scheduling and managing background tasks
// that need to run at regular intervals.
//
// Example usage:
//
//	job := NewJob()
//	job.RegisterHandler(myHandler, time.Second*10)
//	job.Start(context.Background())
//	// ... later ...
//	job.Stop() // to stop all handlers
type Job struct {
	handlers []*HandlerConfig
	wg       sync.WaitGroup
	cancel   context.CancelFunc
}

// AddHandler is deprecated. Use RegisterHandler instead.
// Deprecated: Use RegisterHandler to avoid duplicate handler registration.
// Adds a handler configuration to the job for periodic execution.
func (j *Job) AddHandler(cfg *HandlerConfig) {
	if cfg == nil {
		return
	}

	j.handlers = append(j.handlers, cfg)
}

// RegisterHandler adds a new handler configuration to the job.
// Ensures that the handler is not already registered by name.
func (j *Job) RegisterHandler(cfg *HandlerConfig) {
	if cfg == nil {
		return
	}

	// Ensure the handler is not already registered
	for _, existing := range j.handlers {
		if existing.Handler.Name() == cfg.Handler.Name() {
			logger.Warnf("Handler %s is already registered", cfg.Handler.Name())
			return
		}
	}

	// Add the new handler
	j.handlers = append(j.handlers, cfg)
}

// Start launches all registered job handlers as goroutines.
// Each handler runs periodically at its configured interval.
// Returns immediately after starting all handlers.
func (j *Job) Start(ctx context.Context) error {
	ctx, j.cancel = context.WithCancel(ctx)

	for _, cfg := range j.handlers {
		j.wg.Add(1)
		go func() {
			defer func() {
				if errRecover := recover(); errRecover != nil {
					logger.Errorf("Recovered from panic in job handler %s: %v", cfg.Handler.Name(), errRecover)
				}
			}()

			j.runHandler(ctx, cfg)
		}()
	}

	return nil
}

// Stop cancels all running job handlers and waits for their completion.
// This method should be called to gracefully shut down all handlers.
func (j *Job) Stop(ctx context.Context) error {
	if j.cancel != nil {
		j.cancel()
	}
	j.wg.Wait()
	return nil
}

// runHandler executes a single handler periodically according to its interval.
// Handles context cancellation and logs processing results and errors.
func (j *Job) runHandler(ctx context.Context, cfg *HandlerConfig) {
	defer j.wg.Done()

	ticker := time.NewTicker(cfg.Handler.Interval())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			processed, err := cfg.Handler.Handle(ctx)
			if err != nil {
				logger.Errorf("Error handling job: %v", err)
			}
			logger.Infof("Processed %d items in handler %v", processed, cfg.Handler.Name())
		}
	}
}

// NewJob creates and returns a new Job instance with no handlers.
// Use RegisterHandler to add handlers before starting the job.
func NewJob() *Job {
	return &Job{
		handlers: make([]*HandlerConfig, 0),
	}
}
