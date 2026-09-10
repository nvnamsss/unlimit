package utility

import (
	"github.com/nvnamsss/unlimit/collections"
)

// UniqueArray returns a slice containing only the unique elements from the input slice.
// It is a generic function that works with any comparable type.
// Example usage:
//
//	arr := []int{1, 2, 2, 3, 3, 3}
//	result := UniqueArray(arr)
//	fmt.Println(result) // Output: [1 2 3] (order may vary)
//
// This function is useful for removing duplicate elements from a slice
// while preserving the uniqueness of each element.
func UniqueArray[K comparable](arr []K) (results []K) {
	m := make(map[K]bool)
	for idx := range arr {
		v := arr[idx]
		m[v] = true
	}
	for k := range m {
		results = append(results, k)
	}
	return
}

// Array2Map converts a slice to a map using a function to generate keys.
// It is a generic function that takes a slice of type V and a function that generates
// a key of type K for each element, returning a map from K to V.
// Example usage:
//
//	users := []User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
//	userMap := Array2Map(users, func(u User) int { return u.ID })
//	fmt.Println(userMap[1].Name) // Output: "Alice"
//
// This function is useful for creating lookup maps from arrays of objects.
// If multiple elements would produce the same key, only the first one is included.
func Array2Map[K comparable, V any](arr []V, fn func(V) K) map[K]V {
	m := make(map[K]V)
	if len(arr) == 0 {
		return m
	}
	for idx := range arr {
		value := arr[idx]
		key := fn(value)
		if _, ok := m[key]; ok {
			continue
		}
		m[key] = value
	}
	return m
}

// UniqueKeys extracts unique keys from a slice using a key extraction function.
// It is a generic function that takes a slice of type V and a function that extracts
// a key of type K from each element, returning a slice of unique keys.
// Example usage:
//
//	users := []User{{Role: "admin", Name: "Alice"}, {Role: "user", Name: "Bob"}, {Role: "admin", Name: "Charlie"}}
//	roles := UniqueKeys(users, func(u User) string { return u.Role })
//	fmt.Println(roles) // Output: ["admin", "user"] (order may vary)
//
// This function is useful for extracting a set of distinct values from a collection.
func UniqueKeys[K comparable, V any](array []V, uniqueKeyFn func(V) K) []K {
	uniqueMap := make(map[K]bool)
	var uniqueElements []K
	for _, v := range array {
		k := uniqueKeyFn(v)
		if _, ok := uniqueMap[k]; ok {
			continue
		}
		uniqueMap[k] = true
		uniqueElements = append(uniqueElements, k)
	}
	return uniqueElements
}

// IsContains checks if an element exists in a slice.
// It is a generic function that works with comparable types using direct equality comparison.
// Example usage:
//
//	arr := []string{"apple", "banana", "cherry"}
//	exists := IsContains(arr, "banana")
//	fmt.Println(exists) // Output: true
//
// This function is useful for determining if a slice contains a specific element.
func IsContains[V comparable](array []V, element V) bool {
	for _, v := range array {
		if v == element {
			return true
		}
	}
	return false
}

// ChunkArray returns a slice of the first 'limit' elements from the input slice.
// It is a generic function that works with any type.
// Example usage:
//
//	arr := []int{1, 2, 3, 4, 5}
//	chunk := ChunkArray(arr, 3)
//	fmt.Println(chunk) // Output: [1, 2, 3]
//
// This function is useful for pagination or limiting the size of a result set.
// If the input slice length is less than the limit, the entire slice is returned.
func ChunkArray[T any](slice []T, limit int) []T {
	if len(slice) < limit {
		return slice
	}
	return slice[:limit]
}

// BatchArray divides a slice into batches of a specified size.
// It is a generic function that works with any type.
// Example usage:
//
//	arr := []int{1, 2, 3, 4, 5}
//	batches := BatchArray(arr, 2)
//	fmt.Println(batches) // Output: [[1 2] [3 4] [5]]
//
// This function is useful for processing large slices in smaller chunks.
// If batchSize is less than or equal to zero, it returns nil.
// Each batch will contain at most batchSize elements, with the last batch
// potentially containing fewer elements.
func BatchArray[T any](slice []T, batchSize int) [][]T {
	if batchSize <= 0 {
		return nil
	}

	var batches [][]T
	for i := 0; i < len(slice); i += batchSize {
		end := i + batchSize
		if end > len(slice) {
			end = len(slice)
		}
		batches = append(batches, slice[i:end])
	}
	return batches
}

// BatchArrayIterator creates an iterator that processes a slice in batches without storing all batches in memory.
// It is a generic function that works with any type.
// Example usage:
//
//	arr := []int{1, 2, 3, 4, 5}
//	iterator := BatchArrayIterator(arr, 2)
//	for {
//	    batch, hasMore := iterator.Next()
//	    fmt.Println(batch) // First outputs: [1 2], then [3 4], then [5]
//	    if !hasMore {
//	        break
//	    }
//	}
//
// This function is memory-efficient for processing large slices in smaller chunks
// as it only creates one batch at a time when Next() is called.
// If batchSize is less than or equal to zero, it returns an empty iterator.
func BatchArrayIterator[T any](slice []T, batchSize int) collections.Iterator[[]T] {
	return collections.NewBatchIterator(slice, batchSize)
}

// Reverse reverses the order of elements in a slice in-place.
// It is a generic function that works with any type.
// Example usage:
//
//	arr := []int{1, 2, 3, 4, 5}
//	Reverse(arr)
//	fmt.Println(arr) // Output: [5, 4, 3, 2, 1]
//
// This function is useful for inverting the order of elements in a slice.
// The operation is performed in-place, modifying the original slice.
func Reverse[T any](slice []T) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// Filter returns a new slice containing only the elements that satisfy the predicate.
// It is a generic function that takes a slice of type T and a predicate function
// that returns a boolean value for each element.
// Example usage:
//
//	arr := []int{1, 2, 3, 4, 5}
//	even := Filter(arr, func(n int) bool { return n%2 == 0 })
//	fmt.Println(even) // Output: [2, 4]
//
// This function is useful for extracting a subset of elements that meet specific criteria.
func Filter[T any](slice []T, pred func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if pred(v) {
			result = append(result, v)
		}
	}
	return result
}

// RemoveDuplicates removes duplicate elements from a slice.
// It is a generic function that works with any comparable type.
// Example usage:
//
//	arr := []int{1, 2, 2, 3, 3, 3}
//	unique := RemoveDuplicates(arr)
//	fmt.Println(unique) // Output: [1, 2, 3]
//
// This function is useful for ensuring each element appears only once in the result.
// Unlike UniqueArray, this function preserves the original order of elements.
func RemoveDuplicates[T comparable](slice []T) []T {
	seen := make(map[T]struct{})
	var result []T
	for _, item := range slice {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// FormatList applies a function to each element of an array and returns a new array with the results.
// It is a generic function that takes an array of type V and a function that transforms V
// into type T, returning a slice of type T.
// Example usage:
//
//	arr := []int{1, 2, 3}
//	result := FormatList(arr, func(v int) string {
//		return fmt.Sprintf("Number: %d", v)
//	})
//	fmt.Println(result) // Output: ["Number: 1", "Number: 2", "Number: 3"]
//
// This function is useful for transforming arrays of any type into another type
// using a provided transformation function.
// It can be used for various purposes, such as formatting, mapping, or transforming data.
func FormatList[V, T any](arr []V, fn func(V) T) (results []T) {
	results = make([]T, 0, len(arr))
	for _, v := range arr {
		results = append(results, fn(v))
	}
	return
}

// TransformArray applies a function to each element of an array and returns a new array with the results.
// It is a generic function that takes an array of type V and a function that transforms V
// into type T, returning a slice of type T.
// Example usage:
//
//	arr := []int{1, 2, 3}
//	result := TransformArray(arr, func(v int) string {
//		return fmt.Sprintf("Number: %d", v)
//	})
//	fmt.Println(result) // Output: ["Number: 1", "Number: 2", "Number: 3"]
//
// This function is useful for transforming arrays of any type into another type
// using a provided transformation function.
// It is an alias for FormatList and provides the same functionality with a different name
// that may be more intuitive in certain contexts.
func TransformArray[V any, T any](arr []V, fn func(V) T) []T {
	return FormatList(arr, fn)
}

// Keys returns a slice containing all the keys from the input map.
// It is a generic function that works with any comparable key type.
// Example usage:
//
//	m := map[string]int{"a": 1, "b": 2}
//	keys := Keys(m)
//	fmt.Println(keys) // Output: ["a", "b"] (order may vary)
//
// This function is useful for extracting all keys from a map into a slice.
func Keys[V comparable, T any](m map[V]T) (keys []V) {
	keys = make([]V, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return
}

// Values returns a slice containing all the values from the input map.
// It is a generic function that works with any value type.
// Example usage:
//
//	m := map[string]int{"a": 1, "b": 2}
//	values := Values(m)
//	fmt.Println(values) // Output: [1, 2] (order may vary)
//
// This function is useful for extracting all values from a map into a slice.
func Values[V comparable, T any](m map[V]T) (values []T) {
	values = make([]T, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return
}
