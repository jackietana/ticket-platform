# TicketPlatform
A microservice platform for booking, selling, and managing event tickets.

TicketPlatform is a highly scalable backend developed using modern architectural patterns, ensuring component independence and fault tolerance. It is designed to process complex, real-time ticketing and booking logic.

***

## Overview and Architecture
The project is built using the Go language and employs the microservices pattern, where each functional module (authentication, catalog, orders) is a self-contained, independent service.

### Technology Stack
*   **Language:** Go (Golang) 1.26+
*   **Communication:** gRPC (for high-efficiency inter-service communication) and REST (for public APIs).
*   **Message Broker:** RabbitMQ 4.0+ (for asynchronous event-driven communication and decoupling background tasks).
*   **Databases:**
    *   PostgreSQL 18+ (For transactional data storage, using `golang-migrate`).
    *   Redis 8+ (For caching sessions, rate limiting, and temporary seat/slot booking with TTL).
*   **Object Storage:** Minio (For storing posters, media files, and generated tickets in json format).
*   **Infrastructure:** Docker, Docker Compose (for full containerization).

### Implemented Services
1.  **`auth-service` (Authentication Service):**
    *   Handles the entire user lifecycle: registration, login, and profile management.
    *   **Security:** Implements mechanisms to prevent session hijacking by binding sessions to the user's IP address and User-Agent.
    *   **Reliability:** Features automatic database migration execution upon startup.
2.  **`catalog-service` (Catalog Service):**
    *   Manages the event inventory, including category management and detailed event listings.
    *   **Media Support:** Handles poster uploads and integration with object storage for media files.
    *   **Core Features:** Supports creation of events with specific pricing, date scheduling, and capacity tracking.
3.  **`order-service` (Order Management Service):**
    *   Manages ticket reservations and orders using a high-performance, resilient flow.
    *   **Concurrency & Performance:** Implements a pipeline-based temporary seat-booking mechanism in Redis with a 10-minute TTL to handle concurrent high-load traffic safely.
    *   **Resiliency (Fault Tolerance):** Integrates an abstract `PaymentProvider` layer with simulated payment gateway failures (insufficient funds, network blips) to test robust transactional state rollbacks.
    *   **Event-Driven Background Worker:** Features an independent RabbitMQ Consumer/Worker that safely processes asynchronous `order.paid` events, dynamically generates electronic json tickets in memory, and offloads them to Minio object storage.

***

## Getting Started Guide

Depending on your goal (local development or production simulation), choose the appropriate setup guide below.

### 1. For Local Development (Development Mode)
This mode is ideal for quick debugging and testing, as services are run directly on the host machine.

1.  **Environment Setup:** Create a `.env.local` file in the project root using the provided template, setting `localhost` for DB hosts, Redis and RabbitMQ endpoints.
2.  **Infrastructure Spin-up:** Start PostgreSQL, Redis, RabbitMQ, and MinIO in detached mode:
    ```bash
    docker compose up -d
    ```
3.  **Service Launch:** Run the services directly on the host:
    ```bash
    source .env.local && go run auth-service/cmd/main.go
    source .env.local && go run catalog-service/cmd/main.go
    source .env.local && go run order-service/cmd/main.go
    ```

### 2. For Production-style Testing (Containerized)
Recommended for testing the full interaction flow of all components within an isolated Docker network.

1.  **Compilation:** Compile the binary file for the target container architecture:
    ```bash
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ./auth-service/auth-app ./auth-service/cmd/main.go
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ./catalog-service/catalog-app ./catalog-service/cmd/main.go
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ./order-service/order-app ./order-service/cmd/main.go
    ```
2.  **Build and Run:** Execute the full build and startup (Docker Compose automatically orchestrates healthchecks, ensuring services wait until RabbitMQ and DBs are healthy):
    ```bash
    docker compose up -d --build
    ```

***

## Development Roadmap
The planned development phases focus on maximizing modularity and decoupling services to ensure the platform can handle unpredictable, high-load traffic.

### Completed Modules
*   [x] Base Infrastructure (Docker Compose, Postgres, Redis).
*   [x] `auth-service`: Fully functional authentication system with session protection.
*   [x] **Media Storage**: Integration of an object storage system (Minio) for storing posters and media files.
*   [x] `catalog-service`: Implementation of the core service for managing events, categories, and ticket inventory.
*   [x] **Asynchronous Communication**: Integration of a message broker (RabbitMQ) for service decoupling.
*   [x] `order-service`: Implementation of the order management system with Redis booking pipelines and async ticket generation.
*   [x] **Transactions & Payments**: Integration of a simulated payment gateway handler with transactional rollback capabilities.

### Next Steps
*   [ ] **Physical Worker Decoupling**: Separate the `order-service` API and RabbitMQ Consumer into distinct physical Docker containers to allow independent horizontal scaling.
*   [ ] **Full Payment Microservice**: Extract the payment logic from `order-service` into a standalone `payment-service` interacting via gRPC / Webhooks.
