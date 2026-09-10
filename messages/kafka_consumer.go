package messages

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/nvnamsss/unlimit/logger"
)

var (
	// ErrKafkaConsumerClosed is returned when operations are attempted on a closed consumer
	ErrKafkaConsumerClosed = errors.New("kafka consumer is closed")
	// ErrKafkaNoHandlers is returned when starting consumer without handlers
	ErrKafkaNoHandlers = errors.New("no handlers registered")
)

// Offset reset strategy
const (
	OffsetNewest = "newest" // start from newest offset
	OffsetOldest = "oldest" // start from oldest offset
)

// KafkaConsumerConfig holds configuration for Kafka consumer
type KafkaConsumerConfig struct {
	// Brokers is the list of Kafka broker addresses
	Brokers []string

	// GroupID is the consumer group identifier
	GroupID string

	// Topics is the list of topics to subscribe to
	Topics []string

	// OffsetInitial sets where to start consuming (newest or oldest)
	OffsetInitial string

	// SessionTimeout is the timeout for session keepalive
	SessionTimeout time.Duration

	// RebalanceTimeout is the maximum time for rebalance operations
	RebalanceTimeout time.Duration

	// HeartbeatInterval is the interval for sending heartbeats
	HeartbeatInterval time.Duration

	// MaxProcessingTime is the maximum time for processing a message
	MaxProcessingTime time.Duration

	// CommitInterval is the interval for auto-committing offsets
	CommitInterval time.Duration

	// EnableAutoCommit enables automatic offset commits
	EnableAutoCommit bool

	// IsolationLevel sets read committed/uncommitted (for transactional reads)
	IsolationLevel string

	// ErrorHandler is called when consumer errors occur (optional)
	ErrorHandler func(error)
}

// DefaultKafkaConsumerConfig returns a production-ready consumer configuration
func DefaultKafkaConsumerConfig() *KafkaConsumerConfig {
	return &KafkaConsumerConfig{
		OffsetInitial:     OffsetNewest,
		SessionTimeout:    10 * time.Second,
		RebalanceTimeout:  60 * time.Second,
		HeartbeatInterval: 3 * time.Second,
		MaxProcessingTime: 5 * time.Minute,
		CommitInterval:    1 * time.Second,
		EnableAutoCommit:  true,
		IsolationLevel:    "read_uncommitted",
	}
}

// mapOffsetInitial converts string offset to Sarama offset
func mapOffsetInitial(offset string) int64 {
	switch offset {
	case OffsetNewest:
		return sarama.OffsetNewest
	case OffsetOldest:
		return sarama.OffsetOldest
	default:
		return sarama.OffsetNewest
	}
}

// buildSaramaConsumerConfig creates Sarama config from KafkaConsumerConfig
func buildSaramaConsumerConfig(config *KafkaConsumerConfig) *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V3_0_0_0
	// Consumer settings
	cfg.Consumer.Group.Rebalance.Timeout = config.RebalanceTimeout
	cfg.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	cfg.Consumer.Group.Session.Timeout = config.SessionTimeout
	cfg.Consumer.Group.Heartbeat.Interval = config.HeartbeatInterval
	cfg.Consumer.MaxProcessingTime = config.MaxProcessingTime
	cfg.Consumer.Offsets.Initial = mapOffsetInitial(config.OffsetInitial)
	cfg.Consumer.Offsets.AutoCommit.Enable = config.EnableAutoCommit
	cfg.Consumer.Offsets.AutoCommit.Interval = config.CommitInterval

	// Isolation level for transactional reads
	if config.IsolationLevel == "read_committed" {
		cfg.Consumer.IsolationLevel = sarama.ReadCommitted
	} else {
		cfg.Consumer.IsolationLevel = sarama.ReadUncommitted
	}

	cfg.Consumer.Return.Errors = true

	return cfg
}

type KafkaConsumer struct {
	handlers []ConsumeHandler
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *KafkaConsumer) Setup(session sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	logger.Infof("kafka consumer: session started, member=%s generation=%d",
		session.MemberID(), session.GenerationID())
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *KafkaConsumer) Cleanup(session sarama.ConsumerGroupSession) error {

	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *KafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/main/consumer_group.go#L27-L2
	logger.Infof("kafka consumer: claiming topic=%s partition=%d offset=%d",
		claim.Topic(), claim.Partition(), claim.InitialOffset())

	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				continue
			}

			logger.Infof("kafka consumer: received message topic=%s partition=%d offset=%d",
				message.Topic, message.Partition, message.Offset)

			msg := &SaramaMessage{
				Message: message,
			}

			// Invoke all registered handlers for this message
			for _, handler := range consumer.handlers {
				if err := handler(session.Context(), msg); err != nil {
					logger.Errorf("kafka consumer handler error: topic=%s partition=%d offset=%d error=%v",
						message.Topic, message.Partition, message.Offset, err)
				}
			}

			session.MarkMessage(message, "")
		// Should return when `session.Context()` is done.
		// If not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance. see:
		// https://github.com/Shopify/sarama/issues/1192
		case <-session.Context().Done():
			logger.Infof("kafka consumer: claim context done for topic=%s partition=%d", claim.Topic(), claim.Partition())
			return nil
		}
	}
}

type KafkaConsumerGroup struct {
	consumer sarama.ConsumerGroup
	handlers []ConsumeHandler
	config   *KafkaConsumerConfig
	cancel   context.CancelFunc
	closed   bool
}

func (c *KafkaConsumerGroup) Handle(fn ConsumeHandler) {
	c.handlers = append(c.handlers, fn)
}

func (c *KafkaConsumerGroup) Start(ctx context.Context) error {
	if c.closed {
		return ErrKafkaConsumerClosed
	}
	if len(c.handlers) == 0 {
		return ErrKafkaNoHandlers
	}

	ctx, c.cancel = context.WithCancel(ctx)

	// Start error handler goroutine
	go func() {
		for err := range c.consumer.Errors() {
			if c.config.ErrorHandler != nil {
				c.config.ErrorHandler(fmt.Errorf("consumer group error: %w", err))
			} else {
				logger.Errorf("kafka consumer group error: %v", err)
			}
		}
	}()
	consumer := &KafkaConsumer{
		handlers: c.handlers, // Pass all handlers
	}

	// Start single consume loop that will invoke all handlers
	go func() {
		logger.Infof("kafka consumer: starting consume loop for group=%s topics=%v", c.config.GroupID, c.config.Topics)
		for {

			logger.Infof("kafka consumer: calling Consume() for topics=%v", c.config.Topics)
			// Consume will block until session ends or context is cancelled
			if err := c.consumer.Consume(ctx, c.config.Topics, consumer); err != nil {
				logger.Errorf("Error from consumer: %v", err)
			}
		}
	}()

	return nil
}

func (c *KafkaConsumerGroup) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true

	// Cancel context to stop consumption
	if c.cancel != nil {
		c.cancel()
	}

	// Close the consumer group
	return c.consumer.Close()
}

// NewKafkaConsumer creates a new Kafka consumer group with proper validation and defaults
func NewKafkaConsumer(config *KafkaConsumerConfig) (*KafkaConsumerGroup, error) {
	if config == nil {
		config = DefaultKafkaConsumerConfig()
	}

	// Validate required fields
	if len(config.Brokers) == 0 {
		return nil, fmt.Errorf("%w: no brokers specified", ErrKafkaInvalidConfig)
	}
	if config.GroupID == "" {
		return nil, fmt.Errorf("%w: group ID is required", ErrKafkaInvalidConfig)
	}
	if len(config.Topics) == 0 {
		return nil, fmt.Errorf("%w: no topics specified", ErrKafkaInvalidConfig)
	}

	// Apply defaults for unset fields
	if config.SessionTimeout <= 0 {
		config.SessionTimeout = 10 * time.Second
	}
	if config.RebalanceTimeout <= 0 {
		config.RebalanceTimeout = 60 * time.Second
	}
	if config.HeartbeatInterval <= 0 {
		config.HeartbeatInterval = 3 * time.Second
	}
	if config.MaxProcessingTime <= 0 {
		config.MaxProcessingTime = 5 * time.Minute
	}
	if config.CommitInterval <= 0 {
		config.CommitInterval = 1 * time.Second
	}
	if config.OffsetInitial == "" {
		config.OffsetInitial = OffsetNewest
	}
	if config.IsolationLevel == "" {
		config.IsolationLevel = "read_uncommitted"
	}

	saramaConfig := buildSaramaConsumerConfig(config)

	consumerGroup, err := sarama.NewConsumerGroup(config.Brokers, config.GroupID, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer group: %w", err)
	}

	return &KafkaConsumerGroup{
		consumer: consumerGroup,
		config:   config,
		handlers: make([]ConsumeHandler, 0),
		closed:   false,
	}, nil
}
