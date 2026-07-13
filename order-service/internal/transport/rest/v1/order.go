package rest

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackietana/ticket-platform/order-service/internal/dto"
	"github.com/jackietana/ticket-platform/order-service/internal/service"
)

func (h *Handler) CreateOrder(c *gin.Context) {
	var input dto.CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid input data"})
		return
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid user ID"})
		return
	}

	eventID, err := uuid.Parse(input.EventID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid event ID"})
		return
	}

	if input.SlotsCount <= 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid slots count"})
		return
	}

	order, err := h.service.CreateOrder(c.Request.Context(), userID, eventID, input.SlotsCount)
	if err != nil {
		if errors.Is(err, service.ErrNotEnoughTickets) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.CreateOrderResponse{
		ID:         order.ID.String(),
		Status:     order.Status,
		TotalPrice: float64(order.TotalPrice) / 100.0,
	})
}

func (h *Handler) PayOrder(c *gin.Context) {
	idStr := c.Param("id")
	if len(idStr) == 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "missing id parameter"})
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id parameter"})
		return
	}

	if err := h.service.PayOrder(c.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrPaymentRejected) {
			c.JSON(http.StatusPaymentRequired, dto.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.PayOrderResponse{Message: "order successfully paid"})
}
