package algo

import (
	"sync"
	"testing"
)

// TestLRUCache_New tests the creation of a new LRU cache
func TestLRUCache_New(t *testing.T) {
	// Test with invalid sizes
	_, err := NewLRU[string, int](-1)
	if err != ErrInvalidSize {
		t.Errorf("expected ErrInvalidSize, got: %v", err)
	}

	_, err = NewLRU[string, int](0)
	if err != ErrInvalidSize {
		t.Errorf("expected ErrInvalidSize, got: %v", err)
	}

	// Test with valid size
	l, err := NewLRU[string, int](1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if l.Len() != 0 {
		t.Errorf("expected len 0, got: %d", l.Len())
	}
}

// TestLRUCache_NewLRUNoError tests the creation of a new LRU cache with the no-error constructor
func TestLRUCache_NewLRUNoError(t *testing.T) {
	// Test with invalid size
	l := NewLRUNoError[string, int](-1)
	if l == nil {
		t.Fatalf("expected non-nil LRU")
	}
	if l.size != 1 {
		t.Errorf("expected size 1, got: %d", l.size)
	}

	// Test with valid size
	l = NewLRUNoError[string, int](5)
	if l == nil {
		t.Fatalf("expected non-nil LRU")
	}
	if l.size != 5 {
		t.Errorf("expected size 5, got: %d", l.size)
	}
}

// TestLRUCache_Add tests adding items to the cache
func TestLRUCache_Add(t *testing.T) {
	l, err := NewLRU[string, int](2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Add first item
	evicted := l.Add("key1", 1)
	if evicted {
		t.Errorf("expected no eviction")
	}
	if l.Len() != 1 {
		t.Errorf("expected len 1, got: %d", l.Len())
	}

	// Add second item
	evicted = l.Add("key2", 2)
	if evicted {
		t.Errorf("expected no eviction")
	}
	if l.Len() != 2 {
		t.Errorf("expected len 2, got: %d", l.Len())
	}

	// Add third item, should cause eviction
	evicted = l.Add("key3", 3)
	if !evicted {
		t.Errorf("expected eviction")
	}
	if l.Len() != 2 {
		t.Errorf("expected len 2, got: %d", l.Len())
	}

	// Check first item is gone
	_, ok := l.Get("key1")
	if ok {
		t.Errorf("unexpected key1")
	}

	// Update existing item
	evicted = l.Add("key2", 22)
	if evicted {
		t.Errorf("expected no eviction")
	}
	v, ok := l.Get("key2")
	if !ok {
		t.Errorf("key2 should exist")
	}
	if v != 22 {
		t.Errorf("expected 22, got: %d", v)
	}
}

// TestLRUCache_Get tests retrieving items from the cache
func TestLRUCache_Get(t *testing.T) {
	l, err := NewLRU[string, int](2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Setup
	l.Add("key1", 1)
	l.Add("key2", 2)

	// Exercise & Verify: Get existing item
	v, ok := l.Get("key1")
	if !ok {
		t.Errorf("key1 should exist")
	}
	if v != 1 {
		t.Errorf("expected 1, got: %d", v)
	}

	// Exercise & Verify: Get non-existent item
	v, ok = l.Get("key3")
	if ok {
		t.Errorf("key3 should not exist")
	}
	if v != 0 {
		t.Errorf("expected 0, got: %d", v)
	}
}

// TestLRUCache_LRUBehavior tests the Least Recently Used eviction policy
func TestLRUCache_LRUBehavior(t *testing.T) {
	l, err := NewLRU[string, int](2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Setup
	l.Add("key1", 1)
	l.Add("key2", 2)

	// Access key1 to make key2 the least recently used
	l.Get("key1")

	// Add new item to trigger eviction
	l.Add("key3", 3)

	// Verify key2 was evicted (it was the least recently used)
	_, ok := l.Get("key2")
	if ok {
		t.Errorf("key2 should have been evicted")
	}

	// Verify key1 and key3 are still present
	_, ok = l.Get("key1")
	if !ok {
		t.Errorf("key1 should still exist")
	}
	_, ok = l.Get("key3")
	if !ok {
		t.Errorf("key3 should exist")
	}
}

// TestLRUCache_Contains tests the Contains method
func TestLRUCache_Contains(t *testing.T) {
	l, err := NewLRU[string, int](2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Setup
	l.Add("key1", 1)

	// Exercise & Verify
	if !l.Contains("key1") {
		t.Errorf("key1 should exist")
	}
	if l.Contains("key2") {
		t.Errorf("key2 should not exist")
	}
}

// TestLRUCache_ContainsDoesNotUpdateRecency tests that Contains doesn't affect LRU order
func TestLRUCache_ContainsDoesNotUpdateRecency(t *testing.T) {
	l, err := NewLRU[string, int](2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Setup
	l.Add("key1", 1)
	l.Add("key2", 2)

	// Check contains but should not update recency
	l.Contains("key1")

	// Add third item, should evict key1 since it's still the least recently used
	l.Add("key3", 3)

	// Verify key1 was evicted
	if l.Contains("key1") {
		t.Errorf("key1 should have been evicted")
	}
}

// TestLRUCache_Peek tests the Peek method
func TestLRUCache_Peek(t *testing.T) {
	l, err := NewLRU[string, int](2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Setup
	l.Add("key1", 1)

	// Exercise & Verify
	v, ok := l.Peek("key1")
	if !ok {
		t.Errorf("key1 should exist")
	}
	if v != 1 {
		t.Errorf("expected 1, got: %d", v)
	}
}

// TestLRUCache_PeekDoesNotUpdateRecency tests that Peek doesn't affect LRU order
func TestLRUCache_PeekDoesNotUpdateRecency(t *testing.T) {
	l, err := NewLRU[string, int](2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Setup
	l.Add("key1", 1)
	l.Add("key2", 2)

	// Peek key1 but should not update recency
	l.Peek("key1")

	// Add third item, should evict key1 since it's still the least recently used
	l.Add("key3", 3)

	// Verify key1 was evicted
	_, ok := l.Peek("key1")
	if ok {
		t.Errorf("key1 should have been evicted")
	}
}

// TestLRUCache_Remove tests the Remove method
func TestLRUCache_Remove(t *testing.T) {
	l, err := NewLRU[string, int](2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Setup
	l.Add("key1", 1)
	l.Add("key2", 2)

	// Exercise: Remove an item
	ok := l.Remove("key1")

	// Verify
	if !ok {
		t.Errorf("expected remove to return true")
	}
	if l.Len() != 1 {
		t.Errorf("expected len 1, got: %d", l.Len())
	}

	// Try to remove the same item again
	ok = l.Remove("key1")
	if ok {
		t.Errorf("expected remove to return false")
	}
}

// TestLRUCache_Clear tests the Clear method
func TestLRUCache_Clear(t *testing.T) {
	l, err := NewLRU[string, int](2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Setup
	l.Add("key1", 1)
	l.Add("key2", 2)

	// Exercise: Clear the cache
	l.Clear()

	// Verify
	if l.Len() != 0 {
		t.Errorf("expected len 0, got: %d", l.Len())
	}
	if _, ok := l.Get("key1"); ok {
		t.Errorf("key1 should not exist")
	}
	if _, ok := l.Get("key2"); ok {
		t.Errorf("key2 should not exist")
	}
}

// TestLRUCache_Keys tests the Keys method
func TestLRUCache_Keys(t *testing.T) {
	l, err := NewLRU[string, int](3)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Setup
	l.Add("key1", 1)
	l.Add("key2", 2)
	l.Add("key3", 3)

	// Exercise
	keys := l.Keys()

	// Verify
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got: %d", len(keys))
	}

	// Keys should be ordered from least recently used to most recently used
	if keys[0] != "key1" || keys[1] != "key2" || keys[2] != "key3" {
		t.Errorf("keys in wrong order, got: %v", keys)
	}

	// Update order by getting key1
	l.Get("key1")
	keys = l.Keys()
	if keys[0] != "key2" || keys[1] != "key3" || keys[2] != "key1" {
		t.Errorf("keys in wrong order after Get, got: %v", keys)
	}
}

// TestLRUCache_RemoveOldest tests the RemoveOldest method
func TestLRUCache_RemoveOldest(t *testing.T) {
	l, err := NewLRU[string, int](2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Setup
	l.Add("key1", 1)
	l.Add("key2", 2)

	// Exercise
	key, value, ok := l.RemoveOldest()

	// Verify
	if !ok {
		t.Errorf("expected RemoveOldest to return true")
	}
	if key != "key1" {
		t.Errorf("expected key1, got: %s", key)
	}
	if value != 1 {
		t.Errorf("expected 1, got: %d", value)
	}
	if l.Len() != 1 {
		t.Errorf("expected len 1, got: %d", l.Len())
	}
}

// TestLRUCache_RemoveOldestEmpty tests RemoveOldest on an empty cache
func TestLRUCache_RemoveOldestEmpty(t *testing.T) {
	l, err := NewLRU[string, int](2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Exercise: Remove from empty cache
	key, value, ok := l.RemoveOldest()

	// Verify
	if ok {
		t.Errorf("expected RemoveOldest to return false")
	}
	if key != "" || value != 0 {
		t.Errorf("expected zero values, got key=%s value=%d", key, value)
	}
}

// TestLRUCache_GetOldest tests the GetOldest method
func TestLRUCache_GetOldest(t *testing.T) {
	l, err := NewLRU[string, int](2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Setup
	l.Add("key1", 1)
	l.Add("key2", 2)

	// Exercise
	key, value, ok := l.GetOldest()

	// Verify
	if !ok {
		t.Errorf("expected GetOldest to return true")
	}
	if key != "key1" {
		t.Errorf("expected key1, got: %s", key)
	}
	if value != 1 {
		t.Errorf("expected 1, got: %d", value)
	}
}

// TestLRUCache_GetOldestAfterUpdate tests GetOldest after updating cache
func TestLRUCache_GetOldestAfterUpdate(t *testing.T) {
	l, err := NewLRU[string, int](2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Setup
	l.Add("key1", 1)
	l.Add("key2", 2)

	// Update recency
	l.Add("key1", 11)

	// Exercise
	key, value, ok := l.GetOldest()

	// Verify
	if !ok {
		t.Errorf("expected GetOldest to return true")
	}
	if key != "key2" {
		t.Errorf("expected key2, got: %s", key)
	}
	if value != 2 {
		t.Errorf("expected 2, got: %d", value)
	}
}

// TestLRUCache_GetOldestEmpty tests GetOldest on empty cache
func TestLRUCache_GetOldestEmpty(t *testing.T) {
	l, err := NewLRU[string, int](2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Clear the cache
	l.Clear()

	// Exercise
	key, value, ok := l.GetOldest()

	// Verify
	if ok {
		t.Errorf("expected GetOldest to return false")
	}
	if key != "" || value != 0 {
		t.Errorf("expected zero values, got key=%s value=%d", key, value)
	}
}

// TestLRUCache_IntKeys tests the LRU with integer keys
func TestLRUCache_IntKeys(t *testing.T) {
	// Setup: Create cache with int keys
	intCache, _ := NewLRU[int, string](2)

	// Exercise
	intCache.Add(1, "one")
	intCache.Add(2, "two")

	// Verify
	v, ok := intCache.Get(1)
	if !ok || v != "one" {
		t.Errorf("expected to get 'one', got %v, ok=%v", v, ok)
	}
}

// TestLRUCache_StructValues tests the LRU with struct values
func TestLRUCache_StructValues(t *testing.T) {
	// Define a struct type for testing
	type Person struct {
		Name string
		Age  int
	}

	// Setup: Create cache with struct values
	structCache, _ := NewLRU[string, Person](2)

	// Exercise
	structCache.Add("alice", Person{"Alice", 30})
	structCache.Add("bob", Person{"Bob", 25})

	// Verify
	v, ok := structCache.Get("alice")
	if !ok || v.Name != "Alice" || v.Age != 30 {
		t.Errorf("expected to get Person{Alice, 30}, got %+v, ok=%v", v, ok)
	}
}

// TestLRUCache_ConcurrentAccess tests concurrent access to the LRU cache
func TestLRUCache_ConcurrentAccess(t *testing.T) {
	// Setup
	cache, _ := NewLRU[int, int](100)
	const goroutines = 10
	const itemsPerRoutine = 100

	// Create a wait group to synchronize goroutines
	var wg sync.WaitGroup
	wg.Add(goroutines * 2) // for readers and writers

	// Exercise: Launch writer goroutines
	for i := 0; i < goroutines; i++ {
		go func(base int) {
			defer wg.Done()
			for j := 0; j < itemsPerRoutine; j++ {
				key := base*itemsPerRoutine + j
				cache.Add(key, key*2)
			}
		}(i)
	}

	// Exercise: Launch reader goroutines
	for i := 0; i < goroutines; i++ {
		go func(base int) {
			defer wg.Done()
			for j := 0; j < itemsPerRoutine; j++ {
				key := base*itemsPerRoutine + j
				cache.Get(key)
				cache.Contains(key)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify: The cache should not have panicked
	// If we got here without deadlock or panic, the test passes
	if cache.Len() > 100 {
		t.Errorf("cache exceeded its capacity: %d", cache.Len())
	}
}
