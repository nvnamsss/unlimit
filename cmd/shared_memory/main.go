package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/voidforge-studios/unlimit/messages"
)

const (
	defaultPath      = "/dev/shm/unlimit_queue"
	defaultSlotSize  = 4096
	defaultSlotCount = 1024
)

func main() {
	// Command line flags
	mode := flag.String("mode", "", "Mode: 'producer' or 'consumer' (required)")
	path := flag.String("path", defaultPath, "Shared memory file path")
	slotSize := flag.Int("slot-size", defaultSlotSize, "Size of each message slot in bytes")
	slotCount := flag.Int("slot-count", defaultSlotCount, "Number of slots in the ring buffer")
	messageCount := flag.Int("count", 10, "Number of messages to send (producer only)")
	interval := flag.Duration("interval", 100*time.Nanosecond, "Interval between messages (producer) or poll interval (consumer)")

	flag.Parse()

	if *mode == "" {
		fmt.Println("Usage: shared_memory -mode=<producer|consumer> [options]")
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # Start producer (creates shared memory)")
		fmt.Println("  ./shared_memory -mode=producer -count=100")
		fmt.Println()
		fmt.Println("  # Start consumer (opens existing shared memory)")
		fmt.Println("  ./shared_memory -mode=consumer")
		os.Exit(1)
	}

	switch *mode {
	case "producer":
		runProducer(*path, *slotSize, *slotCount, *messageCount, *interval)
	case "consumer":
		runConsumer(*path, *slotSize, *slotCount, *interval)
	default:
		fmt.Printf("Invalid mode: %s. Use 'producer' or 'consumer'\n", *mode)
		os.Exit(1)
	}
}

func runProducer(path string, slotSize, slotCount, messageCount int, interval time.Duration) {
	fmt.Println("=== Shared Memory Producer ===")
	fmt.Printf("Path: %s\n", path)
	fmt.Printf("Slot Size: %d bytes\n", slotSize)
	fmt.Printf("Slot Count: %d\n", slotCount)
	fmt.Printf("Messages to send: %d\n", messageCount)
	fmt.Printf("Interval: %v\n", interval)
	fmt.Println()

	// Create producer (creates the shared memory)
	producer, err := messages.NewSharedMemoryProducer(&messages.SharedMemoryProducerConfig{
		Path:      path,
		Topic:     "demo-topic",
		Partition: 0,
		SlotSize:  slotSize,
		SlotCount: slotCount,
		Create:    true, // Producer creates the shared memory
	})
	if err != nil {
		fmt.Printf("Failed to create producer: %v\n", err)
		os.Exit(1)
	}
	defer producer.Close()

	fmt.Println("Producer created. Shared memory initialized.")
	fmt.Println("Start the consumer in another terminal with:")
	fmt.Printf("  ./shared_memory -mode=consumer -path=%s\n", path)
	fmt.Println()

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Send messages
	for i := 1; i <= messageCount; i++ {
		select {
		case <-ctx.Done():
			fmt.Println("Producer stopped by signal")
			return
		default:
		}

		key := []byte(fmt.Sprintf("key-%d", i))
		value := []byte(fmt.Sprintf("Message #%d sent at %s", i, time.Now().Format(time.RFC3339Nano)))

		err := producer.SendRaw(key, value)
		if err != nil {
			if err == messages.ErrSharedMemoryFull {
				fmt.Printf("[%d/%d] Buffer full, waiting...\n", i, messageCount)
				time.Sleep(interval * 2)
				i-- // Retry
				continue
			}
			fmt.Printf("[%d/%d] Failed to send: %v\n", i, messageCount, err)
			continue
		}

		fmt.Printf("[%d/%d] Sent: key=%s\n", i, messageCount, key)
		time.Sleep(interval)
	}

	fmt.Println()
	fmt.Println("All messages sent!")
	fmt.Println("Press Ctrl+C to exit and cleanup shared memory, or let consumer drain the buffer.")

	// Wait for signal
	<-sigCh
	cancel()
	fmt.Println("\nProducer shutdown complete.")
}

func runConsumer(path string, slotSize, slotCount int, pollInterval time.Duration) {
	fmt.Println("=== Shared Memory Consumer ===")
	fmt.Printf("Path: %s\n", path)
	fmt.Printf("Slot Size: %d bytes\n", slotSize)
	fmt.Printf("Slot Count: %d\n", slotCount)
	fmt.Printf("Poll Interval: %v\n", pollInterval)
	fmt.Println()

	// Check if shared memory file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("Shared memory file does not exist: %s\n", path)
		fmt.Println("Start the producer first with:")
		fmt.Printf("  ./shared_memory -mode=producer -path=%s\n", path)
		os.Exit(1)
	}

	// Create consumer (opens existing shared memory)
	consumer, err := messages.NewSharedMemoryConsumer(&messages.SharedMemoryConsumerConfig{
		Path:         path,
		SlotSize:     slotSize,
		SlotCount:    slotCount,
		PollInterval: pollInterval,
		Create:       false, // Consumer opens existing shared memory
	})
	if err != nil {
		fmt.Printf("Failed to create consumer: %v\n", err)
		os.Exit(1)
	}
	defer consumer.Close()

	fmt.Println("Consumer connected to shared memory.")
	fmt.Println("Waiting for messages... (Press Ctrl+C to exit)")
	fmt.Println()

	// Track statistics
	var messageReceived int
	startTime := time.Now()

	// Register message handler
	consumer.Handle(func(ctx context.Context, msg messages.Message) error {
		messageReceived++
		fmt.Printf("[%d] Received message:\n", messageReceived)
		fmt.Printf("    Topic:     %s\n", msg.Topic())
		fmt.Printf("    Partition: %d\n", msg.Partition())
		fmt.Printf("    Offset:    %d\n", msg.Offset())
		fmt.Printf("    Key:       %s\n", string(msg.Key()))
		fmt.Printf("    Value:     %s\n", string(msg.Value()))
		fmt.Printf("    Timestamp: %s\n", msg.Timestamp().Format(time.RFC3339Nano))
		// calculate latency
		fmt.Printf("	Latency:   %v\n", time.Since(msg.Timestamp()))
		fmt.Println()
		return nil
	})

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start consuming
	consumer.Start(ctx)

	// Wait for shutdown signal
	<-sigCh
	fmt.Println("\nShutting down consumer...")
	cancel()

	// Print statistics
	elapsed := time.Since(startTime)
	fmt.Println()
	fmt.Println("=== Consumer Statistics ===")
	fmt.Printf("Messages received: %d\n", messageReceived)
	fmt.Printf("Runtime: %v\n", elapsed)
	if messageReceived > 0 {
		fmt.Printf("Average rate: %.2f msg/sec\n", float64(messageReceived)/elapsed.Seconds())
	}
	fmt.Println("Consumer shutdown complete.")
}
