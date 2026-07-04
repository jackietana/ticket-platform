package dto

import (
	"io"
	"time"
)

type ListEventsInput struct {
	CategoryID int64 `form:"category_id" binding:"required"`
}

type CreateEventInput struct {
	CategoryID  int64     `form:"category_id" binding:"required"`
	Title       string    `form:"title" binding:"required"`
	Description string    `form:"description"`
	PriceCents  int64     `form:"price_cents" binding:"required"`
	TotalSlots  int32     `form:"total_slots" binding:"required"`
	EventDate   time.Time `form:"event_date" time_format:"2006-01-02T15:04:05Z07:00" binding:"required"`

	File        io.Reader
	FileSize    int64
	ContentType string
}

type CreateEventResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
