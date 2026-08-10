package messages

import (
	"time"

	"github.com/IBM/sarama"
	"github.com/nats-io/nats.go"
)

// Message represents a message consumed from a messaging system
// This interface abstracts away the details of the underlying messaging implementation
type Message interface {
	// Topic returns the name of the topic this message was consumed from
	Topic() string

	// Partition returns the partition number this message was consumed from
	Partition() int32

	// Offset returns the offset of this message in its partition
	Offset() int64

	// Key returns the message key if present, or nil if there is none
	Key() []byte

	// Value returns the message content
	Value() []byte

	// Headers returns the message headers if present
	Headers() []MessageHeader

	// Timestamp returns the timestamp associated with this message
	Timestamp() time.Time
}

// MessageHeader represents a key-value pair header attached to a message
type MessageHeader interface {
	// Key returns the header key
	Key() string

	// Value returns the header value
	Value() []byte
}

// SaramaMessageAdapter adapts Sarama's ConsumerMessage to our ConsumerMessage interface
// This adapter would be implemented in a separate file that has the Sarama dependency
type SaramaMessageAdapter interface {
	// ToConsumerMessage converts a Sarama message to our generic ConsumerMessage interface
	ToConsumerMessage() SaramaMessage
}

// MessageProcessor defines an interface for components that process consumer messages
type MessageProcessor interface {
	// Process handles a consumer message
	Process(msg SaramaMessage) error
}

var _ Message = (*SaramaMessage)(nil)
var _ MessageHeader = (*saramaMessageHeader)(nil)

type SaramaMessage struct {
	Message *sarama.ConsumerMessage
}

// Implementing ConsumerMessage interface for SaramaMessage
func (m *SaramaMessage) Topic() string {
	return m.Message.Topic
}

func (m *SaramaMessage) Partition() int32 {
	return m.Message.Partition
}

func (m *SaramaMessage) Offset() int64 {
	return m.Message.Offset
}

func (m *SaramaMessage) Key() []byte {
	return m.Message.Key
}

func (m *SaramaMessage) Value() []byte {
	return m.Message.Value
}

func (m *SaramaMessage) Headers() []MessageHeader {
	if len(m.Message.Headers) == 0 {
		return nil
	}

	headers := make([]MessageHeader, len(m.Message.Headers))
	for i, h := range m.Message.Headers {
		headers[i] = &saramaMessageHeader{header: h}
	}
	return headers
}

func (m *SaramaMessage) Timestamp() time.Time {
	return m.Message.Timestamp
}

// Implementation of MessageHeader for Sarama's RecordHeader
type saramaMessageHeader struct {
	header *sarama.RecordHeader
}

func (h *saramaMessageHeader) Key() string {
	return string(h.header.Key)
}

func (h *saramaMessageHeader) Value() []byte {
	return h.header.Value
}

// NewSaramaMessage creates a new SaramaMessage from a Sarama ConsumerMessage
func NewSaramaMessage(message *sarama.ConsumerMessage) Message {
	return &SaramaMessage{Message: message}
}

type SimpleMessage struct {
	topic string
	key   []byte
	value []byte
}

func (m *SimpleMessage) Topic() string {
	return m.topic
}

func (m *SimpleMessage) Partition() int32 {
	return 0 // SimpleMessage does not have partition information
}

func (m *SimpleMessage) Offset() int64 {
	return 0 // SimpleMessage does not have offset information
}

func (m *SimpleMessage) Key() []byte {
	return m.key
}

func (m *SimpleMessage) Value() []byte {
	return m.value
}

func (m *SimpleMessage) Headers() []MessageHeader {
	return nil // SimpleMessage does not have headers
}

func (m *SimpleMessage) Timestamp() time.Time {
	return time.Now() // SimpleMessage does not have a specific timestamp
}

func NewSimpleMessage(topic string, key, value []byte) *SimpleMessage {
	return &SimpleMessage{
		topic: topic,
		key:   key,
		value: value,
	}
}

type natMessage struct {
	msg *nats.Msg
}

// Implementing Message interface for natMessage
func (m *natMessage) Topic() string {
	return m.msg.Subject
}

func (m *natMessage) Partition() int32 {
	return 0 // NATS does not have partition concept
}

func (m *natMessage) Offset() int64 {
	return 0 // NATS does not have offset concept
}

func (m *natMessage) Key() []byte {
	return nil // NATS does not have key concept
}

func (m *natMessage) Value() []byte {
	return m.msg.Data
}

func (m *natMessage) Headers() []MessageHeader {
	if len(m.msg.Header) == 0 {
		return nil
	}

	headers := make([]MessageHeader, 0, len(m.msg.Header))
	for k, vals := range m.msg.Header {
		for _, v := range vals {
			headers = append(headers, &natsMessageHeader{key: k, value: []byte(v)})
		}
	}
	return headers
}

func (m *natMessage) Timestamp() time.Time {
	// NATS does not provide a message timestamp by default
	return time.Time{}
}

// MessageHeader implementation for NATS
type natsMessageHeader struct {
	key   string
	value []byte
}

func (h *natsMessageHeader) Key() string {
	return h.key
}

func (h *natsMessageHeader) Value() []byte {
	return h.value
}

func NewNATsMessage(msg *nats.Msg) Message {
	return &natMessage{
		msg: msg,
	}
}
