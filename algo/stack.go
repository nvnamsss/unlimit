package algo

import (
	"errors"
	"sync"
)

var (
	// ErrStackEmpty is returned when trying to pop from an empty stack
	ErrStackEmpty = errors.New("stack is empty")
)

// Stack interface represents basic stack operations
type Stack[T any] interface {
	// Push adds an element to the top of the stack
	Push(item T)
	// Pop removes and returns the element at the top of the stack
	Pop() (T, error)
	// Peek returns the element at the top of the stack without removing it
	Peek() (T, error)
	// Size returns the number of elements in the stack
	Size() int
	// IsEmpty returns true if the stack contains no elements
	IsEmpty() bool
}

// SliceStack implements Stack interface using a slice
type SliceStack[T any] struct {
	items []T
	mu    sync.RWMutex
}

// Push adds an element to the top of the stack
func (s *SliceStack[T]) Push(item T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
}

// Pop removes and returns the element at the top of the stack
func (s *SliceStack[T]) Pop() (T, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var zero T
	if len(s.items) == 0 {
		return zero, ErrStackEmpty
	}

	// Get the last item
	index := len(s.items) - 1
	item := s.items[index]

	// Remove the last item from the stack
	s.items = s.items[:index]

	return item, nil
}

// Peek returns the element at the top of the stack without removing it
func (s *SliceStack[T]) Peek() (T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var zero T
	if len(s.items) == 0 {
		return zero, ErrStackEmpty
	}

	return s.items[len(s.items)-1], nil
}

// Size returns the number of elements in the stack
func (s *SliceStack[T]) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

// IsEmpty returns true if the stack contains no elements
func (s *SliceStack[T]) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items) == 0
}

// NewStack creates and returns a new stack
func NewStack[T any]() Stack[T] {
	return &SliceStack[T]{
		items: make([]T, 0),
	}
}
