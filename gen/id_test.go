package gen

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestUUIDGenerator_Generate_NoPrefix(t *testing.T) {
	// Create a generator with no prefix
	generator := NewUUIDGenerator("")

	// Generate an ID
	id := generator.Generate()

	// Verify that the ID is non-empty
	if id == "" {
		t.Error("Expected non-empty ID, got empty string")
	}

	// Verify UUID format (basic check - should be 36 chars with hyphens)
	if len(id) != 36 || !strings.Contains(id, "-") {
		t.Errorf("Expected UUID format, got: %s", id)
	}

	// Generate another ID to ensure uniqueness
	id2 := generator.Generate()
	if id == id2 {
		t.Error("Expected unique IDs, got the same ID twice")
	}
}

func TestUUIDGenerator_Generate_WithPrefix(t *testing.T) {
	// Create a generator with a prefix
	prefix := "test"
	generator := NewUUIDGenerator(prefix)

	// Generate an ID
	id := generator.Generate()

	// Verify that the ID starts with the prefix (without underscore)
	if !strings.HasPrefix(id, prefix) {
		t.Errorf("Expected ID to start with '%s', got: %s", prefix, id)
	}

	// Verify the remaining part is a valid UUID (basic check)
	uuidPart := strings.TrimPrefix(id, prefix)
	if len(uuidPart) != 36 || !strings.Contains(uuidPart, "-") {
		t.Errorf("Expected UUID format after prefix, got: %s", uuidPart)
	}
}

func TestNewUUIDGenerator(t *testing.T) {
	// Test that the factory function returns a valid IDGenerator
	prefix := "prefix"
	generator := NewUUIDGenerator(prefix)

	// Type assertion to check if it's a UUIDGenerator
	_, ok := generator.(*UUIDGenerator)
	if !ok {
		t.Error("Expected NewUUIDGenerator to return *UUIDGenerator")
	}

	// Check if it works as expected
	id := generator.Generate()
	if !strings.HasPrefix(id, prefix) {
		t.Errorf("Expected ID to start with '%s', got: %s", prefix, id)
	}
}

// Benchmarks for ID generation
func BenchmarkUUIDGenerator_Generate_NoPrefix(b *testing.B) {
	generator := NewUUIDGenerator("")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generator.Generate()
	}
}

func BenchmarkUUIDGenerator_Generate_WithPrefix(b *testing.B) {
	generator := NewUUIDGenerator("test-prefix")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generator.Generate()
	}
}

func BenchmarkUUIDGenerator_Generate_LongPrefix(b *testing.B) {
	generator := NewUUIDGenerator("very-long-prefix-for-testing-performance-impact")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generator.Generate()
	}
}

func BenchmarkNewUUIDGenerator(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewUUIDGenerator("bench")
	}
}

// RandomTokenGenerator Tests

func TestRandomTokenGenerator_New(t *testing.T) {
	// Test valid length
	generator := NewRandomTokenGenerator(8)
	_, ok := generator.(*RandomTokenGenerator)
	if !ok {
		t.Error("Expected NewRandomTokenGenerator to return *RandomTokenGenerator")
	}

	// Test zero length (should default to 8)
	generator = NewRandomTokenGenerator(0)
	id := generator.Generate()
	if len(id) != 8 {
		t.Errorf("Expected default length of 8, got %d", len(id))
	}

	// Test negative length (should default to 8)
	generator = NewRandomTokenGenerator(-5)
	id = generator.Generate()
	if len(id) != 8 {
		t.Errorf("Expected default length of 8, got %d", len(id))
	}

	// Test length greater than 36 (should default to 8)
	generator = NewRandomTokenGenerator(50)
	id = generator.Generate()
	if len(id) != 8 {
		t.Errorf("Expected default length of 8, got %d", len(id))
	}
}

func TestRandomTokenGenerator_Generate(t *testing.T) {
	// Test with specific length
	length := 12
	generator := NewRandomTokenGenerator(length)

	// Generate an ID
	id := generator.Generate()

	// Verify length
	if len(id) != length {
		t.Errorf("Expected ID length %d, got %d", length, len(id))
	}

	// Verify it's non-empty
	if id == "" {
		t.Error("Expected non-empty ID, got empty string")
	}

	// Generate another ID to check uniqueness
	id2 := generator.Generate()
	if id == id2 {
		t.Error("Expected unique IDs, got the same ID twice")
	}

	// Verify it contains only valid UUID characters (hex + hyphens)
	validChars := "0123456789abcdef-"
	for _, char := range id {
		if !strings.ContainsRune(validChars, char) {
			t.Errorf("ID contains invalid character: %c", char)
		}
	}
}

// SequentialIDGenerator Tests

func TestSequentialIDGenerator_New(t *testing.T) {
	// Test with positive start value
	start := 10
	generator := NewSequentialIDGenerator(start)
	_, ok := generator.(*SequentialIDGenerator)
	if !ok {
		t.Error("Expected NewSequentialIDGenerator to return *SequentialIDGenerator")
	}

	// Test with zero start value - should start generating from 0
	generator = NewSequentialIDGenerator(0)
	id := generator.Generate()
	if id != "0" {
		t.Errorf("Expected first ID to be '0', got '%s'", id)
	}

	// Test with negative start value (should default to 0)
	generator = NewSequentialIDGenerator(-5)
	id = generator.Generate()
	if id != "0" {
		t.Errorf("Expected first ID to be '0' for negative start, got '%s'", id)
	}
}

func TestSequentialIDGenerator_Generate(t *testing.T) {
	// Test sequential generation starting from 1
	generator := NewSequentialIDGenerator(1)

	// Generate first ID (starts at 1, current = 0, increments to 1)
	id1 := generator.Generate()
	if id1 != "1" {
		t.Errorf("Expected first ID to be '1', got '%s'", id1)
	}

	// Generate second ID
	id2 := generator.Generate()
	if id2 != "2" {
		t.Errorf("Expected second ID to be '2', got '%s'", id2)
	}

	// Generate third ID
	id3 := generator.Generate()
	if id3 != "3" {
		t.Errorf("Expected third ID to be '3', got '%s'", id3)
	}
}

func TestSequentialIDGenerator_Generate_DifferentStartValues(t *testing.T) {
	testCases := []struct {
		start    int
		expected []string
	}{
		{0, []string{"0", "1", "2"}},         // Starts at 0, generates 0, 1, 2
		{5, []string{"5", "6", "7"}},         // Starts at 5, generates 5, 6, 7
		{100, []string{"100", "101", "102"}}, // Starts at 100, generates 100, 101, 102
		{-1, []string{"0", "1", "2"}},        // negative should default to 0, then generate 0, 1, 2
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("start_%d", tc.start), func(t *testing.T) {
			generator := NewSequentialIDGenerator(tc.start)

			for i, expectedID := range tc.expected {
				actualID := generator.Generate()
				if actualID != expectedID {
					t.Errorf("Generation %d: expected '%s', got '%s'", i+1, expectedID, actualID)
				}
			}
		})
	}
}

// Concurrency Tests

func TestUUIDGenerator_ConcurrentGeneration(t *testing.T) {
	generator := NewUUIDGenerator("concurrent")
	const numGoroutines = 50
	const numIDsPerGoroutine = 10

	var wg sync.WaitGroup
	idChan := make(chan string, numGoroutines*numIDsPerGoroutine)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIDsPerGoroutine; j++ {
				id := generator.Generate()
				idChan <- id
			}
		}()
	}

	wg.Wait()
	close(idChan)

	// Collect all IDs and verify uniqueness
	ids := make(map[string]bool)
	for id := range idChan {
		if ids[id] {
			t.Errorf("Duplicate ID found: %s", id)
		}
		ids[id] = true
	}

	expectedCount := numGoroutines * numIDsPerGoroutine
	if len(ids) != expectedCount {
		t.Errorf("Expected %d unique IDs, got %d", expectedCount, len(ids))
	}
}

func TestSequentialIDGenerator_ConcurrentGeneration(t *testing.T) {
	generator := NewSequentialIDGenerator(0)
	const numGoroutines = 10
	const numIDsPerGoroutine = 10

	var wg sync.WaitGroup
	idChan := make(chan string, numGoroutines*numIDsPerGoroutine)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIDsPerGoroutine; j++ {
				id := generator.Generate()
				idChan <- id
			}
		}()
	}

	wg.Wait()
	close(idChan)

	// Collect all IDs
	ids := make(map[string]bool)
	duplicateCount := 0
	for id := range idChan {
		if ids[id] {
			duplicateCount++
		}
		ids[id] = true
	}

	// Note: SequentialIDGenerator is NOT thread-safe by design
	// This test demonstrates the race condition issue
	if duplicateCount == 0 {
		// This would be unexpected - we expect some race conditions
		t.Log("Warning: No race conditions detected - this is unusual for SequentialIDGenerator")
	}
}

// Benchmarks for other generators

func BenchmarkRandomTokenGenerator_Generate(b *testing.B) {
	generator := NewRandomTokenGenerator(8)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generator.Generate()
	}
}

func BenchmarkSequentialIDGenerator_Generate(b *testing.B) {
	generator := NewSequentialIDGenerator(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generator.Generate()
	}
}
