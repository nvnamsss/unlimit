package algo

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSliceQueue_New(t *testing.T) {
	q := NewQueue[int]()
	if q == nil {
		t.Fatal("NewQueue() returned nil")
	}

	if !q.IsEmpty() {
		t.Error("New queue should be empty")
	}

	if q.Size() != 0 {
		t.Errorf("New queue size should be 0, got %d", q.Size())
	}
}

func TestSliceQueue_PushPop(t *testing.T) {
	q := NewQueue[int]()

	// Test Push
	q.Push(1)
	if q.IsEmpty() {
		t.Error("Queue should not be empty after Push")
	}
	if q.Size() != 1 {
		t.Errorf("Queue size should be 1, got %d", q.Size())
	}

	// Test multiple Push operations
	q.Push(2)
	q.Push(3)
	if q.Size() != 3 {
		t.Errorf("Queue size should be 3, got %d", q.Size())
	}

	// Test Pop operations
	val, err := q.Pop()
	if err != nil {
		t.Errorf("Pop() returned error: %v", err)
	}
	if val != 1 {
		t.Errorf("Pop() should return 1, got %v", val)
	}
	if q.Size() != 2 {
		t.Errorf("Queue size should be 2 after Pop, got %d", q.Size())
	}

	val, err = q.Pop()
	if err != nil {
		t.Errorf("Pop() returned error: %v", err)
	}
	if val != 2 {
		t.Errorf("Pop() should return 2, got %v", val)
	}

	val, err = q.Pop()
	if err != nil {
		t.Errorf("Pop() returned error: %v", err)
	}
	if val != 3 {
		t.Errorf("Pop() should return 3, got %v", val)
	}

	// Queue should be empty now
	if !q.IsEmpty() {
		t.Error("Queue should be empty after all elements are popped")
	}

	_, err = q.Pop()
	if err == nil {
		t.Error("Pop() on empty queue should return error")
	}
	if !errors.Is(err, ErrQueueEmpty) {
		t.Errorf("Expected %v, got %v", ErrQueueEmpty, err)
	}
}

func TestSliceQueue_Peek(t *testing.T) {
	q := NewQueue[string]()

	// Peek on empty queue
	_, err := q.Peek()
	if err == nil {
		t.Error("Peek() on empty queue should return error")
	}
	if !errors.Is(err, ErrQueueEmpty) {
		t.Errorf("Expected %v, got %v", ErrQueueEmpty, err)
	}

	// Add item and peek
	q.Push("test")
	val, err := q.Peek()
	if err != nil {
		t.Errorf("Peek() returned error: %v", err)
	}
	if val != "test" {
		t.Errorf("Peek() should return 'test', got %v", val)
	}

	// Make sure Peek doesn't remove the item
	if q.Size() != 1 {
		t.Errorf("Queue size should still be 1 after Peek, got %d", q.Size())
	}

	// Add more items and test peek again
	q.Push("test2")
	val, err = q.Peek()
	if err != nil {
		t.Errorf("Peek() returned error: %v", err)
	}
	// Peek should still return the first item
	if val != "test" {
		t.Errorf("Peek() should return 'test', got %v", val)
	}
}

func TestSliceQueue_EmptyOperations(t *testing.T) {
	q := NewQueue[int]()

	// Pop from empty queue
	_, err := q.Pop()
	if err == nil {
		t.Error("Pop() on empty queue should return error")
	}
	if !errors.Is(err, ErrQueueEmpty) {
		t.Errorf("Expected %v, got %v", ErrQueueEmpty, err)
	}

	// Check Size and IsEmpty
	if q.Size() != 0 {
		t.Errorf("Empty queue size should be 0, got %d", q.Size())
	}
	if !q.IsEmpty() {
		t.Error("Queue should be empty")
	}
}

func TestSliceQueue_DifferentTypes(t *testing.T) {
	// Test with integers
	intQ := NewQueue[int]()
	intQ.Push(42)
	intQ.Push(99)
	val, _ := intQ.Pop()
	if val != 42 {
		t.Errorf("Expected 42, got %v", val)
	}

	// Test with strings
	stringQ := NewQueue[string]()
	stringQ.Push("hello")
	stringQ.Push("world")
	strVal, _ := stringQ.Pop()
	if strVal != "hello" {
		t.Errorf("Expected 'hello', got %v", strVal)
	}

	// Test with floats
	floatQ := NewQueue[float64]()
	floatQ.Push(3.14)
	floatQ.Push(2.71)
	floatVal, _ := floatQ.Pop()
	if floatVal != 3.14 {
		t.Errorf("Expected 3.14, got %v", floatVal)
	}

	// Test with structs
	type Person struct{ Name string }
	structQ := NewQueue[Person]()
	structQ.Push(Person{"test"})
	structQ.Push(Person{"another"})
	personVal, _ := structQ.Pop()
	if personVal.Name != "test" {
		t.Errorf("Expected Person with Name='test', got %v", personVal)
	}
}

func TestSliceQueue_ConcurrentOperations(t *testing.T) {
	q := NewQueue[int]()
	const itemCount = 1000
	var wg sync.WaitGroup

	// Push concurrently
	wg.Add(itemCount)
	for i := 0; i < itemCount; i++ {
		go func(val int) {
			defer wg.Done()
			q.Push(val)
		}(i)
	}
	wg.Wait()

	if q.Size() != itemCount {
		t.Errorf("Queue size should be %d, got %d", itemCount, q.Size())
	}

	// Create a map to check if all values are popped exactly once
	valMap := make(map[int]bool)

	// Pop concurrently
	wg.Add(itemCount)
	var mu sync.Mutex
	for i := 0; i < itemCount; i++ {
		go func() {
			defer wg.Done()
			val, err := q.Pop()
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

	if q.Size() != 0 {
		t.Errorf("Queue should be empty after all Pops, size: %d", q.Size())
	}

	if !q.IsEmpty() {
		t.Error("Queue should be empty after all Pops")
	}

	// Check if all values were popped
	if len(valMap) != itemCount {
		t.Errorf("Expected %d unique values to be popped, got %d", itemCount, len(valMap))
	}
}

// Add a comparative test as mentioned in the guidelines
func TestSliceQueue_CompareWithSlice(t *testing.T) {
	q := NewQueue[int]()
	refSlice := make([]int, 0)

	// Test push operations match
	testValues := []int{1, 2, 3, 4}

	for _, val := range testValues {
		q.Push(val)
		refSlice = append(refSlice, val)
	}

	if q.Size() != len(refSlice) {
		t.Errorf("Queue size %d does not match reference slice size %d", q.Size(), len(refSlice))
	}

	// Test that queue behaves like FIFO same as slice
	for i := 0; i < len(testValues); i++ {
		expected := refSlice[0]
		refSlice = refSlice[1:]

		actual, err := q.Pop()
		if err != nil {
			t.Errorf("Queue Pop() returned error: %v", err)
		}

		if actual != expected {
			t.Errorf("Queue Pop() returned %v, expected %v", actual, expected)
		}
	}

	// Both should be empty now
	if !q.IsEmpty() || len(refSlice) != 0 {
		t.Errorf("Both queue and reference slice should be empty")
	}
}

func TestSliceQueue_ParallelOperations(t *testing.T) {
	producersDone = 0 // Reset the global variable to avoid interference between tests

	// Define test item type
	type TestItem struct {
		ProducerID int
		ItemNumber int
	}

	// Test with different numbers of producers and consumers
	testCases := []struct {
		name         string
		producers    int
		consumers    int
		itemsPerProd int
	}{
		{"Equal Producers and Consumers", 10, 10, 100},
		{"More Producers than Consumers", 20, 5, 50},
		{"More Consumers than Producers", 5, 20, 200},
		{"Single Producer Multiple Consumers", 1, 10, 1000},
		{"Multiple Producers Single Consumer", 10, 1, 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q := NewQueue[TestItem]()
			var wg sync.WaitGroup
			totalItems := tc.producers * tc.itemsPerProd

			// Channel to collect consumed items for verification
			consumed := make(chan TestItem, totalItems)

			// Error channel to report any issues during testing
			errCh := make(chan error, tc.producers+tc.consumers)

			// Start producers
			wg.Add(tc.producers)
			for p := 0; p < tc.producers; p++ {
				go func(producerID int) {
					defer wg.Done()
					for i := 0; i < tc.itemsPerProd; i++ {
						// Create an item that identifies producer and item number
						item := TestItem{
							ProducerID: producerID,
							ItemNumber: i,
						}
						q.Push(item)

						// Introduce small random delay to increase contention
						if i%10 == 0 {
							time.Sleep(time.Microsecond)
						}
					}
				}(p)
			}

			// Wait for a bit to allow some items to be pushed before starting consumers
			time.Sleep(time.Millisecond)

			// Start consumers
			wg.Add(tc.consumers)
			for c := 0; c < tc.consumers; c++ {
				go func() {
					defer wg.Done()
					for {
						item, err := q.Pop()
						if err != nil {
							if errors.Is(err, ErrQueueEmpty) {
								// Queue might be temporarily empty, check if all producers are done
								if q.Size() == 0 && atomic.LoadInt32(&producersDone) == int32(tc.producers) {
									return // All producers are done and queue is empty
								}
								// Otherwise, wait a bit and try again
								time.Sleep(time.Microsecond * 10)
								continue
							}
							// Unexpected error
							errCh <- fmt.Errorf("consumer error: %w", err)
							return
						}

						// Successfully got an item, record it
						consumed <- item
					}
				}()
			}

			// Wait for producers to finish
			wg.Wait()
			atomic.StoreInt32(&producersDone, int32(tc.producers))

			// Create a timeout for collecting all consumed items
			timeout := time.After(5 * time.Second)

			// Track all consumed items
			producerItemCounts := make(map[int]map[int]bool)
			for i := 0; i < tc.producers; i++ {
				producerItemCounts[i] = make(map[int]bool)
			}

			// Collect all items until we've received the expected number
			for i := 0; i < totalItems; i++ {
				select {
				case item := <-consumed:
					pid := item.ProducerID
					itemNum := item.ItemNumber

					// Verify we haven't seen this item before
					if producerItemCounts[pid][itemNum] {
						t.Errorf("Item from producer %d with number %d was consumed multiple times", pid, itemNum)
					}
					producerItemCounts[pid][itemNum] = true

				case err := <-errCh:
					t.Errorf("Error during parallel operation: %v", err)
					return
				case <-timeout:
					t.Errorf("Test timed out waiting for all items to be consumed. Expected %d items.", totalItems)
					return
				}
			}

			// Verify all items were consumed exactly once
			for p := 0; p < tc.producers; p++ {
				for i := 0; i < tc.itemsPerProd; i++ {
					if !producerItemCounts[p][i] {
						t.Errorf("Item from producer %d with number %d was not consumed", p, i)
					}
				}
			}

			// Ensure queue is empty at the end
			if !q.IsEmpty() {
				t.Errorf("Queue should be empty after all operations, but has size: %d", q.Size())
			}
		})
	}
}

// Add a new test to check push/peek operations under load
func TestSliceQueue_ParallelPushPeek(t *testing.T) {
	q := NewQueue[string]()
	const numGoroutines = 20
	const operationsPerGoroutine = 100
	var wg sync.WaitGroup

	// Start pushers - they'll continuously add items to the queue
	wg.Add(numGoroutines / 2)
	for i := 0; i < numGoroutines/2; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				q.Push(fmt.Sprintf("item-%d-%d", id, j))
				time.Sleep(time.Microsecond) // Small delay to increase contention
			}
		}(i)
	}

	// Start peekers - they'll continuously peek at the front of the queue without removing items
	peekErrors := int32(0)
	wg.Add(numGoroutines / 2)
	for i := 0; i < numGoroutines/2; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				_, err := q.Peek()
				if err != nil && !errors.Is(err, ErrQueueEmpty) {
					atomic.AddInt32(&peekErrors, 1)
				}
				time.Sleep(time.Microsecond) // Small delay to increase contention
			}
		}()
	}

	wg.Wait()

	// Verify no unexpected errors occurred
	if peekErrors > 0 {
		t.Errorf("Got %d unexpected peek errors", peekErrors)
	}

	// Validate queue has the expected number of items
	expectedItems := (numGoroutines / 2) * operationsPerGoroutine
	if q.Size() != expectedItems {
		t.Errorf("Queue should have %d items, but got %d", expectedItems, q.Size())
	}

	// Now drain the queue and check we can get all items
	for i := 0; i < expectedItems; i++ {
		_, err := q.Pop()
		if err != nil {
			t.Errorf("Failed to pop item %d: %v", i, err)
		}
	}

	// Queue should be empty now
	if !q.IsEmpty() {
		t.Errorf("Queue should be empty but has %d items", q.Size())
	}
}

var producersDone int32 // Used to coordinate producers and consumers
