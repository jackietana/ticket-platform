package mq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func NewRabbitMQChannel(endpoint string) (*amqp.Channel, error) {
	conn, err := amqp.Dial(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	_, err = ch.QueueDeclare(
		"orders_paid_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	if err = ch.QueueBind("orders_paid_queue", "orders.paid", "amq.direct", false, nil); err != nil {
		return nil, fmt.Errorf("failed to bind exchange: %w", err)
	}

	return ch, nil
}
