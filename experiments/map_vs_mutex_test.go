package experiments

import (
	"fmt"
	"sync"
	"testing"
)

// MutexMap uses a regular map protected by RWMutex
type MutexMap struct {
	data  map[string]interface{}
	mutex sync.RWMutex
}

// NewMutexMap creates a new mutex-protected map
func NewMutexMap() *MutexMap {
	return &MutexMap{
		data: make(map[string]interface{}),
	}
}

// Set stores a key-value pair
func (m *MutexMap) Set(key string, value interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.data[key] = value
}

// Get retrieves a value by key
func (m *MutexMap) Get(key string) (interface{}, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	value, exists := m.data[key]
	return value, exists
}

// Delete removes a key-value pair
func (m *MutexMap) Delete(key string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.data, key)
}

// Len returns the number of items
func (m *MutexMap) Len() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.data)
}

// SyncMapWrapper wraps sync.Map with consistent interface
type SyncMapWrapper struct {
	data sync.Map
}

// NewSyncMapWrapper creates a new sync.Map wrapper
func NewSyncMapWrapper() *SyncMapWrapper {
	return &SyncMapWrapper{}
}

// Set stores a key-value pair
func (s *SyncMapWrapper) Set(key string, value interface{}) {
	s.data.Store(key, value)
}

// Get retrieves a value by key
func (s *SyncMapWrapper) Get(key string) (interface{}, bool) {
	return s.data.Load(key)
}

// Delete removes a key-value pair
func (s *SyncMapWrapper) Delete(key string) {
	s.data.Delete(key)
}

// Len returns the number of items (approximate, as sync.Map doesn't have direct Len)
func (s *SyncMapWrapper) Len() int {
	count := 0
	s.data.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// Benchmark helper functions
func generateKeys(n int) []string {
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		keys[i] = fmt.Sprintf("key-%d", i)
	}
	return keys
}

// Benchmark Set operations
func BenchmarkMutexMap_Set(b *testing.B) {
	m := NewMutexMap()

	for b.Loop() {
		key := fmt.Sprintf("key-%d", b.N)
		m.Set(key, "value")
	}
}

func BenchmarkSyncMap_Set(b *testing.B) {
	s := NewSyncMapWrapper()

	for b.Loop() {
		key := fmt.Sprintf("key-%d", b.N)
		s.Set(key, "value")
	}
}

// Benchmark Get operations
func BenchmarkMutexMap_Get(b *testing.B) {
	m := NewMutexMap()
	// Pre-populate with data
	keys := generateKeys(1000)
	for _, key := range keys {
		m.Set(key, "value")
	}

	keyIndex := 0
	for b.Loop() {
		m.Get(keys[keyIndex%len(keys)])
		keyIndex++
	}
}

func BenchmarkSyncMap_Get(b *testing.B) {
	s := NewSyncMapWrapper()
	// Pre-populate with data
	keys := generateKeys(1000)
	for _, key := range keys {
		s.Set(key, "value")
	}

	keyIndex := 0
	for b.Loop() {
		s.Get(keys[keyIndex%len(keys)])
		keyIndex++
	}
}

// Benchmark mixed operations (Set/Get/Delete)
func BenchmarkMutexMap_Mixed(b *testing.B) {
	m := NewMutexMap()
	keys := generateKeys(100)

	// Pre-populate
	for _, key := range keys {
		m.Set(key, "initial-value")
	}

	keyIndex := 0
	for b.Loop() {
		key := keys[keyIndex%len(keys)]
		switch keyIndex % 4 {
		case 0, 1: // 50% Get operations
			m.Get(key)
		case 2: // 25% Set operations
			m.Set(key, "updated-value")
		case 3: // 25% Delete operations
			m.Delete(key)
			m.Set(key, "restored-value") // Restore for next iterations
		}
		keyIndex++
	}
}

func BenchmarkSyncMap_Mixed(b *testing.B) {
	s := NewSyncMapWrapper()
	keys := generateKeys(100)

	// Pre-populate
	for _, key := range keys {
		s.Set(key, "initial-value")
	}

	keyIndex := 0
	for b.Loop() {
		key := keys[keyIndex%len(keys)]
		switch keyIndex % 4 {
		case 0, 1: // 50% Get operations
			s.Get(key)
		case 2: // 25% Set operations
			s.Set(key, "updated-value")
		case 3: // 25% Delete operations
			s.Delete(key)
			s.Set(key, "restored-value") // Restore for next iterations
		}
		keyIndex++
	}
}

// Benchmark parallel operations
func BenchmarkMutexMap_ParallelReadWrite(b *testing.B) {
	m := NewMutexMap()
	keys := generateKeys(100)

	// Pre-populate
	for _, key := range keys {
		m.Set(key, "value")
	}

	b.RunParallel(func(pb *testing.PB) {
		keyIndex := 0
		for pb.Next() {
			key := keys[keyIndex%len(keys)]
			if keyIndex%10 == 0 { // 10% writes, 90% reads
				m.Set(key, "updated")
			} else {
				m.Get(key)
			}
			keyIndex++
		}
	})
}

func BenchmarkSyncMap_ParallelReadWrite(b *testing.B) {
	s := NewSyncMapWrapper()
	keys := generateKeys(100)

	// Pre-populate
	for _, key := range keys {
		s.Set(key, "value")
	}

	b.RunParallel(func(pb *testing.PB) {
		keyIndex := 0
		for pb.Next() {
			key := keys[keyIndex%len(keys)]
			if keyIndex%10 == 0 { // 10% writes, 90% reads
				s.Set(key, "updated")
			} else {
				s.Get(key)
			}
			keyIndex++
		}
	})
}

// Benchmark with different data sizes
func BenchmarkComparison_DifferentSizes(b *testing.B) {
	for _, size := range []int{10, 100, 1000, 10000} {
		b.Run(fmt.Sprintf("MutexMap-Size-%d", size), func(b *testing.B) {
			m := NewMutexMap()
			keys := generateKeys(size)

			// Pre-populate
			for _, key := range keys {
				m.Set(key, "value")
			}

			keyIndex := 0
			for b.Loop() {
				m.Get(keys[keyIndex%len(keys)])
				keyIndex++
			}
		})

		b.Run(fmt.Sprintf("SyncMap-Size-%d", size), func(b *testing.B) {
			s := NewSyncMapWrapper()
			keys := generateKeys(size)

			// Pre-populate
			for _, key := range keys {
				s.Set(key, "value")
			}

			keyIndex := 0
			for b.Loop() {
				s.Get(keys[keyIndex%len(keys)])
				keyIndex++
			}
		})
	}
}

// Memory allocation benchmarks
func BenchmarkMutexMap_SetMemory(b *testing.B) {
	b.ReportAllocs()
	m := NewMutexMap()

	for b.Loop() {
		key := fmt.Sprintf("key-%d", b.N)
		m.Set(key, "value")
	}
}

func BenchmarkSyncMap_SetMemory(b *testing.B) {
	b.ReportAllocs()
	s := NewSyncMapWrapper()

	for b.Loop() {
		key := fmt.Sprintf("key-%d", b.N)
		s.Set(key, "value")
	}
}

// Contention-heavy benchmarks
func BenchmarkMutexMap_HighContention(b *testing.B) {
	m := NewMutexMap()
	key := "contended-key"
	m.Set(key, "initial-value")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// All goroutines contend for the same key
			m.Get(key)
			m.Set(key, "updated-value")
		}
	})
}

func BenchmarkSyncMap_HighContention(b *testing.B) {
	s := NewSyncMapWrapper()
	key := "contended-key"
	s.Set(key, "initial-value")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// All goroutines contend for the same key
			s.Get(key)
			s.Set(key, "updated-value")
		}
	})
}
