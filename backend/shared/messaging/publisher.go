package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQPublisher publishes events to RabbitMQ
type RabbitMQPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewRabbitMQPublisher creates a new RabbitMQ publisher
func NewRabbitMQPublisher(url string) (*RabbitMQPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &RabbitMQPublisher{
		conn:    conn,
		channel: ch,
	}, nil
}

// Publish publishes an event to RabbitMQ
func (p *RabbitMQPublisher) Publish(exchange, routingKey string, event interface{}) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return p.channel.PublishWithContext(
		context.Background(),
		exchange,   // exchange name
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // Survive broker restart
		},
	)
}

// Close closes the publisher connection
func (p *RabbitMQPublisher) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}
