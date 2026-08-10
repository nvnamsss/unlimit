package messages

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

var (
	// ErrSharedMemoryFull is returned when the shared memory buffer is full
	ErrSharedMemoryFull = errors.New("shared memory buffer is full")
	// ErrSharedMemoryClosed is returned when operations are attempted on a closed shared memory
	ErrSharedMemoryClosed = errors.New("shared memory is closed")
	// ErrNoSubscribers is returned when there are no subscribers for a topic
	ErrNoSubscribers = errors.New("no subscribers for topic")
	// ErrMessageTooLarge is returned when a message exceeds the slot size
	ErrMessageTooLarge = errors.New("message too large for slot")
	// ErrSharedMemoryEmpty is returned when trying to read from empty buffer
	ErrSharedMemoryEmpty = errors.New("shared memory buffer is empty")
)

const (
	// Header layout in shared memory:
	// [0-7]   : writeIndex (uint64) - next slot to write
	// [8-15]  : readIndex (uint64) - next slot to read
	// [16-23] : messageCount (uint64) - number of messages in buffer
	// [24-31] : closed (uint64) - 1 if closed, 0 otherwise
	// [32-39] : reserved
	// [40-47] : reserved
	// [48-55] : reserved
	// [56-63] : reserved
	headerSize = 64

	// Each message slot layout:
	// [0-3]   : messageLen (uint32) - length of message data
	// [4-7]   : flags (uint32) - 0=empty, 1=writing, 2=ready, 3=reading
	// [8-11]  : keyLen (uint32)
	// [12-15] : valueLen (uint32)
	// [16-19] : headerCount (uint32)
	// [20-23] : partition (int32)
	// [24-31] : offset (int64)
	// [32-39] : timestamp (int64, unix nano)
	// [40-43] : topicLen (uint32)
	// [44-..] : topic + key + value + headers data
	slotHeaderSize = 48

	// Flag values for slot state
	slotEmpty   uint32 = 0
	slotWriting uint32 = 1
	slotReady   uint32 = 2
	slotReading uint32 = 3
)

// SharedMemoryMessage implements the Message interface for shared memory transport
type SharedMemoryMessage struct {
	topic     string
	partition int32
	offset    int64
	key       []byte
	value     []byte
	headers   []MessageHeader
	timestamp time.Time
}

func (m *SharedMemoryMessage) Topic() string            { return m.topic }
func (m *SharedMemoryMessage) Partition() int32         { return m.partition }
func (m *SharedMemoryMessage) Offset() int64            { return m.offset }
func (m *SharedMemoryMessage) Key() []byte              { return m.key }
func (m *SharedMemoryMessage) Value() []byte            { return m.value }
func (m *SharedMemoryMessage) Headers() []MessageHeader { return m.headers }
func (m *SharedMemoryMessage) Timestamp() time.Time     { return m.timestamp }

// SharedMemoryHeader implements the MessageHeader interface
type SharedMemoryHeader struct {
	key   string
	value []byte
}

func (h *SharedMemoryHeader) Key() string   { return h.key }
func (h *SharedMemoryHeader) Value() []byte { return h.value }

// SharedMemoryRingBuffer represents a lock-free ring buffer in shared memory.
// This can be used for inter-process communication between producer and consumer.
//
// Memory layout:
// - Header (64 bytes): writeIndex, readIndex, messageCount, closed flag
// - Slots (slotCount * slotSize): fixed-size message slots
type SharedMemoryRingBuffer struct {
	data      []byte   // mmap'd memory region
	file      *os.File // backing file for mmap
	slotSize  int      // size of each message slot
	slotCount int      // number of slots in the ring buffer
	isOwner   bool     // true if this instance created the shared memory
	path      string   // path to the shared memory file
}

// SharedMemoryConfig holds configuration for shared memory
type SharedMemoryConfig struct {
	// Path is the file path for the memory-mapped file (e.g., "/dev/shm/myqueue" on Linux)
	Path string
	// SlotSize is the maximum size of each message slot (default: 4096)
	SlotSize int
	// SlotCount is the number of slots in the ring buffer (default: 1024)
	SlotCount int
	// Create indicates whether to create a new shared memory or open existing
	Create bool
}

// NewSharedMemoryRingBuffer creates or opens a shared memory ring buffer
func NewSharedMemoryRingBuffer(config *SharedMemoryConfig) (*SharedMemoryRingBuffer, error) {
	if config == nil {
		return nil, errors.New("config is required")
	}
	if config.Path == "" {
		return nil, errors.New("path is required")
	}
	if config.SlotSize <= 0 {
		config.SlotSize = 4096
	}
	if config.SlotCount <= 0 {
		config.SlotCount = 1024
	}

	totalSize := headerSize + (config.SlotSize * config.SlotCount)

	var file *os.File
	var err error
	var isOwner bool

	if config.Create {
		// Create new shared memory file
		file, err = os.OpenFile(config.Path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to create shared memory file: %w", err)
		}

		// Truncate to required size
		if err := file.Truncate(int64(totalSize)); err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to truncate shared memory file: %w", err)
		}
		isOwner = true
	} else {
		// Open existing shared memory file
		file, err = os.OpenFile(config.Path, os.O_RDWR, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open shared memory file: %w", err)
		}
	}

	// Memory map the file
	data, err := syscall.Mmap(
		int(file.Fd()),
		0,
		totalSize,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED,
	)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to mmap: %w", err)
	}

	rb := &SharedMemoryRingBuffer{
		data:      data,
		file:      file,
		slotSize:  config.SlotSize,
		slotCount: config.SlotCount,
		isOwner:   isOwner,
		path:      config.Path,
	}

	// Initialize header if we're the owner
	if isOwner {
		rb.setWriteIndex(0)
		rb.setReadIndex(0)
		rb.setMessageCount(0)
		rb.setClosed(false)

		// Initialize all slots as empty
		for i := 0; i < config.SlotCount; i++ {
			rb.setSlotFlag(i, slotEmpty)
		}
	}

	return rb, nil
}

// Header accessors using atomic operations
func (rb *SharedMemoryRingBuffer) writeIndexPtr() *uint64 {
	return (*uint64)(unsafe.Pointer(&rb.data[0]))
}

func (rb *SharedMemoryRingBuffer) readIndexPtr() *uint64 {
	return (*uint64)(unsafe.Pointer(&rb.data[8]))
}

func (rb *SharedMemoryRingBuffer) messageCountPtr() *uint64 {
	return (*uint64)(unsafe.Pointer(&rb.data[16]))
}

func (rb *SharedMemoryRingBuffer) closedPtr() *uint64 {
	return (*uint64)(unsafe.Pointer(&rb.data[24]))
}

func (rb *SharedMemoryRingBuffer) getWriteIndex() uint64 {
	return atomic.LoadUint64(rb.writeIndexPtr())
}

func (rb *SharedMemoryRingBuffer) setWriteIndex(v uint64) {
	atomic.StoreUint64(rb.writeIndexPtr(), v)
}

func (rb *SharedMemoryRingBuffer) getReadIndex() uint64 {
	return atomic.LoadUint64(rb.readIndexPtr())
}

func (rb *SharedMemoryRingBuffer) setReadIndex(v uint64) {
	atomic.StoreUint64(rb.readIndexPtr(), v)
}

func (rb *SharedMemoryRingBuffer) getMessageCount() uint64 {
	return atomic.LoadUint64(rb.messageCountPtr())
}

func (rb *SharedMemoryRingBuffer) setMessageCount(v uint64) {
	atomic.StoreUint64(rb.messageCountPtr(), v)
}

func (rb *SharedMemoryRingBuffer) incMessageCount() {
	atomic.AddUint64(rb.messageCountPtr(), 1)
}

func (rb *SharedMemoryRingBuffer) decMessageCount() {
	atomic.AddUint64(rb.messageCountPtr(), ^uint64(0)) // subtract 1
}

func (rb *SharedMemoryRingBuffer) isClosed() bool {
	return atomic.LoadUint64(rb.closedPtr()) == 1
}

func (rb *SharedMemoryRingBuffer) setClosed(closed bool) {
	if closed {
		atomic.StoreUint64(rb.closedPtr(), 1)
	} else {
		atomic.StoreUint64(rb.closedPtr(), 0)
	}
}

// Slot accessors
func (rb *SharedMemoryRingBuffer) slotOffset(index int) int {
	return headerSize + (index * rb.slotSize)
}

func (rb *SharedMemoryRingBuffer) slotFlagPtr(index int) *uint32 {
	offset := rb.slotOffset(index) + 4 // flags at offset 4
	if offset+4 > len(rb.data) {
		return nil
	}
	return (*uint32)(unsafe.Pointer(&rb.data[offset]))
}

func (rb *SharedMemoryRingBuffer) getSlotFlag(index int) uint32 {
	ptr := rb.slotFlagPtr(index)
	if ptr == nil {
		return slotEmpty
	}
	return atomic.LoadUint32(ptr)
}

func (rb *SharedMemoryRingBuffer) setSlotFlag(index int, flag uint32) {
	ptr := rb.slotFlagPtr(index)
	if ptr == nil {
		return
	}
	atomic.StoreUint32(ptr, flag)
}

func (rb *SharedMemoryRingBuffer) casSlotFlag(index int, old, new uint32) bool {
	ptr := rb.slotFlagPtr(index)
	if ptr == nil {
		return false
	}
	return atomic.CompareAndSwapUint32(ptr, old, new)
}

// Write writes a message to the ring buffer
func (rb *SharedMemoryRingBuffer) Write(msg Message) error {
	if len(rb.data) == 0 {
		return ErrSharedMemoryClosed
	}
	if rb.isClosed() {
		return ErrSharedMemoryClosed
	}

	// Serialize message
	msgData, err := rb.serializeMessage(msg)
	if err != nil {
		return err
	}

	if len(msgData)+slotHeaderSize > rb.slotSize {
		return ErrMessageTooLarge
	}

	// Try to claim a slot using CAS
	for {
		if rb.isClosed() {
			return ErrSharedMemoryClosed
		}

		writeIdx := rb.getWriteIndex()
		slotIdx := int(writeIdx % uint64(rb.slotCount))

		// Check if slot is empty
		if rb.getSlotFlag(slotIdx) != slotEmpty {
			// Buffer is full
			return ErrSharedMemoryFull
		}

		// Try to claim the slot
		if !rb.casSlotFlag(slotIdx, slotEmpty, slotWriting) {
			continue // Another writer got it
		}

		// Advance write index
		if !atomic.CompareAndSwapUint64(rb.writeIndexPtr(), writeIdx, writeIdx+1) {
			// Someone else advanced, release slot and retry
			rb.setSlotFlag(slotIdx, slotEmpty)
			continue
		}

		// Write message to slot
		offset := rb.slotOffset(slotIdx)

		// Bounds check before writing
		if offset+slotHeaderSize+len(msgData) > len(rb.data) {
			rb.setSlotFlag(slotIdx, slotEmpty)
			return ErrMessageTooLarge
		}

		binary.LittleEndian.PutUint32(rb.data[offset:], uint32(len(msgData)))
		copy(rb.data[offset+slotHeaderSize:], msgData)

		// Mark slot as ready
		rb.setSlotFlag(slotIdx, slotReady)
		rb.incMessageCount()

		return nil
	}
}

// Read reads a message from the ring buffer
func (rb *SharedMemoryRingBuffer) Read() (Message, error) {
	if len(rb.data) == 0 {
		return nil, ErrSharedMemoryClosed
	}
	if rb.isClosed() && rb.getMessageCount() == 0 {
		return nil, ErrSharedMemoryClosed
	}

	for {
		readIdx := rb.getReadIndex()
		slotIdx := int(readIdx % uint64(rb.slotCount))

		// Validate slot index
		if slotIdx < 0 || slotIdx >= rb.slotCount {
			return nil, ErrSharedMemoryEmpty
		}

		// Check if slot is ready
		flag := rb.getSlotFlag(slotIdx)
		if flag != slotReady {
			if rb.isClosed() {
				return nil, ErrSharedMemoryClosed
			}
			return nil, ErrSharedMemoryEmpty
		}

		// Try to claim the slot for reading
		if !rb.casSlotFlag(slotIdx, slotReady, slotReading) {
			continue // Another reader got it
		}

		// Advance read index
		if !atomic.CompareAndSwapUint64(rb.readIndexPtr(), readIdx, readIdx+1) {
			// Someone else advanced, this shouldn't happen in single consumer
			// but handle gracefully
			rb.setSlotFlag(slotIdx, slotReady)
			continue
		}

		// Read message from slot
		offset := rb.slotOffset(slotIdx)

		// Bounds check before reading
		if offset+slotHeaderSize > len(rb.data) {
			rb.setSlotFlag(slotIdx, slotEmpty)
			return nil, ErrSharedMemoryEmpty
		}

		msgLen := binary.LittleEndian.Uint32(rb.data[offset:])

		// Validate message length
		if int(msgLen) < 0 || offset+slotHeaderSize+int(msgLen) > len(rb.data) {
			rb.setSlotFlag(slotIdx, slotEmpty)
			rb.decMessageCount()
			return nil, ErrSharedMemoryEmpty
		}

		msgData := make([]byte, msgLen)
		copy(msgData, rb.data[offset+slotHeaderSize:offset+slotHeaderSize+int(msgLen)])

		// Clear slot and mark as empty
		rb.setSlotFlag(slotIdx, slotEmpty)
		rb.decMessageCount()

		// Deserialize message
		return rb.deserializeMessage(msgData)
	}
}

// serializeMessage converts a Message to bytes
func (rb *SharedMemoryRingBuffer) serializeMessage(msg Message) ([]byte, error) {
	topic := msg.Topic()
	key := msg.Key()
	value := msg.Value()
	headers := msg.Headers()

	// Calculate total size
	size := 4 + len(topic) + // topicLen + topic
		4 + len(key) + // keyLen + key
		4 + len(value) + // valueLen + value
		4 + // partition
		8 + // offset
		8 + // timestamp
		4 // headerCount

	for _, h := range headers {
		size += 4 + len(h.Key()) + 4 + len(h.Value())
	}

	data := make([]byte, size)
	offset := 0

	// Topic
	binary.LittleEndian.PutUint32(data[offset:], uint32(len(topic)))
	offset += 4
	copy(data[offset:], topic)
	offset += len(topic)

	// Key
	binary.LittleEndian.PutUint32(data[offset:], uint32(len(key)))
	offset += 4
	copy(data[offset:], key)
	offset += len(key)

	// Value
	binary.LittleEndian.PutUint32(data[offset:], uint32(len(value)))
	offset += 4
	copy(data[offset:], value)
	offset += len(value)

	// Partition
	binary.LittleEndian.PutUint32(data[offset:], uint32(msg.Partition()))
	offset += 4

	// Offset
	binary.LittleEndian.PutUint64(data[offset:], uint64(msg.Offset()))
	offset += 8

	// Timestamp
	binary.LittleEndian.PutUint64(data[offset:], uint64(msg.Timestamp().UnixNano()))
	offset += 8

	// Headers
	binary.LittleEndian.PutUint32(data[offset:], uint32(len(headers)))
	offset += 4

	for _, h := range headers {
		// Header key
		hKey := h.Key()
		binary.LittleEndian.PutUint32(data[offset:], uint32(len(hKey)))
		offset += 4
		copy(data[offset:], hKey)
		offset += len(hKey)

		// Header value
		hVal := h.Value()
		binary.LittleEndian.PutUint32(data[offset:], uint32(len(hVal)))
		offset += 4
		copy(data[offset:], hVal)
		offset += len(hVal)
	}

	return data, nil
}

// deserializeMessage converts bytes back to a Message
func (rb *SharedMemoryRingBuffer) deserializeMessage(data []byte) (Message, error) {
	if len(data) < 24 {
		return nil, errors.New("message data too short")
	}

	offset := 0

	// Topic
	topicLen := binary.LittleEndian.Uint32(data[offset:])
	offset += 4
	topic := string(data[offset : offset+int(topicLen)])
	offset += int(topicLen)

	// Key
	keyLen := binary.LittleEndian.Uint32(data[offset:])
	offset += 4
	key := make([]byte, keyLen)
	copy(key, data[offset:offset+int(keyLen)])
	offset += int(keyLen)

	// Value
	valueLen := binary.LittleEndian.Uint32(data[offset:])
	offset += 4
	value := make([]byte, valueLen)
	copy(value, data[offset:offset+int(valueLen)])
	offset += int(valueLen)

	// Partition
	partition := int32(binary.LittleEndian.Uint32(data[offset:]))
	offset += 4

	// Offset
	msgOffset := int64(binary.LittleEndian.Uint64(data[offset:]))
	offset += 8

	// Timestamp
	timestampNano := int64(binary.LittleEndian.Uint64(data[offset:]))
	timestamp := time.Unix(0, timestampNano)
	offset += 8

	// Headers
	headerCount := binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	headers := make([]MessageHeader, headerCount)
	for i := uint32(0); i < headerCount; i++ {
		// Header key
		hKeyLen := binary.LittleEndian.Uint32(data[offset:])
		offset += 4
		hKey := string(data[offset : offset+int(hKeyLen)])
		offset += int(hKeyLen)

		// Header value
		hValLen := binary.LittleEndian.Uint32(data[offset:])
		offset += 4
		hVal := make([]byte, hValLen)
		copy(hVal, data[offset:offset+int(hValLen)])
		offset += int(hValLen)

		headers[i] = &SharedMemoryHeader{key: hKey, value: hVal}
	}

	return &SharedMemoryMessage{
		topic:     topic,
		partition: partition,
		offset:    msgOffset,
		key:       key,
		value:     value,
		headers:   headers,
		timestamp: timestamp,
	}, nil
}

// Close closes the shared memory ring buffer
func (rb *SharedMemoryRingBuffer) Close() error {
	rb.setClosed(true)

	if err := syscall.Munmap(rb.data); err != nil {
		return fmt.Errorf("failed to munmap: %w", err)
	}

	if err := rb.file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	return nil
}

// Destroy removes the shared memory file (should only be called by owner)
func (rb *SharedMemoryRingBuffer) Destroy() error {
	if err := rb.Close(); err != nil {
		return err
	}
	return os.Remove(rb.path)
}

// Size returns the number of messages in the buffer
func (rb *SharedMemoryRingBuffer) Size() int {
	return int(rb.getMessageCount())
}

// IsFull returns true if the buffer is full
func (rb *SharedMemoryRingBuffer) IsFull() bool {
	return rb.getMessageCount() >= uint64(rb.slotCount)
}

// IsEmpty returns true if the buffer is empty
func (rb *SharedMemoryRingBuffer) IsEmpty() bool {
	return rb.getMessageCount() == 0
}

// SharedMemoryProducer implements the Producer interface using shared memory
type SharedMemoryProducer struct {
	rb        *SharedMemoryRingBuffer
	topic     string
	partition int32
	offset    atomic.Int64
}

// SharedMemoryProducerConfig holds configuration for SharedMemoryProducer
type SharedMemoryProducerConfig struct {
	// Path is the file path for the memory-mapped file
	Path string
	// Topic is the default topic to send messages to
	Topic string
	// Partition is the partition to assign to messages
	Partition int32
	// SlotSize is the maximum size of each message (default: 4096)
	SlotSize int
	// SlotCount is the number of slots in the ring buffer (default: 1024)
	SlotCount int
	// Create indicates whether to create new shared memory (true) or open existing (false)
	Create bool
}

// NewSharedMemoryProducer creates a new shared memory producer
func NewSharedMemoryProducer(config *SharedMemoryProducerConfig) (*SharedMemoryProducer, error) {
	if config == nil {
		return nil, errors.New("config is required")
	}

	rb, err := NewSharedMemoryRingBuffer(&SharedMemoryConfig{
		Path:      config.Path,
		SlotSize:  config.SlotSize,
		SlotCount: config.SlotCount,
		Create:    config.Create,
	})
	if err != nil {
		return nil, err
	}

	return &SharedMemoryProducer{
		rb:        rb,
		topic:     config.Topic,
		partition: config.Partition,
	}, nil
}

// Send sends a message to the shared memory buffer
func (p *SharedMemoryProducer) Send(message Message) error {
	topic := message.Topic()
	if topic == "" {
		topic = p.topic
	}

	newOffset := p.offset.Add(1)

	// Wrap the message with metadata
	smMsg := &SharedMemoryMessage{
		topic:     topic,
		partition: p.partition,
		offset:    newOffset,
		key:       message.Key(),
		value:     message.Value(),
		headers:   message.Headers(),
		timestamp: time.Now(),
	}

	return p.rb.Write(smMsg)
}

// SendRaw sends a raw message with key and value
func (p *SharedMemoryProducer) SendRaw(key, value []byte, headers ...MessageHeader) error {
	msg := &SharedMemoryMessage{
		topic:     p.topic,
		partition: p.partition,
		offset:    p.offset.Add(1),
		key:       key,
		value:     value,
		headers:   headers,
		timestamp: time.Now(),
	}
	return p.rb.Write(msg)
}

// Close closes the producer
func (p *SharedMemoryProducer) Close() error {
	return p.rb.Close()
}

// SharedMemoryConsumer implements message consumption using shared memory
type SharedMemoryConsumer struct {
	rb           *SharedMemoryRingBuffer
	handlers     []ConsumeHandler
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	pollInterval time.Duration
}

// SharedMemoryConsumerConfig holds configuration for SharedMemoryConsumer
type SharedMemoryConsumerConfig struct {
	// Path is the file path for the memory-mapped file
	Path string
	// SlotSize must match the producer's slot size
	SlotSize int
	// SlotCount must match the producer's slot count
	SlotCount int
	// PollInterval is how often to check for new messages (default: 1ms)
	PollInterval time.Duration
	// Create should be false for consumer (producer creates the shared memory)
	Create bool
}

// Minimum poll interval to prevent SIGBUS errors from rapid mmap access
const minPollInterval = 10 * time.Microsecond

// NewSharedMemoryConsumer creates a new shared memory consumer
func NewSharedMemoryConsumer(config *SharedMemoryConsumerConfig) (*SharedMemoryConsumer, error) {
	if config == nil {
		return nil, errors.New("config is required")
	}

	if config.PollInterval <= 0 {
		config.PollInterval = time.Millisecond
	} else if config.PollInterval < minPollInterval {
		// Enforce minimum poll interval to prevent SIGBUS from rapid mmap access
		config.PollInterval = minPollInterval
	}

	rb, err := NewSharedMemoryRingBuffer(&SharedMemoryConfig{
		Path:      config.Path,
		SlotSize:  config.SlotSize,
		SlotCount: config.SlotCount,
		Create:    config.Create,
	})
	if err != nil {
		return nil, err
	}

	return &SharedMemoryConsumer{
		rb:           rb,
		handlers:     make([]ConsumeHandler, 0),
		pollInterval: config.PollInterval,
	}, nil
}

// Handle registers a message handler
func (c *SharedMemoryConsumer) Handle(fn ConsumeHandler) {
	c.handlers = append(c.handlers, fn)
}

// Start begins consuming messages
func (c *SharedMemoryConsumer) Start(ctx context.Context) {
	ctx, c.cancel = context.WithCancel(ctx)

	c.wg.Add(1)
	go c.consumeLoop(ctx)
}

// consumeLoop continuously polls for messages
func (c *SharedMemoryConsumer) consumeLoop(ctx context.Context) {
	defer c.wg.Done()

	ticker := time.NewTicker(c.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.pollMessages(ctx)
		}
	}
}

// pollMessages reads available messages from the buffer
func (c *SharedMemoryConsumer) pollMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msg, err := c.rb.Read()
		if err != nil {
			if err == ErrSharedMemoryEmpty || err == ErrSharedMemoryClosed {
				return
			}
			// Log error and continue
			continue
		}

		c.processMessage(ctx, msg)
	}
}

// processMessage invokes all registered handlers for a message
func (c *SharedMemoryConsumer) processMessage(ctx context.Context, msg Message) {
	for _, handler := range c.handlers {
		if err := handler(ctx, msg); err != nil {
			// Log error but continue processing
			continue
		}
	}
}

// Close stops the consumer
func (c *SharedMemoryConsumer) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()
	return c.rb.Close()
}

// ensure SharedMemoryProducer implements Producer interface
var _ Producer = (*SharedMemoryProducer)(nil)
