package algo

import (
	"fmt"
	"testing"
)

func TestConsistentHasher_MassiveCollision(t *testing.T) {
	type args struct {
		bucketCount int
		keyCount    int
	}
	tests := []struct {
		name         string
		args         args
		wantMaxRatio float64
		wantEmpty    bool
	}{
		{
			name:         "should distribute 1M keys over 1000 buckets",
			args:         args{bucketCount: 1000, keyCount: 1000000},
			wantMaxRatio: 2.0,
			wantEmpty:    false,
		},
		{
			name:         "should distribute 1M keys over 10k buckets",
			args:         args{bucketCount: 10000, keyCount: 1000000},
			wantMaxRatio: 2.0,
			wantEmpty:    false,
		},
		{
			name:         "should distribute 10K keys over 100 buckets",
			args:         args{bucketCount: 100, keyCount: 10000},
			wantMaxRatio: 2.0,
			wantEmpty:    false,
		},
		{
			name:         "should distribute 100 keys over 10 buckets",
			args:         args{bucketCount: 10, keyCount: 100},
			wantMaxRatio: 3.0,
			wantEmpty:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := ConsistentHasher{Buckets: tt.args.bucketCount}
			buckets := make([]int, tt.args.bucketCount)
			for i := 0; i < tt.args.keyCount; i++ {
				key := fmt.Sprintf("massive-key-%d", i)
				idx := h.GetBucket(key)
				if idx < 0 || idx >= tt.args.bucketCount {
					t.Fatalf("Bucket index out of range: got %d", idx)
				}
				buckets[idx]++
			}
			min, max, sum := buckets[0], buckets[0], 0
			emptyBuckets := 0
			for _, count := range buckets {
				if count < min {
					min = count
				}
				if count > max {
					max = count
				}
				if count == 0 {
					emptyBuckets++
				}
				sum += count
			}
			avg := float64(sum) / float64(tt.args.bucketCount)
			t.Logf("Buckets: %d, Keys: %d, Min: %d, Max: %d, Avg: %.2f, Empty: %d", tt.args.bucketCount, tt.args.keyCount, min, max, avg, emptyBuckets)
			if tt.wantEmpty {
				if emptyBuckets == 0 {
					t.Errorf("Expected some empty buckets, got none")
				}
			} else {
				if emptyBuckets > 0 {
					t.Errorf("%d buckets are empty, indicating poor distribution", emptyBuckets)
				}
			}
			if float64(max)/float64(min) > tt.wantMaxRatio {
				t.Errorf("Bucket distribution is unbalanced: max/min ratio %.2f (allowed %.2f)", float64(max)/float64(min), tt.wantMaxRatio)
			}
		})
	}
}

func TestConsistentHasher_New(t *testing.T) {
	h := ConsistentHasher{Buckets: 10}
	if h.Buckets != 10 {
		t.Errorf("Expected Buckets=10, got %d", h.Buckets)
	}
}

func TestConsistentHasher_Hash(t *testing.T) {
	h := ConsistentHasher{Buckets: 10}
	hash1 := h.Hash("key1")
	hash2 := h.Hash("key2")
	if hash1 == hash2 {
		t.Error("Hash values for different keys should differ")
	}
	// Determinism
	if h.Hash("key1") != hash1 {
		t.Error("Hash should be deterministic for the same key")
	}
}

func TestConsistentHasher_GetBucket(t *testing.T) {
	h := ConsistentHasher{Buckets: 5}
	buckets := make(map[int]bool)
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%d", i)
		idx := h.GetBucket(key)
		if idx < 0 || idx >= h.Buckets {
			t.Errorf("Bucket index out of range: got %d", idx)
		}
		buckets[idx] = true
	}
	if len(buckets) != h.Buckets {
		t.Errorf("Expected all buckets to be used, got %d", len(buckets))
	}
}

func TestConsistentHasher_GetBucket_ZeroBuckets(t *testing.T) {
	h := ConsistentHasher{Buckets: 0}
	idx := h.GetBucket("anykey")
	if idx != 0 {
		t.Errorf("Expected bucket 0 for zero buckets, got %d", idx)
	}
}

func TestConsistentHasher_GetBucket_NegativeBuckets(t *testing.T) {
	h := ConsistentHasher{Buckets: -3}
	idx := h.GetBucket("anykey")
	if idx != 0 {
		t.Errorf("Expected bucket 0 for negative buckets, got %d", idx)
	}
}

func TestConsistentHasher_Determinism(t *testing.T) {
	h := ConsistentHasher{Buckets: 7}
	key := "deterministic-key"
	idx1 := h.GetBucket(key)
	idx2 := h.GetBucket(key)
	if idx1 != idx2 {
		t.Error("GetBucket should be deterministic for the same key")
	}
}
