package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/voidforge-studios/unlimit/messages"
)

const (
	defaultPath      = "/dev/shm/unlimit_hft_queue"
	defaultSlotCount = 1024
)

func main() {
	mode := flag.String("mode", "", "Mode: 'producer' or 'consumer' (required)")
	path := flag.String("path", defaultPath, "Shared memory file path")
	slotCount := flag.Uint64("slots", defaultSlotCount, "Number of slots (must be power of 2)")
	messageCount := flag.Int("count", 1000000, "Number of messages to send (producer only)")
	symbol := flag.String("symbol", "BTCUSDT", "Trading symbol")
	batchSize := flag.Int("batch", 1000, "Messages per batch for stats reporting")

	flag.Parse()

	if *mode == "" {
		fmt.Println("HFT Shared Memory Demo")
		fmt.Println()
		fmt.Println("Usage: hft_shared_memory -mode=<producer|consumer> [options]")
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # Start producer (creates shared memory)")
		fmt.Println("  ./hft_shared_memory -mode=producer -count=1000000")
		fmt.Println()
		fmt.Println("  # Start consumer (opens existing shared memory)")
		fmt.Println("  ./hft_shared_memory -mode=consumer")
		fmt.Println()
		fmt.Println("For lowest tail latency, run with:")
		fmt.Println("  sudo nice -n -20 GOGC=off ./hft_shared_memory -mode=producer")
		fmt.Println()
		fmt.Println("Or with CPU isolation (requires kernel boot params: isolcpus=2,3):")
		fmt.Println("  sudo taskset -c 2 GOGC=off ./hft_shared_memory -mode=producer")
		os.Exit(1)
	}

	// Set GOMAXPROCS to use all cores
	runtime.GOMAXPROCS(runtime.NumCPU())

	switch *mode {
	case "producer":
		runProducer(*path, *slotCount, *messageCount, *symbol, *batchSize)
	case "consumer":
		runConsumer(*path, *slotCount, *batchSize)
	default:
		fmt.Printf("Invalid mode: %s\n", *mode)
		os.Exit(1)
	}
}

func runProducer(path string, slotCount uint64, messageCount int, symbol string, batchSize int) {
	fmt.Println("=== HFT Producer ===")
	fmt.Printf("Path: %s\n", path)
	fmt.Printf("Slots: %d\n", slotCount)
	fmt.Printf("Messages: %d\n", messageCount)
	fmt.Printf("Symbol: %s\n", symbol)
	fmt.Println()

	producer, err := messages.NewHFTProducer(&messages.HFTProducerConfig{
		Path:      path,
		SlotCount: slotCount,
		Symbol:    symbol,
		AccountID: 12345,
		Create:    true,
	})
	if err != nil {
		fmt.Printf("Failed to create producer: %v\n", err)
		os.Exit(1)
	}
	defer producer.Close()

	fmt.Println("Producer created. Start consumer in another terminal:")
	fmt.Printf("  ./hft_shared_memory -mode=consumer -path=%s\n", path)
	fmt.Println()
	fmt.Println("⚠️  IMPORTANT: Start the consumer BEFORE pressing Enter!")
	fmt.Println("   The buffer only holds", slotCount, "messages.")
	fmt.Println("   Without a consumer, the producer will timeout after buffer is full.")
	fmt.Println()
	fmt.Println("Press Enter to start sending messages...")
	fmt.Scanln()

	// Lock thread for consistent latency
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var totalLatency int64
	var minLatency int64 = int64(^uint64(0) >> 1)
	var maxLatency int64
	var timeoutCount int
	var successCount int

	startTime := time.Now()
	batchStart := time.Now()

	for i := 0; i < messageCount; i++ {
		price := int64(50000_00 + (i % 1000))     // Price in cents
		quantity := int64(1_00000000 + (i % 100)) // Quantity in satoshi
		side := uint8(i % 2)
		orderType := uint8(i % 3)
		orderID := uint32(i + 1)

		sendStart := time.Now()
		err := producer.SendOrderWait(price, quantity, side, orderType, orderID, time.Second)
		latency := time.Since(sendStart).Nanoseconds()

		if err != nil {
			timeoutCount++
			if timeoutCount <= 5 {
				fmt.Printf("Buffer full (timeout #%d) - is consumer running?\n", timeoutCount)
			} else if timeoutCount == 6 {
				fmt.Println("Suppressing further timeout messages...")
			}
			continue
		}

		successCount++
		totalLatency += latency
		if latency < minLatency {
			minLatency = latency
		}
		if latency > maxLatency {
			maxLatency = latency
		}

		if successCount%batchSize == 0 {
			batchDuration := time.Since(batchStart)
			rate := float64(batchSize) / batchDuration.Seconds()
			avgLatency := float64(totalLatency) / float64(successCount)
			fmt.Printf("[%d/%d] Rate: %.0f msg/s, Avg latency: %.0f ns, Min: %d ns, Max: %d ns\n",
				successCount, messageCount, rate, avgLatency, minLatency, maxLatency)
			batchStart = time.Now()
		}
	}

	elapsed := time.Since(startTime)
	avgLatency := float64(0)
	if successCount > 0 {
		avgLatency = float64(totalLatency) / float64(successCount)
	}
	rate := float64(successCount) / elapsed.Seconds()

	fmt.Println()
	fmt.Println("=== Producer Summary ===")
	fmt.Printf("Messages sent: %d\n", successCount)
	fmt.Printf("Messages timed out: %d\n", timeoutCount)
	fmt.Printf("Total time: %v\n", elapsed)
	fmt.Printf("Throughput: %.0f msg/s\n", rate)
	if successCount > 0 {
		fmt.Printf("Avg publish latency: %.0f ns\n", avgLatency)
		fmt.Printf("Min publish latency: %d ns\n", minLatency)
		fmt.Printf("Max publish latency: %d ns\n", maxLatency)
	}
	if timeoutCount > 0 {
		fmt.Println()
		fmt.Println("⚠️  Timeouts occurred because buffer was full.")
		fmt.Println("   Make sure consumer is running before starting producer.")
	}
}

func runConsumer(path string, slotCount uint64, batchSize int) {
	fmt.Println("=== HFT Consumer ===")
	fmt.Printf("Path: %s\n", path)
	fmt.Printf("Slots: %d\n", slotCount)
	fmt.Println()

	// Check if shared memory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("Shared memory not found: %s\n", path)
		fmt.Println("Start the producer first")
		os.Exit(1)
	}

	consumer, err := messages.NewHFTConsumer(&messages.HFTConsumerConfig{
		Path:      path,
		SlotCount: slotCount,
	})
	if err != nil {
		fmt.Printf("Failed to create consumer: %v\n", err)
		os.Exit(1)
	}
	defer consumer.Close()

	fmt.Println("Consumer connected. Busy-spinning for messages...")
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	var messageCount atomic.Uint64
	var totalLatency atomic.Int64
	var minLatency atomic.Int64
	var maxLatency atomic.Int64
	minLatency.Store(int64(^uint64(0) >> 1))

	startTime := time.Now()
	lastReport := time.Now()
	var lastCount uint64

	// Signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Message handler
	consumer.Handle(func(msg *messages.HFTMessage) error {
		latency := messages.MeasureLatency(msg)

		messageCount.Add(1)
		totalLatency.Add(latency)

		// Update min/max atomically
		for {
			current := minLatency.Load()
			if latency >= current || minLatency.CompareAndSwap(current, latency) {
				break
			}
		}
		for {
			current := maxLatency.Load()
			if latency <= current || maxLatency.CompareAndSwap(current, latency) {
				break
			}
		}

		return nil
	})

	// Start consumer in a goroutine
	go consumer.Run()

	// Stats reporting loop
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigCh:
			goto shutdown
		case <-ticker.C:
			count := messageCount.Load()
			elapsed := time.Since(lastReport)
			rate := float64(count-lastCount) / elapsed.Seconds()
			avgLatency := float64(0)
			if count > 0 {
				avgLatency = float64(totalLatency.Load()) / float64(count)
			}
			fmt.Printf("Messages: %d, Rate: %.0f msg/s, Avg e2e latency: %.0f ns, Min: %d ns, Max: %d ns\n",
				count, rate, avgLatency, minLatency.Load(), maxLatency.Load())
			lastReport = time.Now()
			lastCount = count
		}
	}

shutdown:
	consumer.Stop()
	elapsed := time.Since(startTime)
	count := messageCount.Load()
	avgLatency := float64(0)
	if count > 0 {
		avgLatency = float64(totalLatency.Load()) / float64(count)
	}

	fmt.Println()
	fmt.Println("=== Consumer Summary ===")
	fmt.Printf("Total messages: %d\n", count)
	fmt.Printf("Total time: %v\n", elapsed)
	if elapsed.Seconds() > 0 {
		fmt.Printf("Throughput: %.0f msg/s\n", float64(count)/elapsed.Seconds())
	}
	fmt.Printf("Avg end-to-end latency: %.0f ns\n", avgLatency)
	fmt.Printf("Min end-to-end latency: %d ns\n", minLatency.Load())
	fmt.Printf("Max end-to-end latency: %d ns\n", maxLatency.Load())
}
