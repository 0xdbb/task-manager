package queue

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// QueueOptions defines configurable options for declaring a queue.
type QueueOptions struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}

// ConsumerOptions defines options for consuming messages.
type ConsumerOptions struct {
	QueueName string
	WorkerTag string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      amqp.Table
}

// QueueManager manages RabbitMQ connection and operations.
type QueueManager struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewQueueManager initializes a new QueueManager with exponential backoff
func NewQueueManager(amqpURL string) (*QueueManager, error) {
	var (
		conn       *amqp.Connection
		err        error
		maxRetries = 5
	)

	for attempt := 0; attempt <= maxRetries; attempt++ {
		conn, err = amqp.Dial(amqpURL)
		if err == nil {
			ch, err := conn.Channel()
			if err != nil {
				conn.Close()
				return nil, fmt.Errorf("failed to open channel: %w", err)
			}

			return &QueueManager{
				conn:    conn,
				channel: ch,
			}, nil
		}

		backoffDuration := time.Duration(1<<attempt) * time.Second
		fmt.Printf("RMQ Connection attempt %d failed; retrying in %v\n", attempt+1, backoffDuration)
		time.Sleep(backoffDuration)
	}

	return nil, fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %w", maxRetries, err)
}

// DeclareQueue declares a queue with the provided options.
func (qm *QueueManager) DeclareQueue(options QueueOptions) (amqp.Queue, error) {
	return qm.channel.QueueDeclare(
		options.Name,
		options.Durable,
		options.AutoDelete,
		options.Exclusive,
		options.NoWait,
		options.Args,
	)
}

// Set QoS (Prefetch Count) - Ensures workers only take tasks when free
func (qm *QueueManager) SetQos(prefetchCount, prefetchSize int, global bool) error {
	return qm.channel.Qos(
		prefetchCount,
		0,
		false,
	)
}

// Publish sends a message to the specified queue.
func (qm *QueueManager) Publish(queueName string, body []byte, priority uint8) error {
	return qm.channel.Publish(
		"",        // Default exchange
		queueName, // Routing key
		false,     // Mandatory
		false,     // Immediate
		amqp.Publishing{
			Headers:      map[string]interface{}{},
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Priority:     priority,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
}

// Consume starts consuming messages from the queue with given options.
func (qm *QueueManager) Consume(options ConsumerOptions) (<-chan amqp.Delivery, error) {
	return qm.channel.Consume(
		options.QueueName,
		options.WorkerTag,
		options.AutoAck,
		options.Exclusive,
		options.NoLocal,
		options.NoWait,
		options.Args,
	)
}

// Close cleans up the queue connection and channel.
func (qm *QueueManager) Close() {
	if qm.channel != nil {
		qm.channel.Close()
	}
	if qm.conn != nil {
		qm.conn.Close()
	}
}
