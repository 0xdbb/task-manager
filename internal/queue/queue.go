package queue

import (
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

// NewQueueManager initializes a new QueueManager.
func NewQueueManager(amqpURL string) (*QueueManager, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &QueueManager{
		conn:    conn,
		channel: ch,
	}, nil
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

// Publish sends a message to the specified queue.
func (qm *QueueManager) Publish(queueName string, body []byte, priority uint8) error {
	return qm.channel.Publish(
		"",        // Default exchange
		queueName, // Routing key
		false,     // Mandatory
		false,     // Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Priority:    priority,
			Timestamp:   time.Now(),
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
