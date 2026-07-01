# TicketPlatform
A microservice platform for booking, selling, and managing event tickets.

TicketPlatform is a highly scalable backend developed using modern architectural patterns, ensuring component independence and fault tolerance. It is designed to process complex, real-time ticketing and booking logic.

***

## Overview and Architecture
The project is built using the Go language and employs the microservices pattern, where each functional module (authentication, catalog, orders) is a self-contained, independent service.

### Technology Stack
*   **Language:** Go (Golang) 1.26+
*   **Communication:** gRPC (for high-efficiency inter-service communication) and REST (for public APIs).
*   **Databases:**
    *   PostgreSQL 18+ (For transactional data storage, using `golang-migrate`).
    *   Redis 8+ (For caching sessions, rate limiting, and message queues).
*   **Infrastructure:** Docker, Docker Compose (for full containerization).

### Implemented Services
1.  **`auth-service` (Authentication Service):**
    *   Handles the entire user lifecycle: registration, login, and profile management.
    *   **Security:** Implements mechanisms to prevent session hijacking by binding sessions to the user's IP address and User-Agent.
    *   **Reliability:** Features automatic database migration execution upon startup.

***

## Getting Started Guide

Depending on your goal (local development or production simulation), choose the appropriate setup guide below.

### 1. For Local Development (Development Mode)
This mode is ideal for quick debugging and testing, as services are run directly on the host machine.

1.  **Environment Setup:** Create a `.env.local` file in the project root using the provided template, setting `localhost` for DB hosts.
2.  **Infrastructure Spin-up:** Start PostgreSQL and Redis in detached mode:
    ```bash
    docker compose up -d
    ```
3.  **Service Launch:** Run the `auth-service` directly on the host:
    ```bash
    source .env.local && go run auth-service/cmd/main.go
    ```

### 2. For Production-style Testing (Containerized)
Recommended for testing the full interaction flow of all components within an isolated Docker network.

1.  **Compilation:** Compile the binary file for the target container architecture:
    ```bash
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ./auth-service/auth-app ./auth-service/cmd/main.go
    ```
2.  **Build and Run:** Execute the full build and startup:
    ```bash
    docker compose up -d --build
    ```

***

## Development Roadmap
The planned development phases focus on maximizing modularity and decoupling services to ensure the platform can handle unpredictable, high-load traffic.

### Completed Modules
*   [x] Base Infrastructure (Docker Compose, Postgres, Redis).
*   [x] `auth-service`: Fully functional authentication system with session protection.

### Next Steps
*   [ ] **`catalog-service`**: Implementation of the core service for managing events, categories, and ticket inventory.
*   [ ] **Asynchronous Communication**: Integration of a message broker (RabbitMQ / Kafka) for service decoupling.
*   [ ] **Transactions & Payments**: Development of `order-service` (order management) and `payment-service` (mock payment gateway integration).
*   [ ] **Media Storage**: Integration of an object storage system (Minio / S3) for storing posters and media files.
