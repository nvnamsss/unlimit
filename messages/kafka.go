package messages

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

var (
	// ErrKafkaProducerClosed is returned when operations are attempted on a closed producer
	ErrKafkaProducerClosed = errors.New("kafka producer is closed")
	// ErrKafkaNotConnected is returned when the producer is not connected to Kafka
	ErrKafkaNotConnected = errors.New("not connected to kafka")
	// ErrKafkaInvalidConfig is returned when the configuration is invalid
	ErrKafkaInvalidConfig = errors.New("invalid kafka configuration")
)

// KafkaSyncProducer wraps Sarama's synchronous producer
type KafkaSyncProducer struct {
	producer sarama.SyncProducer
	closed   bool
	mu       sync.RWMutex
}

// Compression types
const (
	CompressionNone   = "none"
	CompressionGZIP   = "gzip"
	CompressionSnappy = "snappy"
	CompressionLZ4    = "lz4"
	CompressionZSTD   = "zstd"
)

// RequiredAcks levels
const (
	NoResponse   = 0  // no response from broker
	WaitForLocal = 1  // wait for leader acknowledgement
	WaitForAll   = -1 // wait for all in-sync replicas
)

// ProducerError represents an error from the async producer
type ProducerError struct {
	Topic     string
	Partition int32
	Offset    int64
	Err       error
}

// ProducerSuccess represents a successful message delivery
type ProducerSuccess struct {
	Topic     string
	Partition int32
	Offset    int64
}

// KafkaProducerConfig holds configuration for Kafka producers
type KafkaProducerConfig struct {
	// Brokers is the list of Kafka broker addresses
	Brokers []string

	// Compression type (none, gzip, snappy, lz4, zstd)
	Compression string

	// RequiredAcks determines the number of required acks (-1: all ISRs, 0: no acks, 1: leader only)
	RequiredAcks int

	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// RetryBackoff is the time to wait before retrying
	RetryBackoff time.Duration

	// Idempotent enables idempotent producer for exactly-once semantics
	Idempotent bool

	// Timeout for produce requests
	Timeout time.Duration

	// FlushMessages is the number of messages to buffer before flushing (async only)
	FlushMessages int

	// FlushFrequency is the time to wait before flushing (async only)
	FlushFrequency time.Duration

	// FlushMaxMessages is the maximum number of messages per request (async only)
	FlushMaxMessages int

	// ReturnSuccesses enables success notifications (required for async, default true)
	ReturnSuccesses bool

	// ReturnErrors enables error notifications (default true)
	ReturnErrors bool

	// ErrorHandler is called for async producer errors (optional)
	ErrorHandler func(*ProducerError)

	// SuccessHandler is called for async producer successes (optional)
	SuccessHandler func(*ProducerSuccess)
}

// DefaultKafkaProducerConfig returns a production-ready configuration
func DefaultKafkaProducerConfig() *KafkaProducerConfig {
	return &KafkaProducerConfig{
		Compression:      CompressionSnappy,
		RequiredAcks:     WaitForAll,
		MaxRetries:       5,
		RetryBackoff:     100 * time.Millisecond,
		Idempotent:       false,
		Timeout:          10 * time.Second,
		FlushMessages:    100,
		FlushFrequency:   10 * time.Millisecond,
		FlushMaxMessages: 1000,
		ReturnSuccesses:  true,
		ReturnErrors:     true,
	}
}

// mapCompression converts string compression type to Sarama compression codec
func mapCompression(compression string) sarama.CompressionCodec {
	switch compression {
	case CompressionGZIP:
		return sarama.CompressionGZIP
	case CompressionSnappy:
		return sarama.CompressionSnappy
	case CompressionLZ4:
		return sarama.CompressionLZ4
	case CompressionZSTD:
		return sarama.CompressionZSTD
	case CompressionNone:
		return sarama.CompressionNone
	default:
		return sarama.CompressionNone
	}
}

// mapRequiredAcks converts int acks level to Sarama RequiredAcks
func mapRequiredAcks(acks int) sarama.RequiredAcks {
	switch acks {
	case NoResponse:
		return sarama.NoResponse
	case WaitForLocal:
		return sarama.WaitForLocal
	case WaitForAll:
		return sarama.WaitForAll
	default:
		return sarama.WaitForLocal
	}
}

// buildSaramaConfig creates a Sarama configuration from KafkaProducerConfig
func buildSaramaConfig(config *KafkaProducerConfig) *sarama.Config {
	saramaConf := sarama.NewConfig()

	// Producer settings
	saramaConf.Producer.Return.Successes = config.ReturnSuccesses
	saramaConf.Producer.Return.Errors = config.ReturnErrors
	saramaConf.Producer.RequiredAcks = mapRequiredAcks(config.RequiredAcks)
	saramaConf.Producer.Compression = mapCompression(config.Compression)
	saramaConf.Producer.Retry.Max = config.MaxRetries
	saramaConf.Producer.Retry.Backoff = config.RetryBackoff
	saramaConf.Producer.Timeout = config.Timeout

	// Idempotence settings
	if config.Idempotent {
		saramaConf.Producer.Idempotent = true
		saramaConf.Producer.RequiredAcks = sarama.WaitForAll
		saramaConf.Net.MaxOpenRequests = 1
	}

	// Flush settings for async producer
	if config.FlushMessages > 0 {
		saramaConf.Producer.Flush.Messages = config.FlushMessages
	}
	if config.FlushFrequency > 0 {
		saramaConf.Producer.Flush.Frequency = config.FlushFrequency
	}
	if config.FlushMaxMessages > 0 {
		saramaConf.Producer.Flush.MaxMessages = config.FlushMaxMessages
	}

	return saramaConf
}

// buildProducerMessage creates a Sarama ProducerMessage from our Message interface
func buildProducerMessage(message Message) *sarama.ProducerMessage {
	saramaMessage := &sarama.ProducerMessage{
		Topic:     message.Topic(),
		Value:     sarama.ByteEncoder(message.Value()),
		Timestamp: message.Timestamp(),
	}

	// Set key if present
	if key := message.Key(); key != nil {
		saramaMessage.Key = sarama.ByteEncoder(key)
	}

	// Set partition if specified (use -1 for automatic partitioning)
	if partition := message.Partition(); partition >= 0 {
		saramaMessage.Partition = partition
	}

	// Convert headers
	if headers := message.Headers(); len(headers) > 0 {
		saramaMessage.Headers = make([]sarama.RecordHeader, len(headers))
		for i, h := range headers {
			saramaMessage.Headers[i] = sarama.RecordHeader{
				Key:   []byte(h.Key()),
				Value: h.Value(),
			}
		}
	}

	return saramaMessage
}

func (p *KafkaSyncProducer) Send(message Message) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return ErrKafkaProducerClosed
	}

	saramaMessage := buildProducerMessage(message)
	partition, offset, err := p.producer.SendMessage(saramaMessage)
	if err != nil {
		return fmt.Errorf("failed to send message to kafka: %w", err)
	}

	// Update message with broker-assigned values (if message supports it)
	// Note: Our Message interface is read-only, so we can't update it
	// Applications can use partition/offset from return if needed in future
	_ = partition
	_ = offset

	return nil
}

func (p *KafkaSyncProducer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	return p.producer.Close()
}

// NewKafkaProducer creates a new synchronous Kafka producer
func NewKafkaProducer(config *KafkaProducerConfig) (Producer, error) {
	if config == nil {
		config = DefaultKafkaProducerConfig()
	}

	if len(config.Brokers) == 0 {
		return nil, fmt.Errorf("%w: no brokers specified", ErrKafkaInvalidConfig)
	}

	// Apply default values for required fields if not set
	if config.Timeout <= 0 {
		config.Timeout = 10 * time.Second
	}
	if config.RetryBackoff <= 0 {
		config.RetryBackoff = 100 * time.Millisecond
	}
	if config.Compression == "" {
		config.Compression = CompressionNone
	}

	// Ensure return successes and errors are true for sync producer (required by Sarama)
	config.ReturnSuccesses = true
	config.ReturnErrors = true

	saramaConf := buildSaramaConfig(config)
	producer, err := sarama.NewSyncProducer(config.Brokers, saramaConf)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka sync producer: %w", err)
	}

	return &KafkaSyncProducer{
		producer: producer,
		closed:   false,
	}, nil
}

// KafkaAsyncProducer wraps Sarama's asynchronous producer
type KafkaAsyncProducer struct {
	producer sarama.AsyncProducer
	closed   bool
	mu       sync.RWMutex
	wg       sync.WaitGroup
	cancel   chan struct{}
}

func (p *KafkaAsyncProducer) Send(message Message) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return ErrKafkaProducerClosed
	}

	saramaMessage := buildProducerMessage(message)

	// Non-blocking send to input channel
	select {
	case p.producer.Input() <- saramaMessage:
		return nil
	case <-p.cancel:
		return ErrKafkaProducerClosed
	}
}

func (p *KafkaAsyncProducer) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	close(p.cancel)
	p.mu.Unlock()

	// Close the producer (this will close input channel)
	p.producer.AsyncClose()

	// Wait for error/success handlers to finish draining channels
	p.wg.Wait()

	return nil
}

// NewKafkaAsyncProducer creates a new asynchronous Kafka producer with background goroutines
// to handle errors and successes
func NewKafkaAsyncProducer(config *KafkaProducerConfig) (Producer, error) {
	if config == nil {
		config = DefaultKafkaProducerConfig()
	}

	if len(config.Brokers) == 0 {
		return nil, fmt.Errorf("%w: no brokers specified", ErrKafkaInvalidConfig)
	}

	// Apply default values for required fields if not set
	if config.Timeout <= 0 {
		config.Timeout = 10 * time.Second
	}
	if config.RetryBackoff <= 0 {
		config.RetryBackoff = 100 * time.Millisecond
	}
	if config.Compression == "" {
		config.Compression = CompressionNone
	}

	// Ensure return flags are set for async producer
	config.ReturnSuccesses = true
	config.ReturnErrors = true

	saramaConf := buildSaramaConfig(config)
	producer, err := sarama.NewAsyncProducer(config.Brokers, saramaConf)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka async producer: %w", err)
	}

	ap := &KafkaAsyncProducer{
		producer: producer,
		closed:   false,
		cancel:   make(chan struct{}),
	}

	// Start goroutines to drain error and success channels
	// This prevents goroutine leaks and deadlocks
	ap.wg.Add(2)

	// Error handler goroutine
	go func() {
		defer ap.wg.Done()
		for err := range producer.Errors() {
			prodErr := &ProducerError{
				Topic:     err.Msg.Topic,
				Partition: err.Msg.Partition,
				Offset:    err.Msg.Offset,
				Err:       err.Err,
			}
			if config.ErrorHandler != nil {
				config.ErrorHandler(prodErr)
			} else {
				// Default: log to stderr
				log.Printf("kafka async producer error: topic=%s partition=%d offset=%d error=%v",
					prodErr.Topic, prodErr.Partition, prodErr.Offset, prodErr.Err)
			}
		}
	}()

	// Success handler goroutine
	go func() {
		defer ap.wg.Done()
		for msg := range producer.Successes() {
			if config.SuccessHandler != nil {
				prodSuccess := &ProducerSuccess{
					Topic:     msg.Topic,
					Partition: msg.Partition,
					Offset:    msg.Offset,
				}
				config.SuccessHandler(prodSuccess)
			}
			// Otherwise silently acknowledge success
		}
	}()

	return ap, nil
}
