package algo

import (
	"strconv"
	"testing"
)

// BenchmarkSliceQueue_Push measures the performance of Push operations
func BenchmarkSliceQueue_Push(b *testing.B) {
	q := NewQueue[int]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		q.Push(i)
	}
}

// BenchmarkSliceQueue_Pop measures the performance of Pop operations
func BenchmarkSliceQueue_Pop(b *testing.B) {
	q := NewQueue[int]()

	// Setup: Push N items to the queue
	for i := 0; i < b.N; i++ {
		q.Push(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := q.Pop()
		if err != nil {
			b.Fatal("Unexpected error during Pop:", err)
		}
	}
}

// BenchmarkSliceQueue_PushPop measures the performance of alternating Push and Pop operations
func BenchmarkSliceQueue_PushPop(b *testing.B) {
	q := NewQueue[int]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		q.Push(i)
		_, _ = q.Pop()
	}
}

// BenchmarkSliceQueue_Peek measures the performance of Peek operations
func BenchmarkSliceQueue_Peek(b *testing.B) {
	q := NewQueue[string]()
	q.Push("test-item")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = q.Peek()
	}
}

// BenchmarkSliceQueue_Size measures the performance of Size operations
func BenchmarkSliceQueue_Size(b *testing.B) {
	q := NewQueue[int]()
	for i := 0; i < 1000; i++ {
		q.Push(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = q.Size()
	}
}

// BenchmarkSliceQueue_IsEmpty measures the performance of IsEmpty operations
func BenchmarkSliceQueue_IsEmpty(b *testing.B) {
	q := NewQueue[string]()
	q.Push("test-item")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = q.IsEmpty()
	}
}

func BenchmarkSliceQueue_PushParallel(b *testing.B) {
	q := NewQueue[int]()

	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			q.Push(counter)
			counter++
		}
	})
}

// BenchmarkSliceQueue_PushPopParallel measures the performance of Push and Pop operations in parallel
func BenchmarkSliceQueue_PushPopParallel(b *testing.B) {
	q := NewQueue[int]()

	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			if counter%2 == 0 {
				q.Push(counter)
			} else {
				_, _ = q.Pop()
			}
			counter++
		}
	})
}

// BenchmarkSliceQueue_LargeQueue measures the performance with large number of elements
func BenchmarkSliceQueue_LargeQueue(b *testing.B) {
	b.Run("Push10k", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			q := NewQueue[int]()
			b.StartTimer()

			for j := 0; j < 10000; j++ {
				q.Push(j)
			}
		}
	})

	b.Run("Pop10k", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			q := NewQueue[int]()
			for j := 0; j < 10000; j++ {
				q.Push(j)
			}
			b.StartTimer()

			for j := 0; j < 10000; j++ {
				_, _ = q.Pop()
			}
		}
	})
}

// BenchmarkSliceQueue_QueueOperations compares different queue operations
func BenchmarkSliceQueue_QueueOperations(b *testing.B) {
	// Push benchmarks with different queue sizes
	for _, size := range []int{10, 100, 1000, 10000} {
		b.Run("Push_Size_"+strconv.Itoa(size), func(b *testing.B) {
			q := NewQueue[int]()
			// Pre-fill queue
			for i := 0; i < size; i++ {
				q.Push(i)
			}
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				q.Push(i)
			}
		})
	}

	// Pop benchmarks with different queue sizes
	for _, size := range []int{10, 100, 1000, 10000} {
		b.Run("Pop_Size_"+strconv.Itoa(size), func(b *testing.B) {
			b.StopTimer()
			q := NewQueue[int]()
			// Fill queue with size + b.N elements
			for i := 0; i < size+b.N; i++ {
				q.Push(i)
			}
			b.StartTimer()

			for i := 0; i < b.N; i++ {
				_, _ = q.Pop()
			}
		})
	}
}

// BenchmarkSliceQueue_MemoryUsage simulates scenarios to help measure memory usage
func BenchmarkSliceQueue_MemoryUsage(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Create a queue with 1 million elements to test memory consumption
		b.StopTimer()
		q := NewQueue[int]()
		b.StartTimer()

		for j := 0; j < 1000000; j++ {
			q.Push(j)
		}

		// Force allocation to be measured
		b.StopTimer()
		q.Size() // Use the queue to prevent compiler optimizations
		b.StartTimer()
	}
}

// BenchmarkQueueFIFO_VS_StackLIFO compares FIFO behavior of queue vs LIFO behavior of stack
func BenchmarkQueueFIFO_VS_StackLIFO(b *testing.B) {
	// Queue FIFO benchmark
	b.Run("Queue_FIFO", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			q := NewQueue[int]()
			for j := 0; j < 1000; j++ {
				q.Push(j)
			}
			b.StartTimer()

			// In FIFO, we should get 0, then 1, then 2...
			for j := 0; j < 1000; j++ {
				val, _ := q.Pop()
				if val != j && !testing.Short() {
					b.Fatalf("Expected %d, got %v - Queue not behaving as FIFO", j, val)
				}
			}
		}
	})

	// Stack LIFO benchmark
	b.Run("Stack_LIFO", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			s := NewStack[int]() // Note: You'll need to implement a generic Stack type
			for j := 0; j < 1000; j++ {
				s.Push(j)
			}
			b.StartTimer()

			// In LIFO, we should get 999, then 998, then 997...
			for j := 999; j >= 0; j-- {
				val, _ := s.Pop()
				if val != j && !testing.Short() {
					b.Fatalf("Expected %d, got %v - Stack not behaving as LIFO", j, val)
				}
			}
		}
	})
}

// =============================================================================
// WaitFreeQueue Benchmarks
// =============================================================================

// BenchmarkWaitFreeQueue_Push measures the performance of Push operations
func BenchmarkWaitFreeQueue_Push(b *testing.B) {
	q := NewLockFreeQueue[int]()

	for b.Loop() {
		q.Push(1)
	}
}

// BenchmarkWaitFreeQueue_Pop measures the performance of Pop operations
func BenchmarkWaitFreeQueue_Pop(b *testing.B) {
	q := NewLockFreeQueue[int]()

	// Setup: Push enough items to ensure we can pop during the entire benchmark
	for i := 0; i < b.N; i++ {
		q.Push(i)
	}

	b.ResetTimer()
	for b.Loop() {
		_, _ = q.Pop()
	}
}

// BenchmarkWaitFreeQueue_PushPop measures the performance of alternating Push and Pop operations
func BenchmarkWaitFreeQueue_PushPop(b *testing.B) {
	q := NewLockFreeQueue[int]()

	for b.Loop() {
		q.Push(1)
		_, _ = q.Pop()
	}
}

// BenchmarkWaitFreeQueue_Peek measures the performance of Peek operations
func BenchmarkWaitFreeQueue_Peek(b *testing.B) {
	q := NewLockFreeQueue[string]()
	q.Push("test-item")

	for b.Loop() {
		_, _ = q.Peek()
	}
}

// BenchmarkWaitFreeQueue_Size measures the performance of Size operations
func BenchmarkWaitFreeQueue_Size(b *testing.B) {
	q := NewLockFreeQueue[int]()
	for i := 0; i < 1000; i++ {
		q.Push(i)
	}

	for b.Loop() {
		_ = q.Size()
	}
}

// BenchmarkWaitFreeQueue_IsEmpty measures the performance of IsEmpty operations
func BenchmarkWaitFreeQueue_IsEmpty(b *testing.B) {
	q := NewLockFreeQueue[string]()
	q.Push("test-item")

	for b.Loop() {
		_ = q.IsEmpty()
	}
}

// BenchmarkWaitFreeQueue_PushParallel measures the performance of Push operations in parallel
func BenchmarkWaitFreeQueue_PushParallel(b *testing.B) {
	q := NewLockFreeQueue[int]()

	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			q.Push(counter)
			counter++
		}
	})
}

// BenchmarkWaitFreeQueue_PopParallel measures the performance of Pop operations in parallel
func BenchmarkWaitFreeQueue_PopParallel(b *testing.B) {
	q := NewLockFreeQueue[int]()

	// Pre-fill with enough items
	for i := 0; i < b.N*10; i++ {
		q.Push(i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = q.Pop()
		}
	})
}

// BenchmarkWaitFreeQueue_PushPopParallel measures the performance of Push and Pop operations in parallel
func BenchmarkWaitFreeQueue_PushPopParallel(b *testing.B) {
	q := NewLockFreeQueue[int]()

	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			if counter%2 == 0 {
				q.Push(counter)
			} else {
				_, _ = q.Pop()
			}
			counter++
		}
	})
}

// BenchmarkWaitFreeQueue_LargeQueue measures the performance with large number of elements
func BenchmarkWaitFreeQueue_LargeQueue(b *testing.B) {
	b.Run("Push10k", func(b *testing.B) {
		for b.Loop() {
			b.StopTimer()
			q := NewLockFreeQueue[int]()
			b.StartTimer()

			for j := 0; j < 10000; j++ {
				q.Push(j)
			}
		}
	})

	b.Run("Pop10k", func(b *testing.B) {
		for b.Loop() {
			b.StopTimer()
			q := NewLockFreeQueue[int]()
			for j := 0; j < 10000; j++ {
				q.Push(j)
			}
			b.StartTimer()

			for j := 0; j < 10000; j++ {
				_, _ = q.Pop()
			}
		}
	})
}

// BenchmarkWaitFreeQueue_QueueOperations compares different queue operations with varying sizes
func BenchmarkWaitFreeQueue_QueueOperations(b *testing.B) {
	// Push benchmarks with different queue sizes
	for _, size := range []int{10, 100, 1000, 10000} {
		b.Run("Push_Size_"+strconv.Itoa(size), func(b *testing.B) {
			q := NewLockFreeQueue[int]()
			// Pre-fill queue
			for i := 0; i < size; i++ {
				q.Push(i)
			}

			for b.Loop() {
				q.Push(1)
			}
		})
	}

	// Pop benchmarks with different queue sizes
	for _, size := range []int{10, 100, 1000, 10000} {
		b.Run("Pop_Size_"+strconv.Itoa(size), func(b *testing.B) {
			q := NewLockFreeQueue[int]()
			// Fill queue with size + b.N elements
			for i := 0; i < size+b.N; i++ {
				q.Push(i)
			}

			b.ResetTimer()
			for b.Loop() {
				_, _ = q.Pop()
			}
		})
	}
}

// BenchmarkWaitFreeQueue_MemoryUsage measures memory allocation patterns
func BenchmarkWaitFreeQueue_MemoryUsage(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		b.StopTimer()
		q := NewLockFreeQueue[int]()
		b.StartTimer()

		for j := 0; j < 10000; j++ {
			q.Push(j)
		}

		b.StopTimer()
		_ = q.Size() // Use the queue to prevent compiler optimizations
		b.StartTimer()
	}
}

// BenchmarkWaitFreeQueue_HighContention measures performance under high contention
func BenchmarkWaitFreeQueue_HighContention(b *testing.B) {
	for _, goroutines := range []int{2, 4, 8, 16, 32} {
		b.Run("Goroutines_"+strconv.Itoa(goroutines), func(b *testing.B) {
			q := NewLockFreeQueue[int]()

			b.SetParallelism(goroutines)
			b.RunParallel(func(pb *testing.PB) {
				counter := 0
				for pb.Next() {
					if counter%2 == 0 {
						q.Push(counter)
					} else {
						_, _ = q.Pop()
					}
					counter++
				}
			})
		})
	}
}

// =============================================================================
// Comparison Benchmarks: SliceQueue vs WaitFreeQueue
// =============================================================================

// BenchmarkQueueComparison_Push compares Push performance between SliceQueue and WaitFreeQueue
func BenchmarkQueueComparison_Push(b *testing.B) {
	b.Run("SliceQueue", func(b *testing.B) {
		q := NewQueue[int]()
		for b.Loop() {
			q.Push(1)
		}
	})

	b.Run("WaitFreeQueue", func(b *testing.B) {
		q := NewLockFreeQueue[int]()
		for b.Loop() {
			q.Push(1)
		}
	})
}

// BenchmarkQueueComparison_PushPop compares alternating Push/Pop performance
func BenchmarkQueueComparison_PushPop(b *testing.B) {
	b.Run("SliceQueue", func(b *testing.B) {
		q := NewQueue[int]()
		for b.Loop() {
			q.Push(1)
			_, _ = q.Pop()
		}
	})

	b.Run("WaitFreeQueue", func(b *testing.B) {
		q := NewLockFreeQueue[int]()
		for b.Loop() {
			q.Push(1)
			_, _ = q.Pop()
		}
	})
}

// BenchmarkQueueComparison_Parallel compares parallel performance between SliceQueue and WaitFreeQueue
func BenchmarkQueueComparison_Parallel(b *testing.B) {
	b.Run("SliceQueue", func(b *testing.B) {
		q := NewQueue[int]()
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				if counter%2 == 0 {
					q.Push(counter)
				} else {
					_, _ = q.Pop()
				}
				counter++
			}
		})
	})

	b.Run("WaitFreeQueue", func(b *testing.B) {
		q := NewLockFreeQueue[int]()
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				if counter%2 == 0 {
					q.Push(counter)
				} else {
					_, _ = q.Pop()
				}
				counter++
			}
		})
	})
}

// BenchmarkQueueComparison_ParallelHighContention compares performance under high contention
func BenchmarkQueueComparison_ParallelHighContention(b *testing.B) {
	for _, goroutines := range []int{4, 8, 16} {
		b.Run("Goroutines_"+strconv.Itoa(goroutines), func(b *testing.B) {
			b.Run("SliceQueue", func(b *testing.B) {
				q := NewQueue[int]()
				b.SetParallelism(goroutines)
				b.RunParallel(func(pb *testing.PB) {
					counter := 0
					for pb.Next() {
						if counter%2 == 0 {
							q.Push(counter)
						} else {
							_, _ = q.Pop()
						}
						counter++
					}
				})
			})

			b.Run("WaitFreeQueue", func(b *testing.B) {
				q := NewLockFreeQueue[int]()
				b.SetParallelism(goroutines)
				b.RunParallel(func(pb *testing.PB) {
					counter := 0
					for pb.Next() {
						if counter%2 == 0 {
							q.Push(counter)
						} else {
							_, _ = q.Pop()
						}
						counter++
					}
				})
			})
		})
	}
}

// =============================================================================
// ChannelQueue Benchmarks
// =============================================================================

// BenchmarkChannelQueue_Push measures the performance of Push operations
func BenchmarkChannelQueue_Push(b *testing.B) {
	q := NewChannelQueue[int](b.N + 1)

	for b.Loop() {
		q.Push(1)
	}
}

// BenchmarkChannelQueue_TryPush measures the performance of TryPush operations
func BenchmarkChannelQueue_TryPush(b *testing.B) {
	q := NewChannelQueue[int](b.N + 1)

	for b.Loop() {
		_ = q.TryPush(1)
	}
}

// BenchmarkChannelQueue_Pop measures the performance of Pop operations
func BenchmarkChannelQueue_Pop(b *testing.B) {
	q := NewChannelQueue[int](b.N + 1)

	// Setup: Push N items to the queue
	for i := 0; i < b.N; i++ {
		q.Push(i)
	}

	b.ResetTimer()
	for b.Loop() {
		_, _ = q.Pop()
	}
}

// BenchmarkChannelQueue_BlockingPop measures the performance of BlockingPop operations
func BenchmarkChannelQueue_BlockingPop(b *testing.B) {
	q := NewChannelQueue[int](b.N + 1)

	// Setup: Push N items to the queue
	for i := 0; i < b.N; i++ {
		q.Push(i)
	}

	b.ResetTimer()
	for b.Loop() {
		_ = q.BlockingPop()
	}
}

// BenchmarkChannelQueue_PushPop measures the performance of alternating Push and Pop operations
func BenchmarkChannelQueue_PushPop(b *testing.B) {
	q := NewChannelQueue[int](10)

	for b.Loop() {
		q.Push(1)
		_, _ = q.Pop()
	}
}

// BenchmarkChannelQueue_Size measures the performance of Size operations
func BenchmarkChannelQueue_Size(b *testing.B) {
	q := NewChannelQueue[int](1000)
	for i := 0; i < 500; i++ {
		q.Push(i)
	}

	for b.Loop() {
		_ = q.Size()
	}
}

// BenchmarkChannelQueue_IsEmpty measures the performance of IsEmpty operations
func BenchmarkChannelQueue_IsEmpty(b *testing.B) {
	q := NewChannelQueue[string](10)
	q.Push("test-item")

	for b.Loop() {
		_ = q.IsEmpty()
	}
}

// BenchmarkChannelQueue_IsFull measures the performance of IsFull operations
func BenchmarkChannelQueue_IsFull(b *testing.B) {
	q := NewChannelQueue[int](100)
	for i := 0; i < 100; i++ {
		q.Push(i)
	}

	for b.Loop() {
		_ = q.IsFull()
	}
}

// BenchmarkChannelQueue_PushParallel measures the performance of Push operations in parallel
func BenchmarkChannelQueue_PushParallel(b *testing.B) {
	q := NewChannelQueue[int](b.N + 1000)

	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			q.Push(counter)
			counter++
		}
	})
}

// BenchmarkChannelQueue_TryPushParallel measures the performance of TryPush operations in parallel
func BenchmarkChannelQueue_TryPushParallel(b *testing.B) {
	q := NewChannelQueue[int](b.N + 1000)

	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			_ = q.TryPush(counter)
			counter++
		}
	})
}

// BenchmarkChannelQueue_PopParallel measures the performance of Pop operations in parallel
func BenchmarkChannelQueue_PopParallel(b *testing.B) {
	q := NewChannelQueue[int](b.N * 10)

	// Pre-fill with enough items
	for i := 0; i < b.N*10; i++ {
		q.Push(i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = q.Pop()
		}
	})
}

// BenchmarkChannelQueue_PushPopParallel measures the performance of Push and Pop operations in parallel
func BenchmarkChannelQueue_PushPopParallel(b *testing.B) {
	q := NewChannelQueue[int](1000)

	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			if counter%2 == 0 {
				_ = q.TryPush(counter)
			} else {
				_, _ = q.Pop()
			}
			counter++
		}
	})
}

// BenchmarkChannelQueue_HighContention measures performance under high contention
func BenchmarkChannelQueue_HighContention(b *testing.B) {
	for _, goroutines := range []int{2, 4, 8, 16, 32} {
		b.Run("Goroutines_"+strconv.Itoa(goroutines), func(b *testing.B) {
			q := NewChannelQueue[int](1000)

			b.SetParallelism(goroutines)
			b.RunParallel(func(pb *testing.PB) {
				counter := 0
				for pb.Next() {
					if counter%2 == 0 {
						_ = q.TryPush(counter)
					} else {
						_, _ = q.Pop()
					}
					counter++
				}
			})
		})
	}
}

// BenchmarkChannelQueue_DifferentCapacities measures performance with different capacities
func BenchmarkChannelQueue_DifferentCapacities(b *testing.B) {
	for _, capacity := range []int{10, 100, 1000, 10000} {
		b.Run("Capacity_"+strconv.Itoa(capacity), func(b *testing.B) {
			q := NewChannelQueue[int](capacity)

			for b.Loop() {
				_ = q.TryPush(1)
				_, _ = q.Pop()
			}
		})
	}
}

// BenchmarkChannelQueue_MemoryUsage measures memory allocation patterns
func BenchmarkChannelQueue_MemoryUsage(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		b.StopTimer()
		q := NewChannelQueue[int](1000)
		b.StartTimer()

		for j := 0; j < 1000; j++ {
			q.Push(j)
		}

		b.StopTimer()
		_ = q.Size() // Use the queue to prevent compiler optimizations
		b.StartTimer()
	}
}

// =============================================================================
// Comparison Benchmarks: All Queue Implementations
// =============================================================================

// BenchmarkAllQueues_Push compares Push performance across all implementations
func BenchmarkAllQueues_Push(b *testing.B) {
	b.Run("SliceQueue", func(b *testing.B) {
		q := NewQueue[int]()
		for b.Loop() {
			q.Push(1)
		}
	})

	b.Run("LockFreeQueue", func(b *testing.B) {
		q := NewLockFreeQueue[int]()
		for b.Loop() {
			q.Push(1)
		}
	})

	b.Run("ChannelQueue", func(b *testing.B) {
		q := NewChannelQueue[int](b.N + 1)
		for b.Loop() {
			q.Push(1)
		}
	})
}

// BenchmarkAllQueues_PushPop compares alternating Push/Pop performance
func BenchmarkAllQueues_PushPop(b *testing.B) {
	b.Run("SliceQueue", func(b *testing.B) {
		q := NewQueue[int]()
		for b.Loop() {
			q.Push(1)
			_, _ = q.Pop()
		}
	})

	b.Run("LockFreeQueue", func(b *testing.B) {
		q := NewLockFreeQueue[int]()
		for b.Loop() {
			q.Push(1)
			_, _ = q.Pop()
		}
	})

	b.Run("ChannelQueue", func(b *testing.B) {
		q := NewChannelQueue[int](10)
		for b.Loop() {
			q.Push(1)
			_, _ = q.Pop()
		}
	})
}

// BenchmarkAllQueues_Parallel compares parallel performance across all implementations
func BenchmarkAllQueues_Parallel(b *testing.B) {
	b.Run("SliceQueue", func(b *testing.B) {
		q := NewQueue[int]()
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				if counter%2 == 0 {
					q.Push(counter)
				} else {
					_, _ = q.Pop()
				}
				counter++
			}
		})
	})

	b.Run("LockFreeQueue", func(b *testing.B) {
		q := NewLockFreeQueue[int]()
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				if counter%2 == 0 {
					q.Push(counter)
				} else {
					_, _ = q.Pop()
				}
				counter++
			}
		})
	})

	b.Run("ChannelQueue", func(b *testing.B) {
		q := NewChannelQueue[int](1000)
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				if counter%2 == 0 {
					_ = q.TryPush(counter)
				} else {
					_, _ = q.Pop()
				}
				counter++
			}
		})
	})
}

// BenchmarkAllQueues_ParallelHighContention compares performance under high contention
func BenchmarkAllQueues_ParallelHighContention(b *testing.B) {
	for _, goroutines := range []int{4, 8, 16} {
		b.Run("Goroutines_"+strconv.Itoa(goroutines), func(b *testing.B) {
			b.Run("SliceQueue", func(b *testing.B) {
				q := NewQueue[int]()
				b.SetParallelism(goroutines)
				b.RunParallel(func(pb *testing.PB) {
					counter := 0
					for pb.Next() {
						if counter%2 == 0 {
							q.Push(counter)
						} else {
							_, _ = q.Pop()
						}
						counter++
					}
				})
			})

			b.Run("LockFreeQueue", func(b *testing.B) {
				q := NewLockFreeQueue[int]()
				b.SetParallelism(goroutines)
				b.RunParallel(func(pb *testing.PB) {
					counter := 0
					for pb.Next() {
						if counter%2 == 0 {
							q.Push(counter)
						} else {
							_, _ = q.Pop()
						}
						counter++
					}
				})
			})

			b.Run("ChannelQueue", func(b *testing.B) {
				q := NewChannelQueue[int](1000)
				b.SetParallelism(goroutines)
				b.RunParallel(func(pb *testing.PB) {
					counter := 0
					for pb.Next() {
						if counter%2 == 0 {
							_ = q.TryPush(counter)
						} else {
							_, _ = q.Pop()
						}
						counter++
					}
				})
			})
		})
	}
}
