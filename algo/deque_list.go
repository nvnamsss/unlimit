package algo

// DequeList represents a double-ended queue implemented using a doubly linked list.
// This implementation offers O(1) complexity for all primary operations.
type DequeList[T any] struct {
	head *DoublyNode[T]
	tail *DoublyNode[T]
	size int
}

// NewDequeList creates and returns a new empty DequeList.
func NewDequeList[T any]() *DequeList[T] {
	return &DequeList[T]{
		head: nil,
		tail: nil,
		size: 0,
	}
}

// PushBack adds an element to the back of the deque.
func (d *DequeList[T]) PushBack(item T) {
	newNode := &DoublyNode[T]{
		Value: item,
		Prev:  d.tail,
		Next:  nil,
	}

	if d.IsEmpty() {
		d.head = newNode
		d.tail = newNode
	} else {
		d.tail.Next = newNode
		d.tail = newNode
	}

	d.size++
}

// PushFront adds an element to the front of the deque.
func (d *DequeList[T]) PushFront(item T) {
	newNode := &DoublyNode[T]{
		Value: item,
		Prev:  nil,
		Next:  d.head,
	}

	if d.IsEmpty() {
		d.head = newNode
		d.tail = newNode
	} else {
		d.head.Prev = newNode
		d.head = newNode
	}

	d.size++
}

// PopBack removes and returns the last element from the deque.
// If the deque is empty, returns the zero value of type T and false.
func (d *DequeList[T]) PopBack() (T, bool) {
	var zero T
	if d.IsEmpty() {
		return zero, false
	}

	item := d.tail.Value

	if d.head == d.tail {
		// Only one element in the deque
		d.head = nil
		d.tail = nil
	} else {
		// More than one element
		d.tail = d.tail.Prev
		d.tail.Next = nil
	}

	d.size--
	return item, true
}

// PopFront removes and returns the first element from the deque.
// If the deque is empty, returns the zero value of type T and false.
func (d *DequeList[T]) PopFront() (T, bool) {
	var zero T
	if d.IsEmpty() {
		return zero, false
	}

	item := d.head.Value

	if d.head == d.tail {
		// Only one element in the deque
		d.head = nil
		d.tail = nil
	} else {
		// More than one element
		d.head = d.head.Next
		d.head.Prev = nil
	}

	d.size--
	return item, true
}

// PeekBack returns the last element of the deque without removing it.
// If the deque is empty, returns the zero value of type T and false.
func (d *DequeList[T]) PeekBack() (T, bool) {
	var zero T
	if d.IsEmpty() {
		return zero, false
	}
	return d.tail.Value, true
}

// PeekFront returns the first element of the deque without removing it.
// If the deque is empty, returns the zero value of type T and false.
func (d *DequeList[T]) PeekFront() (T, bool) {
	var zero T
	if d.IsEmpty() {
		return zero, false
	}
	return d.head.Value, true
}

// Size returns the number of elements in the deque.
func (d *DequeList[T]) Size() int {
	return d.size
}

// IsEmpty returns true if the deque is empty, false otherwise.
func (d *DequeList[T]) IsEmpty() bool {
	return d.size == 0
}

// Clear removes all elements from the deque.
func (d *DequeList[T]) Clear() {
	d.head = nil
	d.tail = nil
	d.size = 0
}
