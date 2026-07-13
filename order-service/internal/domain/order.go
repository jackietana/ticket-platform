package domain

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID         uuid.UUID `json:"uuid"`
	UserID     uuid.UUID `json:"user_id"`
	EventID    uuid.UUID `json:"event_id"`
	SlotsCount int       `json:"slots_count"`
	TotalPrice int       `json:"total_price"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type Ticket struct {
	Header  string  `json:"header"`
	OrderID string  `json:"order_id"`
	UserID  string  `json:"user_id"`
	EventID string  `json:"event_id"`
	Seats   int     `json:"seats"`
	Amount  float64 `json:"amount"`
	Status  string  `json:"status"`
	Bottom  string  `json:"bottom"`
}
