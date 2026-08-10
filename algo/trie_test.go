package algo

import (
	"reflect"
	"sort"
	"testing"
)

// TestTrie_New tests the creation of a new Trie
func TestTrie_New(t *testing.T) {
	trie := NewTrie()

	if trie == nil {
		t.Error("NewTrie() returned nil")
		return
	}

	if trie.root == nil {
		t.Error("Trie root should not be nil")
		return
	}

	if trie.root.isEnd {
		t.Error("New trie root should not be marked as end of word")
		return
	}

	if len(trie.root.children) != 0 {
		t.Error("New trie should have empty children map")
	}
}

// TestTrie_Insert tests inserting words into the trie
func TestTrie_Insert(t *testing.T) {
	trie := NewTrie()

	// Insert a single word
	trie.Insert("hello")

	// Verify root has 'h' as child
	if len(trie.root.children) != 1 {
		t.Errorf("Expected root to have 1 child, got %d", len(trie.root.children))
	}

	// Check if path exists for "hello"
	node := trie.root
	for _, ch := range "hello" {
		child, exists := node.children[ch]
		if !exists {
			t.Errorf("Character %c not found in trie path", ch)
		}
		node = child
	}

	if !node.isEnd {
		t.Error("Last node should be marked as end of word")
	}

	// Insert another word that shares prefix
	trie.Insert("help")

	// Now check both words
	if !trie.Search("hello") {
		t.Error("Failed to find 'hello' after insertion")
	}

	if !trie.Search("help") {
		t.Error("Failed to find 'help' after insertion")
	}
}

// TestTrie_InsertEmptyString tests handling of empty strings
func TestTrie_InsertEmptyString(t *testing.T) {
	trie := NewTrie()

	// Insert an empty string
	trie.Insert("")

	// Root should be marked as end of word
	if !trie.root.isEnd {
		t.Error("Root should be marked as end of word after inserting empty string")
	}

	// Search should return true for empty string
	if !trie.Search("") {
		t.Error("Search should find empty string after insertion")
	}
}

// TestTrie_Search tests searching for words in the trie
func TestTrie_Search(t *testing.T) {
	trie := NewTrie()
	words := []string{"apple", "application", "banana", "book", "cat"}

	// Insert test words
	for _, word := range words {
		trie.Insert(word)
	}

	// Test finding existing words
	for _, word := range words {
		if !trie.Search(word) {
			t.Errorf("Search failed to find '%s'", word)
		}
	}

	// Test non-existent words
	nonExistent := []string{"app", "ban", "bo", "doggy"}
	for _, word := range nonExistent {
		if trie.Search(word) {
			t.Errorf("Search incorrectly found '%s'", word)
		}
	}
}

// TestTrie_StartsWith tests prefix matching in the trie
func TestTrie_StartsWith(t *testing.T) {
	trie := NewTrie()
	words := []string{"apple", "application", "banana", "book"}

	// Insert test words
	for _, word := range words {
		trie.Insert(word)
	}

	// Test valid prefixes
	prefixes := []string{"a", "ap", "app", "b", "ba", "bo", ""}
	for _, prefix := range prefixes {
		if !trie.StartsWith(prefix) {
			t.Errorf("StartsWith failed for prefix '%s'", prefix)
		}
	}

	// Test invalid prefixes
	invalidPrefixes := []string{"c", "d", "boo-", "appp"}
	for _, prefix := range invalidPrefixes {
		if trie.StartsWith(prefix) {
			t.Errorf("StartsWith incorrectly matched prefix '%s'", prefix)
		}
	}
}

// TestTrie_Delete tests removing words from the trie
func TestTrie_Delete(t *testing.T) {
	trie := NewTrie()
	words := []string{"apple", "application", "banana", "book"}

	// Insert test words
	for _, word := range words {
		trie.Insert(word)
	}

	// Delete "apple" and verify it's gone but "application" remains
	result := trie.Delete("apple")
	if !result {
		t.Error("Delete should return true for existing word")
	}

	if trie.Search("apple") {
		t.Error("Word 'apple' should be deleted")
	}

	if !trie.Search("application") {
		t.Error("Word 'application' should still exist after deleting 'apple'")
	}

	// Delete non-existent word
	result = trie.Delete("nonexistent")
	if result {
		t.Error("Delete should return false for non-existent word")
	}

	// Delete empty string
	trie.Insert("")
	result = trie.Delete("")
	if !result {
		t.Error("Delete should return true for empty string")
	}

	if trie.Search("") {
		t.Error("Empty string should be deleted")
	}

	// Test with multi-byte rune characters
	multiRuneTrie := NewTrie()

	// Words with multi-byte characters (Chinese, emojis, etc.)
	multiRuneWords := []string{
		"你好",      // Chinese for "hello"
		"世界",      // Chinese for "world"
		"こんにちは",   // Japanese for "hello"
		"🌍🌎🌏",     // Earth emojis
		"café",    // Accented character
		"München", // German city name with umlaut
	}

	// Insert multi-rune words
	for _, word := range multiRuneWords {
		multiRuneTrie.Insert(word)
	}

	// Verify all words were inserted correctly
	for _, word := range multiRuneWords {
		if !multiRuneTrie.Search(word) {
			t.Errorf("Failed to find multi-rune word: %s", word)
		}
	}

	// Delete and verify each multi-rune word
	for _, word := range multiRuneWords {
		// Delete the word
		result = multiRuneTrie.Delete(word)
		if !result {
			t.Errorf("Failed to delete multi-rune word: %s", word)
		}

		// Verify it's gone
		if multiRuneTrie.Search(word) {
			t.Errorf("Word '%s' should be deleted", word)
		}
	}

	// Test partial matches with multi-rune words
	complexTrie := NewTrie()
	complexTrie.Insert("你好世界") // "Hello world" in Chinese
	complexTrie.Insert("你好朋友") // "Hello friend" in Chinese

	// Delete the first word
	result = complexTrie.Delete("你好世界")
	if !result {
		t.Error("Failed to delete '你好世界'")
	}

	// Verify first word is gone but second word remains
	if complexTrie.Search("你好世界") {
		t.Error("Word '你好世界' should be deleted")
	}

	if !complexTrie.Search("你好朋友") {
		t.Error("Word '你好朋友' should still exist")
	}

	// Verify prefix still works
	if !complexTrie.StartsWith("你好") {
		t.Error("Prefix '你好' should still be found")
	}
}

// TestTrie_GetAllWords tests retrieving all words from the trie
func TestTrie_GetAllWords(t *testing.T) {
	trie := NewTrie()
	words := []string{"apple", "application", "banana", "book", "cat"}

	// Insert test words
	for _, word := range words {
		trie.Insert(word)
	}

	// Get all words and compare
	result := trie.GetAllWords()

	// Sort both slices for comparison
	sort.Strings(words)
	sort.Strings(result)

	if !reflect.DeepEqual(words, result) {
		t.Errorf("GetAllWords returned incorrect result\nExpected: %v\nGot: %v", words, result)
	}

	// Test with empty trie
	emptyTrie := NewTrie()
	emptyResult := emptyTrie.GetAllWords()
	if len(emptyResult) != 0 {
		t.Errorf("GetAllWords for empty trie should return empty slice, got %v", emptyResult)
	}
}

// TestTrie_LargeDataset tests trie with a larger set of words
func TestTrie_LargeDataset(t *testing.T) {
	trie := NewTrie()
	words := []string{}

	// Generate 1000 words of various lengths
	for i := 0; i < 1000; i++ {
		word := ""
		length := (i % 10) + 1 // Words of length 1-10
		for j := 0; j < length; j++ {
			word += string(rune('a' + (i+j)%26))
		}
		words = append(words, word)
		trie.Insert(word)
	}

	// Verify all words are found
	for _, word := range words {
		if !trie.Search(word) {
			t.Errorf("Failed to find word '%s' in large dataset", word)
		}
	}
}

func TestTrie_WithModifiers(t *testing.T) {
	// Create a new trie with a bloom filter
	trie := NewTrie().WithBloomFilter(1000, 0.01)

	// Insert some words
	words := []string{"hello", "world", "hi", "hey"}
	for _, word := range words {
		trie.Insert(word)
	}

	// Test search functionality
	for _, word := range words {
		if !trie.Search(word) {
			t.Errorf("Expected word '%s' to be found", word)
		}
	}

	// Test non-existent word
	if trie.Search("notintrie") {
		t.Error("Expected word 'notintrie' to not be found")
	}

	// Test prefix functionality
	if !trie.StartsWith("he") {
		t.Error("Expected prefix 'he' to exist")
	}

	// Test delete functionality
	if !trie.Delete("hello") {
		t.Error("Expected to delete 'hello'")
	}

	if trie.Search("hello") {
		t.Error("Expected 'hello' to be deleted")
	}
}
