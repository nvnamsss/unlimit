package algo

import (
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoaringBitmap_New(t *testing.T) {
	tests := []struct {
		name string
		want func(*RoaringBitmap) bool
	}{
		{
			name: "should create empty bitmap with initialized containers",
			want: func(rb *RoaringBitmap) bool {
				return rb != nil &&
					rb.containers != nil &&
					len(rb.containers) == 0 &&
					rb.Cardinality() == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRoaringBitmap()
			assert.True(t, tt.want(got))
		})
	}
}

func TestRoaringBitmap_Add(t *testing.T) {
	type args struct {
		values []uint32
	}

	tests := []struct {
		name               string
		args               args
		wantCardinality    int
		wantContains       []uint32
		wantNotContains    []uint32
		wantContainerCount int
	}{
		{
			name: "should add single value successfully",
			args: args{
				values: []uint32{100},
			},
			wantCardinality:    1,
			wantContains:       []uint32{100},
			wantNotContains:    []uint32{99, 101},
			wantContainerCount: 1,
		},
		{
			name: "should add multiple values in same container",
			args: args{
				values: []uint32{100, 200, 300},
			},
			wantCardinality:    3,
			wantContains:       []uint32{100, 200, 300},
			wantNotContains:    []uint32{99, 101, 199, 301},
			wantContainerCount: 1,
		},
		{
			name: "should add values across different containers",
			args: args{
				values: []uint32{100, 70000, 130000},
			},
			wantCardinality:    3,
			wantContains:       []uint32{100, 70000, 130000},
			wantNotContains:    []uint32{99, 69999, 129999},
			wantContainerCount: 2,
		},
		{
			name: "should add values across different containers v2",
			args: args{
				values: []uint32{100, 70000, 1e9},
			},
			wantCardinality:    3,
			wantContains:       []uint32{100, 70000, 1e9},
			wantNotContains:    []uint32{99, 69999, 129999},
			wantContainerCount: 3,
		},
		{
			name: "should add values across different containers v3",
			args: args{
				values: []uint32{100, 65537, 131073},
			},
			wantCardinality:    3,
			wantContains:       []uint32{100, 65537, 131073},
			wantNotContains:    []uint32{99, 69999, 129999},
			wantContainerCount: 3,
		},
		{
			name: "should handle duplicate values without increasing cardinality",
			args: args{
				values: []uint32{100, 100, 200, 200, 100},
			},
			wantCardinality:    2,
			wantContains:       []uint32{100, 200},
			wantNotContains:    []uint32{99, 101, 199, 201},
			wantContainerCount: 1,
		},
		{
			name: "should handle empty input",
			args: args{
				values: []uint32{},
			},
			wantCardinality:    0,
			wantContains:       []uint32{},
			wantNotContains:    []uint32{1, 100, 1000},
			wantContainerCount: 0,
		},
		{
			name: "should handle maximum uint32 value",
			args: args{
				values: []uint32{0, 0xFFFFFFFF},
			},
			wantCardinality:    2,
			wantContains:       []uint32{0, 0xFFFFFFFF},
			wantNotContains:    []uint32{1, 0xFFFFFFFE},
			wantContainerCount: 2,
		},
		{
			name: "store a big value with one container",
			args: args{
				values: []uint32{0xFFFFFFFF},
			},
			wantCardinality:    1,
			wantContains:       []uint32{0xFFFFFFFF},
			wantNotContains:    []uint32{0xFFFFFFFE},
			wantContainerCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := NewRoaringBitmap()

			// Add all values
			for _, value := range tt.args.values {
				rb.Add(value)
			}

			// Verify cardinality
			assert.Equal(t, tt.wantCardinality, rb.Cardinality())

			// Verify container count
			assert.Equal(t, tt.wantContainerCount, len(rb.containers))

			// Verify contains
			for _, value := range tt.wantContains {
				assert.True(t, rb.Contains(value), "Should contain %d", value)
			}

			// Verify not contains
			for _, value := range tt.wantNotContains {
				assert.False(t, rb.Contains(value), "Should not contain %d", value)
			}
		})
	}
}

func TestRoaringBitmap_Remove(t *testing.T) {
	type args struct {
		initialValues []uint32
		removeValues  []uint32
	}

	tests := []struct {
		name               string
		args               args
		wantCardinality    int
		wantContains       []uint32
		wantNotContains    []uint32
		wantContainerCount int
	}{
		{
			name: "should remove existing value successfully",
			args: args{
				initialValues: []uint32{100, 200, 300},
				removeValues:  []uint32{200},
			},
			wantCardinality:    2,
			wantContains:       []uint32{100, 300},
			wantNotContains:    []uint32{200},
			wantContainerCount: 1,
		},
		{
			name: "should handle removing non-existent value",
			args: args{
				initialValues: []uint32{100, 200},
				removeValues:  []uint32{300},
			},
			wantCardinality:    2,
			wantContains:       []uint32{100, 200},
			wantNotContains:    []uint32{300},
			wantContainerCount: 1,
		},
		{
			name: "should remove all values and cleanup container",
			args: args{
				initialValues: []uint32{100},
				removeValues:  []uint32{100},
			},
			wantCardinality:    0,
			wantContains:       []uint32{},
			wantNotContains:    []uint32{100},
			wantContainerCount: 0,
		},
		{
			name: "should handle removing from empty bitmap",
			args: args{
				initialValues: []uint32{},
				removeValues:  []uint32{100},
			},
			wantCardinality:    0,
			wantContains:       []uint32{},
			wantNotContains:    []uint32{100},
			wantContainerCount: 0,
		},
		{
			name: "should remove multiple values across containers",
			args: args{
				initialValues: []uint32{100, 70000, 130000},
				removeValues:  []uint32{100, 130000},
			},
			wantCardinality:    1,
			wantContains:       []uint32{70000},
			wantNotContains:    []uint32{100, 130000},
			wantContainerCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := NewRoaringBitmap()

			// Add initial values
			for _, value := range tt.args.initialValues {
				rb.Add(value)
			}

			// Remove values
			for _, value := range tt.args.removeValues {
				rb.Remove(value)
			}

			// Verify results
			assert.Equal(t, tt.wantCardinality, rb.Cardinality())
			assert.Equal(t, tt.wantContainerCount, len(rb.containers))

			for _, value := range tt.wantContains {
				assert.True(t, rb.Contains(value), "Should contain %d", value)
			}

			for _, value := range tt.wantNotContains {
				assert.False(t, rb.Contains(value), "Should not contain %d", value)
			}
		})
	}
}

func TestRoaringBitmap_Union(t *testing.T) {
	type args struct {
		bitmap1Values []uint32
		bitmap2Values []uint32
	}

	tests := []struct {
		name            string
		args            args
		wantCardinality int
		wantValues      []uint32
	}{
		{
			name: "should union two non-overlapping bitmaps",
			args: args{
				bitmap1Values: []uint32{100, 200},
				bitmap2Values: []uint32{300, 400},
			},
			wantCardinality: 4,
			wantValues:      []uint32{100, 200, 300, 400},
		},
		{
			name: "should union two overlapping bitmaps",
			args: args{
				bitmap1Values: []uint32{100, 200, 300},
				bitmap2Values: []uint32{200, 300, 400},
			},
			wantCardinality: 4,
			wantValues:      []uint32{100, 200, 300, 400},
		},
		{
			name: "should union with empty bitmap",
			args: args{
				bitmap1Values: []uint32{100, 200},
				bitmap2Values: []uint32{},
			},
			wantCardinality: 2,
			wantValues:      []uint32{100, 200},
		},
		{
			name: "should union two empty bitmaps",
			args: args{
				bitmap1Values: []uint32{},
				bitmap2Values: []uint32{},
			},
			wantCardinality: 0,
			wantValues:      []uint32{},
		},
		{
			name: "should union bitmaps across different containers",
			args: args{
				bitmap1Values: []uint32{100, 70000},
				bitmap2Values: []uint32{200, 130000},
			},
			wantCardinality: 4,
			wantValues:      []uint32{100, 200, 70000, 130000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb1 := NewRoaringBitmap()
			rb2 := NewRoaringBitmap()

			// Setup bitmaps
			for _, value := range tt.args.bitmap1Values {
				rb1.Add(value)
			}
			for _, value := range tt.args.bitmap2Values {
				rb2.Add(value)
			}

			// Perform union
			result := rb1.Union(rb2)

			// Verify results
			assert.Equal(t, tt.wantCardinality, result.Cardinality())

			resultArray := result.ToArray()
			sort.Slice(tt.wantValues, func(i, j int) bool { return tt.wantValues[i] < tt.wantValues[j] })
			assert.Equal(t, tt.wantValues, resultArray)

			// Verify all expected values are present
			for _, value := range tt.wantValues {
				assert.True(t, result.Contains(value), "Union should contain %d", value)
			}
		})
	}
}

func TestRoaringBitmap_Intersection(t *testing.T) {
	type args struct {
		bitmap1Values []uint32
		bitmap2Values []uint32
	}

	tests := []struct {
		name            string
		args            args
		wantCardinality int
		wantValues      []uint32
	}{
		{
			name: "should intersect overlapping bitmaps",
			args: args{
				bitmap1Values: []uint32{100, 200, 300},
				bitmap2Values: []uint32{200, 300, 400},
			},
			wantCardinality: 2,
			wantValues:      []uint32{200, 300},
		},
		{
			name: "should return empty for non-overlapping bitmaps",
			args: args{
				bitmap1Values: []uint32{100, 200},
				bitmap2Values: []uint32{300, 400},
			},
			wantCardinality: 0,
			wantValues:      []uint32{},
		},
		{
			name: "should return empty for intersection with empty bitmap",
			args: args{
				bitmap1Values: []uint32{100, 200},
				bitmap2Values: []uint32{},
			},
			wantCardinality: 0,
			wantValues:      []uint32{},
		},
		{
			name: "should intersect identical bitmaps",
			args: args{
				bitmap1Values: []uint32{100, 200, 300},
				bitmap2Values: []uint32{100, 200, 300},
			},
			wantCardinality: 3,
			wantValues:      []uint32{100, 200, 300},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb1 := NewRoaringBitmap()
			rb2 := NewRoaringBitmap()

			// Setup bitmaps
			for _, value := range tt.args.bitmap1Values {
				rb1.Add(value)
			}
			for _, value := range tt.args.bitmap2Values {
				rb2.Add(value)
			}

			// Perform intersection
			result := rb1.Intersection(rb2)

			// Verify results
			assert.Equal(t, tt.wantCardinality, result.Cardinality())

			resultArray := result.ToArray()
			sort.Slice(tt.wantValues, func(i, j int) bool { return tt.wantValues[i] < tt.wantValues[j] })
			assert.Equal(t, tt.wantValues, resultArray)
		})
	}
}

func TestRoaringBitmap_Difference(t *testing.T) {
	type args struct {
		bitmap1Values []uint32
		bitmap2Values []uint32
	}

	tests := []struct {
		name            string
		args            args
		wantCardinality int
		wantValues      []uint32
	}{
		{
			name: "should calculate difference with overlapping bitmaps",
			args: args{
				bitmap1Values: []uint32{100, 200, 300, 400},
				bitmap2Values: []uint32{200, 300},
			},
			wantCardinality: 2,
			wantValues:      []uint32{100, 400},
		},
		{
			name: "should return original bitmap for non-overlapping difference",
			args: args{
				bitmap1Values: []uint32{100, 200},
				bitmap2Values: []uint32{300, 400},
			},
			wantCardinality: 2,
			wantValues:      []uint32{100, 200},
		},
		{
			name: "should return original bitmap for difference with empty bitmap",
			args: args{
				bitmap1Values: []uint32{100, 200, 300},
				bitmap2Values: []uint32{},
			},
			wantCardinality: 3,
			wantValues:      []uint32{100, 200, 300},
		},
		{
			name: "should return empty for identical bitmaps",
			args: args{
				bitmap1Values: []uint32{100, 200, 300},
				bitmap2Values: []uint32{100, 200, 300},
			},
			wantCardinality: 0,
			wantValues:      []uint32{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb1 := NewRoaringBitmap()
			rb2 := NewRoaringBitmap()

			// Setup bitmaps
			for _, value := range tt.args.bitmap1Values {
				rb1.Add(value)
			}
			for _, value := range tt.args.bitmap2Values {
				rb2.Add(value)
			}

			// Perform difference
			result := rb1.Difference(rb2)

			// Verify results
			assert.Equal(t, tt.wantCardinality, result.Cardinality())

			resultArray := result.ToArray()
			sort.Slice(tt.wantValues, func(i, j int) bool { return tt.wantValues[i] < tt.wantValues[j] })
			assert.Equal(t, tt.wantValues, resultArray)
		})
	}
}

func TestRoaringBitmap_ToArray(t *testing.T) {
	type args struct {
		values []uint32
	}

	tests := []struct {
		name      string
		args      args
		wantArray []uint32
	}{
		{
			name: "should return empty array for empty bitmap",
			args: args{
				values: []uint32{},
			},
			wantArray: []uint32{},
		},
		{
			name: "should return sorted array for single container",
			args: args{
				values: []uint32{300, 100, 200},
			},
			wantArray: []uint32{100, 200, 300},
		},
		{
			name: "should return sorted array for multiple containers",
			args: args{
				values: []uint32{70000, 100, 130000, 200},
			},
			wantArray: []uint32{100, 200, 70000, 130000},
		},
		{
			name: "should handle single value",
			args: args{
				values: []uint32{42},
			},
			wantArray: []uint32{42},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := NewRoaringBitmap()

			// Add values
			for _, value := range tt.args.values {
				rb.Add(value)
			}

			// Get array
			result := rb.ToArray()

			// Verify result is sorted
			for i := 1; i < len(result); i++ {
				assert.True(t, result[i-1] < result[i], "Array should be sorted")
			}

			// Verify expected values
			assert.Equal(t, tt.wantArray, result)
		})
	}
}

func TestRoaringBitmap_ContainerConversion(t *testing.T) {
	tests := []struct {
		name                    string
		addCount                int
		removeCount             int
		wantCardinality         int
		shouldTriggerConversion bool
	}{
		{
			name:                    "should trigger array to bitmap conversion",
			addCount:                arrayToBitmapThreshold + 100,
			removeCount:             0,
			wantCardinality:         arrayToBitmapThreshold + 100,
			shouldTriggerConversion: true,
		},
		{
			name:                    "should trigger bitmap to array conversion after removal",
			addCount:                arrayToBitmapThreshold + 100,
			removeCount:             200,
			wantCardinality:         arrayToBitmapThreshold - 100,
			shouldTriggerConversion: true,
		},
		{
			name:                    "should not trigger conversion with small dataset",
			addCount:                100,
			removeCount:             0,
			wantCardinality:         100,
			shouldTriggerConversion: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := NewRoaringBitmap()

			// Add values
			for i := uint32(0); i < uint32(tt.addCount); i++ {
				rb.Add(i)
			}

			// Remove values from the end
			for i := uint32(tt.addCount - tt.removeCount); i < uint32(tt.addCount); i++ {
				rb.Remove(i)
			}

			// Verify cardinality
			assert.Equal(t, tt.wantCardinality, rb.Cardinality())

			// Verify all remaining values are present
			for i := uint32(0); i < uint32(tt.wantCardinality); i++ {
				assert.True(t, rb.Contains(i), "Should contain value %d after container conversion", i)
			}

			// Verify removed values are not present
			for i := uint32(tt.wantCardinality); i < uint32(tt.addCount); i++ {
				assert.False(t, rb.Contains(i), "Should not contain removed value %d", i)
			}
		})
	}
}

func TestRoaringBitmap_ConcurrentOperations(t *testing.T) {
	tests := []struct {
		name            string
		numGoroutines   int
		operationsPerGR int
		wantCardinality int
	}{
		{
			name:            "should handle concurrent additions",
			numGoroutines:   10,
			operationsPerGR: 100,
			wantCardinality: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := NewRoaringBitmap()
			var wg sync.WaitGroup
			wg.Add(tt.numGoroutines)

			// Concurrent additions
			for i := 0; i < tt.numGoroutines; i++ {
				go func(base int) {
					defer wg.Done()
					for j := 0; j < tt.operationsPerGR; j++ {
						rb.Add(uint32(base*tt.operationsPerGR + j))
					}
				}(i)
			}

			wg.Wait()

			// Verify results
			assert.Equal(t, tt.wantCardinality, rb.Cardinality())

			// Verify all values were added
			for i := 0; i < tt.numGoroutines; i++ {
				for j := 0; j < tt.operationsPerGR; j++ {
					value := uint32(i*tt.operationsPerGR + j)
					assert.True(t, rb.Contains(value), "Should contain concurrently added value %d", value)
				}
			}
		})
	}
}

func TestRoaringBitmap_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		values []uint32
		check  func(*testing.T, *RoaringBitmap)
	}{
		{
			name:   "should handle boundary values correctly",
			values: []uint32{0, 65535, 65536, 0xFFFFFFFF},
			check: func(t *testing.T, rb *RoaringBitmap) {
				assert.Equal(t, 4, rb.Cardinality())
				assert.True(t, rb.Contains(0))
				assert.True(t, rb.Contains(65535))
				assert.True(t, rb.Contains(65536))
				assert.True(t, rb.Contains(0xFFFFFFFF))
			},
		},
		{
			name:   "should handle operations on empty bitmap",
			values: []uint32{},
			check: func(t *testing.T, rb *RoaringBitmap) {
				// Test operations on empty bitmap
				assert.False(t, rb.Contains(100))
				rb.Remove(100) // Should not panic
				assert.Equal(t, 0, rb.Cardinality())

				// Test set operations with empty bitmap
				other := NewRoaringBitmap()
				other.Add(100)

				union := rb.Union(other)
				assert.Equal(t, 1, union.Cardinality())

				intersection := rb.Intersection(other)
				assert.Equal(t, 0, intersection.Cardinality())

				difference := rb.Difference(other)
				assert.Equal(t, 0, difference.Cardinality())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := NewRoaringBitmap()

			// Add values
			for _, value := range tt.values {
				rb.Add(value)
			}

			// Run custom checks
			tt.check(t, rb)
		})
	}
}

func TestRoaringBitmap_SetOperationsConsistency(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (*RoaringBitmap, *RoaringBitmap)
		validate func(*testing.T, *RoaringBitmap, *RoaringBitmap)
	}{
		{
			name: "should satisfy commutative property for union",
			setup: func() (*RoaringBitmap, *RoaringBitmap) {
				rb1 := NewRoaringBitmap()
				rb2 := NewRoaringBitmap()

				for i := uint32(0); i < 1000; i += 2 {
					rb1.Add(i)
				}
				for i := uint32(0); i < 1000; i += 3 {
					rb2.Add(i)
				}

				return rb1, rb2
			},
			validate: func(t *testing.T, rb1, rb2 *RoaringBitmap) {
				union1 := rb1.Union(rb2)
				union2 := rb2.Union(rb1)

				assert.Equal(t, union1.Cardinality(), union2.Cardinality())

				// Verify both unions contain the same elements
				array1 := union1.ToArray()
				array2 := union2.ToArray()
				assert.Equal(t, array1, array2)
			},
		},
		{
			name: "should satisfy commutative property for intersection",
			setup: func() (*RoaringBitmap, *RoaringBitmap) {
				rb1 := NewRoaringBitmap()
				rb2 := NewRoaringBitmap()

				for i := uint32(0); i < 1000; i += 2 {
					rb1.Add(i)
				}
				for i := uint32(0); i < 1000; i += 3 {
					rb2.Add(i)
				}

				return rb1, rb2
			},
			validate: func(t *testing.T, rb1, rb2 *RoaringBitmap) {
				intersection1 := rb1.Intersection(rb2)
				intersection2 := rb2.Intersection(rb1)

				assert.Equal(t, intersection1.Cardinality(), intersection2.Cardinality())

				// Verify specific values that should be in intersection (divisible by 6)
				for i := uint32(0); i < 1000; i += 6 {
					assert.True(t, intersection1.Contains(i), "Value %d should be in intersection", i)
					assert.True(t, intersection2.Contains(i), "Value %d should be in intersection", i)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb1, rb2 := tt.setup()
			tt.validate(t, rb1, rb2)
		})
	}
}

func TestRoaringBitmap_SimpleTest(t *testing.T) {
	rb := NewRoaringBitmap()
	rb.Add(1)
	rb.Add(1e5)

	assert.True(t, rb.Contains(1))
	assert.False(t, rb.Contains(2))
	assert.True(t, rb.Contains(1e5))
}
