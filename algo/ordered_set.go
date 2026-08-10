package algo

import (
	"cmp"
	"fmt"
	"strings"
)

// OrderedSet represents a set data structure that maintains elements in sorted order
// using a Red-Black Tree as the underlying storage mechanism
type OrderedSet[T cmp.Ordered] struct {
	tree *RedBlackTree[T, struct{}] // Use empty struct as value since we only care about keys
	size int
}

// NewOrderedSet creates a new empty OrderedSet
func NewOrderedSet[T cmp.Ordered]() *OrderedSet[T] {
	return &OrderedSet[T]{
		tree: NewRedBlackTree[T, struct{}](),
		size: 0,
	}
}

// Add inserts an element into the set
// Returns true if the element was added (not already present), false if already exists
func (s *OrderedSet[T]) Add(element T) bool {
	if s.Contains(element) {
		return false
	}
	s.tree.Insert(element, struct{}{})
	s.size++
	return true
}

// Remove deletes an element from the set
// Returns true if the element was removed, false if not found
func (s *OrderedSet[T]) Remove(element T) bool {
	if s.tree.Delete(element) {
		s.size--
		return true
	}
	return false
}

// Contains checks if an element exists in the set
func (s *OrderedSet[T]) Contains(element T) bool {
	return s.tree.Search(element) != nil
}

// Size returns the number of elements in the set
func (s *OrderedSet[T]) Size() int {
	return s.size
}

// IsEmpty returns true if the set contains no elements
func (s *OrderedSet[T]) IsEmpty() bool {
	return s.size == 0
}

// Clear removes all elements from the set
func (s *OrderedSet[T]) Clear() {
	s.tree = NewRedBlackTree[T, struct{}]()
	s.size = 0
}

// ToSlice returns all elements as a sorted slice
func (s *OrderedSet[T]) ToSlice() []T {
	return s.tree.InOrderTraversal()
}

// Min returns the smallest element in the set
// Returns zero value and false if set is empty
func (s *OrderedSet[T]) Min() (T, bool) {
	var zero T
	if s.IsEmpty() {
		return zero, false
	}

	// Find minimum by traversing left from root
	current := s.tree.Root
	for current.Left != s.tree.Nil {
		current = current.Left
	}
	return current.Key, true
}

// Max returns the largest element in the set
// Returns zero value and false if set is empty
func (s *OrderedSet[T]) Max() (T, bool) {
	var zero T
	if s.IsEmpty() {
		return zero, false
	}

	// Find maximum by traversing right from root
	current := s.tree.Root
	for current.Right != s.tree.Nil {
		current = current.Right
	}
	return current.Key, true
}

// Union returns a new set containing all elements from both sets
func (s *OrderedSet[T]) Union(other *OrderedSet[T]) *OrderedSet[T] {
	result := NewOrderedSet[T]()

	// Add all elements from this set
	for _, element := range s.ToSlice() {
		result.Add(element)
	}

	// Add all elements from other set
	for _, element := range other.ToSlice() {
		result.Add(element)
	}

	return result
}

// Intersection returns a new set containing elements present in both sets
func (s *OrderedSet[T]) Intersection(other *OrderedSet[T]) *OrderedSet[T] {
	result := NewOrderedSet[T]()

	// Use the smaller set for iteration efficiency
	smaller, larger := s, other
	if other.Size() < s.Size() {
		smaller, larger = other, s
	}

	for _, element := range smaller.ToSlice() {
		if larger.Contains(element) {
			result.Add(element)
		}
	}

	return result
}

// Difference returns a new set containing elements in this set but not in other
func (s *OrderedSet[T]) Difference(other *OrderedSet[T]) *OrderedSet[T] {
	result := NewOrderedSet[T]()

	for _, element := range s.ToSlice() {
		if !other.Contains(element) {
			result.Add(element)
		}
	}

	return result
}

// SymmetricDifference returns a new set containing elements in either set but not in both
func (s *OrderedSet[T]) SymmetricDifference(other *OrderedSet[T]) *OrderedSet[T] {
	result := NewOrderedSet[T]()

	// Add elements from this set that are not in other
	for _, element := range s.ToSlice() {
		if !other.Contains(element) {
			result.Add(element)
		}
	}

	// Add elements from other set that are not in this
	for _, element := range other.ToSlice() {
		if !s.Contains(element) {
			result.Add(element)
		}
	}

	return result
}

// IsSubset returns true if all elements in this set are also in other
func (s *OrderedSet[T]) IsSubset(other *OrderedSet[T]) bool {
	if s.Size() > other.Size() {
		return false
	}

	for _, element := range s.ToSlice() {
		if !other.Contains(element) {
			return false
		}
	}

	return true
}

// IsSuperset returns true if this set contains all elements in other
func (s *OrderedSet[T]) IsSuperset(other *OrderedSet[T]) bool {
	return other.IsSubset(s)
}

// Equals returns true if both sets contain exactly the same elements
func (s *OrderedSet[T]) Equals(other *OrderedSet[T]) bool {
	if s.Size() != other.Size() {
		return false
	}

	return s.IsSubset(other)
}

// Filter returns a new set containing only elements that satisfy the predicate
func (s *OrderedSet[T]) Filter(predicate func(T) bool) *OrderedSet[T] {
	result := NewOrderedSet[T]()

	for _, element := range s.ToSlice() {
		if predicate(element) {
			result.Add(element)
		}
	}

	return result
}

// ForEach applies a function to each element in the set (in sorted order)
func (s *OrderedSet[T]) ForEach(fn func(T)) {
	for _, element := range s.ToSlice() {
		fn(element)
	}
}

// String returns a string representation of the set
func (s *OrderedSet[T]) String() string {
	if s.IsEmpty() {
		return "{}"
	}

	elements := s.ToSlice()
	strElements := make([]string, len(elements))
	for i, element := range elements {
		strElements[i] = fmt.Sprintf("%v", element)
	}

	return "{" + strings.Join(strElements, ", ") + "}"
}

// Clone returns a shallow copy of the set
func (s *OrderedSet[T]) Clone() *OrderedSet[T] {
	clone := NewOrderedSet[T]()
	for _, element := range s.ToSlice() {
		clone.Add(element)
	}
	return clone
}

// AddAll adds multiple elements to the set
// Returns the number of elements actually added (excluding duplicates)
func (s *OrderedSet[T]) AddAll(elements ...T) int {
	added := 0
	for _, element := range elements {
		if s.Add(element) {
			added++
		}
	}
	return added
}

// RemoveAll removes multiple elements from the set
// Returns the number of elements actually removed
func (s *OrderedSet[T]) RemoveAll(elements ...T) int {
	removed := 0
	for _, element := range elements {
		if s.Remove(element) {
			removed++
		}
	}
	return removed
}

// ContainsAll returns true if the set contains all specified elements
func (s *OrderedSet[T]) ContainsAll(elements ...T) bool {
	for _, element := range elements {
		if !s.Contains(element) {
			return false
		}
	}
	return true
}

// ContainsAny returns true if the set contains at least one of the specified elements
func (s *OrderedSet[T]) ContainsAny(elements ...T) bool {
	for _, element := range elements {
		if s.Contains(element) {
			return true
		}
	}
	return false
}
