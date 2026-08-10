package algo

import (
	"sync"
)

/*
https://dl.acm.org/doi/pdf/10.1145/103418.103458
*/

// UnionFind represents a disjoint-set data structure.
// It supports union and find operations, along with
// methods for determining whether two elements are in the same set.
type UnionFind struct {
	parent []int        // parent[i] = parent of i
	rank   []int        // rank[i] = rank of subtree rooted at i (never more than 31)
	count  int          // number of components
	mu     sync.RWMutex // mutex for thread safety
}

// NewUnionFind initializes a new UnionFind data structure with n elements.
// Initially, each element is in its own set.
func NewUnionFind(n int) *UnionFind {
	parent := make([]int, n)
	rank := make([]int, n)

	// Initialize: each element is its own parent
	for i := 0; i < n; i++ {
		parent[i] = i
		rank[i] = 0
	}

	return &UnionFind{
		parent: parent,
		rank:   rank,
		count:  n,
	}
}

// Find returns the canonical element (root) of the set containing element p.
// Uses path compression to keep the tree flat.
func (uf *UnionFind) Find(p int) int {
	// Validate p is a valid index
	if p < 0 || p >= len(uf.parent) {
		panic("index out of bounds")
	}

	uf.mu.RLock()
	// First check if p is already a root to avoid unnecessary locking
	if uf.parent[p] == p {
		uf.mu.RUnlock()
		return p
	}
	uf.mu.RUnlock()

	// If not a root, we need to traverse and compress the path
	uf.mu.Lock()
	defer uf.mu.Unlock()

	// Recheck condition in case another thread changed it while we were waiting for the lock
	if p < 0 || p >= len(uf.parent) {
		panic("index out of bounds")
	}

	// Find the root with path compression
	root := p
	for root != uf.parent[root] {
		root = uf.parent[root]
	}

	// Compress path leading back to root
	for p != root {
		newp := uf.parent[p]
		uf.parent[p] = root
		p = newp
	}

	return root
}

// Union merges the sets containing elements p and q.
// Uses union by rank to keep the tree balanced.
func (uf *UnionFind) Union(p, q int) {
	uf.mu.Lock()
	defer uf.mu.Unlock()

	rootP := uf.findWithoutLocking(p)
	rootQ := uf.findWithoutLocking(q)

	// Already in the same set
	if rootP == rootQ {
		return
	}

	// Make root of smaller rank point to root of larger rank
	if uf.rank[rootP] < uf.rank[rootQ] {
		uf.parent[rootP] = rootQ
	} else if uf.rank[rootP] > uf.rank[rootQ] {
		uf.parent[rootQ] = rootP
	} else {
		uf.parent[rootQ] = rootP
		uf.rank[rootP]++
	}

	// Decrease the number of components
	uf.count--
}

// findWithoutLocking is an internal helper that performs Find without locking.
// It assumes the caller has already acquired an appropriate lock.
func (uf *UnionFind) findWithoutLocking(p int) int {
	// Validate p is a valid index
	if p < 0 || p >= len(uf.parent) {
		panic("index out of bounds")
	}

	// Find the root with path compression
	root := p
	for root != uf.parent[root] {
		root = uf.parent[root]
	}

	// Compress path leading back to root
	for p != root {
		newp := uf.parent[p]
		uf.parent[p] = root
		p = newp
	}

	return root
}

// Connected returns true if the two elements are in the same set.
func (uf *UnionFind) Connected(p, q int) bool {
	uf.mu.RLock()
	defer uf.mu.RUnlock()

	return uf.findWithoutLocking(p) == uf.findWithoutLocking(q)
}

// Count returns the number of disjoint sets.
func (uf *UnionFind) Count() int {
	uf.mu.RLock()
	defer uf.mu.RUnlock()

	return uf.count
}
