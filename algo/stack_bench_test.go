package algo

import (
	"testing"
)

// BenchmarkSliceStack_Push measures the performance of Push operations
func BenchmarkSliceStack_Push(b *testing.B) {
	s := NewStack[int]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Push(i)
	}
}

// BenchmarkSliceStack_Pop measures the performance of Pop operations
func BenchmarkSliceStack_Pop(b *testing.B) {
	s := NewStack[int]()

	// Setup: Push N items to the stack
	for i := 0; i < b.N; i++ {
		s.Push(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.Pop()
		if err != nil {
			b.Fatal("Unexpected error during Pop:", err)
		}
	}
}

// BenchmarkSliceStack_PushPop measures the performance of alternating Push and Pop operations
func BenchmarkSliceStack_PushPop(b *testing.B) {
	s := NewStack[int]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Push(i)
		_, _ = s.Pop()
	}
}

// BenchmarkSliceStack_Peek measures the performance of Peek operations
func BenchmarkSliceStack_Peek(b *testing.B) {
	s := NewStack[string]()
	s.Push("test-item")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.Peek()
	}
}

// BenchmarkSliceStack_Size measures the performance of Size operations
func BenchmarkSliceStack_Size(b *testing.B) {
	s := NewStack[int]()
	for i := 0; i < 1000; i++ {
		s.Push(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Size()
	}
}

// BenchmarkSliceStack_IsEmpty measures the performance of IsEmpty operations
func BenchmarkSliceStack_IsEmpty(b *testing.B) {
	s := NewStack[string]()
	s.Push("test-item")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.IsEmpty()
	}
}

// BenchmarkSliceStack_PushPopParallel measures the performance of Push and Pop operations in parallel
func BenchmarkSliceStack_PushPopParallel(b *testing.B) {
	s := NewStack[int]()

	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			if counter%2 == 0 {
				s.Push(counter)
			} else {
				_, _ = s.Pop()
			}
			counter++
		}
	})
}

// BenchmarkSliceStack_LargeStack measures the performance with large number of elements
func BenchmarkSliceStack_LargeStack(b *testing.B) {
	b.Run("Push10k", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			s := NewStack[int]()
			b.StartTimer()

			for j := 0; j < 10000; j++ {
				s.Push(j)
			}
		}
	})

	b.Run("Pop10k", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			s := NewStack[int]()
			for j := 0; j < 10000; j++ {
				s.Push(j)
			}
			b.StartTimer()

			for j := 0; j < 10000; j++ {
				_, _ = s.Pop()
			}
		}
	})
}

// BenchmarkSliceStack_VS_SliceQueue compares stack and queue performance for the same operations
func BenchmarkSliceStack_VS_SliceQueue(b *testing.B) {
	// Stack benchmarks
	b.Run("Stack_Push", func(b *testing.B) {
		s := NewStack[int]()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			s.Push(i)
		}
	})

	b.Run("Stack_Pop", func(b *testing.B) {
		b.StopTimer()
		s := NewStack[int]()
		for i := 0; i < b.N; i++ {
			s.Push(i)
		}
		b.StartTimer()

		for i := 0; i < b.N; i++ {
			_, _ = s.Pop()
		}
	})

	// Queue benchmarks
	b.Run("Queue_Push", func(b *testing.B) {
		q := NewQueue[int]()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			q.Push(i)
		}
	})

	b.Run("Queue_Pop", func(b *testing.B) {
		b.StopTimer()
		q := NewQueue[int]()
		for i := 0; i < b.N; i++ {
			q.Push(i)
		}
		b.StartTimer()

		for i := 0; i < b.N; i++ {
			_, _ = q.Pop()
		}
	})
}

// BenchmarkStack_MemoryUsage simulates scenarios to help measure memory usage
func BenchmarkStack_MemoryUsage(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Create a stack with 1 million elements to test memory consumption
		b.StopTimer()
		s := NewStack[int]()
		b.StartTimer()

		for j := 0; j < 1000000; j++ {
			s.Push(j)
		}

		// Force allocation to be measured
		b.StopTimer()
		s.Size() // Use the stack to prevent compiler optimizations
		b.StartTimer()
	}
}
