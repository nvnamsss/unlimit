package algo

import (
	"reflect"
	"sync"
	"testing"
)

// TestOrderedSet_New tests the creation of a new OrderedSet
func TestOrderedSet_New(t *testing.T) {
	set := NewOrderedSet[int]()

	if set == nil {
		t.Error("NewOrderedSet() should not return nil")
	}

	if set.Size() != 0 {
		t.Errorf("New set should have size 0, got %d", set.Size())
	}

	if !set.IsEmpty() {
		t.Error("New set should be empty")
	}

	if set.tree == nil {
		t.Error("Internal tree should be initialized")
	}
}

// TestOrderedSet_Add tests basic add operations using table-driven approach
func TestOrderedSet_Add(t *testing.T) {
	// Define test inputs
	type args struct {
		element int
	}

	// Sample test data
	testElement := 10
	duplicateElement := 10
	additionalElements := []int{5, 15, 3, 7}

	// Define test cases
	tests := []struct {
		name         string
		args         args
		setup        func(*OrderedSet[int]) // Setup initial state
		wantAdded    bool
		wantSize     int
		wantContains bool
	}{
		{
			name: "should add first element successfully",
			args: args{
				element: testElement,
			},
			setup:        func(set *OrderedSet[int]) {}, // Empty set
			wantAdded:    true,
			wantSize:     1,
			wantContains: true,
		},
		{
			name: "should reject duplicate element",
			args: args{
				element: duplicateElement,
			},
			setup: func(set *OrderedSet[int]) {
				set.Add(duplicateElement) // Pre-add the element
			},
			wantAdded:    false,
			wantSize:     1,
			wantContains: true,
		},
		{
			name: "should add element to non-empty set",
			args: args{
				element: 20,
			},
			setup: func(set *OrderedSet[int]) {
				for _, elem := range additionalElements {
					set.Add(elem)
				}
			},
			wantAdded:    true,
			wantSize:     len(additionalElements) + 1,
			wantContains: true,
		},
		{
			name: "should add negative element",
			args: args{
				element: -5,
			},
			setup:        func(set *OrderedSet[int]) {},
			wantAdded:    true,
			wantSize:     1,
			wantContains: true,
		},
	}

	// Execute tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup - create new set for each test
			set := NewOrderedSet[int]()

			// Apply setup if provided
			if tt.setup != nil {
				tt.setup(set)
			}

			// Execute - perform the add operation
			gotAdded := set.Add(tt.args.element)

			// Verify results
			if gotAdded != tt.wantAdded {
				t.Errorf("Add() returned %v, want %v", gotAdded, tt.wantAdded)
			}

			if set.Size() != tt.wantSize {
				t.Errorf("Expected size %d, got %d", tt.wantSize, set.Size())
			}

			if set.Contains(tt.args.element) != tt.wantContains {
				t.Errorf("Contains(%d) = %v, want %v", tt.args.element, set.Contains(tt.args.element), tt.wantContains)
			}

			// Verify set is not empty after successful add
			if tt.wantAdded && set.IsEmpty() {
				t.Error("Set should not be empty after successful add")
			}

			// Verify set maintains ordering
			slice := set.ToSlice()
			for i := 1; i < len(slice); i++ {
				if slice[i-1] >= slice[i] {
					t.Error("Set should maintain sorted order")
				}
			}
		})
	}
}

// TestOrderedSet_Remove tests remove operations using table-driven approach
func TestOrderedSet_Remove(t *testing.T) {
	// Define test inputs
	type args struct {
		element int
	}

	// Sample test data
	existingElements := []int{10, 5, 15, 3, 7}
	targetElement := 5
	nonExistentElement := 100

	// Define test cases
	tests := []struct {
		name         string
		args         args
		setup        func(*OrderedSet[int]) // Setup initial state
		wantRemoved  bool
		wantSize     int
		wantContains bool
	}{
		{
			name: "should return false when removing from empty set",
			args: args{
				element: targetElement,
			},
			setup:        func(set *OrderedSet[int]) {}, // Empty set
			wantRemoved:  false,
			wantSize:     0,
			wantContains: false,
		},
		{
			name: "should remove existing element successfully",
			args: args{
				element: targetElement,
			},
			setup: func(set *OrderedSet[int]) {
				for _, elem := range existingElements {
					set.Add(elem)
				}
			},
			wantRemoved:  true,
			wantSize:     len(existingElements) - 1,
			wantContains: false,
		},
		{
			name: "should return false when removing non-existing element",
			args: args{
				element: nonExistentElement,
			},
			setup: func(set *OrderedSet[int]) {
				for _, elem := range existingElements {
					set.Add(elem)
				}
			},
			wantRemoved:  false,
			wantSize:     len(existingElements),
			wantContains: false,
		},
		{
			name: "should remove last element from single-element set",
			args: args{
				element: targetElement,
			},
			setup: func(set *OrderedSet[int]) {
				set.Add(targetElement)
			},
			wantRemoved:  true,
			wantSize:     0,
			wantContains: false,
		},
		{
			name: "should remove element and maintain ordering",
			args: args{
				element: 15, // Remove max element
			},
			setup: func(set *OrderedSet[int]) {
				set.AddAll(10, 5, 15, 20) // Add in random order
			},
			wantRemoved:  true,
			wantSize:     3,
			wantContains: false,
		},
	}

	// Execute tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup - create new set for each test
			set := NewOrderedSet[int]()

			// Apply setup if provided
			if tt.setup != nil {
				tt.setup(set)
			}

			// Store initial slice for ordering verification
			initialSlice := set.ToSlice()

			// Execute - perform the remove operation
			gotRemoved := set.Remove(tt.args.element)

			// Verify results
			if gotRemoved != tt.wantRemoved {
				t.Errorf("Remove() returned %v, want %v", gotRemoved, tt.wantRemoved)
			}

			if set.Size() != tt.wantSize {
				t.Errorf("Expected size %d after removal, got %d", tt.wantSize, set.Size())
			}

			if set.Contains(tt.args.element) != tt.wantContains {
				t.Errorf("Contains(%d) = %v after removal, want %v", tt.args.element, set.Contains(tt.args.element), tt.wantContains)
			}

			// Verify set emptiness state
			if tt.wantSize == 0 && !set.IsEmpty() {
				t.Error("Set should be empty when size is 0")
			}

			if tt.wantSize > 0 && set.IsEmpty() {
				t.Error("Set should not be empty when size > 0")
			}

			// Verify remaining elements maintain ordering
			remainingSlice := set.ToSlice()
			for i := 1; i < len(remainingSlice); i++ {
				if remainingSlice[i-1] >= remainingSlice[i] {
					t.Error("Set should maintain sorted order after removal")
				}
			}

			// Verify that only the target element was removed
			if tt.wantRemoved {
				for _, elem := range initialSlice {
					if elem != tt.args.element && !set.Contains(elem) {
						t.Errorf("Element %d should still be in set after removing %d", elem, tt.args.element)
					}
				}
			}
		})
	}
}

// TestOrderedSet_Contains tests membership checking
func TestOrderedSet_Contains(t *testing.T) {
	set := NewOrderedSet[int]()

	// Check empty set
	if set.Contains(10) {
		t.Error("Empty set should not contain any elements")
	}

	// Add elements and check
	elements := []int{10, 5, 15}
	for _, elem := range elements {
		set.Add(elem)
	}

	for _, elem := range elements {
		if !set.Contains(elem) {
			t.Errorf("Set should contain element %d", elem)
		}
	}

	// Check non-existing element
	if set.Contains(100) {
		t.Error("Set should not contain non-existing element")
	}
}

// TestOrderedSet_ToSlice tests ordered traversal
func TestOrderedSet_ToSlice(t *testing.T) {
	set := NewOrderedSet[int]()

	// Empty set
	slice := set.ToSlice()
	if len(slice) != 0 {
		t.Error("Empty set should return empty slice")
	}

	// Add elements in random order
	elements := []int{15, 3, 10, 7, 1}
	for _, elem := range elements {
		set.Add(elem)
	}

	slice = set.ToSlice()
	expected := []int{1, 3, 7, 10, 15} // Should be sorted

	if !reflect.DeepEqual(slice, expected) {
		t.Errorf("Expected %v, got %v", expected, slice)
	}
}

// TestOrderedSet_MinMax tests min/max operations
func TestOrderedSet_MinMax(t *testing.T) {
	set := NewOrderedSet[int]()

	// Empty set
	_, ok := set.Min()
	if ok {
		t.Error("Min() on empty set should return false")
	}

	_, ok = set.Max()
	if ok {
		t.Error("Max() on empty set should return false")
	}

	// Add elements
	elements := []int{15, 3, 10, 7, 1}
	for _, elem := range elements {
		set.Add(elem)
	}

	min, ok := set.Min()
	if !ok {
		t.Error("Min() should return true for non-empty set")
	}
	if min != 1 {
		t.Errorf("Expected min 1, got %d", min)
	}

	max, ok := set.Max()
	if !ok {
		t.Error("Max() should return true for non-empty set")
	}
	if max != 15 {
		t.Errorf("Expected max 15, got %d", max)
	}
}

// TestOrderedSet_Clear tests clearing the set
func TestOrderedSet_Clear(t *testing.T) {
	set := NewOrderedSet[int]()

	// Add elements
	elements := []int{10, 5, 15}
	for _, elem := range elements {
		set.Add(elem)
	}

	// Clear and verify
	set.Clear()

	if set.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", set.Size())
	}

	if !set.IsEmpty() {
		t.Error("Set should be empty after clear")
	}

	for _, elem := range elements {
		if set.Contains(elem) {
			t.Errorf("Set should not contain element %d after clear", elem)
		}
	}
}

// TestOrderedSet_Union tests union operations
func TestOrderedSet_Union(t *testing.T) {
	set1 := NewOrderedSet[int]()
	set2 := NewOrderedSet[int]()

	// Union of empty sets
	union := set1.Union(set2)
	if union.Size() != 0 {
		t.Error("Union of empty sets should be empty")
	}

	// Add elements to sets
	set1.AddAll(1, 3, 5)
	set2.AddAll(2, 4, 5) // 5 is common

	union = set1.Union(set2)
	expected := []int{1, 2, 3, 4, 5}

	if union.Size() != len(expected) {
		t.Errorf("Expected union size %d, got %d", len(expected), union.Size())
	}

	unionSlice := union.ToSlice()
	if !reflect.DeepEqual(unionSlice, expected) {
		t.Errorf("Expected %v, got %v", expected, unionSlice)
	}
}

// TestOrderedSet_Intersection tests intersection operations
func TestOrderedSet_Intersection(t *testing.T) {
	set1 := NewOrderedSet[int]()
	set2 := NewOrderedSet[int]()

	// Intersection of empty sets
	intersection := set1.Intersection(set2)
	if intersection.Size() != 0 {
		t.Error("Intersection of empty sets should be empty")
	}

	// Add elements with some overlap
	set1.AddAll(1, 3, 5, 7)
	set2.AddAll(3, 5, 9, 11)

	intersection = set1.Intersection(set2)
	expected := []int{3, 5}

	if intersection.Size() != len(expected) {
		t.Errorf("Expected intersection size %d, got %d", len(expected), intersection.Size())
	}

	intersectionSlice := intersection.ToSlice()
	if !reflect.DeepEqual(intersectionSlice, expected) {
		t.Errorf("Expected %v, got %v", expected, intersectionSlice)
	}
}

// TestOrderedSet_Difference tests difference operations
func TestOrderedSet_Difference(t *testing.T) {
	set1 := NewOrderedSet[int]()
	set2 := NewOrderedSet[int]()

	// Add elements
	set1.AddAll(1, 3, 5, 7)
	set2.AddAll(3, 5, 9)

	difference := set1.Difference(set2)
	expected := []int{1, 7} // Elements in set1 but not in set2

	if difference.Size() != len(expected) {
		t.Errorf("Expected difference size %d, got %d", len(expected), difference.Size())
	}

	differenceSlice := difference.ToSlice()
	if !reflect.DeepEqual(differenceSlice, expected) {
		t.Errorf("Expected %v, got %v", expected, differenceSlice)
	}
}

// TestOrderedSet_SubsetSuperset tests subset/superset operations
func TestOrderedSet_SubsetSuperset(t *testing.T) {
	set1 := NewOrderedSet[int]()
	set2 := NewOrderedSet[int]()

	// Empty set is subset of any set
	if !set1.IsSubset(set2) {
		t.Error("Empty set should be subset of empty set")
	}

	set2.AddAll(1, 2, 3)
	if !set1.IsSubset(set2) {
		t.Error("Empty set should be subset of any set")
	}

	set1.AddAll(1, 3)
	if !set1.IsSubset(set2) {
		t.Error("set1 should be subset of set2")
	}

	if !set2.IsSuperset(set1) {
		t.Error("set2 should be superset of set1")
	}

	set1.Add(4) // Add element not in set2
	if set1.IsSubset(set2) {
		t.Error("set1 should not be subset of set2 after adding unique element")
	}
}

// TestOrderedSet_Equals tests equality operations
func TestOrderedSet_Equals(t *testing.T) {
	set1 := NewOrderedSet[int]()
	set2 := NewOrderedSet[int]()

	// Empty sets are equal
	if !set1.Equals(set2) {
		t.Error("Empty sets should be equal")
	}

	// Add same elements in different order
	set1.AddAll(3, 1, 5)
	set2.AddAll(1, 5, 3)

	if !set1.Equals(set2) {
		t.Error("Sets with same elements should be equal")
	}

	set2.Add(7)
	if set1.Equals(set2) {
		t.Error("Sets with different elements should not be equal")
	}
}

// TestOrderedSet_DifferentTypes tests with different data types
func TestOrderedSet_DifferentTypes(t *testing.T) {
	// Test with strings
	stringSet := NewOrderedSet[string]()
	stringSet.AddAll("banana", "apple", "cherry")

	stringSlice := stringSet.ToSlice()
	expectedStrings := []string{"apple", "banana", "cherry"}
	if !reflect.DeepEqual(stringSlice, expectedStrings) {
		t.Errorf("String set: expected %v, got %v", expectedStrings, stringSlice)
	}

	// Test with floats
	floatSet := NewOrderedSet[float64]()
	floatSet.AddAll(3.14, 2.71, 1.41)

	min, ok := floatSet.Min()
	if !ok || min != 1.41 {
		t.Errorf("Float set min: expected 1.41, got %v", min)
	}
}

// TestOrderedSet_Filter tests filtering operations
func TestOrderedSet_Filter(t *testing.T) {
	set := NewOrderedSet[int]()
	set.AddAll(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

	// Filter even numbers
	evens := set.Filter(func(x int) bool { return x%2 == 0 })
	expectedEvens := []int{2, 4, 6, 8, 10}

	evensSlice := evens.ToSlice()
	if !reflect.DeepEqual(evensSlice, expectedEvens) {
		t.Errorf("Expected %v, got %v", expectedEvens, evensSlice)
	}

	// Filter numbers greater than 5
	greaterThan5 := set.Filter(func(x int) bool { return x > 5 })
	expectedGreater := []int{6, 7, 8, 9, 10}

	greaterSlice := greaterThan5.ToSlice()
	if !reflect.DeepEqual(greaterSlice, expectedGreater) {
		t.Errorf("Expected %v, got %v", expectedGreater, greaterSlice)
	}
}

// TestOrderedSet_Clone tests cloning operations
func TestOrderedSet_Clone(t *testing.T) {
	original := NewOrderedSet[int]()
	original.AddAll(1, 3, 5, 7)

	clone := original.Clone()

	// Verify clone has same elements
	if !original.Equals(clone) {
		t.Error("Clone should equal original")
	}

	// Verify independence
	clone.Add(9)
	if original.Contains(9) {
		t.Error("Changes to clone should not affect original")
	}

	original.Remove(1)
	if !clone.Contains(1) {
		t.Error("Changes to original should not affect clone")
	}
}

// TestOrderedSet_EmptyOperations tests operations on empty set
func TestOrderedSet_EmptyOperations(t *testing.T) {
	set := NewOrderedSet[int]()

	// Remove from empty set
	if set.Remove(1) {
		t.Error("Remove from empty set should return false")
	}

	// Contains on empty set
	if set.Contains(1) {
		t.Error("Empty set should not contain any elements")
	}

	// Min/Max on empty set
	if _, ok := set.Min(); ok {
		t.Error("Min on empty set should return false")
	}

	if _, ok := set.Max(); ok {
		t.Error("Max on empty set should return false")
	}

	// ToSlice on empty set
	slice := set.ToSlice()
	if len(slice) != 0 {
		t.Error("Empty set should return empty slice")
	}

	// String representation
	str := set.String()
	if str != "{}" {
		t.Errorf("Expected '{}', got '%s'", str)
	}
}

// TestOrderedSet_ConcurrentOperations tests thread-safety awareness
func TestOrderedSet_ConcurrentOperations(t *testing.T) {
	set := NewOrderedSet[int]()

	// Note: OrderedSet is not thread-safe by design
	// This test verifies data structure integrity after concurrent operations
	// In a real scenario, external synchronization would be needed

	const goroutines = 10
	const opsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Perform concurrent additions
	for i := 0; i < goroutines; i++ {
		go func(offset int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				element := offset*opsPerGoroutine + j
				set.Add(element)
			}
		}(i)
	}

	wg.Wait()

	// Verify some additions succeeded
	if set.Size() == 0 {
		t.Error("No elements found after concurrent additions")
	}

	// Verify ordering is maintained
	slice := set.ToSlice()
	for i := 1; i < len(slice); i++ {
		if slice[i-1] >= slice[i] {
			t.Error("Set ordering violated after concurrent operations")
		}
	}
}

// TestOrderedSet_String tests string representation
func TestOrderedSet_String(t *testing.T) {
	set := NewOrderedSet[int]()

	// Empty set
	if set.String() != "{}" {
		t.Errorf("Expected '{}', got '%s'", set.String())
	}

	// Single element
	set.Add(5)
	if set.String() != "{5}" {
		t.Errorf("Expected '{5}', got '%s'", set.String())
	}

	// Multiple elements
	set.AddAll(1, 3)
	expected := "{1, 3, 5}" // Should be in sorted order
	if set.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, set.String())
	}
}

// TestOrderedSet_AddAllRemoveAll tests bulk operations
func TestOrderedSet_AddAllRemoveAll(t *testing.T) {
	set := NewOrderedSet[int]()

	// AddAll
	added := set.AddAll(1, 3, 5, 3, 7) // 3 is duplicate
	if added != 4 {                    // Should add 4 unique elements
		t.Errorf("Expected 4 additions, got %d", added)
	}

	if set.Size() != 4 {
		t.Errorf("Expected size 4, got %d", set.Size())
	}

	// RemoveAll
	removed := set.RemoveAll(1, 3, 9) // 9 doesn't exist
	if removed != 2 {                 // Should remove 2 existing elements
		t.Errorf("Expected 2 removals, got %d", removed)
	}

	if set.Size() != 2 {
		t.Errorf("Expected size 2 after removal, got %d", set.Size())
	}
}

// TestOrderedSet_ContainsAllAny tests bulk membership operations
func TestOrderedSet_ContainsAllAny(t *testing.T) {
	set := NewOrderedSet[int]()
	set.AddAll(1, 3, 5, 7)

	// ContainsAll
	if !set.ContainsAll(1, 3, 5) {
		t.Error("Should contain all specified elements")
	}

	if set.ContainsAll(1, 3, 9) {
		t.Error("Should not contain all elements when some are missing")
	}

	// ContainsAny
	if !set.ContainsAny(1, 9, 11) {
		t.Error("Should contain at least one specified element")
	}

	if set.ContainsAny(9, 11, 13) {
		t.Error("Should not contain any of the non-existing elements")
	}
}
