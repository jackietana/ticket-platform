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

	"github.com/jackietana/ticket-platform/api/gen/authv1"
	pb "github.com/jackietana/ticket-platform/api/gen/catalogv1"
	"github.com/jackietana/ticket-platform/catalog-service/internal/config"
	"github.com/jackietana/ticket-platform/catalog-service/internal/repository"
	"github.com/jackietana/ticket-platform/catalog-service/internal/service"
	grpcsrv "github.com/jackietana/ticket-platform/catalog-service/internal/transport/grpc"
	"github.com/jackietana/ticket-platform/catalog-service/internal/transport/rest/v1"
	"github.com/jackietana/ticket-platform/pkg/database/psql"
	pkgMinio "github.com/jackietana/ticket-platform/pkg/minio"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const MAX_HEADER_SIZE = 5 * 1 << 20

func main() {
	// REST dependencies
	configPath := os.Getenv("APP_CONFIG_PATH")
	if configPath == "" {
		configPath = "./catalog-service/configs/local.yaml"
	}

	cfg, err := config.NewConfig(configPath)
	if err != nil {
		log.Fatalf("error creating config: %v", err)
	}

	if err := psql.RunUpMigrations(&cfg.Postgres, "catalog"); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	psqlDB, err := psql.NewPostgresConnection(&cfg.Postgres)
	if err != nil {
		log.Fatalf("error connecting to db: %v", err)
	}

	minio, err := pkgMinio.NewMinioClient(&cfg.Minio)
	if err != nil {
		log.Fatalf("failed to connect to minio: %v", err)
	}

	conn, err := grpc.NewClient(cfg.GetAuthServiceEndpoint(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to create gRPC channel: %v", err)
	}
	defer conn.Close()

	authClient := authv1.NewAuthServiceClient(conn)

	repository := repository.NewRepository(context.Background(), psqlDB, minio, cfg.Minio)
	if err := repository.InitStorage(context.Background()); err != nil {
		log.Fatalf("failed to init minio storage: %v", err)
	}

	service := service.NewService(repository, repository)
	handler := rest.NewHandler(authClient, service)

	router := handler.InitRouter()
	httpSrv := &http.Server{
		Addr:           ":" + cfg.Server.RESTPort,
		Handler:        router,
		MaxHeaderBytes: MAX_HEADER_SIZE,
	}

	go func() {
		log.Printf("REST server started on port %s", cfg.Server.RESTPort)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("error starting server: %v", err)
		}
	}()

	// gRPC dependencies
	lis, err := net.Listen("tcp", ":"+cfg.Server.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCatalogServiceServer(grpcServer, grpcsrv.NewCatalogServer(service))

	go func() {
		log.Printf("gRPC server started on port %s", cfg.Server.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
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
