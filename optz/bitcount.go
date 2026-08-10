package optz

// MIT HAKMEM bit count
// Time complexity: O(1)
// Space complexity: O(1)
func MITHakmemBitCount(v uint32) uint32 {
	var n uint32 = uint32(v)
	var x uint32 = n

	n = (x >> 1) & 0x77777777
	x = x - n
	n = (n >> 1) & 0x77777777
	x = x - n
	n = (n >> 1) & 0x77777777
	x = x - n
	x = (x + (x >> 4)) & 0x0F0F0F0F
	x = x * 0x01010101
	return uint32(x >> 24)
}

// Brian Kernighan's Algorithm
// Time complexity: O(k)
// Space complexity: O(1)
func HammingWeightBitcount(v uint32) uint32 {
	var n uint32 = uint32(v)
	var c uint32
	for c = 0; n > 0; c++ {
		n &= n - 1
	}
	return c
}

func NativeBitcount(v uint32) uint32 {
	var n uint32 = uint32(v)
	var c uint32
	for c = 0; n > 0; c++ {
		n &= n - 1
	}
	return c
}
