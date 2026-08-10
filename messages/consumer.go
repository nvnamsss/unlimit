package messages

import (
	"context"
	"errors"
)

var (
	ErrTopicRegistered = errors.New("topic has already been registered")
)

type ConsumeHandler func(ctx context.Context, message Message) error
