package grpc_client

import (
	"fmt"

	"github.com/jackietana/ticket-platform/api/gen/authv1"
	"github.com/jackietana/ticket-platform/api/gen/catalogv1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewAuthClient(endpoint string) (authv1.AuthServiceClient, error) {
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc channel: %w", err)
	}

	return authv1.NewAuthServiceClient(conn), nil
}

func NewCatalogClient(endpoint string) (catalogv1.CatalogServiceClient, error) {
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc channel: %w", err)
	}

	return catalogv1.NewCatalogServiceClient(conn), nil
}
