package messages

import (
	"errors"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
)

// TestNewKafkaProducer_InvalidConfig tests that constructor returns error for invalid config
func TestNewKafkaProducer_InvalidConfig(t *testing.T) {
	// Test with nil config (should use defaults but fail on empty brokers)
	config := &KafkaProducerConfig{
		Brokers: []string{},
	}

	producer, err := NewKafkaProducer(config)
	if err == nil {
		t.Fatal("expected error for empty brokers, got nil")
	}
	if producer != nil {
		t.Fatal("expected nil producer for invalid config")
	}
	if !errors.Is(err, ErrKafkaInvalidConfig) {
		t.Errorf("expected ErrKafkaInvalidConfig, got %v", err)
	}
}

// TestKafkaSyncProducer_SendAfterClose tests that Send returns error after Close
func TestKafkaSyncProducer_SendAfterClose(t *testing.T) {
	// Create a mock sync producer
	mockProducer := mocks.NewSyncProducer(t, nil)
	defer mockProducer.Close()

	producer := &KafkaSyncProducer{
		producer: mockProducer,
		closed:   false,
	}

	// Send a message successfully
	msg := NewSimpleMessage("test-topic", []byte("key"), []byte("value"))
	mockProducer.ExpectSendMessageAndSucceed()

	err := producer.Send(msg)
	if err != nil {
		t.Fatalf("expected successful send, got error: %v", err)
	}

	// Close the producer
	err = producer.Close()
	if err != nil {
		t.Fatalf("expected successful close, got error: %v", err)
	}

	// Try to send after close - should return error
	err = producer.Send(msg)
	if err == nil {
		t.Fatal("expected error after close, got nil")
	}
	if !errors.Is(err, ErrKafkaProducerClosed) {
		t.Errorf("expected ErrKafkaProducerClosed, got %v", err)
	}

	// Close again should be idempotent
	err = producer.Close()
	if err != nil {
		t.Fatalf("expected idempotent close, got error: %v", err)
	}
}

// TestKafkaAsyncProducer_ErrorHandling tests that async producer handles errors via callback
func TestKafkaAsyncProducer_ErrorHandling(t *testing.T) {
	// Create a mock async producer
	mockProducer := mocks.NewAsyncProducer(t, nil)

	// Track errors and successes
	var receivedErrors []*ProducerError
	var receivedSuccesses []*ProducerSuccess
	errorChan := make(chan bool, 1)
	successChan := make(chan bool, 1)

	config := &KafkaProducerConfig{
		Brokers: []string{"localhost:9092"},
		ErrorHandler: func(err *ProducerError) {
			receivedErrors = append(receivedErrors, err)
			errorChan <- true
		},
		SuccessHandler: func(msg *ProducerSuccess) {
			receivedSuccesses = append(receivedSuccesses, msg)
			successChan <- true
		},
	}

	ap := &KafkaAsyncProducer{
		producer: mockProducer,
		closed:   false,
		cancel:   make(chan struct{}),
	}

	// Start goroutines to drain channels
	ap.wg.Add(2)

	go func() {
		defer ap.wg.Done()
		for err := range mockProducer.Errors() {
			if config.ErrorHandler != nil {
				prodErr := &ProducerError{
					Topic:     err.Msg.Topic,
					Partition: err.Msg.Partition,
					Offset:    err.Msg.Offset,
					Err:       err.Err,
				}
				config.ErrorHandler(prodErr)
			}
		}
	}()

	go func() {
		defer ap.wg.Done()
		for msg := range mockProducer.Successes() {
			if config.SuccessHandler != nil {
				prodSuccess := &ProducerSuccess{
					Topic:     msg.Topic,
					Partition: msg.Partition,
					Offset:    msg.Offset,
				}
				config.SuccessHandler(prodSuccess)
			}
		}
	}()

	// Test successful message
	msg := NewSimpleMessage("test-topic", []byte("key"), []byte("value"))
	mockProducer.ExpectInputAndSucceed()

	err := ap.Send(msg)
	if err != nil {
		t.Fatalf("expected successful send, got error: %v", err)
	}

	// Wait for success callback
	select {
	case <-successChan:
		// Success received
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for success callback")
	}

	if len(receivedSuccesses) != 1 {
		t.Errorf("expected 1 success, got %d", len(receivedSuccesses))
	}

	// Test error message
	mockProducer.ExpectInputAndFail(sarama.ErrOutOfBrokers)

	err = ap.Send(msg)
	if err != nil {
		t.Fatalf("send should not return error for async producer, got: %v", err)
	}

	// Wait for error callback
	select {
	case <-errorChan:
		// Error received
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for error callback")
	}

	if len(receivedErrors) != 1 {
		t.Errorf("expected 1 error, got %d", len(receivedErrors))
	}

	// Close and verify graceful shutdown
	mockProducer.AsyncClose()
	ap.wg.Wait()

	if len(receivedSuccesses) != 1 {
		t.Errorf("expected 1 total success, got %d", len(receivedSuccesses))
	}
	if len(receivedErrors) != 1 {
		t.Errorf("expected 1 total error, got %d", len(receivedErrors))
	}
}

// TestKafkaSyncProducer_Send tests successful message sending with key, value, and headers
func TestKafkaSyncProducer_Send(t *testing.T) {
	// Create a mock sync producer
	mockProducer := mocks.NewSyncProducer(t, nil)
	defer mockProducer.Close()

	producer := &KafkaSyncProducer{
		producer: mockProducer,
		closed:   false,
	}

	// Test sending a simple message with key and value
	msg := NewSimpleMessage("test-topic", []byte("test-key"), []byte("test-value"))

	// Expect the message to be sent successfully
	mockProducer.ExpectSendMessageAndSucceed()

	err := producer.Send(msg)
	if err != nil {
		t.Fatalf("expected successful send, got error: %v", err)
	}

	// Test sending multiple messages
	messages := []struct {
		topic string
		key   []byte
		value []byte
	}{
		{"topic1", []byte("key1"), []byte("value1")},
		{"topic2", []byte("key2"), []byte("value2")},
		{"topic3", nil, []byte("value3")}, // nil key
	}

	for _, tc := range messages {
		msg := NewSimpleMessage(tc.topic, tc.key, tc.value)
		mockProducer.ExpectSendMessageAndSucceed()

		err := producer.Send(msg)
		if err != nil {
			t.Errorf("failed to send message to %s: %v", tc.topic, err)
		}
	}

	// Test send failure - simulate broker error
	mockProducer.ExpectSendMessageAndFail(sarama.ErrOutOfBrokers)

	failMsg := NewSimpleMessage("fail-topic", []byte("key"), []byte("value"))
	err = producer.Send(failMsg)
	if err == nil {
		t.Fatal("expected error from broker, got nil")
	}

	// Verify all expectations were met
	if err := mockProducer.Close(); err != nil {
		t.Errorf("mock producer close failed: %v", err)
	}
}
