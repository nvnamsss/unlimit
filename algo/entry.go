package algo

// Entry is used to hold a value in the evictList
type Entry[K comparable, V any] struct {
	Key   K
	Value V
}
