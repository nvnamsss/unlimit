package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/voidforge-studios/unlimit/scheduler"
)

func main() {
	s := scheduler.New()
	s.RegisterOneOffTask("task1", func() error {
		log.Printf("One-off task executed at %v", time.Now().Format(time.RFC3339))
		return nil
	}, time.Now().Add(5*time.Second))

	s.RegisterIntervalTask("task2", func() error {
		log.Printf("Interval task executed at %v", time.Now().Format(time.RFC3339))
		return nil
	}, 3*time.Second)

	s.RegisterCronTask("task3", func() error {
		log.Printf("Cron task executed at %v", time.Now().Format(time.RFC3339))
		return nil
	}, "*/1 * * * *") // every minute

	s.Start()
	defer s.Stop()

	// Keep the main function running to listen for messages
	// wait for a signal to exit
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)
	<-sc

	log.Println("Shutting down...")
}
