package efflux

import (
	"context"

	"github.com/google/uuid"
)

// Worker processes tasks assigned by a manager
// Implementations define the actual task processing logic
type Worker[T any] interface {
	// Process executes the task and returns the result
	Process(ctx context.Context, task T) error
	// GetID returns the worker's unique identifier
	GetID() string
}

type BasicWorker[T any] struct {
	retryCount  int
	taskHandler TaskHandler[T]
	id          string
}

func NewBasicWorker[T any](id string, handler TaskHandler[T]) Worker[T] {
	if id == "" {
		id = uuid.New().String()
	}

	return &BasicWorker[T]{
		id:          id,
		taskHandler: handler,
	}
}

func (w *BasicWorker[T]) Process(ctx context.Context, task T) error {
	w.taskHandler(ctx, task)
	return nil
}

func (w *BasicWorker[T]) GetID() string {
	return w.id
}

type WorkerWR[T any, R any] interface {
	Process(ctx context.Context, task T) (R, error)
	GetID() string
}

type BasicWorkerWR[T any, R any] struct {
	handler TaskHandlerWR[T, R]
}

func NewBasicWorkerWR[T any, R any](handler TaskHandlerWR[T, R]) WorkerWR[T, R] {
	return &BasicWorkerWR[T, R]{
		handler: handler,
	}
}

func (w *BasicWorkerWR[T, R]) Process(ctx context.Context, task T) (R, error) {
	return w.handler(ctx, task)
}

func (w *BasicWorkerWR[T, R]) GetID() string {
	return uuid.New().String()
}
