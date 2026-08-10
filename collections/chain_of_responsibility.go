package collections

import "context"

type Handler[T any] interface {
	Handle(ctx context.Context, request T) (T, bool)
	SetNext(handler Handler[T])
}
