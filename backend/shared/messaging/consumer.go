package messaging

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQConsumer consumes events from RabbitMQ
type RabbitMQConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewRabbitMQConsumer creates a new RabbitMQ consumer
func NewRabbitMQConsumer(url, queueName string) (*RabbitMQConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare queue (idempotent)
	_, err = ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // auto-delete
		false,     // exclusive
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Set prefetch count (process 1 message at a time)
	if err := ch.Qos(1, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	return &RabbitMQConsumer{
		conn:    conn,
		channel: ch,
	}, nil
}

// Consume consumes messages from the queue
func (c *RabbitMQConsumer) Consume(queueName string, handler func([]byte) error) error {
	msgs, err := c.channel.Consume(
		queueName,
		"",    // consumer tag
		false, // auto-ack (we'll ack manually after processing)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	forever := make(chan bool)

	go func() {
		for msg := range msgs {
			if err := handler(msg.Body); err != nil {
				// Negative ack - requeue the message
				msg.Nack(false, true)
			} else {
				// Positive ack - message processed successfully
				msg.Ack(false)
			}
		}
	}()

	<-forever
	return nil
}

// Close closes the consumer connection
func (c *RabbitMQConsumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
