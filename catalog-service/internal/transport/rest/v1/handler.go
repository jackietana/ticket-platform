package rest

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackietana/ticket-platform/catalog-service/internal/domain"
	"github.com/jackietana/ticket-platform/catalog-service/internal/dto"
	pb "github.com/jackietana/ticket-platform/catalog-service/internal/transport/grpc/gen"
)

type CatalogService interface {
	GetEventByID(ctx context.Context, id uuid.UUID) (*domain.Event, error)
	ListEvents(ctx context.Context, categoryID *int64) ([]domain.Event, error)
	CreateEvent(ctx context.Context, input dto.CreateEventInput) (*domain.Event, error)
}

type Handler struct {
	authClient pb.AuthServiceClient
	service    CatalogService
}

func NewHandler(client pb.AuthServiceClient, service CatalogService) *Handler {
	return &Handler{authClient: client, service: service}
}

func (h *Handler) InitRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	v1 := router.Group("/api/v1")
	{
		events := v1.Group("/events")
		events.Use(h.AuthMiddleware())
		{
			events.GET("/:id", h.GetEventByID)
			events.GET("", h.ListEvents)
			events.POST("", h.CreateEvent)
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

		c.Set("client_ip", c.ClientIP())
		c.Set("user-agent", c.GetHeader("User-Agent"))

		resp, err := h.authClient.ValidateToken(c.Request.Context(), &pb.ValidateTokenRequest{Token: authValues[1]})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "token is invalid"})
			return
		}

		if !resp.IsValid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "token is expired"})
			return
		}

		c.Set("user_id", resp.UserId)

		c.Next()
	}
}
