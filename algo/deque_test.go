package algo

import (
	"testing"
)

func TestDeque_New(t *testing.T) {
	deque := NewDeque[int]()
	if deque == nil {
		t.Error("NewDeque() should return a non-nil deque")
	}
	if !deque.IsEmpty() {
		t.Error("New deque should be empty")
	}
	if deque.Size() != 0 {
		t.Errorf("New deque should have size 0, got %d", deque.Size())
	}
}

func TestDeque_PushBack(t *testing.T) {
	deque := NewDeque[int]()

	// Test single push
	deque.PushBack(1)
	if deque.Size() != 1 {
		t.Errorf("Expected size 1, got %d", deque.Size())
	}
	val, ok := deque.PeekBack()
	if !ok || val != 1 {
		t.Errorf("Expected PeekBack() to return 1, got %v, %v", val, ok)
	}

	// Test multiple pushes
	deque.PushBack(2)
	deque.PushBack(3)
	if deque.Size() != 3 {
		t.Errorf("Expected size 3, got %d", deque.Size())
	}
	val, ok = deque.PeekBack()
	if !ok || val != 3 {
		t.Errorf("Expected PeekBack() to return 3, got %v, %v", val, ok)
	}
}

func TestDeque_PushFront(t *testing.T) {
	deque := NewDeque[int]()

	// Test single push
	deque.PushFront(1)
	if deque.Size() != 1 {
		t.Errorf("Expected size 1, got %d", deque.Size())
	}
	val, ok := deque.PeekFront()
	if !ok || val != 1 {
		t.Errorf("Expected PeekFront() to return 1, got %v, %v", val, ok)
	}

	// Test multiple pushes
	deque.PushFront(2)
	deque.PushFront(3)
	if deque.Size() != 3 {
		t.Errorf("Expected size 3, got %d", deque.Size())
	}
	val, ok = deque.PeekFront()
	if !ok || val != 3 {
		t.Errorf("Expected PeekFront() to return 3, got %v, %v", val, ok)
	}
}

func TestDeque_PopBack(t *testing.T) {
	var (
		val   int
		ok    bool
		deque = NewDeque[int]()
	)

	// Test pop on empty deque
	_, ok = deque.PopBack()
	if ok {
		t.Errorf("PopBack() on empty deque should return false, got %v", ok)
	}

	// Test pop after push
	deque.PushBack(1)
	deque.PushBack(2)

	val, ok = deque.PopBack()
	if !ok || val != 2 {
		t.Errorf("Expected PopBack() to return 2, got %v, %v", val, ok)
	}
	if deque.Size() != 1 {
		t.Errorf("Expected size 1 after pop, got %d", deque.Size())
	}

	val, ok = deque.PopBack()
	if !ok || val != 1 {
		t.Errorf("Expected PopBack() to return 1, got %v, %v", val, ok)
	}
	if !deque.IsEmpty() {
		t.Error("Deque should be empty after popping all elements")
	}
}

func TestDeque_PopFront(t *testing.T) {
	var (
		val   int
		ok    bool
		deque = NewDeque[int]()
	)

	// Test pop on empty deque
	_, ok = deque.PopFront()
	if ok {
		t.Errorf("PopFront() on empty deque should return false, got %v", ok)
	}

	// Test pop after push
	deque.PushBack(1)
	deque.PushBack(2)

	val, ok = deque.PopFront()
	if !ok || val != 1 {
		t.Errorf("Expected PopFront() to return 1, got %v, %v", val, ok)
	}
	if deque.Size() != 1 {
		t.Errorf("Expected size 1 after pop, got %d", deque.Size())
	}

	val, ok = deque.PopFront()
	if !ok || val != 2 {
		t.Errorf("Expected PopFront() to return 2, got %v, %v", val, ok)
	}
	if !deque.IsEmpty() {
		t.Error("Deque should be empty after popping all elements")
	}
}

func TestDeque_PeekBack(t *testing.T) {
	deque := NewDeque[int]()

	// Test peek on empty deque
	_, ok := deque.PeekBack()
	if ok {
		t.Error("PeekBack() on empty deque should return false")
	}

	// Test peek after push
	deque.PushBack(1)
	val, ok := deque.PeekBack()
	if !ok || val != 1 {
		t.Errorf("Expected PeekBack() to return 1, got %v, %v", val, ok)
	}

	// Ensure peek doesn't remove the item
	if deque.Size() != 1 {
		t.Errorf("PeekBack() should not remove items, expected size 1, got %d", deque.Size())
	}
}

func TestDeque_PeekFront(t *testing.T) {
	deque := NewDeque[int]()

	// Test peek on empty deque
	_, ok := deque.PeekFront()
	if ok {
		t.Error("PeekFront() on empty deque should return false")
	}

	// Test peek after push
	deque.PushBack(1)
	val, ok := deque.PeekFront()
	if !ok || val != 1 {
		t.Errorf("Expected PeekFront() to return 1, got %v, %v", val, ok)
	}

	// Test peek with multiple items
	deque.PushBack(2)
	val, ok = deque.PeekFront()
	if !ok || val != 1 {
		t.Errorf("Expected PeekFront() to return 1, got %v, %v", val, ok)
	}

	// Ensure peek doesn't remove the item
	if deque.Size() != 2 {
		t.Errorf("PeekFront() should not remove items, expected size 2, got %d", deque.Size())
	}
}

func TestDeque_Clear(t *testing.T) {
	deque := NewDeque[int]()
	deque.PushBack(1)
	deque.PushBack(2)
	deque.PushBack(3)

	deque.Clear()
	if !deque.IsEmpty() {
		t.Error("Deque should be empty after Clear()")
	}
	if deque.Size() != 0 {
		t.Errorf("Deque size should be 0 after Clear(), got %d", deque.Size())
	}
}

func TestDeque_IsEmpty(t *testing.T) {
	deque := NewDeque[int]()
	if !deque.IsEmpty() {
		t.Error("New deque should be empty")
	}

	deque.PushBack(1)
	if deque.IsEmpty() {
		t.Error("Deque with elements should not be empty")
	}

	deque.PopBack()
	if !deque.IsEmpty() {
		t.Error("Deque should be empty after removing all elements")
	}
}

func TestDeque_Size(t *testing.T) {
	deque := NewDeque[int]()
	if deque.Size() != 0 {
		t.Errorf("Expected size 0, got %d", deque.Size())
	}

	deque.PushBack(1)
	if deque.Size() != 1 {
		t.Errorf("Expected size 1, got %d", deque.Size())
	}

	deque.PushFront(2)
	if deque.Size() != 2 {
		t.Errorf("Expected size 2, got %d", deque.Size())
	}

	deque.PopBack()
	if deque.Size() != 1 {
		t.Errorf("Expected size 1, got %d", deque.Size())
	}

	deque.Clear()
	if deque.Size() != 0 {
		t.Errorf("Expected size 0 after Clear(), got %d", deque.Size())
	}
}

func TestDeque_DifferentTypes(t *testing.T) {
	// Test with strings
	stringDeque := NewDeque[string]()
	stringDeque.PushBack("hello")
	stringDeque.PushBack("world")
	val, ok := stringDeque.PopFront()
	if !ok || val != "hello" {
		t.Errorf("Expected 'hello', got %v", val)
	}

	// Test with custom struct
	type Person struct {
		Name string
		Age  int
	}

	personDeque := NewDeque[Person]()
	alice := Person{"Alice", 30}
	bob := Person{"Bob", 25}

	personDeque.PushBack(alice)
	personDeque.PushBack(bob)

	person, ok := personDeque.PopBack()
	if !ok || person.Name != "Bob" || person.Age != 25 {
		t.Errorf("Expected Bob/25, got %v/%v", person.Name, person.Age)
	}
}

func TestDeque_IntegrationTest(t *testing.T) {
	deque := NewDeque[int]()

	// Perform a series of operations
	deque.PushBack(1)
	deque.PushFront(2)
	deque.PushBack(3)
	deque.PushFront(4)
	// Now should have: [4, 2, 1, 3]

	val, _ := deque.PopFront()
	if val != 4 {
		t.Errorf("Expected 4, got %d", val)
	}

	val, _ = deque.PopBack()
	if val != 3 {
		t.Errorf("Expected 3, got %d", val)
	}

	val, _ = deque.PeekFront()
	if val != 2 {
		t.Errorf("Expected 2, got %d", val)
	}

	val, _ = deque.PeekBack()
	if val != 1 {
		t.Errorf("Expected 1, got %d", val)
	}

	if deque.Size() != 2 {
		t.Errorf("Expected size 2, got %d", deque.Size())
	}
}
