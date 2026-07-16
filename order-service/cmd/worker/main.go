package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackietana/ticket-platform/order-service/internal/config"
	"github.com/jackietana/ticket-platform/order-service/internal/rabbitmq"
	"github.com/jackietana/ticket-platform/order-service/internal/repository"
	"github.com/jackietana/ticket-platform/order-service/pkg/mq"
	pkgminio "github.com/jackietana/ticket-platform/pkg/minio"
)

func main() {
	// RabbitMQ dependencies
	configPath := os.Getenv("APP_CONFIG_PATH")
	if configPath == "" {
		configPath = "./order-service/configs/local.yaml"
	}

	cfg, err := config.NewConfig(configPath)
	if err != nil {
		log.Fatalf("error creating config: %v", err)
	}

	mqChan, err := mq.NewRabbitMQChannel(cfg.GetRabbitmqEndpoint())
	if err != nil {
		log.Fatalf("failed to create rabbitmq channel: %v", err)
	}

	minioClient, err := pkgminio.NewMinioClient(&cfg.Minio)
	if err != nil {
		log.Fatalf("failed to connect to minio: %v", err)
	}

	rabbitmqCtx, cancel := context.WithCancel(context.Background())

	fileStorage := repository.NewMinioStorage(minioClient)
	if err := fileStorage.InitStorage(rabbitmqCtx); err != nil {
		log.Fatalf("failed to init minio storage: %v", err)
	}

	mqServer := rabbitmq.NewOrderConsumer(mqChan, fileStorage)

	go func() {
		if err := mqServer.StartListen(rabbitmqCtx); err != nil {
			log.Fatalf("failed to listed rabbitmq: %v", err)
		}
		log.Println("rabbitmq server started")
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Println("received stop signal, closing rabbitmq server")

	cancel()
}
