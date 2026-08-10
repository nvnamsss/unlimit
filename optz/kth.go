package optz

// FindKthElement finds the kth element in the slice using the provided comparator
// Time complexity: O(n)
// Space complexity: O(1)
func FindKthElement[T any](arr []T, k int, less func(a, b T) bool) T {
	left, right := 0, len(arr)-1

	for left <= right {
		pivotIdx := partition(arr, left, right, less)

		if pivotIdx == k-1 {
			return arr[pivotIdx]
		} else if pivotIdx < k-1 {
			left = pivotIdx + 1
		} else {
			right = pivotIdx - 1
		}
	}

	return arr[0] // return first element if k is invalid
}

func partition[T any](arr []T, left, right int, less func(a, b T) bool) int {
	pivot := arr[right]
	i := left - 1

	for j := left; j < right; j++ {
		if less(arr[j], pivot) {
			i++
			swap(arr, i, j)
		}
	}

	swap(arr, i+1, right)
	return i + 1
}

func swap[T any](arr []T, i, j int) {
	arr[i], arr[j] = arr[j], arr[i]
}
