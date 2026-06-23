package grpcsrv

import (
	"context"
	"errors"

	pb "github.com/jackietana/ticket-platform/auth-service/internal/transport/grpc/gen"
	"github.com/jackietana/ticket-platform/auth-service/internal/transport/rest/v1"
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
		return nil, errors.New("empty token provided")
	}

	userId, err := s.authService.GetUserIdByToken(ctx, token)
	if err != nil {
		return &pb.ValidateTokenResponse{IsValid: false}, nil
	}

	return &pb.ValidateTokenResponse{
		UserId:  userId,
		IsValid: true,
	}, nil
}
