package cache

import (
	"runtime"
	"sync"

	"github.com/voidforge-studios/unlimit/algo"
)

// DynamicLFU implements a Least Frequently Used (LFU) cache with dynamic capacity
// Features:
// - O(1) Get and Set operations (amortized)
// - Pluggable capacity control strategies via CapacityController interface
// - LFU eviction policy with LRU tie-breaking
// - Generic types for keys and values
// - Thread-safe implementation using mutex
type DynamicLFU[K comparable, V any] struct {
	// Embed the base LFU cache for core functionality
	*algo.LFU[K, V]

	minCapacity int // Minimum capacity (won't shrink below this)
	maxCapacity int // Maximum capacity (won't grow beyond this)

	// Capacity control
	controller     CapacityController
	autoGrow       bool // Enable automatic growth
	autoShrink     bool // Enable automatic shrinking
	lastResizeSize int  // Track last size to prevent thrashing

	// Stats tracking for hit/miss rates
	hits   uint64
	misses uint64

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// NewDynamicLFU creates a new DynamicLFU cache with the specified initial capacity
// By default, uses DefaultCapacityController with auto-grow enabled and auto-shrink disabled
// Default min capacity is 1, max capacity is 1000x initial capacity
// Capacity must be at least 1
func NewDynamicLFU[K comparable, V any](capacity int) *DynamicLFU[K, V] {
	if capacity < 1 {
		capacity = 1
	}

	return &DynamicLFU[K, V]{
		LFU:            algo.NewLFU[K, V](capacity),
		minCapacity:    1,               // Default min is 1, not initial capacity
		maxCapacity:    capacity * 1000, // Default max is 1000x initial
		controller:     NewDefaultCapacityController(),
		autoGrow:       true,
		autoShrink:     false,
		lastResizeSize: 0,
		hits:           0,
		misses:         0,
	}
}

// SetCapacityController sets a custom capacity controller strategy
func (c *DynamicLFU[K, V]) SetCapacityController(controller CapacityController) {
	if controller != nil {
		c.controller = controller
	}
}

// EnableAutoGrow enables automatic capacity growth when cache is full
// When enabled, capacity doubles when size >= capacity * loadFactor
func (c *DynamicLFU[K, V]) EnableAutoGrow(enabled bool) {
	c.autoGrow = enabled
}

// EnableAutoShrink enables automatic capacity shrinking based on memory pressure
// When enabled, capacity reduces when memory usage exceeds memThreshold
func (c *DynamicLFU[K, V]) EnableAutoShrink(enabled bool) {
	c.autoShrink = enabled
}

// SetCapacityLimits sets the minimum and maximum capacity bounds
func (c *DynamicLFU[K, V]) SetCapacityLimits(min, max int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if min < 1 {
		min = 1
	}
	if max < min {
		max = min
	}
	c.minCapacity = min
	c.maxCapacity = max

	// Adjust current capacity if out of bounds
	currentCapacity := c.LFU.Capacity()
	if currentCapacity < c.minCapacity {
		c.LFU.Resize(c.minCapacity)
	} else if currentCapacity > c.maxCapacity {
		c.LFU.Resize(c.maxCapacity)
	}
}

// SetAutoGrowParameters configures auto-growth behavior
// loadFactor: grow when size/capacity >= loadFactor (default 0.9)
// Only works if the controller is a DefaultCapacityController
func (c *DynamicLFU[K, V]) SetAutoGrowParameters(loadFactor float64) {
	if defaultCtrl, ok := c.controller.(*DefaultCapacityController); ok {
		if loadFactor > 0 && loadFactor <= 1.0 {
			defaultCtrl.LoadFactor = loadFactor
		}
	}
}

// SetAutoShrinkParameters configures auto-shrink behavior
// shrinkFactor: shrink to this ratio when triggered (default 0.5)
// memThreshold: trigger when memory usage exceeds this (default 0.85)
// Only works if the controller is a DefaultCapacityController
func (c *DynamicLFU[K, V]) SetAutoShrinkParameters(shrinkFactor, memThreshold float64) {
	if defaultCtrl, ok := c.controller.(*DefaultCapacityController); ok {
		if shrinkFactor > 0 && shrinkFactor < 1.0 {
			defaultCtrl.ShrinkFactor = shrinkFactor
		}
		if memThreshold > 0 && memThreshold <= 1.0 {
			defaultCtrl.MemThreshold = memThreshold
		}
	}
}

// Get retrieves a value from the cache by key
// Returns (value, true) if found, (zero-value, false) if not found
// Time complexity: O(1) amortized
func (c *DynamicLFU[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	value, exists := c.LFU.Get(key)
	if !exists {
		c.misses++
		return value, false
	}

	c.hits++
	return value, true
}

// Set stores a value in the cache with the specified key
// If the key exists, updates the value and increments frequency
// If the cache is full and auto-grow is enabled, uses the controller to determine growth
// If auto-grow is disabled, evicts the least frequently used item (LRU among ties)
// Time complexity: O(1) amortized
// Returns true on success
func (c *DynamicLFU[K, V]) Set(key K, value V) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to grow capacity
	if c.autoGrow {
		c.tryAutoGrow()
	}

	// Check if we need to shrink due to memory pressure
	if c.autoShrink {
		c.tryAutoShrink()
	}

	return c.LFU.Set(key, value)
}

// getStats returns current cache statistics for the controller
func (c *DynamicLFU[K, V]) getStats() CacheStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	usedRatio := float64(0)
	if m.Sys > 0 {
		usedRatio = float64(m.Alloc) / float64(m.Sys)
	}

	totalAccess := c.hits + c.misses
	hitRate := float64(0)
	missRate := float64(0)
	if totalAccess > 0 {
		hitRate = float64(c.hits) / float64(totalAccess)
		missRate = float64(c.misses) / float64(totalAccess)
	}

	return CacheStats{
		Capacity:       c.LFU.Capacity(),
		Size:           c.LFU.Len(),
		HitRate:        hitRate,
		MissRate:       missRate,
		AllocBytes:     m.Alloc,
		SysBytes:       m.Sys,
		UsedRatio:      usedRatio,
		LastResizeSize: c.lastResizeSize,
	}
}

// tryAutoGrow attempts to grow the cache capacity using the controller
func (c *DynamicLFU[K, V]) tryAutoGrow() {
	stats := c.getStats()

	if !c.controller.ShouldGrow(stats) {
		return
	}

	currentCapacity := c.LFU.Capacity()
	newCapacity := c.controller.NextCapacity(currentCapacity, true)

	// Don't exceed max capacity
	if newCapacity > c.maxCapacity {
		newCapacity = c.maxCapacity
	}

	// Only grow if we can increase capacity
	if newCapacity > currentCapacity {
		c.LFU.Resize(newCapacity)
		c.lastResizeSize = c.LFU.Len()
	}
}

// tryAutoShrink attempts to shrink the cache using the controller
func (c *DynamicLFU[K, V]) tryAutoShrink() {
	stats := c.getStats()

	if !c.controller.ShouldShrink(stats) {
		return
	}

	currentCapacity := c.LFU.Capacity()
	newCapacity := c.controller.NextCapacity(currentCapacity, false)

	// Don't go below min capacity
	if newCapacity < c.minCapacity {
		newCapacity = c.minCapacity
	}

	// Only shrink if it's a meaningful reduction
	if newCapacity < currentCapacity {
		c.Resize(newCapacity)
		c.lastResizeSize = c.LFU.Len()
	}
}

// Delete removes a key from the cache
// Returns true if the key was found and deleted, false otherwise
// Time complexity: O(1)
func (c *DynamicLFU[K, V]) Delete(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.LFU.Delete(key)
}

// Len returns the number of items currently in the cache
// Time complexity: O(1)
func (c *DynamicLFU[K, V]) Len() int {
	return c.LFU.Len()
}

// Capacity returns the current capacity of the cache
// Time complexity: O(1)
func (c *DynamicLFU[K, V]) Capacity() int {
	return c.LFU.Capacity()
}

// Resize changes the cache capacity
// If newCapacity < current size, evicts least frequently used items until size <= newCapacity
// If newCapacity < 1, sets capacity to 1
// Respects min and max capacity limits
// Time complexity: O(evictions) where evictions = max(0, currentSize - newCapacity)
func (c *DynamicLFU[K, V]) Resize(newCapacity int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if newCapacity < 1 {
		newCapacity = 1
	}

	// Enforce capacity limits
	if newCapacity < c.minCapacity {
		newCapacity = c.minCapacity
	}
	if newCapacity > c.maxCapacity {
		newCapacity = c.maxCapacity
	}

	c.LFU.Resize(newCapacity)
}

// Clear removes all items from the cache
// Time complexity: O(1) - just creates new maps
func (c *DynamicLFU[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LFU.Clear()
	c.hits = 0
	c.misses = 0
}
