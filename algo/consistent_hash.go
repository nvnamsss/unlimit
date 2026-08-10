package algo

type Hasher interface {
	Hash(key string) uint32
}

// ConsistentHasher implements Hasher using FNV-1a algorithm and supports bucket mapping.
type ConsistentHasher struct {
	Buckets int // Number of buckets for consistent hashing
}

// Hash returns a uint32 hash value for the given key using FNV-1a.
func (ConsistentHasher) Hash(key string) uint32 {
	const (
		offset32 = 2166136261
		prime32  = 16777619
	)
	hash := uint32(offset32)
	for i := 0; i < len(key); i++ {
		hash ^= uint32(key[i])
		hash *= prime32
	}
	return hash
}

// GetBucket returns the bucket index for the given key.
// The index is in the range [0, Buckets-1].
func (h ConsistentHasher) GetBucket(key string) int {
	if h.Buckets <= 0 {
		return 0 // fallback to 0 if not configured
	}
	hash := h.Hash(key)
	return int(hash % uint32(h.Buckets))
}
