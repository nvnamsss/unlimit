package algo

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnionFind_New(t *testing.T) {
	// Test with different sizes
	sizes := []int{1, 5, 100}

	for _, size := range sizes {
		t.Run("Size_"+string(rune('0'+size)), func(t *testing.T) {
			uf := NewUnionFind(size)

			// Check proper initialization
			assert.Equal(t, size, len(uf.parent), "Parent slice should have correct size")
			assert.Equal(t, size, len(uf.rank), "Rank slice should have correct size")
			assert.Equal(t, size, uf.Count(), "Count should equal initial size")

			// Check that each element is its own parent initially
			for i := 0; i < size; i++ {
				assert.Equal(t, i, uf.parent[i], "Each element should initially be its own parent")
				assert.Equal(t, 0, uf.rank[i], "Initial rank should be 0")
			}
		})
	}

	// Test zero size
	t.Run("Size_0", func(t *testing.T) {
		uf := NewUnionFind(0)
		assert.Equal(t, 0, len(uf.parent))
		assert.Equal(t, 0, len(uf.rank))
		assert.Equal(t, 0, uf.Count())
	})
}

func TestUnionFind_Find(t *testing.T) {
	t.Run("Simple_Find", func(t *testing.T) {
		uf := NewUnionFind(10)
		for i := 0; i < 10; i++ {
			assert.Equal(t, i, uf.Find(i), "Initially, Find(i) should return i")
		}
	})

	t.Run("Find_After_Union", func(t *testing.T) {
		uf := NewUnionFind(10)
		uf.Union(1, 2)
		uf.Union(2, 3)
		uf.Union(3, 4)

		root := uf.Find(1)
		assert.Equal(t, root, uf.Find(2), "Elements 1 and 2 should have the same root")
		assert.Equal(t, root, uf.Find(3), "Elements 1 and 3 should have the same root")
		assert.Equal(t, root, uf.Find(4), "Elements 1 and 4 should have the same root")

		// Verify elements in different sets have different roots
		assert.NotEqual(t, root, uf.Find(0), "Elements 0 and 1 should have different roots")
		assert.NotEqual(t, root, uf.Find(5), "Elements 1 and 5 should have different roots")
	})

	t.Run("Find_Path_Compression", func(t *testing.T) {
		uf := NewUnionFind(5)

		// Create a tall tree without path compression
		uf.parent[1] = 0
		uf.parent[2] = 1
		uf.parent[3] = 2
		uf.parent[4] = 3

		// Find should compress paths
		assert.Equal(t, 0, uf.Find(4), "Root should be 0")

		// After compression, 4's parent should be the root (0)
		assert.Equal(t, 0, uf.parent[4], "Path compression should set parent to root")
	})

	t.Run("Find_Out_Of_Bounds", func(t *testing.T) {
		uf := NewUnionFind(5)

		// Test negative index
		assert.Panics(t, func() { uf.Find(-1) }, "Should panic with negative index")

		// Test index >= size
		assert.Panics(t, func() { uf.Find(5) }, "Should panic with index >= size")
	})
}

func TestUnionFind_Union(t *testing.T) {
	t.Run("Basic_Union", func(t *testing.T) {
		uf := NewUnionFind(10)
		uf.Union(1, 2)
		assert.True(t, uf.Connected(1, 2), "Elements should be connected after union")
		assert.Equal(t, 9, uf.Count(), "Count should decrease after union")
	})

	t.Run("Redundant_Union", func(t *testing.T) {
		uf := NewUnionFind(10)
		initialCount := uf.Count()
		uf.Union(1, 2)
		assert.Equal(t, initialCount-1, uf.Count())

		// Union same elements again
		uf.Union(1, 2)
		assert.Equal(t, initialCount-1, uf.Count(), "Redundant union should not decrease count")

		// Union two elements already in same component
		uf.Union(2, 1)
		assert.Equal(t, initialCount-1, uf.Count(), "Union of connected elements should not decrease count")
	})

	t.Run("Union_By_Rank", func(t *testing.T) {
		uf := NewUnionFind(10)

		// Create two trees with different ranks
		// First tree: 1-2-3 (rank 2)
		uf.Union(1, 2)
		uf.Union(2, 3)

		// Second tree: 4-5 (rank 1)
		uf.Union(4, 5)

		// Get roots before merger
		root1 := uf.Find(1)
		root2 := uf.Find(4)

		// Union the two trees
		uf.Union(3, 5)

		// Root of smaller rank should point to root of larger rank
		assert.Equal(t, uf.Find(root2), root1, "Root of smaller rank tree should point to root of larger rank tree")
	})

	t.Run("Union_Same_Rank", func(t *testing.T) {
		uf := NewUnionFind(10)

		// Create two single-node trees (rank 0)
		root1 := 1
		root2 := 2

		// Initial ranks should be equal
		assert.Equal(t, uf.rank[root1], uf.rank[root2])

		// Union them
		uf.Union(root1, root2)

		// Find actual roots after union
		newRoot1 := uf.Find(root1)
		newRoot2 := uf.Find(root2)

		// Roots should be the same
		assert.Equal(t, newRoot1, newRoot2)

		// Rank of the new root should be increased
		assert.Equal(t, 1, uf.rank[newRoot1], "Rank should increase when unioning trees of same rank")
	})

	t.Run("Union_Out_Of_Bounds", func(t *testing.T) {
		uf := NewUnionFind(5)

		// Test negative indices
		assert.Panics(t, func() { uf.Union(-1, 2) })
		assert.Panics(t, func() { uf.Union(1, -1) })

		// Test indices >= size
		assert.Panics(t, func() { uf.Union(5, 2) })
		assert.Panics(t, func() { uf.Union(1, 5) })
	})
}

func TestUnionFind_Connected(t *testing.T) {
	t.Run("Initial_State", func(t *testing.T) {
		uf := NewUnionFind(5)

		// Initially, elements are only connected to themselves
		for i := 0; i < 5; i++ {
			assert.True(t, uf.Connected(i, i), "Element should be connected to itself")

			// Check no other connections
			for j := 0; j < 5; j++ {
				if i != j {
					assert.False(t, uf.Connected(i, j), "Different elements should not be connected initially")
				}
			}
		}
	})

	t.Run("After_Union", func(t *testing.T) {
		uf := NewUnionFind(10)

		// Create two components: {0,1,2,3} and {4,5}
		uf.Union(0, 1)
		uf.Union(1, 2)
		uf.Union(2, 3)
		uf.Union(4, 5)

		// Check connections within components
		assert.True(t, uf.Connected(0, 3))
		assert.True(t, uf.Connected(1, 3))
		assert.True(t, uf.Connected(2, 0))
		assert.True(t, uf.Connected(4, 5))

		// Check connections between components (should be false)
		assert.False(t, uf.Connected(0, 4))
		assert.False(t, uf.Connected(3, 5))

		// Check with unconnected elements
		assert.False(t, uf.Connected(0, 6))
		assert.False(t, uf.Connected(5, 7))
	})

	t.Run("Connected_Out_Of_Bounds", func(t *testing.T) {
		uf := NewUnionFind(5)

		assert.Panics(t, func() { uf.Connected(-1, 2) })
		assert.Panics(t, func() { uf.Connected(1, -1) })
		assert.Panics(t, func() { uf.Connected(5, 2) })
		assert.Panics(t, func() { uf.Connected(1, 5) })
	})
}

func TestUnionFind_Count(t *testing.T) {
	t.Run("Initial_Count", func(t *testing.T) {
		sizes := []int{0, 1, 10, 100}
		for _, size := range sizes {
			uf := NewUnionFind(size)
			assert.Equal(t, size, uf.Count(), "Initial count should equal size")
		}
	})

	t.Run("Count_After_Union", func(t *testing.T) {
		uf := NewUnionFind(10)
		initialCount := uf.Count()

		// Connect 3 components
		uf.Union(0, 1)
		assert.Equal(t, initialCount-1, uf.Count())

		uf.Union(2, 3)
		assert.Equal(t, initialCount-2, uf.Count())

		uf.Union(4, 5)
		assert.Equal(t, initialCount-3, uf.Count())

		// Join two existing components
		uf.Union(1, 3)
		assert.Equal(t, initialCount-4, uf.Count(), "Joining existing components should decrease count by 1")

		// Redundant union should not change count
		uf.Union(0, 2)
		assert.Equal(t, initialCount-4, uf.Count(), "Redundant union should not change count")
	})
}

func TestUnionFind_Complex(t *testing.T) {
	t.Run("Large_Component", func(t *testing.T) {
		size := 100
		uf := NewUnionFind(size)

		// Create one large component
		for i := 1; i < size; i++ {
			uf.Union(0, i)
		}

		// All elements should be connected to 0
		for i := 1; i < size; i++ {
			assert.True(t, uf.Connected(0, i))
		}

		assert.Equal(t, 1, uf.Count(), "Should have exactly 1 component")
	})

	t.Run("Multiple_Operations", func(t *testing.T) {
		uf := NewUnionFind(10)

		// Create a few unions
		uf.Union(0, 1)
		uf.Union(2, 3)
		uf.Union(4, 5)
		uf.Union(6, 7)
		uf.Union(8, 9)

		// Now we have 5 components
		assert.Equal(t, 5, uf.Count())

		// Join some components
		uf.Union(1, 3) // Join {0,1} and {2,3}
		uf.Union(5, 7) // Join {4,5} and {6,7}

		// Now we have 3 components
		assert.Equal(t, 3, uf.Count())

		// Verify connections
		assert.True(t, uf.Connected(0, 3))
		assert.True(t, uf.Connected(4, 7))
		assert.False(t, uf.Connected(0, 4))
		assert.False(t, uf.Connected(0, 8))

		// Join remaining components
		uf.Union(0, 4)
		uf.Union(4, 8)

		// Should have 1 component
		assert.Equal(t, 1, uf.Count())

		// Everything should be connected
		for i := 0; i < 9; i++ {
			for j := i + 1; j < 10; j++ {
				assert.True(t, uf.Connected(i, j))
			}
		}
	})
}

func TestUnionFind_Basic(t *testing.T) {
	uf := NewUnionFind(10)
	uf.Union(1, 2)
	uf.Union(2, 3)
	assert.Equal(t, true, uf.Connected(1, 3)) // Should be true
}

func TestUnionFind_Parallel(t *testing.T) {
	t.Run("High_Contention_Test", func(t *testing.T) {
		// Create a moderately sized union-find structure
		const size = 10
		uf := NewUnionFind(size)

		// Number of goroutines performing operations in parallel
		const numWorkers = 100

		// Create a specific structure that's vulnerable to race conditions
		// First, we'll make several separate "chains" of elements
		for i := 0; i <= size-2; i += 2 {
			uf.Union(i, i+1)
		}

		// At this point, we have size/2 separate components
		// Each component is a pair like (0,1), (2,3), etc.
		expectedCount := size / 2
		assert.Equal(t, expectedCount, uf.Count())

		// Now create a high-contention scenario by having multiple goroutines
		// try to connect these components in different orders simultaneously
		var wg sync.WaitGroup

		// Synchronization point to maximize contention
		ready := make(chan struct{})

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)

			// Each worker connects components in different patterns
			go func(workerID int) {
				defer wg.Done()

				// Wait for all goroutines to be ready
				<-ready

				// Different workers follow different patterns to maximize contention
				start := (workerID * 4) % size

				// Try to create conflicts by connecting different sets
				// This creates a situation where multiple elements could become
				// the canonical element if parallelism isn't handled properly
				for j := 0; j < 10; j++ {
					idx1 := (start + j*17) % size
					idx2 := (start + j*29) % size

					// Connect these two elements, potentially creating a race condition
					// if the implementation isn't thread-safe
					uf.Union(idx1, idx2)

					// Also try some finds in between to add more contention
					_ = uf.Find((idx1 + idx2) % size)
				}
			}(i)
		}

		// Start all goroutines simultaneously to maximize contention
		close(ready)

		// Wait for all operations to complete
		wg.Wait()

		// Now verify the data structure integrity

		// 1. Count should be less than what we started with
		assert.Less(t, uf.Count(), expectedCount, "After parallel unions, component count should decrease")

		// 2. For each element, verify it has a valid canonical element (no cycles)
		for i := 0; i < size; i++ {
			root := uf.Find(i)
			// The root should be its own parent
			assert.Equal(t, root, uf.parent[root], "Root element should be its own parent")
		}

		// 3. Verify consistency: if a and b are connected, and b and c are connected,
		// then a and c must be connected
		for i := 0; i < size; i += 50 { // Sample some elements to avoid too much testing
			for j := i + 1; j < size; j += 50 {
				for k := j + 1; k < size; k += 50 {
					if uf.Connected(i, j) && uf.Connected(j, k) {
						assert.True(t, uf.Connected(i, k),
							"Transitivity violated: %d and %d connected, %d and %d connected, but %d and %d not connected",
							i, j, j, k, i, k)
					}
				}
			}
		}
	})

	t.Run("Conflicting_Canonical_Elements", func(t *testing.T) {
		// This test specifically targets the scenario where different threads
		// might disagree on canonical elements
		const size = 200
		uf := NewUnionFind(size)

		// Create initial structure: several isolated trees
		// Tree 1: 0->10->20->30
		uf.parent[10] = 0
		uf.parent[20] = 10
		uf.parent[30] = 20
		uf.rank[0] = 3

		// Tree 2: 1->11->21->31
		uf.parent[11] = 1
		uf.parent[21] = 11
		uf.parent[31] = 21
		uf.rank[1] = 3

		// Update count to reflect our manual modifications
		uf.count = size - 6

		// Now create a situation where multiple threads try to union
		// these trees in different ways simultaneously
		var wg sync.WaitGroup
		ready := make(chan struct{})

		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				<-ready

				// Different threads connecting different parts of the trees
				// This can cause race conditions in determining the canonical element
				switch id % 5 {
				case 0:
					uf.Union(0, 1) // Connecting roots directly
				case 1:
					uf.Union(30, 31) // Connecting leaf nodes
				case 2:
					uf.Union(10, 21) // Connecting middle nodes
				case 3:
					uf.Union(20, 11) // Connecting different level nodes
				case 4:
					// Just do finds to add more contention during the unions
					_ = uf.Find(10)
					_ = uf.Find(11)
					_ = uf.Find(20)
					_ = uf.Find(21)
				}
			}(i)
		}

		// Start all goroutines at once
		close(ready)
		wg.Wait()

		// After all operations, 0, 1, 10, 11, 20, 21, 30, 31 should all be connected
		// Let's verify they all have the same root
		root := uf.Find(0)
		connectPoints := []int{0, 1, 10, 11, 20, 21, 30, 31}

		for _, p := range connectPoints {
			assert.Equal(t, root, uf.Find(p), "Element %d should have the same root as element 0", p)
		}

		// Also verify no cycles were created (each element can reach a root)
		for i := 0; i < size; i++ {
			p := i
			visited := make(map[int]bool)

			// Follow parent pointers until we reach a root or detect a cycle
			for p != uf.parent[p] {
				if visited[p] {
					assert.Fail(t, "Cycle detected in UnionFind structure starting from element %d", i)
					break
				}
				visited[p] = true
				p = uf.parent[p]
			}
		}
	})
}
