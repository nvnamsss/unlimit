package trigger

// Action represents an action to be executed when a trigger fires
type Action func(event Event) error

// TypedAction is a type-safe action function that works with typed data
type TypedAction[T any] func(data T) error
