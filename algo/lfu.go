package algo

import (
	"container/list"
	"sync"
)

// LFU implements a Least Frequently Used (LFU) cache
// Items with the lowest access frequency are evicted when capacity is reached
// Thread-safe implementation using mutex
type LFU[K comparable, V any] struct {
	capacity int
	minFreq  int
	mu       sync.RWMutex

	// items maps key to cache entry
	items map[K]*lfuEntry[K, V]

	// freqList maps frequency to list of items with that frequency
	freqList map[int]*list.List
}

// lfuEntry represents a single cache entry
type lfuEntry[K comparable, V any] struct {
	key   K
	value V
	freq  int
	elem  *list.Element // pointer to element in frequency list
}

// NewLFU creates a new LFU cache with the specified capacity
// Returns error if capacity is less than 1
func NewLFU[K comparable, V any](capacity int) *LFU[K, V] {
	if capacity < 1 {
		capacity = 1
	}

	return &LFU[K, V]{
		capacity: capacity,
		minFreq:  0,
		items:    make(map[K]*lfuEntry[K, V]),
		freqList: make(map[int]*list.List),
	}
}

// Get retrieves a value from the cache
// Returns (value, true) if found, (zero-value, false) if not found
// Increments the frequency count for the accessed item
func (c *LFU[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.items[key]
	if !exists {
		var zero V
		return zero, false
	}

	// Increment frequency
	c.incrementFreq(entry)
	return entry.value, true
}

// Set stores a value in the cache with the specified key
// If the key exists, updates the value and increments frequency
// If the cache is full, evicts the least frequently used item
// Returns true on success
func (c *LFU[K, V]) Set(key K, value V) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update existing entry
	if entry, exists := c.items[key]; exists {
		entry.value = value
		c.incrementFreq(entry)
		return true
	}

	// Evict if at capacity
	if len(c.items) >= c.capacity {
		c.evict()
	}

	// Add new entry with frequency 1
	entry := &lfuEntry[K, V]{
		key:   key,
		value: value,
		freq:  1,
	}

	// Create frequency list if it doesn't exist
	if c.freqList[1] == nil {
		c.freqList[1] = list.New()
	}

	entry.elem = c.freqList[1].PushFront(entry)
	c.items[key] = entry
	c.minFreq = 1
	return true
}

// Delete removes a key from the cache
// Returns true if the key was found and removed, false otherwise
func (c *LFU[K, V]) Delete(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.items[key]
	if !exists {
		return false
	}

	c.removeEntry(entry)
	return true
}

// incrementFreq increments the frequency of an entry and moves it to the appropriate list
func (c *LFU[K, V]) incrementFreq(entry *lfuEntry[K, V]) {
	freq := entry.freq

	// Remove from current frequency list
	c.freqList[freq].Remove(entry.elem)

	// If current frequency list is empty and it's the minimum, update minFreq
	if c.freqList[freq].Len() == 0 {
		delete(c.freqList, freq)
		if c.minFreq == freq {
			c.minFreq++
		}
	}

	// Increment frequency
	entry.freq++
	newFreq := entry.freq

	// Create new frequency list if needed
	if c.freqList[newFreq] == nil {
		c.freqList[newFreq] = list.New()
	}

	// Add to new frequency list
	entry.elem = c.freqList[newFreq].PushFront(entry)
}

// evict removes the least frequently used item from the cache
func (c *LFU[K, V]) evict() {
	if len(c.items) == 0 {
		return
	}

	// Get the list with minimum frequency
	minList := c.freqList[c.minFreq]
	if minList == nil || minList.Len() == 0 {
		return
	}

	// Remove the least recently used item within the minimum frequency
	// (items are added to the front, so back is least recent)
	elem := minList.Back()
	if elem != nil {
		entry := elem.Value.(*lfuEntry[K, V])
		c.removeEntry(entry)
	}
}

// removeEntry removes an entry from the cache
func (c *LFU[K, V]) removeEntry(entry *lfuEntry[K, V]) {
	// Remove from frequency list
	c.freqList[entry.freq].Remove(entry.elem)

	// Clean up empty frequency list
	if c.freqList[entry.freq].Len() == 0 {
		delete(c.freqList, entry.freq)
	}

	// Remove from items map
	delete(c.items, entry.key)
}

func (c *LFU[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Capacity returns the current capacity of the cache
func (c *LFU[K, V]) Capacity() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.capacity
}

// Resize changes the cache capacity
// If newCapacity < current size, evicts least frequently used items until size <= newCapacity
// If newCapacity < 1, sets capacity to 1
func (c *LFU[K, V]) Resize(newCapacity int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if newCapacity < 1 {
		newCapacity = 1
	}

	c.capacity = newCapacity

	// Evict items if new capacity is smaller than current size
	for len(c.items) > c.capacity {
		c.evict()
	}
}

// Clear removes all items from the cache
func (c *LFU[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[K]*lfuEntry[K, V])
	c.freqList = make(map[int]*list.List)
	c.minFreq = 0
}
