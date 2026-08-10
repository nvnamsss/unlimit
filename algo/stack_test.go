package algo

import (
	"sync"
	"testing"
)

func TestSliceStack_New(t *testing.T) {
	s := NewStack[int]()
	if s == nil {
		t.Fatal("NewStack() returned nil")
	}

	if !s.IsEmpty() {
		t.Error("New stack should be empty")
	}

	if s.Size() != 0 {
		t.Errorf("New stack size should be 0, got %d", s.Size())
	}
}

func TestSliceStack_PushPop(t *testing.T) {
	s := NewStack[int]()

	// Test Push
	s.Push(1)
	if s.IsEmpty() {
		t.Error("Stack should not be empty after Push")
	}
	if s.Size() != 1 {
		t.Errorf("Stack size should be 1, got %d", s.Size())
	}

	// Test multiple Push operations
	s.Push(2)
	s.Push(3)
	if s.Size() != 3 {
		t.Errorf("Stack size should be 3, got %d", s.Size())
	}

	// Test Pop operations - should return in reverse order (LIFO)
	val, err := s.Pop()
	if err != nil {
		t.Errorf("Pop() returned error: %v", err)
	}
	if val != 3 {
		t.Errorf("Pop() should return 3, got %v", val)
	}
	if s.Size() != 2 {
		t.Errorf("Stack size should be 2 after Pop, got %d", s.Size())
	}

	val, err = s.Pop()
	if err != nil {
		t.Errorf("Pop() returned error: %v", err)
	}
	if val != 2 {
		t.Errorf("Pop() should return 2, got %v", val)
	}

	val, err = s.Pop()
	if err != nil {
		t.Errorf("Pop() returned error: %v", err)
	}
	if val != 1 {
		t.Errorf("Pop() should return 1, got %v", val)
	}

	// Stack should be empty now
	if !s.IsEmpty() {
		t.Error("Stack should be empty after all elements are popped")
	}

	_, err = s.Pop()
	if err == nil {
		t.Error("Pop() on empty stack should return error")
	}
	if err != ErrStackEmpty {
		t.Errorf("Expected ErrStackEmpty, got %v", err)
	}
}

func TestSliceStack_Peek(t *testing.T) {
	s := NewStack[string]()

	// Peek on empty stack
	_, err := s.Peek()
	if err == nil {
		t.Error("Peek() on empty stack should return error")
	}
	if err != ErrStackEmpty {
		t.Errorf("Expected ErrStackEmpty, got %v", err)
	}

	// Add item and peek
	s.Push("test")
	val, err := s.Peek()
	if err != nil {
		t.Errorf("Peek() returned error: %v", err)
	}
	if val != "test" {
		t.Errorf("Peek() should return 'test', got %v", val)
	}

	// Make sure Peek doesn't remove the item
	if s.Size() != 1 {
		t.Errorf("Stack size should still be 1 after Peek, got %d", s.Size())
	}

	// Add more items and test peek again
	s.Push("test2")
	val, err = s.Peek()
	if err != nil {
		t.Errorf("Peek() returned error: %v", err)
	}
	// Peek should return the last item in a stack (unlike a queue)
	if val != "test2" {
		t.Errorf("Peek() should return 'test2', got %v", val)
	}
}

func TestSliceStack_EmptyOperations(t *testing.T) {
	s := NewStack[int]()

	// Pop from empty stack
	_, err := s.Pop()
	if err == nil {
		t.Error("Pop() on empty stack should return error")
	}
	if err != ErrStackEmpty {
		t.Errorf("Expected ErrStackEmpty, got %v", err)
	}

	// Check Size and IsEmpty
	if s.Size() != 0 {
		t.Errorf("Empty stack size should be 0, got %d", s.Size())
	}
	if !s.IsEmpty() {
		t.Error("Stack should be empty")
	}
}

func TestSliceStack_DifferentTypes(t *testing.T) {
	// Test with integers
	intStack := NewStack[int]()
	intStack.Push(42)
	val, _ := intStack.Pop()
	if val != 42 {
		t.Errorf("Expected 42, got %v", val)
	}

	// Test with strings
	stringStack := NewStack[string]()
	stringStack.Push("string")
	strVal, _ := stringStack.Pop()
	if strVal != "string" {
		t.Errorf("Expected 'string', got %v", strVal)
	}

	// Test with floats
	floatStack := NewStack[float64]()
	floatStack.Push(3.14)
	floatVal, _ := floatStack.Pop()
	if floatVal != 3.14 {
		t.Errorf("Expected 3.14, got %v", floatVal)
	}

	// Test with structs
	type Person struct{ Name string }
	structStack := NewStack[Person]()
	structStack.Push(Person{"test"})
	personVal, _ := structStack.Pop()
	if personVal.Name != "test" {
		t.Errorf("Expected struct with Name='test', got %v", personVal)
	}
}

func TestSliceStack_ConcurrentOperations(t *testing.T) {
	s := NewStack[int]()
	const itemCount = 1000
	var wg sync.WaitGroup

	// Push concurrently
	wg.Add(itemCount)
	for i := 0; i < itemCount; i++ {
		go func(val int) {
			defer wg.Done()
			s.Push(val)
		}(i)
	}
	wg.Wait()

	if s.Size() != itemCount {
		t.Errorf("Stack size should be %d, got %d", itemCount, s.Size())
	}

	// Create a map to check if all values are popped exactly once
	valMap := make(map[int]bool)

	// Pop concurrently
	wg.Add(itemCount)
	var mu sync.Mutex
	for i := 0; i < itemCount; i++ {
		go func() {
			defer wg.Done()
			val, err := s.Pop()
			if err != nil {
				t.Errorf("Pop() returned error: %v", err)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			if valMap[val] {
				t.Errorf("Value %v was already popped", val)
			}
			valMap[val] = true
		}()
	}
	wg.Wait()

	if s.Size() != 0 {
		t.Errorf("Stack should be empty after all Pops, size: %d", s.Size())
	}

	if !s.IsEmpty() {
		t.Error("Stack should be empty after all Pops")
	}

	// Check if all values were popped
	if len(valMap) != itemCount {
		t.Errorf("Expected %d unique values to be popped, got %d", itemCount, len(valMap))
	}
}

func TestSliceStack_CompareWithQueue(t *testing.T) {
	stack := NewStack[int]()
	queue := NewQueue[int]()

	// Push same values to both
	for i := 1; i <= 3; i++ {
		stack.Push(i)
		queue.Push(i)
	}

	// Stack should be LIFO (last in, first out)
	for i := 3; i >= 1; i-- {
		val, _ := stack.Pop()
		if val != i {
			t.Errorf("Stack Pop() should return %d, got %v", i, val)
		}
	}

	// Queue should be FIFO (first in, first out)
	for i := 1; i <= 3; i++ {
		val, _ := queue.Pop()
		if val != i {
			t.Errorf("Queue Pop() should return %d, got %v", i, val)
		}
	}
}
