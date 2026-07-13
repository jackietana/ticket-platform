package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	pb "github.com/jackietana/ticket-platform/api/gen/catalogv1"
	"github.com/jackietana/ticket-platform/order-service/internal/domain"
)

var (
	ErrNotEnoughTickets = errors.New("not enough tickets available")
	ErrPaymentRejected  = errors.New("payment rejected error")
)

type PsqlRepository interface {
	CreateOrder(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, orderID uuid.UUID) (*domain.Order, error)
	UpdateStatus(ctx context.Context, orderID uuid.UUID, status string) error
}

type RedisRepository interface {
	BookSlots(ctx context.Context, eventID, orderID string, totalSlots, slotsCount int, ttl time.Duration) (bool, error)
	ConfirmBooking(ctx context.Context, orderID string) error
	CancelBooking(ctx context.Context, eventID, orderID string, slotsCount int) error
}

type OrderMQProducer interface {
	PublishOrderPaid(ctx context.Context, order *domain.Order) error
}

type PaymentProvider interface {
	// ProcessPayment simulates a debit.
	// Returns the transaction ID from the payment system or an error (e.g., "insufficient funds")
	ProcessPayment(ctx context.Context, orderID uuid.UUID, amountCents int) (string, error)
}

type OrderService struct {
	catalogClient   pb.CatalogServiceClient
	repo            PsqlRepository
	cache           RedisRepository
	bookingTTL      time.Duration
	paymentProvider PaymentProvider
	mqProducer      OrderMQProducer
}

func NewService(catalog pb.CatalogServiceClient, repo PsqlRepository, cache RedisRepository, ttl time.Duration,
	payment PaymentProvider, mqClient OrderMQProducer) *OrderService {

	return &OrderService{
		catalogClient:   catalog,
		repo:            repo,
		cache:           cache,
		bookingTTL:      ttl,
		paymentProvider: payment,
		mqProducer:      mqClient,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID, eventID uuid.UUID, slotsCount int) (*domain.Order,
	error) {

	resp, err := s.catalogClient.GetEventAvailability(ctx, &pb.GetEventAvailabilityRequest{
		EventId: eventID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get event info: %w", err)
	}

	orderID := uuid.New()
	totalPrice := int(resp.PriceCents) * slotsCount

	booked, err := s.cache.BookSlots(ctx, eventID.String(), orderID.String(), int(resp.TotalSlots), slotsCount,
		s.bookingTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to book slots: %w", err)
	} else if !booked {
		return nil, ErrNotEnoughTickets
	}

	order := &domain.Order{
		ID:         orderID,
		UserID:     userID,
		EventID:    eventID,
		SlotsCount: slotsCount,
		TotalPrice: totalPrice,
		Status:     "pending",
	}

	if err = s.repo.CreateOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return order, nil
}

func (s *OrderService) PayOrder(ctx context.Context, orderID uuid.UUID) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order.Status != "pending" {
		return errors.New("invalid order status")
	}

	txID, err := s.paymentProvider.ProcessPayment(ctx, order.ID, order.TotalPrice)
	if err != nil {
		_ = s.repo.UpdateStatus(ctx, orderID, "failed")
		return ErrPaymentRejected
	}
	log.Printf("order %s successfully paid via tx %s", orderID, txID)

	if err := s.repo.UpdateStatus(ctx, orderID, "paid"); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	order.Status = "paid"
	if err := s.mqProducer.PublishOrderPaid(ctx, order); err != nil {
		// roll back the status to pending if broker crashes
		_ = s.repo.UpdateStatus(ctx, orderID, "pending")
		return fmt.Errorf("failed to publish order paid event: %w", err)
	}

	if err := s.cache.ConfirmBooking(ctx, orderID.String()); err != nil {
		log.Printf("warning: failed to clear redis booking: %v", err)
	}

	return nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, orderID uuid.UUID) (*domain.Order, error) {
	return s.repo.GetByID(ctx, orderID)
}
