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

	pb "github.com/jackietana/ticket-platform/api/gen/authv1"
	"github.com/jackietana/ticket-platform/auth-service/internal/config"
	"github.com/jackietana/ticket-platform/auth-service/internal/repository"
	"github.com/jackietana/ticket-platform/auth-service/internal/service"
	grpcsrv "github.com/jackietana/ticket-platform/auth-service/internal/transport/grpc"
	"github.com/jackietana/ticket-platform/auth-service/internal/transport/rest/v1"
	"github.com/jackietana/ticket-platform/auth-service/pkg/hash"
	pkgcache "github.com/jackietana/ticket-platform/pkg/cache"
	pkgpsql "github.com/jackietana/ticket-platform/pkg/database"
	"google.golang.org/grpc"
)

// @title Auth Service API
// @version 1.0
// @description Authentication service providing user registration, login and token validation via PostgreSQL and Redis.
//
// @host localhost:8080
// @BasePath /
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
//
// @schemes http
//
// @contact.name API Support
// @contact.email support@example.com

func main() {
	// REST deps
	configPath := os.Getenv("APP_CONFIG_PATH")
	if configPath == "" {
		configPath = "./auth-service/configs/local.yaml"
	}

	cfg, err := config.NewConfig(configPath)
	if err != nil {
		log.Fatalf("error creating config: %v", err)
	}

	if err := pkgpsql.RunUpMigrations(&cfg.Postgres, "auth"); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	hasher := hash.NewSHA1Hasher(cfg.Salt)
	cacheDB := pkgcache.NewRedisConnection(fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port), cfg.Redis.Pass)

	repoDB, err := pkgpsql.NewPostgresConnection(&cfg.Postgres)
	if err != nil {
		log.Fatalf("error connecting to db: %v", err)
	}

	repo := repository.NewRepository(repoDB)
	cacher := repository.NewCache(cacheDB)
	authService := service.NewAuthService(hasher, repo, cacher)
	handler := rest.NewHandler(authService)

	router := handler.Init()
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

	// gRPC deps
	lis, err := net.Listen("tcp", ":"+cfg.Server.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, grpcsrv.NewAuthServer(authService))

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
