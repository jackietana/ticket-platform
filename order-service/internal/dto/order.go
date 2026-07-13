package dto

type CreateOrderInput struct {
	UserID     string `json:"user_id" binding:"required"`
	EventID    string `json:"event_id" binding:"required"`
	SlotsCount int    `json:"slots_count" binding:"required"`
}

type CreateOrderResponse struct {
	ID         string  `json:"id"`
	Status     string  `json:"status"`
	TotalPrice float64 `json:"total_price"`
}

type PayOrderResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
