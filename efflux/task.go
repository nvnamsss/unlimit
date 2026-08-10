package efflux

import (
	"context"
	"time"
)

// TaskHandler defines a generic function that processes a job of any type
type TaskHandler[T any] func(ctx context.Context, task T) error
type TaskHandlerWR[T any, R any] func(ctx context.Context, task T) (R, error)

// Task represents a generic task interface that all task types must implement
type Task interface {
	GetID() string
	GetRetries() int
	SetRetries(retries int)
	GetMaxRetries() int
	GetExecuteAt() time.Time
	SetExecuteAt(t time.Time)
	SetError(err error)
	GetError() error
}

type ShardedTask interface {
	GetID() string
}
