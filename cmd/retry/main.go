package main

import (
	"context"
	"log"
	"time"

	"github.com/voidforge-studios/unlimit/collections"
)

func main() {
	r := collections.NewRetrier(func(i int) time.Duration {
		if i >= 3 {
			return 0
		}
		return time.Duration(1<<uint(i)) * time.Second
	}, func(err error) bool {
		return err == nil
	})

	t := 0
	_ = r.Run(context.Background(), func(ctx context.Context) error {
		// Your operation here
		log.Println("Attempting operation...")
		// Simulate operation failure 5 times
		t++
		if t < 5 {
			log.Println("Operation failed")
			return context.DeadlineExceeded
		}
		return nil
	})

	log.Println("Operation completed")
}
