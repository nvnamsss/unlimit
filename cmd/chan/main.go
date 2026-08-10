package main

import (
	"log"
	"time"
)

func main() {
	c := make(chan bool)

	for i := 0; i < 5; i++ {
		go gogo(c, i)
	}

	// Give goroutines time to start and block on channel
	time.Sleep(100 * time.Millisecond)

	// Broadcast signal
	close(c)

	// Wait so goroutines can print
	time.Sleep(100 * time.Millisecond)
}

func gogo(c chan bool, id int) {
	<-c // block until channel is closed

	log.Printf("goroutine %d: hi mom", id)
}
