package algo

import (
	"sort"
)

// RoaringBitmap represents a compressed bitmap using the Roaring bitmap format
type RoaringBitmap struct {
	containers map[uint16]container
}

// container interface for different container types
type container interface {
	contains(x uint16) bool
	add(x uint16) container
	remove(x uint16) container
	cardinality() int
	union(other container) container
	intersection(other container) container
	difference(other container) container
	toArray() []uint16
	clone() container
}

// arrayContainer stores values in a sorted array (used for sparse data)
type arrayContainer struct {
	values []uint16
}

// bitmapContainer stores values in a bitmap (used for dense data)
type bitmapContainer struct {
	bitmap [8192]uint64 // 2^16 bits / 64 = 1024 uint64s, but we use 8192 for safety
}

const (
	// Threshold for switching between array and bitmap containers
	arrayToBitmapThreshold = 4096
	bitmapToArrayThreshold = 4096
)

// NewRoaringBitmap creates a new empty roaring bitmap
func NewRoaringBitmap() *RoaringBitmap {
	return &RoaringBitmap{
		containers: make(map[uint16]container),
	}
}

// Add adds a value to the bitmap
func (rb *RoaringBitmap) Add(x uint32) {
	high := uint16(x >> 16)
	low := uint16(x & 0xFFFF)

	if container, exists := rb.containers[high]; exists {
		rb.containers[high] = container.add(low)
	} else {
		// Create new array container with the value
		newContainer := &arrayContainer{values: []uint16{low}}
		rb.containers[high] = newContainer
	}
}

// Contains checks if a value exists in the bitmap
func (rb *RoaringBitmap) Contains(x uint32) bool {
	high := uint16(x >> 16)
	low := uint16(x & 0xFFFF)

	if container, exists := rb.containers[high]; exists {
		return container.contains(low)
	}
	return false
}

// Remove removes a value from the bitmap
func (rb *RoaringBitmap) Remove(x uint32) {
	high := uint16(x >> 16)
	low := uint16(x & 0xFFFF)

	if container, exists := rb.containers[high]; exists {
		newContainer := container.remove(low)
		if newContainer.cardinality() == 0 {
			delete(rb.containers, high)
		} else {
			rb.containers[high] = newContainer
		}
	}
}

// Cardinality returns the number of elements in the bitmap
func (rb *RoaringBitmap) Cardinality() int {
	total := 0
	for _, container := range rb.containers {
		total += container.cardinality()
	}
	return total
}

// Union returns a new bitmap that is the union of this bitmap and another
func (rb *RoaringBitmap) Union(other *RoaringBitmap) *RoaringBitmap {
	result := NewRoaringBitmap()

	// Copy all containers from this bitmap
	for key, container := range rb.containers {
		result.containers[key] = container.clone()
	}

	// Union with containers from other bitmap
	for key, otherContainer := range other.containers {
		if thisContainer, exists := result.containers[key]; exists {
			result.containers[key] = thisContainer.union(otherContainer)
		} else {
			result.containers[key] = otherContainer.clone()
		}
	}

	return result
}

// Intersection returns a new bitmap that is the intersection of this bitmap and another
func (rb *RoaringBitmap) Intersection(other *RoaringBitmap) *RoaringBitmap {
	result := NewRoaringBitmap()

	for key, thisContainer := range rb.containers {
		if otherContainer, exists := other.containers[key]; exists {
			intersectionContainer := thisContainer.intersection(otherContainer)
			if intersectionContainer.cardinality() > 0 {
				result.containers[key] = intersectionContainer
			}
		}
	}

	return result
}

// Difference returns a new bitmap that contains elements in this bitmap but not in other
func (rb *RoaringBitmap) Difference(other *RoaringBitmap) *RoaringBitmap {
	result := NewRoaringBitmap()

	for key, thisContainer := range rb.containers {
		if otherContainer, exists := other.containers[key]; exists {
			diffContainer := thisContainer.difference(otherContainer)
			if diffContainer.cardinality() > 0 {
				result.containers[key] = diffContainer
			}
		} else {
			result.containers[key] = thisContainer.clone()
		}
	}

	return result
}

// ToArray returns all values in the bitmap as a sorted slice
func (rb *RoaringBitmap) ToArray() []uint32 {
	var result []uint32

	// Get sorted keys
	keys := make([]uint16, 0, len(rb.containers))
	for key := range rb.containers {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	// Collect values from each container
	for _, key := range keys {
		container := rb.containers[key]
		values := container.toArray()
		for _, val := range values {
			result = append(result, (uint32(key)<<16)|uint32(val))
		}
	}

	return result
}

// arrayContainer implementation

func (ac *arrayContainer) contains(x uint16) bool {
	// Binary search
	left, right := 0, len(ac.values)-1
	for left <= right {
		mid := (left + right) / 2
		if ac.values[mid] == x {
			return true
		} else if ac.values[mid] < x {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return false
}

func (ac *arrayContainer) add(x uint16) container {
	// Check if already exists
	if ac.contains(x) {
		return ac
	}

	// Find insertion point
	insertPos := 0
	for insertPos < len(ac.values) && ac.values[insertPos] < x {
		insertPos++
	}

	// Create new array with the value inserted
	newValues := make([]uint16, len(ac.values)+1)
	copy(newValues[:insertPos], ac.values[:insertPos])
	newValues[insertPos] = x
	copy(newValues[insertPos+1:], ac.values[insertPos:])

	newContainer := &arrayContainer{values: newValues}

	// Convert to bitmap container if threshold exceeded
	if len(newValues) > arrayToBitmapThreshold {
		return newContainer.toBitmapContainer()
	}

	return newContainer
}

func (ac *arrayContainer) remove(x uint16) container {
	// Find the value
	removePos := -1
	for i, val := range ac.values {
		if val == x {
			removePos = i
			break
		}
	}

	// Value not found
	if removePos == -1 {
		return ac
	}

	// Create new array without the value
	newValues := make([]uint16, len(ac.values)-1)
	copy(newValues[:removePos], ac.values[:removePos])
	copy(newValues[removePos:], ac.values[removePos+1:])

	return &arrayContainer{values: newValues}
}

func (ac *arrayContainer) cardinality() int {
	return len(ac.values)
}

func (ac *arrayContainer) union(other container) container {
	switch otherContainer := other.(type) {
	case *arrayContainer:
		return ac.unionArray(otherContainer)
	case *bitmapContainer:
		return otherContainer.union(ac)
	default:
		return ac
	}
}

func (ac *arrayContainer) unionArray(other *arrayContainer) container {
	// Merge two sorted arrays
	result := make([]uint16, 0, len(ac.values)+len(other.values))
	i, j := 0, 0

	for i < len(ac.values) && j < len(other.values) {
		if ac.values[i] < other.values[j] {
			result = append(result, ac.values[i])
			i++
		} else if ac.values[i] > other.values[j] {
			result = append(result, other.values[j])
			j++
		} else {
			result = append(result, ac.values[i])
			i++
			j++
		}
	}

	// Add remaining elements
	for i < len(ac.values) {
		result = append(result, ac.values[i])
		i++
	}
	for j < len(other.values) {
		result = append(result, other.values[j])
		j++
	}

	newContainer := &arrayContainer{values: result}
	if len(result) > arrayToBitmapThreshold {
		return newContainer.toBitmapContainer()
	}
	return newContainer
}

func (ac *arrayContainer) intersection(other container) container {
	switch otherContainer := other.(type) {
	case *arrayContainer:
		return ac.intersectionArray(otherContainer)
	case *bitmapContainer:
		return otherContainer.intersection(ac)
	default:
		return &arrayContainer{values: []uint16{}}
	}
}

func (ac *arrayContainer) intersectionArray(other *arrayContainer) container {
	result := make([]uint16, 0)
	i, j := 0, 0

	for i < len(ac.values) && j < len(other.values) {
		if ac.values[i] < other.values[j] {
			i++
		} else if ac.values[i] > other.values[j] {
			j++
		} else {
			result = append(result, ac.values[i])
			i++
			j++
		}
	}

	return &arrayContainer{values: result}
}

func (ac *arrayContainer) difference(other container) container {
	switch otherContainer := other.(type) {
	case *arrayContainer:
		return ac.differenceArray(otherContainer)
	case *bitmapContainer:
		result := make([]uint16, 0)
		for _, val := range ac.values {
			if !otherContainer.contains(val) {
				result = append(result, val)
			}
		}
		return &arrayContainer{values: result}
	default:
		return ac.clone()
	}
}

func (ac *arrayContainer) differenceArray(other *arrayContainer) container {
	result := make([]uint16, 0)
	i, j := 0, 0

	for i < len(ac.values) && j < len(other.values) {
		if ac.values[i] < other.values[j] {
			result = append(result, ac.values[i])
			i++
		} else if ac.values[i] > other.values[j] {
			j++
		} else {
			i++
			j++
		}
	}

	// Add remaining elements from this array
	for i < len(ac.values) {
		result = append(result, ac.values[i])
		i++
	}

	return &arrayContainer{values: result}
}

func (ac *arrayContainer) toArray() []uint16 {
	result := make([]uint16, len(ac.values))
	copy(result, ac.values)
	return result
}

func (ac *arrayContainer) clone() container {
	newValues := make([]uint16, len(ac.values))
	copy(newValues, ac.values)
	return &arrayContainer{values: newValues}
}

func (ac *arrayContainer) toBitmapContainer() *bitmapContainer {
	bc := &bitmapContainer{}
	for _, val := range ac.values {
		bc.setBit(val)
	}
	return bc
}

// bitmapContainer implementation

func (bc *bitmapContainer) contains(x uint16) bool {
	wordIndex := x / 64
	bitIndex := x % 64
	return (bc.bitmap[wordIndex] & (1 << bitIndex)) != 0
}

func (bc *bitmapContainer) add(x uint16) container {
	newContainer := bc.clone().(*bitmapContainer)
	newContainer.setBit(x)
	return newContainer
}

func (bc *bitmapContainer) remove(x uint16) container {
	newContainer := bc.clone().(*bitmapContainer)
	newContainer.clearBit(x)

	// Convert to array container if threshold reached
	if newContainer.cardinality() < bitmapToArrayThreshold {
		return newContainer.toArrayContainer()
	}

	return newContainer
}

func (bc *bitmapContainer) setBit(x uint16) {
	wordIndex := x / 64
	bitIndex := x % 64
	bc.bitmap[wordIndex] |= (1 << bitIndex)
}

func (bc *bitmapContainer) clearBit(x uint16) {
	wordIndex := x / 64
	bitIndex := x % 64
	bc.bitmap[wordIndex] &^= (1 << bitIndex)
}

func (bc *bitmapContainer) cardinality() int {
	count := 0
	for _, word := range bc.bitmap {
		count += popcount(word)
	}
	return count
}

func (bc *bitmapContainer) union(other container) container {
	switch otherContainer := other.(type) {
	case *bitmapContainer:
		result := &bitmapContainer{}
		for i := 0; i < len(bc.bitmap); i++ {
			result.bitmap[i] = bc.bitmap[i] | otherContainer.bitmap[i]
		}
		return result
	case *arrayContainer:
		result := bc.clone().(*bitmapContainer)
		for _, val := range otherContainer.values {
			result.setBit(val)
		}
		return result
	default:
		return bc.clone()
	}
}

func (bc *bitmapContainer) intersection(other container) container {
	switch otherContainer := other.(type) {
	case *bitmapContainer:
		result := &bitmapContainer{}
		for i := 0; i < len(bc.bitmap); i++ {
			result.bitmap[i] = bc.bitmap[i] & otherContainer.bitmap[i]
		}
		// Convert to array if sparse
		if result.cardinality() < bitmapToArrayThreshold {
			return result.toArrayContainer()
		}
		return result
	case *arrayContainer:
		result := make([]uint16, 0)
		for _, val := range otherContainer.values {
			if bc.contains(val) {
				result = append(result, val)
			}
		}
		return &arrayContainer{values: result}
	default:
		return &arrayContainer{values: []uint16{}}
	}
}

func (bc *bitmapContainer) difference(other container) container {
	switch otherContainer := other.(type) {
	case *bitmapContainer:
		result := &bitmapContainer{}
		for i := 0; i < len(bc.bitmap); i++ {
			result.bitmap[i] = bc.bitmap[i] &^ otherContainer.bitmap[i]
		}
		// Convert to array if sparse
		if result.cardinality() < bitmapToArrayThreshold {
			return result.toArrayContainer()
		}
		return result
	case *arrayContainer:
		result := bc.clone().(*bitmapContainer)
		for _, val := range otherContainer.values {
			result.clearBit(val)
		}
		if result.cardinality() < bitmapToArrayThreshold {
			return result.toArrayContainer()
		}
		return result
	default:
		return bc.clone()
	}
}

func (bc *bitmapContainer) toArray() []uint16 {
	result := make([]uint16, 0, bc.cardinality())
	for i, word := range bc.bitmap {
		if word == 0 {
			continue
		}
		baseIndex := uint16(i * 64)
		for j := 0; j < 64; j++ {
			if (word & (1 << j)) != 0 {
				result = append(result, baseIndex+uint16(j))
			}
		}
	}
	return result
}

func (bc *bitmapContainer) clone() container {
	newContainer := &bitmapContainer{}
	copy(newContainer.bitmap[:], bc.bitmap[:])
	return newContainer
}

func (bc *bitmapContainer) toArrayContainer() *arrayContainer {
	return &arrayContainer{values: bc.toArray()}
}

// popcount counts the number of set bits in a uint64
func popcount(x uint64) int {
	count := 0
	for x != 0 {
		count++
		x &= x - 1 // Clear the lowest set bit
	}
	return count
}
