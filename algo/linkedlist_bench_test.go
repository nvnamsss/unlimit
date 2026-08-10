package algo

import (
	"container/list"
	"fmt"
	"testing"
)

// Benchmark sizes to test with
var benchSizes = []int{100, 1000, 10000}

// BenchmarkLinkedList_Add tests adding elements to the end of a LinkedList
func BenchmarkLinkedList_Add(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				ll := NewLinkedList[int]()
				for j := 0; j < size; j++ {
					ll.Add(j)
				}
			}
		})
	}
}

// BenchmarkLinkedList_AddFront tests adding elements to the front of a LinkedList
func BenchmarkLinkedList_AddFront(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				ll := NewLinkedList[int]()
				for j := 0; j < size; j++ {
					ll.AddFront(j)
				}
			}
		})
	}
}

// BenchmarkLinkedList_Iteration tests iterating through all elements in a LinkedList
func BenchmarkLinkedList_Iteration(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			// Setup
			ll := NewLinkedList[int]()
			for i := 0; i < size; i++ {
				ll.Add(i)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				sum := 0
				ll.ForEach(func(data int) {
					sum += data
				})
				// Prevent compiler optimization
				if sum < 0 && !testing.Short() {
					b.Fatal("Unexpected result")
				}
			}
		})
	}
}

// BenchmarkLinkedList_Get tests retrieving elements at specific positions
func BenchmarkLinkedList_Get(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			// Setup
			ll := NewLinkedList[int]()
			for i := 0; i < size; i++ {
				ll.Add(i)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Get elements at different positions (beginning, middle, end)
				var val int
				step := size / 10
				if step == 0 {
					step = 1
				}

				for j := 0; j < size; j += step {
					val, _ = ll.Get(j)
				}

				// Prevent compiler optimization
				if val < 0 && !testing.Short() {
					b.Fatal("Unexpected result")
				}
			}
		})
	}
}

// BenchmarkLinkedList_Remove tests removing elements from a LinkedList
func BenchmarkLinkedList_Remove(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Setup
				ll := NewLinkedList[int]()
				for j := 0; j < size; j++ {
					ll.Add(j)
				}

				b.StartTimer()
				// Remove half of the elements
				for j := 0; j < size/2; j++ {
					ll.Remove(j * 2)
				}
				b.StopTimer()
			}
		})
	}
}

// BenchmarkStdList_PushBack tests adding elements to the end of container/list
func BenchmarkStdList_PushBack(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				l := list.New()
				for j := 0; j < size; j++ {
					l.PushBack(j)
				}
			}
		})
	}
}

// BenchmarkStdList_PushFront tests adding elements to the front of container/list
func BenchmarkStdList_PushFront(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				l := list.New()
				for j := 0; j < size; j++ {
					l.PushFront(j)
				}
			}
		})
	}
}

// BenchmarkStdList_Iteration tests iterating through all elements in container/list
func BenchmarkStdList_Iteration(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			// Setup
			l := list.New()
			for i := 0; i < size; i++ {
				l.PushBack(i)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				sum := 0
				for e := l.Front(); e != nil; e = e.Next() {
					sum += e.Value.(int)
				}
				// Prevent compiler optimization
				if sum < 0 && !testing.Short() {
					b.Fatal("Unexpected result")
				}
			}
		})
	}
}

// BenchmarkStdList_Get tests retrieving elements at specific positions
func BenchmarkStdList_Get(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			// Setup
			l := list.New()
			elements := make([]*list.Element, size)
			for i := 0; i < size; i++ {
				elements[i] = l.PushBack(i)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				var val interface{}
				step := size / 10
				if step == 0 {
					step = 1
				}

				for j := 0; j < size; j += step {
					val = elements[j].Value
				}

				// Prevent compiler optimization
				if val.(int) < 0 && !testing.Short() {
					b.Fatal("Unexpected result")
				}
			}
		})
	}
}

// BenchmarkStdList_Remove tests removing elements from container/list
func BenchmarkStdList_Remove(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Setup
				l := list.New()
				elements := make([]*list.Element, size)
				for j := 0; j < size; j++ {
					elements[j] = l.PushBack(j)
				}

				b.StartTimer()
				// Remove half of the elements
				for j := 0; j < size/2; j++ {
					l.Remove(elements[j*2])
				}
				b.StopTimer()
			}
		})
	}
}

// BenchmarkComparison_AddOperation compares adding elements to both list implementations
func BenchmarkComparison_AddOperation(b *testing.B) {
	size := 10000

	b.Run("LinkedList.Add", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ll := NewLinkedList[int]()
			for j := 0; j < size; j++ {
				ll.Add(j)
			}
		}
	})

	b.Run("StdList.PushBack", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			l := list.New()
			for j := 0; j < size; j++ {
				l.PushBack(j)
			}
		}
	})
}

// BenchmarkComparison_IterationOperation compares iteration performance
func BenchmarkComparison_IterationOperation(b *testing.B) {
	size := 10000

	// Setup for LinkedList
	ll := NewLinkedList[int]()
	for i := 0; i < size; i++ {
		ll.Add(i)
	}

	// Setup for std/list
	l := list.New()
	for i := 0; i < size; i++ {
		l.PushBack(i)
	}

	b.Run("LinkedList.ForEach", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sum := 0
			ll.ForEach(func(data int) {
				sum += data
			})
			// Prevent compiler optimization
			if sum < 0 && !testing.Short() {
				b.Fatal("Unexpected result")
			}
		}
	})

	b.Run("StdList.Iteration", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sum := 0
			for e := l.Front(); e != nil; e = e.Next() {
				sum += e.Value.(int)
			}
			// Prevent compiler optimization
			if sum < 0 && !testing.Short() {
				b.Fatal("Unexpected result")
			}
		}
	})
}
