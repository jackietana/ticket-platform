package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackietana/ticket-platform/api/gen/orderv1"
	"github.com/jackietana/ticket-platform/order-service/internal/config"
	"github.com/jackietana/ticket-platform/order-service/internal/rabbitmq"
	"github.com/jackietana/ticket-platform/order-service/internal/repository"
	"github.com/jackietana/ticket-platform/order-service/internal/service"
	grpcsrv "github.com/jackietana/ticket-platform/order-service/internal/transport/grpc"
	"github.com/jackietana/ticket-platform/order-service/internal/transport/rest/v1"
	"github.com/jackietana/ticket-platform/order-service/pkg/mq"
	pkgcache "github.com/jackietana/ticket-platform/pkg/cache"
	pkgpsql "github.com/jackietana/ticket-platform/pkg/database"
	pkgclient "github.com/jackietana/ticket-platform/pkg/grpc"
	"google.golang.org/grpc"
)

const ORDER_TTL = time.Minute * 10

func main() {
	// REST dependencies
	configPath := os.Getenv("APP_CONFIG_PATH")
	if configPath == "" {
		configPath = "./order-service/configs/local.yaml"
	}

	cfg, err := config.NewConfig(configPath)
	if err != nil {
		log.Fatalf("error creating config: %v", err)
	}

	if err := pkgpsql.RunUpMigrations(&cfg.Postgres, "order"); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	psqlDB, err := pkgpsql.NewPostgresConnection(&cfg.Postgres)
	if err != nil {
		log.Fatalf("error connecting to db: %v", err)
	}

	redisConn := pkgcache.NewRedisConnection(fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port), cfg.Redis.Pass)
	mqChan, err := mq.NewRabbitMQChannel(cfg.GetRabbitmqEndpoint())
	if err != nil {
		log.Fatalf("failed to create rabbitmq channel: %v", err)
	}

	catalogClient, err := pkgclient.NewCatalogClient(cfg.GetCatalogServiceEndpoint())
	if err != nil {
		log.Fatalf("failed to create catalog client: %v", err)
	}

	authClient, err := pkgclient.NewAuthClient(cfg.AuthService.GetAuthServiceEndpoint())
	if err != nil {
		log.Fatalf("failed to create auth client: %v", err)
	}

	psqlRepo := repository.NewPsqlRepository(psqlDB)
	redisRepo := repository.NewRedisRepository(redisConn)
	payProvider := repository.NewFakePaymentProvider()
	mqClient := rabbitmq.NewOrderProducer(mqChan)

	service := service.NewService(catalogClient, psqlRepo, redisRepo, ORDER_TTL, payProvider, mqClient)
	handler := rest.NewHandler(authClient, service)
	router := handler.InitRouter()
	httpSrv := &http.Server{
		Addr:    ":" + cfg.Server.RESTPort,
		Handler: router,
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
	orderv1.RegisterOrderServiceServer(grpcServer, grpcsrv.NewOrderServer(service))

	go func() {
		log.Printf("gRPC server started on port %s", cfg.Server.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
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
