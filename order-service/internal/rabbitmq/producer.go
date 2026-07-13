package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackietana/ticket-platform/order-service/internal/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

type OrderProducer struct {
	channel *amqp.Channel
}

func NewOrderProducer(ch *amqp.Channel) *OrderProducer {
	return &OrderProducer{channel: ch}
}

func (p *OrderProducer) PublishOrderPaid(ctx context.Context, order *domain.Order) error {
	body, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order to json: %w", err)
	}

	err = p.channel.PublishWithContext(
		ctx,
		"amq.direct",
		"orders.paid",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish rabbitmq message: %w", err)
	}

	return nil
}
