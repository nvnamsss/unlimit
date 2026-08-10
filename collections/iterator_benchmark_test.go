package collections

import (
	"math/rand"
	"strconv"
	"testing"
)

// generateLargeSlice creates a slice with random numbers for benchmarking
func generateLargeSlice(size int) []int {
	slice := make([]int, size)
	for i := 0; i < size; i++ {
		slice[i] = rand.Intn(1000)
	}
	return slice
}

// BenchmarkBatchIteratorMemory tests memory usage of the BatchIterator for processing data
func BenchmarkBatchIteratorMemory(b *testing.B) {
	// Test with different data sizes to show the effect more clearly
	dataSizes := []int{10000, 100000, 1000000}

	for _, size := range dataSizes {
		b.Run("BatchIterator_"+strconv.Itoa(size), func(b *testing.B) {
			data := generateLargeSlice(size)
			batchSize := 100

			b.ResetTimer()
			b.ReportAllocs() // Track allocations

			for i := 0; i < b.N; i++ {
				iterator := NewBatchIterator(data, batchSize)
				var sum int
				// Process data through iterator
				for {
					batch, hasMore := iterator.Next()
					if len(batch) > 0 {
						for _, v := range batch {
							sum += v
						}
					}
					if !hasMore {
						break
					}
				}
				_ = sum // Prevent compiler optimization
			}
		})
	}
}

// BenchmarkBatchArrayMemory tests memory usage of the traditional approach that loads all batches into memory
func BenchmarkBatchArrayMemory(b *testing.B) {
	// Use the same data sizes as BatchIterator benchmark
	dataSizes := []int{10000, 100000, 1000000}

	for _, size := range dataSizes {
		b.Run("BatchArray_"+string(rune(size)), func(b *testing.B) {
			data := generateLargeSlice(size)
			batchSize := 100

			b.ResetTimer()
			b.ReportAllocs() // Track allocations

			for i := 0; i < b.N; i++ {
				// Create all batches at once - this is what we're comparing against
				var batches [][]int
				for j := 0; j < len(data); j += batchSize {
					end := j + batchSize
					if end > len(data) {
						end = len(data)
					}
					batches = append(batches, data[j:end])
				}

				// Process data from batches
				var sum int
				for _, batch := range batches {
					for _, v := range batch {
						sum += v
					}
				}
				_ = sum // Prevent compiler optimization
			}
		})
	}
}

// BenchmarkIteratorVsArrayComparison directly compares the two approaches with the same data
func BenchmarkIteratorVsArrayComparison(b *testing.B) {
	dataSizes := []int{1000000} // Use a large data size to emphasize difference

	for _, size := range dataSizes {
		data := generateLargeSlice(size)
		batchSize := 100

		b.Run("BatchIterator", func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				iterator := NewBatchIterator(data, batchSize)
				var sum int
				for {
					batch, hasMore := iterator.Next()
					for _, v := range batch {
						sum += v
					}
					if !hasMore {
						break
					}
				}
				_ = sum
			}
		})

		b.Run("BatchArray", func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				var batches [][]int
				for j := 0; j < len(data); j += batchSize {
					end := j + batchSize
					if end > len(data) {
						end = len(data)
					}
					batches = append(batches, data[j:end])
				}

				var sum int
				for _, batch := range batches {
					for _, v := range batch {
						sum += v
					}
				}
				_ = sum
			}
		})
	}
}

// BenchmarkBatchSizeComparison tests how different batch sizes affect memory usage
func BenchmarkBatchSizeComparison(b *testing.B) {
	dataSize := 1000000
	data := generateLargeSlice(dataSize)
	batchSizes := []int{10, 100, 1000, 10000}

	for _, batchSize := range batchSizes {
		b.Run("Iterator_BatchSize_"+string(rune(batchSize)), func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				iterator := NewBatchIterator(data, batchSize)
				var sum int
				for {
					batch, hasMore := iterator.Next()
					for _, v := range batch {
						sum += v
					}
					if !hasMore {
						break
					}
				}
				_ = sum
			}
		})

		b.Run("Array_BatchSize_"+string(rune(batchSize)), func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				var batches [][]int
				for j := 0; j < len(data); j += batchSize {
					end := j + batchSize
					if end > len(data) {
						end = len(data)
					}
					batches = append(batches, data[j:end])
				}

				var sum int
				for _, batch := range batches {
					for _, v := range batch {
						sum += v
					}
				}
				_ = sum
			}
		})
	}
}
