package algo

// LinkedList represents a singly linked list data structure
type LinkedList[T any] struct {
	Head *ListNode[T]
	Tail *ListNode[T]
	Size int
}

// NewLinkedList creates and returns a new empty linked list
func NewLinkedList[T any]() *LinkedList[T] {
	return &LinkedList[T]{
		Head: nil,
		Tail: nil,
		Size: 0,
	}
}

// IsEmpty checks if the linked list is empty
func (ll *LinkedList[T]) IsEmpty() bool {
	return ll.Size == 0
}

// Add appends a new element to the end of the list
func (ll *LinkedList[T]) Add(data T) {
	newNode := &ListNode[T]{Data: data, Next: nil}

	if ll.IsEmpty() {
		ll.Head = newNode
	} else {
		ll.Tail.Next = newNode
	}

	ll.Tail = newNode
	ll.Size++
}

// AddFront adds a new element to the beginning of the list
func (ll *LinkedList[T]) AddFront(data T) {
	newNode := &ListNode[T]{Data: data, Next: ll.Head}
	ll.Head = newNode

	if ll.Tail == nil {
		ll.Tail = newNode
	}

	ll.Size++
}

// RemoveFirst removes the first element from the list
func (ll *LinkedList[T]) RemoveFirst() (T, bool) {
	var zero T
	if ll.IsEmpty() {
		return zero, false
	}

	data := ll.Head.Data
	ll.Head = ll.Head.Next
	ll.Size--

	if ll.Head == nil {
		ll.Tail = nil
	}

	return data, true
}

// Remove removes the element with the specified data
func (ll *LinkedList[T]) Remove(data T) bool {
	if ll.IsEmpty() {
		return false
	}

	// If head node holds the data to be deleted
	if compareEqual(ll.Head.Data, data) {
		ll.RemoveFirst()
		return true
	}

	// Search for the data to delete
	current := ll.Head
	for current.Next != nil && !compareEqual(current.Next.Data, data) {
		current = current.Next
	}

	// If data was not found
	if current.Next == nil {
		return false
	}

	// Unlink the node from the list
	if current.Next == ll.Tail {
		ll.Tail = current
	}

	current.Next = current.Next.Next
	ll.Size--

	return true
}

// Contains checks if the list contains the specified element
func (ll *LinkedList[T]) Contains(data T) bool {
	current := ll.Head
	for current != nil {
		if compareEqual(current.Data, data) {
			return true
		}
		current = current.Next
	}
	return false
}

// Get returns the element at the specified position
func (ll *LinkedList[T]) Get(index int) (T, bool) {
	var zero T
	if index < 0 || index >= ll.Size {
		return zero, false
	}

	current := ll.Head
	for i := 0; i < index; i++ {
		current = current.Next
	}

	return current.Data, true
}

// GetFirst returns the first element in the list
func (ll *LinkedList[T]) GetFirst() (T, bool) {
	var zero T
	if ll.IsEmpty() {
		return zero, false
	}
	return ll.Head.Data, true
}

// GetLast returns the last element in the list
func (ll *LinkedList[T]) GetLast() (T, bool) {
	var zero T
	if ll.IsEmpty() {
		return zero, false
	}
	return ll.Tail.Data, true
}

// Clear removes all elements from the list
func (ll *LinkedList[T]) Clear() {
	ll.Head = nil
	ll.Tail = nil
	ll.Size = 0
}

// ToSlice converts the linked list to a slice
func (ll *LinkedList[T]) ToSlice() []T {
	result := make([]T, ll.Size)
	current := ll.Head
	for i := 0; current != nil; i++ {
		result[i] = current.Data
		current = current.Next
	}
	return result
}

// ForEach executes the provided function once for each list element
func (ll *LinkedList[T]) ForEach(f func(data T)) {
	current := ll.Head
	for current != nil {
		f(current.Data)
		current = current.Next
	}
}

// Helper function to compare values for equality
// This is needed since Go doesn't allow direct comparison with == for all types
func compareEqual[T any](a, b T) bool {
	// For comparable types, we can use reflection
	// This is a simplified implementation - in production code you might want to
	// handle special types differently or use constraints like comparable
	return any(a) == any(b)
}
