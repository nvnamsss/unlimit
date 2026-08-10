package algo

import (
	"container/list"
	"sync"
)

// LRUCache is a thread-safe fixed size LRU cache with generic types
type LRUCache[K comparable, V any] struct {
	size      int
	evictList *list.List
	items     map[K]*list.Element
	lock      sync.RWMutex
}

// NewLRU creates an LRU of the given size
func NewLRU[K comparable, V any](size int) (*LRUCache[K, V], error) {
	if size <= 0 {
		return nil, ErrInvalidSize
	}
	c := &LRUCache[K, V]{
		size:      size,
		evictList: list.New(),
		items:     make(map[K]*list.Element),
	}
	return c, nil
}

// NewLRUNoError creates an LRU of the given size but ignores errors
func NewLRUNoError[K comparable, V any](size int) *LRUCache[K, V] {
	if size <= 0 {
		size = 1
	}
	c := &LRUCache[K, V]{
		size:      size,
		evictList: list.New(),
		items:     make(map[K]*list.Element),
	}
	return c
}

// Add adds a value to the cache. Returns true if an eviction occurred.
func (c *LRUCache[K, V]) Add(key K, value V) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Check for existing item
	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		ent.Value.(*Entry[K, V]).Value = value
		return false
	}

	// Add new item
	ent := &Entry[K, V]{key, value}
	element := c.evictList.PushFront(ent)
	c.items[key] = element

	// Verify size not exceeded
	evict := c.evictList.Len() > c.size
	if evict {
		c.removeOldest()
	}
	return evict
}

// Get looks up a key's value from the cache
func (c *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		return ent.Value.(*Entry[K, V]).Value, true
	}
	return value, false
}

func (c *LRUCache[K, V]) GetOldest() (key K, value V, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if ent := c.evictList.Back(); ent != nil {
		kv := ent.Value.(*Entry[K, V])
		return kv.Key, kv.Value, true
	}
	return key, value, false
}

// Contains checks if a key exists in cache without updating the recent-ness
func (c *LRUCache[K, V]) Contains(key K) (ok bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	_, ok = c.items[key]
	return ok
}

// Peek returns the key value without updating the recent-ness
func (c *LRUCache[K, V]) Peek(key K) (value V, ok bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if ent, ok := c.items[key]; ok {
		return ent.Value.(*Entry[K, V]).Value, true
	}
	return value, false
}

// Remove removes the provided key from the cache
func (c *LRUCache[K, V]) Remove(key K) (present bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if ent, ok := c.items[key]; ok {
		c.removeElement(ent)
		return true
	}
	return false
}

// RemoveOldest removes the oldest item from the cache
func (c *LRUCache[K, V]) RemoveOldest() (key K, value V, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.removeOldest()
}

// removeOldest removes the oldest item from the cache (internal, no lock)
func (c *LRUCache[K, V]) removeOldest() (key K, value V, ok bool) {
	ent := c.evictList.Back()
	if ent != nil {
		c.removeElement(ent)
		kv := ent.Value.(*Entry[K, V])
		return kv.Key, kv.Value, true
	}
	return key, value, false
}

// removeElement is used to remove a given list element from the cache
func (c *LRUCache[K, V]) removeElement(e *list.Element) {
	c.evictList.Remove(e)
	kv := e.Value.(*Entry[K, V])
	delete(c.items, kv.Key)
}

// Len returns the number of items in the cache
func (c *LRUCache[K, V]) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.evictList.Len()
}

// Clear purges all stored items from the cache
func (c *LRUCache[K, V]) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.evictList = list.New()
	c.items = make(map[K]*list.Element)
}

// Keys returns a slice of the keys in the cache, ordered from least recently
// used to most recently used
func (c *LRUCache[K, V]) Keys() []K {
	c.lock.RLock()
	defer c.lock.RUnlock()

	keys := make([]K, 0, len(c.items))
	for ent := c.evictList.Back(); ent != nil; ent = ent.Prev() {
		keys = append(keys, ent.Value.(*Entry[K, V]).Key)
	}
	return keys
}

// Error types
var (
	ErrInvalidSize = Error("invalid size")
)

// Error represents an LRU error
type Error string

func (e Error) Error() string { return string(e) }
