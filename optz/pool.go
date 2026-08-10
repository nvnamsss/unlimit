package optz

import (
	"math"
	"math/rand"
	"sync/atomic"
)

type Pool[T any] interface {
	Get() T
	Add(item T) bool
	Put(T)
	Do(func(T) error) error
}

// this approach is not thread-safe, but it is fast
// it is used for testing purposes only
type randomBoundedPool[T any] struct {
	pools     []*boundedPool[T]
	capacity  int
	consensus int
	backup    func() T
}

func (bp *randomBoundedPool[T]) Get() T {
	var (
		index = rand.Int31() % int32(bp.consensus)
	)
	return bp.pools[index].Get()
}

func (bp *randomBoundedPool[T]) Put(item T) {
	var (
		index = rand.Int31() % int32(bp.consensus)
	)
	bp.pools[index].Put(item)
}

func (bp *randomBoundedPool[T]) Add(item T) bool {
	var (
		index = 0
	)

	for index < bp.consensus {
		if bp.pools[index].Add(item) {
			break
		}
		index++
	}

	return index < bp.consensus
}

func (bp *randomBoundedPool[T]) Do(f func(T) error) error {
	var (
		index = rand.Int31() % int32(bp.consensus)
	)
	v := bp.pools[index].Get()
	err := f(v)
	bp.pools[index].Put(v)

	return err
}

func newRandomBoundedPool[T any](consensus int, size int, backup func() T) *randomBoundedPool[T] {
	var (
		pool = &randomBoundedPool[T]{
			consensus: consensus,
			backup:    backup,
		}
		bps = size / consensus
	)

	for loop := 0; loop < consensus; loop++ {
		bp := newBoundedPool(bps, backup)
		pool.pools = append(pool.pools, bp)
	}

	return pool
}

/*
Theoretically, Pool be a free list of knowledge will be used later,
Pool freely to scale or shrink depend on the throughput
*/
type boundedPool[T any] struct {
	// Padding to prevent false sharing
	_padding1 [56]byte

	items        []T
	capacity     int
	capacityMask int // for fast modulo with bitwise AND
	backup       func() T

	// Separate cache lines for counters to prevent false sharing
	_padding2    [56]byte
	used         int64
	_padding3    [56]byte
	released     int64
	_padding4    [56]byte
	releasedSlow int64
}

func newBoundedPool[T any](capacity int, backup func() T) *boundedPool[T] {
	// Round up to power of 2 for fast modulo
	capacity = 1 << uint(math.Ceil(math.Log2(float64(capacity))))

	return &boundedPool[T]{
		items:        make([]T, capacity), // Pre-allocate full capacity
		capacity:     capacity,
		capacityMask: capacity - 1, // For fast modulo with bitwise AND
		backup:       backup,
	}
}

func (bp *boundedPool[T]) Get() T {
	u := atomic.AddInt64(&bp.used, 1) - 1
	r := atomic.LoadInt64(&bp.releasedSlow)

	if u-r >= int64(bp.capacity) {
		if bp.backup != nil {
			return bp.backup()
		}
		var zero T
		return zero
	}

	// Fast modulo using bitwise AND with power-of-2 mask
	index := u & int64(bp.capacityMask)
	item := bp.items[index]
	var zero T
	bp.items[index] = zero // Clear the slot
	return item
}

func (bp *boundedPool[T]) Put(item T) {
	r := atomic.AddInt64(&bp.released, 1) - 1
	// Fast modulo using bitwise AND with power-of-2 mask
	index := r & int64(bp.capacityMask)
	bp.items[index] = item
	atomic.AddInt64(&bp.releasedSlow, 1)
}

func (bp *boundedPool[T]) Add(item T) bool {
	currentLen := len(bp.items)
	if currentLen == bp.capacity {
		return false
	}
	// Direct assignment instead of append
	bp.items[currentLen] = item
	return true
}

func (bp *boundedPool[T]) Do(f func(T) error) error {
	i := bp.Get()
	err := f(i)
	bp.Put(i)
	return err
}

type indexer struct {
	c        []uint32
	used     int32
	capacity int32
	wait     chan bool
	level    int32
}

func newIndexer(capacity int32) *indexer {
	i := &indexer{
		capacity: capacity,
		wait:     make(chan bool),
		level:    int32(math.Max(1, math.Ceil(math.Log2(float64(capacity))/5))),
	}

	var size int
	size = int(math.Pow(32, float64(i.level-1))-1) / 31
	size += int(math.Ceil(float64(capacity) / 32))
	i.c = make([]uint32, size)
	return i
}

func (i *indexer) Use() int32 {
	var (
		set   bool
		index int32
		shift uint32   = 1
		ns    []uint32 = make([]uint32, i.level+1)
		fs    []uint32 = make([]uint32, i.level+1)
	)

	if atomic.AddInt32(&i.used, 1) > i.capacity {
		<-i.wait
	}

	for !set {
		index = 0
		for loop := int32(0); loop < i.level; loop++ {
			v := atomic.LoadUint32(&i.c[ns[loop]])
			fs[loop] = i.firstZeroBit(v)
			ns[loop+1] = ns[loop]*32 + fs[loop] + 1
			index += int32(fs[loop]) * int32(math.Pow(32, float64(i.level-loop-1)))
		}

		if index >= i.capacity {
			atomic.AddInt32(&i.used, 1)
			<-i.wait
			continue
		}

		level := i.level - 1
		lvup := true
		for level >= 0 && lvup {
			v := atomic.LoadUint32(&i.c[ns[level]])
			m := v | (shift << fs[level])

			if m == v {
				set = false
				break
			}

			set = atomic.CompareAndSwapUint32(&i.c[ns[level]], v, m)
			if set {
				lvup = m == math.MaxUint32
				level--
			}
		}
	}

	return index
}

func (i *indexer) Release(index int32) {
	var (
		set   bool
		ns    []uint32 = make([]uint32, i.level+1)
		ics   []uint32 = make([]uint32, i.level+1)
		shift uint32   = 1
		m              = uint32(index)
	)

	for loop := i.level - 1; loop >= 0; loop-- {
		s := uint32(math.Pow(32, float64(loop))-1) / 31
		ns[loop] = m/32 + s
		ics[loop] = m % 32
		m /= 32
	}

	for !set {
		level := i.level - 1
		lvdown := true
		for level >= 0 && lvdown {
			v := atomic.LoadUint32(&i.c[ns[level]])
			set = atomic.CompareAndSwapUint32(&i.c[ns[level]], v, v & ^(shift<<ics[level]))
			if set {
				lvdown = v == math.MaxUint32
				level--
			}
		}

	}

	if atomic.AddInt32(&i.used, -1) >= i.capacity {
		i.wait <- true
	}

}

func (i *indexer) firstZeroBit(v uint32) uint32 {
	nv := ^v
	return i.bitCount(nv&(-nv) - 1)
}

// MIT HAKMEM bit count
func (i *indexer) bitCount(v uint32) uint32 {
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
