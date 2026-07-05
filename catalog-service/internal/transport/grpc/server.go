package grpcsrv

import (
	"context"

	"github.com/google/uuid"
	pb "github.com/jackietana/ticket-platform/catalog-service/internal/transport/grpc/gen"
	"github.com/jackietana/ticket-platform/catalog-service/internal/transport/rest/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CatalogServer struct {
	pb.UnimplementedCatalogServiceServer
	catalogService rest.CatalogService
}

func NewCatalogServer(service rest.CatalogService) *CatalogServer {
	return &CatalogServer{catalogService: service}
}

func (s *CatalogServer) GetEventAvailability(ctx context.Context, req *pb.GetEventAvailabilityRequest) (
	*pb.GetEventAvailabilityResponse, error) {

	eventID, err := uuid.Parse(req.EventId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid uuid format")
	}

	event, err := s.catalogService.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "event not found")
	}

	return &pb.GetEventAvailabilityResponse{
		EventId:    req.EventId,
		TotalSlots: event.TotalSlots,
		PriceCents: event.PriceCents,
	}, nil
}
