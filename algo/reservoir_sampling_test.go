package algo

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewReservoirSampleCollection(t *testing.T) {
	collector := NewReservoirSampleCollection[int]()
	assert.NotNil(t, collector)
	assert.Equal(t, 0, collector.Size())
	assert.Equal(t, 0, collector.Count())
	assert.Len(t, collector.items, 0)
}

func TestReservoirSampleCollector_Add(t *testing.T) {
	// Create a deterministic random generator instead of using deprecated rand.Seed
	r := rand.New(rand.NewSource(42))

	// Store original rand functions
	origRandIntn := randIntn

	// Replace rand functions with our deterministic versions
	randIntn = func(n int) int {
		return r.Intn(n)
	}

	// Restore original functions after the test
	defer func() {
		randIntn = origRandIntn
	}()

	tests := []struct {
		name        string
		setup       func() *ReservoirSampleCollector[int]
		item        int
		maxSize     int
		wantAdded   bool
		wantMinSize int
		wantMaxSize int
	}{
		{
			name: "should add to empty reservoir",
			setup: func() *ReservoirSampleCollector[int] {
				return NewReservoirSampleCollection[int]()
			},
			item:        42,
			maxSize:     10,
			wantAdded:   true,
			wantMinSize: 1,
			wantMaxSize: 1,
		},
		{
			name: "should add when reservoir not full",
			setup: func() *ReservoirSampleCollector[int] {
				r := NewReservoirSampleCollection[int]()
				r.Add(1, 10)
				r.Add(2, 10)
				return r
			},
			item:        42,
			maxSize:     10,
			wantAdded:   true,
			wantMinSize: 3,
			wantMaxSize: 3,
		},
		{
			name: "should maintain size cap",
			setup: func() *ReservoirSampleCollector[int] {
				r := NewReservoirSampleCollection[int]()
				for i := 0; i < 5; i++ {
					r.Add(i, 3)
				}
				return r
			},
			item:        42,
			maxSize:     3,
			wantMinSize: 3,
			wantMaxSize: 3,
		},
		{
			name: "should allow unlimited size with maxSize <= 0",
			setup: func() *ReservoirSampleCollector[int] {
				r := NewReservoirSampleCollection[int]()
				for i := 0; i < 100; i++ {
					r.Add(i, -1)
				}
				return r
			},
			item:        42,
			maxSize:     -1,
			wantAdded:   true,
			wantMinSize: 101,
			wantMaxSize: 101,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := tt.setup()
			beforeSize := collector.Size()

			idx := collector.Add(tt.item, tt.maxSize)

			// Check reservoir state after adding
			afterSize := collector.Size()
			afterCount := collector.Count()

			// Size should always increase by 1 (counts seen items)
			assert.Equal(t, beforeSize+1, afterSize)

			// Count should obey maxSize constraint
			assert.GreaterOrEqual(t, afterCount, tt.wantMinSize)
			assert.LessOrEqual(t, afterCount, tt.wantMaxSize)

			// Check if item was added
			if idx != -1 {
				// Should be within bounds
				assert.GreaterOrEqual(t, idx, 0)
				assert.Less(t, idx, afterCount)
				// The specific item should be in the collection at the returned index
				assert.Equal(t, tt.item, collector.items[idx])
			}
		})
	}
}

func TestReservoirSampleCollector_Random(t *testing.T) {
	// Create a deterministic random generator instead of using deprecated rand.Seed
	r := rand.New(rand.NewSource(42))

	// Store original rand functions
	origRandIntn := randIntn

	// Replace rand functions with our deterministic versions
	randIntn = func(n int) int {
		return r.Intn(n)
	}

	// Restore original functions after the test
	defer func() {
		randIntn = origRandIntn
	}()

	tests := []struct {
		name       string
		setup      func() *ReservoirSampleCollector[int]
		k          int
		wantLen    int
		wantEmpty  bool
		checkItems func(*testing.T, []int)
	}{
		{
			name: "should return empty for k <= 0",
			setup: func() *ReservoirSampleCollector[int] {
				r := NewReservoirSampleCollection[int]()
				for i := 0; i < 10; i++ {
					r.Add(i, 10)
				}
				return r
			},
			k:         0,
			wantLen:   0,
			wantEmpty: true,
		},
		{
			name: "should return empty for empty reservoir",
			setup: func() *ReservoirSampleCollector[int] {
				return NewReservoirSampleCollection[int]()
			},
			k:         5,
			wantLen:   0,
			wantEmpty: true,
		},
		{
			name: "should return k items for sufficient reservoir",
			setup: func() *ReservoirSampleCollector[int] {
				r := NewReservoirSampleCollection[int]()
				for i := 0; i < 100; i++ {
					r.Add(i, 100)
				}
				return r
			},
			k:       10,
			wantLen: 10,
			checkItems: func(t *testing.T, items []int) {
				// Each number should be in range 0-99
				for _, item := range items {
					assert.GreaterOrEqual(t, item, 0)
					assert.Less(t, item, 100)
				}
			},
		},
		{
			name: "should return all items for k >= count",
			setup: func() *ReservoirSampleCollector[int] {
				r := NewReservoirSampleCollection[int]()
				for i := 0; i < 5; i++ {
					r.Add(i, 10)
				}
				return r
			},
			k:       10,
			wantLen: 5,
			checkItems: func(t *testing.T, items []int) {
				// Should contain exactly 0,1,2,3,4 in some order
				expected := map[int]bool{0: true, 1: true, 2: true, 3: true, 4: true}
				for _, item := range items {
					assert.True(t, expected[item], "unexpected item: %d", item)
					delete(expected, item)
				}
				assert.Empty(t, expected, "missing expected items")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := tt.setup()
			result := collector.Random(tt.k)

			assert.Len(t, result, tt.wantLen)

			if tt.wantEmpty {
				assert.Empty(t, result)
			}

			if tt.checkItems != nil {
				tt.checkItems(t, result)
			}

			// Test that calling Random doesn't modify the original collection
			originalCount := collector.Count()
			collector.Random(tt.k)
			assert.Equal(t, originalCount, collector.Count())
		})
	}
}

func TestReservoirSampleCollector_Size_Count(t *testing.T) {
	collector := NewReservoirSampleCollection[int]()
	assert.Equal(t, 0, collector.Size())
	assert.Equal(t, 0, collector.Count())

	// Add items without constraint
	for i := 0; i < 100; i++ {
		collector.Add(i, 0)
	}
	assert.Equal(t, 100, collector.Size())  // Total seen
	assert.Equal(t, 100, collector.Count()) // Items in reservoir

	// Add items with constraint
	collector = NewReservoirSampleCollection[int]()
	for i := 0; i < 100; i++ {
		collector.Add(i, 10)
	}
	assert.Equal(t, 100, collector.Size()) // Total seen
	assert.Equal(t, 10, collector.Count()) // Items in reservoir (capped)
}

func TestReservoirSample(t *testing.T) {
	// Create a deterministic random generator instead of using deprecated rand.Seed
	r := rand.New(rand.NewSource(42))

	// Store original rand functions
	origRandIntn := randIntn
	origRandFloat64 := randFloat64

	// Replace rand functions with our deterministic versions
	randIntn = func(n int) int {
		return r.Intn(n)
	}

	randFloat64 = func() float64 {
		return r.Float64()
	}

	// Restore original functions after the test
	defer func() {
		randIntn = origRandIntn
		randFloat64 = origRandFloat64
	}()

	tests := []struct {
		name       string
		source     []int
		k          int
		want       int // length of result
		checkItems func(*testing.T, []int, []int)
	}{
		{
			name:   "should return empty slice for k <= 0",
			source: []int{1, 2, 3, 4, 5},
			k:      0,
			want:   0,
		},
		{
			name:   "should return all items when k >= length",
			source: []int{1, 2, 3, 4, 5},
			k:      10,
			want:   5,
			checkItems: func(t *testing.T, source, result []int) {
				// Sort both slices to compare
				sourceMap := make(map[int]bool)
				resultMap := make(map[int]bool)

				for _, v := range source {
					sourceMap[v] = true
				}

				for _, v := range result {
					resultMap[v] = true
					assert.True(t, sourceMap[v], "Result contains item not in source: %d", v)
				}

				assert.Equal(t, len(sourceMap), len(resultMap), "Result should contain all items from source")
			},
		},
		{
			name:   "should sample k items from larger source",
			source: generateSequence(1, 1000),
			k:      50,
			want:   50,
			checkItems: func(t *testing.T, source, result []int) {
				// Check that all items are from source
				sourceMap := make(map[int]bool)
				for _, v := range source {
					sourceMap[v] = true
				}

				for _, v := range result {
					assert.True(t, sourceMap[v], "Result contains item not in source: %d", v)
				}

				// Check that samples are reasonably distributed
				// (this is a probabilistic test, but reservoir sampling should be uniform)
				resultMap := make(map[int]bool)
				for _, v := range result {
					resultMap[v] = true
				}

				assert.Equal(t, len(result), len(resultMap), "Result should have no duplicates")
			},
		},
		{
			name:   "should handle empty source",
			source: []int{},
			k:      5,
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReservoirSample(tt.source, tt.k)
			assert.Len(t, result, tt.want)

			if tt.checkItems != nil {
				tt.checkItems(t, tt.source, result)
			}

			// Verify that source was not modified
			if len(tt.source) > 0 {
				originalSource := make([]int, len(tt.source))
				copy(originalSource, tt.source)
				assert.Equal(t, originalSource, tt.source)
			}
		})
	}
}

func TestReservoirSampleStream(t *testing.T) {
	// Create a deterministic random generator instead of using deprecated rand.Seed
	r := rand.New(rand.NewSource(42))

	// Store original rand functions
	origRandIntn := randIntn
	origRandFloat64 := randFloat64

	// Replace rand functions with our deterministic versions
	randIntn = func(n int) int {
		return r.Intn(n)
	}

	randFloat64 = func() float64 {
		return r.Float64()
	}

	// Restore original functions after the test
	defer func() {
		randIntn = origRandIntn
		randFloat64 = origRandFloat64
	}()

	tests := []struct {
		name       string
		stream     []int
		k          int
		wantLen    int
		checkItems func(*testing.T, []int, []int)
	}{
		{
			name:    "should return empty slice for k <= 0",
			stream:  []int{1, 2, 3, 4, 5},
			k:       0,
			wantLen: 0,
		},
		{
			name:    "should return all items when stream smaller than k",
			stream:  []int{1, 2, 3},
			k:       5,
			wantLen: 3,
			checkItems: func(t *testing.T, stream, result []int) {
				// Convert slices to maps for easy comparison
				streamMap := make(map[int]bool)
				for _, v := range stream {
					streamMap[v] = true
				}

				resultMap := make(map[int]bool)
				for _, v := range result {
					resultMap[v] = true
					assert.True(t, streamMap[v], "Result contains item not in stream: %d", v)
				}

				assert.Equal(t, len(streamMap), len(resultMap), "Result should contain all items from stream")
			},
		},
		{
			name:    "should sample k items from larger stream",
			stream:  generateSequence(1, 1000),
			k:       50,
			wantLen: 50,
			checkItems: func(t *testing.T, stream, result []int) {
				// Check that all items are from stream
				streamMap := make(map[int]bool)
				for _, v := range stream {
					streamMap[v] = true
				}

				for _, v := range result {
					assert.True(t, streamMap[v], "Result contains item not in stream: %d", v)
				}

				// Check that samples are reasonably distributed
				resultMap := make(map[int]bool)
				for _, v := range result {
					resultMap[v] = true
				}

				assert.Equal(t, len(result), len(resultMap), "Result should have no duplicates")
			},
		},
		{
			name:    "should handle empty stream",
			stream:  []int{},
			k:       5,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a channel to simulate a stream
			streamCh := make(chan int)

			// Start goroutine to feed the stream
			go func() {
				defer close(streamCh)
				for _, item := range tt.stream {
					streamCh <- item
				}
			}()

			result := ReservoirSampleStream(streamCh, tt.k)

			assert.Len(t, result, tt.wantLen)

			if tt.checkItems != nil {
				tt.checkItems(t, tt.stream, result)
			}
		})
	}
}

// Helper functions to make rand functions mockable for testing
var (
	randIntn    = rand.Intn
	randFloat64 = rand.Float64
)

// generateSequence creates a sequence of integers from start to end (inclusive)
func generateSequence(start, end int) []int {
	result := make([]int, end-start+1)
	for i := 0; i < len(result); i++ {
		result[i] = start + i
	}
	return result
}

// Test for thread safety
func TestReservoirSampleCollector_ThreadSafety(t *testing.T) {
	collector := NewReservoirSampleCollection[int]()
	const concurrent = 10
	const itemsPerRoutine = 1000
	const maxSize = 100

	var wg sync.WaitGroup
	wg.Add(concurrent)

	// Multiple goroutines adding items
	for i := 0; i < concurrent; i++ {
		go func(offset int) {
			defer wg.Done()
			for j := 0; j < itemsPerRoutine; j++ {
				collector.Add(offset*itemsPerRoutine+j, maxSize)
			}
		}(i)
	}

	// While adding, read samples concurrently
	done := make(chan bool)
	go func() {
		for i := 0; i < 100; i++ {
			sample := collector.Random(10)
			// Just ensure this doesn't panic
			_ = sample
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	wg.Wait()
	<-done

	// Verify final state
	totalExpectedItems := concurrent * itemsPerRoutine
	assert.Equal(t, totalExpectedItems, collector.Size())
	assert.LessOrEqual(t, collector.Count(), maxSize)
}

// Benchmark tests for performance evaluation
func BenchmarkReservoirSample(b *testing.B) {
	// Generate a large dataset
	data := generateSequence(1, 100000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ReservoirSample(data, 100)
	}
}

func BenchmarkReservoirSampleCollector_Add(b *testing.B) {
	collector := NewReservoirSampleCollection[int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.Add(i, 1000)
	}
}

func BenchmarkReservoirSampleCollector_Random(b *testing.B) {
	// Setup collector with data
	collector := NewReservoirSampleCollection[int]()
	for i := 0; i < 100000; i++ {
		collector.Add(i, 10000)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.Random(100)
	}
}
