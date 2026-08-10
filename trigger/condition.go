package trigger

// Condition represents a condition that must be met for a trigger to execute
type Condition func(event Event) bool

// TypedCondition is a type-safe condition function
type TypedCondition[T any] func(data T) bool
