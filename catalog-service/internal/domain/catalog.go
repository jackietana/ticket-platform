package domain

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Event struct {
	ID          uuid.UUID `json:"id"`
	CategoryID  *int64    `json:"category_id,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	PosterUrl   string    `json:"poster_url"`
	PriceCents  int64     `json:"price_cents"`
	TotalSlots  int32     `json:"total_slots"`
	EventDate   time.Time `json:"event_date"`
	CreatedAt   time.Time `json:"created_at"`
}
