package collections

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

// Test Batch function
func TestBatch_BasicOperation(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	var processed [][]int

	err := Batch(slice, 2, func(batch []int) error {
		processed = append(processed, batch)
		return nil
	})

	if err != nil {
		t.Errorf("Batch() returned unexpected error: %v", err)
	}
}

func TestBatch_ErrorHandling(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	expectedErr := errors.New("test error")

	err := Batch(slice, 2, func(batch []int) error {
		if len(batch) > 0 && batch[0] == 3 {
			return expectedErr
		}
		return nil
	})

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestBatch_EmptySlice(t *testing.T) {
	var slice []int
	var processed [][]int

	err := Batch(slice, 2, func(batch []int) error {
		processed = append(processed, batch)
		return nil
	})

	if err != nil {
		t.Errorf("Batch() with empty slice returned error: %v", err)
	}

	if len(processed) != 1 || len(processed[0]) != 0 {
		t.Errorf("Expected one empty batch, got %v", processed)
	}
}

// Test BatchIterator creation
func TestBatchIterator_New(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	iterator := NewBatchIterator(slice, 2)

	if iterator == nil {
		t.Error("NewBatchIterator() returned nil")
	}
	if iterator.batchSize != 2 {
		t.Errorf("Expected batch size 2, got %d", iterator.batchSize)
	}
	if iterator.position != 0 {
		t.Errorf("Expected position 0, got %d", iterator.position)
	}
}

func TestBatchIterator_NewWithInvalidBatchSize(t *testing.T) {
	slice := []int{1, 2, 3}
	iterator := NewBatchIterator(slice, 0)

	if iterator.slice != nil {
		t.Error("Expected nil slice for invalid batch size")
	}
	if iterator.batchSize != 0 {
		t.Errorf("Expected batch size 0, got %d", iterator.batchSize)
	}
}

// Test Next method
func TestBatchIterator_Next(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	iterator := NewBatchIterator(slice, 2)

	// First batch
	batch, hasMore := iterator.Next()
	expected := []int{1, 2}
	if !reflect.DeepEqual(batch, expected) {
		t.Errorf("First batch: expected %v, got %v", expected, batch)
	}
	if !hasMore {
		t.Error("First batch should indicate more batches available")
	}

	// Second batch
	batch, hasMore = iterator.Next()
	expected = []int{3, 4}
	if !reflect.DeepEqual(batch, expected) {
		t.Errorf("Second batch: expected %v, got %v", expected, batch)
	}
	if !hasMore {
		t.Error("Second batch should indicate more batches available")
	}

	// Third batch
	batch, hasMore = iterator.Next()
	expected = []int{5}
	if !reflect.DeepEqual(batch, expected) {
		t.Errorf("Third batch: expected %v, got %v", expected, batch)
	}
	if hasMore {
		t.Error("Third batch should indicate no more batches")
	}

	// Fourth call should return empty
	batch, hasMore = iterator.Next()
	if len(batch) != 0 {
		t.Errorf("Expected empty batch, got %v", batch)
	}
	if hasMore {
		t.Error("Should indicate no more batches")
	}
}

func TestBatchIterator_NextEmptySlice(t *testing.T) {
	var slice []int
	iterator := NewBatchIterator(slice, 2)

	batch, hasMore := iterator.Next()
	if len(batch) != 0 {
		t.Errorf("Expected empty batch, got %v", batch)
	}
	if hasMore {
		t.Error("Empty slice should not have more batches")
	}
}

// Test Map method
func TestBatchIterator_Map(t *testing.T) {
	slice := []int{1, 2, 3, 4}
	iterator := NewBatchIterator(slice, 2)

	mapped := iterator.Map(func(batch []int) []int {
		result := make([]int, len(batch))
		for i, v := range batch {
			result[i] = v * 2
		}
		return result
	})

	// First batch
	batch, hasMore := mapped.Next()
	expected := []int{2, 4}
	if !reflect.DeepEqual(batch, expected) {
		t.Errorf("First mapped batch: expected %v, got %v", expected, batch)
	}
	if !hasMore {
		t.Error("Should have more batches")
	}

	// Second batch
	batch, hasMore = mapped.Next()
	expected = []int{6, 8}
	if !reflect.DeepEqual(batch, expected) {
		t.Errorf("Second mapped batch: expected %v, got %v", expected, batch)
	}
	if hasMore {
		t.Error("Should not have more batches")
	}
}

// Test Filter method
func TestBatchIterator_Filter(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6}
	iterator := NewBatchIterator(slice, 2)

	filtered := iterator.Filter(func(batch []int) bool {
		l := len(batch)
		three := batch[0] != 3
		return l == 2 && three // Only full batches not starting with 3
	})

	// Should get first batch [1, 2] and third batch [5, 6]
	batch, hasMore := filtered.Next()
	expected := []int{1, 2}
	if !reflect.DeepEqual(batch, expected) {
		t.Errorf("First filtered batch: expected %v, got %v", expected, batch)
	}
	if !hasMore {
		t.Error("Should have more batches")
	}

	batch, hasMore = filtered.Next()
	expected = []int{5, 6}
	if !reflect.DeepEqual(batch, expected) {
		t.Errorf("Second filtered batch: expected %v, got %v", expected, batch)
	}
	if hasMore {
		t.Error("Should not have more batches")
	}
}

// Test Take method
func TestBatchIterator_Take(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6}
	iterator := NewBatchIterator(slice, 2)

	taken := iterator.Take(2)

	// First batch
	batch, hasMore := taken.Next()
	expected := []int{1, 2}
	if !reflect.DeepEqual(batch, expected) {
		t.Errorf("First taken batch: expected %v, got %v", expected, batch)
	}
	if !hasMore {
		t.Error("Should have more batches")
	}

	// Second batch
	batch, hasMore = taken.Next()
	expected = []int{3, 4}
	if !reflect.DeepEqual(batch, expected) {
		t.Errorf("Second taken batch: expected %v, got %v", expected, batch)
	}
	if hasMore {
		t.Error("Should not have more batches after take limit")
	}

	// Third call should return empty
	batch, hasMore = taken.Next()
	if len(batch) != 0 {
		t.Errorf("Expected empty batch after take limit, got %v", batch)
	}
	if hasMore {
		t.Error("Should not have more batches after take limit")
	}
}

// Test Reduce method
func TestBatchIterator_Reduce(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	iterator := NewBatchIterator(slice, 2)

	result := iterator.Reduce(func(acc, batch []int) []int {
		return append(acc, batch...)
	})

	expected := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Reduce result: expected %v, got %v", expected, result)
	}
}

func TestBatchIterator_ReduceEmptySlice(t *testing.T) {
	var slice []int
	iterator := NewBatchIterator(slice, 2)

	result := iterator.Reduce(func(acc, batch []int) []int {
		return append(acc, batch...)
	})

	if len(result) != 0 {
		t.Errorf("Expected empty result for empty slice, got %v", result)
	}
}

// Test Collect method
func TestBatchIterator_Collect(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	iterator := NewBatchIterator(slice, 2)

	batches := iterator.Collect()

	expected := [][]int{{1, 2}, {3, 4}, {5}}
	if !reflect.DeepEqual(batches, expected) {
		t.Errorf("Collect result: expected %v, got %v", expected, batches)
	}
}

// Test Execute method
func TestBatchIterator_Execute(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	iterator := NewBatchIterator(slice, 2)

	var processed [][]int
	err := iterator.Execute(func(batch []int) error {
		processed = append(processed, batch)
		return nil
	})

	if err != nil {
		t.Errorf("Execute returned unexpected error: %v", err)
	}

	expected := [][]int{{1, 2}, {3, 4}, {5}}
	if !reflect.DeepEqual(processed, expected) {
		t.Errorf("Execute result: expected %v, got %v", expected, processed)
	}
}

func TestBatchIterator_ExecuteWithError(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	iterator := NewBatchIterator(slice, 2)

	expectedErr := errors.New("test error")
	var processed [][]int

	err := iterator.Execute(func(batch []int) error {
		processed = append(processed, batch)
		if len(processed) == 2 {
			return expectedErr
		}
		return nil
	})

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	if len(processed) != 2 {
		t.Errorf("Expected 2 processed batches before error, got %d", len(processed))
	}
}

// Test with different types
func TestBatchIterator_DifferentTypes(t *testing.T) {
	// Test with strings
	stringSlice := []string{"a", "b", "c", "d"}
	stringIterator := NewBatchIterator(stringSlice, 2)

	batch, hasMore := stringIterator.Next()
	expected := []string{"a", "b"}
	if !reflect.DeepEqual(batch, expected) {
		t.Errorf("String batch: expected %v, got %v", expected, batch)
	}
	if !hasMore {
		t.Error("Should have more batches")
	}

	// Test with custom struct
	type Person struct {
		Name string
		Age  int
	}

	people := []Person{
		{"Alice", 30},
		{"Bob", 25},
		{"Charlie", 35},
	}
	personIterator := NewBatchIterator(people, 2)

	personBatch, hasMore := personIterator.Next()
	expectedPeople := []Person{{"Alice", 30}, {"Bob", 25}}
	if !reflect.DeepEqual(personBatch, expectedPeople) {
		t.Errorf("Person batch: expected %v, got %v", expectedPeople, personBatch)
	}
	if !hasMore {
		t.Error("Should have more batches")
	}
}

// Test chaining operations
func TestBatchIterator_ChainedOperations(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6, 7, 8}
	iterator := NewBatchIterator(slice, 2)

	result := iterator.
		Map(func(batch []int) []int {
			// Double each element
			doubled := make([]int, len(batch))
			for i, v := range batch {
				doubled[i] = v * 2
			}
			return doubled
		}).
		Filter(func(batch []int) bool {
			// Only batches where first element is less than 10
			return len(batch) > 0 && batch[0] < 10
		}).
		Take(2).
		Collect()

	expected := [][]int{{2, 4}, {6, 8}}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Chained operations result: expected %v, got %v", expected, result)
	}
}

// Test ExecuteWithReturn function
func TestExecuteWithReturn_BasicOperation(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	iterator := NewBatchIterator(slice, 2)

	results, err := ExecuteWithReturn(iterator, func(batch []int) (string, error) {
		return fmt.Sprintf("Batch: %v", batch), nil
	})

	if err != nil {
		t.Errorf("ExecuteWithReturn() returned unexpected error: %v", err)
	}

	expected := []string{"Batch: [1 2]", "Batch: [3 4]", "Batch: [5]"}
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("Expected %v, got %v", expected, results)
	}
}

func TestExecuteWithReturn_ErrorHandling(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	iterator := NewBatchIterator(slice, 2)
	expectedErr := errors.New("test error")

	results, err := ExecuteWithReturn(iterator, func(batch []int) (string, error) {
		if len(batch) > 0 && batch[0] == 3 {
			return "", expectedErr
		}
		return fmt.Sprintf("Batch: %v", batch), nil
	})

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	if results != nil {
		t.Errorf("Expected nil results on error, got %v", results)
	}
}

func TestExecuteWithReturn_EmptySlice(t *testing.T) {
	var slice []int
	iterator := NewBatchIterator(slice, 2)

	results, err := ExecuteWithReturn(iterator, func(batch []int) (int, error) {
		return len(batch), nil
	})

	if err != nil {
		t.Errorf("ExecuteWithReturn() with empty slice returned error: %v", err)
	}

	expected := []int{0}
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("Expected %v, got %v", expected, results)
	}
}

func TestExecuteWithReturn_IntegerReturn(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6}
	iterator := NewBatchIterator(slice, 3)

	results, err := ExecuteWithReturn(iterator, func(batch []int) (int, error) {
		sum := 0
		for _, v := range batch {
			sum += v
		}
		return sum, nil
	})

	if err != nil {
		t.Errorf("ExecuteWithReturn() returned unexpected error: %v", err)
	}

	expected := []int{6, 15} // [1+2+3, 4+5+6]
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("Expected %v, got %v", expected, results)
	}
}

func TestExecuteWithReturn_StructReturn(t *testing.T) {
	type BatchInfo struct {
		Size  int
		First int
		Last  int
	}

	slice := []int{1, 2, 3, 4, 5}
	iterator := NewBatchIterator(slice, 2)

	results, err := ExecuteWithReturn(iterator, func(batch []int) (BatchInfo, error) {
		if len(batch) == 0 {
			return BatchInfo{}, nil
		}
		return BatchInfo{
			Size:  len(batch),
			First: batch[0],
			Last:  batch[len(batch)-1],
		}, nil
	})

	if err != nil {
		t.Errorf("ExecuteWithReturn() returned unexpected error: %v", err)
	}

	expected := []BatchInfo{
		{Size: 2, First: 1, Last: 2},
		{Size: 2, First: 3, Last: 4},
		{Size: 1, First: 5, Last: 5},
	}
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("Expected %v, got %v", expected, results)
	}
}

func TestExecuteWithReturn_SliceReturn(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6}
	iterator := NewBatchIterator(slice, 2)

	results, err := ExecuteWithReturn(iterator, func(batch []int) ([]int, error) {
		doubled := make([]int, len(batch))
		for i, v := range batch {
			doubled[i] = v * 2
		}
		return doubled, nil
	})

	if err != nil {
		t.Errorf("ExecuteWithReturn() returned unexpected error: %v", err)
	}

	expected := [][]int{{2, 4}, {6, 8}, {10, 12}}
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("Expected %v, got %v", expected, results)
	}
}

func TestExecuteWithReturn_BooleanReturn(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	iterator := NewBatchIterator(slice, 2)

	results, err := ExecuteWithReturn(iterator, func(batch []int) (bool, error) {
		// Return true if batch contains even numbers
		for _, v := range batch {
			if v%2 == 0 {
				return true, nil
			}
		}
		return false, nil
	})

	if err != nil {
		t.Errorf("ExecuteWithReturn() returned unexpected error: %v", err)
	}

	expected := []bool{true, true, false} // [1,2] has 2, [3,4] has 4, [5] has no even
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("Expected %v, got %v", expected, results)
	}
}

func TestExecuteWithReturn_WithChainedOperations(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6, 7, 8}
	iterator := NewBatchIterator(slice, 2)

	// Apply some transformations before ExecuteWithReturn
	transformed := iterator.
		Map(func(batch []int) []int {
			doubled := make([]int, len(batch))
			for i, v := range batch {
				doubled[i] = v * 2
			}
			return doubled
		}).
		Filter(func(batch []int) bool {
			return len(batch) > 0 && batch[0] < 10
		}).
		Take(2)

	results, err := ExecuteWithReturn(transformed.(*BatchIterator[int]), func(batch []int) (string, error) {
		return fmt.Sprintf("Transformed: %v", batch), nil
	})

	if err != nil {
		t.Errorf("ExecuteWithReturn() with chained operations returned error: %v", err)
	}

	expected := []string{"Transformed: [2 4]", "Transformed: [6 8]"}
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("Expected %v, got %v", expected, results)
	}
}

func TestExecuteWithReturn_DifferentInputTypes(t *testing.T) {
	// Test with string slice
	stringSlice := []string{"hello", "world", "test"}
	stringIterator := NewBatchIterator(stringSlice, 2)

	stringResults, err := ExecuteWithReturn(stringIterator, func(batch []string) (int, error) {
		totalLength := 0
		for _, s := range batch {
			totalLength += len(s)
		}
		return totalLength, nil
	})

	if err != nil {
		t.Errorf("ExecuteWithReturn() with strings returned error: %v", err)
	}

	expectedStringResults := []int{10, 4} // ["hello","world"] = 5+5=10, ["test"] = 4
	if !reflect.DeepEqual(stringResults, expectedStringResults) {
		t.Errorf("Expected %v, got %v", expectedStringResults, stringResults)
	}
}
