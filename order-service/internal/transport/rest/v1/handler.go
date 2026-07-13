package rest

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackietana/ticket-platform/api/gen/authv1"
	"github.com/jackietana/ticket-platform/order-service/internal/domain"
	"github.com/jackietana/ticket-platform/order-service/internal/dto"
	"google.golang.org/grpc/metadata"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID, eventID uuid.UUID, slotsCount int) (*domain.Order, error)
	PayOrder(ctx context.Context, orderID uuid.UUID) error
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*domain.Order, error)
}

type Handler struct {
	authClient authv1.AuthServiceClient
	service    OrderService
}

func NewHandler(auth authv1.AuthServiceClient, service OrderService) *Handler {
	return &Handler{authClient: auth, service: service}
}

func (h *Handler) InitRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	v1 := router.Group("/api/v1")
	{
		events := v1.Group("/orders")
		events.Use(h.AuthMiddleware())
		{
			events.POST("", h.CreateOrder)
			events.POST("/:id/pay", h.PayOrder)
		}
	}

	return router
}

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		authValues := strings.Split(authHeader, " ")

		if len(authValues) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid auth header"})
			return
		}

		if authValues[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid auth header"})
			return
		}

		if len(authValues[1]) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid auth token"})
			return
		}

		md := metadata.New(map[string]string{
			"client-ip": c.ClientIP(),
			"client-ua": c.GetHeader("User-Agent"),
		})

		grpcCtx := metadata.NewOutgoingContext(c.Request.Context(), md)
		resp, err := h.authClient.ValidateToken(grpcCtx, &authv1.ValidateTokenRequest{Token: authValues[1]})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "token is expired or invalid"})
			return
		}

		if !resp.IsValid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "token is invalid"})
			return
		}

		c.Set("user_id", resp.UserId)

		c.Next()
	}
}
