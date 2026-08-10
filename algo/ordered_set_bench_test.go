package algo

import (
	"cmp"
	"fmt"
	"math/rand"
	"slices"
	"testing"
)

// ArraySet represents a simple array-based set for comparison
type ArraySet[T cmp.Ordered] struct {
	data []T
}

func NewArraySet[T cmp.Ordered]() *ArraySet[T] {
	return &ArraySet[T]{
		data: make([]T, 0),
	}
}

func (s *ArraySet[T]) Add(element T) bool {
	if s.Contains(element) {
		return false
	}
	s.data = append(s.data, element)
	return true
}

func (s *ArraySet[T]) Remove(element T) bool {
	for i, v := range s.data {
		if v == element {
			s.data = append(s.data[:i], s.data[i+1:]...)
			return true
		}
	}
	return false
}

func (s *ArraySet[T]) Contains(element T) bool {
	for _, v := range s.data {
		if v == element {
			return true
		}
	}
	return false
}

func (s *ArraySet[T]) Size() int {
	return len(s.data)
}

func (s *ArraySet[T]) ToSlice() []T {
	result := make([]T, len(s.data))
	copy(result, s.data)
	slices.Sort(result)
	return result
}

// BenchmarkOrderedSet_Add benchmarks add operations
func BenchmarkOrderedSet_Add(b *testing.B) {
	b.Run("Sequential", func(b *testing.B) {
		set := NewOrderedSet[int]()

		for b.Loop() {
			b.StopTimer()
			i := b.N % 10000 // Reset after 10000 to avoid infinite growth
			b.StartTimer()

			set.Add(i)
		}
	})

	b.Run("Random", func(b *testing.B) {
		set := NewOrderedSet[int]()
		rng := rand.New(rand.NewSource(42))

		for b.Loop() {
			value := rng.Intn(100000)
			set.Add(value)
		}
	})
}

// BenchmarkArraySet_Add benchmarks array-based add operations for comparison
func BenchmarkArraySet_Add(b *testing.B) {
	b.Run("Sequential", func(b *testing.B) {
		set := NewArraySet[int]()

		for b.Loop() {
			b.StopTimer()
			i := b.N % 10000 // Reset after 10000 to avoid infinite growth
			b.StartTimer()

			set.Add(i)
		}
	})

	b.Run("Random", func(b *testing.B) {
		set := NewArraySet[int]()
		rng := rand.New(rand.NewSource(42))

		for b.Loop() {
			value := rng.Intn(100000)
			set.Add(value)
		}
	})
}

// BenchmarkOrderedSet_Contains benchmarks lookup operations
func BenchmarkOrderedSet_Contains(b *testing.B) {
	for _, size := range []int{100, 1000, 10000, 100000} {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			set := NewOrderedSet[int]()
			rng := rand.New(rand.NewSource(42))

			// Setup: populate set
			for i := 0; i < size; i++ {
				set.Add(rng.Intn(size * 2))
			}

			searchRng := rand.New(rand.NewSource(123))

			for b.Loop() {
				value := searchRng.Intn(size * 2)
				set.Contains(value)
			}
		})
	}
}

// BenchmarkArraySet_Contains benchmarks array-based lookup operations
func BenchmarkArraySet_Contains(b *testing.B) {
	for _, size := range []int{100, 1000, 10000, 100000} {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			set := NewArraySet[int]()
			rng := rand.New(rand.NewSource(42))

			// Setup: populate set
			for i := 0; i < size; i++ {
				set.Add(rng.Intn(size * 2))
			}

			searchRng := rand.New(rand.NewSource(123))

			for b.Loop() {
				value := searchRng.Intn(size * 2)
				set.Contains(value)
			}
		})
	}
}

// BenchmarkOrderedSet_Remove benchmarks remove operations
func BenchmarkOrderedSet_Remove(b *testing.B) {
	b.Run("ExistingElements", func(b *testing.B) {
		const setupSize = 10000

		for b.Loop() {
			b.StopTimer()
			set := NewOrderedSet[int]()
			elements := make([]int, setupSize)
			for i := 0; i < setupSize; i++ {
				elements[i] = i
				set.Add(i)
			}
			removeIndex := b.N % setupSize
			b.StartTimer()

			set.Remove(elements[removeIndex])
		}
	})

	b.Run("NonExistingElements", func(b *testing.B) {
		set := NewOrderedSet[int]()
		for i := 0; i < 10000; i++ {
			set.Add(i)
		}

		for b.Loop() {
			// Try to remove elements that don't exist
			set.Remove(10000 + (b.N % 1000))
		}
	})
}

// BenchmarkArraySet_Remove benchmarks array-based remove operations
func BenchmarkArraySet_Remove(b *testing.B) {
	b.Run("ExistingElements", func(b *testing.B) {
		const setupSize = 1000 // Smaller size for array due to O(n) complexity

		for b.Loop() {
			b.StopTimer()
			set := NewArraySet[int]()
			elements := make([]int, setupSize)
			for i := 0; i < setupSize; i++ {
				elements[i] = i
				set.Add(i)
			}
			removeIndex := b.N % setupSize
			b.StartTimer()

			set.Remove(elements[removeIndex])
		}
	})
}

// BenchmarkOrderedSet_ToSlice benchmarks ordered traversal
func BenchmarkOrderedSet_ToSlice(b *testing.B) {
	for _, size := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			set := NewOrderedSet[int]()
			rng := rand.New(rand.NewSource(42))

			// Setup: populate set with random data
			for i := 0; i < size; i++ {
				set.Add(rng.Intn(size * 2))
			}

			var result []int
			for b.Loop() {
				result = set.ToSlice()
			}

			// Prevent optimization
			if len(result) == 0 && !testing.Short() {
				b.Fatal("Unexpected empty result")
			}
		})
	}
}

// BenchmarkArraySet_ToSlice benchmarks array-based sorted traversal
func BenchmarkArraySet_ToSlice(b *testing.B) {
	for _, size := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			set := NewArraySet[int]()
			rng := rand.New(rand.NewSource(42))

			// Setup: populate set with random data
			for i := 0; i < size; i++ {
				set.Add(rng.Intn(size * 2))
			}

			var result []int
			for b.Loop() {
				result = set.ToSlice()
			}

			// Prevent optimization
			if len(result) == 0 && !testing.Short() {
				b.Fatal("Unexpected empty result")
			}
		})
	}
}

// BenchmarkOrderedSet_Union benchmarks set union operations
func BenchmarkOrderedSet_Union(b *testing.B) {
	for _, size := range []int{100, 1000, 5000} {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			set1 := NewOrderedSet[int]()
			set2 := NewOrderedSet[int]()
			rng := rand.New(rand.NewSource(42))

			// Setup: populate sets with overlapping data
			for i := 0; i < size; i++ {
				set1.Add(rng.Intn(size))
				set2.Add(rng.Intn(size))
			}

			var result *OrderedSet[int]
			for b.Loop() {
				result = set1.Union(set2)
			}

			// Prevent optimization
			if result.Size() == 0 && !testing.Short() {
				b.Fatal("Unexpected empty result")
			}
		})
	}
}

// BenchmarkOrderedSet_Intersection benchmarks set intersection operations
func BenchmarkOrderedSet_Intersection(b *testing.B) {
	for _, size := range []int{100, 1000, 5000} {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			set1 := NewOrderedSet[int]()
			set2 := NewOrderedSet[int]()
			rng := rand.New(rand.NewSource(42))

			// Setup: populate sets with overlapping data
			for i := 0; i < size; i++ {
				val := rng.Intn(size / 2) // Higher chance of overlap
				set1.Add(val)
				set2.Add(val + size/4) // Partial overlap
			}

			var result *OrderedSet[int]
			for b.Loop() {
				result = set1.Intersection(set2)
			}

			// Prevent optimization
			if result == nil && !testing.Short() {
				b.Fatal("Unexpected nil result")
			}
		})
	}
}

// BenchmarkOrderedSet_MinMax benchmarks min/max operations
func BenchmarkOrderedSet_MinMax(b *testing.B) {
	for _, size := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprintf("Min-Size-%d", size), func(b *testing.B) {
			set := NewOrderedSet[int]()
			rng := rand.New(rand.NewSource(42))

			// Setup: populate set
			for i := 0; i < size; i++ {
				set.Add(rng.Intn(size * 2))
			}

			var result int
			for b.Loop() {
				result, _ = set.Min()
			}

			// Prevent optimization
			if result == 0 && !testing.Short() {
				// Note: result could legitimately be 0, so this is a weak check
			}
		})

		b.Run(fmt.Sprintf("Max-Size-%d", size), func(b *testing.B) {
			set := NewOrderedSet[int]()
			rng := rand.New(rand.NewSource(42))

			// Setup: populate set
			for i := 0; i < size; i++ {
				set.Add(rng.Intn(size * 2))
			}

			var result int
			for b.Loop() {
				result, _ = set.Max()
			}

			// Prevent optimization
			if result == 0 && !testing.Short() {
				// Note: result could legitimately be 0, so this is a weak check
			}
		})
	}
}

// BenchmarkOrderedSet_Mixed benchmarks mixed operations
func BenchmarkOrderedSet_Mixed(b *testing.B) {
	b.Run("AddContainsRemove", func(b *testing.B) {
		set := NewOrderedSet[int]()
		rng := rand.New(rand.NewSource(42))

		for b.Loop() {
			value := rng.Intn(10000)

			// 50% chance to add, 30% chance to check contains, 20% chance to remove
			op := rng.Intn(100)
			if op < 50 {
				set.Add(value)
			} else if op < 80 {
				set.Contains(value)
			} else {
				set.Remove(value)
			}
		}
	})
}

// BenchmarkArraySet_Mixed benchmarks mixed operations on array set
func BenchmarkArraySet_Mixed(b *testing.B) {
	b.Run("AddContainsRemove", func(b *testing.B) {
		set := NewArraySet[int]()
		rng := rand.New(rand.NewSource(42))

		for b.Loop() {
			value := rng.Intn(1000) // Smaller range for array due to O(n) operations

			// 50% chance to add, 30% chance to check contains, 20% chance to remove
			op := rng.Intn(100)
			if op < 50 {
				set.Add(value)
			} else if op < 80 {
				set.Contains(value)
			} else {
				set.Remove(value)
			}
		}
	})
}

// BenchmarkOrderedSet_Parallel benchmarks concurrent operations
func BenchmarkOrderedSet_Parallel(b *testing.B) {
	b.Run("ParallelContains", func(b *testing.B) {
		set := NewOrderedSet[int]()

		// Setup: populate set
		for i := 0; i < 10000; i++ {
			set.Add(i)
		}

		b.RunParallel(func(pb *testing.PB) {
			rng := rand.New(rand.NewSource(42))
			for pb.Next() {
				value := rng.Intn(20000) // Some hits, some misses
				set.Contains(value)
			}
		})
	})
}

// BenchmarkOrderedSet_Memory benchmarks memory allocation
func BenchmarkOrderedSet_Memory(b *testing.B) {
	b.ReportAllocs()

	b.Run("Add", func(b *testing.B) {
		for b.Loop() {
			b.StopTimer()
			set := NewOrderedSet[int]()
			b.StartTimer()

			set.Add(42)
		}
	})

	b.Run("ToSlice", func(b *testing.B) {
		set := NewOrderedSet[int]()
		for i := 0; i < 100; i++ {
			set.Add(i)
		}

		for b.Loop() {
			_ = set.ToSlice()
		}
	})

	b.Run("Union", func(b *testing.B) {
		set1 := NewOrderedSet[int]()
		set2 := NewOrderedSet[int]()
		for i := 0; i < 50; i++ {
			set1.Add(i)
			set2.Add(i + 25) // Some overlap
		}

		for b.Loop() {
			_ = set1.Union(set2)
		}
	})
}

// BenchmarkComparison_AddContains benchmarks direct comparison between implementations
func BenchmarkComparison_AddContains(b *testing.B) {
	const dataSize = 1000

	b.Run("OrderedSet", func(b *testing.B) {
		rng := rand.New(rand.NewSource(42))

		for b.Loop() {
			b.StopTimer()
			set := NewOrderedSet[int]()
			data := make([]int, dataSize)
			for i := 0; i < dataSize; i++ {
				data[i] = rng.Intn(dataSize * 2)
			}
			b.StartTimer()

			// Add all elements
			for _, v := range data {
				set.Add(v)
			}

			// Check contains for all elements
			for _, v := range data {
				set.Contains(v)
			}
		}
	})

	b.Run("ArraySet", func(b *testing.B) {
		rng := rand.New(rand.NewSource(42))

		for b.Loop() {
			b.StopTimer()
			set := NewArraySet[int]()
			data := make([]int, dataSize)
			for i := 0; i < dataSize; i++ {
				data[i] = rng.Intn(dataSize * 2)
			}
			b.StartTimer()

			// Add all elements
			for _, v := range data {
				set.Add(v)
			}

			// Check contains for all elements
			for _, v := range data {
				set.Contains(v)
			}
		}
	})
}

// BenchmarkOrderedSet_CompareWithArraySet_Contains compares Contains performance between implementations
func BenchmarkOrderedSet_CompareWithArraySet_Contains(b *testing.B) {
	for _, size := range []int{10, 1000, 100000, 1000000} {
		b.Run(fmt.Sprintf("OrderedSet-Size-%d", size), func(b *testing.B) {
			set := NewOrderedSet[int]()
			rng := rand.New(rand.NewSource(42))

			// Setup: populate set with deterministic data
			for i := 0; i < size; i++ {
				set.Add(rng.Intn(size * 2))
			}

			searchRng := rand.New(rand.NewSource(123))

			for b.Loop() {
				value := searchRng.Intn(size * 2)
				set.Contains(value)
			}
		})

		b.Run(fmt.Sprintf("ArraySet-Size-%d", size), func(b *testing.B) {
			set := NewArraySet[int]()
			rng := rand.New(rand.NewSource(42))

			// Setup: populate set with same deterministic data
			for i := 0; i < size; i++ {
				set.Add(rng.Intn(size * 2))
			}

			searchRng := rand.New(rand.NewSource(123))

			for b.Loop() {
				value := searchRng.Intn(size * 2)
				set.Contains(value)
			}
		})
	}
}
