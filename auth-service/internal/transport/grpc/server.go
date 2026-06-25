package grpcsrv

import (
	"context"

	pb "github.com/jackietana/ticket-platform/auth-service/internal/transport/grpc/gen"
	"github.com/jackietana/ticket-platform/auth-service/internal/transport/rest/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	authService rest.AuthService
}

func NewAuthServer(service rest.AuthService) *AuthServer {
	return &AuthServer{authService: service}
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	token := req.GetToken()
	if token == "" {
		return nil, status.Error(codes.InvalidArgument, "empty token provided")
	}

	// TODO: send client-ip in 'order-service'
	metaData, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get metadata")
	}

	var ipAddr string
	if data, ok := metaData["client-ip"]; ok && len(data) > 0 {
		ipAddr = data[0]
	}

	var userAgent string
	if data, ok := metaData["user-agent"]; ok && len(data) > 0 {
		userAgent = data[0]
	}

	userId, err := s.authService.ValidateSession(ctx, token, ipAddr, userAgent)
	if err != nil {
		return &pb.ValidateTokenResponse{IsValid: false}, nil
	}

	return &pb.ValidateTokenResponse{
		UserId:  userId,
		IsValid: true,
	}, nil
}
