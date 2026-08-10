package algo

import (
	"strconv"
	"testing"
)

// BenchmarkDeque_PushBack measures the performance of PushBack operation
func BenchmarkDeque_PushBack(b *testing.B) {
	for _, size := range []int{10, 100, 1000, 10000} {
		b.Run("Size-"+strconv.Itoa(size), func(b *testing.B) {
			b.ReportAllocs()
			deque := NewDeque[int]()

			// Pre-fill with size elements
			for i := 0; i < size; i++ {
				deque.PushBack(i)
			}

			// Benchmark adding one more element
			for b.Loop() {
				deque.PushBack(size + 1)
				if deque.Size() > size+100 {
					deque = NewDeque[int]()
					for i := 0; i < size; i++ {
						deque.PushBack(i)
					}
				}
			}
		})
	}
}

// BenchmarkDeque_PushFront measures the performance of PushFront operation
func BenchmarkDeque_PushFront(b *testing.B) {
	for _, size := range []int{10, 100, 1000, 10000} {
		b.Run("Size-"+strconv.Itoa(size), func(b *testing.B) {
			b.ReportAllocs()
			deque := NewDeque[int]()

			// Pre-fill with size elements
			for i := 0; i < size; i++ {
				deque.PushBack(i)
			}

			// Benchmark adding to front
			for b.Loop() {
				deque.PushFront(size + 1)

				if deque.Size() > size+100 {
					deque = NewDeque[int]()
					for i := 0; i < size; i++ {
						deque.PushBack(i)
					}
				}
			}
		})
	}
}

// BenchmarkDeque_PopBack measures the performance of PopBack operation
func BenchmarkDeque_PopBack(b *testing.B) {
	for _, size := range []int{10, 100, 1000, 10000} {
		b.Run("Size-"+strconv.Itoa(size), func(b *testing.B) {
			b.ReportAllocs()

			var result int
			var ok bool

			for b.Loop() {
				// Setup fresh deque with elements
				// b.StopTimer()
				deque := NewDeque[int]()
				for i := 0; i < size; i++ {
					deque.PushBack(i)
				}
				// b.StartTimer()

				// Measure single PopBack operation
				result, ok = deque.PopBack()
			}

			// Use result to prevent compiler optimization
			if !ok && !testing.Short() {
				b.Fatal("PopBack failed")
			}
			_ = result
		})
	}
}

// BenchmarkDeque_PopFront measures the performance of PopFront operation
func BenchmarkDeque_PopFront(b *testing.B) {
	for _, size := range []int{10, 100, 1000, 10000} {
		b.Run("Size-"+strconv.Itoa(size), func(b *testing.B) {
			b.ReportAllocs()

			var result int
			var ok bool

			for b.Loop() {
				// Setup fresh deque with elements
				// b.StopTimer()
				deque := NewDeque[int]()
				for i := 0; i < size; i++ {
					deque.PushBack(i)
				}
				// b.StartTimer()

				// Measure single PopFront operation
				result, ok = deque.PopFront()
			}

			// Use result to prevent compiler optimization
			if !ok && !testing.Short() {
				b.Fatal("PopFront failed")
			}
			_ = result
		})
	}
}

// BenchmarkDeque_PeekOperations measures the performance of peek operations
func BenchmarkDeque_PeekOperations(b *testing.B) {
	b.Run("PeekFront", func(b *testing.B) {
		b.ReportAllocs()
		deque := NewDeque[int]()
		for i := 0; i < 1000; i++ {
			deque.PushBack(i)
		}

		var val int
		var ok bool
		for b.Loop() {
			val, ok = deque.PeekFront()
		}

		if !ok && !testing.Short() {
			b.Fatal("PeekFront failed")
		}
		_ = val
	})

	b.Run("PeekBack", func(b *testing.B) {
		b.ReportAllocs()
		deque := NewDeque[int]()
		for i := 0; i < 1000; i++ {
			deque.PushBack(i)
		}

		var val int
		var ok bool
		for b.Loop() {
			val, ok = deque.PeekBack()
		}

		if !ok && !testing.Short() {
			b.Fatal("PeekBack failed")
		}
		_ = val
	})
}

// BenchmarkDeque_StackUsage simulates using the deque as a stack (LIFO)
func BenchmarkDeque_StackUsage(b *testing.B) {
	for _, size := range []int{10, 100, 1000, 10000} {
		b.Run("Size-"+strconv.Itoa(size), func(b *testing.B) {
			b.ReportAllocs()
			deque := NewDeque[int]()

			// Pre-fill with size elements
			for i := 0; i < size; i++ {
				deque.PushBack(i)
			}

			var result int
			for b.Loop() {
				// Stack operation: Push and Pop from the same end (back)
				deque.PushBack(size)
				result, _ = deque.PopBack()
			}

			_ = result
		})
	}
}

// BenchmarkDeque_QueueUsage simulates using the deque as a queue (FIFO)
func BenchmarkDeque_QueueUsage(b *testing.B) {
	for _, size := range []int{10, 100, 1000, 10000} {
		b.Run("Size-"+strconv.Itoa(size), func(b *testing.B) {
			b.ReportAllocs()
			deque := NewDeque[int]()

			// Pre-fill with size elements
			for i := 0; i < size; i++ {
				deque.PushBack(i)
			}

			var result int
			for b.Loop() {
				// Queue operation: Push at back, Pop from front
				deque.PushBack(size)
				result, _ = deque.PopFront()
			}

			_ = result
		})
	}
}

// BenchmarkDeque_MixedOperations measures a mix of all operations
func BenchmarkDeque_MixedOperations(b *testing.B) {
	for _, size := range []int{10, 100, 1000} {
		b.Run("Size-"+strconv.Itoa(size), func(b *testing.B) {
			b.ReportAllocs()
			deque := NewDeque[int]()

			// Pre-fill with size elements
			for i := 0; i < size; i++ {
				deque.PushBack(i)
			}

			counter := 0
			var result int
			var ok bool

			for b.Loop() {
				// Mix of all operations based on counter
				switch counter % 6 {
				case 0:
					deque.PushBack(counter)
				case 1:
					deque.PushFront(counter)
				case 2:
					result, ok = deque.PopBack()
				case 3:
					result, ok = deque.PopFront()
				case 4:
					result, ok = deque.PeekBack()
				case 5:
					result, ok = deque.PeekFront()
				}

				counter++

				// Ensure deque doesn't get empty
				b.StopTimer()
				if deque.Size() < size/2 {
					for i := 0; i < size/2; i++ {
						deque.PushBack(i)
					}
				}
				b.StartTimer()
			}

			_ = result
			_ = ok
		})
	}
}
