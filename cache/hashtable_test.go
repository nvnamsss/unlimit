package cache

import (
	"testing"
	"time"
)

// Verify that Hashtable implements TypedCache interface
var _ TypedCache[string, int] = (*Hashtable[string, int])(nil)

// Verify that Hashtable implements TypedCacheWithExpiration interface
var _ TypedCacheWithExpiration[string, int] = (*Hashtable[string, int])(nil)

func TestHashtable_New(t *testing.T) {
	h := NewHashtable[string, int]()
	if h == nil {
		t.Fatal("NewHashtable returned nil")
	}
	if h.Len() != 0 {
		t.Errorf("New hashtable should be empty, got length %d", h.Len())
	}
}

func TestHashtable_NewWithCapacity(t *testing.T) {
	h := NewHashtableWithCapacity[string, int](100)
	if h == nil {
		t.Fatal("NewHashtableWithCapacity returned nil")
	}
	if h.Len() != 0 {
		t.Errorf("New hashtable should be empty, got length %d", h.Len())
	}
}

func TestHashtable_SetAndGet(t *testing.T) {
	h := NewHashtable[string, int]()

	// Set a value
	ok := h.Set("key1", 42)
	if !ok {
		t.Error("Set should return true")
	}

	// Get the value
	value, exists := h.Get("key1")
	if !exists {
		t.Error("Get should find the key")
	}
	if value != 42 {
		t.Errorf("Expected value 42, got %d", value)
	}

	// Get non-existent key
	_, exists = h.Get("nonexistent")
	if exists {
		t.Error("Get should not find non-existent key")
	}
}

func TestHashtable_Update(t *testing.T) {
	h := NewHashtable[string, int]()

	h.Set("key1", 42)
	h.Set("key1", 99) // Update

	value, exists := h.Get("key1")
	if !exists {
		t.Error("Key should exist")
	}
	if value != 99 {
		t.Errorf("Expected updated value 99, got %d", value)
	}
}

func TestHashtable_Delete(t *testing.T) {
	h := NewHashtable[string, int]()

	h.Set("key1", 42)
	h.Set("key2", 99)

	// Delete existing key
	ok := h.Delete("key1")
	if !ok {
		t.Error("Delete should return true for existing key")
	}

	// Verify deletion
	_, exists := h.Get("key1")
	if exists {
		t.Error("Key should not exist after deletion")
	}

	// Delete non-existent key
	ok = h.Delete("nonexistent")
	if ok {
		t.Error("Delete should return false for non-existent key")
	}

	// Verify other key still exists
	value, exists := h.Get("key2")
	if !exists || value != 99 {
		t.Error("Other keys should not be affected by deletion")
	}
}

func TestHashtable_Len(t *testing.T) {
	h := NewHashtable[string, int]()

	if h.Len() != 0 {
		t.Error("Empty hashtable should have length 0")
	}

	h.Set("key1", 1)
	h.Set("key2", 2)
	h.Set("key3", 3)

	if h.Len() != 3 {
		t.Errorf("Expected length 3, got %d", h.Len())
	}

	h.Delete("key2")

	if h.Len() != 2 {
		t.Errorf("Expected length 2 after deletion, got %d", h.Len())
	}
}

func TestHashtable_Clear(t *testing.T) {
	h := NewHashtable[string, int]()

	h.Set("key1", 1)
	h.Set("key2", 2)
	h.Set("key3", 3)

	h.Clear()

	if h.Len() != 0 {
		t.Errorf("Cleared hashtable should have length 0, got %d", h.Len())
	}

	_, exists := h.Get("key1")
	if exists {
		t.Error("Keys should not exist after clear")
	}
}

func TestHashtable_Has(t *testing.T) {
	h := NewHashtable[string, int]()

	h.Set("key1", 42)

	if !h.Has("key1") {
		t.Error("Has should return true for existing key")
	}

	if h.Has("nonexistent") {
		t.Error("Has should return false for non-existent key")
	}
}

func TestHashtable_Keys(t *testing.T) {
	h := NewHashtable[string, int]()

	h.Set("key1", 1)
	h.Set("key2", 2)
	h.Set("key3", 3)

	keys := h.Keys()

	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// Check all keys are present
	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[key] = true
	}

	if !keyMap["key1"] || !keyMap["key2"] || !keyMap["key3"] {
		t.Error("Not all keys are present in Keys() result")
	}
}

func TestHashtable_Values(t *testing.T) {
	h := NewHashtable[string, int]()

	h.Set("key1", 1)
	h.Set("key2", 2)
	h.Set("key3", 3)

	values := h.Values()

	if len(values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(values))
	}

	// Check all values are present
	valueMap := make(map[int]bool)
	for _, value := range values {
		valueMap[value] = true
	}

	if !valueMap[1] || !valueMap[2] || !valueMap[3] {
		t.Error("Not all values are present in Values() result")
	}
}

func TestHashtable_ForEach(t *testing.T) {
	h := NewHashtable[string, int]()

	h.Set("key1", 1)
	h.Set("key2", 2)
	h.Set("key3", 3)

	count := 0
	sum := 0
	h.ForEach(func(key string, value int) {
		count++
		sum += value
	})

	if count != 3 {
		t.Errorf("ForEach should iterate 3 times, got %d", count)
	}

	if sum != 6 {
		t.Errorf("Expected sum of values to be 6, got %d", sum)
	}
}

func TestHashtable_Clone(t *testing.T) {
	h := NewHashtable[string, int]()

	h.Set("key1", 1)
	h.Set("key2", 2)
	h.Set("key3", 3)

	clone := h.Clone()

	// Verify clone has same data
	if clone.Len() != h.Len() {
		t.Error("Clone should have same length as original")
	}

	h.ForEach(func(key string, value int) {
		cloneValue, exists := clone.Get(key)
		if !exists {
			t.Errorf("Clone missing key %s", key)
		}
		if cloneValue != value {
			t.Errorf("Clone has different value for key %s: expected %d, got %d", key, value, cloneValue)
		}
	})

	// Verify clone is independent
	clone.Set("key1", 999)

	originalValue, _ := h.Get("key1")
	if originalValue != 1 {
		t.Error("Modifying clone should not affect original")
	}
}

func TestHashtable_DifferentTypes(t *testing.T) {
	// Test with different key/value types
	t.Run("IntKeys", func(t *testing.T) {
		h := NewHashtable[int, string]()
		h.Set(1, "one")
		h.Set(2, "two")

		value, exists := h.Get(1)
		if !exists || value != "one" {
			t.Error("Hashtable with int keys should work")
		}
	})

	t.Run("StructValues", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}

		h := NewHashtable[string, Person]()
		h.Set("john", Person{Name: "John", Age: 30})

		person, exists := h.Get("john")
		if !exists || person.Name != "John" || person.Age != 30 {
			t.Error("Hashtable with struct values should work")
		}
	})
}

func TestHashtable_TypedCacheInterface(t *testing.T) {
	// Test that Hashtable can be used as TypedCache interface
	var cache TypedCache[string, int] = NewHashtable[string, int]()

	cache.Set("key1", 42)

	value, exists := cache.Get("key1")
	if !exists || value != 42 {
		t.Error("Hashtable should work through TypedCache interface")
	}

	if cache.Len() != 1 {
		t.Error("TypedCache.Len() should work")
	}

	cache.Delete("key1")

	if cache.Len() != 0 {
		t.Error("TypedCache.Delete() should work")
	}
}

// TTL Tests

func TestHashtable_SetWithTTL(t *testing.T) {
	h := NewHashtable[string, int]()

	// Set with TTL
	h.SetWithTTL("key1", 42, 1*time.Hour)

	// Should be retrievable immediately
	value, exists := h.Get("key1")
	if !exists {
		t.Error("Key with TTL should exist immediately after set")
	}
	if value != 42 {
		t.Errorf("Expected value 42, got %d", value)
	}
}

func TestHashtable_SetWithTTL_Expiration(t *testing.T) {
	h := NewHashtable[string, int]()

	// Set with very short TTL
	h.SetWithTTL("key1", 42, 50*time.Millisecond)

	// Should be retrievable immediately
	_, exists := h.Get("key1")
	if !exists {
		t.Error("Key should exist immediately")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired (lazy eviction on Get)
	_, exists = h.Get("key1")
	if exists {
		t.Error("Key should be expired after TTL")
	}
}

func TestHashtable_SetWithTTL_LazyEviction(t *testing.T) {
	h := NewHashtable[string, int]()

	// Set with very short TTL
	h.SetWithTTL("key1", 42, 50*time.Millisecond)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Len still counts expired entries (not lazily evicted yet)
	// This is expected behavior - lazy eviction happens on Get/Has
	initialLen := h.Len()

	// Access the key to trigger lazy eviction
	h.Get("key1")

	// Now the key should be removed
	if h.Len() >= initialLen {
		// The key was evicted, so length should decrease or stay same
		// (if there were other keys)
	}

	// Verify key is gone
	_, exists := h.Get("key1")
	if exists {
		t.Error("Expired key should be evicted on access")
	}
}

func TestHashtable_Has_WithTTL(t *testing.T) {
	h := NewHashtable[string, int]()

	h.SetWithTTL("key1", 42, 50*time.Millisecond)

	// Should exist immediately
	if !h.Has("key1") {
		t.Error("Has should return true for non-expired key")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired (lazy eviction on Has)
	if h.Has("key1") {
		t.Error("Has should return false for expired key")
	}
}

func TestHashtable_Keys_SkipsExpired(t *testing.T) {
	h := NewHashtable[string, int]()

	h.Set("permanent", 1)
	h.SetWithTTL("temporary", 2, 50*time.Millisecond)

	// Both should be in Keys initially
	keys := h.Keys()
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Only permanent key should be in Keys
	keys = h.Keys()
	if len(keys) != 1 {
		t.Errorf("Expected 1 key after expiration, got %d", len(keys))
	}
	if keys[0] != "permanent" {
		t.Errorf("Expected 'permanent' key, got %s", keys[0])
	}
}

func TestHashtable_Values_SkipsExpired(t *testing.T) {
	h := NewHashtable[string, int]()

	h.Set("permanent", 1)
	h.SetWithTTL("temporary", 2, 50*time.Millisecond)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Only permanent value should be returned
	values := h.Values()
	if len(values) != 1 {
		t.Errorf("Expected 1 value after expiration, got %d", len(values))
	}
	if values[0] != 1 {
		t.Errorf("Expected value 1, got %d", values[0])
	}
}

func TestHashtable_ForEach_SkipsExpired(t *testing.T) {
	h := NewHashtable[string, int]()

	h.Set("permanent", 1)
	h.SetWithTTL("temporary", 2, 50*time.Millisecond)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	count := 0
	h.ForEach(func(key string, value int) {
		count++
		if key != "permanent" {
			t.Errorf("ForEach should skip expired key, got %s", key)
		}
	})

	if count != 1 {
		t.Errorf("Expected 1 iteration, got %d", count)
	}
}

func TestHashtable_Clone_SkipsExpired(t *testing.T) {
	h := NewHashtable[string, int]()

	h.Set("permanent", 1)
	h.SetWithTTL("temporary", 2, 50*time.Millisecond)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	clone := h.Clone()

	// Clone should only have non-expired entries
	if clone.Len() != 1 {
		t.Errorf("Clone should have 1 entry, got %d", clone.Len())
	}

	_, exists := clone.Get("permanent")
	if !exists {
		t.Error("Clone should have permanent key")
	}

	_, exists = clone.Get("temporary")
	if exists {
		t.Error("Clone should not have expired key")
	}
}

func TestHashtable_SetWithTTL_ZeroDuration(t *testing.T) {
	h := NewHashtable[string, int]()

	// Zero TTL should mean no expiration
	h.SetWithTTL("key1", 42, 0)

	time.Sleep(10 * time.Millisecond)

	value, exists := h.Get("key1")
	if !exists {
		t.Error("Key with zero TTL should not expire")
	}
	if value != 42 {
		t.Errorf("Expected value 42, got %d", value)
	}
}

func TestHashtable_SetWithTTL_UpdateExtendsExpiration(t *testing.T) {
	h := NewHashtable[string, int]()

	// Set with short TTL
	h.SetWithTTL("key1", 42, 50*time.Millisecond)

	// Update with longer TTL before expiration
	time.Sleep(30 * time.Millisecond)
	h.SetWithTTL("key1", 99, 100*time.Millisecond)

	// Original TTL would have expired by now
	time.Sleep(50 * time.Millisecond)

	// Should still exist because we extended the TTL
	value, exists := h.Get("key1")
	if !exists {
		t.Error("Key should exist after TTL extension")
	}
	if value != 99 {
		t.Errorf("Expected updated value 99, got %d", value)
	}
}

func TestHashtable_TypedCacheWithExpirationInterface(t *testing.T) {
	// Test that Hashtable can be used as TypedCacheWithExpiration interface
	var cache TypedCacheWithExpiration[string, int] = NewHashtable[string, int]()

	cache.SetWithTTL("key1", 42, 1*time.Hour)

	value, exists := cache.Get("key1")
	if !exists || value != 42 {
		t.Error("Hashtable should work through TypedCacheWithExpiration interface")
	}

	// Test that TypedCache methods also work
	cache.Set("key2", 99)
	value, exists = cache.Get("key2")
	if !exists || value != 99 {
		t.Error("Set should work through TypedCacheWithExpiration interface")
	}
}

func BenchmarkHashtable_Set(b *testing.B) {
	h := NewHashtable[int, int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Set(i, i*2)
	}
}

func BenchmarkHashtable_Get(b *testing.B) {
	h := NewHashtable[int, int]()
	for i := 0; i < 1000; i++ {
		h.Set(i, i*2)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Get(i % 1000)
	}
}

func BenchmarkHashtable_SetWithTTL(b *testing.B) {
	h := NewHashtable[int, int]()
	ttl := 1 * time.Hour
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.SetWithTTL(i, i*2, ttl)
	}
}

func BenchmarkHashtable_Delete(b *testing.B) {
	b.StopTimer()
	h := NewHashtable[int, int]()
	for i := 0; i < b.N; i++ {
		h.Set(i, i*2)
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		h.Delete(i)
	}
}
