package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackietana/ticket-platform/order-service/internal/domain"
)

var ErrOrderNotFound = errors.New("order not found")

type PsqlRepository struct {
	db *sql.DB
}

func NewPsqlRepository(db *sql.DB) *PsqlRepository {
	return &PsqlRepository{db}
}

func (r *PsqlRepository) CreateOrder(ctx context.Context, order *domain.Order) error {
	strExec := `INSERT INTO orders (user_id, event_id, slots_count, total_price, status) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, strExec, order.UserID, order.EventID, order.SlotsCount,
		order.TotalPrice, order.Status).Scan(&order.ID, &order.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	return nil
}

func (r *PsqlRepository) GetByID(ctx context.Context, orderID uuid.UUID) (*domain.Order, error) {
	var order domain.Order
	strExec := "SELECT id, user_id, event_id, slots_count, total_price, status, created_at from orders WHERE id=$1"

	err := r.db.QueryRowContext(ctx, strExec, orderID).Scan(&order.ID, &order.UserID, &order.EventID,
		&order.SlotsCount, &order.TotalPrice, &order.Status, &order.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order by id: %w", err)
	}

	return &order, nil
}

func (r *PsqlRepository) UpdateStatus(ctx context.Context, orderID uuid.UUID, status string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE orders SET status=$1 WHERE id=$2", status, orderID)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}
