package service

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/jackietana/ticket-platform/catalog-service/internal/domain"
	"github.com/jackietana/ticket-platform/catalog-service/internal/dto"
)

type EventRepository interface {
	CreateEvent(ctx context.Context, event *domain.Event) error
	GetEventByID(ctx context.Context, id uuid.UUID) (*domain.Event, error)
	ListEvents(ctx context.Context, categoryID *int64) ([]domain.Event, error)
}

type StorageRepository interface {
	UploadPoster(ctx context.Context, file io.Reader, size int64, contentType string) (string, error)
}

type CatalogService struct {
	eventRepo   EventRepository
	storageRepo StorageRepository
}

func NewService(eventRepo EventRepository, storageRepo StorageRepository) *CatalogService {
	return &CatalogService{eventRepo: eventRepo, storageRepo: storageRepo}
}

func (s *CatalogService) CreateEvent(ctx context.Context, input dto.CreateEventInput) (*domain.Event, error) {
	objectURL, err := s.storageRepo.UploadPoster(ctx, input.File, input.FileSize, input.ContentType)
	if err != nil {
		return nil, err
	}

	event := &domain.Event{
		ID:          uuid.New(),
		CategoryID:  &input.CategoryID,
		Title:       input.Title,
		Description: input.Description,
		PosterUrl:   objectURL,
		PriceCents:  input.PriceCents,
		TotalSlots:  input.TotalSlots,
		EventDate:   input.EventDate,
	}

	return event, s.eventRepo.CreateEvent(ctx, event)
}

func (s *CatalogService) GetEventByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	return s.eventRepo.GetEventByID(ctx, id)
}

func (s *CatalogService) ListEvents(ctx context.Context, categoryID *int64) ([]domain.Event, error) {
	return s.eventRepo.ListEvents(ctx, categoryID)
}
