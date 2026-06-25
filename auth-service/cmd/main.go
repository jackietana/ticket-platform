package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackietana/ticket-platform/auth-service/internal/config"
	"github.com/jackietana/ticket-platform/auth-service/internal/repository"
	"github.com/jackietana/ticket-platform/auth-service/internal/service"
	grpcsrv "github.com/jackietana/ticket-platform/auth-service/internal/transport/grpc"
	pb "github.com/jackietana/ticket-platform/auth-service/internal/transport/grpc/gen"
	"github.com/jackietana/ticket-platform/auth-service/internal/transport/rest/v1"
	"github.com/jackietana/ticket-platform/auth-service/pkg/cache"
	"github.com/jackietana/ticket-platform/auth-service/pkg/hash"
	"github.com/jackietana/ticket-platform/auth-service/pkg/psql"
	"google.golang.org/grpc"
)

// @title auth-service
// @version 1.0
// @description auth-service providing authentication via PostgreSQL and Redis.

// @host localhost:8080
// @BasePath /

func main() {
	// REST deps
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("error creating config: %v", err)
	}

	hasher := hash.NewSHA1Hasher(cfg.Salt)
	cacheDB := cache.NewRedisConnection("localhost:6379")

	repoDB, err := psql.NewPostgresConnection(cfg.DB)
	if err != nil {
		log.Fatalf("error connecting to db: %v", err)
	}

	repo := repository.NewRepository(repoDB)
	cacher := repository.NewCache(cacheDB)
	authService := service.NewAuthService(hasher, repo, cacher)
	handler := rest.NewHandler(authService)

	router := handler.Init()
	httpSrv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// gRPC deps
	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, grpcsrv.NewAuthServer(authService))

	go func() {
		log.Println("gRPC server started on port 9000")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	go func() {
		log.Println("REST server started on port 8080")
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("error starting server: %v", err)
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Println("received stop signal, closing services...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err = httpSrv.Shutdown(ctx); err != nil {
		log.Fatalf("error stoping http server: %v", err)
	}

	grpcServer.GracefulStop()
	log.Println("services successfully stoped")
}
