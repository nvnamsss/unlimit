package utility

import (
	"fmt"
	"testing"
)

// Benchmark for UniqueArray function
func BenchmarkUniqueArray(b *testing.B) {
	b.Run("SmallSlice", func(b *testing.B) {
		data := []int{1, 2, 2, 3, 3, 3, 4, 4, 4, 4}
		for b.Loop() {
			UniqueArray(data)
		}
	})

	b.Run("LargeSlice", func(b *testing.B) {
		data := make([]int, 10000)
		for i := range data {
			data[i] = i % 100 // Create duplicates
		}
		for b.Loop() {
			UniqueArray(data)
		}
	})

	b.Run("NoDuplicates", func(b *testing.B) {
		data := make([]int, 1000)
		for i := range data {
			data[i] = i // No duplicates
		}
		for b.Loop() {
			UniqueArray(data)
		}
	})
}

// Benchmark for Array2Map function
func BenchmarkArray2Map(b *testing.B) {
	type User struct {
		ID   int
		Name string
	}

	b.Run("SmallSlice", func(b *testing.B) {
		users := make([]User, 100)
		for i := range users {
			users[i] = User{ID: i, Name: fmt.Sprintf("User%d", i)}
		}
		for b.Loop() {
			Array2Map(users, func(u User) int { return u.ID })
		}
	})

	b.Run("LargeSlice", func(b *testing.B) {
		users := make([]User, 10000)
		for i := range users {
			users[i] = User{ID: i, Name: fmt.Sprintf("User%d", i)}
		}
		for b.Loop() {
			Array2Map(users, func(u User) int { return u.ID })
		}
	})

	b.Run("WithDuplicateKeys", func(b *testing.B) {
		users := make([]User, 1000)
		for i := range users {
			users[i] = User{ID: i % 10, Name: fmt.Sprintf("User%d", i)} // Duplicate keys
		}
		for b.Loop() {
			Array2Map(users, func(u User) int { return u.ID })
		}
	})
}

// Benchmark for UniqueKeys function
func BenchmarkUniqueKeys(b *testing.B) {
	type User struct {
		Role string
		Name string
	}

	b.Run("SmallSlice", func(b *testing.B) {
		users := make([]User, 100)
		roles := []string{"admin", "user", "guest"}
		for i := range users {
			users[i] = User{Role: roles[i%len(roles)], Name: fmt.Sprintf("User%d", i)}
		}
		for b.Loop() {
			UniqueKeys(users, func(u User) string { return u.Role })
		}
	})

	b.Run("LargeSlice", func(b *testing.B) {
		users := make([]User, 10000)
		roles := []string{"admin", "user", "guest", "moderator", "viewer"}
		for i := range users {
			users[i] = User{Role: roles[i%len(roles)], Name: fmt.Sprintf("User%d", i)}
		}
		for b.Loop() {
			UniqueKeys(users, func(u User) string { return u.Role })
		}
	})
}

// Benchmark for IsContains function
func BenchmarkIsContains(b *testing.B) {
	b.Run("SmallSlice-Found", func(b *testing.B) {
		data := make([]int, 100)
		for i := range data {
			data[i] = i
		}
		target := 50
		for b.Loop() {
			IsContains(data, target)
		}
	})

	b.Run("SmallSlice-NotFound", func(b *testing.B) {
		data := make([]int, 100)
		for i := range data {
			data[i] = i
		}
		target := 200
		for b.Loop() {
			IsContains(data, target)
		}
	})

	b.Run("LargeSlice-Found", func(b *testing.B) {
		data := make([]int, 10000)
		for i := range data {
			data[i] = i
		}
		target := 5000
		for b.Loop() {
			IsContains(data, target)
		}
	})

	b.Run("LargeSlice-NotFound", func(b *testing.B) {
		data := make([]int, 10000)
		for i := range data {
			data[i] = i
		}
		target := 20000
		for b.Loop() {
			IsContains(data, target)
		}
	})
}

// Benchmark for ChunkArray function
func BenchmarkChunkArray(b *testing.B) {
	b.Run("SmallSlice", func(b *testing.B) {
		data := make([]int, 100)
		for i := range data {
			data[i] = i
		}
		for b.Loop() {
			ChunkArray(data, 10)
		}
	})

	b.Run("LargeSlice", func(b *testing.B) {
		data := make([]int, 10000)
		for i := range data {
			data[i] = i
		}
		for b.Loop() {
			ChunkArray(data, 100)
		}
	})

	b.Run("LimitExceedsSize", func(b *testing.B) {
		data := make([]int, 100)
		for i := range data {
			data[i] = i
		}
		for b.Loop() {
			ChunkArray(data, 200)
		}
	})
}

// Benchmark for BatchArray function
func BenchmarkBatchArray(b *testing.B) {
	b.ReportAllocs()

	b.Run("SmallSlice", func(b *testing.B) {
		data := make([]int, 100)
		for i := range data {
			data[i] = i
		}
		for b.Loop() {
			BatchArray(data, 10)
		}
	})

	b.Run("LargeSlice", func(b *testing.B) {
		data := make([]int, 10000)
		for i := range data {
			data[i] = i
		}
		for b.Loop() {
			BatchArray(data, 100)
		}
	})

	b.Run("SmallBatchSize", func(b *testing.B) {
		data := make([]int, 1000)
		for i := range data {
			data[i] = i
		}
		for b.Loop() {
			BatchArray(data, 5)
		}
	})

	b.Run("LargeBatchSize", func(b *testing.B) {
		data := make([]int, 1000)
		for i := range data {
			data[i] = i
		}
		for b.Loop() {
			BatchArray(data, 500)
		}
	})
}

// Benchmark for BatchArrayIterator function
func BenchmarkBatchArrayIterator(b *testing.B) {
	b.Run("SmallSlice", func(b *testing.B) {
		data := make([]int, 100)
		for i := range data {
			data[i] = i
		}
		for b.Loop() {
			iterator := BatchArrayIterator(data, 10)
			for {
				_, hasMore := iterator.Next()
				if !hasMore {
					break
				}
			}
		}
	})

	b.Run("LargeSlice", func(b *testing.B) {
		data := make([]int, 10000)
		for i := range data {
			data[i] = i
		}
		for b.Loop() {
			iterator := BatchArrayIterator(data, 100)
			for {
				_, hasMore := iterator.Next()
				if !hasMore {
					break
				}
			}
		}
	})
}

// Benchmark for Reverse function
func BenchmarkReverse(b *testing.B) {
	b.Run("SmallSlice", func(b *testing.B) {
		for b.Loop() {
			b.StopTimer()
			data := make([]int, 100)
			for i := range data {
				data[i] = i
			}
			b.StartTimer()
			Reverse(data)
		}
	})

	b.Run("LargeSlice", func(b *testing.B) {
		for b.Loop() {
			b.StopTimer()
			data := make([]int, 10000)
			for i := range data {
				data[i] = i
			}
			b.StartTimer()
			Reverse(data)
		}
	})
}

// Benchmark for Filter function
func BenchmarkFilter(b *testing.B) {
	b.ReportAllocs()

	b.Run("SmallSlice-EvenNumbers", func(b *testing.B) {
		data := make([]int, 100)
		for i := range data {
			data[i] = i
		}
		for b.Loop() {
			Filter(data, func(n int) bool { return n%2 == 0 })
		}
	})

	b.Run("LargeSlice-EvenNumbers", func(b *testing.B) {
		data := make([]int, 10000)
		for i := range data {
			data[i] = i
		}
		for b.Loop() {
			Filter(data, func(n int) bool { return n%2 == 0 })
		}
	})

	b.Run("SmallSlice-HighSelectivity", func(b *testing.B) {
		data := make([]int, 100)
		for i := range data {
			data[i] = i
		}
		for b.Loop() {
			Filter(data, func(n int) bool { return n > 90 }) // Few matches
		}
	})

	b.Run("SmallSlice-LowSelectivity", func(b *testing.B) {
		data := make([]int, 100)
		for i := range data {
			data[i] = i
		}
		for b.Loop() {
			Filter(data, func(n int) bool { return n < 90 }) // Many matches
		}
	})
}

// Benchmark for RemoveDuplicates function
func BenchmarkRemoveDuplicates(b *testing.B) {
	b.ReportAllocs()

	b.Run("SmallSlice-ManyDuplicates", func(b *testing.B) {
		data := make([]int, 100)
		for i := range data {
			data[i] = i % 10 // Lots of duplicates
		}
		for b.Loop() {
			RemoveDuplicates(data)
		}
	})

	b.Run("LargeSlice-ManyDuplicates", func(b *testing.B) {
		data := make([]int, 10000)
		for i := range data {
			data[i] = i % 100 // Lots of duplicates
		}
		for b.Loop() {
			RemoveDuplicates(data)
		}
	})

	b.Run("SmallSlice-NoDuplicates", func(b *testing.B) {
		data := make([]int, 100)
		for i := range data {
			data[i] = i // No duplicates
		}
		for b.Loop() {
			RemoveDuplicates(data)
		}
	})
}

// Benchmark for FormatList function
func BenchmarkFormatList(b *testing.B) {
	b.ReportAllocs()

	b.Run("SmallSlice-IntToString", func(b *testing.B) {
		data := make([]int, 100)
		for i := range data {
			data[i] = i
		}
		for b.Loop() {
			FormatList(data, func(v int) string {
				return fmt.Sprintf("Number: %d", v)
			})
		}
	})

	b.Run("LargeSlice-IntToString", func(b *testing.B) {
		data := make([]int, 10000)
		for i := range data {
			data[i] = i
		}
		for b.Loop() {
			FormatList(data, func(v int) string {
				return fmt.Sprintf("Number: %d", v)
			})
		}
	})

	b.Run("SmallSlice-SimpleTransform", func(b *testing.B) {
		data := make([]int, 100)
		for i := range data {
			data[i] = i
		}
		for b.Loop() {
			FormatList(data, func(v int) int {
				return v * 2
			})
		}
	})
}

// Comparison benchmark between different implementations
func BenchmarkUniqueComparison(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i % 100 // Create duplicates
	}

	b.Run("UniqueArray", func(b *testing.B) {
		for b.Loop() {
			UniqueArray(data)
		}
	})

	b.Run("RemoveDuplicates", func(b *testing.B) {
		for b.Loop() {
			RemoveDuplicates(data)
		}
	})
}
