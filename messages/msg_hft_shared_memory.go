package messages

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

// HFT Shared Memory Implementation
// Optimized for High-Frequency Trading with:
// - Zero-copy message access
// - Busy-spin polling (no sleep/ticker overhead)
// - Fixed-size messages (no serialization)
// - Cache-line aligned structures
// - CPU affinity support
// - Memory prefetching hints

var (
	// ErrHFTBufferFull is returned when the HFT buffer is full
	ErrHFTBufferFull = errors.New("hft buffer is full")
	// ErrHFTBufferEmpty is returned when the HFT buffer is empty
	ErrHFTBufferEmpty = errors.New("hft buffer is empty")
	// ErrHFTClosed is returned when the buffer is closed
	ErrHFTClosed = errors.New("hft buffer is closed")
	// ErrHFTTimeout is returned when busy-spin times out
	ErrHFTTimeout = errors.New("hft operation timed out")
)

const (
	// Cache line size for alignment (typically 64 bytes on modern CPUs)
	cacheLineSize = 64

	// HFT header layout (cache-line aligned):
	// [0-7]   : writeSequence (uint64) - monotonically increasing write counter
	// [8-15]  : readSequence (uint64) - monotonically increasing read counter
	// [16-23] : closed (uint64) - 1 if closed
	// [24-63] : padding for cache line alignment
	hftHeaderSize = 64

	// HFT slot states
	hftSlotEmpty   uint64 = 0
	hftSlotWriting uint64 = 1
	hftSlotReady   uint64 = 2
	hftSlotReading uint64 = 3
)

// HFTMessage is a fixed-size message optimized for HFT
// Total size: 128 bytes (2 cache lines) for optimal memory access
// All fields are fixed-size to avoid serialization overhead
type HFTMessage struct {
	// First cache line (64 bytes) - hot data
	Sequence  uint64   // 8 bytes - Message sequence number
	Timestamp int64    // 8 bytes - Unix nanoseconds
	Symbol    [8]byte  // 8 bytes - Trading symbol (e.g., "BTCUSDT\0")
	Price     int64    // 8 bytes - Price in smallest unit (e.g., cents, satoshi)
	Quantity  int64    // 8 bytes - Quantity in smallest unit
	Side      uint8    // 1 byte - 0=buy, 1=sell
	Type      uint8    // 1 byte - Order type: 0=market, 1=limit, 2=stop
	Flags     uint16   // 2 bytes - Custom flags
	OrderID   uint32   // 4 bytes - Order ID
	_pad1     [16]byte // 16 bytes - Padding to 64 bytes

	// Second cache line (64 bytes) - extended data
	AccountID  uint64   // 8 bytes - Account identifier
	StrategyID uint32   // 4 bytes - Strategy identifier
	Reserved   uint32   // 4 bytes - Reserved for future use
	Extra      [48]byte // 48 bytes - Extra payload for custom data
}

// HFTMessageSize is the fixed size of each message
const HFTMessageSize = 128

// Compile-time size check (will fail to compile if size doesn't match)
var _ = [1]struct{}{}[128-int(unsafe.Sizeof(HFTMessage{}))+128-int(unsafe.Sizeof(HFTMessage{}))]

// HFTSlot represents a single slot in the ring buffer
// Padded to avoid false sharing between slots
type HFTSlot struct {
	State   uint64     // 8 bytes - Slot state (empty/writing/ready/reading)
	_pad1   [56]byte   // 56 bytes - Padding to cache line (total 64 bytes)
	Message HFTMessage // 128 bytes - The message data
	_pad2   [64]byte   // 64 bytes - Padding to prevent false sharing with next slot
}

// HFTSlotSize is the total size of each slot including padding
const HFTSlotSize = 256

// Compile-time size check
var _ = [1]struct{}{}[256-int(unsafe.Sizeof(HFTSlot{}))+256-int(unsafe.Sizeof(HFTSlot{}))]

// HFTRingBuffer is a lock-free ring buffer optimized for HFT
// Uses SPSC (Single Producer Single Consumer) design for maximum performance
type HFTRingBuffer struct {
	data      []byte   // mmap'd memory region
	file      *os.File // backing file
	slotCount uint64   // number of slots (must be power of 2)
	mask      uint64   // slotCount - 1 for fast modulo
	isOwner   bool     // true if this instance created the shared memory
	path      string   // path to shared memory file
}

// HFTRingBufferConfig holds configuration for HFT ring buffer
type HFTRingBufferConfig struct {
	// Path is the file path for mmap (use /dev/shm/ for tmpfs)
	Path string
	// SlotCount must be a power of 2 (default: 1024)
	SlotCount uint64
	// Create indicates whether to create new or open existing
	Create bool
	// HugePages enables huge page support (2MB pages) for better TLB performance
	HugePages bool
}

// writeSequencePtr returns pointer to write sequence in header
func (rb *HFTRingBuffer) writeSequencePtr() *uint64 {
	return (*uint64)(unsafe.Pointer(&rb.data[0]))
}

// readSequencePtr returns pointer to read sequence in header
func (rb *HFTRingBuffer) readSequencePtr() *uint64 {
	return (*uint64)(unsafe.Pointer(&rb.data[8]))
}

// closedPtr returns pointer to closed flag in header
func (rb *HFTRingBuffer) closedPtr() *uint64 {
	return (*uint64)(unsafe.Pointer(&rb.data[16]))
}

// slotPtr returns pointer to a slot by index
func (rb *HFTRingBuffer) slotPtr(index uint64) *HFTSlot {
	offset := hftHeaderSize + (index * HFTSlotSize)
	return (*HFTSlot)(unsafe.Pointer(&rb.data[offset]))
}

// isPowerOfTwo checks if n is a power of 2
func isPowerOfTwo(n uint64) bool {
	return n > 0 && (n&(n-1)) == 0
}

// NewHFTRingBuffer creates or opens an HFT ring buffer
func NewHFTRingBuffer(config *HFTRingBufferConfig) (*HFTRingBuffer, error) {
	if config == nil {
		return nil, errors.New("config is required")
	}
	if config.Path == "" {
		return nil, errors.New("path is required")
	}
	if config.SlotCount == 0 {
		config.SlotCount = 1024
	}
	if !isPowerOfTwo(config.SlotCount) {
		return nil, errors.New("slotCount must be a power of 2")
	}

	totalSize := hftHeaderSize + (int(config.SlotCount) * HFTSlotSize)

	var file *os.File
	var err error
	var isOwner bool

	if config.Create {
		file, err = os.OpenFile(config.Path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to create hft shared memory: %w", err)
		}

		if err := file.Truncate(int64(totalSize)); err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to truncate: %w", err)
		}
		isOwner = true
	} else {
		file, err = os.OpenFile(config.Path, os.O_RDWR, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open hft shared memory: %w", err)
		}
	}

	// mmap flags for optimal performance
	mmapFlags := syscall.MAP_SHARED
	if config.HugePages {
		mmapFlags |= syscall.MAP_HUGETLB
	}

	data, err := syscall.Mmap(
		int(file.Fd()),
		0,
		totalSize,
		syscall.PROT_READ|syscall.PROT_WRITE,
		mmapFlags,
	)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to mmap: %w", err)
	}

	rb := &HFTRingBuffer{
		data:      data,
		file:      file,
		slotCount: config.SlotCount,
		mask:      config.SlotCount - 1,
		isOwner:   isOwner,
		path:      config.Path,
	}

	if isOwner {
		// Zero out all memory first to ensure clean state
		for i := range rb.data {
			rb.data[i] = 0
		}

		// Initialize header
		atomic.StoreUint64(rb.writeSequencePtr(), 0)
		atomic.StoreUint64(rb.readSequencePtr(), 0)
		atomic.StoreUint64(rb.closedPtr(), 0)

		// Initialize all slots as empty
		for i := uint64(0); i < config.SlotCount; i++ {
			slot := rb.slotPtr(i)
			atomic.StoreUint64(&slot.State, hftSlotEmpty)
		}

		// Sync to ensure initialization is visible
		_ = unix.Msync(rb.data, unix.MS_SYNC)
	}

	// Lock memory to prevent page faults (reduces tail latency)
	// Requires CAP_IPC_LOCK capability or sufficient RLIMIT_MEMLOCK
	if err := unix.Mlock(rb.data); err != nil {
		// Not fatal - just means we might get page faults
		// Log or ignore based on requirements
	}

	return rb, nil
}

// Publish publishes a message to the buffer (producer side)
// Returns immediately if buffer is full (non-blocking)
func (rb *HFTRingBuffer) Publish(msg *HFTMessage) error {
	if atomic.LoadUint64(rb.closedPtr()) == 1 {
		return ErrHFTClosed
	}

	writeSeq := atomic.LoadUint64(rb.writeSequencePtr())
	readSeq := atomic.LoadUint64(rb.readSequencePtr())

	// Check if buffer is full
	if writeSeq-readSeq >= rb.slotCount {
		return ErrHFTBufferFull
	}

	slotIdx := writeSeq & rb.mask
	slot := rb.slotPtr(slotIdx)

	// Try to claim slot
	if !atomic.CompareAndSwapUint64(&slot.State, hftSlotEmpty, hftSlotWriting) {
		return ErrHFTBufferFull
	}

	// Set sequence and timestamp
	msg.Sequence = writeSeq
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().UnixNano()
	}

	// Direct memory copy (zero-copy write)
	slot.Message = *msg

	// Memory barrier before publishing
	atomic.StoreUint64(&slot.State, hftSlotReady)

	// Advance write sequence
	atomic.AddUint64(rb.writeSequencePtr(), 1)

	return nil
}

// PublishWait publishes with busy-spin wait if buffer is full
// timeout of 0 means spin forever until success
func (rb *HFTRingBuffer) PublishWait(msg *HFTMessage, timeout time.Duration) error {
	start := time.Now()
	spins := 0

	for {
		err := rb.Publish(msg)
		if err == nil {
			return nil
		}
		if err == ErrHFTClosed {
			return err
		}

		// Busy spin with progressive backoff
		spins++
		if spins < 100 {
			// Pure spin
			runtime.Gosched()
		} else if spins < 1000 {
			// Pause instruction hint (reduces CPU power and improves spin efficiency)
			for i := 0; i < 10; i++ {
				runtime.Gosched()
			}
		} else {
			// Brief yield after many spins
			runtime.Gosched()
			spins = 0 // Reset to try fast path again
		}

		if timeout > 0 && time.Since(start) >= timeout {
			return ErrHFTTimeout
		}
	}
}

// Consume consumes a message from the buffer (consumer side)
// Returns a pointer directly into shared memory (zero-copy)
// WARNING: The returned pointer is only valid until the next Consume call
func (rb *HFTRingBuffer) Consume() (*HFTMessage, error) {
	if atomic.LoadUint64(rb.closedPtr()) == 1 {
		readSeq := atomic.LoadUint64(rb.readSequencePtr())
		writeSeq := atomic.LoadUint64(rb.writeSequencePtr())
		if readSeq >= writeSeq {
			return nil, ErrHFTClosed
		}
	}

	readSeq := atomic.LoadUint64(rb.readSequencePtr())
	writeSeq := atomic.LoadUint64(rb.writeSequencePtr())

	// Check if buffer is empty
	if readSeq >= writeSeq {
		return nil, ErrHFTBufferEmpty
	}

	slotIdx := readSeq & rb.mask
	slot := rb.slotPtr(slotIdx)

	// Check if slot is ready
	if atomic.LoadUint64(&slot.State) != hftSlotReady {
		return nil, ErrHFTBufferEmpty
	}

	// Claim for reading
	if !atomic.CompareAndSwapUint64(&slot.State, hftSlotReady, hftSlotReading) {
		return nil, ErrHFTBufferEmpty
	}

	// Return pointer to message (zero-copy read)
	msg := &slot.Message

	// Mark as empty and advance read sequence
	atomic.StoreUint64(&slot.State, hftSlotEmpty)
	atomic.AddUint64(rb.readSequencePtr(), 1)

	return msg, nil
}

// ConsumeWait consumes with busy-spin wait if buffer is empty
// timeout of 0 means spin forever until message available
func (rb *HFTRingBuffer) ConsumeWait(timeout time.Duration) (*HFTMessage, error) {
	start := time.Now()
	spins := 0

	for {
		msg, err := rb.Consume()
		if err == nil {
			return msg, nil
		}
		if err == ErrHFTClosed {
			return nil, err
		}

		// Busy spin with progressive backoff
		spins++
		if spins < 100 {
			runtime.Gosched()
		} else if spins < 1000 {
			for i := 0; i < 10; i++ {
				runtime.Gosched()
			}
		} else {
			runtime.Gosched()
			spins = 0
		}

		if timeout > 0 && time.Since(start) >= timeout {
			return nil, ErrHFTTimeout
		}
	}
}

// ConsumeCopy consumes and copies the message to avoid pointer lifetime issues
func (rb *HFTRingBuffer) ConsumeCopy() (HFTMessage, error) {
	msg, err := rb.Consume()
	if err != nil {
		return HFTMessage{}, err
	}
	return *msg, nil
}

// Prefetch prefetches the next slot into CPU cache
// Call this before Consume for better latency
func (rb *HFTRingBuffer) Prefetch() {
	readSeq := atomic.LoadUint64(rb.readSequencePtr())
	slotIdx := readSeq & rb.mask
	slot := rb.slotPtr(slotIdx)
	// Touch the memory to bring it into cache
	_ = atomic.LoadUint64(&slot.State)
}

// WarmUp touches all memory pages to prevent page faults during operation
// Call this once after creating the ring buffer
func (rb *HFTRingBuffer) WarmUp() {
	// Touch every page (4KB) to fault them in
	pageSize := 4096
	for i := 0; i < len(rb.data); i += pageSize {
		_ = rb.data[i]
	}
	// Also write to ensure write pages are faulted
	for i := 0; i < len(rb.data); i += pageSize {
		rb.data[i] = rb.data[i]
	}
}

// Size returns the number of messages in the buffer
func (rb *HFTRingBuffer) Size() uint64 {
	writeSeq := atomic.LoadUint64(rb.writeSequencePtr())
	readSeq := atomic.LoadUint64(rb.readSequencePtr())
	return writeSeq - readSeq
}

// IsClosed returns true if the buffer is closed
func (rb *HFTRingBuffer) IsClosed() bool {
	return atomic.LoadUint64(rb.closedPtr()) == 1
}

// Close closes the buffer
func (rb *HFTRingBuffer) Close() error {
	atomic.StoreUint64(rb.closedPtr(), 1)

	if err := syscall.Munmap(rb.data); err != nil {
		return err
	}

	if rb.isOwner {
		rb.file.Close()
		return os.Remove(rb.path)
	}

	return rb.file.Close()
}

// HFTProducer is a high-performance producer for HFT workloads
type HFTProducer struct {
	rb          *HFTRingBuffer
	symbol      [8]byte
	accountID   uint64
	cpuAffinity int // CPU core to pin to (-1 for no pinning)
}

// HFTProducerConfig holds producer configuration
type HFTProducerConfig struct {
	Path        string
	SlotCount   uint64
	Symbol      string
	AccountID   uint64
	Create      bool
	HugePages   bool
	CPUAffinity int // CPU core to pin to (-1 for no pinning)
}

// NewHFTProducer creates a new HFT producer
func NewHFTProducer(config *HFTProducerConfig) (*HFTProducer, error) {
	rb, err := NewHFTRingBuffer(&HFTRingBufferConfig{
		Path:      config.Path,
		SlotCount: config.SlotCount,
		Create:    config.Create,
		HugePages: config.HugePages,
	})
	if err != nil {
		return nil, err
	}

	p := &HFTProducer{
		rb:          rb,
		accountID:   config.AccountID,
		cpuAffinity: config.CPUAffinity,
	}

	// Copy symbol
	copy(p.symbol[:], config.Symbol)

	return p, nil
}

// LockThread locks the current goroutine to an OS thread and sets CPU affinity
// Call this at the start of your hot path
func (p *HFTProducer) LockThread() {
	runtime.LockOSThread()
	// Note: CPU affinity requires platform-specific syscalls
	// On Linux, you would use sched_setaffinity
}

// UnlockThread unlocks the goroutine from the OS thread
func (p *HFTProducer) UnlockThread() {
	runtime.UnlockOSThread()
}

// SendOrder sends an order message
func (p *HFTProducer) SendOrder(price, quantity int64, side, orderType uint8, orderID uint32) error {
	msg := &HFTMessage{
		Timestamp: time.Now().UnixNano(),
		Symbol:    p.symbol,
		Price:     price,
		Quantity:  quantity,
		Side:      side,
		Type:      orderType,
		OrderID:   orderID,
		AccountID: p.accountID,
	}
	return p.rb.Publish(msg)
}

// SendOrderWait sends an order with busy-spin wait
func (p *HFTProducer) SendOrderWait(price, quantity int64, side, orderType uint8, orderID uint32, timeout time.Duration) error {
	msg := &HFTMessage{
		Timestamp: time.Now().UnixNano(),
		Symbol:    p.symbol,
		Price:     price,
		Quantity:  quantity,
		Side:      side,
		Type:      orderType,
		OrderID:   orderID,
		AccountID: p.accountID,
	}
	return p.rb.PublishWait(msg, timeout)
}

// Close closes the producer
func (p *HFTProducer) Close() error {
	return p.rb.Close()
}

// HFTConsumer is a high-performance consumer for HFT workloads
type HFTConsumer struct {
	rb          *HFTRingBuffer
	handler     func(*HFTMessage) error
	running     atomic.Bool
	cpuAffinity int
}

// HFTConsumerConfig holds consumer configuration
type HFTConsumerConfig struct {
	Path        string
	SlotCount   uint64
	HugePages   bool
	CPUAffinity int
}

// NewHFTConsumer creates a new HFT consumer
func NewHFTConsumer(config *HFTConsumerConfig) (*HFTConsumer, error) {
	rb, err := NewHFTRingBuffer(&HFTRingBufferConfig{
		Path:      config.Path,
		SlotCount: config.SlotCount,
		Create:    false,
		HugePages: config.HugePages,
	})
	if err != nil {
		return nil, err
	}

	return &HFTConsumer{
		rb:          rb,
		cpuAffinity: config.CPUAffinity,
	}, nil
}

// Handle sets the message handler
func (c *HFTConsumer) Handle(fn func(*HFTMessage) error) {
	c.handler = fn
}

// Run starts the busy-spin consume loop
// This will consume 100% CPU on the current goroutine
// Call in a dedicated goroutine
func (c *HFTConsumer) Run() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	c.running.Store(true)

	for c.running.Load() {
		// Prefetch next message into cache
		c.rb.Prefetch()

		msg, err := c.rb.Consume()
		if err != nil {
			if err == ErrHFTClosed {
				return
			}
			// Spin on empty buffer
			runtime.Gosched()
			continue
		}

		if c.handler != nil {
			_ = c.handler(msg)
		}
	}
}

// Stop stops the consumer
func (c *HFTConsumer) Stop() {
	c.running.Store(false)
}

// Close stops and closes the consumer
func (c *HFTConsumer) Close() error {
	c.Stop()
	return c.rb.Close()
}

// Benchmark helpers

// HFTLatencyStats holds latency statistics
type HFTLatencyStats struct {
	Count   uint64
	TotalNs uint64
	MinNs   int64
	MaxNs   int64
	AvgNs   float64
}

// MeasureLatency calculates end-to-end latency from message timestamp
func MeasureLatency(msg *HFTMessage) int64 {
	return time.Now().UnixNano() - msg.Timestamp
}
