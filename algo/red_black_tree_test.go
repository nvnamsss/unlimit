package algo

import (
	"reflect"
	"sort"
	"sync"
	"testing"
)

// TestRedBlackTree_New tests the creation of a new Red-Black Tree
func TestRedBlackTree_New(t *testing.T) {
	tree := NewRedBlackTree[int, string]()

	if tree == nil {
		t.Error("NewRedBlackTree() should not return nil")
	}

	if tree.Root != tree.Nil {
		t.Error("New tree should have Root pointing to Nil sentinel")
	}

	if tree.Nil == nil {
		t.Error("Nil sentinel should not be nil")
	}

	if tree.Nil.Color != Black {
		t.Error("Nil sentinel should be Black")
	}
}

// TestRedBlackTree_Insert tests insertion operations using table-driven approach
func TestRedBlackTree_Insert(t *testing.T) {
	// Define test inputs
	type args struct {
		key   int
		value string
	}

	// Sample test data
	testKey := 10
	testValue := "ten"
	multipleKeys := []int{5, 15, 3, 7}
	multipleValues := []string{"five", "fifteen", "three", "seven"}

	// Define test cases
	tests := []struct {
		name           string
		args           args
		additionalOps  func(*RedBlackTree[int, string]) // Additional operations before verification
		wantRootKey    int
		wantRootValue  string
		wantRootColor  Color
		wantRootNotNil bool
	}{
		{
			name: "should insert single element successfully",
			args: args{
				key:   testKey,
				value: testValue,
			},
			wantRootKey:    testKey,
			wantRootValue:  testValue,
			wantRootColor:  Black,
			wantRootNotNil: true,
		},
		{
			name: "should maintain root as black after multiple insertions",
			args: args{
				key:   testKey,
				value: testValue,
			},
			additionalOps: func(tree *RedBlackTree[int, string]) {
				for i, key := range multipleKeys {
					tree.Insert(key, multipleValues[i])
				}
			},
			wantRootColor:  Black,
			wantRootNotNil: true,
		},
		{
			name: "should handle duplicate key insertion",
			args: args{
				key:   testKey,
				value: testValue,
			},
			additionalOps: func(tree *RedBlackTree[int, string]) {
				// Insert same key again with different value
				tree.Insert(testKey, "updated_ten")
			},
			wantRootKey:    testKey,
			wantRootColor:  Black,
			wantRootNotNil: true,
		},
	}

	// Execute tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup - create new tree for each test
			tree := NewRedBlackTree[int, string]()

			// Execute - perform the insertion
			tree.Insert(tt.args.key, tt.args.value)

			// Execute additional operations if specified
			if tt.additionalOps != nil {
				tt.additionalOps(tree)
			}

			// Verify results
			if tt.wantRootNotNil {
				if tree.Root == tree.Nil {
					t.Error("Root should not be Nil after insertion")
					return
				}

				if tree.Root.Color != tt.wantRootColor {
					t.Errorf("Expected root color %v, got %v", tt.wantRootColor, tree.Root.Color)
				}

				// Verify specific root key/value only if specified
				if tt.wantRootKey != 0 {
					if tree.Root.Key != tt.wantRootKey {
						t.Errorf("Expected root key %d, got %d", tt.wantRootKey, tree.Root.Key)
					}
				}

				if tt.wantRootValue != "" {
					if tree.Root.Value != tt.wantRootValue {
						t.Errorf("Expected root value '%s', got '%s'", tt.wantRootValue, tree.Root.Value)
					}
				}
			}

			// Verify tree structure integrity
			if tree.Root != tree.Nil {
				// Ensure root parent is Nil
				if tree.Root.Parent != tree.Nil {
					t.Error("Root parent should be Nil")
				}

				// Ensure root is always black (Red-Black Tree property)
				if tree.Root.Color != Black {
					t.Error("Root should always be Black")
				}
			}
		})
	}
}

// TestRedBlackTree_Search tests search functionality
func TestRedBlackTree_Search(t *testing.T) {
	tree := NewRedBlackTree[int, string]()

	// Search in empty tree
	result := tree.Search(10)
	if result != nil {
		t.Error("Search in empty tree should return nil")
	}

	// Insert elements and search
	tree.Insert(10, "ten")
	tree.Insert(5, "five")
	tree.Insert(15, "fifteen")

	// Search existing element
	result = tree.Search(5)
	if result == nil {
		t.Error("Search for existing element should not return nil")
	}
	if result.Key != 5 || result.Value != "five" {
		t.Errorf("Expected key=5, value='five', got key=%d, value=%s", result.Key, result.Value)
	}

	// Search non-existing element
	result = tree.Search(100)
	if result != nil {
		t.Error("Search for non-existing element should return nil")
	}
}

// TestRedBlackTree_Delete tests deletion operations
func TestRedBlackTree_Delete(t *testing.T) {
	tree := NewRedBlackTree[int, string]()

	// Delete from empty tree
	deleted := tree.Delete(10)
	if deleted {
		t.Error("Delete from empty tree should return false")
	}

	// Insert elements
	keys := []int{10, 5, 15, 3, 7, 12, 18}
	for _, key := range keys {
		tree.Insert(key, "value")
	}

	// Delete existing element
	deleted = tree.Delete(5)
	if !deleted {
		t.Error("Delete existing element should return true")
	}

	// Verify element is deleted
	result := tree.Search(5)
	if result != nil {
		t.Error("Deleted element should not be found")
	}

	// Delete non-existing element
	deleted = tree.Delete(100)
	if deleted {
		t.Error("Delete non-existing element should return false")
	}

	// Verify root is still black after deletions
	if tree.Root != tree.Nil && tree.Root.Color != Black {
		t.Error("Root should always be Black after deletions")
	}
}

// TestRedBlackTree_InOrderTraversal tests in-order traversal
func TestRedBlackTree_InOrderTraversal(t *testing.T) {
	tree := NewRedBlackTree[int, string]()

	// Traversal of empty tree
	result := tree.InOrderTraversal()
	if len(result) != 0 {
		t.Error("In-order traversal of empty tree should return empty slice")
	}

	// Insert elements
	keys := []int{10, 5, 15, 3, 7, 12, 18}
	for _, key := range keys {
		tree.Insert(key, "value")
	}

	// Get traversal result
	result = tree.InOrderTraversal()

	// Verify sorted order
	expected := make([]int, len(keys))
	copy(expected, keys)
	sort.Ints(expected)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// TestRedBlackTree_EmptyOperations tests operations on empty tree
func TestRedBlackTree_EmptyOperations(t *testing.T) {
	tree := NewRedBlackTree[int, string]()

	// Search in empty tree
	if tree.Search(1) != nil {
		t.Error("Search in empty tree should return nil")
	}

	// Delete from empty tree
	if tree.Delete(1) {
		t.Error("Delete from empty tree should return false")
	}

	// Traversal of empty tree
	result := tree.InOrderTraversal()
	if len(result) != 0 {
		t.Error("Traversal of empty tree should return empty slice")
	}

	// Root should be Nil
	if tree.Root != tree.Nil {
		t.Error("Empty tree root should be Nil")
	}
}

// TestRedBlackTree_DifferentTypes tests with different key/value types
func TestRedBlackTree_DifferentTypes(t *testing.T) {
	// Test with string keys and int values
	stringTree := NewRedBlackTree[string, int]()
	stringTree.Insert("apple", 1)
	stringTree.Insert("banana", 2)
	stringTree.Insert("cherry", 3)

	result := stringTree.Search("banana")
	if result == nil || result.Value != 2 {
		t.Error("String key tree should work correctly")
	}

	traversal := stringTree.InOrderTraversal()
	expected := []string{"apple", "banana", "cherry"}
	if !reflect.DeepEqual(traversal, expected) {
		t.Errorf("Expected %v, got %v", expected, traversal)
	}

	// Test with float keys
	floatTree := NewRedBlackTree[float64, string]()
	floatTree.Insert(3.14, "pi")
	floatTree.Insert(2.71, "e")
	floatTree.Insert(1.41, "sqrt2")

	floatResult := floatTree.Search(2.71)
	if floatResult == nil || floatResult.Value != "e" {
		t.Error("Float key tree should work correctly")
	}
}

// TestRedBlackTree_ConcurrentOperations tests thread-safety
func TestRedBlackTree_ConcurrentOperations(t *testing.T) {
	tree := NewRedBlackTree[int, string]()

	// Note: Red-Black Tree is not thread-safe by design
	// This test verifies the data structure integrity after concurrent operations
	// In a real scenario, external synchronization would be needed

	const goroutines = 10
	const opsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Perform concurrent insertions
	for i := 0; i < goroutines; i++ {
		go func(offset int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				key := offset*opsPerGoroutine + j
				tree.Insert(key, "value")
			}
		}(i)
	}

	wg.Wait()

	// Verify some insertions succeeded (exact count may vary due to race conditions)
	traversal := tree.InOrderTraversal()
	if len(traversal) == 0 {
		t.Error("No elements found after concurrent insertions")
	}

	// Verify ordering is maintained
	for i := 1; i < len(traversal); i++ {
		if traversal[i-1] >= traversal[i] {
			t.Error("Tree ordering violated after concurrent operations")
		}
	}
}

// TestRedBlackTree_LargeDataset tests with a large number of elements
func TestRedBlackTree_LargeDataset(t *testing.T) {
	tree := NewRedBlackTree[int, int]()

	const size = 1000

	// Insert elements
	for i := 0; i < size; i++ {
		tree.Insert(i, i*2)
	}

	// Verify all elements can be found
	for i := 0; i < size; i++ {
		result := tree.Search(i)
		if result == nil {
			t.Errorf("Element %d not found", i)
		}
		if result.Value != i*2 {
			t.Errorf("Expected value %d, got %d", i*2, result.Value)
		}
	}

	// Verify traversal produces sorted order
	traversal := tree.InOrderTraversal()
	if len(traversal) != size {
		t.Errorf("Expected %d elements, got %d", size, len(traversal))
	}

	for i := 0; i < size; i++ {
		if traversal[i] != i {
			t.Errorf("Expected element %d at position %d, got %d", i, i, traversal[i])
		}
	}

	// Delete half the elements
	for i := 0; i < size; i += 2 {
		deleted := tree.Delete(i)
		if !deleted {
			t.Errorf("Failed to delete element %d", i)
		}
	}

	// Verify deleted elements are gone and remaining elements are still there
	for i := 0; i < size; i++ {
		result := tree.Search(i)
		if i%2 == 0 {
			// Should be deleted
			if result != nil {
				t.Errorf("Deleted element %d still found", i)
			}
		} else {
			// Should still exist
			if result == nil {
				t.Errorf("Existing element %d not found after deletion", i)
			}
		}
	}
}

// TestRedBlackTree_RootAlwaysBlack verifies the root is always black
func TestRedBlackTree_RootAlwaysBlack(t *testing.T) {
	tree := NewRedBlackTree[int, string]()

	// Insert various elements and check root color
	keys := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	for _, key := range keys {
		tree.Insert(key, "value")
		if tree.Root.Color != Black {
			t.Errorf("Root should always be Black after inserting %d", key)
		}
	}

	// Delete elements and check root color
	for _, key := range keys {
		tree.Delete(key)
		if tree.Root != tree.Nil && tree.Root.Color != Black {
			t.Errorf("Root should always be Black after deleting %d", key)
		}
	}
}

// TestRedBlackTree_PrintTree tests the print functionality (basic smoke test)
func TestRedBlackTree_PrintTree(t *testing.T) {
	tree := NewRedBlackTree[int, string]()

	// Should not panic on empty tree
	tree.PrintTree()

	// Should not panic with elements
	tree.Insert(10, "ten")
	tree.Insert(5, "five")
	tree.Insert(15, "fifteen")
	tree.PrintTree()
}
