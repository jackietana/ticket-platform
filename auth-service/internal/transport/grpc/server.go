package grpcsrv

import (
	"context"
	"errors"

	"github.com/jackietana/ticket-platform/auth-service/internal/service"
	pb "github.com/jackietana/ticket-platform/auth-service/internal/transport/grpc/gen"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	authService *service.AuthService
}

func NewAuthServer(service *service.AuthService) *AuthServer {
	return &AuthServer{authService: service}
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	token := req.GetToken()
	if token == "" {
		return nil, errors.New("empty token provided")
	}

	userId, err := s.authService.GetUserIdByToken(ctx, token)
	if err != nil {
		return &pb.ValidateTokenResponse{IsValid: false}, err
	}

	return &pb.ValidateTokenResponse{
		UserId:  userId,
		IsValid: true,
	}, nil
}
