package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackietana/ticket-platform/order-service/internal/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

type FileStorage interface {
	UploadTicket(ctx context.Context, filename string, content []byte) error
}

type OrderConsumer struct {
	channel *amqp.Channel
	storage FileStorage
}

func NewOrderConsumer(ch *amqp.Channel, fs FileStorage) *OrderConsumer {
	return &OrderConsumer{channel: ch, storage: fs}
}

func (c *OrderConsumer) StartListen(ctx context.Context) error {
	msgs, err := c.channel.Consume("orders_paid_queue", "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	go func() {
		for msg := range msgs {
			var order domain.Order
			if err := json.Unmarshal(msg.Body, &order); err != nil {
				log.Printf("ERROR: failed to unmarshal message body: %v. Content: %s\n", err, string(msg.Body))
				_ = msg.Nack(false, false)
				continue
			}

			ticketContent := domain.Ticket{
				Header:  "=== ELECTRONIC TICKET ===",
				OrderID: order.ID.String(),
				UserID:  order.UserID.String(),
				EventID: order.EventID.String(),
				Seats:   order.SlotsCount,
				Amount:  float64(order.TotalPrice / 100.0),
				Status:  order.Status,
				Bottom:  "=========================",
			}
			ticketBytes, err := json.Marshal(ticketContent)
			if err != nil {
				log.Printf("ERROR: failed to marshal ticket: %v\n", err)
				_ = msg.Nack(false, true)
				continue
			}

			filename := fmt.Sprintf("ticket_%s.json", order.ID.String())
			if err = c.storage.UploadTicket(ctx, filename, ticketBytes); err != nil {
				log.Printf("ERROR: failed to upload ticket to Minio: %v\n", err)
				_ = msg.Nack(false, true)
				continue
			}
			log.Printf("successfully generated and uploaded ticket: %s\n", filename)

			if err := msg.Ack(false); err != nil {
				log.Printf("ERROR: failed to ACK message for order %s: %v\n", order.ID.String(), err)
			}
		}
	}()

	<-ctx.Done()
	log.Println("shutting down worker...")

	if err := c.channel.Close(); err != nil {
		return fmt.Errorf("failed to close rabbitmq channel: %w", err)
	}

	return nil
}
