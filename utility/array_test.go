package utility

import (
	"reflect"
	"strconv"
	"testing"
)

func TestUniqueArray_Basic(t *testing.T) {
	// Setup
	arr := []int{1, 2, 2, 3, 3, 3, 4, 4, 4, 4}

	// Exercise
	result := UniqueArray(arr)

	// Verify
	// Convert to map for easy comparison regardless of order
	resultMap := make(map[int]bool)
	for _, v := range result {
		resultMap[v] = true
	}

	expected := map[int]bool{1: true, 2: true, 3: true, 4: true}
	if !reflect.DeepEqual(resultMap, expected) {
		t.Errorf("UniqueArray failed, expected unique elements %v, got %v", expected, resultMap)
	}

	if len(result) != 4 {
		t.Errorf("UniqueArray failed, expected 4 elements, got %d", len(result))
	}
}

func TestUniqueArray_Empty(t *testing.T) {
	// Setup
	var arr []int

	// Exercise
	result := UniqueArray(arr)

	// Verify
	if len(result) != 0 {
		t.Errorf("UniqueArray on empty array should return empty array, got %v", result)
	}
}

func TestUniqueArray_String(t *testing.T) {
	// Setup
	arr := []string{"apple", "banana", "apple", "cherry", "banana", "apple"}

	// Exercise
	result := UniqueArray(arr)

	// Verify
	resultMap := make(map[string]bool)
	for _, v := range result {
		resultMap[v] = true
	}

	expected := map[string]bool{"apple": true, "banana": true, "cherry": true}
	if !reflect.DeepEqual(resultMap, expected) {
		t.Errorf("UniqueArray failed for strings, expected %v, got %v", expected, resultMap)
	}
}

type testUser struct {
	ID   int
	Name string
}

func TestArray2Map_Basic(t *testing.T) {
	// Setup
	users := []testUser{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Charlie"},
	}

	// Exercise
	result := Array2Map(users, func(u testUser) int { return u.ID })

	// Verify
	if len(result) != 3 {
		t.Errorf("Array2Map failed, expected 3 elements, got %d", len(result))
	}

	if result[1].Name != "Alice" {
		t.Errorf("Array2Map failed, expected Name 'Alice' for key 1, got %s", result[1].Name)
	}

	if result[2].Name != "Bob" {
		t.Errorf("Array2Map failed, expected Name 'Bob' for key 2, got %s", result[2].Name)
	}

	if result[3].Name != "Charlie" {
		t.Errorf("Array2Map failed, expected Name 'Charlie' for key 3, got %s", result[3].Name)
	}
}

func TestArray2Map_DuplicateKeys(t *testing.T) {
	// Setup
	users := []testUser{
		{ID: 1, Name: "Alice"},
		{ID: 1, Name: "Bob"}, // Duplicate key
		{ID: 2, Name: "Charlie"},
	}

	// Exercise
	result := Array2Map(users, func(u testUser) int { return u.ID })

	// Verify
	if len(result) != 2 {
		t.Errorf("Array2Map with duplicate keys failed, expected 2 elements, got %d", len(result))
	}

	if result[1].Name != "Alice" {
		t.Errorf("Array2Map failed, should keep first entry for duplicate key, expected 'Alice', got %s", result[1].Name)
	}
}

func TestArray2Map_Empty(t *testing.T) {
	// Setup
	var users []testUser

	// Exercise
	result := Array2Map(users, func(u testUser) int { return u.ID })

	// Verify
	if len(result) != 0 {
		t.Errorf("Array2Map on empty array should return empty map, got %v", result)
	}
}

func TestUniqueKeys_Basic(t *testing.T) {
	// Setup
	users := []testUser{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 1, Name: "Charlie"}, // Duplicate ID
		{ID: 3, Name: "Dave"},
		{ID: 2, Name: "Eve"}, // Duplicate ID
	}

	// Exercise
	result := UniqueKeys(users, func(u testUser) int { return u.ID })

	// Verify
	resultMap := make(map[int]bool)
	for _, v := range result {
		resultMap[v] = true
	}

	expected := map[int]bool{1: true, 2: true, 3: true}
	if !reflect.DeepEqual(resultMap, expected) {
		t.Errorf("UniqueKeys failed, expected keys %v, got %v", expected, resultMap)
	}

	if len(result) != 3 {
		t.Errorf("UniqueKeys failed, expected 3 elements, got %d", len(result))
	}
}

func TestUniqueKeys_Empty(t *testing.T) {
	// Setup
	var users []testUser

	// Exercise
	result := UniqueKeys(users, func(u testUser) int { return u.ID })

	// Verify
	if len(result) != 0 {
		t.Errorf("UniqueKeys on empty array should return empty array, got %v", result)
	}
}

func TestIsContains_Basic(t *testing.T) {
	// Setup
	arr := []string{"apple", "banana", "cherry"}

	// Exercise and Verify
	if !IsContains(arr, "apple") {
		t.Errorf("IsContains failed, expected to find 'apple'")
	}

	if !IsContains(arr, "banana") {
		t.Errorf("IsContains failed, expected to find 'banana'")
	}

	if !IsContains(arr, "cherry") {
		t.Errorf("IsContains failed, expected to find 'cherry'")
	}

	if IsContains(arr, "dragonfruit") {
		t.Errorf("IsContains failed, did not expect to find 'dragonfruit'")
	}
}

func TestIsContains_Complex(t *testing.T) {
	// Setup
	arr := []testUser{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Charlie"},
	}

	// Exercise and Verify
	if !IsContains(arr, testUser{ID: 1, Name: "Alice"}) {
		t.Errorf("IsContains failed for complex type, expected to find user Alice")
	}

	if IsContains(arr, testUser{ID: 4, Name: "Dave"}) {
		t.Errorf("IsContains failed for complex type, did not expect to find user Dave")
	}
}

func TestIsContains_Empty(t *testing.T) {
	// Setup
	var arr []string

	// Exercise and Verify
	if IsContains(arr, "apple") {
		t.Errorf("IsContains on empty array should return false")
	}
}

func TestChunkArray_Basic(t *testing.T) {
	// Setup
	arr := []int{1, 2, 3, 4, 5}

	// Exercise
	result := ChunkArray(arr, 3)

	// Verify
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("ChunkArray failed, expected %v, got %v", expected, result)
	}
}

func TestChunkArray_LargerLimit(t *testing.T) {
	// Setup
	arr := []int{1, 2, 3}

	// Exercise
	result := ChunkArray(arr, 5)

	// Verify
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("ChunkArray with larger limit failed, expected %v, got %v", expected, result)
	}
}

func TestChunkArray_Empty(t *testing.T) {
	// Setup
	var arr []int

	// Exercise
	result := ChunkArray(arr, 3)

	// Verify
	if len(result) != 0 {
		t.Errorf("ChunkArray on empty array should return empty array, got %v", result)
	}
}

func TestReverse_Basic(t *testing.T) {
	// Setup
	arr := []int{1, 2, 3, 4, 5}
	expected := []int{5, 4, 3, 2, 1}

	// Exercise
	Reverse(arr)

	// Verify
	if !reflect.DeepEqual(arr, expected) {
		t.Errorf("Reverse failed, expected %v, got %v", expected, arr)
	}
}

func TestReverse_SingleElement(t *testing.T) {
	// Setup
	arr := []int{1}
	expected := []int{1}

	// Exercise
	Reverse(arr)

	// Verify
	if !reflect.DeepEqual(arr, expected) {
		t.Errorf("Reverse on single element failed, expected %v, got %v", expected, arr)
	}
}

func TestReverse_Empty(t *testing.T) {
	// Setup
	var arr []int

	// Exercise
	Reverse(arr)

	// Verify
	if len(arr) != 0 {
		t.Errorf("Reverse on empty array should still be empty, got %v", arr)
	}
}

func TestFilter_Basic(t *testing.T) {
	// Setup
	arr := []int{1, 2, 3, 4, 5}

	// Exercise
	result := Filter(arr, func(n int) bool { return n%2 == 0 })

	// Verify
	expected := []int{2, 4}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Filter failed, expected %v, got %v", expected, result)
	}
}

func TestFilter_AllMatch(t *testing.T) {
	// Setup
	arr := []int{2, 4, 6, 8}

	// Exercise
	result := Filter(arr, func(n int) bool { return n%2 == 0 })

	// Verify
	expected := []int{2, 4, 6, 8}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Filter with all matches failed, expected %v, got %v", expected, result)
	}
}

func TestFilter_NoMatch(t *testing.T) {
	// Setup
	arr := []int{1, 3, 5, 7}

	// Exercise
	result := Filter(arr, func(n int) bool { return n%2 == 0 })

	// Verify
	if len(result) != 0 {
		t.Errorf("Filter with no matches should return empty array, got %v", result)
	}
}

func TestFilter_Empty(t *testing.T) {
	// Setup
	var arr []int

	// Exercise
	result := Filter(arr, func(n int) bool { return n%2 == 0 })

	// Verify
	if len(result) != 0 {
		t.Errorf("Filter on empty array should return empty array, got %v", result)
	}
}

func TestRemoveDuplicates_Basic(t *testing.T) {
	// Setup
	arr := []int{1, 2, 2, 3, 3, 3, 4, 4, 4, 4}

	// Exercise
	result := RemoveDuplicates(arr)

	// Verify
	expected := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("RemoveDuplicates failed, expected %v, got %v", expected, result)
	}
}

func TestRemoveDuplicates_NoDuplicates(t *testing.T) {
	// Setup
	arr := []int{1, 2, 3, 4, 5}

	// Exercise
	result := RemoveDuplicates(arr)

	// Verify
	expected := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("RemoveDuplicates on array without duplicates failed, expected %v, got %v", expected, result)
	}
}

func TestRemoveDuplicates_Empty(t *testing.T) {
	// Setup
	var arr []int

	// Exercise
	result := RemoveDuplicates(arr)

	// Verify
	if len(result) != 0 {
		t.Errorf("RemoveDuplicates on empty array should return empty array, got %v", result)
	}
}

func TestFormatList_Basic(t *testing.T) {
	// Setup
	arr := []int{1, 2, 3}

	// Exercise
	result := FormatList(arr, func(v int) string {
		return "Number: " + strconv.Itoa(v)
	})

	// Verify
	expected := []string{"Number: 1", "Number: 2", "Number: 3"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("FormatList failed, expected %v, got %v", expected, result)
	}
}

func TestFormatList_TypeConversion(t *testing.T) {
	// Setup
	arr := []int{1, 2, 3}

	// Exercise
	result := FormatList(arr, func(v int) int {
		return v * 2
	})

	// Verify
	expected := []int{2, 4, 6}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("FormatList with type conversion failed, expected %v, got %v", expected, result)
	}
}

func TestFormatList_Empty(t *testing.T) {
	// Setup
	var arr []int

	// Exercise
	result := FormatList(arr, func(v int) string {
		return strconv.Itoa(v)
	})

	// Verify
	if len(result) != 0 {
		t.Errorf("FormatList on empty array should return empty array, got %v", result)
	}
}

func TestBatchArrayIterator(t *testing.T) {
	t.Run("Basic Functionality", func(t *testing.T) {
		// Setup
		arr := []int{1, 2, 3, 4, 5}
		batchSize := 2

		// Exercise
		iterator := BatchArrayIterator(arr, batchSize)

		// Verify
		var results [][]int
		for {
			batch, hasMore := iterator.Next()
			if len(batch) > 0 {
				results = append(results, batch)
			}
			if !hasMore {
				break
			}
		}

		expected := [][]int{{1, 2}, {3, 4}, {5}}
		if len(results) != 3 {
			t.Errorf("BatchArrayIterator failed, expected 3 batches, got %d", len(results))
		}

		for i := 0; i < len(expected); i++ {
			if !reflect.DeepEqual(results[i], expected[i]) {
				t.Errorf("BatchArrayIterator batch %d failed, expected %v, got %v", i, expected[i], results[i])
			}
		}
	})

	t.Run("Empty Array", func(t *testing.T) {
		// Setup
		var arr []int
		batchSize := 2

		// Exercise
		iterator := BatchArrayIterator(arr, batchSize)

		// Verify
		batch, hasMore := iterator.Next()
		if len(batch) != 0 {
			t.Errorf("BatchArrayIterator on empty array should return empty batch, got %v", batch)
		}
		if hasMore {
			t.Errorf("BatchArrayIterator on empty array should indicate no more batches")
		}
	})

	t.Run("Batch Size Equal to Array Length", func(t *testing.T) {
		// Setup
		arr := []int{1, 2, 3}
		batchSize := 3

		// Exercise
		iterator := BatchArrayIterator(arr, batchSize)

		// Verify
		batch, hasMore := iterator.Next()
		if !reflect.DeepEqual(batch, []int{1, 2, 3}) {
			t.Errorf("BatchArrayIterator failed, expected [1, 2, 3], got %v", batch)
		}
		if hasMore {
			t.Errorf("BatchArrayIterator should indicate no more batches after full batch")
		}

		// Ensure no more batches
		batch, hasMore = iterator.Next()
		if len(batch) != 0 {
			t.Errorf("BatchArrayIterator should return empty batch after exhaustion, got %v", batch)
		}
		if hasMore {
			t.Errorf("BatchArrayIterator should indicate no more batches after exhaustion")
		}
	})

	t.Run("Batch Size Greater Than Array Length", func(t *testing.T) {
		// Setup
		arr := []int{1, 2}
		batchSize := 5

		// Exercise
		iterator := BatchArrayIterator(arr, batchSize)

		// Verify
		batch, hasMore := iterator.Next()
		if !reflect.DeepEqual(batch, []int{1, 2}) {
			t.Errorf("BatchArrayIterator failed, expected [1, 2], got %v", batch)
		}
		if hasMore {
			t.Errorf("BatchArrayIterator should indicate no more batches after full batch")
		}
	})

	t.Run("Invalid Batch Size", func(t *testing.T) {
		// Setup
		arr := []int{1, 2, 3, 4, 5}
		batchSize := 0

		// Exercise
		iterator := BatchArrayIterator(arr, batchSize)

		// Verify
		batch, hasMore := iterator.Next()
		if len(batch) != 0 {
			t.Errorf("BatchArrayIterator with batch size 0 should return empty batch, got %v", batch)
		}
		if hasMore {
			t.Errorf("BatchArrayIterator with batch size 0 should indicate no more batches")
		}

		// Test negative batch size
		batchSize = -1
		iterator = BatchArrayIterator(arr, batchSize)
		batch, hasMore = iterator.Next()
		if len(batch) != 0 {
			t.Errorf("BatchArrayIterator with negative batch size should return empty batch, got %v", batch)
		}
		if hasMore {
			t.Errorf("BatchArrayIterator with negative batch size should indicate no more batches")
		}
	})

	t.Run("Collect Method", func(t *testing.T) {
		// Setup
		arr := []int{1, 2, 3, 4, 5}
		batchSize := 2

		// Exercise
		iterator := BatchArrayIterator(arr, batchSize)
		results := iterator.Collect()

		// Verify
		expected := [][]int{{1, 2}, {3, 4}, {5}}
		if len(results) != 3 {
			t.Errorf("BatchArrayIterator Collect failed, expected 3 batches, got %d", len(results))
		}

		for i := 0; i < len(expected); i++ {
			if !reflect.DeepEqual(results[i], expected[i]) {
				t.Errorf("BatchArrayIterator Collect batch %d failed, expected %v, got %v", i, expected[i], results[i])
			}
		}

		// After collecting, Next should return empty batch
		batch, hasMore := iterator.Next()
		if len(batch) != 0 || hasMore {
			t.Errorf("BatchArrayIterator Next after Collect should return empty batch and false")
		}
	})

	t.Run("String Type", func(t *testing.T) {
		// Setup
		arr := []string{"a", "b", "c", "d", "e"}
		batchSize := 2

		// Exercise
		iterator := BatchArrayIterator(arr, batchSize)

		// Verify
		var results [][]string
		for {
			batch, hasMore := iterator.Next()
			if len(batch) > 0 {
				results = append(results, batch)
			}
			if !hasMore {
				break
			}
		}

		expected := [][]string{{"a", "b"}, {"c", "d"}, {"e"}}
		if len(results) != 3 {
			t.Errorf("BatchArrayIterator failed for strings, expected 3 batches, got %d", len(results))
		}

		for i := 0; i < len(expected); i++ {
			if !reflect.DeepEqual(results[i], expected[i]) {
				t.Errorf("BatchArrayIterator batch %d failed for strings, expected %v, got %v", i, expected[i], results[i])
			}
		}
	})
}
