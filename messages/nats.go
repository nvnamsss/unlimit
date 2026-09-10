package messages

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nvnamsss/unlimit/utility"
)

type NATsProducer struct {
	conn *nats.Conn
}

func NewNATsProducer(serverUrl string) (*NATsProducer, error) {
	// Connect to the first broker in the list
	nc, err := nats.Connect(serverUrl)
	if err != nil {
		return nil, err
	}
	return &NATsProducer{
		conn: nc,
	}, nil
}

func (p *NATsProducer) Send(message Message) error {
	data := message.Value()
	return p.conn.Publish(message.Topic(), data)
}

func (p *NATsProducer) Close() error {
	p.conn.Close()
	return nil
}

type NATsConsumer struct {
	serverUrl     string
	conn          *nats.Conn
	subscriptions map[string]*nats.Subscription
	handlers      map[string]ConsumeHandler
}

func NewNATsConsumer(serverUrl string) *NATsConsumer {
	return &NATsConsumer{
		serverUrl:     serverUrl,
		subscriptions: make(map[string]*nats.Subscription),
		handlers:      make(map[string]ConsumeHandler),
	}
}

func (c *NATsConsumer) Start(ctx context.Context) error {
	var err error
	c.conn, err = nats.Connect(c.serverUrl)
	if err != nil {
		return err
	}

	for k, v := range c.handlers {
		c.subscriptions[k], err = c.conn.Subscribe(k, func(msg *nats.Msg) {
			var message Message = NewNATsMessage(msg)
			if err := v(ctx, message); err != nil {
				return
			}
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// Handler function receives raw []byte and should process/convert to Message as needed
func (c *NATsConsumer) Subscribe(topic string, handler ConsumeHandler) error {
	if _, ok := c.subscriptions[topic]; ok {
		return ErrTopicRegistered
	}

	c.handlers[topic] = handler
	return nil
}

func (c *NATsConsumer) ListeningTopics() []string {
	return utility.Keys(c.handlers)
}

// Optional: Close the consumer connection
func (c *NATsConsumer) Close() error {
	for _, v := range c.subscriptions {
		v.Unsubscribe()
	}

	c.conn.Close()
	return nil
}
