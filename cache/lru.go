package cache

import (
	"sync"
	"time"

	"github.com/voidforge-studios/unlimit/algo"
)

// lruEntry wraps a value with optional expiration time for TTL support.
type lruEntry[V any] struct {
	value      V
	expiration int64 // Unix nano timestamp, 0 means no expiration
}

// isExpired checks if the entry has expired.
func (e *lruEntry[V]) isExpired() bool {
	return e.expiration > 0 && time.Now().UnixNano() > e.expiration
}

// LRUCache implements TypedCache interface using an LRU eviction policy.
// It wraps the algo.LRUCache to provide a simpler interface for cache operations.
// The cache is thread-safe and automatically evicts the least recently used items
// when the capacity is reached.
// Example usage:
//
//	cache, _ := NewLRUCache[string, int](100)
//	cache.Set("key1", 42)
//	value, ok := cache.Get("key1")
//	fmt.Println(value) // Output: 42
//
// This implementation is useful for caching frequently accessed data with
// automatic memory management through LRU eviction.
type LRUCache[K comparable, V any] struct {
	cache      *algo.LRUCache[K, *lruEntry[V]]
	expiration map[K]int64 // Track expiration times for lazy eviction
	mu         sync.RWMutex
}

// NewLRUCache creates a new LRU cache with the specified capacity.
// The capacity determines the maximum number of items the cache can hold.
// When the capacity is exceeded, the least recently used item is evicted.
// Example usage:
//
//	cache, err := NewLRUCache[string, *User](1000)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// This function returns an error if the capacity is less than or equal to zero.
func NewLRUCache[K comparable, V any](capacity int) (TypedCacheWithExpiration[K, V], error) {
	lru, err := algo.NewLRU[K, *lruEntry[V]](capacity)
	if err != nil {
		return nil, err
	}
	return &LRUCache[K, V]{
		cache:      lru,
		expiration: make(map[K]int64),
	}, nil
}

// Get retrieves a value from the cache by key.
// It uses lazy eviction - if the entry has expired, it will be deleted and (zero-value, false) returned.
// Accessing a key updates its position in the LRU order (marks it as recently used).
// Example usage:
//
//	value, ok := cache.Get("user:123")
//	if !ok {
//	    fmt.Println("Cache miss or expired")
//	}
//
// This function is useful for retrieving cached values while maintaining LRU ordering.
func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	entry, ok := c.cache.Get(key)
	if !ok {
		var zero V
		return zero, false
	}

	// Check if expired (lazy eviction)
	if entry.isExpired() {
		c.cache.Remove(key)
		c.mu.Lock()
		delete(c.expiration, key)
		c.mu.Unlock()
		var zero V
		return zero, false
	}

	return entry.value, true
}

// Set adds or updates a key-value pair in the cache without expiration.
// It returns true if adding this item caused an eviction of the oldest item.
// If the key already exists, its value is updated and it's marked as recently used.
// Example usage:
//
//	evicted := cache.Set("session:abc", sessionData)
//	if evicted {
//	    fmt.Println("An old item was evicted")
//	}
//
// This function is useful for storing values in the cache with automatic
// capacity management through LRU eviction.
func (c *LRUCache[K, V]) Set(key K, value V) bool {
	entry := &lruEntry[V]{value: value, expiration: 0}
	evicted := c.cache.Add(key, entry)
	c.mu.Lock()
	delete(c.expiration, key) // Remove any existing expiration
	c.mu.Unlock()
	return evicted
}

// SetWithTTL adds or updates a key-value pair with a time-to-live duration.
// The entry will be lazily removed after the TTL expires (checked on next Get).
// It returns true if adding this item caused an eviction of the oldest item.
// Example usage:
//
//	evicted := cache.SetWithTTL("session:abc", session, 30*time.Minute)
//	if evicted {
//	    fmt.Println("An old item was evicted")
//	}
//
// This function is useful for caching data that should automatically expire.
func (c *LRUCache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) bool {
	var expiration int64
	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	}
	entry := &lruEntry[V]{value: value, expiration: expiration}
	evicted := c.cache.Add(key, entry)
	c.mu.Lock()
	if expiration > 0 {
		c.expiration[key] = expiration
	} else {
		delete(c.expiration, key)
	}
	c.mu.Unlock()
	return evicted
}

// Delete removes a key-value pair from the cache.
// It returns true if the key existed and was removed, false if the key was not found.
// Example usage:
//
//	deleted := cache.Delete("temp:xyz")
//	if deleted {
//	    fmt.Println("Item removed from cache")
//	}
//
// This function is useful for explicitly removing items from the cache
// before they would naturally be evicted.
func (c *LRUCache[K, V]) Delete(key K) bool {
	removed := c.cache.Remove(key)
	if removed {
		c.mu.Lock()
		delete(c.expiration, key)
		c.mu.Unlock()
	}
	return removed
}

// Len returns the current number of items in the cache.
// Note: This count may include expired entries that haven't been lazily evicted yet.
// Example usage:
//
//	count := cache.Len()
//	fmt.Printf("Cache contains %d items\n", count)
//
// This function is useful for monitoring cache usage and determining
// how full the cache is relative to its capacity.
func (c *LRUCache[K, V]) Len() int {
	return c.cache.Len()
}
