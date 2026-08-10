package collections

// Iterator is a generic interface for iterating over a collection of elements
type Iterator[T any] interface {
	Next() (T, bool)                 // returns the next item and a boolean indicating if there's more
	Map(func(T) T) Iterator[T]       // transforms each element using the provided function
	Reduce(func(T, T) T) T           // combines all elements using the provided function
	Filter(func(T) bool) Iterator[T] // filters elements based on the predicate function
	Take(n int) Iterator[T]          // limits the number of elements returned by the iterator
	Collect() []T                    // returns all remaining elements as a slice
}

// Batch processes a slice in batches with the provided function.
// It is a generic function that takes a slice of type T, a batch size,
// and a function that processes each batch and may return an error.
// Example usage:
//
//	arr := []int{1, 2, 3, 4, 5}
//	err := collections.Batch(arr, 2, func(batch []int) error {
//		fmt.Println(batch)
//		return nil
//	})
//	// Output:
//	// [1 2]
//	// [3 4]
//	// [5]
//
// This function is useful for efficiently processing large slices in smaller batches
// while handling potential errors from the batch processing function.
// Processing stops on the first error encountered and returns it.
// If no errors occur, nil is returned after all batches are processed.
func Batch[T any](slice []T, batchSize int, f func([]T) error) error {
	it := NewBatchIterator(slice, batchSize)
	return it.Execute(f)
}

// BatchIterator implements Iterator interface for processing slices in batches
type BatchIterator[T any] struct {
	slice     []T // slice is the underlying data to be processed in batches
	batchSize int // batchSize specifies the maximum number of elements per batch
	position  int // position tracks the current index in the slice for iteration

	mapper  func([]T) []T      // mapper is an optional function to transform each batch
	filter  func([]T) bool     // filter is an optional predicate to include/exclude batches
	taker   int                // taker limits the number of batches to process (0 means no limit)
	reducer func([]T, []T) []T // reducer is an optional function to combine batches
}

// NewBatchIterator creates a new BatchIterator with the given slice and batch size
func NewBatchIterator[T any](slice []T, batchSize int) *BatchIterator[T] {
	if batchSize <= 0 {
		return &BatchIterator[T]{
			slice:     nil,
			batchSize: 0,
			position:  0,
		}
	}

	return &BatchIterator[T]{
		slice:     slice,
		batchSize: batchSize,
		position:  0,
		taker:     -1,
	}
}

// Map transforms each batch in the iterator using the provided function.
// It is a generic method that takes a function which receives a batch (slice of T)
// and returns a transformed batch (slice of T). The method returns a new iterator
// that yields the transformed batches.
// Example usage:
//
//	iterator := NewBatchIterator([]int{1, 2, 3, 4}, 2)
//	mapped := iterator.Map(func(batch []int) []int {
//	    for i, v := range batch {
//	        batch[i] = v * 2
//	    }
//	    return batch
//	})
//	for {
//	    batch, hasMore := mapped.Next()
//	    fmt.Println(batch)
//	    if !hasMore {
//	        break
//	    }
//	}
//	// Output:
//	// [2 4]
//	// [6 8]
//
// This method is useful for applying transformations to each batch of elements
// in a collection, such as mapping, formatting, or modifying data in bulk.
func (it *BatchIterator[T]) Map(fn func([]T) []T) Iterator[[]T] {
	return &BatchIterator[T]{
		slice:     it.slice,
		batchSize: it.batchSize,
		position:  it.position,
		mapper:    fn,
		taker:     it.taker,
		filter:    it.filter,
		reducer:   it.reducer,
	}
}

// Filter keeps only batches that satisfy the predicate function.
// It is a generic method that takes a function which receives a batch (slice of T)
// and returns a boolean indicating whether the batch should be included.
// The method returns a new iterator that yields only the batches for which the predicate returns true.
// Example usage:
//
//	iterator := NewBatchIterator([]int{1, 2, 3, 4, 5, 6}, 2)
//	filtered := iterator.Filter(func(batch []int) bool {
//	    return len(batch) == 2 && batch[0]%2 == 1 // Only full batches starting with odd number
//	})
//	for {
//	    batch, hasMore := filtered.Next()
//	    fmt.Println(batch)
//	    if !hasMore {
//	        break
//	    }
//	}
//	// Output:
//	// [1 2]
//	// [5 6]
//
// This method is useful for extracting only those batches that meet specific criteria,
// enabling selective processing or analysis of data in bulk.
func (it *BatchIterator[T]) Filter(fn func([]T) bool) Iterator[[]T] {
	return &BatchIterator[T]{
		slice:     it.slice,
		batchSize: it.batchSize,
		position:  it.position,
		filter:    fn,
		mapper:    it.mapper,
		taker:     it.taker,
		reducer:   it.reducer,
	}
}

// Take limits the number of batches returned by the iterator.
// It is a generic method that takes an integer n and returns a new iterator
// that will yield at most n batches before stopping iteration.
// Example usage:
//
//	iterator := NewBatchIterator([]int{1, 2, 3, 4, 5, 6}, 2)
//	taken := iterator.Take(2)
//	for {
//	    batch, hasMore := taken.Next()
//	    fmt.Println(batch)
//	    if !hasMore {
//	        break
//	    }
//	}
//	// Output:
//	// [1 2]
//	// [3 4]
//
// This method is useful for limiting the number of batches processed or collected,
// such as for pagination, previewing, or restricting resource usage.
func (it *BatchIterator[T]) Take(n int) Iterator[[]T] {
	return &BatchIterator[T]{
		slice:     it.slice,
		batchSize: it.batchSize,
		position:  it.position,
		taker:     n,
		mapper:    it.mapper,
		filter:    it.filter,
		reducer:   it.reducer,
	}
}

// Reduce combines all batches using the provided function.
// It is a generic method that takes a function which receives an accumulator (slice of T)
// and the next batch (slice of T), and returns the updated accumulator.
// The method repeatedly applies the function to accumulate a single result from all batches.
// Example usage:
//
//	iterator := NewBatchIterator([]int{1, 2, 3, 4, 5}, 2)
//	result := iterator.Reduce(func(acc, batch []int) []int {
//	    return append(acc, batch...)
//	})
//	fmt.Println(result) // Output: [1 2 3 4 5]
//
// This method is useful for reducing a collection of batches into a single result,
// such as flattening, aggregating, or merging data from all batches.
func (it *BatchIterator[T]) Reduce(fn func([]T, []T) []T) []T {
	var result []T
	first := true

	// Process all batches
	for {
		batch, hasMore := it.Next()
		if len(batch) == 0 && !hasMore {
			break
		}

		// For the first batch, just set it as the initial result
		if first {
			result = batch
			first = false
		} else {
			// For subsequent batches, combine with the accumulated result
			result = fn(result, batch)
		}

		if !hasMore {
			break
		}
	}

	return result
}

// Next returns the next batch of elements
// It applies mapper, filter and take operations as specified
func (it *BatchIterator[T]) Next() ([]T, bool) {
	// If taker is set and has reached 0, we've taken all requested batches
	if it.taker == 0 {
		var zeroVal []T
		return zeroVal, false
	}

	// Try to get the next batch
	batch, hasMore := getNextBatch(it)

	// If we got a batch, process it according to transformations
	if len(batch) > 0 || hasMore {
		// Apply mapper if specified
		if it.mapper != nil {
			batch = it.mapper(batch)
		}

		// Apply filter if specified
		if it.filter != nil && !it.filter(batch) {
			// If filter rejected this batch, recursively try to get the next one
			return it.Next()
		}

		// Decrement taker if specified
		if it.taker > 0 {
			it.taker--
			// If we just used the last take, indicate no more items
			if it.taker == 0 {
				hasMore = false
			}
		}
	}

	return batch, hasMore
}

// getNextBatch retrieves the next batch from the underlying slice
// This is a helper function to avoid code duplication
func getNextBatch[T any](it *BatchIterator[T]) ([]T, bool) {
	if it.slice == nil || it.position >= len(it.slice) {
		var zeroVal []T
		return zeroVal, false
	}

	end := it.position + it.batchSize
	if end > len(it.slice) {
		end = len(it.slice)
	}

	batch := it.slice[it.position:end]
	it.position = end

	return batch, it.position < len(it.slice)
}

// Collect returns all remaining batches as a slice of slices
// It applies any mapper, filter and take operations before collecting
func (it *BatchIterator[T]) Collect() [][]T {
	var results [][]T
	batchCount := 0

	for {
		batch, hasMore := it.Next()

		// Only add non-empty batches
		if len(batch) > 0 {
			results = append(results, batch)
			batchCount++
		}

		// Exit loop if no more batches or we've reached the take limit
		if !hasMore {
			break
		}
	}

	return results
}

// Execute processes each batch from the iterator with the provided function.
// It is a method that works with any type in the BatchIterator.
// Example usage:
//
//	arr := []int{1, 2, 3, 4, 5}
//	iterator := NewBatchIterator(arr, 2)
//	err := iterator.Execute(func(batch []int) error {
//		fmt.Println(batch)
//		return nil
//	})
//	// Output:
//	// [1 2]
//	// [3 4]
//	// [5]
//
// This method is useful for performing operations on batches from the iterator
// while handling potential errors. Processing stops on the first error encountered
// and returns it. If no errors occur, nil is returned after all batches are processed.
func (it *BatchIterator[T]) Execute(f func([]T) error) error {
	for {
		batch, hasMore := it.Next()
		if err := f(batch); err != nil {
			return err
		}

		if !hasMore {
			break
		}
	}

	return nil
}

// ExecuteWithReturn processes each batch from the iterator with the provided function and collects results.
// It is a generic method that works with any type in the BatchIterator and any return type R.
// Example usage:
//
//	arr := []int{1, 2, 3, 4, 5}
//	iterator := NewBatchIterator(arr, 2)
//	results, err := ExecuteWithReturn(iterator, func(batch []int) (string, error) {
//		return fmt.Sprintf("Batch: %v", batch), nil
//	})
//	// results: ["Batch: [1 2]", "Batch: [3 4]", "Batch: [5]"]
//
// This method is useful for transforming batches into a different type while collecting all results.
// Processing stops on the first error encountered and returns it along with nil results.
// If no errors occur, all results are returned after all batches are processed.
func ExecuteWithReturn[T any, R any](it *BatchIterator[T], f func([]T) (R, error)) ([]R, error) {
	var results []R

	for {
		batch, hasMore := it.Next()
		result, err := f(batch)
		if err != nil {
			return nil, err
		}
		results = append(results, result)

		if !hasMore {
			break
		}
	}

	return results, nil
}
