package algo

import (
	"math"
	"math/rand"
	"sync"
)

// Helper functions to make rand functions mockable for testing
var (
// randIntn    = rand.Intn
// randFloat64 = rand.Float64
)

// ReservoirSampleCollector maintains a random sample of items using reservoir sampling.
// It is a generic struct that works with any type.
// Example usage:
//
//	collector := NewReservoirSampleCollection[int]()
//	for i := 0; i < 1000; i++ {
//	    collector.Add(i, 10)
//	    collector.Random(1)
//	}
//	sample := collector.Random(5)
//	fmt.Println(sample) // Output: A random sample of 5 integers from the collection
//
// This struct is useful for maintaining a representative random sample from a stream of data
// where the total size is unknown or too large to store completely. It implements Algorithm L
// for optimal performance with O(k(1+log(n/k))) expected running time. The collection is
// thread-safe and can be accessed concurrently.
type ReservoirSampleCollector[T any] struct {
	items []T
	size  int
	mu    sync.RWMutex
}

func NewReservoirSampleCollection[T any]() *ReservoirSampleCollector[T] {
	return &ReservoirSampleCollector[T]{items: make([]T, 0)}
}

// Add adds an item to the reservoir using reservoir sampling algorithm.
// If maxSize is > 0, it maintains at most maxSize items.
// Returns the index at which the item was added, or -1 if it wasn't added.
func (r *ReservoirSampleCollector[T]) Add(item T, maxSize int) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.size++ // Increment total items seen count

	// If we haven't reached maxSize yet, simply append
	if maxSize <= 0 || len(r.items) < maxSize {
		r.items = append(r.items, item)
		return len(r.items) - 1
	}

	// Apply reservoir sampling: with probability maxSize/size, replace a random element
	if rand.Intn(r.size) < maxSize {
		// Replace a random item in the reservoir
		idx := rand.Intn(maxSize)
		r.items[idx] = item
		return idx
	}

	return -1 // Item wasn't added
}

// Random returns k random items from the collection.
// If k > number of items in collection, returns all items.
func (r *ReservoirSampleCollector[T]) Random(k int) []T {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if k <= 0 {
		return []T{}
	}

	itemCount := len(r.items)
	if itemCount == 0 {
		return []T{}
	}

	if k >= itemCount {
		// Return copy of all items
		result := make([]T, itemCount)
		copy(result, r.items)
		return result
	}

	// Use the existing ReservoirSample function
	return ReservoirSample(r.items, k)
}

// Size returns the total number of items seen so far
func (r *ReservoirSampleCollector[T]) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.size
}

// Count returns the number of items currently in the reservoir
func (r *ReservoirSampleCollector[T]) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.items)
}

// ReservoirSample implements Algorithm L for reservoir sampling.
// It efficiently selects k random elements from a slice of any type.
// This algorithm has an optimal expected running time of O(k(1+log(n/k))).
//
// Parameters:
//   - source: The slice to sample from
//   - k: The number of items to sample
//
// Returns:
//   - A slice containing k randomly selected items from the source
func ReservoirSample[T any](source []T, k int) []T {
	n := len(source)

	// Edge cases
	if k <= 0 {
		return []T{}
	}
	if k >= n {
		result := make([]T, n)
		copy(result, source)
		return result
	}

	// Initialize reservoir with the first k items
	reservoir := make([]T, k)
	for i := 0; i < k; i++ {
		reservoir[i] = source[i]
	}

	// Algorithm L implementation
	W := math.Exp(math.Log(rand.Float64()) / float64(k))
	i := k

	for {
		// Skip items based on geometric distribution
		i += int(math.Floor(math.Log(rand.Float64())/math.Log(1-W))) + 1

		if i < n {
			// Replace a random item in the reservoir with item i
			randIndex := rand.Intn(k)
			reservoir[randIndex] = source[i]
			W *= math.Exp(math.Log(rand.Float64()) / float64(k))
		} else {
			// We've processed all items
			break
		}
	}

	return reservoir
}

// ReservoirSampleStream implements Algorithm L for reservoir sampling from a stream.
// It efficiently selects k random elements from a stream of unknown size.
//
// Parameters:
//   - stream: A channel providing the stream of items
//   - k: The number of items to sample
//
// Returns:
//   - A slice containing k randomly selected items from the stream
func ReservoirSampleStream[T any](stream <-chan T, k int) []T {
	// Edge case
	if k <= 0 {
		return []T{}
	}

	// Initialize reservoir
	reservoir := make([]T, 0, k)

	// Read first k items
	i := 0
	for item := range stream {
		if i < k {
			reservoir = append(reservoir, item)
		} else {
			break
		}
		i++
	}

	// If stream had fewer than k items, return what we got
	if i < k {
		return reservoir
	}

	// Continue with Algorithm L for the rest of the stream
	W := math.Exp(math.Log(rand.Float64()) / float64(k))
	skip := int(math.Floor(math.Log(rand.Float64())/math.Log(1-W))) + 1

	// Process the remaining items
	skipped := 0
	for item := range stream {
		if skipped >= skip {
			// Replace a random item in the reservoir
			randIndex := rand.Intn(k)
			reservoir[randIndex] = item

			// Update W and calculate next skip
			W *= math.Exp(math.Log(rand.Float64()) / float64(k))
			skip = int(math.Floor(math.Log(rand.Float64())/math.Log(1-W))) + 1
			skipped = 0
		} else {
			skipped++
		}
	}

	return reservoir
}
