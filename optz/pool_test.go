package optz

import (
	"sync"
	"testing"
)

type testItem struct {
	value int
}

func TestRandomBoundedPool(t *testing.T) {
	backup := func() *testItem {
		return &testItem{value: -1}
	}

	t.Run("basic operations", func(t *testing.T) {
		pool := newRandomBoundedPool[*testItem](3, 9, backup)

		// Test Add
		item1 := &testItem{value: 1}
		if !pool.Add(item1) {
			t.Error("Failed to add item to pool")
		}

		// Test Get and Put
		got := pool.Get()
		if got == nil {
			t.Error("Expected non-nil item from Get")
		}
		pool.Put(got)
	})

	t.Run("concurrent operations", func(t *testing.T) {
		pool := newRandomBoundedPool[*testItem](3, 90, backup)
		var wg sync.WaitGroup
		numGoroutines := 10

		// Pre-fill pool
		for i := 0; i < 90; i++ {
			pool.Add(&testItem{value: i})
		}

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					item := pool.Get()
					if item == nil {
						t.Error("Got nil item from pool")
					}
					pool.Put(item)
				}
			}()
		}
		wg.Wait()
	})
}

func TestBoundedPool(t *testing.T) {
	backup := func() *testItem {
		return &testItem{value: -1}
	}

	t.Run("capacity limits", func(t *testing.T) {
		pool := newBoundedPool[*testItem](5, backup)

		// Test adding items up to capacity
		for i := 0; i < 5; i++ {
			if !pool.Add(&testItem{value: i}) {
				t.Errorf("Failed to add item %d within capacity", i)
			}
		}

		// Test adding beyond capacity
		if pool.Add(&testItem{value: 100}) {
			t.Error("Should not be able to add item beyond capacity")
		}
	})

	t.Run("get put operations", func(t *testing.T) {
		pool := newBoundedPool[*testItem](5, backup)
		item := &testItem{value: 42}

		pool.Add(item)
		got := pool.Get()
		if got == nil || got.value != 42 {
			t.Errorf("Expected item with value 42, got %+v", got)
		}

		pool.Put(item)
	})
}

func TestIndexer(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		idx := newIndexer(32)

		// Test Use
		index := idx.Use()
		if index < 0 || index >= 32 {
			t.Errorf("Index out of range: %d", index)
		}

		// Test Release
		idx.Release(index)
	})

	t.Run("concurrent operations", func(t *testing.T) {
		idx := newIndexer(100)
		var wg sync.WaitGroup
		numGoroutines := 10

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					index := idx.Use()
					if index < 0 || index >= 100 {
						t.Errorf("Invalid index: %d", index)
					}
					idx.Release(index)
				}
			}()
		}
		wg.Wait()
	})
}

func TestBitOperations(t *testing.T) {
	idx := newIndexer(32)

	t.Run("firstZeroBit", func(t *testing.T) {
		cases := []struct {
			input    uint32
			expected uint32
		}{
			{0, 0},
			{1, 1},
			{2, 0},
			{3, 2},
			{4, 0},
		}

		for _, tc := range cases {
			got := idx.firstZeroBit(tc.input)
			if got != tc.expected {
				t.Errorf("firstZeroBit(%d) = %d; want %d", tc.input, got, tc.expected)
			}
		}
	})

	t.Run("bitCount", func(t *testing.T) {
		cases := []struct {
			input    uint32
			expected uint32
		}{
			{0, 0},
			{1, 1},
			{3, 2},
			{7, 3},
			{15, 4},
		}

		for _, tc := range cases {
			got := idx.bitCount(tc.input)
			if got != tc.expected {
				t.Errorf("bitCount(%d) = %d; want %d", tc.input, got, tc.expected)
			}
		}
	})
}

func BenchmarkRandomBoundedPool(b *testing.B) {
	backup := func() *testItem {
		return &testItem{value: -1}
	}
	pool := newRandomBoundedPool[*testItem](3, 1000, backup)

	// Pre-fill pool
	for i := 0; i < 1000; i++ {
		pool.Add(&testItem{value: i})
	}

	b.Run("sequential-get-put", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			item := pool.Get()
			pool.Put(item)
		}
	})

	b.Run("parallel-get-put", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				item := pool.Get()
				pool.Put(item)
			}
		})
	})
}

func BenchmarkBoundedPool(b *testing.B) {
	backup := func() *testItem {
		return &testItem{value: -1}
	}
	pool := newBoundedPool[*testItem](1000, backup)

	// Pre-fill pool
	for i := 0; i < 1000; i++ {
		pool.Add(&testItem{value: i})
	}

	b.Run("sequential-get-put", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			item := pool.Get()
			pool.Put(item)
		}
	})

	b.Run("parallel-get-put", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				item := pool.Get()
				pool.Put(item)
			}
		})
	})
}

func BenchmarkIndexer(b *testing.B) {
	idx := newIndexer(1000)

	b.Run("sequential-use-release", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			index := idx.Use()
			idx.Release(index)
		}
	})

	b.Run("parallel-use-release", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				index := idx.Use()
				idx.Release(index)
			}
		})
	})

	b.Run("bit-operations", func(b *testing.B) {
		var v uint32 = 0xFFFF0000
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			idx.firstZeroBit(v)
			idx.bitCount(v)
		}
	})
}
