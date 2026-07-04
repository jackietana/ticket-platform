package rest

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackietana/ticket-platform/catalog-service/internal/domain"
	"github.com/jackietana/ticket-platform/catalog-service/internal/dto"
)

type CatalogService interface {
	GetEventByID(ctx context.Context, id uuid.UUID) (*domain.Event, error)
	ListEvents(ctx context.Context, categoryID *int64) ([]domain.Event, error)
	CreateEvent(ctx context.Context, input dto.CreateEventInput) (*domain.Event, error)
}

type Handler struct {
	service CatalogService
}

func NewHandler(service CatalogService) *Handler {
	return &Handler{service}
}

func (h *Handler) Init() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	v1 := router.Group("/api/v1")
	{
		events := v1.Group("/events")
		{
			events.GET("/:id", h.GetEventByID)
			events.GET("", h.ListEvents)
			events.POST("", h.CreateEvent)
		}
	}

	return router
}
