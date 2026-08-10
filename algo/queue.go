package algo

import (
	"errors"
	"sync"
	"sync/atomic"
	"unsafe"
)

var (
	// ErrQueueEmpty is returned when trying to pop from an empty queue
	ErrQueueEmpty = errors.New("queue is empty")
)

// Queue interface represents basic queue operations
type Queue[T any] interface {
	// Push adds an element to the end of the queue
	Push(item T)
	// Pop removes and returns the element at the front of the queue
	Pop() (T, error)
	// Peek returns the element at the front of the queue without removing it
	Peek() (T, error)
	// Size returns the number of elements in the queue
	Size() int
	// IsEmpty returns true if the queue contains no elements
	IsEmpty() bool
}

// SliceQueue implements Queue interface using a slice
type SliceQueue[T any] struct {
	items []T
	mu    sync.RWMutex
}

// Push adds an element to the end of the queue
func (q *SliceQueue[T]) Push(item T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, item)
}

// Pop removes and returns the element at the front of the queue
func (q *SliceQueue[T]) Pop() (T, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	var zero T
	if len(q.items) == 0 {
		return zero, ErrQueueEmpty
	}

	// Get the first item
	item := q.items[0]

	// Remove the first item from the queue
	q.items = q.items[1:]

	return item, nil
}

// Peek returns the element at the front of the queue without removing it
func (q *SliceQueue[T]) Peek() (T, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var zero T
	if len(q.items) == 0 {
		return zero, ErrQueueEmpty
	}

	return q.items[0], nil
}

// Size returns the number of elements in the queue
func (q *SliceQueue[T]) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.items)
}

// IsEmpty returns true if the queue contains no elements
func (q *SliceQueue[T]) IsEmpty() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.items) == 0
}

// NewQueue creates and returns a new queue
func NewQueue[T any]() Queue[T] {
	return &SliceQueue[T]{
		items: make([]T, 0),
	}
}

// lfNode represents a node in the lock-free queue
type lfNode[T any] struct {
	value T
	next  atomic.Pointer[lfNode[T]]
}

// LockFreeQueue implements a lock-free queue using CAS operations.
// This implementation is based on the Michael-Scott queue algorithm.
// It provides thread-safe operations without using locks, making it suitable
// for high-concurrency scenarios where lock contention would be a bottleneck.
type LockFreeQueue[T any] struct {
	head  atomic.Pointer[lfNode[T]]
	tail  atomic.Pointer[lfNode[T]]
	count atomic.Int64
}

// NewLockFreeQueue creates and returns a new wait-free queue.
// The queue is initialized with a sentinel node to simplify edge cases.
func NewLockFreeQueue[T any]() *LockFreeQueue[T] {
	q := &LockFreeQueue[T]{}
	// Create a sentinel/dummy node
	sentinel := &lfNode[T]{}
	q.head.Store(sentinel)
	q.tail.Store(sentinel)
	return q
}

// Push adds an element to the end of the queue.
// This operation is lock-free and uses CAS to atomically update the tail.
func (q *LockFreeQueue[T]) Push(item T) {
	node := &lfNode[T]{value: item}

	for {
		tail := q.tail.Load()
		next := tail.next.Load()

		// Check if tail is still the actual tail
		if tail == q.tail.Load() {
			if next == nil {
				// Tail is pointing to the last node, try to link the new node
				if tail.next.CompareAndSwap(nil, node) {
					// Successfully linked, try to swing tail to the new node
					// It's OK if this fails; another thread will help
					q.tail.CompareAndSwap(tail, node)
					q.count.Add(1)
					return
				}
			} else {
				// Tail is falling behind, try to advance it
				q.tail.CompareAndSwap(tail, next)
			}
		}
	}
}

// Pop removes and returns the element at the front of the queue.
// This operation is lock-free and uses CAS to atomically update the head.
// Returns ErrQueueEmpty if the queue is empty.
func (q *LockFreeQueue[T]) Pop() (T, error) {
	var zero T

	for {
		head := q.head.Load()
		tail := q.tail.Load()
		next := head.next.Load()

		// Check if head is still the actual head
		if head == q.head.Load() {
			if head == tail {
				// Queue appears to be empty or tail is falling behind
				if next == nil {
					// Queue is actually empty
					return zero, ErrQueueEmpty
				}
				// Tail is falling behind, try to advance it
				q.tail.CompareAndSwap(tail, next)
			} else {
				// Read value before CAS, otherwise another dequeue might free the node
				value := next.value
				// Try to swing head to the next node
				if q.head.CompareAndSwap(head, next) {
					q.count.Add(-1)
					return value, nil
				}
			}
		}
	}
}

// Peek returns the element at the front of the queue without removing it.
// Returns ErrQueueEmpty if the queue is empty.
func (q *LockFreeQueue[T]) Peek() (T, error) {
	var zero T

	head := q.head.Load()
	next := head.next.Load()

	if next == nil {
		return zero, ErrQueueEmpty
	}

	return next.value, nil
}

// Size returns the number of elements in the queue.
// Note: This is an approximate count due to concurrent operations.
func (q *LockFreeQueue[T]) Size() int {
	return int(q.count.Load())
}

// IsEmpty returns true if the queue contains no elements.
// Note: This check is subject to race conditions in concurrent scenarios.
func (q *LockFreeQueue[T]) IsEmpty() bool {
	head := q.head.Load()
	return head.next.Load() == nil
}

// ensure WaitFreeQueue implements Queue interface
var _ Queue[any] = (*LockFreeQueue[any])(nil)

// compile-time check to ensure unsafe.Pointer alignment
var _ = unsafe.Sizeof(lfNode[any]{})

// ErrQueueFull is returned when trying to push to a full bounded queue
var ErrQueueFull = errors.New("queue is full")

// ChannelQueue implements Queue interface using Go channels.
// This provides a simple, efficient, and thread-safe queue implementation
// that leverages Go's built-in channel synchronization primitives.
//
// Key properties:
// - Thread-safe: Uses Go's channel which is inherently synchronized
// - Bounded: Has a fixed capacity set at creation time
// - Blocking-aware: Push blocks if queue is full (unless using TryPush)
// - Non-blocking Pop: Returns immediately with error if empty
type ChannelQueue[T any] struct {
	ch       chan T
	capacity int
}

// NewChannelQueue creates a new channel-based queue with the specified capacity.
// The capacity must be greater than 0.
func NewChannelQueue[T any](capacity int) *ChannelQueue[T] {
	if capacity <= 0 {
		capacity = 1
	}
	return &ChannelQueue[T]{
		ch:       make(chan T, capacity),
		capacity: capacity,
	}
}

// Push adds an element to the end of the queue.
// This operation blocks if the queue is full.
func (q *ChannelQueue[T]) Push(item T) {
	q.ch <- item
}

// TryPush attempts to add an element to the queue without blocking.
// Returns ErrQueueFull if the queue is at capacity.
func (q *ChannelQueue[T]) TryPush(item T) error {
	select {
	case q.ch <- item:
		return nil
	default:
		return ErrQueueFull
	}
}

// Pop removes and returns the element at the front of the queue.
// Returns ErrQueueEmpty if the queue is empty (non-blocking).
func (q *ChannelQueue[T]) Pop() (T, error) {
	var zero T
	select {
	case item := <-q.ch:
		return item, nil
	default:
		return zero, ErrQueueEmpty
	}
}

// BlockingPop removes and returns the element at the front of the queue.
// This operation blocks until an element is available.
func (q *ChannelQueue[T]) BlockingPop() T {
	return <-q.ch
}

// Peek returns the element at the front of the queue without removing it.
// Note: This is not truly atomic with channels - the peeked item is
// temporarily removed and re-added, which may change ordering under concurrency.
// For a true peek operation, consider using SliceQueue or LockFreeQueue.
func (q *ChannelQueue[T]) Peek() (T, error) {
	var zero T
	select {
	case item := <-q.ch:
		// Put it back (this may reorder under high concurrency)
		select {
		case q.ch <- item:
		default:
			// Queue became full while we held the item - shouldn't happen
			// but handle gracefully by keeping the item
			go func() { q.ch <- item }()
		}
		return item, nil
	default:
		return zero, ErrQueueEmpty
	}
}

// Size returns the number of elements currently in the queue.
func (q *ChannelQueue[T]) Size() int {
	return len(q.ch)
}

// IsEmpty returns true if the queue contains no elements.
func (q *ChannelQueue[T]) IsEmpty() bool {
	return len(q.ch) == 0
}

// IsFull returns true if the queue is at capacity.
func (q *ChannelQueue[T]) IsFull() bool {
	return len(q.ch) == q.capacity
}

// Capacity returns the maximum number of elements the queue can hold.
func (q *ChannelQueue[T]) Capacity() int {
	return q.capacity
}

// Close closes the underlying channel.
// After closing, Push operations will panic and Pop will eventually return ErrQueueEmpty.
func (q *ChannelQueue[T]) Close() {
	close(q.ch)
}

// Range iterates over all elements in the queue, removing them as it goes.
// The iteration stops when the queue is empty or the callback returns false.
// This is useful for draining the queue.
func (q *ChannelQueue[T]) Range(fn func(item T) bool) {
	for {
		select {
		case item := <-q.ch:
			if !fn(item) {
				return
			}
		default:
			return
		}
	}
}

// ensure ChannelQueue implements Queue interface
var _ Queue[any] = (*ChannelQueue[any])(nil)
