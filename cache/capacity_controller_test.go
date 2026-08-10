package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCapacityController_DefaultController(t *testing.T) {
	t.Run("DefaultControllerShouldGrow", func(t *testing.T) {
		ctrl := NewDefaultCapacityController()

		// Should grow when load >= 0.9
		stats := CacheStats{
			Capacity: 10,
			Size:     9,
		}
		assert.True(t, ctrl.ShouldGrow(stats))

		// Should not grow when load < 0.9
		stats.Size = 8
		assert.False(t, ctrl.ShouldGrow(stats))
	})

	t.Run("DefaultControllerShouldShrink", func(t *testing.T) {
		ctrl := NewDefaultCapacityController()

		// Should shrink when memory usage > 0.85
		stats := CacheStats{
			Capacity:  100,
			Size:      20,
			UsedRatio: 0.9, // > 0.85 threshold
		}
		assert.True(t, ctrl.ShouldShrink(stats))

		// Should not shrink when memory usage <= 0.85
		stats.UsedRatio = 0.8
		assert.False(t, ctrl.ShouldShrink(stats))
	})

	t.Run("DefaultControllerNextCapacity", func(t *testing.T) {
		ctrl := NewDefaultCapacityController()

		// Growth: 2x
		newCap := ctrl.NextCapacity(10, true)
		assert.Equal(t, 20, newCap)

		// Shrink: 0.5x
		newCap = ctrl.NextCapacity(100, false)
		assert.Equal(t, 50, newCap)
	})
}

func TestCapacityController_LoadBasedController(t *testing.T) {
	t.Run("LoadBasedGrow", func(t *testing.T) {
		ctrl := &LoadBasedController{
			HighWaterMark: 0.8,
			LowWaterMark:  0.3,
			GrowFactor:    1.5,
			ShrinkFactor:  0.7,
		}

		// Should grow when load >= high water mark
		stats := CacheStats{
			Capacity: 10,
			Size:     8,
		}
		assert.True(t, ctrl.ShouldGrow(stats))

		stats.Size = 7
		assert.False(t, ctrl.ShouldGrow(stats))
	})

	t.Run("LoadBasedShrink", func(t *testing.T) {
		ctrl := &LoadBasedController{
			HighWaterMark: 0.8,
			LowWaterMark:  0.3,
			GrowFactor:    1.5,
			ShrinkFactor:  0.7,
		}

		// Should shrink when load <= low water mark
		stats := CacheStats{
			Capacity: 100,
			Size:     30,
		}
		assert.True(t, ctrl.ShouldShrink(stats))

		stats.Size = 31
		assert.False(t, ctrl.ShouldShrink(stats))
	})

	t.Run("LoadBasedNextCapacity", func(t *testing.T) {
		ctrl := &LoadBasedController{
			HighWaterMark: 0.8,
			LowWaterMark:  0.3,
			GrowFactor:    1.5,
			ShrinkFactor:  0.7,
		}

		// Growth: 1.5x
		newCap := ctrl.NextCapacity(10, true)
		assert.Equal(t, 15, newCap)

		// Shrink: 0.7x
		newCap = ctrl.NextCapacity(100, false)
		assert.Equal(t, 70, newCap)
	})
}

func TestCapacityController_MemoryBasedController(t *testing.T) {
	t.Run("MemoryBasedGrow", func(t *testing.T) {
		ctrl := &MemoryBasedController{
			TargetRatio:  0.7,
			GrowFactor:   1.2,
			ShrinkFactor: 0.8,
		}

		// Should grow when memory usage < target * 0.5 (0.35)
		stats := CacheStats{
			UsedRatio: 0.3,
		}
		assert.True(t, ctrl.ShouldGrow(stats))

		stats.UsedRatio = 0.5
		assert.False(t, ctrl.ShouldGrow(stats))
	})

	t.Run("MemoryBasedShrink", func(t *testing.T) {
		ctrl := &MemoryBasedController{
			TargetRatio:  0.7,
			GrowFactor:   1.2,
			ShrinkFactor: 0.8,
		}

		// Should shrink when memory usage > target (0.7)
		stats := CacheStats{
			UsedRatio: 0.8,
		}
		assert.True(t, ctrl.ShouldShrink(stats))

		stats.UsedRatio = 0.6
		assert.False(t, ctrl.ShouldShrink(stats))
	})

	t.Run("MemoryBasedNextCapacity", func(t *testing.T) {
		ctrl := &MemoryBasedController{
			TargetRatio:  0.7,
			GrowFactor:   1.2,
			ShrinkFactor: 0.8,
		}

		// Growth: 1.2x
		newCap := ctrl.NextCapacity(10, true)
		assert.Equal(t, 12, newCap)

		// Shrink: 0.8x
		newCap = ctrl.NextCapacity(100, false)
		assert.Equal(t, 80, newCap)
	})
}

func TestDynamicLFU_WithCustomController(t *testing.T) {
	t.Run("SwitchToLoadBasedController", func(t *testing.T) {
		cache := NewDynamicLFU[string, int](10)

		// Switch to load-based controller with aggressive growth
		loadCtrl := &LoadBasedController{
			HighWaterMark: 0.7, // Grow at 70% load
			LowWaterMark:  0.2,
			GrowFactor:    3.0, // Triple the capacity
			ShrinkFactor:  0.5,
		}
		cache.SetCapacityController(loadCtrl)
		cache.EnableAutoGrow(true)

		// Fill to 70% (7 items)
		for i := 0; i < 7; i++ {
			cache.Set(string(rune('a'+i)), i)
		}

		// Should trigger growth with next insert
		cache.Set("h", 8)

		// With 3x growth, capacity should be 30 (10 * 3)
		assert.Equal(t, 30, cache.Capacity())
	})

	t.Run("SwitchToMemoryBasedController", func(t *testing.T) {
		cache := NewDynamicLFU[string, int](10)

		// Switch to memory-based controller
		memCtrl := &MemoryBasedController{
			TargetRatio:  0.7,
			GrowFactor:   1.5,
			ShrinkFactor: 0.6,
		}
		cache.SetCapacityController(memCtrl)
		cache.EnableAutoGrow(true)

		// Memory-based controller uses runtime.MemStats
		// Growth happens when UsedRatio < 0.35 (0.7 * 0.5)
		// This is harder to test deterministically, so we just verify the controller is set
		assert.NotNil(t, cache.controller)
	})

	t.Run("ConfigureDefaultController", func(t *testing.T) {
		cache := NewDynamicLFU[string, int](10)

		// Modify default controller parameters
		cache.SetAutoGrowParameters(0.75) // Grow at 75% instead of 90%

		// Should still use DefaultCapacityController
		defaultCtrl, ok := cache.controller.(*DefaultCapacityController)
		assert.True(t, ok)
		assert.Equal(t, 0.75, defaultCtrl.LoadFactor)
	})

	t.Run("ControllerRespectsBounds", func(t *testing.T) {
		cache := NewDynamicLFU[string, int](10)
		cache.SetCapacityLimits(5, 50)

		// Even with aggressive growth controller
		aggressiveCtrl := &LoadBasedController{
			HighWaterMark: 0.5,
			GrowFactor:    10.0, // Try to grow 10x
			ShrinkFactor:  0.1,
		}
		cache.SetCapacityController(aggressiveCtrl)
		cache.EnableAutoGrow(true)

		// Fill cache
		for i := 0; i < 20; i++ {
			cache.Set(string(rune('a'+i)), i)
		}

		// Capacity should be clamped to max (50)
		assert.LessOrEqual(t, cache.Capacity(), 50)
		assert.GreaterOrEqual(t, cache.Capacity(), 5)
	})
}

func TestDynamicLFU_ControllerStats(t *testing.T) {
	t.Run("StatsIncludeHitMissRates", func(t *testing.T) {
		cache := NewDynamicLFU[string, int](10)

		// Add items
		cache.Set("a", 1)
		cache.Set("b", 2)
		cache.Set("c", 3)

		// Generate hits and misses
		cache.Get("a") // hit
		cache.Get("b") // hit
		cache.Get("x") // miss
		cache.Get("y") // miss
		cache.Get("z") // miss

		stats := cache.getStats()

		// 2 hits, 3 misses = 40% hit rate, 60% miss rate
		assert.Equal(t, 0.4, stats.HitRate)
		assert.Equal(t, 0.6, stats.MissRate)
		assert.Equal(t, 3, stats.Size)
		assert.Equal(t, 10, stats.Capacity)
	})

	t.Run("StatsIncludeMemoryInfo", func(t *testing.T) {
		cache := NewDynamicLFU[string, int](10)

		stats := cache.getStats()

		// Memory stats should be populated
		assert.Greater(t, stats.AllocBytes, uint64(0))
		assert.Greater(t, stats.SysBytes, uint64(0))
		assert.GreaterOrEqual(t, stats.UsedRatio, 0.0)
		assert.LessOrEqual(t, stats.UsedRatio, 1.0)
	})
}
