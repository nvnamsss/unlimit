package algo

// Deque represents a double-ended queue that allows adding and removing elements
// from both the front and back of the queue.
type Deque[T any] struct {
	elements []T
}

// NewDeque creates and returns a new empty Deque.
func NewDeque[T any]() *Deque[T] {
	return &Deque[T]{
		elements: make([]T, 0),
	}
}

// PushBack adds an element to the back of the deque.
func (d *Deque[T]) PushBack(item T) {
	d.elements = append(d.elements, item)
}

// PushFront adds an element to the front of the deque.
func (d *Deque[T]) PushFront(item T) {
	d.elements = append([]T{item}, d.elements...)
}

// PopBack removes and returns the last element from the deque.
// If the deque is empty, returns the zero value of type T and false.
func (d *Deque[T]) PopBack() (T, bool) {
	var zero T
	if d.IsEmpty() {
		return zero, false
	}

	lastIndex := len(d.elements) - 1
	item := d.elements[lastIndex]
	d.elements = d.elements[:lastIndex]
	return item, true
}

// PopFront removes and returns the first element from the deque.
// If the deque is empty, returns the zero value of type T and false.
func (d *Deque[T]) PopFront() (T, bool) {
	var zero T
	if d.IsEmpty() {
		return zero, false
	}

	item := d.elements[0]
	d.elements = d.elements[1:]
	return item, true
}

// PeekBack returns the last element of the deque without removing it.
// If the deque is empty, returns the zero value of type T and false.
func (d *Deque[T]) PeekBack() (T, bool) {
	var zero T
	if d.IsEmpty() {
		return zero, false
	}
	return d.elements[len(d.elements)-1], true
}

// PeekFront returns the first element of the deque without removing it.
// If the deque is empty, returns the zero value of type T and false.
func (d *Deque[T]) PeekFront() (T, bool) {
	var zero T
	if d.IsEmpty() {
		return zero, false
	}
	return d.elements[0], true
}

// Size returns the number of elements in the deque.
func (d *Deque[T]) Size() int {
	return len(d.elements)
}

// IsEmpty returns true if the deque is empty, false otherwise.
func (d *Deque[T]) IsEmpty() bool {
	return len(d.elements) == 0
}

// Clear removes all elements from the deque.
func (d *Deque[T]) Clear() {
	d.elements = make([]T, 0)
}
