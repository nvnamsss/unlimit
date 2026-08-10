package cache

import (
	"sync"
	"time"
)

// hashtableEntry holds a value and its optional expiration time.
type hashtableEntry[V any] struct {
	value      V
	expiration int64 // Unix nano timestamp, 0 means no expiration
}

// isExpired checks if the entry has expired.
func (e *hashtableEntry[V]) isExpired() bool {
	return e.expiration > 0 && time.Now().UnixNano() > e.expiration
}

// Hashtable is a thread-safe generic hash table implementation with optional TTL support.
// It provides a concurrent-safe, in-memory key-value store using read-write mutex for synchronization.
// Multiple readers can access the hashtable simultaneously, while writers have exclusive access.
// Expiration is handled lazily - expired entries are only removed when accessed.
// Example usage:
//
//	cache := NewHashtable[string, *User]()
//	cache.Set("user:123", &User{ID: 123, Name: "Alice"})
//	cache.SetWithTTL("session:abc", session, 30*time.Minute)
//	user, ok := cache.Get("user:123")
//
// This implementation is useful for concurrent access to shared data with high read throughput.
type Hashtable[K comparable, V any] struct {
	items map[K]*hashtableEntry[V]
	mu    sync.RWMutex
}

// NewHashtable creates a new empty thread-safe hashtable.
// Example usage:
//
//	cache := NewHashtable[int, string]()
//	cache.Set(1, "one")
//
// This function is useful for creating a concurrent-safe key-value store
// without pre-allocating capacity.
func NewHashtable[K comparable, V any]() *Hashtable[K, V] {
	return &Hashtable[K, V]{
		items: make(map[K]*hashtableEntry[V]),
	}
}

// NewHashtableWithCapacity creates a new thread-safe hashtable with a specified initial capacity.
// Pre-allocating capacity can improve performance when the expected size is known.
// Example usage:
//
//	cache := NewHashtableWithCapacity[string, int](1000)
//
// This function is useful for reducing memory allocations when you know
// approximately how many items will be stored.
func NewHashtableWithCapacity[K comparable, V any](capacity int) *Hashtable[K, V] {
	return &Hashtable[K, V]{
		items: make(map[K]*hashtableEntry[V], capacity),
	}
}

// Get retrieves a value from the hashtable by key in a thread-safe manner.
// It uses lazy eviction - if the entry has expired, it will be deleted and (zero-value, false) returned.
// Returns (value, true) if found and not expired, (zero-value, false) if not found or expired.
// Example usage:
//
//	value, ok := cache.Get("user:123")
//	if !ok {
//	    fmt.Println("Key not found or expired")
//	}
//
// This function is useful for safe concurrent reads from the hashtable.
func (h *Hashtable[K, V]) Get(key K) (V, bool) {
	h.mu.RLock()
	e, exists := h.items[key]
	if !exists {
		h.mu.RUnlock()
		var zero V
		return zero, false
	}

	// Check expiration
	if e.isExpired() {
		h.mu.RUnlock()
		// Upgrade to write lock to delete expired entry
		h.mu.Lock()
		// Double-check after acquiring write lock
		if e, exists := h.items[key]; exists && e.isExpired() {
			delete(h.items, key)
		}
		h.mu.Unlock()
		var zero V
		return zero, false
	}

	value := e.value
	h.mu.RUnlock()
	return value, true
}

// Set stores a value in the hashtable with the specified key in a thread-safe manner.
// It uses a write lock, ensuring exclusive access during the update.
// If the key exists, updates the value. The entry will not expire.
// Example usage:
//
//	cache.Set("session:abc", sessionData)
//	cache.Set("counter", 42) // Updates if key exists
//
// This function is useful for safe concurrent writes to the hashtable.
// Always returns true to indicate successful storage.
func (h *Hashtable[K, V]) Set(key K, value V) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.items[key] = &hashtableEntry[V]{value: value, expiration: 0}
	return true
}

// SetWithTTL stores a value with a time-to-live duration in a thread-safe manner.
// The entry will be lazily removed after the TTL expires (checked on next Get).
// Example usage:
//
//	cache.SetWithTTL("session:abc", session, 30*time.Minute)
//	cache.SetWithTTL("rate:user:123", count, 1*time.Hour)
//
// This function is useful for caching data that should automatically expire.
// Always returns false as this implementation has no capacity limit.
func (h *Hashtable[K, V]) SetWithTTL(key K, value V, ttl time.Duration) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	var expiration int64
	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	}
	h.items[key] = &hashtableEntry[V]{value: value, expiration: expiration}
	return false
}

// Delete removes a key from the hashtable in a thread-safe manner.
// It uses a write lock, ensuring exclusive access during the deletion.
// Returns true if the key was found and removed, false otherwise.
// Example usage:
//
//	if cache.Delete("temp:xyz") {
//	    fmt.Println("Item removed")
//	}
//
// This function is useful for safe concurrent deletions from the hashtable.
func (h *Hashtable[K, V]) Delete(key K) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, exists := h.items[key]; exists {
		delete(h.items, key)
		return true
	}
	return false
}

// Len returns the number of items currently in the hashtable in a thread-safe manner.
// It uses a read lock, allowing concurrent length checks.
// Note: This count may include expired entries that haven't been lazily evicted yet.
// Example usage:
//
//	count := cache.Len()
//	fmt.Printf("Cache contains %d items\n", count)
//
// This function is useful for monitoring hashtable size safely during concurrent access.
func (h *Hashtable[K, V]) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.items)
}

// Clear removes all items from the hashtable in a thread-safe manner.
// It uses a write lock, ensuring exclusive access during the clear operation.
// Example usage:
//
//	cache.Clear()
//	fmt.Println("All items removed")
//
// This function is useful for resetting the hashtable safely during concurrent access.
func (h *Hashtable[K, V]) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.items = make(map[K]*hashtableEntry[V])
}

// Has checks if a key exists in the hashtable in a thread-safe manner.
// It uses lazy eviction - if the entry has expired, it will be deleted and false returned.
// Example usage:
//
//	if cache.Has("user:123") {
//	    fmt.Println("User exists in cache")
//	}
//
// This function is useful for checking key existence safely during concurrent access.
func (h *Hashtable[K, V]) Has(key K) bool {
	h.mu.RLock()
	e, exists := h.items[key]
	if !exists {
		h.mu.RUnlock()
		return false
	}

	if e.isExpired() {
		h.mu.RUnlock()
		// Upgrade to write lock to delete expired entry
		h.mu.Lock()
		if e, exists := h.items[key]; exists && e.isExpired() {
			delete(h.items, key)
		}
		h.mu.Unlock()
		return false
	}

	h.mu.RUnlock()
	return true
}

// Keys returns a slice of all non-expired keys in the hashtable in a thread-safe manner.
// It uses a read lock, creating a snapshot of keys at the time of the call.
// Expired entries are skipped but not deleted (lazy eviction only on Get/Has).
// The iteration order is non-deterministic.
// Example usage:
//
//	keys := cache.Keys()
//	for _, key := range keys {
//	    fmt.Println(key)
//	}
//
// This function is useful for iterating over all keys safely during concurrent access.
func (h *Hashtable[K, V]) Keys() []K {
	h.mu.RLock()
	defer h.mu.RUnlock()
	keys := make([]K, 0, len(h.items))
	for key, e := range h.items {
		if !e.isExpired() {
			keys = append(keys, key)
		}
	}
	return keys
}

// Values returns a slice of all non-expired values in the hashtable in a thread-safe manner.
// It uses a read lock, creating a snapshot of values at the time of the call.
// Expired entries are skipped but not deleted (lazy eviction only on Get/Has).
// The iteration order is non-deterministic.
// Example usage:
//
//	values := cache.Values()
//	for _, value := range values {
//	    fmt.Println(value)
//	}
//
// This function is useful for iterating over all values safely during concurrent access.
func (h *Hashtable[K, V]) Values() []V {
	h.mu.RLock()
	defer h.mu.RUnlock()
	values := make([]V, 0, len(h.items))
	for _, e := range h.items {
		if !e.isExpired() {
			values = append(values, e.value)
		}
	}
	return values
}

// ForEach iterates over all non-expired key-value pairs in the hashtable in a thread-safe manner.
// It uses a read lock for the entire iteration, blocking writes until iteration completes.
// Expired entries are skipped but not deleted (lazy eviction only on Get/Has).
// The iteration order is non-deterministic.
// Example usage:
//
//	cache.ForEach(func(key string, value int) {
//	    fmt.Printf("%s: %d\n", key, value)
//	})
//
// This function is useful for processing all items safely, but note that the
// read lock is held during the entire callback execution.
func (h *Hashtable[K, V]) ForEach(fn func(key K, value V)) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for key, e := range h.items {
		if !e.isExpired() {
			fn(key, e.value)
		}
	}
}

// Clone creates a shallow copy of the hashtable in a thread-safe manner.
// It uses a read lock to create a snapshot of the current state.
// Only non-expired entries are copied to the new hashtable.
// The returned hashtable is a new independent instance.
// Example usage:
//
//	clone := cache.Clone()
//	clone.Set("new", value) // Does not affect original
//
// This function is useful for creating snapshots or backups of the hashtable.
// Note: this is a shallow copy; if values are pointers, they point to the same objects.
func (h *Hashtable[K, V]) Clone() *Hashtable[K, V] {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clone := &Hashtable[K, V]{
		items: make(map[K]*hashtableEntry[V], len(h.items)),
	}
	for key, e := range h.items {
		if !e.isExpired() {
			clone.items[key] = &hashtableEntry[V]{value: e.value, expiration: e.expiration}
		}
	}
	return clone
}
