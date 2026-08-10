package cache

// CacheStats provides context for growth/shrink decisions.
type CacheStats struct {
	Capacity int
	Size     int
	HitRate  float64
	MissRate float64

	AllocBytes uint64  // runtime.MemStats.Alloc
	SysBytes   uint64  // runtime.MemStats.Sys
	UsedRatio  float64 // Alloc/Sys

	// For tracking thrashing prevention
	LastResizeSize int
}

// CapacityController defines the interface for adaptive cache resizing strategies.
type CapacityController interface {
	// ShouldGrow returns true if the cache should increase its capacity.
	// The parameters give context like current usage and memory stats.
	ShouldGrow(stats CacheStats) bool

	// ShouldShrink returns true if the cache should decrease its capacity.
	ShouldShrink(stats CacheStats) bool

	// NextCapacity computes the next capacity value based on the current one.
	NextCapacity(current int, grow bool) int
}

// DefaultCapacityController implements the original DynamicLFU auto-resize strategy
// - Grows by 2x when size/capacity >= loadFactor
// - Shrinks when memory usage exceeds memThreshold
type DefaultCapacityController struct {
	LoadFactor   float64 // Trigger growth when size/capacity >= loadFactor (default 0.9)
	GrowFactor   float64 // Growth multiplier (default 2.0 for 2x growth)
	ShrinkFactor float64 // Shrink to this ratio when memory pressure (default 0.5)
	MemThreshold float64 // Memory usage threshold to trigger shrink (default 0.85)
}

// NewDefaultCapacityController creates a controller with the original DynamicLFU defaults
func NewDefaultCapacityController() *DefaultCapacityController {
	return &DefaultCapacityController{
		LoadFactor:   0.9,
		GrowFactor:   2.0,
		ShrinkFactor: 0.5,
		MemThreshold: 0.85,
	}
}

func NewCapacityController(
	loadFactor float64,
	growFactor float64,
	shrinkFactor float64,
	memThreshold float64,
) CapacityController {
	return &DefaultCapacityController{
		LoadFactor:   loadFactor,
		GrowFactor:   growFactor,
		ShrinkFactor: shrinkFactor,
		MemThreshold: memThreshold,
	}
}

func (d *DefaultCapacityController) ShouldGrow(stats CacheStats) bool {
	// Grow when size/capacity >= loadFactor
	if stats.Capacity == 0 {
		return false
	}
	loadRatio := float64(stats.Size) / float64(stats.Capacity)
	return loadRatio >= d.LoadFactor
}

func (d *DefaultCapacityController) ShouldShrink(stats CacheStats) bool {
	// Shrink when memory usage exceeds threshold
	if stats.UsedRatio <= d.MemThreshold {
		return false
	}

	// Prevent thrashing: only shrink if size has decreased significantly since last resize
	if stats.LastResizeSize > 0 && stats.Size >= int(float64(stats.LastResizeSize)*0.8) {
		return false
	}

	// Calculate potential new capacity
	newCapacity := int(float64(stats.Size) / d.ShrinkFactor)

	// Only shrink if it's a significant reduction (at least 25%)
	return newCapacity < int(float64(stats.Capacity)*0.75)
}

func (d *DefaultCapacityController) NextCapacity(current int, grow bool) int {
	if grow {
		return int(float64(current) * d.GrowFactor)
	}
	// When shrinking, base it on the shrink factor
	return int(float64(current) * d.ShrinkFactor)
}

// MemoryBasedController implements a memory-pressure based resizing strategy
type MemoryBasedController struct {
	TargetRatio  float64 // e.g. 0.7 means grow if under 70% usage
	GrowFactor   float64 // e.g. 1.2 = +20%
	ShrinkFactor float64 // e.g. 0.8 = -20%
}

func (m *MemoryBasedController) ShouldGrow(stats CacheStats) bool {
	return stats.UsedRatio < m.TargetRatio*0.5
}

func (m *MemoryBasedController) ShouldShrink(stats CacheStats) bool {
	return stats.UsedRatio > m.TargetRatio
}

func (m *MemoryBasedController) NextCapacity(current int, grow bool) int {
	if grow {
		return int(float64(current) * m.GrowFactor)
	}
	return int(float64(current) * m.ShrinkFactor)
}

// LoadBasedController implements a simple load-based resizing strategy
// Ignores memory pressure, only considers cache load
type LoadBasedController struct {
	HighWaterMark float64 // Grow when size/capacity >= this (e.g., 0.9)
	LowWaterMark  float64 // Shrink when size/capacity <= this (e.g., 0.3)
	GrowFactor    float64 // Growth multiplier (e.g., 2.0)
	ShrinkFactor  float64 // Shrink multiplier (e.g., 0.5)
}

func (l *LoadBasedController) ShouldGrow(stats CacheStats) bool {
	if stats.Capacity == 0 {
		return false
	}
	loadRatio := float64(stats.Size) / float64(stats.Capacity)
	return loadRatio >= l.HighWaterMark
}

func (l *LoadBasedController) ShouldShrink(stats CacheStats) bool {
	if stats.Capacity == 0 {
		return false
	}
	loadRatio := float64(stats.Size) / float64(stats.Capacity)
	return loadRatio <= l.LowWaterMark
}

func (l *LoadBasedController) NextCapacity(current int, grow bool) int {
	if grow {
		return int(float64(current) * l.GrowFactor)
	}
	return int(float64(current) * l.ShrinkFactor)
}
