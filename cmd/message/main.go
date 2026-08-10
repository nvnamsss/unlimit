package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/voidforge-studios/unlimit/messages"
)

const (
	brokerKafka = "localhost:29092"
)

func main() {
	mode := "producer" // Change to "consumer" to run consumer mode
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	switch mode {
	case "producer":
		runProducer()
	case "consumer":
		runConsumer()
	default:
		fmt.Println("Usage: go run main.go [producer|consumer]")
		os.Exit(1)
	}
}

func runProducer() {
	fmt.Println("=== Kafka Producer ===")
	fmt.Println()

	config := &messages.KafkaProducerConfig{
		Brokers:         []string{brokerKafka},
		Compression:     messages.CompressionSnappy,
		RequiredAcks:    messages.WaitForLocal,
		MaxRetries:      5,
		RetryBackoff:    100 * time.Millisecond,
		Timeout:         10 * time.Second,
		ReturnSuccesses: true,
		ReturnErrors:    true,
	}

	// Create synchronous producer
	producer, err := messages.NewKafkaProducer(config)
	if err != nil {
		log.Fatalf("failed to create kafka producer: %v", err)
	}
	defer producer.Close()

	// Send a simple message
	msg := messages.NewSimpleMessage(
		"demo-topic",
		[]byte("user-123"),
		[]byte("Hello from Kafka producer!"),
	)

	err = producer.Send(msg)
	if err != nil {
		log.Fatalf("failed to send message: %v", err)
	}

	fmt.Println("✓ Message sent successfully to topic: demo-topic")

	// Send multiple messages
	for i := 0; i < 5; i++ {
		msg := messages.NewSimpleMessage(
			"demo-topic",
			[]byte(fmt.Sprintf("key-%d", i)),
			[]byte(fmt.Sprintf("Message number %d", i)),
		)

		err = producer.Send(msg)
		if err != nil {
			log.Printf("failed to send message %d: %v", i, err)
			continue
		}

		fmt.Printf("✓ Sent message %d\n", i)
	}

	// Example with async producer
	asyncConfig := messages.DefaultKafkaProducerConfig()
	asyncConfig.Brokers = []string{"localhost:9092"}
	asyncConfig.ErrorHandler = func(err *messages.ProducerError) {
		log.Printf("❌ Async error: topic=%s partition=%d error=%v",
			err.Topic, err.Partition, err.Err)
	}
	asyncConfig.SuccessHandler = func(msg *messages.ProducerSuccess) {
		fmt.Printf("✓ Async success: topic=%s partition=%d offset=%d\n",
			msg.Topic, msg.Partition, msg.Offset)
	}

	asyncProducer, err := messages.NewKafkaAsyncProducer(asyncConfig)
	if err != nil {
		log.Fatalf("failed to create async kafka producer: %v", err)
	}
	defer asyncProducer.Close()

	// Send messages asynchronously
	for i := 0; i < 10; i++ {
		msg := messages.NewSimpleMessage(
			"async-topic",
			[]byte(fmt.Sprintf("async-key-%d", i)),
			[]byte(fmt.Sprintf("Async message %d", i)),
		)

		err = asyncProducer.Send(msg)
		if err != nil {
			log.Printf("failed to queue async message %d: %v", i, err)
		}
	}

	// Wait a bit for async messages to be processed
	time.Sleep(2 * time.Second)

	fmt.Println("\n✓ All messages sent successfully!")
}

func runConsumer() {
	fmt.Println("=== Kafka Consumer ===")
	fmt.Println()

	// Configure Kafka consumer
	config := &messages.KafkaConsumerConfig{
		Brokers:           []string{brokerKafka},
		GroupID:           "demo-consumer-group-2",
		Topics:            []string{"demo-topic"},
		OffsetInitial:     messages.OffsetOldest, // Changed to OffsetNewest - only consume new messages
		SessionTimeout:    10 * time.Second,
		HeartbeatInterval: 3 * time.Second,
		EnableAutoCommit:  true,
		CommitInterval:    1 * time.Second,
		ErrorHandler: func(err error) {
			log.Printf("❌ Consumer error: %v", err)
		},
	}

	fmt.Println("Consumer Configuration:")
	fmt.Printf("  Brokers: %v\n", config.Brokers)
	fmt.Printf("  Group ID: %s\n", config.GroupID)
	fmt.Printf("  Topics: %v\n", config.Topics)
	fmt.Printf("  Offset Strategy: %s\n", config.OffsetInitial)
	fmt.Println()

	// Create consumer
	consumer, err := messages.NewKafkaConsumer(config)
	if err != nil {
		log.Fatalf("failed to create kafka consumer: %v", err)
	}
	defer consumer.Close()

	// Track message count
	var messageCount int

	// Register message handler
	consumer.Handle(func(ctx context.Context, msg messages.Message) error {
		messageCount++
		fmt.Printf("\n[Message #%d]\n", messageCount)
		fmt.Printf("  Topic:     %s\n", msg.Topic())
		fmt.Printf("  Partition: %d\n", msg.Partition())
		fmt.Printf("  Offset:    %d\n", msg.Offset())
		fmt.Printf("  Key:       %s\n", string(msg.Key()))
		fmt.Printf("  Value:     %s\n", string(msg.Value()))
		fmt.Printf("  Timestamp: %s\n", msg.Timestamp().Format(time.RFC3339))
		return nil
	})

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start consuming in background
	fmt.Println("Starting consumer...")
	fmt.Println("⚠️  NOTE: Consumer will only read NEW messages (OffsetNewest)")
	fmt.Println("⚠️  Run 'go run main.go producer' in another terminal to send messages")
	fmt.Println()
	fmt.Println("Waiting for messages... (Press Ctrl+C to exit)")
	fmt.Println()

	err = consumer.Start(ctx)
	if err != nil {
		log.Fatalf("failed to start consumer: %v", err)
	}

	fmt.Println("✓ Consumer is ready and listening...")
	fmt.Println()

	// Wait for shutdown signal
	<-sigCh
	fmt.Println("\n\nShutting down consumer...")
	cancel()

	// Close consumer (waits for goroutines to finish)
	if err := consumer.Close(); err != nil {
		log.Printf("error closing consumer: %v", err)
	}

	fmt.Printf("\n✓ Consumer stopped. Processed %d messages.\n", messageCount)
}
