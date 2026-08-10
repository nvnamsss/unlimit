package algo

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
)

// Helper functions for generating test data
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func generateWordList(count int, minLength, maxLength int) []string {
	words := make([]string, count)
	for i := 0; i < count; i++ {
		length := minLength
		if maxLength > minLength {
			length = rand.Intn(maxLength-minLength) + minLength
		}
		words[i] = generateRandomString(length)
	}
	return words
}

// BenchmarkTrie_Insert benchmarks the Insert operation
func BenchmarkTrie_Insert(b *testing.B) {
	words := generateWordList(1000, 5, 15)
	trie := NewTrie()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use modulo to cycle through the word list
		trie.Insert(words[i%len(words)])
	}
}

// BenchmarkTrie_Search benchmarks the Search operation
func BenchmarkTrie_Search(b *testing.B) {
	// Setup
	words := generateWordList(1000, 5, 15)
	trie := NewTrie()
	for _, word := range words {
		trie.Insert(word)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Search for words that exist in the trie
		trie.Search(words[i%len(words)])
	}
}

// BenchmarkTrie_SearchMiss benchmarks the Search operation for non-existent words
func BenchmarkTrie_SearchMiss(b *testing.B) {
	// Setup
	words := generateWordList(1000, 5, 15)
	trie := NewTrie()
	for _, word := range words {
		trie.Insert(word)
	}

	// Generate words that don't exist in the trie
	missWords := generateWordList(1000, 5, 15)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Search for words that don't exist in the trie
		trie.Search(missWords[i%len(missWords)])
	}
}

// BenchmarkTrie_StartsWith benchmarks the StartsWith operation
func BenchmarkTrie_StartsWith(b *testing.B) {
	// Setup
	words := generateWordList(1000, 5, 15)
	trie := NewTrie()
	for _, word := range words {
		trie.Insert(word)
	}

	// Create prefixes of various lengths
	prefixes := make([]string, len(words))
	for i, word := range words {
		prefixLen := len(word) / 2
		if prefixLen > 0 {
			prefixes[i] = word[:prefixLen]
		} else {
			prefixes[i] = word
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Check if trie contains words with the prefix
		trie.StartsWith(prefixes[i%len(prefixes)])
	}
}

// BenchmarkTrie_Delete benchmarks the Delete operation
func BenchmarkTrie_Delete(b *testing.B) {
	b.Run("ExistingWords", func(b *testing.B) {
		words := generateWordList(1000, 5, 15)
		trie := NewTrie()
		for _, word := range words {
			trie.Insert(word)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			trie.Delete(words[i%len(words)])
		}
	})

	b.Run("ExistingWordsWithModifier", func(b *testing.B) {
		words := generateWordList(1000, 5, 15)
		trie := NewTrie().WithBloomFilter(1000, 0.01)
		for _, word := range words {
			trie.Insert(word)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			trie.Delete(words[i%len(words)])
		}
	})

	b.Run("NonExistentWords", func(b *testing.B) {
		words := generateWordList(1000, 5, 15)
		missWords := generateWordList(1000, 5, 15)
		trie := NewTrie()
		for _, word := range words {
			trie.Insert(word)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			trie.Delete(missWords[i%len(missWords)])
		}
	})

	b.Run("NonExistentWordsWithModifier", func(b *testing.B) {
		words := generateWordList(1000, 5, 15)
		missWords := generateWordList(1000, 5, 15)
		trie := NewTrie().WithBloomFilter(1000, 0.01)
		for _, word := range words {
			trie.Insert(word)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			trie.Delete(missWords[i%len(missWords)])
		}
	})
}

// BenchmarkTrie_GetAllWords benchmarks the GetAllWords operation
func BenchmarkTrie_GetAllWords(b *testing.B) {
	// Test with different dictionary sizes
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			words := generateWordList(size, 5, 15)
			trie := NewTrie()
			for _, word := range words {
				trie.Insert(word)
			}

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = trie.GetAllWords()
			}
		})
	}
}

// BenchmarkTrie_InsertWithDifferentSizes benchmarks Insert with different data sizes
func BenchmarkTrie_InsertWithDifferentSizes(b *testing.B) {
	// Test different word lengths and counts
	scenarios := []struct {
		wordCount int
		minLength int
		maxLength int
		name      string
	}{
		{100, 3, 5, "Small-Short"},
		{100, 10, 20, "Small-Long"},
		{1000, 3, 5, "Medium-Short"},
		{1000, 10, 20, "Medium-Long"},
		{5000, 3, 5, "Large-Short"},
		{5000, 10, 20, "Large-Long"},
	}

	for _, s := range scenarios {
		words := generateWordList(s.wordCount, s.minLength, s.maxLength)

		b.Run(s.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				trie := NewTrie()
				b.StopTimer()
				// Only insert a subset to avoid timing becoming dominated by insert count
				insertCount := 100
				if s.wordCount < insertCount {
					insertCount = s.wordCount
				}
				b.StartTimer()

				for j := 0; j < insertCount; j++ {
					trie.Insert(words[j])
				}
			}
		})
	}
}

// BenchmarkTrie_ParallelOperations benchmarks parallel operations
func BenchmarkTrie_ParallelOperations(b *testing.B) {
	words := generateWordList(1000, 5, 15)
	trie := NewTrie()
	for _, word := range words {
		trie.Insert(word)
	}

	b.Run("ParallelSearch", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				trie.Search(words[i%len(words)])
				i++
			}
		})
	})

	b.Run("ParallelStartsWith", func(b *testing.B) {
		// Create prefixes of various lengths
		prefixes := make([]string, len(words))
		for i, word := range words {
			prefixLen := len(word) / 2
			if prefixLen > 0 {
				prefixes[i] = word[:prefixLen]
			} else {
				prefixes[i] = word
			}
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				trie.StartsWith(prefixes[i%len(prefixes)])
				i++
			}
		})
	})
}

// BenchmarkTrie_RealWorldScenario simulates a more realistic usage pattern
func BenchmarkTrie_RealWorldScenario(b *testing.B) {
	// Create a dictionary of words
	words := generateWordList(5000, 3, 15)
	prefixes := make([]string, len(words))
	for i, word := range words {
		prefixLen := 1 + rand.Intn(len(word))
		prefixes[i] = word[:prefixLen]
	}

	b.Run("DictionaryLookup", func(b *testing.B) {
		trie := NewTrie()
		// Insert all words
		for _, word := range words {
			trie.Insert(word)
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Mix of operations: search, prefix check, insert
			op := i % 3
			idx := i % len(words)

			switch op {
			case 0:
				trie.Search(words[idx])
			case 1:
				trie.StartsWith(prefixes[idx])
			case 2:
				// New word based on index
				newWord := words[idx] + strconv.Itoa(i)
				trie.Insert(newWord)
			}
		}
	})
}

// BenchmarkTrie_MemoryUsage specifically focuses on memory allocation patterns
func BenchmarkTrie_MemoryUsage(b *testing.B) {
	b.Run("Insert", func(b *testing.B) {
		words := generateWordList(100, 5, 15)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			trie := NewTrie()
			for j := 0; j < 100; j++ {
				trie.Insert(words[j])
			}
		}
	})

	b.Run("Search", func(b *testing.B) {
		words := generateWordList(100, 5, 15)
		trie := NewTrie()
		for _, word := range words {
			trie.Insert(word)
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			trie.Search(words[i%len(words)])
		}
	})
}
