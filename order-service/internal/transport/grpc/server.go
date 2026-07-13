package grpcsrv

import (
	"context"
	"errors"

	"github.com/google/uuid"
	pb "github.com/jackietana/ticket-platform/api/gen/orderv1"
	"github.com/jackietana/ticket-platform/order-service/internal/repository"
	"github.com/jackietana/ticket-platform/order-service/internal/transport/rest/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderServer struct {
	pb.UnimplementedOrderServiceServer
	service rest.OrderService
}

func NewOrderServer(service rest.OrderService) *OrderServer {
	return &OrderServer{service: service}
}

func (s *OrderServer) GetOrderStatus(ctx context.Context, req *pb.GetOrderStatusRequest) (*pb.GetOrderStatusResponse,
	error) {

	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid uuid format")
	}

	order, err := s.service.GetOrderByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return nil, status.Error(codes.NotFound, "order not found")
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetOrderStatusResponse{
		OrderId: order.ID.String(),
		Status:  order.Status,
	}, nil
}
