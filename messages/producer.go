package messages

type Producer interface {
	Send(message Message) error
	Close() error
}
