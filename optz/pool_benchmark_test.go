package optz

import (
	"sync"
	"testing"
)

func BenchmarkPoolComparison(b *testing.B) {
	// Setup pools with same capacity
	const poolSize = 1000

	// Custom pools
	customPool := newRandomBoundedPool[*testItem](3, poolSize, func() *testItem {
		return &testItem{value: -1}
	})
	boundPool := newBoundedPool[*testItem](poolSize, func() *testItem {
		return &testItem{value: -1}
	})

	// Native sync.Pool
	nativePool := &sync.Pool{
		New: func() interface{} {
			return &testItem{value: -1}
		},
	}

	// Pre-fill pools
	for i := 0; i < poolSize; i++ {
		item := &testItem{value: i}
		customPool.Add(item)
		boundPool.Add(item)
		nativePool.Put(&testItem{value: i})
	}

	b.Run("sequential-operations", func(b *testing.B) {
		b.Run("sync.Pool", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				item := nativePool.Get().(*testItem)
				nativePool.Put(item)
			}
		})

		b.Run("randomBoundedPool", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				item := customPool.Get()
				customPool.Put(item)
			}
		})

		b.Run("boundedPool", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				item := boundPool.Get()
				boundPool.Put(item)
			}
		})
	})

	b.Run("parallel-operations", func(b *testing.B) {
		b.Run("sync.Pool", func(b *testing.B) {
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					item := nativePool.Get().(*testItem)
					nativePool.Put(item)
				}
			})
		})

		b.Run("randomBoundedPool", func(b *testing.B) {
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					item := customPool.Get()
					customPool.Put(item)
				}
			})
		})

		b.Run("boundedPool", func(b *testing.B) {
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					item := boundPool.Get()
					boundPool.Put(item)
				}
			})
		})
	})

	b.Run("high-contention", func(b *testing.B) {
		workers := 100
		opsPerWorker := b.N / workers

		b.Run("sync.Pool", func(b *testing.B) {
			b.ResetTimer()
			var wg sync.WaitGroup
			for i := 0; i < workers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < opsPerWorker; j++ {
						item := nativePool.Get().(*testItem)
						item.value++
						nativePool.Put(item)
					}
				}()
			}
			wg.Wait()
		})

		b.Run("randomBoundedPool", func(b *testing.B) {
			b.ResetTimer()
			var wg sync.WaitGroup
			for i := 0; i < workers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < opsPerWorker; j++ {
						item := customPool.Get()
						item.value++
						customPool.Put(item)
					}
				}()
			}
			wg.Wait()
		})
	})
}
