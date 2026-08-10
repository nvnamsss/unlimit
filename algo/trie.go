package algo

import (
	boom "github.com/tylertreat/BoomFilters"
)

// TrieNode represents a node in the Trie data structure
type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool
}

// TrieModifier defines the interface for the chain of responsibility
type TrieModifier interface {
	Insert(word string, next func(word string))
	Search(word string, next func(word string) bool) bool
	StartsWith(prefix string, next func(prefix string) bool) bool
	Delete(word string, next func(word string) bool) bool
}

// Trie represents a trie data structure
type Trie struct {
	root     *TrieNode
	modifier TrieModifier
}

// BloomFilterModifier implements TrieModifier with a bloom filter
type BloomFilterModifier struct {
	bloomFilter *boom.StableBloomFilter
	next        TrieModifier
}

// WithBloomFilter adds a bloom filter to the Trie's chain of modifiers
func (t *Trie) WithBloomFilter(size int, falsePositiveRate float64) *Trie {
	bf := &BloomFilterModifier{
		bloomFilter: boom.NewDefaultStableBloomFilter(uint(size), falsePositiveRate),
	}

	if t.modifier == nil {
		t.modifier = bf
	} else {
		// Find the end of the chain
		current := t.modifier
		for {
			if bfMod, ok := current.(*BloomFilterModifier); ok {
				if bfMod.next == nil {
					bfMod.next = bf
					break
				}
				current = bfMod.next
			} else {
				break
			}
		}
	}

	return t
}

// Insert implements TrieModifier.Insert for BloomFilterModifier
func (bf *BloomFilterModifier) Insert(word string, next func(word string)) {
	bf.bloomFilter.Add([]byte(word))
	next(word)
}

// Search implements TrieModifier.Search for BloomFilterModifier
func (bf *BloomFilterModifier) Search(word string, next func(word string) bool) bool {
	if !bf.bloomFilter.Test([]byte(word)) {
		return false // Word is definitely not in the trie
	}
	return next(word)
}

// StartsWith implements TrieModifier.StartsWith for BloomFilterModifier
func (bf *BloomFilterModifier) StartsWith(prefix string, next func(prefix string) bool) bool {
	// Bloom filter can't efficiently check prefixes, so pass through
	return next(prefix)
}

// Delete implements TrieModifier.Delete for BloomFilterModifier
func (bf *BloomFilterModifier) Delete(word string, next func(word string) bool) bool {
	if !bf.bloomFilter.Test([]byte(word)) {
		return false // Word is definitely not in the trie
	}
	return next(word)
}

// NewTrie creates a new Trie
func NewTrie() *Trie {
	return &Trie{
		root: &TrieNode{
			children: make(map[rune]*TrieNode),
			isEnd:    false,
		},
	}
}

// Insert adds a word to the trie
func (t *Trie) Insert(word string) {
	if t.modifier != nil {
		t.modifier.Insert(word, func(w string) {
			t.insertIntoTrie(w)
		})
	} else {
		t.insertIntoTrie(word)
	}
}

// insertIntoTrie is the internal implementation of Insert
func (t *Trie) insertIntoTrie(word string) {
	node := t.root
	for _, ch := range word {
		if _, exists := node.children[ch]; !exists {
			node.children[ch] = &TrieNode{
				children: make(map[rune]*TrieNode),
				isEnd:    false,
			}
		}
		node = node.children[ch]
	}
	node.isEnd = true
}

// Search checks if the word exists in the trie
func (t *Trie) Search(word string) bool {
	if t.modifier != nil {
		return t.modifier.Search(word, func(w string) bool {
			return t.searchInTrie(w)
		})
	}
	return t.searchInTrie(word)
}

// searchInTrie is the internal implementation of Search
func (t *Trie) searchInTrie(word string) bool {
	node := t.root
	for _, ch := range word {
		if _, exists := node.children[ch]; !exists {
			return false
		}
		node = node.children[ch]
	}
	return node.isEnd
}

// StartsWith checks if there is any word in the trie that starts with the given prefix
func (t *Trie) StartsWith(prefix string) bool {
	if t.modifier != nil {
		return t.modifier.StartsWith(prefix, func(p string) bool {
			return t.startsWithInTrie(p)
		})
	}
	return t.startsWithInTrie(prefix)
}

// startsWithInTrie is the internal implementation of StartsWith
func (t *Trie) startsWithInTrie(prefix string) bool {
	node := t.root
	for _, ch := range prefix {
		if _, exists := node.children[ch]; !exists {
			return false
		}
		node = node.children[ch]
	}
	return true
}

// Delete removes a word from the trie if it exists
func (t *Trie) Delete(word string) bool {
	if t.modifier != nil {
		return t.modifier.Delete(word, func(w string) bool {
			return deleteHelper(t.root, w, 0)
		})
	}
	return deleteHelper(t.root, word, 0)
}

// deleteHelper is an iterative helper function for Delete
func deleteHelper(node *TrieNode, word string, index int) bool {
	cur := node
	runes := []rune(word)
	for i := index; i < len(runes); i++ {
		ch := runes[i]
		if _, exists := cur.children[ch]; !exists {
			return false // Word not found
		}
		cur = cur.children[ch]
	}

	if cur == nil {
		return false
	}

	if !cur.isEnd {
		return false
	}

	cur.isEnd = false
	return true
}

// GetAllWords returns all words in the trie
func (t *Trie) GetAllWords() []string {
	result := []string{}
	getAllWordsHelper(t.root, "", &result)
	return result
}

// getAllWordsHelper is a recursive helper function for GetAllWords
func getAllWordsHelper(node *TrieNode, prefix string, result *[]string) {
	if node.isEnd {
		*result = append(*result, prefix)
	}

	for ch, childNode := range node.children {
		getAllWordsHelper(childNode, prefix+string(ch), result)
	}
}
