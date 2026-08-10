package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/nats-io/nats.go"
	"github.com/voidforge-studios/unlimit/messages"
)

func main() {
	// c := messages.NewNATsConsumer(nats.DefaultURL)

	// c.Subscribe("my-topic", func(ctx context.Context, message messages.Message) error {
	// 	log.Printf("Received message on topic %s: %s", message.Topic(), string(message.Value()))
	// 	return nil
	// })

	// go c.Start(context.Background())
	// defer c.Close()

	var (
		topic = "void.credit.transaction"
	)

	p, err := messages.NewNATsProducer(nats.DefaultURL)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 10; i++ {
		v := fmt.Sprintf("Message %d", i)
		if err := p.Send(messages.NewSimpleMessage(topic, []byte("my-key"), []byte(v))); err != nil {
			panic(err)
		}
	}

	log.Printf("Message sent successfully to topic %v", topic)

	// Keep the main function running to listen for messages
	// wait for a signal to exit
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)
	<-sc

	log.Println("Shutting down...")
}
