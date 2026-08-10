package algo

import (
	"fmt"
	"math/rand"
	"testing"
)

// BenchmarkUnionFind_New benchmarks the creation of UnionFind with different sizes
func BenchmarkUnionFind_New(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = NewUnionFind(size)
			}
		})
	}
}

// BenchmarkUnionFind_Find benchmarks the Find operation
func BenchmarkUnionFind_Find(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			// Create the UnionFind structure with specified size
			uf := NewUnionFind(size)

			// Create some connections to make the structure non-trivial
			// Connect about 20% of elements randomly
			numConnections := size / 5
			for i := 0; i < numConnections; i++ {
				p := rand.Intn(size)
				q := rand.Intn(size)
				uf.Union(p, q)
			}

			b.ResetTimer() // Reset timer after setup

			// Benchmark the Find operation
			for i := 0; i < b.N; i++ {
				_ = uf.Find(rand.Intn(size))
			}
		})
	}
}

// BenchmarkUnionFind_Union benchmarks the Union operation
func BenchmarkUnionFind_Union(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			// Create the UnionFind structure with specified size
			uf := NewUnionFind(size)

			b.ResetTimer() // Reset timer after setup

			// Benchmark Union operations with random elements
			for i := 0; i < b.N; i++ {
				p := rand.Intn(size)
				q := rand.Intn(size)
				uf.Union(p, q)
			}
		})
	}
}

// BenchmarkUnionFind_Connected benchmarks the Connected operation
func BenchmarkUnionFind_Connected(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			// Create the UnionFind structure with specified size
			uf := NewUnionFind(size)

			// Create some connections to make the structure non-trivial
			// Connect about 20% of elements randomly
			numConnections := size / 5
			for i := 0; i < numConnections; i++ {
				p := rand.Intn(size)
				q := rand.Intn(size)
				uf.Union(p, q)
			}

			b.ResetTimer() // Reset timer after setup

			// Benchmark the Connected operation
			for i := 0; i < b.N; i++ {
				p := rand.Intn(size)
				q := rand.Intn(size)
				_ = uf.Connected(p, q)
			}
		})
	}
}

// BenchmarkUnionFind_Operations benchmarks a mixed workload of operations
func BenchmarkUnionFind_Operations(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			// Create the UnionFind structure with specified size
			uf := NewUnionFind(size)

			b.ResetTimer() // Reset timer after setup

			// Benchmark a mix of operations
			for i := 0; i < b.N; i++ {
				op := rand.Intn(3) // 0: Union, 1: Find, 2: Connected
				p := rand.Intn(size)
				q := rand.Intn(size)

				switch op {
				case 0:
					uf.Union(p, q)
				case 1:
					_ = uf.Find(p)
				case 2:
					_ = uf.Connected(p, q)
				}
			}
		})
	}
}

// BenchmarkUnionFind_WorstCase benchmarks the worst-case scenario
// where we create a deep tree and then perform finds
func BenchmarkUnionFind_WorstCase(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			// Create the UnionFind structure
			uf := NewUnionFind(size)

			// Create a deep tree (worst case for find without path compression)
			// We're specifically bypassing the rank-based balancing to create a tall tree
			for i := 1; i < size; i++ {
				uf.parent[i] = i - 1
			}

			b.ResetTimer() // Reset timer after setup

			// Benchmark Find on the deepest leaf
			for i := 0; i < b.N; i++ {
				_ = uf.Find(size - 1)
			}
		})
	}
}

// BenchmarkUnionFind_Parallel benchmarks parallel access to UnionFind
func BenchmarkUnionFind_Parallel(b *testing.B) {
	size := 10000
	uf := NewUnionFind(size)

	// Create some initial connections
	for i := 0; i < size/5; i++ {
		p := rand.Intn(size)
		q := rand.Intn(size)
		uf.Union(p, q)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		// Create a local random source to avoid contention
		localRand := rand.New(rand.NewSource(rand.Int63()))

		for pb.Next() {
			op := localRand.Intn(3) // 0: Union, 1: Find, 2: Connected
			p := localRand.Intn(size)
			q := localRand.Intn(size)

			switch op {
			case 0:
				uf.Union(p, q)
			case 1:
				_ = uf.Find(p)
			case 2:
				_ = uf.Connected(p, q)
			}
		}
	})
}
